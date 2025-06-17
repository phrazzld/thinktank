package models

// Imports will be added as functions are implemented

// ModelInfo contains metadata for a single LLM model.
type ModelInfo struct {
	Provider        string                 `json:"provider"`
	APIModelID      string                 `json:"api_model_id"`
	ContextWindow   int                    `json:"context_window"`
	MaxOutputTokens int                    `json:"max_output_tokens"`
	DefaultParams   map[string]interface{} `json:"default_params"`
}

// TODO: Implement modelDefinitions map and functions
// This file will be completed in subsequent tasks
