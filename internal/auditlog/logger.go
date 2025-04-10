// Package auditlog provides structured logging capabilities for the architect tool.
package auditlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// errLogger is used to log internal errors without causing recursive logging issues
var errLogger = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// StructuredLogger defines the interface for structured audit logging.
// It provides methods for logging structured events and cleaning up resources.
type StructuredLogger interface {
	// Log records a structured audit event.
	// Implementations should ensure this method is safe for concurrent use
	// and should handle any errors internally to prevent disruption to the
	// application flow (e.g., by logging errors to the standard logger).
	Log(event AuditEvent)

	// Close releases any resources held by the logger.
	// This should be called when the logger is no longer needed,
	// typically using the defer pattern after logger creation.
	// Implementations should ensure this method is idempotent and
	// safe to call multiple times.
	// 
	// Returns an error if cleanup fails, which the caller may choose
	// to log but typically should not cause the application to fail.
	Close() error
}

// FileLogger implements StructuredLogger by writing JSON lines to a file.
// It ensures thread-safety using a mutex and properly manages file resources.
type FileLogger struct {
	file *os.File    // The file handle for writing logs
	mu   sync.Mutex  // Mutex for ensuring thread-safety
}

// NewFileLogger creates a new structured logger that writes to the specified file path.
// It automatically creates the directory if it doesn't exist and opens the file in append mode.
// 
// The function handles:
// - Path validation (empty or invalid paths)
// - Directory creation with proper permissions
// - File opening with appropriate flags
// - Error wrapping for better diagnostics
func NewFileLogger(filePath string) (*FileLogger, error) {
	// Validate file path
	if filePath == "" {
		return nil, fmt.Errorf("log file path cannot be empty")
	}

	// Clean and normalize the path
	filePath = filepath.Clean(filePath)

	// Ensure directory exists with proper permissions
	dir := filepath.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create log directory '%s': %w", dir, err)
		}
	}

	// Open file for appending with create if not exists
	// Using appropriate flags for atomic operations when possible
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	
	// Open the file with proper permissions
	// 0644 = user rw, group r, others r
	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file '%s': %w", filePath, err)
	}

	return &FileLogger{
		file: file,
	}, nil
}

// Log writes an audit event to the log file as a JSON line.
// 
// This method is completely thread-safe and can be called concurrently from multiple goroutines.
// It handles all error conditions gracefully without panicking, including:
// - Nil file handle
// - Closed file
// - JSON marshaling errors
// - File write errors
// 
// Errors are logged to stderr but don't cause the application to fail. This is essential for
// logging systems, which should never disrupt the main application flow.
//
// The method also ensures that events have proper default values for required fields,
// such as timestamp and log level.
func (l *FileLogger) Log(event AuditEvent) {
	// Protect against nil receiver
	if l == nil {
		errLogger("[ERROR] Attempted to log to a nil FileLogger")
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Protect against nil or closed file handle
	if l.file == nil {
		errLogger("[ERROR] Attempted to log to a FileLogger with a nil file handle")
		return
	}

	// Clone the event to avoid modifying the caller's copy
	eventCopy := event
	
	// Ensure timestamp is set
	if eventCopy.Timestamp.IsZero() {
		eventCopy.Timestamp = time.Now().UTC()
	}

	// Ensure level is set
	if eventCopy.Level == "" {
		eventCopy.Level = "INFO"
	}

	// Marshal to JSON with graceful error handling
	jsonBytes, err := json.Marshal(eventCopy)
	if err != nil {
		// Log to stderr but avoid recursion
		errLogger("[ERROR] Failed to marshal audit event: %v", err)
		// Try a simplified version with just the core fields
		simplifiedEvent := AuditEvent{
			Timestamp: eventCopy.Timestamp,
			Level:     eventCopy.Level,
			Operation: eventCopy.Operation,
			Message:   eventCopy.Message + " [marshaling error: full event could not be serialized]",
		}
		
		jsonBytes, err = json.Marshal(simplifiedEvent)
		if err != nil {
			// If even simplified event fails, give up but don't crash
			errLogger("[ERROR] Failed to marshal simplified audit event: %v", err)
			return
		}
	}

	// Add newline and write to file with proper error handling
	jsonBytes = append(jsonBytes, '\n')
	
	_, err = l.file.Write(jsonBytes)
	if err != nil {
		// Log write error with context
		errLogger("[ERROR] Failed to write audit event to log file: %v", err)
		
		// Handle specific error types with contextual information
		if os.IsPermission(err) {
			errLogger("[ERROR] Permission denied when writing to log file. Check file permissions.")
		} else if os.IsNotExist(err) {
			errLogger("[ERROR] Log file no longer exists. It may have been deleted.")
		}
	}
}

// Close flushes any buffered data and closes the underlying file.
// It is safe to call Close multiple times; subsequent calls will return nil.
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// NoopLogger implements StructuredLogger but performs no operations.
// It's used when audit logging is disabled to avoid nil checks in the application code.
type NoopLogger struct{}

// Log implements StructuredLogger.Log but does nothing.
func (l *NoopLogger) Log(event AuditEvent) {
	// Do nothing
}

// Close implements StructuredLogger.Close but does nothing and returns nil.
func (l *NoopLogger) Close() error {
	// Do nothing, return no error
	return nil
}

// NewNoopLogger creates and returns a new NoopLogger instance.
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}
