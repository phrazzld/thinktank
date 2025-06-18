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
- [Error Handling in Test Files](#error-handling-in-test-files)
  - [golangci-lint errcheck Compliance](#golangci-lint-errcheck-compliance)
  - [When to Use t.Fatalf() vs t.Errorf()](#when-to-use-tfatalf-vs-terrorf)
  - [Common Patterns for Test Error Handling](#common-patterns-for-test-error-handling)
  - [Pre-commit and CI Integration](#pre-commit-and-ci-integration)
  - [Best Practices Summary](#best-practices-summary)

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

Our logging system provides structured, context-aware logging with support for correlation IDs and log stream separation. The system now includes a dual-output approach:

- **Console Output**: Clean, human-readable progress and status reporting via ConsoleWriter
- **Structured Logs**: JSON-formatted logs for debugging and audit trails

### Dual-Output Architecture

The logging system operates on two parallel channels:

1. **ConsoleWriter**: Provides clean, user-facing output to stdout with interactive features
2. **Structured Logger**: Maintains JSON logs for debugging, auditing, and machine processing

This separation ensures users see clean progress information while maintaining comprehensive structured logging for troubleshooting.

### LoggerInterface

All structured logging is done through the `LoggerInterface` defined in `internal/logutil/logutil.go`:

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

### Output Routing and CLI Flags

The logging system provides flexible output routing controlled by CLI flags:

#### Default Behavior
- **Console Output**: Clean, formatted messages to stdout (via ConsoleWriter)
- **Structured Logs**: JSON logs saved to `thinktank.log` file in output directory

#### CLI Flag Options

| Flag | Description | Effect |
|------|-------------|--------|
| `--quiet`, `-q` | Suppress console output (errors only) | Only error messages shown on console |
| `--json-logs` | Show JSON logs on stderr | Preserves legacy behavior for scripts |
| `--no-progress` | Disable progress indicators | Show only start/complete messages |
| `--verbose` | Enable both console AND JSON logs | Console output + JSON to stderr |

#### Stream Separation

When `--json-logs` or `--verbose` is used, structured logs are sent to stderr with severity-based separation:
- INFO and DEBUG logs go to STDOUT
- WARN and ERROR logs go to STDERR

This separation helps with usability in CLI applications and maintains compatibility with existing log processing scripts.

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

## Error Handling in Test Files

Test files require special attention to error handling to ensure CI pipeline compliance and proper test behavior. This section outlines best practices for handling errors in Go test files.

### golangci-lint errcheck Compliance

The `errcheck` linter enforces that all error return values are checked. This is particularly important in test files where unchecked errors can lead to false positives or incomplete test coverage.

### When to Use t.Fatalf() vs t.Errorf()

Choose the appropriate error reporting method based on the criticality of the operation:

#### Use `t.Fatalf()` for Critical Setup Operations

Use `t.Fatalf()` when the error prevents the test from continuing meaningfully:

```go
// Critical setup that must succeed for the test to be valid
tempDir := t.TempDir()
configDir := filepath.Join(tempDir, ".config", "thinktank")
err := os.MkdirAll(configDir, 0755)
if err != nil {
    t.Fatalf("Failed to create test config directory: %v", err)
}

// Changing working directory - critical for test isolation
originalWd, err := os.Getwd()
if err != nil {
    t.Fatalf("Failed to get current working directory: %v", err)
}
if err := os.Chdir(tempDir); err != nil {
    t.Fatalf("Failed to change to temp directory: %v", err)
}
```

#### Use `t.Errorf()` for Cleanup Operations

Use `t.Errorf()` for operations that should succeed but won't invalidate the test if they fail:

```go
// Cleanup in defer - use t.Errorf() to report but not fail the test
defer func() {
    if err := os.Chdir(originalWd); err != nil {
        t.Errorf("Failed to restore working directory: %v", err)
    }
}()

// Environment variable restoration
defer func() {
    if originalHome != "" {
        if err := os.Setenv("HOME", originalHome); err != nil {
            t.Errorf("Failed to restore HOME environment variable: %v", err)
        }
    } else {
        if err := os.Unsetenv("HOME"); err != nil {
            t.Errorf("Failed to unset HOME environment variable: %v", err)
        }
    }
}()
```

### Common Patterns for Test Error Handling

#### 1. Environment Variable Manipulation

Always check errors when setting or unsetting environment variables:

```go
// Save original value
originalHome := os.Getenv("HOME")

// Set new value with error checking
if err := os.Setenv("HOME", tempDir); err != nil {
    t.Errorf("Failed to set HOME environment variable: %v", err)
}

// Restore in defer with proper error handling
defer func() {
    if originalHome != "" {
        if err := os.Setenv("HOME", originalHome); err != nil {
            t.Errorf("Failed to restore HOME environment variable: %v", err)
        }
    } else {
        if err := os.Unsetenv("HOME"); err != nil {
            t.Errorf("Failed to unset HOME environment variable: %v", err)
        }
    }
}()
```

#### 2. File Operations

Handle file operation errors appropriately:

```go
// File creation - usually critical
err := os.WriteFile(configFile, []byte(testConfig), 0644)
if err != nil {
    t.Fatalf("Failed to write test config file: %v", err)
}

// File removal in cleanup - non-critical
defer func() {
    if err := os.Remove(tempFile.Name()); err != nil {
        t.Errorf("Failed to remove temporary file: %v", err)
    }
}()

// File closing - should always be checked
if err := tmpFile.Close(); err != nil {
    t.Errorf("Failed to close temporary file: %v", err)
}
```

#### 3. Directory Operations

Handle directory changes carefully to maintain test isolation:

```go
// Save current directory
originalWd, err := os.Getwd()
if err != nil {
    t.Fatalf("Failed to get current working directory: %v", err)
}

// Change directory with error checking
if err := os.Chdir(tempDir); err != nil {
    t.Fatalf("Failed to change to temp directory: %v", err)
}

// Always restore in defer
defer func() {
    if err := os.Chdir(originalWd); err != nil {
        t.Errorf("Failed to restore working directory: %v", err)
    }
}()
```

### Pre-commit and CI Integration

To prevent errcheck violations from reaching CI:

1. **Run golangci-lint locally before committing**:
   ```bash
   golangci-lint run ./...
   ```

2. **Check specific packages after modifications**:
   ```bash
   golangci-lint run internal/providers/
   ```

3. **Fix all errcheck violations before pushing**:
   - Never suppress errors with `_` unless absolutely necessary
   - If an error truly can be ignored, document why with a comment
   - Consider if the operation is actually necessary if the error doesn't matter

### Best Practices Summary

1. **Always check error returns** - Never ignore errors from OS operations, even in tests
2. **Use appropriate error methods** - `t.Fatalf()` for critical setup, `t.Errorf()` for cleanup
3. **Maintain test isolation** - Always restore original state (working directory, environment variables)
4. **Document error handling decisions** - If an error is intentionally ignored, explain why
5. **Run linters locally** - Catch errcheck violations before they reach CI
6. **Follow existing patterns** - Consistency across the codebase makes maintenance easier

By following these patterns, you'll avoid common errcheck violations and ensure your tests are robust, maintainable, and CI-compliant.
