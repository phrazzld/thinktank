// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// Test-only error definitions for provider-related errors
var (
	// ErrProviderNotFound indicates a provider was not found in the registry
	ErrProviderNotFound = errors.New("provider not found in registry")

	// ErrProviderInitialization indicates provider initialization failed
	ErrProviderInitialization = errors.New("error initializing provider")
)

// TestRegistryAPIServiceBoundaries verifies that the registryAPIService
// implementation correctly handles operations across system boundaries.
// This focuses on:
// 1. Registry interaction - how the API interacts with the registry
// 2. Context propagation - how context is maintained across boundaries
// 3. Error handling - how errors are translated across boundaries
// 4. Authentication - how API keys are resolved and used
func TestRegistryAPIServiceBoundaries(t *testing.T) {
	// Set up a detailed boundary test logger
	logger := testutil.NewMockLogger()

	// Create a mock registry with tracking of method calls
	mockRegistry := &BoundaryMockRegistry{
		models: map[string]*registry.ModelDefinition{
			"test-model": {
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
				},
			},
		},
		providers: map[string]*registry.ProviderDefinition{
			"test-provider": {
				Name:    "test-provider",
				BaseURL: "https://api.test-provider.example.com",
			},
		},
		implementations: map[string]ProviderImplementation{
			"test-provider": &BoundaryMockProvider{
				CreateClientResult: &BoundaryMockLLMClient{
					ModelName: "test-model-id",
				},
			},
		},
		// Track calls for boundary verification
		methodCalls: make(map[string][]BoundaryMethodCall),
	}

	// Create the service with our instrumented mock registry
	service := NewRegistryAPIService(mockRegistry, logger)

	// 1. Test Registry Boundary: Context Propagation
	t.Run("ContextPropagation", func(t *testing.T) {
		// Create a context with a deadline
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Set a correlation ID in the context
		ctx = logutil.WithCustomCorrelationID(ctx, "test-correlation-id")

		// Call a method that should accept the context
		_, _ = service.InitLLMClient(ctx, "test-api-key", "test-model", "")

		// In real code, the context would be passed to the provider implementation's CreateClient method
		// Let's verify that we got that far in the call chain
		providerCalls := mockRegistry.GetMethodCalls("GetProviderImplementation")
		if len(providerCalls) == 0 {
			t.Fatal("Expected GetProviderImplementation to be called")
		}

		// Verify a provider implementation was retrieved
		provider, ok := mockRegistry.implementations["test-provider"].(*BoundaryMockProvider)
		if !ok {
			t.Fatal("Expected to find test-provider implementation")
		}

		// Verify CreateClient was called with our context
		if len(provider.CreateClientCalls) == 0 {
			t.Fatal("Expected CreateClient to be called")
		}

		callCtx := provider.CreateClientCalls[0].Ctx

		// Check if context has our correlation ID
		correlationID := logutil.GetCorrelationID(callCtx)
		if correlationID != "test-correlation-id" {
			t.Errorf("Expected correlation ID to be propagated, got %q", correlationID)
		}

		// Check that deadline was propagated
		deadline, hasDeadline := callCtx.Deadline()
		if !hasDeadline {
			t.Error("Expected context deadline to be propagated")
		} else {
			// The deadline should be roughly 1 second from when we created it
			expectedDeadline := time.Now().Add(1 * time.Second)
			if deadline.Sub(expectedDeadline) > 500*time.Millisecond {
				t.Errorf("Deadline not propagated correctly: got %v", deadline)
			}
		}

		// Verify that Done channel is properly connected
		select {
		case <-callCtx.Done():
			t.Error("Context should not be done yet")
		default:
			// This is expected
		}

		// Cancel the parent context
		cancel()

		// Now the propagated context should also be done
		select {
		case <-callCtx.Done():
			// This is expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Context cancellation was not propagated")
		}

		// Check that calling with already cancelled context immediately returns error
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		_, cancelErr := service.InitLLMClient(cancelledCtx, "test-api-key", "test-model", "")
		if cancelErr == nil || !strings.Contains(cancelErr.Error(), "context") {
			t.Errorf("Expected context error for cancelled context, got: %v", cancelErr)
		}
	})

	// 2. Test Registry Boundary: Error Translation
	t.Run("ErrorTranslation", func(t *testing.T) {
		// The "error-model" is specially handled in the registryAPIService implementation
		// and is not actually looked up in the registry, so we don't need to add it.

		// Create a new registry implementation that returns errors
		errorRegistry := &BoundaryMockRegistry{
			models: map[string]*registry.ModelDefinition{
				"test-error-model": {
					Name:     "test-error-model",
					Provider: "error-provider",
				},
			},
			providers: map[string]*registry.ProviderDefinition{
				"error-provider": {
					Name: "error-provider",
				},
			},
			implementations: map[string]ProviderImplementation{
				"error-provider": &BoundaryMockProvider{
					CreateClientError: fmt.Errorf("provider error: %w", llm.ErrSafetyBlocked),
				},
			},
			methodCalls: make(map[string][]BoundaryMethodCall),
		}

		// Create a new service with the error registry
		errorService := NewRegistryAPIService(errorRegistry, testutil.NewMockLogger())

		// Clear previous call logs
		mockRegistry.ClearMethodCalls()

		ctx := context.Background()

		// Test case 1: The special error-model case
		t.Run("SpecialErrorModel", func(t *testing.T) {
			// Call method with the special error-model name that triggers a direct error return
			_, err := service.InitLLMClient(ctx, "test-api-key", "error-model", "")

			// Verify error was returned
			if err == nil {
				t.Fatal("Expected error from InitLLMClient with error-model")
			}

			// Check the error message
			expectedMsg := "test model error"
			if err.Error() != expectedMsg {
				t.Errorf("Expected error message %q, got: %v", expectedMsg, err.Error())
			}
		})

		// Test case 2: Error during CreateClient
		t.Run("CreateClientError", func(t *testing.T) {
			// Call method with the provider that returns an error during CreateClient
			_, err := errorService.InitLLMClient(ctx, "test-api-key", "test-error-model", "")

			// Verify error is correctly wrapped
			if err == nil {
				t.Fatal("Expected error from InitLLMClient with error provider")
			}

			// Check that the error is properly wrapped with the appropriate domain error
			if !errors.Is(err, llm.ErrClientInitialization) {
				t.Errorf("Expected error to be wrapped with ErrClientInitialization, got: %v", err)
			}

			// The error should at least mention safety filters
			if !strings.Contains(err.Error(), "safety") && !strings.Contains(err.Error(), "blocked") {
				t.Errorf("Expected error to mention safety filters or blocking, got: %v", err)
			}

			// Check that error message is helpful
			expectedSubstring := "provider error"
			if !strings.Contains(err.Error(), expectedSubstring) {
				t.Errorf("Expected error message to contain %q, got: %v", expectedSubstring, err)
			}
		})

		// Test case 3: Non-existent model error translation
		t.Run("NonExistentModel", func(t *testing.T) {
			// Call method with a non-existent model
			_, modelErr := service.GetModelDefinition(ctx, "non-existent-model")
			if modelErr == nil {
				t.Fatal("Expected error from GetModelDefinition with non-existent model")
			}

			// Check that the error is properly wrapped with the appropriate domain error
			if !errors.Is(modelErr, llm.ErrModelNotFound) {
				t.Errorf("Expected error to be wrapped with ErrModelNotFound, got: %v", modelErr)
			}
		})
	})

	// 3. Test Registry Boundary: Registry Interaction
	t.Run("RegistryInteraction", func(t *testing.T) {
		// Create a fresh registry for this test to avoid accumulating calls
		testRegistry := &BoundaryMockRegistry{
			models: map[string]*registry.ModelDefinition{
				"test-model": {
					Name:       "test-model",
					Provider:   "test-provider",
					APIModelID: "test-model-id",
					Parameters: map[string]registry.ParameterDefinition{
						"temperature": {
							Type:    "float",
							Default: 0.7,
						},
					},
				},
			},
			providers: map[string]*registry.ProviderDefinition{
				"test-provider": {
					Name:    "test-provider",
					BaseURL: "https://api.test-provider.example.com",
				},
			},
			implementations: map[string]ProviderImplementation{
				"test-provider": &BoundaryMockProvider{
					CreateClientResult: &BoundaryMockLLMClient{
						ModelName: "test-model-id",
					},
				},
			},
			methodCalls: make(map[string][]BoundaryMethodCall),
		}

		// Create a service with the test registry
		testService := NewRegistryAPIService(testRegistry, testutil.NewMockLogger())

		// Create context for tests
		ctx := context.Background()

		// 1. Test GetModelParameters
		t.Run("GetModelParameters", func(t *testing.T) {
			// Call the method
			params, err := testService.GetModelParameters(ctx, "test-model")
			if err != nil {
				t.Errorf("GetModelParameters failed: %v", err)
			}

			// Verify parameters returned
			if params["temperature"] != 0.7 {
				t.Errorf("Expected temperature parameter to be 0.7, got %v", params["temperature"])
			}

			// Verify the registry was called correctly
			modelCalls := testRegistry.GetMethodCalls("GetModel")
			if len(modelCalls) != 1 {
				t.Errorf("Expected 1 call to GetModel, got %d", len(modelCalls))
			} else if modelName, ok := modelCalls[0].Args[0].(string); !ok || modelName != "test-model" {
				t.Errorf("GetModel was called with wrong arguments: %v", modelCalls[0].Args)
			}
		})

		// 2. Test InitLLMClient
		t.Run("InitLLMClient", func(t *testing.T) {
			// Reset call tracking for this test
			testRegistry.ClearMethodCalls()
			ctx := context.Background()

			// Call the method
			_, err := testService.InitLLMClient(ctx, "test-api-key", "test-model", "")
			if err != nil {
				t.Errorf("InitLLMClient failed: %v", err)
			}

			// Get the call sequence
			callSequence := testRegistry.GetCallSequence()

			// Verify the 3 methods were called in the correct order
			expectedPattern := []string{"GetModel", "GetProvider", "GetProviderImplementation"}
			if len(callSequence) < 3 {
				t.Errorf("Expected at least 3 registry calls, got %d: %v", len(callSequence), callSequence)
			} else {
				// Check just the first 3 calls match our pattern
				for i := 0; i < 3; i++ {
					if i >= len(callSequence) || callSequence[i] != expectedPattern[i] {
						if i < len(callSequence) {
							t.Errorf("Expected call %d to be %s, got %s", i, expectedPattern[i], callSequence[i])
						} else {
							t.Errorf("Expected call %d to be %s, but didn't have enough calls", i, expectedPattern[i])
						}
					}
				}
			}

			// Verify provider implementation's CreateClient was called with correct args
			provider := testRegistry.implementations["test-provider"].(*BoundaryMockProvider)
			if len(provider.CreateClientCalls) != 1 {
				t.Errorf("Expected 1 call to CreateClient, got %d", len(provider.CreateClientCalls))
			} else {
				call := provider.CreateClientCalls[0]
				if call.APIKey != "test-api-key" {
					t.Errorf("Expected APIKey to be 'test-api-key', got %q", call.APIKey)
				}
				if call.ModelID != "test-model-id" {
					t.Errorf("Expected ModelID to be 'test-model-id', got %q", call.ModelID)
				}
			}
		})
	})
}

