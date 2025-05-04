// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// MockRegistryAPI is a mock implementation of the Registry API for testing
type MockRegistryAPI struct {
	models             map[string]*registry.ModelDefinition
	providers          map[string]*registry.ProviderDefinition
	implementations    map[string]MockProviderAPI
	getModelErr        error
	getProviderErr     error
	getProviderImplErr error
}

// MockProviderAPI implements a provider for testing
type MockProviderAPI struct {
	createClientErr error
	client          llm.LLMClient
}

// Use the built-in MockLLMClient from the llm package instead of creating our own

// MockAPIConfigLoader implements registry.ConfigLoaderInterface for testing
type MockAPIConfigLoader struct {
	config  *registry.ModelsConfig
	loadErr error
}

// Load implements registry.ConfigLoaderInterface
func (m *MockAPIConfigLoader) Load() (*registry.ModelsConfig, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.config, nil
}

// Placeholder for future use when testing API key resolution
// var originalConfigLoader func() registry.ConfigLoaderInterface

// CreateClient implements the Provider interface
func (m MockProviderAPI) CreateClient(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error) {
	if m.createClientErr != nil {
		return nil, m.createClientErr
	}
	return m.client, nil
}

// GetModel implements the registry.Registry method
func (m *MockRegistryAPI) GetModel(name string) (*registry.ModelDefinition, error) {
	if m.getModelErr != nil {
		return nil, m.getModelErr
	}
	model, ok := m.models[name]
	if !ok {
		return nil, errors.New("model not found")
	}
	return model, nil
}

// GetProvider implements the registry.Registry method
func (m *MockRegistryAPI) GetProvider(name string) (*registry.ProviderDefinition, error) {
	if m.getProviderErr != nil {
		return nil, m.getProviderErr
	}
	provider, ok := m.providers[name]
	if !ok {
		return nil, errors.New("provider not found")
	}
	return provider, nil
}

// GetProviderImplementation implements the registry.Registry method
func (m *MockRegistryAPI) GetProviderImplementation(name string) (interface{}, error) {
	if m.getProviderImplErr != nil {
		return nil, m.getProviderImplErr
	}
	impl, ok := m.implementations[name]
	if !ok {
		return nil, errors.New("provider implementation not found")
	}
	return impl, nil
}

// setupTest creates test fixtures for each test case
func setupTest(t *testing.T) (*registryAPIService, *MockRegistryAPI, *testutil.MockLogger) {
	t.Helper()
	logger := testutil.NewMockLogger()
	mockRegistry := &MockRegistryAPI{
		models:          make(map[string]*registry.ModelDefinition),
		providers:       make(map[string]*registry.ProviderDefinition),
		implementations: make(map[string]MockProviderAPI),
	}

	// Add a test model
	mockRegistry.models["test-model"] = &registry.ModelDefinition{
		Name:       "test-model",
		Provider:   "test-provider",
		APIModelID: "test-model-id",
		Parameters: map[string]registry.ParameterDefinition{
			"temperature": {
				Type:    "float",
				Default: 0.7,
				Min:     0.0,
				Max:     1.0,
			},
			"max_tokens": {
				Type:    "int",
				Default: 1024,
				Min:     1,
				Max:     4096,
			},
			"model_type": {
				Type:       "string",
				Default:    "default",
				EnumValues: []string{"default", "fast", "creative"},
			},
		},
	}

	// Add a test provider
	mockRegistry.providers["test-provider"] = &registry.ProviderDefinition{
		Name:    "test-provider",
		BaseURL: "https://test-api.example.com",
	}

	// Add a test provider implementation
	mockRegistry.implementations["test-provider"] = MockProviderAPI{
		client: &llm.MockLLMClient{
			GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
				return &llm.ProviderResult{Content: "Test response"}, nil
			},
			GetModelNameFunc: func() string { return "test-model" },
			CloseFunc:        func() error { return nil },
		},
	}

	service := &registryAPIService{
		registry: mockRegistry,
		logger:   logger,
	}

	return service, mockRegistry, logger
}

// TestNewRegistryAPIService tests the constructor function
func TestNewRegistryAPIService(t *testing.T) {
	logger := testutil.NewMockLogger()
	mockRegistry := &MockRegistryAPI{}

	service := NewRegistryAPIService(mockRegistry, logger)

	if service == nil {
		t.Fatal("NewRegistryAPIService returned nil")
	}

	// Check that the service is of the correct type
	_, ok := service.(*registryAPIService)
	if !ok {
		t.Errorf("NewRegistryAPIService did not return a registryAPIService, got %T", service)
	}
}

