// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"github.com/phrazzld/thinktank/internal/providers"
	"github.com/phrazzld/thinktank/internal/providers/provider"
)

// ProviderFactoryFunc creates a new provider implementation
type ProviderFactoryFunc func() provider.Provider

// ProviderRegistry allows registering and retrieving provider implementations
type ProviderRegistry interface {
	// RegisterProvider registers a provider implementation factory
	RegisterProvider(name string, factory ProviderFactoryFunc)

	// GetProvider returns a provider implementation
	GetProvider(name string) (provider.Provider, error)
}

// defaultProviderRegistry is the implementation of the ProviderRegistry interface
type defaultProviderRegistry struct {
	providers map[string]ProviderFactoryFunc
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() ProviderRegistry {
	return &defaultProviderRegistry{
		providers: make(map[string]ProviderFactoryFunc),
	}
}

// RegisterProvider registers a provider implementation factory
func (r *defaultProviderRegistry) RegisterProvider(name string, factory ProviderFactoryFunc) {
	r.providers[name] = factory
}

// GetProvider returns a provider implementation
func (r *defaultProviderRegistry) GetProvider(name string) (provider.Provider, error) {
	factory, ok := r.providers[name]
	if !ok {
		return nil, providers.ErrProviderNotFound
	}
	return factory(), nil
}
