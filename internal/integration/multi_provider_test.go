// internal/integration/multi_provider_test.go
package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// TestAPIKeyIsolation verifies that the API key isolation logic works correctly
// across different providers (Gemini, OpenAI, OpenRouter)
func TestAPIKeyIsolation(t *testing.T) {
	// Create a logger that we can inspect for API key info
	testLogger := logutil.NewTestLogger(t)

	// Create a new registry API service
	registryManager := registry.NewManager(testLogger)
	registryInstance := registryManager.GetRegistry()

	// Initialize the registry with test models and providers
	initializeTestRegistry(t, registryInstance)

	// Register provider implementations
	registerTestProviders(t, registryInstance)

	apiService := thinktank.NewRegistryAPIService(registryInstance, testLogger)

	// Save the original environment variables so we can restore them
	originalGeminiKey := os.Getenv("GEMINI_API_KEY")
	originalOpenAIKey := os.Getenv("OPENAI_API_KEY")
	originalOpenRouterKey := os.Getenv("OPENROUTER_API_KEY")

	// Clean up environment after the test
	defer func() {
		// Restore the original environment variables
		if err := setEnvOrUnset("GEMINI_API_KEY", originalGeminiKey); err != nil {
			t.Logf("Error restoring GEMINI_API_KEY: %v", err)
		}
		if err := setEnvOrUnset("OPENAI_API_KEY", originalOpenAIKey); err != nil {
			t.Logf("Error restoring OPENAI_API_KEY: %v", err)
		}
		if err := setEnvOrUnset("OPENROUTER_API_KEY", originalOpenRouterKey); err != nil {
			t.Logf("Error restoring OPENROUTER_API_KEY: %v", err)
		}
	}()

	// Setup test API keys with distinctive prefixes
	// This makes it easy to verify in logs which key was used
	geminiKey := "gemini-test-key"
	openaiKey := "openai-test-key"
	openrouterKey := "sk-or-test-key"

	// Test cases for API key isolation
	tests := []struct {
		name         string
		setupEnv     func()
		models       []string
		passedKey    string
		expectedKeys map[string]string
	}{
		{
			name: "Environment variables are prioritized",
			setupEnv: func() {
				// Set all environment variables
				if err := os.Setenv("GEMINI_API_KEY", geminiKey); err != nil {
					t.Fatalf("Failed to set GEMINI_API_KEY: %v", err)
				}
				if err := os.Setenv("OPENAI_API_KEY", openaiKey); err != nil {
					t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
				}
				if err := os.Setenv("OPENROUTER_API_KEY", openrouterKey); err != nil {
					t.Fatalf("Failed to set OPENROUTER_API_KEY: %v", err)
				}
			},
			models:    []string{"gemini-pro", "gpt-3.5-turbo"},
			passedKey: "incorrect-key", // This should be ignored
			expectedKeys: map[string]string{
				"gemini-pro":    geminiKey,
				"gpt-3.5-turbo": openaiKey,
			},
		},
		{
			name: "Fallback to passed key when env var is not set",
			setupEnv: func() {
				// Set only Gemini key
				if err := os.Setenv("GEMINI_API_KEY", geminiKey); err != nil {
					t.Fatalf("Failed to set GEMINI_API_KEY: %v", err)
				}
				// Unset other keys
				if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
					t.Fatalf("Failed to unset OPENAI_API_KEY: %v", err)
				}
				if err := os.Unsetenv("OPENROUTER_API_KEY"); err != nil {
					t.Fatalf("Failed to unset OPENROUTER_API_KEY: %v", err)
				}
			},
			models:    []string{"gemini-pro", "gpt-3.5-turbo"},
			passedKey: openaiKey, // Should be used for OpenAI only
			expectedKeys: map[string]string{
				"gemini-pro":    geminiKey, // From env var
				"gpt-3.5-turbo": openaiKey, // From passed key
			},
		},
	}

	// Run each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment for this test case
			tc.setupEnv()

			// Create context
			ctx := context.Background()

			// Create a client for each model and verify key isolation
			for _, modelName := range tc.models {
				// Clear the test logger
				testLogger.ClearTestLogs()

				// Initialize client with the model and passed key
				// We expect an error because we're using fake keys
				client, err := apiService.InitLLMClient(ctx, tc.passedKey, modelName, "")
				t.Logf("Expected error for %s: %v", modelName, err)

				// Check the logs to verify which API key was used
				logs := testLogger.GetTestLogs()

				// Verify the expected API key was logged
				expectedKey := tc.expectedKeys[modelName]
				expectedPrefix := expectedKey
				if len(expectedPrefix) > 5 {
					expectedPrefix = expectedPrefix[:5]
				}

				keyFound := false
				for _, logMsg := range logs {
					if strings.Contains(logMsg, expectedPrefix) {
						keyFound = true
						break
					}
				}

				if !keyFound {
					t.Errorf("Model %s did not use expected API key with prefix: %s",
						modelName, expectedPrefix)
					t.Logf("Expected prefix: %s", expectedPrefix)
					for i, log := range logs {
						t.Logf("Log[%d]: %s", i, log)
					}
				}

				// Clean up client if it was created
				if client != nil {
					if err := client.Close(); err != nil {
						t.Logf("Error closing client: %v", err)
					}
				}
			}
		})
	}
}

