package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/stretchr/testify/assert"
)

// mockAuditLogger implements auditlog.AuditLogger for testing
type mockAuditLogger struct {
	logOpCalled bool
	logOpError  error
}

func (m *mockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	return m.logOpError
}

func (m *mockAuditLogger) LogOp(ctx context.Context, operation, result string, request, response map[string]interface{}, err error) error {
	m.logOpCalled = true
	return m.logOpError
}

func (m *mockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	return m.logOpError
}

func (m *mockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return m.logOpError
}

func (m *mockAuditLogger) Close() error {
	return nil
}

// TestHandleError tests the handleError function with various error types
func TestHandleError(t *testing.T) {
	if os.Getenv("TEST_HANDLE_ERROR") != "" {
		// This is the subprocess that will call handleError and os.Exit
		ctx := context.Background()
		logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")
		auditLogger := &mockAuditLogger{}

		// Parse the error type from environment variable
		errorType := os.Getenv("TEST_ERROR_TYPE")

		var err error
		switch errorType {
		case "nil":
			err = nil
		case "auth":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Authentication failed",
				ErrorCategory: llm.CategoryAuth,
			}
		case "ratelimit":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Rate limit exceeded",
				ErrorCategory: llm.CategoryRateLimit,
			}
		case "invalid":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Invalid request",
				ErrorCategory: llm.CategoryInvalidRequest,
			}
		case "server":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Server error",
				ErrorCategory: llm.CategoryServer,
			}
		case "network":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Network error",
				ErrorCategory: llm.CategoryNetwork,
			}
		case "inputlimit":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Input limit exceeded",
				ErrorCategory: llm.CategoryInputLimit,
			}
		case "contentfiltered":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Content filtered",
				ErrorCategory: llm.CategoryContentFiltered,
			}
		case "insufficientcredits":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Insufficient credits",
				ErrorCategory: llm.CategoryInsufficientCredits,
			}
		case "cancelled":
			err = &llm.LLMError{
				Provider:      "test",
				Message:       "Request cancelled",
				ErrorCategory: llm.CategoryCancelled,
			}
		case "partialsuccess":
			err = thinktank.ErrPartialSuccess
		case "generic":
			err = errors.New("generic error")
		default:
			err = errors.New("unknown error")
		}

		handleError(ctx, err, logger, auditLogger, "test_operation")
		return
	}

	tests := []struct {
		name         string
		errorType    string
		expectedExit int
	}{
		{
			name:         "nil error should not exit",
			errorType:    "nil",
			expectedExit: 0, // Special case - nil error returns early
		},
		{
			name:         "auth error",
			errorType:    "auth",
			expectedExit: ExitCodeAuthError,
		},
		{
			name:         "rate limit error",
			errorType:    "ratelimit",
			expectedExit: ExitCodeRateLimitError,
		},
		{
			name:         "invalid request error",
			errorType:    "invalid",
			expectedExit: ExitCodeInvalidRequest,
		},
		{
			name:         "server error",
			errorType:    "server",
			expectedExit: ExitCodeServerError,
		},
		{
			name:         "network error",
			errorType:    "network",
			expectedExit: ExitCodeNetworkError,
		},
		{
			name:         "input limit error",
			errorType:    "inputlimit",
			expectedExit: ExitCodeInputError,
		},
		{
			name:         "content filtered error",
			errorType:    "contentfiltered",
			expectedExit: ExitCodeContentFiltered,
		},
		{
			name:         "insufficient credits error",
			errorType:    "insufficientcredits",
			expectedExit: ExitCodeInsufficientCredits,
		},
		{
			name:         "cancelled error",
			errorType:    "cancelled",
			expectedExit: ExitCodeCancelled,
		},
		{
			name:         "partial success error",
			errorType:    "partialsuccess",
			expectedExit: ExitCodeGenericError,
		},
		{
			name:         "generic error",
			errorType:    "generic",
			expectedExit: ExitCodeGenericError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.errorType == "nil" {
				// Test nil error case directly without subprocess
				ctx := context.Background()
				logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")
				auditLogger := &mockAuditLogger{}

				// This should return early and not call os.Exit
				handleError(ctx, nil, logger, auditLogger, "test_operation")

				// If we get here, the function returned properly (didn't exit)
				assert.False(t, auditLogger.logOpCalled, "LogOp should not be called for nil error")
				return
			}

			// For non-nil errors, use subprocess to test os.Exit behavior
			cmd := exec.Command(os.Args[0], "-test.run", "TestHandleError")
			cmd.Env = append(os.Environ(),
				"TEST_HANDLE_ERROR=1",
				fmt.Sprintf("TEST_ERROR_TYPE=%s", tt.errorType))

			err := cmd.Run()

			// Check the exit code
			if exitError, ok := err.(*exec.ExitError); ok {
				assert.Equal(t, tt.expectedExit, exitError.ExitCode(),
					"Expected exit code %d for %s error, got %d",
					tt.expectedExit, tt.errorType, exitError.ExitCode())
			} else if err == nil {
				// Command exited with code 0
				assert.Equal(t, 0, tt.expectedExit,
					"Expected exit code 0 for %s error, but command succeeded", tt.errorType)
			} else {
				t.Fatalf("Unexpected error running subprocess: %v", err)
			}
		})
	}
}

