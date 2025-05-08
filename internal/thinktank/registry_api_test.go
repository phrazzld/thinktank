// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers"
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

// MockProviderAPI implements providers.Provider for testing
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
func (m *MockRegistryAPI) GetProviderImplementation(name string) (providers.Provider, error) {
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
	// Test cases to cover different scenarios
	testCases := []struct {
		name          string
		registry      interface{}
		logger        logutil.LoggerInterface
		expectSuccess bool
		expectedType  interface{}
		checkFields   bool
	}{
		{
			name:          "valid registry and logger",
			registry:      &MockRegistryAPI{},
			logger:        testutil.NewMockLogger(),
			expectSuccess: true,
			expectedType:  &registryAPIService{},
			checkFields:   true,
		},
		{
			name:          "nil registry",
			registry:      nil,
			logger:        testutil.NewMockLogger(),
			expectSuccess: true, // The constructor doesn't validate inputs
			expectedType:  &registryAPIService{},
			checkFields:   true,
		},
		{
			name:          "non-registry type",
			registry:      "not a registry",
			logger:        testutil.NewMockLogger(),
			expectSuccess: true, // The constructor accepts any interface{}
			expectedType:  &registryAPIService{},
			checkFields:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the constructor
			service := NewRegistryAPIService(tc.registry, tc.logger)

			// Check if service was created successfully
			if tc.expectSuccess {
				if service == nil {
					t.Fatal("NewRegistryAPIService returned nil, expected a valid service")
				}

				// Check that the service is of the correct type
				serviceImpl, ok := service.(*registryAPIService)
				if !ok {
					t.Fatalf("NewRegistryAPIService did not return a registryAPIService, got %T", service)
				}

				// Check fields only if specified
				if tc.checkFields {
					// Check that registry field was set correctly
					if serviceImpl.registry != tc.registry {
						t.Errorf("Expected registry field to be %v, got %v", tc.registry, serviceImpl.registry)
					}

					// Check that logger field was set correctly
					if serviceImpl.logger != tc.logger {
						t.Errorf("Expected logger field to be %v, got %v", tc.logger, serviceImpl.logger)
					}
				}
			} else {
				if service != nil {
					t.Errorf("Expected NewRegistryAPIService to return nil, got %v", service)
				}
			}
		})
	}

	// Test that the service implements the APIService interface
	t.Run("implements APIService interface", func(t *testing.T) {
		logger := testutil.NewMockLogger()
		mockRegistry := &MockRegistryAPI{}
		service := NewRegistryAPIService(mockRegistry, logger)

		// Simple check that service is not nil (interface checking done at compile time)
		if service == nil {
			t.Fatal("Service should not be nil")
		}
	})
}

// TestProcessLLMResponse moved to registry_api_error_handling_test.go

// TestErrorClassificationMethods moved to registry_api_error_handling_test.go

