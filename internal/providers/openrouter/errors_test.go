// Package openrouter contains tests for the OpenRouter client
package openrouter

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseErrorResponse(t *testing.T) {
	t.Parallel() // Pure CPU-bound JSON response parsing test
	tests := []struct {
		name           string
		responseBody   []byte
		wantErrMsg     string
		wantErrType    string
		wantErrParam   string
		wantEmptyParse bool
	}{
		{
			name:           "Empty response body",
			responseBody:   []byte(""),
			wantEmptyParse: true,
		},
		{
			name:           "Invalid JSON",
			responseBody:   []byte("not json"),
			wantEmptyParse: true,
		},
		{
			name:           "Valid JSON but not error format",
			responseBody:   []byte(`{"foo": "bar"}`),
			wantEmptyParse: true,
		},
		{
			name:         "Error message only",
			responseBody: []byte(`{"error": {"message": "API error occurred"}}`),
			wantErrMsg:   "API error occurred",
			wantErrType:  "",
			wantErrParam: "",
		},
		{
			name:         "Complete error response",
			responseBody: []byte(`{"error": {"code": "insufficient_quota", "message": "You exceeded your quota", "type": "billing_error", "param": "account_balance"}}`),
			wantErrMsg:   "You exceeded your quota",
			wantErrType:  "billing_error",
			wantErrParam: "account_balance",
		},
		{
			name:         "Error with numeric code",
			responseBody: []byte(`{"error": {"code": 429, "message": "Rate limit exceeded", "type": "rate_limit_error"}}`),
			wantErrMsg:   "Rate limit exceeded",
			wantErrType:  "rate_limit_error",
			wantErrParam: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg, errType, errParam := ParseErrorResponse(tt.responseBody)

			if tt.wantEmptyParse {
				assert.Empty(t, errMsg, "Expected empty error message")
				assert.Empty(t, errType, "Expected empty error type")
				assert.Empty(t, errParam, "Expected empty error param")
			} else {
				assert.Equal(t, tt.wantErrMsg, errMsg, "Error message mismatch")
				assert.Equal(t, tt.wantErrType, errType, "Error type mismatch")
				assert.Equal(t, tt.wantErrParam, errParam, "Error param mismatch")
			}
		})
	}
}

func TestFormatErrorDetails(t *testing.T) {
	t.Parallel() // Pure CPU-bound error detail formatting test
	tests := []struct {
		name        string
		errorMsg    string
		errorType   string
		errorParam  string
		wantDetails string
	}{
		{
			name:        "Empty message",
			errorMsg:    "",
			errorType:   "",
			errorParam:  "",
			wantDetails: "",
		},
		{
			name:        "Message only",
			errorMsg:    "An error occurred",
			errorType:   "",
			errorParam:  "",
			wantDetails: "API Error: An error occurred",
		},
		{
			name:        "Message and type",
			errorMsg:    "Rate limit exceeded",
			errorType:   "rate_limit",
			errorParam:  "",
			wantDetails: "API Error: Rate limit exceeded (Type: rate_limit)",
		},
		{
			name:        "Message, type, and param",
			errorMsg:    "Invalid parameter",
			errorType:   "invalid_request",
			errorParam:  "temperature",
			wantDetails: "API Error: Invalid parameter (Type: invalid_request) (Param: temperature)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := FormatErrorDetails(tt.errorMsg, tt.errorType, tt.errorParam)
			assert.Equal(t, tt.wantDetails, details)
		})
	}
}

func TestIsOpenRouterError(t *testing.T) {
	t.Parallel() // Pure CPU-bound error type validation test
	tests := []struct {
		name        string
		err         error
		wantIsORErr bool
	}{
		{
			name:        "Nil error",
			err:         nil,
			wantIsORErr: false,
		},
		{
			name:        "Standard error",
			err:         errors.New("standard error"),
			wantIsORErr: false,
		},
		{
			name:        "OpenRouter LLM error",
			err:         llm.New("openrouter", "", 0, "test error", "", nil, llm.CategoryAuth),
			wantIsORErr: true,
		},
		{
			name:        "Other provider LLM error",
			err:         llm.New("openai", "", 0, "test error", "", nil, llm.CategoryAuth),
			wantIsORErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmErr, isORErr := IsOpenRouterError(tt.err)
			assert.Equal(t, tt.wantIsORErr, isORErr)

			if tt.wantIsORErr {
				require.NotNil(t, llmErr)
				assert.Equal(t, "openrouter", llmErr.Provider)
			} else if tt.err == nil {
				assert.Nil(t, llmErr)
			}
		})
	}
}

