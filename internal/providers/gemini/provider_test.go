// internal/providers/gemini/provider_test.go
package gemini

import (
	"context"
	"errors"
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
	return &llm.ProviderResult{Content: "Test response"}, nil
}

// CountTokens implements the llm.LLMClient interface for testing
func (m *MockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return &llm.ProviderTokenCount{Total: 10}, nil
}

// GetModelInfo implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.GetModelInfoFunc != nil {
		return m.GetModelInfoFunc(ctx)
	}
	return &llm.ProviderModelInfo{
		Name:             "test-model",
		InputTokenLimit:  8192,
		OutputTokenLimit: 4096,
	}, nil
}

// GetModelName implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GetModelName() string {
	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "test-model"
}

// Close implements the llm.LLMClient interface for testing
func (m *MockLLMClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// TestGeminiProviderImplementsProviderInterface verifies that GeminiProvider implements the Provider interface
func TestGeminiProviderImplementsProviderInterface(t *testing.T) {
	// This is a compile-time check to ensure GeminiProvider implements Provider
	var _ providers.Provider = (*GeminiProvider)(nil)
}

// TestNewProvider verifies that NewProvider creates a valid GeminiProvider
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

// TestGeminiClientAdapter verifies that GeminiClientAdapter correctly wraps a client
func TestGeminiClientAdapter(t *testing.T) {
	// Create a mock client
	mockClient := &MockLLMClient{}

	// Create an adapter
	adapter := NewGeminiClientAdapter(mockClient)

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
		return &llm.ProviderResult{Content: "Mock response"}, nil
	}
	result, err := adapter.GenerateContent(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Expected no error from GenerateContent, got: %v", err)
	}
	if result.Content != "Mock response" {
		t.Errorf("Expected 'Mock response', got: %s", result.Content)
	}

	// Test error case
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
		return nil, errors.New("mock error")
	}
	_, err = adapter.GenerateContent(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("Expected error from GenerateContent, got nil")
	}
}

// TestGeminiClientGetModelInfo tests the GetModelInfo method
func TestGeminiClientGetModelInfo(t *testing.T) {
	// Create a mock client with a specific model info response
	mockClient := &MockLLMClient{
		GetModelInfoFunc: func(ctx context.Context) (*llm.ProviderModelInfo, error) {
			return &llm.ProviderModelInfo{
				Name:             "gemini-1.5-pro",
				InputTokenLimit:  32768,
				OutputTokenLimit: 8192,
			}, nil
		},
	}

	// Create an adapter
	adapter := NewGeminiClientAdapter(mockClient)

	// Test GetModelInfo
	info, err := adapter.GetModelInfo(context.Background())
	if err != nil {
		t.Fatalf("Expected no error from GetModelInfo, got: %v", err)
	}

	// Verify the model info
	if info.Name != "gemini-1.5-pro" {
		t.Errorf("Expected model name 'gemini-1.5-pro', got: %s", info.Name)
	}
	if info.InputTokenLimit != 32768 {
		t.Errorf("Expected input token limit 32768, got: %d", info.InputTokenLimit)
	}
	if info.OutputTokenLimit != 8192 {
		t.Errorf("Expected output token limit 8192, got: %d", info.OutputTokenLimit)
	}
}
