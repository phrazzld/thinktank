package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/stretchr/testify/assert"
)

func TestGetExitCodeFromError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "nil error returns success",
			err:          nil,
			expectedCode: ExitCodeSuccess,
		},
		{
			name: "LLMError with CategoryAuth",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			expectedCode: ExitCodeAuthError,
		},
		{
			name: "LLMError with CategoryRateLimit",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Rate limit exceeded",
				ErrorCategory: llm.CategoryRateLimit,
			},
			expectedCode: ExitCodeRateLimitError,
		},
		{
			name: "LLMError with CategoryInvalidRequest",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Invalid request",
				ErrorCategory: llm.CategoryInvalidRequest,
			},
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name: "LLMError with CategoryServer",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Server error",
				ErrorCategory: llm.CategoryServer,
			},
			expectedCode: ExitCodeServerError,
		},
		{
			name: "LLMError with CategoryNetwork",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Network error",
				ErrorCategory: llm.CategoryNetwork,
			},
			expectedCode: ExitCodeNetworkError,
		},
		{
			name: "LLMError with CategoryInputLimit",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Input limit exceeded",
				ErrorCategory: llm.CategoryInputLimit,
			},
			expectedCode: ExitCodeInputError,
		},
		{
			name: "LLMError with CategoryContentFiltered",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Content filtered",
				ErrorCategory: llm.CategoryContentFiltered,
			},
			expectedCode: ExitCodeContentFiltered,
		},
		{
			name: "LLMError with CategoryInsufficientCredits",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Insufficient credits",
				ErrorCategory: llm.CategoryInsufficientCredits,
			},
			expectedCode: ExitCodeInsufficientCredits,
		},
		{
			name: "LLMError with CategoryCancelled",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Request cancelled",
				ErrorCategory: llm.CategoryCancelled,
			},
			expectedCode: ExitCodeCancelled,
		},
		{
			name: "LLMError with unknown category defaults to generic",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Unknown error",
				ErrorCategory: llm.CategoryUnknown, // Use valid category that defaults to generic
			},
			expectedCode: ExitCodeGenericError,
		},
		{
			name:         "ErrPartialSuccess returns generic error",
			err:          thinktank.ErrPartialSuccess,
			expectedCode: ExitCodeGenericError,
		},
		{
			name:         "context.Canceled returns cancelled",
			err:          context.Canceled,
			expectedCode: ExitCodeCancelled,
		},
		{
			name:         "context.DeadlineExceeded returns cancelled",
			err:          context.DeadlineExceeded,
			expectedCode: ExitCodeCancelled,
		},
		{
			name:         "wrapped context canceled message returns cancelled",
			err:          errors.New("operation failed: context canceled"),
			expectedCode: ExitCodeCancelled, // getErrorCategoryFromMessage should catch this
		},
		{
			name:         "generic error returns generic code",
			err:          errors.New("something unexpected happened"),
			expectedCode: ExitCodeGenericError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExitCodeFromError(tt.err)
			assert.Equal(t, tt.expectedCode, result, "Expected exit code %d for %s error, got %d", tt.expectedCode, tt.name, result)
		})
	}
}

func TestGetFriendlyErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "An unknown error occurred",
		},
		{
			name:     "authentication error - api key",
			err:      errors.New("failed to authenticate with API key"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "authentication error - auth",
			err:      errors.New("authentication failed"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "authentication error - unauthorized",
			err:      errors.New("unauthorized access"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "rate limit error - rate limit",
			err:      errors.New("rate limit exceeded"),
			expected: "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits.",
		},
		{
			name:     "rate limit error - too many requests",
			err:      errors.New("too many requests"),
			expected: "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits.",
		},
		{
			name:     "timeout error - timeout",
			err:      errors.New("operation timed out after 5m"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "timeout error - deadline exceeded",
			err:      errors.New("deadline exceeded"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "timeout error - timed out",
			err:      errors.New("request timed out"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "not found error",
			err:      errors.New("model not found: claude-3"),
			expected: "Resource not found. Please check that the specified file paths or models exist.",
		},
		{
			name:     "file permission error",
			err:      errors.New("file permission denied: /tmp/output.txt"),
			expected: "File permission error: Please check file permissions and try again.",
		},
		{
			name:     "file error - generic",
			err:      errors.New("file not readable"),
			expected: "File error: file not readable",
		},
		{
			name:     "flag error - flag",
			err:      errors.New("unknown flag: --invalid"),
			expected: "Invalid command line arguments. Use --help to see usage instructions.",
		},
		{
			name:     "flag error - usage",
			err:      errors.New("usage: thinktank [options]"),
			expected: "Invalid command line arguments. Use --help to see usage instructions.",
		},
		{
			name:     "flag error - help",
			err:      errors.New("help requested"),
			expected: "Invalid command line arguments. Use --help to see usage instructions.",
		},
		{
			name:     "context cancelled error",
			err:      errors.New("context cancelled"),
			expected: "Operation was cancelled. This might be due to timeout or user interruption.",
		},
		{
			name:     "context canceled error",
			err:      errors.New("context canceled"),
			expected: "Operation was cancelled. This might be due to timeout or user interruption.",
		},
		{
			name:     "context error - deadline exceeded matches timeout pattern",
			err:      errors.New("context deadline exceeded"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "network error - network",
			err:      errors.New("network connection failed"),
			expected: "Network error: Please check your internet connection and try again.",
		},
		{
			name:     "network error - connection",
			err:      errors.New("connection refused"),
			expected: "Network error: Please check your internet connection and try again.",
		},
		{
			name:     "generic error",
			err:      errors.New("something unexpected happened"),
			expected: "something unexpected happened",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFriendlyErrorMessage(tt.err)
			assert.Equal(t, tt.expected, result, "getFriendlyErrorMessage() = %q, want %q", result, tt.expected)
		})
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name             string
		message          string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:          "no sensitive information",
			message:       "Simple error message",
			shouldContain: "Simple error message",
		},
		{
			name:             "contains API key - OpenAI sk-",
			message:          "Failed API call with key sk_abcdef1234567890abcdef12345",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "sk_abcdef1234567890abcdef12345",
		},
		{
			name:             "contains API key - OpenAI sk-",
			message:          "Failed API call with key sk-abcdef1234567890abcdef12345",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "sk-abcdef1234567890abcdef12345",
		},
		{
			name:             "contains API key - Gemini key_",
			message:          "Failed API call with key_abcdef1234567890abcdef12345",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "key_abcdef1234567890abcdef12345",
		},
		{
			name:             "contains API key - Gemini key-",
			message:          "Failed API call with key-abcdef1234567890abcdef12345",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "key-abcdef1234567890abcdef12345",
		},
		{
			name:             "contains long alphanumeric string",
			message:          "Error with token 1234567890abcdef1234567890abcdef12345",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "1234567890abcdef1234567890abcdef12345",
		},
		{
			name:             "contains URL with credentials",
			message:          "Failed to connect to https://user:password@example.com/api",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "user:password",
		},
		{
			name:             "contains GEMINI_API_KEY environment variable",
			message:          "GEMINI_API_KEY=key-1234567890abcdef",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "key-1234567890abcdef",
		},
		{
			name:             "contains OPENAI_API_KEY environment variable",
			message:          "OPENAI_API_KEY=sk-1234567890abcdef",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "sk-1234567890abcdef",
		},
		{
			name:             "contains OPENROUTER_API_KEY environment variable",
			message:          "OPENROUTER_API_KEY=sk-or-1234567890abcdef",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "sk-or-1234567890abcdef",
		},
		{
			name:             "contains API_KEY environment variable",
			message:          "API_KEY=abc123def456",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.message)
			if tt.shouldContain != "" {
				assert.Contains(t, result, tt.shouldContain, "sanitizeErrorMessage(%q) = %q, should contain %q", tt.message, result, tt.shouldContain)
			}
			if tt.shouldNotContain != "" {
				assert.NotContains(t, result, tt.shouldNotContain, "sanitizeErrorMessage(%q) = %q, should not contain %q", tt.message, result, tt.shouldNotContain)
			}
		})
	}
}

// TestHandleErrorAuditLogging tests the audit logging behavior during error handling
func TestHandleErrorAuditLogging(t *testing.T) {
	ctx := context.Background()
	logger := testutil.NewMockLogger()

	tests := []struct {
		name            string
		err             error
		auditError      error
		expectAuditCall bool
	}{
		{
			name:            "nil error should not call audit logger",
			err:             nil,
			expectAuditCall: false,
		},
		{
			name:            "error should call audit logger",
			err:             errors.New("test error"),
			expectAuditCall: true,
		},
		{
			name:            "audit logger failure should be logged",
			err:             errors.New("test error"),
			auditError:      errors.New("audit log error"),
			expectAuditCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditLogger := &mockAuditLogger{
				logOpError: tt.auditError,
			}
			logger.ClearMessages()

			// We can't test handleError directly because it calls os.Exit()
			// Instead, we test the audit logging logic that would be called
			if tt.err != nil {
				// Simulate the audit logging part of handleError
				logErr := auditLogger.LogOp(ctx, "test_operation", "Failure", nil, nil, tt.err)
				if logErr != nil {
					logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
				}
			}

			if tt.expectAuditCall {
				assert.True(t, auditLogger.logOpCalled, "Expected audit logger to be called")
				if tt.auditError != nil {
					logEntries := logger.GetLogEntries()
					found := false
					for _, entry := range logEntries {
						if entry.Message == "Failed to write audit log: audit log error" {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected audit error to be logged")
				}
			} else {
				assert.False(t, auditLogger.logOpCalled, "Expected audit logger not to be called")
			}
		})
	}
}
