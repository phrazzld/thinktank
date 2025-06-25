// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
)

// TestCLIErrorConstants tests that CLI error constants work with existing error infrastructure
// Following TDD - RED phase: this test will fail until we implement the error constants
func TestCLIErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{
			name:     "missing instructions error",
			err:      ErrMissingInstructions,
			wantCode: ExitCodeInvalidRequest,
			wantMsg:  "instructions file required",
		},
		{
			name:     "invalid model error",
			err:      ErrInvalidModel,
			wantCode: ExitCodeInvalidRequest,
			wantMsg:  "invalid model",
		},
		{
			name:     "no API key error",
			err:      ErrNoAPIKey,
			wantCode: ExitCodeAuthError,
			wantMsg:  "API key not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that errors work with existing getExitCodeFromError function
			// CLI errors need to be wrapped as LLM errors for exit code determination
			if cliErr, ok := IsCLIError(tt.err); ok {
				llmErr := WrapAsLLMError(cliErr)
				exitCode := getExitCodeFromError(llmErr)
				assert.Equal(t, tt.wantCode, exitCode)
			} else {
				exitCode := getExitCodeFromError(tt.err)
				assert.Equal(t, tt.wantCode, exitCode)
			}

			// Test that error messages contain expected content
			assert.Contains(t, tt.err.Error(), tt.wantMsg)

			// Test that errors are not nil
			assert.NotNil(t, tt.err)
		})
	}
}

// TestCLIErrorWrapping tests error wrapping with context preservation
func TestCLIErrorWrapping(t *testing.T) {
	originalErr := errors.New("original error")

	// Test wrapping with context
	wrappedErr := WrapCLIError(originalErr, CLIErrorInvalidValue, "parsing failed", "check your syntax", "test context")

	// Should preserve original error
	assert.ErrorIs(t, wrappedErr, originalErr)

	// Should work with existing exit code determination
	exitCode := getExitCodeFromError(WrapAsLLMError(wrappedErr))
	assert.Equal(t, ExitCodeInvalidRequest, exitCode)

	// Should contain helpful message
	assert.Contains(t, wrappedErr.Error(), "parsing failed")
}

// TestCLIErrorWithCorrelationID tests correlation ID integration
func TestCLIErrorWithCorrelationID(t *testing.T) {
	correlationID := "test-correlation-123"

	err := NewCLIErrorWithCorrelation("test error", "suggestion", correlationID)

	// Should work with existing error infrastructure
	exitCode := getExitCodeFromError(err)
	assert.Equal(t, ExitCodeInvalidRequest, exitCode)

	// Should preserve correlation ID (if LLM package supports extraction)
	extractedID := llm.ExtractCorrelationID(err)
	assert.Equal(t, correlationID, extractedID)
}

// TestCLIErrorIntegrationWithExistingFunctions tests integration with existing error handling
func TestCLIErrorIntegrationWithExistingFunctions(t *testing.T) {
	// Test that CLI errors work with getExitCodeFromError function
	cliErr := ErrMissingInstructions

	// CLI errors need to be wrapped as LLM errors for proper exit code determination
	if typedCliErr, ok := IsCLIError(cliErr); ok {
		llmErr := WrapAsLLMError(typedCliErr)
		exitCode := getExitCodeFromError(llmErr)
		assert.Equal(t, ExitCodeInvalidRequest, exitCode)

		// Test message generation
		message := generateErrorMessage(cliErr)
		assert.NotEmpty(t, message)
		assert.Contains(t, message, "instructions file required")
	}
}