func TestFormatAPIErrorFromResponse(t *testing.T) {
	t.Parallel() // Pure CPU-bound API error formatting test
	tests := []struct {
		name         string
		err          error
		statusCode   int
		responseBody []byte
		wantCategory llm.ErrorCategory
		wantProvider string
		wantContains string
	}{
		{
			name:         "Nil error",
			err:          nil,
			statusCode:   0,
			responseBody: nil,
			wantCategory: llm.CategoryUnknown,
			wantProvider: "",
			wantContains: "",
		},
		{
			name:         "Authentication error",
			err:          errors.New("invalid api key"),
			statusCode:   401,
			responseBody: []byte(`{"error": {"message": "Invalid API key", "type": "auth_error"}}`),
			wantCategory: llm.CategoryAuth,
			wantProvider: "openrouter",
			wantContains: "Invalid API key",
		},
		{
			name:         "Rate limit error",
			err:          errors.New("too many requests"),
			statusCode:   429,
			responseBody: []byte(`{"error": {"message": "Rate limit exceeded", "type": "rate_limit"}}`),
			wantCategory: llm.CategoryRateLimit,
			wantProvider: "openrouter",
			wantContains: "Rate limit exceeded",
		},
		{
			name:         "Invalid request error",
			err:          errors.New("invalid request"),
			statusCode:   400,
			responseBody: []byte(`{"error": {"message": "Invalid request parameters", "type": "invalid_request", "param": "temperature"}}`),
			wantCategory: llm.CategoryInvalidRequest,
			wantProvider: "openrouter",
			wantContains: "temperature",
		},
		{
			name:         "Server error",
			err:          errors.New("server error"),
			statusCode:   500,
			responseBody: []byte(`{"error": {"message": "Internal server error", "type": "server_error"}}`),
			wantCategory: llm.CategoryServer,
			wantProvider: "openrouter",
			wantContains: "Internal server error",
		},
		{
			name:         "Not found error",
			err:          errors.New("model not found"),
			statusCode:   404,
			responseBody: []byte(`{"error": {"message": "Model not found", "type": "not_found"}}`),
			wantCategory: llm.CategoryNotFound,
			wantProvider: "openrouter",
			wantContains: "Model not found",
		},
		{
			name:         "Error with existing LLMError",
			err:          llm.New("openrouter", "", 0, "existing llm error", "", nil, llm.CategoryAuth),
			statusCode:   401,
			responseBody: nil,
			wantCategory: llm.CategoryAuth,
			wantProvider: "openrouter",
			wantContains: "existing llm error",
		},
		{
			name:         "Error with empty response body",
			err:          errors.New("connection error"),
			statusCode:   0,
			responseBody: nil,
			wantCategory: llm.CategoryNetwork, // Should detect network error from the message
			wantProvider: "openrouter",
			wantContains: "connection error",
		},
		{
			name:         "Category detection from HTTP status only",
			err:          errors.New("http error"),
			statusCode:   403,
			responseBody: []byte(``),
			wantCategory: llm.CategoryAuth, // 403 should be categorized as auth
			wantProvider: "openrouter",
			wantContains: "http error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				result := FormatAPIErrorFromResponse(nil, tt.statusCode, tt.responseBody)
				assert.Nil(t, result)
				return
			}

			result := FormatAPIErrorFromResponse(tt.err, tt.statusCode, tt.responseBody)

			require.NotNil(t, result, "FormatAPIError returned nil")
			assert.Equal(t, tt.wantProvider, result.Provider)
			assert.Equal(t, tt.wantCategory, result.Category())

			if tt.wantContains != "" {
				assert.Contains(t, result.Error(), tt.wantContains)
			}

			// Verify suggestions are provided
			if tt.name != "Error with existing LLMError" { // Skip this check for existing LLMError
				assert.NotEmpty(t, result.Suggestion, "Error suggestion should not be empty")
			}
		})
	}
}

