# SES to LMTP Forwarder

A Go application that retrieves SES messages from an SQS queue, fetches email bodies from S3, and forwards them to a local LMTP server.

## Features

- Polls SQS for SES notification messages
- Retrieves email content from S3
- Forwards emails via LMTP protocol
- Graceful shutdown on SIGINT/SIGTERM signals
- HTTP health check endpoint at `/stats.json`
- Runs as non-root user for security
- Docker health checks included

## Quick Start

```bash
# Pull the image
docker pull harrisonhjones/ses2lmtp:latest

# Create environment file
cat > .env << EOF
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/your-queue
LMTP_HOST=192.168.0.123:31024
LMTP_FROM=sqs2lmtp@domain.tld
MAILBOXES=user1@domain.tld,user2@domain.tld
DEFAULT_MAILBOX=default@domain.tld
EOF

# Run the container
docker run -d \
  --name ses2lmtp \
  --env-file .env \
  -p 8080:8080 \
  harrisonhjones/ses2lmtp:latest
```

## Configuration

### Required Environment Variables

- `SQS_QUEUE_URL`: SES SQS queue URL
- `LMTP_HOST`: LMTP server host and port (e.g., 192.168.0.123:31024)
- `LMTP_FROM`: From address for LMTP forwarding
- `MAILBOXES`: Comma-separated list of allowed mailboxes
- `DEFAULT_MAILBOX`: Default mailbox for forwarding

### Optional Environment Variables

- `AWS_REGION`: AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID`: AWS access key (or use IAM roles)
- `AWS_SECRET_ACCESS_KEY`: AWS secret key (or use IAM roles)

## Health Check

The container includes a health check that queries the `/stats.json` endpoint:

```bash
# Check health status
docker inspect --format='{{.State.Health.Status}}' ses2lmtp

# Query stats endpoint directly
curl http://localhost:8080/stats.json
```

## Tags

- `latest`: Latest stable release
- `latest-dev`: Latest development build from main branch
- `v*.*.*`: Specific version tags
- `<commit-sha>`: Development builds from main branch (specific commits)

## Source Code

Full documentation and source code available at: https://github.com/harrisonhjones/ses2lmtp

## License

See the [GitHub repository](https://github.com/harrisonhjones/ses2lmtp) for license information.
