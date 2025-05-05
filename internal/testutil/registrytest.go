// Package testutil provides utilities for testing in the thinktank project
package testutil

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

// RegistryTestSetup contains all the components for a registry API test
type RegistryTestSetup struct {
	Registry        *MockRegistry
	Logger          *MockLogger
	TestClient      *MockLLMClient
	FailureClient   *MockLLMClient
	ModelName       string
	ProviderName    string
	CorrelationID   string
	ConfiguredError error
}

// MockLLMClient is a mock implementation of llm.LLMClient for testing
type MockLLMClient struct {
	// ModelName for GetModelName method to return
	ModelName string

	// Response for GenerateContent to return
	Response *llm.ProviderResult

	// Error for GenerateContent to return (nil for success)
	Error error

	// Record of calls to GenerateContent
	Calls []MockGenerateContentCall
}

// MockGenerateContentCall records the parameters used in a GenerateContent call
type MockGenerateContentCall struct {
	Context context.Context
	Prompt  string
	Params  map[string]interface{}
}

// GenerateContent implements llm.LLMClient.GenerateContent
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// Record the call
	m.Calls = append(m.Calls, MockGenerateContentCall{
		Context: ctx,
		Prompt:  prompt,
		Params:  params,
	})

	// Return configured response/error
	return m.Response, m.Error
}

// GetModelName implements llm.LLMClient.GetModelName
func (m *MockLLMClient) GetModelName() string {
	return m.ModelName
}

// Close implements llm.LLMClient.Close
func (m *MockLLMClient) Close() error {
	return nil
}

// NewMockLLMClient creates a new mock LLM client with the given parameters
func NewMockLLMClient(modelName string, response *llm.ProviderResult, err error) *MockLLMClient {
	return &MockLLMClient{
		ModelName: modelName,
		Response:  response,
		Error:     err,
		Calls:     make([]MockGenerateContentCall, 0),
	}
}

// MockProvider is a provider for testing registry API interactions
type MockProvider struct {
	// Name is the provider name
	Name string

	// Client to return from CreateClient
	Client *MockLLMClient

	// Error to return from CreateClient
	Error error

	// Record of calls to CreateClient
	Calls []MockCreateClientCall
}

// MockCreateClientCall records the parameters used in a CreateClient call
type MockCreateClientCall struct {
	Context     context.Context
	APIKey      string
	ModelID     string
	APIEndpoint string
}

// CreateClient implements providers.Provider.CreateClient
func (m *MockProvider) CreateClient(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error) {
	// Record the call
	m.Calls = append(m.Calls, MockCreateClientCall{
		Context:     ctx,
		APIKey:      apiKey,
		ModelID:     modelID,
		APIEndpoint: apiEndpoint,
	})

	// Return configured client/error
	return m.Client, m.Error
}

// NewMockProvider creates a new mock provider with the given parameters
func NewMockProvider(name string, client *MockLLMClient, err error) *MockProvider {
	return &MockProvider{
		Name:   name,
		Client: client,
		Error:  err,
		Calls:  make([]MockCreateClientCall, 0),
	}
}

// NewRegistryTestSetup creates a complete test setup for registry API testing
func NewRegistryTestSetup() *RegistryTestSetup {
	// Create a context with correlation ID for testing
	correlationID := "test-correlation-id"

	// Create standard model and provider names
	modelName := "test-model"
	providerName := "test-provider"

	// Create mock clients
	successClient := NewMockLLMClient(modelName, BasicSuccessResponse, nil)
	failureClient := NewMockLLMClient(modelName, nil, llm.ErrSafetyBlocked)

	// Create a new registry with standard test fixtures
	registry := NewMockRegistry()

	// Create a new logger
	logger := NewMockLogger()

	return &RegistryTestSetup{
		Registry:      registry,
		Logger:        logger,
		TestClient:    successClient,
		FailureClient: failureClient,
		ModelName:     modelName,
		ProviderName:  providerName,
		CorrelationID: correlationID,
	}
}

