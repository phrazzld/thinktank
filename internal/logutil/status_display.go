package logutil

import (
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// StatusDisplay handles environment-aware rendering of model processing status
type StatusDisplay struct {
	isInteractive     bool
	terminalWidth     int
	colors            *ColorScheme
	lastLineCount     int // Track lines printed for cursor positioning
	spinner           spinner.Model
	spinnerTick       *time.Ticker
	spinnerDone       chan struct{}
	progress          *GridProgress
	onSpinnerTick     func()
	maxModelNameWidth int
	indexColumnWidth  int
	mu                sync.Mutex // protect spinner state
}

// NewStatusDisplay creates a new status display with environment detection
func NewStatusDisplay(isInteractive bool) *StatusDisplay {
	width := 80 // default width
	if isInteractive {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			width = w
		}
	}

	display := &StatusDisplay{
		isInteractive: isInteractive,
		terminalWidth: width,
		colors:        NewColorScheme(isInteractive),
	}
	if isInteractive {
		display.spinner = spinner.New()
		display.spinner.Spinner = spinner.Dot
		display.spinnerTick = time.NewTicker(100 * time.Millisecond)
		display.spinnerDone = make(chan struct{})
		display.progress = NewGridProgress(display.colors, progressWidth(width))
		go display.runSpinnerTicker()
	}
	return display
}

func (d *StatusDisplay) SetOnSpinnerTick(cb func()) {
	d.mu.Lock()
	d.onSpinnerTick = cb
	d.mu.Unlock()
}

func (d *StatusDisplay) runSpinnerTicker() {
	d.mu.Lock()
	ticker := d.spinnerTick
	done := d.spinnerDone
	d.mu.Unlock()
	if ticker == nil || done == nil {
		return
	}

	for {
		select {
		case <-ticker.C:
			var cb func()
			d.mu.Lock()
			d.spinner, _ = d.spinner.Update(d.spinner.Tick())
			cb = d.onSpinnerTick
			d.mu.Unlock()
			if cb != nil {
				cb()
			}
		case <-done:
			return
		}
	}
}

// Stop halts the spinner ticker if running.
func (d *StatusDisplay) Stop() {
	d.mu.Lock()
	ticker := d.spinnerTick
	done := d.spinnerDone
	d.spinnerTick = nil
	d.spinnerDone = nil
	d.mu.Unlock()

	if ticker == nil || done == nil {
		return
	}

	ticker.Stop()
	close(done)
}

