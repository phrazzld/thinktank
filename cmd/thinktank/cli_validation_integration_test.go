// Package main provides integration tests for CLI validation functionality
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
)

// TestValidateInputsIntegration tests the main ValidateInputs function with real environment variables
func TestValidateInputsIntegration(t *testing.T) {
	// Removed t.Parallel() - modifies environment variables
	// Save original environment variables

	originalGeminiKey := os.Getenv(apiKeyEnvVar)
	originalOpenAIKey := os.Getenv(openaiAPIKeyEnvVar)
	originalOpenRouterKey := os.Getenv("OPENROUTER_API_KEY")

	defer func() {
		// Restore original environment
		if originalGeminiKey != "" {
			_ = os.Setenv(apiKeyEnvVar, originalGeminiKey)
		} else {
			_ = os.Unsetenv(apiKeyEnvVar)
		}
		if originalOpenAIKey != "" {
			_ = os.Setenv(openaiAPIKeyEnvVar, originalOpenAIKey)
		} else {
			_ = os.Unsetenv(openaiAPIKeyEnvVar)
		}
		if originalOpenRouterKey != "" {
			_ = os.Setenv("OPENROUTER_API_KEY", originalOpenRouterKey)
		} else {
			_ = os.Unsetenv("OPENROUTER_API_KEY")
		}
	}()

	// Create a temporary instructions file
	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	if err := os.WriteFile(instructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}

	// Use buffer logger instead of test logger to avoid failing on expected error logs
	logger := logutil.NewBufferLogger(logutil.InfoLevel)

	// Get actual supported models for testing
	supportedModels := models.ListAllModels()
	var geminiModel, openAIModel, openRouterModel string
	for _, model := range supportedModels {
		provider, _ := models.GetProviderForModel(model)
		switch provider {
		case "gemini":
			if geminiModel == "" {
				geminiModel = model
			}
		case "openai":
			if openAIModel == "" {
				openAIModel = model
			}
		case "openrouter":
			if openRouterModel == "" {
				openRouterModel = model
			}
		}
	}

	tests := []struct {
		name          string
		config        *config.CliConfig
		envVars       map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid configuration with Gemini model",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{geminiModel},
			},
			envVars: map[string]string{
				apiKeyEnvVar: "test-gemini-api-key",
			},
			expectError: false,
		},
		{
			name: "Valid configuration with OpenAI model",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{openAIModel},
			},
			envVars: map[string]string{
				openaiAPIKeyEnvVar: "test-openai-api-key",
			},
			expectError: false,
		},
		{
			name: "Valid configuration with OpenRouter model",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{openRouterModel},
			},
			envVars: map[string]string{
				"OPENROUTER_API_KEY": "test-openrouter-api-key",
			},
			expectError: false,
		},
		{
			name: "Missing Gemini API key",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{geminiModel},
			},
			envVars:       map[string]string{}, // No API key set
			expectError:   true,
			errorContains: "gemini API key not set",
		},
		{
			name: "Missing OpenAI API key",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{openAIModel},
			},
			envVars:       map[string]string{}, // No API key set
			expectError:   true,
			errorContains: "openAI API key not set",
		},
		{
			name: "Missing OpenRouter API key",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{openRouterModel},
			},
			envVars:       map[string]string{}, // No API key set
			expectError:   true,
			errorContains: "openRouter API key not set",
		},
		{
			name: "Multiple models with mixed providers",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{openAIModel, geminiModel},
			},
			envVars: map[string]string{
				apiKeyEnvVar:       "test-gemini-api-key",
				openaiAPIKeyEnvVar: "test-openai-api-key",
			},
			expectError: false,
		},
		{
			name: "Dry run mode bypasses API key requirements",
			config: &config.CliConfig{
				InstructionsFile: "", // Not required for dry run
				Paths:            []string{"src/"},
				ModelNames:       []string{}, // Not required for dry run
				DryRun:           true,
			},
			envVars:     map[string]string{}, // No API keys needed
			expectError: false,
		},
		{
			name: "Missing instructions file (non-dry-run)",
			config: &config.CliConfig{
				InstructionsFile: "",
				Paths:            []string{"src/"},
				ModelNames:       []string{geminiModel},
				DryRun:           false,
			},
			envVars: map[string]string{
				apiKeyEnvVar: "test-gemini-api-key",
			},
			expectError:   true,
			errorContains: "missing required --instructions flag",
		},
		{
			name: "Missing paths",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{}, // No paths provided
				ModelNames:       []string{geminiModel},
				DryRun:           false,
			},
			envVars: map[string]string{
				apiKeyEnvVar: "test-gemini-api-key",
			},
			expectError:   true,
			errorContains: "no paths specified",
		},
		{
			name: "Missing models (non-dry-run)",
			config: &config.CliConfig{
				InstructionsFile: instructionsFile,
				Paths:            []string{"src/"},
				ModelNames:       []string{}, // No models specified
				DryRun:           false,
			},
			envVars: map[string]string{
				apiKeyEnvVar: "test-gemini-api-key",
			},
			expectError:   true,
			errorContains: "no models specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			_ = os.Unsetenv(apiKeyEnvVar)
			_ = os.Unsetenv(openaiAPIKeyEnvVar)
			_ = os.Unsetenv("OPENROUTER_API_KEY")

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Call the actual ValidateInputs function
			err := ValidateInputs(tt.config, logger)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			// Check for unexpected error
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestValidateInputsEdgeCases tests additional edge cases to improve ValidateInputsWithEnv coverage
func TestValidateInputsEdgeCases(t *testing.T) {
	// Removed t.Parallel() - uses environment variables
	// Create a temporary instructions file

	tempDir := t.TempDir()
	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	if err := os.WriteFile(instructionsFile, []byte("test instructions"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}

	// Use buffer logger instead of test logger to avoid failing on expected error logs
	logger := logutil.NewBufferLogger(logutil.InfoLevel)

	t.Run("Synthesis model with invalid model", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: instructionsFile,
			Paths:            []string{"src/"},
			ModelNames:       []string{"gemini-2.5-pro"},
			SynthesisModel:   "totally-invalid-model-name",
		}

		getenv := func(key string) string {
			if key == apiKeyEnvVar {
				return "test-gemini-key"
			}
			return ""
		}

		err := ValidateInputsWithEnv(config, logger, getenv)
		if err == nil {
			t.Error("Expected error for invalid synthesis model")
		}
		if !strings.Contains(err.Error(), "invalid synthesis model") {
			t.Errorf("Expected error to contain synthesis model validation message, got: %v", err)
		}
	})

	t.Run("Synthesis model with valid supported model", func(t *testing.T) {
		config := &config.CliConfig{
			InstructionsFile: instructionsFile,
			Paths:            []string{"src/"},
			ModelNames:       []string{"gemini-2.5-pro"},
			SynthesisModel:   "gpt-4.1", // Valid supported model
		}

		getenv := func(key string) string {
			if key == apiKeyEnvVar {
				return "test-gemini-key"
			}
			if key == openaiAPIKeyEnvVar {
				return "test-openai-key"
			}
			return ""
		}

		err := ValidateInputsWithEnv(config, logger, getenv)
		if err != nil {
			t.Errorf("Unexpected error for valid synthesis model: %v", err)
		}
	})
}
