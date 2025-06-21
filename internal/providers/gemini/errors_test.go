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

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name         string
		responseBody []byte
		wantMessage  string
		wantStatus   string
		wantCode     int
	}{
		{
			name:         "empty response body",
			responseBody: []byte{},
			wantMessage:  "",
			wantStatus:   "",
			wantCode:     0,
		},
		{
			name:         "nil response body",
			responseBody: nil,
			wantMessage:  "",
			wantStatus:   "",
			wantCode:     0,
		},
		{
			name:         "invalid JSON",
			responseBody: []byte(`{invalid json}`),
			wantMessage:  "",
			wantStatus:   "",
			wantCode:     0,
		},
		{
			name:         "minimal valid error response",
			responseBody: []byte(`{"error": {"message": "Authentication failed"}}`),
			wantMessage:  "Authentication failed",
			wantStatus:   "",
			wantCode:     0,
		},
		{
			name:         "complete error response",
			responseBody: []byte(`{"error": {"message": "Invalid parameter value", "status": "INVALID_ARGUMENT", "code": 400}}`),
			wantMessage:  "Invalid parameter value",
			wantStatus:   "INVALID_ARGUMENT",
			wantCode:     400,
		},
		{
			name:         "error response with status and code only",
			responseBody: []byte(`{"error": {"message": "Rate limit exceeded", "status": "RESOURCE_EXHAUSTED", "code": 429}}`),
			wantMessage:  "Rate limit exceeded",
			wantStatus:   "RESOURCE_EXHAUSTED",
			wantCode:     429,
		},
		{
			name:         "error response with missing fields",
			responseBody: []byte(`{"error": {"status": "INTERNAL", "code": 500}}`),
			wantMessage:  "",
			wantStatus:   "INTERNAL",
			wantCode:     500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, status, code := ParseErrorResponse(tt.responseBody)

			assert.Equal(t, tt.wantMessage, message, "message mismatch")
			assert.Equal(t, tt.wantStatus, status, "status mismatch")
			assert.Equal(t, tt.wantCode, code, "code mismatch")
		})
	}
}

func TestFormatErrorDetails(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		errorStatus  string
		errorCode    int
		wantDetails  string
	}{
		{
			name:         "empty message returns empty",
			errorMessage: "",
			errorStatus:  "INVALID_ARGUMENT",
			errorCode:    400,
			wantDetails:  "",
		},
		{
			name:         "message only",
			errorMessage: "Authentication failed",
			errorStatus:  "",
			errorCode:    0,
			wantDetails:  "API Error: Authentication failed",
		},
		{
			name:         "message with status",
			errorMessage: "Invalid request",
			errorStatus:  "INVALID_ARGUMENT",
			errorCode:    0,
			wantDetails:  "API Error: Invalid request (Status: INVALID_ARGUMENT)",
		},
		{
			name:         "message with code",
			errorMessage: "Context length exceeded",
			errorStatus:  "",
			errorCode:    400,
			wantDetails:  "API Error: Context length exceeded (Code: 400)",
		},
		{
			name:         "message with status and code",
			errorMessage: "Rate limit exceeded",
			errorStatus:  "RESOURCE_EXHAUSTED",
			errorCode:    429,
			wantDetails:  "API Error: Rate limit exceeded (Status: RESOURCE_EXHAUSTED) (Code: 429)",
		},
		{
			name:         "complete error details",
			errorMessage: "Invalid parameter value",
			errorStatus:  "INVALID_ARGUMENT",
			errorCode:    400,
			wantDetails:  "API Error: Invalid parameter value (Status: INVALID_ARGUMENT) (Code: 400)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorDetails(tt.errorMessage, tt.errorStatus, tt.errorCode)
			assert.Equal(t, tt.wantDetails, result)
		})
	}
}

