// Package auditlog provides structured logging for audit purposes
package auditlog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// AuditLogger defines the interface for logging audit events.
// Implementations of this interface will handle persisting audit
// log entries in various formats (e.g., JSON Lines file, no-op).
type AuditLogger interface {
	// Log records a single audit entry.
	// The entry contains information about operations, status, and relevant metadata.
	// Returns an error if the logging operation fails.
	// NOTE: Prefer using the LogOp method instead of this method directly.
	Log(ctx context.Context, entry AuditEntry) error

	// LogOp is a helper method for logging operations with minimal parameters.
	// It creates an AuditEntry with the given operation, status, and optional data,
	// sets a timestamp, and logs it. The method returns any error from logging.
	// This is the recommended way to log audit events to ensure consistency.
	LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error

	// Close releases any resources used by the logger (e.g., open file handles).
	// Should be called when the logger is no longer needed.
	// Returns an error if the closing operation fails.
	Close() error

	// The following methods are provided for backward compatibility.
	// New code should use the context-aware methods above.

	// LogLegacy is the non-context version of Log for backward compatibility.
	// This method will be deprecated in the future.
	LogLegacy(entry AuditEntry) error

	// LogOpLegacy is the non-context version of LogOp for backward compatibility.
	// This method will be deprecated in the future.
	LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error
}

// FileAuditLogger implements AuditLogger by writing JSON Lines to a file.
type FileAuditLogger struct {
	file   *os.File
	mu     sync.Mutex
	logger logutil.LoggerInterface // For logging errors within the audit logger itself
}

// NewFileAuditLogger creates a new FileAuditLogger that writes to the specified file path.
// If the file doesn't exist, it will be created. If it does exist, logs will be appended.
// The provided internal logger is used to log any errors that occur during audit logging operations.
func NewFileAuditLogger(filePath string, internalLogger logutil.LoggerInterface) (*FileAuditLogger, error) {
	// Create a context with a correlation ID for initialization logs
	ctx := logutil.WithCorrelationID(context.Background())

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		internalLogger.ErrorContext(ctx, "Failed to open audit log file '%s': %v", filePath, err)
		return nil, fmt.Errorf("failed to open audit log file %s: %w", filePath, err)
	}
	internalLogger.InfoContext(ctx, "Audit logging enabled to file: %s", filePath)
	return &FileAuditLogger{
		file:   file,
		logger: internalLogger,
	}, nil
}

// Log records a single audit entry by marshaling it to JSON and writing it to the log file.
// It sets the entry timestamp if not already set and ensures thread safety with a mutex lock.
// The context is used to extract correlation ID and for context-aware logging.
func (l *FileAuditLogger) Log(ctx context.Context, entry AuditEntry) error {
	// Use default context if nil is provided for backward compatibility
	if ctx == nil {
		ctx = context.Background()
	}

	// Get a context-aware logger
	contextLogger := l.logger.WithContext(ctx)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure timestamp is set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Add correlation ID from context if not already present in inputs
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		if entry.Inputs == nil {
			entry.Inputs = make(map[string]interface{})
		}
		// Only add if not already present
		if _, exists := entry.Inputs["correlation_id"]; !exists {
			entry.Inputs["correlation_id"] = correlationID
		}
	}

	// Marshal entry to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Failed to marshal audit entry to JSON: %v, Entry: %+v", err, entry)
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Write JSON line to file
	if _, err := l.file.Write(append(jsonData, '\n')); err != nil {
		contextLogger.ErrorContext(ctx, "Failed to write audit entry to file '%s': %v", l.file.Name(), err)
		return fmt.Errorf("failed to write audit entry: %w", err)
	}
	return nil
}

// LogLegacy is the non-context version of Log for backward compatibility.
// It calls Log with a background context.
func (l *FileAuditLogger) LogLegacy(entry AuditEntry) error {
	return l.Log(context.Background(), entry)
}

// LogOp implements the AuditLogger interface's LogOp method.
// It creates an AuditEntry with the provided parameters and logs it.
// The context is used to extract correlation ID and for context-aware logging.
func (l *FileAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	// Use default context if nil is provided for backward compatibility
	if ctx == nil {
		ctx = context.Background()
	}

	// Make a copy of inputs to avoid modifying the original map
	inputsCopy := make(map[string]interface{})
	if inputs != nil {
		for k, v := range inputs {
			inputsCopy[k] = v
		}
	}

	// Add correlation ID from context if not already present in inputs
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" && inputsCopy != nil {
		// Only add if not already present
		if _, exists := inputsCopy["correlation_id"]; !exists {
			inputsCopy["correlation_id"] = correlationID
		}
	}

	// Create a new entry with current timestamp
	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: operation,
		Status:    status,
		Inputs:    inputsCopy,
		Outputs:   outputs,
	}

	// Add message based on status and operation
	var message string
	switch status {
	case "Success":
		message = fmt.Sprintf("%s completed successfully", operation)
	case "InProgress":
		message = fmt.Sprintf("%s started", operation)
	case "Failure":
		message = fmt.Sprintf("%s failed", operation)
	default:
		message = fmt.Sprintf("%s - %s", operation, status)
	}
	entry.Message = message

	// Add error information if provided
	if err != nil {
		errorType := "GeneralError"

		// Extract error category for categorized errors
		if catErr, ok := llm.IsCategorizedError(err); ok {
			category := catErr.Category()
			errorType = fmt.Sprintf("Error:%s", category.String())
		}

		entry.Error = &ErrorInfo{
			Message: err.Error(),
			Type:    errorType,
		}
	}

	// Log the entry with context
	return l.Log(ctx, entry)
}

// LogOpLegacy is the non-context version of LogOp for backward compatibility.
// It calls LogOp with a background context.
func (l *FileAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return l.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

// Close properly closes the log file.
// It ensures thread safety with a mutex lock and prevents double-closing.
func (l *FileAuditLogger) Close() error {
	// Create a context with a correlation ID for close logs
	ctx := logutil.WithCorrelationID(context.Background())

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.logger.InfoContext(ctx, "Closing audit log file: %s", l.file.Name())
		err := l.file.Close()
		l.file = nil // Prevent double close
		if err != nil {
			l.logger.ErrorContext(ctx, "Error closing audit log file: %v", err)
			return err
		}
	}
	return nil
}

// NoOpAuditLogger implements AuditLogger with no-op methods.
// This implementation is used when audit logging is disabled.
type NoOpAuditLogger struct{}

// NewNoOpAuditLogger creates a new NoOpAuditLogger instance.
func NewNoOpAuditLogger() *NoOpAuditLogger {
	return &NoOpAuditLogger{}
}

// Log implements the AuditLogger interface but performs no action.
// It always returns nil (no error).
func (l *NoOpAuditLogger) Log(ctx context.Context, entry AuditEntry) error {
	return nil // Do nothing
}

// LogLegacy implements the backward compatibility method.
func (l *NoOpAuditLogger) LogLegacy(entry AuditEntry) error {
	return nil // Do nothing
}

// LogOp implements the AuditLogger interface's LogOp method but performs no action.
// It always returns nil (no error).
func (l *NoOpAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return nil // Do nothing
}

// LogOpLegacy implements the backward compatibility method.
func (l *NoOpAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return nil // Do nothing
}

// Close implements the AuditLogger interface but performs no action.
// It always returns nil (no error).
func (l *NoOpAuditLogger) Close() error {
	return nil // Do nothing
}

// Compile-time checks to ensure implementations satisfy the AuditLogger interface.
var _ AuditLogger = (*FileAuditLogger)(nil)
var _ AuditLogger = (*NoOpAuditLogger)(nil)
