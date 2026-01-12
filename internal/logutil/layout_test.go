package logutil

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCalculateLayout_NarrowTerminal(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		expectNarrow  bool
	}{
		{"Very narrow", 20, true},
		{"Narrow", 40, true},
		{"Just under standard", 49, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := CalculateLayout(tt.terminalWidth)

			if layout.TerminalWidth != tt.terminalWidth {
				t.Errorf("Expected TerminalWidth %d, got %d", tt.terminalWidth, layout.TerminalWidth)
			}

			if layout.IsNarrowTerminal() != tt.expectNarrow {
				t.Errorf("Expected IsNarrowTerminal() %v, got %v", tt.expectNarrow, layout.IsNarrowTerminal())
			}

			// Verify basic constraints
			if layout.ModelNameWidth < 1 {
				t.Error("ModelNameWidth should be at least 1")
			}
			if layout.StatusWidth < 1 {
				t.Error("StatusWidth should be at least 1")
			}
			if layout.MinPadding < 1 {
				t.Error("MinPadding should be at least 1")
			}

			// Total space should not exceed terminal width
			totalUsed := layout.ModelNameWidth + layout.StatusWidth + layout.MinPadding
			if totalUsed > tt.terminalWidth {
				t.Errorf("Total used space %d exceeds terminal width %d", totalUsed, tt.terminalWidth)
			}
		})
	}
}

func TestCalculateLayout_StandardTerminal(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
	}{
		{"Standard 80", 80},
		{"Standard 100", 100},
		{"Just under wide", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := CalculateLayout(tt.terminalWidth)

			if layout.TerminalWidth != tt.terminalWidth {
				t.Errorf("Expected TerminalWidth %d, got %d", tt.terminalWidth, layout.TerminalWidth)
			}

			if layout.IsNarrowTerminal() {
				t.Error("Standard terminal should not be considered narrow")
			}
			if layout.IsWideTerminal() {
				t.Error("Standard terminal should not be considered wide")
			}

			// Verify reasonable proportions for standard terminals
			if layout.ModelNameWidth < 20 {
				t.Error("ModelNameWidth should be at least 20 for standard terminals")
			}
			if layout.StatusWidth < 15 {
				t.Error("StatusWidth should be at least 15 for standard terminals")
			}

			// Total space should not exceed terminal width
			totalUsed := layout.ModelNameWidth + layout.StatusWidth + layout.MinPadding
			if totalUsed > tt.terminalWidth {
				t.Errorf("Total used space %d exceeds terminal width %d", totalUsed, tt.terminalWidth)
			}
		})
	}
}

func TestCalculateLayout_WideTerminal(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		expectWide    bool
	}{
		{"Wide 130", 130, true},
		{"Very wide 200", 200, true},
		{"Extremely wide 300", 300, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := CalculateLayout(tt.terminalWidth)

			if layout.TerminalWidth != tt.terminalWidth {
				t.Errorf("Expected TerminalWidth %d, got %d", tt.terminalWidth, layout.TerminalWidth)
			}

			if layout.IsWideTerminal() != tt.expectWide {
				t.Errorf("Expected IsWideTerminal() %v, got %v", tt.expectWide, layout.IsWideTerminal())
			}

			if layout.IsNarrowTerminal() {
				t.Error("Wide terminal should not be considered narrow")
			}

			// Wide terminals should have generous allocations
			if layout.ModelNameWidth < 30 {
				t.Error("ModelNameWidth should be generous for wide terminals")
			}
			if layout.StatusWidth < 20 {
				t.Error("StatusWidth should be generous for wide terminals")
			}
			if layout.MinPadding < 2 {
				t.Error("MinPadding should be at least 2 for wide terminals")
			}
		})
	}
}

func TestFormatAlignedText(t *testing.T) {
	tests := []struct {
		name         string
		layout       LayoutConfig
		leftText     string
		rightText    string
		expectLength int
	}{
		{
			name: "Standard alignment",
			layout: LayoutConfig{
				TerminalWidth:  80,
				ModelNameWidth: 40,
				StatusWidth:    20,
				MinPadding:     2,
			},
			leftText:     "gemini-3-flash",
			rightText:    "✓ 68.5s",
			expectLength: 60, // 40 + 20
		},
		{
			name: "Long left text truncation",
			layout: LayoutConfig{
				TerminalWidth:  50,
				ModelNameWidth: 20,
				StatusWidth:    15,
				MinPadding:     2,
			},
			leftText:  "very-long-model-name-that-exceeds-width",
			rightText: "✓ done",
			// Should truncate with "..."
		},
		{
			name: "Long right text truncation",
			layout: LayoutConfig{
				TerminalWidth:  50,
				ModelNameWidth: 20,
				StatusWidth:    10,
				MinPadding:     2,
			},
			leftText:  "model",
			rightText: "very long status message",
			// Should truncate with "..."
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.layout.FormatAlignedText(tt.leftText, tt.rightText)

			if tt.expectLength > 0 {
				if len(result) < tt.expectLength {
					t.Errorf("Expected minimum length %d, got %d", tt.expectLength, len(result))
				}
			}

			// Verify the result contains both texts (possibly truncated)
			if !strings.Contains(result, tt.leftText[:min(len(tt.leftText), tt.layout.ModelNameWidth)]) &&
				!strings.Contains(result, tt.leftText[:max(0, min(len(tt.leftText), tt.layout.ModelNameWidth-3))]) {
				t.Error("Result should contain left text or its truncation")
			}

			// Verify padding exists
			paddingCount := strings.Count(result, " ")
			if paddingCount < tt.layout.MinPadding {
				t.Errorf("Expected at least %d spaces for padding, got %d", tt.layout.MinPadding, paddingCount)
			}
		})
	}
}

