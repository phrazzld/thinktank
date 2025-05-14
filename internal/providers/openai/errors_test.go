package openai

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
)

func TestIsOpenAIError(t *testing.T) {
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
			name: "OpenAI LLM error",
			err: &llm.LLMError{
				Provider:      "openai",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			wantIsError: true,
		},
		{
			name: "Non-OpenAI LLM error",
			err: &llm.LLMError{
				Provider:      "gemini",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			wantIsError: false,
		},
		{
			name: "Wrapped OpenAI error",
			err: fmt.Errorf("wrapper: %w", &llm.LLMError{
				Provider:      "openai",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			}),
			wantIsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmErr, isOpenAIErr := IsOpenAIError(tt.err)
			assert.Equal(t, tt.wantIsError, isOpenAIErr)

			if tt.wantIsError {
				assert.NotNil(t, llmErr)
				assert.Equal(t, "openai", llmErr.Provider)
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
			wantProvider:  "openai",
			wantMsgPrefix: "Authentication failed",
		},
		{
			name:          "rate limit error with status code",
			err:           fmt.Errorf("too many requests"),
			statusCode:    429,
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Request rate limit exceeded",
		},
		{
			name:          "server error with status code",
			err:           fmt.Errorf("internal server error"),
			statusCode:    500,
			wantCategory:  llm.CategoryServer,
			wantProvider:  "openai",
			wantMsgPrefix: "openai API server",
		},
		{
			name:          "invalid request with status code",
			err:           fmt.Errorf("bad request"),
			statusCode:    400,
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "openai",
			wantMsgPrefix: "Invalid request",
		},
		{
			name:          "not found with status code",
			err:           fmt.Errorf("not found"),
			statusCode:    404,
			wantCategory:  llm.CategoryNotFound,
			wantProvider:  "openai",
			wantMsgPrefix: "The requested model or resource",
		},
		{
			name:          "auth error from message",
			err:           fmt.Errorf("authorization failed"),
			statusCode:    0,
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openai",
			wantMsgPrefix: "Authentication failed",
		},
		{
			name:          "rate limit error from message",
			err:           fmt.Errorf("rate limit exceeded"),
			statusCode:    0,
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Request rate limit exceeded",
		},
		{
			name:          "billing error from message",
			err:           fmt.Errorf("billing quota exceeded"),
			statusCode:    0,
			wantCategory:  llm.CategoryInsufficientCredits,
			wantProvider:  "openai",
			wantMsgPrefix: "Insufficient credits",
		},
		{
			name:          "content filtering from message",
			err:           fmt.Errorf("content_filter triggered"),
			statusCode:    0,
			wantCategory:  llm.CategoryContentFiltered,
			wantProvider:  "openai",
			wantMsgPrefix: "Content was filtered",
		},
		{
			name:          "token limit from message",
			err:           fmt.Errorf("token limit exceeded"),
			statusCode:    0,
			wantCategory:  llm.CategoryInputLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Input token limit exceeded",
		},
		{
			name:          "network error from message",
			err:           fmt.Errorf("network timeout"),
			statusCode:    0,
			wantCategory:  llm.CategoryNetwork,
			wantProvider:  "openai",
			wantMsgPrefix: "Network error",
		},
		{
			name:          "cancelled error from message",
			err:           fmt.Errorf("request cancelled"),
			statusCode:    0,
			wantCategory:  llm.CategoryCancelled,
			wantProvider:  "openai",
			wantMsgPrefix: "Request to openai API was cancelled",
		},
		{
			name:          "status code takes precedence over message",
			err:           fmt.Errorf("network timeout"), // Would be CategoryNetwork from message
			statusCode:    429,                           // But CategoryRateLimit from status code
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Request rate limit exceeded",
		},
		{
			name: "already LLMError",
			err: &llm.LLMError{
				Provider:      "openai",
				Message:       "Existing error",
				ErrorCategory: llm.CategoryAuth,
			},
			statusCode:          0,
			wantCategory:        llm.CategoryAuth,
			wantProvider:        "openai",
			wantMsgPrefix:       "Existing error",
			skipSuggestionCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pass nil as the response body for the test
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
			providerName:  "openai",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "",
			wantMsgPrefix: "",
		},
		{
			name:          "standard error",
			err:           fmt.Errorf("some error"),
			providerName:  "openai",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "network error",
			err:           fmt.Errorf("connection timeout"),
			providerName:  "openai",
			wantCategory:  llm.CategoryNetwork,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "auth error",
			err:           fmt.Errorf("authentication failed"),
			providerName:  "openai",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "rate limit error",
			err:           fmt.Errorf("rate limit exceeded"),
			providerName:  "openai",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name: "existing LLMError from same provider",
			err: &llm.LLMError{
				Provider:      "openai",
				Message:       "Existing error",
				ErrorCategory: llm.CategoryAuth,
			},
			providerName:  "openai",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openai",
			wantMsgPrefix: "Existing error",
		},
		{
			name: "existing LLMError from different provider",
			err: &llm.LLMError{
				Provider:      "gemini",
				Message:       "Error from gemini",
				ErrorCategory: llm.CategoryAuth,
			},
			providerName:  "openai",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from gemini",
		},
		{
			name:          "insufficient credits error from message",
			err:           fmt.Errorf("billing quota exceeded"),
			providerName:  "openai",
			wantCategory:  llm.CategoryInsufficientCredits,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "content filtered error from message",
			err:           fmt.Errorf("content_filter triggered"),
			providerName:  "openai",
			wantCategory:  llm.CategoryContentFiltered,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "input limit error from message",
			err:           fmt.Errorf("token limit exceeded"),
			providerName:  "openai",
			wantCategory:  llm.CategoryInputLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "cancelled error from message",
			err:           fmt.Errorf("request cancelled"),
			providerName:  "openai",
			wantCategory:  llm.CategoryCancelled,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "not found error from message",
			err:           fmt.Errorf("model not found"),
			providerName:  "openai",
			wantCategory:  llm.CategoryNotFound,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "invalid request error from message",
			err:           fmt.Errorf("invalid request parameters"),
			providerName:  "openai",
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "wrapped error",
			err:           fmt.Errorf("outer error: %w", fmt.Errorf("inner rate limit exceeded")),
			providerName:  "openai",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Error from openai provider",
			checkUnwrap:   true,
		},
		{
			name:          "error with custom provider name",
			err:           fmt.Errorf("some error"),
			providerName:  "custom-provider",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "custom-provider",
			wantMsgPrefix: "Error from custom-provider provider",
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

func TestCreateAPIError(t *testing.T) {
	tests := []struct {
		name                string
		category            llm.ErrorCategory
		errMsg              string
		originalErr         error
		details             string
		wantCategory        llm.ErrorCategory
		wantProvider        string
		wantMsgPrefix       string
		wantDetails         string
		skipSuggestionCheck bool
	}{
		{
			name:          "auth error",
			category:      llm.CategoryAuth,
			errMsg:        "Authentication failed",
			originalErr:   fmt.Errorf("invalid auth"),
			details:       "API key invalid",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openai",
			wantMsgPrefix: "Authentication failed",
			wantDetails:   "API key invalid",
		},
		{
			name:          "rate limit error",
			category:      llm.CategoryRateLimit,
			errMsg:        "Rate limit exceeded",
			originalErr:   fmt.Errorf("too many requests"),
			details:       "Retry after 60s",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openai",
			wantMsgPrefix: "Rate limit exceeded",
			wantDetails:   "Retry after 60s",
		},
		{
			name:          "invalid request error",
			category:      llm.CategoryInvalidRequest,
			errMsg:        "Invalid request parameters",
			originalErr:   fmt.Errorf("bad request"),
			details:       "temperature must be between 0 and 2",
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "openai",
			wantMsgPrefix: "Invalid request parameters",
			wantDetails:   "temperature must be between 0 and 2",
		},
		{
			name:          "unknown error",
			category:      llm.CategoryUnknown,
			errMsg:        "Unknown error occurred",
			originalErr:   errors.New("some error"),
			details:       "",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "openai",
			wantMsgPrefix: "Unknown error occurred",
			wantDetails:   "",
		},
		{
			name:          "no original error",
			category:      llm.CategoryAuth,
			errMsg:        "Authentication failed",
			originalErr:   nil,
			details:       "",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openai",
			wantMsgPrefix: "Authentication failed",
			wantDetails:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateAPIError(tt.category, tt.errMsg, tt.originalErr, tt.details)

			assert.NotNil(t, result)
			assert.Equal(t, tt.wantProvider, result.Provider)
			assert.Equal(t, tt.wantCategory, result.Category())
			assert.Equal(t, tt.wantDetails, result.Details)

			assert.True(t, strings.HasPrefix(result.Message, tt.wantMsgPrefix),
				"Expected message to start with %q, got %q", tt.wantMsgPrefix, result.Message)

			// Check that suggestions are present for known error categories
			if tt.wantCategory != llm.CategoryUnknown && !tt.skipSuggestionCheck {
				assert.NotEmpty(t, result.Suggestion, "Expected non-empty suggestion for %s error", tt.wantCategory)
			}

			// Verify original error is preserved
			assert.Equal(t, tt.originalErr, result.Original)
		})
	}
}
