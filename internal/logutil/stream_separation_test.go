package logutil

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestSlogLoggerWithStreamSeparation_Basic(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparation(&stdoutBuf, &stderrBuf, slog.LevelDebug)

	// Test basic logging methods
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Verify stdout only contains DEBUG and INFO logs
	stdoutOutput := stdoutBuf.String()
	if !strings.Contains(stdoutOutput, `"level":"DEBUG"`) {
		t.Error("DEBUG log level not found in stdout")
	}
	if !strings.Contains(stdoutOutput, `"level":"INFO"`) {
		t.Error("INFO log level not found in stdout")
	}
	if strings.Contains(stdoutOutput, `"level":"WARN"`) {
		t.Error("WARN log level should not be in stdout")
	}
	if strings.Contains(stdoutOutput, `"level":"ERROR"`) {
		t.Error("ERROR log level should not be in stdout")
	}

	// Verify stderr only contains WARN and ERROR logs
	stderrOutput := stderrBuf.String()
	if strings.Contains(stderrOutput, `"level":"DEBUG"`) {
		t.Error("DEBUG log level should not be in stderr")
	}
	if strings.Contains(stderrOutput, `"level":"INFO"`) {
		t.Error("INFO log level should not be in stderr")
	}
	if !strings.Contains(stderrOutput, `"level":"WARN"`) {
		t.Error("WARN log level not found in stderr")
	}
	if !strings.Contains(stderrOutput, `"level":"ERROR"`) {
		t.Error("ERROR log level not found in stderr")
	}

	// Verify messages appear in the correct output stream
	if !strings.Contains(stdoutOutput, `"msg":"debug message"`) {
		t.Error("Debug message not found in stdout")
	}
	if !strings.Contains(stdoutOutput, `"msg":"info message"`) {
		t.Error("Info message not found in stdout")
	}
	if !strings.Contains(stderrOutput, `"msg":"warn message"`) {
		t.Error("Warn message not found in stderr")
	}
	if !strings.Contains(stderrOutput, `"msg":"error message"`) {
		t.Error("Error message not found in stderr")
	}
}

func TestSlogLoggerWithStreamSeparation_ContextAware(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparation(&stdoutBuf, &stderrBuf, slog.LevelDebug)

	// Create context with correlation ID
	ctx := WithCustomCorrelationID(context.Background(), "test-separation-correlation-id")

	// Test context-aware logging methods
	logger.DebugContext(ctx, "debug context message")
	logger.InfoContext(ctx, "info context message")
	logger.WarnContext(ctx, "warn context message")
	logger.ErrorContext(ctx, "error context message")

	// Verify correlation ID appears in both output streams
	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	if !strings.Contains(stdoutOutput, `"correlation_id":"test-separation-correlation-id"`) {
		t.Error("Correlation ID not found in stdout")
	}
	if !strings.Contains(stderrOutput, `"correlation_id":"test-separation-correlation-id"`) {
		t.Error("Correlation ID not found in stderr")
	}

	// Verify log entries are properly parsed as JSON
	stdoutLines := strings.Split(strings.TrimSpace(stdoutOutput), "\n")
	stderrLines := strings.Split(strings.TrimSpace(stderrOutput), "\n")

	for i, line := range stdoutLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse stdout JSON log entry %d: %v", i, err)
		}
	}

	for i, line := range stderrLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse stderr JSON log entry %d: %v", i, err)
		}
	}
}

func TestSlogLoggerWithStreamSeparation_WithContext(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparation(&stdoutBuf, &stderrBuf, slog.LevelDebug)

	// Create context with correlation ID
	ctx := WithCustomCorrelationID(context.Background(), "test-withcontext-id")

	// Create a new logger with the context
	contextLogger := logger.WithContext(ctx)

	// Log with context methods, but without explicitly passing the context
	// to the logging methods
	contextLogger.Info("info with implicit context")
	contextLogger.Error("error with implicit context")

	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	// Since we're not using the Context methods, correlation ID won't appear
	// But the logs should still go to the correct streams
	if !strings.Contains(stdoutOutput, `"msg":"info with implicit context"`) {
		t.Error("Info message not found in stdout")
	}
	if !strings.Contains(stderrOutput, `"msg":"error with implicit context"`) {
		t.Error("Error message not found in stderr")
	}
}

func TestEnableStreamSeparation(t *testing.T) {
	// Create a logger without stream separation
	var singleBuf bytes.Buffer
	originalLogger := NewSlogLogger(&singleBuf, slog.LevelDebug)

	// Use custom buffers instead of os.Stdout and os.Stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	separatedLogger := NewSlogLoggerWithStreamSeparation(&stdoutBuf, &stderrBuf, slog.LevelDebug)

	// Verify the original logger is not using stream separation
	if originalLogger.streamSplit {
		t.Error("Original logger should not have stream separation enabled")
	}

	// Verify the separated logger has stream separation enabled
	if !separatedLogger.streamSplit {
		t.Error("Separated logger should have stream separation enabled")
	}

	// Test that logs go to the correct streams
	separatedLogger.Info("info after separation")
	separatedLogger.Error("error after separation")

	// Verify logs go to the correct streams
	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	if !strings.Contains(stdoutOutput, `"msg":"info after separation"`) {
		t.Error("Info message not found in stdout after separation")
	}
	if !strings.Contains(stderrOutput, `"msg":"error after separation"`) {
		t.Error("Error message not found in stderr after separation")
	}
}

