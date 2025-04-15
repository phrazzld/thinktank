package orchestrator

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
)

// TestAPIServiceAdapter_InitLLMClient tests that APIServiceAdapter correctly delegates to the
// underlying APIService.InitLLMClient method.
func TestAPIServiceAdapter_InitLLMClient(t *testing.T) {
	// Setup a mock APIService that tracks calls to InitLLMClient
	mockService := &mockAPIService{}

	// Define expected values
	expectedApiKey := "test-api-key"
	expectedModelName := "test-model"
	expectedApiEndpoint := "test-endpoint"
	expectedClient := &mockLLMClient{modelName: expectedModelName}

	// Configure the mock's InitLLMClient method to return the expected client
	mockService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		// Verify the arguments match what we expect
		if apiKey != expectedApiKey {
			t.Errorf("Expected apiKey %q, got %q", expectedApiKey, apiKey)
		}
		if modelName != expectedModelName {
			t.Errorf("Expected modelName %q, got %q", expectedModelName, modelName)
		}
		if apiEndpoint != expectedApiEndpoint {
			t.Errorf("Expected apiEndpoint %q, got %q", expectedApiEndpoint, apiEndpoint)
		}
		return expectedClient, nil
	}

	// Create the adapter with the mock service
	adapter := &APIServiceAdapter{APIService: mockService}

	// Call the method being tested
	client, err := adapter.InitLLMClient(context.Background(), expectedApiKey, expectedModelName, expectedApiEndpoint)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the client returned is the expected one
	if client != expectedClient {
		t.Errorf("Expected client %v, got %v", expectedClient, client)
	}

	// Verify the method was called
	if len(mockService.InitLLMClientCalls) != 1 {
		t.Errorf("Expected 1 call to InitLLMClient, got %d", len(mockService.InitLLMClientCalls))
	}
}

// TestAPIServiceAdapter_ProcessLLMResponse tests that APIServiceAdapter correctly delegates to the
// underlying APIService.ProcessLLMResponse method.
func TestAPIServiceAdapter_ProcessLLMResponse(t *testing.T) {
	// Setup a mock APIService that tracks calls to ProcessLLMResponse
	mockService := &mockAPIService{}

	// Define expected values
	expectedContent := "test content"
	expectedResult := &llm.ProviderResult{
		Content:    expectedContent,
		TokenCount: 100,
	}

	// Configure the mock's ProcessLLMResponse method to return the expected content
	mockService.ProcessLLMResponseFunc = func(result *llm.ProviderResult) (string, error) {
		// Verify the result matches what we expect
		if result.Content != expectedResult.Content {
			t.Errorf("Expected content %q, got %q", expectedResult.Content, result.Content)
		}
		if result.TokenCount != expectedResult.TokenCount {
			t.Errorf("Expected tokenCount %d, got %d", expectedResult.TokenCount, result.TokenCount)
		}
		return expectedContent, nil
	}

	// Create the adapter with the mock service
	adapter := &APIServiceAdapter{APIService: mockService}

	// Call the method being tested
	content, err := adapter.ProcessLLMResponse(expectedResult)

	// Verify no error occurred
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the content returned is the expected one
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}

	// Verify the method was called
	if len(mockService.ProcessLLMResponseCalls) != 1 {
		t.Errorf("Expected 1 call to ProcessLLMResponse, got %d", len(mockService.ProcessLLMResponseCalls))
	}
}

// mockLLMClient is a mock implementation of llm.LLMClient for testing
type mockLLMClient struct {
	modelName           string
	tokenResponse       *llm.ProviderTokenCount
	modelInfo           *llm.ProviderModelInfo
	generateContentFunc func(ctx context.Context, prompt string) (*llm.ProviderResult, error)
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt)
	}
	return &llm.ProviderResult{
		Content:      "Generated content for: " + prompt,
		TokenCount:   100,
		FinishReason: "STOP",
	}, nil
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.tokenResponse != nil {
		return m.tokenResponse, nil
	}
	return &llm.ProviderTokenCount{
		Total: 100,
	}, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.modelInfo != nil {
		return m.modelInfo, nil
	}
	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  1000,
		OutputTokenLimit: 1000,
	}, nil
}

func (m *mockLLMClient) GetModelName() string {
	return m.modelName
}

func (m *mockLLMClient) Close() error {
	return nil
}
