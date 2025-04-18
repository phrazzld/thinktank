// internal/providers/provider_test.go
package providers

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	CreateClientFunc func(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error)
}

// CreateClient implementation for MockProvider
func (m *MockProvider) CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error) {
	if m.CreateClientFunc != nil {
		return m.CreateClientFunc(ctx, apiKey, modelID, apiEndpoint)
	}
	return &llm.MockLLMClient{}, nil
}

// TestProviderInterface verifies that MockProvider satisfies the Provider interface
func TestProviderInterface(t *testing.T) {
	// This is a compile-time check to ensure MockProvider implements Provider
	var _ Provider = (*MockProvider)(nil)

	// Create a mock provider
	provider := &MockProvider{}

	// Test creating a client
	client, err := provider.CreateClient(context.Background(), "test-key", "test-model", "")

	// Verify there was no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify client is not nil
	if client == nil {
		t.Fatal("Expected non-nil client, got nil")
	}

	// Verify client implements LLMClient interface
	modelName := client.GetModelName()
	if modelName == "" {
		t.Fatal("Expected non-empty model name from client")
	}
}
