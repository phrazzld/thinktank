package logutil

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestSlogLogger_Basic(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelDebug)

	// Test basic logging methods
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	// Verify each log level appears in the output
	if !strings.Contains(output, `"level":"DEBUG"`) {
		t.Error("Debug log level not found in output")
	}
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("Info log level not found in output")
	}
	if !strings.Contains(output, `"level":"WARN"`) {
		t.Error("Warn log level not found in output")
	}
	if !strings.Contains(output, `"level":"ERROR"`) {
		t.Error("Error log level not found in output")
	}

	// Verify messages appear in the output
	if !strings.Contains(output, `"msg":"debug message"`) {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, `"msg":"info message"`) {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, `"msg":"warn message"`) {
		t.Error("Warn message not found in output")
	}
	if !strings.Contains(output, `"msg":"error message"`) {
		t.Error("Error message not found in output")
	}
}

func TestSlogLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelDebug)

	// Create context with correlation ID
	ctx := WithCustomCorrelationID(context.Background(), "test-correlation-id")

	// Create a new logger with the context
	contextLogger := logger.WithContext(ctx)

	// Log with context methods
	contextLogger.InfoContext(ctx, "message with context")

	output := buf.String()

	// Verify correlation ID appears in the output
	if !strings.Contains(output, `"correlation_id":"test-correlation-id"`) {
		t.Error("Correlation ID not found in output")
	}
}

func TestSlogLogger_ContextAwareMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelDebug)

	// Create context with correlation ID
	ctx := WithCustomCorrelationID(context.Background(), "test-correlation-id")

	// Test context-aware logging methods
	logger.DebugContext(ctx, "debug context message")
	logger.InfoContext(ctx, "info context message")
	logger.WarnContext(ctx, "warn context message")
	logger.ErrorContext(ctx, "error context message")

	output := buf.String()

	// Verify correlation ID appears in all messages
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 4 log lines, got %d", len(lines))
	}

	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Fatalf("Failed to parse JSON log entry %d: %v", i, err)
		}

		correlationID, ok := logEntry["correlation_id"]
		if !ok {
			t.Errorf("Log entry %d missing correlation_id field", i)
		} else if id, ok := correlationID.(string); !ok || id != "test-correlation-id" {
			t.Errorf("Log entry %d has incorrect correlation_id: %v", i, correlationID)
		}
	}
}

func TestSlogLogger_FormatWithArgs(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelDebug)

	// Test formatting with arguments
	logger.Info("formatted %s with %d arguments", "message", 2)

	output := buf.String()

	// Verify formatted message appears in the output
	if !strings.Contains(output, `"msg":"formatted message with 2 arguments"`) {
		t.Error("Formatted message not found in output")
	}
}

func TestSlogLogger_StructuredLogging(t *testing.T) {
	// Use a direct slog.Logger for testing structured logging
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	slogLogger := slog.New(handler)

	// Log with structured fields
	slogLogger.Info("structured message",
		"user_id", "user123",
		"items_count", 42,
		"verified", true,
		"metadata.source", "api",
		"metadata.version", "v1.2.3",
	)

	output := buf.String()

	// Verify fields are present
	if !strings.Contains(output, `"user_id":"user123"`) {
		t.Error("user_id field not found in output")
	}
	if !strings.Contains(output, `"items_count":42`) {
		t.Error("items_count field not found in output")
	}
	if !strings.Contains(output, `"verified":true`) {
		t.Error("verified field not found in output")
	}
	if !strings.Contains(output, `"metadata.source":"api"`) {
		t.Error("metadata.source field not found in output")
	}
}

func TestSlogLogger_CorrelationIDInContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelDebug)

	// Create context with correlation ID
	ctx := WithCustomCorrelationID(context.Background(), "test-correlation-id")

	// Log a simple message with context
	logger.InfoContext(ctx, "test message with correlation id")

	output := buf.String()

	// Verify correlation ID appears in the output
	if !strings.Contains(output, `"correlation_id":"test-correlation-id"`) {
		t.Error("Correlation ID not found in output")
	}

	// Verify message appears in the output
	if !strings.Contains(output, `"msg":"test message with correlation id"`) {
		t.Error("Message not found in output")
	}
}

