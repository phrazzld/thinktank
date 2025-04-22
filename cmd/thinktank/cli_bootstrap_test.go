// Package architect provides the command-line interface for the thinktank tool
package thinktank

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
	// Skip this test for now as the ValidateInputs function doesn't
	// directly check for config file existence - it's handled at initialization
	t.Skip("Skipping test as ValidateInputs doesn't validate config path existence")
}