// TestRegistryAPIWithAdapterBoundary tests the interaction between
// the RegistryAPIService and adapters at system boundaries, including:
// - Type conversion between different interfaces
// - Error propagation through adapter layers
// - Fallback behavior for missing implementation details
func TestRegistryAPIWithAdapterBoundary(t *testing.T) {
	// Create the basic mock registry API
	mockRegistry := &BoundaryMockRegistry{
		models: map[string]*registry.ModelDefinition{
			"test-model": {
				Name:       "test-model",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
			},
		},
		providers: map[string]*registry.ProviderDefinition{
			"test-provider": {
				Name: "test-provider",
			},
		},
		implementations: map[string]ProviderImplementation{
			"test-provider": &BoundaryMockProvider{},
		},
		methodCalls: make(map[string][]BoundaryMethodCall),
	}

	// Create a registry API service
	service := NewRegistryAPIService(mockRegistry, testutil.NewMockLogger())

	// Create an adapter around the service
	adapter := &APIServiceAdapter{
		APIService: service,
	}

	t.Run("AdapterDelegation", func(t *testing.T) {
		// Create context for tests
		ctx := context.Background()

		// Clear previous calls
		mockRegistry.ClearMethodCalls()

		// Call through the adapter
		_, err := adapter.GetModelDefinition(ctx, "test-model")
		if err != nil {
			t.Errorf("GetModelDefinition through adapter failed: %v", err)
		}

		// Verify the registry was called
		modelCalls := mockRegistry.GetMethodCalls("GetModel")
		if len(modelCalls) != 1 {
			t.Errorf("Expected 1 call to GetModel, got %d", len(modelCalls))
		}
	})

	t.Run("AdapterErrorPropagation", func(t *testing.T) {
		// Create context for tests
		ctx := context.Background()

		// Set up error case
		mockRegistry.getModelErr = fmt.Errorf("registry error: %w", llm.ErrModelNotFound)

		// Call through the adapter
		_, err := adapter.GetModelDefinition(ctx, "error-model")

		// Verify error propagation
		if err == nil {
			t.Fatal("Expected error from GetModelDefinition through adapter")
		}

		// Check error wrapping
		if !errors.Is(err, llm.ErrModelNotFound) {
			t.Errorf("Expected error to be wrapped with ErrModelNotFound, got: %v", err)
		}

		// Reset for next test
		mockRegistry.getModelErr = nil
	})
}

