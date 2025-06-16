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
	"sync"
	"time"

	"golang.org/x/term"
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
	// The index parameter indicates the model's position in the processing order (1-based).
	//
	// This typically displays "[X/N] model-name: processing..." type messages.
	ModelStarted(modelName string, index int)

	// ModelCompleted reports that processing has finished for a specific model.
	// The index parameter indicates the model's position in the processing order (1-based).
	// The duration parameter specifies how long the processing took.
	// The err parameter indicates whether processing succeeded (nil) or failed (non-nil).
	//
	// Success displays: "[X/N] model-name: ✓ completed (1.2s)"
	// Failure displays: "[X/N] model-name: ✗ failed (error message)"
	ModelCompleted(modelName string, index int, duration time.Duration, err error)

	// ModelRateLimited reports that a model's processing has been delayed due to rate limiting.
	// The index parameter indicates the model's position in the processing order (1-based).
	// The delay parameter specifies how long the delay will be.
	//
	// This displays: "[X/N] model-name: rate limited, waiting 2s..."
	ModelRateLimited(modelName string, index int, delay time.Duration)

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
}

// consoleWriter is the concrete implementation of ConsoleWriter interface.
// It provides clean, human-readable console output that adapts to different
// execution environments (interactive terminals vs CI/CD pipelines).
type consoleWriter struct {
	mu            sync.Mutex // Protects concurrent access to all fields
	isInteractive bool       // Whether running in interactive terminal
	quiet         bool       // Whether to suppress non-essential output
	noProgress    bool       // Whether to suppress detailed progress indicators
	modelCount    int        // Total number of models to process
	modelIndex    int        // Current model index (for progress tracking)

	// Dependency injection for testing
	isTerminalFunc func() bool
}

// Ensure consoleWriter implements ConsoleWriter interface
var _ ConsoleWriter = (*consoleWriter)(nil)

// NewConsoleWriter creates a new ConsoleWriter with automatic environment detection.
// It detects whether running in an interactive terminal vs CI/CD environment
// and configures output accordingly.
func NewConsoleWriter() ConsoleWriter {
	return &consoleWriter{
		isTerminalFunc: defaultIsTerminal,
		isInteractive:  detectInteractiveEnvironment(defaultIsTerminal),
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

	return &consoleWriter{
		isTerminalFunc: isTerminalFunc,
		isInteractive:  detectInteractiveEnvironment(isTerminalFunc),
	}
}

// defaultIsTerminal uses golang.org/x/term to detect if stdout is a terminal
func defaultIsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// detectInteractiveEnvironment determines if we're running in an interactive
// environment based on TTY detection and CI environment variables
func detectInteractiveEnvironment(isTerminalFunc func() bool) bool {
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
		if os.Getenv(envVar) == "true" {
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

	if c.isInteractive {
		fmt.Printf("🚀 Processing %d models...\n", modelCount)
	} else {
		fmt.Printf("Starting processing with %d models\n", modelCount)
	}
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
func (c *consoleWriter) ModelStarted(modelName string, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	if c.noProgress {
		return
	}

	if c.isInteractive {
		fmt.Printf("[%d/%d] %s: processing...\n", index, c.modelCount, modelName)
	} else {
		fmt.Printf("Processing model %d/%d: %s\n", index, c.modelCount, modelName)
	}
}

// ModelCompleted reports that processing has finished for a specific model
func (c *consoleWriter) ModelCompleted(modelName string, index int, duration time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	durationStr := formatDuration(duration)

	if err != nil {
		// Always show errors, even in no-progress mode
		if c.isInteractive {
			fmt.Printf("[%d/%d] %s: ✗ failed (%s)\n", index, c.modelCount, modelName, err.Error())
		} else {
			fmt.Printf("Failed model %d/%d: %s (%s)\n", index, c.modelCount, modelName, err.Error())
		}
	} else {
		if c.noProgress {
			return
		}

		if c.isInteractive {
			fmt.Printf("[%d/%d] %s: ✓ completed (%s)\n", index, c.modelCount, modelName, durationStr)
		} else {
			fmt.Printf("Completed model %d/%d: %s (%s)\n", index, c.modelCount, modelName, durationStr)
		}
	}
}

// ModelRateLimited reports that a model's processing has been delayed due to rate limiting
func (c *consoleWriter) ModelRateLimited(modelName string, index int, delay time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.quiet {
		return
	}

	delayStr := formatDuration(delay)

	if c.isInteractive {
		fmt.Printf("[%d/%d] %s: rate limited, waiting %s...\n", index, c.modelCount, modelName, delayStr)
	} else {
		fmt.Printf("Rate limited for model %d/%d: %s (waiting %s)\n", index, c.modelCount, modelName, delayStr)
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

	if c.isInteractive {
		fmt.Printf("✨ Done! Output saved to: %s\n", outputPath)
	} else {
		fmt.Printf("Synthesis complete. Output: %s\n", outputPath)
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
}
