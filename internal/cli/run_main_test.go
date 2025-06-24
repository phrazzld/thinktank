package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunMain_BasicContract tests the fundamental contract of RunMain
func TestRunMain_BasicContract(t *testing.T) {
	t.Run("returns MainResult with exit code and error", func(t *testing.T) {
		config := createValidMainConfig(t) // Already includes --dry-run

		result := RunMain(config)

		assert.NotNil(t, result, "RunMain should always return a MainResult")
		assert.IsType(t, &MainResult{}, result, "RunMain should return *MainResult")
		assert.GreaterOrEqual(t, result.ExitCode, 0, "Exit code should be non-negative")
	})
}

// TestRunMain_FlagParsingErrors tests error conditions during flag parsing
func TestRunMain_FlagParsingErrors(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
		errorPattern string
	}{
		{
			name:         "invalid flag",
			args:         []string{"thinktank", "--invalid-flag"},
			expectedCode: ExitCodeInvalidRequest,
			errorPattern: "flag provided but not defined",
		},
		{
			name:         "invalid flag value",
			args:         []string{"thinktank", "--timeout", "invalid"},
			expectedCode: ExitCodeInvalidRequest,
			errorPattern: "parse error",
		},
		{
			name:         "conflicting flags",
			args:         []string{"thinktank", "--quiet", "--verbose"},
			expectedCode: ExitCodeInvalidRequest,
			errorPattern: "conflicting flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createBasicMainConfig(t)
			config.Args = tt.args

			result := RunMain(config)

			assert.Equal(t, tt.expectedCode, result.ExitCode, "Should return correct exit code")
			assert.Error(t, result.Error, "Should return error for flag parsing failure")
			if tt.errorPattern != "" {
				assert.Contains(t, result.Error.Error(), tt.errorPattern, "Error should contain expected pattern")
			}
		})
	}
}

// TestRunMain_ValidationErrors tests input validation failures
func TestRunMain_ValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		configMods   func(*MainConfig)
		expectedCode int
		errorPattern string
	}{
		{
			name: "missing instructions in non-dry-run",
			configMods: func(mc *MainConfig) {
				mc.Args = []string{"thinktank", "--model", "gemini-2.5-pro", "src/"}
			},
			expectedCode: ExitCodeInvalidRequest,
			errorPattern: "missing required --instructions flag",
		},
		{
			name: "missing models in non-dry-run",
			configMods: func(mc *MainConfig) {
				tempDir := t.TempDir()
				instructionsFile := filepath.Join(tempDir, "instructions.md")
				require.NoError(t, os.WriteFile(instructionsFile, []byte("test instructions"), 0644))
				// Add explicit --model flag with empty value - this fails during context gathering
				mc.Args = []string{"thinktank", "--instructions", instructionsFile, "--model", "", "src/"}
			},
			expectedCode: ExitCodeGenericError, // Context gathering fails, not validation
			errorPattern: "context gathering failed",
		},
		{
			name: "missing paths",
			configMods: func(mc *MainConfig) {
				tempDir := t.TempDir()
				instructionsFile := filepath.Join(tempDir, "instructions.md")
				require.NoError(t, os.WriteFile(instructionsFile, []byte("test instructions"), 0644))
				mc.Args = []string{"thinktank", "--instructions", instructionsFile, "--model", "gemini-2.5-pro"}
			},
			expectedCode: ExitCodeInvalidRequest,
			errorPattern: "no paths specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createBasicMainConfig(t)
			tt.configMods(config)

			result := RunMain(config)

			assert.Equal(t, tt.expectedCode, result.ExitCode, "Should return correct exit code")
			assert.Error(t, result.Error, "Should return error for validation failure")
			if tt.errorPattern != "" && result.Error != nil {
				assert.Contains(t, result.Error.Error(), tt.errorPattern, "Error should contain expected pattern")
			}
		})
	}
}

