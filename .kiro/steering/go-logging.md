---
inclusion: always
---

# Go Logging Standards

## Use slog for Structured Logging

All Go code in this project should use the `log/slog` package for structured logging instead of the traditional `log` package.

### Benefits

- Structured logging with key-value pairs for better log parsing and analysis
- Multiple output formats (JSON, text)
- Configurable log levels (Debug, Info, Warn, Error)
- Better integration with modern observability tools

### Usage

```go
import (
    "log/slog"
    "os"
)

// Initialize logger (typically in main)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Log with structured fields
logger.Info("server started", "address", addr, "port", port)
logger.Error("failed to connect", "error", err, "host", hostname)
```

### Guidelines

- Use JSON handler for production services
- Include relevant context as key-value pairs
- Use appropriate log levels (Debug, Info, Warn, Error)
- Pass logger instances to handlers and functions that need logging
- Avoid string formatting in log messages; use structured fields instead
- Log messages should start with lowercase letters (e.g., "starting server" not "Starting server")
