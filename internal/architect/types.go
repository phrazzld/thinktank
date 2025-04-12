// Package architect contains the core application logic for the architect tool
package architect

import "github.com/phrazzld/architect/internal/logutil"

// CliConfig represents the command-line configuration used by the application
type CliConfig struct {
	// Instructions configuration
	InstructionsFile string

	// Output configuration
	OutputDir    string
	AuditLogFile string // Path to write structured audit logs (JSON Lines)
	Format       string

	// Context gathering options
	Paths        []string
	Include      string
	Exclude      string
	ExcludeNames string
	DryRun       bool
	Verbose      bool

	// API configuration
	ApiKey     string
	ModelNames []string

	// Token management
	ConfirmTokens int

	// Logging
	LogLevel logutil.LogLevel
}

// Constants for configuration
const (
	// APIKeyEnvVar is the environment variable name for the API key
	APIKeyEnvVar = "GEMINI_API_KEY"
)
