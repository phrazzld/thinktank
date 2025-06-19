package logutil

import (
	"strings"
)

// LayoutConfig defines column widths and spacing for different terminal sizes.
// It provides responsive layout that adapts to narrow, standard, and wide terminals
// while maintaining proper alignment and readability.
type LayoutConfig struct {
	TerminalWidth  int // Total terminal width in characters
	ModelNameWidth int // Width allocated for model names (left column)
	StatusWidth    int // Width allocated for status text (right column)
	FileNameWidth  int // Width allocated for file names in file listings
	FileSizeWidth  int // Width allocated for file sizes (right-aligned)
	MinPadding     int // Minimum padding between columns
}

// Terminal width constants for responsive layout calculation
const (
	// MinLayoutWidth is the absolute minimum width we can work with
	MinLayoutWidth = 50
	// StandardLayoutWidth is the optimal width for most use cases
	StandardLayoutWidth = 80
	// WideLayoutWidth is the threshold for wide terminal optimizations
	WideLayoutWidth = 120
	// MaxLayoutWidth is the maximum width we'll optimize for
	MaxLayoutWidth = 160

	// Default column allocations for different layouts
	defaultModelNameWidth = 30
	defaultStatusWidth    = 20
	defaultFileNameWidth  = 40
	defaultFileSizeWidth  = 8
	defaultMinPadding     = 2
)

// CalculateLayout computes optimal column widths for the given terminal width.
// It provides responsive behavior that adapts to different terminal sizes
// while ensuring content remains readable and properly aligned.
//
// Layout strategies:
// - Narrow terminals (< 50 chars): Minimal layout with reduced padding
// - Standard terminals (50-120 chars): Balanced layout with good readability
// - Wide terminals (> 120 chars): Expanded layout with generous spacing
func CalculateLayout(terminalWidth int) LayoutConfig {
	// Ensure minimum width
	if terminalWidth < MinLayoutWidth {
		return calculateNarrowLayout(terminalWidth)
	}

	// Standard width handling
	if terminalWidth <= WideLayoutWidth {
		return calculateStandardLayout(terminalWidth)
	}

	// Wide terminal handling
	return calculateWideLayout(terminalWidth)
}

// calculateNarrowLayout handles terminals with limited width (< 50 chars).
// Prioritizes essential information with minimal padding.
func calculateNarrowLayout(width int) LayoutConfig {
	// For very narrow terminals, use minimal viable layout
	minPadding := 1
	if width < 20 {
		// Extremely narrow - just try to fit something useful
		return LayoutConfig{
			TerminalWidth:  width,
			ModelNameWidth: width/2 - 1,
			StatusWidth:    width/2 - 1,
			FileNameWidth:  width/2 - 1,
			FileSizeWidth:  6, // Minimum for "1.2K" format
			MinPadding:     1,
		}
	}

	// Narrow but workable
	modelNameWidth := (width * 6) / 10 // 60% for model name
	statusWidth := width - modelNameWidth - minPadding

	return LayoutConfig{
		TerminalWidth:  width,
		ModelNameWidth: modelNameWidth,
		StatusWidth:    statusWidth,
		FileNameWidth:  modelNameWidth, // Reuse for file listings
		FileSizeWidth:  8,
		MinPadding:     minPadding,
	}
}

// calculateStandardLayout handles standard terminal widths (50-120 chars).
// Provides balanced layout with good readability and proper alignment.
func calculateStandardLayout(width int) LayoutConfig {
	minPadding := defaultMinPadding

	// For standard terminals, use proportional allocation
	// Model name gets 60-70% of available space, status gets remainder
	availableSpace := width - minPadding
	modelNameWidth := (availableSpace * 65) / 100 // 65% for model name
	statusWidth := availableSpace - modelNameWidth

	// File listing layout - filename gets more space than status
	fileNameWidth := (availableSpace * 75) / 100 // 75% for filename
	fileSizeWidth := availableSpace - fileNameWidth

	// Ensure minimum viable widths
	if statusWidth < 15 {
		modelNameWidth = width - 15 - minPadding
		statusWidth = 15
	}

	return LayoutConfig{
		TerminalWidth:  width,
		ModelNameWidth: modelNameWidth,
		StatusWidth:    statusWidth,
		FileNameWidth:  fileNameWidth,
		FileSizeWidth:  fileSizeWidth,
		MinPadding:     minPadding,
	}
}

