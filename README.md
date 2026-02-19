# SES to LMTP Forwarder

A Go application that retrieves SES messages from an SQS queue, fetches email bodies from S3, and forwards them to a local LMTP server.

## Features

- Polls SQS for SES notification messages
- Retrieves email content from S3
- Forwards emails via LMTP protocol
- Graceful shutdown on SIGINT/SIGTERM signals
- HTTP health check endpoint at `/stats.json`
- Runs as non-root user for security
- Docker support with automated health checks

## Quick Start

### Using Docker (Recommended)

1. **Pull the image:**
   ```bash
   docker pull harrisonhjones/ses2lmtp:latest
   ```

2. **Create environment file:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run the container:**
   ```bash
   docker run -d \
     --name ses2lmtp \
     --env-file .env \
     -p 8080:8080 \
     harrisonhjones/ses2lmtp:latest
   ```

4. **Check health:**
   ```bash
   curl http://localhost:8080/stats.json
   ```

### Running Locally

```bash
# Install dependencies
go mod tidy

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Run
go run .
```

## Configuration

### Required Environment Variables

- `SQS_QUEUE_URL`: SES SQS queue URL
- `LMTP_HOST`: LMTP server host and port (e.g., `192.168.0.123:31024`)
- `LMTP_FROM`: From address for LMTP forwarding
- `MAILBOXES`: Comma-separated list of allowed mailboxes
- `DEFAULT_MAILBOX`: Default mailbox for forwarding

### Optional Environment Variables

- `AWS_REGION`: AWS region (default: `us-east-1`)
- `AWS_ACCESS_KEY_ID`: AWS access key (or use IAM roles)
- `AWS_SECRET_ACCESS_KEY`: AWS secret key (or use IAM roles)
- `HEALTH_CHECK_PORT`: HTTP server port (default: `8080`)

### AWS Credentials

You can provide AWS credentials in several ways:

1. Environment variables (as shown in `.env`)
2. AWS credentials file (mount `~/.aws` to `/root/.aws` in container)
3. IAM roles (if running on EC2 or EKS)

## Health Check

The application exposes a health check endpoint at `/stats.json`:

```bash
# Query the endpoint
curl http://localhost:8080/stats.json

# Check Docker container health
docker inspect --format='{{.State.Health.Status}}' ses2lmtp
```

## Available Docker Tags

- `latest`: Latest stable release
- `latest-dev`: Latest development build from main branch
- `v*.*.*`: Specific version tags (e.g., `v1.0.0`)
- `<commit-sha>`: Specific commit builds

## Development

For information on local development, testing, building, and publishing, see [DEVELOPMENT.md](DEVELOPMENT.md).

## License

See [LICENSE.md](LICENSE.md).
