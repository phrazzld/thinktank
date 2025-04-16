// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"context"
	"testing"

	"github.com/openai/openai-go"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAIClientImplementsLLMClient tests that openaiClient correctly implements the LLMClient interface
func TestOpenAIClientImplementsLLMClient(t *testing.T) {
	// Create a mock OpenAI API
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Test content",
							Role:    "assistant",
						},
						FinishReason: "stop",
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			}, nil
		},
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			// Use the same response format as createChatCompletionFunc for consistency
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Test content",
							Role:    "assistant",
						},
						FinishReason: "stop",
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			}, nil
		},
	}

	// Create a mock tokenizer
	mockTokenizer := &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			return 10, nil
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: mockTokenizer,
		modelName: "gpt-4",
		modelLimits: map[string]*modelInfo{
			"gpt-4": {
				inputTokenLimit:  8192,
				outputTokenLimit: 4096,
			},
		},
	}

	// Test interface method implementations
	ctx := context.Background()

	// Test GenerateContent
	t.Run("GenerateContent", func(t *testing.T) {
		result, err := client.GenerateContent(ctx, "test prompt", nil)
		require.NoError(t, err)
		assert.Equal(t, "Test content", result.Content)
		assert.Equal(t, "stop", result.FinishReason)
		assert.Equal(t, int32(5), result.TokenCount)
		assert.False(t, result.Truncated)
	})

	// Test CountTokens
	t.Run("CountTokens", func(t *testing.T) {
		tokenCount, err := client.CountTokens(ctx, "test prompt")
		require.NoError(t, err)
		assert.Equal(t, int32(10), tokenCount.Total)
	})

	// Test GetModelInfo
	t.Run("GetModelInfo", func(t *testing.T) {
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", modelInfo.Name)
		assert.Equal(t, int32(8192), modelInfo.InputTokenLimit)
		assert.Equal(t, int32(4096), modelInfo.OutputTokenLimit)
	})

	// Test GetModelName
	t.Run("GetModelName", func(t *testing.T) {
		assert.Equal(t, "gpt-4", client.GetModelName())
	})

	// Test Close
	t.Run("Close", func(t *testing.T) {
		assert.NoError(t, client.Close())
	})
}

// Verify at compile time that openaiClient implements the LLMClient interface
var _ llm.LLMClient = (*openaiClient)(nil)