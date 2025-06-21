package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

func TestNewFlagsParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected struct {
			Quiet      bool
			JsonLogs   bool
			NoProgress bool
		}
	}{
		{
			name: "All new flags set to true",
			args: []string{"--quiet", "--json-logs", "--no-progress", "--instructions", "test.txt", "test-path"},
			expected: struct {
				Quiet      bool
				JsonLogs   bool
				NoProgress bool
			}{
				Quiet:      true,
				JsonLogs:   true,
				NoProgress: true,
			},
		},
		{
			name: "No new flags set (defaults to false)",
			args: []string{"--instructions", "test.txt", "test-path"},
			expected: struct {
				Quiet      bool
				JsonLogs   bool
				NoProgress bool
			}{
				Quiet:      false,
				JsonLogs:   false,
				NoProgress: false,
			},
		},
		{
			name: "Only quiet flag set",
			args: []string{"--quiet", "--instructions", "test.txt", "test-path"},
			expected: struct {
				Quiet      bool
				JsonLogs   bool
				NoProgress bool
			}{
				Quiet:      true,
				JsonLogs:   false,
				NoProgress: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			getenv := func(key string) string {
				switch key {
				case config.APIKeyEnvVar:
					return "test-api-key"
				default:
					return ""
				}
			}

			cfg, err := ParseFlagsWithEnv(flagSet, tt.args, getenv)
			if err != nil {
				t.Fatalf("ParseFlagsWithEnv failed: %v", err)
			}

			if cfg.Quiet != tt.expected.Quiet {
				t.Errorf("Expected Quiet=%v, got %v", tt.expected.Quiet, cfg.Quiet)
			}
			if cfg.JsonLogs != tt.expected.JsonLogs {
				t.Errorf("Expected JsonLogs=%v, got %v", tt.expected.JsonLogs, cfg.JsonLogs)
			}
			if cfg.NoProgress != tt.expected.NoProgress {
				t.Errorf("Expected NoProgress=%v, got %v", tt.expected.NoProgress, cfg.NoProgress)
			}
		})
	}
}

