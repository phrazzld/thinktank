package logutil

import (
	"context"
	"strings"
	"testing"
)

func TestTestLogger_NewTestLogger(t *testing.T) {
	logger := NewTestLogger(t)

	if logger == nil {
		t.Error("Expected non-nil TestLogger")
	}

	// Verify initial state
	logs := logger.GetTestLogs()
	if len(logs) != 0 {
		t.Errorf("Expected empty logs initially, got %d logs", len(logs))
	}
}

func TestTestLogger_BasicLogging(t *testing.T) {
	logger := NewTestLogger(t)

	// Test Debug
	logger.Debug("debug message")
	logs := logger.GetTestLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log after Debug, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "debug message") {
		t.Errorf("Expected log to contain 'debug message', got: %s", logs[0])
	}

	// Test Info
	logger.Info("info message")
	logs = logger.GetTestLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs after Info, got %d", len(logs))
	}

	// Test Warn
	logger.Warn("warn message")
	logs = logger.GetTestLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs after Warn, got %d", len(logs))
	}

	// Test Error
	logger.Error("error message")
	logs = logger.GetTestLogs()
	if len(logs) != 4 {
		t.Errorf("Expected 4 logs after Error, got %d", len(logs))
	}

	// Test Fatal
	logger.Fatal("fatal message")
	logs = logger.GetTestLogs()
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs after Fatal, got %d", len(logs))
	}
}

func TestTestLogger_PrintFunctions(t *testing.T) {
	logger := NewTestLogger(t)

	// Test Println
	logger.Println("println message")
	logs := logger.GetTestLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log after Println, got %d", len(logs))
	}

	// Test Printf
	logger.Printf("printf message %d", 42)
	logs = logger.GetTestLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs after Printf, got %d", len(logs))
	}
	if !strings.Contains(logs[1], "42") {
		t.Errorf("Expected log to contain '42', got: %s", logs[1])
	}
}

func TestTestLogger_ClearTestLogs(t *testing.T) {
	logger := NewTestLogger(t)

	logger.Info("message 1")
	logger.Info("message 2")

	logs := logger.GetTestLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs before clear, got %d", len(logs))
	}

	logger.ClearTestLogs()

	logs = logger.GetTestLogs()
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs after clear, got %d", len(logs))
	}
}

func TestTestLogger_ContextLogging(t *testing.T) {
	logger := NewTestLogger(t)
	ctx := context.Background()

	// Test DebugContext
	logger.DebugContext(ctx, "debug context message")
	logs := logger.GetTestLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log after DebugContext, got %d", len(logs))
	}

	// Test InfoContext
	logger.InfoContext(ctx, "info context message")
	logs = logger.GetTestLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs after InfoContext, got %d", len(logs))
	}

	// Test WarnContext
	logger.WarnContext(ctx, "warn context message")
	logs = logger.GetTestLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs after WarnContext, got %d", len(logs))
	}

	// Test ErrorContext
	logger.ErrorContext(ctx, "error context message")
	logs = logger.GetTestLogs()
	if len(logs) != 4 {
		t.Errorf("Expected 4 logs after ErrorContext, got %d", len(logs))
	}

	// Test FatalContext
	logger.FatalContext(ctx, "fatal context message")
	logs = logger.GetTestLogs()
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs after FatalContext, got %d", len(logs))
	}
}

func TestTestLogger_WithContext(t *testing.T) {
	logger := NewTestLogger(t)
	ctx := WithCorrelationID(context.Background())

	contextLogger := logger.WithContext(ctx)
	if contextLogger == nil {
		t.Error("Expected non-nil context logger")
	}

	// The returned logger should be the same instance since TestLogger
	// implements context handling directly
	if contextLogger != logger {
		t.Error("Expected WithContext to return the same logger instance")
	}
}

func TestTestLogger_ConcurrentAccess(t *testing.T) {
	logger := NewTestLogger(t)

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

	logs := logger.GetTestLogs()
	if len(logs) != 20 {
		t.Errorf("Expected 20 logs from concurrent access, got %d", len(logs))
	}
}

func TestTestLogger_EmptyState(t *testing.T) {
	logger := NewTestLogger(t)

	// Test empty state operations
	logs := logger.GetTestLogs()
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs initially, got %d", len(logs))
	}

	// Test clear on empty logger
	logger.ClearTestLogs()
	logs = logger.GetTestLogs()
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs after clearing empty logger, got %d", len(logs))
	}
}

func TestTestLogger_MessageFormatting(t *testing.T) {
	logger := NewTestLogger(t)

	// Test formatted messages
	logger.Info("formatted message with %s and %d", "string", 123)
	logs := logger.GetTestLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if !strings.Contains(logs[0], "string") || !strings.Contains(logs[0], "123") {
		t.Errorf("Expected log to contain formatted values, got: %s", logs[0])
	}

	// Test multiple formatted messages
	logger.Debug("debug %v", []int{1, 2, 3})
	logger.Error("error %t", true)

	logs = logger.GetTestLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestTestLogger_LogLevels(t *testing.T) {
	logger := NewTestLogger(t)

	// Test all log levels are captured
	logger.Debug("debug level")
	logger.Info("info level")
	logger.Warn("warn level")
	logger.Error("error level")
	logger.Fatal("fatal level")

	logs := logger.GetTestLogs()
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs for all levels, got %d", len(logs))
	}

	// Verify each level is captured
	logText := strings.Join(logs, "\n")
	levels := []string{"debug level", "info level", "warn level", "error level", "fatal level"}
	for _, level := range levels {
		if !strings.Contains(logText, level) {
			t.Errorf("Expected logs to contain '%s', got: %s", level, logText)
		}
	}
}

func TestTestLogger_ContextMessageFormatting(t *testing.T) {
	logger := NewTestLogger(t)
	ctx := WithCorrelationID(context.Background())

	// Test context-based formatted messages
	logger.InfoContext(ctx, "context message with %s", "formatting")
	logs := logger.GetTestLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if !strings.Contains(logs[0], "formatting") {
		t.Errorf("Expected log to contain 'formatting', got: %s", logs[0])
	}

	// Test context messages with correlation ID
	logger.ErrorContext(ctx, "error with correlation")
	logs = logger.GetTestLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
}
