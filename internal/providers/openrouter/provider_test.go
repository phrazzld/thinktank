package openrouter

import (
	"context"
	"os"
	"testing"

	"github.com/phrazzld/architect/internal/logutil"
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

	// Set test env vars
	err := os.Setenv("OPENROUTER_API_KEY", "test-env-key")
	assert.NoError(t, err, "Failed to set environment variable")

	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	// Create the provider
	provider := NewProvider(logger)

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