// Full adapter registry integration tests are in registry_api_full_test.go

// Test original service contract for backward compatibility
func TestRegistryAPIServiceContract(t *testing.T) {
	// Logger is required but not used directly in this test
	_ = logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")

	// Create a mock APIService
	mockService := &mockAPIService{
		isEmptyResponseErrorFunc: func(err error) bool {
			return err != nil && errors.Is(err, llm.ErrEmptyResponse)
		},
		isSafetyBlockedErrorFunc: func(err error) bool {
			return err != nil && errors.Is(err, llm.ErrSafetyBlocked)
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			if result == nil {
				return "", llm.ErrEmptyResponse
			}
			if result.Content == "" {
				if len(result.SafetyInfo) > 0 {
					for _, info := range result.SafetyInfo {
						if info.Blocked {
							return "", llm.ErrSafetyBlocked
						}
					}
				}
				return "", llm.ErrEmptyResponse
			}
			return result.Content, nil
		},
	}

	// Test ProcessLLMResponse with various test cases
	t.Run("ProcessLLMResponse", func(t *testing.T) {
		tests := []struct {
			name    string
			result  *llm.ProviderResult
			want    string
			wantErr bool
			errType error
		}{
			{
				name:    "nil result",
				result:  nil,
				want:    "",
				wantErr: true,
				errType: llm.ErrEmptyResponse,
			},
			{
				name: "empty content",
				result: &llm.ProviderResult{
					Content: "",
				},
				want:    "",
				wantErr: true,
				errType: llm.ErrEmptyResponse,
			},
			{
				name: "safety blocked content",
				result: &llm.ProviderResult{
					Content: "",
					SafetyInfo: []llm.Safety{
						{Category: "harmful", Blocked: true},
					},
				},
				want:    "",
				wantErr: true,
				errType: llm.ErrSafetyBlocked,
			},
			{
				name: "valid content",
				result: &llm.ProviderResult{
					Content: "valid content",
				},
				want:    "valid content",
				wantErr: false,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := mockService.ProcessLLMResponse(tc.result)
				if (err != nil) != tc.wantErr {
					t.Errorf("ProcessLLMResponse() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.wantErr && tc.errType != nil && !errors.Is(err, tc.errType) {
					t.Errorf("ProcessLLMResponse() error type = %v, want %v", err, tc.errType)
				}
				if got != tc.want {
					t.Errorf("ProcessLLMResponse() = %v, want %v", got, tc.want)
				}
			})
		}
	})

	// Test Error Classification methods
	t.Run("ErrorClassification", func(t *testing.T) {
		// Create errors for testing
		emptyError := llm.ErrEmptyResponse
		safetyError := llm.ErrSafetyBlocked
		otherError := errors.New("other error")

		// Test IsEmptyResponseError classification
		if !mockService.IsEmptyResponseError(emptyError) {
			t.Errorf("IsEmptyResponseError() should return true for ErrEmptyResponse")
		}
		if mockService.IsEmptyResponseError(safetyError) {
			t.Errorf("IsEmptyResponseError() should return false for ErrSafetyBlocked")
		}
		if mockService.IsEmptyResponseError(otherError) {
			t.Errorf("IsEmptyResponseError() should return false for other errors")
		}

		// Test IsSafetyBlockedError classification
		if !mockService.IsSafetyBlockedError(safetyError) {
			t.Errorf("IsSafetyBlockedError() should return true for ErrSafetyBlocked")
		}
		if mockService.IsSafetyBlockedError(emptyError) {
			t.Errorf("IsSafetyBlockedError() should return false for ErrEmptyResponse")
		}
		if mockService.IsSafetyBlockedError(otherError) {
			t.Errorf("IsSafetyBlockedError() should return false for other errors")
		}
	})
}