// SetupStandardRegistry configures the mock registry with standard test models and providers
func (s *RegistryTestSetup) SetupStandardRegistry() {
	// Add standard test model
	s.Registry.AddModel(registry.ModelDefinition{
		Name:       s.ModelName,
		Provider:   s.ProviderName,
		APIModelID: "api-" + s.ModelName,
		Parameters: StandardParameters,
	})

	// Add standard test provider
	s.Registry.AddProvider(registry.ProviderDefinition{
		Name:    s.ProviderName,
		BaseURL: "https://test.example.com",
	})

	// Create and add provider implementation
	mockProvider := NewMockProvider(
		s.ProviderName,
		s.TestClient,
		nil,
	)
	s.Registry.AddProviderImplementation(s.ProviderName, mockProvider)
}

// SetupErrorCase configures the registry to return the specified error
func (s *RegistryTestSetup) SetupErrorCase(errorType string) {
	// Configure the error based on the type
	switch errorType {
	case "model_not_found":
		s.Registry.SetGetModelError(errors.New("model not found"))
		s.ConfiguredError = llm.ErrModelNotFound

	case "provider_not_found":
		s.Registry.SetGetProviderError(errors.New("provider not found"))
		s.ConfiguredError = errors.New("provider not found")

	case "implementation_not_found":
		s.Registry.SetGetProviderImplementationError(errors.New("provider implementation not found"))
		s.ConfiguredError = errors.New("provider implementation not found")

	case "client_creation_failed":
		// Update the provider to return an error
		provider := NewMockProvider(
			s.ProviderName,
			nil,
			llm.ErrClientInitialization,
		)
		s.Registry.AddProviderImplementation(s.ProviderName, provider)
		s.ConfiguredError = llm.ErrClientInitialization

	case "safety_blocked":
		// Update the provider to return a client that returns safety blocked errors
		provider := NewMockProvider(
			s.ProviderName,
			s.FailureClient,
			nil,
		)
		s.Registry.AddProviderImplementation(s.ProviderName, provider)
		s.ConfiguredError = llm.ErrSafetyBlocked

	default:
		s.Registry.SetGetModelError(errors.New("unknown error"))
		s.ConfiguredError = errors.New("unknown error")
	}
}

// CreateTestContext creates a context with the test correlation ID
func (s *RegistryTestSetup) CreateTestContext() context.Context {
	return context.Background()
}

// AssertRegistryMethodCalled checks that a specified registry method was called
func AssertRegistryMethodCalled(t testing.TB, registry *MockRegistry, methodName string) {
	t.Helper()
	calls := registry.GetMethodCalls(methodName)
	if len(calls) == 0 {
		t.Errorf("Expected registry method %s to be called, but it wasn't", methodName)
	}
}

// AssertRegistryMethodNotCalled checks that a specified registry method was not called
func AssertRegistryMethodNotCalled(t testing.TB, registry *MockRegistry, methodName string) {
	t.Helper()
	calls := registry.GetMethodCalls(methodName)
	if len(calls) > 0 {
		t.Errorf("Expected registry method %s not to be called, but it was called %d times", methodName, len(calls))
	}
}

