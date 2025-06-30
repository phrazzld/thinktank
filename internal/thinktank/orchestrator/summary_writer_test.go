package orchestrator

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// SimpleTestLogger is a simplified test logger for unit tests
type SimpleTestLogger struct{}

func NewSimpleTestLogger() *SimpleTestLogger {
	return &SimpleTestLogger{}
}

func (l *SimpleTestLogger) Println(v ...interface{})                                             {}
func (l *SimpleTestLogger) Printf(format string, v ...interface{})                               {}
func (l *SimpleTestLogger) Debug(format string, args ...interface{})                             {}
func (l *SimpleTestLogger) Info(format string, args ...interface{})                              {}
func (l *SimpleTestLogger) Warn(format string, args ...interface{})                              {}
func (l *SimpleTestLogger) Error(format string, args ...interface{})                             {}
func (l *SimpleTestLogger) Fatal(format string, args ...interface{})                             {}
func (l *SimpleTestLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {}
func (l *SimpleTestLogger) InfoContext(ctx context.Context, format string, args ...interface{})  {}
func (l *SimpleTestLogger) WarnContext(ctx context.Context, format string, args ...interface{})  {}
func (l *SimpleTestLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {}
func (l *SimpleTestLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {}

// MockConsoleWriter is a simplified test console writer for unit tests
type MockConsoleWriter struct{}

func (m *MockConsoleWriter) StartProcessing(modelCount int)                             {}
func (m *MockConsoleWriter) ModelQueued(modelName string, index int)                    {}
func (m *MockConsoleWriter) ModelStarted(modelIndex, totalModels int, modelName string) {}
func (m *MockConsoleWriter) ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration) {
}
func (m *MockConsoleWriter) ModelFailed(modelIndex, totalModels int, modelName string, reason string) {
}
func (m *MockConsoleWriter) ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration) {
}
func (m *MockConsoleWriter) ShowProcessingLine(modelName string)                  {}
func (m *MockConsoleWriter) UpdateProcessingLine(modelName string, status string) {}
func (m *MockConsoleWriter) ShowFileOperations(message string)                    {}
func (m *MockConsoleWriter) ShowSummarySection(summary logutil.SummaryData)       {}
func (m *MockConsoleWriter) ShowOutputFiles(files []logutil.OutputFile)           {}
func (m *MockConsoleWriter) ShowFailedModels(failed []logutil.FailedModel)        {}
func (m *MockConsoleWriter) SynthesisStarted()                                    {}
func (m *MockConsoleWriter) SynthesisCompleted(outputPath string)                 {}
func (m *MockConsoleWriter) StatusMessage(message string)                         {}
func (m *MockConsoleWriter) SetQuiet(quiet bool)                                  {}
func (m *MockConsoleWriter) SetNoProgress(noProgress bool)                        {}
func (m *MockConsoleWriter) IsInteractive() bool                                  { return false }
func (m *MockConsoleWriter) GetTerminalWidth() int                                { return 80 }
func (m *MockConsoleWriter) FormatMessage(message string) string                  { return message }
func (m *MockConsoleWriter) ErrorMessage(message string)                          {}
func (m *MockConsoleWriter) WarningMessage(message string)                        {}
func (m *MockConsoleWriter) SuccessMessage(message string)                        {}
func (m *MockConsoleWriter) StartStatusTracking(modelNames []string)              {}
func (m *MockConsoleWriter) UpdateModelStatus(modelName string, status logutil.ModelStatus, duration time.Duration, errorMsg string) {
}
func (m *MockConsoleWriter) UpdateModelRateLimited(modelName string, retryAfter time.Duration) {}
func (m *MockConsoleWriter) RefreshStatusDisplay()                                             {}
func (m *MockConsoleWriter) FinishStatusTracking()                                             {}
func (l *SimpleTestLogger) WithContext(ctx context.Context) logutil.LoggerInterface            { return l }

// stripAnsiColors removes ANSI color codes from a string
func stripAnsiColors(s string) string {
	// Regular expression to match ANSI escape sequences for colors
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

// TestStripAnsiColors tests the utility function that removes ANSI color codes
func TestStripAnsiColors(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No color codes",
			input:    "Plain text with no colors",
			expected: "Plain text with no colors",
		},
		{
			name:     "Simple color code",
			input:    "\033[31mRed text\033[0m",
			expected: "Red text",
		},
		{
			name:     "Multiple color codes",
			input:    "\033[32mGreen\033[0m and \033[31mRed\033[0m",
			expected: "Green and Red",
		},
		{
			name:     "Complex formatting",
			input:    "\033[1;36mBold Cyan\033[0m and \033[3;33;41mItalic Yellow on Red\033[0m",
			expected: "Bold Cyan and Italic Yellow on Red",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripAnsiColors(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	logger := NewSimpleTestLogger()
	consoleWriter := &MockConsoleWriter{}
	summaryWriter := NewSummaryWriter(logger, consoleWriter)

	tests := []struct {
		name           string
		summary        *ResultsSummary
		expectedParts  []string
		notExpectedStr string
	}{
		{
			name: "CompleteSuccess",
			summary: &ResultsSummary{
				TotalModels:      2,
				SuccessfulModels: 2,
				SuccessfulNames:  []string{"model1", "model2"},
				SynthesisPath:    "/path/to/synthesis.md",
			},
			expectedParts: []string{
				"SUCCESS",
				"Models: 2 total, 2 successful, 0 failed",
				"Synthesis file:",
				"/path/to/synthesis.md",
				"Successful models: 2 models",
			},
			notExpectedStr: "Failed models:",
		},
		{
			name: "PartialSuccess",
			summary: &ResultsSummary{
				TotalModels:      3,
				SuccessfulModels: 1,
				SuccessfulNames:  []string{"model1"},
				FailedModels:     []string{"model2", "model3"},
				SynthesisPath:    "/path/to/synthesis.md",
			},
			expectedParts: []string{
				"PARTIAL SUCCESS",
				"Models: 3 total, 1 successful, 2 failed",
				"Synthesis file:",
				"/path/to/synthesis.md",
				"Successful models:",
				"Failed models:",
			},
		},
		{
			name: "CompleteFailure",
			summary: &ResultsSummary{
				TotalModels:      2,
				SuccessfulModels: 0,
				FailedModels:     []string{"model1", "model2"},
			},
			expectedParts: []string{
				"FAILED",
				"Models: 2 total, 0 successful, 2 failed",
				"Failed models:",
			},
			notExpectedStr: "Synthesis file:",
		},
		{
			name: "IndividualOutputs",
			summary: &ResultsSummary{
				TotalModels:      2,
				SuccessfulModels: 2,
				SuccessfulNames:  []string{"model1", "model2"},
				OutputPaths:      []string{"/path/to/model1.md", "/path/to/model2.md"},
			},
			expectedParts: []string{
				"SUCCESS",
				"Models: 2 total, 2 successful, 0 failed",
				"Output files:",
				"/path/to/model1.md",
				"/path/to/model2.md",
			},
			notExpectedStr: "Synthesis file:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := summaryWriter.GenerateSummary(tc.summary)
			plainResult := stripAnsiColors(result)

			// Check that the result contains all expected parts
			for _, expected := range tc.expectedParts {
				if !strings.Contains(plainResult, expected) {
					t.Errorf("Expected summary to contain %q but it did not.\nSummary:\n%s", expected, result)
				}
			}

			// Check that the result does not contain unexpected parts
			if tc.notExpectedStr != "" && strings.Contains(plainResult, tc.notExpectedStr) {
				t.Errorf("Summary should not contain %q but it did.\nSummary:\n%s", tc.notExpectedStr, result)
			}
		})
	}
}

func TestDisplaySummary(t *testing.T) {
	logger := NewSimpleTestLogger()
	consoleWriter := &MockConsoleWriter{}
	summaryWriter := NewSummaryWriter(logger, consoleWriter)

	summary := &ResultsSummary{
		TotalModels:      3,
		SuccessfulModels: 2,
		SuccessfulNames:  []string{"model1", "model2"},
		FailedModels:     []string{"model3"},
		SynthesisPath:    "/path/to/synthesis.md",
	}

	// Just test that it doesn't panic
	summaryWriter.DisplaySummary(context.Background(), summary)

	// Since SimpleTestLogger doesn't track log entries, we can only test that it doesn't panic
}

func TestTruncateHelpers(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		maxLen   int
		expected string
	}{
		{
			name:     "Short path not truncated",
			path:     "/short/path.md",
			maxLen:   20,
			expected: "/short/path.md",
		},
		{
			name:     "Long path truncated",
			path:     "/very/long/path/to/some/file/with/a/very/long/name.md",
			maxLen:   15,
			expected: "...long/name.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := truncatePath(tc.path, tc.maxLen)
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}

	// Test truncateList
	items := []string{"model1", "model2", "model3"}
	result := truncateList(items, 30)
	if !strings.Contains(result, "3 models") || !strings.Contains(result, "model1") {
		t.Errorf("Expected truncateList to include count and model names, got: %s", result)
	}

	// Test with a very short max length
	shortResult := truncateList(items, 10)
	if len(shortResult) > 10 {
		t.Errorf("truncateList exceeded max length: got %d, expected ≤ 10", len(shortResult))
	}
}