// TestGetModelParameters tests the GetModelParameters method
func TestGetModelParameters(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name            string
		modelName       string
		modelDefinition *registry.ModelDefinition
		getModelErr     error
		registryImpl    interface{}
		expectedParams  map[string]interface{}
		expectError     bool
		errorSubstring  string
	}{
		{
			name:         "existing model with parameters",
			modelName:    "test-model",
			getModelErr:  nil,
			registryImpl: nil, // Use the default mock registry
			expectedParams: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  1024,
				"model_type":  "default",
			},
			expectError: false,
		},
		{
			name:           "non-existent model",
			modelName:      "non-existent-model",
			getModelErr:    errors.New("model not found"),
			registryImpl:   nil, // Use the default mock registry
			expectedParams: map[string]interface{}{},
			expectError:    false,
		},
		{
			name:           "registry does not implement GetModel",
			modelName:      "test-model",
			getModelErr:    nil,
			registryImpl:   "not a registry", // String instead of proper mock registry
			expectError:    true,
			errorSubstring: "does not implement GetModel method",
		},
		{
			name:      "model with no parameters",
			modelName: "empty-model",
			modelDefinition: &registry.ModelDefinition{
				Name:       "empty-model",
				Provider:   "test-provider",
				APIModelID: "empty-model-id",
				Parameters: map[string]registry.ParameterDefinition{},
			},
			registryImpl:   nil, // Use the default mock registry
			expectedParams: map[string]interface{}{},
			expectError:    false,
		},
		{
			name:      "model with parameters but no defaults",
			modelName: "no-defaults-model",
			modelDefinition: &registry.ModelDefinition{
				Name:       "no-defaults-model",
				Provider:   "test-provider",
				APIModelID: "no-defaults-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"param1": {
						Type: "float",
						// No default value
					},
					"param2": {
						Type: "int",
						// No default value
					},
				},
			},
			registryImpl:   nil, // Use the default mock registry
			expectedParams: map[string]interface{}{},
			expectError:    false,
		},
		{
			name:      "model with mixed parameters (with and without defaults)",
			modelName: "mixed-defaults-model",
			modelDefinition: &registry.ModelDefinition{
				Name:       "mixed-defaults-model",
				Provider:   "test-provider",
				APIModelID: "mixed-defaults-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"with_default": {
						Type:    "float",
						Default: 0.5,
					},
					"without_default": {
						Type: "int",
						// No default value
					},
					"string_param": {
						Type:    "string",
						Default: "default-value",
					},
				},
			},
			registryImpl: nil, // Use the default mock registry
			expectedParams: map[string]interface{}{
				"with_default": 0.5,
				"string_param": "default-value",
			},
			expectError: false,
		},
		{
			name:      "model with nil default value",
			modelName: "nil-default-model",
			modelDefinition: &registry.ModelDefinition{
				Name:       "nil-default-model",
				Provider:   "test-provider",
				APIModelID: "nil-default-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"nil_default": {
						Type:    "string",
						Default: nil,
					},
				},
			},
			registryImpl:   nil, // Use the default mock registry
			expectedParams: map[string]interface{}{},
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			var service *registryAPIService
			var mockRegistry *MockRegistryAPI

			if tc.registryImpl == nil {
				service, mockRegistry, _ = setupTest(t)

				// Add custom model definition if provided
				if tc.modelDefinition != nil {
					mockRegistry.models[tc.modelName] = tc.modelDefinition
				}

				// Set get model error if provided
				mockRegistry.getModelErr = tc.getModelErr
			} else {
				// Use the provided registry implementation
				logger := testutil.NewMockLogger()
				service = &registryAPIService{
					registry: tc.registryImpl,
					logger:   logger,
				}
			}

			// Call the method being tested
			params, err := service.GetModelParameters(tc.modelName)

			// Verify expected error behavior
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tc.errorSubstring)
				}
				if !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error containing '%s', got '%v'", tc.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

				// Verify the parameters map
				if len(params) != len(tc.expectedParams) {
					t.Errorf("Expected %d parameters, got %d", len(tc.expectedParams), len(params))
				}

				// Check each expected parameter
				for key, expectedValue := range tc.expectedParams {
					actualValue, ok := params[key]
					if !ok {
						t.Errorf("Expected parameter '%s' not found in result", key)
						continue
					}

					if actualValue != expectedValue {
						t.Errorf("Parameter '%s': expected '%v', got '%v'", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

// TestInitLLMClient tests the InitLLMClient method
func TestInitLLMClient(t *testing.T) {
	// Store and restore environment variables
	origEnv := make(map[string]string)
	restoreEnv := func() {
		for name, value := range origEnv {
			if value != "" {
				if err := os.Setenv(name, value); err != nil {
					t.Logf("Warning: failed to restore env var %s: %v", name, err)
				}
			} else {
				if err := os.Unsetenv(name); err != nil {
					t.Logf("Warning: failed to unset env var %s: %v", name, err)
				}
			}
		}
	}

	// Save environment variables we'll modify
	envVarsToSave := []string{"TEST_PROVIDER_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY"}
	for _, name := range envVarsToSave {
		origEnv[name], _ = os.LookupEnv(name)
	}

	// We'll restore environment variables at the end
	defer restoreEnv()

	// Create a done context we can cancel to test cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("basic validation", func(t *testing.T) {
		service, _, _ := setupTest(t)

		// Test empty model name
		_, err := service.InitLLMClient(ctx, "test-api-key", "", "")
		if err == nil {
			t.Error("Expected error for empty model name, got nil")
		}

		// Test cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		_, err = service.InitLLMClient(cancelledCtx, "test-api-key", "test-model", "")
		if err == nil {
			t.Error("Expected error for cancelled context, got nil")
		}
		if !strings.Contains(err.Error(), "context") {
			t.Errorf("Expected context error, got: %v", err)
		}
	})

	t.Run("test error model", func(t *testing.T) {
		service, _, _ := setupTest(t)

		// Test the special error model
		_, err := service.InitLLMClient(ctx, "test-api-key", "error-model", "")
		if err == nil {
			t.Error("Expected error for error-model, got nil")
		}
		if err.Error() != "test model error" {
			t.Errorf("Expected 'test model error', got: %v", err)
		}
	})

	t.Run("model lookup failure", func(t *testing.T) {
		service, mockRegistry, _ := setupTest(t)

		// Set error for GetModel
		mockRegistry.getModelErr = errors.New("model lookup failed")

		_, err := service.InitLLMClient(ctx, "test-api-key", "test-model", "")
		if err == nil {
			t.Error("Expected error for model lookup failure, got nil")
		}
		if !errors.Is(err, llm.ErrModelNotFound) {
			t.Errorf("Expected ErrModelNotFound, got: %v", err)
		}
	})

	t.Run("provider lookup failure", func(t *testing.T) {
		service, mockRegistry, _ := setupTest(t)

		// Set error for GetProvider
		mockRegistry.getProviderErr = errors.New("provider lookup failed")

		_, err := service.InitLLMClient(ctx, "test-api-key", "test-model", "")
		if err == nil {
			t.Error("Expected error for provider lookup failure, got nil")
		}
		if !errors.Is(err, llm.ErrClientInitialization) {
			t.Errorf("Expected ErrClientInitialization, got: %v", err)
		}
	})

	t.Run("provider implementation lookup failure", func(t *testing.T) {
		service, mockRegistry, _ := setupTest(t)

		// Set error for GetProviderImplementation
		mockRegistry.getProviderImplErr = errors.New("provider implementation lookup failed")

		_, err := service.InitLLMClient(ctx, "test-api-key", "test-model", "")
		if err == nil {
			t.Error("Expected error for provider implementation lookup failure, got nil")
		}
		if !errors.Is(err, llm.ErrClientInitialization) {
			t.Errorf("Expected ErrClientInitialization, got: %v", err)
		}
	})

	t.Run("client creation failure", func(t *testing.T) {
		service, mockRegistry, _ := setupTest(t)

		// Update the provider implementation to return an error
		providerImpl := mockRegistry.implementations["test-provider"]
		providerImpl.createClientErr = errors.New("client creation failed")
		mockRegistry.implementations["test-provider"] = providerImpl

		_, err := service.InitLLMClient(ctx, "test-api-key", "test-model", "")
		if err == nil {
			t.Error("Expected error for client creation failure, got nil")
		}
		if !errors.Is(err, llm.ErrClientInitialization) {
			t.Errorf("Expected ErrClientInitialization, got: %v", err)
		}
	})

	t.Run("custom endpoint is logged", func(t *testing.T) {
		service, _, logger := setupTest(t)

		// Test with a custom endpoint (we don't care if it ultimately fails due to API key)
		_, err := service.InitLLMClient(ctx, "test-api-key", "test-model", "https://custom-endpoint.example.com")
		// We expect this to fail due to API key issues, but we don't care about that for this test
		if err == nil {
			t.Log("Unexpectedly succeeded with no API key configured")
		}

		// Check that the custom endpoint was logged
		if !logger.ContainsMessage("Using custom API endpoint: https://custom-endpoint.example.com") {
			t.Error("Expected log message about custom endpoint, not found")
		}
	})
}

// TestValidateModelParameter moved to registry_api_validation_test.go

// TestGetModelDefinition tests the GetModelDefinition method
func TestGetModelDefinition(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		modelName      string
		modelDef       *registry.ModelDefinition
		getModelErr    error
		registryImpl   interface{}
		expectSuccess  bool
		expectError    bool
		errorSubstring string
		wrapsWith      error // Expected error to be wrapped with
	}{
		{
			name:          "existing model",
			modelName:     "test-model",
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:           "model not found",
			modelName:      "non-existent-model",
			getModelErr:    errors.New("model not found"),
			expectSuccess:  false,
			expectError:    true,
			errorSubstring: "non-existent-model",
			wrapsWith:      llm.ErrModelNotFound,
		},
		{
			name:           "empty model name",
			modelName:      "",                            // Empty model name should still attempt lookup
			getModelErr:    errors.New("model not found"), // Registry returns this error for empty name
			expectSuccess:  false,
			expectError:    true,
			errorSubstring: "", // Don't check specific error message
			wrapsWith:      llm.ErrModelNotFound,
		},
		{
			name:           "registry does not implement GetModel",
			modelName:      "test-model",
			registryImpl:   "not a registry", // String instead of proper mock registry
			expectSuccess:  false,
			expectError:    true,
			errorSubstring: "does not implement GetModel method",
		},
		{
			name:      "custom model definition",
			modelName: "custom-model",
			modelDef: &registry.ModelDefinition{
				Name:       "custom-model",
				Provider:   "custom-provider",
				APIModelID: "custom-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"custom_param": {
						Type:    "string",
						Default: "custom-value",
						EnumValues: []string{
							"custom-value",
							"other-value",
						},
					},
					"numeric_param": {
						Type:    "float",
						Default: 0.5,
						Min:     0.0,
						Max:     1.0,
					},
				},
			},
			expectSuccess: true,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			var service *registryAPIService
			var mockRegistry *MockRegistryAPI

			if tc.registryImpl == nil {
				service, mockRegistry, _ = setupTest(t)

				// Add custom model definition if provided
				if tc.modelDef != nil {
					mockRegistry.models[tc.modelName] = tc.modelDef
				}

				// Set get model error if provided
				mockRegistry.getModelErr = tc.getModelErr
			} else {
				// Use the provided registry implementation
				logger := testutil.NewMockLogger()
				service = &registryAPIService{
					registry: tc.registryImpl,
					logger:   logger,
				}
			}

			// Call the method being tested
			modelDef, err := service.GetModelDefinition(tc.modelName)

			// Verify expected error behavior
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error for case '%s', got nil", tc.name)
				}

				// Verify error message contains expected substring
				if tc.errorSubstring != "" && !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstring, err)
				}

				// Verify error wrapping if expected
				if tc.wrapsWith != nil && !errors.Is(err, tc.wrapsWith) {
					t.Errorf("Expected error to be wrapped with '%v', got: %v", tc.wrapsWith, err)
				}
			} else if tc.expectSuccess {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				// Verify model definition is not nil
				if modelDef == nil {
					t.Fatalf("Expected non-nil model definition, got nil")
				}

				// Determine expected model definition
				var expectedDef *registry.ModelDefinition
				if tc.modelDef != nil {
					expectedDef = tc.modelDef
				} else {
					expectedDef = mockRegistry.models[tc.modelName]
				}

				// Verify model definition fields
				if modelDef.Name != expectedDef.Name {
					t.Errorf("Expected Name '%s', got '%s'", expectedDef.Name, modelDef.Name)
				}

				if modelDef.Provider != expectedDef.Provider {
					t.Errorf("Expected Provider '%s', got '%s'", expectedDef.Provider, modelDef.Provider)
				}

				if modelDef.APIModelID != expectedDef.APIModelID {
					t.Errorf("Expected APIModelID '%s', got '%s'", expectedDef.APIModelID, modelDef.APIModelID)
				}

				// Verify parameters if expected model has them
				if len(expectedDef.Parameters) > 0 {
					if len(modelDef.Parameters) != len(expectedDef.Parameters) {
						t.Errorf("Expected %d parameters, got %d", len(expectedDef.Parameters), len(modelDef.Parameters))
					}

					// Check each parameter
					for paramName, expectedParam := range expectedDef.Parameters {
						actualParam, ok := modelDef.Parameters[paramName]
						if !ok {
							t.Errorf("Expected parameter '%s' not found in result", paramName)
							continue
						}

						// Verify parameter type
						if actualParam.Type != expectedParam.Type {
							t.Errorf("Parameter '%s': expected type '%s', got '%s'", paramName, expectedParam.Type, actualParam.Type)
						}

						// Verify default value if exists
						if expectedParam.Default != nil {
							if actualParam.Default != expectedParam.Default {
								t.Errorf("Parameter '%s': expected default '%v', got '%v'", paramName, expectedParam.Default, actualParam.Default)
							}
						}

						// Verify min/max for numeric parameters
						if expectedParam.Min != nil && actualParam.Min != expectedParam.Min {
							t.Errorf("Parameter '%s': expected min '%v', got '%v'", paramName, expectedParam.Min, actualParam.Min)
						}

						if expectedParam.Max != nil && actualParam.Max != expectedParam.Max {
							t.Errorf("Parameter '%s': expected max '%v', got '%v'", paramName, expectedParam.Max, actualParam.Max)
						}

						// Verify enum values for string parameters
						if len(expectedParam.EnumValues) > 0 {
							if len(actualParam.EnumValues) != len(expectedParam.EnumValues) {
								t.Errorf("Parameter '%s': expected %d enum values, got %d", paramName, len(expectedParam.EnumValues), len(actualParam.EnumValues))
							} else {
								for i, enumVal := range expectedParam.EnumValues {
									if i < len(actualParam.EnumValues) && actualParam.EnumValues[i] != enumVal {
										t.Errorf("Parameter '%s': expected enum value at index %d to be '%s', got '%s'", paramName, i, enumVal, actualParam.EnumValues[i])
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

// TestGetModelTokenLimits tests the GetModelTokenLimits method
func TestGetModelTokenLimits(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name                  string
		modelName             string
		modelDef              *registry.ModelDefinition
		getModelErr           error
		registryImpl          interface{}
		expectedContextWindow int32
		expectedMaxTokens     int32
		expectError           bool
		errorSubstring        string
		wrapsWith             error // Expected error to be wrapped with
	}{
		{
			name:      "existing model with token limits",
			modelName: "model-with-limits",
			modelDef: &registry.ModelDefinition{
				Name:            "model-with-limits",
				Provider:        "test-provider",
				APIModelID:      "model-id",
				ContextWindow:   50000,
				MaxOutputTokens: 10000,
			},
			expectedContextWindow: 50000,
			expectedMaxTokens:     10000,
			expectError:           false,
		},
		{
			name:      "existing model without token limits",
			modelName: "model-without-limits",
			modelDef: &registry.ModelDefinition{
				Name:       "model-without-limits",
				Provider:   "test-provider",
				APIModelID: "model-id",
				// No token limits specified
			},
			expectedContextWindow: 1000000, // Should use high default
			expectedMaxTokens:     65000,   // Should use high default
			expectError:           false,
		},
		{
			name:           "model not found",
			modelName:      "non-existent-model",
			getModelErr:    errors.New("model not found"),
			expectError:    true,
			errorSubstring: "non-existent-model",
			wrapsWith:      llm.ErrModelNotFound,
		},
		{
			name:           "empty model name",
			modelName:      "",
			getModelErr:    errors.New("model not found"), // Registry returns this error for empty name
			expectError:    true,
			errorSubstring: "", // Don't check specific error message
			wrapsWith:      llm.ErrModelNotFound,
		},
		{
			name:           "registry does not implement GetModel",
			modelName:      "test-model",
			registryImpl:   "not a registry", // String instead of proper mock registry
			expectError:    true,
			errorSubstring: "does not implement GetModel method",
		},
		{
			name:      "synthesis model with very large limits",
			modelName: "synthesis-model",
			modelDef: &registry.ModelDefinition{
				Name:            "synthesis-model",
				Provider:        "test-provider",
				APIModelID:      "synthesis-model-id",
				ContextWindow:   1000000, // 1M context window
				MaxOutputTokens: 200000,  // 200K output tokens
				Parameters: map[string]registry.ParameterDefinition{
					"temperature": {
						Type:    "float",
						Default: 0.7,
					},
				},
			},
			expectedContextWindow: 1000000,
			expectedMaxTokens:     200000,
			expectError:           false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			var service *registryAPIService
			var mockRegistry *MockRegistryAPI

			if tc.registryImpl == nil {
				service, mockRegistry, _ = setupTest(t)

				// Add custom model definition if provided
				if tc.modelDef != nil {
					mockRegistry.models[tc.modelName] = tc.modelDef
				}

				// Set get model error if provided
				mockRegistry.getModelErr = tc.getModelErr
			} else {
				// Use the provided registry implementation
				logger := testutil.NewMockLogger()
				service = &registryAPIService{
					registry: tc.registryImpl,
					logger:   logger,
				}
			}

			// Call the method being tested
			contextWindow, maxTokens, err := service.GetModelTokenLimits(tc.modelName)

			// Verify expected error behavior
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error for case '%s', got nil", tc.name)
				}

				// Verify error message contains expected substring
				if tc.errorSubstring != "" && !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstring, err)
				}

				// Verify error wrapping if expected
				if tc.wrapsWith != nil && !errors.Is(err, tc.wrapsWith) {
					t.Errorf("Expected error to be wrapped with '%v', got: %v", tc.wrapsWith, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				// Verify return values
				if contextWindow != tc.expectedContextWindow {
					t.Errorf("Expected contextWindow=%d, got %d", tc.expectedContextWindow, contextWindow)
				}

				if maxTokens != tc.expectedMaxTokens {
					t.Errorf("Expected maxTokens=%d, got %d", tc.expectedMaxTokens, maxTokens)
				}
			}
		})
	}
}

// Additional test functions would follow for remaining methods:
// - TestGetEnvVarNameForProvider
// - TestGetErrorDetails