// TestHandleErrorAuditLogFailure tests handleError when audit logging fails
func TestHandleErrorAuditLogFailure(t *testing.T) {
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")

	// Create a mock audit logger that returns an error
	auditLogger := &mockAuditLogger{
		logOpError: errors.New("audit log error"),
	}

	// We need to test this in a subprocess since handleError calls os.Exit
	if os.Getenv("TEST_AUDIT_FAILURE") != "" {
		err := errors.New("test error")
		handleError(ctx, err, logger, auditLogger, "test_operation")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHandleErrorAuditLogFailure")
	cmd.Env = append(os.Environ(), "TEST_AUDIT_FAILURE=1")

	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		// Should exit with generic error code since we passed a generic error
		assert.Equal(t, ExitCodeGenericError, exitError.ExitCode())
	} else {
		t.Fatal("Expected command to exit with error code")
	}
}

// TestSetupGracefulShutdown tests the setupGracefulShutdown function
func TestSetupGracefulShutdown(t *testing.T) {
	t.Run("context creation", func(t *testing.T) {
		ctx := context.Background()

		signalCtx, cancel := setupGracefulShutdown(ctx)
		defer cancel()

		// Verify we got a valid context and cancel function
		assert.NotNil(t, signalCtx, "Signal context should not be nil")
		assert.NotNil(t, cancel, "Cancel function should not be nil")

		// Verify the context is not cancelled initially
		select {
		case <-signalCtx.Done():
			t.Fatal("Signal context should not be cancelled initially")
		default:
			// Expected - context is not cancelled
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx := context.Background()

		signalCtx, cancel := setupGracefulShutdown(ctx)

		// Cancel the signal context
		cancel()

		// Wait a brief moment for the goroutine to process
		time.Sleep(10 * time.Millisecond)

		// Verify the context is cancelled
		select {
		case <-signalCtx.Done():
			// Expected - context should be cancelled
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Signal context should be cancelled after calling cancel")
		}
	})

	t.Run("parent context cancellation", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())

		signalCtx, signalCancel := setupGracefulShutdown(parentCtx)
		defer signalCancel()

		// Cancel the parent context
		parentCancel()

		// Wait a brief moment for the goroutine to process
		time.Sleep(10 * time.Millisecond)

		// Verify the signal context is also cancelled
		select {
		case <-signalCtx.Done():
			// Expected - signal context should be cancelled when parent is cancelled
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Signal context should be cancelled when parent context is cancelled")
		}
	})
}

// TestMainFunction tests the Main function
func TestMainFunction(t *testing.T) {
	// We need to test Main in a subprocess since it calls os.Exit on errors
	if os.Getenv("TEST_MAIN_FUNCTION") != "" {
		// This is the subprocess - call Main()
		Main()
		return
	}

	t.Run("main with invalid flags", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run", "TestMainFunction")
		cmd.Env = append(os.Environ(), "TEST_MAIN_FUNCTION=1")
		// Add invalid flags that should cause ParseFlags to fail
		cmd.Args = append(cmd.Args, "--invalid-flag")

		err := cmd.Run()

		// Should exit with error code 2 (standard flag parsing error)
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 2, exitError.ExitCode(),
				"Expected exit code 2 for invalid flags")
		} else {
			t.Fatal("Expected command to exit with error code for invalid flags")
		}
	})

	t.Run("main with missing required flags", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run", "TestMainFunction")
		cmd.Env = append(os.Environ(), "TEST_MAIN_FUNCTION=1")
		// Don't provide required flags like --instructions

		err := cmd.Run()

		// Should exit with some error code
		if exitError, ok := err.(*exec.ExitError); ok {
			// The exact exit code depends on which validation fails first
			assert.NotEqual(t, 0, exitError.ExitCode(),
				"Expected non-zero exit code for missing required flags")
		} else {
			t.Fatal("Expected command to exit with error code for missing required flags")
		}
	})
}

