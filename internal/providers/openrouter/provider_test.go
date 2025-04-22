package openrouter

import (
	"context"
	"os"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
)

// TestNewClientWithEnvVars tests creating a client with environment variables
func TestNewClientWithEnvVars(t *testing.T) {
	// Save current env vars
	oldAPIKey := os.Getenv("OPENROUTER_API_KEY")
	defer func() {
		var err error
		if oldAPIKey != "" {
			err = os.Setenv("OPENROUTER_API_KEY", oldAPIKey)
		} else {
			err = os.Unsetenv("OPENROUTER_API_KEY")
		}
		if err != nil {
			t.Logf("Failed to restore environment variable: %v", err)
		}
	}()

	// This is now set below to a valid key format

	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	// Create the provider
	provider := NewProvider(logger)

	// Set test env vars with valid prefix
	err := os.Setenv("OPENROUTER_API_KEY", "sk-or-test-env-key")
	assert.NoError(t, err, "Failed to set environment variable")

	// Test with empty API key (should use env var)
	client, err := provider.CreateClient(context.Background(), "", "anthropic/claude-3-opus", "")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "anthropic/claude-3-opus", client.GetModelName())

	// We can only test functionality through public methods
	assert.Equal(t, "anthropic/claude-3-opus", client.GetModelName())
}

// TestProviderInitialization tests the provider initialization
func TestProviderInitialization(t *testing.T) {
	// Initialize provider with nil logger (should use default)
	provider := NewProvider(nil)
	assert.NotNil(t, provider, "Provider should not be nil even with nil logger")

	// Initialize provider with custom logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test-provider] ")
	provider = NewProvider(logger)
	assert.NotNil(t, provider, "Provider should not be nil with custom logger")
}

// TestAPIKeyValidation tests the API key validation logic
func TestAPIKeyValidation(t *testing.T) {
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")
	provider := NewProvider(logger)

	testCases := []struct {
		name          string
		apiKey        string
		expectedError bool
		errorContains string
	}{
		{
			name:          "Valid API key format",
			apiKey:        "sk-or-abcdefghijklmnopqrstuvwxyz",
			expectedError: false,
		},
		{
			name:          "Invalid API key format - OpenAI format",
			apiKey:        "sk-abcdefghijklmnopqrstuvwxyz",
			expectedError: true,
			errorContains: "invalid OpenRouter API key format",
		},
		{
			name:          "Invalid API key format - Gemini format",
			apiKey:        "gemini-123456789",
			expectedError: true,
			errorContains: "invalid OpenRouter API key format",
		},
		{
			name:          "Invalid API key format - random string",
			apiKey:        "not-a-valid-key",
			expectedError: true,
			errorContains: "invalid OpenRouter API key format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.CreateClient(context.Background(), tc.apiKey, "anthropic/claude-3-opus", "")

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
