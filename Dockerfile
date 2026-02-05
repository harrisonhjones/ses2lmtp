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

# Install ca-certificates for HTTPS requests and procps for pgrep
RUN apk --no-cache add ca-certificates procps

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/ses2lmtp .

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD pgrep ses2lmtp || exit 1

# Run the binary
CMD ["./ses2lmtp"]
