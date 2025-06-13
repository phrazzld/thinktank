package logutil

import (
	"context"
	"strings"
	"testing"
)

func TestBufferLogger_NewBufferLogger(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	if logger == nil {
		t.Error("Expected non-nil BufferLogger")
	}

	// Verify initial state
	logs := logger.GetLogs()
	if len(logs) != 0 {
		t.Errorf("Expected empty logs initially, got %d logs", len(logs))
	}
}

func TestBufferLogger_BasicLogging(t *testing.T) {
	// Test with DebugLevel to capture all messages
	logger := NewBufferLogger(DebugLevel)

	// Test Debug
	logger.Debug("debug message")
	logs := logger.GetLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log after Debug, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "debug message") {
		t.Errorf("Expected log to contain 'debug message', got: %s", logs[0])
	}

	// Test Info
	logger.Info("info message")
	logs = logger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs after Info, got %d", len(logs))
	}

	// Test Warn
	logger.Warn("warn message")
	logs = logger.GetLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs after Warn, got %d", len(logs))
	}

	// Test Error
	logger.Error("error message")
	logs = logger.GetLogs()
	if len(logs) != 4 {
		t.Errorf("Expected 4 logs after Error, got %d", len(logs))
	}

	// Test Fatal
	logger.Fatal("fatal message")
	logs = logger.GetLogs()
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs after Fatal, got %d", len(logs))
	}
}

func TestBufferLogger_PrintFunctions(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	// Test Println
	logger.Println("println message")
	logs := logger.GetLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log after Println, got %d", len(logs))
	}

	// Test Printf
	logger.Printf("printf message %d", 42)
	logs = logger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs after Printf, got %d", len(logs))
	}
	if !strings.Contains(logs[1], "42") {
		t.Errorf("Expected log to contain '42', got: %s", logs[1])
	}
}

func TestBufferLogger_GetLogsAsString(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	logger.Info("first message")
	logger.Error("second message")

	logsString := logger.GetLogsAsString()
	if !strings.Contains(logsString, "first message") {
		t.Errorf("Expected logs string to contain 'first message', got: %s", logsString)
	}
	if !strings.Contains(logsString, "second message") {
		t.Errorf("Expected logs string to contain 'second message', got: %s", logsString)
	}
}

func TestBufferLogger_ClearLogs(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	logger.Info("message 1")
	logger.Info("message 2")

	logs := logger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs before clear, got %d", len(logs))
	}

	logger.ClearLogs()

	logs = logger.GetLogs()
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs after clear, got %d", len(logs))
	}
}

func TestBufferLogger_GetLogEntries(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	logger.Info("info message")
	logger.Error("error message")

	entries := logger.GetLogEntries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(entries))
	}

	// Verify entries contain the expected information
	if entries[0].Message != "info message" {
		t.Errorf("Expected first entry message 'info message', got: %s", entries[0].Message)
	}
	if entries[1].Message != "error message" {
		t.Errorf("Expected second entry message 'error message', got: %s", entries[1].Message)
	}
}

func TestBufferLogger_GetAllCorrelationIDs(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	// Test with context containing correlation ID
	ctx := WithCorrelationID(context.Background())
	logger.InfoContext(ctx, "message with correlation ID")

	correlationIDs := logger.GetAllCorrelationIDs()
	if len(correlationIDs) == 0 {
		t.Error("Expected at least one correlation ID")
	}
}

func TestBufferLogger_ContextLogging(t *testing.T) {
	// Test with DebugLevel to capture all messages including Debug
	logger := NewBufferLogger(DebugLevel)
	ctx := context.Background()

	// Test DebugContext
	logger.DebugContext(ctx, "debug context message")
	logs := logger.GetLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log after DebugContext, got %d", len(logs))
	}

	// Test InfoContext
	logger.InfoContext(ctx, "info context message")
	logs = logger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs after InfoContext, got %d", len(logs))
	}

	// Test WarnContext
	logger.WarnContext(ctx, "warn context message")
	logs = logger.GetLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs after WarnContext, got %d", len(logs))
	}

	// Test ErrorContext
	logger.ErrorContext(ctx, "error context message")
	logs = logger.GetLogs()
	if len(logs) != 4 {
		t.Errorf("Expected 4 logs after ErrorContext, got %d", len(logs))
	}

	// Test FatalContext
	logger.FatalContext(ctx, "fatal context message")
	logs = logger.GetLogs()
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs after FatalContext, got %d", len(logs))
	}
}

