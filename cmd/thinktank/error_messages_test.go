package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// handleErrorMessageTest is a test version of handleError for subprocess testing
func handleErrorMessageTest(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string) {
	// Simple implementation that just calls os.Exit with appropriate code
	if err == nil {
		os.Exit(0)
		return
	}

	// Basic error categorization for testing
	if llmErr, ok := err.(*llm.LLMError); ok {
		switch llmErr.ErrorCategory {
		case llm.CategoryAuth:
			os.Exit(2) // cli.ExitCodeAuthError
		case llm.CategoryRateLimit:
			os.Exit(3) // cli.ExitCodeRateLimitError
		case llm.CategoryInvalidRequest:
			os.Exit(4) // cli.ExitCodeInvalidRequest
		default:
			os.Exit(1) // cli.ExitCodeGenericError
		}
	}
	os.Exit(1) // cli.ExitCodeGenericError
}

// TestHandleErrorMessages checks that handleError generates appropriate user-facing messages
func TestHandleErrorMessages(t *testing.T) {
	// We need a way to capture stderr output to test the messages
	// This requires running the test in a subprocess pattern
	if os.Getenv("TEST_ERROR_MESSAGE") == "1" {
		// In subprocess mode
		errMsg := os.Getenv("ERROR_MESSAGE")
		errType := os.Getenv("ERROR_TYPE")

		// Create error based on type
		var err error
		switch errType {
		case "auth":
			err = llm.New("test", "AUTH_ERR", 401, "Authentication failed", "req123", errors.New(errMsg), llm.CategoryAuth)
		case "rate_limit":
			err = llm.New("test", "RATE_LIMIT", 429, "Rate limit exceeded", "req456", errors.New(errMsg), llm.CategoryRateLimit)
		case "generic":
			err = errors.New(errMsg)
		case "partial_success":
			err = thinktank.ErrPartialSuccess
		default:
			err = errors.New(errMsg)
		}

		// Setup minimal context
		ctx := context.Background()
		logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
		auditLogger := auditlog.NewNoOpAuditLogger()

		// Keep the original os.Exit to actually exit the subprocess
		// No need to save the original since this is a subprocess that will exit

		// Handle error - this will print to stderr and exit
		// Note: handleError was moved to internal/cli package
		// Note: handleError function is not exported from cli package
		// This test should be converted to direct function testing
		// For now, create a local handleError that just exits with the error code
		handleErrorMessageTest(ctx, err, logger, auditLogger, "test_operation")
		return
	}

	// Main test process - run tests using subprocesses
	tests := []struct {
		name           string
		errorType      string
		errorMessage   string
		expectedOutput string
	}{
		{
			name:           "Auth error message",
			errorType:      "auth",
			errorMessage:   "invalid API key",
			expectedOutput: "Error: Authentication failed: invalid API key",
		},
		{
			name:           "Rate limit error message",
			errorType:      "rate_limit",
			errorMessage:   "too many requests",
			expectedOutput: "Error: Rate limit exceeded: too many requests",
		},
		{
			name:           "Generic error message",
			errorType:      "generic",
			errorMessage:   "something unexpected happened",
			expectedOutput: "Error: something unexpected happened",
		},
		{
			name:           "Partial success error message",
			errorType:      "partial_success",
			errorMessage:   "",
			expectedOutput: "Error: Some model executions failed, but partial results were generated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subprocess to test the error message
			cmd := exec.Command(os.Args[0], "-test.run=TestHandleErrorMessages")

			// Set up environment for the subprocess
			cmd.Env = append(os.Environ(),
				"TEST_ERROR_MESSAGE=1",
				"ERROR_TYPE="+tt.errorType,
				"ERROR_MESSAGE="+tt.errorMessage,
			)

			// Capture stderr
			var stderr strings.Builder
			cmd.Stderr = &stderr

			// Run the subprocess
			err := cmd.Run()

			// The subprocess should exit with a non-zero code, so err should not be nil
			if err == nil {
				t.Errorf("Expected subprocess to exit with non-zero code")
			}

			// Check that the stderr output contains the expected message
			if !strings.Contains(stderr.String(), tt.expectedOutput) {
				t.Errorf("Expected stderr to contain %q, got %q", tt.expectedOutput, stderr.String())
			}
		})
	}
}

// TestUserFacingError tests the UserFacingError method of LLMError
func TestUserFacingError(t *testing.T) {
	tests := []struct {
		name          string
		llmErr        *llm.LLMError
		shouldContain []string
	}{
		{
			name: "Error with original error",
			llmErr: &llm.LLMError{
				Message:    "Authentication failed",
				Original:   errors.New("API key expired"),
				Suggestion: "Renew your API key",
			},
			shouldContain: []string{
				"Authentication failed: API key expired",
				"Suggestion: Renew your API key",
			},
		},
		{
			name: "Error without original error",
			llmErr: &llm.LLMError{
				Message:    "Rate limit exceeded",
				Suggestion: "Try again later",
			},
			shouldContain: []string{
				"Rate limit exceeded",
				"Suggestion: Try again later",
			},
		},
		{
			name: "Error without suggestion",
			llmErr: &llm.LLMError{
				Message:  "Network error",
				Original: errors.New("connection timeout"),
			},
			shouldContain: []string{
				"Network error: connection timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.llmErr.UserFacingError()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected UserFacingError() to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

// TestDebugInfo tests that LLMError.DebugInfo returns all expected details
func TestDebugInfo(t *testing.T) {
	originalErr := errors.New("original error")
	llmErr := &llm.LLMError{
		Provider:      "test-provider",
		Code:          "TEST_ERR_CODE",
		StatusCode:    429,
		Message:       "Rate limit exceeded",
		RequestID:     "req-12345",
		Original:      originalErr,
		ErrorCategory: llm.CategoryRateLimit,
		Suggestion:    "Try again later",
		Details:       "Additional details about the error",
	}

	debugInfo := llmErr.DebugInfo()

	// Check that all fields are included in the debug info
	expectedFields := []string{
		"Provider: test-provider",
		"Error Category: RateLimit",
		"Message: Rate limit exceeded",
		"Error Code: TEST_ERR_CODE",
		"Status Code: 429",
		"Request ID: req-12345",
		"Original Error: original error",
		"Details: Additional details about the error",
		"Suggestion: Try again later",
	}

	for _, field := range expectedFields {
		if !strings.Contains(debugInfo, field) {
			t.Errorf("Expected debug info to contain %q, but it didn't.\nDebug info: %s", field, debugInfo)
		}
	}
}
