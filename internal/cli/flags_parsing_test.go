package cli

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
)

// TestParseFlagsWithEnv tests direct flag parsing without subprocess
func TestParseFlagsWithEnv(t *testing.T) {
	// Mock environment function that returns empty values
	mockGetenv := func(key string) string {
		return ""
	}

	t.Run("valid flags", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--instructions", "/path/to/instructions.md",
			"--model", "gemini-2.5-pro",
			"--output-dir", "/tmp/output",
			"--verbose",
			"--dry-run",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for valid flags")
		assert.NotNil(t, cfg, "Should return a valid config")
		assert.Equal(t, "/path/to/instructions.md", cfg.InstructionsFile)
		assert.Equal(t, []string{"gemini-2.5-pro"}, cfg.ModelNames)
		assert.Equal(t, "/tmp/output", cfg.OutputDir)
		assert.True(t, cfg.Verbose)
		assert.True(t, cfg.DryRun)
		assert.Equal(t, []string{"/path/to/input"}, cfg.Paths)
		assert.Equal(t, logutil.DebugLevel, cfg.LogLevel) // Verbose should set debug level
	})

	t.Run("multiple models", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--instructions", "/path/to/instructions.md",
			"--model", "gemini-2.5-pro",
			"--model", "gpt-4",
			"--model", "claude-3-opus",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for multiple models")
		assert.Equal(t, []string{"gemini-2.5-pro", "gpt-4", "claude-3-opus"}, cfg.ModelNames)
	})

	t.Run("default values", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{"/path/to/input"} // Minimal args

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for minimal valid flags")
		assert.Equal(t, []string{config.DefaultModel}, cfg.ModelNames) // Should use default model
		assert.Equal(t, logutil.InfoLevel, cfg.LogLevel)               // Default log level
		assert.False(t, cfg.Verbose)
		assert.False(t, cfg.Quiet)
		assert.False(t, cfg.DryRun)
		assert.Equal(t, config.DefaultTimeout, cfg.Timeout)
		assert.Equal(t, 5, cfg.MaxConcurrentRequests)       // Default value
		assert.Equal(t, 60, cfg.RateLimitRequestsPerMinute) // Default value
	})

	t.Run("log level parsing", func(t *testing.T) {
		tests := []struct {
			logLevel string
			expected logutil.LogLevel
		}{
			{"debug", logutil.DebugLevel},
			{"info", logutil.InfoLevel},
			{"warn", logutil.WarnLevel},
			{"error", logutil.ErrorLevel},
		}

		for _, tt := range tests {
			t.Run("log_level_"+tt.logLevel, func(t *testing.T) {
				flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
				args := []string{
					"--log-level", tt.logLevel,
					"/path/to/input",
				}

				cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

				assert.NoError(t, err, "Should not return error for valid log level")
				assert.Equal(t, tt.expected, cfg.LogLevel)
			})
		}
	})

	t.Run("boolean flags", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--verbose",
			"--quiet",
			"--json-logs",
			"--no-progress",
			"--dry-run",
			"--partial-success-ok",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for boolean flags")
		assert.True(t, cfg.Verbose)
		assert.True(t, cfg.Quiet)
		assert.True(t, cfg.JsonLogs)
		assert.True(t, cfg.NoProgress)
		assert.True(t, cfg.DryRun)
		assert.True(t, cfg.PartialSuccessOk)
	})

	t.Run("rate limiting flags", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--max-concurrent", "10",
			"--rate-limit", "120",
			"--openai-rate-limit", "3000",
			"--gemini-rate-limit", "60",
			"--openrouter-rate-limit", "20",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for rate limiting flags")
		assert.Equal(t, 10, cfg.MaxConcurrentRequests)
		assert.Equal(t, 120, cfg.RateLimitRequestsPerMinute)
		assert.Equal(t, 3000, cfg.OpenAIRateLimit)
		assert.Equal(t, 60, cfg.GeminiRateLimit)
		assert.Equal(t, 20, cfg.OpenRouterRateLimit)
	})

	t.Run("timeout flag", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--timeout", "5m",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for timeout flag")
		assert.Equal(t, 5*time.Minute, cfg.Timeout)
	})

	t.Run("file filtering flags", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--include", ".go,.md,.txt",
			"--exclude", ".exe,.bin",
			"--exclude-names", "node_modules,dist",
			"--format", "File: {path}\nContent: {content}",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for filtering flags")
		assert.Equal(t, ".go,.md,.txt", cfg.Include)
		assert.Equal(t, ".exe,.bin", cfg.Exclude)
		assert.Equal(t, "node_modules,dist", cfg.ExcludeNames)
		assert.Equal(t, "File: {path}\nContent: {content}", cfg.Format)
	})
}