// Initialize registry with test models and providers
func initializeTestRegistry(t *testing.T, reg *registry.Registry) {
	// Create a models config with API key sources
	modelConfig := registry.ModelsConfig{
		APIKeySources: map[string]string{
			"gemini":     "GEMINI_API_KEY",
			"openai":     "OPENAI_API_KEY",
			"openrouter": "OPENROUTER_API_KEY",
		},
		Providers: []registry.ProviderDefinition{
			{
				Name:    "gemini",
				BaseURL: "",
			},
			{
				Name:    "openai",
				BaseURL: "",
			},
			{
				Name:    "openrouter",
				BaseURL: "",
			},
		},
		Models: []registry.ModelDefinition{
			{
				Name:       "gemini-pro",
				Provider:   "gemini",
				APIModelID: "gemini-pro",
				Parameters: map[string]registry.ParameterDefinition{
					"temperature": {
						Type:    "float",
						Default: 0.7,
						Min:     0.0,
						Max:     1.0,
					},
					"max_tokens": {
						Type:    "int",
						Default: 1024,
						Min:     1,
						Max:     8192,
					},
				},
			},
			{
				Name:       "gpt-3.5-turbo",
				Provider:   "openai",
				APIModelID: "gpt-3.5-turbo",
				Parameters: map[string]registry.ParameterDefinition{
					"temperature": {
						Type:    "float",
						Default: 0.7,
						Min:     0.0,
						Max:     1.0,
					},
					"max_tokens": {
						Type:    "int",
						Default: 1024,
						Min:     1,
						Max:     4096,
					},
				},
			},
		},
	}

	// Create a mock loader that returns our test config
	mockLoader := &mockConfigLoader{config: &modelConfig}

	// Load the config into the registry
	err := reg.LoadConfig(mockLoader)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
}

// Register test provider implementations with the registry
func registerTestProviders(t *testing.T, reg *registry.Registry) {
	// Register mock provider implementations
	if err := reg.RegisterProviderImplementation("gemini", &mockProvider{providerName: "gemini"}); err != nil {
		t.Fatalf("Failed to register Gemini provider: %v", err)
	}
	if err := reg.RegisterProviderImplementation("openai", &mockProvider{providerName: "openai"}); err != nil {
		t.Fatalf("Failed to register OpenAI provider: %v", err)
	}
	if err := reg.RegisterProviderImplementation("openrouter", &mockProvider{providerName: "openrouter"}); err != nil {
		t.Fatalf("Failed to register OpenRouter provider: %v", err)
	}
}

// mockConfigLoader is a mock implementation of registry.ConfigLoaderInterface
type mockConfigLoader struct {
	config *registry.ModelsConfig
}

func (m *mockConfigLoader) Load() (*registry.ModelsConfig, error) {
	return m.config, nil
}

// mockProvider is a mock implementation of providers.Provider
type mockProvider struct {
	providerName string
}

func (m *mockProvider) CreateClient(ctx context.Context, apiKey, modelId, apiEndpoint string) (llm.LLMClient, error) {
	// Return an error to simulate client creation failure
	// The test doesn't need the actual client, it just verifies the API key usage
	return nil, fmt.Errorf("mock provider %s error: key=%s", m.providerName, apiKey)
}

// Helper to set or unset an environment variable
func setEnvOrUnset(key, value string) error {
	if value != "" {
		return os.Setenv(key, value)
	}
	return os.Unsetenv(key)
}
