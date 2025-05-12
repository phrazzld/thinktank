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
	getModelFunc                func(ctx context.Context, name string) (*registry.ModelDefinition, error)
	getProviderFunc             func(ctx context.Context, name string) (*registry.ProviderDefinition, error)
	getProviderImplFunc         func(ctx context.Context, name string) (providers.Provider, error)
	createLLMClientFunc         func(ctx context.Context, apiKey, modelName string) (llm.LLMClient, error)
	getAllModelNamesFunc        func(ctx context.Context) []string
	getModelNamesByProviderFunc func(ctx context.Context, providerName string) []string

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
	model, err := m.GetModelWithContext(ctx, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model definition: %w", err)
	}

	// Get provider definition
	providerName := model.Provider
	provider, err := m.GetProviderWithContext(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider definition: %w", err)
	}

	// Get provider implementation
	impl, err := m.GetProviderImplementationWithContext(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider implementation: %w", err)
	}

	// Create client
	return impl.CreateClient(ctx, apiKey, model.APIModelID, provider.BaseURL)
}

// GetAllModelNames implements overloaded registry method - see GetAllModelNamesWithContext for actual implementation

// GetModelNamesByProvider implements overloaded registry method - see GetModelNamesByProviderWithContext for actual implementation

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
func (m *MockRegistry) SetGetModelFunc(fn func(ctx context.Context, name string) (*registry.ModelDefinition, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getModelFunc = fn
}

// SetGetProviderFunc sets a custom implementation for GetProvider
func (m *MockRegistry) SetGetProviderFunc(fn func(ctx context.Context, name string) (*registry.ProviderDefinition, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getProviderFunc = fn
}

// SetGetProviderImplementationFunc sets a custom implementation for GetProviderImplementation
func (m *MockRegistry) SetGetProviderImplementationFunc(fn func(ctx context.Context, name string) (providers.Provider, error)) {
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
func (m *MockRegistry) SetGetAllModelNamesFunc(fn func(ctx context.Context) []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getAllModelNamesFunc = fn
}

// SetGetModelNamesByProviderFunc sets a custom implementation for GetModelNamesByProvider
func (m *MockRegistry) SetGetModelNamesByProviderFunc(fn func(ctx context.Context, providerName string) []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getModelNamesByProviderFunc = fn
}

// Backward-compatible methods for tests that haven't been updated yet

// GetModel implements backward compatibility with non-context API
func (m *MockRegistry) GetModel(nameOrCtx interface{}, name ...string) (*registry.ModelDefinition, error) {
	if ctx, ok := nameOrCtx.(context.Context); ok && len(name) > 0 {
		// New context-based API
		return m.GetModelWithContext(ctx, name[0])
	}
	// Legacy API
	if strName, ok := nameOrCtx.(string); ok {
		return m.GetModelWithContext(context.Background(), strName)
	}
	return nil, fmt.Errorf("invalid arguments to GetModel")
}

// GetModelWithContext is the internal implementation with context
func (m *MockRegistry) GetModelWithContext(ctx context.Context, name string) (*registry.ModelDefinition, error) {
	m.recordCall("GetModel", ctx, name)

	// Use custom function if provided
	if m.getModelFunc != nil {
		return m.getModelFunc(ctx, name)
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

// GetProvider implements backward compatibility with non-context API
func (m *MockRegistry) GetProvider(nameOrCtx interface{}, name ...string) (*registry.ProviderDefinition, error) {
	if ctx, ok := nameOrCtx.(context.Context); ok && len(name) > 0 {
		// New context-based API
		return m.GetProviderWithContext(ctx, name[0])
	}
	// Legacy API
	if strName, ok := nameOrCtx.(string); ok {
		return m.GetProviderWithContext(context.Background(), strName)
	}
	return nil, fmt.Errorf("invalid arguments to GetProvider")
}

// GetProviderWithContext is the internal implementation with context
func (m *MockRegistry) GetProviderWithContext(ctx context.Context, name string) (*registry.ProviderDefinition, error) {
	m.recordCall("GetProvider", ctx, name)

	// Use custom function if provided
	if m.getProviderFunc != nil {
		return m.getProviderFunc(ctx, name)
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

// RegisterProviderImplementation implements backward compatibility with non-context API
func (m *MockRegistry) RegisterProviderImplementation(ctxOrName interface{}, nameOrImpl interface{}, implOpt ...providers.Provider) error {
	if ctx, ok := ctxOrName.(context.Context); ok {
		// New context-based API
		if strName, ok := nameOrImpl.(string); ok && len(implOpt) > 0 {
			return m.RegisterProviderImplementationWithContext(ctx, strName, implOpt[0])
		}
		return fmt.Errorf("invalid arguments to RegisterProviderImplementation with context")
	}
	// Legacy API
	if strName, ok := ctxOrName.(string); ok {
		if impl, ok := nameOrImpl.(providers.Provider); ok {
			return m.RegisterProviderImplementationWithContext(context.Background(), strName, impl)
		}
	}
	return fmt.Errorf("invalid arguments to RegisterProviderImplementation")
}

// RegisterProviderImplementationWithContext is the internal implementation with context
func (m *MockRegistry) RegisterProviderImplementationWithContext(ctx context.Context, name string, impl providers.Provider) error {
	m.recordCall("RegisterProviderImplementation", ctx, name, impl)

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

// GetProviderImplementation implements backward compatibility with non-context API
func (m *MockRegistry) GetProviderImplementation(ctxOrName interface{}, name ...string) (providers.Provider, error) {
	if ctx, ok := ctxOrName.(context.Context); ok && len(name) > 0 {
		// New context-based API
		return m.GetProviderImplementationWithContext(ctx, name[0])
	}
	// Legacy API
	if strName, ok := ctxOrName.(string); ok {
		return m.GetProviderImplementationWithContext(context.Background(), strName)
	}
	return nil, fmt.Errorf("invalid arguments to GetProviderImplementation")
}

// GetProviderImplementationWithContext is the internal implementation with context
func (m *MockRegistry) GetProviderImplementationWithContext(ctx context.Context, name string) (providers.Provider, error) {
	m.recordCall("GetProviderImplementation", ctx, name)

	// Use custom function if provided
	if m.getProviderImplFunc != nil {
		return m.getProviderImplFunc(ctx, name)
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

// GetAllModelNames implements backward compatibility with non-context API
func (m *MockRegistry) GetAllModelNames(ctx ...context.Context) []string {
	var useCtx context.Context
	if len(ctx) > 0 {
		useCtx = ctx[0]
	} else {
		useCtx = context.Background()
	}
	return m.GetAllModelNamesWithContext(useCtx)
}

// GetAllModelNamesWithContext is the internal implementation with context
func (m *MockRegistry) GetAllModelNamesWithContext(ctx context.Context) []string {
	m.recordCall("GetAllModelNames", ctx)

	// Use custom function if provided
	if m.getAllModelNamesFunc != nil {
		return m.getAllModelNamesFunc(ctx)
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

// GetModelNamesByProvider implements backward compatibility with non-context API
func (m *MockRegistry) GetModelNamesByProvider(ctxOrName interface{}, providerName ...string) []string {
	if ctx, ok := ctxOrName.(context.Context); ok && len(providerName) > 0 {
		// New context-based API
		return m.GetModelNamesByProviderWithContext(ctx, providerName[0])
	}
	// Legacy API
	if strName, ok := ctxOrName.(string); ok {
		return m.GetModelNamesByProviderWithContext(context.Background(), strName)
	}
	return nil
}

// GetModelNamesByProviderWithContext is the internal implementation with context
func (m *MockRegistry) GetModelNamesByProviderWithContext(ctx context.Context, providerName string) []string {
	m.recordCall("GetModelNamesByProvider", ctx, providerName)

	// Use custom function if provided
	if m.getModelNamesByProviderFunc != nil {
		return m.getModelNamesByProviderFunc(ctx, providerName)
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
// This method supports both context and non-context invocations for backward compatibility
func (m *MockRegistry) AddProviderImplementation(ctxOrName interface{}, nameOrImpl interface{}, implOpt ...providers.Provider) {
	// Context-based API (from RegisterProviderImplementation)
	if _, ok := ctxOrName.(context.Context); ok {
		if strName, ok := nameOrImpl.(string); ok && len(implOpt) > 0 {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.implementations[strName] = implOpt[0]
			return
		}
	}

	// Legacy API
	if strName, ok := ctxOrName.(string); ok {
		if impl, ok := nameOrImpl.(providers.Provider); ok {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.implementations[strName] = impl
			return
		}
	}
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
