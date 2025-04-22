// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"context"
	"testing"

	"github.com/openai/openai-go"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/stretchr/testify/assert"
)

// TestOpenAIClientImplementsLLMClient verifies that the OpenAI client implementation
// satisfies the LLMClient interface required by the architect framework
func TestOpenAIClientImplementsLLMClient(t *testing.T) {
	// Verify at compile time that openaiClient implements the LLMClient interface
	var _ llm.LLMClient = (*openaiClient)(nil)

	// Create a mock API and tokenizer
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
					CompletionTokens: 5,
				},
			}, nil
		},
	}

	mockTokenizer := &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			return 10, nil
		},
	}

	// Create an openaiClient instance manually
	client := &openaiClient{
		api:         mockAPI,
		tokenizer:   mockTokenizer,
		modelName:   "gpt-4",
		modelLimits: make(map[string]*modelInfo),
	}

	// Verify the client implements the interface at runtime
	var llmClient llm.LLMClient = client
	assert.NotNil(t, llmClient, "LLM client should not be nil")

	// Type assert to original type to verify it's the same object
	openaiClient, ok := llmClient.(*openaiClient)
	assert.True(t, ok, "Client should be of type *openaiClient")
	assert.Equal(t, "gpt-4", openaiClient.modelName, "Model name should match")

	// Test some basic functionality
	modelName := client.GetModelName()
	assert.Equal(t, "gpt-4", modelName, "GetModelName should return the correct model name")

	// Verify the client can be used to generate content
	ctx := context.Background()
	result, err := client.GenerateContent(ctx, "Test prompt", nil)
	assert.NoError(t, err, "GenerateContent should not return an error")
	assert.Equal(t, "Test content", result.Content, "Result should contain expected content")
}

// Verify at compile time that openaiClient implements the LLMClient interface
var _ llm.LLMClient = (*openaiClient)(nil)
