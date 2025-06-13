package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers"
)

// TestRegistry_GetAvailableModels tests the GetAvailableModels function
func TestRegistry_GetAvailableModels(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	t.Run("empty registry", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		models, err := registry.GetAvailableModels(ctx)

		if err != nil {
			t.Errorf("Expected no error for empty registry, got: %v", err)
		}
		if len(models) != 0 {
			t.Errorf("Expected empty slice for empty registry, got %d models", len(models))
		}
	})

	t.Run("registry with models", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		// Manually add some models to the registry for testing
		testModels := map[string]ModelDefinition{
			"gpt-4": {
				Name:       "GPT-4",
				Provider:   "openai",
				APIModelID: "gpt-4",
			},
			"gemini-pro": {
				Name:       "Gemini Pro",
				Provider:   "gemini",
				APIModelID: "gemini-pro",
			},
			"claude-3": {
				Name:       "Claude 3",
				Provider:   "anthropic",
				APIModelID: "claude-3",
			},
		}

		// Use reflection or direct field access to add models
		// Since we can't access private fields directly, we'll use a different approach
		registry.models = testModels

		models, err := registry.GetAvailableModels(ctx)

		if err != nil {
			t.Errorf("Expected no error for populated registry, got: %v", err)
		}
		if len(models) != len(testModels) {
			t.Errorf("Expected %d models, got %d", len(testModels), len(models))
		}

		// Verify all expected models are present
		modelSet := make(map[string]bool)
		for _, model := range models {
			modelSet[model] = true
		}

		for expectedModel := range testModels {
			if !modelSet[expectedModel] {
				t.Errorf("Expected model %s not found in returned list", expectedModel)
			}
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		// Add some test models
		registry.models = map[string]ModelDefinition{
			"test-model": {
				Name:       "Test Model",
				Provider:   "test",
				APIModelID: "test-model",
			},
		}

		// Test concurrent access (this tests the mutex locking)
		done := make(chan bool, 2)

		go func() {
			models, err := registry.GetAvailableModels(ctx)
			if err != nil {
				t.Errorf("Goroutine 1: Expected no error, got: %v", err)
			}
			if len(models) != 1 {
				t.Errorf("Goroutine 1: Expected 1 model, got %d", len(models))
			}
			done <- true
		}()

		go func() {
			models, err := registry.GetAvailableModels(ctx)
			if err != nil {
				t.Errorf("Goroutine 2: Expected no error, got: %v", err)
			}
			if len(models) != 1 {
				t.Errorf("Goroutine 2: Expected 1 model, got %d", len(models))
			}
			done <- true
		}()

		// Wait for both goroutines to complete
		<-done
		<-done
	})
}

// TestRegistry_CreateLLMClient_ErrorScenarios tests error scenarios in CreateLLMClient
func TestRegistry_CreateLLMClient_ErrorScenarios(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	t.Run("model not found", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		client, err := registry.CreateLLMClient(ctx, "test-api-key", "non-existent-model")

		if client != nil {
			t.Error("Expected nil client for non-existent model")
		}
		if err == nil {
			t.Error("Expected error for non-existent model, got nil")
		}
	})

	t.Run("provider not found", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		// Add a model with non-existent provider
		registry.models = map[string]ModelDefinition{
			"test-model": {
				Name:       "Test Model",
				Provider:   "non-existent-provider",
				APIModelID: "test-model",
			},
		}

		client, err := registry.CreateLLMClient(ctx, "test-api-key", "test-model")

		if client != nil {
			t.Error("Expected nil client for non-existent provider")
		}
		if err == nil {
			t.Error("Expected error for non-existent provider, got nil")
		}
	})

	t.Run("provider implementation not found", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		// Add provider and model but no implementation
		registry.providers = map[string]ProviderDefinition{
			"test-provider": {
				Name: "Test Provider",
			},
		}
		registry.models = map[string]ModelDefinition{
			"test-model": {
				Name:       "Test Model",
				Provider:   "test-provider",
				APIModelID: "test-model",
			},
		}

		client, err := registry.CreateLLMClient(ctx, "test-api-key", "test-model")

		if client != nil {
			t.Error("Expected nil client for missing provider implementation")
		}
		if err == nil {
			t.Error("Expected error for missing provider implementation, got nil")
		}
	})

	t.Run("provider client creation failure", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		// Add provider, model, and failing implementation
		registry.providers = map[string]ProviderDefinition{
			"test-provider": {
				Name: "Test Provider",
			},
		}
		registry.models = map[string]ModelDefinition{
			"test-model": {
				Name:       "Test Model",
				Provider:   "test-provider",
				APIModelID: "test-model",
			},
		}

		// Add a failing provider implementation
		failingProvider := &mockFailingProvider{}
		registry.implementations = map[string]providers.Provider{
			"test-provider": failingProvider,
		}

		client, err := registry.CreateLLMClient(ctx, "test-api-key", "test-model")

		if client != nil {
			t.Error("Expected nil client for failing provider")
		}
		if err == nil {
			t.Error("Expected error for failing provider, got nil")
		}
	})

	t.Run("successful client creation", func(t *testing.T) {
		registry := NewRegistry(logger)
		ctx := context.Background()

		// Add provider, model, and working implementation
		registry.providers = map[string]ProviderDefinition{
			"test-provider": {
				Name: "Test Provider",
			},
		}
		registry.models = map[string]ModelDefinition{
			"test-model": {
				Name:       "Test Model",
				Provider:   "test-provider",
				APIModelID: "test-model",
			},
		}

		// Add a working provider implementation
		workingProvider := &mockWorkingProvider{}
		registry.implementations = map[string]providers.Provider{
			"test-provider": workingProvider,
		}

		client, err := registry.CreateLLMClient(ctx, "test-api-key", "test-model")

		if err != nil {
			t.Errorf("Expected no error for working provider, got: %v", err)
		}
		if client == nil {
			t.Error("Expected non-nil client for working provider")
		}

		// Verify the provider received the correct parameters
		if workingProvider.lastAPIKey != "test-api-key" {
			t.Errorf("Expected API key 'test-api-key', provider received '%s'", workingProvider.lastAPIKey)
		}
		// The API endpoint should be empty since we didn't set BaseURL in the provider
		if workingProvider.lastAPIEndpoint != "" {
			t.Errorf("Expected empty API endpoint, provider received '%s'", workingProvider.lastAPIEndpoint)
		}
	})
}

// mockFailingProvider is a mock provider that always fails to create clients
type mockFailingProvider struct{}

func (m *mockFailingProvider) CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error) {
	return nil, errors.New("mock provider client creation failure")
}

// mockWorkingProvider is a mock provider that successfully creates clients
type mockWorkingProvider struct {
	lastAPIKey      string
	lastAPIEndpoint string
}

func (m *mockWorkingProvider) CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error) {
	m.lastAPIKey = apiKey
	m.lastAPIEndpoint = apiEndpoint
	return &mockClient{}, nil
}

// mockClient is a simple mock client
type mockClient struct{}

func (m *mockClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{Content: "mock response"}, nil
}

func (m *mockClient) GetModelName() string {
	return "mock-model"
}

func (m *mockClient) Close() error {
	return nil
}
