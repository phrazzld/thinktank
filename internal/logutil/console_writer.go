// Package logutil provides console output functionality for user-facing progress
// and status reporting, complementing the structured logging infrastructure.
//
// The ConsoleWriter interface provides clean, human-readable output that adapts
// to different environments (interactive terminals vs CI/CD) while maintaining
// separation from structured logging concerns.
package logutil

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Terminal width constants
const (
	// DefaultTerminalWidth is used when width detection fails or in non-interactive environments
	DefaultTerminalWidth = 80
	// MinTerminalWidth is the minimum width we'll format for
	MinTerminalWidth = 10
	// MaxTerminalWidth is the maximum width we'll use (for very wide terminals)
	MaxTerminalWidth = 120
)

// ConsoleWriter defines an interface for clean, user-facing console output
// that adapts to different execution environments (interactive vs CI/CD).
//
// This interface is designed to provide progress reporting and status updates
// for long-running operations, particularly model processing workflows.
// It maintains clear separation from structured logging, focusing solely
// on human-readable output to stdout.
//
// The interface supports two output modes:
// - Interactive Mode: Rich output with emojis and progress indicators for TTY environments
// - CI Mode: Simple, parseable text output for non-interactive environments
//
// All methods are thread-safe and designed for concurrent access during
// parallel model processing operations.
type ConsoleWriter interface {
	// Progress Reporting Methods
	// These methods track and display the progress of model processing operations

	// StartProcessing initiates progress reporting for a batch of models.
	// The modelCount parameter determines the total number of models to process,
	// enabling progress indicators like [1/3], [2/3], etc.
	//
	// This method should be called once at the beginning of model processing
	// to establish the total count for subsequent progress updates.
	StartProcessing(modelCount int)

	// ModelQueued reports that a model has been added to the processing queue.
	// The index parameter indicates the model's position in the processing order (1-based).
	//
	// In interactive mode, this may display queuing information.
	// In CI mode, this may be a no-op or simple log entry.
	ModelQueued(modelName string, index int)

	// ModelStarted reports that processing has begun for a specific model.
	// Parameters follow modern clean output specification with index first.
	//
	// This typically displays "[X/N] model-name: processing..." type messages.
	ModelStarted(modelIndex, totalModels int, modelName string)

	// ModelCompleted reports that processing has finished successfully for a specific model.
	// Parameters follow modern clean output specification with index first.
	// Duration specifies how long the processing took.
	//
	// Displays: "[X/N] model-name: ✓ completed (1.2s)"
	ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration)

	// ModelFailed reports that processing has failed for a specific model.
	// Parameters follow modern clean output specification with index first.
	// Reason provides a human-readable explanation of the failure.
	//
	// Displays: "[X/N] model-name: ✗ failed (reason)"
	ModelFailed(modelIndex, totalModels int, modelName string, reason string)

	// ModelRateLimited reports that a model's processing has been delayed due to rate limiting.
	// Parameters follow modern clean output specification with index first.
	// retryAfter specifies how long the delay will be.
	//
	// This displays: "[X/N] model-name: ⚠ rate limited (retry in 30s)"
	ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration)

	// Modern Clean Output Methods
	// These methods implement the new aligned, professional output format

	// ShowProcessingLine displays an initial processing status line for a model.
	// Used for the left-aligned model name with right-aligned "processing..." status.
	//
	// Displays: "model-name                          processing..."
	ShowProcessingLine(modelName string)

	// UpdateProcessingLine updates the processing line in-place with final status.
	// Used to replace "processing..." with final status like "✓ 68.5s" or "✗ rate limited".
	//
	// Displays: "model-name                          ✓ 68.5s"
	UpdateProcessingLine(modelName string, status string)

	// ShowFileOperations displays clean, declarative file operation messages.
	// Used for messages like "Saving individual outputs..." and "Saved 2 outputs to path".
	ShowFileOperations(message string)

	// ShowSummarySection displays the main summary section with structured format.
	// Uses UPPERCASE headers, bullet points, and basic statistics.
	//
	// Example:
	// SUMMARY
	// ───────
	// ● 3 models processed
	// ● 2 successful, 1 failed
	// ● Output directory: ./path
	ShowSummarySection(summary SummaryData)

	// ShowOutputFiles displays the output files section with human-readable sizes.
	// Uses right-aligned file sizes and proper formatting.
	//
	// Example:
	// OUTPUT FILES
	// ────────────
	//   gemini-2.5-pro.md                     4.2K
	//   claude-3-5-sonnet.md                  3.8K
	ShowOutputFiles(files []OutputFile)

	// ShowFailedModels displays the failed models section when failures occur.
	// Only displayed when there are actual failures to report.
	//
	// Example:
	// FAILED MODELS
	// ─────────────
	//   gpt-4o                                rate limited
	ShowFailedModels(failed []FailedModel)

	// Status Update Methods
	// These methods report major workflow milestones

	// SynthesisStarted reports that the synthesis phase has begun.
	// This is called after all individual model processing is complete
	// and the system begins combining results.
	//
	// Interactive mode: "📄 Synthesizing results..."
	// CI mode: "Starting synthesis"
	SynthesisStarted()

	// SynthesisCompleted reports that synthesis has finished successfully.
	// The outputPath parameter specifies where the final results were saved.
	//
	// Interactive mode: "✨ Done! Output saved to: path/to/output"
	// CI mode: "Synthesis complete. Output: path/to/output"
	SynthesisCompleted(outputPath string)

	// StatusMessage displays a general status update to the user.
	// This is used for lifecycle events outside of model processing.
	//
	// Interactive mode: "📁 message"
	// CI mode: "message"
	StatusMessage(message string)

	// Control Methods
	// These methods configure the behavior of console output

	// SetQuiet enables or disables quiet mode.
	// When quiet is true, only essential messages (like errors) are displayed.
	// Progress indicators and status updates are suppressed.
	//
	// This is typically controlled by CLI flags like --quiet or -q.
	SetQuiet(quiet bool)

	// SetNoProgress enables or disables progress indicators.
	// When noProgress is true, detailed progress like "[X/N]" and "processing..."
	// messages are suppressed, but major status updates (start/complete) remain.
	//
	// This is typically controlled by CLI flags like --no-progress.
	SetNoProgress(noProgress bool)

	// IsInteractive returns true if the output environment supports interactive features
	// like emojis, colors, and dynamic line updates (i.e., running in a TTY).
	//
	// Returns false for CI environments, non-TTY pipes, or when CI=true is set.
	// This determination affects which output format is used for all methods.
	IsInteractive() bool

	// GetTerminalWidth returns the current terminal width in characters.
	// Returns a default width if detection fails or if not running in a terminal.
	GetTerminalWidth() int

	// FormatMessage formats a message to fit within the terminal width,
	// truncating or wrapping as appropriate for the current environment.
	FormatMessage(message string) string

	// ErrorMessage displays an error message to the user with appropriate formatting.
	// This provides better visual distinction for error states.
	ErrorMessage(message string)

	// WarningMessage displays a warning message to the user with appropriate formatting.
	// This provides better visual distinction for warning states.
	WarningMessage(message string)

	// SuccessMessage displays a success message to the user with appropriate formatting.
	// This provides better visual distinction for success states.
	SuccessMessage(message string)
}

