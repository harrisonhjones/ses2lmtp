package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	sqsQueueURL := MustGetEnv("SQS_QUEUE_URL", nil)
	//lmtpHost := MustGetEnv("LMTP_HOST", nil)
	//lmtpPort := MustGetEnv("LMTP_PORT", nil)

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	Check(err, "failed to load AWS config")

	// Create AWS service clients
	sqsClient := sqs.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	processMessage := newMessageProcessor(s3Client)

	slog.Info("starting ses to lmtp forwarder")

	// Main processing loop
	for {
		// Receive messages from SQS
		result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsQueueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // Long polling
		})
		if err != nil {
			slog.Error("failed to receive messages from sqs", "err", err)
			time.Sleep(time.Second)
			continue
		}

		// Process each message
		for _, message := range result.Messages {
			if err := processMessage(message); err != nil {
				slog.Error("failed to process message", "err", err)
				continue
			}

			/*
							Bucket: aws.String(sesEvent.Receipt.Action.BucketName),
				Key:    aws.String(sesEvent.Receipt.Action.ObjectKey),
			*/
			//var sesEvent events.SimpleEmailService
			/*
				err := processMessage(s3Client, message, lmtpHost+":"+lmtpPort)
				if err != nil {
					log.Printf("Error processing message: %v", err)
					continue
				}

				// Delete message from queue after successful processing
				_, err = sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(sqsQueueURL),
					ReceiptHandle: message.ReceiptHandle,
				})
				if err != nil {
					log.Printf("Failed to delete message from queue: %v", err)
				}*/
		}
	}
}

func newMessageProcessor(s3Client *s3.Client) func(message sqsTypes.Message) error {
	return func(message sqsTypes.Message) error {
		slog.Info("got message", "message", message)

		slog.Info("parsing message as sns entity")
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
		goOut, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
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

		return nil
	}
}
