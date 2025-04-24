// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
)

// TestSynthesisModelParsing tests that the synthesis model flag is correctly
// parsed from command line arguments
func TestSynthesisModelParsing(t *testing.T) {
	// Create a test instructions file
	tempFile, err := os.CreateTemp("", "instructions-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary instructions file: %v", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	_, err = tempFile.WriteString("Test instructions content")
	if err != nil {
		t.Fatalf("Failed to write to temporary instructions file: %v", err)
	}
	_ = tempFile.Close()

	tests := []struct {
		name            string
		args            []string
		expectSynthesis string
	}{
		{
			name:            "With synthesis model",
			args:            []string{"--instructions", tempFile.Name(), "--synthesis-model", "gpt-4"},
			expectSynthesis: "gpt-4",
		},
		{
			name:            "With synthesis model using equals sign",
			args:            []string{"--instructions", tempFile.Name(), "--synthesis-model=gemini-1.0-pro"},
			expectSynthesis: "gemini-1.0-pro",
		},
		{
			name:            "Without synthesis model",
			args:            []string{"--instructions", tempFile.Name()},
			expectSynthesis: "", // Default empty
		},
		{
			name:            "With empty synthesis model",
			args:            []string{"--instructions", tempFile.Name(), "--synthesis-model", ""},
			expectSynthesis: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Append a test path to the args (required for validation)
			args := append(tt.args, "testpath")

			// Create a flag set for testing
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard) // Suppress error output

			// Parse the flags using our exported function
			cfg, err := ParseFlagsWithEnv(fs, args, func(key string) string {
				return "test-key" // Mock environment function
			})

			// We should not get parsing errors
			if err != nil {
				t.Fatalf("ParseFlagsWithEnv() error = %v", err)
			}

			// Check that the synthesis model was correctly parsed
			if cfg.SynthesisModel != tt.expectSynthesis {
				t.Errorf("Expected SynthesisModel = %q, got %q", tt.expectSynthesis, cfg.SynthesisModel)
			}
		})
	}
}

// TestSynthesisModelValidation tests the validation logic for the synthesis model flag
func TestSynthesisModelValidation(t *testing.T) {
	// Create a test instructions file
	tempFile, err := os.CreateTemp("", "instructions-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary instructions file: %v", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	_, err = tempFile.WriteString("Test instructions content")
	if err != nil {
		t.Fatalf("Failed to write to temporary instructions file: %v", err)
	}
	_ = tempFile.Close()

	// Configure test cases, focusing on empty synthesis model which doesn't require registry mocking
	tests := []struct {
		name          string
		config        *config.CliConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "Empty synthesis model",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				SynthesisModel:   "", // Empty - should pass validation
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger that captures log calls
			logger := &errorTrackingLogger{}

			// Run the validation
			err := ValidateInputsWithEnv(tt.config, logger, func(key string) string {
				return "mock-value" // Return a valid value for any key
			})

			// Check if error matches expectation
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateInputs() error = %v, expectError %v", err, tt.expectError)
			}

			// Verify error contains expected text
			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error message %q doesn't contain expected text %q", err.Error(), tt.errorContains)
				}
			}

			// Verify logger recorded errors for error cases
			if tt.expectError && !logger.errorCalled {
				t.Error("Expected error to be logged, but no error was logged")
			}
		})
	}
}
