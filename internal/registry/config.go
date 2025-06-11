// Package registry provides a configuration-driven registry
// for LLM providers and models, allowing for flexible configuration
// and easier addition of new models and providers.
package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phrazzld/thinktank/internal/logutil"
	"gopkg.in/yaml.v3"
)

const (
	// ConfigDirName is the name of the configuration directory
	ConfigDirName = ".config/thinktank"
	// ModelsConfigFileName is the name of the models configuration file
	ModelsConfigFileName = "models.yaml"

	// Environment variable configuration
	// These environment variables can be used to override or supplement configuration
	// when models.yaml is missing or incomplete
	EnvConfigProvider      = "THINKTANK_CONFIG_PROVIDER"       // Default provider (e.g., "gemini", "openai")
	EnvConfigModel         = "THINKTANK_CONFIG_MODEL"          // Default model name
	EnvConfigAPIModelID    = "THINKTANK_CONFIG_API_MODEL_ID"   // API model ID for default model
	EnvConfigContextWindow = "THINKTANK_CONFIG_CONTEXT_WINDOW" // Context window for default model
	EnvConfigMaxOutput     = "THINKTANK_CONFIG_MAX_OUTPUT"     // Max output tokens for default model
	EnvConfigBaseURL       = "THINKTANK_CONFIG_BASE_URL"       // Custom base URL for provider
)

// getDefaultConfiguration returns a minimal default configuration
// that can be used when no configuration file is available and no environment variables are set.
// This ensures the application can run in containerized environments without external dependencies.
func getDefaultConfiguration() *ModelsConfig {
	return &ModelsConfig{
		APIKeySources: map[string]string{
			"openai":     "OPENAI_API_KEY",
			"gemini":     "GEMINI_API_KEY",
			"openrouter": "OPENROUTER_API_KEY",
		},
		Providers: []ProviderDefinition{
			{Name: "openai"},
			{Name: "gemini"},
			{Name: "openrouter"},
		},
		Models: []ModelDefinition{
			{
				Name:            "gemini-2.5-pro-preview-03-25",
				Provider:        "gemini",
				APIModelID:      "gemini-2.5-pro-preview-03-25",
				ContextWindow:   1000000,
				MaxOutputTokens: 65000,
				Parameters: map[string]ParameterDefinition{
					"temperature": {Type: "float", Default: 0.7},
					"top_p":       {Type: "float", Default: 0.95},
					"top_k":       {Type: "int", Default: 40},
				},
			},
			{
				Name:            "gpt-4",
				Provider:        "openai",
				APIModelID:      "gpt-4",
				ContextWindow:   128000,
				MaxOutputTokens: 4096,
				Parameters: map[string]ParameterDefinition{
					"temperature":       {Type: "float", Default: 0.7},
					"top_p":             {Type: "float", Default: 1.0},
					"frequency_penalty": {Type: "float", Default: 0.0},
					"presence_penalty":  {Type: "float", Default: 0.0},
				},
			},
			{
				Name:            "gpt-4.1",
				Provider:        "openai",
				APIModelID:      "gpt-4.1",
				ContextWindow:   1000000,
				MaxOutputTokens: 200000,
				Parameters: map[string]ParameterDefinition{
					"temperature":       {Type: "float", Default: 0.7},
					"top_p":             {Type: "float", Default: 1.0},
					"frequency_penalty": {Type: "float", Default: 0.0},
					"presence_penalty":  {Type: "float", Default: 0.0},
				},
			},
		},
	}
}