func TestFormatAPIError(t *testing.T) {
	t.Parallel() // Pure CPU-bound API error formatting test
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
			providerName:  "openrouter",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "",
			wantMsgPrefix: "",
		},
		{
			name:          "standard error",
			err:           errors.New("some error"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "network error",
			err:           errors.New("connection timeout"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryNetwork,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "auth error",
			err:           errors.New("authorization failed"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "rate limit error",
			err:           errors.New("rate limit exceeded"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "custom provider name",
			err:           errors.New("some error"),
			providerName:  "custom-provider",
			wantCategory:  llm.CategoryUnknown,
			wantProvider:  "custom-provider",
			wantMsgPrefix: "Error from custom-provider provider",
			checkUnwrap:   true,
		},
		{
			name: "existing LLMError from same provider",
			err: &llm.LLMError{
				Provider:      "openrouter",
				Message:       "Existing error",
				ErrorCategory: llm.CategoryAuth,
			},
			providerName:  "openrouter",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Existing error",
		},
		{
			name: "existing LLMError from different provider",
			err: &llm.LLMError{
				Provider:      "openai",
				Message:       "Error from openai",
				ErrorCategory: llm.CategoryAuth,
			},
			providerName:  "openrouter",
			wantCategory:  llm.CategoryAuth,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openai",
		},
		{
			name:          "insufficient credits error from message",
			err:           errors.New("billing quota exceeded"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryInsufficientCredits,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "content filtered error from message",
			err:           errors.New("content filtered by safety settings"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryContentFiltered,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "input limit error from message",
			err:           errors.New("token limit exceeded"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryInputLimit,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "cancelled error from message",
			err:           errors.New("request cancelled"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryCancelled,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "not found error from message",
			err:           errors.New("model not found"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryNotFound,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "invalid request error from message",
			err:           errors.New("invalid request parameters"),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
			checkUnwrap:   true,
		},
		{
			name:          "wrapped error",
			err:           fmt.Errorf("outer error: %w", errors.New("inner rate limit exceeded")),
			providerName:  "openrouter",
			wantCategory:  llm.CategoryRateLimit,
			wantProvider:  "openrouter",
			wantMsgPrefix: "Error from openrouter provider",
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
	t.Parallel() // Pure CPU-bound API error creation test
	tests := []struct {
		name           string
		category       llm.ErrorCategory
		errMsg         string
		originalErr    error
		details        string
		wantSuggestion bool
	}{
		{
			name:           "Auth error",
			category:       llm.CategoryAuth,
			errMsg:         "Authentication failed",
			originalErr:    errors.New("invalid api key"),
			details:        "API key invalid",
			wantSuggestion: true,
		},
		{
			name:           "Rate limit error",
			category:       llm.CategoryRateLimit,
			errMsg:         "Rate limit exceeded",
			originalErr:    errors.New("too many requests"),
			details:        "Wait and try again",
			wantSuggestion: true,
		},
		{
			name:           "Invalid request error",
			category:       llm.CategoryInvalidRequest,
			errMsg:         "Invalid parameters",
			originalErr:    errors.New("bad request"),
			details:        "Parameter 'temperature' out of range",
			wantSuggestion: true,
		},
		{
			name:           "Server error",
			category:       llm.CategoryServer,
			errMsg:         "Server error",
			originalErr:    errors.New("internal error"),
			details:        "OpenRouter servers returned 500",
			wantSuggestion: true,
		},
		{
			name:           "Network error",
			category:       llm.CategoryNetwork,
			errMsg:         "Network error",
			originalErr:    errors.New("connection failed"),
			details:        "Unable to reach OpenRouter API",
			wantSuggestion: true,
		},
		{
			name:           "Unknown error",
			category:       llm.CategoryUnknown,
			errMsg:         "Unknown error",
			originalErr:    errors.New("something went wrong"),
			details:        "Unexpected error occurred",
			wantSuggestion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateAPIError(tt.category, tt.errMsg, tt.originalErr, tt.details)

			require.NotNil(t, result)
			assert.Equal(t, "openrouter", result.Provider)

			// Compare the integer values directly
			gotCategory := result.Category()
			assert.Equal(t, int(tt.category), int(gotCategory))

			assert.Equal(t, tt.details, result.Details)
			assert.Contains(t, result.Error(), tt.errMsg)

			// Verify original error is wrapped
			assert.Equal(t, tt.originalErr.Error(), result.Unwrap().Error())

			// Verify suggestion is provided
			if tt.wantSuggestion {
				assert.NotEmpty(t, result.Suggestion)
			}
		})
	}
}
