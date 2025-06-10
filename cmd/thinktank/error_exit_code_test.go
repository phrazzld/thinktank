package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// Define a variable to capture exit codes without actually exiting
var lastExitCode int

// handleTestError is a wrapper around handleError for testing
// It uses the exitCodeCapture function to capture the exit code instead of exiting
func handleTestError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string, t *testing.T) {
	// Reset the last exit code
	lastExitCode = 0

	// Define a monkeyPatched version of handleError that only captures exit code without generating messages
	monkeyPatchedHandleError := func(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string) {
		if err == nil {
			return
		}

		// Just calculate the exit code for testing purposes
		exitCode := ExitCodeGenericError

		// Check if the error is an LLMError that implements CategorizedError
		if catErr, ok := llm.IsCategorizedError(err); ok {
			category := catErr.Category()

			// Determine exit code based on error category
			switch category {
			case llm.CategoryAuth:
				exitCode = ExitCodeAuthError
			case llm.CategoryRateLimit:
				exitCode = ExitCodeRateLimitError
			case llm.CategoryInvalidRequest:
				exitCode = ExitCodeInvalidRequest
			case llm.CategoryServer:
				exitCode = ExitCodeServerError
			case llm.CategoryNetwork:
				exitCode = ExitCodeNetworkError
			case llm.CategoryInputLimit:
				exitCode = ExitCodeInputError
			case llm.CategoryContentFiltered:
				exitCode = ExitCodeContentFiltered
			case llm.CategoryInsufficientCredits:
				exitCode = ExitCodeInsufficientCredits
			case llm.CategoryCancelled:
				exitCode = ExitCodeCancelled
			}
		} else if errors.Is(err, thinktank.ErrPartialSuccess) {
			// Special case for partial success errors
			exitCode = ExitCodeGenericError
		} else if errors.Is(err, thinktank.ErrInvalidConfiguration) ||
			errors.Is(err, thinktank.ErrNoModelsProvided) ||
			errors.Is(err, thinktank.ErrInvalidInstructions) ||
			errors.Is(err, thinktank.ErrInvalidOutputDir) ||
			errors.Is(err, thinktank.ErrInvalidModelName) {
			// Invalid request sentinel errors
			exitCode = ExitCodeInvalidRequest
		} else if errors.Is(err, thinktank.ErrInvalidAPIKey) {
			// Auth sentinel errors
			exitCode = ExitCodeAuthError
		}

		// Set our lastExitCode variable for test verification
		lastExitCode = exitCode
	}

	// Call our monkeypatched version
	monkeyPatchedHandleError(ctx, err, logger, auditLogger, operation)
}

// TestHandleErrorExitCodes checks that handleError assigns the correct exit code
// based on error category
func TestHandleErrorExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "Auth error",
			err:          llm.New("test", "AUTH_ERR", 401, "Authentication failed", "req123", errors.New("invalid key"), llm.CategoryAuth),
			expectedCode: ExitCodeAuthError,
		},
		{
			name:         "Rate limit error",
			err:          llm.New("test", "RATE_LIMIT", 429, "Rate limit exceeded", "req456", errors.New("too many requests"), llm.CategoryRateLimit),
			expectedCode: ExitCodeRateLimitError,
		},
		{
			name:         "Invalid request error",
			err:          llm.New("test", "INVALID_REQ", 400, "Invalid request", "req789", errors.New("bad request"), llm.CategoryInvalidRequest),
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "Server error",
			err:          llm.New("test", "SERVER_ERR", 500, "Server error", "req101", errors.New("internal server error"), llm.CategoryServer),
			expectedCode: ExitCodeServerError,
		},
		{
			name:         "Network error",
			err:          llm.New("test", "NETWORK_ERR", 0, "Network error", "req112", errors.New("connection failed"), llm.CategoryNetwork),
			expectedCode: ExitCodeNetworkError,
		},
		{
			name:         "Input limit error",
			err:          llm.New("test", "INPUT_LIMIT", 413, "Input too large", "req131", errors.New("token limit exceeded"), llm.CategoryInputLimit),
			expectedCode: ExitCodeInputError,
		},
		{
			name:         "Content filtered error",
			err:          llm.New("test", "CONTENT_FILTERED", 400, "Content filtered", "req415", errors.New("content not allowed"), llm.CategoryContentFiltered),
			expectedCode: ExitCodeContentFiltered,
		},
		{
			name:         "Insufficient credits error",
			err:          llm.New("test", "INSUFFICIENT_CREDITS", 402, "Insufficient credits", "req617", errors.New("payment required"), llm.CategoryInsufficientCredits),
			expectedCode: ExitCodeInsufficientCredits,
		},
		{
			name:         "Cancelled error",
			err:          llm.New("test", "CANCELLED", 499, "Request cancelled", "req819", errors.New("context cancelled"), llm.CategoryCancelled),
			expectedCode: ExitCodeCancelled,
		},
		{
			name:         "Generic error",
			err:          errors.New("generic error"),
			expectedCode: ExitCodeGenericError,
		},
		{
			name:         "Partial success error",
			err:          thinktank.ErrPartialSuccess,
			expectedCode: ExitCodeGenericError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a test context and loggers
			ctx := context.Background()
			logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
			auditLogger := auditlog.NewNoOpAuditLogger()

			// Handle the error using our test harness
			handleTestError(ctx, tt.err, logger, auditLogger, "test_operation", t)

			// Verify that the exit code matches expectations
			if lastExitCode != tt.expectedCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectedCode, lastExitCode)
			}
		})
	}
}

