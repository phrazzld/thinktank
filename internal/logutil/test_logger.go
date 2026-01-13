// Package logutil provides logging utilities for the thinktank project
package logutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestLogger is a logger implementation for testing that captures log messages
type TestLogger struct {
	t                     *testing.T
	logs                  []string
	errorLogs             []string
	logsMutex             sync.Mutex
	prefix                string
	level                 LogLevel
	autoFailMode          bool
	expectedErrorPatterns []string
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) *TestLogger {
	logger := &TestLogger{
		t:                     t,
		logs:                  []string{},
		errorLogs:             []string{},
		level:                 DebugLevel,
		autoFailMode:          true,
		expectedErrorPatterns: []string{},
	}

	// Use t.Cleanup to automatically fail tests that logged errors
	t.Cleanup(func() {
		if logger.autoFailMode && logger.HasUnexpectedErrorLogs() {
			unexpectedErrors := logger.GetUnexpectedErrorLogs()
			t.Errorf("Test failed due to error-level logs captured:\n%s",
				strings.Join(unexpectedErrors, "\n"))
		}
	})

	return logger
}

// NewTestLoggerWithoutAutoFail creates a test logger that won't automatically fail on error logs
func NewTestLoggerWithoutAutoFail(t *testing.T) *TestLogger {
	return &TestLogger{
		t:                     t,
		logs:                  []string{},
		errorLogs:             []string{},
		level:                 DebugLevel,
		autoFailMode:          false,
		expectedErrorPatterns: []string{},
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
		formattedMsg := fmt.Sprintf("[ERROR] %s%s", l.prefix, msg)
		l.t.Logf("%s", formattedMsg)
		l.captureLog(formattedMsg)
		l.captureErrorLog(formattedMsg)
	}
}

// Fatal logs a fatal message
func (l *TestLogger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	formattedMsg := fmt.Sprintf("[FATAL] %s%s", l.prefix, msg)
	l.t.Logf("%s", formattedMsg)
	l.captureLog(formattedMsg)
	l.captureErrorLog(formattedMsg)
	// Don't call os.Exit in tests
}

// Println implements LoggerInterface by logging at info level
func (l *TestLogger) Println(v ...interface{}) {
	l.Info("%s", fmt.Sprintln(v...))
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

// captureErrorLog captures an error-level log message for later inspection
func (l *TestLogger) captureErrorLog(msg string) {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.errorLogs = append(l.errorLogs, msg)
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

// ClearTestLogs clears all captured log messages including error logs
func (l *TestLogger) ClearTestLogs() {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.logs = []string{}
	l.errorLogs = []string{}
}

// GetErrorLogs returns all captured error-level log messages (ERROR and FATAL)
func (l *TestLogger) GetErrorLogs() []string {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	// Return a copy to avoid race conditions
	errorLogs := make([]string, len(l.errorLogs))
	copy(errorLogs, l.errorLogs)
	return errorLogs
}

// HasErrorLogs returns true if any error-level logs were captured
func (l *TestLogger) HasErrorLogs() bool {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	return len(l.errorLogs) > 0
}

// ClearErrorLogs clears all captured error-level log messages
func (l *TestLogger) ClearErrorLogs() {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.errorLogs = []string{}
}

// AssertNoErrorLogs fails the test if any error-level logs were captured
func (l *TestLogger) AssertNoErrorLogs() {
	if l.HasErrorLogs() {
		errorLogs := l.GetErrorLogs()
		l.t.Errorf("Expected no error-level logs, but found %d:\n%s",
			len(errorLogs), strings.Join(errorLogs, "\n"))
	}
}

// DisableAutoFail disables automatic test failure on error logs
func (l *TestLogger) DisableAutoFail() {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.autoFailMode = false
}

// EnableAutoFail enables automatic test failure on error logs
func (l *TestLogger) EnableAutoFail() {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.autoFailMode = true
}

// ExpectError declares that an error message matching the given pattern is expected
// and should not cause test failure. Pattern matching uses substring matching.
func (l *TestLogger) ExpectError(pattern string) {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.expectedErrorPatterns = append(l.expectedErrorPatterns, pattern)
}

// isExpectedError checks if an error log matches any expected error pattern
func (l *TestLogger) isExpectedError(errorLog string) bool {
	for _, pattern := range l.expectedErrorPatterns {
		if strings.Contains(errorLog, pattern) {
			return true
		}
	}
	return false
}

// GetUnexpectedErrorLogs returns error logs that don't match expected patterns
func (l *TestLogger) GetUnexpectedErrorLogs() []string {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()

	var unexpectedErrors []string
	for _, errorLog := range l.errorLogs {
		if !l.isExpectedError(errorLog) {
			unexpectedErrors = append(unexpectedErrors, errorLog)
		}
	}
	return unexpectedErrors
}

// HasUnexpectedErrorLogs returns true if any unexpected error logs were captured
func (l *TestLogger) HasUnexpectedErrorLogs() bool {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()

	for _, errorLog := range l.errorLogs {
		if !l.isExpectedError(errorLog) {
			return true
		}
	}
	return false
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
		l.captureErrorLog(logMsg)
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
	l.captureErrorLog(logMsg)
	// Don't call os.Exit in tests
}

// WithContext returns a logger with context information
func (l *TestLogger) WithContext(ctx context.Context) LoggerInterface {
	// For test logger, we just return the same logger
	// A real implementation might create a new logger with context attached
	return l
}