// RenderStatus displays the current status of all models.
// Caller is responsible for serialization (consoleWriter holds its own mutex).
func (d *StatusDisplay) RenderStatus(states []*ModelState, forceRefresh bool) {
	if d.isInteractive && d.lastLineCount > 0 && !forceRefresh {
		// Move cursor up to overwrite previous status
		fmt.Printf("\033[%dA", d.lastLineCount)
	}

	lineCount := 0
	totalModels := len(states)
	d.CalculateLayout(states)

	// Render each model's status line
	for _, state := range states {
		line := d.formatModelLine(state, totalModels)
		if d.isInteractive {
			fmt.Printf("\033[2K%s\n", line)
		} else {
			fmt.Println(line)
		}
		lineCount++
	}

	completed, percent := d.calculateProgress(states, totalModels)
	if d.isInteractive {
		separator := strings.Repeat("─", d.terminalWidth)
		fmt.Println(d.colors.ColorSeparator(separator))
		d.progress.SetTotalCells(progressWidth(d.terminalWidth))
		percentText := fmt.Sprintf("%.0f%%", percent*100)
		fmt.Printf("Overall: [%s] %s\n", d.progress.Render(percent), percentText)
		lineCount += 2
	} else {
		percentText := fmt.Sprintf("%.0f%%", percent*100)
		fmt.Printf("Overall: %s (%d/%d completed)\n", percentText, completed, totalModels)
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

	indicator := d.formatIndicator(state)
	statusText := d.formatStatusText(state)

	if d.indexColumnWidth == 0 || d.maxModelNameWidth == 0 {
		d.indexColumnWidth = len(fmt.Sprintf("[%0*d/%d]", totalDigits, totalModels, totalModels))
		statusWidth := d.getDisplayWidth(statusText)
		indicatorWidth := d.getDisplayWidth(indicator)
		availableNameWidth := d.terminalWidth - indicatorWidth - d.indexColumnWidth - statusWidth - 3
		if availableNameWidth < 1 {
			availableNameWidth = 1
		}
		if len(state.DisplayName) > availableNameWidth {
			d.maxModelNameWidth = availableNameWidth
		} else {
			d.maxModelNameWidth = len(state.DisplayName)
		}
	}

	paddedID := fmt.Sprintf("%-*s", d.indexColumnWidth, modelId)
	modelName := state.DisplayName
	modelName = truncateWithEllipsis(modelName, d.maxModelNameWidth)
	paddedName := fmt.Sprintf("%-*s", d.maxModelNameWidth, modelName)

	// Color the model name
	coloredName := d.colors.ColorModelName(paddedName)

	left := fmt.Sprintf("%s %s %s", indicator, paddedID, coloredName)
	leftWidth := d.getDisplayWidth(left)
	statusWidth := d.getDisplayWidth(statusText)
	terminalWidth := d.terminalWidth
	if terminalWidth <= 0 {
		terminalWidth = leftWidth + statusWidth + 1
	}
	padding := terminalWidth - leftWidth - statusWidth
	if padding < 1 {
		padding = 1
	}

	return fmt.Sprintf("%s%s%s", left, strings.Repeat(" ", padding), statusText)
}

func truncateWithEllipsis(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	width := runewidth.StringWidth(text)
	if width <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	runes := []rune(text)
	for len(runes) > 0 {
		truncated := string(runes) + "..."
		if runewidth.StringWidth(truncated) <= maxWidth {
			return truncated
		}
		runes = runes[:len(runes)-1]
	}
	return strings.Repeat(".", maxWidth)
}

// formatIndicator creates a fixed-width leading indicator for the model state
func (d *StatusDisplay) formatIndicator(state *ModelState) string {
	switch state.Status {
	case StatusQueued:
		return d.formatProcessingIndicator()

	case StatusStarting:
		return d.formatProcessingIndicator()

	case StatusProcessing:
		return d.formatProcessingIndicator()

	case StatusRateLimited:
		return fmt.Sprintf("  %s ", d.colors.ColorWarning("⚠"))

	case StatusCompleted:
		return fmt.Sprintf("  %s ", d.colors.ColorSuccess("✓"))

	case StatusFailed:
		return fmt.Sprintf("  %s ", d.colors.ColorError("✗"))

	default:
		return "    "
	}
}

func (d *StatusDisplay) formatProcessingIndicator() string {
	if !d.isInteractive {
		return "  … "
	}

	d.mu.Lock()
	spinnerView := ""
	if d.spinnerTick != nil {
		spinnerView = d.spinner.View()
	}
	d.mu.Unlock()
	if spinnerView == "" {
		spinnerView = "…"
	}

	return fmt.Sprintf("  %s ", d.colors.ColorDuration(spinnerView))
}

func (d *StatusDisplay) formatStatusText(state *ModelState) string {
	switch state.Status {
	case StatusQueued:
		return d.colors.ColorDuration("queued")
	case StatusStarting:
		return d.colors.ColorInfo("starting")
	case StatusProcessing:
		return d.colors.ColorInfo("processing")
	case StatusRateLimited:
		return d.colors.ColorWarning(fmt.Sprintf("retry in %s", FormatDuration(state.RetryAfter)))
	case StatusCompleted:
		return d.colors.ColorDuration(fmt.Sprintf("completed (%s)", FormatDuration(state.Duration)))
	case StatusFailed:
		errorMsg := state.ErrorMsg
		if errorMsg == "" {
			errorMsg = "error"
		}
		return d.colors.ColorError(errorMsg)
	default:
		return "unknown"
	}
}

// RenderSummaryHeader displays initial processing information
func (d *StatusDisplay) RenderSummaryHeader(totalModels int) {
	fmt.Printf("\nProcessing %d models...\n", totalModels)
}

// RenderCompletion displays final completion message and clears status area.
// Caller is responsible for serialization.
func (d *StatusDisplay) RenderCompletion() {
	d.Stop()
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

// RenderPeriodicUpdate renders status in CI mode with headers.
// Caller is responsible for serialization. isInteractive is immutable after construction.
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

func (d *StatusDisplay) CalculateLayout(states []*ModelState) {
	totalModels := len(states)
	if totalModels < 1 {
		totalModels = 1
	}

	totalDigits := len(fmt.Sprintf("%d", totalModels))
	modelIdSample := fmt.Sprintf("[%0*d/%d]", totalDigits, totalModels, totalModels)
	d.indexColumnWidth = len(modelIdSample)

	maxNameWidth := 0
	maxStatusWidth := 0
	for _, state := range states {
		if len(state.DisplayName) > maxNameWidth {
			maxNameWidth = len(state.DisplayName)
		}
		statusWidth := d.getDisplayWidth(d.formatStatusText(state))
		if statusWidth > maxStatusWidth {
			maxStatusWidth = statusWidth
		}
	}

	indicatorWidth := d.getDisplayWidth("    ")
	availableNameWidth := d.terminalWidth - indicatorWidth - d.indexColumnWidth - maxStatusWidth - 3
	if availableNameWidth < 1 {
		availableNameWidth = 1
	}

	if maxNameWidth > availableNameWidth {
		maxNameWidth = availableNameWidth
	}
	d.maxModelNameWidth = maxNameWidth
}

// getDisplayWidth calculates the actual display width of a string, excluding ANSI color codes
func (d *StatusDisplay) getDisplayWidth(s string) int {
	cleaned := ansiPattern.ReplaceAllString(s, "")
	return runewidth.StringWidth(cleaned)
}

func (d *StatusDisplay) calculateProgress(states []*ModelState, totalModels int) (int, float64) {
	if totalModels == 0 {
		return 0, 0
	}

	completed := 0
	for _, state := range states {
		if state.Status == StatusCompleted || state.Status == StatusFailed {
			completed++
		}
	}

	return completed, float64(completed) / float64(totalModels)
}

type GridProgress struct {
	totalCells int
	colors     *ColorScheme
}

func NewGridProgress(colors *ColorScheme, totalCells int) *GridProgress {
	if totalCells < 1 {
		totalCells = 1
	}
	return &GridProgress{
		totalCells: totalCells,
		colors:     colors,
	}
}

func (g *GridProgress) SetTotalCells(totalCells int) {
	if totalCells < 1 {
		totalCells = 1
	}
	g.totalCells = totalCells
}

func (g *GridProgress) Render(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	filled := int(math.Floor(percent * float64(g.totalCells)))
	if filled > g.totalCells {
		filled = g.totalCells
	}
	empty := g.totalCells - filled

	// Use ASCII characters for non-interactive mode
	filledChar := "█"
	emptyChar := "░"
	if g.colors == nil {
		filledChar = "="
		emptyChar = "-"
	}

	filledCells := strings.Repeat(filledChar, filled)
	emptyCells := strings.Repeat(emptyChar, empty)

	if g.colors != nil {
		filledCells = g.colors.ApplyColor("#3B82F6", filledCells)
	}

	return filledCells + emptyCells
}

func progressWidth(terminalWidth int) int {
	width := terminalWidth - 16
	if width < 1 {
		return 1
	}
	return width
}
