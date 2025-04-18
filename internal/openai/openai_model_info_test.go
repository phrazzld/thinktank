// Package openai provides an implementation of the LLM client for the OpenAI API
package openai

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetModelInfo verifies that the client's GetModelInfo method
// correctly retrieves model information (token limits, etc.)
func TestGetModelInfo(t *testing.T) {
	// Test context
	ctx := context.Background()

	// Create a client with known model limits
	client := &openaiClient{
		modelName: "gpt-4",
		modelLimits: map[string]*modelInfo{
			"gpt-4": {
				inputTokenLimit:  8192,
				outputTokenLimit: 2048,
			},
		},
	}

	// Get model info
	modelInfo, err := client.GetModelInfo(ctx)

	// Verify results
	require.NoError(t, err, "GetModelInfo should not return an error")
	require.NotNil(t, modelInfo, "Model info should not be nil")
	assert.Equal(t, "gpt-4", modelInfo.Name, "Model name should match")
	assert.Equal(t, int32(8192), modelInfo.InputTokenLimit, "Input token limit should match")
	assert.Equal(t, int32(2048), modelInfo.OutputTokenLimit, "Output token limit should match")
}

// TestGetModelName verifies that the client's GetModelName method
// correctly returns the model name
func TestGetModelName(t *testing.T) {
	// Create a client with a specific model name
	client := &openaiClient{
		modelName: "gpt-4",
	}

	// Get model name
	modelName := client.GetModelName()

	// Verify result
	assert.Equal(t, "gpt-4", modelName, "GetModelName should return the correct model name")
}

// TestGetModelInfoWithUnknownModel tests how GetModelInfo handles unknown models
func TestGetModelInfoWithUnknownModel(t *testing.T) {
	// Test context
	ctx := context.Background()

	// Create a client with a model name that isn't in the model limits map
	unknownModelName := "unknown-model"
	client := &openaiClient{
		modelName:   unknownModelName,
		modelLimits: map[string]*modelInfo{},
	}

	// Get model info
	modelInfo, err := client.GetModelInfo(ctx)

	// Verify it falls back to defaults
	require.NoError(t, err, "GetModelInfo should not return an error for unknown model")
	require.NotNil(t, modelInfo, "Model info should not be nil")
	assert.Equal(t, unknownModelName, modelInfo.Name, "Model name should match")
	// Should use more generous defaults for unknown models (updated values)
	assert.Equal(t, int32(200000), modelInfo.InputTokenLimit, "Input token limit should use default")
	assert.Equal(t, int32(4096), modelInfo.OutputTokenLimit, "Output token limit should use default")
}

// TestGetModelInfoIntegrationWithMockProvider tests the integration between
// the client and the model info provider
func TestGetModelInfoIntegrationWithMockProvider(t *testing.T) {
	ctx := context.Background()
	testModel := "gpt-4"

	t.Run("Fixed model info using mock provider", func(t *testing.T) {
		// Create mock provider that returns fixed model info
		fixedInputLimit := int32(10000)
		fixedOutputLimit := int32(2000)

		mockProvider := MockModelInfo(fixedInputLimit, fixedOutputLimit, nil)

		// Get model info using the mock
		info, err := mockProvider.getModelInfo(ctx, testModel)

		// Verify mock provider works as expected
		require.NoError(t, err, "getModelInfo should not return an error")
		require.NotNil(t, info, "Model info should not be nil")
		assert.Equal(t, fixedInputLimit, info.inputTokenLimit, "Input token limit should match")
		assert.Equal(t, fixedOutputLimit, info.outputTokenLimit, "Output token limit should match")
	})

	t.Run("Error handling with mock provider", func(t *testing.T) {
		// Create mock provider that returns an error
		expectedError := errors.New("model info retrieval error")
		mockProvider := MockModelInfo(0, 0, expectedError)

		// Try to get model info
		info, err := mockProvider.getModelInfo(ctx, testModel)

		// Verify error is returned
		require.Error(t, err, "getModelInfo should return an error")
		assert.Equal(t, expectedError, err, "Error should match expected")
		assert.Nil(t, info, "Model info should be nil on error")
	})
}
