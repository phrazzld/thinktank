package registry

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestModelConfigParsing validates that the configuration structs can be properly
// unmarshaled from YAML, with all fields correctly mapped.
func TestModelConfigParsing(t *testing.T) {
	// Define a test YAML configuration
	yamlData := `
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
      top_p:
        type: float
        default: 1.0
      max_tokens:
        type: int
        default: 2048

  - name: gemini-1.5-pro
    provider: gemini
    api_model_id: gemini-1.5-pro-latest
    context_window: 1000000
    max_output_tokens: 8192
    parameters:
      temperature:
        type: float
        default: 0.8
      top_p:
        type: float
        default: 0.95
      top_k:
        type: int
        default: 40
`

	// Parse the YAML into our config structs
	var config ModelsConfig
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Validate API key sources
	if len(config.APIKeySources) != 2 {
		t.Errorf("Expected 2 API key sources, got %d", len(config.APIKeySources))
	}
	if config.APIKeySources["openai"] != "OPENAI_API_KEY" {
		t.Errorf("Expected OpenAI API key env var to be OPENAI_API_KEY, got %s", config.APIKeySources["openai"])
	}
	if config.APIKeySources["gemini"] != "GEMINI_API_KEY" {
		t.Errorf("Expected Gemini API key env var to be GEMINI_API_KEY, got %s", config.APIKeySources["gemini"])
	}

	// Validate providers
	if len(config.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(config.Providers))
	}

	openaiProvider := config.Providers[0]
	if openaiProvider.Name != "openai" || openaiProvider.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("OpenAI provider not parsed correctly: %+v", openaiProvider)
	}

	geminiProvider := config.Providers[1]
	if geminiProvider.Name != "gemini" || geminiProvider.BaseURL != "https://generativelanguage.googleapis.com" {
		t.Errorf("Gemini provider not parsed correctly: %+v", geminiProvider)
	}

	// Validate models
	if len(config.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(config.Models))
	}

	gptModel := config.Models[0]
	if gptModel.Name != "gpt-4-turbo" ||
		gptModel.Provider != "openai" ||
		gptModel.APIModelID != "gpt-4-turbo-preview" ||
		gptModel.ContextWindow != 128000 ||
		gptModel.MaxOutputTokens != 4096 {
		t.Errorf("GPT model not parsed correctly: %+v", gptModel)
	}

	// Check GPT model parameters
	if len(gptModel.Parameters) != 3 {
		t.Errorf("Expected 3 parameters for GPT model, got %d", len(gptModel.Parameters))
	}

	tempParam := gptModel.Parameters["temperature"]
	if tempParam.Type != "float" || tempParam.Default.(float64) != 0.7 {
		t.Errorf("Temperature parameter not parsed correctly: %+v", tempParam)
	}

	// Validate Gemini model
	geminiModel := config.Models[1]
	if geminiModel.Name != "gemini-1.5-pro" ||
		geminiModel.Provider != "gemini" ||
		geminiModel.APIModelID != "gemini-1.5-pro-latest" ||
		geminiModel.ContextWindow != 1000000 ||
		geminiModel.MaxOutputTokens != 8192 {
		t.Errorf("Gemini model not parsed correctly: %+v", geminiModel)
	}

	// Check Gemini model parameters
	if len(geminiModel.Parameters) != 3 {
		t.Errorf("Expected 3 parameters for Gemini model, got %d", len(geminiModel.Parameters))
	}

	topKParam := geminiModel.Parameters["top_k"]
	if topKParam.Type != "int" || topKParam.Default.(int) != 40 {
		t.Errorf("top_k parameter not parsed correctly: %+v", topKParam)
	}
}

// TestMarshalingAndUnmarshaling validates that the config structs can be
// marshaled to YAML and then unmarshaled back correctly.
func TestMarshalingAndUnmarshaling(t *testing.T) {
	// Create a simple config
	originalConfig := ModelsConfig{
		APIKeySources: map[string]string{
			"test": "TEST_API_KEY",
		},
		Providers: []ProviderDefinition{
			{
				Name:    "test-provider",
				BaseURL: "https://api.test.com",
			},
		},
		Models: []ModelDefinition{
			{
				Name:            "test-model",
				Provider:        "test-provider",
				APIModelID:      "test-model-v1",
				ContextWindow:   4096,
				MaxOutputTokens: 1024,
				Parameters: map[string]ParameterDefinition{
					"test-param": {
						Type:    "string",
						Default: "default-value",
					},
				},
			},
		},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(originalConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config to YAML: %v", err)
	}

	// Unmarshal back to a new config
	var newConfig ModelsConfig
	err = yaml.Unmarshal(yamlData, &newConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML to config: %v", err)
	}

	// Validate the unmarshaled config matches the original
	if len(newConfig.APIKeySources) != len(originalConfig.APIKeySources) {
		t.Errorf("API key sources count mismatch after unmarshal")
	}
	if newConfig.APIKeySources["test"] != originalConfig.APIKeySources["test"] {
		t.Errorf("API key source mismatch after unmarshal")
	}

	if len(newConfig.Providers) != len(originalConfig.Providers) {
		t.Errorf("Providers count mismatch after unmarshal")
	}
	if newConfig.Providers[0].Name != originalConfig.Providers[0].Name {
		t.Errorf("Provider name mismatch after unmarshal")
	}

	if len(newConfig.Models) != len(originalConfig.Models) {
		t.Errorf("Models count mismatch after unmarshal")
	}

	origModel := originalConfig.Models[0]
	newModel := newConfig.Models[0]

	if newModel.Name != origModel.Name ||
		newModel.Provider != origModel.Provider ||
		newModel.APIModelID != origModel.APIModelID ||
		newModel.ContextWindow != origModel.ContextWindow ||
		newModel.MaxOutputTokens != origModel.MaxOutputTokens {
		t.Errorf("Model fields mismatch after unmarshal")
	}

	if len(newModel.Parameters) != len(origModel.Parameters) {
		t.Errorf("Parameters count mismatch after unmarshal")
	}

	origParam := origModel.Parameters["test-param"]
	newParam := newModel.Parameters["test-param"]

	if newParam.Type != origParam.Type {
		t.Errorf("Parameter type mismatch after unmarshal")
	}
}
