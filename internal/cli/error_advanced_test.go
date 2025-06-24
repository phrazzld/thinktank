package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/stretchr/testify/assert"
)

func TestGenerateErrorMessage(t *testing.T) {
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
			name: "LLMError with auth category - enhanced message",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			expected: "Authentication failed. Please check your API key and permissions.",
		},
		{
			name: "LLMError with rate limit category - enhanced message",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Rate limit exceeded",
				ErrorCategory: llm.CategoryRateLimit,
			},
			expected: "Rate limit exceeded. Please try again later or adjust rate limits.",
		},
		{
			name: "LLMError with auth category and custom UserFacingError",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
				Suggestion:    "Check your API key",
			},
			expected: "Authentication failed\n\nSuggestion: Check your API key",
		},
		{
			name: "LLMError with rate limit category and custom UserFacingError",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Rate limit exceeded",
				ErrorCategory: llm.CategoryRateLimit,
				Suggestion:    "Wait before retrying",
			},
			expected: "Rate limit exceeded\n\nSuggestion: Wait before retrying",
		},
		{
			name: "LLMError with network category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Network connection failed",
				ErrorCategory: llm.CategoryNetwork,
			},
			expected: "Network connection failed",
		},
		{
			name: "LLMError with server category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Internal server error",
				ErrorCategory: llm.CategoryServer,
			},
			expected: "Internal server error",
		},
		{
			name: "LLMError with input limit category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Token limit exceeded",
				ErrorCategory: llm.CategoryInputLimit,
			},
			expected: "Token limit exceeded",
		},
		{
			name: "LLMError with content filtered category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Content was filtered",
				ErrorCategory: llm.CategoryContentFiltered,
			},
			expected: "Content was filtered",
		},
		{
			name: "LLMError with insufficient credits category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Insufficient credits",
				ErrorCategory: llm.CategoryInsufficientCredits,
			},
			expected: "Insufficient credits",
		},
		{
			name: "LLMError with cancelled category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Request cancelled",
				ErrorCategory: llm.CategoryCancelled,
			},
			expected: "Request cancelled",
		},
		{
			name: "LLMError with not found category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Model not found",
				ErrorCategory: llm.CategoryNotFound,
			},
			expected: "Model not found",
		},
		{
			name: "LLMError with unknown category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Unknown error type",
				ErrorCategory: llm.CategoryUnknown,
			},
			expected: "Unknown error type",
		},
		{
			name: "LLMError with invalid request category (no enhancement)",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Invalid request format",
				ErrorCategory: llm.CategoryInvalidRequest,
			},
			expected: "Invalid request format",
		},
		{
			name: "Non-LLMError CategorizedError",
			err: &llm.MockError{
				Message:       "Mock categorized error",
				ErrorCategory: llm.CategoryAuth,
			},
			expected: "Mock categorized error",
		},
		{
			name:     "partial success error",
			err:      thinktank.ErrPartialSuccess,
			expected: "Some model executions failed, but partial results were generated. Use --partial-success-ok flag to exit with success code in this case.",
		},
		{
			name:     "generic error with auth pattern",
			err:      errors.New("failed to authenticate with API key"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "generic error with rate limit pattern",
			err:      errors.New("rate limit exceeded"),
			expected: "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits.",
		},
		{
			name:     "generic error with timeout pattern",
			err:      errors.New("operation timed out after 5m"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "generic error with context deadline pattern",
			err:      errors.New("context deadline exceeded"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "generic error with file permission pattern",
			err:      errors.New("file permission denied"),
			expected: "File permission error: Please check file permissions and try again.",
		},
		{
			name:     "generic error with network pattern",
			err:      errors.New("network connection failed"),
			expected: "Network error: Please check your internet connection and try again.",
		},
		{
			name:     "generic error with cancellation pattern",
			err:      errors.New("context cancelled"),
			expected: "Operation was cancelled. This might be due to timeout or user interruption.",
		},
		{
			name:     "generic error with no known pattern",
			err:      errors.New("something unexpected happened"),
			expected: "something unexpected happened",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateErrorMessage(tt.err)
			assert.Equal(t, tt.expected, result, "generateErrorMessage() = %q, want %q", result, tt.expected)
		})
	}
}