// BoundaryMethodCall records a method call for boundary testing
type BoundaryMethodCall struct {
	Method string
	Args   []interface{}
}

// BoundaryMockRegistry is a mock implementation of registry.Registry that
// records method calls for boundary testing
type BoundaryMockRegistry struct {
	models             map[string]*registry.ModelDefinition
	providers          map[string]*registry.ProviderDefinition
	implementations    map[string]ProviderImplementation
	getModelErr        error
	getProviderErr     error
	getProviderImplErr error
	methodCalls        map[string][]BoundaryMethodCall
	callSequence       []string
}

// GetModel implements registry.Registry
func (m *BoundaryMockRegistry) GetModel(ctx context.Context, name string) (*registry.ModelDefinition, error) {
	// Record the call
	m.methodCalls["GetModel"] = append(m.methodCalls["GetModel"], BoundaryMethodCall{
		Method: "GetModel",
		Args:   []interface{}{name},
	})
	m.callSequence = append(m.callSequence, "GetModel")

	if m.getModelErr != nil {
		return nil, m.getModelErr
	}
	model, ok := m.models[name]
	if !ok {
		return nil, llm.Wrap(llm.ErrModelNotFound, "", fmt.Sprintf("model '%s' not found in registry", name), llm.CategoryNotFound)
	}
	return model, nil
}

