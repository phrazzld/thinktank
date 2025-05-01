package gemini_test

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers/gemini"
)

func TestGeminiProviderSecretHandling(t *testing.T) {
	// Create a test API key with a pattern that should be detected if leaked
	testAPIKey := "AIzaSyTESTKEY1234567890abcdefghijklmnopq"

	// Create a logger with secret detection
	testLogger := logutil.NewBufferLogger()
	secretLogger := logutil.WithSecretDetection(testLogger)
	// Don't panic on detection, just record for the test
	secretLogger.SetFailOnSecretDetect(false)

	t.Run("Client creation should not leak API key", func(t *testing.T) {
		// Clear any previously detected secrets
		secretLogger.ClearDetectedSecrets()

		// Create a provider with the secret-detecting logger
		provider := gemini.NewProvider(secretLogger)

		// Create a client which will trigger logging
		client, err := provider.CreateClient(
			context.Background(),
			testAPIKey,   // The API key that should never appear in logs
			"gemini-pro", // Model name
			"",           // Default endpoint
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
		provider := gemini.NewProvider(secretLogger)

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
}
