package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/mail"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	smtp "github.com/emersion/go-smtp"
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Build information set via ldflags
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"

	// Error counter for health check
	errCount     = 0
	errCountLock = &sync.RWMutex{}
)

type metrics struct {
	// SQS
	sqsMessagesReceived prometheus.Counter
	sqsPollDuration     prometheus.Histogram
	sqsPollErrors       prometheus.Counter
	sqsDeleteErrors     prometheus.Counter

	// Processing
	emailsProcessed          prometheus.Counter
	messageProcessingErrors  prometheus.Counter
	emailProcessingDuration  prometheus.Histogram
	defaultMailboxDeliveries prometheus.Counter

	// S3
	s3FetchErrors prometheus.Counter
	s3GetDuration prometheus.Histogram
	s3ObjectSize  prometheus.Histogram

	// LMTP
	lmtpSendErrors   prometheus.Counter
	lmtpSendDuration prometheus.Histogram
}

func newMetrics(reg *prometheus.Registry, version, commit, buildDate string) *metrics {
	m := &metrics{
		sqsMessagesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sqs_messages_received_total",
			Help: "Total number of messages received from SQS.",
		}),
		sqsPollDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sqs_poll_duration_seconds",
			Help:    "Duration of SQS poll operations.",
			Buckets: prometheus.DefBuckets,
		}),
		sqsPollErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sqs_poll_errors_total",
			Help: "Total number of SQS poll errors.",
		}),
		sqsDeleteErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sqs_delete_errors_total",
			Help: "Total number of SQS message delete errors.",
		}),
		emailsProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "emails_processed_total",
			Help: "Total number of emails successfully processed end-to-end.",
		}),
		messageProcessingErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "message_processing_errors_total",
			Help: "Total number of message processing errors.",
		}),
		emailProcessingDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "email_processing_duration_seconds",
			Help:    "End-to-end duration of processing a single email message.",
			Buckets: prometheus.DefBuckets,
		}),
		defaultMailboxDeliveries: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "default_mailbox_deliveries_total",
			Help: "Total number of emails delivered to the default mailbox due to no matching recipient.",
		}),
		s3FetchErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "s3_fetch_errors_total",
			Help: "Total number of S3 object fetch errors.",
		}),
		s3GetDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "s3_get_duration_seconds",
			Help:    "Duration of S3 GetObject operations.",
			Buckets: prometheus.DefBuckets,
		}),
		s3ObjectSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "s3_object_size_bytes",
			Help:    "Size of S3 objects (email bodies) retrieved.",
			Buckets: prometheus.ExponentialBuckets(1024, 4, 8), // 1KB to ~16MB
		}),
		lmtpSendErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lmtp_send_errors_total",
			Help: "Total number of LMTP send errors.",
		}),
		lmtpSendDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "lmtp_send_duration_seconds",
			Help:    "Duration of LMTP send operations.",
			Buckets: prometheus.DefBuckets,
		}),
	}

	buildInfo := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "build_info",
		Help: "Build information.",
		ConstLabels: prometheus.Labels{
			"version":    version,
			"commit":     commit,
			"build_date": buildDate,
		},
	})
	buildInfo.Set(1)

	reg.MustRegister(
		buildInfo,
		m.sqsMessagesReceived,
		m.sqsPollDuration,
		m.sqsPollErrors,
		m.sqsDeleteErrors,
		m.emailsProcessed,
		m.messageProcessingErrors,
		m.emailProcessingDuration,
		m.defaultMailboxDeliveries,
		m.s3FetchErrors,
		m.s3GetDuration,
		m.s3ObjectSize,
		m.lmtpSendErrors,
		m.lmtpSendDuration,
	)

	return m
}

