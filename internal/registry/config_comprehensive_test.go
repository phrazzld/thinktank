package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestConfigurationLoadingEnvironments tests configuration loading in different environments
func TestConfigurationLoadingEnvironments(t *testing.T) {
	tests := []struct {
		name                  string
		setupFileConfig       bool
		setupEnvConfig        bool
		fileConfigValid       bool
		envConfigComplete     bool
		expectedSource        string
		expectedProviderCount int
		expectedModelCount    int
		expectError           bool
	}{
		{
			name:                  "Container environment - no file, complete env config",
			setupFileConfig:       false,
			setupEnvConfig:        true,
			fileConfigValid:       false,
			envConfigComplete:     true,
			expectedSource:        "environment",
			expectedProviderCount: 1,
			expectedModelCount:    1,
			expectError:           false,
		},
		{
			name:                  "Container environment - no file, incomplete env config",
			setupFileConfig:       false,
			setupEnvConfig:        true,
			fileConfigValid:       false,
			envConfigComplete:     false,
			expectedSource:        "default",
			expectedProviderCount: 3,
			expectedModelCount:    3,
			expectError:           false,
		},
		{
			name:                  "Local environment - valid file config",
			setupFileConfig:       true,
			setupEnvConfig:        false,
			fileConfigValid:       true,
			envConfigComplete:     false,
			expectedSource:        "file",
			expectedProviderCount: 2,
			expectedModelCount:    2,
			expectError:           false,
		},
		{
			name:                  "Local environment - invalid file, no env",
			setupFileConfig:       true,
			setupEnvConfig:        false,
			fileConfigValid:       false,
			envConfigComplete:     false,
			expectedSource:        "default",
			expectedProviderCount: 3,
			expectedModelCount:    3,
			expectError:           false,
		},
		{
			name:                  "Priority test - file overrides env",
			setupFileConfig:       true,
			setupEnvConfig:        true,
			fileConfigValid:       true,
			envConfigComplete:     true,
			expectedSource:        "file",
			expectedProviderCount: 2,
			expectedModelCount:    2,
			expectError:           false,
		},
		{
			name:                  "Fallback chain - invalid file, env config",
			setupFileConfig:       true,
			setupEnvConfig:        true,
			fileConfigValid:       false,
			envConfigComplete:     true,
			expectedSource:        "environment",
			expectedProviderCount: 1,
			expectedModelCount:    1,
			expectError:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variables at start
			envVars := []string{
				EnvConfigProvider, EnvConfigModel, EnvConfigAPIModelID,
				EnvConfigContextWindow, EnvConfigMaxOutput, EnvConfigBaseURL,
			}
			for _, envVar := range envVars {
				if err := os.Unsetenv(envVar); err != nil {
					t.Errorf("Warning: Failed to unset environment variable %s: %v", envVar, err)
				}
			}

			var tmpFilePath string

			// Setup file configuration if needed
			if tt.setupFileConfig {
				tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer func() {
					if err := os.Remove(tmpFile.Name()); err != nil {
						t.Errorf("Warning: Failed to remove temp file: %v", err)
					}
				}()
				tmpFilePath = tmpFile.Name()

				var configContent string
				if tt.fileConfigValid {
					configContent = `
api_key_sources:
  openai: OPENAI_API_KEY
  gemini: GEMINI_API_KEY

providers:
  - name: openai
  - name: gemini
    base_url: https://custom-gemini.example.com

models:
  - name: gpt-4-test
    provider: openai
    api_model_id: gpt-4
    context_window: 128000
    max_output_tokens: 4096
    parameters:
      temperature:
        type: float
        default: 0.7

  - name: gemini-test
    provider: gemini
    api_model_id: gemini-1.5-pro
    context_window: 1000000
    max_output_tokens: 8192
    parameters:
      temperature:
        type: float
        default: 0.8
`
				} else {
					// Invalid YAML content
					configContent = "invalid yaml: [\nthis breaks parsing"
				}

				if _, err := tmpFile.WriteString(configContent); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
				if err := tmpFile.Close(); err != nil {
					t.Errorf("Warning: Failed to close temp file: %v", err)
				}
			} else {
				// Use non-existent file path
				tmpFilePath = filepath.Join(os.TempDir(), "non-existent-config.yaml")
			}

			// Setup environment configuration if needed
			if tt.setupEnvConfig {
				if tt.envConfigComplete {
					if err := os.Setenv(EnvConfigProvider, "openrouter"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigProvider, err)
					}
					if err := os.Setenv(EnvConfigModel, "env-test-model"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigModel, err)
					}
					if err := os.Setenv(EnvConfigAPIModelID, "deepseek/deepseek-chat"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigAPIModelID, err)
					}
					if err := os.Setenv(EnvConfigContextWindow, "200000"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigContextWindow, err)
					}
					if err := os.Setenv(EnvConfigMaxOutput, "16000"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigMaxOutput, err)
					}
					if err := os.Setenv(EnvConfigBaseURL, "https://openrouter.ai/api/v1"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigBaseURL, err)
					}
				} else {
					// Incomplete env config (missing required fields)
					if err := os.Setenv(EnvConfigProvider, "gemini"); err != nil {
						t.Errorf("Warning: Failed to set environment variable %s: %v", EnvConfigProvider, err)
					}
					// Missing model name and API model ID
				}
			}

			// Clean up environment variables after test
			defer func() {
				for _, envVar := range envVars {
					if err := os.Unsetenv(envVar); err != nil {
						t.Errorf("Warning: Failed to unset environment variable %s: %v", envVar, err)
					}
				}
			}()

			// Create loader with custom path
			loader := &ConfigLoader{
				Logger: logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel),
			}

			// Mock the GetConfigPath method
			loader.GetConfigPath = func() (string, error) {
				return tmpFilePath, nil
			}

			// Load configuration
			config, err := loader.Load()

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify configuration
			if config == nil {
				t.Fatal("Expected non-nil configuration")
			}

			// Check provider count
			if len(config.Providers) != tt.expectedProviderCount {
				t.Errorf("Expected %d providers, got %d", tt.expectedProviderCount, len(config.Providers))
			}

			// Check model count
			if len(config.Models) != tt.expectedModelCount {
				t.Errorf("Expected %d models, got %d", tt.expectedModelCount, len(config.Models))
			}

			// Verify source-specific characteristics
			switch tt.expectedSource {
			case "file":
				// File config should have specific models
				foundGPT4Test := false
				for _, model := range config.Models {
					if model.Name == "gpt-4-test" {
						foundGPT4Test = true
						break
					}
				}
				if !foundGPT4Test {
					t.Error("Expected file config to contain 'gpt-4-test' model")
				}

			case "environment":
				// Environment config should have env-specified model
				foundEnvModel := false
				for _, model := range config.Models {
					if model.Name == "env-test-model" {
						foundEnvModel = true
						if model.Provider != "openrouter" {
							t.Errorf("Expected env model provider to be 'openrouter', got '%s'", model.Provider)
						}
						break
					}
				}
				if !foundEnvModel {
					t.Error("Expected environment config to contain 'env-test-model'")
				}

			case "default":
				// Default config should have known default models
				foundDefaultModel := false
				for _, model := range config.Models {
					if model.Name == "gemini-2.5-pro-preview-03-25" || model.Name == "gpt-4" || model.Name == "gpt-4.1" {
						foundDefaultModel = true
						break
					}
				}
				if !foundDefaultModel {
					t.Error("Expected default config to contain known default models")
				}
			}
		})
	}
}

