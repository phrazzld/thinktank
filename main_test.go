// main_test.go
package main

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

func TestCheckTokenLimit(t *testing.T) {
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test case 1: Token count is within limits
	t.Run("TokenCountWithinLimits", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 1000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err != nil {
			t.Errorf("Expected no error for token count within limits, got: %v", err)
		}
	})

	// Test case 2: Token count exceeds limits
	t.Run("TokenCountExceedsLimits", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 3000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error for token count exceeding limits, got nil")
		}
	})

	// Test case 3: Error getting model info
	t.Run("ErrorGettingModelInfo", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, errors.New("model info error")
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error when getting model info fails, got nil")
		}
	})

	// Test case 4: Error counting tokens
	t.Run("ErrorCountingTokens", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return nil, errors.New("token counting error")
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error when counting tokens fails, got nil")
		}
	})
}
