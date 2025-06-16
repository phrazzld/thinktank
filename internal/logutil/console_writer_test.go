package logutil

import (
	"testing"
	"time"
)

// mockConsoleWriter is a test implementation of ConsoleWriter
// that verifies the interface can be properly implemented
type mockConsoleWriter struct {
	quiet       bool
	noProgress  bool
	interactive bool
	modelCount  int
	calls       []string
}

// Ensure mockConsoleWriter implements ConsoleWriter interface
var _ ConsoleWriter = (*mockConsoleWriter)(nil)

func newMockConsoleWriter() *mockConsoleWriter {
	return &mockConsoleWriter{
		interactive: true,
		calls:       make([]string, 0),
	}
}

func (m *mockConsoleWriter) StartProcessing(modelCount int) {
	m.modelCount = modelCount
	m.calls = append(m.calls, "StartProcessing")
}

func (m *mockConsoleWriter) ModelQueued(modelName string, index int) {
	m.calls = append(m.calls, "ModelQueued")
}

func (m *mockConsoleWriter) ModelStarted(modelName string, index int) {
	m.calls = append(m.calls, "ModelStarted")
}

func (m *mockConsoleWriter) ModelCompleted(modelName string, index int, duration time.Duration, err error) {
	m.calls = append(m.calls, "ModelCompleted")
}

func (m *mockConsoleWriter) ModelRateLimited(modelName string, index int, delay time.Duration) {
	m.calls = append(m.calls, "ModelRateLimited")
}

func (m *mockConsoleWriter) SynthesisStarted() {
	m.calls = append(m.calls, "SynthesisStarted")
}

func (m *mockConsoleWriter) SynthesisCompleted(outputPath string) {
	m.calls = append(m.calls, "SynthesisCompleted")
}

func (m *mockConsoleWriter) SetQuiet(quiet bool) {
	m.quiet = quiet
}

func (m *mockConsoleWriter) SetNoProgress(noProgress bool) {
	m.noProgress = noProgress
}

func (m *mockConsoleWriter) IsInteractive() bool {
	return m.interactive
}

// TestConsoleWriterInterface verifies that the ConsoleWriter interface
// can be implemented and used correctly
func TestConsoleWriterInterface(t *testing.T) {
	writer := newMockConsoleWriter()

	// Test control methods
	if !writer.IsInteractive() {
		t.Error("Expected interactive mode to be true by default")
	}

	writer.SetQuiet(true)
	if !writer.quiet {
		t.Error("SetQuiet(true) should set quiet mode")
	}

	writer.SetNoProgress(true)
	if !writer.noProgress {
		t.Error("SetNoProgress(true) should set no-progress mode")
	}

	// Test progress reporting methods
	writer.StartProcessing(3)
	if writer.modelCount != 3 {
		t.Errorf("Expected model count 3, got %d", writer.modelCount)
	}

	// Test all methods can be called without panic
	writer.ModelQueued("test-model", 1)
	writer.ModelStarted("test-model", 1)
	writer.ModelCompleted("test-model", 1, time.Second, nil)
	writer.ModelRateLimited("test-model", 2, time.Second*2)
	writer.SynthesisStarted()
	writer.SynthesisCompleted("/path/to/output")

	// Verify all methods were called
	expectedCalls := []string{
		"StartProcessing",
		"ModelQueued",
		"ModelStarted",
		"ModelCompleted",
		"ModelRateLimited",
		"SynthesisStarted",
		"SynthesisCompleted",
	}

	if len(writer.calls) != len(expectedCalls) {
		t.Errorf("Expected %d method calls, got %d", len(expectedCalls), len(writer.calls))
	}

	for i, expected := range expectedCalls {
		if i >= len(writer.calls) || writer.calls[i] != expected {
			t.Errorf("Expected call %d to be %s, got %s", i, expected, writer.calls[i])
		}
	}
}

// TestConsoleWriterWithError verifies error handling in ModelCompleted
func TestConsoleWriterWithError(t *testing.T) {
	writer := newMockConsoleWriter()

	// Test with error
	testErr := time.ParseError{Layout: "test", Value: "test", Message: "test error"}
	writer.ModelCompleted("test-model", 1, time.Second, &testErr)

	if len(writer.calls) != 1 || writer.calls[0] != "ModelCompleted" {
		t.Error("ModelCompleted should handle errors without panic")
	}
}
