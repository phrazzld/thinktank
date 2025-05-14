# Error Handling and Logging Standards

This document describes the standardized approach to error handling and logging used in this project.

## Table of Contents

- [Error Handling](#error-handling)
  - [LLMError and Error Categories](#llmerror-and-error-categories)
  - [Error Helper Functions](#error-helper-functions)
  - [Error Handling Patterns](#error-handling-patterns)
  - [Provider-Specific Error Translation](#provider-specific-error-translation)
- [Structured Logging](#structured-logging)
  - [LoggerInterface](#loggerinterface)
  - [Context-Aware Logging](#context-aware-logging)
  - [Correlation IDs](#correlation-ids)
  - [Stream Separation](#stream-separation)
  - [Logging Patterns](#logging-patterns)

## Error Handling

Our error handling system provides a structured, categorized approach to handling errors, especially those from LLM providers. This allows for consistent error handling, meaningful user feedback, and proper error categorization.

### LLMError and Error Categories

The foundation of our error handling system is the `LLMError` struct defined in `internal/llm/errors.go`:

```go
type LLMError struct {
    // Provider is the name of the provider that generated the error
    Provider string

    // Code is the provider-specific error code (if available)
    Code string

    // StatusCode is the HTTP status code (if applicable)
    StatusCode int

    // Message is a user-friendly error message
    Message string

    // RequestID is the ID of the request that failed (if available)
    RequestID string

    // Original is the underlying error that caused this error
    Original error

    // Category is the standardized error category
    ErrorCategory ErrorCategory

    // Suggestion is a helpful suggestion for resolving the error
    Suggestion string

    // Details contains additional error details
    Details string
}
```

Every error is categorized using the `ErrorCategory` enum:

```go
type ErrorCategory int

const (
    // CategoryUnknown represents an unknown or uncategorized error
    CategoryUnknown ErrorCategory = iota
    // CategoryAuth represents authentication and authorization errors
    CategoryAuth
    // CategoryRateLimit represents rate limiting or quota errors
    CategoryRateLimit
    // CategoryInvalidRequest represents invalid request errors
    CategoryInvalidRequest
    // CategoryNotFound represents model not found errors
    CategoryNotFound
    // CategoryServer represents server errors
    CategoryServer
    // CategoryNetwork represents network connectivity errors
    CategoryNetwork
    // CategoryCancelled represents cancelled context errors
    CategoryCancelled
    // CategoryInputLimit represents input token limit exceeded errors
    CategoryInputLimit
    // CategoryContentFiltered represents content filtered by safety settings errors
    CategoryContentFiltered
    // CategoryInsufficientCredits represents insufficient credits or payment required errors
    CategoryInsufficientCredits
)
```

### Error Helper Functions

We provide several helper functions to work with errors:

1. **Wrap**: Wraps an existing error with additional context
   ```go
   // Wrap wraps an existing error with additional LLM-specific context
   func Wrap(err error, provider string, message string, category ErrorCategory) *LLMError
   ```

2. **IsCategory**: Checks if an error belongs to a specific category
   ```go
   // IsCategory checks if an error belongs to a specific category
   func IsCategory(err error, category ErrorCategory) bool
   ```

3. **Category-specific helpers**: Convenience functions for common categories
   ```go
   func IsAuth(err error) bool
   func IsRateLimit(err error) bool
   func IsInvalidRequest(err error) bool
   func IsNotFound(err error) bool
   func IsServer(err error) bool
   func IsNetwork(err error) bool
   func IsCancelled(err error) bool
   func IsInputLimit(err error) bool
   func IsContentFiltered(err error) bool
   func IsInsufficientCredits(err error) bool
   ```

4. **Error detection and formatting**:
   ```go
   // DetectErrorCategory determines the error category from status code and message
   func DetectErrorCategory(err error, statusCode int) ErrorCategory

   // FormatAPIError creates a standardized LLMError from a generic error
   func FormatAPIError(provider string, err error, statusCode int, responseBody string) *LLMError
   ```

### Error Handling Patterns

When handling errors in your code, follow these patterns:

1. **Always check and handle errors explicitly**:
   ```go
   result, err := someFunction()
   if err != nil {
       // Handle the error
   }
   ```

2. **Wrap errors with context**:
   ```go
   result, err := someFunction()
   if err != nil {
       return fmt.Errorf("failed to process request: %w", err)
       // OR for LLM-specific errors:
       return llm.Wrap(err, "provider-name", "Failed to process request", llm.CategoryInvalidRequest)
   }
   ```

3. **Check error categories for specific handling**:
   ```go
   if err != nil {
       if llm.IsRateLimit(err) {
           // Handle rate limit error (e.g., retry with backoff)
       } else if llm.IsAuth(err) {
           // Handle authentication error
       } else {
           // Handle other errors
       }
   }
   ```

4. **Handle error categories at boundaries**:
   In top-level components (like CLI handlers), handle errors based on their category to provide appropriate feedback and exit codes.

### Provider-Specific Error Translation

For provider-specific code, implement a `FormatAPIError` function that translates provider-specific errors into the standard `LLMError` format:

```go
// Example implementation for a specific provider
func FormatAPIError(err error, statusCode int, responseBody string) *LLMError {
    if err == nil {
        return nil
    }

    // Determine most specific error category
    category := llm.DetectErrorCategory(err, statusCode)

    // Create standardized error with provider-specific details
    return llm.CreateStandardErrorWithMessage("provider-name", category, err, responseBody)
}
```

## Structured Logging

Our logging system provides structured, context-aware logging with support for correlation IDs and log stream separation.

### LoggerInterface

All logging is done through the `LoggerInterface` defined in `internal/logutil/logutil.go`:

```go
type LoggerInterface interface {
    // Context-aware logging methods with structured key-value pairs
    DebugContext(ctx context.Context, msg string, args ...any)
    InfoContext(ctx context.Context, msg string, args ...any)
    WarnContext(ctx context.Context, msg string, args ...any)
    ErrorContext(ctx context.Context, msg string, args ...any)
    FatalContext(ctx context.Context, msg string, args ...any)

    // Standard logging methods (prefer context-aware methods when possible)
    Debug(format string, v ...interface{})
    Info(format string, v ...interface{})
    Warn(format string, v ...interface{})
    Error(format string, v ...interface{})
    Fatal(format string, v ...interface{})

    // Legacy compatibility methods (use Info/InfoContext instead when possible)
    Println(v ...interface{})
    Printf(format string, v ...interface{})

    // WithContext returns a logger with context information attached
    WithContext(ctx context.Context) LoggerInterface
}
```

The primary implementation uses Go's `log/slog` package for structured JSON logging.

### Context-Aware Logging

Always prefer context-aware logging methods when possible. These methods automatically include correlation IDs and other relevant context information:

```go
// Instead of:
logger.Info("Processing request for user %s", userID)

// Use:
logger.InfoContext(ctx, "Processing request for user %s", userID)
```

### Correlation IDs

Correlation IDs are used to track related operations and are propagated through contexts:

```go
// Create a new context with a correlation ID
ctx = logutil.WithCorrelationID(ctx)

// Or set a specific correlation ID
ctx = logutil.WithCorrelationID(ctx, "custom-request-id-123")

// Retrieve the correlation ID from a context
id := logutil.GetCorrelationID(ctx)
```

When using context-aware logging methods, correlation IDs are automatically included in log entries.

### Stream Separation

Our logging system supports separating logs by severity level:
- INFO and DEBUG logs go to STDOUT
- WARN and ERROR logs go to STDERR

This separation helps with usability in CLI applications. Stream separation can be enabled via configuration:

```go
// Enable stream separation in the app configuration
config.SplitLogs = true
```

### Logging Patterns

Follow these practices for effective logging:

1. **Use context-aware methods**:
   ```go
   logger.InfoContext(ctx, "Processing file %s", filename)
   ```

2. **Include relevant key-value pairs for structured logging**:
   ```go
   logger.InfoContext(ctx, "User authenticated",
       "user_id", user.ID,
       "role", user.Role,
       "ip_address", ipAddress)
   ```

3. **Use appropriate log levels**:
   - `DEBUG`: Detailed diagnostic information
   - `INFO`: Normal operation events
   - `WARN`: Potential issues that didn't prevent operation
   - `ERROR`: Errors that caused operation failure
   - `FATAL`: Critical errors causing immediate termination

4. **Include operation results and durations**:
   ```go
   startTime := time.Now()
   result, err := someOperation()
   duration := time.Since(startTime)

   if err != nil {
       logger.ErrorContext(ctx, "Operation failed",
           "operation", "someOperation",
           "duration_ms", duration.Milliseconds(),
           "error", err)
       return err
   }

   logger.InfoContext(ctx, "Operation completed successfully",
       "operation", "someOperation",
       "duration_ms", duration.Milliseconds(),
       "result_count", len(result))
   ```

5. **NEVER log sensitive information**:
   - API keys
   - Credentials
   - Personal information
   - Full request/response bodies that might contain sensitive data

For proper sanitization of sensitive information, use the `SanitizingLogger` which automatically removes API keys and other sensitive information from log messages.
