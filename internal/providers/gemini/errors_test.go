package gemini

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/gemini"
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
			llmErr, isGeminiErr := gemini.IsGeminiError(tt.err)
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

func TestFormatAPIError(t *testing.T) {
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
			wantMsgPrefix: "Gemini API server",
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
			wantCategory:  llm.CategoryUnknown, // Falls back to unknown without status code
			wantProvider:  "gemini",
			wantMsgPrefix: "Error calling Gemini API",
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
			wantMsgPrefix: "Request to Gemini API was cancelled",
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
			result := gemini.FormatAPIError(tt.err, tt.statusCode)

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
			wantProvider:  "gemini",
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
			wantProvider:  "gemini",
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
			wantProvider:  "gemini",
			wantMsgPrefix: "Invalid request parameters",
			wantDetails:   "temperature must be between 0 and 2",
		},
		{
			name:          "empty message uses default",
			category:      llm.CategoryAuth,
			errMsg:        "",
			originalErr:   fmt.Errorf("invalid auth"),
			details:       "",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "gemini",
			wantMsgPrefix: "Authentication failed", // Default message should be used
			wantDetails:   "",
		},
		{
			name:          "unknown error",
			category:      llm.CategoryUnknown,
			errMsg:        "Unknown error occurred",
			originalErr:   errors.New("some error"),
			details:       "",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "gemini",
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
			wantProvider:  "gemini",
			wantMsgPrefix: "Authentication failed",
			wantDetails:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gemini.CreateAPIError(tt.category, tt.errMsg, tt.originalErr, tt.details)

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

func TestIsAPIError(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmErr, isAPIErr := gemini.IsAPIError(tt.err)
			assert.Equal(t, tt.wantIsError, isAPIErr)

			if tt.wantIsError {
				assert.NotNil(t, llmErr)
				assert.Equal(t, "gemini", llmErr.Provider)
			} else {
				assert.Nil(t, llmErr)
			}
		})
	}
}
