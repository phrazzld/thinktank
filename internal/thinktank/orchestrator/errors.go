// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"errors"

	"github.com/misty-step/thinktank/internal/llm"
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

// CategorizeOrchestratorError maps orchestrator errors to standard LLM error categories.
// This function is used to provide consistent error categorization across the application.
func CategorizeOrchestratorError(err error) llm.ErrorCategory {
	switch {
	case errors.Is(err, ErrInvalidSynthesisModel):
		return llm.CategoryInvalidRequest
	case errors.Is(err, ErrNoValidModels):
		return llm.CategoryInvalidRequest
	case errors.Is(err, ErrPartialProcessingFailure):
		// This is a partial failure, so we treat it as a server error
		// since some models may have failed due to server issues
		return llm.CategoryServer
	case errors.Is(err, ErrAllProcessingFailed):
		// This could be due to various reasons, default to server error
		return llm.CategoryServer
	case errors.Is(err, ErrSynthesisFailed):
		// Synthesis failures could be due to invalid inputs or server issues
		return llm.CategoryServer
	case errors.Is(err, ErrOutputFileSaveFailed):
		// File system errors are considered server-side issues
		return llm.CategoryServer
	case errors.Is(err, ErrModelProcessingCancelled):
		return llm.CategoryCancelled
	default:
		// If we can't identify the error, we check if it's already a LLMError
		if catErr, ok := llm.IsCategorizedError(err); ok {
			return catErr.Category()
		}
		return llm.CategoryUnknown
	}
}

// WrapOrchestratorError wraps an error with orchestrator-specific context.
// It uses llm.Wrap to ensure proper error categorization.
func WrapOrchestratorError(err error, message string) error {
	if err == nil {
		return nil
	}

	category := CategorizeOrchestratorError(err)
	return llm.Wrap(err, "orchestrator", message, category)
}
