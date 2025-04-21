// Package logutil provides logging utilities for the architect project
package logutil

import (
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
