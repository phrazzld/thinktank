package cli

import (
	"flag"
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
