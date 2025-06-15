// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestSynthesisModelPatternValidation tests the pattern validation logic
// for synthesis models when registry is not available
func TestSynthesisModelPatternValidation(t *testing.T) {
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
	// Create configurations with various synthesis model names
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
			name:           "GPT model pattern",
			synthesisModel: "gpt-4.1",
			expectValid:    true,
		},
		{
			name:           "Gemini model pattern",
			synthesisModel: "gemini-1.0-pro",
			expectValid:    true,
		},
		{
			name:           "Claude model pattern",
			synthesisModel: "claude-3",
			expectValid:    true,
		},
		{
			name:           "OpenRouter model pattern",
			synthesisModel: "openrouter/anthropic/claude-3",
			expectValid:    true,
		},
		{
			name:           "Text model pattern",
			synthesisModel: "text-embedding-ada-002",
			expectValid:    true,
		},
		{
			name:           "Invalid model pattern",
			synthesisModel: "invalid-model-name",
			expectValid:    false,
		},
		{
			name:           "Another invalid pattern",
			synthesisModel: "foo-bar-baz",
			expectValid:    false,
		},
	}

	// Helper function for case insensitive prefix check
	hasPrefix := func(name, prefix string) bool {
		if len(name) < len(prefix) {
			return false
		}
		for i := 0; i < len(prefix); i++ {
			if (name[i] | 32) != (prefix[i] | 32) { // case insensitive compare
				return false
			}
		}
		return true
	}

	// Create a simplified validation helper that only checks the pattern
	// This mimics the pattern-matching fallback in ValidateInputsWithEnv
	validateModelPattern := func(modelName string) bool {
		if modelName == "" {
			return true
		}

		// Basic model validation based on naming patterns
		return hasPrefix(modelName, "gpt-") ||
			hasPrefix(modelName, "text-") ||
			hasPrefix(modelName, "gemini-") ||
			hasPrefix(modelName, "claude-") ||
			hasPrefix(modelName, "openrouter/")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateModelPattern(tt.synthesisModel)
			if result != tt.expectValid {
				t.Errorf("Expected validation result %v for model '%s', but got %v",
					tt.expectValid, tt.synthesisModel, result)
			}
		})
	}

	// Now test that our actual code in cli.go has the same behavior
	// This ensures the pattern matching in the production code matches our expectations
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "")

	for _, tt := range tests {
		t.Run("CLI_"+tt.name, func(t *testing.T) {
			// Create a config with just the synthesis model
			cfg := &config.CliConfig{
				InstructionsFile: "test.txt",
				Paths:            []string{"testfile"},
				ModelNames:       []string{"test-model"},
				SynthesisModel:   tt.synthesisModel,
			}

			// Override getenv to return a valid key
			mockGetenv := func(string) string {
				return "test-key"
			}

			// Mock the registry as nil to force pattern matching
			// Store the original regManager and restore it after
			origGetManager := getRegistryManagerForValidation
			defer func() {
				getRegistryManagerForValidation = origGetManager
			}()

			// Override the registry manager getter to return nil
			getRegistryManagerForValidation = func(logger logutil.LoggerInterface) interface{} {
				return nil
			}

			// Call the validation function
			err := ValidateInputsWithEnv(cfg, logger, mockGetenv)

			// If we expect the model to be valid, there shouldn't be a synthesis model error
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