// AssertRegistryMethodCalledWith checks that a registry method was called with specific arguments
func AssertRegistryMethodCalledWith(t testing.TB, registry *MockRegistry, methodName string, expectedArg interface{}) {
	t.Helper()
	calls := registry.GetMethodCalls(methodName)
	if len(calls) == 0 {
		t.Errorf("Expected registry method %s to be called, but it wasn't", methodName)
		return
	}

	found := false
	for _, call := range calls {
		for _, arg := range call.Args {
			if fmt.Sprintf("%v", arg) == fmt.Sprintf("%v", expectedArg) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Errorf("Expected registry method %s to be called with %v, but it wasn't", methodName, expectedArg)
	}
}

// AssertRegistryMethodCalledTimes checks that a registry method was called a specific number of times
func AssertRegistryMethodCalledTimes(t testing.TB, registry *MockRegistry, methodName string, expectedCalls int) {
	t.Helper()
	calls := registry.GetMethodCalls(methodName)
	if len(calls) != expectedCalls {
		t.Errorf("Expected registry method %s to be called %d times, but it was called %d times",
			methodName, expectedCalls, len(calls))
	}
}

// AssertRegistryCallOrder checks that registry methods were called in the expected order
func AssertRegistryCallOrder(t testing.TB, registry *MockRegistry, expectedOrder []string) {
	t.Helper()
	sequence := registry.GetCallSequence()

	// Check that sequence contains at least the expected calls
	if len(sequence) < len(expectedOrder) {
		t.Errorf("Expected at least %d registry calls, got %d", len(expectedOrder), len(sequence))
		return
	}

	// Check each expected method in order
	var lastIndex = -1
	for i, method := range expectedOrder {
		// Find this method in the sequence
		found := false
		for j := lastIndex + 1; j < len(sequence); j++ {
			if sequence[j] == method {
				found = true
				lastIndex = j
				break
			}
		}

		if !found {
			t.Errorf("Expected registry method %s at position %d in call sequence, not found after position %d",
				method, i, lastIndex)
			break
		}
	}
}

// AssertErrorMatches checks if an error matches the expected one
func AssertErrorMatches(t testing.TB, err error, expected error) {
	t.Helper()
	if err == nil && expected != nil {
		t.Errorf("Expected error %v, got nil", expected)
		return
	}
	if err != nil && expected == nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	if err != nil && expected != nil && !errors.Is(err, expected) {
		t.Errorf("Expected error %v, got %v", expected, err)
	}
}

// CreatePreConfiguredMockRegistry returns a mock registry pre-configured with standard test models and providers
func CreatePreConfiguredMockRegistry() *MockRegistry {
	registry := NewMockRegistry()

	// Add standard models
	for _, model := range CreateTestModels() {
		registry.AddModel(model)
	}

	// Add standard providers
	for _, provider := range CreateTestProviders() {
		registry.AddProvider(provider)
	}

	return registry
}

// CreateSuccessProviderImpl creates a provider implementation that returns a success response
func CreateSuccessProviderImpl(providerName string, content string) *MockProvider {
	response := CreateSuccessResponse(content)
	client := NewMockLLMClient(providerName+"-model", response, nil)
	return NewMockProvider(providerName, client, nil)
}

// CreateErrorProviderImpl creates a provider implementation that returns an error
func CreateErrorProviderImpl(providerName string, err error) *MockProvider {
	client := NewMockLLMClient(providerName+"-model", nil, err)
	return NewMockProvider(providerName, client, nil)
}

// CreateClientCreationErrorProviderImpl creates a provider implementation that fails during client creation
func CreateClientCreationErrorProviderImpl(providerName string, err error) *MockProvider {
	return NewMockProvider(providerName, nil, err)
}

// GetTestProviderMap returns a map of provider names to mock providers for different test scenarios
func GetTestProviderMap() map[string]*MockProvider {
	testProviders := map[string]*MockProvider{
		"success": CreateSuccessProviderImpl("success", "Success response"),
		"empty": NewMockProvider(
			"empty",
			NewMockLLMClient("empty-model", EmptyResponse, nil),
			nil,
		),
		"safety_blocked": NewMockProvider(
			"safety_blocked",
			NewMockLLMClient("blocked-model", nil, llm.ErrSafetyBlocked),
			nil,
		),
		"client_error": NewMockProvider(
			"client_error",
			nil,
			llm.ErrClientInitialization,
		),
		"rate_limited": NewMockProvider(
			"rate_limited",
			NewMockLLMClient("rate-limited-model", nil, CreateRateLimitError("test").Original),
			nil,
		),
	}

	return testProviders
}
