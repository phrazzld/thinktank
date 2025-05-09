// Package logutil provides unified logging functionality with support for
// structured logging, contextual information, and correlation IDs.
//
// The package provides a LoggerInterface that abstracts over different logging
// implementations. This interface supports:
// - Context-aware logging methods to propagate correlation IDs
// - Structured logging with key-value pairs
// - Multiple log levels (Debug, Info, Warn, Error, Fatal)
// - Stream separation for different log levels
//
// Key components:
// - LoggerInterface: The primary interface that all loggers implement
// - Logger: A basic logger implementation with correlation ID support
// - SlogLogger: An implementation using Go's log/slog package for structured logging
package logutil

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Variable to allow mocking os.Exit in tests
var osExit = os.Exit

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

// CorrelationIDKey is the context key for correlation ID
const CorrelationIDKey ContextKey = "correlation_id"

// WithCorrelationID adds a correlation ID to the context.
// If an ID already exists in the context, it is preserved.
// If no correlation ID is present, a new UUID is generated.
//
// This function has been enhanced to accept an optional ID parameter.
// When called with no parameters, it behaves like the original WithCorrelationID.
// When called with an ID parameter, it behaves like WithCustomCorrelationID.
//
// Usage:
//
//	// Generate and add a new correlation ID (preserves any existing ID)
//	ctx = logutil.WithCorrelationID(ctx)
//
//	// Set a specific correlation ID (replaces any existing ID)
//	ctx = logutil.WithCorrelationID(ctx, "custom-id-123")
func WithCorrelationID(ctx context.Context, id ...string) context.Context {
	// Check if correlation ID already exists and if we're using the no-args version
	// or the empty string version - in both cases, preserve existing ID
	if existingID := GetCorrelationID(ctx); existingID != "" {
		if len(id) == 0 || (len(id) > 0 && id[0] == "") {
			// Preserve existing ID
			return ctx
		}
	}

	// If a custom ID is provided and it's not empty, use it
	if len(id) > 0 && id[0] != "" {
		return context.WithValue(ctx, CorrelationIDKey, id[0])
	}

	// Otherwise generate a new ID
	newID := uuid.New().String()
	return context.WithValue(ctx, CorrelationIDKey, newID)
}

// WithCustomCorrelationID adds a custom correlation ID to the context.
// This is maintained for backward compatibility, but new code should use
// WithCorrelationID(ctx, id) instead.
func WithCustomCorrelationID(ctx context.Context, id string) context.Context {
	return WithCorrelationID(ctx, id)
}

// GetCorrelationID retrieves correlation ID from context, or returns
// empty string if not present
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	id, ok := ctx.Value(CorrelationIDKey).(string)
	if !ok {
		return ""
	}
	return id
}

