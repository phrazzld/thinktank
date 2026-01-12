package logutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestShowFailedModels verifies that the ShowFailedModels method displays
// failed models with proper formatting, colors, and alignment
func TestShowFailedModels(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test failed models with various failure reasons
	failedModels := []FailedModel{
		{Name: "gpt-4o", Reason: "rate limited"},
		{Name: "claude-3-5-sonnet", Reason: "invalid API key"},
		{Name: "gemini-3-flash", Reason: "quota exceeded"},
	}

	// Create console writer with test options
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false }, // Non-interactive for cleaner output
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	// Call ShowFailedModels
	writer.ShowFailedModels(failedModels)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify the output contains expected elements
	expectedElements := []string{
		"FAILED MODELS",     // Header
		"gpt-4o",            // Model name 1
		"rate limited",      // Failure reason 1
		"claude-3-5-sonnet", // Model name 2
		"invalid API key",   // Failure reason 2
		"gemini-3-flash",    // Model name 3
		"quota exceeded",    // Failure reason 3
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain %q, but it was missing.\nActual output:\n%s", element, output)
		}
	}

	// Verify header format (should be uppercase)
	if !strings.Contains(output, "FAILED MODELS") {
		t.Errorf("Expected uppercase 'FAILED MODELS' header, got: %s", output)
	}

	// Verify each model and reason appears in the same line (proper alignment)
	lines := strings.Split(output, "\n")
	var modelLines []string
	for _, line := range lines {
		if strings.Contains(line, "gpt-4o") || strings.Contains(line, "claude-3-5-sonnet") || strings.Contains(line, "gemini-3-flash") {
			modelLines = append(modelLines, line)
		}
	}

	if len(modelLines) != 3 {
		t.Errorf("Expected 3 model lines, got %d: %v", len(modelLines), modelLines)
	}

	// Verify that each model line contains both the model name and reason
	modelReasonPairs := map[string]string{
		"gpt-4o":            "rate limited",
		"claude-3-5-sonnet": "invalid API key",
		"gemini-3-flash":    "quota exceeded",
	}

	for modelName, expectedReason := range modelReasonPairs {
		found := false
		for _, line := range modelLines {
			if strings.Contains(line, modelName) && strings.Contains(line, expectedReason) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find line with both %q and %q, but didn't find it in: %v", modelName, expectedReason, modelLines)
		}
	}

	fmt.Printf("Test output:\n%s", output)
}

// TestShowFailedModelsEmpty verifies that empty failed list doesn't display anything
func TestShowFailedModelsEmpty(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create console writer
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	// Call ShowFailedModels with empty slice
	writer.ShowFailedModels([]FailedModel{})

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should be empty (no output for empty failed models list)
	if output != "" {
		t.Errorf("Expected no output for empty failed models list, got: %q", output)
	}
}

// TestShowFailedModelsQuietMode verifies that quiet mode suppresses output
func TestShowFailedModelsQuietMode(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test failed models
	failedModels := []FailedModel{
		{Name: "test-model", Reason: "test failure"},
	}

	// Create console writer and enable quiet mode
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})
	writer.SetQuiet(true)

	// Call ShowFailedModels
	writer.ShowFailedModels(failedModels)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should be empty (quiet mode suppresses output)
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %q", output)
	}
}

// TestShowFailedModelsSingleFailure verifies proper formatting for single failure
func TestShowFailedModelsSingleFailure(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create single failed model
	failedModels := []FailedModel{
		{Name: "problematic-model", Reason: "connection timeout"},
	}

	// Create console writer
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	// Call ShowFailedModels
	writer.ShowFailedModels(failedModels)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify essential elements are present
	if !strings.Contains(output, "FAILED MODELS") {
		t.Errorf("Expected 'FAILED MODELS' header, got: %s", output)
	}
	if !strings.Contains(output, "problematic-model") {
		t.Errorf("Expected model name 'problematic-model', got: %s", output)
	}
	if !strings.Contains(output, "connection timeout") {
		t.Errorf("Expected failure reason 'connection timeout', got: %s", output)
	}
}