// TestProcessLLMResponse tests the ProcessLLMResponse method
func TestProcessLLMResponse(t *testing.T) {
	service, _, _ := setupTest(t)

	testCases := []struct {
		name            string
		result          *llm.ProviderResult
		expectedContent string
		expectError     bool
		expectedError   error
	}{
		{
			name:            "nil result",
			result:          nil,
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
		},
		{
			name: "empty content",
			result: &llm.ProviderResult{
				Content:      "",
				FinishReason: "stop",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
		},
		{
			name: "safety blocked",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{
						Category: "harmful_content",
						Blocked:  true,
					},
				},
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrSafetyBlocked,
		},
		{
			name: "whitespace only content",
			result: &llm.ProviderResult{
				Content: "   \n   ",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrWhitespaceContent,
		},
		{
			name: "valid content",
			result: &llm.ProviderResult{
				Content: "This is a valid response",
			},
			expectedContent: "This is a valid response",
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := service.ProcessLLMResponse(tc.result)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}

				if !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error to be '%v', got '%v'", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

				if content != tc.expectedContent {
					t.Errorf("Expected content to be '%s', got '%s'", tc.expectedContent, content)
				}
			}
		})
	}
}

// TestErrorClassificationMethods tests the error classification methods
func TestErrorClassificationMethods(t *testing.T) {
	service, _, _ := setupTest(t)

	// Test IsEmptyResponseError
	t.Run("IsEmptyResponseError", func(t *testing.T) {
		testCases := []struct {
			name     string
			err      error
			expected bool
		}{
			{
				name:     "nil error",
				err:      nil,
				expected: false,
			},
			{
				name:     "empty response error",
				err:      llm.ErrEmptyResponse,
				expected: true,
			},
			{
				name:     "whitespace content error",
				err:      llm.ErrWhitespaceContent,
				expected: true,
			},
			{
				name:     "message contains empty response",
				err:      errors.New("received empty response from API"),
				expected: true,
			},
			{
				name:     "unrelated error",
				err:      errors.New("some other error"),
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := service.IsEmptyResponseError(tc.err)
				if result != tc.expected {
					t.Errorf("Expected IsEmptyResponseError to return %v, got %v", tc.expected, result)
				}
			})
		}
	})

	// Test IsSafetyBlockedError
	t.Run("IsSafetyBlockedError", func(t *testing.T) {
		testCases := []struct {
			name     string
			err      error
			expected bool
		}{
			{
				name:     "nil error",
				err:      nil,
				expected: false,
			},
			{
				name:     "safety blocked error",
				err:      llm.ErrSafetyBlocked,
				expected: true,
			},
			{
				name:     "message contains safety",
				err:      errors.New("content blocked by safety filters"),
				expected: true,
			},
			{
				name:     "unrelated error",
				err:      errors.New("some other error"),
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := service.IsSafetyBlockedError(tc.err)
				if result != tc.expected {
					t.Errorf("Expected IsSafetyBlockedError to return %v, got %v", tc.expected, result)
				}
			})
		}
	})
}

// TestGetModelParameters tests the GetModelParameters method
func TestGetModelParameters(t *testing.T) {
	service, mockRegistry, _ := setupTest(t)

	t.Run("existing model", func(t *testing.T) {
		params, err := service.GetModelParameters("test-model")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that the parameters were correctly retrieved
		if len(params) != 3 {
			t.Errorf("Expected 3 parameters, got %d", len(params))
		}

		// Check specific parameter values
		if temp, ok := params["temperature"]; !ok || temp != 0.7 {
			t.Errorf("Expected temperature parameter to be 0.7, got %v", temp)
		}

		if maxTokens, ok := params["max_tokens"]; !ok || maxTokens != 1024 {
			t.Errorf("Expected max_tokens parameter to be 1024, got %v", maxTokens)
		}

		if modelType, ok := params["model_type"]; !ok || modelType != "default" {
			t.Errorf("Expected model_type parameter to be 'default', got %v", modelType)
		}
	})

	t.Run("non-existent model", func(t *testing.T) {
		mockRegistry.getModelErr = errors.New("model not found")

		params, err := service.GetModelParameters("non-existent-model")

		// Service should return an empty map rather than error for missing models
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(params) != 0 {
			t.Errorf("Expected empty parameter map, got %v", params)
		}
	})
}

// Additional test functions would follow for remaining methods:
// - TestValidateModelParameter
// - TestGetModelDefinition
// - TestGetModelTokenLimits
// - TestGetEnvVarNameForProvider
// - TestGetErrorDetails
