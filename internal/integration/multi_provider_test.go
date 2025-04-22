// internal/integration/multi_provider_test.go
package integration

import (
	"context"
	"os"
	"strings"
	"testing"

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
	apiService := thinktank.NewRegistryAPIService(registryManager, testLogger)

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

// Helper to set or unset an environment variable
func setEnvOrUnset(key, value string) error {
	if value != "" {
		return os.Setenv(key, value)
	}
	return os.Unsetenv(key)
}
