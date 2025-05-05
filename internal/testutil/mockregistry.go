// Package testutil provides utilities for testing in the thinktank project
package testutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers"
	"github.com/phrazzld/thinktank/internal/registry"
)

// Test-only type definitions to avoid dependencies in tests
type testModelDefinition struct {
	Name       string
	Provider   string
	APIModelID string
}

type testProviderDefinition struct {
	Name    string
	BaseURL string
}

// MockRegistry is a mock implementation of the registry.Registry interface for testing.
// It provides:
// 1. Full implementation of all Registry methods with configurable behaviors
// 2. Call tracking for verification in tests
// 3. Customizable error responses for testing error handling
// 4. Support for custom function implementations to override default behavior
//
// Usage:
//
//	// Create a registry with basic functionality
//	registry := testutil.NewMockRegistry()
//
//	// Add models and providers
//	registry.AddModel(registry.ModelDefinition{
//	    Name:       "test-model",
//	    Provider:   "test-provider",
//	    APIModelID: "test-api-id",
//	})
//
//	// Set up custom error cases
//	registry.SetGetModelError(errors.New("model not found"))
//
//	// Track method calls
//	calls := registry.GetMethodCalls("GetModel")
//	sequence := registry.GetCallSequence()
type MockRegistry struct {
	// Storage
	models          map[string]registry.ModelDefinition
	providers       map[string]registry.ProviderDefinition
	implementations map[string]providers.Provider

	// Call tracking
	methodCalls  map[string][]MockMethodCall
	callSequence []string

	// Error cases
	getModelErr                  error
	getProviderErr               error
	getProviderImplementationErr error
	registerImplErr              error
	createLLMClientErr           error

	// Custom behavior
	getModelFunc                func(name string) (*registry.ModelDefinition, error)
	getProviderFunc             func(name string) (*registry.ProviderDefinition, error)
	getProviderImplFunc         func(name string) (providers.Provider, error)
	createLLMClientFunc         func(ctx context.Context, apiKey, modelName string) (llm.LLMClient, error)
	getAllModelNamesFunc        func() []string
	getModelNamesByProviderFunc func(providerName string) []string

	// Thread safety
	mu sync.RWMutex

	// Logger
	logger logutil.LoggerInterface
}

// MockMethodCall records information about a method call for verification
type MockMethodCall struct {
	Method string
	Args   []interface{}
}

// NewMockRegistry creates a new mock registry for testing
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		models:          make(map[string]registry.ModelDefinition),
		providers:       make(map[string]registry.ProviderDefinition),
		implementations: make(map[string]providers.Provider),
		methodCalls:     make(map[string][]MockMethodCall),
		callSequence:    make([]string, 0),
		logger:          NewMockLogger(),
	}
}

// recordCall records a method call for tracking
func (m *MockRegistry) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.methodCalls[method] = append(m.methodCalls[method], MockMethodCall{
		Method: method,
		Args:   args,
	})
	m.callSequence = append(m.callSequence, method)
}

// GetModel implements the Registry.GetModel method
func (m *MockRegistry) GetModel(name string) (*registry.ModelDefinition, error) {
	m.recordCall("GetModel", name)

	// Use custom function if provided
	if m.getModelFunc != nil {
		return m.getModelFunc(name)
	}

	// Return error if configured
	if m.getModelErr != nil {
		return nil, m.getModelErr
	}

	// Standard implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	model, ok := m.models[name]
	if !ok {
		return nil, fmt.Errorf("model '%s' not found in registry", name)
	}

	// Return a copy to prevent mutation
	modelCopy := model
	return &modelCopy, nil
}

// GetProvider implements the Registry.GetProvider method
func (m *MockRegistry) GetProvider(name string) (*registry.ProviderDefinition, error) {
	m.recordCall("GetProvider", name)

	// Use custom function if provided
	if m.getProviderFunc != nil {
		return m.getProviderFunc(name)
	}

	// Return error if configured
	if m.getProviderErr != nil {
		return nil, m.getProviderErr
	}

	// Standard implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found in registry", name)
	}

	// Return a copy to prevent mutation
	providerCopy := provider
	return &providerCopy, nil
}

// RegisterProviderImplementation implements the Registry.RegisterProviderImplementation method
func (m *MockRegistry) RegisterProviderImplementation(name string, impl providers.Provider) error {
	m.recordCall("RegisterProviderImplementation", name, impl)

	// Return error if configured
	if m.registerImplErr != nil {
		return m.registerImplErr
	}

	// Standard implementation
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if provider exists
	_, ok := m.providers[name]
	if !ok {
		return fmt.Errorf("provider '%s' not defined in configuration", name)
	}

	m.implementations[name] = impl
	return nil
}

// GetProviderImplementation implements the Registry.GetProviderImplementation method
func (m *MockRegistry) GetProviderImplementation(name string) (providers.Provider, error) {
	m.recordCall("GetProviderImplementation", name)

	// Use custom function if provided
	if m.getProviderImplFunc != nil {
		return m.getProviderImplFunc(name)
	}

	// Return error if configured
	if m.getProviderImplementationErr != nil {
		return nil, m.getProviderImplementationErr
	}

	// Standard implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	impl, ok := m.implementations[name]
	if !ok {
		return nil, fmt.Errorf("provider implementation for '%s' not registered", name)
	}

	return impl, nil
}