func TestFormatFileListItem(t *testing.T) {
	tests := []struct {
		name     string
		layout   LayoutConfig
		filename string
		size     string
	}{
		{
			name: "Standard file listing",
			layout: LayoutConfig{
				FileNameWidth: 30,
				FileSizeWidth: 8,
				MinPadding:    2,
			},
			filename: "gemini-3-flash.md",
			size:     "4.2K",
		},
		{
			name: "Long filename truncation",
			layout: LayoutConfig{
				FileNameWidth: 15,
				FileSizeWidth: 8,
				MinPadding:    2,
			},
			filename: "very-long-filename-that-exceeds-width.md",
			size:     "1.5M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.layout.FormatFileListItem(tt.filename, tt.size)

			// Should contain the size
			if !strings.Contains(result, tt.size) {
				t.Errorf("Result should contain size %q, got %q", tt.size, result)
			}

			// Should have appropriate padding
			paddingCount := strings.Count(result, " ")
			if paddingCount < tt.layout.MinPadding {
				t.Errorf("Expected at least %d spaces for padding, got %d", tt.layout.MinPadding, paddingCount)
			}

			// Should not exceed total width
			totalWidth := tt.layout.FileNameWidth + tt.layout.FileSizeWidth
			if len(result) > totalWidth+tt.layout.MinPadding {
				t.Errorf("Result length %d exceeds expected max %d", len(result), totalWidth+tt.layout.MinPadding)
			}
		})
	}
}

func TestGetSeparatorLine(t *testing.T) {
	tests := []struct {
		name          string
		layout        LayoutConfig
		requestLength int
		expectLength  int
	}{
		{
			name:          "Default length",
			layout:        LayoutConfig{TerminalWidth: 80},
			requestLength: 0,
			expectLength:  40, // Half of terminal width
		},
		{
			name:          "Custom length",
			layout:        LayoutConfig{TerminalWidth: 80},
			requestLength: 20,
			expectLength:  20,
		},
		{
			name:          "Length exceeds terminal width",
			layout:        LayoutConfig{TerminalWidth: 80},
			requestLength: 100,
			expectLength:  80, // Clamped to terminal width
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.layout.GetSeparatorLine(tt.requestLength)

			if utf8.RuneCountInString(result) != tt.expectLength {
				t.Errorf("Expected length %d, got %d", tt.expectLength, utf8.RuneCountInString(result))
			}

			// Verify it's all separator characters
			if !strings.Contains(result, "─") {
				t.Error("Separator line should contain separator characters")
			}

			// Verify consistency
			expectedLine := strings.Repeat("─", tt.expectLength)
			if result != expectedLine {
				t.Errorf("Expected %q, got %q", expectedLine, result)
			}
		})
	}
}

func TestLayoutEdgeCases(t *testing.T) {
	t.Run("Extremely narrow terminal", func(t *testing.T) {
		layout := CalculateLayout(10)

		// Should still provide usable layout even for very narrow terminals
		if layout.ModelNameWidth < 1 {
			t.Error("Should provide at least 1 character for model name")
		}
		if layout.StatusWidth < 1 {
			t.Error("Should provide at least 1 character for status")
		}
	})

	t.Run("Zero width terminal", func(t *testing.T) {
		layout := CalculateLayout(0)

		// Should handle gracefully without panicking
		if layout.TerminalWidth != 0 {
			t.Error("Should preserve terminal width even if zero")
		}
	})

	t.Run("Negative width terminal", func(t *testing.T) {
		layout := CalculateLayout(-10)

		// Should handle gracefully
		if layout.ModelNameWidth < 0 || layout.StatusWidth < 0 {
			t.Error("Widths should not be negative")
		}
	})
}

func TestLayoutConstraints(t *testing.T) {
	widths := []int{20, 40, 60, 80, 100, 120, 150, 200}

	for _, width := range widths {
		t.Run(fmt.Sprintf("Width_%d", width), func(t *testing.T) {
			layout := CalculateLayout(width)

			// Basic sanity checks
			if layout.TerminalWidth != width {
				t.Errorf("Terminal width mismatch: expected %d, got %d", width, layout.TerminalWidth)
			}

			// Widths should be positive
			if layout.ModelNameWidth <= 0 {
				t.Error("ModelNameWidth should be positive")
			}
			if layout.StatusWidth <= 0 {
				t.Error("StatusWidth should be positive")
			}
			if layout.FileNameWidth <= 0 {
				t.Error("FileNameWidth should be positive")
			}
			if layout.FileSizeWidth <= 0 {
				t.Error("FileSizeWidth should be positive")
			}

			// MinPadding should be reasonable
			if layout.MinPadding < 1 || layout.MinPadding > 5 {
				t.Errorf("MinPadding should be between 1-5, got %d", layout.MinPadding)
			}
		})
	}
}

// Helper functions for test readability
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
