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
