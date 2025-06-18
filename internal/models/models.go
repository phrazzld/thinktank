package models

// Imports will be added as functions are implemented

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

// TODO: Implement modelDefinitions map and functions
// This file will be completed in subsequent tasks
