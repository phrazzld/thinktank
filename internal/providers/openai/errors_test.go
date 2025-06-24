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
	t.Parallel() // Pure CPU-bound error type validation test
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
	t.Parallel() // Pure CPU-bound API error formatting test
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
	t.Parallel() // Pure CPU-bound API error creation test
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

func TestParseErrorResponse(t *testing.T) {
	t.Parallel() // Pure CPU-bound JSON response parsing test
	tests := []struct {
		name         string
		responseBody []byte
		wantMessage  string
		wantType     string
		wantCode     string
		wantParam    string
	}{
		{
			name:         "empty response body",
			responseBody: []byte{},
			wantMessage:  "",
			wantType:     "",
			wantCode:     "",
			wantParam:    "",
		},
		{
			name:         "nil response body",
			responseBody: nil,
			wantMessage:  "",
			wantType:     "",
			wantCode:     "",
			wantParam:    "",
		},
		{
			name:         "invalid JSON",
			responseBody: []byte(`{invalid json}`),
			wantMessage:  "",
			wantType:     "",
			wantCode:     "",
			wantParam:    "",
		},
		{
			name:         "minimal valid error response",
			responseBody: []byte(`{"error": {"message": "Authentication failed"}}`),
			wantMessage:  "Authentication failed",
			wantType:     "",
			wantCode:     "",
			wantParam:    "",
		},
		{
			name:         "complete error response",
			responseBody: []byte(`{"error": {"message": "Invalid parameter value", "type": "invalid_request_error", "code": "invalid_parameter", "param": "temperature"}}`),
			wantMessage:  "Invalid parameter value",
			wantType:     "invalid_request_error",
			wantCode:     "invalid_parameter",
			wantParam:    "temperature",
		},
		{
			name:         "error response with type and code only",
			responseBody: []byte(`{"error": {"message": "Rate limit exceeded", "type": "rate_limit_error", "code": "rate_limit_exceeded"}}`),
			wantMessage:  "Rate limit exceeded",
			wantType:     "rate_limit_error",
			wantCode:     "rate_limit_exceeded",
			wantParam:    "",
		},
		{
			name:         "error response with missing fields",
			responseBody: []byte(`{"error": {"type": "server_error"}}`),
			wantMessage:  "",
			wantType:     "server_error",
			wantCode:     "",
			wantParam:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, errorType, code, param := ParseErrorResponse(tt.responseBody)

			assert.Equal(t, tt.wantMessage, message, "message mismatch")
			assert.Equal(t, tt.wantType, errorType, "type mismatch")
			assert.Equal(t, tt.wantCode, code, "code mismatch")
			assert.Equal(t, tt.wantParam, param, "param mismatch")
		})
	}
}

func TestFormatErrorDetails(t *testing.T) {
	t.Parallel() // Pure CPU-bound error detail formatting test
	tests := []struct {
		name         string
		errorMessage string
		errorType    string
		errorCode    string
		errorParam   string
		wantDetails  string
	}{
		{
			name:         "empty message returns empty",
			errorMessage: "",
			errorType:    "invalid_request_error",
			errorCode:    "invalid_parameter",
			errorParam:   "temperature",
			wantDetails:  "",
		},
		{
			name:         "message only",
			errorMessage: "Authentication failed",
			errorType:    "",
			errorCode:    "",
			errorParam:   "",
			wantDetails:  "API Error: Authentication failed",
		},
		{
			name:         "message with type",
			errorMessage: "Invalid request",
			errorType:    "invalid_request_error",
			errorCode:    "",
			errorParam:   "",
			wantDetails:  "API Error: Invalid request (Type: invalid_request_error)",
		},
		{
			name:         "message with code",
			errorMessage: "Context length exceeded",
			errorType:    "",
			errorCode:    "context_length_exceeded",
			errorParam:   "",
			wantDetails:  "API Error: Context length exceeded (Code: context_length_exceeded)",
		},
		{
			name:         "message with param",
			errorMessage: "Invalid parameter value",
			errorType:    "",
			errorCode:    "",
			errorParam:   "temperature",
			wantDetails:  "API Error: Invalid parameter value (Param: temperature)",
		},
		{
			name:         "message with type and code",
			errorMessage: "Rate limit exceeded",
			errorType:    "rate_limit_error",
			errorCode:    "rate_limit_exceeded",
			errorParam:   "",
			wantDetails:  "API Error: Rate limit exceeded (Type: rate_limit_error) (Code: rate_limit_exceeded)",
		},
		{
			name:         "complete error details",
			errorMessage: "Invalid parameter value",
			errorType:    "invalid_request_error",
			errorCode:    "invalid_parameter",
			errorParam:   "temperature",
			wantDetails:  "API Error: Invalid parameter value (Type: invalid_request_error) (Code: invalid_parameter) (Param: temperature)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorDetails(tt.errorMessage, tt.errorType, tt.errorCode, tt.errorParam)
			assert.Equal(t, tt.wantDetails, result)
		})
	}
}

func TestMapOpenAIErrorToCategory(t *testing.T) {
	t.Parallel() // Pure CPU-bound error categorization mapping test
	tests := []struct {
		name         string
		errorType    string
		errorCode    string
		wantCategory llm.ErrorCategory
	}{
		// Error type mapping
		{
			name:         "authentication error type",
			errorType:    "authentication_error",
			errorCode:    "",
			wantCategory: llm.CategoryAuth,
		},
		{
			name:         "invalid request error type",
			errorType:    "invalid_request_error",
			errorCode:    "",
			wantCategory: llm.CategoryInvalidRequest,
		},
		{
			name:         "rate limit error type",
			errorType:    "rate_limit_error",
			errorCode:    "",
			wantCategory: llm.CategoryRateLimit,
		},
		{
			name:         "server error type",
			errorType:    "server_error",
			errorCode:    "",
			wantCategory: llm.CategoryServer,
		},

		// Error code mapping
		{
			name:         "context length exceeded code",
			errorType:    "",
			errorCode:    "context_length_exceeded",
			wantCategory: llm.CategoryInputLimit,
		},
		{
			name:         "model not found code",
			errorType:    "",
			errorCode:    "model_not_found",
			wantCategory: llm.CategoryNotFound,
		},
		{
			name:         "insufficient quota code",
			errorType:    "",
			errorCode:    "insufficient_quota",
			wantCategory: llm.CategoryInsufficientCredits,
		},
		{
			name:         "content filter code",
			errorType:    "",
			errorCode:    "content_filter",
			wantCategory: llm.CategoryContentFiltered,
		},

		// Type takes precedence over code
		{
			name:         "type overrides code",
			errorType:    "authentication_error",
			errorCode:    "context_length_exceeded",
			wantCategory: llm.CategoryAuth, // Should be auth, not input limit
		},

		// Unknown cases
		{
			name:         "unknown error type",
			errorType:    "unknown_error_type",
			errorCode:    "",
			wantCategory: llm.CategoryUnknown,
		},
		{
			name:         "unknown error code",
			errorType:    "",
			errorCode:    "unknown_error_code",
			wantCategory: llm.CategoryUnknown,
		},
		{
			name:         "empty type and code",
			errorType:    "",
			errorCode:    "",
			wantCategory: llm.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapOpenAIErrorToCategory(tt.errorType, tt.errorCode)
			assert.Equal(t, tt.wantCategory, result)
		})
	}
}