func TestProcessError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		err            error
		operation      string
		auditError     error
		expectedResult *ErrorProcessingResult
	}{
		{
			name:      "nil error returns no-op result",
			err:       nil,
			operation: "test_operation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeSuccess,
				UserMessage: "",
				ShouldExit:  false,
				AuditLogged: false,
				AuditError:  nil,
			},
		},
		{
			name: "auth error with successful audit logging",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			},
			operation: "authentication",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeAuthError,
				UserMessage: "Authentication failed. Please check your API key and permissions.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "rate limit error with audit failure",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Rate limit exceeded",
				ErrorCategory: llm.CategoryRateLimit,
			},
			operation:  "rate_limiting",
			auditError: errors.New("audit log write failed"),
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeRateLimitError,
				UserMessage: "Rate limit exceeded. Please try again later or adjust rate limits.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  errors.New("audit log write failed"),
			},
		},
		{
			name: "server error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Internal server error",
				ErrorCategory: llm.CategoryServer,
			},
			operation: "api_call",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeServerError,
				UserMessage: "Internal server error",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "network error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Network connection failed",
				ErrorCategory: llm.CategoryNetwork,
			},
			operation: "network_operation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeNetworkError,
				UserMessage: "Network connection failed",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "input limit error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Token limit exceeded",
				ErrorCategory: llm.CategoryInputLimit,
			},
			operation: "input_processing",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeInputError,
				UserMessage: "Token limit exceeded",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "content filtered error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Content was filtered",
				ErrorCategory: llm.CategoryContentFiltered,
			},
			operation: "content_filtering",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeContentFiltered,
				UserMessage: "Content was filtered",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "insufficient credits error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Insufficient credits",
				ErrorCategory: llm.CategoryInsufficientCredits,
			},
			operation: "billing",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeInsufficientCredits,
				UserMessage: "Insufficient credits",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "cancelled error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Request cancelled",
				ErrorCategory: llm.CategoryCancelled,
			},
			operation: "cancellation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeCancelled,
				UserMessage: "Request cancelled",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "not found error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Model not found",
				ErrorCategory: llm.CategoryNotFound,
			},
			operation: "model_lookup",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeGenericError,
				UserMessage: "Model not found",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name: "invalid request error",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Invalid request format",
				ErrorCategory: llm.CategoryInvalidRequest,
			},
			operation: "request_validation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeInvalidRequest,
				UserMessage: "Invalid request format",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name:      "partial success error",
			err:       thinktank.ErrPartialSuccess,
			operation: "execution",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeGenericError,
				UserMessage: "Some model executions failed, but partial results were generated. Use --partial-success-ok flag to exit with success code in this case.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name:      "context canceled error",
			err:       context.Canceled,
			operation: "context_operation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeCancelled,
				UserMessage: "Operation was cancelled. This might be due to timeout or user interruption.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name:      "context deadline exceeded error",
			err:       context.DeadlineExceeded,
			operation: "timeout_operation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeCancelled,
				UserMessage: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name:      "filesystem permission error",
			err:       errors.New("file permission denied: /path/to/file"),
			operation: "file_operation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeGenericError,
				UserMessage: "File permission error: Please check file permissions and try again.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name:      "network connection error",
			err:       errors.New("network connection failed: dial tcp refused"),
			operation: "network_dial",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeGenericError,
				UserMessage: "Network error: Please check your internet connection and try again.",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
		{
			name:      "generic error",
			err:       errors.New("something unexpected happened"),
			operation: "generic_operation",
			expectedResult: &ErrorProcessingResult{
				ExitCode:    ExitCodeGenericError,
				UserMessage: "something unexpected happened",
				ShouldExit:  true,
				AuditLogged: true,
				AuditError:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := testutil.NewMockLogger()
			auditLogger := &mockAuditLogger{logOpError: tt.auditError}

			result := processError(ctx, tt.err, logger, auditLogger, tt.operation)

			assert.Equal(t, tt.expectedResult.ExitCode, result.ExitCode)
			assert.Equal(t, tt.expectedResult.UserMessage, result.UserMessage)
			assert.Equal(t, tt.expectedResult.ShouldExit, result.ShouldExit)
			assert.Equal(t, tt.expectedResult.AuditLogged, result.AuditLogged)

			if tt.expectedResult.AuditError != nil {
				assert.Error(t, result.AuditError)
				assert.Equal(t, tt.expectedResult.AuditError.Error(), result.AuditError.Error())
			} else {
				assert.NoError(t, result.AuditError)
			}

			// Verify audit logging was called when expected
			if tt.expectedResult.AuditLogged {
				assert.True(t, auditLogger.logOpCalled, "Expected audit logger to be called")
			} else {
				assert.False(t, auditLogger.logOpCalled, "Expected audit logger not to be called")
			}
		})
	}
}