func TestValidateInputs(t *testing.T) {
	logger := logutil.NewTestLoggerWithoutAutoFail(t)

	tests := []struct {
		name        string
		config      *config.CliConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidateInputs wrapper passes through success case",
			config: &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"gemini-2.5-pro"}, // Use known valid model
			},
			expectError: false,
		},
		{
			name: "ValidateInputs wrapper passes through error case",
			config: &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"gemini-2.5-pro"},
				Quiet:            true,
				Verbose:          true, // This should cause an error
			},
			expectError: true,
			errorMsg:    "conflicting flags: --quiet and --verbose are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the wrapper function
			err := ValidateInputs(tt.config, logger)

			// Verify it behaves the same as calling ValidateInputsWithEnv with os.Getenv
			expectedErr := ValidateInputsWithEnv(tt.config, logger, os.Getenv)

			// Both should have the same error state
			if (err == nil) != (expectedErr == nil) {
				t.Errorf("ValidateInputs error state doesn't match ValidateInputsWithEnv: got err=%v, expected err=%v", err, expectedErr)
			}

			// If both have errors, they should be the same error
			if err != nil && expectedErr != nil && err.Error() != expectedErr.Error() {
				t.Errorf("ValidateInputs error message doesn't match ValidateInputsWithEnv: got %q, expected %q", err.Error(), expectedErr.Error())
			}

			// Verify the expected test behavior
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateInputsWithEnv_QuietVerboseConflict(t *testing.T) {
	tests := []struct {
		name        string
		quiet       bool
		verbose     bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Conflicting quiet and verbose flags should error",
			quiet:       true,
			verbose:     true,
			expectError: true,
			errorMsg:    "conflicting flags: --quiet and --verbose are mutually exclusive",
		},
		{
			name:        "Quiet flag alone should not error",
			quiet:       true,
			verbose:     false,
			expectError: false,
		},
		{
			name:        "Verbose flag alone should not error",
			quiet:       false,
			verbose:     true,
			expectError: false,
		},
		{
			name:        "Neither flag set should not error",
			quiet:       false,
			verbose:     false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logutil.NewTestLoggerWithoutAutoFail(t)

			config := &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"gemini-2.5-pro"},
				Quiet:            tt.quiet,
				Verbose:          tt.verbose,
			}

			// Simple getenv that provides required API key
			getenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-api-key"
				default:
					return ""
				}
			}

			err := ValidateInputsWithEnv(config, logger, getenv)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateInputsWithEnv_InstructionsFileValidation(t *testing.T) {
	tests := []struct {
		name             string
		instructionsFile string
		dryRun           bool
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "Missing instructions file with dry run should not error",
			instructionsFile: "",
			dryRun:           true,
			expectError:      false,
		},
		{
			name:             "Missing instructions file without dry run should error",
			instructionsFile: "",
			dryRun:           false,
			expectError:      true,
			errorMsg:         "missing required --instructions flag",
		},
		{
			name:             "Has instructions file with dry run should not error",
			instructionsFile: "test.txt",
			dryRun:           true,
			expectError:      false,
		},
		{
			name:             "Has instructions file without dry run should not error",
			instructionsFile: "test.txt",
			dryRun:           false,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logutil.NewTestLoggerWithoutAutoFail(t)

			config := &config.CliConfig{
				InstructionsFile: tt.instructionsFile,
				Paths:            []string{"test-path"},
				ModelNames:       []string{"gemini-2.5-pro"},
				DryRun:           tt.dryRun,
			}

			// Simple getenv that provides required API key
			getenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-api-key"
				default:
					return ""
				}
			}

			err := ValidateInputsWithEnv(config, logger, getenv)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateInputsWithEnv_InputPathsValidation(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Missing input paths should error",
			paths:       []string{},
			expectError: true,
			errorMsg:    "no paths specified",
		},
		{
			name:        "Has input paths should not error",
			paths:       []string{"test-path"},
			expectError: false,
		},
		{
			name:        "Multiple input paths should not error",
			paths:       []string{"path1", "path2", "path3"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logutil.NewTestLoggerWithoutAutoFail(t)

			config := &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            tt.paths,
				ModelNames:       []string{"gemini-2.5-pro"},
			}

			// Simple getenv that provides required API key
			getenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-api-key"
				default:
					return ""
				}
			}

			err := ValidateInputsWithEnv(config, logger, getenv)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateInputsWithEnv_ModelValidation(t *testing.T) {
	tests := []struct {
		name           string
		modelNames     []string
		synthesisModel string
		dryRun         bool
		expectError    bool
		errorMsg       string
	}{
		{
			name:        "Missing model names with dry run should not error",
			modelNames:  []string{},
			dryRun:      true,
			expectError: false,
		},
		{
			name:        "Missing model names without dry run should error",
			modelNames:  []string{},
			dryRun:      false,
			expectError: true,
			errorMsg:    "no models specified",
		},
		{
			name:        "Has model names should not error",
			modelNames:  []string{"gemini-2.5-pro"},
			dryRun:      false,
			expectError: false,
		},
		{
			name:           "Valid synthesis model gpt- prefix should not error",
			modelNames:     []string{"gemini-2.5-pro"},
			synthesisModel: "gpt-4",
			expectError:    false,
		},
		{
			name:           "Valid synthesis model text- prefix should not error",
			modelNames:     []string{"gemini-2.5-pro"},
			synthesisModel: "text-davinci-003",
			expectError:    false,
		},
		{
			name:           "Valid synthesis model gemini- prefix should not error",
			modelNames:     []string{"gpt-4"},
			synthesisModel: "gemini-2.5-pro",
			expectError:    false,
		},
		{
			name:           "Valid synthesis model claude- prefix should not error",
			modelNames:     []string{"gpt-4"},
			synthesisModel: "claude-3-opus",
			expectError:    false,
		},
		{
			name:           "Valid synthesis model containing openai should not error",
			modelNames:     []string{"gemini-2.5-pro"},
			synthesisModel: "some-openai-model",
			expectError:    false,
		},
		{
			name:           "Valid synthesis model containing openrouter/ should not error",
			modelNames:     []string{"gemini-2.5-pro"},
			synthesisModel: "openrouter/anthropic/claude-3",
			expectError:    false,
		},
		{
			name:           "Invalid synthesis model should error",
			modelNames:     []string{"gemini-2.5-pro"},
			synthesisModel: "unknown-invalid-model",
			expectError:    true,
			errorMsg:       "invalid synthesis model: 'unknown-invalid-model' does not match any known model pattern",
		},
		{
			name:           "Empty synthesis model should not error",
			modelNames:     []string{"gemini-2.5-pro"},
			synthesisModel: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logutil.NewTestLoggerWithoutAutoFail(t)

			config := &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       tt.modelNames,
				SynthesisModel:   tt.synthesisModel,
				DryRun:           tt.dryRun,
			}

			// Simple getenv that provides required API key
			getenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-api-key"
				default:
					return ""
				}
			}

			err := ValidateInputsWithEnv(config, logger, getenv)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	// Test the ParseFlags wrapper function
	// This function calls ParseFlagsWithEnv with flag.NewFlagSet, os.Args[1:], and os.Getenv

	// Since ParseFlags uses os.Args and os.Getenv, we need to test it indirectly
	// by verifying it produces the same result as calling ParseFlagsWithEnv directly

	// Save original os.Args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Set up test args
	testArgs := []string{"thinktank", "--instructions", "test.txt", "--model", "gemini-2.5-pro", "test-path"}
	os.Args = testArgs

	// Set up environment
	_ = os.Setenv("GEMINI_API_KEY", "test-api-key")
	defer func() { _ = os.Unsetenv("GEMINI_API_KEY") }()

	// Call ParseFlags
	cfg, err := ParseFlags()

	// Verify no error occurred
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Verify basic configuration was set correctly
	if cfg.InstructionsFile != "test.txt" {
		t.Errorf("Expected InstructionsFile='test.txt', got %q", cfg.InstructionsFile)
	}

	if len(cfg.ModelNames) != 1 || cfg.ModelNames[0] != "gemini-2.5-pro" {
		t.Errorf("Expected ModelNames=['gemini-2.5-pro'], got %v", cfg.ModelNames)
	}

	if len(cfg.Paths) != 1 || cfg.Paths[0] != "test-path" {
		t.Errorf("Expected Paths=['test-path'], got %v", cfg.Paths)
	}
}

