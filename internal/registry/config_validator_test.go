package registry

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestDefaultModelsYAML validates that the default models.yaml file
// in the config directory is valid and can be loaded and parsed
func TestDefaultModelsYAML(t *testing.T) {
	// Get the path to the default models.yaml
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate up to the project root
	projectRoot := filepath.Join(wd, "..", "..")
	defaultConfigPath := filepath.Join(projectRoot, "config", "models.yaml")

	// Check if the file exists
	if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
		t.Fatalf("Default config file not found at %s", defaultConfigPath)
	}

	// Read the file
	data, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		t.Fatalf("Failed to read default config file: %v", err)
	}

	// Parse the YAML
	config, err := parseAndValidateYAML(data)
	if err != nil {
		t.Fatalf("Failed to parse default config file: %v", err)
	}

	// Validate the configuration
	if len(config.APIKeySources) == 0 {
		t.Error("Default config should have API key sources")
	}

	if len(config.Providers) == 0 {
		t.Error("Default config should have providers")
	}

	if len(config.Models) == 0 {
		t.Error("Default config should have models")
	}

	// Validate that required providers exist
	hasOpenAI := false
	hasGemini := false
	for _, provider := range config.Providers {
		if provider.Name == "openai" {
			hasOpenAI = true
		}
		if provider.Name == "gemini" {
			hasGemini = true
		}
	}

	if !hasOpenAI {
		t.Error("Default config should include OpenAI provider")
	}

	if !hasGemini {
		t.Error("Default config should include Gemini provider")
	}

	// Validate that all models have the required parameters
	for _, model := range config.Models {
		if model.Name == "" {
			t.Error("All models should have a name")
		}

		if model.Provider == "" {
			t.Errorf("Model %s is missing provider", model.Name)
		}

		if model.APIModelID == "" {
			t.Errorf("Model %s is missing api_model_id", model.Name)
		}

		if model.ContextWindow <= 0 {
			t.Errorf("Model %s has invalid context_window", model.Name)
		}

		if model.MaxOutputTokens <= 0 {
			t.Errorf("Model %s has invalid max_output_tokens", model.Name)
		}

		if len(model.Parameters) == 0 {
			t.Errorf("Model %s should have parameters", model.Name)
		} else {
			// Check for common parameters
			if _, ok := model.Parameters["temperature"]; !ok {
				t.Errorf("Model %s should have temperature parameter", model.Name)
			}
		}
	}
}

// Helper function to parse and validate YAML
func parseAndValidateYAML(data []byte) (*ModelsConfig, error) {
	var config ModelsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Create a config loader to validate the config
	loader := &ConfigLoader{}
	if err := loader.validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