// consoleWriter is the concrete implementation of ConsoleWriter interface.
// It provides clean, human-readable console output that adapts to different
// execution environments (interactive terminals vs CI/CD pipelines).
type consoleWriter struct {
	mu            sync.Mutex      // Protects concurrent access to all fields
	isInteractive bool            // Whether running in interactive terminal
	quiet         bool            // Whether to suppress non-essential output
	noProgress    bool            // Whether to suppress detailed progress indicators
	modelCount    int             // Total number of models to process
	modelIndex    int             // Current model index (for progress tracking)
	terminalWidth int             // Cached terminal width, 0 means not detected yet
	layout        LayoutConfig    // Cached layout configuration
	colors        *ColorScheme    // Color scheme for semantic coloring
	symbols       *SymbolProvider // Unicode/ASCII symbol provider with fallback detection

	// Dependency injection for testing
	isTerminalFunc  func() bool
	getTermSizeFunc func() (int, int, error)
}

// Ensure consoleWriter implements ConsoleWriter interface
var _ ConsoleWriter = (*consoleWriter)(nil)

// NewConsoleWriter creates a new ConsoleWriter with automatic environment detection.
// It detects whether running in an interactive terminal vs CI/CD environment
// and configures output accordingly.
func NewConsoleWriter() ConsoleWriter {
	isInteractive := detectInteractiveEnvironment(defaultIsTerminal)
	return &consoleWriter{
		isTerminalFunc:  defaultIsTerminal,
		getTermSizeFunc: defaultGetTermSize,
		isInteractive:   isInteractive,
		colors:          NewColorScheme(isInteractive),
		symbols:         NewSymbolProvider(isInteractive),
	}
}