func TestParseFlagsWithEnv_RateLimitingFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected struct {
			MaxConcurrentRequests      int
			RateLimitRequestsPerMinute int
		}
	}{
		{
			name: "Default rate limiting values",
			args: []string{"--instructions", "test.txt", "test-path"},
			expected: struct {
				MaxConcurrentRequests      int
				RateLimitRequestsPerMinute int
			}{
				MaxConcurrentRequests:      5,
				RateLimitRequestsPerMinute: 60,
			},
		},
		{
			name: "Custom max concurrent requests",
			args: []string{"--instructions", "test.txt", "--max-concurrent", "10", "test-path"},
			expected: struct {
				MaxConcurrentRequests      int
				RateLimitRequestsPerMinute int
			}{
				MaxConcurrentRequests:      10,
				RateLimitRequestsPerMinute: 60,
			},
		},
		{
			name: "Custom rate limit requests per minute",
			args: []string{"--instructions", "test.txt", "--rate-limit", "120", "test-path"},
			expected: struct {
				MaxConcurrentRequests      int
				RateLimitRequestsPerMinute int
			}{
				MaxConcurrentRequests:      5,
				RateLimitRequestsPerMinute: 120,
			},
		},
		{
			name: "Both rate limiting flags set",
			args: []string{"--instructions", "test.txt", "--max-concurrent", "3", "--rate-limit", "30", "test-path"},
			expected: struct {
				MaxConcurrentRequests      int
				RateLimitRequestsPerMinute int
			}{
				MaxConcurrentRequests:      3,
				RateLimitRequestsPerMinute: 30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			getenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-api-key"
				default:
					return ""
				}
			}

			cfg, err := ParseFlagsWithEnv(flagSet, tt.args, getenv)
			if err != nil {
				t.Fatalf("ParseFlagsWithEnv failed: %v", err)
			}

			if cfg.MaxConcurrentRequests != tt.expected.MaxConcurrentRequests {
				t.Errorf("Expected MaxConcurrentRequests=%d, got %d", tt.expected.MaxConcurrentRequests, cfg.MaxConcurrentRequests)
			}
			if cfg.RateLimitRequestsPerMinute != tt.expected.RateLimitRequestsPerMinute {
				t.Errorf("Expected RateLimitRequestsPerMinute=%d, got %d", tt.expected.RateLimitRequestsPerMinute, cfg.RateLimitRequestsPerMinute)
			}
		})
	}
}

func TestParseFlagsWithEnv_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Invalid flag should cause parse error",
			args:        []string{"--invalid-flag", "value", "test-path"},
			expectError: true,
			errorMsg:    "error parsing flags",
		},
		{
			name:        "Invalid octal permission for dir-permissions",
			args:        []string{"--instructions", "test.txt", "--dir-permissions", "invalid", "test-path"},
			expectError: true,
			errorMsg:    "invalid directory permission format",
		},
		{
			name:        "Invalid octal permission for file-permissions",
			args:        []string{"--instructions", "test.txt", "--file-permissions", "invalid", "test-path"},
			expectError: true,
			errorMsg:    "invalid file permission format",
		},
		{
			name:        "Invalid log level",
			args:        []string{"--instructions", "test.txt", "--log-level", "invalid", "test-path"},
			expectError: false, // ParseLogLevel error is silently ignored, using default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			getenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-api-key"
				default:
					return ""
				}
			}

			cfg, err := ParseFlagsWithEnv(flagSet, tt.args, getenv)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify config was created successfully
				if cfg == nil {
					t.Errorf("Expected config to be created but got nil")
				}
			}
		})
	}
}

