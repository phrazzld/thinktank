package models

import "fmt"

// ModelInfo contains metadata for a single LLM model.
// This struct replaces the complex registry system with simple hardcoded definitions.
type ModelInfo struct {
	// Provider identifies the LLM service provider (openai, gemini, openrouter)
	Provider string `json:"provider"`

	// APIModelID is the model identifier used in API calls to the provider
	APIModelID string `json:"api_model_id"`

	// ContextWindow is the maximum combined tokens for input + output
	ContextWindow int `json:"context_window"`

	// MaxOutputTokens is the maximum tokens allowed for generation
	MaxOutputTokens int `json:"max_output_tokens"`

	// DefaultParams contains provider-specific default parameters (temperature, top_p, etc.)
	DefaultParams map[string]interface{} `json:"default_params"`
}

// ModelDefinitions contains hardcoded metadata for all supported LLM models.
// This replaces the complex YAML-based registry system with simple, direct access.
// TODO: Make this private once access functions are implemented
var ModelDefinitions = map[string]ModelInfo{
	// OpenAI Models
	"gpt-4.1": {
		Provider:        "openai",
		APIModelID:      "gpt-4.1",
		ContextWindow:   1000000,
		MaxOutputTokens: 200000,
		DefaultParams: map[string]interface{}{
			"temperature":       0.7,
			"top_p":             1.0,
			"frequency_penalty": 0.0,
			"presence_penalty":  0.0,
		},
	},
	"o4-mini": {
		Provider:        "openai",
		APIModelID:      "o4-mini",
		ContextWindow:   200000,
		MaxOutputTokens: 200000,
		DefaultParams: map[string]interface{}{
			"temperature":       1.0,
			"top_p":             1.0,
			"frequency_penalty": 0.0,
			"presence_penalty":  0.0,
			"reasoning": map[string]interface{}{
				"effort": "high",
			},
		},
	},

	// Gemini Models
	"gemini-2.5-pro": {
		Provider:        "gemini",
		APIModelID:      "gemini-2.5-pro",
		ContextWindow:   1000000,
		MaxOutputTokens: 65000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
			"top_k":       40,
		},
	},
	"gemini-2.5-flash": {
		Provider:        "gemini",
		APIModelID:      "gemini-2.5-flash",
		ContextWindow:   1000000,
		MaxOutputTokens: 65000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
			"top_k":       40,
		},
	},

	// OpenRouter Models
	"openrouter/deepseek/deepseek-chat-v3-0324": {
		Provider:        "openrouter",
		APIModelID:      "deepseek/deepseek-chat-v3-0324",
		ContextWindow:   65536,
		MaxOutputTokens: 8192,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
	},
	"openrouter/deepseek/deepseek-r1": {
		Provider:        "openrouter",
		APIModelID:      "deepseek/deepseek-r1",
		ContextWindow:   131072,
		MaxOutputTokens: 33792,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
	},
	"openrouter/x-ai/grok-3-beta": {
		Provider:        "openrouter",
		APIModelID:      "x-ai/grok-3-beta",
		ContextWindow:   131072,
		MaxOutputTokens: 131072,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
	},
}

// GetModelInfo returns model metadata for the given model name.
// Returns an error if the model is not supported.
func GetModelInfo(name string) (ModelInfo, error) {
	if info, exists := ModelDefinitions[name]; exists {
		return info, nil
	}
	return ModelInfo{}, fmt.Errorf("unknown model: %s", name)
}

// GetProviderForModel returns the provider name for the given model.
// Returns an error if the model is not supported.
func GetProviderForModel(name string) (string, error) {
	info, err := GetModelInfo(name)
	if err != nil {
		return "", err
	}
	return info.Provider, nil
}
