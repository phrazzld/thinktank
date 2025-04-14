// internal/logutil/logutil_test.go
package logutil

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogLevelString(t *testing.T) {
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
