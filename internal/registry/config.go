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
	ConfigDirName = ".config/architect"
	// ModelsConfigFileName is the name of the models configuration file
	ModelsConfigFileName = "models.yaml"
)

// ConfigLoader is responsible for loading the models configuration
type ConfigLoader struct {
	// GetConfigPath is a function that returns the path to the models.yaml configuration file
	// It can be replaced in tests to return a test file path
	GetConfigPath func() (string, error)
}

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
		return nil, err
	}

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

	// Parse the YAML
	var config ModelsConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML in configuration file at %s: %w", configPath, err)
	}

	// Validate the configuration
	if err := c.validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validate checks that the required fields in the configuration are present
func (c *ConfigLoader) validate(config *ModelsConfig) error {
	// Check API key sources
	if len(config.APIKeySources) == 0 {
		return fmt.Errorf("configuration must include api_key_sources")
	}

	// Check providers
	if len(config.Providers) == 0 {
		return fmt.Errorf("configuration must include at least one provider")
	}

	providerNames := make(map[string]bool)
	for i, provider := range config.Providers {
		if provider.Name == "" {
			return fmt.Errorf("provider at index %d is missing name", i)
		}
		providerNames[provider.Name] = true
	}

	// Check models
	if len(config.Models) == 0 {
		return fmt.Errorf("configuration must include at least one model")
	}

	for i, model := range config.Models {
		if model.Name == "" {
			return fmt.Errorf("model at index %d is missing name", i)
		}
		if model.Provider == "" {
			return fmt.Errorf("model %s is missing provider", model.Name)
		}
		if !providerNames[model.Provider] {
			return fmt.Errorf("model %s references unknown provider %s", model.Name, model.Provider)
		}
		if model.APIModelID == "" {
			return fmt.Errorf("model %s is missing api_model_id", model.Name)
		}
		if model.ContextWindow <= 0 {
			return fmt.Errorf("model %s has invalid context_window", model.Name)
		}
		if model.MaxOutputTokens <= 0 {
			return fmt.Errorf("model %s has invalid max_output_tokens", model.Name)
		}
	}

	return nil
}
