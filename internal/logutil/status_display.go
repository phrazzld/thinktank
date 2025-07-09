package logutil

import (
	"fmt"
	"os"
	"regexp"

	"golang.org/x/term"
)

// StatusDisplay handles environment-aware rendering of model processing status
type StatusDisplay struct {
	isInteractive bool
	terminalWidth int
	colors        *ColorScheme
	symbols       *SymbolProvider
	lastLineCount int // Track lines printed for cursor positioning
}

// NewStatusDisplay creates a new status display with environment detection
func NewStatusDisplay(isInteractive bool) *StatusDisplay {
	width := 80 // default width
	if isInteractive {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			width = w
		}
	}

	return &StatusDisplay{
		isInteractive: isInteractive,
		terminalWidth: width,
		colors:        NewColorScheme(isInteractive),
		symbols:       NewSymbolProvider(isInteractive),
	}
}

// RenderStatus displays the current status of all models
func (d *StatusDisplay) RenderStatus(states []*ModelState, forceRefresh bool) {
	if d.isInteractive && d.lastLineCount > 0 && !forceRefresh {
		// Move cursor up to overwrite previous status
		fmt.Printf("\033[%dA", d.lastLineCount)
	}

	lineCount := 0
	totalModels := len(states)

	// Render each model's status line
	for _, state := range states {
		line := d.formatModelLine(state, totalModels)
		fmt.Println(line)
		lineCount++
	}

	// Add a blank line for separation in CI mode
	if !d.isInteractive {
		fmt.Println()
		lineCount++
	}

	d.lastLineCount = lineCount
}

// formatModelLine creates a formatted status line for a single model
func (d *StatusDisplay) formatModelLine(state *ModelState, totalModels int) string {
	// Model identifier with index - pad the current index to match total digits
	totalDigits := len(fmt.Sprintf("%d", totalModels))
	modelId := fmt.Sprintf("[%0*d/%d]", totalDigits, state.Index, totalModels)

	// Format status with appropriate colors and symbols
	status := d.formatStatus(state)

	// Calculate available space for model name (accounting for colors in status)
	// We need to strip color codes to get the actual display width of the status
	statusDisplayWidth := d.getDisplayWidth(status)
	maxNameWidth := d.terminalWidth - len(modelId) - statusDisplayWidth - 3 // 3 for spacing

	// Truncate model name if too long - use DisplayName for user-facing output
	modelName := state.DisplayName
	if len(modelName) > maxNameWidth {
		if maxNameWidth > 3 {
			modelName = modelName[:maxNameWidth-3] + "..."
		} else {
			modelName = "..."
		}
	}

	// Color the model name
	coloredName := d.colors.ColorModelName(modelName)

	// Calculate padding to right-align the status
	leftSide := fmt.Sprintf("%s %s", modelId, coloredName)
	leftSideDisplayWidth := d.getDisplayWidth(leftSide)
	padding := d.terminalWidth - leftSideDisplayWidth - statusDisplayWidth
	if padding < 1 {
		padding = 1
	}

	return fmt.Sprintf("%s%*s%s", leftSide, padding, "", status)
}

// formatStatus creates a colored status string based on the model's current state
func (d *StatusDisplay) formatStatus(state *ModelState) string {
	switch state.Status {
	case StatusQueued:
		return d.colors.ColorSymbol("queued...")

	case StatusStarting:
		return d.colors.ColorSymbol("starting...")

	case StatusProcessing:
		return d.colors.ColorSymbol("processing...")

	case StatusRateLimited:
		symbol := d.colors.ColorWarning(d.symbols.GetSymbols().Warning)
		retryText := d.colors.ColorDuration(FormatDuration(state.RetryAfter))
		return fmt.Sprintf("%s rate limited (retry in %s)", symbol, retryText)

	case StatusCompleted:
		symbol := d.colors.ColorSuccess(d.symbols.GetSymbols().Success)
		duration := d.colors.ColorDuration(FormatDuration(state.Duration))
		return fmt.Sprintf("%s completed (%s)", symbol, duration)

	case StatusFailed:
		symbol := d.colors.ColorError(d.symbols.GetSymbols().Error)
		errorMsg := state.ErrorMsg
		if errorMsg == "" {
			errorMsg = "error"
		}
		coloredError := d.colors.ColorError(errorMsg)
		return fmt.Sprintf("%s failed (%s)", symbol, coloredError)

	default:
		return "unknown"
	}
}

// RenderSummaryHeader displays initial processing information
func (d *StatusDisplay) RenderSummaryHeader(totalModels int) {
	fmt.Printf("\nProcessing %d models...\n", totalModels)
}

// RenderCompletion displays final completion message and clears status area
func (d *StatusDisplay) RenderCompletion() {
	if d.isInteractive && d.lastLineCount > 0 {
		// Clear the status lines by overwriting with empty lines
		for i := 0; i < d.lastLineCount; i++ {
			fmt.Printf("\033[2K\n") // Clear line and move to next
		}
		// Move cursor back up
		fmt.Printf("\033[%dA", d.lastLineCount)
	}
	d.lastLineCount = 0
	fmt.Println() // Add spacing before next section
}

// RenderPeriodicUpdate renders status in CI mode with headers
func (d *StatusDisplay) RenderPeriodicUpdate(states []*ModelState, summary StatusSummary) {
	if d.isInteractive {
		return // Interactive mode uses real-time updates
	}

	fmt.Printf("Status Update - %d/%d completed (%.0f%%):\n",
		summary.CompletedCount+summary.FailedCount,
		summary.TotalModels,
		summary.GetCompletionRate())

	d.RenderStatus(states, true)
}

// getDisplayWidth calculates the actual display width of a string, excluding ANSI color codes
func (d *StatusDisplay) getDisplayWidth(s string) int {
	// Regular expression to match ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	// Remove ANSI codes and return the length
	cleaned := ansiRegex.ReplaceAllString(s, "")
	return len(cleaned)
}
