package logutil

import (
	"strings"
	"testing"
	"time"
)

func TestStatusDisplay_Basic(t *testing.T) {
	display := NewStatusDisplay(true) // Interactive mode

	// Create test model states
	states := []*ModelState{
		{
			Name:     "model1",
			Index:    1,
			Status:   StatusProcessing,
			Duration: 100 * time.Millisecond,
		},
		{
			Name:     "model2",
			Index:    2,
			Status:   StatusCompleted,
			Duration: 200 * time.Millisecond,
		},
	}

	// Test rendering
	output := captureOutput(func() {
		display.RenderSummaryHeader(2)
		display.RenderStatus(states, true)
		display.RenderCompletion()
	})

	if !strings.Contains(output, "Processing 2 models") {
		t.Errorf("Expected summary header in output, got: %q", output)
	}
}

func TestStatusDisplay_NonInteractive(t *testing.T) {
	display := NewStatusDisplay(false) // Non-interactive mode

	states := []*ModelState{
		{
			Name:   "model1",
			Index:  1,
			Status: StatusCompleted,
		},
	}

	summary := StatusSummary{
		TotalModels:    1,
		CompletedCount: 1,
	}

	output := captureOutput(func() {
		display.RenderPeriodicUpdate(states, summary)
	})

	if !strings.Contains(output, "Status Update") {
		t.Errorf("Expected periodic update header in CI mode, got: %q", output)
	}
}

func TestStatusDisplay_FormatModelLine(t *testing.T) {
	display := NewStatusDisplay(true)
	display.terminalWidth = 80

	testCases := []struct {
		name        string
		state       *ModelState
		totalModels int
		expectParts []string
	}{
		{
			name: "Processing state",
			state: &ModelState{
				Name:        "test-model",
				DisplayName: "test-model",
				Index:       1,
				Status:      StatusProcessing,
			},
			totalModels: 5,
			expectParts: []string{"[1/5]", "test-model", "processing"},
		},
		{
			name: "Completed state",
			state: &ModelState{
				Name:        "test-model",
				DisplayName: "test-model",
				Index:       2,
				Status:      StatusCompleted,
				Duration:    150 * time.Millisecond,
			},
			totalModels: 5,
			expectParts: []string{"[2/5]", "test-model", "completed", "150ms"},
		},
		{
			name: "Failed state",
			state: &ModelState{
				Name:        "test-model",
				DisplayName: "test-model",
				Index:       3,
				Status:      StatusFailed,
				ErrorMsg:    "timeout",
			},
			totalModels: 5,
			expectParts: []string{"[3/5]", "test-model", "timeout"},
		},
		{
			name: "Rate limited state",
			state: &ModelState{
				Name:        "test-model",
				DisplayName: "test-model",
				Index:       4,
				Status:      StatusRateLimited,
				RetryAfter:  2 * time.Second,
			},
			totalModels: 5,
			expectParts: []string{"[4/5]", "test-model", "retry in", "2.0s"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			line := display.formatModelLine(tc.state, tc.totalModels)

			for _, part := range tc.expectParts {
				if !strings.Contains(line, part) {
					t.Errorf("Expected %q to contain %q", line, part)
				}
			}
		})
	}
}

func TestStatusDisplay_ZeroPadding(t *testing.T) {
	display := NewStatusDisplay(true)

	testCases := []struct {
		index       int
		totalModels int
		expected    string
	}{
		{1, 9, "[1/9]"},
		{1, 10, "[01/10]"},
		{1, 99, "[01/99]"},
		{1, 100, "[001/100]"},
		{15, 100, "[015/100]"},
	}

	for _, tc := range testCases {
		state := &ModelState{
			Name:        "test",
			DisplayName: "test",
			Index:       tc.index,
			Status:      StatusQueued,
		}

		line := display.formatModelLine(state, tc.totalModels)
		if !strings.Contains(line, tc.expected) {
			t.Errorf("Expected %q in line for index %d/%d, got: %q", tc.expected, tc.index, tc.totalModels, line)
		}
	}
}

func TestStatusDisplay_GetDisplayWidth(t *testing.T) {
	display := NewStatusDisplay(true)

	testCases := []struct {
		input    string
		expected int
	}{
		{"plain text", 10},
		{"\033[31mred text\033[0m", 8}, // ANSI codes removed
		{"\033[1;36mBold Cyan\033[0m", 9},
		{"", 0},
	}

	for _, tc := range testCases {
		width := display.getDisplayWidth(tc.input)
		if width != tc.expected {
			t.Errorf("getDisplayWidth(%q) = %d, want %d", tc.input, width, tc.expected)
		}
	}
}
