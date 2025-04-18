// internal/llm/client.go
package llm

import (
	"context"
)

// ProviderResult holds the response from a content generation call
type ProviderResult struct {
	Content      string   // The generated content
	FinishReason string   // Why generation stopped, e.g., "stop", "length", "safety"
	TokenCount   int32    // Number of tokens in the response
	Truncated    bool     // Whether the response was truncated
	SafetyInfo   []Safety // Optional safety information
}

// Safety represents content safety evaluation information
type Safety struct {
	Category string  // Safety category name
	Blocked  bool    // Whether content was blocked due to this category
	Score    float32 // Severity score (provider-specific scale)
}

// ProviderTokenCount holds the result of a token counting operation
type ProviderTokenCount struct {
	Total int32 // Total token count
}

// ProviderModelInfo holds model capabilities and limits
type ProviderModelInfo struct {
	Name             string // Model name
	InputTokenLimit  int32  // Maximum input tokens allowed
	OutputTokenLimit int32  // Maximum output tokens allowed
}

// LLMClient defines the interface for interacting with any LLM provider
type LLMClient interface {
	// GenerateContent sends a text prompt to the LLM and returns the generated content
	// If params is provided, these parameters will override the default model parameters
	GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*ProviderResult, error)

	// CountTokens counts the tokens in a given prompt
	CountTokens(ctx context.Context, prompt string) (*ProviderTokenCount, error)

	// GetModelInfo retrieves information about the current model
	GetModelInfo(ctx context.Context) (*ProviderModelInfo, error)

	// GetModelName returns the name of the model being used
	GetModelName() string

	// Close releases resources used by the client
	Close() error
}

// MockLLMClient is a testing mock for the LLMClient interface
type MockLLMClient struct {
	GenerateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*ProviderResult, error)
	CountTokensFunc     func(ctx context.Context, prompt string) (*ProviderTokenCount, error)
	GetModelInfoFunc    func(ctx context.Context) (*ProviderModelInfo, error)
	GetModelNameFunc    func() string
	CloseFunc           func() error
}

// GenerateContent implementation for MockLLMClient
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*ProviderResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, params)
	}
	return &ProviderResult{Content: "Mock response"}, nil
}

// CountTokens implementation for MockLLMClient
func (m *MockLLMClient) CountTokens(ctx context.Context, prompt string) (*ProviderTokenCount, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return &ProviderTokenCount{Total: int32(len(prompt) / 4)}, nil // Simple approximation
}

// GetModelInfo implementation for MockLLMClient
func (m *MockLLMClient) GetModelInfo(ctx context.Context) (*ProviderModelInfo, error) {
	if m.GetModelInfoFunc != nil {
		return m.GetModelInfoFunc(ctx)
	}
	return &ProviderModelInfo{
		Name:             "mock-model",
		InputTokenLimit:  8192,
		OutputTokenLimit: 4096,
	}, nil
}

// GetModelName implementation for MockLLMClient
func (m *MockLLMClient) GetModelName() string {
	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "mock-model"
}

// Close implementation for MockLLMClient
func (m *MockLLMClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
