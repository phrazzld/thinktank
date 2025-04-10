package architect

import (
	"context"
	"github.com/phrazzld/architect/internal/gemini"
)

// MockClient is a mock implementation of gemini.Client for testing
type MockClient struct {
	CountTokensFunc  func(ctx context.Context, prompt string) (*gemini.TokenCount, error)
	GenerateContentFunc func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	GetModelInfoFunc func(ctx context.Context) (*gemini.ModelInfo, error)
	CloseFunc       func() error
}

func (m *MockClient) CountTokens(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return nil, nil
}

func (m *MockClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt)
	}
	return nil, nil
}

func (m *MockClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	if m.GetModelInfoFunc != nil {
		return m.GetModelInfoFunc(ctx)
	}
	return nil, nil
}

func (m *MockClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}