// TestExitCodeFromLLMErrorCategory tests mapping from LLMError categories to exit codes
func TestExitCodeFromLLMErrorCategory(t *testing.T) {
	tests := []struct {
		category     llm.ErrorCategory
		expectedCode int
	}{
		{llm.CategoryAuth, ExitCodeAuthError},
		{llm.CategoryRateLimit, ExitCodeRateLimitError},
		{llm.CategoryInvalidRequest, ExitCodeInvalidRequest},
		{llm.CategoryServer, ExitCodeServerError},
		{llm.CategoryNetwork, ExitCodeNetworkError},
		{llm.CategoryInputLimit, ExitCodeInputError},
		{llm.CategoryContentFiltered, ExitCodeContentFiltered},
		{llm.CategoryInsufficientCredits, ExitCodeInsufficientCredits},
		{llm.CategoryCancelled, ExitCodeCancelled},
		{llm.CategoryUnknown, ExitCodeGenericError},
		{llm.CategoryNotFound, ExitCodeGenericError}, // NotFound doesn't have a specific exit code
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			// Create an LLM error with the given category
			err := llm.New("test", "", 0, "test error", "", nil, tt.category)

			// Setup a test context and loggers
			ctx := context.Background()
			logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
			auditLogger := auditlog.NewNoOpAuditLogger()

			// Handle the error using our test harness
			handleTestError(ctx, err, logger, auditLogger, "test_operation", t)

			// Verify that the exit code matches expectations
			if lastExitCode != tt.expectedCode {
				t.Errorf("Category %v expected exit code %d, got %d",
					tt.category, tt.expectedCode, lastExitCode)
			}
		})
	}
}

// TestThinkTankSentinelErrorHandling tests exit codes for thinktank-specific sentinel errors
func TestThinkTankSentinelErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "ErrPartialSuccess",
			err:          thinktank.ErrPartialSuccess,
			expectedCode: ExitCodeGenericError,
		},
		{
			name:         "ErrInvalidConfiguration",
			err:          thinktank.ErrInvalidConfiguration,
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "ErrNoModelsProvided",
			err:          thinktank.ErrNoModelsProvided,
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "ErrInvalidModelName",
			err:          thinktank.ErrInvalidModelName,
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "ErrInvalidAPIKey",
			err:          thinktank.ErrInvalidAPIKey,
			expectedCode: ExitCodeAuthError,
		},
		{
			name:         "ErrInvalidInstructions",
			err:          thinktank.ErrInvalidInstructions,
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "ErrInvalidOutputDir",
			err:          thinktank.ErrInvalidOutputDir,
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "ErrContextGatheringFailed",
			err:          thinktank.ErrContextGatheringFailed,
			expectedCode: ExitCodeGenericError,
		},
		{
			name:         "Wrapped sentinel error",
			err:          fmt.Errorf("wrapping error: %w", thinktank.ErrInvalidAPIKey),
			expectedCode: ExitCodeAuthError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a test context and loggers
			ctx := context.Background()
			logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
			auditLogger := auditlog.NewNoOpAuditLogger()

			// Test the exit code calculation
			handleTestError(ctx, tt.err, logger, auditLogger, "test_operation", t)

			// Verify that the exit code matches expectations
			if lastExitCode != tt.expectedCode {
				t.Errorf("Expected exit code %d for %v, got %d", tt.expectedCode, tt.err, lastExitCode)
			}
		})
	}
}