func main() {
	// Log build information
	slog.Info("build information", "version", version, "commit", commit, "buildDate", buildDate)

	sqsQueueURL := MustGetEnv("SQS_QUEUE_URL", nil)
	lmtpHost := MustGetEnv("LMTP_HOST", nil)
	lmtpFrom := MustGetEnv("LMTP_FROM", nil)
	healthCheckPort := MustGetEnv("HEALTH_CHECK_PORT", aws.String("8080"))
	mailboxes := Map(strings.Split(MustGetEnv("MAILBOXES", nil), ","), func(v string) string {
		return strings.TrimSpace(v)
	})
	defaultMailbox := MustGetEnv("DEFAULT_MAILBOX", nil)

	slog.Info("starting up", "config", map[string]string{
		"mailboxes":       strings.Join(mailboxes, ","),
		"defaultMailbox":  defaultMailbox,
		"lmtpHost":        lmtpHost,
		"lmtpFrom":        lmtpFrom,
		"sqsQueueURL":     sqsQueueURL,
		"healthCheckPort": healthCheckPort,
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		slog.Info("main is exiting; cancelling the context")
		cancel()
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		slog.Info("received shutdown signal", "signal", sig)
		cancel() // Cancel the context to trigger graceful shutdown
	}()

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	Check(err, "failed to load AWS config")

	// Create AWS service clients
	sqsClient := sqs.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	lmtpSender := newLMTPSender(lmtpHost, lmtpFrom)

	// Set up Prometheus registry and metrics
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	m := newMetrics(reg, version, commit, buildDate)

	processMessage := newMessageProcessor(mailboxes, defaultMailbox, s3Client, lmtpSender, m)

	// Start HTTP server
	httpServer := &http.Server{
		Addr: ":" + healthCheckPort,
	}
	http.HandleFunc("/stats.json", statsHandler)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	go func() {
		slog.Info("starting http server", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "err", err)
		}
	}()

	slog.Info("starting ses to lmtp forwarder")

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down gracefully...")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				slog.Error("failed to shutdown http server", "err", err)
			}
			return
		default:
			slog.Info("polling messages from sqs")
			pollStart := time.Now()
			result, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(sqsQueueURL),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20, // Long polling
			})
			m.sqsPollDuration.Observe(time.Since(pollStart).Seconds())

			if err != nil {
				if ctx.Err() != nil {
					slog.Info("context cancelled, stopping queue polling", "ctxErr", err.Error())
					return
				}
				slog.Error("failed to receive messages from sqs", "err", err)
				m.sqsPollErrors.Inc()
				errCountLock.Lock()
				errCount++
				errCountLock.Unlock()
				time.Sleep(time.Second)
				continue
			}

			slog.Info("polled messages from sqs", "count", len(result.Messages))
			m.sqsMessagesReceived.Add(float64(len(result.Messages)))
			errCountLock.Lock()
			errCount = 0
			errCountLock.Unlock()

			// Process each message
			for _, message := range result.Messages {
				// Check context before processing each message
				if ctx.Err() != nil {
					slog.Info("context cancelled, stopping message processing", "ctxErr", err.Error())
					return
				}

				slog.Info("processing message")
				processStart := time.Now()
				if err := processMessage(ctx, message); err != nil {
					slog.Error("failed to process message", "err", err)
					m.messageProcessingErrors.Inc()
					continue
				}
				m.emailProcessingDuration.Observe(time.Since(processStart).Seconds())
				slog.Info("processed message")

				slog.Info("deleting message")
				_, err = sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(sqsQueueURL),
					ReceiptHandle: message.ReceiptHandle,
				})
				if err != nil {
					slog.Info("failed to delete message from queue", "err", err)
					m.sqsDeleteErrors.Inc()
					continue
				}
				m.emailsProcessed.Inc()
				slog.Info("deleted message")
			}
		}
	}
}

