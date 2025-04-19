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
	GenerateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	CountTokensFunc     func(ctx context.Context, prompt string) (int32, error)
	GetModelLimitsFunc  func(ctx context.Context) (*llm.ModelLimits, error)
	GetModelNameFunc    func() string
	CloseFunc           func() error
}

// GenerateContent implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{Content: "Test response"}, nil
}

// CountTokens implements the llm.LLMClient interface for testing
func (m *MockLLMClient) CountTokens(ctx context.Context, prompt string) (int32, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return 10, nil
}

// GetModelLimits implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GetModelLimits(ctx context.Context) (*llm.ModelLimits, error) {
	if m.GetModelLimitsFunc != nil {
		return m.GetModelLimitsFunc(ctx)
	}
	return &llm.ModelLimits{
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
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{Content: "Mock response"}, nil
	}
	result, err := adapter.GenerateContent(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Expected no error from GenerateContent, got: %v", err)
	}
	if result.Content != "Mock response" {
		t.Errorf("Expected 'Mock response', got: %s", result.Content)
	}

	// Test error case
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return nil, errors.New("mock error")
	}
	_, err = adapter.GenerateContent(context.Background(), "test prompt", nil)
	if err == nil {
		t.Fatal("Expected error from GenerateContent, got nil")
	}
}

// TestGeminiClientGetModelLimits tests the GetModelLimits method
func TestGeminiClientGetModelLimits(t *testing.T) {
	// Create a mock client with a specific model limits response
	mockClient := &MockLLMClient{
		GetModelLimitsFunc: func(ctx context.Context) (*llm.ModelLimits, error) {
			return &llm.ModelLimits{
				InputTokenLimit:  32768,
				OutputTokenLimit: 8192,
			}, nil
		},
		GetModelNameFunc: func() string {
			return "gemini-1.5-pro"
		},
	}

	// Create an adapter
	adapter := NewGeminiClientAdapter(mockClient)

	// Test GetModelLimits
	limits, err := adapter.GetModelLimits(context.Background())
	if err != nil {
		t.Fatalf("Expected no error from GetModelLimits, got: %v", err)
	}

	// Verify the model limits
	if adapter.GetModelName() != "gemini-1.5-pro" {
		t.Errorf("Expected model name 'gemini-1.5-pro', got: %s", adapter.GetModelName())
	}
	if limits.InputTokenLimit != 32768 {
		t.Errorf("Expected input token limit 32768, got: %d", limits.InputTokenLimit)
	}
	if limits.OutputTokenLimit != 8192 {
		t.Errorf("Expected output token limit 8192, got: %d", limits.OutputTokenLimit)
	}
}
