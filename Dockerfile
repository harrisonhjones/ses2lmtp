# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the application with version information
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.buildDate=${BUILD_DATE}'" -o ses2lmtp .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests, curl for healthcheck, and jq for JSON parsing
RUN apk --no-cache add ca-certificates curl jq

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/ses2lmtp .

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Change ownership of the app directory
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${HEALTH_CHECK_PORT:-8080}/stats.json | jq -e '.healthy == true' || exit 1

# Run the binary
CMD ["./ses2lmtp"]
