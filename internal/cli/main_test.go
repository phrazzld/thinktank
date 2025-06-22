package cli

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
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
// TestHandleError has been converted to direct function tests in error_handling_test.go
// to eliminate subprocess test complexity and improve reliability

// TestHandleErrorAuditLogFailure has been converted to direct function test in error_handling_test.go
// to eliminate subprocess test complexity and improve reliability

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

// TestMainFunction has been converted to direct function tests in flags_parsing_test.go
// to eliminate subprocess test complexity and improve reliability.
// Flag parsing errors are now tested directly via ParseFlagsWithEnv().
// Input validation errors are tested directly via TestValidationErrors().
func TestMainFunction(t *testing.T) {
	t.Run("Main function components work independently", func(t *testing.T) {
		// Test that we can call the main components without subprocess execution

		// Test flag parsing works (this uses the testable ParseFlagsWithEnv function)
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--instructions", "/path/to/instructions.md",
			"--model", "gemini-2.5-pro",
			"--dry-run", // This avoids needing real API calls or file operations
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, func(string) string { return "" })
		assert.NoError(t, err, "ParseFlagsWithEnv should work with valid args")
		assert.NotNil(t, cfg, "Should return valid config")

		// Test that input validation works
		logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")
		validateErr := ValidateInputs(cfg, logger)
		assert.NoError(t, validateErr, "ValidateInputs should pass with valid config in dry-run mode")

		// Test that SetupLogging works
		setupLogger := SetupLogging(cfg)
		assert.NotNil(t, setupLogger, "SetupLogging should return a valid logger")
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
