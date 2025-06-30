package logutil

import (
	"testing"
	"time"
)

func TestConsoleWriter_StatusTracking(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})

	modelNames := []string{"model1", "model2"}

	// Test StartStatusTracking
	output := captureOutput(func() {
		cw.StartStatusTracking(modelNames)
	})

	if output == "" {
		t.Error("Expected output from StartStatusTracking")
	}

	// Test UpdateModelStatus
	captureOutput(func() {
		cw.UpdateModelStatus("model1", StatusProcessing, 0, "")
		cw.UpdateModelStatus("model1", StatusCompleted, 100*time.Millisecond, "")
	})

	// Test UpdateModelRateLimited
	captureOutput(func() {
		cw.UpdateModelRateLimited("model2", 2*time.Second)
	})

	// Test RefreshStatusDisplay
	captureOutput(func() {
		cw.RefreshStatusDisplay()
	})

	// Test FinishStatusTracking
	captureOutput(func() {
		cw.FinishStatusTracking()
	})
}

func TestConsoleWriter_StatusTracking_QuietMode(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})
	cw.SetQuiet(true)

	modelNames := []string{"model1"}

	// Should produce no output in quiet mode
	output := captureOutput(func() {
		cw.StartStatusTracking(modelNames)
		cw.UpdateModelStatus("model1", StatusCompleted, 100*time.Millisecond, "")
		cw.FinishStatusTracking()
	})

	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %q", output)
	}
}

func TestConsoleWriter_StatusTracking_NoProgressMode(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})
	cw.SetNoProgress(true)

	modelNames := []string{"model1"}

	// Should produce no output in no-progress mode
	output := captureOutput(func() {
		cw.StartStatusTracking(modelNames)
		cw.UpdateModelStatus("model1", StatusCompleted, 100*time.Millisecond, "")
		cw.FinishStatusTracking()
	})

	if output != "" {
		t.Errorf("Expected no output in no-progress mode, got: %q", output)
	}
}
