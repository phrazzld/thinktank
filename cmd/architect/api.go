// Package architect provides the command-line interface for the architect tool
package architect

import (
	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
)

// Re-export error types from internal/architect for backward compatibility with tests
var (
	ErrEmptyResponse        = architect.ErrEmptyResponse
	ErrWhitespaceContent    = architect.ErrWhitespaceContent
	ErrSafetyBlocked        = architect.ErrSafetyBlocked
	ErrAPICall              = architect.ErrAPICall
	ErrClientInitialization = architect.ErrClientInitialization
)

// APIService is an alias to the interfaces one
type APIService = interfaces.APIService

// NewAPIService is a wrapper for the internal one
// It uses the registry-based implementation for better flexibility
func NewAPIService(logger logutil.LoggerInterface) APIService {
	registryManager := registry.GetGlobalManager(logger)
	return architect.NewRegistryAPIService(registryManager, logger)
}