func newMessageProcessor(
	mailboxes []string,
	defaultMailbox string,
	s3Client *s3.Client,
	emailSender func(to []string, body io.Reader) error,
	m *metrics,
) func(ctx context.Context, message sqsTypes.Message) error {
	return func(ctx context.Context, message sqsTypes.Message) error {
		// Check if context is cancelled before processing
		if ctx.Err() != nil {
			return ctx.Err()
		}

		slog.Info("parsing message as sns entity", "message", message)
		var snsEntity events.SNSEntity
		if err := json.Unmarshal([]byte(Value(message.Body)), &snsEntity); err != nil {
			return fmt.Errorf("failed to unmarshal sns entity: %w", err)
		}
		slog.Info("parsed message", "entity", snsEntity)

		slog.Info("parsing entity message as ses event")
		var sesEvent events.SimpleEmailService
		if err := json.Unmarshal([]byte(snsEntity.Message), &sesEvent); err != nil {
			return fmt.Errorf("failed to unmarshal ses entity: %w", err)
		}
		slog.Info("parsed entity message as ses event", "sesEvent", sesEvent)

		if at := sesEvent.Receipt.Action.Type; at != "S3" {
			slog.Error("unsupported action type", "type", at)
			return nil
		}

		slog.Info("getting mail body from s3")
		s3Start := time.Now()
		goOut, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(sesEvent.Receipt.Action.BucketName),
			Key:    aws.String(sesEvent.Receipt.Action.ObjectKey),
		})
		m.s3GetDuration.Observe(time.Since(s3Start).Seconds())
		if err != nil {
			m.s3FetchErrors.Inc()
			return fmt.Errorf("failed to get object from s3: %w", err)
		}
		defer func() {
			if err := goOut.Body.Close(); err != nil {
				slog.Warn("failed to close S3 object body", "err", err)
			}
		}()
		if goOut.ContentLength != nil {
			m.s3ObjectSize.Observe(float64(*goOut.ContentLength))
		}
		slog.Info("got mail body from s3", "data", goOut)

		slog.Info("reading s3 object body")
		emailBody, err := io.ReadAll(goOut.Body)
		if err != nil {
			return fmt.Errorf("failed to get s3 object body: %w", err)
		}
		slog.Info("read s3 object body", "bodyLength", len(emailBody))

		slog.Info("parsing email body")
		emailMsg, err := mail.ReadMessage(bytes.NewBuffer(emailBody))
		if err != nil {
			return fmt.Errorf("failed to read email: %v", err)
		}
		slog.Info("parsed email", "email", emailMsg)

		slog.Info("got recipients from ses event", "recipients", sesEvent.Receipt.Recipients)

		slog.Info("filtering recipients")
		recipients := Filter(sesEvent.Receipt.Recipients, func(r string) bool {
			return Contains(mailboxes, r)
		})
		if len(recipients) == 0 {
			slog.Info("no valid recipients found, using default mailbox", "defaultMailbox", defaultMailbox)
			recipients = []string{defaultMailbox}
			m.defaultMailboxDeliveries.Inc()
		}
		slog.Info("filtered recipients", "recipients", recipients)

		slog.Info("sending email")
		lmtpStart := time.Now()
		if err := emailSender(recipients, bytes.NewBuffer(emailBody)); err != nil {
			m.lmtpSendErrors.Inc()
			return fmt.Errorf("failed to send email: %w", err)
		}
		m.lmtpSendDuration.Observe(time.Since(lmtpStart).Seconds())
		slog.Info("sent email")
		return nil
	}
}

func newLMTPSender(host, from string) func(to []string, body io.Reader) error {
	return func(to []string, body io.Reader) error {
		conn, err := net.Dial("tcp", host)
		Check(err, "failed to dial")

		lmtpClient := smtp.NewClientLMTP(conn)
		defer func() {
			_ = lmtpClient.Quit()
			_ = conn.Close()
		}()
		return lmtpClient.SendMail(from, to, body)
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	errCountLock.RLock()
	errCount := errCount
	errCountLock.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"healthy":    errCount < 3,
		"errorCount": errCount,
	})
}
