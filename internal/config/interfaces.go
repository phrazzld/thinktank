// Package config provides configuration management for the architect application
package config

// ManagerInterface defines the interface for application configuration management
type ManagerInterface interface {
	// GetConfig returns the current configuration
	GetConfig() *AppConfig

	// GetUserConfigDir returns the user-specific configuration directory
	GetUserConfigDir() string

	// GetSystemConfigDirs returns the system-wide configuration directories
	GetSystemConfigDirs() []string

	// GetUserTemplateDir returns the directory for user-specific templates
	GetUserTemplateDir() string

	// GetSystemTemplateDirs returns the system-wide template directories
	GetSystemTemplateDirs() []string

	// GetConfigDirs returns all configuration directories
	GetConfigDirs() ConfigDirectories

	// GetTemplatePath finds the path to a template file using the configured precedence
	GetTemplatePath(name string) (string, error)

	// LoadFromFiles loads configuration from files (user, system) according to precedence
	LoadFromFiles() error

	// MergeWithFlags merges loaded configuration with command-line flags
	MergeWithFlags(cliFlags map[string]interface{}) error

	// EnsureConfigDirs creates necessary configuration directories if they don't exist
	EnsureConfigDirs() error

	// WriteDefaultConfig writes the default configuration to the user's config file
	WriteDefaultConfig() error
}