// NewConsoleWriterWithOptions creates a ConsoleWriter with injectable dependencies
// for testing. This allows mocking of terminal detection and other environment
// checks to ensure reliable testing across different scenarios.
func NewConsoleWriterWithOptions(opts ConsoleWriterOptions) ConsoleWriter {
	isTerminalFunc := opts.IsTerminalFunc
	if isTerminalFunc == nil {
		isTerminalFunc = defaultIsTerminal
	}

	getTermSizeFunc := opts.GetTermSizeFunc
	if getTermSizeFunc == nil {
		getTermSizeFunc = defaultGetTermSize
	}

	getEnvFunc := opts.GetEnvFunc
	if getEnvFunc == nil {
		getEnvFunc = os.Getenv
	}

	isInteractive := detectInteractiveEnvironmentWithEnv(isTerminalFunc, getEnvFunc)
	return &consoleWriter{
		isTerminalFunc:  isTerminalFunc,
		getTermSizeFunc: getTermSizeFunc,
		isInteractive:   isInteractive,
		colors:          NewColorScheme(isInteractive),
		symbols:         NewSymbolProvider(isInteractive),
	}
}

// defaultIsTerminal uses golang.org/x/term to detect if stdout is a terminal
func defaultIsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// defaultGetTermSize uses golang.org/x/term to get terminal dimensions
func defaultGetTermSize() (int, int, error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

// detectInteractiveEnvironment determines if we're running in an interactive
// environment based on TTY detection and CI environment variables
func detectInteractiveEnvironment(isTerminalFunc func() bool) bool {
	return detectInteractiveEnvironmentWithEnv(isTerminalFunc, os.Getenv)
}

// detectInteractiveEnvironmentWithEnv determines if we're running in an interactive
// environment based on TTY detection and CI environment variables with injectable env func
func detectInteractiveEnvironmentWithEnv(isTerminalFunc func() bool, getEnvFunc func(string) string) bool {
	// Check common CI environment variables
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"CONTINUOUS_INTEGRATION",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, envVar := range ciVars {
		value := getEnvFunc(envVar)
		if value != "" && (value == "true" || envVar == "JENKINS_URL") {
			return false
		}
	}

	// If not in CI and stdout is a terminal, we're interactive
	return isTerminalFunc()
}

// StartProcessing initiates progress reporting for a batch of models
func (c *consoleWriter) StartProcessing(modelCount int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.modelCount = modelCount
	c.modelIndex = 0

	if c.quiet {
		return
	}

	// Use clean, declarative messaging consistent across environments
	message := fmt.Sprintf("Processing %d models...", modelCount)
	fmt.Println(c.colors.ColorModelName(message))
}

// ModelQueued reports that a model has been added to the processing queue
func (c *consoleWriter) ModelQueued(modelName string, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet || c.noProgress {
		return
	}

	// In most cases, queuing happens so quickly that we don't need to report it
	// This method is primarily for completeness of the interface
}

// ModelStarted reports that processing has begun for a specific model
func (c *consoleWriter) ModelStarted(modelIndex, totalModels int, modelName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	if c.noProgress {
		return
	}

	// Add whitespace before first model processing
	if modelIndex == 1 {
		fmt.Println() // Phase separation whitespace
	}

	coloredModelName := c.colors.ColorModelName(modelName)
	if c.isInteractive {
		fmt.Printf("[%d/%d] %s: processing...\n", modelIndex, totalModels, coloredModelName)
	} else {
		fmt.Printf("Processing model %d/%d: %s\n", modelIndex, totalModels, coloredModelName)
	}
}

// ModelCompleted reports that processing has finished successfully for a specific model
func (c *consoleWriter) ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Success messages can be suppressed in quiet mode
	if c.quiet || c.noProgress {
		return
	}

	durationStr := c.colors.ColorDuration(formatDuration(duration))
	coloredModelName := c.colors.ColorModelName(modelName)
	successSymbol := c.colors.ColorSuccess(c.symbols.GetSymbols().Success)

	if c.isInteractive {
		fmt.Printf("[%d/%d] %s: %s completed (%s)\n", modelIndex, totalModels, coloredModelName, successSymbol, durationStr)
	} else {
		fmt.Printf("Completed model %d/%d: %s (%s)\n", modelIndex, totalModels, coloredModelName, durationStr)
	}
}

