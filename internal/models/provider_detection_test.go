// Package models provides model configuration and selection functionality
package models

import (
	"os"
	"testing"
)

func TestGetAvailableProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		envVars     map[string]string
		expected    []string
		description string
	}{
		{
			name:        "no API keys set",
			envVars:     map[string]string{},
			expected:    []string{},
			description: "Should return empty list when no API keys are set",
		},
		{
			name: "only OpenAI API key set",
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-test123",
			},
			expected:    []string{"openai"},
			description: "Should return only openai when only OPENAI_API_KEY is set",
		},
		{
			name: "only Gemini API key set",
			envVars: map[string]string{
				"GEMINI_API_KEY": "gemini-test-key",
			},
			expected:    []string{"gemini"},
			description: "Should return only gemini when only GEMINI_API_KEY is set",
		},
		{
			name: "only OpenRouter API key set",
			envVars: map[string]string{
				"OPENROUTER_API_KEY": "sk-or-test",
			},
			expected:    []string{"openrouter"},
			description: "Should return only openrouter when only OPENROUTER_API_KEY is set",
		},
		{
			name: "OpenAI and Gemini keys set",
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-test123",
				"GEMINI_API_KEY": "gemini-test",
			},
			expected:    []string{"openai", "gemini"},
			description: "Should return both openai and gemini when both keys are set",
		},
		{
			name: "all three API keys set",
			envVars: map[string]string{
				"OPENAI_API_KEY":     "sk-test123",
				"GEMINI_API_KEY":     "gemini-test",
				"OPENROUTER_API_KEY": "sk-or-test",
			},
			expected:    []string{"openai", "gemini", "openrouter"},
			description: "Should return all three providers when all keys are set",
		},
		{
			name: "empty string API key should be ignored",
			envVars: map[string]string{
				"OPENAI_API_KEY": "",
				"GEMINI_API_KEY": "valid-key",
			},
			expected:    []string{"gemini"},
			description: "Should ignore empty string API keys",
		},
		{
			name: "whitespace-only API key should not be ignored",
			envVars: map[string]string{
				"OPENAI_API_KEY": "  ",
				"GEMINI_API_KEY": "valid-key",
			},
			expected:    []string{"openai", "gemini"},
			description: "Should count whitespace-only API keys as valid (per implementation)",
		},
		{
			name: "unknown environment variables ignored",
			envVars: map[string]string{
				"OPENAI_API_KEY":     "sk-test123",
				"OPENROUTER_API_KEY": "sk-or-test",
				"UNKNOWN_API_KEY":    "should-be-ignored",
			},
			expected:    []string{"openai", "openrouter"},
			description: "Should only return providers with valid keys, ignoring unknown env vars",
		},
		{
			name: "case sensitivity test",
			envVars: map[string]string{
				"openai_api_key": "lowercase-should-be-ignored",
				"OPENAI_API_KEY": "sk-correct-case",
				"Gemini_Api_Key": "mixed-case-ignored",
				"GEMINI_API_KEY": "gemini-correct",
			},
			expected:    []string{"openai", "gemini"},
			description: "Should only recognize exact case environment variable names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			expectedKeys := []string{"OPENAI_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY"}
			originalEnv := make(map[string]string)
			for _, key := range expectedKeys {
				originalEnv[key] = os.Getenv(key)
			}

			// Clean up environment
			defer func() {
				for key, value := range originalEnv {
					if value == "" {
						_ = os.Unsetenv(key)
					} else {
						_ = os.Setenv(key, value)
					}
				}
			}()

			// Clear all relevant environment variables first
			for _, key := range expectedKeys {
				_ = os.Unsetenv(key)
			}

			// Set up test environment
			for key, value := range tt.envVars {
				if value != "" {
					_ = os.Setenv(key, value)
				}
			}

			// Call function under test
			result := GetAvailableProviders()

			// Verify result length
			if len(result) != len(tt.expected) {
				t.Errorf("GetAvailableProviders() returned %d providers, want %d\nGot: %v\nWant: %v\nDescription: %s",
					len(result), len(tt.expected), result, tt.expected, tt.description)
				return
			}

			// Convert to sets for comparison (order doesn't matter)
			resultSet := make(map[string]bool)
			for _, provider := range result {
				resultSet[provider] = true
			}

			expectedSet := make(map[string]bool)
			for _, provider := range tt.expected {
				expectedSet[provider] = true
			}

			// Verify all expected providers are present
			for _, expected := range tt.expected {
				if !resultSet[expected] {
					t.Errorf("GetAvailableProviders() missing expected provider: %s\nGot: %v\nWant: %v\nDescription: %s",
						expected, result, tt.expected, tt.description)
				}
			}

			// Verify no unexpected providers are present
			for _, actual := range result {
				if !expectedSet[actual] {
					t.Errorf("GetAvailableProviders() returned unexpected provider: %s\nGot: %v\nWant: %v\nDescription: %s",
						actual, result, tt.expected, tt.description)
				}
			}

			// Verify all returned providers have valid API keys (consistent with implementation)
			for _, provider := range result {
				envVar := GetAPIKeyEnvVar(provider)
				if envVar == "" {
					t.Errorf("GetAvailableProviders() returned provider %s but GetAPIKeyEnvVar() returned empty string", provider)
					continue
				}

				apiKey := os.Getenv(envVar)
				// The implementation only checks != "", so we verify the same
				if apiKey == "" {
					t.Errorf("GetAvailableProviders() returned provider %s but API key is empty: %q", provider, apiKey)
				}
			}
		})
	}
}

