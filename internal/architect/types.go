// Package architect contains the core application logic for the architect tool
package architect

import (
	// Import the config package which contains the canonical CliConfig
	_ "github.com/phrazzld/architect/internal/config"
)

// Configuration constants have been moved to internal/config
const (
	// APIKeyEnvVar is the environment variable name for the API key
	// Kept for backward compatibility but prefer using config.APIKeyEnvVar
	APIKeyEnvVar = "GEMINI_API_KEY"
)