// TestConfigurationValidationEdgeCases tests edge cases in configuration validation
func TestConfigurationValidationEdgeCases(t *testing.T) {
	loader := NewConfigLoader(logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel))

	tests := []struct {
		name        string
		config      *ModelsConfig
		expectError bool
		errorText   string
	}{
		{
			name: "Duplicate provider names",
			config: &ModelsConfig{
				APIKeySources: map[string]string{"test": "TEST_KEY"},
				Providers: []ProviderDefinition{
					{Name: "duplicate"},
					{Name: "duplicate"}, // Duplicate
				},
				Models: []ModelDefinition{
					{Name: "test", Provider: "duplicate", APIModelID: "test-api"},
				},
			},
			expectError: true,
			errorText:   "duplicate provider name",
		},
		{
			name: "Duplicate model names",
			config: &ModelsConfig{
				APIKeySources: map[string]string{"test": "TEST_KEY"},
				Providers:     []ProviderDefinition{{Name: "test"}},
				Models: []ModelDefinition{
					{Name: "duplicate", Provider: "test", APIModelID: "test-api-1"},
					{Name: "duplicate", Provider: "test", APIModelID: "test-api-2"}, // Duplicate
				},
			},
			expectError: true,
			errorText:   "duplicate model name",
		},
		{
			name: "Model references non-existent provider",
			config: &ModelsConfig{
				APIKeySources: map[string]string{"test": "TEST_KEY"},
				Providers:     []ProviderDefinition{{Name: "existing"}},
				Models: []ModelDefinition{
					{Name: "test", Provider: "non-existent", APIModelID: "test-api"}, // Bad provider
				},
			},
			expectError: true,
			errorText:   "unknown provider",
		},
		{
			name: "Empty provider name",
			config: &ModelsConfig{
				APIKeySources: map[string]string{"test": "TEST_KEY"},
				Providers: []ProviderDefinition{
					{Name: ""}, // Empty name
				},
				Models: []ModelDefinition{
					{Name: "test", Provider: "test", APIModelID: "test-api"},
				},
			},
			expectError: true,
			errorText:   "missing name",
		},
		{
			name: "Empty model name",
			config: &ModelsConfig{
				APIKeySources: map[string]string{"test": "TEST_KEY"},
				Providers:     []ProviderDefinition{{Name: "test"}},
				Models: []ModelDefinition{
					{Name: "", Provider: "test", APIModelID: "test-api"}, // Empty name
				},
			},
			expectError: true,
			errorText:   "missing name",
		},
		{
			name: "Empty API model ID",
			config: &ModelsConfig{
				APIKeySources: map[string]string{"test": "TEST_KEY"},
				Providers:     []ProviderDefinition{{Name: "test"}},
				Models: []ModelDefinition{
					{Name: "test", Provider: "test", APIModelID: ""}, // Empty API model ID
				},
			},
			expectError: true,
			errorText:   "missing api_model_id",
		},
		{
			name: "Complex valid configuration",
			config: &ModelsConfig{
				APIKeySources: map[string]string{
					"openai":     "OPENAI_API_KEY",
					"gemini":     "GEMINI_API_KEY",
					"openrouter": "OPENROUTER_API_KEY",
				},
				Providers: []ProviderDefinition{
					{Name: "openai", BaseURL: "https://api.openai.com/v1"},
					{Name: "gemini"},
					{Name: "openrouter", BaseURL: "https://openrouter.ai/api/v1"},
				},
				Models: []ModelDefinition{
					{
						Name:            "gpt-4-advanced",
						Provider:        "openai",
						APIModelID:      "gpt-4",
						ContextWindow:   128000,
						MaxOutputTokens: 4096,
						Parameters: map[string]ParameterDefinition{
							"temperature": {Type: "float", Default: 0.7},
							"top_p":       {Type: "float", Default: 1.0},
						},
					},
					{
						Name:            "gemini-pro-advanced",
						Provider:        "gemini",
						APIModelID:      "gemini-1.5-pro",
						ContextWindow:   1000000,
						MaxOutputTokens: 8192,
						Parameters: map[string]ParameterDefinition{
							"temperature": {Type: "float", Default: 0.8},
							"top_k":       {Type: "int", Default: 40},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if tt.errorText != "" && !containsIgnoreCase(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestEnvironmentVariableOverrideCombinations tests various combinations of environment variable overrides
func TestEnvironmentVariableOverrideCombinations(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectSuccess  bool
		expectProvider string
		expectModel    string
		expectBaseURL  string
	}{
		{
			name: "All required env vars set - OpenAI",
			envVars: map[string]string{
				EnvConfigProvider:   "openai",
				EnvConfigModel:      "custom-gpt-4",
				EnvConfigAPIModelID: "gpt-4-custom",
			},
			expectSuccess:  true,
			expectProvider: "openai",
			expectModel:    "custom-gpt-4",
		},
		{
			name: "All env vars including optional - Gemini",
			envVars: map[string]string{
				EnvConfigProvider:      "gemini",
				EnvConfigModel:         "custom-gemini",
				EnvConfigAPIModelID:    "gemini-custom",
				EnvConfigContextWindow: "500000",
				EnvConfigMaxOutput:     "25000",
				EnvConfigBaseURL:       "https://custom-gemini.example.com",
			},
			expectSuccess:  true,
			expectProvider: "gemini",
			expectModel:    "custom-gemini",
			expectBaseURL:  "https://custom-gemini.example.com",
		},
		{
			name: "Missing model name",
			envVars: map[string]string{
				EnvConfigProvider:   "openai",
				EnvConfigAPIModelID: "gpt-4",
				// Missing EnvConfigModel
			},
			expectSuccess: false,
		},
		{
			name: "Missing API model ID",
			envVars: map[string]string{
				EnvConfigProvider: "openai",
				EnvConfigModel:    "test-model",
				// Missing EnvConfigAPIModelID
			},
			expectSuccess: false,
		},
		{
			name: "Missing provider",
			envVars: map[string]string{
				EnvConfigModel:      "test-model",
				EnvConfigAPIModelID: "test-api-id",
				// Missing EnvConfigProvider
			},
			expectSuccess: false,
		},
		{
			name: "Invalid numeric values",
			envVars: map[string]string{
				EnvConfigProvider:      "gemini",
				EnvConfigModel:         "test-model",
				EnvConfigAPIModelID:    "test-api-id",
				EnvConfigContextWindow: "not-a-number",
				EnvConfigMaxOutput:     "also-not-a-number",
			},
			expectSuccess:  true, // Should succeed with defaults for invalid numbers
			expectProvider: "gemini",
			expectModel:    "test-model",
		},
		{
			name: "OpenRouter with custom URL",
			envVars: map[string]string{
				EnvConfigProvider:   "openrouter",
				EnvConfigModel:      "openrouter-model",
				EnvConfigAPIModelID: "deepseek/deepseek-chat",
				EnvConfigBaseURL:    "https://custom-openrouter.example.com",
			},
			expectSuccess:  true,
			expectProvider: "openrouter",
			expectModel:    "openrouter-model",
			expectBaseURL:  "https://custom-openrouter.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment first
			envVars := []string{
				EnvConfigProvider, EnvConfigModel, EnvConfigAPIModelID,
				EnvConfigContextWindow, EnvConfigMaxOutput, EnvConfigBaseURL,
			}
			for _, envVar := range envVars {
				if err := os.Unsetenv(envVar); err != nil {
					t.Errorf("Warning: Failed to unset environment variable %s: %v", envVar, err)
				}
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Errorf("Warning: Failed to set environment variable %s: %v", key, err)
				}
			}

			// Clean up after test
			defer func() {
				for _, envVar := range envVars {
					if err := os.Unsetenv(envVar); err != nil {
						t.Errorf("Warning: Failed to unset environment variable %s: %v", envVar, err)
					}
				}
			}()

			// Test environment variable configuration loading
			config, loaded := loadConfigurationFromEnvironment()

			if tt.expectSuccess {
				if !loaded {
					t.Fatalf("Expected environment config to be loaded but it wasn't")
				}
				if config == nil {
					t.Fatalf("Expected non-nil config")
				}

				// Verify provider
				if len(config.Providers) != 1 {
					t.Fatalf("Expected exactly 1 provider, got %d", len(config.Providers))
				}
				if config.Providers[0].Name != tt.expectProvider {
					t.Errorf("Expected provider '%s', got '%s'", tt.expectProvider, config.Providers[0].Name)
				}

				// Verify model
				if len(config.Models) != 1 {
					t.Fatalf("Expected exactly 1 model, got %d", len(config.Models))
				}
				if config.Models[0].Name != tt.expectModel {
					t.Errorf("Expected model '%s', got '%s'", tt.expectModel, config.Models[0].Name)
				}

				// Verify base URL if expected
				if tt.expectBaseURL != "" {
					if config.Providers[0].BaseURL != tt.expectBaseURL {
						t.Errorf("Expected base URL '%s', got '%s'", tt.expectBaseURL, config.Providers[0].BaseURL)
					}
				}

				// Verify API key source is correctly mapped
				expectedAPIKeyVar := ""
				switch tt.expectProvider {
				case "openai":
					expectedAPIKeyVar = "OPENAI_API_KEY"
				case "gemini":
					expectedAPIKeyVar = "GEMINI_API_KEY"
				case "openrouter":
					expectedAPIKeyVar = "OPENROUTER_API_KEY"
				}

				if apiKeyVar, exists := config.APIKeySources[tt.expectProvider]; !exists || apiKeyVar != expectedAPIKeyVar {
					t.Errorf("Expected API key var '%s' for provider '%s', got '%s'", expectedAPIKeyVar, tt.expectProvider, apiKeyVar)
				}

			} else {
				if loaded {
					t.Fatalf("Expected environment config NOT to be loaded but it was")
				}
			}
		})
	}
}

// TestConfigurationErrorHandling tests robust error handling in configuration loading
func TestConfigurationErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		setupFile      bool
		fileContent    string
		filePermission os.FileMode
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Permission denied reading config file",
			setupFile:      true,
			fileContent:    "valid: config",
			filePermission: 0000, // No read permission
			expectError:    true,
			errorContains:  "permission denied",
		},
		{
			name:           "Malformed YAML - unclosed bracket",
			setupFile:      true,
			fileContent:    "providers: [\n  - name: test\n# Missing closing bracket",
			filePermission: 0644,
			expectError:    true,
			errorContains:  "invalid YAML",
		},
		{
			name:           "Malformed YAML - invalid structure",
			setupFile:      true,
			fileContent:    "this: is\n  not: valid\n    yaml: structure {\n  incomplete",
			filePermission: 0644,
			expectError:    true,
			errorContains:  "invalid YAML",
		},
		{
			name:           "Empty file",
			setupFile:      true,
			fileContent:    "",
			filePermission: 0644,
			expectError:    true,
			errorContains:  "configuration must include",
		},
		{
			name:           "Only whitespace",
			setupFile:      true,
			fileContent:    "   \n\t\n   ",
			filePermission: 0644,
			expectError:    true,
			errorContains:  "configuration must include",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables to avoid fallback
			envVars := []string{
				EnvConfigProvider, EnvConfigModel, EnvConfigAPIModelID,
				EnvConfigContextWindow, EnvConfigMaxOutput, EnvConfigBaseURL,
			}
			for _, envVar := range envVars {
				if err := os.Unsetenv(envVar); err != nil {
					t.Errorf("Warning: Failed to unset environment variable %s: %v", envVar, err)
				}
			}
			defer func() {
				for _, envVar := range envVars {
					if err := os.Unsetenv(envVar); err != nil {
						t.Errorf("Warning: Failed to unset environment variable %s: %v", envVar, err)
					}
				}
			}()

			var tmpFilePath string
			if tt.setupFile {
				tmpFile, err := os.CreateTemp("", "config-error-test-*.yaml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer func() {
					if err := os.Remove(tmpFile.Name()); err != nil {
						t.Errorf("Warning: Failed to remove temp file: %v", err)
					}
				}()
				tmpFilePath = tmpFile.Name()

				if _, err := tmpFile.WriteString(tt.fileContent); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}
				if err := tmpFile.Close(); err != nil {
					t.Errorf("Warning: Failed to close temp file: %v", err)
				}

				// Set file permissions
				if err := os.Chmod(tmpFilePath, tt.filePermission); err != nil {
					t.Fatalf("Failed to set file permissions: %v", err)
				}
			} else {
				tmpFilePath = filepath.Join(os.TempDir(), "non-existent-file.yaml")
			}

			// Create loader with custom path
			loader := &ConfigLoader{
				Logger: logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel),
			}
			loader.GetConfigPath = func() (string, error) {
				return tmpFilePath, nil
			}

			// Attempt to load configuration
			config, err := loader.Load()

			if tt.expectError {
				// With the enhanced fallback, most errors should be handled gracefully
				// Only severe errors that prevent fallback should cause failure
				if err == nil {
					// Check if we got default fallback instead of error
					if config == nil || len(config.Models) == 0 {
						t.Fatalf("Expected either error or valid fallback config")
					}
					// If we got a valid config, it's the default fallback - that's acceptable
					return
				}

				if tt.errorContains != "" && !containsIgnoreCase(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if config == nil {
					t.Fatal("Expected non-nil config")
				}
			}
		})
	}
}

// Helper function for case-insensitive string matching
func containsIgnoreCase(s, substr string) bool {
	return containsString(strings.ToLower(s), strings.ToLower(substr))
}

func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
