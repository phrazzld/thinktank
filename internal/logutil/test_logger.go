// Package logutil provides logging utilities for the thinktank project
package logutil

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// TestLogger is a logger implementation for testing that captures log messages
type TestLogger struct {
	t         *testing.T
	logs      []string
	logsMutex sync.Mutex
	prefix    string
	level     LogLevel
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{
		t:     t,
		logs:  []string{},
		level: DebugLevel,
	}
}

// Debug logs a debug message
func (l *TestLogger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		msg := fmt.Sprintf(format, args...)
		l.t.Logf("[DEBUG] %s%s", l.prefix, msg)
		l.captureLog(fmt.Sprintf("[DEBUG] %s%s", l.prefix, msg))
	}
}

// Info logs an info message
func (l *TestLogger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		msg := fmt.Sprintf(format, args...)
		l.t.Logf("[INFO] %s%s", l.prefix, msg)
		l.captureLog(fmt.Sprintf("[INFO] %s%s", l.prefix, msg))
	}
}

// Warn logs a warning message
func (l *TestLogger) Warn(format string, args ...interface{}) {
	if l.level <= WarnLevel {
		msg := fmt.Sprintf(format, args...)
		l.t.Logf("[WARN] %s%s", l.prefix, msg)
		l.captureLog(fmt.Sprintf("[WARN] %s%s", l.prefix, msg))
	}
}

// Error logs an error message
func (l *TestLogger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		msg := fmt.Sprintf(format, args...)
		l.t.Logf("[ERROR] %s%s", l.prefix, msg)
		l.captureLog(fmt.Sprintf("[ERROR] %s%s", l.prefix, msg))
	}
}

// Fatal logs a fatal message
func (l *TestLogger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.t.Logf("[FATAL] %s%s", l.prefix, msg)
	l.captureLog(fmt.Sprintf("[FATAL] %s%s", l.prefix, msg))
	// Don't call os.Exit in tests
}

// Println implements LoggerInterface by logging at info level
func (l *TestLogger) Println(v ...interface{}) {
	l.Info(fmt.Sprintln(v...))
}

// Printf implements LoggerInterface by logging at info level
func (l *TestLogger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// captureLog captures a log message for later inspection
func (l *TestLogger) captureLog(msg string) {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.logs = append(l.logs, msg)
}

// GetTestLogs returns all captured log messages
func (l *TestLogger) GetTestLogs() []string {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	// Return a copy to avoid race conditions
	logs := make([]string, len(l.logs))
	copy(logs, l.logs)
	return logs
}

// ClearTestLogs clears all captured log messages
func (l *TestLogger) ClearTestLogs() {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.logs = []string{}
}

// Context-aware logging methods

// DebugContext logs a debug message with context
func (l *TestLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= DebugLevel {
		msg := fmt.Sprintf(format, args...)
		correlationID := GetCorrelationID(ctx)
		// Format the log message with correlation ID as a structured field
		logMsg := fmt.Sprintf("[DEBUG] %s%s [correlation_id=%s]", l.prefix, msg, correlationID)
		l.t.Logf("%s", logMsg)
		l.captureLog(logMsg)
	}
}

// InfoContext logs an info message with context
func (l *TestLogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= InfoLevel {
		msg := fmt.Sprintf(format, args...)
		correlationID := GetCorrelationID(ctx)
		// Format the log message with correlation ID as a structured field
		logMsg := fmt.Sprintf("[INFO] %s%s [correlation_id=%s]", l.prefix, msg, correlationID)
		l.t.Logf("%s", logMsg)
		l.captureLog(logMsg)
	}
}

// WarnContext logs a warning message with context
func (l *TestLogger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= WarnLevel {
		msg := fmt.Sprintf(format, args...)
		correlationID := GetCorrelationID(ctx)
		// Format the log message with correlation ID as a structured field
		logMsg := fmt.Sprintf("[WARN] %s%s [correlation_id=%s]", l.prefix, msg, correlationID)
		l.t.Logf("%s", logMsg)
		l.captureLog(logMsg)
	}
}

// ErrorContext logs an error message with context
func (l *TestLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		msg := fmt.Sprintf(format, args...)
		correlationID := GetCorrelationID(ctx)
		// Format the log message with correlation ID as a structured field
		logMsg := fmt.Sprintf("[ERROR] %s%s [correlation_id=%s]", l.prefix, msg, correlationID)
		l.t.Logf("%s", logMsg)
		l.captureLog(logMsg)
	}
}

// FatalContext logs a fatal message with context
func (l *TestLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	correlationID := GetCorrelationID(ctx)
	// Format the log message with correlation ID as a structured field
	logMsg := fmt.Sprintf("[FATAL] %s%s [correlation_id=%s]", l.prefix, msg, correlationID)
	l.t.Logf("%s", logMsg)
	l.captureLog(logMsg)
	// Don't call os.Exit in tests
}

// WithContext returns a logger with context information
func (l *TestLogger) WithContext(ctx context.Context) LoggerInterface {
	// For test logger, we just return the same logger
	// A real implementation might create a new logger with context attached
	return l
}
