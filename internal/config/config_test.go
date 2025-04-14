package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/logutil"
)

func TestNewDefaultCliConfig(t *testing.T) {
	cfg := NewDefaultCliConfig()

	// Verify default values
	if cfg.Format != DefaultFormat {
		t.Errorf("Expected Format to be %q, got %q", DefaultFormat, cfg.Format)
	}

	if cfg.Exclude != DefaultExcludes {
		t.Errorf("Expected Exclude to be %q, got %q", DefaultExcludes, cfg.Exclude)
	}

	if cfg.ExcludeNames != DefaultExcludeNames {
		t.Errorf("Expected ExcludeNames to be %q, got %q", DefaultExcludeNames, cfg.ExcludeNames)
	}

	if len(cfg.ModelNames) != 1 || cfg.ModelNames[0] != DefaultModel {
		t.Errorf("Expected ModelNames to be [%q], got %v", DefaultModel, cfg.ModelNames)
	}

	if cfg.LogLevel != logutil.InfoLevel {
		t.Errorf("Expected LogLevel to be %v, got %v", logutil.InfoLevel, cfg.LogLevel)
	}

	if cfg.MaxConcurrentRequests != DefaultMaxConcurrentRequests {
		t.Errorf("Expected MaxConcurrentRequests to be %d, got %d", DefaultMaxConcurrentRequests, cfg.MaxConcurrentRequests)
	}

	if cfg.RateLimitRequestsPerMinute != DefaultRateLimitRequestsPerMinute {
		t.Errorf("Expected RateLimitRequestsPerMinute to be %d, got %d", DefaultRateLimitRequestsPerMinute, cfg.RateLimitRequestsPerMinute)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default values for all fields
	if cfg.OutputFile != DefaultOutputFile {
		t.Errorf("Expected OutputFile to be %q, got %q", DefaultOutputFile, cfg.OutputFile)
	}

	if cfg.ModelName != DefaultModel {
		t.Errorf("Expected ModelName to be %q, got %q", DefaultModel, cfg.ModelName)
	}

	if cfg.Format != DefaultFormat {
		t.Errorf("Expected Format to be %q, got %q", DefaultFormat, cfg.Format)
	}

	if cfg.LogLevel != logutil.InfoLevel {
		t.Errorf("Expected LogLevel to be %v, got %v", logutil.InfoLevel, cfg.LogLevel)
	}

	if cfg.ConfirmTokens != 0 {
		t.Errorf("Expected ConfirmTokens to be 0, got %d", cfg.ConfirmTokens)
	}

	// Verify nested ExcludeConfig values
	if cfg.Excludes.Extensions != DefaultExcludes {
		t.Errorf("Expected Excludes.Extensions to be %q, got %q", DefaultExcludes, cfg.Excludes.Extensions)
	}

	if cfg.Excludes.Names != DefaultExcludeNames {
		t.Errorf("Expected Excludes.Names to be %q, got %q", DefaultExcludeNames, cfg.Excludes.Names)
	}

	// Verify default values for fields that should be empty or zero
	if cfg.Include != "" {
		t.Errorf("Expected Include to be empty, got %q", cfg.Include)
	}

	if cfg.Verbose != false {
		t.Errorf("Expected Verbose to be false, got %v", cfg.Verbose)
	}
}

// MockLogger is a simple logger implementation for testing
type MockLogger struct {
	ErrorCalled   bool
	ErrorMessages []string
}

func (l *MockLogger) Error(format string, args ...interface{}) {
	l.ErrorCalled = true
	l.ErrorMessages = append(l.ErrorMessages, fmt.Sprintf(format, args...))
}

func (l *MockLogger) Debug(format string, args ...interface{})  {}
func (l *MockLogger) Info(format string, args ...interface{})   {}
func (l *MockLogger) Warn(format string, args ...interface{})   {}
func (l *MockLogger) Fatal(format string, args ...interface{})  {}
func (l *MockLogger) Printf(format string, args ...interface{}) {}
func (l *MockLogger) Println(v ...interface{})                  {}

// TestValidateConfig tests the ValidateConfig function
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *CliConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid configuration",
			config: &CliConfig{
				InstructionsFile: "instructions.md",
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
			},
			expectError: false,
		},
		{
			name: "Missing instructions file",
			config: &CliConfig{
				InstructionsFile: "", // Missing
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
			},
			expectError:   true,
			errorContains: "missing required --instructions flag",
		},
		{
			name: "Missing paths",
			config: &CliConfig{
				InstructionsFile: "instructions.md",
				Paths:            []string{}, // Empty
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
			},
			expectError:   true,
			errorContains: "no paths specified",
		},
		{
			name: "Missing API key",
			config: &CliConfig{
				InstructionsFile: "instructions.md",
				Paths:            []string{"testfile"},
				APIKey:           "", // Missing
				ModelNames:       []string{"model1"},
			},
			expectError:   true,
			errorContains: "API key not set",
		},
		{
			name: "Missing models",
			config: &CliConfig{
				InstructionsFile: "instructions.md",
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{}, // Empty
			},
			expectError:   true,
			errorContains: "no models specified",
		},
		{
			name: "Dry run allows missing instructions file",
			config: &CliConfig{
				InstructionsFile: "", // Missing but allowed in dry run
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				DryRun:           true,
			},
			expectError: false,
		},
		{
			name: "Dry run still requires paths",
			config: &CliConfig{
				InstructionsFile: "",         // Missing but allowed in dry run
				Paths:            []string{}, // Empty paths still invalid
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				DryRun:           true,
			},
			expectError:   true,
			errorContains: "no paths specified",
		},
		{
			name: "Dry run still requires API key",
			config: &CliConfig{
				InstructionsFile: "", // Missing but allowed in dry run
				Paths:            []string{"testfile"},
				APIKey:           "", // Missing - still required in dry run
				ModelNames:       []string{"model1"},
				DryRun:           true,
			},
			expectError:   true,
			errorContains: "API key not set",
		},
		{
			name: "Dry run allows missing models",
			config: &CliConfig{
				InstructionsFile: "", // Missing but allowed in dry run
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{}, // Empty - allowed in dry run
				DryRun:           true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &MockLogger{}
			err := ValidateConfig(tt.config, logger)

			// Check if error matches expectation
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateConfig() error = %v, expectError %v", err, tt.expectError)
			}

			// Verify error contains expected text
			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error message %q doesn't contain expected text %q", err.Error(), tt.errorContains)
				}
			}

			// Verify logger recorded errors for error cases
			if tt.expectError && !logger.ErrorCalled {
				t.Error("Expected error to be logged, but no error was logged")
			}

			if !tt.expectError && logger.ErrorCalled {
				t.Error("No error expected, but error was logged")
			}
		})
	}
}
