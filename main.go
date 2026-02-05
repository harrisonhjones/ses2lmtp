package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/mail"
	"os"
	"os/signal"
	"strings"
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
)

func main() {
	sqsQueueURL := MustGetEnv("SQS_QUEUE_URL", nil)
	lmtpHost := MustGetEnv("LMTP_HOST", nil)
	lmtpFrom := MustGetEnv("LMTP_FROM", nil)
	mailboxes := Map(strings.Split(MustGetEnv("MAILBOXES", nil), ","), func(v string) string {
		return strings.TrimSpace(v)
	})
	defaultMailbox := MustGetEnv("DEFAULT_MAILBOX", nil)

	slog.Info("starting up", "config", map[string]string{
		"mailboxes":      strings.Join(mailboxes, ","),
		"defaultMailbox": defaultMailbox,
		"lmtpHost":       lmtpHost,
		"lmtpFrom":       lmtpFrom,
		"sqsQueueURL":    sqsQueueURL,
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	processMessage := newMessageProcessor(mailboxes, defaultMailbox, s3Client, lmtpSender)

	slog.Info("starting ses to lmtp forwarder")

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down gracefully...")
			return
		default:
			slog.Info("polling messages from sqs")
			result, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(sqsQueueURL),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20, // Long polling
			})
			if err != nil {
				if ctx.Err() != nil {
					slog.Info("context cancelled, stopping message processing")
					return
				}
				slog.Error("failed to receive messages from sqs", "err", err)
				time.Sleep(time.Second)
				continue
			}
			slog.Info("polled messages from sqs", "count", len(result.Messages))

			// Process each message
			for _, message := range result.Messages {
				// Check context before processing each message
				if ctx.Err() != nil {
					slog.Info("context cancelled, stopping message processing")
					return
				}

				slog.Info("processing message")
				if err := processMessage(ctx, message); err != nil {
					slog.Error("failed to process message", "err", err)
					continue
				}
				slog.Info("processed message")

				slog.Info("deleting message")
				_, err = sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(sqsQueueURL),
					ReceiptHandle: message.ReceiptHandle,
				})
				if err != nil {
					slog.Info("failed to delete message from queue", "err", err)
					continue
				}
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
		goOut, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(sesEvent.Receipt.Action.BucketName),
			Key:    aws.String(sesEvent.Receipt.Action.ObjectKey),
		})
		if err != nil {
			return fmt.Errorf("failed to get object from s3: %w", err)
		}
		defer func() {
			if err := goOut.Body.Close(); err != nil {
				slog.Warn("failed to close S3 object body", "err", err)
			}
		}()
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
		}
		slog.Info("filtered recipients", "recipients", recipients)

		slog.Info("sending email")
		if err := emailSender(recipients, bytes.NewBuffer(emailBody)); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
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
