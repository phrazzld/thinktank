# Structured JSON Logging Guide

This document describes how to use structured JSON logging with correlation IDs in the thinktank project.

## Overview

The project uses a comprehensive structured logging system built on Go's `log/slog` package that provides:

- **JSON output** for machine-readable logs
- **Correlation ID support** for tracing request flow
- **Context-aware logging** methods
- **Stream separation** (info/debug to stdout, warn/error to stderr)
- **Structured key-value pairs** for rich log data

## Quick Start

### 1. Import the logging package

```go
import (
    "context"
    "log/slog"
    "github.com/phrazzld/thinktank/internal/logutil"
)
```

### 2. Create a structured logger

```go
// Create a JSON logger that outputs to stderr
logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)

// Or create a logger with stream separation (info to stdout, errors to stderr)
logger := logutil.NewSlogLoggerWithStreamSeparationFromLogLevel(
    os.Stdout, os.Stderr, logutil.InfoLevel)
```

### 3. Create context with correlation ID

```go
// Generate a new correlation ID
ctx := logutil.WithCorrelationID(context.Background())

// Or use a custom correlation ID
ctx := logutil.WithCorrelationID(context.Background(), "req-123")
```

### 4. Log with structured data

```go
// Use slog.Attr functions for structured key-value pairs
logger.InfoContext(ctx, "user operation completed",
    slog.String("user_id", "user123"),
    slog.String("operation", "login"),
    slog.Int("duration_ms", 250))

logger.ErrorContext(ctx, "database connection failed",
    slog.String("database", "users"),
    slog.String("error", err.Error()),
    slog.Int("retry_count", 3))
```

## JSON Output Example

The above logging calls produce JSON output like:

```json
{
  "time": "2025-06-09T21:45:30.123456-07:00",
  "level": "INFO",
  "msg": "user operation completed",
  "correlation_id": "7c88c317-2766-4a33-8ca3-23135411c7b1",
  "user_id": "user123",
  "operation": "login",
  "duration_ms": 250
}

{
  "time": "2025-06-09T21:45:31.456789-07:00",
  "level": "ERROR",
  "msg": "database connection failed",
  "correlation_id": "7c88c317-2766-4a33-8ca3-23135411c7b1",
  "database": "users",
  "error": "connection timeout after 5s",
  "retry_count": 3
}
```

## Logger Interface

All loggers implement the `logutil.LoggerInterface` which provides both context-aware and standard logging methods:

```go
type LoggerInterface interface {
    // Context-aware methods (preferred)
    DebugContext(ctx context.Context, msg string, args ...any)
    InfoContext(ctx context.Context, msg string, args ...any)
    WarnContext(ctx context.Context, msg string, args ...any)
    ErrorContext(ctx context.Context, msg string, args ...any)
    FatalContext(ctx context.Context, msg string, args ...any)

    // Standard methods
    Debug(format string, v ...interface{})
    Info(format string, v ...interface{})
    Warn(format string, v ...interface{})
    Error(format string, v ...interface{})
    Fatal(format string, v ...interface{})

    // WithContext creates a logger with context attached
    WithContext(ctx context.Context) LoggerInterface
}
```

## Correlation ID Usage

### Generating Correlation IDs

```go
// Generate a new UUID correlation ID
ctx := logutil.WithCorrelationID(context.Background())

// Use a custom correlation ID (e.g., from HTTP header)
requestID := r.Header.Get("X-Request-ID")
ctx := logutil.WithCorrelationID(context.Background(), requestID)

// Preserve existing correlation ID
ctx := logutil.WithCorrelationID(existingCtx) // Keeps existing ID if present
```

### Retrieving Correlation IDs

```go
correlationID := logutil.GetCorrelationID(ctx)
if correlationID != "" {
    // Use the correlation ID for other purposes (e.g., response headers)
    w.Header().Set("X-Correlation-ID", correlationID)
}
```

### Propagating Context

Always pass context through your call chain to maintain correlation ID:

```go
func ProcessRequest(ctx context.Context, userID string) error {
    // Context with correlation ID is automatically propagated
    logger.InfoContext(ctx, "processing request", slog.String("user_id", userID))

    // Pass context to other functions
    result, err := CallDatabase(ctx, userID)
    if err != nil {
        logger.ErrorContext(ctx, "database call failed",
            slog.String("user_id", userID),
            slog.String("error", err.Error()))
        return err
    }

    logger.InfoContext(ctx, "request completed",
        slog.String("user_id", userID),
        slog.Any("result", result))
    return nil
}

func CallDatabase(ctx context.Context, userID string) (interface{}, error) {
    // Correlation ID is automatically included in logs
    logger.DebugContext(ctx, "executing database query",
        slog.String("user_id", userID),
        slog.String("query", "SELECT * FROM users WHERE id = ?"))

    // Database operations...
    return result, nil
}
```

## Structured Data Types

Use appropriate slog functions for different data types:

```go
// Strings
slog.String("key", "value")

// Numbers
slog.Int("count", 42)
slog.Int64("timestamp", time.Now().Unix())
slog.Float64("percentage", 95.5)

// Booleans
slog.Bool("success", true)

// Time
slog.Time("created_at", time.Now())

// Duration
slog.Duration("elapsed", time.Since(start))

// Any type (uses reflection)
slog.Any("data", complexObject)

// Groups for nested objects
slog.Group("user",
    slog.String("id", "123"),
    slog.String("name", "John Doe"),
    slog.String("email", "john@example.com"))
```