// calculateWideLayout handles wide terminals (> 120 chars).
// Provides generous spacing and expanded columns for optimal readability.
func calculateWideLayout(width int) LayoutConfig {
	minPadding := 3 // More generous padding for wide terminals

	// Cap the effective width to avoid excessive stretching
	effectiveWidth := width
	if effectiveWidth > MaxLayoutWidth {
		effectiveWidth = MaxLayoutWidth
	}

	// For wide terminals, provide generous fixed widths
	modelNameWidth := 45 // Generous space for long model names
	statusWidth := 25    // Ample space for status messages
	fileNameWidth := 60  // Comfortable space for file names
	fileSizeWidth := 12  // Extra space for large file sizes

	// Ensure we don't exceed terminal width
	totalUsed := modelNameWidth + statusWidth + minPadding
	if totalUsed > effectiveWidth {
		// Fall back to proportional allocation if fixed widths don't fit
		return calculateStandardLayout(effectiveWidth)
	}

	return LayoutConfig{
		TerminalWidth:  width,
		ModelNameWidth: modelNameWidth,
		StatusWidth:    statusWidth,
		FileNameWidth:  fileNameWidth,
		FileSizeWidth:  fileSizeWidth,
		MinPadding:     minPadding,
	}
}

// FormatAlignedText formats text with proper alignment for the layout.
// Handles left-alignment for names and right-alignment for status/sizes.
func (lc *LayoutConfig) FormatAlignedText(leftText, rightText string) string {
	// Calculate total available space for content
	totalSpace := lc.ModelNameWidth + lc.StatusWidth

	// Truncate left text if it's too long
	if len(leftText) > lc.ModelNameWidth {
		if lc.ModelNameWidth > 3 {
			leftText = leftText[:lc.ModelNameWidth-3] + "..."
		} else {
			leftText = leftText[:lc.ModelNameWidth]
		}
	}

	// Truncate right text if it's too long
	if len(rightText) > lc.StatusWidth {
		if lc.StatusWidth > 3 {
			rightText = rightText[:lc.StatusWidth-3] + "..."
		} else {
			rightText = rightText[:lc.StatusWidth]
		}
	}

	// Calculate padding needed
	usedSpace := len(leftText) + len(rightText)
	padding := totalSpace - usedSpace

	// Ensure minimum padding
	if padding < lc.MinPadding {
		padding = lc.MinPadding
	}

	return leftText + strings.Repeat(" ", padding) + rightText
}

// FormatFileListItem formats a file listing item with proper alignment.
// Used for output file listings with right-aligned sizes.
func (lc *LayoutConfig) FormatFileListItem(filename, size string) string {
	// Calculate total available space
	totalSpace := lc.FileNameWidth + lc.FileSizeWidth

	// Truncate filename if needed
	if len(filename) > lc.FileNameWidth {
		if lc.FileNameWidth > 3 {
			filename = filename[:lc.FileNameWidth-3] + "..."
		} else {
			filename = filename[:lc.FileNameWidth]
		}
	}

	// Calculate padding
	usedSpace := len(filename) + len(size)
	padding := totalSpace - usedSpace

	// Ensure minimum padding
	if padding < lc.MinPadding {
		padding = lc.MinPadding
	}

	return filename + strings.Repeat(" ", padding) + size
}

// GetSeparatorLine returns a separator line that fits the terminal width.
// Used for section headers like "SUMMARY" and "OUTPUT FILES".
func (lc *LayoutConfig) GetSeparatorLine(length int) string {
	if length <= 0 {
		length = lc.TerminalWidth / 2 // Default to half terminal width
	}

	// Don't exceed terminal width
	if length > lc.TerminalWidth {
		length = lc.TerminalWidth
	}

	return strings.Repeat("─", length)
}

// IsNarrowTerminal returns true if the terminal is considered narrow.
// Used to adjust output behavior for space-constrained environments.
func (lc *LayoutConfig) IsNarrowTerminal() bool {
	return lc.TerminalWidth < StandardLayoutWidth
}

// IsWideTerminal returns true if the terminal is considered wide.
// Used to enable enhanced formatting for spacious environments.
func (lc *LayoutConfig) IsWideTerminal() bool {
	return lc.TerminalWidth > WideLayoutWidth
}
