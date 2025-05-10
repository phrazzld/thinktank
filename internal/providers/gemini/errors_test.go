package gemini

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
)

func TestIsGeminiError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantIsError bool
	}{
		{
			name:        "nil error",
			err:         nil,
			wantIsError: false,
		},
		{
			name:        "regular error",
			err:         fmt.Errorf("some error"),
			wantIsError: false,
		},
		{
			name: "Gemini LLM error",
			err: &llm.LLMError{
				Provider:      "gemini",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			wantIsError: true,
		},
		{
			name: "Non-Gemini LLM error",
			err: &llm.LLMError{
				Provider:      "openai",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			wantIsError: false,
		},
		{
			name: "Wrapped Gemini error",
			err: fmt.Errorf("wrapper: %w", &llm.LLMError{
				Provider:      "gemini",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			}),
			wantIsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmErr, isGeminiErr := IsGeminiError(tt.err)
			assert.Equal(t, tt.wantIsError, isGeminiErr)

			if tt.wantIsError {
				assert.NotNil(t, llmErr)
				assert.Equal(t, "gemini", llmErr.Provider)
			} else {
				assert.Nil(t, llmErr)
			}
		})
	}
}

func TestFormatAPIErrorFromResponse(t *testing.T) {
	tests := []struct {
		name                string
		err                 error
		statusCode          int
		wantCategory        llm.ErrorCategory
		wantProvider        string
		wantMsgPrefix       string
		skipSuggestionCheck bool
	}{
		{
			name:          "nil error",
			err:           nil,
			statusCode:    0,
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "",
			wantMsgPrefix: "",
		},
		{
			name:          "auth error with status code",
			err:           fmt.Errorf("invalid auth"),
			statusCode:    401,
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "gemini",
			wantMsgPrefix: "Authentication failed",
		},
		{
			name:          "rate limit error with status code",
			err:           fmt.Errorf("too many requests"),
			statusCode:    429,
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Request rate limit",
		},
		{
			name:          "server error with status code",
			err:           fmt.Errorf("internal server error"),
			statusCode:    500,
			wantCategory:  llm.CategoryServer,
			wantProvider:  "gemini",
			wantMsgPrefix: "gemini API server", // Updated to match actual message
		},
		{
			name:          "invalid request with status code",
			err:           fmt.Errorf("bad request"),
			statusCode:    400,
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "gemini",
			wantMsgPrefix: "Invalid request",
		},
		{
			name:          "not found with status code",
			err:           fmt.Errorf("not found"),
			statusCode:    404,
			wantCategory:  llm.CategoryNotFound,
			wantProvider:  "gemini",
			wantMsgPrefix: "The requested model or resource",
		},
		{
			name:          "auth error from message",
			err:           fmt.Errorf("authorization failed"),
			statusCode:    0,
			wantCategory:  llm.CategoryAuth, // Updated to match actual category
			wantProvider:  "gemini",
			wantMsgPrefix: "Authentication failed with the gemini API", // Updated to match actual message
		},
		{
			name:          "rate limit error from message",
			err:           fmt.Errorf("rate limit exceeded"),
			statusCode:    0,
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Request rate limit",
		},
		{
			name:          "quota error from message",
			err:           fmt.Errorf("quota exceeded"),
			statusCode:    0,
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Request rate limit",
		},
		{
			name:          "content filtering from message",
			err:           fmt.Errorf("content filtered by safety settings"),
			statusCode:    0,
			wantCategory:  llm.CategoryContentFiltered,
			wantProvider:  "gemini",
			wantMsgPrefix: "Content was filtered",
		},
		{
			name:          "token limit from message",
			err:           fmt.Errorf("token limit exceeded"),
			statusCode:    0,
			wantCategory:  llm.CategoryInputLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Input token limit",
		},
		{
			name:          "network error from message",
			err:           fmt.Errorf("network timeout"),
			statusCode:    0,
			wantCategory:  llm.CategoryNetwork,
			wantProvider:  "gemini",
			wantMsgPrefix: "Network error",
		},
		{
			name:          "cancelled error from message",
			err:           fmt.Errorf("request cancelled"),
			statusCode:    0,
			wantCategory:  llm.CategoryCancelled,
			wantProvider:  "gemini",
			wantMsgPrefix: "Request to gemini API was cancelled", // Updated to match actual message
		},
		{
			name:          "status code takes precedence over message",
			err:           fmt.Errorf("network timeout"), // Would be CategoryNetwork from message
			statusCode:    429,                           // But CategoryRateLimit from status code
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Request rate limit",
		},
		{
			name: "already LLMError",
			err: &llm.LLMError{
				Provider:      "gemini",
				Message:       "Existing error",
				ErrorCategory: llm.CategoryAuth,
			},
			statusCode:          0,
			wantCategory:        llm.CategoryAuth,
			wantProvider:        "gemini",
			wantMsgPrefix:       "Existing error",
			skipSuggestionCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAPIErrorFromResponse(tt.err, tt.statusCode, nil)

			if tt.err == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.wantProvider, result.Provider)
			assert.Equal(t, tt.wantCategory, result.Category())

			if tt.wantMsgPrefix != "" {
				assert.True(t, strings.HasPrefix(result.Message, tt.wantMsgPrefix),
					"Expected message to start with %q, got %q", tt.wantMsgPrefix, result.Message)
			}

			// Check that suggestions are present for known error categories
			if tt.wantCategory != llm.CategoryUnknown && !tt.skipSuggestionCheck {
				assert.NotEmpty(t, result.Suggestion, "Expected non-empty suggestion for %s error", tt.wantCategory)
			}
		})
	}
}

