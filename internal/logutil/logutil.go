// internal/logutil/logutil.go
package logutil

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Variable to allow mocking os.Exit in tests
var osExit = os.Exit

// LoggerInterface defines the common logging interface
// This allows both our structured Logger and the standard log.Logger
// to be used interchangeably
type LoggerInterface interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}

// StdLoggerAdapter provides a compatibility layer for standard log.Logger
type StdLoggerAdapter struct {
	*log.Logger
}

// NewStdLoggerAdapter wraps a standard logger
func NewStdLoggerAdapter(logger *log.Logger) *StdLoggerAdapter {
	return &StdLoggerAdapter{Logger: logger}
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
	mu     sync.Mutex // Mutex to protect concurrent access
	level  LogLevel   // Current log level
	writer io.Writer  // Where to write logs (typically os.Stderr)
	prefix string     // Prefix for all log messages
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
	}

	return logger
}

// formatMessage creates a formatted log message with timestamp and level
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s%s", timestamp, level.String(), l.prefix, msg)
}

// log logs a message at the specified level if it meets the threshold
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	formattedMsg := l.formatMessage(level, format, args...)
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
