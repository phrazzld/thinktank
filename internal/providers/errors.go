// Package providers defines the interfaces and implementations for LLM provider adapters
package providers

import "errors"

// Common errors for provider implementations
var (
	// ErrProviderNotFound is returned when a requested provider is not registered
	ErrProviderNotFound = errors.New("provider not found")

	// ErrInvalidAPIKey is returned when an API key is invalid or empty
	ErrInvalidAPIKey = errors.New("invalid API key")

	// ErrInvalidModelID is returned when a model ID is invalid or empty
	ErrInvalidModelID = errors.New("invalid model ID")

	// ErrInvalidEndpoint is returned when an API endpoint is invalid
	ErrInvalidEndpoint = errors.New("invalid API endpoint")

	// ErrClientCreation is returned when client creation fails
	ErrClientCreation = errors.New("failed to create client")
)