// GetProvider implements registry.Registry
func (m *BoundaryMockRegistry) GetProvider(ctx context.Context, name string) (*registry.ProviderDefinition, error) {
	// Record the call
	m.methodCalls["GetProvider"] = append(m.methodCalls["GetProvider"], BoundaryMethodCall{
		Method: "GetProvider",
		Args:   []interface{}{name},
	})
	m.callSequence = append(m.callSequence, "GetProvider")

	if m.getProviderErr != nil {
		return nil, m.getProviderErr
	}
	provider, ok := m.providers[name]
	if !ok {
		return nil, llm.Wrap(llm.ErrProviderNotFound, "", fmt.Sprintf("provider '%s' not found in registry", name), llm.CategoryNotFound)
	}
	return provider, nil
}

// GetProviderImplementation implements registry.Registry
func (m *BoundaryMockRegistry) GetProviderImplementation(ctx context.Context, name string) (providers.Provider, error) {
	// Record the call
	m.methodCalls["GetProviderImplementation"] = append(m.methodCalls["GetProviderImplementation"], BoundaryMethodCall{
		Method: "GetProviderImplementation",
		Args:   []interface{}{name},
	})
	m.callSequence = append(m.callSequence, "GetProviderImplementation")

	if m.getProviderImplErr != nil {
		return nil, m.getProviderImplErr
	}
	impl, ok := m.implementations[name]
	if !ok {
		return nil, llm.Wrap(llm.ErrProviderNotFound, "", fmt.Sprintf("provider implementation '%s' not found in registry", name), llm.CategoryNotFound)
	}
	// We need to cast the impl to providers.Provider
	providerImpl, ok := impl.(providers.Provider)
	if !ok {
		return nil, llm.Wrap(llm.ErrClientInitialization, "", fmt.Sprintf("implementation for '%s' does not implement providers.Provider", name), llm.CategoryInvalidRequest)
	}
	return providerImpl, nil
}

