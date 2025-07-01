package models

import (
	"testing"
)

// TestObsoleteProvidersRemoved ensures that OpenAI and Gemini providers
// are no longer available after consolidation to OpenRouter.
// This test should initially FAIL, driving the cleanup process.
func TestObsoleteProvidersRemoved(t *testing.T) {
	t.Parallel()

	t.Run("No models should use openai provider", func(t *testing.T) {
		allModels := ListAllModels()
		for _, modelName := range allModels {
			model, err := GetModelInfo(modelName)
			if err != nil {
				t.Fatalf("Failed to get model info for %s: %v", modelName, err)
			}
			if model.Provider == "openai" {
				t.Errorf("Model %s still uses 'openai' provider - should be 'openrouter'", modelName)
			}
		}
	})

	t.Run("No models should use gemini provider", func(t *testing.T) {
		allModels := ListAllModels()
		for _, modelName := range allModels {
			model, err := GetModelInfo(modelName)
			if err != nil {
				t.Fatalf("Failed to get model info for %s: %v", modelName, err)
			}
			if model.Provider == "gemini" {
				t.Errorf("Model %s still uses 'gemini' provider - should be 'openrouter'", modelName)
			}
		}
	})

	t.Run("GetAPIKeyEnvVar should not return obsolete provider env vars", func(t *testing.T) {
		// This will initially fail - forcing us to clean up the function
		envVar := GetAPIKeyEnvVar("openai")
		if envVar == "OPENAI_API_KEY" {
			t.Error("GetAPIKeyEnvVar should not return OPENAI_API_KEY for obsolete openai provider")
		}

		envVar = GetAPIKeyEnvVar("gemini")
		if envVar == "GEMINI_API_KEY" {
			t.Error("GetAPIKeyEnvVar should not return GEMINI_API_KEY for obsolete gemini provider")
		}
	})

	t.Run("GetAvailableProviders should not include obsolete providers", func(t *testing.T) {
		providers := GetAvailableProviders()
		for _, provider := range providers {
			if provider == "openai" {
				t.Error("GetAvailableProviders should not include 'openai' - models migrated to OpenRouter")
			}
			if provider == "gemini" {
				t.Error("GetAvailableProviders should not include 'gemini' - models migrated to OpenRouter")
			}
		}
	})

	t.Run("Only openrouter provider should be available", func(t *testing.T) {
		providers := GetAvailableProviders()

		// Should have at least openrouter and test provider
		foundOpenRouter := false
		for _, provider := range providers {
			if provider == "openrouter" {
				foundOpenRouter = true
			}
		}

		if !foundOpenRouter {
			t.Error("GetAvailableProviders should include 'openrouter'")
		}
	})
}
