// Package architect contains the core application logic for the architect tool
package architect

import "github.com/phrazzld/architect/internal/logutil"

// CliConfig represents the command-line configuration used by the application
type CliConfig struct {
	// Task configuration
	TaskDescription string
	TaskFile        string

	// Output configuration
	OutputFile string
	Format     string

	// Context gathering options
	Paths        []string
	Include      string
	Exclude      string
	ExcludeNames string
	DryRun       bool
	Verbose      bool

	// API configuration
	ApiKey    string
	ModelName string

	// Token management
	ConfirmTokens int

	// Template options
	Template     string
	ListExamples bool
	ShowExample  string

	// Logging
	LogLevel logutil.LogLevel
}

// Constants for configuration
const (
	// APIKeyEnvVar is the environment variable name for the API key
	APIKeyEnvVar = "GEMINI_API_KEY"
)