func TestHandleErrorOrchestration(t *testing.T) {
	// Test that handleError properly orchestrates processError functionality
	// We can't test handleError directly due to os.Exit(), but we can verify
	// that the orchestration logic works through processError integration

	ctx := context.Background()
	logger := testutil.NewMockLogger()
	auditLogger := &mockAuditLogger{}

	t.Run("nil error should not exit", func(t *testing.T) {
		result := processError(ctx, nil, logger, auditLogger, "test_operation")

		// Verify that handleError would return early (not exit) for nil errors
		assert.False(t, result.ShouldExit)
		assert.Equal(t, ExitCodeSuccess, result.ExitCode)
		assert.Empty(t, result.UserMessage)
	})

	t.Run("error orchestration produces expected result", func(t *testing.T) {
		testErr := &llm.LLMError{
			Provider:      "test",
			Message:       "Authentication failed",
			ErrorCategory: llm.CategoryAuth,
		}

		result := processError(ctx, testErr, logger, auditLogger, "authentication")

		// Verify that handleError would exit with correct code and message
		assert.True(t, result.ShouldExit)
		assert.Equal(t, ExitCodeAuthError, result.ExitCode)
		assert.Equal(t, "Authentication failed. Please check your API key and permissions.", result.UserMessage)
		assert.True(t, result.AuditLogged)

		// Verify the orchestration components were called
		assert.True(t, auditLogger.logOpCalled)

		// Verify logger was called for the error
		logEntries := logger.GetLogEntries()
		found := false
		for _, entry := range logEntries {
			if strings.Contains(entry.Message, "Error: Authentication failed") {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error to be logged")
	})
}

// TestErrorWrappingAndUnwrapping tests complex error wrapping scenarios
func TestErrorWrappingAndUnwrapping(t *testing.T) {
	t.Run("wrapped LLMError maintains category", func(t *testing.T) {
		originalErr := &llm.LLMError{
			Provider:      "openai",
			Message:       "Authentication failed",
			ErrorCategory: llm.CategoryAuth,
		}
		wrappedErr := fmt.Errorf("operation failed: %w", originalErr)

		exitCode := getExitCodeFromError(wrappedErr)
		assert.Equal(t, ExitCodeAuthError, exitCode, "Wrapped LLMError should maintain auth exit code")

		userMessage := generateErrorMessage(wrappedErr)
		assert.Equal(t, "Authentication failed. Please check your API key and permissions.", userMessage)
	})

	t.Run("double wrapped LLMError maintains category", func(t *testing.T) {
		originalErr := &llm.LLMError{
			Provider:      "gemini",
			Message:       "Rate limit exceeded",
			ErrorCategory: llm.CategoryRateLimit,
		}
		firstWrap := fmt.Errorf("API call failed: %w", originalErr)
		secondWrap := fmt.Errorf("request processing failed: %w", firstWrap)

		exitCode := getExitCodeFromError(secondWrap)
		assert.Equal(t, ExitCodeRateLimitError, exitCode, "Double wrapped LLMError should maintain rate limit exit code")
	})

	t.Run("wrapped context errors maintain cancellation category", func(t *testing.T) {
		wrappedCanceled := fmt.Errorf("operation cancelled: %w", context.Canceled)
		wrappedDeadline := fmt.Errorf("operation timed out: %w", context.DeadlineExceeded)

		assert.Equal(t, ExitCodeCancelled, getExitCodeFromError(wrappedCanceled))
		assert.Equal(t, ExitCodeCancelled, getExitCodeFromError(wrappedDeadline))
	})

	t.Run("wrapped partial success error maintains generic category", func(t *testing.T) {
		wrappedPartial := fmt.Errorf("execution completed with issues: %w", thinktank.ErrPartialSuccess)

		exitCode := getExitCodeFromError(wrappedPartial)
		assert.Equal(t, ExitCodeGenericError, exitCode)

		userMessage := generateErrorMessage(wrappedPartial)
		assert.Equal(t, "Some model executions failed, but partial results were generated. Use --partial-success-ok flag to exit with success code in this case.", userMessage)
	})

	t.Run("LLMError with original error shows both messages", func(t *testing.T) {
		originalErr := errors.New("network timeout")
		llmErr := &llm.LLMError{
			Provider:      "test",
			Message:       "API call failed",
			ErrorCategory: llm.CategoryNetwork,
			Original:      originalErr,
		}

		userMessage := generateErrorMessage(llmErr)
		// The LLMError.UserFacingError() method should include both the message and original error
		assert.Contains(t, userMessage, "API call failed")
		assert.Contains(t, userMessage, "network timeout")
	})
}

// TestErrorCodeMapping tests the comprehensive deterministic error code mapping
func TestErrorCodeMapping(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		description  string
	}{
		// LLM Error Categories - Complete Coverage
		{
			name: "CategoryAuth",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Invalid API key",
				ErrorCategory: llm.CategoryAuth,
			},
			expectedCode: ExitCodeAuthError,
			description:  "Authentication errors should map to auth exit code",
		},
		{
			name: "CategoryRateLimit",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Too many requests",
				ErrorCategory: llm.CategoryRateLimit,
			},
			expectedCode: ExitCodeRateLimitError,
			description:  "Rate limit errors should map to rate limit exit code",
		},
		{
			name: "CategoryInvalidRequest",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Malformed request",
				ErrorCategory: llm.CategoryInvalidRequest,
			},
			expectedCode: ExitCodeInvalidRequest,
			description:  "Invalid request errors should map to invalid request exit code",
		},
		{
			name: "CategoryServer",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Internal server error",
				ErrorCategory: llm.CategoryServer,
			},
			expectedCode: ExitCodeServerError,
			description:  "Server errors should map to server exit code",
		},
		{
			name: "CategoryNetwork",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Connection failed",
				ErrorCategory: llm.CategoryNetwork,
			},
			expectedCode: ExitCodeNetworkError,
			description:  "Network errors should map to network exit code",
		},
		{
			name: "CategoryInputLimit",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Input too long",
				ErrorCategory: llm.CategoryInputLimit,
			},
			expectedCode: ExitCodeInputError,
			description:  "Input limit errors should map to input exit code",
		},
		{
			name: "CategoryContentFiltered",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Content blocked",
				ErrorCategory: llm.CategoryContentFiltered,
			},
			expectedCode: ExitCodeContentFiltered,
			description:  "Content filtered errors should map to content filtered exit code",
		},
		{
			name: "CategoryInsufficientCredits",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "No credits remaining",
				ErrorCategory: llm.CategoryInsufficientCredits,
			},
			expectedCode: ExitCodeInsufficientCredits,
			description:  "Insufficient credits errors should map to insufficient credits exit code",
		},
		{
			name: "CategoryCancelled",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Request cancelled",
				ErrorCategory: llm.CategoryCancelled,
			},
			expectedCode: ExitCodeCancelled,
			description:  "Cancelled errors should map to cancelled exit code",
		},
		{
			name: "CategoryNotFound",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Resource not found",
				ErrorCategory: llm.CategoryNotFound,
			},
			expectedCode: ExitCodeGenericError,
			description:  "Not found errors should map to generic exit code (no specific exit code)",
		},
		{
			name: "CategoryUnknown",
			err: &llm.LLMError{
				Provider:      "test",
				Message:       "Unknown error",
				ErrorCategory: llm.CategoryUnknown,
			},
			expectedCode: ExitCodeGenericError,
			description:  "Unknown category errors should map to generic exit code",
		},

		// Context Errors
		{
			name:         "context.Canceled",
			err:          context.Canceled,
			expectedCode: ExitCodeCancelled,
			description:  "Context canceled should map to cancelled exit code",
		},
		{
			name:         "context.DeadlineExceeded",
			err:          context.DeadlineExceeded,
			expectedCode: ExitCodeCancelled,
			description:  "Context deadline exceeded should map to cancelled exit code",
		},

		// Standard Library Errors
		{
			name:         "thinktank.ErrPartialSuccess",
			err:          thinktank.ErrPartialSuccess,
			expectedCode: ExitCodeGenericError,
			description:  "Partial success should map to generic exit code",
		},

		// Pattern-based Error Detection
		{
			name:         "Message pattern: context canceled",
			err:          errors.New("operation failed: context canceled"),
			expectedCode: ExitCodeCancelled,
			description:  "Messages containing 'context canceled' should map to cancelled exit code",
		},
		{
			name:         "Message pattern: context cancelled (UK spelling)",
			err:          errors.New("operation failed: context cancelled"),
			expectedCode: ExitCodeCancelled,
			description:  "Messages containing 'context cancelled' should map to cancelled exit code",
		},
		{
			name:         "Message pattern: deadline exceeded",
			err:          errors.New("operation failed: deadline exceeded"),
			expectedCode: ExitCodeCancelled,
			description:  "Messages containing 'deadline exceeded' should map to cancelled exit code",
		},

		// Generic Errors
		{
			name:         "Filesystem error",
			err:          errors.New("permission denied: cannot access file"),
			expectedCode: ExitCodeGenericError,
			description:  "Filesystem errors should map to generic exit code",
		},
		{
			name:         "Network error",
			err:          errors.New("dial tcp: connection refused"),
			expectedCode: ExitCodeGenericError,
			description:  "Low-level network errors should map to generic exit code",
		},
		{
			name:         "Configuration error",
			err:          errors.New("invalid configuration: missing required field"),
			expectedCode: ExitCodeGenericError,
			description:  "Configuration errors should map to generic exit code",
		},
		{
			name:         "Unknown error",
			err:          errors.New("unexpected error occurred"),
			expectedCode: ExitCodeGenericError,
			description:  "Unknown errors should map to generic exit code",
		},

		// Edge Cases
		{
			name:         "nil error",
			err:          nil,
			expectedCode: ExitCodeSuccess,
			description:  "nil error should return success exit code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExitCodeFromError(tt.err)
			assert.Equal(t, tt.expectedCode, result, tt.description)
		})
	}
}

// Simple test mock for audit logger
type TestMockAuditLogger struct{}

func (m *TestMockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error { return nil }
func (m *TestMockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error                { return nil }
func (m *TestMockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs, outputs map[string]interface{}, err error) error {
	return nil
}
func (m *TestMockAuditLogger) LogOpLegacy(operation, status string, inputs, outputs map[string]interface{}, err error) error {
	return nil
}
func (m *TestMockAuditLogger) Close() error { return nil }
