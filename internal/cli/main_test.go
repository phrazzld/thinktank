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
			cmd := exec.Command(os.Args[0], "-test.run=TestHandleError")
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

	cmd := exec.Command(os.Args[0], "-test.run=TestHandleErrorAuditLogFailure")
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
		cmd := exec.Command(os.Args[0], "-test.run=TestMainFunction")
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
		cmd := exec.Command(os.Args[0], "-test.run=TestMainFunction")
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

// TestMainDryRun tests the Main function with dry-run mode to avoid API calls
func TestMainDryRun(t *testing.T) {
	if os.Getenv("TEST_MAIN_DRY_RUN") != "" {
		// This is the subprocess that will run Main with dry-run
		// Override os.Args to simulate command line arguments
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set up args from environment variables
		instructionsFile := os.Getenv("TEST_INSTRUCTIONS_FILE")
		outputDir := os.Getenv("TEST_OUTPUT_DIR")
		testFile := os.Getenv("TEST_FILE")
		testMode := os.Getenv("TEST_MODE")

		os.Args = []string{"thinktank"}
		if instructionsFile != "" {
			os.Args = append(os.Args, "--instructions", instructionsFile)
		}
		if testMode == "dry-run" {
			os.Args = append(os.Args, "--dry-run")
		}
		if testMode == "audit" {
			auditFile := os.Getenv("TEST_AUDIT_FILE")
			os.Args = append(os.Args, "--dry-run", "--audit-log-file", auditFile)
		}
		if testMode == "verbose" {
			os.Args = append(os.Args, "--dry-run", "--verbose")
		}
		if testMode == "quiet" {
			os.Args = append(os.Args, "--dry-run", "--quiet")
		}
		if outputDir != "" {
			os.Args = append(os.Args, "--output-dir", outputDir)
		}
		if testFile != "" {
			os.Args = append(os.Args, testFile)
		}

		Main()
		return
	}

	// Create a temporary instructions file
	tmpFile, err := os.CreateTemp("", "test_instructions_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("Test instructions")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_main_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test file in the directory
	testFile := tmpDir + "/test.go"
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("main dry run success", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainDryRun")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_DRY_RUN=1",
			"TEST_MODE=dry-run",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code for dry run
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success for dry run, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error running dry run: %v", err)
			}
		}
	})

	t.Run("main with audit logging", func(t *testing.T) {
		auditFile := tmpDir + "/audit.log"

		cmd := exec.Command(os.Args[0], "-test.run=TestMainDryRun")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_DRY_RUN=1",
			"TEST_MODE=audit",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_AUDIT_FILE="+auditFile,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with audit logging, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with audit logging: %v", err)
			}
		}

		// Verify audit log was created
		if _, err := os.Stat(auditFile); os.IsNotExist(err) {
			t.Error("Expected audit log file to be created")
		}
	})

	t.Run("main with verbose logging", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainDryRun")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_DRY_RUN=1",
			"TEST_MODE=verbose",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with verbose logging, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with verbose logging: %v", err)
			}
		}
	})

	t.Run("main with quiet mode", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainDryRun")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_DRY_RUN=1",
			"TEST_MODE=quiet",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with quiet mode, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with quiet mode: %v", err)
			}
		}
	})
}

// TestMainValidationErrors tests Main function with various validation errors
func TestMainValidationErrors(t *testing.T) {
	if os.Getenv("TEST_MAIN_VALIDATION") != "" {
		// This is the subprocess that will run Main
		Main()
		return
	}

	// Create a temporary instructions file
	tmpFile, err := os.CreateTemp("", "test_instructions_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("Test instructions")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_main_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := tmpDir + "/test.go"
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("main with missing instructions file", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainValidationErrors")
		cmd.Env = append(os.Environ(), "TEST_MAIN_VALIDATION=1")
		// Missing --instructions flag
		cmd.Args = append(cmd.Args, testFile)

		err := cmd.Run()

		// Should exit with error code for validation failure
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.NotEqual(t, 0, exitError.ExitCode(),
				"Expected non-zero exit code for missing instructions")
		} else {
			t.Fatal("Expected command to exit with error code for missing instructions")
		}
	})

	t.Run("main with conflicting flags", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainValidationErrors")
		cmd.Env = append(os.Environ(), "TEST_MAIN_VALIDATION=1")
		// Conflicting --quiet and --verbose flags
		cmd.Args = append(cmd.Args,
			"--instructions", tmpFile.Name(),
			"--quiet",
			"--verbose",
			testFile)

		err := cmd.Run()

		// Should exit with error code for validation failure
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.NotEqual(t, 0, exitError.ExitCode(),
				"Expected non-zero exit code for conflicting flags")
		} else {
			t.Fatal("Expected command to exit with error code for conflicting flags")
		}
	})

	t.Run("main with invalid synthesis model", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainValidationErrors")
		cmd.Env = append(os.Environ(), "TEST_MAIN_VALIDATION=1")
		// Invalid synthesis model pattern
		cmd.Args = append(cmd.Args,
			"--instructions", tmpFile.Name(),
			"--synthesis-model", "invalid-model-pattern",
			"--dry-run",
			testFile)

		err := cmd.Run()

		// Should exit with error code for validation failure
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.NotEqual(t, 0, exitError.ExitCode(),
				"Expected non-zero exit code for invalid synthesis model")
		} else {
			t.Fatal("Expected command to exit with error code for invalid synthesis model")
		}
	})

	t.Run("main with no paths", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainValidationErrors")
		cmd.Env = append(os.Environ(), "TEST_MAIN_VALIDATION=1")
		// No file paths provided
		cmd.Args = append(cmd.Args,
			"--instructions", tmpFile.Name())

		err := cmd.Run()

		// Should exit with error code for missing paths
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.NotEqual(t, 0, exitError.ExitCode(),
				"Expected non-zero exit code for missing paths")
		} else {
			t.Fatal("Expected command to exit with error code for missing paths")
		}
	})
}

