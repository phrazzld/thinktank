package models

import (
	"os"
	"strings"
	"testing"
)

// TestObsoleteProvidersRemoved ensures that OpenAI and Gemini providers
// are no longer available after consolidation to OpenRouter.
//
// CONSOLIDATION VALIDATION: This test suite validates the complete removal
// of obsolete providers after the OpenRouter consolidation project.
//
// Purpose:
// - Ensures no models still reference 'openai' or 'gemini' providers
// - Validates that API key environment variable mapping is updated
// - Confirms all models now use the unified 'openrouter' provider
//
// This test was initially designed to FAIL during the consolidation process,
// driving the cleanup and ensuring no obsolete provider references remain.
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
		// DEBUG: Print environment variables at test runtime
		envVars := []string{"OPENROUTER_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "THINKTANK_ENABLE_TEST_MODELS"}
		for _, envVar := range envVars {
			value := os.Getenv(envVar)
			if value != "" {
				t.Logf("DEBUG: Environment variable %s = %s (length: %d)", envVar, value, len(value))
			} else {
				t.Logf("DEBUG: Environment variable %s is empty or not set", envVar)
			}
		}

		// DEBUG: Print all environment variables with relevant prefixes
		for _, env := range os.Environ() {
			if strings.Contains(env, "OPENROUTER") || strings.Contains(env, "OPENAI") || strings.Contains(env, "GEMINI") || strings.Contains(env, "THINKTANK") {
				t.Logf("DEBUG: Found relevant env var: %s", env)
			}
		}

		providers := GetAvailableProviders()
		t.Logf("DEBUG: GetAvailableProviders returned: %v", providers)
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
