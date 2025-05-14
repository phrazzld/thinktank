package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestGetConfigPath tests the GetConfigPath function
func TestGetConfigPath(t *testing.T) {
	loader := NewConfigLoader(nil)

	path, err := loader.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ConfigDirName, ModelsConfigFileName)
	if path != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, path)
	}
}

// TestLoadConfigFileNotFound tests the Load function when the config file is not found
func TestLoadConfigFileNotFound(t *testing.T) {
	// Create a temporary setup with a non-existent file path
	tmpPath := filepath.Join(os.TempDir(), "non-existent-file.yaml")

	// Create a custom loader with an overridden GetConfigPath method
	loader := &ConfigLoader{
		Logger: logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel),
	}

	// Mock the GetConfigPath method
	origGetConfigPath := loader.GetConfigPath
	loader.GetConfigPath = func() (string, error) {
		return tmpPath, nil
	}
	defer func() { loader.GetConfigPath = origGetConfigPath }()

	// Attempt to load the config
	_, err := loader.Load()
	if err == nil {
		t.Fatal("Expected error when config file not found, got nil")
	}

	// Check that the error message contains "not found"
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error message should indicate file not found, got: %v", err)
	}
}

// TestLoadInvalidYAML tests the Load function with an invalid YAML file
func TestLoadInvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: Failed to remove temp file: %v", err)
		}
	}()

	// Write invalid YAML to the file
	invalidYAML := "invalid yaml: [\nthis is not valid"
	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Create a custom loader with an overridden GetConfigPath method
	loader := &ConfigLoader{
		Logger: logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel),
	}

	// Mock the GetConfigPath method
	origGetConfigPath := loader.GetConfigPath
	loader.GetConfigPath = func() (string, error) {
		return tmpFile.Name(), nil
	}
	defer func() { loader.GetConfigPath = origGetConfigPath }()

	_, err = loader.Load()
	if err == nil {
		t.Fatal("Expected error when loading invalid YAML, got nil")
	}

	// Check that the error message contains "invalid YAML"
	if err.Error() == "" {
		t.Errorf("Error message should indicate invalid YAML: %v", err)
	}
}

// TestValidateConfig tests the validate function with various invalid configs
func TestValidateConfig(t *testing.T) {
	loader := NewConfigLoader(nil)

	// Test with empty config
	emptyConfig := &ModelsConfig{}
	err := loader.validate(emptyConfig)
	if err == nil {
		t.Error("Expected error with empty config, got nil")
	}

	// Test with missing providers
	configNoProviders := &ModelsConfig{
		APIKeySources: map[string]string{"test": "TEST_API_KEY"},
	}
	err = loader.validate(configNoProviders)
	if err == nil {
		t.Error("Expected error with missing providers, got nil")
	}

	// Test with missing models
	configNoModels := &ModelsConfig{
		APIKeySources: map[string]string{"test": "TEST_API_KEY"},
		Providers:     []ProviderDefinition{{Name: "test"}},
	}
	err = loader.validate(configNoModels)
	if err == nil {
		t.Error("Expected error with missing models, got nil")
	}

	// Test with invalid model (missing provider)
	configInvalidModel := &ModelsConfig{
		APIKeySources: map[string]string{"test": "TEST_API_KEY"},
		Providers:     []ProviderDefinition{{Name: "test"}},
		Models: []ModelDefinition{
			{
				Name: "test-model",
				// Missing Provider
				APIModelID: "test-model-id",
			},
		},
	}
	err = loader.validate(configInvalidModel)
	if err == nil {
		t.Error("Expected error with invalid model (missing provider), got nil")
	}

	// Test with invalid model (unknown provider)
	configUnknownProvider := &ModelsConfig{
		APIKeySources: map[string]string{"test": "TEST_API_KEY"},
		Providers:     []ProviderDefinition{{Name: "test"}},
		Models: []ModelDefinition{
			{
				Name:       "test-model",
				Provider:   "unknown", // Doesn't match any provider
				APIModelID: "test-model-id",
			},
		},
	}
	err = loader.validate(configUnknownProvider)
	if err == nil {
		t.Error("Expected error with invalid model (unknown provider), got nil")
	}

	// Test with valid config
	validConfig := &ModelsConfig{
		APIKeySources: map[string]string{"test": "TEST_API_KEY"},
		Providers:     []ProviderDefinition{{Name: "test"}},
		Models: []ModelDefinition{
			{
				Name:       "test-model",
				Provider:   "test",
				APIModelID: "test-model-id",
			},
		},
	}
	err = loader.validate(validConfig)
	if err != nil {
		t.Errorf("Expected no error with valid config, got: %v", err)
	}
}

// TestLoadValid tests the complete Load function with a valid config file
func TestLoadValid(t *testing.T) {
	// Create a temporary file with valid YAML
	tmpFile, err := os.CreateTemp("", "valid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: Failed to remove temp file: %v", err)
		}
	}()

	// Write valid YAML to the file
	validYAML := `
api_key_sources:
  openai: OPENAI_API_KEY
  gemini: GEMINI_API_KEY

providers:
  - name: openai
    base_url: https://api.openai.com/v1
  - name: gemini
    base_url: https://generativelanguage.googleapis.com

models:
  - name: gpt-4-turbo
    provider: openai
    api_model_id: gpt-4-turbo-preview
    context_window: 128000
    max_output_tokens: 4096
    parameters:
      temperature:
        type: float
        default: 0.7

  - name: gemini-1.5-pro
    provider: gemini
    api_model_id: gemini-1.5-pro-latest
    context_window: 1000000
    max_output_tokens: 8192
    parameters:
      temperature:
        type: float
        default: 0.8
`
	if _, err := tmpFile.WriteString(validYAML); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Create a custom loader with an overridden GetConfigPath method
	loader := &ConfigLoader{
		Logger: logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel),
	}

	// Mock the GetConfigPath method
	origGetConfigPath := loader.GetConfigPath
	loader.GetConfigPath = func() (string, error) {
		return tmpFile.Name(), nil
	}
	defer func() { loader.GetConfigPath = origGetConfigPath }()

	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Unexpected error loading valid config: %v", err)
	}

	// Verify the loaded config
	if len(config.APIKeySources) != 2 {
		t.Errorf("Expected 2 API key sources, got %d", len(config.APIKeySources))
	}

	if len(config.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(config.Providers))
	}

	if config.Providers[0].Name != "openai" {
		t.Errorf("Expected first provider to be 'openai', got '%s'", config.Providers[0].Name)
	}

	if len(config.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(config.Models))
	}

	if config.Models[0].Name != "gpt-4-turbo" {
		t.Errorf("Expected first model to be 'gpt-4-turbo', got '%s'", config.Models[0].Name)
	}

	// Token-related validation was removed in T036E
}
