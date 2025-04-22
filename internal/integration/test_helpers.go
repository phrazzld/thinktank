// Package integration provides testing utilities for the thinktank project
// This is a simplified version with token functionality removed (T036C)
package integration

import (
	"context"

	"github.com/phrazzld/thinktank/internal/gemini"
)

// TestEnv represents a test environment for integration tests
type TestEnv struct {
	// MockClient is a mock implementation of the Gemini client
	MockClient *gemini.MockClient
}

// NewTestEnv creates a new test environment
func NewTestEnv() *TestEnv {
	mockClient := &gemini.MockClient{
		GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
			return &gemini.GenerationResult{
				Content:      "Test content",
				FinishReason: "stop",
				Truncated:    false,
			}, nil
		},
		GetModelNameFunc: func() string {
			return "gemini-pro"
		},
	}

	return &TestEnv{
		MockClient: mockClient,
	}
}

// SimpleResult represents a simple test result
type SimpleResult struct {
	Content string
}

// MockResponse generates a mock response for testing
func MockResponse(content string) *SimpleResult {
	return &SimpleResult{
		Content: content,
	}
}