func TestParseOctalPermission_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		permStr     string
		expectError bool
	}{
		{
			name:        "Valid octal permission",
			permStr:     "0755",
			expectError: false,
		},
		{
			name:        "Valid octal permission without leading zero",
			permStr:     "755",
			expectError: false,
		},
		{
			name:        "Invalid octal permission - non-numeric",
			permStr:     "invalid",
			expectError: true,
		},
		{
			name:        "Invalid octal permission - contains 8",
			permStr:     "0788",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseOctalPermission(tt.permStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetFriendlyErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: "An unknown error occurred",
		},
		{
			name:     "API key error",
			err:      fmt.Errorf("api key not found"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "Auth error",
			err:      fmt.Errorf("unauthorized access"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "Rate limit error",
			err:      fmt.Errorf("rate limit exceeded"),
			expected: "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits.",
		},
		{
			name:     "Too many requests error",
			err:      fmt.Errorf("too many requests"),
			expected: "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits.",
		},
		{
			name:     "Timeout error",
			err:      fmt.Errorf("operation timed out"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "Deadline exceeded error",
			err:      fmt.Errorf("deadline exceeded"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "Not found error",
			err:      fmt.Errorf("file not found"),
			expected: "Resource not found. Please check that the specified file paths or models exist.",
		},
		{
			name:     "File permission error",
			err:      fmt.Errorf("file permission denied"),
			expected: "File permission error: Please check file permissions and try again.",
		},
		{
			name:     "General file error",
			err:      fmt.Errorf("file access error"),
			expected: "File error: file access error",
		},
		{
			name:     "Flag usage error",
			err:      fmt.Errorf("flag usage: invalid argument"),
			expected: "Invalid command line arguments. Use --help to see usage instructions.",
		},
		{
			name:     "Context cancelled error",
			err:      fmt.Errorf("context canceled"),
			expected: "Operation was cancelled. This might be due to timeout or user interruption.",
		},
		{
			name:     "Network error",
			err:      fmt.Errorf("network connection failed"),
			expected: "Network error: Please check your internet connection and try again.",
		},
		{
			name:     "Generic error",
			err:      fmt.Errorf("something went wrong"),
			expected: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFriendlyErrorMessage(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Clean message",
			message:  "simple error message",
			expected: "simple error message",
		},
		{
			name:     "OpenAI API key with sk- prefix",
			message:  "error with sk-1234567890abcdef",
			expected: "error with [REDACTED]",
		},
		{
			name:     "Gemini API key with key- prefix",
			message:  "error with key-1234567890abcdef",
			expected: "error with [REDACTED]",
		},
		{
			name:     "Long alphanumeric string",
			message:  "error with abcdef1234567890abcdef1234567890abcdef",
			expected: "error with [REDACTED]",
		},
		{
			name:     "URL with credentials",
			message:  "error connecting to https://user:pass@example.com/api",
			expected: "error connecting to [REDACTED]/api",
		},
		{
			name:     "GEMINI_API_KEY environment variable",
			message:  "GEMINI_API_KEY=sk-1234567890abcdef",
			expected: "[REDACTED]",
		},
		{
			name:     "OPENAI_API_KEY environment variable",
			message:  "OPENAI_API_KEY=sk-1234567890abcdef",
			expected: "[REDACTED]",
		},
		{
			name:     "OPENROUTER_API_KEY environment variable",
			message:  "OPENROUTER_API_KEY=sk-1234567890abcdef",
			expected: "[REDACTED]",
		},
		{
			name:     "API_KEY environment variable",
			message:  "API_KEY=sk-1234567890abcdef",
			expected: "[REDACTED]",
		},
		{
			name:     "Multiple patterns",
			message:  "GEMINI_API_KEY=key-1234567890abcdef and sk-abcdef1234567890",
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.message)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFlagValidation(t *testing.T) {
	logger := logutil.NewTestLoggerWithoutAutoFail(t)

	tests := []struct {
		name        string
		config      *config.CliConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Quiet and verbose flags conflict",
			config: &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"test-model"},
				Quiet:            true,
				Verbose:          true,
			},
			expectError: true,
			errorMsg:    "conflicting flags: --quiet and --verbose are mutually exclusive",
		},
		{
			name: "Quiet and json-logs can coexist",
			config: &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"test-model"},
				Quiet:            true,
				JsonLogs:         true,
			},
			expectError: false,
		},
		{
			name: "All new flags together (without verbose)",
			config: &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"test-model"},
				Quiet:            true,
				JsonLogs:         true,
				NoProgress:       true,
			},
			expectError: false,
		},
		{
			name: "Verbose alone is fine",
			config: &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"test-path"},
				ModelNames:       []string{"test-model"},
				Verbose:          true,
			},
			expectError: false,
		},
	}

	getenv := func(key string) string {
		switch key {
		case config.APIKeyEnvVar:
			return "test-api-key"
		default:
			return ""
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputsWithEnv(tt.config, logger, getenv)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
