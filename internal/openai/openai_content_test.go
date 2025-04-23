// Package openai provides a mock implementation of the OpenAI client for testing
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
