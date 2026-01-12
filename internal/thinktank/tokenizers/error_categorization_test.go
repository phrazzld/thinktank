package tokenizers

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
)

// TestTokenizerError_ImplementsCategorizedError tests that TokenizerError implements CategorizedError interface
func TestTokenizerError_ImplementsCategorizedError(t *testing.T) {
	t.Parallel()

	err := NewTokenizerError("openai", "gpt-4", "test error", errors.New("underlying error"))

	// Verify it implements CategorizedError interface
	var categorizedErr llm.CategorizedError
	assert.Implements(t, &categorizedErr, err, "TokenizerError should implement CategorizedError interface")

	// Verify we can extract the category
	category := err.Category()
	assert.NotEqual(t, llm.CategoryUnknown, category, "Should determine a specific category")
}

// TestCategorizeTokenizerError_TimeoutErrors tests timeout error categorization
func TestCategorizeTokenizerError_TimeoutErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		cause    error
		expected llm.ErrorCategory
	}{
		{
			name:     "context deadline exceeded",
			cause:    context.DeadlineExceeded,
			expected: llm.CategoryCancelled,
		},
		{
			name:     "context canceled",
			cause:    context.Canceled,
			expected: llm.CategoryCancelled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewTokenizerError("openai", "gpt-4", "timeout", tc.cause)
			assert.Equal(t, tc.expected, err.Category())
		})
	}
}

// TestCategorizeTokenizerError_CircuitBreakerErrors tests circuit breaker error categorization
func TestCategorizeTokenizerError_CircuitBreakerErrors(t *testing.T) {
	t.Parallel()

	err := NewTokenizerError("openai", "gpt-4", "circuit breaker open", errors.New("too many failures"))
	assert.Equal(t, llm.CategoryRateLimit, err.Category())
}

// TestCategorizeTokenizerError_AuthenticationErrors tests authentication error categorization
func TestCategorizeTokenizerError_AuthenticationErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "initialization failure",
			message: "tokenizer initialization failed",
			cause:   errors.New("api key invalid"),
		},
		{
			name:    "auth error",
			message: "authentication error",
			cause:   errors.New("credential validation failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewTokenizerError("openai", "gpt-4", tc.message, tc.cause)
			assert.Equal(t, llm.CategoryAuth, err.Category())
		})
	}
}

// TestCategorizeTokenizerError_NetworkErrors tests network error categorization
func TestCategorizeTokenizerError_NetworkErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "connection timeout",
			message: "network error",
			cause:   errors.New("connection timeout"),
		},
		{
			name:    "dial error",
			message: "failed to connect",
			cause:   errors.New("dial tcp: network unreachable"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewTokenizerError("openai", "gpt-4", tc.message, tc.cause)
			assert.Equal(t, llm.CategoryNetwork, err.Category())
		})
	}
}

// TestCategorizeTokenizerError_ModelNotFoundErrors tests model not found error categorization
func TestCategorizeTokenizerError_ModelNotFoundErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "unsupported model",
			message: "unsupported model",
			cause:   errors.New("model not found"),
		},
		{
			name:    "encoding not found",
			message: "encoding error",
			cause:   errors.New("unknown model encoding"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewTokenizerError("openai", "unknown-model", tc.message, tc.cause)
			assert.Equal(t, llm.CategoryNotFound, err.Category())
		})
	}
}

// TestCategorizeTokenizerError_InvalidRequestErrors tests invalid request error categorization
func TestCategorizeTokenizerError_InvalidRequestErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		message string
		cause   error
	}{
		{
			name:    "malformed input",
			message: "tokenization failed",
			cause:   errors.New("malformed input text"),
		},
		{
			name:    "encoding failed",
			message: "invalid input",
			cause:   errors.New("encoding failed due to invalid characters"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewTokenizerError("openai", "gpt-4", tc.message, tc.cause)
			assert.Equal(t, llm.CategoryInvalidRequest, err.Category())
		})
	}
}

// TestCategorizeTokenizerError_DefaultServerError tests default server error categorization
func TestCategorizeTokenizerError_DefaultServerError(t *testing.T) {
	t.Parallel()

	err := NewTokenizerError("openai", "gpt-4", "general tokenization failure", errors.New("unexpected tokenizer error"))
	assert.Equal(t, llm.CategoryServer, err.Category())
}

// TestTokenizerError_Details tests that error details are properly constructed
func TestTokenizerError_Details(t *testing.T) {
	t.Parallel()

	err := NewTokenizerError("openai", "gpt-4", "test error", errors.New("underlying cause"))

	assert.Contains(t, err.Details, "provider=openai")
	assert.Contains(t, err.Details, "model=gpt-4")
	assert.Contains(t, err.Details, "cause=underlying cause")
}

// TestNewTokenizerErrorWithCategory tests explicit category setting
func TestNewTokenizerErrorWithCategory(t *testing.T) {
	t.Parallel()

	err := NewTokenizerErrorWithCategory("gemini", "gemini-3-flash", "custom error",
		errors.New("cause"), llm.CategoryInsufficientCredits)

	assert.Equal(t, llm.CategoryInsufficientCredits, err.Category())
	assert.Equal(t, "gemini", err.Provider)
	assert.Equal(t, "gemini-3-flash", err.Model)
	assert.Equal(t, "custom error", err.Message)
}

// TestTokenizerError_LLMErrorCompatibility tests compatibility with llm.IsCategorizedError
func TestTokenizerError_LLMErrorCompatibility(t *testing.T) {
	t.Parallel()

	tokenizerErr := NewTokenizerError("openai", "gpt-4", "test error", errors.New("cause"))

	// Should be detectable as a categorized error
	categorizedErr, ok := llm.IsCategorizedError(tokenizerErr)
	assert.True(t, ok, "Should be detectable as CategorizedError")
	assert.NotNil(t, categorizedErr, "Should return the categorized error")
	assert.Equal(t, tokenizerErr.Category(), categorizedErr.Category(), "Category should match")
}
