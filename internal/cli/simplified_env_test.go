package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSimplifiedParserEnvironmentVariables tests that the simplified parser respects environment variables
func TestSimplifiedParserEnvironmentVariables(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	targetDir := filepath.Join(tempDir, "src")

	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions"), 0644))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	testCases := []struct {
		name        string
		envVars     map[string]string
		args        []string
		expected    func(*testing.T, *config.CliConfig)
		description string
	}{
		{
			name: "Environment variables applied to simplified mode",
			envVars: map[string]string{
				"THINKTANK_MODEL":      "env-model",
				"THINKTANK_OUTPUT_DIR": "/env/output",
				"THINKTANK_VERBOSE":    "true",
				"THINKTANK_RATE_LIMIT": "500",
				"THINKTANK_TIMEOUT":    "15m",
			},
			args: []string{"thinktank", instructionsFile, targetDir, "--dry-run"},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				assert.Equal(t, []string{"env-model"}, cfg.ModelNames, "Environment model should be applied")
				assert.Equal(t, "/env/output", cfg.OutputDir, "Environment output dir should be applied")
				assert.True(t, cfg.Verbose, "Environment verbose should be applied")
				assert.Equal(t, 500, cfg.RateLimitRequestsPerMinute, "Environment rate limit should be applied")
				assert.Equal(t, 15*time.Minute, cfg.Timeout, "Environment timeout should be applied")
				assert.True(t, cfg.DryRun, "CLI dry-run flag should still work")
			},
			description: "Should apply environment variables to simplified mode",
		},
		{
			name: "Simplified CLI flags override environment",
			envVars: map[string]string{
				"THINKTANK_VERBOSE": "false",
			},
			args: []string{"thinktank", instructionsFile, targetDir, "--verbose", "--dry-run"},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				assert.True(t, cfg.Verbose, "CLI verbose flag should override environment")
				assert.True(t, cfg.DryRun, "CLI dry-run flag should work")
			},
			description: "Should prioritize CLI flags over environment in simplified mode",
		},
		{
			name: "Complex environment variables in simplified mode",
			envVars: map[string]string{
				"THINKTANK_INCLUDE":           "*.go,*.ts",
				"THINKTANK_EXCLUDE":           "*.test.go",
				"THINKTANK_MAX_CONCURRENT":    "8",
				"THINKTANK_RATE_LIMIT_OPENAI": "2000",
				"THINKTANK_RATE_LIMIT_GEMINI": "100",
				"GEMINI_API_KEY":              "mock-api-key", // Add API key to avoid validation error
			},
			args: []string{"thinktank", instructionsFile, targetDir, "--dry-run"},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				assert.Equal(t, "*.go,*.ts", cfg.Include, "Environment include should be applied")
				assert.Equal(t, "*.test.go", cfg.Exclude, "Environment exclude should be applied")
				assert.Equal(t, 8, cfg.MaxConcurrentRequests, "Environment max concurrent should be applied")
				assert.Equal(t, 2000, cfg.OpenAIRateLimit, "Environment OpenAI rate limit should be applied")
				assert.Equal(t, 100, cfg.GeminiRateLimit, "Environment Gemini rate limit should be applied")
			},
			description: "Should handle complex environment variables in simplified mode",
		},
		{
			name:    "No environment variables in simplified mode",
			envVars: map[string]string{}, // No environment variables
			args:    []string{"thinktank", instructionsFile, targetDir, "--dry-run"},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				// Should use default values
				assert.Equal(t, []string{config.DefaultModel}, cfg.ModelNames, "Should use default model")
				assert.Equal(t, "", cfg.OutputDir, "Should use default output dir")
				assert.False(t, cfg.Verbose, "Should use default verbose setting")
				assert.True(t, cfg.DryRun, "Should parse CLI dry-run flag")
			},
			description: "Should use defaults when no environment variables are set in simplified mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create logger for ParserRouter
			logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)

			// Create mock environment getter
			getenv := func(key string) string {
				return tc.envVars[key]
			}

			// Create ParserRouter with environment function
			router := NewParserRouterWithEnv(logger, getenv)

			// Parse arguments - should detect simplified mode and apply environment variables
			result := router.ParseArguments(tc.args)

			require.NoError(t, result.Error, "Parsing should not error")
			require.NotNil(t, result.Config, "Config should not be nil")
			assert.Equal(t, SimplifiedMode, result.Mode, "Should detect simplified mode")

			// Run test-specific assertions
			tc.expected(t, result.Config)
		})
	}
}

