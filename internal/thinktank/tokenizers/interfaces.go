// Package tokenizers provides accurate token counting implementations for various LLM providers.
// This package enables precise token counting to replace estimation-based model selection.
package tokenizers

import (
	"context"
)

// AccurateTokenCounter provides model-specific accurate token counting.
// Implementations should handle provider-specific tokenization algorithms.
type AccurateTokenCounter interface {
	// CountTokens returns accurate token count for the given text and model.
	// Returns error if the model is not supported or tokenization fails.
	CountTokens(ctx context.Context, text string, modelName string) (int, error)

	// SupportsModel returns true if accurate counting is available for the model.
	SupportsModel(modelName string) bool

	// GetEncoding returns the tokenizer encoding name for the given model.
	// Returns error if the model is not supported.
	GetEncoding(modelName string) (string, error)
}

// TokenizerManager handles lazy loading and caching of tokenizers.
// Provides centralized access to all tokenizer implementations.
type TokenizerManager interface {
	// GetTokenizer returns a tokenizer for the given provider.
	// Handles lazy initialization and caching for performance.
	GetTokenizer(provider string) (AccurateTokenCounter, error)

	// SupportsProvider returns true if accurate tokenization is available for the provider.
	SupportsProvider(provider string) bool

	// ClearCache clears all cached tokenizers to free memory.
	ClearCache()
}

// TokenizerError represents errors from tokenization operations.
type TokenizerError struct {
	Provider string
	Model    string
	Message  string
	Cause    error
}

func (e *TokenizerError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *TokenizerError) Unwrap() error {
	return e.Cause
}

// NewTokenizerError creates a new tokenizer error with context.
func NewTokenizerError(provider, model, message string, cause error) *TokenizerError {
	return &TokenizerError{
		Provider: provider,
		Model:    model,
		Message:  message,
		Cause:    cause,
	}
}
