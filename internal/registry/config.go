// Package registry provides a configuration-driven registry
// for LLM providers and models, allowing for flexible configuration
// and easier addition of new models and providers.
package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/phrazzld/thinktank/internal/logutil"
	"gopkg.in/yaml.v3"
)

const (
	// ConfigDirName is the name of the configuration directory
	ConfigDirName = ".config/thinktank"
	// ModelsConfigFileName is the name of the models configuration file
	ModelsConfigFileName = "models.yaml"
)

// ConfigLoader is responsible for loading the models configuration
type ConfigLoader struct {
	// GetConfigPath is a function that returns the path to the models.yaml configuration file
	// It can be replaced in tests to return a test file path
	GetConfigPath func() (string, error)
	// Logger is used for logging
	Logger logutil.LoggerInterface
}

// Compile-time check to ensure ConfigLoader implements ConfigLoaderInterface
var _ ConfigLoaderInterface = (*ConfigLoader)(nil)

// NewConfigLoader creates a new ConfigLoader
func NewConfigLoader(logger logutil.LoggerInterface) *ConfigLoader {
	// If no logger is provided, create a default one
	if logger == nil {
		logger = logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	}

	loader := &ConfigLoader{
		Logger: logger,
	}

	// Set the default implementation of GetConfigPath
	loader.GetConfigPath = func() (string, error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		configDir := filepath.Join(homeDir, ConfigDirName)
		configPath := filepath.Join(configDir, ModelsConfigFileName)

		return configPath, nil
	}

	return loader
}

// Load reads and parses the models.yaml configuration file
func (c *ConfigLoader) Load() (*ModelsConfig, error) {
	// Create a background context for logging
	ctx := context.Background()

	configPath, err := c.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine configuration path: %w", err)
	}

	// Log the configuration file path being used
	c.Logger.InfoContext(ctx, "Loading model configuration from: %s", configPath)

	// Read the configuration file
	//nolint:gosec // G304: Config file path validated by GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found at %s: %w", configPath, err)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading configuration file at %s: %w", configPath, err)
		}
		return nil, fmt.Errorf("error reading configuration file at %s: %w", configPath, err)
	}

	// Log configuration file size
	c.Logger.InfoContext(ctx, "Read configuration file (%d bytes)", len(data))

	// Parse the YAML
	var config ModelsConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML in configuration file at %s: %w", configPath, err)
	}

	// Log successful parsing
	c.Logger.InfoContext(ctx, "Successfully parsed YAML configuration")

	// Validate the configuration
	if err := c.validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log validation success
	c.Logger.InfoContext(ctx, "Configuration validated successfully: %d providers, %d models defined",
		len(config.Providers), len(config.Models))

	return &config, nil
}

// validate performs comprehensive validation of the configuration
func (c *ConfigLoader) validate(config *ModelsConfig) error {
	// Create a background context for logging
	ctx := context.Background()

	// Check API key sources
	if len(config.APIKeySources) == 0 {
		return fmt.Errorf("configuration must include api_key_sources")
	}
	c.Logger.InfoContext(ctx, "Validated API key sources: %d sources defined", len(config.APIKeySources))

	// Display found API key sources for debugging
	for provider, envVar := range config.APIKeySources {
		// Check if the environment variable exists (without revealing its value)
		_, exists := os.LookupEnv(envVar)
		if exists {
			c.Logger.InfoContext(ctx, "✓ API key for provider '%s' found in environment variable %s", provider, envVar)
		} else {
			c.Logger.WarnContext(ctx, "⚠ API key for provider '%s' not found in environment variable %s", provider, envVar)
		}
	}

	// Check providers
	if len(config.Providers) == 0 {
		return fmt.Errorf("configuration must include at least one provider")
	}
	c.Logger.InfoContext(ctx, "Validating %d providers...", len(config.Providers))

	// Check for provider name uniqueness
	providerNames := make(map[string]bool)
	for i, provider := range config.Providers {
		if provider.Name == "" {
			return fmt.Errorf("provider at index %d is missing name", i)
		}
		if providerNames[provider.Name] {
			return fmt.Errorf("duplicate provider name '%s' detected", provider.Name)
		}
		providerNames[provider.Name] = true

		// Log provider details
		if provider.BaseURL != "" {
			c.Logger.InfoContext(ctx, "Provider '%s' configured with custom base URL: %s", provider.Name, provider.BaseURL)
		} else {
			c.Logger.InfoContext(ctx, "Provider '%s' configured with default base URL", provider.Name)
		}
	}

	// Check models
	if len(config.Models) == 0 {
		return fmt.Errorf("configuration must include at least one model")
	}
	c.Logger.InfoContext(ctx, "Validating %d models...", len(config.Models))

	// Check for model name uniqueness
	modelNames := make(map[string]bool)

	for i, model := range config.Models {
		// Validate model name
		if model.Name == "" {
			return fmt.Errorf("model at index %d is missing name", i)
		}
		if modelNames[model.Name] {
			return fmt.Errorf("duplicate model name '%s' detected", model.Name)
		}
		modelNames[model.Name] = true

		// Validate provider
		if model.Provider == "" {
			return fmt.Errorf("model '%s' is missing provider", model.Name)
		}
		if !providerNames[model.Provider] {
			return fmt.Errorf("model '%s' references unknown provider '%s'", model.Name, model.Provider)
		}

		// Validate API model ID
		if model.APIModelID == "" {
			return fmt.Errorf("model '%s' is missing api_model_id", model.Name)
		}

		// Token validation removed as part of T036C
		// The token validation logic has been removed since the token-related fields
		// have been removed from the ModelDefinition struct.
		// Token handling is now the responsibility of each provider.

		// Token warnings removed as part of T036C
		// These warnings have been removed since token-related fields were removed.

		// Parameter validation
		if len(model.Parameters) == 0 {
			c.Logger.WarnContext(ctx, "⚠ Warning: Model '%s' has no parameters defined", model.Name)
		} else {
			// Log parameters with invalid/suspicious values
			for paramName, paramDef := range model.Parameters {
				// Validate parameter type
				if paramDef.Type == "" {
					c.Logger.WarnContext(ctx, "⚠ Warning: Parameter '%s' for model '%s' is missing type", paramName, model.Name)
				}

				// Check for default value presence
				if paramDef.Default == nil {
					c.Logger.WarnContext(ctx, "⚠ Warning: Parameter '%s' for model '%s' has no default value", paramName, model.Name)
				}

				// Check numeric constraints for consistency
				if (paramDef.Type == "float" || paramDef.Type == "int") &&
					paramDef.Min != nil && paramDef.Max != nil {
					// Attempt type assertion - this is simplified and may need refinement
					minFloat, minOk := paramDef.Min.(float64)
					maxFloat, maxOk := paramDef.Max.(float64)
					if minOk && maxOk && minFloat > maxFloat {
						c.Logger.WarnContext(ctx, "⚠ Warning: Parameter '%s' for model '%s' has min (%v) > max (%v)",
							paramName, model.Name, paramDef.Min, paramDef.Max)
					}
				}
			}
		}

		// Log successful model validation
		c.Logger.InfoContext(ctx, "✓ Validated model '%s' (provider: '%s')",
			model.Name, model.Provider)
	}

	return nil
}
