# Structured JSON Logging Guide

This document describes how to use structured JSON logging with correlation IDs in the thinktank project.

## Overview

The project uses a dual-output logging system that combines clean console output with comprehensive structured logging:

### Console Output (New Default)
- **Clean, human-readable progress** via ConsoleWriter interface
- **Interactive features** like emojis and progress indicators (TTY environments)
- **CI-friendly output** with simple text formatting (non-TTY environments)
- **User-facing status updates** to stdout

### Structured JSON Logging
- **Machine-readable logs** built on Go's `log/slog` package
- **Correlation ID support** for tracing request flow across operations
- **Context-aware logging** methods with structured key-value pairs
- **Flexible output routing** based on CLI flags

### Output Routing Modes

| Mode | Console Output | JSON Logs | Use Case |
|------|----------------|-----------|----------|
| **Default** | Clean progress to stdout | Saved to `thinktank.log` file | Normal user interaction |
| **`--quiet`** | Errors only | Saved to `thinktank.log` file | Automated scripts |
| **`--json-logs`** | Clean progress to stdout | JSON to stderr | Legacy compatibility |
| **`--verbose`** | Clean progress to stdout | JSON to stderr | Development/debugging |

## Quick Start

### 1. Import the required packages

```go
import (
    "context"
    "log/slog"
    "github.com/misty-step/thinktank/internal/logutil"
)
```

### 2. Set up dual-output logging

```go
// Create a ConsoleWriter for user-facing output
consoleWriter := logutil.NewConsoleWriter()

// Create structured logger (output depends on CLI flags)
logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)

// Configure based on CLI flags
consoleWriter.SetQuiet(config.Quiet)
consoleWriter.SetNoProgress(config.NoProgress)
```

### 3. Use ConsoleWriter for user-facing progress

```go
// Start processing workflow
consoleWriter.StartProcessing(3)

// Report model progress
consoleWriter.ModelStarted("gemini-1.5-flash", 1)
consoleWriter.ModelCompleted("gemini-1.5-flash", 1, duration, nil)

// Report workflow completion
consoleWriter.SynthesisCompleted("/path/to/output")
```

### 4. Use structured logging for debugging

```go
// Create context with correlation ID
ctx := logutil.WithCorrelationID(context.Background())

// Log structured data for debugging/auditing
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

## ConsoleWriter Interface

The `ConsoleWriter` interface provides clean, user-facing console output that adapts to different environments:

```go
type ConsoleWriter interface {
    // Progress reporting
    StartProcessing(modelCount int)
    ModelStarted(modelName string, index int)
    ModelCompleted(modelName string, index int, duration time.Duration, err error)
    ModelRateLimited(modelName string, index int, delay time.Duration)

    // Status updates
    SynthesisStarted()
    SynthesisCompleted(outputPath string)
    StatusMessage(message string)

    // Message formatting
    ErrorMessage(message string)
    WarningMessage(message string)
    SuccessMessage(message string)

    // Control
    SetQuiet(quiet bool)
    SetNoProgress(noProgress bool)
    IsInteractive() bool
}
```

### Environment Detection

ConsoleWriter automatically detects the execution environment:

- **Interactive Mode** (TTY + not CI): Rich output with emojis and progress indicators
- **CI Mode** (non-TTY or CI=true): Simple, parseable text output

```go
// Interactive mode output
🚀 Processing 3 models...
[1/3] gemini-1.5-flash: processing...
[1/3] gemini-1.5-flash: ✓ completed (0.8s)
✨ Done! Output saved to: output_20240115_120000/

// CI mode output
Starting processing with 3 models
Processing model 1/3: gemini-1.5-flash
Completed model 1/3: gemini-1.5-flash (0.8s)
Synthesis complete. Output: output_20240115_120000/
```

## Structured Logger Interface

All structured loggers implement the `logutil.LoggerInterface` which provides both context-aware and standard logging methods:

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

## Migration Guide

### Migrating from JSON-Only Logging (Old System)

The previous system only provided structured JSON logs. The new system adds clean console output while maintaining JSON logging.

#### Before (JSON-only)
```go
import "log/slog"

// Old: Only JSON output to stderr
logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
logger.InfoContext(ctx, "processing model", slog.String("model", "gemini-1.5-flash"))
```

#### After (Dual-output)
```go
import "log/slog"

// New: Clean console output + structured logging
consoleWriter := logutil.NewConsoleWriter()
logger := logutil.NewSlogLoggerFromLogLevel(outputDest, logutil.InfoLevel)

// User-facing progress
consoleWriter.ModelStarted("gemini-1.5-flash", 1)

// Structured logging for debugging
logger.InfoContext(ctx, "processing model", slog.String("model", "gemini-1.5-flash"))
```

### Backward Compatibility

Existing scripts that depend on JSON output can use the `--json-logs` flag:

```bash
# Old behavior: JSON to stderr
thinktank --json-logs --instructions task.txt ./src

# New default: Clean console + JSON to file
thinktank --instructions task.txt ./src

# Development: Both console and JSON to stderr
thinktank --verbose --instructions task.txt ./src
```

### Migrating from Standard Logging

#### Before (standard log)
```go
import "log"

log.Printf("User %s logged in successfully", userID)
log.Printf("Error processing request: %v", err)
```

#### After (dual-output logging)
```go
import (
    "context"
    "log/slog"
    "github.com/misty-step/thinktank/internal/logutil"
)

