// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
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
			args:            []string{"--instructions", tempFile.Name(), "--synthesis-model", "gpt-4.1"},
			expectSynthesis: "gpt-4.1",
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
	// Save the original function to restore later
	origGetManager := getRegistryManagerForValidation
	// Override with a function that returns nil to force pattern matching
	getRegistryManagerForValidation = func(logger logutil.LoggerInterface) interface{} {
		return nil
	}
	// Restore at the end
	defer func() {
		getRegistryManagerForValidation = origGetManager
	}()
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

	// Store and restore the global manager
	originalManager := registry.GetGlobalManager(nil)
	registry.SetGlobalManagerForTesting(nil) // Force pattern matching for this test
	defer registry.SetGlobalManagerForTesting(originalManager)

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
		{
			name: "Valid model pattern - gemini",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				SynthesisModel:   "gemini-1.0-pro", // Valid pattern
			},
			expectError: false,
		},
		{
			name: "Valid model pattern - gpt",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				SynthesisModel:   "gpt-4.1", // Valid pattern
			},
			expectError: false,
		},
		{
			name: "Valid model pattern - claude",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				SynthesisModel:   "claude-3", // Valid pattern
			},
			expectError: false,
		},
		{
			name: "Valid model pattern - openrouter",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				SynthesisModel:   "openrouter/llama-3", // Valid pattern
			},
			expectError: false,
		},
		{
			name: "Invalid synthesis model",
			config: &config.CliConfig{
				InstructionsFile: tempFile.Name(),
				Paths:            []string{"testfile"},
				APIKey:           "test-key",
				ModelNames:       []string{"model1"},
				SynthesisModel:   "invalid-model-name", // Invalid pattern
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
// without using a real registry
func TestInvalidSynthesisModelValidation(t *testing.T) {
	// Save the original function to restore later
	origGetManager := getRegistryManagerForValidation
	// Override with a function that returns nil to force pattern matching
	getRegistryManagerForValidation = func(logger logutil.LoggerInterface) interface{} {
		return nil
	}
	// Restore at the end
	defer func() {
		getRegistryManagerForValidation = origGetManager
	}()
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

	// Store and restore the global manager
	originalManager := registry.GetGlobalManager(nil)
	registry.SetGlobalManagerForTesting(nil) // Set to nil to force string pattern matching
	defer registry.SetGlobalManagerForTesting(originalManager)

	// Create a logger for the test
	logger := &errorTrackingLogger{}

	// Create a config with an invalid synthesis model
	invalidConfig := &config.CliConfig{
		InstructionsFile: tempFile.Name(),
		Paths:            []string{"testfile"},
		APIKey:           "test-key",
		ModelNames:       []string{"gemini-1.0-pro"},
		SynthesisModel:   "invalid-model-name", // Invalid pattern
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

	// Test with valid synthesis model pattern
	// Reset the logger
	logger = &errorTrackingLogger{}

	// Create a config with a valid synthesis model pattern
	validConfig := &config.CliConfig{
		InstructionsFile: tempFile.Name(),
		Paths:            []string{"testfile"},
		APIKey:           "test-key",
		ModelNames:       []string{"gemini-1.0-pro"},
		SynthesisModel:   "gpt-4.1", // Valid pattern
	}

	// Validate the inputs - this should pass the synthesis model check
	err = ValidateInputsWithEnv(validConfig, logger, mockGetenv)

	// When registry is nil, pattern-based validation should pass for a valid model name
	if err != nil && strings.Contains(err.Error(), "invalid synthesis model") {
		t.Errorf("Got unexpected synthesis model error for valid model pattern: %v", err)
	}
}
