// Package openai provides shared test utilities for OpenAI client tests
package openai

import (
	"context"
	"fmt"

	"github.com/phrazzld/thinktank/internal/llm"
)

// Helper function to convert to a pointer
//
//nolint:unused
func toPtr[T any](v T) *T {
	return &v
}

// CreateMockOpenAIClientForTesting creates a mocked OpenAI client for testing
func CreateMockOpenAIClientForTesting(modelName string, responseFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)) (llm.LLMClient, error) {
	// Create a new mock client instead
	mockClient := &MockClient{
		GetModelNameFunc: func() string {
			return modelName
		},
		GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
			if responseFunc != nil {
				return responseFunc(ctx, prompt, params)
			}
			return &llm.ProviderResult{
				Content: fmt.Sprintf("Mock OpenAI response for: %s", prompt),
			}, nil
		},
	}

	return mockClient, nil
}

// MockAPIErrorResponse creates a mock error response with specifics about the error
// This is now just a convenience function that forwards to the version in errors.go
func MockAPIErrorResponseOld(errorType int, statusCode int, message string, details string) *llm.LLMError {
	return MockAPIErrorResponse(errorType, statusCode, message, details)
}

// MockClient is a mock implementation of the llm.LLMClient interface for testing
type MockClient struct {
	GenerateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	GetModelNameFunc    func() string
	CloseFunc           func() error
	SetTemperatureFunc  func(temp float32)
	SetTopPFunc         func(topP float32)
	SetMaxTokensFunc    func(tokens int32)
}

// GenerateContent implements the llm.LLMClient interface
func (m *MockClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{
		Content: fmt.Sprintf("Mock response for: %s", prompt),
	}, nil
}

// GetModelName implements the llm.LLMClient interface
func (m *MockClient) GetModelName() string {
	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "gpt-4"
}

// Close implements the llm.LLMClient interface
func (m *MockClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// SetTemperature implements parameter setting for the OpenAI client
func (m *MockClient) SetTemperature(temp float32) {
	if m.SetTemperatureFunc != nil {
		m.SetTemperatureFunc(temp)
	}
}

// SetTopP implements parameter setting for the OpenAI client
func (m *MockClient) SetTopP(topP float32) {
	if m.SetTopPFunc != nil {
		m.SetTopPFunc(topP)
	}
}

// SetMaxTokens implements parameter setting for the OpenAI client
func (m *MockClient) SetMaxTokens(tokens int32) {
	if m.SetMaxTokensFunc != nil {
		m.SetMaxTokensFunc(tokens)
	}
}

// SetFrequencyPenalty implements parameter setting for the OpenAI client
func (m *MockClient) SetFrequencyPenalty(penalty float32) {
	// No-op implementation for the mock
}

// SetPresencePenalty implements parameter setting for the OpenAI client
func (m *MockClient) SetPresencePenalty(penalty float32) {
	// No-op implementation for the mock
}
