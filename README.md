# SES to LMTP Forwarder

A Go application that retrieves SES messages from an SQS queue, fetches email bodies from S3, and forwards them to a local LMTP server. The application includes graceful shutdown handling and can be run locally or in a Docker container.

## Features

- Polls SQS for SES notification messages
- Retrieves email content from S3
- Graceful shutdown on SIGINT/SIGTERM signals
- Context-based cancellation throughout the application
- Docker support with health checks

## Quick Start with Docker

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your configuration:**
   - Set your AWS credentials
   - Update the SQS queue URL
   - Configure LMTP server details (if needed)

3. **Build and run:**
   ```bash
   # Build the Docker image
   docker build -t ses2lmtp .
   
   # Run the container
   docker run --env-file .env --network host ses2lmtp
   ```

## Configuration

### Environment Variables

- `AWS_REGION`: AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `SQS_QUEUE_URL`: SES SQS queue URL (required)
- `LMTP_HOST`: LMTP server host and port (e.g., 192.168.0.123:31024)
- `LMTP_FROM`: From address for LMTP forwarding (e.g., sqs2lmtp@domain1.tld)
- `MAILBOXES`: Comma-separated list of allowed mailboxes (e.g., mb1@domain2.tld,mb2@domain3.tld)
- `DEFAULT_MAILBOX`: Default mailbox for forwarding (e.g., user@domain2.tld)
- `HEALTH_CHECK_PORT`: Port for health check endpoint (optional, defaults to 8080)

### AWS Credentials

You can provide AWS credentials in several ways:

1. **Environment variables** (as shown in .env)
2. **AWS credentials file** (mount `~/.aws` to `/root/.aws` in container)
3. **IAM roles** (if running on EC2 or EKS)

## Docker Commands

```bash
# Build the image
docker build -t ses2lmtp .

# Run in foreground with environment file
docker run --env-file .env ses2lmtp

# Run in background
docker run -d --env-file .env --name ses2lmtp ses2lmtp

# View logs
docker logs -f ses2lmtp

# Stop the container
docker stop ses2lmtp

# Remove the container
docker rm ses2lmtp
```

## Local Development

```bash
# Install dependencies
go mod tidy

# Run locally (make sure .env is configured)
go run .

# Build locally
go build -o ses2lmtp .
./ses2lmtp
```

## Graceful Shutdown

The application supports graceful shutdown:
- Send SIGINT (Ctrl+C) or SIGTERM to trigger shutdown
- The application will finish processing current messages and exit cleanly
- All AWS operations respect the cancellation context

## Network Configuration

The Docker container uses `--network host` to access local services. If your LMTP server is running in another container, you may need to:
- Use a custom Docker network
- Update the network configuration accordingly
- Modify the LMTP_HOST environment variable to point to the correct container

## Health Check

The Docker image includes a health check that queries the `/stats.json` endpoint on port 8080. The health check runs every 30 seconds and will mark the container as unhealthy if the endpoint returns an unhealthy status or fails to respond.

You can manually check the health status:
```bash
# Query the stats endpoint
curl http://localhost:8080/stats.json

# Check Docker health status
docker inspect --format='{{.State.Health.Status}}' ses2lmtp
```

## Publishing to Docker Hub

The image is published to [harrisonhjones/ses2lmtp](https://hub.docker.com/r/harrisonhjones/ses2lmtp/).

### Publishing a New Version

1. **Build the image with tags:**
   ```bash
   docker build -t harrisonhjones/ses2lmtp:latest .
   ```

2. **Login to Docker Hub:**
   ```bash
   docker login
   ```

3. **Push the image:**
   ```bash
   docker push harrisonhjones/ses2lmtp:latest
   ```

### Using the Published Image

Instead of building locally, you can pull and run the published image:

```bash
# Pull the latest version
docker pull harrisonhjones/ses2lmtp:latest

# Run the published image
docker run --env-file .env -p 8080:8080 harrisonhjones/ses2lmtp:latest
```

## Testing

The project includes comprehensive unit tests for all utility functions.

### Running Tests

```bash
# Run all tests
go test

# Run tests with verbose output
go test -v

# Run tests with coverage
go test -cover
```

### Test Coverage Visualization

To generate and view a detailed HTML coverage report:

```bash
# Generate coverage profile
go test -coverprofile coverage.out

# Open HTML coverage report in browser
go tool cover -html coverage.out
```

The HTML report will show:
- Line-by-line coverage highlighting
- Coverage percentages for each function
- Uncovered code paths in red
- Covered code paths in green

### Function-Level Coverage

To see coverage breakdown by function:

```bash
go tool cover -func coverage.out
```
