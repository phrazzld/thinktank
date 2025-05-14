package thinktank

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestGetModelTokenLimits_UsesConfigValues ensures that the GetModelTokenLimits method
// correctly uses token limits from model configuration rather than hardcoded defaults
func TestGetModelTokenLimits_UsesConfigValues(t *testing.T) {
	// Setup a mock registry with predefined models
	models := map[string]*registry.ModelDefinition{
		"test-huge-model": {
			Name:            "test-huge-model",
			Provider:        "test-provider",
			APIModelID:      "test-huge-model",
			ContextWindow:   1000000, // 1M tokens
			MaxOutputTokens: 200000,  // 200K tokens
		},
		"test-model-no-limits": {
			Name:       "test-model-no-limits",
			Provider:   "test-provider",
			APIModelID: "test-model-no-limits",
			// No explicit token limits
		},
	}

	mockRegistry := &MockRegistry{
		Models: models,
	}

	// Create a registry API service with the mock registry
	logger := testutil.NewMockLogger()
	service := NewRegistryAPIService(mockRegistry, logger)

	// Create a context for tests
	ctx := context.Background()

	// Test 1: Model with explicit limits should use those limits
	contextWindow, maxOutputTokens, err := service.GetModelTokenLimits(ctx, "test-huge-model")
	if err != nil {
		t.Fatalf("Unexpected error getting token limits: %v", err)
	}

	if contextWindow != 1000000 {
		t.Errorf("Expected context window to be 1000000, got %d", contextWindow)
	}

	if maxOutputTokens != 200000 {
		t.Errorf("Expected max output tokens to be 200000, got %d", maxOutputTokens)
	}

	// Test 2: Model with no limits should use high default values
	contextWindow, maxOutputTokens, err = service.GetModelTokenLimits(ctx, "test-model-no-limits")
	if err != nil {
		t.Fatalf("Unexpected error getting token limits: %v", err)
	}

	// Should use the high default values
	if contextWindow != 1000000 {
		t.Errorf("Expected default context window to be 1000000, got %d", contextWindow)
	}

	if maxOutputTokens != 65000 {
		t.Errorf("Expected default max output tokens to be 65000, got %d", maxOutputTokens)
	}
}

// MockRegistry implements the necessary methods for testing the registry API service
type MockRegistry struct {
	Models map[string]*registry.ModelDefinition
}

func (m *MockRegistry) GetModel(ctx context.Context, name string) (*registry.ModelDefinition, error) {
	model, ok := m.Models[name]
	if !ok {
		return nil, errors.New("model not found")
	}
	return model, nil
}

func (m *MockRegistry) GetProvider(ctx context.Context, name string) (*registry.ProviderDefinition, error) {
	// Not used in these tests
	return &registry.ProviderDefinition{
		Name: name,
	}, nil
}

func (m *MockRegistry) GetProviderImplementation(ctx context.Context, name string) (interface{}, error) {
	// Not used in these tests
	return nil, nil
}