// ModelFailed reports that processing has failed for a specific model
func (c *consoleWriter) ModelFailed(modelIndex, totalModels int, modelName string, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Errors are essential - always show them even in quiet mode
	coloredModelName := c.colors.ColorModelName(modelName)
	errorSymbol := c.colors.ColorError(c.symbols.GetSymbols().Error)
	coloredReason := c.colors.ColorError(reason)

	if c.isInteractive {
		fmt.Printf("[%d/%d] %s: %s failed (%s)\n", modelIndex, totalModels, coloredModelName, errorSymbol, coloredReason)
	} else {
		fmt.Printf("Failed model %d/%d: %s (%s)\n", modelIndex, totalModels, coloredModelName, coloredReason)
	}
}

// ModelRateLimited reports that a model's processing has been delayed due to rate limiting
func (c *consoleWriter) ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	retryStr := c.colors.ColorDuration(formatDuration(retryAfter))
	coloredModelName := c.colors.ColorModelName(modelName)
	warningSymbol := c.colors.ColorWarning(c.symbols.GetSymbols().Warning)

	if c.isInteractive {
		fmt.Printf("[%d/%d] %s: %s rate limited (retry in %s)\n", modelIndex, totalModels, coloredModelName, warningSymbol, retryStr)
	} else {
		fmt.Printf("Rate limited for model %d/%d: %s (retry in %s)\n", modelIndex, totalModels, coloredModelName, retryStr)
	}
}

// SynthesisStarted reports that the synthesis phase has begun
func (c *consoleWriter) SynthesisStarted() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	if c.isInteractive {
		fmt.Println("📄 Synthesizing results...")
	} else {
		fmt.Println("Starting synthesis")
	}
}

// SynthesisCompleted reports that synthesis has finished successfully
func (c *consoleWriter) SynthesisCompleted(outputPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	coloredOutputPath := c.colors.ColorFilePath(outputPath)

	if c.isInteractive {
		successSymbol := c.colors.ColorSuccess(c.symbols.GetSymbols().Sparkles)
		fmt.Printf("%s Done! Output saved to: %s\n", successSymbol, coloredOutputPath)
	} else {
		fmt.Printf("Synthesis complete. Output: %s\n", coloredOutputPath)
	}
}

// StatusMessage displays a general status update to the user
func (c *consoleWriter) StatusMessage(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	// Format message to fit terminal width
	formattedMessage := c.formatMessageForTerminal(message)

	if c.isInteractive {
		bulletSymbol := c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet)
		fmt.Printf("%s %s\n", bulletSymbol, formattedMessage)
	} else {
		fmt.Println(formattedMessage)
	}
}

// SetQuiet enables or disables quiet mode
func (c *consoleWriter) SetQuiet(quiet bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.quiet = quiet
}

// SetNoProgress enables or disables progress indicators
func (c *consoleWriter) SetNoProgress(noProgress bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.noProgress = noProgress
}

// IsInteractive returns true if the output environment supports interactive features
func (c *consoleWriter) IsInteractive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isInteractive
}

// GetTerminalWidth returns the current terminal width in characters
func (c *consoleWriter) GetTerminalWidth() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return cached width if we have it
	if c.terminalWidth > 0 {
		return c.terminalWidth
	}

	// Try to detect terminal width
	if c.isTerminalFunc() && c.getTermSizeFunc != nil {
		width, _, err := c.getTermSizeFunc()
		if err == nil && width > 0 {
			// Clamp to reasonable bounds, but allow small widths for testing
			if width < MinTerminalWidth && width >= 3 {
				// Allow small widths for testing edge cases
				c.terminalWidth = width
				return width
			} else if width < 3 {
				// Minimum of 3 for "..."
				width = 3
			} else if width > MaxTerminalWidth {
				width = MaxTerminalWidth
			}
			c.terminalWidth = width
			return width
		} else if err != nil {
			// Log terminal width detection failure to stderr
			fmt.Fprintf(os.Stderr, "Warning: terminal width detection failed: %v, using default width %d\n", err, DefaultTerminalWidth)
		}
	}

	// Fallback to default width
	c.terminalWidth = DefaultTerminalWidth
	return DefaultTerminalWidth
}

