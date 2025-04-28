// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"errors"
)

// Sentinel errors for synthesis and model processing failures.
// These error variables can be used with errors.Is() for reliable error type checking.
var (
	// ErrInvalidSynthesisModel is returned when a synthesis model is specified
	// but is invalid or cannot be initialized.
	ErrInvalidSynthesisModel = errors.New("invalid synthesis model specified")

	// ErrNoValidModels is returned when no valid model outputs are available
	// for processing or synthesis.
	ErrNoValidModels = errors.New("no valid models available for processing")

	// ErrPartialProcessingFailure is returned when some (but not all) models
	// fail during processing. The orchestrator continues with the successful models.
	ErrPartialProcessingFailure = errors.New("some models failed during processing")

	// ErrAllProcessingFailed is returned when all models fail to process.
	ErrAllProcessingFailed = errors.New("all models failed during processing")

	// ErrSynthesisFailed is returned when the synthesis step fails.
	ErrSynthesisFailed = errors.New("synthesis of model outputs failed")

	// ErrOutputFileSaveFailed is returned when there's an error saving output to a file.
	ErrOutputFileSaveFailed = errors.New("failed to save output to file")

	// ErrModelProcessingCancelled is returned when model processing is cancelled by context.
	ErrModelProcessingCancelled = errors.New("model processing cancelled")
)
