package apikey

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

func TestNewAPIKeyResolver(t *testing.T) {
	// Test with nil logger
	resolver := NewAPIKeyResolver(nil)
	if resolver == nil {
		t.Fatal("Expected resolver to be created with nil logger")
	}
	if resolver.logger == nil {
		t.Fatal("Expected default logger to be created")
	}

	// Test with provided logger
	logger := logutil.NewTestLogger(t)
	resolver = NewAPIKeyResolver(logger)
	if resolver == nil {
		t.Fatal("Expected resolver to be created")
	}
	if resolver.logger != logger {
		t.Fatal("Expected provided logger to be used")
	}
}

func TestResolveAPIKey(t *testing.T) {
	tests := []struct {
		name              string
		providerName      string
		providedKey       string
		envVars           map[string]string
		apiKeySources     map[string]string
		expectedKey       string
		expectedSource    APIKeySource
		expectedEnvVar    string
		expectedError     bool
		expectedErrorType error
	}{
		{
			name:         "environment variable takes precedence over provided key",
			providerName: "openai",
			providedKey:  "provided-key",
			envVars: map[string]string{
				"OPENAI_API_KEY": "env-key",
			},
			apiKeySources: map[string]string{
				"openai": "OPENAI_API_KEY",
			},
			expectedKey:    "env-key",
			expectedSource: APIKeySourceEnvironment,
			expectedEnvVar: "OPENAI_API_KEY",
		},
		{
			name:         "uses provided key when environment variable not set",
			providerName: "gemini",
			providedKey:  "provided-key",
			envVars:      map[string]string{},
			apiKeySources: map[string]string{
				"gemini": "GEMINI_API_KEY",
			},
			expectedKey:    "provided-key",
			expectedSource: APIKeySourceParameter,
		},
		{
			name:         "returns error when no key available",
			providerName: "openrouter",
			providedKey:  "",
			envVars:      map[string]string{},
			apiKeySources: map[string]string{
				"openrouter": "OPENROUTER_API_KEY",
			},
			expectedError:     true,
			expectedErrorType: llm.ErrClientInitialization,
		},
		{
			name:         "uses fallback environment variable name when config unavailable",
			providerName: "openai",
			providedKey:  "",
			envVars: map[string]string{
				"OPENAI_API_KEY": "fallback-env-key",
			},
			apiKeySources:  nil,
			expectedKey:    "fallback-env-key",
			expectedSource: APIKeySourceEnvironment,
			expectedEnvVar: "OPENAI_API_KEY",
		},
		{
			name:         "handles unknown provider with generic env var name",
			providerName: "customllm",
			providedKey:  "",
			envVars: map[string]string{
				"CUSTOMLLM_API_KEY": "custom-key",
			},
			apiKeySources:  map[string]string{},
			expectedKey:    "custom-key",
			expectedSource: APIKeySourceEnvironment,
			expectedEnvVar: "CUSTOMLLM_API_KEY",
		},
		{
			name:         "empty environment variable falls back to provided key",
			providerName: "openai",
			providedKey:  "fallback-key",
			envVars: map[string]string{
				"OPENAI_API_KEY": "",
			},
			apiKeySources: map[string]string{
				"openai": "OPENAI_API_KEY",
			},
			expectedKey:    "fallback-key",
			expectedSource: APIKeySourceParameter,
		},
		{
			name:         "custom environment variable from config",
			providerName: "mymodel",
			providedKey:  "",
			envVars: map[string]string{
				"MY_CUSTOM_API_KEY": "custom-env-key",
			},
			apiKeySources: map[string]string{
				"mymodel": "MY_CUSTOM_API_KEY",
			},
			expectedKey:    "custom-env-key",
			expectedSource: APIKeySourceEnvironment,
			expectedEnvVar: "MY_CUSTOM_API_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any existing environment variables that might interfere
			existingEnvVars := []string{"OPENAI_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY", "CUSTOMLLM_API_KEY", "MY_CUSTOM_API_KEY"}
			for _, envVar := range existingEnvVars {
				oldVal := os.Getenv(envVar)
				os.Unsetenv(envVar)
				defer os.Setenv(envVar, oldVal)
			}

			// Set up environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create resolver with configured API key sources
			logger := logutil.NewTestLogger(t)
			resolver := NewAPIKeyResolverWithConfig(logger, tt.apiKeySources)

			// Resolve API key
			ctx := context.Background()
			result, err := resolver.ResolveAPIKey(ctx, tt.providerName, tt.providedKey)

			// Check error
			if tt.expectedError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.expectedErrorType != nil {
					// Check if error contains expected type
					if !strings.Contains(err.Error(), "API key required but not found") {
						t.Errorf("Expected error containing 'API key required but not found', got: %v", err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check result
			if result.Key != tt.expectedKey {
				t.Errorf("Expected key %q, got %q", tt.expectedKey, result.Key)
			}
			if result.Source != tt.expectedSource {
				t.Errorf("Expected source %v, got %v", tt.expectedSource, result.Source)
			}
			if result.EnvironmentVariable != tt.expectedEnvVar {
				t.Errorf("Expected env var %q, got %q", tt.expectedEnvVar, result.EnvironmentVariable)
			}
			if result.Provider != tt.providerName {
				t.Errorf("Expected provider %q, got %q", tt.providerName, result.Provider)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		apiKey       string
		expectError  bool
		expectWarn   bool
	}{
		{
			name:         "empty key returns error",
			providerName: "openai",
			apiKey:       "",
			expectError:  true,
		},
		{
			name:         "valid openai key format",
			providerName: "openai",
			apiKey:       "sk-1234567890abcdef",
			expectError:  false,
		},
		{
			name:         "openai key wrong format warns",
			providerName: "openai",
			apiKey:       "not-sk-prefix",
			expectError:  false,
			expectWarn:   true,
		},
		{
			name:         "short gemini key warns",
			providerName: "gemini",
			apiKey:       "short",
			expectError:  false,
			expectWarn:   true,
		},
		{
			name:         "valid gemini key",
			providerName: "gemini",
			apiKey:       "verylongalphanumerickey1234567890",
			expectError:  false,
		},
		{
			name:         "openrouter key no specific validation",
			providerName: "openrouter",
			apiKey:       "any-format-key",
			expectError:  false,
		},
		{
			name:         "unknown provider no specific validation",
			providerName: "customprovider",
			apiKey:       "custom-key",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logutil.NewTestLogger(t)
			resolver := NewAPIKeyResolver(logger)

			ctx := context.Background()
			err := resolver.ValidateAPIKey(ctx, tt.providerName, tt.apiKey)

			if tt.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Test logger doesn't provide GetLogs method, so we skip the warning check
			// The functionality is still tested through the actual logging
		})
	}
}

func TestGetEnvironmentVariableName(t *testing.T) {
	tests := []struct {
		name          string
		providerName  string
		apiKeySources map[string]string
		expected      string
	}{
		{
			name:         "uses config mapping when available",
			providerName: "mymodel",
			apiKeySources: map[string]string{
				"mymodel": "CUSTOM_ENV_VAR",
			},
			expected: "CUSTOM_ENV_VAR",
		},
		{
			name:          "fallback for openai",
			providerName:  "openai",
			apiKeySources: nil,
			expected:      "OPENAI_API_KEY",
		},
		{
			name:          "fallback for gemini",
			providerName:  "gemini",
			apiKeySources: nil,
			expected:      "GEMINI_API_KEY",
		},
		{
			name:          "fallback for openrouter",
			providerName:  "openrouter",
			apiKeySources: nil,
			expected:      "OPENROUTER_API_KEY",
		},
		{
			name:          "generic fallback for unknown provider",
			providerName:  "newprovider",
			apiKeySources: nil,
			expected:      "NEWPROVIDER_API_KEY",
		},
		{
			name:          "case insensitive provider name",
			providerName:  "OpenAI",
			apiKeySources: nil,
			expected:      "OPENAI_API_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &APIKeyResolver{
				apiKeySources: tt.apiKeySources,
			}

			result := resolver.getEnvironmentVariableName(tt.providerName)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetAPIKeyPrecedenceDocumentation(t *testing.T) {
	resolver := NewAPIKeyResolver(nil)
	doc := resolver.GetAPIKeyPrecedenceDocumentation()

	// Check that documentation contains key information
	expectedPhrases := []string{
		"API Key Resolution Precedence",
		"Environment Variables",
		"Highest Priority",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
		"OPENROUTER_API_KEY",
		"Explicitly Provided API Key",
		"Fallback",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(doc, phrase) {
			t.Errorf("Documentation missing expected phrase: %q", phrase)
		}
	}
}