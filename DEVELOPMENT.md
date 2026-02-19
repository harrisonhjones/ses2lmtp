# Development Guide

This guide covers local development, testing, building, and publishing the SES to LMTP Forwarder.

## Local Development

### Prerequisites

- Go 1.23 or later
- Docker (optional, for container testing)
- AWS credentials with access to SQS and S3

### Setup

1. **Clone the repository**

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run locally:**
   ```bash
   go run .
   ```

5. **Build locally:**
   ```bash
   go build -o ses2lmtp .
   ./ses2lmtp
   ```

### Build with Version Information

To build with embedded version information:

```bash
go build -ldflags="-X 'main.version=v1.0.0' -X 'main.commit=$(git rev-parse HEAD)' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o ses2lmtp .
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

Generate and view a detailed HTML coverage report:

```bash
# Generate coverage profile
go test -coverprofile coverage.out

# Open HTML coverage report in browser
go tool cover -html coverage.out
```

The HTML report shows:
- Line-by-line coverage highlighting
- Coverage percentages for each function
- Uncovered code paths in red
- Covered code paths in green

### Function-Level Coverage

To see coverage breakdown by function:

```bash
go tool cover -func coverage.out
```

## Docker Development

### Build and Run Locally

```bash
# Build the Docker image
docker build -t ses2lmtp .

# Run in foreground with environment file
docker run --env-file .env -p 8080:8080 ses2lmtp

# Run in background
docker run -d --env-file .env -p 8080:8080 --name ses2lmtp ses2lmtp

# View logs
docker logs -f ses2lmtp

# Stop and remove
docker stop ses2lmtp
docker rm ses2lmtp
```

### Build and Run in One Command

```bash
# Windows CMD
docker build -t ses2lmtp . & docker run --rm --env-file .env -p 8080:8080 ses2lmtp

# PowerShell
docker build -t ses2lmtp . ; docker run --rm --env-file .env -p 8080:8080 ses2lmtp
```

### Test Health Check

```bash
# Query the stats endpoint
curl http://localhost:8080/stats.json

# Check Docker health status
docker inspect --format='{{.State.Health.Status}}' ses2lmtp
```

## Publishing to Docker Hub

The image is published to [harrisonhjones/ses2lmtp](https://hub.docker.com/r/harrisonhjones/ses2lmtp/).

### Automated Publishing with GitHub Actions

The repository uses GitHub Actions to automatically build and push Docker images:

- **On push to `main`**: Builds and pushes with `latest-dev` tag and commit SHA
- **On GitHub release**: Builds and pushes with `latest` and the release tag

#### Setup Requirements

Configure these secrets in your GitHub repository settings (Settings → Secrets and variables → Actions):

- `DOCKERHUB_USERNAME`: Your Docker Hub username
- `DOCKERHUB_TOKEN`: A Docker Hub access token (create at https://hub.docker.com/settings/security)

### Manual Publishing

You can also publish manually:

1. **Build the image with tags:**
   ```bash
   docker build -t harrisonhjones/ses2lmtp:latest -t harrisonhjones/ses2lmtp:v1.0.0 .
   ```

2. **Login to Docker Hub:**
   ```bash
   docker login
   ```

3. **Push the image:**
   ```bash
   docker push harrisonhjones/ses2lmtp:latest
   docker push harrisonhjones/ses2lmtp:v1.0.0
   ```

## Project Structure

```
.
├── main.go              # Main application logic
├── util.go              # Utility functions
├── util_test.go         # Unit tests
├── Dockerfile           # Docker build configuration
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── .env.example         # Example environment configuration
├── README.md            # User-facing documentation
├── DEVELOPMENT.md       # This file
├── DOCKERHUB.md         # Docker Hub description
└── .github/
    └── workflows/
        └── docker-build.yml  # CI/CD pipeline
```

## Graceful Shutdown

The application supports graceful shutdown:
- Send SIGINT (Ctrl+C) or SIGTERM to trigger shutdown
- The application will finish processing current messages and exit cleanly
- All AWS operations respect the cancellation context
- HTTP server shuts down gracefully with a 5-second timeout

## Network Configuration

The Docker container can be run with different network configurations:

- `--network host`: Access local services directly (default in examples)
- Custom Docker network: For container-to-container communication
- Port mapping with `-p 8080:8080`: Expose health check endpoint

If your LMTP server is running in another container:
- Create a custom Docker network
- Update the `LMTP_HOST` environment variable to point to the container name

## Documentation

When making changes, remember to update:
- **README.md**: User-facing "what is this and how do I get started"
- **DEVELOPMENT.md**: This file - development and contribution guide
- **DOCKERHUB.md**: Docker Hub repository description
- **.kiro/steering/documentation-updates.md**: If adding new .md files
