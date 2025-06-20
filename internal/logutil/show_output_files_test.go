package logutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestShowOutputFiles verifies that the ShowOutputFiles method displays files
// with proper formatting, human-readable sizes, and correct structure
func TestShowOutputFiles(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test files with various sizes
	testFiles := []OutputFile{
		{Name: "model-output-1.md", Size: 1024},      // Should show as "1.0K"
		{Name: "model-output-2.md", Size: 4567},      // Should show as "4.5K"
		{Name: "synthesis.md", Size: 1536000},        // Should show as "1.5M"
	}

	// Create console writer with test options
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // Non-interactive for cleaner output
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc: func(key string) string { return "" },
	})

	// Call ShowOutputFiles
	writer.ShowOutputFiles(testFiles)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = old
	
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Verify the output contains expected elements
	expectedElements := []string{
		"OUTPUT FILES",            // Header
		"model-output-1.md",      // File name 1
		"1.0K",                   // Human readable size 1
		"model-output-2.md",      // File name 2
		"4.5K",                   // Human readable size 2
		"synthesis.md",           // File name 3
		"1.5M",                   // Human readable size 3
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain %q, but it was missing.\nActual output:\n%s", element, output)
		}
	}

	// Verify header format (should be uppercase)
	if !strings.Contains(output, "OUTPUT FILES") {
		t.Errorf("Expected uppercase 'OUTPUT FILES' header, got: %s", output)
	}

	fmt.Printf("Test output:\n%s", output)
}

// TestShowOutputFilesEmpty verifies that empty file list doesn't display anything
func TestShowOutputFilesEmpty(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create console writer
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc: func(key string) string { return "" },
	})

	// Call ShowOutputFiles with empty slice
	writer.ShowOutputFiles([]OutputFile{})

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = old
	
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Should be empty (no output for empty file list)
	if output != "" {
		t.Errorf("Expected no output for empty file list, got: %q", output)
	}
}

// TestShowOutputFilesQuietMode verifies that quiet mode suppresses output
func TestShowOutputFilesQuietMode(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test files
	testFiles := []OutputFile{
		{Name: "test.md", Size: 1024},
	}

	// Create console writer and enable quiet mode
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc: func(key string) string { return "" },
	})
	writer.SetQuiet(true)

	// Call ShowOutputFiles
	writer.ShowOutputFiles(testFiles)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = old
	
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Should be empty (quiet mode suppresses output)
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %q", output)
	}
}