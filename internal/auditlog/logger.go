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
// It handles errors internally, logs them to stderr, but doesn't fail the application.
func (l *FileLogger) Log(event AuditEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Ensure level is set
	if event.Level == "" {
		event.Level = "INFO"
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		// Log to stderr but avoid recursion
		errLogger("[ERROR] Failed to marshal audit event: %v", err)
		return
	}

	// Add newline and write to file
	jsonBytes = append(jsonBytes, '\n')
	if _, err := l.file.Write(jsonBytes); err != nil {
		errLogger("[ERROR] Failed to write audit event: %v", err)
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