func TestStreamSeparationFromLogLevel(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparationFromLogLevel(&stdoutBuf, &stderrBuf, InfoLevel)

	// Test that DEBUG logs are filtered out but INFO logs go to stdout
	logger.Debug("debug message") // Should be filtered
	logger.Info("info message")   // Should appear in stdout
	logger.Error("error message") // Should appear in stderr

	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	// Debug should be filtered due to log level
	if strings.Contains(stdoutOutput, `"msg":"debug message"`) {
		t.Error("Debug message should have been filtered out")
	}

	// Info should go to stdout
	if !strings.Contains(stdoutOutput, `"msg":"info message"`) {
		t.Error("Info message not found in stdout")
	}

	// Error should go to stderr
	if !strings.Contains(stderrOutput, `"msg":"error message"`) {
		t.Error("Error message not found in stderr")
	}
}

func TestStreamSeparation_CorrelationIDPropagation(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparation(&stdoutBuf, &stderrBuf, slog.LevelDebug)

	// Create a context with correlation ID
	correlationID := "propagated-correlation-id"
	ctx := WithCorrelationID(context.Background(), correlationID)

	// Create a derived logger with the context
	ctxLogger := logger.WithContext(ctx)

	// Test context-aware methods with both the original context and nil/TODO context
	ctxLogger.InfoContext(ctx, "explicit context log")
	ctxLogger.InfoContext(context.TODO(), "implicit context log")
	ctxLogger.ErrorContext(ctx, "explicit context error log")
	ctxLogger.ErrorContext(context.TODO(), "implicit context error log")

	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	// Parse JSON logs for detailed checking
	stdoutLines := strings.Split(strings.TrimSpace(stdoutOutput), "\n")
	stderrLines := strings.Split(strings.TrimSpace(stderrOutput), "\n")

	// Check info logs in stdout
	for i, line := range stdoutLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Fatalf("Failed to parse JSON log entry %d from stdout: %v", i, err)
		}

		// Check for correlation ID in each log
		id, ok := logEntry["correlation_id"].(string)
		if !ok {
			t.Errorf("Log %d in stdout missing correlation_id", i)
			continue
		}

		if id != correlationID {
			t.Errorf("Log %d in stdout has wrong correlation_id: got %s, want %s", i, id, correlationID)
		}
	}

	// Check error logs in stderr
	for i, line := range stderrLines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Fatalf("Failed to parse JSON log entry %d from stderr: %v", i, err)
		}

		// Check for correlation ID in each log
		id, ok := logEntry["correlation_id"].(string)
		if !ok {
			t.Errorf("Log %d in stderr missing correlation_id", i)
			continue
		}

		if id != correlationID {
			t.Errorf("Log %d in stderr has wrong correlation_id: got %s, want %s", i, id, correlationID)
		}
	}
}

func TestStreamSeparation_ContextFallbacks(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparation(&stdoutBuf, &stderrBuf, slog.LevelDebug)

	// Test different context scenarios
	ctx1 := WithCorrelationID(context.Background(), "id-1")
	ctx2 := WithCorrelationID(context.Background(), "id-2")

	// Create a logger with ctx1
	ctxLogger := logger.WithContext(ctx1)

	// Use explicit ctx2 (should override logger's ctx1)
	ctxLogger.InfoContext(ctx2, "should use id-2")

	// Use empty context (should fall back to logger's ctx1)
	ctxLogger.InfoContext(context.TODO(), "should use id-1")

	// Use empty context (should fall back to logger's ctx1)
	ctxLogger.InfoContext(context.Background(), "should also use id-1")

	stdoutOutput := stdoutBuf.String()
	stdoutLines := strings.Split(strings.TrimSpace(stdoutOutput), "\n")

	// Check each log line
	if len(stdoutLines) != 3 {
		t.Fatalf("Expected 3 log lines, got %d", len(stdoutLines))
	}

	// Parse and check the first log (should have id-2)
	var log1 map[string]interface{}
	if err := json.Unmarshal([]byte(stdoutLines[0]), &log1); err != nil {
		t.Fatalf("Failed to parse first log: %v", err)
	}

	if id, ok := log1["correlation_id"].(string); !ok || id != "id-2" {
		t.Errorf("First log should have correlation_id 'id-2', got: %v", log1["correlation_id"])
	}

	// Parse and check the second log (should have id-1)
	var log2 map[string]interface{}
	if err := json.Unmarshal([]byte(stdoutLines[1]), &log2); err != nil {
		t.Fatalf("Failed to parse second log: %v", err)
	}

	if id, ok := log2["correlation_id"].(string); !ok || id != "id-1" {
		t.Errorf("Second log should have correlation_id 'id-1', got: %v", log2["correlation_id"])
	}

	// Parse and check the third log (should have id-1)
	var log3 map[string]interface{}
	if err := json.Unmarshal([]byte(stdoutLines[2]), &log3); err != nil {
		t.Fatalf("Failed to parse third log: %v", err)
	}

	if id, ok := log3["correlation_id"].(string); !ok || id != "id-1" {
		t.Errorf("Third log should have correlation_id 'id-1', got: %v", log3["correlation_id"])
	}
}
