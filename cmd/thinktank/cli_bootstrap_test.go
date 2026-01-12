// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestParseFlags_UnknownFlag tests parsing an unknown flag
func TestParseFlags_UnknownFlag(t *testing.T) {
	// Removed t.Parallel() - uses os.Getenv
	// Create a flag set

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress flag error output

	// Parse with unknown flag
	_, err := ParseFlagsWithEnv(fs, []string{"--unknown-flag"}, os.Getenv)

	// Should get an error
	if err == nil {
		t.Fatal("Expected error when parsing unknown flag, got nil")
	}

	// Error should mention the unknown flag
	if err.Error() == "" {
		t.Error("Expected error message to be non-empty")
	}
}

// TestParseFlags_InvalidValue tests parsing an invalid value for a flag
func TestParseFlags_InvalidValue(t *testing.T) {
	// Removed t.Parallel() - uses os.Getenv
	// Create a flag set

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress flag error output

	// Test cases with different invalid values
	testCases := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			name:   "Invalid max-concurrent",
			args:   []string{"--max-concurrent=invalid"},
			errMsg: "invalid value",
		},
		{
			name:   "Invalid rate-limit",
			args:   []string{"--rate-limit=invalid"},
			errMsg: "invalid value",
		},
		{
			name:   "Invalid timeout",
			args:   []string{"--timeout=abc"},
			errMsg: "invalid value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)

			// Parse with invalid value
			_, err := ParseFlagsWithEnv(fs, tc.args, os.Getenv)

			// Should get an error
			if err == nil {
				t.Fatal("Expected error when parsing invalid value, got nil")
			}

			// Error message should contain expected text
			if !strings.Contains(fmt.Sprint(err), tc.errMsg) {
				t.Errorf("Expected error message to contain %q, got: %v", tc.errMsg, err)
			}
		})
	}
}

// TestValidateInputs_MissingModels tests validation with missing models
func TestValidateInputs_MissingModels(t *testing.T) {
	// Removed t.Parallel() - creates temporary files and uses env
	// Create a test instructions file

	tmpInstructions, err := os.CreateTemp("", "test-instructions-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp instructions file: %v", err)
	}
	defer func() { _ = os.Remove(tmpInstructions.Name()) }()

	// Empty model list in non-dry-run mode
	config := &config.CliConfig{
		InstructionsFile: tmpInstructions.Name(),
		Paths:            []string{"testpath"},
		ModelNames:       []string{}, // empty
		DryRun:           false,      // not dry run
	}

	// Create a mock environment function
	mockGetenv := func(key string) string {
		return "mock-key" // Return a mock key for any env var
	}

	// Create a logger
	logger := logutil.NewLogger(logutil.InfoLevel, io.Discard, "[test] ")

	// Validate inputs
	err = ValidateInputsWithEnv(config, logger, mockGetenv)

	// Should get an error about missing models
	if err == nil {
		t.Fatal("Expected error when missing models in non-dry-run mode, got nil")
	}

	// Check the error message
	if !strings.Contains(fmt.Sprint(err), "no models specified") {
		t.Errorf("Expected error message to mention missing models, got: %v", err)
	}
}

// TestValidateInputs_ConfigPathNotFound tests validation with a missing config path
func TestValidateInputs_ConfigPathNotFound(t *testing.T) {
	// Removed t.Parallel() - test uses filesystem operations
	// Skip this test for now as the ValidateInputs function doesn't
	// directly check for config file existence - it's handled at initialization

	t.Skip("Skipping test as ValidateInputs doesn't validate config path existence")
}

// TestParseFlags_Timeout tests parsing of the timeout flag
func TestParseFlags_Timeout(t *testing.T) {
	// Removed t.Parallel() - uses mock environment function
	// Create a flag set

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress flag error output

	// Test cases
	testCases := []struct {
		name            string
		args            []string
		expectedTimeout string
		expectError     bool
		errorSubstring  string
	}{
		{
			name:            "Set 2 minute timeout",
			args:            []string{"--timeout=2m"},
			expectedTimeout: "2m0s",
			expectError:     false,
		},
		{
			name:            "Set 30 second timeout",
			args:            []string{"--timeout", "30s"},
			expectedTimeout: "30s",
			expectError:     false,
		},
		{
			name:            "Default timeout",
			args:            []string{"--instructions=test.txt"},
			expectedTimeout: "10m0s", // Default is 10 minutes
			expectError:     false,
		},
		{
			name:            "Invalid timeout",
			args:            []string{"--timeout=invalid"},
			expectedTimeout: "",
			expectError:     true,
			errorSubstring:  "invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag set for each test case
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)

			// Parse with the test args
			cfg, err := ParseFlagsWithEnv(fs, tc.args, func(string) string { return "mock-value" })

			// Check error expectations
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}

			// If expecting an error, check error message
			if tc.expectError && err != nil && tc.errorSubstring != "" {
				if !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorSubstring, err)
				}
			}

			// Check the parsed timeout
			if !tc.expectError && cfg != nil {
				if cfg.Timeout.String() != tc.expectedTimeout {
					t.Errorf("Expected Timeout: %s, got: %s", tc.expectedTimeout, cfg.Timeout.String())
				}
			}
		})
	}
}

// TestParseFlags_SynthesisModel tests parsing of the synthesis-model flag
func TestParseFlags_SynthesisModel(t *testing.T) {
	// Removed t.Parallel() - uses mock environment function
	// Create a flag set

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress flag error output

	// Test cases
	testCases := []struct {
		name           string
		args           []string
		expectedModel  string
		expectError    bool
		errorSubstring string
	}{
		{
			name:          "Set synthesis model",
			args:          []string{"--synthesis-model=gpt-5.2"},
			expectedModel: "gpt-5.2",
			expectError:   false,
		},
		{
			name:          "Set synthesis model with space",
			args:          []string{"--synthesis-model", "gemini-3-flash"},
			expectedModel: "gemini-3-flash",
			expectError:   false,
		},
		{
			name:          "No synthesis model specified",
			args:          []string{"--instructions=test.txt"},
			expectedModel: "", // Default is empty string
			expectError:   false,
		},
		{
			name:          "Invalid flag format",
			args:          []string{"--synthesis-model="},
			expectedModel: "",
			expectError:   false, // Go's flag package allows empty values
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag set for each test case
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)

			// Parse with the test args
			cfg, err := ParseFlagsWithEnv(fs, tc.args, func(string) string { return "mock-value" })

			// Check error expectations
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}

			// If expecting an error, check error message
			if tc.expectError && err != nil && tc.errorSubstring != "" {
				if !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorSubstring, err)
				}
			}

			// Check the parsed synthesis model
			if !tc.expectError && cfg != nil {
				if cfg.SynthesisModel != tc.expectedModel {
					t.Errorf("Expected SynthesisModel: %s, got: %s", tc.expectedModel, cfg.SynthesisModel)
				}
			}
		})
	}
}
