// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
)

// TestConfigurationIntegrationScenarios tests end-to-end configuration loading scenarios
// This simulates real-world usage patterns in different deployment environments
func TestConfigurationIntegrationScenarios(t *testing.T) {
	tests := []struct {
		name                  string
		simulateEnvironment   string
		setupConfigFile       bool
		configFileContent     string
		setupEnvVars          map[string]string
		expectedSuccess       bool
		expectedModelCount    int
		expectedProviderCount int
		description           string
	}{
		{
			name:                "Local development environment",
			simulateEnvironment: "local",
			setupConfigFile:     true,
			configFileContent: `
api_key_sources:
  openai: OPENAI_API_KEY
  gemini: GEMINI_API_KEY

providers:
  - name: openai
  - name: gemini

models:
  - name: gpt-4-dev
    provider: openai
    api_model_id: gpt-4
    context_window: 128000
    max_output_tokens: 4096
    parameters:
      temperature:
        type: float
        default: 0.5

  - name: gemini-dev
    provider: gemini
    api_model_id: gemini-1.5-pro
    context_window: 1000000
    max_output_tokens: 8192
    parameters:
      temperature:
        type: float
        default: 0.7
`,
			setupEnvVars:          nil,
			expectedSuccess:       true,
			expectedModelCount:    2,
			expectedProviderCount: 2,
			description:           "Local development with explicit config file should work",
		},
		{
			name:                "Docker container environment",
			simulateEnvironment: "container",
			setupConfigFile:     false, // No config file in container
			configFileContent:   "",
			setupEnvVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER":       "gemini",
				"THINKTANK_CONFIG_MODEL":          "gemini-container",
				"THINKTANK_CONFIG_API_MODEL_ID":   "gemini-1.5-pro",
				"THINKTANK_CONFIG_CONTEXT_WINDOW": "500000",
				"THINKTANK_CONFIG_MAX_OUTPUT":     "32000",
			},
			expectedSuccess:       true,
			expectedModelCount:    1,
			expectedProviderCount: 1,
			description:           "Container environment with env vars should work",
		},
		{
			name:                "Kubernetes pod environment",
			simulateEnvironment: "kubernetes",
			setupConfigFile:     false,
			configFileContent:   "",
			setupEnvVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER":       "openrouter",
				"THINKTANK_CONFIG_MODEL":          "k8s-deepseek",
				"THINKTANK_CONFIG_API_MODEL_ID":   "deepseek/deepseek-chat",
				"THINKTANK_CONFIG_BASE_URL":       "https://openrouter.ai/api/v1",
				"THINKTANK_CONFIG_CONTEXT_WINDOW": "131072",
				"THINKTANK_CONFIG_MAX_OUTPUT":     "65536",
			},
			expectedSuccess:       true,
			expectedModelCount:    1,
			expectedProviderCount: 1,
			description:           "Kubernetes pod with OpenRouter config should work",
		},
		{
			name:                  "Minimal environment - default fallback",
			simulateEnvironment:   "minimal",
			setupConfigFile:       false,
			configFileContent:     "",
			setupEnvVars:          nil, // No env vars
			expectedSuccess:       true,
			expectedModelCount:    3, // Default config has 3 models
			expectedProviderCount: 2, // Default config has 2 unique providers used by models (openai, gemini)
			description:           "Minimal environment should fall back to defaults",
		},
		{
			name:                "CI/CD environment",
			simulateEnvironment: "ci",
			setupConfigFile:     false,
			configFileContent:   "",
			setupEnvVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER":       "openai",
				"THINKTANK_CONFIG_MODEL":          "ci-gpt-4",
				"THINKTANK_CONFIG_API_MODEL_ID":   "gpt-4",
				"THINKTANK_CONFIG_CONTEXT_WINDOW": "100000",
				"THINKTANK_CONFIG_MAX_OUTPUT":     "8192",
			},
			expectedSuccess:       true,
			expectedModelCount:    1,
			expectedProviderCount: 1,
			description:           "CI/CD environment with OpenAI config should work",
		},
		{
			name:                "Corrupted config with env fallback",
			simulateEnvironment: "mixed",
			setupConfigFile:     true,
			configFileContent:   "invalid: yaml: content [\n unclosed",
			setupEnvVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER":     "gemini",
				"THINKTANK_CONFIG_MODEL":        "fallback-gemini",
				"THINKTANK_CONFIG_API_MODEL_ID": "gemini-1.5-pro",
			},
			expectedSuccess:       true,
			expectedModelCount:    1,
			expectedProviderCount: 1,
			description:           "Corrupted config should fall back to env vars",
		},
		{
			name:                "Partial env config with default fallback",
			simulateEnvironment: "partial",
			setupConfigFile:     false,
			configFileContent:   "",
			setupEnvVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER": "gemini",
				// Missing model name and API model ID
			},
			expectedSuccess:       true,
			expectedModelCount:    3, // Should fall back to defaults
			expectedProviderCount: 2, // Default config has 2 unique providers used by models (openai, gemini)
			description:           "Partial env config should fall back to defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger
			testLogger := logutil.NewTestLogger(t)

			// Create a context with correlation ID for tracking
			correlationID := "test-config-" + tt.simulateEnvironment
			ctx := logutil.WithCorrelationID(context.Background(), correlationID)

			// Clean up environment variables before test
			envVarsToClean := []string{
				"THINKTANK_CONFIG_PROVIDER", "THINKTANK_CONFIG_MODEL", "THINKTANK_CONFIG_API_MODEL_ID",
				"THINKTANK_CONFIG_CONTEXT_WINDOW", "THINKTANK_CONFIG_MAX_OUTPUT", "THINKTANK_CONFIG_BASE_URL",
			}
			for _, envVar := range envVarsToClean {
				os.Unsetenv(envVar)
			}

			// Setup environment variables if specified
			if tt.setupEnvVars != nil {
				for key, value := range tt.setupEnvVars {
					os.Setenv(key, value)
				}
			}

			// Clean up environment variables after test
			defer func() {
				for _, envVar := range envVarsToClean {
					os.Unsetenv(envVar)
				}
			}()

			var configPath string
			if tt.setupConfigFile {
				// Create temporary config file
				tmpFile, err := os.CreateTemp("", "integration-config-*.yaml")
				if err != nil {
					t.Fatalf("Failed to create temp config file: %v", err)
				}
				defer os.Remove(tmpFile.Name())
				configPath = tmpFile.Name()

				if _, err := tmpFile.WriteString(tt.configFileContent); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
				tmpFile.Close()
			} else {
				// Use non-existent path to simulate missing config file
				configPath = filepath.Join(os.TempDir(), "non-existent-config.yaml")
			}

			// Create config loader with custom path
			configLoader := &registry.ConfigLoader{
				Logger: testLogger,
			}
			configLoader.GetConfigPath = func() (string, error) {
				return configPath, nil
			}

			// Create registry and attempt to load configuration
			reg := registry.NewRegistry(testLogger)

			// Load configuration using the registry
			err := reg.LoadConfig(ctx, configLoader)

			// Check if the result matches expectations
			if tt.expectedSuccess {
				if err != nil {
					t.Fatalf("Expected successful config loading but got error: %v\nDescription: %s", err, tt.description)
				}

				// Verify the registry is properly initialized
				if reg == nil {
					t.Fatalf("Registry should not be nil after successful config loading\nDescription: %s", tt.description)
				}

				// Test registry functionality by trying to get available models
				availableModels, err := reg.GetAvailableModels(ctx)
				if err != nil {
					t.Fatalf("Failed to get available models: %v\nDescription: %s", err, tt.description)
				}

				if len(availableModels) < tt.expectedModelCount {
					t.Errorf("Expected at least %d models, got %d\nDescription: %s\nModels: %v",
						tt.expectedModelCount, len(availableModels), tt.description, availableModels)
				}

				// Test getting providers for each model
				testedProviders := make(map[string]bool)
				for _, modelName := range availableModels {
					model, err := reg.GetModel(ctx, modelName)
					if err != nil {
						t.Errorf("Failed to get model '%s': %v\nDescription: %s", modelName, err, tt.description)
						continue
					}

					// Get the provider for this model
					provider, err := reg.GetProvider(ctx, model.Provider)
					if err != nil {
						t.Errorf("Failed to get provider '%s' for model '%s': %v\nDescription: %s",
							model.Provider, modelName, err, tt.description)
						continue
					}

					if provider.Name != model.Provider {
						t.Errorf("Provider name mismatch: expected '%s', got '%s'\nDescription: %s",
							model.Provider, provider.Name, tt.description)
					}

					testedProviders[provider.Name] = true
				}

				if len(testedProviders) < tt.expectedProviderCount {
					t.Errorf("Expected at least %d providers, got %d\nDescription: %s\nProviders: %v",
						tt.expectedProviderCount, len(testedProviders), tt.description, getKeys(testedProviders))
				}

			} else {
				if err == nil {
					t.Fatalf("Expected config loading to fail but it succeeded\nDescription: %s", tt.description)
				}
			}

			// Verify that correlation ID is present in logs
			logs := testLogger.GetTestLogs()
			foundCorrelationID := false
			for _, logMsg := range logs {
				if containsString(logMsg, correlationID) {
					foundCorrelationID = true
					break
				}
			}
			if !foundCorrelationID && len(logs) > 0 {
				t.Errorf("Correlation ID '%s' not found in logs\nDescription: %s\nLogs: %v",
					correlationID, tt.description, logs)
			}
		})
	}
}

