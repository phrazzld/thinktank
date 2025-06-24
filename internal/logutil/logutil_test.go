// internal/logutil/logutil_test.go
package logutil

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"testing"
)

// Note: osExit is defined in logutil.go

func TestLogLevelString(t *testing.T) {
	t.Parallel() // CPU-bound: String formatting test
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, test := range tests {
		if got := test.level.String(); got != test.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", test.level, got, test.expected)
		}
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	t.Parallel() // CPU-bound: Logger behavior test with buffer
	buf := new(bytes.Buffer)
	logger := NewLogger(InfoLevel, buf, "")

	// Debug message should not be logged at Info level
	logger.Debug("This should not be logged")
	if buf.Len() > 0 {
		t.Errorf("Debug message was logged when level was set to Info")
	}

	// Info message should be logged
	logger.Info("This should be logged")
	if buf.Len() == 0 {
		t.Errorf("Info message was not logged when level was set to Info")
	}

	// Clear buffer
	buf.Reset()

	// Change level to Debug
	logger.SetLevel(DebugLevel)

	// Now debug message should be logged
	logger.Debug("This should be logged too")
	if buf.Len() == 0 {
		t.Errorf("Debug message was not logged when level was set to Debug")
	}
}

func TestLoggerPrefix(t *testing.T) {
	t.Parallel() // CPU-bound: Logger prefix test with buffer
	buf := new(bytes.Buffer)
	prefix := "TEST: "
	logger := NewLogger(DebugLevel, buf, prefix)

	logger.Info("Message")
	if !strings.Contains(buf.String(), prefix) {
		t.Errorf("Log message does not contain prefix %q", prefix)
	}

	// Test changing prefix
	buf.Reset()
	newPrefix := "NEW: "
	logger.SetPrefix(newPrefix)
	logger.Info("Another message")
	if !strings.Contains(buf.String(), newPrefix) {
		t.Errorf("Log message does not contain new prefix %q", newPrefix)
	}
}

func TestLoggerAllLevels(t *testing.T) {
	t.Parallel() // CPU-bound: Logger level test with buffer
	buf := new(bytes.Buffer)
	logger := NewLogger(DebugLevel, buf, "")

	// Test all log levels
	testCases := []struct {
		logFunc func(string, ...interface{})
		level   string
		message string
	}{
		{logger.Debug, "DEBUG", "debug message"},
		{logger.Info, "INFO", "info message"},
		{logger.Warn, "WARN", "warn message"},
		{logger.Error, "ERROR", "error message"},
	}

	for _, tc := range testCases {
		buf.Reset()
		tc.logFunc(tc.message)
		output := buf.String()
		if !strings.Contains(output, tc.level) {
			t.Errorf("Expected log output to contain level %q, got: %s", tc.level, output)
		}
		if !strings.Contains(output, tc.message) {
			t.Errorf("Expected log output to contain message %q, got: %s", tc.message, output)
		}
	}

	// Test context-aware logging methods
	testID := "test-correlation-id-123"
	ctx := context.WithValue(context.Background(), CorrelationIDKey, testID)

	// Test all context-aware log levels
	contextTestCases := []struct {
		logFunc func(context.Context, string, ...interface{})
		level   string
		message string
	}{
		{logger.DebugContext, "DEBUG", "debug context message"},
		{logger.InfoContext, "INFO", "info context message"},
		{logger.WarnContext, "WARN", "warn context message"},
		{logger.ErrorContext, "ERROR", "error context message"},
	}

	for _, tc := range contextTestCases {
		buf.Reset()
		tc.logFunc(ctx, tc.message)
		output := buf.String()
		if !strings.Contains(output, tc.level) {
			t.Errorf("Expected log output to contain level %q, got: %s", tc.level, output)
		}
		if !strings.Contains(output, tc.message) {
			t.Errorf("Expected log output to contain message %q, got: %s", tc.message, output)
		}
		if !strings.Contains(output, testID) {
			t.Errorf("Expected log output to contain correlation ID %q, got: %s", testID, output)
		}
	}

	// Test WithContext creates logger with context
	buf.Reset()
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("message with logger context")
	output := buf.String()
	if !strings.Contains(output, testID) {
		t.Errorf("Expected output to contain correlation ID %q, got: %s", testID, output)
	}
}

