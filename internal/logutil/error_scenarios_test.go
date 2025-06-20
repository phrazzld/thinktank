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
				"* 3 models processed",
				"* 0 successful, 3 failed",
				"[!] All models failed to process",
				"* Check your API keys and network connectivity",
				"* Review error details above for specific failure reasons",
				"* Verify model names and rate limits with providers",
			},
			notExpectedMsg: []string{
				"Partial success",
				"All models processed successfully",
				"Success rate:",
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
				"* 5 models processed",
				"* 3 successful, 2 failed",
				"* Synthesis: [OK] completed",
				"[!] Partial success - some models failed",
				"* Success rate: 60% (3/5 models)",
				"* Check failed model details above for specific issues",
				"* Consider retrying failed models or adjusting configuration",
			},
			notExpectedMsg: []string{
				"All models failed to process",
				"All models processed successfully",
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
				"* 4 models processed",
				"* 4 successful, 0 failed",
				"* Synthesis: [OK] completed",
				"[OK] All models processed successfully",
				"* Synthesis completed - check the combined output above",
			},
			notExpectedMsg: []string{
				"All models failed to process",
				"Partial success",
				"Individual model outputs saved",
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
				"* 2 models processed",
				"* 2 successful, 0 failed",
				"[OK] All models processed successfully",
				"* Individual model outputs saved - see file list above",
			},
			notExpectedMsg: []string{
				"All models failed to process",
				"Partial success",
				"Synthesis completed",
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
				"* 1 models processed",
				"* 1 successful, 0 failed",
				"[OK] All models processed successfully",
			},
			notExpectedMsg: []string{
				"All models failed to process",
				"Partial success",
				"Individual model outputs saved", // Should not appear for single model
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

// TestSuccessRateCalculation verifies that success rate calculations are accurate
func TestSuccessRateCalculation(t *testing.T) {
	testCases := []struct {
		name         string
		successful   int
		total        int
		expectedRate string
	}{
		{"75% Success", 3, 4, "75%"},
		{"50% Success", 2, 4, "50%"},
		{"33% Success", 1, 3, "33%"},
		{"67% Success", 2, 3, "67%"},
		{"80% Success", 4, 5, "80%"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout for testing
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			summaryData := SummaryData{
				ModelsProcessed:  tc.total,
				SuccessfulModels: tc.successful,
				FailedModels:     tc.total - tc.successful,
				SynthesisStatus:  "skipped",
				OutputDirectory:  "/tmp/output",
			}

			// Create console writer
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc:  func() bool { return false },
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

			// Verify success rate calculation
			expectedSuccessLine := fmt.Sprintf("* Success rate: %s (%d/%d models)", tc.expectedRate, tc.successful, tc.total)
			if !strings.Contains(output, expectedSuccessLine) {
				t.Errorf("Expected success rate line %q not found in output:\n%s", expectedSuccessLine, output)
			}
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