func TestMapGeminiErrorToCategory(t *testing.T) {
	tests := []struct {
		name         string
		errorStatus  string
		errorCode    int
		wantCategory llm.ErrorCategory
	}{
		// Status mapping tests
		{
			name:         "UNAUTHENTICATED status",
			errorStatus:  "UNAUTHENTICATED",
			errorCode:    0,
			wantCategory: llm.CategoryAuth,
		},
		{
			name:         "PERMISSION_DENIED status",
			errorStatus:  "PERMISSION_DENIED",
			errorCode:    0,
			wantCategory: llm.CategoryAuth,
		},
		{
			name:         "RESOURCE_EXHAUSTED status",
			errorStatus:  "RESOURCE_EXHAUSTED",
			errorCode:    0,
			wantCategory: llm.CategoryRateLimit,
		},
		{
			name:         "INVALID_ARGUMENT status",
			errorStatus:  "INVALID_ARGUMENT",
			errorCode:    0,
			wantCategory: llm.CategoryInvalidRequest,
		},
		{
			name:         "NOT_FOUND status",
			errorStatus:  "NOT_FOUND",
			errorCode:    0,
			wantCategory: llm.CategoryNotFound,
		},
		{
			name:         "UNAVAILABLE status",
			errorStatus:  "UNAVAILABLE",
			errorCode:    0,
			wantCategory: llm.CategoryServer,
		},
		{
			name:         "INTERNAL status",
			errorStatus:  "INTERNAL",
			errorCode:    0,
			wantCategory: llm.CategoryServer,
		},
		{
			name:         "DEADLINE_EXCEEDED status",
			errorStatus:  "DEADLINE_EXCEEDED",
			errorCode:    0,
			wantCategory: llm.CategoryCancelled,
		},
		{
			name:         "OUT_OF_RANGE status",
			errorStatus:  "OUT_OF_RANGE",
			errorCode:    0,
			wantCategory: llm.CategoryInputLimit,
		},

		// Error code mapping tests
		{
			name:         "401 error code",
			errorStatus:  "",
			errorCode:    401,
			wantCategory: llm.CategoryAuth,
		},
		{
			name:         "403 error code",
			errorStatus:  "",
			errorCode:    403,
			wantCategory: llm.CategoryAuth,
		},
		{
			name:         "404 error code",
			errorStatus:  "",
			errorCode:    404,
			wantCategory: llm.CategoryNotFound,
		},
		{
			name:         "429 error code",
			errorStatus:  "",
			errorCode:    429,
			wantCategory: llm.CategoryRateLimit,
		},
		{
			name:         "400 error code",
			errorStatus:  "",
			errorCode:    400,
			wantCategory: llm.CategoryInvalidRequest,
		},
		{
			name:         "500 error code",
			errorStatus:  "",
			errorCode:    500,
			wantCategory: llm.CategoryServer,
		},
		{
			name:         "502 error code",
			errorStatus:  "",
			errorCode:    502,
			wantCategory: llm.CategoryServer,
		},
		{
			name:         "503 error code",
			errorStatus:  "",
			errorCode:    503,
			wantCategory: llm.CategoryServer,
		},

		// Status takes precedence over code
		{
			name:         "status overrides code",
			errorStatus:  "UNAUTHENTICATED",
			errorCode:    429,
			wantCategory: llm.CategoryAuth, // Should be auth, not rate limit
		},

		// Unknown cases
		{
			name:         "unknown status",
			errorStatus:  "UNKNOWN_STATUS",
			errorCode:    0,
			wantCategory: llm.CategoryUnknown,
		},
		{
			name:         "unknown error code",
			errorStatus:  "",
			errorCode:    999,
			wantCategory: llm.CategoryUnknown,
		},
		{
			name:         "empty status and code",
			errorStatus:  "",
			errorCode:    0,
			wantCategory: llm.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapGeminiErrorToCategory(tt.errorStatus, tt.errorCode)
			assert.Equal(t, tt.wantCategory, result)
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
			details:       "prompt format invalid",
			wantCategory:  llm.CategoryInvalidRequest,
			wantProvider:  "gemini",
			wantMsgPrefix: "Invalid request parameters",
			wantDetails:   "prompt format invalid",
		},
		{
			name:          "insufficient credits error",
			category:      llm.CategoryInsufficientCredits,
			errMsg:        "Billing quota exceeded",
			originalErr:   fmt.Errorf("quota exceeded"),
			details:       "Check billing account",
			wantCategory:  llm.CategoryInsufficientCredits,
			wantProvider:  "gemini",
			wantMsgPrefix: "Billing quota exceeded",
			wantDetails:   "Check billing account",
		},
		{
			name:          "not found error",
			category:      llm.CategoryNotFound,
			errMsg:        "Model not found",
			originalErr:   fmt.Errorf("model not found"),
			details:       "Check model name",
			wantCategory:  llm.CategoryNotFound,
			wantProvider:  "gemini",
			wantMsgPrefix: "Model not found",
			wantDetails:   "Check model name",
		},
		{
			name:          "server error",
			category:      llm.CategoryServer,
			errMsg:        "Internal server error",
			originalErr:   fmt.Errorf("server error"),
			details:       "Temporary issue",
			wantCategory:  llm.CategoryServer,
			wantProvider:  "gemini",
			wantMsgPrefix: "Internal server error",
			wantDetails:   "Temporary issue",
		},
		{
			name:          "input limit error",
			category:      llm.CategoryInputLimit,
			errMsg:        "Context length exceeded",
			originalErr:   fmt.Errorf("input too long"),
			details:       "Reduce input size",
			wantCategory:  llm.CategoryInputLimit,
			wantProvider:  "gemini",
			wantMsgPrefix: "Context length exceeded",
			wantDetails:   "Reduce input size",
		},
		{
			name:          "content filtered error",
			category:      llm.CategoryContentFiltered,
			errMsg:        "Content was filtered",
			originalErr:   fmt.Errorf("safety filter triggered"),
			details:       "Modify prompt",
			wantCategory:  llm.CategoryContentFiltered,
			wantProvider:  "gemini",
			wantMsgPrefix: "Content was filtered",
			wantDetails:   "Modify prompt",
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

func TestIsSafetyFilter(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		want         bool
	}{
		{
			name:         "safety keyword",
			errorMessage: "Content blocked by safety filters",
			want:         true,
		},
		{
			name:         "blocked keyword",
			errorMessage: "Request was blocked",
			want:         true,
		},
		{
			name:         "content filter keyword",
			errorMessage: "Triggered content filter",
			want:         true,
		},
		{
			name:         "content policy keyword",
			errorMessage: "Violates content policy",
			want:         true,
		},
		{
			name:         "case insensitive safety",
			errorMessage: "SAFETY filter triggered",
			want:         true,
		},
		{
			name:         "case insensitive blocked",
			errorMessage: "Request BLOCKED by system",
			want:         true,
		},
		{
			name:         "no safety keywords",
			errorMessage: "Network timeout error",
			want:         false,
		},
		{
			name:         "empty message",
			errorMessage: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSafetyFilter(tt.errorMessage)
			assert.Equal(t, tt.want, result)
		})
	}
}
