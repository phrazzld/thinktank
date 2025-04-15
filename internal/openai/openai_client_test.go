package openai

import (
	"context"
	"os"
	"testing"

	"github.com/openai/openai-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define mocks for our internal interfaces
type mockOpenAIAPI struct {
	createChatCompletionFunc func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error)
}

func (m *mockOpenAIAPI) createChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
	return m.createChatCompletionFunc(ctx, messages, model)
}

type mockTokenizer struct {
	countTokensFunc func(text string, model string) (int, error)
}

func (m *mockTokenizer) countTokens(text string, model string) (int, error) {
	return m.countTokensFunc(text, model)
}

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
		result, err := client.GenerateContent(ctx, "test prompt")
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

// TestNewClient tests the NewClient constructor function
func TestNewClient(t *testing.T) {
	t.Skip("Skipping test that relies on environment variables")
	// This test would check that NewClient correctly sets up the client
	// It's skipped here since it would require setting up environment variables
}

// TestOpenAIClientErrorHandling tests how the OpenAI client handles API errors
func TestOpenAIClientErrorHandling(t *testing.T) {
	// Create a mock OpenAI API that returns an error
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return nil, &APIError{
				Type:    ErrorTypeRateLimit,
				Message: "Rate limit exceeded",
			}
		},
	}

	// Create a mock tokenizer that returns an error
	mockTokenizer := &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			return 0, &APIError{
				Type:    ErrorTypeInvalidRequest,
				Message: "Invalid request",
			}
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: mockTokenizer,
		modelName: "gpt-4",
	}

	ctx := context.Background()

	// Test GenerateContent error handling
	t.Run("GenerateContent error", func(t *testing.T) {
		_, err := client.GenerateContent(ctx, "test prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rate limit exceeded")
	})

	// Test CountTokens error handling
	t.Run("CountTokens error", func(t *testing.T) {
		_, err := client.CountTokens(ctx, "test prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid request")
	})
}

// TestUnknownModelFallback tests how the client handles unknown models
func TestUnknownModelFallback(t *testing.T) {
	// Create mock API
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Test content",
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	// Create a client with an unknown model
	client := &openaiClient{
		api:         mockAPI,
		tokenizer:   &mockTokenizer{},
		modelName:   "unknown-model",
		modelLimits: map[string]*modelInfo{},
	}

	ctx := context.Background()

	// Test GetModelInfo with unknown model
	t.Run("GetModelInfo with unknown model", func(t *testing.T) {
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "unknown-model", modelInfo.Name)
		// Should return default values for unknown models
		assert.True(t, modelInfo.InputTokenLimit > 0)
		assert.True(t, modelInfo.OutputTokenLimit > 0)
	})
}

// TestTruncatedResponse tests how the client handles truncated responses
func TestTruncatedResponse(t *testing.T) {
	// Create mock API that returns a response with "length" finish reason
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Truncated content",
							Role:    "assistant",
						},
						FinishReason: "length",
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 100,
					TotalTokens:      110,
				},
			}, nil
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
		modelLimits: map[string]*modelInfo{
			"gpt-4": {
				inputTokenLimit:  8192,
				outputTokenLimit: 2048,
			},
		},
	}

	ctx := context.Background()

	// Test truncated response
	result, err := client.GenerateContent(ctx, "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "Truncated content", result.Content)
	assert.Equal(t, "length", result.FinishReason)
	assert.Equal(t, int32(100), result.TokenCount)
	assert.True(t, result.Truncated)
}

// TestEmptyResponseHandling tests how the client handles empty responses
func TestEmptyResponseHandling(t *testing.T) {
	// Create mock API that returns an empty response
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 0,
					TotalTokens:      10,
				},
			}, nil
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
	}

	ctx := context.Background()

	// Test empty response handling
	_, err := client.GenerateContent(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no completion choices returned")
}

// TestContentFilterHandling tests handling of content filter errors
func TestContentFilterHandling(t *testing.T) {
	// Create mock API that returns a content filter error
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return nil, &APIError{
				Type:    ErrorTypeContentFiltered,
				Message: "Content was filtered",
			}
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
	}

	ctx := context.Background()

	// Test content filter handling
	_, err := client.GenerateContent(ctx, "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Content was filtered")
}

// TestModelEncodingSelection tests the getEncodingForModel function
func TestModelEncodingSelection(t *testing.T) {
	tests := []struct {
		modelName        string
		expectedEncoding string
	}{
		{"gpt-4", "cl100k_base"},
		{"gpt-4-32k", "cl100k_base"},
		{"gpt-4-turbo", "cl100k_base"},
		{"gpt-4o", "cl100k_base"},
		{"gpt-3.5-turbo", "cl100k_base"},
		{"gpt-3.5-turbo-16k", "cl100k_base"},
		{"text-embedding-ada-002", "cl100k_base"},
		{"text-davinci-003", "p50k_base"}, // Older model should use p50k_base
		{"unknown-model", "p50k_base"},    // Unknown models should use p50k_base
	}

	for _, test := range tests {
		t.Run(test.modelName, func(t *testing.T) {
			encoding := getEncodingForModel(test.modelName)
			assert.Equal(t, test.expectedEncoding, encoding)
		})
	}
}

// TestNewClientErrorHandling tests error handling in NewClient
func TestNewClientErrorHandling(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test with empty API key
	err := os.Unsetenv("OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("Failed to unset OPENAI_API_KEY: %v", err)
	}
	client, err := NewClient("gpt-4")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY environment variable not set")

	// Set an invalid API key (too short)
	err = os.Setenv("OPENAI_API_KEY", "invalid-key")
	if err != nil {
		t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
	}
	client, err = NewClient("gpt-4")
	// This should succeed since we're just creating the client (error would occur on API calls)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