// TestRegistryManagerConfigurationFallback tests the registry manager's configuration fallback behavior
func TestRegistryManagerConfigurationFallback(t *testing.T) {
	testLogger := logutil.NewTestLogger(t)

	// Test with missing config file - should use default fallback through manager
	manager := registry.NewManager(testLogger)

	// Initialize the manager - this should succeed with fallback
	err := manager.Initialize()
	if err != nil {
		t.Fatalf("Manager initialization should succeed with fallback: %v", err)
	}

	// Get the registry and test basic functionality
	reg := manager.GetRegistry()
	if reg == nil {
		t.Fatal("Registry should not be nil after manager initialization")
	}

	ctx := context.Background()

	// Test getting available models
	models, err := reg.GetAvailableModels(ctx)
	if err != nil {
		t.Fatalf("Failed to get available models: %v", err)
	}

	if len(models) == 0 {
		t.Error("Should have at least some default models available")
	}

	// Test getting a known default model
	knownModels := []string{"gemini-2.5-pro-preview-03-25", "gpt-4", "gpt-4.1"}
	foundModel := false
	for _, knownModel := range knownModels {
		_, err := reg.GetModel(ctx, knownModel)
		if err == nil {
			foundModel = true
			break
		}
	}

	if !foundModel {
		t.Errorf("Should be able to get at least one of the known default models: %v", knownModels)
	}
}

