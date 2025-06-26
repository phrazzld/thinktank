package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentVariableIntegration tests end-to-end environment variable integration
func TestEnvironmentVariableIntegration(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions"), 0644))

	testCases := []struct {
		name        string
		envVars     map[string]string
		args        []string
		expected    func(*testing.T, *config.CliConfig)
		description string
	}{
		{
			name: "Environment variables set defaults",
			envVars: map[string]string{
				"THINKTANK_MODEL":      "gpt-4o",
				"THINKTANK_OUTPUT_DIR": "/tmp/output",
				"THINKTANK_DRY_RUN":    "true",
				"THINKTANK_VERBOSE":    "1",
				"THINKTANK_RATE_LIMIT": "1000",
				"THINKTANK_TIMEOUT":    "5m",
			},
			args: []string{"thinktank", "--instructions", instructionsFile, tempDir},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				assert.Equal(t, []string{"gpt-4o"}, cfg.ModelNames, "Model should be set from environment")
				assert.Equal(t, "/tmp/output", cfg.OutputDir, "Output dir should be set from environment")
				assert.True(t, cfg.DryRun, "Dry run should be enabled from environment")
				assert.True(t, cfg.Verbose, "Verbose should be enabled from environment")
				assert.Equal(t, 1000, cfg.RateLimitRequestsPerMinute, "Rate limit should be set from environment")
				assert.Equal(t, 5*time.Minute, cfg.Timeout, "Timeout should be set from environment")
			},
			description: "Should apply environment variable defaults",
		},
		{
			name: "CLI flags override environment variables",
			envVars: map[string]string{
				"THINKTANK_MODEL":      "env-model",
				"THINKTANK_OUTPUT_DIR": "/env/output",
				"THINKTANK_DRY_RUN":    "false",
				"THINKTANK_VERBOSE":    "false",
			},
			args: []string{
				"thinktank",
				"--instructions", instructionsFile,
				"--model", "cli-model",
				"--output-dir", "/cli/output",
				"--dry-run",
				"--verbose",
				tempDir,
			},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				assert.Equal(t, []string{"cli-model"}, cfg.ModelNames, "CLI model should override environment")
				assert.Equal(t, "/cli/output", cfg.OutputDir, "CLI output dir should override environment")
				assert.True(t, cfg.DryRun, "CLI dry run should override environment")
				assert.True(t, cfg.Verbose, "CLI verbose should override environment")
			},
			description: "Should prioritize CLI flags over environment variables",
		},
		{
			name: "Complex environment configuration",
			envVars: map[string]string{
				"THINKTANK_INCLUDE":               "*.go,*.ts",
				"THINKTANK_EXCLUDE":               "*.test.go",
				"THINKTANK_EXCLUDE_NAMES":         ".git,node_modules",
				"THINKTANK_MAX_CONCURRENT":        "10",
				"THINKTANK_RATE_LIMIT_OPENAI":     "3000",
				"THINKTANK_RATE_LIMIT_GEMINI":     "60",
				"THINKTANK_RATE_LIMIT_OPENROUTER": "20",
			},
			args: []string{"thinktank", "--instructions", instructionsFile, tempDir},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				assert.Equal(t, "*.go,*.ts", cfg.Include, "Include patterns should be set from environment")
				assert.Equal(t, "*.test.go", cfg.Exclude, "Exclude patterns should be set from environment")
				assert.Equal(t, ".git,node_modules", cfg.ExcludeNames, "Exclude names should be set from environment")
				assert.Equal(t, 10, cfg.MaxConcurrentRequests, "Max concurrent should be set from environment")
				assert.Equal(t, 3000, cfg.OpenAIRateLimit, "OpenAI rate limit should be set from environment")
				assert.Equal(t, 60, cfg.GeminiRateLimit, "Gemini rate limit should be set from environment")
				assert.Equal(t, 20, cfg.OpenRouterRateLimit, "OpenRouter rate limit should be set from environment")
			},
			description: "Should handle complex environment configuration",
		},
		{
			name:    "No environment variables - uses defaults",
			envVars: map[string]string{}, // No environment variables
			args:    []string{"thinktank", "--instructions", instructionsFile, tempDir},
			expected: func(t *testing.T, cfg *config.CliConfig) {
				// Should use default values
				assert.Equal(t, []string{"gemini-2.5-pro"}, cfg.ModelNames, "Should use default model")
				assert.Equal(t, "", cfg.OutputDir, "Should use default output dir")
				assert.False(t, cfg.DryRun, "Should use default dry run setting")
				assert.False(t, cfg.Verbose, "Should use default verbose setting")
			},
			description: "Should use defaults when no environment variables are set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock environment getter
			getenv := func(key string) string {
				return tc.envVars[key]
			}

			// Parse flags with environment integration
			cfg, err := ParseFlagsWithArgsAndEnv(tc.args, getenv)

			require.NoError(t, err, "ParseFlagsWithArgsAndEnv should not error")
			require.NotNil(t, cfg, "Config should not be nil")

			// Run test-specific assertions
			tc.expected(t, cfg)
		})
	}
}

