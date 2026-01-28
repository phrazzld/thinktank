package logutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestCompleteErrorScenarioHandling verifies that different failure scenarios
// display appropriate contextual messaging and actionable guidance
func TestCompleteErrorScenarioHandling(t *testing.T) {
	tests := []struct {
		name             string
		summaryData      SummaryData
		expectedMessages []string
		notExpectedMsg   []string
	}{
		{
			name: "All Models Failed",
			summaryData: SummaryData{
				ModelsProcessed:  3,
				SuccessfulModels: 0,
				FailedModels:     3,
				SynthesisStatus:  "skipped",
				OutputDirectory:  "/tmp/output",
			},
			expectedMessages: []string{
				"SUMMARY",
				"3 processed",
				"x x x",
				"All models failed - review errors above",
			},
			notExpectedMsg: []string{
				"Synthesis",
			},
		},
		{
			name: "Partial Success",
			summaryData: SummaryData{
				ModelsProcessed:  5,
				SuccessfulModels: 3,
				FailedModels:     2,
				SynthesisStatus:  "completed",
				OutputDirectory:  "/tmp/output",
			},
			expectedMessages: []string{
				"SUMMARY",
				"5 processed",
				"o o o x x",
				"Synthesis",
				"[OK] completed",
				"2 models failed - review errors above",
			},
			notExpectedMsg: []string{
				"All models failed",
			},
		},
		{
			name: "Complete Success with Synthesis",
			summaryData: SummaryData{
				ModelsProcessed:  4,
				SuccessfulModels: 4,
				FailedModels:     0,
				SynthesisStatus:  "completed",
				OutputDirectory:  "/tmp/output",
			},
			expectedMessages: []string{
				"SUMMARY",
				"4 processed",
				"o o o o",
				"Synthesis",
				"[OK] completed",
			},
			notExpectedMsg: []string{
				"failed - review errors above",
			},
		},
		{
			name: "Complete Success without Synthesis",
			summaryData: SummaryData{
				ModelsProcessed:  2,
				SuccessfulModels: 2,
				FailedModels:     0,
				SynthesisStatus:  "skipped",
				OutputDirectory:  "/tmp/output",
			},
			expectedMessages: []string{
				"SUMMARY",
				"2 processed",
				"o o",
			},
			notExpectedMsg: []string{
				"Synthesis",
				"failed - review errors above",
			},
		},
		{
			name: "Single Model Success",
			summaryData: SummaryData{
				ModelsProcessed:  1,
				SuccessfulModels: 1,
				FailedModels:     0,
				SynthesisStatus:  "skipped",
				OutputDirectory:  "/tmp/output",
			},
			expectedMessages: []string{
				"SUMMARY",
				"1 processed",
				"o",
			},
			notExpectedMsg: []string{
				"Synthesis",
				"failed - review errors above",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout for testing
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create console writer with test options
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc:  func() bool { return false }, // Non-interactive for cleaner output
				GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
				GetEnvFunc:      func(key string) string { return "" },
			})

			// Call ShowSummarySection
			writer.ShowSummarySection(tt.summaryData)

			// Restore stdout and read captured output
			_ = w.Close()
			os.Stdout = old

			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify expected messages are present
			for _, expected := range tt.expectedMessages {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it was missing.\nActual output:\n%s", expected, output)
				}
			}

			// Verify unexpected messages are not present
			for _, notExpected := range tt.notExpectedMsg {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output NOT to contain %q, but it was present.\nActual output:\n%s", notExpected, output)
				}
			}

			fmt.Printf("Test %s output:\n%s\n", tt.name, output)
		})
	}
}

// TestScenarioGuidanceQuietMode verifies that scenario guidance respects quiet mode
func TestScenarioGuidanceQuietMode(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create summary data that would normally trigger guidance
	summaryData := SummaryData{
		ModelsProcessed:  3,
		SuccessfulModels: 0,
		FailedModels:     3,
		SynthesisStatus:  "skipped",
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

	// Should be empty (quiet mode suppresses all summary output)
	if output != "" {
		t.Errorf("Expected no output in quiet mode, got: %q", output)
	}
}
