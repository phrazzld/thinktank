// Package tokenizers provides accurate token counting using tiktoken o200k_base.
// This package enables precise token counting for model selection decisions.
package tokenizers

import (
	"context"
	"errors"
	"strings"

	"github.com/misty-step/thinktank/internal/llm"
)

// AccurateTokenCounter provides model-specific accurate token counting.
// All OpenRouter models are normalized to o200k_base encoding.
type AccurateTokenCounter interface {
	// CountTokens returns accurate token count for the given text and model.
	CountTokens(ctx context.Context, text string, modelName string) (int, error)

	// SupportsModel returns true if accurate counting is available for the model.
	SupportsModel(modelName string) bool

	// GetEncoding returns the tokenizer encoding name for the given model.
	GetEncoding(modelName string) (string, error)
}

// TokenizerManager handles lazy loading and caching of tokenizers.
// Thread-safe for concurrent access.
type TokenizerManager interface {
	// GetTokenizer returns a tokenizer for the given provider.
	GetTokenizer(provider string) (AccurateTokenCounter, error)

	// SupportsProvider returns true if accurate tokenization is available for the provider.
	SupportsProvider(provider string) bool

	// ClearCache clears all cached tokenizers to free memory.
	ClearCache()
}

// TokenizerError represents errors from tokenization operations.
type TokenizerError struct {
	Provider      string
	Model         string
	Message       string
	Cause         error
	ErrorCategory llm.ErrorCategory
	Details       string
}

func (e *TokenizerError) Error() string {
	var errorMsg string
	if e.Cause != nil {
		errorMsg = e.Message + ": " + e.Cause.Error()
	} else {
		errorMsg = e.Message
	}

	if e.Details != "" {
		errorMsg += " (" + e.Details + ")"
	}

	return errorMsg
}

func (e *TokenizerError) Unwrap() error {
	return e.Cause
}

// Category implements the CategorizedError interface.
func (e *TokenizerError) Category() llm.ErrorCategory {
	return e.ErrorCategory
}

// NewTokenizerError creates a new tokenizer error with context.
func NewTokenizerError(provider, model, message string, cause error) *TokenizerError {
	category := categorizeTokenizerError(provider, model, message, cause)
	details := buildErrorDetails(provider, model, cause)

	return &TokenizerError{
		Provider:      provider,
		Model:         model,
		Message:       message,
		Cause:         cause,
		ErrorCategory: category,
		Details:       details,
	}
}

// NewTokenizerErrorWithCategory creates a new tokenizer error with explicit category.
func NewTokenizerErrorWithCategory(provider, model, message string, cause error, category llm.ErrorCategory) *TokenizerError {
	details := buildErrorDetails(provider, model, cause)

	return &TokenizerError{
		Provider:      provider,
		Model:         model,
		Message:       message,
		Cause:         cause,
		ErrorCategory: category,
		Details:       details,
	}
}

// NewTokenizerErrorWithDetails creates a new tokenizer error with tokenizer type info.
func NewTokenizerErrorWithDetails(provider, model, message string, cause error, tokenizerType string) *TokenizerError {
	category := categorizeTokenizerError(provider, model, message, cause)
	details := buildEnhancedErrorDetails(provider, model, cause, tokenizerType)

	return &TokenizerError{
		Provider:      provider,
		Model:         model,
		Message:       message,
		Cause:         cause,
		ErrorCategory: category,
		Details:       details,
	}
}

// categorizeTokenizerError determines the error category based on error context.
// Checks both cause (using errors.Is for wrapped errors) and message content.
func categorizeTokenizerError(provider, model, message string, cause error) llm.ErrorCategory {
	// Check for context cancellation/timeout using errors.Is (handles wrapped errors)
	if cause != nil {
		if errors.Is(cause, context.DeadlineExceeded) || errors.Is(cause, context.Canceled) {
			return llm.CategoryCancelled
		}
	}

	// Build searchable text from message and cause
	var messageAndCause string
	if cause != nil {
		messageAndCause = strings.ToLower(message + " " + cause.Error())
	} else {
		messageAndCause = strings.ToLower(message)
	}

	if strings.Contains(messageAndCause, "circuit breaker") {
		return llm.CategoryRateLimit
	}
	if containsAny(messageAndCause, "initialization", "auth", "api key", "credential") {
		return llm.CategoryAuth
	}
	if containsAny(messageAndCause, "network", "connection", "dial", "timeout", "unreachable") {
		return llm.CategoryNetwork
	}
	if containsAny(messageAndCause, "invalid", "malformed", "encoding failed") {
		return llm.CategoryInvalidRequest
	}
	if containsAny(messageAndCause, "not found", "unsupported", "unknown model") {
		return llm.CategoryNotFound
	}
	if strings.Contains(messageAndCause, "encoding") {
		return llm.CategoryNotFound
	}

	// Default: Unknown if no cause and no matching patterns, Server otherwise
	if cause == nil && messageAndCause == "" {
		return llm.CategoryUnknown
	}
	return llm.CategoryServer
}

// containsAny checks if text contains any of the keywords.
func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// buildErrorDetails constructs error details for debugging.
func buildErrorDetails(provider, model string, cause error) string {
	details := "Tokenizer error details:"
	if provider != "" {
		details += " provider=" + provider
	}
	if model != "" {
		details += " model=" + model
	}
	if cause != nil {
		details += " cause=" + cause.Error()
	}
	return details
}

// buildEnhancedErrorDetails includes tokenizer type in error details.
func buildEnhancedErrorDetails(provider, model string, cause error, tokenizerType string) string {
	details := "Tokenizer error details:"
	if provider != "" {
		details += " provider=" + provider
	}
	if model != "" {
		details += " model=" + model
	}
	if tokenizerType != "" {
		details += " tokenizer=" + tokenizerType
	}
	if cause != nil {
		details += " cause=" + cause.Error()
	}
	return details
}
