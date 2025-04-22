// Package registry provides a configuration-driven registry
// for LLM providers and models, allowing for flexible configuration
// and easier addition of new models and providers.
package registry

import (
	"fmt"
	"os"
	"path/filepath"

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
}

// Compile-time check to ensure ConfigLoader implements ConfigLoaderInterface
var _ ConfigLoaderInterface = (*ConfigLoader)(nil)

// NewConfigLoader creates a new ConfigLoader
func NewConfigLoader() *ConfigLoader {
	loader := &ConfigLoader{}

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
	configPath, err := c.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine configuration path: %w", err)
	}

	// Log the configuration file path being used
	fmt.Printf("Loading model configuration from: %s\n", configPath)

	// Read the configuration file
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
	fmt.Printf("Read configuration file (%d bytes)\n", len(data))

	// Parse the YAML
	var config ModelsConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML in configuration file at %s: %w", configPath, err)
	}

	// Log successful parsing
	fmt.Printf("Successfully parsed YAML configuration\n")

	// Validate the configuration
	if err := c.validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Log validation success
	fmt.Printf("Configuration validated successfully: %d providers, %d models defined\n",
		len(config.Providers), len(config.Models))

	return &config, nil
}

// validate performs comprehensive validation of the configuration
func (c *ConfigLoader) validate(config *ModelsConfig) error {
	// Check API key sources
	if len(config.APIKeySources) == 0 {
		return fmt.Errorf("configuration must include api_key_sources")
	}
	fmt.Printf("Validated API key sources: %d sources defined\n", len(config.APIKeySources))

	// Display found API key sources for debugging
	for provider, envVar := range config.APIKeySources {
		// Check if the environment variable exists (without revealing its value)
		_, exists := os.LookupEnv(envVar)
		if exists {
			fmt.Printf("✓ API key for provider '%s' found in environment variable %s\n", provider, envVar)
		} else {
			fmt.Printf("⚠ API key for provider '%s' not found in environment variable %s\n", provider, envVar)
		}
	}

	// Check providers
	if len(config.Providers) == 0 {
		return fmt.Errorf("configuration must include at least one provider")
	}
	fmt.Printf("Validating %d providers...\n", len(config.Providers))

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
			fmt.Printf("Provider '%s' configured with custom base URL: %s\n", provider.Name, provider.BaseURL)
		} else {
			fmt.Printf("Provider '%s' configured with default base URL\n", provider.Name)
		}
	}

	// Check models
	if len(config.Models) == 0 {
		return fmt.Errorf("configuration must include at least one model")
	}
	fmt.Printf("Validating %d models...\n", len(config.Models))

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
			fmt.Printf("⚠ Warning: Model '%s' has no parameters defined\n", model.Name)
		} else {
			// Log parameters with invalid/suspicious values
			for paramName, paramDef := range model.Parameters {
				// Validate parameter type
				if paramDef.Type == "" {
					fmt.Printf("⚠ Warning: Parameter '%s' for model '%s' is missing type\n", paramName, model.Name)
				}

				// Check for default value presence
				if paramDef.Default == nil {
					fmt.Printf("⚠ Warning: Parameter '%s' for model '%s' has no default value\n", paramName, model.Name)
				}

				// Check numeric constraints for consistency
				if (paramDef.Type == "float" || paramDef.Type == "int") &&
					paramDef.Min != nil && paramDef.Max != nil {
					// Attempt type assertion - this is simplified and may need refinement
					minFloat, minOk := paramDef.Min.(float64)
					maxFloat, maxOk := paramDef.Max.(float64)
					if minOk && maxOk && minFloat > maxFloat {
						fmt.Printf("⚠ Warning: Parameter '%s' for model '%s' has min (%v) > max (%v)\n",
							paramName, model.Name, paramDef.Min, paramDef.Max)
					}
				}
			}
		}

		// Log successful model validation
		fmt.Printf("✓ Validated model '%s' (provider: '%s')\n",
			model.Name, model.Provider)
	}

	return nil
}
