// Package tokenizers provides accurate token counting implementations for various LLM providers.
// This package enables precise token counting to replace estimation-based model selection.
package tokenizers

import (
	"context"
	"io"
	"strings"

	"github.com/phrazzld/thinktank/internal/llm"
)

// AccurateTokenCounter provides model-specific accurate token counting.
// Implementations should handle provider-specific tokenization algorithms like tiktoken for OpenAI
// or SentencePiece for Gemini models. This interface enables precise token counting to replace
// estimation-based model selection decisions.
type AccurateTokenCounter interface {
	// CountTokens returns accurate token count for the given text and model.
	// Context is used for cancellation and timeout handling.
	// Returns error if the model is not supported or tokenization fails.
	// Empty text returns 0 tokens without error.
	CountTokens(ctx context.Context, text string, modelName string) (int, error)

	// SupportsModel returns true if accurate counting is available for the model.
	// This method should be used to check compatibility before calling CountTokens.
	SupportsModel(modelName string) bool

	// GetEncoding returns the tokenizer encoding name for the given model.
	// Examples: "cl100k_base" for GPT-4, "o200k_base" for GPT-4o, "gemini" for Gemini models.
	// Returns error if the model is not supported.
	GetEncoding(modelName string) (string, error)
}

// TokenizerManager handles lazy loading and caching of tokenizers.
// Provides centralized access to all tokenizer implementations with provider-aware routing.
// Implements the factory pattern for creating appropriate tokenizers based on provider names.
// Thread-safe for concurrent access.
type TokenizerManager interface {
	// GetTokenizer returns a tokenizer for the given provider.
	// Supported providers: "openai" (tiktoken), "gemini" (SentencePiece).
	// Handles lazy initialization and caching for performance.
	// Returns error for unsupported providers.
	GetTokenizer(provider string) (AccurateTokenCounter, error)

	// SupportsProvider returns true if accurate tokenization is available for the provider.
	// Use this method to check compatibility before calling GetTokenizer.
	SupportsProvider(provider string) bool

	// ClearCache clears all cached tokenizers to free memory.
	// Useful for testing or memory management in long-running processes.
	ClearCache()
}

// StreamingTokenCounter extends AccurateTokenCounter with streaming capabilities
// for handling very large inputs that don't fit in memory.
// Automatically switches to streaming mode for inputs larger than memory thresholds.
// Provides the same accuracy as AccurateTokenCounter while handling massive files.
type StreamingTokenCounter interface {
	AccurateTokenCounter
	// CountTokensStreaming tokenizes content from a reader in chunks.
	// Useful for files too large to fit in memory (>50MB).
	// Provides same accuracy as CountTokens but with streaming processing.
	CountTokensStreaming(ctx context.Context, reader io.Reader, modelName string) (int, error)

	// CountTokensStreamingWithAdaptiveChunking tokenizes content using adaptive chunk sizing.
	// Uses input size to determine optimal chunk size for better performance:
	// - Small inputs (<5MB): 8KB chunks for responsiveness
	// - Medium inputs (5-25MB): 32KB chunks for balanced performance
	// - Large inputs (>25MB): 64KB chunks for maximum throughput
	CountTokensStreamingWithAdaptiveChunking(ctx context.Context, reader io.Reader, modelName string, inputSizeBytes int) (int, error)

	// GetChunkSizeForInput returns the optimal chunk size based on input size.
	// Implements adaptive chunking algorithm for performance optimization.
	GetChunkSizeForInput(inputSizeBytes int) int
}