ctx := logutil.WithCorrelationID(context.Background())
consoleWriter := logutil.NewConsoleWriter()
logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)

// User-facing messages
consoleWriter.SuccessMessage("User logged in successfully")
consoleWriter.ErrorMessage("Error processing request")

// Structured logs for debugging
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

## Token Counting and Model Selection Logging

The thinktank project includes comprehensive logging for token counting operations and model selection decisions. This section documents the structured logging patterns used for these operations.

### Tokenization Service Logging

When using the `TokenCountingService` with logging enabled, structured log entries track tokenization decisions:

```go
// Create service with logger for structured logging
logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
service := NewTokenCountingServiceWithLogger(logger)

// Model compatibility check with structured logging
ctx := logutil.WithCorrelationID(context.Background())
results, err := service.GetCompatibleModels(ctx, req, []string{"openai", "gemini"})
```

### Model Selection Log Fields

The tokenization service logs the following structured fields:

#### Initial Model Selection Context
```json
{
  "timestamp": "2024-12-27T22:30:00Z",
  "level": "INFO",
  "message": "Starting model compatibility check",
  "correlation_id": "req_abc123",
  "provider_count": 2,
  "file_count": 5,
  "has_instructions": true
}
```

#### Individual Model Evaluation
```json
{
  "timestamp": "2024-12-27T22:30:01Z",
  "level": "INFO",
  "message": "Model evaluation:",
  "correlation_id": "req_abc123",
  "model": "gpt-5.2",
  "provider": "openai",
  "context_window": 1000000,
  "status": "COMPATIBLE",
  "tokenizer": "tiktoken",
  "accurate": true
}
```

#### Model Selection Summary
```json
{
  "timestamp": "2024-12-27T22:30:02Z",
  "level": "INFO",
  "message": "Model compatibility check completed",
  "correlation_id": "req_abc123",
  "total_models": 8,
  "compatible_models": 6,
  "accurate_count": 3,
  "estimated_count": 5
}
```

### Tokenizer Selection Logic

The logging system tracks which tokenizer is used for each model:

| Provider | Tokenizer Used | Log Field Value | Accuracy |
|----------|----------------|-----------------|----------|
| `openai` | tiktoken | `"tiktoken"` | High (90%+) |
| `gemini` | SentencePiece | `"sentencepiece"` | High (90%+) |
| `openrouter` | Estimation fallback | `"estimation"` | Lower (75%) |
| Unknown | Estimation fallback | `"estimation"` | Lower (75%) |

### Model Compatibility Reasons

When models are marked as incompatible, the `reason` field provides detailed context:

```json
{
  "model": "gpt-5.2",
  "status": "SKIPPED",
  "reason": "requires 1200000 tokens but model only has 800000 usable tokens (1000000 total - 200000 safety margin)"
}
```

### Troubleshooting Token Counting

#### Missing tokenization logs
- Ensure you're using `NewTokenCountingServiceWithLogger()` constructor
- Verify the logger is properly configured with appropriate log level
- Check that correlation IDs are propagated through context

#### Inaccurate token counts
- Look for `"accurate": false` in log entries indicating fallback to estimation
- Check `tokenizer` field to understand which method was used
- Review model provider support (OpenAI uses tiktoken, others may use estimation)

#### Performance issues with tokenization
- Monitor tokenization time in structured logs
- Check for repeated tokenization of the same content
- Consider caching strategies for large file sets

## Tokenizer Selection Logic

thinktank automatically selects the best tokenization method based on model provider and availability:

### Selection Process

1. **Provider Detection**: Determines provider from model name using `models.GetModelInfo()`
2. **Accurate Tokenizer Check**: Attempts provider-specific tokenizer
3. **Fallback to Estimation**: Uses character-based estimation if accurate tokenizer unavailable

### Decision Flow

```
Model Name → Provider Detection → Tokenizer Selection
    ↓
OpenAI models (gpt-4*, o4-*) → tiktoken encoding
    ↓
Gemini models (gemini-*, gemma-*) → SentencePiece encoding
    ↓
Other providers → Estimation fallback (0.75 tokens/char)
```

### Implementation Details

The tokenization service uses a lazy-loading architecture:

```go
// Provider-aware tokenizer manager
manager := tokenizers.NewTokenizerManager()

// Automatic provider selection
tokenizer, err := manager.GetTokenizer(provider)
if err != nil {
    // Falls back to estimation automatically
}

count, err := tokenizer.CountTokens(ctx, text, modelName)
```

## Provider Support Matrix

| Provider   | Tokenizer     | Accuracy | Encoding | Status |
|------------|---------------|----------|----------|--------|
| OpenAI     | tiktoken      | Exact    | cl100k_base, o200k_base | ✓ |
| Gemini     | SentencePiece | Exact    | Gemini-specific | ✓ |
| OpenRouter | Estimation    | ~95%     | Character-based | △ |
| Others     | Estimation    | ~75%     | Character-based | △ |

### Tokenizer Accuracy

- **Exact (90%+ accuracy)**: Uses provider's official tokenizer
- **High (~95% accuracy)**: Provider-compatible tokenizer with minor variations
- **Estimation (~75% accuracy)**: Character-based calculation (1 token ≈ 4 chars)

### Circuit Breaker Integration

Tokenizers include circuit breaker protection:

- **Failure Threshold**: 5 consecutive failures triggers circuit open
- **Recovery Time**: 30 seconds before attempting retry
- **Automatic Fallback**: Estimation used when circuit is open