// TestRunMain_BootstrapComponents tests the bootstrap initialization logic
func TestRunMain_BootstrapComponents(t *testing.T) {
	t.Run("creates context with timeout", func(t *testing.T) {
		config := createValidMainConfig(t)
		config.Args = withTimeout(config.Args, "30s")

		result := RunMain(config)

		// Should succeed or fail gracefully, but not panic from context issues
		assert.NotNil(t, result, "Should handle context creation")
		assert.GreaterOrEqual(t, result.ExitCode, 0, "Should have valid exit code")
	})

	t.Run("handles audit log file creation", func(t *testing.T) {
		tempDir := t.TempDir()
		auditFile := filepath.Join(tempDir, "audit.log")

		config := createValidMainConfig(t)
		config.Args = withAuditLogFile(config.Args, auditFile)
		// Remove --dry-run for this test to ensure audit logging actually occurs
		config.Args = removeDryRunFlag(config.Args)
		// Provide a mock API key to avoid immediate authentication failure
		config.Getenv = func(key string) string {
			if key == "GEMINI_API_KEY" {
				return "mock-api-key-for-testing"
			}
			return ""
		}

		result := RunMain(config)

		// Should succeed in creating audit file (even if execution fails due to mock API key)
		assert.FileExists(t, auditFile, "Should create audit log file")
		assert.NotNil(t, result, "Should complete execution")
		// Don't assert on exit code since it might fail due to mock API key, but audit file should exist
	})

	t.Run("handles correlation ID generation", func(t *testing.T) {
		config := createValidMainConfig(t)

		result := RunMain(config)

		// Should complete without correlation ID errors
		assert.NotNil(t, result, "Should handle correlation ID setup")
	})
}

// TestRunMain_DryRunBehavior tests dry-run mode execution
func TestRunMain_DryRunBehavior(t *testing.T) {
	t.Run("dry run succeeds without API keys", func(t *testing.T) {
		config := createValidMainConfig(t)
		config.Args = withDryRun(config.Args)

		result := RunMain(config)

		assert.Equal(t, ExitCodeSuccess, result.ExitCode, "Dry run should succeed without API keys")
		assert.NoError(t, result.Error, "Dry run should not error")
		assert.NotNil(t, result.RunResult, "Should have run result from dry run")
	})

	t.Run("dry run allows missing instructions", func(t *testing.T) {
		config := createBasicMainConfig(t)
		config.Args = []string{"thinktank", "--dry-run", "src/"}

		result := RunMain(config)

		assert.Equal(t, ExitCodeSuccess, result.ExitCode, "Dry run should allow missing instructions")
		assert.NoError(t, result.Error, "Dry run should not error")
	})
}

// TestRunMain_ProductionExecution tests full execution scenarios
func TestRunMain_ProductionExecution(t *testing.T) {
	t.Run("missing API key fails in production mode", func(t *testing.T) {
		config := createValidMainConfig(t)
		// Remove --dry-run to test actual production mode
		config.Args = removeDryRunFlag(config.Args)
		// Don't set API keys
		config.Getenv = func(key string) string { return "" }

		result := RunMain(config)

		// Should fail due to missing API keys when actually trying to execute
		assert.NotEqual(t, ExitCodeSuccess, result.ExitCode, "Should fail without API keys in production")
		assert.Error(t, result.Error, "Should return error for missing API keys")
	})

	t.Run("handles partial success with tolerance", func(t *testing.T) {
		config := createValidMainConfig(t)
		config.Args = withPartialSuccessOk(config.Args)
		// Set a minimal API key for execution
		config.Getenv = func(key string) string {
			if key == "GEMINI_API_KEY" {
				return "test-key-value"
			}
			return ""
		}

		result := RunMain(config)

		// Should handle execution, potentially with partial success tolerance
		assert.NotNil(t, result, "Should handle partial success scenarios")
		assert.GreaterOrEqual(t, result.ExitCode, 0, "Should have valid exit code")
	})
}

