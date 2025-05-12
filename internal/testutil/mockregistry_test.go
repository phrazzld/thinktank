package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// Mock provider and client implementations for testing

// MockProvider implements a test Provider
type mockTestProvider struct {
	ClientResult llm.LLMClient
	ClientError  error
	Calls        []mockProviderCall
}

type mockProviderCall struct {
	Ctx         context.Context
	APIKey      string
	ModelID     string
	APIEndpoint string
}

func (m *mockTestProvider) CreateClient(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error) {
	m.Calls = append(m.Calls, mockProviderCall{
		Ctx:         ctx,
		APIKey:      apiKey,
		ModelID:     modelID,
		APIEndpoint: apiEndpoint,
	})

	if m.ClientError != nil {
		return nil, m.ClientError
	}
	return m.ClientResult, nil
}

// MockClient implements a test LLM client
type mockTestClient struct {
	ModelName string
}

func (m *mockTestClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content: "mock response",
	}, nil
}

func (m *mockTestClient) GetModelName() string {
	return m.ModelName
}

func (m *mockTestClient) Close() error {
	return nil
}

// TestMockRegistryBasics tests basic functionality of the MockRegistry
func TestMockRegistryBasics(t *testing.T) {
	// Create a new registry
	registry := NewMockRegistry()

	// Check initialization
	if registry.methodCalls == nil {
		t.Error("methodCalls map not initialized")
	}
	if registry.callSequence == nil {
		t.Error("callSequence slice not initialized")
	}

	// Test call tracking
	registry.recordCall("TestMethod", "arg1", "arg2")
	calls := registry.GetMethodCalls("TestMethod")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	// Test call sequence
	sequence := registry.GetCallSequence()
	if len(sequence) != 1 || sequence[0] != "TestMethod" {
		t.Errorf("Expected call sequence [TestMethod], got %v", sequence)
	}

	// Test clearing calls
	registry.ClearMethodCalls()
	calls = registry.GetMethodCalls("TestMethod")
	if len(calls) != 0 {
		t.Errorf("Expected 0 calls after clear, got %d", len(calls))
	}
}

// TestMockRegistryCallInterception tests that the MockRegistry properly
// intercepts and records method calls.
func TestMockRegistryCallInterception(t *testing.T) {
	// Create a new registry
	registry := NewMockRegistry()

	// Set error for GetModel method
	expectedErr := errors.New("test error")
	registry.SetGetModelError(expectedErr)

	// Call the method
	ctx := context.Background()
	_, err := registry.GetModel(ctx, "test-model")

	// Verify error is returned
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Verify call was recorded
	calls := registry.GetMethodCalls("GetModel")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call to GetModel, got %d", len(calls))
	} else if calls[0].Args[1] != "test-model" { // Arg 0 is context, arg 1 is model name
		t.Errorf("Expected arg 'test-model', got %v", calls[0].Args[1])
	}
}

// TestMockRegistryGetProviderImplementation tests the GetProviderImplementation method
func TestMockRegistryGetProviderImplementation(t *testing.T) {
	// Create a new registry
	registry := NewMockRegistry()

	// Test with error
	expectedErr := errors.New("implementation not found")
	registry.SetGetProviderImplementationError(expectedErr)

	ctx := context.Background()
	_, err := registry.GetProviderImplementation(ctx, "test-provider")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Set up custom implementation
	provider := &mockTestProvider{}
	registry.AddProvider(testProviderDefinition{Name: "test-provider"})
	ctx2 := context.Background()
	registry.AddProviderImplementation(ctx2, "test-provider", provider)
	registry.SetGetProviderImplementationError(nil) // Clear error

	// Get implementation
	ctx3 := context.Background()
	impl, err := registry.GetProviderImplementation(ctx3, "test-provider")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if impl != provider {
		t.Error("Retrieved provider not the same as added provider")
	}

	// Verify call was recorded
	calls := registry.GetMethodCalls("GetProviderImplementation")
	if len(calls) != 2 { // We called it twice
		t.Errorf("Expected 2 calls to GetProviderImplementation, got %d", len(calls))
	}
}

// TestMockRegistryChain tests a chain of method calls that depend on each other
func TestMockRegistryChain(t *testing.T) {
	// Create a new registry
	registry := NewMockRegistry()
	ctx := context.Background()

	// Set up the registry with model, provider, and implementation
	registry.AddModel(testModelDefinition{
		Name:       "test-model",
		Provider:   "test-provider",
		APIModelID: "test-api-id",
	})

	registry.AddProvider(testProviderDefinition{
		Name:    "test-provider",
		BaseURL: "https://test.example.com",
	})

	mockClient := &mockTestClient{ModelName: "test-model"}
	mockProvider := &mockTestProvider{ClientResult: mockClient}
	registry.AddProviderImplementation("test-provider", mockProvider)

	// Call CreateLLMClient - this should chain through GetModel, GetProvider, GetProviderImplementation
	client, err := registry.CreateLLMClient(ctx, "test-api-key", "test-model")
	if err != nil {
		t.Errorf("Unexpected error creating client: %v", err)
	}
	if client != mockClient {
		t.Error("Retrieved client not the same as mock client")
	}

	// Verify call sequence
	sequence := registry.GetCallSequence()
	expectedMethods := []string{"CreateLLMClient", "GetModel", "GetProvider", "GetProviderImplementation"}

	// Verify expected method calls are in the sequence (order may vary due to implementation details)
	for _, method := range expectedMethods {
		found := false
		for _, call := range sequence {
			if call == method {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected method %s not found in call sequence", method)
		}
	}

	// Verify provider was called with correct parameters
	if len(mockProvider.Calls) != 1 {
		t.Errorf("Expected 1 call to provider, got %d", len(mockProvider.Calls))
	} else {
		call := mockProvider.Calls[0]
		if call.APIKey != "test-api-key" {
			t.Errorf("Expected API key 'test-api-key', got '%s'", call.APIKey)
		}
		if call.ModelID != "test-api-id" {
			t.Errorf("Expected model ID 'test-api-id', got '%s'", call.ModelID)
		}
		if call.APIEndpoint != "https://test.example.com" {
			t.Errorf("Expected endpoint 'https://test.example.com', got '%s'", call.APIEndpoint)
		}
	}
}