// TokenizerError represents errors from tokenization operations with enhanced categorization.
// Implements the llm.CategorizedError interface for consistent error handling.
// Provides structured error information for debugging and circuit breaker integration.
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

	// Include enhanced details for better debugging
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
// Automatically categorizes the error based on patterns in the message and cause.
// Use this constructor for standard error creation with automatic categorization.
func NewTokenizerError(provider, model, message string, cause error) *TokenizerError {
	category := categorizeTokenizerError(provider, model, message, cause)
	details := buildErrorDetails(provider, model, message, cause)

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
// Use this constructor when you know the specific error category to avoid automatic detection.
// Useful for consistent categorization in specific error scenarios.
func NewTokenizerErrorWithCategory(provider, model, message string, cause error, category llm.ErrorCategory) *TokenizerError {
	details := buildErrorDetails(provider, model, message, cause)

	return &TokenizerError{
		Provider:      provider,
		Model:         model,
		Message:       message,
		Cause:         cause,
		ErrorCategory: category,
		Details:       details,
	}
}

// NewTokenizerErrorWithDetails creates a new tokenizer error with enhanced details including tokenizer type.
// Provides the most comprehensive error information for debugging and monitoring.
// Includes provider, model, tokenizer type, and cause in structured details.
// Recommended constructor for production error handling.
func NewTokenizerErrorWithDetails(provider, model, message string, cause error, tokenizerType string) *TokenizerError {
	category := categorizeTokenizerError(provider, model, message, cause)
	details := buildEnhancedErrorDetails(provider, model, message, cause, tokenizerType)

	return &TokenizerError{
		Provider:      provider,
		Model:         model,
		Message:       message,
		Cause:         cause,
		ErrorCategory: category,
		Details:       details,
	}
}

// categorizeTokenizerError determines the appropriate error category based on error context.
// Analyzes error messages and causes to automatically assign llm.ErrorCategory values.
// Used for consistent error categorization across all tokenizer implementations.
// Returns llm.CategoryUnknown if no specific pattern is detected.
func categorizeTokenizerError(provider, model, message string, cause error) llm.ErrorCategory {
	if cause == nil {
		return llm.CategoryUnknown
	}

	// Check for context cancellation (timeouts)
	if cause.Error() == "context deadline exceeded" || cause.Error() == "context canceled" {
		return llm.CategoryCancelled
	}

	// Check for specific tokenizer failure patterns
	messageAndCause := message + " " + cause.Error()

	// Circuit breaker related errors
	if contains(messageAndCause, "circuit breaker") {
		return llm.CategoryRateLimit // Circuit breaker often indicates rate limiting or overload
	}

	// Authentication/initialization errors
	if contains(messageAndCause, "initialization", "auth", "api key", "credential") {
		return llm.CategoryAuth
	}

	// Network connectivity issues
	if contains(messageAndCause, "network", "connection", "dial", "timeout", "unreachable") {
		return llm.CategoryNetwork
	}

	// Invalid input/request (check before model not found to catch "encoding failed")
	if contains(messageAndCause, "invalid", "malformed", "encoding failed") {
		return llm.CategoryInvalidRequest
	}

	// Model/encoding not found (check after invalid request)
	if contains(messageAndCause, "not found", "unsupported", "unknown model") {
		return llm.CategoryNotFound
	}

	// General encoding issues (fallback for encoding-related problems)
	if contains(messageAndCause, "encoding") {
		return llm.CategoryNotFound
	}

	// Default to server error for tokenization failures
	return llm.CategoryServer
}

// buildErrorDetails constructs detailed error information for debugging.
// Creates structured error details including provider, model, and cause information.
// Used by NewTokenizerError and NewTokenizerErrorWithCategory constructors.
// Provides consistent formatting for error details across all tokenizers.
func buildErrorDetails(provider, model, message string, cause error) string {
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

// contains checks if any of the keywords exist in the text (case-insensitive).
// Utility function for pattern matching in error categorization.
// Returns true if any keyword is found in the text, false otherwise.
// Used internally by categorizeTokenizerError for error pattern detection.
func contains(text string, keywords ...string) bool {
	textLower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// buildEnhancedErrorDetails constructs detailed error information with tokenizer type for debugging.
// Enhanced version of buildErrorDetails that includes tokenizer type (tiktoken, sentencepiece, etc.).
// Used by NewTokenizerErrorWithDetails for the most comprehensive error information.
// Provides full context for debugging tokenization failures.
func buildEnhancedErrorDetails(provider, model, message string, cause error, tokenizerType string) string {
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

// getTokenizerType returns the tokenizer type for a given provider.
// Maps provider names to their corresponding tokenizer implementation names.
// Used for error reporting and debugging to identify which tokenizer was used.
// Returns "unknown" for unrecognized providers.
func getTokenizerType(provider string) string {
	switch provider {
	case "openai":
		return "tiktoken"
	case "gemini":
		return "sentencepiece"
	case "openrouter":
		return "tiktoken-o200k"
	default:
		return "unknown"
	}
}