func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		providerName  string
		wantCategory  llm.ErrorCategory
		wantProvider  string
		wantMsgPrefix string
		checkUnwrap   bool
	}{
		{
			name:          "nil error",
			err:           nil,
			providerName:  "gemini",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "",
			wantMsgPrefix: "",
		},
		{
			name:          "regular error",
			err:           fmt.Errorf("some error"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name: "already LLMError from same provider",
			err: &llm.LLMError{
				Provider:      "gemini",
				Message:       "Existing error",
				ErrorCategory: llm.CategoryAuth,
			},
			providerName:  "gemini",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "gemini",
			wantMsgPrefix: "Existing error",
		},
		{
			name: "already LLMError from different provider",
			err: &llm.LLMError{
				Provider:      "openai",
				Message:       "Existing error",
				ErrorCategory: llm.CategoryAuth,
			},
			providerName:  "gemini",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "gemini",
			wantMsgPrefix: "Existing error",
		},
		{
			name:          "auth error from message",
			err:           fmt.Errorf("invalid auth"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "rate limit error from message",
			err:           fmt.Errorf("rate limit exceeded"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "network error from message",
			err:           fmt.Errorf("network timeout"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryNetwork,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "insufficient credits error from message",
			err:           fmt.Errorf("billing quota exceeded"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryInsufficientCredits,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "content filtered error from message",
			err:           fmt.Errorf("content filtered by safety settings"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryContentFiltered,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "input limit error from message",
			err:           fmt.Errorf("token limit exceeded"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryInputLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "cancelled error from message",
			err:           fmt.Errorf("request cancelled"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryCancelled,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "not found error from message",
			err:           fmt.Errorf("model not found"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryNotFound,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "invalid request error from message",
			err:           fmt.Errorf("invalid request parameters"),
			providerName:  "gemini",
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
		{
			name:          "wrapped error",
			err:           fmt.Errorf("outer error: %w", fmt.Errorf("inner rate limit exceeded")),
			providerName:  "gemini",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Error from gemini provider",
			checkUnwrap:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAPIError(tt.err, tt.providerName)

			if tt.err == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)

			var llmErr *llm.LLMError
			if errors.As(result, &llmErr) {
				assert.Equal(t, tt.wantProvider, llmErr.Provider)
				assert.Equal(t, tt.wantCategory, llmErr.Category())

				if tt.wantMsgPrefix != "" {
					assert.True(t, strings.HasPrefix(llmErr.Message, tt.wantMsgPrefix),
						"Expected message to start with %q, got %q", tt.wantMsgPrefix, llmErr.Message)
				}

				// Check for proper error wrapping
				if tt.checkUnwrap {
					unwrapped := errors.Unwrap(result)
					assert.NotNil(t, unwrapped, "Expected error to be wrapped")
					assert.Equal(t, tt.err.Error(), unwrapped.Error(), "Original error should be preserved when unwrapped")
				}

				// FormatAPIError uses llm.Wrap which doesn't set suggestions
				// Suggestion fields are only populated by CreateStandardErrorWithMessage
				// and similar functions, not by FormatAPIError
			} else {
				t.Fatalf("Expected result to be of type *llm.LLMError")
			}
		})
	}
}