// TestEnvironmentVariableErrorHandling tests error handling for invalid environment values
func TestEnvironmentVariableErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions"), 0644))

	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorSubstr string
		description string
	}{
		{
			name: "Invalid timeout format",
			envVars: map[string]string{
				"THINKTANK_TIMEOUT": "invalid-duration",
			},
			expectError: true,
			errorSubstr: "invalid timeout format",
			description: "Should return error for invalid timeout",
		},
		{
			name: "Invalid rate limit",
			envVars: map[string]string{
				"THINKTANK_RATE_LIMIT": "not-a-number",
			},
			expectError: true,
			errorSubstr: "invalid rate limit",
			description: "Should return error for invalid rate limit",
		},
		{
			name: "Negative max concurrent",
			envVars: map[string]string{
				"THINKTANK_MAX_CONCURRENT": "-5",
			},
			expectError: true,
			errorSubstr: "max concurrent requests must be positive",
			description: "Should return error for negative max concurrent",
		},
		{
			name: "Valid boolean variations",
			envVars: map[string]string{
				"THINKTANK_DRY_RUN": "yes",
				"THINKTANK_VERBOSE": "on",
				"THINKTANK_QUIET":   "1",
			},
			expectError: false,
			description: "Should accept valid boolean variations",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tc.envVars[key]
			}

			args := []string{"thinktank", "--instructions", instructionsFile, tempDir}
			cfg, err := ParseFlagsWithArgsAndEnv(args, getenv)

			if tc.expectError {
				require.Error(t, err, tc.description)
				if tc.errorSubstr != "" {
					assert.Contains(t, err.Error(), tc.errorSubstr, "Error should contain expected substring")
				}
			} else {
				require.NoError(t, err, tc.description)
				require.NotNil(t, cfg, "Config should not be nil on success")
			}
		})
	}
}

// TestEnvironmentVariablePrecedence tests the specific precedence behavior
func TestEnvironmentVariablePrecedence(t *testing.T) {
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	require.NoError(t, os.WriteFile(instructionsFile, []byte("Test instructions"), 0644))

	t.Run("CLI > Environment > Default precedence", func(t *testing.T) {
		// Environment sets one value, CLI sets another
		envVars := map[string]string{
			"THINKTANK_MODEL":      "env-model",
			"THINKTANK_OUTPUT_DIR": "env-output",
			"THINKTANK_TIMEOUT":    "10m",
		}

		getenv := func(key string) string {
			return envVars[key]
		}

		// CLI overrides model and output, but not timeout
		args := []string{
			"thinktank",
			"--instructions", instructionsFile,
			"--model", "cli-model",
			"--output-dir", "cli-output",
			// Note: no --timeout flag, so should use environment value
			tempDir,
		}

		cfg, err := ParseFlagsWithArgsAndEnv(args, getenv)
		require.NoError(t, err)

		// CLI values should override environment
		assert.Equal(t, []string{"cli-model"}, cfg.ModelNames, "CLI model should override environment")
		assert.Equal(t, "cli-output", cfg.OutputDir, "CLI output should override environment")

		// Environment value should be used when no CLI flag provided
		assert.Equal(t, 10*time.Minute, cfg.Timeout, "Environment timeout should be used when no CLI flag")
	})
}

// TestEnvironmentVariableHelp tests that help text reflects environment-loaded defaults
func TestEnvironmentVariableHelp(t *testing.T) {
	t.Run("Help text shows environment-loaded defaults", func(t *testing.T) {
		// This is a basic test to ensure help text generation doesn't crash
		// More detailed help text testing could be added later
		envVars := map[string]string{
			"THINKTANK_MODEL": "env-model",
		}

		getenv := func(key string) string {
			return envVars[key]
		}

		// This should not crash when generating help text
		args := []string{"thinktank", "--help"}
		_, err := ParseFlagsWithArgsAndEnv(args, getenv)

		// Help will cause parsing to "fail" but shouldn't crash
		// The important thing is that it doesn't panic
		assert.Error(t, err) // Expected because --help causes flag.Parse to return an error
	})
}
