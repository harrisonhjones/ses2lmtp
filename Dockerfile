# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o ses2lmtp .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests, curl for healthcheck, and jq for JSON parsing
RUN apk --no-cache add ca-certificates curl jq

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/ses2lmtp .

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl http://localhost:8080/stats.json | jq -e '.healthy == true' || exit 1

# Run the binary
CMD ["./ses2lmtp"]
