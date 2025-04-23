// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"context"
	"errors"
	"testing"

	"github.com/openai/openai-go"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOpenAIAPI is a mock implementation of the openaiAPI interface for testing
type mockOpenAIAPI struct {
	createChatCompletionFunc           func(ctx context.Context, model string, prompt string, systemPrompt string) (*openai.ChatCompletion, error)
	createChatCompletionWithParamsFunc func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

func (m *mockOpenAIAPI) createChatCompletion(ctx context.Context, model string, prompt string, systemPrompt string) (*openai.ChatCompletion, error) {
	if m.createChatCompletionFunc != nil {
		return m.createChatCompletionFunc(ctx, model, prompt, systemPrompt)
	}
	return nil, errors.New("not implemented")
}

func (m *mockOpenAIAPI) createChatCompletionWithParams(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	if m.createChatCompletionWithParamsFunc != nil {
		return m.createChatCompletionWithParamsFunc(ctx, params)
	}
	return nil, errors.New("not implemented")
}

// TestExtractTag verifies the XML tag extraction function
func TestExtractTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tag      string
		expected string
	}{
		{"Basic tag", "<system>Hello</system>", "system", "Hello"},
		{"S tag", "<s>Hello</s>", "s", "Hello"},
		{"Nested content", "<system>Hello <world></system>", "system", "Hello <world>"},
		{"No matching tag", "Hello world", "system", ""},
		{"Incomplete tag", "<system>Hello", "system", ""},
		{"Multiple tags", "<system>First</system><system>Second</system>", "system", "First"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTag(tt.input, tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRemoveTag verifies the XML tag removal function
func TestRemoveTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tag      string
		expected string
	}{
		{"Basic tag", "<system>Hello</system>Content", "system", "Content"},
		{"S tag", "<s>Hello</s>Content", "s", "Content"},
		{"No matching tag", "Hello world", "system", "Hello world"},
		{"Incomplete tag", "<system>Hello", "system", "<system>Hello"},
		{"Multiple tags", "<system>First</system><system>Second</system>", "system", "<system>Second</system>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTag(tt.input, tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEmptyPromptError tests GenerateContent with an empty prompt
func TestEmptyPromptError(t *testing.T) {
	// Create a simple mock implementation that returns errors for empty prompts
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return nil, errors.New("mock not reached")
		},
	}

	client := &openaiClient{
		api:       mockAPI,
		modelName: "gpt-4",
	}

	// Call GenerateContent with empty prompt
	_, err := client.GenerateContent(context.Background(), "", nil)

	// Verify error is returned
	assert.Error(t, err)

	// Check error is correct type
	var llmErr *llm.LLMError
	require.True(t, errors.As(err, &llmErr))
	assert.Equal(t, llm.CategoryInvalidRequest, llmErr.ErrorCategory)
}

// TestSystemPromptExtraction tests the system prompt extraction functionality
func TestSystemPromptExtraction(t *testing.T) {
	// Test with standard system tag
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			// Check that we have exactly 2 messages (system + user)
			if len(params.Messages) != 2 {
				t.Errorf("Expected 2 messages, got %d", len(params.Messages))
			}

			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Mock response",
						},
					},
				},
			}, nil
		},
	}

	client := &openaiClient{
		api:       mockAPI,
		modelName: "gpt-4",
	}

	// Call with a prompt containing a system tag
	_, err := client.GenerateContent(context.Background(), "<system>System instruction</system>User prompt", nil)
	require.NoError(t, err)

	// Test with s tag
	_, err = client.GenerateContent(context.Background(), "<s>System instruction</s>User prompt", nil)
	require.NoError(t, err)
}

// TestParameterHandling tests parameter handling in the GenerateContent method
func TestParameterHandling(t *testing.T) {
	// Test that parameter settings are properly applied
	// We'll use simpler tests focused on the client setters

	// Test temperature setting
	client := &openaiClient{
		modelName: "gpt-4",
	}
	tempValue := float32(0.7)
	client.SetTemperature(tempValue)
	assert.Equal(t, &tempValue, client.temperature)

	// Test top_p setting
	topPValue := float32(0.95)
	client.SetTopP(topPValue)
	assert.Equal(t, &topPValue, client.topP)

	// Test max tokens setting
	maxTokensValue := int32(100)
	client.SetMaxTokens(maxTokensValue)
	assert.Equal(t, &maxTokensValue, client.maxTokens)

	// Test frequency penalty setting
	freqPenaltyValue := float32(0.5)
	client.SetFrequencyPenalty(freqPenaltyValue)
	assert.Equal(t, &freqPenaltyValue, client.frequencyPenalty)

	// Test presence penalty setting
	presPenaltyValue := float32(0.7)
	client.SetPresencePenalty(presPenaltyValue)
	assert.Equal(t, &presPenaltyValue, client.presencePenalty)

	// Test applyOpenAIParameters function
	params := &openai.ChatCompletionNewParams{}
	customParams := map[string]interface{}{
		"temperature":       0.7,
		"top_p":             0.95,
		"max_tokens":        100,
		"frequency_penalty": 0.5,
		"presence_penalty":  0.7,
	}

	// Apply the parameters
	applyOpenAIParameters(params, customParams)

	// Verify the parameters were set properly
	// We can't directly check the values, but we can verify the parameters were changed
	// from their default nil state by the applyOpenAIParameters function
	assert.NotNil(t, params.Temperature)
	assert.NotNil(t, params.TopP)
	assert.NotNil(t, params.MaxTokens)
	assert.NotNil(t, params.FrequencyPenalty)
	assert.NotNil(t, params.PresencePenalty)
}

// TestSuccessfulGeneration tests a successful content generation
func TestSuccessfulGeneration(t *testing.T) {
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Generated response",
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	client := &openaiClient{
		api:       mockAPI,
		modelName: "gpt-4",
	}

	// Call GenerateContent
	result, err := client.GenerateContent(context.Background(), "Test prompt", nil)

	// Verify success
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Generated response", result.Content)
	assert.Equal(t, "stop", result.FinishReason)
	assert.False(t, result.Truncated)
}

// TestTruncatedResponse tests response truncation detection
func TestTruncatedResponse(t *testing.T) {
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Truncated response...",
						},
						FinishReason: "length",
					},
				},
			}, nil
		},
	}

	client := &openaiClient{
		api:       mockAPI,
		modelName: "gpt-4",
	}

	// Call GenerateContent
	result, err := client.GenerateContent(context.Background(), "Test prompt", nil)

	// Verify success but with truncation
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Truncated response...", result.Content)
	assert.Equal(t, "length", result.FinishReason)
	assert.True(t, result.Truncated)
}

// TestEmptyChoicesError tests empty choices error handling
func TestEmptyChoicesError(t *testing.T) {
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
			}, nil
		},
	}

	client := &openaiClient{
		api:       mockAPI,
		modelName: "gpt-4",
	}

	// Call GenerateContent
	result, err := client.GenerateContent(context.Background(), "Test prompt", nil)

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no completions returned")
}

// TestAPIError tests API error handling
func TestAPIError(t *testing.T) {
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return nil, errors.New("API error")
		},
	}

	client := &openaiClient{
		api:       mockAPI,
		modelName: "gpt-4",
	}

	// Call GenerateContent
	result, err := client.GenerateContent(context.Background(), "Test prompt", nil)

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "API error")
}
