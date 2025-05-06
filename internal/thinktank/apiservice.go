// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// APIService is a type alias for interfaces.APIService to avoid circular imports
// in tests while still allowing the use of the interface.
type APIService = interfaces.APIService