## Logger Creation Patterns

### Service Initialization

```go
type UserService struct {
    logger logutil.LoggerInterface
    db     Database
}

func NewUserService(logger logutil.LoggerInterface, db Database) *UserService {
    return &UserService{
        logger: logger,
        db:     db,
    }
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    s.logger.InfoContext(ctx, "fetching user", slog.String("user_id", userID))

    user, err := s.db.GetUser(ctx, userID)
    if err != nil {
        s.logger.ErrorContext(ctx, "failed to fetch user",
            slog.String("user_id", userID),
            slog.String("error", err.Error()))
        return nil, err
    }

    s.logger.InfoContext(ctx, "user fetched successfully",
        slog.String("user_id", userID),
        slog.String("username", user.Username))

    return user, nil
}
```

### HTTP Handler Pattern

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Generate correlation ID for this request
    ctx := logutil.WithCorrelationID(r.Context())

    // Create a logger with context attached
    logger := h.logger.WithContext(ctx)

    logger.InfoContext(ctx, "creating new user",
        slog.String("method", r.Method),
        slog.String("path", r.URL.Path),
        slog.String("remote_addr", r.RemoteAddr))

    // Set correlation ID in response header for client reference
    w.Header().Set("X-Correlation-ID", logutil.GetCorrelationID(ctx))

    // Process request...
    user, err := h.userService.CreateUser(ctx, userRequest)
    if err != nil {
        logger.ErrorContext(ctx, "user creation failed",
            slog.String("error", err.Error()))
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    logger.InfoContext(ctx, "user created successfully",
        slog.String("user_id", user.ID))

    // Return response...
}
```

### Testing Pattern

```go
func TestUserService_GetUser(t *testing.T) {
    // Create a test logger that captures output
    var buf bytes.Buffer
    logger := logutil.NewSlogLoggerFromLogLevel(&buf, logutil.DebugLevel)

    // Create context with test correlation ID
    ctx := logutil.WithCorrelationID(context.Background(), "test-123")

    // Run test
    service := NewUserService(logger, mockDB)
    user, err := service.GetUser(ctx, "user123")

    // Verify logs
    logOutput := buf.String()
    assert.Contains(t, logOutput, "test-123") // Correlation ID present
    assert.Contains(t, logOutput, "user123")  // User ID logged

    // Parse JSON logs for detailed verification
    var logEntry map[string]interface{}
    json.Unmarshal([]byte(logOutput), &logEntry)
    assert.Equal(t, "test-123", logEntry["correlation_id"])
}
```

## Migration from Standard Logging

### Before (standard log)

```go
import "log"

log.Printf("User %s logged in successfully", userID)
log.Printf("Error processing request: %v", err)
```

### After (structured logging)

```go
import (
    "context"
    "log/slog"
    "github.com/phrazzld/thinktank/internal/logutil"
)

ctx := logutil.WithCorrelationID(context.Background())
logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)

logger.InfoContext(ctx, "user logged in successfully", slog.String("user_id", userID))
logger.ErrorContext(ctx, "error processing request", slog.String("error", err.Error()))
```

## Performance Considerations

1. **Reuse loggers**: Create logger instances once and reuse them
2. **Use appropriate log levels**: Debug logs have overhead, use INFO for production
3. **Avoid expensive operations in log arguments**: Don't compute complex values unless the log level is enabled
4. **Use slog.Any sparingly**: It uses reflection and can be slower

## Best Practices

1. **Always use context-aware methods** (`InfoContext`, `ErrorContext`, etc.) for correlation ID support
2. **Generate correlation IDs early** in request processing (HTTP handlers, CLI commands, etc.)
3. **Pass context through all function calls** to maintain correlation ID
4. **Use structured fields** instead of string interpolation for machine-readable logs
5. **Log the right amount**: Too little = hard to debug, too much = noise
6. **Use consistent field names** across your application (e.g., always use "user_id", not "userId")
7. **Include error context**: Log error messages along with the operation that failed
8. **Test your logging**: Write tests that verify log output and correlation ID propagation

## Stream Separation

For applications that need different output streams:

```go
// Info/Debug logs go to stdout, Warn/Error logs go to stderr
logger := logutil.NewSlogLoggerWithStreamSeparationFromLogLevel(
    os.Stdout,   // Info/Debug output
    os.Stderr,   // Warn/Error output
    logutil.InfoLevel)

// Enable stream separation on existing logger
separatedLogger := logutil.EnableStreamSeparation(existingLogger)
```

This is useful for:
- Containerized applications where stdout and stderr are handled differently
- CLI tools where you want errors separate from normal output
- Applications with log aggregation systems that route streams differently

## Troubleshooting

### Correlation ID not appearing in logs

- Ensure you're using context-aware methods (`InfoContext` vs `Info`)
- Verify the context was created with `WithCorrelationID`
- Check that context is being passed through all function calls

### JSON output not formatted correctly

- Verify you're using `NewSlogLogger` or `NewSlogLoggerFromLogLevel`
- Ensure you're using `slog.String()`, `slog.Int()` etc. for structured fields
- Check that you're not mixing format strings with structured arguments

### Performance issues

- Use appropriate log levels (avoid Debug in production)
- Profile log-heavy code paths
- Consider async logging for high-throughput applications
