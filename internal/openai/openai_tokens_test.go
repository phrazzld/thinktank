// Package openai provides an implementation of the LLM client for the OpenAI API
package openai

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCountTokens verifies that the client's CountTokens method
// correctly counts tokens in a given prompt
func TestCountTokens(t *testing.T) {
	t.Skip("TODO: Update this test for the new llm.LLMClient token interface")
	// Test context
	ctx := context.Background()

	// Expected token count
	expectedCount := int32(42)

	// Create a client with a mock tokenizer
	client := &openaiClient{
		modelName: "gpt-4",
		tokenizer: &mockTokenizer{
			countTokensFunc: func(text string, model string) (int, error) {
				// Verify the model name is passed correctly
				assert.Equal(t, "gpt-4", model, "Model should match client's model name")
				// Verify the prompt text is passed correctly
				assert.Equal(t, "Test prompt", text, "Text should match the prompt")
				return int(expectedCount), nil
			},
		},
	}

	// Count tokens
	result, err := client.CountTokens(ctx, "Test prompt")

	// Verify results
	require.NoError(t, err, "CountTokens should not return an error")
	assert.Equal(t, expectedCount, result, "Token count should match")
}

// TestCountTokensWithError tests how CountTokens handles errors from the tokenizer
func TestCountTokensWithError(t *testing.T) {
	// Test context
	ctx := context.Background()

	// Expected error
	expectedError := errors.New("tokenizer error")

	// Create a client with a mock tokenizer that returns an error
	client := &openaiClient{
		modelName: "gpt-4",
		tokenizer: &mockTokenizer{
			countTokensFunc: func(text string, model string) (int, error) {
				return 0, expectedError
			},
		},
	}

	// Count tokens, expecting error
	result, err := client.CountTokens(ctx, "Test prompt")

	// Verify error handling
	require.Error(t, err, "CountTokens should return an error")
	assert.Contains(t, err.Error(), "token counting error", "Error should indicate token counting failed")
	assert.Nil(t, result, "Token count result should be nil on error")
}

// TestModelEncodingSelection tests the selection of the correct tokenizer encoding for different models
func TestModelEncodingSelection(t *testing.T) {
	testCases := []struct {
		name             string
		modelName        string
		expectedEncoding string
	}{
		{"GPT-4 Model", "gpt-4", "cl100k_base"},
		{"GPT-4 model (lowercase)", "gpt-4", "cl100k_base"},
		{"GPT-4 Turbo", "gpt-4-turbo", "cl100k_base"},
		{"GPT-3.5 Turbo", "gpt-3.5-turbo", "cl100k_base"},
		{"Text Embedding", "text-embedding-ada-002", "cl100k_base"},
		{"Unknown Model", "unknown-model", "cl100k_base"}, // Now all models use cl100k_base
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoding := getEncodingForModel(tc.modelName)
			assert.Equal(t, tc.expectedEncoding, encoding, "Encoding should match expected for model")
		})
	}
}