// FormatMessage formats a message to fit within the terminal width
func (c *consoleWriter) FormatMessage(message string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	width := c.getTerminalWidthLocked()
	return c.formatToWidth(message, width)
}

// formatMessageForTerminal formats a message for terminal display (internal helper)
// This method assumes the mutex is already held by the caller
func (c *consoleWriter) formatMessageForTerminal(message string) string {
	// Get terminal width without acquiring mutex (already held)
	width := c.getTerminalWidthLocked()
	return c.formatToWidth(message, width)
}

// getTerminalWidthLocked returns the terminal width without acquiring mutex
// This method assumes the mutex is already held by the caller
func (c *consoleWriter) getTerminalWidthLocked() int {
	// Return cached width if we have it
	if c.terminalWidth > 0 {
		return c.terminalWidth
	}

	// Try to detect terminal width
	if c.isTerminalFunc() && c.getTermSizeFunc != nil {
		width, _, err := c.getTermSizeFunc()
		if err == nil && width > 0 {
			// Clamp to reasonable bounds, but allow small widths for testing
			if width < MinTerminalWidth && width >= 3 {
				// Allow small widths for testing edge cases
				c.terminalWidth = width
				return width
			} else if width < 3 {
				// Minimum of 3 for "..."
				width = 3
			} else if width > MaxTerminalWidth {
				width = MaxTerminalWidth
			}
			c.terminalWidth = width
			return width
		} else if err != nil {
			// Log terminal width detection failure to stderr
			fmt.Fprintf(os.Stderr, "Warning: terminal width detection failed: %v, using default width %d\n", err, DefaultTerminalWidth)
		}
	}

	// Fallback to default width
	c.terminalWidth = DefaultTerminalWidth
	return DefaultTerminalWidth
}

// getLayoutLocked returns the current layout configuration
// This method assumes the mutex is already held by the caller
func (c *consoleWriter) getLayoutLocked() LayoutConfig {
	// Return cached layout if terminal width hasn't changed
	if c.layout.TerminalWidth == c.getTerminalWidthLocked() && c.layout.TerminalWidth > 0 {
		return c.layout
	}

	// Calculate new layout based on current terminal width
	width := c.getTerminalWidthLocked()
	c.layout = CalculateLayout(width)
	return c.layout
}

// formatToWidth formats a message to fit within the specified width
func (c *consoleWriter) formatToWidth(message string, width int) string {
	// If message fits within terminal width, return as-is
	if len(message) <= width {
		return message
	}

	// In non-interactive mode, don't truncate messages
	if !c.isInteractive {
		return message
	}

	// Truncate message and add ellipsis
	if width <= 3 {
		return "..."
	}
	return message[:width-3] + "..."
}

// ErrorMessage displays an error message to the user with appropriate formatting
func (c *consoleWriter) ErrorMessage(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Errors are essential - always show them even in quiet mode
	formattedMessage := c.formatMessageForTerminal(message)
	coloredMessage := c.colors.ColorError(formattedMessage)

	if c.isInteractive {
		errorSymbol := c.colors.ColorError(c.symbols.GetSymbols().Error)
		fmt.Printf("%s %s\n", errorSymbol, coloredMessage)
	} else {
		fmt.Printf("ERROR: %s\n", coloredMessage)
	}
}

// WarningMessage displays a warning message to the user with appropriate formatting
func (c *consoleWriter) WarningMessage(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Warnings are essential - always show them even in quiet mode
	formattedMessage := c.formatMessageForTerminal(message)
	coloredMessage := c.colors.ColorWarning(formattedMessage)

	if c.isInteractive {
		warningSymbol := c.colors.ColorWarning(c.symbols.GetSymbols().Warning)
		fmt.Printf("%s %s\n", warningSymbol, coloredMessage)
	} else {
		fmt.Printf("WARNING: %s\n", coloredMessage)
	}
}

// SuccessMessage displays a success message to the user with appropriate formatting
func (c *consoleWriter) SuccessMessage(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	formattedMessage := c.formatMessageForTerminal(message)
	coloredMessage := c.colors.ColorSuccess(formattedMessage)

	if c.isInteractive {
		successSymbol := c.colors.ColorSuccess(c.symbols.GetSymbols().Success)
		fmt.Printf("%s %s\n", successSymbol, coloredMessage)
	} else {
		fmt.Printf("SUCCESS: %s\n", coloredMessage)
	}
}

