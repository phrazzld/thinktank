// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
)

// TestSynthesisModelDirectValidation tests the synthesis model validation logic
// using the models package for direct model support validation
func TestSynthesisModelDirectValidation(t *testing.T) {
	// Get actual supported models for testing

	supportedModels := models.ListAllModels()

	// Create test cases with actual supported and unsupported models
	tests := []struct {
		name           string
		synthesisModel string
		expectValid    bool
	}{
		{
			name:           "Empty model name",
			synthesisModel: "",
			expectValid:    true,
		},
		{
			name:           "Supported GPT model",
			synthesisModel: "gpt-4.1",
			expectValid:    true,
		},
		{
			name:           "Supported Gemini model",
			synthesisModel: "gemini-2.5-pro",
			expectValid:    true,
		},
		{
			name:           "Supported OpenRouter model",
			synthesisModel: "openrouter/deepseek/deepseek-r1-0528",
			expectValid:    true,
		},
		{
			name:           "Unsupported model",
			synthesisModel: "invalid-model-name",
			expectValid:    false,
		},
		{
			name:           "Another unsupported model",
			synthesisModel: "foo-bar-baz",
			expectValid:    false,
		},
		{
			name:           "Unsupported but valid-looking model",
			synthesisModel: "gpt-99-turbo",
			expectValid:    false,
		},
	}

	// Verify our test expectations match the models package
	for _, tt := range tests {
		if tt.synthesisModel != "" {
			actualSupport := models.IsModelSupported(tt.synthesisModel)
			if actualSupport != tt.expectValid {
				t.Errorf("Test expectation mismatch for model '%s': expected %v, models package says %v",
					tt.synthesisModel, tt.expectValid, actualSupport)
			}
		}
	}

	// Test the actual CLI validation function
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "")

	for _, tt := range tests {
		t.Run("CLI_"+tt.name, func(t *testing.T) {
			// Create a config with a valid model for primary ModelNames
			cfg := &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"testfile"},
				ModelNames:       []string{supportedModels[0]}, // Use first supported model
				SynthesisModel:   tt.synthesisModel,
			}

			// Override getenv to return valid API keys for all providers
			mockGetenv := func(key string) string {
				switch key {
				case "GEMINI_API_KEY":
					return "test-gemini-key"
				case "OPENAI_API_KEY":
					return "test-openai-key"
				case "OPENROUTER_API_KEY":
					return "test-openrouter-key"
				default:
					return "test-key"
				}
			}

			// Call the validation function
			err := ValidateInputsWithEnv(cfg, logger, mockGetenv)

			// Check validation results
			if tt.expectValid {
				if err != nil && containsSynthesisModelError(err.Error()) {
					t.Errorf("Expected valid synthesis model '%s', but got error: %v",
						tt.synthesisModel, err)
				}
			} else {
				// If we expect the model to be invalid, we should get a synthesis model error
				if err == nil || !containsSynthesisModelError(err.Error()) {
					t.Errorf("Expected synthesis model error for '%s', but got: %v",
						tt.synthesisModel, err)
				}
			}
		})
	}
}

// Helper to check if error message contains a synthesis model error
func containsSynthesisModelError(errMsg string) bool {
	return strings.Contains(errMsg, "invalid synthesis model")
}
