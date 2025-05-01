package logutil

import (
	"context"
	"fmt"
	"sync"
)

// BufferLogger is a simple logger that captures log messages in memory
// It's useful for tests where you want to capture logs but don't have a testing.T
type BufferLogger struct {
	logs      []string
	logsMutex sync.Mutex
	prefix    string
	level     LogLevel
	ctx       context.Context
}

// NewBufferLogger creates a new buffer logger
func NewBufferLogger() *BufferLogger {
	return &BufferLogger{
		logs:  []string{},
		level: DebugLevel,
		ctx:   context.Background(),
	}
}

// Debug logs a debug message
func (l *BufferLogger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		msg := fmt.Sprintf(format, args...)
		l.captureLog(fmt.Sprintf("[DEBUG] %s%s", l.prefix, msg))
	}
}

// Info logs an info message
func (l *BufferLogger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		msg := fmt.Sprintf(format, args...)
		l.captureLog(fmt.Sprintf("[INFO] %s%s", l.prefix, msg))
	}
}

// Warn logs a warning message
func (l *BufferLogger) Warn(format string, args ...interface{}) {
	if l.level <= WarnLevel {
		msg := fmt.Sprintf(format, args...)
		l.captureLog(fmt.Sprintf("[WARN] %s%s", l.prefix, msg))
	}
}

// Error logs an error message
func (l *BufferLogger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		msg := fmt.Sprintf(format, args...)
		l.captureLog(fmt.Sprintf("[ERROR] %s%s", l.prefix, msg))
	}
}

// Fatal logs a fatal message
func (l *BufferLogger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.captureLog(fmt.Sprintf("[FATAL] %s%s", l.prefix, msg))
	// Don't call os.Exit in tests
}

// Println implements LoggerInterface by logging at info level
func (l *BufferLogger) Println(v ...interface{}) {
	l.Info(fmt.Sprintln(v...))
}

// Printf implements LoggerInterface by logging at info level
func (l *BufferLogger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// captureLog captures a log message for later inspection
func (l *BufferLogger) captureLog(msg string) {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.logs = append(l.logs, msg)
}

// GetLogs returns all captured log messages
func (l *BufferLogger) GetLogs() []string {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	// Return a copy to avoid race conditions
	logs := make([]string, len(l.logs))
	copy(logs, l.logs)
	return logs
}

// ClearLogs clears all captured log messages
func (l *BufferLogger) ClearLogs() {
	l.logsMutex.Lock()
	defer l.logsMutex.Unlock()
	l.logs = []string{}
}

// Context-aware logging methods

// DebugContext logs a debug message with context
func (l *BufferLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= DebugLevel {
		// Use a new logger with context instead of manual formatting
		withCtx := l.WithContext(ctx)
		withCtx.Debug(format, args...)
	}
}

// InfoContext logs an info message with context
func (l *BufferLogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= InfoLevel {
		// Use a new logger with context instead of manual formatting
		withCtx := l.WithContext(ctx)
		withCtx.Info(format, args...)
	}
}

// WarnContext logs a warning message with context
func (l *BufferLogger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= WarnLevel {
		// Use a new logger with context instead of manual formatting
		withCtx := l.WithContext(ctx)
		withCtx.Warn(format, args...)
	}
}

// ErrorContext logs an error message with context
func (l *BufferLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		// Use a new logger with context instead of manual formatting
		withCtx := l.WithContext(ctx)
		withCtx.Error(format, args...)
	}
}

// FatalContext logs a fatal message with context
func (l *BufferLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	// Use a new logger with context instead of manual formatting
	withCtx := l.WithContext(ctx)
	withCtx.Fatal(format, args...)
	// Don't call os.Exit in tests
}

// WithContext returns a logger with context information
func (l *BufferLogger) WithContext(ctx context.Context) LoggerInterface {
	newLogger := &BufferLogger{
		logs:   l.logs,
		level:  l.level,
		prefix: l.prefix,
		ctx:    ctx,
	}
	return newLogger
}
