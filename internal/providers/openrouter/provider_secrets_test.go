package openrouter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/providers/openrouter"
)

func TestOpenRouterProviderSecretHandling(t *testing.T) {
	// Create a test API key with a pattern that should be detected if leaked
	testAPIKey := "sk-or-test-1234567890abcdefghijklmnopqrstuvwxyz1234567890"

	// Create a logger with secret detection
	testLogger := logutil.NewBufferLogger(logutil.DebugLevel)
	secretLogger := logutil.WithSecretDetection(testLogger)
	// Don't panic on detection, just record for the test
	secretLogger.SetFailOnSecretDetect(false)

	t.Run("Client creation should not leak API key", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := openrouter.NewProvider(secretLogger)

		// Create a client which will trigger logging
		client, err := provider.CreateClient(
			context.Background(),
			testAPIKey,                // The API key that should never appear in logs
			"anthropic/claude-3-opus", // Model name
			"",                        // Default endpoint
		)

		// Whether client creation succeeds or fails doesn't matter
		// What matters is that the API key is not logged
		_ = err // Ignore error - we're only testing for leaked secrets
		if client != nil {
			_ = client.Close() // Clean up if client was created
		}

		// Check if any secrets were detected
		if secretLogger.HasDetectedSecrets() {
			t.Errorf("API key leaked in logs during client creation:\n%v",
				secretLogger.GetDetectedSecrets())
		}
	})

	t.Run("Error handling should not leak API key", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := openrouter.NewProvider(secretLogger)

		// Create a client with an intentionally invalid model name to trigger an error
		client, err := provider.CreateClient(
			context.Background(),
			testAPIKey,           // The API key that should never appear in logs
			"invalid-model-name", // Invalid model to trigger an error
			"",                   // Default endpoint
		)

		// There should be an error
		if err == nil {
			t.Skip("Expected an error with invalid model, but got none")
		}
		if client != nil {
			_ = client.Close() // Clean up if client was created
		}

		// Check if any secrets were detected
		if secretLogger.HasDetectedSecrets() {
			t.Errorf("API key leaked in logs during error handling:\n%v",
				secretLogger.GetDetectedSecrets())
		}
	})

	// New test for API key prefixes not being logged
	t.Run("API key prefix should not be logged", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a buffered logger to capture logs
		bufferLogger := logutil.NewBufferLogger(logutil.DebugLevel)

		// Create a secret-detecting logger that wraps the buffer logger
		logger := logutil.WithSecretDetection(bufferLogger)
		logger.SetFailOnSecretDetect(false)

		// Create a provider with the logger
		provider := openrouter.NewProvider(logger)

		// Create a client which will trigger key prefix logging
		_, _ = provider.CreateClient(
			context.Background(),
			testAPIKey,
			"anthropic/claude-3-opus",
			"",
		)

		// Check if key prefix was logged
		logs := bufferLogger.GetLogs()
		logOutput := strings.Join(logs, "\n")

		if logger.HasDetectedSecrets() {
			t.Errorf("API key prefix leaked in logs:\n%v", logger.GetDetectedSecrets())
		}

		// Additional check: make sure logs don't contain "starts with: sk-or"
		if strings.Contains(logOutput, testAPIKey[:5]) {
			t.Errorf("API key prefix found in logs: %s", testAPIKey[:5])
		}
	})

	// Test for environment variable key handling
	t.Run("Environment variable key should be handled securely", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := openrouter.NewProvider(secretLogger)

		// This will fall back to environment variable (which may not be set in test)
		// But we're testing the logging behavior, not the actual key lookup
		_, _ = provider.CreateClient(
			context.Background(),
			"", // Empty key to trigger env var fallback
			"anthropic/claude-3-opus",
			"",
		)

		// Check if any secrets were detected in the logging
		if secretLogger.HasDetectedSecrets() {
			t.Errorf("Environment variable handling leaked secrets:\n%v",
				secretLogger.GetDetectedSecrets())
		}
	})

	// Test API call execution with mock server
	t.Run("API call should not leak API key", func(t *testing.T) {
		// Set up a mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that the Authorization header has the expected format
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer "+testAPIKey {
				t.Errorf("Expected Authorization header to be 'Bearer %s', got %s", testAPIKey, authHeader)
			}

			// Return a minimal valid response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "test-id",
				"object": "chat.completion",
				"created": 1677858242,
				"model": "test-model",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "test response"
						},
						"finish_reason": "stop"
					}
				],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 20,
					"total_tokens": 30
				}
			}`))
		}))
		defer server.Close()

		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := openrouter.NewProvider(secretLogger)

		// Create a client pointing to our mock server
		client, err := provider.CreateClient(
			context.Background(),
			testAPIKey,
			"anthropic/claude-3-opus",
			server.URL,
		)

		if err != nil || client == nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Make an API call
		_, err = client.GenerateContent(context.Background(), "Hello", nil)
		if err != nil {
			t.Logf("API call error (expected in tests): %v", err)
		}

		// Check if any secrets were detected in the logs during the API call
		if secretLogger.HasDetectedSecrets() {
			t.Errorf("API key leaked in logs during API call:\n%v",
				secretLogger.GetDetectedSecrets())
		}
	})

	// Test edge cases with different key formats
	t.Run("Edge case: Malformed API key should be handled securely", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := openrouter.NewProvider(secretLogger)

		// Try with a malformed key (missing proper prefix)
		malformedKey := "not-a-proper-key-12345"

		// This should fail, but shouldn't leak the key
		_, _ = provider.CreateClient(
			context.Background(),
			malformedKey,
			"anthropic/claude-3-opus",
			"",
		)

		// Check if any secrets were detected
		if secretLogger.HasDetectedSecrets() {
			t.Errorf("Malformed API key leaked in logs:\n%v",
				secretLogger.GetDetectedSecrets())
		}
	})

	// Test edge case with empty key
	t.Run("Edge case: Empty API key should be handled securely", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := openrouter.NewProvider(secretLogger)

		// Create client with empty key - should try to use env var
		_, _ = provider.CreateClient(
			context.Background(),
			"", // Empty key
			"anthropic/claude-3-opus",
			"",
		)

		// No secrets to detect with empty key, but checking anyway
		if secretLogger.HasDetectedSecrets() {
			t.Errorf("Empty API key handling leaked secrets:\n%v",
				secretLogger.GetDetectedSecrets())
		}
	})
}
