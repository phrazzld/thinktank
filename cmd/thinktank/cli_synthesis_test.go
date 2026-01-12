// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/models"
)

// TestSynthesisModelParsing tests that the synthesis model flag is correctly
// parsed from command line arguments
func TestSynthesisModelParsing(t *testing.T) {
	t.Parallel(
	// Create a test instructions file
	)

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
			args:            []string{"--instructions", tempFile.Name(), "--synthesis-model", "gpt-5.2"},
			expectSynthesis: "gpt-5.2",
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

	// Get actual supported models for testing
	supportedModels := models.ListAllModels()

	// Configure test cases using actual supported models
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
				ModelNames:       []string{supportedModels[0]}, // Use actual supported model
				SynthesisModel:   "",                           // Empty - should pass validation
			},
			expectError: false,
		},
		{
			name: "Valid supported model - gemini",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{supportedModels[0]},
				SynthesisModel:   "gemini-3-flash", // Actual supported model
			},
			expectError: false,
		},
		{
			name: "Valid supported model - gpt",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{supportedModels[0]},
				SynthesisModel:   "gpt-5.2", // Actual supported model
			},
			expectError: false,
		},
		{
			name: "Valid supported openrouter model",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{supportedModels[0]},
				SynthesisModel:   "deepseek-v3.2", // Actual supported model
			},
			expectError: false,
		},
		{
			name: "Invalid unsupported model",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{supportedModels[0]},
				SynthesisModel:   "claude-3", // Not in our supported models list
			},
			expectError:   true,
			errorContains: "invalid synthesis model",
		},
		{
			name: "Another invalid synthesis model",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{supportedModels[0]},
				SynthesisModel:   "invalid-model-name", // Invalid model
			},
			expectError:   true,
			errorContains: "invalid synthesis model",
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

// TestInvalidSynthesisModelValidation tests the validation behavior for invalid synthesis models
// using the models package for direct validation
func TestInvalidSynthesisModelValidation(t *testing.T) {
	t.Parallel(
	// Create a test instructions file
	)

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

	// Create a logger for the test
	logger := &errorTrackingLogger{}

	// Get a supported model for the test
	supportedModels := models.ListAllModels()

	// Create a config with an invalid synthesis model
	invalidConfig := &config.CliConfig{
		InstructionsFile: tempFile.Name(),
		Paths:            []string{"testfile"},
		APIKey:           "test-key",
		ModelNames:       []string{supportedModels[0]}, // Use actual supported model
		SynthesisModel:   "invalid-model-name",         // Invalid model
	}

	// Mock environment function that returns test keys
	mockGetenv := func(key string) string {
		return "test-api-key" // Return test key for any env var
	}

	// Validate the inputs - this should fail due to invalid synthesis model
	err = ValidateInputsWithEnv(invalidConfig, logger, mockGetenv)

	// Check for expected error
	if err == nil {
		t.Error("Expected error for invalid synthesis model, but got nil")
	} else if !strings.Contains(err.Error(), "invalid synthesis model") {
		t.Errorf("Expected error to contain 'invalid synthesis model', but got: %v", err)
	}

	// Verify that the error was logged
	if !logger.errorCalled {
		t.Error("Expected error to be logged, but no error was logged")
	}

	// Test with valid synthesis model
	// Reset the logger
	logger = &errorTrackingLogger{}

	// Create a config with a valid synthesis model
	validConfig := &config.CliConfig{
		InstructionsFile: tempFile.Name(),
		Paths:            []string{"testfile"},
		APIKey:           "test-key",
		ModelNames:       []string{supportedModels[0]},
		SynthesisModel:   "gpt-5.2", // Valid supported model
	}

	// Validate the inputs - this should pass the synthesis model check
	err = ValidateInputsWithEnv(validConfig, logger, mockGetenv)

	// Should not get synthesis model error for valid model
	if err != nil && strings.Contains(err.Error(), "invalid synthesis model") {
		t.Errorf("Got unexpected synthesis model error for valid model: %v", err)
	}
}