func TestBufferLogger_WithContext(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)
	ctx := WithCorrelationID(context.Background())

	contextLogger := logger.WithContext(ctx)
	if contextLogger == nil {
		t.Error("Expected non-nil context logger")
	}

	// The returned logger should be a new instance with shared entries
	// but different context
	if contextLogger == logger {
		t.Error("Expected WithContext to return a new logger instance")
	}

	// Test that the new logger shares the same log entries
	logger.Info("original logger message")
	contextLogger.Info("context logger message")

	// Both should see both messages since they share the same entries slice
	logs := logger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs in original logger, got %d", len(logs))
	}

	// Cast back to BufferLogger to access GetLogs method
	contextBufferLogger, ok := contextLogger.(*BufferLogger)
	if !ok {
		t.Error("Expected WithContext to return a *BufferLogger")
		return
	}

	contextLogs := contextBufferLogger.GetLogs()
	if len(contextLogs) != 2 {
		t.Errorf("Expected 2 logs in context logger, got %d", len(contextLogs))
	}
}

func TestBufferLogger_ConcurrentAccess(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	// Test concurrent logging
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			logger.Info("goroutine 1 message %d", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			logger.Error("goroutine 2 message %d", i)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	logs := logger.GetLogs()
	if len(logs) != 20 {
		t.Errorf("Expected 20 logs from concurrent access, got %d", len(logs))
	}
}

func TestBufferLogger_EmptyState(t *testing.T) {
	logger := NewBufferLogger(InfoLevel)

	// Test empty state operations
	logs := logger.GetLogs()
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs initially, got %d", len(logs))
	}

	logsString := logger.GetLogsAsString()
	if logsString != "" {
		t.Errorf("Expected empty logs string initially, got: %s", logsString)
	}

	entries := logger.GetLogEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 log entries initially, got %d", len(entries))
	}

	correlationIDs := logger.GetAllCorrelationIDs()
	if len(correlationIDs) != 0 {
		t.Errorf("Expected 0 correlation IDs initially, got %d", len(correlationIDs))
	}
}

func TestBufferLogger_LogLevelFiltering(t *testing.T) {
	// Test that log level filtering works correctly

	// Test InfoLevel logger - should filter out Debug messages
	infoLogger := NewBufferLogger(InfoLevel)
	infoLogger.Debug("debug message should be filtered")
	infoLogger.Info("info message should appear")
	infoLogger.Warn("warn message should appear")
	infoLogger.Error("error message should appear")

	logs := infoLogger.GetLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs with InfoLevel filtering, got %d", len(logs))
	}

	// Test WarnLevel logger - should filter out Debug and Info messages
	warnLogger := NewBufferLogger(WarnLevel)
	warnLogger.Debug("debug message should be filtered")
	warnLogger.Info("info message should be filtered")
	warnLogger.Warn("warn message should appear")
	warnLogger.Error("error message should appear")

	logs = warnLogger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with WarnLevel filtering, got %d", len(logs))
	}

	// Test ErrorLevel logger - should only show Error messages
	errorLogger := NewBufferLogger(ErrorLevel)
	errorLogger.Debug("debug message should be filtered")
	errorLogger.Info("info message should be filtered")
	errorLogger.Warn("warn message should be filtered")
	errorLogger.Error("error message should appear")

	logs = errorLogger.GetLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log with ErrorLevel filtering, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "error message should appear") {
		t.Errorf("Expected error message to appear, got: %s", logs[0])
	}
}