// loadConfigurationFromEnvironment creates a configuration based on environment variables.
// This is used as a fallback when no configuration file is available.
func loadConfigurationFromEnvironment() (*ModelsConfig, bool) {
	provider := os.Getenv(EnvConfigProvider)
	model := os.Getenv(EnvConfigModel)
	apiModelID := os.Getenv(EnvConfigAPIModelID)

	// If key environment variables are not set, return false
	if provider == "" || model == "" || apiModelID == "" {
		return nil, false
	}

	// Parse numeric values with defaults
	contextWindow := int32(1000000) // Default 1M tokens
	if envContext := os.Getenv(EnvConfigContextWindow); envContext != "" {
		if parsed, err := strconv.ParseInt(envContext, 10, 32); err == nil {
			contextWindow = int32(parsed)
		}
	}

	maxOutput := int32(65000) // Default 65k tokens
	if envOutput := os.Getenv(EnvConfigMaxOutput); envOutput != "" {
		if parsed, err := strconv.ParseInt(envOutput, 10, 32); err == nil {
			maxOutput = int32(parsed)
		}
	}

	// Determine API key environment variable based on provider
	var apiKeyEnvVar string
	switch strings.ToLower(provider) {
	case "openai":
		apiKeyEnvVar = "OPENAI_API_KEY"
	case "gemini":
		apiKeyEnvVar = "GEMINI_API_KEY"
	case "openrouter":
		apiKeyEnvVar = "OPENROUTER_API_KEY"
	default:
		apiKeyEnvVar = "GEMINI_API_KEY" // Default fallback
	}

	// Create provider definition
	providerDef := ProviderDefinition{Name: provider}
	if baseURL := os.Getenv(EnvConfigBaseURL); baseURL != "" {
		providerDef.BaseURL = baseURL
	}

	// Create basic parameters based on provider
	var parameters map[string]ParameterDefinition
	switch strings.ToLower(provider) {
	case "openai":
		parameters = map[string]ParameterDefinition{
			"temperature":       {Type: "float", Default: 0.7},
			"top_p":             {Type: "float", Default: 1.0},
			"frequency_penalty": {Type: "float", Default: 0.0},
			"presence_penalty":  {Type: "float", Default: 0.0},
		}
	case "gemini":
		parameters = map[string]ParameterDefinition{
			"temperature": {Type: "float", Default: 0.7},
			"top_p":       {Type: "float", Default: 0.95},
			"top_k":       {Type: "int", Default: 40},
		}
	case "openrouter":
		parameters = map[string]ParameterDefinition{
			"temperature": {Type: "float", Default: 0.7},
			"top_p":       {Type: "float", Default: 0.95},
		}
	default:
		parameters = map[string]ParameterDefinition{
			"temperature": {Type: "float", Default: 0.7},
		}
	}

	return &ModelsConfig{
		APIKeySources: map[string]string{
			provider: apiKeyEnvVar,
		},
		Providers: []ProviderDefinition{providerDef},
		Models: []ModelDefinition{
			{
				Name:            model,
				Provider:        provider,
				APIModelID:      apiModelID,
				ContextWindow:   contextWindow,
				MaxOutputTokens: maxOutput,
				Parameters:      parameters,
			},
		},
	}, true
}

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

// Load reads and parses the models.yaml configuration file with enhanced fallback mechanisms.
// This method tries multiple approaches in order:
// 1. Load from configuration file (traditional approach)
// 2. Load from environment variables (containerized environments)
// 3. Use embedded default configuration (minimal fallback)
func (c *ConfigLoader) Load() (*ModelsConfig, error) {
	// Create a background context for logging
	ctx := context.Background()

	// Strategy 1: Try to load from configuration file
	config, err := c.loadFromFile(ctx)
	if err == nil {
		c.Logger.InfoContext(ctx, "✅ Configuration loaded successfully from file")
		return config, nil
	}

	// Log the file loading error but continue with fallbacks
	c.Logger.WarnContext(ctx, "Failed to load configuration from file: %v", err)

	// Strategy 2: Try to load from environment variables
	config, loaded := loadConfigurationFromEnvironment()
	if loaded {
		c.Logger.InfoContext(ctx, "✅ Configuration loaded from environment variables")
		c.Logger.InfoContext(ctx, "Using environment-based configuration: provider=%s, model=%s",
			config.Models[0].Provider, config.Models[0].Name)

		// Validate the environment-based configuration
		if err := c.validate(config); err != nil {
			c.Logger.WarnContext(ctx, "Environment-based configuration validation failed: %v", err)
		} else {
			c.Logger.InfoContext(ctx, "Environment-based configuration validated successfully")
			return config, nil
		}
	}

	c.Logger.InfoContext(ctx, "Environment variables not set for configuration override")

	// Strategy 3: Use embedded default configuration
	c.Logger.InfoContext(ctx, "Using embedded default configuration as final fallback")
	config = getDefaultConfiguration()

	// Validate the default configuration
	if err := c.validate(config); err != nil {
		return nil, fmt.Errorf("default configuration validation failed: %w", err)
	}

	c.Logger.InfoContext(ctx, "✅ Default configuration loaded and validated successfully")
	c.Logger.InfoContext(ctx, "Configuration source: embedded defaults (%d providers, %d models)",
		len(config.Providers), len(config.Models))

	return config, nil
}

// loadFromFile attempts to load configuration from the models.yaml file
func (c *ConfigLoader) loadFromFile(ctx context.Context) (*ModelsConfig, error) {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine configuration path: %w", err)
	}

	// Log the configuration file path being used
	c.Logger.InfoContext(ctx, "Attempting to load model configuration from: %s", configPath)

	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found at %s", configPath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading configuration file at %s", configPath)
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

