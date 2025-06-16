// Package logutil provides console output functionality for user-facing progress
// and status reporting, complementing the structured logging infrastructure.
//
// The ConsoleWriter interface provides clean, human-readable output that adapts
// to different environments (interactive terminals vs CI/CD) while maintaining
// separation from structured logging concerns.
package logutil

import (
	"time"
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