// TestRunMain_ErrorPropagation tests error handling and exit code determination
func TestRunMain_ErrorPropagation(t *testing.T) {
	t.Run("propagates exit codes correctly", func(t *testing.T) {
		tests := []struct {
			name         string
			setupError   func(*MainConfig)
			expectedCode int
		}{
			{
				name: "invalid request",
				setupError: func(mc *MainConfig) {
					mc.Args = []string{"thinktank", "--invalid-flag"}
				},
				expectedCode: ExitCodeInvalidRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := createBasicMainConfig(t)
				tt.setupError(config)

				result := RunMain(config)

				assert.Equal(t, tt.expectedCode, result.ExitCode, "Should propagate correct exit code")
			})
		}
	})

	t.Run("handles wrapped errors", func(t *testing.T) {
		config := createBasicMainConfig(t)
		config.Args = []string{"thinktank", "--instructions", "/nonexistent/file", "--model", "gemini-2.5-pro", "src/"}

		result := RunMain(config)

		assert.NotEqual(t, ExitCodeSuccess, result.ExitCode, "Should fail for nonexistent instructions file")
		assert.Error(t, result.Error, "Should return error")
	})
}

// Helper functions for test setup

func createBasicMainConfig(t *testing.T) *MainConfig {
	t.Helper()
	return &MainConfig{
		FileSystem:  &OSFileSystem{},
		ExitHandler: NewMockExitHandler(),
		Args:        []string{"thinktank"},
		Getenv:      func(string) string { return "" },
		Now:         func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) },
	}
}

func createValidMainConfig(t *testing.T) *MainConfig {
	t.Helper()

	// Create temporary directory and files
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.md")
	inputDir := filepath.Join(tempDir, "input")

	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions for analysis"), 0644))
	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "test.go"), []byte("package main\n\nfunc main() {}\n"), 0644))

	config := createBasicMainConfig(t)
	config.Args = []string{
		"thinktank",
		"--instructions", instructionsFile,
		"--model", "gemini-2.5-pro",
		"--output-dir", filepath.Join(tempDir, "output"),
		"--dry-run", // Add dry-run by default to avoid API calls
		inputDir,
	}

	// Set up environment with minimal config for dry-run
	config.Getenv = func(key string) string {
		return "" // No API keys needed for dry-run
	}

	return config
}

func withTimeout(args []string, timeout string) []string {
	return append(args, "--timeout", timeout)
}

func withDryRun(args []string) []string {
	return append(args, "--dry-run")
}

func withPartialSuccessOk(args []string) []string {
	return append(args, "--partial-success-ok")
}

func withAuditLogFile(args []string, filename string) []string {
	// Flags that don't take values (boolean flags)
	boolFlags := map[string]bool{
		"--dry-run":            true,
		"--verbose":            true,
		"--quiet":              true,
		"--json-logs":          true,
		"--no-progress":        true,
		"--partial-success-ok": true,
	}

	// Find the position of the first non-flag argument (path)
	// We need to insert the audit flag before paths to ensure proper parsing
	for i, arg := range args {
		// Skip the program name and check if this is not a flag or flag value
		if i > 0 && !strings.HasPrefix(arg, "-") {
			// Check if previous arg was a flag that takes a value
			if i > 1 && strings.HasPrefix(args[i-1], "-") {
				// If previous flag is a boolean flag, this is a path
				if boolFlags[args[i-1]] {
					// This is the first path argument, insert before it
					result := make([]string, 0, len(args)+2)
					result = append(result, args[:i]...)
					result = append(result, "--audit-log-file", filename)
					result = append(result, args[i:]...)
					return result
				}
				// Previous was a flag that takes a value, so this is its value, keep looking
				continue
			}
			// This is the first path argument, insert before it
			result := make([]string, 0, len(args)+2)
			result = append(result, args[:i]...)
			result = append(result, "--audit-log-file", filename)
			result = append(result, args[i:]...)
			return result
		}
	}
	// No paths found, append at the end
	return append(args, "--audit-log-file", filename)
}

func removeDryRunFlag(args []string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg != "--dry-run" {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// Mock implementations for testing
// Note: NewMockExitHandler() is already defined in run_mocks.go
