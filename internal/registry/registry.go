// Package registry provides a configuration-driven registry
// for LLM providers and models, allowing for flexible configuration
// and easier addition of new models and providers.
package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/providers"
)

// Registry holds loaded provider and model definitions and their implementations.
// It provides thread-safe access to the configuration data and registered provider implementations.
type Registry struct {
	models          map[string]ModelDefinition
	providers       map[string]ProviderDefinition
	implementations map[string]providers.Provider
	mu              sync.RWMutex
	logger          logutil.LoggerInterface
}

// NewRegistry creates a new Registry instance with initialized maps and the provided logger.
func NewRegistry(logger logutil.LoggerInterface) *Registry {
	return &Registry{
		models:          make(map[string]ModelDefinition),
		providers:       make(map[string]ProviderDefinition),
		implementations: make(map[string]providers.Provider),
		logger:          logger,
	}
}

// ConfigLoaderInterface defines the interface for configuration loaders
type ConfigLoaderInterface interface {
	Load() (*ModelsConfig, error)
}

// LoadConfig loads and validates the models configuration using the provided ConfigLoader.
// It populates the Registry with models and providers from the configuration.
func (r *Registry) LoadConfig(loader ConfigLoaderInterface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Debug("Loading models configuration")

	config, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Clear existing maps and populate with new data
	for _, p := range config.Providers {
		r.providers[p.Name] = p
	}

	for _, m := range config.Models {
		r.models[m.Name] = m
	}

	r.logger.Info("Models configuration loaded: %d providers, %d models",
		len(r.providers), len(r.models))

	return nil
}

// GetModel retrieves a model definition by name.
// Returns an error if the model is not found.
func (r *Registry) GetModel(name string) (*ModelDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, ok := r.models[name]
	if !ok {
		return nil, fmt.Errorf("model '%s' not found in registry", name)
	}
	return &model, nil
}

// GetProvider retrieves a provider definition by name.
// Returns an error if the provider is not found.
func (r *Registry) GetProvider(name string) (*ProviderDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found in registry", name)
	}
	return &provider, nil
}

// RegisterProviderImplementation registers a provider implementation.
// The implementation must correspond to a provider defined in the configuration.
func (r *Registry) RegisterProviderImplementation(name string, impl providers.Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify the provider exists in the configuration
	_, ok := r.providers[name]
	if !ok {
		return fmt.Errorf("provider '%s' not defined in configuration", name)
	}

	r.implementations[name] = impl
	r.logger.Debug("Registered provider implementation for '%s'", name)
	return nil
}

// GetProviderImplementation retrieves a registered provider implementation.
// Returns an error if the implementation is not registered.
func (r *Registry) GetProviderImplementation(name string) (providers.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	impl, ok := r.implementations[name]
	if !ok {
		return nil, fmt.Errorf("provider implementation for '%s' not registered", name)
	}
	return impl, nil
}

// CreateLLMClient creates an LLM client for the specified model using the registry.
// This is a convenience method that coordinates the retrieval of model and provider information
// and uses the registered provider implementation to create an appropriate client.
func (r *Registry) CreateLLMClient(ctx context.Context, apiKey, modelName string) (llm.LLMClient, error) {
	// Get the model definition
	model, err := r.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model definition: %w", err)
	}

	// Get the provider definition
	provider, err := r.GetProvider(model.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider definition: %w", err)
	}

	// Get the provider implementation
	impl, err := r.GetProviderImplementation(model.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider implementation: %w", err)
	}

	// Create the client using the provider implementation
	r.logger.Debug("Creating LLM client for model '%s' using provider '%s'",
		modelName, model.Provider)

	return impl.CreateClient(ctx, apiKey, model.APIModelID, provider.BaseURL)
}
