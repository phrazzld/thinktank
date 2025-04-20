package openai

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/openai"
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
			llmErr, isOpenAIErr := openai.IsOpenAIError(tt.err)
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
			result := openai.FormatAPIError(tt.err, tt.statusCode)

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
			result := openai.CreateAPIError(tt.category, tt.errMsg, tt.originalErr, tt.details)

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

func TestMockAPIErrorResponse(t *testing.T) {
	tests := []struct {
		name                string
		errorType           int
		statusCode          int
		message             string
		details             string
		wantCategory        llm.ErrorCategory
		wantProvider        string
		wantMsg             string
		wantDetails         string
		wantStatusCode      int
		skipSuggestionCheck bool
	}{
		{
			name:           "auth error",
			errorType:      1, // ErrorTypeAuth
			statusCode:     401,
			message:        "Invalid authentication",
			details:        "API key expired",
			wantCategory:   llm.CategoryAuth,
			wantProvider:   "openai",
			wantMsg:        "Invalid authentication",
			wantDetails:    "API key expired",
			wantStatusCode: 401,
		},
		{
			name:           "rate limit error",
			errorType:      2, // ErrorTypeRateLimit
			statusCode:     429,
			message:        "Too many requests",
			details:        "Retry-After: 30",
			wantCategory:   llm.CategoryRateLimit,
			wantProvider:   "openai",
			wantMsg:        "Too many requests",
			wantDetails:    "Retry-After: 30",
			wantStatusCode: 429,
		},
		{
			name:           "invalid request error",
			errorType:      3, // ErrorTypeInvalidRequest
			statusCode:     400,
			message:        "Invalid request parameters",
			details:        "Bad temperature",
			wantCategory:   llm.CategoryInvalidRequest,
			wantProvider:   "openai",
			wantMsg:        "Invalid request parameters",
			wantDetails:    "Bad temperature",
			wantStatusCode: 400,
		},
		{
			name:           "not found error",
			errorType:      4, // ErrorTypeNotFound
			statusCode:     404,
			message:        "Model not found",
			details:        "gpt-9 does not exist",
			wantCategory:   llm.CategoryNotFound,
			wantProvider:   "openai",
			wantMsg:        "Model not found",
			wantDetails:    "gpt-9 does not exist",
			wantStatusCode: 404,
		},
		{
			name:           "server error",
			errorType:      5, // ErrorTypeServer
			statusCode:     500,
			message:        "Internal server error",
			details:        "Service unavailable",
			wantCategory:   llm.CategoryServer,
			wantProvider:   "openai",
			wantMsg:        "Internal server error",
			wantDetails:    "Service unavailable",
			wantStatusCode: 500,
		},
		{
			name:           "unknown error type",
			errorType:      999, // Unknown type
			statusCode:     418, // I'm a teapot
			message:        "Teapot error",
			details:        "Cannot brew coffee",
			wantCategory:   llm.CategoryUnknown,
			wantProvider:   "openai",
			wantMsg:        "Teapot error",
			wantDetails:    "Cannot brew coffee",
			wantStatusCode: 418,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openai.MockAPIErrorResponse(tt.errorType, tt.statusCode, tt.message, tt.details)

			assert.NotNil(t, result)
			assert.Equal(t, tt.wantProvider, result.Provider)
			assert.Equal(t, tt.wantCategory, result.Category())
			assert.Equal(t, tt.wantDetails, result.Details)
			assert.Equal(t, tt.wantMsg, result.Message)
			assert.Equal(t, tt.wantStatusCode, result.StatusCode)

			// Check that suggestions are present for known error categories
			if tt.wantCategory != llm.CategoryUnknown && !tt.skipSuggestionCheck {
				assert.NotEmpty(t, result.Suggestion, "Expected non-empty suggestion for %s error", tt.wantCategory)
			}
		})
	}
}
