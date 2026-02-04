# SES to LMTP Forwarder

A Go application that retrieves SES messages from an SQS queue, fetches email bodies from S3, and forwards them to a local LMTP server.

## Quick Start with Docker

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your configuration:**
   - Set your AWS credentials
   - Update the SQS queue URL
   - Configure LMTP server details

3. **Build and run:**
   ```bash
   docker-compose up --build
   ```

## Configuration

### Environment Variables

- `AWS_REGION`: AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `SQS_QUEUE_URL`: SES SQS queue URL
- `LMTP_HOST`: LMTP server host (default: 192.168.0.123)
- `LMTP_PORT`: LMTP server port (default: 31024)

### AWS Credentials

You can provide AWS credentials in several ways:

1. **Environment variables** (as shown in .env)
2. **AWS credentials file** (uncomment the volume mount in docker-compose.yml)
3. **IAM roles** (if running on EC2)

## Docker Commands

```bash
# Build the image
docker-compose build

# Run in foreground
docker-compose up

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down

# Rebuild and restart
docker-compose up --build
```

## Local Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run main.go
```

## Network Configuration

The container uses `network_mode: host` to access your local LMTP server. If your LMTP server is running in another container, you may need to adjust the network configuration.