func TestConvertLogLevelToSlog(t *testing.T) {
	testCases := []struct {
		level       LogLevel
		expectLevel slog.Level
	}{
		{DebugLevel, slog.LevelDebug},
		{InfoLevel, slog.LevelInfo},
		{WarnLevel, slog.LevelWarn},
		{ErrorLevel, slog.LevelError},
		{LogLevel(99), slog.LevelInfo}, // Unknown level should default to Info
	}

	for _, tc := range testCases {
		t.Run(tc.level.String(), func(t *testing.T) {
			result := ConvertLogLevelToSlog(tc.level)
			if result != tc.expectLevel {
				t.Errorf("Expected level %v, got %v", tc.expectLevel, result)
			}
		})
	}
}

func TestNewSlogLoggerFromLogLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewSlogLoggerFromLogLevel(&buf, InfoLevel)

	// Debug messages should be filtered out
	logger.Debug("debug message") // Should not appear
	logger.Info("info message")   // Should appear

	output := buf.String()

	// Debug message should not appear in output
	if strings.Contains(output, `"msg":"debug message"`) {
		t.Error("Debug message should have been filtered out")
	}

	// Info message should appear in output
	if !strings.Contains(output, `"msg":"info message"`) {
		t.Error("Info message not found in output")
	}
}

func TestSlogLoggerWithStreamSeparation(t *testing.T) {
	// Create buffers for info and error logs
	var infoBuf, errorBuf bytes.Buffer
	logger := NewSlogLoggerWithStreamSeparation(&infoBuf, &errorBuf, slog.LevelDebug)

	// Log messages at different levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	infoOutput := infoBuf.String()
	errorOutput := errorBuf.String()

	// Info buffer should contain DEBUG and INFO messages
	if !strings.Contains(infoOutput, `"level":"DEBUG"`) {
		t.Error("Debug level not found in info output")
	}
	if !strings.Contains(infoOutput, `"level":"INFO"`) {
		t.Error("Info level not found in info output")
	}
	if !strings.Contains(infoOutput, `"msg":"debug message"`) {
		t.Error("Debug message not found in info output")
	}
	if !strings.Contains(infoOutput, `"msg":"info message"`) {
		t.Error("Info message not found in info output")
	}

	// Info buffer should NOT contain WARN and ERROR messages
	if strings.Contains(infoOutput, `"level":"WARN"`) {
		t.Error("Warn level found in info output but should be in error output")
	}
	if strings.Contains(infoOutput, `"level":"ERROR"`) {
		t.Error("Error level found in info output but should be in error output")
	}

	// Error buffer should contain WARN and ERROR messages
	if !strings.Contains(errorOutput, `"level":"WARN"`) {
		t.Error("Warn level not found in error output")
	}
	if !strings.Contains(errorOutput, `"level":"ERROR"`) {
		t.Error("Error level not found in error output")
	}
	if !strings.Contains(errorOutput, `"msg":"warn message"`) {
		t.Error("Warn message not found in error output")
	}
	if !strings.Contains(errorOutput, `"msg":"error message"`) {
		t.Error("Error message not found in error output")
	}

	// Error buffer should NOT contain DEBUG and INFO messages
	if strings.Contains(errorOutput, `"level":"DEBUG"`) {
		t.Error("Debug level found in error output but should be in info output")
	}
	if strings.Contains(errorOutput, `"level":"INFO"`) {
		t.Error("Info level found in error output but should be in info output")
	}
}

func TestEnableStreamSeparation_Basic(t *testing.T) {
	// Create a standard logger without stream separation
	var buf bytes.Buffer
	logger := NewSlogLogger(&buf, slog.LevelDebug)

	// Convert to stream-separated logger
	streamLogger := EnableStreamSeparation(logger)

	// Verify it was converted
	if !streamLogger.streamSplit {
		t.Error("Logger should have streamSplit=true after EnableStreamSeparation")
	}

	// Verify a logger that's already stream-separated stays as is
	sameLogger := EnableStreamSeparation(streamLogger)
	if sameLogger != streamLogger {
		t.Error("EnableStreamSeparation should return same logger if already stream-separated")
	}
}
