// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"testing"

	"github.com/phrazzld/thinktank/internal/registry"
)

// TestGetEnvVarNameForProvider tests the getEnvVarNameForProvider helper function
func TestGetEnvVarNameForProvider(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		providerName   string
		modelConfig    *registry.ModelsConfig
		expectedEnvVar string
	}{
		// Config-based API key environment variable names
		{
			name:         "provider found in config",
			providerName: "custom_provider",
			modelConfig: &registry.ModelsConfig{
				APIKeySources: map[string]string{
					"custom_provider": "CUSTOM_ENV_VAR",
				},
			},
			expectedEnvVar: "CUSTOM_ENV_VAR",
		},
		{
			name:         "empty env var name in config",
			providerName: "custom_provider",
			modelConfig: &registry.ModelsConfig{
				APIKeySources: map[string]string{
					"custom_provider": "",
				},
			},
			expectedEnvVar: "CUSTOM_PROVIDER_API_KEY", // Should use fallback
		},
		{
			name:         "provider not found in config",
			providerName: "missing_provider",
			modelConfig: &registry.ModelsConfig{
				APIKeySources: map[string]string{
					"other_provider": "OTHER_ENV_VAR",
				},
			},
			expectedEnvVar: "MISSING_PROVIDER_API_KEY", // Should use fallback
		},
		{
			name:         "nil APIKeySources map",
			providerName: "openai",
			modelConfig: &registry.ModelsConfig{
				APIKeySources: nil,
			},
			expectedEnvVar: "OPENAI_API_KEY", // Should use fallback
		},

		// Fallback for known providers with nil config
		{
			name:           "nil config - openai provider",
			providerName:   "openai",
			modelConfig:    nil,
			expectedEnvVar: "OPENAI_API_KEY",
		},
		{
			name:           "nil config - gemini provider",
			providerName:   "gemini",
			modelConfig:    nil,
			expectedEnvVar: "GEMINI_API_KEY",
		},
		{
			name:           "nil config - openrouter provider",
			providerName:   "openrouter",
			modelConfig:    nil,
			expectedEnvVar: "OPENROUTER_API_KEY",
		},

		// Hard-coded defaults for known providers (with empty config)
		{
			name:           "empty config - openai provider",
			providerName:   "openai",
			modelConfig:    &registry.ModelsConfig{},
			expectedEnvVar: "OPENAI_API_KEY",
		},
		{
			name:           "empty config - gemini provider",
			providerName:   "gemini",
			modelConfig:    &registry.ModelsConfig{},
			expectedEnvVar: "GEMINI_API_KEY",
		},
		{
			name:           "empty config - openrouter provider",
			providerName:   "openrouter",
			modelConfig:    &registry.ModelsConfig{},
			expectedEnvVar: "OPENROUTER_API_KEY",
		},

		// Generated names for unknown providers
		{
			name:           "unknown provider - custom",
			providerName:   "custom",
			modelConfig:    nil,
			expectedEnvVar: "CUSTOM_API_KEY",
		},
		{
			name:           "unknown provider - mixed case",
			providerName:   "azureOpenAI",
			modelConfig:    nil,
			expectedEnvVar: "AZUREOPENAI_API_KEY",
		},
		{
			name:           "unknown provider - with underscore",
			providerName:   "custom_provider",
			modelConfig:    nil,
			expectedEnvVar: "CUSTOM_PROVIDER_API_KEY",
		},
		{
			name:           "empty provider name",
			providerName:   "",
			modelConfig:    nil,
			expectedEnvVar: "_API_KEY", // Edge case
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function being tested
			envVarName := getEnvVarNameForProvider(tc.providerName, tc.modelConfig)

			// Verify the result
			if envVarName != tc.expectedEnvVar {
				t.Errorf("Expected environment variable name '%s', got '%s'",
					tc.expectedEnvVar, envVarName)
			}
		})
	}
}