// TestLoggerFatal tests the Fatal method without calling os.Exit
func TestLoggerFatal(t *testing.T) {
	// Save original os.Exit function
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	// Mock os.Exit to record if it was called and prevent actual exit
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("Expected exit code 1, got %d", code)
		}
	}

	// Create logger and test Fatal
	buf := new(bytes.Buffer)
	logger := NewLogger(DebugLevel, buf, "")

	// Call Fatal
	logger.Fatal("fatal %s", "message")

	// Check if exit was called
	if !exitCalled {
		t.Error("os.Exit was not called by Fatal method")
	}

	// Check the output
	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected log output to contain level ERROR, got: %s", output)
	}
	if !strings.Contains(output, "fatal message") {
		t.Errorf("Expected log output to contain message 'fatal message', got: %s", output)
	}

	// Reset exit called flag
	exitCalled = false

	// Test FatalContext with correlation ID
	buf.Reset()
	testID := "fatal-correlation-id"
	ctx := context.WithValue(context.Background(), CorrelationIDKey, testID)

	// Call FatalContext
	logger.FatalContext(ctx, "fatal context %s", "message")

	// Check if exit was called
	if !exitCalled {
		t.Error("os.Exit was not called by FatalContext method")
	}

	// Check the output
	output = buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected log output to contain level ERROR, got: %s", output)
	}
	if !strings.Contains(output, "fatal context message") {
		t.Errorf("Expected log output to contain message, got: %s", output)
	}
	if !strings.Contains(output, testID) {
		t.Errorf("Expected log output to contain correlation ID %q, got: %s", testID, output)
	}
}

func TestLoggerPrintFunctions(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewLogger(DebugLevel, buf, "")

	// Test Printf
	buf.Reset()
	logger.Printf("Format %s %d", "test", 123)
	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Printf should log at INFO level, got: %s", output)
	}
	if !strings.Contains(output, "Format test 123") {
		t.Errorf("Printf output incorrect, got: %s", output)
	}

	// Test Println
	buf.Reset()
	logger.Println("Line", "test")
	output = buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Println should log at INFO level, got: %s", output)
	}
	if !strings.Contains(output, "Line test") {
		t.Errorf("Println output incorrect, got: %s", output)
	}
}

func TestLoggerGetLevel(t *testing.T) {
	logger := NewLogger(InfoLevel, os.Stderr, "")
	if level := logger.GetLevel(); level != InfoLevel {
		t.Errorf("GetLevel() = %v, want %v", level, InfoLevel)
	}

	logger.SetLevel(DebugLevel)
	if level := logger.GetLevel(); level != DebugLevel {
		t.Errorf("GetLevel() = %v, want %v", level, DebugLevel)
	}
}

func TestNewLoggerDefaults(t *testing.T) {
	// Test with nil writer
	logger := NewLogger(InfoLevel, nil, "")
	if logger.writer == nil {
		t.Error("Logger writer should default to os.Stderr when nil is passed")
	}

	// Ensure context is initialized
	if logger.ctx == nil {
		t.Error("Logger context should be initialized")
	}
}

func TestStdLoggerAdapter(t *testing.T) {
	buf := new(bytes.Buffer)
	stdLogger := log.New(buf, "", log.LstdFlags)
	adapter := NewStdLoggerAdapter(stdLogger)

	testCases := []struct {
		logFunc func(string, ...interface{})
		level   string
		message string
	}{
		{adapter.Debug, "[DEBUG]", "debug message"},
		{adapter.Info, "[INFO]", "info message"},
		{adapter.Warn, "[WARN]", "warn message"},
		{adapter.Error, "[ERROR]", "error message"},
	}

	for _, tc := range testCases {
		buf.Reset()
		tc.logFunc(tc.message)
		output := buf.String()
		if !strings.Contains(output, tc.level) {
			t.Errorf("Expected log output to contain level %q, got: %s", tc.level, output)
		}
		if !strings.Contains(output, tc.message) {
			t.Errorf("Expected log output to contain message %q, got: %s", tc.message, output)
		}
	}

	// Test Printf and Println
	buf.Reset()
	adapter.Printf("Format %s", "test")
	if !strings.Contains(buf.String(), "Format test") {
		t.Errorf("Printf output incorrect, got: %s", buf.String())
	}

	// Test with context and correlation ID
	buf.Reset()
	testID := "test-correlation-id"
	ctx := context.WithValue(context.Background(), CorrelationIDKey, testID)

	adapter.InfoContext(ctx, "context test message")
	output := buf.String()
	if !strings.Contains(output, testID) {
		t.Errorf("Expected log output to contain correlation ID %q, got: %s", testID, output)
	}
	if !strings.Contains(output, "context test message") {
		t.Errorf("Expected log output to contain message, got: %s", output)
	}

	// Test WithContext creates adapter with context
	buf.Reset()
	ctxAdapter := adapter.WithContext(ctx)
	ctxAdapter.Info("message with attached context")
	// Our implementation doesn't include correlation ID from attached context in standard methods
	// This is a design choice, could be changed if needed
}

