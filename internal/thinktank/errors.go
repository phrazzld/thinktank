// Package thinktank provides the core functionality for the thinktank application.
package thinktank

import (
	"errors"
)

// Sentinel errors for the thinktank package.
// These error variables can be used with errors.Is() for reliable error type checking.
var (
	// ErrInvalidConfiguration is returned when the provided configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrNoModelsProvided is returned when no models are specified.
	ErrNoModelsProvided = errors.New("no models provided")

	// ErrInvalidModelName is returned when a model name is invalid or not recognized.
	ErrInvalidModelName = errors.New("invalid model name")

	// ErrInvalidAPIKey is returned when the API key is missing or invalid.
	ErrInvalidAPIKey = errors.New("invalid or missing API key")

	// ErrInvalidInstructions is returned when instructions are missing or invalid.
	ErrInvalidInstructions = errors.New("invalid or missing instructions")

	// ErrInvalidOutputDir is returned when the output directory is invalid.
	ErrInvalidOutputDir = errors.New("invalid output directory")

	// ErrContextGatheringFailed is returned when context gathering fails.
	ErrContextGatheringFailed = errors.New("context gathering failed")
)
