package gemini

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeminiClientImplementsLLMAdapter tests that geminiLLMAdapter correctly adapts Client methods
func TestGeminiClientImplementsLLMAdapter(t *testing.T) {
	// Create a mock Gemini client
	mockClient := &mockGeminiClient{
		generateResult: &GenerationResult{
			Content:       "Test content",
			FinishReason:  "stop",
			SafetyRatings: []SafetyRating{{Category: "test", Blocked: false, Score: 0.1}},
			TokenCount:    10,
			Truncated:     false,
		},
		countTokensResult: &TokenCount{Total: 5},
		modelInfoResult: &ModelInfo{
			Name:             "gemini-model",
			InputTokenLimit:  1000,
			OutputTokenLimit: 500,
		},
	}

	// Get the adapter
	adapter := AsLLMClient(mockClient)

	// Verify the adapter is not nil
	assert.NotNil(t, adapter)

	// Test interface method implementations
	ctx := context.Background()

	// Test GenerateContent
	result, err := adapter.GenerateContent(ctx, "test prompt", nil)
	require.NoError(t, err)
	assert.Equal(t, "Test content", result.Content)
	assert.Equal(t, "stop", result.FinishReason)
	assert.Equal(t, int32(10), result.TokenCount)
	assert.False(t, result.Truncated)
	assert.Len(t, result.SafetyInfo, 1)
	assert.Equal(t, "test", result.SafetyInfo[0].Category)

	// Test CountTokens
	tokenCount, err := adapter.CountTokens(ctx, "test prompt")
	require.NoError(t, err)
	assert.Equal(t, int32(5), tokenCount.Total)

	// Test GetModelInfo
	modelInfo, err := adapter.GetModelInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "gemini-model", modelInfo.Name)
	assert.Equal(t, int32(1000), modelInfo.InputTokenLimit)
	assert.Equal(t, int32(500), modelInfo.OutputTokenLimit)

	// Test GetModelName
	assert.Equal(t, "gemini-model", adapter.GetModelName())

	// Test Close
	assert.NoError(t, adapter.Close())
}

// Test that the constructor function correctly returns the adapter
func TestAsLLMClient(t *testing.T) {
	// Create a mock Gemini client
	mockClient := &mockGeminiClient{}

	// Test using the constructor function
	adapter := AsLLMClient(mockClient)

	// Verify that we get a non-nil adapter
	assert.NotNil(t, adapter)

	// Check that it uses the right client
	geminiAdapter, ok := adapter.(*geminiLLMAdapter)
	assert.True(t, ok)
	assert.Equal(t, mockClient, geminiAdapter.client)
}

// mockGeminiClient is a mock implementation of the gemini.Client interface
type mockGeminiClient struct {
	generateResult    *GenerationResult
	generateErr       error
	countTokensResult *TokenCount
	countTokensErr    error
	modelInfoResult   *ModelInfo
	modelInfoErr      error
}

// Implement gemini.Client interface
func (m *mockGeminiClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*GenerationResult, error) {
	return m.generateResult, m.generateErr
}

func (m *mockGeminiClient) CountTokens(ctx context.Context, prompt string) (*TokenCount, error) {
	return m.countTokensResult, m.countTokensErr
}

func (m *mockGeminiClient) GetModelInfo(ctx context.Context) (*ModelInfo, error) {
	return m.modelInfoResult, m.modelInfoErr
}

func (m *mockGeminiClient) GetModelName() string {
	if m.modelInfoResult != nil {
		return m.modelInfoResult.Name
	}
	return "mock-model"
}

func (m *mockGeminiClient) GetTemperature() float32 {
	return 0.5
}

func (m *mockGeminiClient) GetMaxOutputTokens() int32 {
	return 1000
}

func (m *mockGeminiClient) GetTopP() float32 {
	return 0.9
}

func (m *mockGeminiClient) Close() error {
	return nil
}
