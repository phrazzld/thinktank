// internal/logutil/logutil.go
package logutil

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

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
	s.Logger.Printf("[DEBUG] "+format, v...)
}

// Info implements LoggerInterface.Info
func (s *StdLoggerAdapter) Info(format string, v ...interface{}) {
	s.Logger.Printf("[INFO] "+format, v...)
}

// Warn implements LoggerInterface.Warn
func (s *StdLoggerAdapter) Warn(format string, v ...interface{}) {
	s.Logger.Printf("[WARN] "+format, v...)
}

// Error implements LoggerInterface.Error
func (s *StdLoggerAdapter) Error(format string, v ...interface{}) {
	s.Logger.Printf("[ERROR] "+format, v...)
}

// Fatal implements LoggerInterface.Fatal
func (s *StdLoggerAdapter) Fatal(format string, v ...interface{}) {
	s.Logger.Fatalf("[FATAL] "+format, v...)
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

// Logger provides structured logging with levels and color support
type Logger struct {
	level      LogLevel  // Current log level
	writer     io.Writer // Where to write logs (typically os.Stderr)
	prefix     string    // Prefix for all log messages
	useColors  bool      // Whether to use colors in output
	debugColor *color.Color
	infoColor  *color.Color
	warnColor  *color.Color
	errorColor *color.Color
}

// Ensure Logger implements LoggerInterface
var _ LoggerInterface = (*Logger)(nil)

// NewLogger creates a new logger with the specified configuration
func NewLogger(level LogLevel, writer io.Writer, prefix string, useColors bool) *Logger {
	if writer == nil {
		writer = os.Stderr
	}

	logger := &Logger{
		level:      level,
		writer:     writer,
		prefix:     prefix,
		useColors:  useColors,
		debugColor: color.New(color.FgCyan),
		infoColor:  color.New(color.FgGreen),
		warnColor:  color.New(color.FgYellow),
		errorColor: color.New(color.FgRed, color.Bold),
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
	if level < l.level {
		return
	}

	formattedMsg := l.formatMessage(level, format, args...)

	if l.useColors {
		switch level {
		case DebugLevel:
			fmt.Fprintln(l.writer, l.debugColor.Sprint(formattedMsg))
		case InfoLevel:
			fmt.Fprintln(l.writer, l.infoColor.Sprint(formattedMsg))
		case WarnLevel:
			fmt.Fprintln(l.writer, l.warnColor.Sprint(formattedMsg))
		case ErrorLevel:
			fmt.Fprintln(l.writer, l.errorColor.Sprint(formattedMsg))
		}
	} else {
		fmt.Fprintln(l.writer, formattedMsg)
	}
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
	os.Exit(1)
}

// SetLevel changes the current log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetPrefix changes the prefix used in log messages
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// SetUseColors toggles color output
func (l *Logger) SetUseColors(useColors bool) {
	l.useColors = useColors
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