// TestParseFlagsWithEnvErrors tests error conditions in flag parsing
func TestParseFlagsWithEnvErrors(t *testing.T) {
	mockGetenv := func(key string) string {
		return ""
	}

	t.Run("invalid directory permissions", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--dir-permissions", "invalid",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.Error(t, err, "Should return error for invalid directory permissions")
		assert.Nil(t, cfg, "Should not return config on error")
		assert.Contains(t, err.Error(), "invalid directory permission format")
	})

	t.Run("invalid file permissions", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--file-permissions", "999", // Invalid octal (9 is not valid in octal)
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.Error(t, err, "Should return error for invalid file permissions")
		assert.Nil(t, cfg, "Should not return config on error")
		assert.Contains(t, err.Error(), "invalid file permission format")
	})

	t.Run("valid octal permissions", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--dir-permissions", "0755",
			"--file-permissions", "0644",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.NoError(t, err, "Should not return error for valid octal permissions")
		assert.Equal(t, os.FileMode(0755), cfg.DirPermissions)
		assert.Equal(t, os.FileMode(0644), cfg.FilePermissions)
	})

	t.Run("invalid flag", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--invalid-flag", "value",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.Error(t, err, "Should return error for invalid flag")
		assert.Nil(t, cfg, "Should not return config on error")
		assert.Contains(t, err.Error(), "error parsing flags")
	})

	t.Run("invalid timeout format", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--timeout", "invalid-duration",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

		assert.Error(t, err, "Should return error for invalid timeout format")
		assert.Nil(t, cfg, "Should not return config on error")
		assert.Contains(t, err.Error(), "error parsing flags")
	})

	t.Run("invalid integer flags", func(t *testing.T) {
		tests := []struct {
			flag  string
			value string
		}{
			{"--max-concurrent", "not-a-number"},
			{"--rate-limit", "invalid"},
			{"--openai-rate-limit", "abc"},
		}

		for _, tt := range tests {
			t.Run(tt.flag+"_invalid", func(t *testing.T) {
				flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
				args := []string{
					tt.flag, tt.value,
					"/path/to/input",
				}

				cfg, err := ParseFlagsWithEnv(flagSet, args, mockGetenv)

				assert.Error(t, err, "Should return error for invalid integer flag")
				assert.Nil(t, cfg, "Should not return config on error")
				assert.Contains(t, err.Error(), "error parsing flags")
			})
		}
	})
}

// TestParseOctalPermission tests the octal permission parsing helper
func TestParseOctalPermission(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    os.FileMode
		shouldError bool
	}{
		{
			name:     "valid octal with 0 prefix",
			input:    "0755",
			expected: 0755,
		},
		{
			name:     "valid octal without 0 prefix",
			input:    "644",
			expected: 0644,
		},
		{
			name:     "full permission",
			input:    "0777",
			expected: 0777,
		},
		{
			name:     "minimal permission",
			input:    "0000",
			expected: 0000,
		},
		{
			name:        "invalid octal digit",
			input:       "0999", // 9 is not valid in octal
			shouldError: true,
		},
		{
			name:        "non-numeric",
			input:       "invalid",
			shouldError: true,
		},
		{
			name:        "empty string",
			input:       "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseOctalPermission(tt.input)

			if tt.shouldError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "Expected permission %#o, got %#o", tt.expected, result)
			}
		})
	}
}

// TestMainFunctionValidation tests Main() validation logic directly
func TestMainFunctionValidation(t *testing.T) {
	// This tests the Main() function's flag parsing and validation flow
	// without using subprocess execution

	t.Run("ParseFlags can be called without error in normal case", func(t *testing.T) {
		// We can't easily test ParseFlags() directly because it uses os.Args
		// But we can test that ParseFlagsWithEnv works as the core logic
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"--instructions", "/path/to/instructions.md",
			"--model", "gemini-2.5-pro",
			"/path/to/input",
		}

		cfg, err := ParseFlagsWithEnv(flagSet, args, func(string) string { return "" })

		assert.NoError(t, err, "ParseFlagsWithEnv should work with valid args")
		assert.NotNil(t, cfg, "Should return valid config")

		// Test that ValidateInputs works with this config
		logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")
		validateErr := ValidateInputs(cfg, logger)
		assert.NoError(t, validateErr, "ValidateInputs should pass with valid config")
	})
}