// TestConfigurationErrorRecovery tests how the system recovers from various configuration errors
func TestConfigurationErrorRecovery(t *testing.T) {
	testLogger := logutil.NewTestLogger(t)

	scenarios := []struct {
		name           string
		configContent  string
		envVars        map[string]string
		expectRecovery bool
		description    string
	}{
		{
			name:          "Malformed YAML with complete env fallback",
			configContent: "invalid: yaml [\n broken",
			envVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER":     "gemini",
				"THINKTANK_CONFIG_MODEL":        "recovery-model",
				"THINKTANK_CONFIG_API_MODEL_ID": "gemini-1.5-pro",
			},
			expectRecovery: true,
			description:    "Malformed YAML should recover with env vars",
		},
		{
			name:          "Empty file with env fallback",
			configContent: "",
			envVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER":     "openai",
				"THINKTANK_CONFIG_MODEL":        "recovery-gpt",
				"THINKTANK_CONFIG_API_MODEL_ID": "gpt-4",
			},
			expectRecovery: true,
			description:    "Empty config file should recover with env vars",
		},
		{
			name:          "Invalid config with partial env vars",
			configContent: "completely: invalid\n structure: here",
			envVars: map[string]string{
				"THINKTANK_CONFIG_PROVIDER": "gemini",
				// Missing other required env vars
			},
			expectRecovery: true, // Should fall back to defaults
			description:    "Invalid config with partial env should fall back to defaults",
		},
		{
			name: "Valid config with syntax errors",
			configContent: `
api_key_sources:
  openai: OPENAI_API_KEY
providers:
  - name: openai
models:
  - name: test
    provider: openai
    # Missing required api_model_id field
`,
			envVars:        nil,
			expectRecovery: true, // Should fall back to defaults
			description:    "Config with validation errors should fall back to defaults",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Clean environment
			envVarsToClean := []string{
				"THINKTANK_CONFIG_PROVIDER", "THINKTANK_CONFIG_MODEL", "THINKTANK_CONFIG_API_MODEL_ID",
				"THINKTANK_CONFIG_CONTEXT_WINDOW", "THINKTANK_CONFIG_MAX_OUTPUT", "THINKTANK_CONFIG_BASE_URL",
			}
			for _, envVar := range envVarsToClean {
				os.Unsetenv(envVar)
			}

			// Set up environment variables
			if scenario.envVars != nil {
				for key, value := range scenario.envVars {
					os.Setenv(key, value)
				}
			}

			defer func() {
				for _, envVar := range envVarsToClean {
					os.Unsetenv(envVar)
				}
			}()

			// Create temporary config file
			tmpFile, err := os.CreateTemp("", "error-recovery-*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(scenario.configContent); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}
			tmpFile.Close()

			// Create config loader with custom path
			configLoader := &registry.ConfigLoader{
				Logger: testLogger,
			}
			configLoader.GetConfigPath = func() (string, error) {
				return tmpFile.Name(), nil
			}

			// Create registry and attempt to load configuration
			reg := registry.NewRegistry(testLogger)
			ctx := context.Background()

			err = reg.LoadConfig(ctx, configLoader)

			if scenario.expectRecovery {
				if err != nil {
					t.Fatalf("Expected successful recovery but got error: %v\nDescription: %s", err, scenario.description)
				}

				// Test that the registry is functional
				models, err := reg.GetAvailableModels(ctx)
				if err != nil {
					t.Fatalf("Registry should be functional after recovery: %v\nDescription: %s", err, scenario.description)
				}

				if len(models) == 0 {
					t.Errorf("Should have models available after recovery\nDescription: %s", scenario.description)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected failure but recovery succeeded\nDescription: %s", scenario.description)
				}
			}
		})
	}
}

// Helper functions
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func containsString(s, substr string) bool {
	return len(substr) == 0 || indexString(s, substr) >= 0
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