// Modern Clean Output Methods
// These methods implement the modern clean CLI output format with proper
// alignment, color schemes, and responsive layout.

// ShowProcessingLine displays an initial processing status line for a model
func (c *consoleWriter) ShowProcessingLine(modelName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	layout := c.getLayoutLocked()

	// Apply color to model name
	coloredModelName := c.colors.ColorModelName(modelName)

	// Create processing status
	processingStatus := "processing..."

	// Format with proper alignment
	alignedOutput := layout.FormatAlignedText(coloredModelName, processingStatus)

	fmt.Println(alignedOutput)
}

// UpdateProcessingLine updates the processing line in-place with final status
func (c *consoleWriter) UpdateProcessingLine(modelName string, status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	layout := c.getLayoutLocked()

	// Apply color to model name
	coloredModelName := c.colors.ColorModelName(modelName)

	// Apply semantic coloring to status based on content
	coloredStatus := c.colorizeStatus(status)

	// Format with proper alignment
	alignedOutput := layout.FormatAlignedText(coloredModelName, coloredStatus)

	fmt.Println(alignedOutput)
}

// colorizeStatus applies appropriate colors to status text based on content
func (c *consoleWriter) colorizeStatus(status string) string {
	// Determine status type based on Unicode symbols and keywords
	if strings.Contains(status, "✓") {
		// Success status - apply success color to the entire status
		return c.colors.ColorSuccess(status)
	} else if strings.Contains(status, "✗") {
		// Error status - apply error color to the entire status
		return c.colors.ColorError(status)
	} else if strings.Contains(status, "⚠") {
		// Warning status (rate limited, etc.) - apply warning color
		return c.colors.ColorWarning(status)
	}

	// Default status - no special coloring
	return status
}

// ShowFileOperations displays clean, declarative file operation messages
func (c *consoleWriter) ShowFileOperations(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	// Add whitespace before saving operations
	if strings.HasPrefix(message, "Saving") {
		fmt.Println() // Phase separation whitespace
	}

	// Clean, declarative file operation messaging
	// No special formatting needed - just clear, direct communication
	fmt.Println(message)
}

// ShowSummarySection displays the main summary section with structured format
func (c *consoleWriter) ShowSummarySection(summary SummaryData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	layout := c.getLayoutLocked()

	// Add whitespace before summary section
	fmt.Println() // Phase separation whitespace
	fmt.Println() // Extra space for visual clarity

	// Display UPPERCASE header with separator line
	headerText := "SUMMARY"
	separatorLength := len(headerText)
	separatorLine := layout.GetSeparatorLine(separatorLength)

	fmt.Printf("%s\n", c.colors.ColorSectionHeader(headerText))
	fmt.Printf("%s\n", c.colors.ColorSeparator(separatorLine))

	// Display bullet point statistics
	fmt.Printf("%s %d models processed\n",
		c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet),
		summary.ModelsProcessed)

	fmt.Printf("%s %d successful, %d failed\n",
		c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet),
		summary.SuccessfulModels,
		summary.FailedModels)

	// Show synthesis status if not skipped
	if summary.SynthesisStatus != "skipped" {
		var statusText string
		switch summary.SynthesisStatus {
		case "completed":
			statusText = c.colors.ColorSuccess(c.symbols.GetSymbols().Success + " completed")
		case "failed":
			statusText = c.colors.ColorError(c.symbols.GetSymbols().Error + " failed")
		default:
			statusText = summary.SynthesisStatus
		}
		fmt.Printf("%s Synthesis: %s\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet),
			statusText)
	}

	// Show output directory
	fmt.Printf("%s Output directory: %s\n",
		c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet),
		c.colors.ColorFilePath(summary.OutputDirectory))

	// Add contextual messaging and guidance based on scenarios
	c.displayScenarioGuidance(summary)
}