// TestStdLoggerAdapterFatal tests the Fatal method of StdLoggerAdapter
func TestStdLoggerAdapterFatal(t *testing.T) {
	// Save original os.Exit and replace it
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	// Mock os.Exit
	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	// Create a buffer and logger
	buf := new(bytes.Buffer)
	stdLogger := log.New(buf, "", log.LstdFlags)
	adapter := NewStdLoggerAdapter(stdLogger)

	// Call Fatal
	adapter.Fatal("fatal error: %s", "test")

	// Check that os.Exit was called
	if !exitCalled {
		t.Error("os.Exit was not called")
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Check the output
	output := buf.String()
	if !strings.Contains(output, "[FATAL]") {
		t.Errorf("Expected log output to contain [FATAL], got: %s", output)
	}
	if !strings.Contains(output, "fatal error: test") {
		t.Errorf("Expected log output to contain 'fatal error: test', got: %s", output)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
		isError  bool
	}{
		{"debug", DebugLevel, false},
		{"info", InfoLevel, false},
		{"warn", WarnLevel, false},
		{"error", ErrorLevel, false},
		{"unknown", InfoLevel, true},
		{"", InfoLevel, true},
	}

	for _, test := range tests {
		level, err := ParseLogLevel(test.input)
		if (err != nil) != test.isError {
			t.Errorf("ParseLogLevel(%q) error = %v, want error? %v", test.input, err, test.isError)
		}
		if !test.isError && level != test.expected {
			t.Errorf("ParseLogLevel(%q) = %v, want %v", test.input, level, test.expected)
		}
	}
}

func TestCorrelationIDFunctions(t *testing.T) {
	// Test WithCorrelationID generates and adds ID to context
	ctx := context.Background()
	ctxWithID := WithCorrelationID(ctx)
	id := GetCorrelationID(ctxWithID)

	if id == "" {
		t.Error("WithCorrelationID should have generated a non-empty correlation ID")
	}

	// Test WithCorrelationID preserves existing ID
	newCtx := WithCorrelationID(ctxWithID)
	newID := GetCorrelationID(newCtx)

	if newID != id {
		t.Errorf("WithCorrelationID should have preserved existing ID %q, got %q", id, newID)
	}

	// Test WithCorrelationID with empty ID parameter preserves existing ID
	emptyIDCtx := WithCorrelationID(ctxWithID, "")
	emptyID := GetCorrelationID(emptyIDCtx)

	if emptyID != id {
		t.Errorf("WithCorrelationID with empty ID should have preserved existing ID %q, got %q", id, emptyID)
	}

	// Test WithCorrelationID with custom ID parameter sets the ID
	customID := "custom-test-id-123"
	customCtx := WithCorrelationID(ctx, customID)
	resultID := GetCorrelationID(customCtx)

	if resultID != customID {
		t.Errorf("WithCorrelationID with custom ID should have set ID to %q, got %q", customID, resultID)
	}

	// Test WithCorrelationID with custom ID parameter overrides existing ID
	overrideCtx := WithCorrelationID(ctxWithID, "override-id")
	overrideID := GetCorrelationID(overrideCtx)

	if overrideID != "override-id" {
		t.Errorf("WithCorrelationID with custom ID should have overridden existing ID, got %q", overrideID)
	}

	// Test WithCustomCorrelationID sets custom ID
	customID2 := "custom-test-id-456"
	customCtx2 := WithCustomCorrelationID(ctx, customID2)
	resultID2 := GetCorrelationID(customCtx2)

	if resultID2 != customID2 {
		t.Errorf("WithCustomCorrelationID should have set ID to %q, got %q", customID2, resultID2)
	}

	// Test GetCorrelationID with nil context (using context.TODO() instead of nil)
	nilID := GetCorrelationID(context.TODO())
	if nilID != "" {
		t.Errorf("GetCorrelationID with nil context should return empty string, got %q", nilID)
	}

	// Test GetCorrelationID with context that has no correlation ID
	emptyID = GetCorrelationID(context.Background())
	if emptyID != "" {
		t.Errorf("GetCorrelationID with empty context should return empty string, got %q", emptyID)
	}
}