// CreateLLMClient implements the Registry.CreateLLMClient method
func (m *MockRegistry) CreateLLMClient(ctx context.Context, apiKey, modelName string) (llm.LLMClient, error) {
	m.recordCall("CreateLLMClient", ctx, apiKey, modelName)

	// Use custom function if provided
	if m.createLLMClientFunc != nil {
		return m.createLLMClientFunc(ctx, apiKey, modelName)
	}

	// Return error if configured
	if m.createLLMClientErr != nil {
		return nil, m.createLLMClientErr
	}

	// Standard implementation - validate API key
	if apiKey == "" {
		return nil, fmt.Errorf("empty API key provided for model '%s'", modelName)
	}

	// Get model definition
	model, err := m.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model definition: %w", err)
	}

	// Get provider definition
	providerName := model.Provider
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider definition: %w", err)
	}

	// Get provider implementation
	impl, err := m.GetProviderImplementation(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider implementation: %w", err)
	}

	// Create client
	return impl.CreateClient(ctx, apiKey, model.APIModelID, provider.BaseURL)
}

// GetAllModelNames implements the Registry.GetAllModelNames method
func (m *MockRegistry) GetAllModelNames() []string {
	m.recordCall("GetAllModelNames")

	// Use custom function if provided
	if m.getAllModelNamesFunc != nil {
		return m.getAllModelNamesFunc()
	}

	// Standard implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	modelNames := make([]string, 0, len(m.models))
	for name := range m.models {
		modelNames = append(modelNames, name)
	}

	return modelNames
}

// GetModelNamesByProvider implements the Registry.GetModelNamesByProvider method
func (m *MockRegistry) GetModelNamesByProvider(providerName string) []string {
	m.recordCall("GetModelNamesByProvider", providerName)

	// Use custom function if provided
	if m.getModelNamesByProviderFunc != nil {
		return m.getModelNamesByProviderFunc(providerName)
	}

	// Standard implementation
	m.mu.RLock()
	defer m.mu.RUnlock()

	var modelNames []string
	for name, model := range m.models {
		if model.Provider == providerName {
			modelNames = append(modelNames, name)
		}
	}

	return modelNames
}

// Helper methods for setting up test scenarios

// SetGetModelError sets the error to return from GetModel
func (m *MockRegistry) SetGetModelError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getModelErr = err
}

// SetGetProviderError sets the error to return from GetProvider
func (m *MockRegistry) SetGetProviderError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProviderErr = err
}

// SetGetProviderImplementationError sets the error to return from GetProviderImplementation
func (m *MockRegistry) SetGetProviderImplementationError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProviderImplementationErr = err
}

// SetRegisterProviderImplementationError sets the error to return from RegisterProviderImplementation
func (m *MockRegistry) SetRegisterProviderImplementationError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registerImplErr = err
}

// SetCreateLLMClientError sets the error to return from CreateLLMClient
func (m *MockRegistry) SetCreateLLMClientError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createLLMClientErr = err
}

// SetGetModelFunc sets a custom implementation for GetModel
func (m *MockRegistry) SetGetModelFunc(fn func(name string) (*registry.ModelDefinition, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getModelFunc = fn
}

// SetGetProviderFunc sets a custom implementation for GetProvider
func (m *MockRegistry) SetGetProviderFunc(fn func(name string) (*registry.ProviderDefinition, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProviderFunc = fn
}

// SetGetProviderImplementationFunc sets a custom implementation for GetProviderImplementation
func (m *MockRegistry) SetGetProviderImplementationFunc(fn func(name string) (providers.Provider, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProviderImplFunc = fn
}

// SetCreateLLMClientFunc sets a custom implementation for CreateLLMClient
func (m *MockRegistry) SetCreateLLMClientFunc(fn func(ctx context.Context, apiKey, modelName string) (llm.LLMClient, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createLLMClientFunc = fn
}

// SetGetAllModelNamesFunc sets a custom implementation for GetAllModelNames
func (m *MockRegistry) SetGetAllModelNamesFunc(fn func() []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getAllModelNamesFunc = fn
}

// SetGetModelNamesByProviderFunc sets a custom implementation for GetModelNamesByProvider
func (m *MockRegistry) SetGetModelNamesByProviderFunc(fn func(providerName string) []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getModelNamesByProviderFunc = fn
}

// AddModel adds a model to the mock registry
func (m *MockRegistry) AddModel(model interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Handle both registry.ModelDefinition and test-only types
	switch md := model.(type) {
	case registry.ModelDefinition:
		m.models[md.Name] = md
	case testModelDefinition:
		// Convert test type to registry type
		m.models[md.Name] = registry.ModelDefinition{
			Name:       md.Name,
			Provider:   md.Provider,
			APIModelID: md.APIModelID,
		}
	}
}

// AddProvider adds a provider to the mock registry
func (m *MockRegistry) AddProvider(provider interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Handle both registry.ProviderDefinition and test-only types
	switch pd := provider.(type) {
	case registry.ProviderDefinition:
		m.providers[pd.Name] = pd
	case testProviderDefinition:
		// Convert test type to registry type
		m.providers[pd.Name] = registry.ProviderDefinition{
			Name:    pd.Name,
			BaseURL: pd.BaseURL,
		}
	}
}

// AddProviderImplementation adds a provider implementation to the mock registry
func (m *MockRegistry) AddProviderImplementation(name string, impl providers.Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.implementations[name] = impl
}

// Call tracking helper methods

// GetMethodCalls returns the recorded calls for a specific method
func (m *MockRegistry) GetMethodCalls(method string) []MockMethodCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.methodCalls[method]
}

// GetCallSequence returns the sequence of method calls
func (m *MockRegistry) GetCallSequence() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callSequence
}

// ClearMethodCalls clears all recorded method calls
func (m *MockRegistry) ClearMethodCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.methodCalls = make(map[string][]MockMethodCall)
	m.callSequence = make([]string, 0)
}
