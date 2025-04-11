// internal/logutil/logutil_test.go
package logutil

import (
	"bytes"
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
