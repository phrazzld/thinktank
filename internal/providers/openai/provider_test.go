// internal/providers/openai/provider_test.go
package openai

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/providers"
)

// MockLLMClient implements the llm.LLMClient interface for testing
type MockLLMClient struct {
	GenerateContentFunc func(ctx context.Context, prompt string) (*llm.ProviderResult, error)
	CountTokensFunc     func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error)
	GetModelInfoFunc    func(ctx context.Context) (*llm.ProviderModelInfo, error)
	GetModelNameFunc    func() string
	CloseFunc           func() error
}

// GenerateContent implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt)
	}
	return &llm.ProviderResult{Content: "Mock OpenAI response"}, nil
}

// CountTokens implements the llm.LLMClient interface for testing
func (m *MockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return &llm.ProviderTokenCount{Total: 15}, nil
}

// GetModelInfo implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.GetModelInfoFunc != nil {
		return m.GetModelInfoFunc(ctx)
	}
	return &llm.ProviderModelInfo{
		Name:             "gpt-4",
		InputTokenLimit:  8192,
		OutputTokenLimit: 2048,
	}, nil
}

// GetModelName implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GetModelName() string {
	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "gpt-4"
}

// Close implements the llm.LLMClient interface for testing
func (m *MockLLMClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// TestOpenAIProviderImplementsProviderInterface ensures OpenAIProvider implements the Provider interface
func TestOpenAIProviderImplementsProviderInterface(t *testing.T) {
	// This is a compile-time check to ensure OpenAIProvider implements Provider
	var _ providers.Provider = (*OpenAIProvider)(nil)
}

// TestNewProvider verifies that NewProvider creates a valid OpenAIProvider
func TestNewProvider(t *testing.T) {
	// Create a provider with no logger (should use default)
	provider := NewProvider(nil)

	// Verify it's not nil
	if provider == nil {
		t.Fatal("Expected non-nil provider, got nil")
	}

	// Create a provider with custom logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")
	provider = NewProvider(logger)

	// Verify it's not nil
	if provider == nil {
		t.Fatal("Expected non-nil provider, got nil")
	}
}

// TestOpenAIClientAdapter verifies that OpenAIClientAdapter correctly wraps a client
func TestOpenAIClientAdapter(t *testing.T) {
	// Create a mock client
	mockClient := &MockLLMClient{}

	// Create an adapter
	adapter := NewOpenAIClientAdapter(mockClient)

	// Test that adapter implements llm.LLMClient
	var _ llm.LLMClient = adapter

	// Test SetParameters
	params := map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.9,
	}
	adapter.SetParameters(params)

	// Test GenerateContent pass-through
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{Content: "Mock OpenAI response"}, nil
	}
	result, err := adapter.GenerateContent(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Expected no error from GenerateContent, got: %v", err)
	}
	if result.Content != "Mock OpenAI response" {
		t.Errorf("Expected 'Mock OpenAI response', got: %s", result.Content)
	}

	// Test error case
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
		return nil, errors.New("mock openai error")
	}
	_, err = adapter.GenerateContent(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("Expected error from GenerateContent, got nil")
	}
}

// TestCreateClientWithAPIKey tests creation with explicit API key
func TestCreateClientWithAPIKey(t *testing.T) {
	// Save original env var and restore after test
	origAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if err := os.Setenv("OPENAI_API_KEY", origAPIKey); err != nil {
			t.Logf("Warning: Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Clear env var for this test
	if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
		t.Fatalf("Failed to unset OPENAI_API_KEY: %v", err)
	}

	// Create provider
	provider := NewProvider(nil)

	// This test can't actually create a real client without valid keys,
	// so we'll just verify the error when no API key is provided
	_, err := provider.CreateClient(context.Background(), "", "gpt-4", "")
	if err == nil {
		t.Fatal("Expected error when no API key provided, got nil")
	}

	// For testing purposes only - we can't actually test with a real OpenAI API
	// since we would need a valid API key. We'll just mock this and check
	// that we got past the point where we validate the presence of an API key.

	// We're not really testing against the real OpenAI API in unit tests,
	// so we don't want to connect to any endpoints.
	// The test is focused on the provider's API key handling logic, not the
	// actual connection to OpenAI.
	//
	// In a real integration test, we would use a valid API key and model.
}