// GetMethodCalls returns the method calls for a specific method
func (m *BoundaryMockRegistry) GetMethodCalls(method string) []BoundaryMethodCall {
	return m.methodCalls[method]
}

// GetCallSequence returns the sequence of method calls
func (m *BoundaryMockRegistry) GetCallSequence() []string {
	return m.callSequence
}

// ClearMethodCalls clears the recorded method calls
func (m *BoundaryMockRegistry) ClearMethodCalls() {
	m.methodCalls = make(map[string][]BoundaryMethodCall)
	m.callSequence = nil
}

// ProviderImplementation is now deprecated, use providers.Provider instead
type ProviderImplementation interface {
	CreateClient(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error)
}

// BoundaryMockProvider implements providers.Provider for boundary testing
type BoundaryMockProvider struct {
	CreateClientResult llm.LLMClient
	CreateClientError  error
	CreateClientCalls  []CreateClientCall
}

type CreateClientCall struct {
	Ctx         context.Context
	APIKey      string
	ModelID     string
	APIEndpoint string
}

// CreateClient implements the Provider interface
func (m *BoundaryMockProvider) CreateClient(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error) {
	// Record the call
	m.CreateClientCalls = append(m.CreateClientCalls, CreateClientCall{
		Ctx:         ctx,
		APIKey:      apiKey,
		ModelID:     modelID,
		APIEndpoint: apiEndpoint,
	})

	if m.CreateClientError != nil {
		return nil, m.CreateClientError
	}
	return m.CreateClientResult, nil
}

// BoundaryMockLLMClient implements llm.LLMClient for boundary testing
type BoundaryMockLLMClient struct {
	ModelName            string
	GenerateContentCalls []GenerateContentCall
}

type GenerateContentCall struct {
	Ctx    context.Context
	Prompt string
	Params map[string]interface{}
}

// GenerateContent implements llm.LLMClient
func (m *BoundaryMockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// Record the call
	m.GenerateContentCalls = append(m.GenerateContentCalls, GenerateContentCall{
		Ctx:    ctx,
		Prompt: prompt,
		Params: params,
	})

	return &llm.ProviderResult{
		Content: "mock response",
	}, nil
}

// GetModelName implements llm.LLMClient
func (m *BoundaryMockLLMClient) GetModelName() string {
	return m.ModelName
}

// Close implements llm.LLMClient
func (m *BoundaryMockLLMClient) Close() error {
	return nil
}

// mockAPIService is a mock implementation of the APIService for testing
type mockAPIService struct {
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	processLLMResponseFunc   func(result *llm.ProviderResult) (string, error)
}

func (m *mockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.processLLMResponseFunc != nil {
		return m.processLLMResponseFunc(result)
	}
	return "", errors.New("not implemented")
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	if m.isEmptyResponseErrorFunc != nil {
		return m.isEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	if m.isSafetyBlockedErrorFunc != nil {
		return m.isSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return "not implemented"
}

func (m *mockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	// For testing fallback
	return map[string]interface{}{}, nil
}

func (m *mockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	// For testing fallback
	return true, nil
}

func (m *mockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*registry.ModelDefinition, error) {
	// This should still return an error for the test case
	return nil, errors.New("not implemented")
}

func (m *mockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Implement token limit fallbacks based on model name
	switch modelName {
	case "gemini-1.5-pro-latest":
		return 1000000, 8192, nil
	case "gemini-1.0-pro":
		return 32768, 8192, nil
	case "gpt-4o":
		return 128000, 4096, nil
	case "gpt-4":
		return 8192, 4096, nil
	case "gpt-3.5-turbo":
		return 16385, 4096, nil
	default:
		return 200000, 4096, nil
	}
}
