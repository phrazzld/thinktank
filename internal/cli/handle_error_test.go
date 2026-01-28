// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/thinktank"
)

// mockExitFunc captures os.Exit calls for testing
var mockExitCode int
var mockExitCalled bool

func mockExit(code int) {
	mockExitCode = code
	mockExitCalled = true
}

func TestHandleError(t *testing.T) {
	// Save original osExit
	originalExit := osExit
	osExit = mockExit
	defer func() {
		osExit = originalExit
	}()

	logger := logutil.NewSlogLoggerFromLogLevel(nil, logutil.InfoLevel)
	ctx := context.Background()

	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "LLM auth error",
			err:          &llm.LLMError{ErrorCategory: llm.CategoryAuth},
			expectedCode: ExitCodeAuthError,
		},
		{
			name:         "CLI error",
			err:          &CLIError{Type: CLIErrorInvalidValue},
			expectedCode: ExitCodeInvalidRequest,
		},
		{
			name:         "context canceled",
			err:          context.Canceled,
			expectedCode: ExitCodeCancelled,
		},
		{
			name:         "partial success error",
			err:          thinktank.ErrPartialSuccess,
			expectedCode: ExitCodeGenericError,
		},
		{
			name:         "generic error",
			err:          errors.New("generic error"),
			expectedCode: ExitCodeGenericError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExitCalled = false
			mockExitCode = 0

			handleError(ctx, tt.err, logger)

			if !mockExitCalled {
				t.Error("Expected os.Exit to be called")
			}

			if mockExitCode != tt.expectedCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectedCode, mockExitCode)
			}
		})
	}
}

// Note: We don't test Main() directly with invalid args since it would cause
// a panic when trying to use a nil SimplifiedConfig. The parsing error handling
// is tested separately in the parser tests.