// TestMainConfigurationOptions tests various configuration combinations
func TestMainConfigurationOptions(t *testing.T) {
	if os.Getenv("TEST_MAIN_CONFIG") != "" {
		// This is the subprocess that will run Main
		// Override os.Args to simulate command line arguments
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set up args from environment variables
		instructionsFile := os.Getenv("TEST_INSTRUCTIONS_FILE")
		outputDir := os.Getenv("TEST_OUTPUT_DIR")
		testFile := os.Getenv("TEST_FILE")
		configMode := os.Getenv("TEST_CONFIG_MODE")

		os.Args = []string{"thinktank", "--dry-run"}
		if instructionsFile != "" {
			os.Args = append(os.Args, "--instructions", instructionsFile)
		}
		if outputDir != "" {
			os.Args = append(os.Args, "--output-dir", outputDir)
		}

		// Add mode-specific flags
		switch configMode {
		case "timeout":
			os.Args = append(os.Args, "--timeout", "5s")
		case "ratelimit":
			os.Args = append(os.Args, "--rate-limit", "30", "--max-concurrent", "3")
		case "permissions":
			os.Args = append(os.Args, "--dir-permissions", "0755", "--file-permissions", "0644")
		case "multimodel":
			os.Args = append(os.Args, "--model", "gemini-2.5-pro", "--model", "gemini-2.5-flash")
		case "filtering":
			os.Args = append(os.Args, "--include", ".go,.md", "--exclude", ".exe,.bin", "--exclude-names", "node_modules,dist")
		}

		if testFile != "" {
			os.Args = append(os.Args, testFile)
		}

		Main()
		return
	}

	// Create a temporary instructions file
	tmpFile, err := os.CreateTemp("", "test_instructions_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("Test instructions")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_main_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := tmpDir + "/test.go"
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("main with custom timeout", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainConfigurationOptions")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_CONFIG=1",
			"TEST_CONFIG_MODE=timeout",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with custom timeout, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with custom timeout: %v", err)
			}
		}
	})

	t.Run("main with rate limiting", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainConfigurationOptions")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_CONFIG=1",
			"TEST_CONFIG_MODE=ratelimit",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with rate limiting, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with rate limiting: %v", err)
			}
		}
	})

	t.Run("main with custom permissions", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainConfigurationOptions")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_CONFIG=1",
			"TEST_CONFIG_MODE=permissions",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with custom permissions, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with custom permissions: %v", err)
			}
		}
	})

	t.Run("main with multiple models", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainConfigurationOptions")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_CONFIG=1",
			"TEST_CONFIG_MODE=multimodel",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with multiple models, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with multiple models: %v", err)
			}
		}
	})

	t.Run("main with file filtering", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestMainConfigurationOptions")
		cmd.Env = append(os.Environ(),
			"TEST_MAIN_CONFIG=1",
			"TEST_CONFIG_MODE=filtering",
			"TEST_INSTRUCTIONS_FILE="+tmpFile.Name(),
			"TEST_OUTPUT_DIR="+tmpDir,
			"TEST_FILE="+testFile)

		err := cmd.Run()

		// Should exit with success code
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				t.Fatalf("Expected success with file filtering, got exit code %d", exitError.ExitCode())
			} else {
				t.Fatalf("Unexpected error with file filtering: %v", err)
			}
		}
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
