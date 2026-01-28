package testutil

import (
	"context"

	"github.com/misty-step/thinktank/internal/thinktank/tokenizers"
)

// MockTokenizerManager implements TokenizerManager for testing.
type MockTokenizerManager struct {
	SupportedProviders map[string]bool
	Tokenizers         map[string]tokenizers.AccurateTokenCounter
	GetTokenizerError  error
}

func (m *MockTokenizerManager) GetTokenizer(provider string) (tokenizers.AccurateTokenCounter, error) {
	if m.GetTokenizerError != nil {
		return nil, m.GetTokenizerError
	}

	if tokenizer, exists := m.Tokenizers[provider]; exists {
		return tokenizer, nil
	}

	// Return basic mock tokenizer for any provider
	return &MockAccurateTokenCounter{}, nil
}

func (m *MockTokenizerManager) SupportsProvider(provider string) bool {
	if m.SupportedProviders == nil {
		return false
	}
	return m.SupportedProviders[provider]
}

func (m *MockTokenizerManager) ClearCache() {
	// No-op for mock
}

// MockAccurateTokenCounter implements AccurateTokenCounter for testing.
type MockAccurateTokenCounter struct {
	CountTokensFunc   func(ctx context.Context, text string, modelName string) (int, error)
	SupportsModelFunc func(modelName string) bool
	GetEncodingFunc   func(modelName string) (string, error)
	TokenCount        int
	Error             error
}

func (m *MockAccurateTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, text, modelName)
	}

	if m.Error != nil {
		return 0, m.Error
	}

	if m.TokenCount > 0 {
		return m.TokenCount, nil
	}

	// Default: simple character-based count for testing
	return len(text), nil
}

func (m *MockAccurateTokenCounter) SupportsModel(modelName string) bool {
	if m.SupportsModelFunc != nil {
		return m.SupportsModelFunc(modelName)
	}
	return true // Default: support all models
}

func (m *MockAccurateTokenCounter) GetEncoding(modelName string) (string, error) {
	if m.GetEncodingFunc != nil {
		return m.GetEncodingFunc(modelName)
	}
	return "mock-encoding", nil
}