// TestValidationErrors tests input validation logic directly
func TestValidationErrors(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")

	t.Run("missing instructions file", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "", // Missing instructions
			Paths:            []string{"/some/path"},
			ModelNames:       []string{"gemini-2.5-pro"},
			DryRun:           false,
		}

		err := ValidateInputs(config, logger)
		assert.Error(t, err, "Should return error for missing instructions file")
		assert.Contains(t, err.Error(), "missing required --instructions flag")
	})

	t.Run("conflicting flags", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "/some/instructions.md",
			Paths:            []string{"/some/path"},
			ModelNames:       []string{"gemini-2.5-pro"},
			Quiet:            true, // Conflicting
			Verbose:          true, // Conflicting
			DryRun:           false,
		}

		err := ValidateInputs(config, logger)
		assert.Error(t, err, "Should return error for conflicting flags")
		assert.Contains(t, err.Error(), "conflicting flags: --quiet and --verbose are mutually exclusive")
	})

	t.Run("invalid synthesis model", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "/some/instructions.md",
			Paths:            []string{"/some/path"},
			ModelNames:       []string{"gemini-2.5-pro"},
			SynthesisModel:   "invalid-model-pattern", // Invalid model pattern
			DryRun:           false,
		}

		err := ValidateInputs(config, logger)
		assert.Error(t, err, "Should return error for invalid synthesis model")
		assert.Contains(t, err.Error(), "invalid synthesis model: 'invalid-model-pattern' does not match any known model pattern")
	})

	t.Run("no paths provided", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "/some/instructions.md",
			Paths:            []string{}, // No paths
			ModelNames:       []string{"gemini-2.5-pro"},
			DryRun:           false,
		}

		err := ValidateInputs(config, logger)
		assert.Error(t, err, "Should return error for missing paths")
		assert.Contains(t, err.Error(), "no paths specified")
	})

	t.Run("no models in non-dry-run mode", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "/some/instructions.md",
			Paths:            []string{"/some/path"},
			ModelNames:       []string{}, // No models
			DryRun:           false,
		}

		err := ValidateInputs(config, logger)
		assert.Error(t, err, "Should return error for missing models in non-dry-run mode")
		assert.Contains(t, err.Error(), "no models specified")
	})

	t.Run("valid configuration", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "/some/instructions.md",
			Paths:            []string{"/some/path"},
			ModelNames:       []string{"gemini-2.5-pro"},
			SynthesisModel:   "gpt-4", // Valid model pattern
			DryRun:           false,
		}

		err := ValidateInputs(config, logger)
		assert.NoError(t, err, "Should not return error for valid configuration")
	})

	t.Run("dry run allows missing instructions", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: "", // Missing instructions
			Paths:            []string{"/some/path"},
			ModelNames:       []string{}, // No models
			DryRun:           true,       // Dry run mode allows these
		}

		err := ValidateInputs(config, logger)
		assert.NoError(t, err, "Should not return error for missing instructions/models in dry run mode")
	})
}

// Note: TestGetFriendlyErrorMessage already exists in flags_new_test.go

// TestExitCodes verifies that all exit codes are properly defined
func TestExitCodes(t *testing.T) {
	// Test that exit codes are properly defined and unique
	exitCodes := map[string]int{
		"ExitCodeSuccess":             ExitCodeSuccess,
		"ExitCodeGenericError":        ExitCodeGenericError,
		"ExitCodeAuthError":           ExitCodeAuthError,
		"ExitCodeRateLimitError":      ExitCodeRateLimitError,
		"ExitCodeInvalidRequest":      ExitCodeInvalidRequest,
		"ExitCodeServerError":         ExitCodeServerError,
		"ExitCodeNetworkError":        ExitCodeNetworkError,
		"ExitCodeInputError":          ExitCodeInputError,
		"ExitCodeContentFiltered":     ExitCodeContentFiltered,
		"ExitCodeInsufficientCredits": ExitCodeInsufficientCredits,
		"ExitCodeCancelled":           ExitCodeCancelled,
	}

	// Verify success code is 0
	assert.Equal(t, 0, ExitCodeSuccess, "Success exit code should be 0")

	// Verify all other codes are non-zero and unique
	usedCodes := make(map[int]string)
	for name, code := range exitCodes {
		if name != "ExitCodeSuccess" {
			assert.NotEqual(t, 0, code, "%s should not be 0", name)
		}

		if existing, exists := usedCodes[code]; exists {
			t.Errorf("Exit code %d is used by both %s and %s", code, existing, name)
		}
		usedCodes[code] = name
	}
}
