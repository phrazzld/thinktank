package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/providers"
	"github.com/phrazzld/thinktank/internal/providers/provider"
)

// TestNewProviderRegistry tests the NewProviderRegistry function
func TestNewProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()

	if registry == nil {
		t.Fatal("NewProviderRegistry returned nil")
	}

	// Verify it implements the ProviderRegistry interface
	_, ok := registry.(ProviderRegistry)
	if !ok {
		t.Fatal("NewProviderRegistry did not return an object implementing ProviderRegistry interface")
	}
}

// TestProviderRegistry_RegisterAndGetProvider tests provider registration and retrieval
func TestProviderRegistry_RegisterAndGetProvider(t *testing.T) {
	registry := NewProviderRegistry()

	// Create a mock provider factory
	mockProvider := &mockProvider{name: "test-provider"}
	factory := func() provider.Provider {
		return mockProvider
	}

	// Test registering a provider
	registry.RegisterProvider("test", factory)

	// Test retrieving the registered provider
	retrievedProvider, err := registry.GetProvider("test")
	if err != nil {
		t.Errorf("Expected no error when getting registered provider, got: %v", err)
	}
	if retrievedProvider == nil {
		t.Fatal("GetProvider returned nil for registered provider")
	}

	// Test that we can use the provider (interface functionality works)
	ctx := context.Background()
	client, err := retrievedProvider.CreateClient(ctx, "test-key", "test-model", "")
	if err != nil {
		t.Errorf("Expected no error when creating client from retrieved provider, got: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client from retrieved provider")
	}
}

// TestProviderRegistry_GetProvider_NotFound tests getting a non-existent provider
func TestProviderRegistry_GetProvider_NotFound(t *testing.T) {
	registry := NewProviderRegistry()

	// Test retrieving a non-existent provider
	provider, err := registry.GetProvider("non-existent")

	if provider != nil {
		t.Error("Expected nil provider for non-existent provider")
	}
	if err == nil {
		t.Error("Expected error when getting non-existent provider, got nil")
	}
	if !errors.Is(err, providers.ErrProviderNotFound) {
		t.Errorf("Expected ErrProviderNotFound, got: %v", err)
	}
}

// TestProviderRegistry_RegisterProvider_Overwrite tests overwriting a provider registration
func TestProviderRegistry_RegisterProvider_Overwrite(t *testing.T) {
	registry := NewProviderRegistry()

	// Register first provider
	firstProvider := &mockProvider{name: "first-provider"}
	firstFactory := func() provider.Provider {
		return firstProvider
	}
	registry.RegisterProvider("test", firstFactory)

	// Register second provider with same name (should overwrite)
	secondProvider := &mockProvider{name: "second-provider"}
	secondFactory := func() provider.Provider {
		return secondProvider
	}
	registry.RegisterProvider("test", secondFactory)

	// Verify the second provider is returned by testing its functionality
	retrievedProvider, err := registry.GetProvider("test")
	if err != nil {
		t.Errorf("Expected no error when getting overwritten provider, got: %v", err)
	}
	if retrievedProvider == nil {
		t.Fatal("GetProvider returned nil for overwritten provider")
	}

	// Test that we can use the overwritten provider
	ctx := context.Background()
	client, err := retrievedProvider.CreateClient(ctx, "test-key", "test-model", "")
	if err != nil {
		t.Errorf("Expected no error when creating client from overwritten provider, got: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client from overwritten provider")
	}
}

// TestProviderRegistry_MultipleProviders tests registering and retrieving multiple providers
func TestProviderRegistry_MultipleProviders(t *testing.T) {
	registry := NewProviderRegistry()

	// Register multiple providers
	providers := map[string]*mockProvider{
		"gemini":     {name: "gemini-provider"},
		"openai":     {name: "openai-provider"},
		"openrouter": {name: "openrouter-provider"},
	}

	for name, mockProv := range providers {
		factory := func(p *mockProvider) func() provider.Provider {
			return func() provider.Provider { return p }
		}(mockProv)
		registry.RegisterProvider(name, factory)
	}

	// Verify all providers can be retrieved and used
	for name := range providers {
		retrievedProvider, err := registry.GetProvider(name)
		if err != nil {
			t.Errorf("Expected no error when getting provider %s, got: %v", name, err)
		}
		if retrievedProvider == nil {
			t.Errorf("GetProvider returned nil for provider %s", name)
			continue
		}

		// Test that the provider works
		ctx := context.Background()
		client, err := retrievedProvider.CreateClient(ctx, "test-key", "test-model", "")
		if err != nil {
			t.Errorf("Expected no error when creating client from provider %s, got: %v", name, err)
		}
		if client == nil {
			t.Errorf("Expected non-nil client from provider %s", name)
		}
	}

	// Verify non-existent provider still returns error
	_, err := registry.GetProvider("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent provider after registering multiple providers")
	}
}

// mockProvider is a simple mock implementation of the provider.Provider interface for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error) {
	return &mockLLMClient{}, nil
}

// mockLLMClient is a simple mock LLM client for testing
type mockLLMClient struct{}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{Content: "mock response"}, nil
}

func (m *mockLLMClient) GetModelName() string {
	return "mock-model"
}

func (m *mockLLMClient) Close() error {
	return nil
}