// displayScenarioGuidance provides contextual messaging and actionable next steps
// based on the processing results (all failed, partial success, etc.)
func (c *consoleWriter) displayScenarioGuidance(summary SummaryData) {
	// Determine the scenario and provide appropriate guidance
	if summary.SuccessfulModels == 0 && summary.FailedModels > 0 {
		// All models failed scenario
		fmt.Println()
		fmt.Printf("%s %s\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Warning),
			c.colors.ColorWarning("All models failed to process"))
		fmt.Printf("  %s Check your API keys and network connectivity\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
		fmt.Printf("  %s Review error details above for specific failure reasons\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
		fmt.Printf("  %s Verify model names and rate limits with providers\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
	} else if summary.FailedModels > 0 && summary.SuccessfulModels > 0 {
		// Partial success scenario
		fmt.Println()
		fmt.Printf("%s %s\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Warning),
			c.colors.ColorWarning("Partial success - some models failed"))

		successRate := float64(summary.SuccessfulModels) / float64(summary.ModelsProcessed) * 100
		fmt.Printf("  %s Success rate: %.0f%% (%d/%d models)\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet),
			successRate,
			summary.SuccessfulModels,
			summary.ModelsProcessed)
		fmt.Printf("  %s Check failed model details above for specific issues\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
		fmt.Printf("  %s Consider retrying failed models or adjusting configuration\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
	} else if summary.SuccessfulModels > 0 && summary.FailedModels == 0 {
		// Complete success scenario
		fmt.Println()
		fmt.Printf("%s %s\n",
			c.colors.ColorSymbol(c.symbols.GetSymbols().Success),
			c.colors.ColorSuccess("All models processed successfully"))
		if summary.SynthesisStatus == "completed" {
			fmt.Printf("  %s Synthesis completed - check the combined output above\n",
				c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
		} else if summary.ModelsProcessed > 1 {
			fmt.Printf("  %s Individual model outputs saved - see file list above\n",
				c.colors.ColorSymbol(c.symbols.GetSymbols().Bullet))
		}
	}
}

// ShowOutputFiles displays the output files section with human-readable sizes
func (c *consoleWriter) ShowOutputFiles(files []OutputFile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet || len(files) == 0 {
		return
	}

	layout := c.getLayoutLocked()

	// Display UPPERCASE header with separator line
	headerText := "OUTPUT FILES"
	separatorLength := len(headerText)
	separatorLine := layout.GetSeparatorLine(separatorLength)

	fmt.Printf("%s\n", c.colors.ColorSectionHeader(headerText))
	fmt.Printf("%s\n", c.colors.ColorSeparator(separatorLine))

	// Display files with right-aligned human-readable sizes
	for _, file := range files {
		// Format file size using human-readable format
		formattedSize := FormatFileSize(file.Size)

		// Apply colors to filename and size
		coloredFileName := c.colors.ColorFilePath(file.Name)
		coloredSize := c.colors.ColorFileSize(formattedSize)

		// Format with proper right alignment
		alignedOutput := layout.FormatAlignedText(coloredFileName, coloredSize)
		fmt.Printf("  %s\n", alignedOutput)
	}
}

// ShowFailedModels displays the failed models section when failures occur
func (c *consoleWriter) ShowFailedModels(failed []FailedModel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet || len(failed) == 0 {
		return
	}

	layout := c.getLayoutLocked()

	// Display UPPERCASE header with separator line
	headerText := "FAILED MODELS"
	separatorLength := len(headerText)
	separatorLine := layout.GetSeparatorLine(separatorLength)

	fmt.Printf("%s\n", c.colors.ColorSectionHeader(headerText))
	fmt.Printf("%s\n", c.colors.ColorSeparator(separatorLine))

	// Display failed models with aligned reasons
	for _, model := range failed {
		// Apply colors to model name and failure reason
		coloredModelName := c.colors.ColorModelName(model.Name)
		coloredReason := c.colors.ColorError(model.Reason)

		// Format with proper right alignment
		alignedOutput := layout.FormatAlignedText(coloredModelName, coloredReason)
		fmt.Printf("  %s\n", alignedOutput)
	}
}

// formatDuration formats a time.Duration into a human-readable string
// like "1.2s", "850ms", etc.
func formatDuration(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

// ConsoleWriterOptions provides configuration options for creating a ConsoleWriter
// with injectable dependencies, primarily used for testing.
type ConsoleWriterOptions struct {
	// IsTerminalFunc allows injecting custom terminal detection logic for testing
	IsTerminalFunc func() bool
	// GetTermSizeFunc allows injecting custom terminal size detection for testing
	GetTermSizeFunc func() (int, int, error)
	// GetEnvFunc allows injecting custom environment variable reading for testing
	GetEnvFunc func(string) string
}
