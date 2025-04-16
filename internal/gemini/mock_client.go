// internal/gemini/mock_client.go
package gemini

import (
	"context"
)

// MockClient implements Client interface for testing
type MockClient struct {
	GenerateContentFunc    func(ctx context.Context, prompt string, params map[string]interface{}) (*GenerationResult, error)
	CountTokensFunc        func(ctx context.Context, prompt string) (*TokenCount, error)
	GetModelInfoFunc       func(ctx context.Context) (*ModelInfo, error)
	GetModelNameFunc       func() string
	GetTemperatureFunc     func() float32
	GetMaxOutputTokensFunc func() int32
	GetTopPFunc            func() float32
	CloseFunc              func() error
}

// GenerateContent calls the mocked implementation
func (m *MockClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*GenerationResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, params)
	}
	return &GenerationResult{Content: "Mock response"}, nil
}

// CountTokens calls the mocked implementation
func (m *MockClient) CountTokens(ctx context.Context, prompt string) (*TokenCount, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return &TokenCount{Total: 10}, nil // Simple default
}

// GetModelInfo calls the mocked implementation
func (m *MockClient) GetModelInfo(ctx context.Context) (*ModelInfo, error) {
	if m.GetModelInfoFunc != nil {
		return m.GetModelInfoFunc(ctx)
	}
	return &ModelInfo{
		Name:             "mock-model",
		InputTokenLimit:  32000,
		OutputTokenLimit: 8192,
	}, nil
}

// Close calls the mocked implementation
func (m *MockClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// GetModelName returns the mocked model name
func (m *MockClient) GetModelName() string {
	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "mock-model"
}

// GetTemperature returns the mocked temperature
func (m *MockClient) GetTemperature() float32 {
	if m.GetTemperatureFunc != nil {
		return m.GetTemperatureFunc()
	}
	return DefaultModelConfig().Temperature
}

// GetMaxOutputTokens returns the mocked max output tokens
func (m *MockClient) GetMaxOutputTokens() int32 {
	if m.GetMaxOutputTokensFunc != nil {
		return m.GetMaxOutputTokensFunc()
	}
	return DefaultModelConfig().MaxOutputTokens
}

// GetTopP returns the mocked topP
func (m *MockClient) GetTopP() float32 {
	if m.GetTopPFunc != nil {
		return m.GetTopPFunc()
	}
	return DefaultModelConfig().TopP
}

// NewMockClient creates a new mock client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		GetModelInfoFunc: func(ctx context.Context) (*ModelInfo, error) {
			return &ModelInfo{
				Name:             "mock-model",
				InputTokenLimit:  32000,
				OutputTokenLimit: 8192,
			}, nil
		},
	}
}