// TestSimplifiedParserEnvironmentVariableErrors tests error handling in simplified mode
func TestSimplifiedParserEnvironmentVariableErrors(t *testing.T) {
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	targetDir := filepath.Join(tempDir, "src")

	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions"), 0644))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	testCases := []struct {
		name        string
		envVars     map[string]string
		args        []string
		expectError bool
		errorSubstr string
		description string
	}{
		{
			name: "Invalid timeout in simplified mode",
			envVars: map[string]string{
				"THINKTANK_TIMEOUT": "invalid-timeout",
			},
			args:        []string{"thinktank", instructionsFile, targetDir, "--dry-run"},
			expectError: true,
			errorSubstr: "invalid timeout format",
			description: "Should return error for invalid timeout in simplified mode",
		},
		{
			name: "Invalid rate limit in simplified mode",
			envVars: map[string]string{
				"THINKTANK_RATE_LIMIT": "not-a-number",
			},
			args:        []string{"thinktank", instructionsFile, targetDir, "--dry-run"},
			expectError: true,
			errorSubstr: "invalid rate limit",
			description: "Should return error for invalid rate limit in simplified mode",
		},
		{
			name: "Valid environment variables in simplified mode",
			envVars: map[string]string{
				"THINKTANK_DRY_RUN": "yes",
				"THINKTANK_VERBOSE": "on",
				"THINKTANK_MODEL":   "gpt-4o",
				"GEMINI_API_KEY":    "mock-api-key", // Add API key to avoid validation error
			},
			args:        []string{"thinktank", instructionsFile, targetDir, "--dry-run"},
			expectError: false,
			description: "Should accept valid environment variables in simplified mode",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)

			getenv := func(key string) string {
				return tc.envVars[key]
			}

			router := NewParserRouterWithEnv(logger, getenv)
			result := router.ParseArguments(tc.args)

			if tc.expectError {
				require.Error(t, result.Error, tc.description)
				if tc.errorSubstr != "" {
					assert.Contains(t, result.Error.Error(), tc.errorSubstr, "Error should contain expected substring")
				}
			} else {
				require.NoError(t, result.Error, tc.description)
				require.NotNil(t, result.Config, "Config should not be nil on success")
				assert.Equal(t, SimplifiedMode, result.Mode, "Should detect simplified mode")
			}
		})
	}
}

// TestEnvironmentVariablePrecedenceInBothModes tests precedence across both simplified and complex modes
func TestEnvironmentVariablePrecedenceInBothModes(t *testing.T) {
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	targetDir := filepath.Join(tempDir, "src")

	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions"), 0644))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	envVars := map[string]string{
		"THINKTANK_MODEL":      "env-model",
		"THINKTANK_OUTPUT_DIR": "env-output",
		"THINKTANK_VERBOSE":    "true",
		"THINKTANK_TIMEOUT":    "20m",
	}

	getenv := func(key string) string {
		return envVars[key]
	}

	logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	router := NewParserRouterWithEnv(logger, getenv)

	t.Run("Simplified mode with environment variables", func(t *testing.T) {
		args := []string{"thinktank", instructionsFile, targetDir, "--dry-run"}
		result := router.ParseArguments(args)

		require.NoError(t, result.Error)
		assert.Equal(t, SimplifiedMode, result.Mode)

		// Environment variables should be applied
		assert.Equal(t, []string{"env-model"}, result.Config.ModelNames)
		assert.Equal(t, "env-output", result.Config.OutputDir)
		assert.True(t, result.Config.Verbose)
		assert.Equal(t, 20*time.Minute, result.Config.Timeout)
		assert.True(t, result.Config.DryRun) // From CLI flag
	})

	t.Run("Complex mode with environment variables", func(t *testing.T) {
		args := []string{"thinktank", "--instructions", instructionsFile, "--dry-run", targetDir}
		result := router.ParseArguments(args)

		require.NoError(t, result.Error)
		assert.Equal(t, ComplexMode, result.Mode)

		// Environment variables should be applied in complex mode too
		assert.Equal(t, []string{"env-model"}, result.Config.ModelNames)
		assert.Equal(t, "env-output", result.Config.OutputDir)
		assert.True(t, result.Config.Verbose)
		assert.Equal(t, 20*time.Minute, result.Config.Timeout)
		assert.True(t, result.Config.DryRun) // From CLI flag
	})

	t.Run("CLI flags override environment in both modes", func(t *testing.T) {
		// Test simplified mode with CLI override
		simplifiedArgs := []string{"thinktank", instructionsFile, targetDir, "--dry-run"}
		// Note: simplified mode doesn't support --model or --output-dir flags yet
		// but environment variables should still be applied for other settings

		simplifiedResult := router.ParseArguments(simplifiedArgs)
		require.NoError(t, simplifiedResult.Error)
		assert.Equal(t, SimplifiedMode, simplifiedResult.Mode)

		// Test complex mode with CLI override
		complexArgs := []string{
			"thinktank",
			"--instructions", instructionsFile,
			"--model", "cli-model",
			"--output-dir", "cli-output",
			"--dry-run",
			targetDir,
		}

		complexResult := router.ParseArguments(complexArgs)
		require.NoError(t, complexResult.Error)
		assert.Equal(t, ComplexMode, complexResult.Mode)

		// CLI flags should override environment
		assert.Equal(t, []string{"cli-model"}, complexResult.Config.ModelNames)
		assert.Equal(t, "cli-output", complexResult.Config.OutputDir)
		// But environment values should still be used for flags not specified
		assert.True(t, complexResult.Config.Verbose)
		assert.Equal(t, 20*time.Minute, complexResult.Config.Timeout)
	})
}
