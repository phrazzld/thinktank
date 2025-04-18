// Package openrouter contains tests for the OpenRouter client
package openrouter

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Skip client tests that require direct access to the unexported client type.
// These tests would need to be reworked if the client is exported or access is provided through exported functions.

// TestCreateClientThroughProvider tests client creation through the provider
func TestCreateClientThroughProvider(t *testing.T) {
	// Create a provider
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
	provider := NewProvider(logger)

	tests := []struct {
		name        string
		apiKey      string
		modelID     string
		apiEndpoint string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Valid inputs with default endpoint",
			apiKey:      "test-api-key",
			modelID:     "anthropic/claude-3-opus-20240229",
			apiEndpoint: "",
			wantErr:     false,
		},
		{
			name:        "Valid inputs with custom endpoint",
			apiKey:      "test-api-key",
			modelID:     "anthropic/claude-3-opus-20240229",
			apiEndpoint: "https://custom-endpoint.com/api/v1",
			wantErr:     false,
		},
		{
			name:        "Empty API key",
			apiKey:      "",
			modelID:     "anthropic/claude-3-opus-20240229",
			apiEndpoint: "",
			wantErr:     true,
			errContains: "no OpenRouter API key provided",
		},
		{
			name:        "Empty model ID",
			apiKey:      "test-api-key",
			modelID:     "",
			apiEndpoint: "",
			wantErr:     true,
			errContains: "model ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OPENROUTER_API_KEY", "")

			// Create client via provider
			client, err := provider.CreateClient(context.Background(), tt.apiKey, tt.modelID, tt.apiEndpoint)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)

				// Check that the client's model name matches what we provided
				assert.Equal(t, tt.modelID, client.GetModelName())
			}
		})
	}
}

// TestCountTokensThroughProvider tests the token counting functionality through a client obtained from the provider
func TestCountTokensThroughProvider(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		prompt    string
		expectErr bool
	}{
		{
			name:      "Claude model token count",
			modelID:   "anthropic/claude-3-opus-20240229",
			prompt:    "Hello, world!",
			expectErr: false,
		},
		{
			name:      "OpenAI model token count",
			modelID:   "openai/gpt-4",
			prompt:    "This is a test prompt for counting tokens.",
			expectErr: false,
		},
		{
			name:      "Empty prompt",
			modelID:   "anthropic/claude-3-opus-20240229",
			prompt:    "",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create provider
			logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
			provider := NewProvider(logger)

			// Get client
			client, err := provider.CreateClient(context.Background(), "test-api-key", tt.modelID, "")
			require.NoError(t, err)

			// Call CountTokens
			count, err := client.CountTokens(context.Background(), tt.prompt)

			// Run assertions
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, count)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, count)
				if tt.prompt == "" {
					assert.Equal(t, int32(0), count.Total)
				} else {
					assert.Greater(t, count.Total, int32(0))
				}
			}
		})
	}
}

// TestGetModelInfoThroughProvider tests the model info retrieval
func TestGetModelInfoThroughProvider(t *testing.T) {
	tests := []struct {
		name         string
		modelID      string
		expectedInfo *llm.ProviderModelInfo
	}{
		{
			name:    "Claude 3 Opus",
			modelID: "anthropic/claude-3-opus-20240229",
			expectedInfo: &llm.ProviderModelInfo{
				Name:             "anthropic/claude-3-opus-20240229",
				InputTokenLimit:  200000,
				OutputTokenLimit: 4096,
			},
		},
		{
			name:    "GPT-4 Turbo",
			modelID: "openai/gpt-4-turbo",
			expectedInfo: &llm.ProviderModelInfo{
				Name:             "openai/gpt-4-turbo",
				InputTokenLimit:  128000,
				OutputTokenLimit: 4096,
			},
		},
		{
			name:    "Gemini 1.5",
			modelID: "google/gemini-1.5-pro",
			expectedInfo: &llm.ProviderModelInfo{
				Name:             "google/gemini-1.5-pro",
				InputTokenLimit:  1000000,
				OutputTokenLimit: 8192,
			},
		},
		{
			name:    "Unknown provider",
			modelID: "unknown/model",
			expectedInfo: &llm.ProviderModelInfo{
				Name:             "unknown/model",
				InputTokenLimit:  8192,
				OutputTokenLimit: 2048,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create provider
			logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
			provider := NewProvider(logger)

			// Get client
			client, err := provider.CreateClient(context.Background(), "test-api-key", tt.modelID, "")
			require.NoError(t, err)

			// Call GetModelInfo
			info, err := client.GetModelInfo(context.Background())

			// Check result
			assert.NoError(t, err)
			assert.NotNil(t, info)
			assert.Equal(t, tt.expectedInfo.Name, info.Name)
			assert.Equal(t, tt.expectedInfo.InputTokenLimit, info.InputTokenLimit)
			assert.Equal(t, tt.expectedInfo.OutputTokenLimit, info.OutputTokenLimit)
		})
	}
}

// TestClientMethodsThroughProvider tests the client's methods
func TestClientMethodsThroughProvider(t *testing.T) {
	// Create provider
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
	provider := NewProvider(logger)

	// Get client
	client, err := provider.CreateClient(context.Background(), "test-api-key", "anthropic/claude-3-opus-20240229", "")
	require.NoError(t, err)

	// Test GetModelName
	assert.Equal(t, "anthropic/claude-3-opus-20240229", client.GetModelName())

	// Test Close
	err = client.Close()
	assert.NoError(t, err)
}