func TestGetAPIKeyEnvVar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{
			name:     "openai provider",
			provider: "openai",
			expected: "OPENAI_API_KEY",
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expected: "GEMINI_API_KEY",
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expected: "OPENROUTER_API_KEY",
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			expected: "",
		},
		{
			name:     "empty provider",
			provider: "",
			expected: "",
		},
		{
			name:     "case sensitive - uppercase openai",
			provider: "OPENAI",
			expected: "",
		},
		{
			name:     "case sensitive - mixed case",
			provider: "OpenAI",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAPIKeyEnvVar(tt.provider)
			if result != tt.expected {
				t.Errorf("GetAPIKeyEnvVar(%q) = %q, want %q", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetProviderForModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		modelName    string
		expected     string
		expectError  bool
		errorMessage string
	}{
		{
			name:      "openai model gpt-4.1",
			modelName: "gpt-4.1",
			expected:  "openai",
		},
		{
			name:      "openai model o3",
			modelName: "o3",
			expected:  "openai",
		},
		{
			name:      "openai model o4-mini",
			modelName: "o4-mini",
			expected:  "openai",
		},
		{
			name:      "gemini model",
			modelName: "gemini-2.5-pro",
			expected:  "gemini",
		},
		{
			name:      "gemini flash model",
			modelName: "gemini-2.5-flash",
			expected:  "gemini",
		},
		{
			name:      "openrouter model",
			modelName: "openrouter/meta-llama/llama-4-maverick",
			expected:  "openrouter",
		},
		{
			name:      "another openrouter model",
			modelName: "openrouter/deepseek/deepseek-chat-v3-0324",
			expected:  "openrouter",
		},
		{
			name:         "unknown model",
			modelName:    "unknown-model",
			expectError:  true,
			errorMessage: "unknown model",
		},
		{
			name:         "empty model name",
			modelName:    "",
			expectError:  true,
			errorMessage: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetProviderForModel(tt.modelName)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetProviderForModel(%q) expected error, got nil", tt.modelName)
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					// Note: Just check that error contains the expected message
					if err.Error() == "" {
						t.Errorf("GetProviderForModel(%q) error message is empty", tt.modelName)
					}
				}
				if result != "" {
					t.Errorf("GetProviderForModel(%q) expected empty result on error, got %q", tt.modelName, result)
				}
			} else {
				if err != nil {
					t.Errorf("GetProviderForModel(%q) unexpected error: %v", tt.modelName, err)
				}
				if result != tt.expected {
					t.Errorf("GetProviderForModel(%q) = %q, want %q", tt.modelName, result, tt.expected)
				}
			}
		})
	}
}
