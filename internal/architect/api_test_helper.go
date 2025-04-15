package architect

import (
	"context"
	"errors"

	"github.com/phrazzld/architect/internal/llm"
)

// newGeminiClientWrapperForTest adapts the original Gemini client creation function to return llm.LLMClient
// We use this version specifically for tests to avoid importing the actual gemini package
func newGeminiClientWrapperForTest(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// Mock implementation for tests
	if apiKey == "error-key" {
		return nil, errors.New("test API key error")
	}

	if modelName == "error-model" {
		return nil, errors.New("test model error")
	}

	// Log endpoint for tests that verify it was passed through
	if apiEndpoint != "" {
		// Make this something the test can check for
		// We'll use the modelName to include the endpoint info
		return &testLLMClient{
			modelName: modelName + " with endpoint " + apiEndpoint,
		}, nil
	}

	// Create a test client that implements llm.LLMClient
	client := &testLLMClient{
		modelName: modelName,
	}
	return client, nil
}

// newOpenAIClientWrapperForTest adapts the OpenAI client creation function to handle test cases
func newOpenAIClientWrapperForTest(modelName string) (llm.LLMClient, error) {
	// Mock implementation for tests
	if modelName == "error-model" {
		return nil, errors.New("test model error")
	}

	// Create a test client that implements llm.LLMClient
	client := &testLLMClient{
		modelName: modelName,
	}
	return client, nil
}

// testLLMClient is a test implementation of llm.LLMClient
type testLLMClient struct {
	modelName string
}

func (c *testLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:    "test content",
		TokenCount: 10,
	}, nil
}

func (c *testLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return &llm.ProviderTokenCount{
		Total: 10,
	}, nil
}

func (c *testLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	return &llm.ProviderModelInfo{
		Name:             c.modelName,
		InputTokenLimit:  1000,
		OutputTokenLimit: 1000,
	}, nil
}

func (c *testLLMClient) GetModelName() string {
	return c.modelName
}

func (c *testLLMClient) Close() error {
	return nil
}

// IsModelSupported checks if a model is supported by any provider
func IsModelSupported(modelName string) bool {
	return DetectProviderFromModel(modelName) != ProviderUnknown
}
