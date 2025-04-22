// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// Re-export error types from internal/thinktank for backward compatibility with tests
var (
	ErrEmptyResponse        = thinktank.ErrEmptyResponse
	ErrWhitespaceContent    = thinktank.ErrWhitespaceContent
	ErrSafetyBlocked        = thinktank.ErrSafetyBlocked
	ErrAPICall              = thinktank.ErrAPICall
	ErrClientInitialization = thinktank.ErrClientInitialization
)

// APIService is an alias to the interfaces one
type APIService = interfaces.APIService

// NewAPIService is a wrapper for the internal one
// It uses the registry-based implementation for better flexibility
func NewAPIService(logger logutil.LoggerInterface) APIService {
	registryManager := registry.GetGlobalManager(logger)
	return thinktank.NewRegistryAPIService(registryManager, logger)
}
