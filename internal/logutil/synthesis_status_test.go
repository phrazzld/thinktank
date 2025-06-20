package logutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestSynthesisStatusIntegration verifies that synthesis status is properly
// displayed in the summary section with correct formatting and colors
func TestSynthesisStatusIntegration(t *testing.T) {
	tests := []struct {
		name              string
		synthesisStatus   string
		expectedInOutput  []string
		notExpectedOutput []string
	}{
		{
			name:            "Synthesis Completed",
			synthesisStatus: "completed",
			expectedInOutput: []string{
				"SUMMARY",
				"Synthesis:",
				"✓ completed",
			},
			notExpectedOutput: []string{},
		},
		{
			name:            "Synthesis Failed",
			synthesisStatus: "failed",
			expectedInOutput: []string{
				"SUMMARY",
				"Synthesis:",
				"✗ failed",
			},
			notExpectedOutput: []string{},
		},
		{
			name:            "Synthesis Skipped",
			synthesisStatus: "skipped",
			expectedInOutput: []string{
				"SUMMARY",
			},
			notExpectedOutput: []string{
				"Synthesis:",
			},
		},
		{
			name:            "Custom Status",
			synthesisStatus: "in-progress",
			expectedInOutput: []string{
				"SUMMARY",
				"Synthesis:",
				"in-progress",
			},
			notExpectedOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout for testing
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create test summary data
			summaryData := SummaryData{
				ModelsProcessed:  3,
				SuccessfulModels: 2,
				FailedModels:     1,
				SynthesisStatus:  tt.synthesisStatus,
				OutputDirectory:  "/tmp/output",
			}

			// Create console writer with test options
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc:  func() bool { return false }, // Non-interactive for cleaner output
				GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
				GetEnvFunc:      func(key string) string { return "" },
			})

			// Call ShowSummarySection
			writer.ShowSummarySection(summaryData)

			// Restore stdout and read captured output
			_ = w.Close()
			os.Stdout = old

			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify expected elements are present
			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it was missing.\nActual output:\n%s", expected, output)
				}
			}

			// Verify unexpected elements are not present
			for _, notExpected := range tt.notExpectedOutput {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output NOT to contain %q, but it was present.\nActual output:\n%s", notExpected, output)
				}
			}

			fmt.Printf("Test %s output:\n%s\n", tt.name, output)
		})
	}
}

// TestSynthesisStatusWithOtherSections verifies synthesis status works alongside other sections
func TestSynthesisStatusWithOtherSections(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test data for a complete workflow
	summaryData := SummaryData{
		ModelsProcessed:  4,
		SuccessfulModels: 3,
		FailedModels:     1,
		SynthesisStatus:  "completed",
		OutputDirectory:  "/tmp/thinktank-output",
	}

	outputFiles := []OutputFile{
		{Name: "model1.md", Size: 2048},
		{Name: "model2.md", Size: 4096},
		{Name: "synthesis.md", Size: 8192},
	}

	failedModels := []FailedModel{
		{Name: "problematic-model", Reason: "API timeout"},
	}

	// Create console writer
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	// Call all sections to test integration
	writer.ShowSummarySection(summaryData)
	writer.ShowOutputFiles(outputFiles)
	writer.ShowFailedModels(failedModels)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify all sections are present and synthesis status is correct
	expectedSections := []string{
		"SUMMARY",
		"● 4 models processed",
		"● 3 successful, 1 failed",
		"● Synthesis: ✓ completed",
		"● Output directory:",
		"OUTPUT FILES",
		"model1.md",
		"2.0K",
		"synthesis.md",
		"8.0K",
		"FAILED MODELS",
		"problematic-model",
		"API timeout",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected complete workflow output to contain %q, but it was missing.\nActual output:\n%s", section, output)
		}
	}

	fmt.Printf("Complete workflow output:\n%s", output)
}

// TestSynthesisStatusQuietMode verifies synthesis status respects quiet mode
func TestSynthesisStatusQuietMode(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test summary data with synthesis completed
	summaryData := SummaryData{
		ModelsProcessed:  2,
		SuccessfulModels: 2,
		FailedModels:     0,
		SynthesisStatus:  "completed",
		OutputDirectory:  "/tmp/output",
	}

	// Create console writer and enable quiet mode
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})
	writer.SetQuiet(true)

	// Call ShowSummarySection
	writer.ShowSummarySection(summaryData)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should be empty (quiet mode suppresses summary output)
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %q", output)
	}
}