// LoggerInterface defines a comprehensive logging interface with context-awareness
// that can be implemented by different logger backends (e.g., slog, zerolog, etc.)
type LoggerInterface interface {
	// Context-aware logging methods with structured key-value pairs
	// The args parameter supports alternating key-value pairs that will be included
	// in the structured log output. For example:
	//   logger.InfoContext(ctx, "user logged in", "user_id", 123, "ip", "192.168.1.1")
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

// StdLoggerAdapter provides a compatibility layer for standard log.Logger
type StdLoggerAdapter struct {
	*log.Logger
	ctx context.Context
}

// NewStdLoggerAdapter wraps a standard logger
func NewStdLoggerAdapter(logger *log.Logger) *StdLoggerAdapter {
	return &StdLoggerAdapter{Logger: logger, ctx: context.Background()}
}

// WithContext returns a new logger with the given context
func (s *StdLoggerAdapter) WithContext(ctx context.Context) LoggerInterface {
	return &StdLoggerAdapter{
		Logger: s.Logger,
		ctx:    ctx,
	}
}

// Debug implements LoggerInterface.Debug
func (s *StdLoggerAdapter) Debug(format string, v ...interface{}) {
	s.Printf("[DEBUG] "+format, v...)
}

// Info implements LoggerInterface.Info
func (s *StdLoggerAdapter) Info(format string, v ...interface{}) {
	s.Printf("[INFO] "+format, v...)
}

// Warn implements LoggerInterface.Warn
func (s *StdLoggerAdapter) Warn(format string, v ...interface{}) {
	s.Printf("[WARN] "+format, v...)
}

// Error implements LoggerInterface.Error
func (s *StdLoggerAdapter) Error(format string, v ...interface{}) {
	s.Printf("[ERROR] "+format, v...)
}

// Fatal implements LoggerInterface.Fatal
func (s *StdLoggerAdapter) Fatal(format string, v ...interface{}) {
	s.Printf("[FATAL] "+format, v...)
	osExit(1)
}

// DebugContext implements context-aware debug logging
func (s *StdLoggerAdapter) DebugContext(ctx context.Context, format string, v ...interface{}) {
	id := GetCorrelationID(ctx)
	msg := fmt.Sprintf(format, v...)
	if id != "" {
		s.Printf("[DEBUG] %s [correlation_id=%s]", msg, id)
	} else {
		s.Debug(format, v...)
	}
}

// InfoContext implements context-aware info logging
func (s *StdLoggerAdapter) InfoContext(ctx context.Context, format string, v ...interface{}) {
	id := GetCorrelationID(ctx)
	msg := fmt.Sprintf(format, v...)
	if id != "" {
		s.Printf("[INFO] %s [correlation_id=%s]", msg, id)
	} else {
		s.Info(format, v...)
	}
}

// WarnContext implements context-aware warn logging
func (s *StdLoggerAdapter) WarnContext(ctx context.Context, format string, v ...interface{}) {
	id := GetCorrelationID(ctx)
	msg := fmt.Sprintf(format, v...)
	if id != "" {
		s.Printf("[WARN] %s [correlation_id=%s]", msg, id)
	} else {
		s.Warn(format, v...)
	}
}

// ErrorContext implements context-aware error logging
func (s *StdLoggerAdapter) ErrorContext(ctx context.Context, format string, v ...interface{}) {
	id := GetCorrelationID(ctx)
	msg := fmt.Sprintf(format, v...)
	if id != "" {
		s.Printf("[ERROR] %s [correlation_id=%s]", msg, id)
	} else {
		s.Error(format, v...)
	}
}

// FatalContext implements context-aware fatal logging
func (s *StdLoggerAdapter) FatalContext(ctx context.Context, format string, v ...interface{}) {
	id := GetCorrelationID(ctx)
	msg := fmt.Sprintf(format, v...)
	if id != "" {
		s.Printf("[FATAL] %s [correlation_id=%s]", msg, id)
	} else {
		s.Fatal(format, v...)
	}
	osExit(1)
}

// LogLevel represents different logging severity levels
type LogLevel int

const (
	// Log levels in increasing order of severity
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with levels
type Logger struct {
	mu     sync.Mutex      // Mutex to protect concurrent access
	level  LogLevel        // Current log level
	writer io.Writer       // Where to write logs (typically os.Stderr)
	prefix string          // Prefix for all log messages
	ctx    context.Context // Context for correlation ID
}

// Ensure Logger implements LoggerInterface
var _ LoggerInterface = (*Logger)(nil)

// NewLogger creates a new logger with the specified configuration
func NewLogger(level LogLevel, writer io.Writer, prefix string) *Logger {
	if writer == nil {
		writer = os.Stderr
	}

	logger := &Logger{
		level:  level,
		writer: writer,
		prefix: prefix,
		ctx:    context.Background(),
	}

	return logger
}

// WithContext returns a new logger with context information attached
func (l *Logger) WithContext(ctx context.Context) LoggerInterface {
	newLogger := &Logger{
		level:  l.level,
		writer: l.writer,
		prefix: l.prefix,
		ctx:    ctx,
	}
	return newLogger
}

// formatMessage creates a formatted log message with timestamp and level
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s%s", timestamp, level.String(), l.prefix, msg)
}

// formatMessageWithCorrelationID creates a formatted log message with timestamp, level, and correlation ID
func (l *Logger) formatMessageWithCorrelationID(ctx context.Context, level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		// Format with correlation ID as a structured field, separated from the message itself
		return fmt.Sprintf("%s [%s] %s%s [correlation_id=%s]",
			timestamp, level.String(), l.prefix, msg, correlationID)
	}
	return fmt.Sprintf("%s [%s] %s%s", timestamp, level.String(), l.prefix, msg)
}

// log logs a message at the specified level if it meets the threshold
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Check if we have a correlation ID in the logger's context
	correlationID := GetCorrelationID(l.ctx)
	var formattedMsg string

	if correlationID != "" {
		formattedMsg = l.formatMessageWithCorrelationID(l.ctx, level, format, args...)
	} else {
		formattedMsg = l.formatMessage(level, format, args...)
	}

	_, _ = fmt.Fprintln(l.writer, formattedMsg)
}

// logWithContext logs a message with context-specific correlation ID
func (l *Logger) logWithContext(ctx context.Context, level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	formattedMsg := l.formatMessageWithCorrelationID(ctx, level, format, args...)
	_, _ = fmt.Fprintln(l.writer, formattedMsg)
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info logs a message at INFO level
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal logs a message at ERROR level and then exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
	osExit(1)
}

// DebugContext logs a message at DEBUG level with context information
func (l *Logger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, DebugLevel, format, args...)
}

// InfoContext logs a message at INFO level with context information
func (l *Logger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, InfoLevel, format, args...)
}

// WarnContext logs a message at WARN level with context information
func (l *Logger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, WarnLevel, format, args...)
}

// ErrorContext logs a message at ERROR level with context information
func (l *Logger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, ErrorLevel, format, args...)
}

// FatalContext logs a message at ERROR level with context information and then exits the program
func (l *Logger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, ErrorLevel, format, args...)
	osExit(1)
}

// SetLevel changes the current log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetPrefix changes the prefix used in log messages
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Println implements LoggerInterface by logging at info level
func (l *Logger) Println(v ...interface{}) {
	l.Info(fmt.Sprintln(v...))
}

// Printf implements LoggerInterface by logging at info level
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// ParseLogLevel converts a string to a LogLevel
func ParseLogLevel(level string) (LogLevel, error) {
	switch level {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}