// validate performs comprehensive validation of the configuration with enhanced diagnostics
func (c *ConfigLoader) validate(config *ModelsConfig) error {
	// Create a background context for logging
	ctx := context.Background()

	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	var validationErrors []string

	// Check API key sources
	if len(config.APIKeySources) == 0 {
		validationErrors = append(validationErrors, "configuration must include api_key_sources")
	} else {
		c.Logger.InfoContext(ctx, "Validated API key sources: %d sources defined", len(config.APIKeySources))

		// Display found API key sources for debugging
		var foundKeys, missingKeys []string
		for provider, envVar := range config.APIKeySources {
			// Check if the environment variable exists (without revealing its value)
			_, exists := os.LookupEnv(envVar)
			if exists {
				foundKeys = append(foundKeys, fmt.Sprintf("%s (%s)", provider, envVar))
				c.Logger.InfoContext(ctx, "✓ API key for provider '%s' found in environment variable %s", provider, envVar)
			} else {
				missingKeys = append(missingKeys, fmt.Sprintf("%s (%s)", provider, envVar))
				c.Logger.WarnContext(ctx, "⚠ API key for provider '%s' not found in environment variable %s", provider, envVar)
			}
		}

		if len(foundKeys) > 0 {
			c.Logger.InfoContext(ctx, "Available API keys: %s", strings.Join(foundKeys, ", "))
		}
		if len(missingKeys) > 0 {
			c.Logger.WarnContext(ctx, "Missing API keys: %s", strings.Join(missingKeys, ", "))
			c.Logger.InfoContext(ctx, "💡 Tip: Set missing API key environment variables to enable those providers")
		}
	}

	// Check providers
	if len(config.Providers) == 0 {
		validationErrors = append(validationErrors, "configuration must include at least one provider")
	} else {
		c.Logger.InfoContext(ctx, "Validating %d providers...", len(config.Providers))

		// Check for provider name uniqueness
		providerNames := make(map[string]bool)
		for i, provider := range config.Providers {
			if provider.Name == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("provider at index %d is missing name", i))
				continue
			}
			if providerNames[provider.Name] {
				validationErrors = append(validationErrors, fmt.Sprintf("duplicate provider name '%s' detected", provider.Name))
				continue
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
			validationErrors = append(validationErrors, "configuration must include at least one model")
		} else {
			c.Logger.InfoContext(ctx, "Validating %d models...", len(config.Models))

			// Check for model name uniqueness
			modelNames := make(map[string]bool)

			for i, model := range config.Models {
				// Validate model name
				if model.Name == "" {
					validationErrors = append(validationErrors, fmt.Sprintf("model at index %d is missing name", i))
					continue
				}
				if modelNames[model.Name] {
					validationErrors = append(validationErrors, fmt.Sprintf("duplicate model name '%s' detected", model.Name))
					continue
				}
				modelNames[model.Name] = true

				// Validate provider
				if model.Provider == "" {
					validationErrors = append(validationErrors, fmt.Sprintf("model '%s' is missing provider", model.Name))
					continue
				}
				if !providerNames[model.Provider] {
					validationErrors = append(validationErrors, fmt.Sprintf("model '%s' references unknown provider '%s'", model.Name, model.Provider))
					continue
				}

				// Validate API model ID
				if model.APIModelID == "" {
					validationErrors = append(validationErrors, fmt.Sprintf("model '%s' is missing api_model_id", model.Name))
					continue
				}

				// Token validation removed as part of T036C
				// The token validation logic has been removed since the token-related fields
				// have been removed from the ModelDefinition struct.
				// Token handling is now the responsibility of each provider.

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
		}
	}

	// Return validation errors if any were found
	if len(validationErrors) > 0 {
		c.Logger.WarnContext(ctx, "Configuration validation failed with %d errors:", len(validationErrors))
		for i, err := range validationErrors {
			c.Logger.WarnContext(ctx, "  %d. %s", i+1, err)
		}

		// Provide helpful troubleshooting information
		c.Logger.InfoContext(ctx, "💡 Configuration troubleshooting tips:")
		c.Logger.InfoContext(ctx, "   • For missing config file: Set environment variables (THINKTANK_CONFIG_*) or install models.yaml")
		c.Logger.InfoContext(ctx, "   • For missing API keys: Set OPENAI_API_KEY, GEMINI_API_KEY, or OPENROUTER_API_KEY")
		c.Logger.InfoContext(ctx, "   • For containerized environments: Use environment variable configuration")

		return fmt.Errorf("configuration validation failed with %d errors: %s",
			len(validationErrors), strings.Join(validationErrors, "; "))
	}

	return nil
}
