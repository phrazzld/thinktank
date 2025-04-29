// Package modelproc provides model processing functionality for the thinktank tool.
// It encapsulates the logic for interacting with AI models, managing tokens,
// writing outputs, and logging operations.
package modelproc

import (
	"errors"
)

// Sentinel errors for model processing failures.
// These error variables can be used with errors.Is() for reliable error type checking.
var (
	// ErrModelProcessingFailed is a general error indicating model processing failed.
	ErrModelProcessingFailed = errors.New("model processing failed")

	// ErrModelInitializationFailed is returned when the model client cannot be initialized.
	ErrModelInitializationFailed = errors.New("model initialization failed")

	// ErrInvalidModelResponse is returned when a model response cannot be processed.
	ErrInvalidModelResponse = errors.New("invalid or malformed model response")

	// ErrEmptyModelResponse is returned when a model returns an empty response.
	ErrEmptyModelResponse = errors.New("model returned empty response")

	// ErrContentFiltered is returned when content is blocked by safety filters.
	ErrContentFiltered = errors.New("content blocked by safety filters")

	// ErrModelRateLimited is returned when API requests are rate limited.
	ErrModelRateLimited = errors.New("model API rate limited")

	// ErrModelTokenLimitExceeded is returned when the input exceeds the model's token limit.
	ErrModelTokenLimitExceeded = errors.New("model token limit exceeded")

	// ErrOutputWriteFailed is returned when the output cannot be written to the filesystem.
	ErrOutputWriteFailed = errors.New("failed to write model output to file")
)
