// internal/gemini/mock_client.go
package gemini

import (
	"context"
)

// MockClient implements Client interface for testing
type MockClient struct {
	GenerateContentFunc func(ctx context.Context, prompt string) (*GenerationResult, error)
	CountTokensFunc     func(ctx context.Context, prompt string) (*TokenCount, error)
	GetModelInfoFunc    func(ctx context.Context) (*ModelInfo, error)
	CloseFunc           func() error
}

// GenerateContent calls the mocked implementation
func (m *MockClient) GenerateContent(ctx context.Context, prompt string) (*GenerationResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt)
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
