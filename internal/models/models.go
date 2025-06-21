package models

import (
	"fmt"
	"sort"
)

// ParameterConstraint defines validation rules for a single model parameter
type ParameterConstraint struct {
	// Type specifies the expected parameter type ("int", "float", "string")
	Type string `json:"type"`

	// MinValue is the minimum allowed value for numeric types (optional)
	MinValue *float64 `json:"min_value,omitempty"`

	// MaxValue is the maximum allowed value for numeric types (optional)
	MaxValue *float64 `json:"max_value,omitempty"`

	// EnumValues lists allowed string values for string type parameters (optional)
	EnumValues []string `json:"enum_values,omitempty"`
}

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

	// ParameterConstraints defines validation rules for each parameter this model supports
	ParameterConstraints map[string]ParameterConstraint `json:"parameter_constraints"`

	// MaxConcurrentRequests limits concurrent requests for this specific model (optional)
	// If nil, uses global concurrency settings. If set, enforces per-model limit.
	MaxConcurrentRequests *int `json:"max_concurrent_requests,omitempty"`
}

// Helper functions for creating parameter constraints

func floatConstraint(min, max float64) ParameterConstraint {
	return ParameterConstraint{
		Type:     "float",
		MinValue: &min,
		MaxValue: &max,
	}
}

func intConstraint(min, max float64) ParameterConstraint {
	return ParameterConstraint{
		Type:     "int",
		MinValue: &min,
		MaxValue: &max,
	}
}

// modelDefinitions contains hardcoded metadata for all supported LLM models.
// This replaces the complex YAML-based registry system with simple, direct access.
// Access is provided through public functions like GetModelInfo, ListAllModels, etc.
var modelDefinitions = map[string]ModelInfo{
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 1000000),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 200000),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"top_k":             intConstraint(1, 100),
			"max_output_tokens": intConstraint(1, 65000),
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"top_k":             intConstraint(1, 100),
			"max_output_tokens": intConstraint(1, 65000),
		},
	},

	"o3": {
		Provider:        "openai",
		APIModelID:      "o3",
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 200000),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 65536),
		},
	},
	"openrouter/deepseek/deepseek-r1-0528": {
		Provider:        "openrouter",
		APIModelID:      "deepseek/deepseek-r1-0528",
		ContextWindow:   128000,
		MaxOutputTokens: 32768,
		DefaultParams: map[string]interface{}{
			"temperature": 1.0,
			"top_p":       1.0,
			"stop":        []string{"<｜User｜>", "<｜end▁of▁sentence｜>"},
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 32768),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
			"top_k":             intConstraint(1, 100),
		},
		MaxConcurrentRequests: &[]int{1}[0], // Force sequential processing to avoid concurrency conflicts
	},
	"openrouter/deepseek/deepseek-chat-v3-0324:free": {
		Provider:        "openrouter",
		APIModelID:      "deepseek/deepseek-chat-v3-0324:free",
		ContextWindow:   65536,
		MaxOutputTokens: 8192,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 8192),
		},
	},
	"openrouter/deepseek/deepseek-r1-0528:free": {
		Provider:        "openrouter",
		APIModelID:      "deepseek/deepseek-r1-0528:free",
		ContextWindow:   163840,
		MaxOutputTokens: 32768,
		DefaultParams: map[string]interface{}{
			"temperature": 1.0,
			"top_p":       1.0,
			"stop":        []string{"<｜User｜>", "<｜end▁of▁sentence｜>"},
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":        floatConstraint(0.0, 2.0),
			"top_p":              floatConstraint(0.0, 1.0),
			"max_tokens":         intConstraint(1, 32768),
			"frequency_penalty":  floatConstraint(-2.0, 2.0),
			"presence_penalty":   floatConstraint(-2.0, 2.0),
			"repetition_penalty": floatConstraint(0.0, 2.0),
			"top_k":              intConstraint(1, 100),
		},
		MaxConcurrentRequests: &[]int{1}[0], // Force sequential processing to avoid concurrency conflicts
	},
	"openrouter/meta-llama/llama-3.3-70b-instruct": {
		Provider:        "openrouter",
		APIModelID:      "meta-llama/llama-3.3-70b-instruct",
		ContextWindow:   131072,
		MaxOutputTokens: 4096,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 4096),
		},
	},
	"openrouter/meta-llama/llama-4-maverick": {
		Provider:        "openrouter",
		APIModelID:      "meta-llama/llama-4-maverick",
		ContextWindow:   1048576,
		MaxOutputTokens: 16384,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":        floatConstraint(0.0, 2.0),
			"top_p":              floatConstraint(0.0, 1.0),
			"max_tokens":         intConstraint(1, 16384),
			"frequency_penalty":  floatConstraint(-2.0, 2.0),
			"presence_penalty":   floatConstraint(-2.0, 2.0),
			"repetition_penalty": floatConstraint(0.0, 2.0),
			"top_k":              intConstraint(1, 100),
		},
	},
	"openrouter/meta-llama/llama-4-scout": {
		Provider:        "openrouter",
		APIModelID:      "meta-llama/llama-4-scout",
		ContextWindow:   1048576,
		MaxOutputTokens: 4096,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":        floatConstraint(0.0, 2.0),
			"top_p":              floatConstraint(0.0, 1.0),
			"max_tokens":         intConstraint(1, 4096),
			"frequency_penalty":  floatConstraint(-2.0, 2.0),
			"presence_penalty":   floatConstraint(-2.0, 2.0),
			"repetition_penalty": floatConstraint(0.0, 2.0),
			"top_k":              intConstraint(1, 100),
		},
	},
	"openrouter/x-ai/grok-3-mini-beta": {
		Provider:        "openrouter",
		APIModelID:      "x-ai/grok-3-mini-beta",
		ContextWindow:   131072,
		MaxOutputTokens: 131072,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 131072),
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
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 131072),
		},
	},
	"openrouter/google/gemma-3-27b-it": {
		Provider:        "openrouter",
		APIModelID:      "google/gemma-3-27b-it",
		ContextWindow:   8192,
		MaxOutputTokens: 8192,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 8192),
		},
	},
}

// GetModelInfo returns model metadata for the given model name.
// Returns an error if the model is not supported.
func GetModelInfo(name string) (ModelInfo, error) {
	if info, exists := modelDefinitions[name]; exists {
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

// ListAllModels returns a sorted slice of all supported model names.
func ListAllModels() []string {
	models := make([]string, 0, len(modelDefinitions))
	for name := range modelDefinitions {
		models = append(models, name)
	}
	sort.Strings(models)
	return models
}

// ListModelsForProvider returns a sorted slice of model names for the given provider.
// Returns an empty slice if no models are found for the provider.
func ListModelsForProvider(provider string) []string {
	var models []string
	for name, info := range modelDefinitions {
		if info.Provider == provider {
			models = append(models, name)
		}
	}
	sort.Strings(models)
	return models
}

// GetAPIKeyEnvVar returns the environment variable name for the given provider's API key.
// Returns an empty string for unknown providers.
func GetAPIKeyEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	default:
		return ""
	}
}

// IsModelSupported returns true if the given model name is supported.
func IsModelSupported(name string) bool {
	_, exists := modelDefinitions[name]
	return exists
}

// ValidateParameter validates a parameter value against the constraints defined for the given model.
// Returns an error if the parameter is invalid, nil if valid.
func ValidateParameter(modelName, paramName string, value interface{}) error {
	modelInfo, err := GetModelInfo(modelName)
	if err != nil {
		return fmt.Errorf("model '%s' not supported: %w", modelName, err)
	}

	constraint, exists := modelInfo.ParameterConstraints[paramName]
	if !exists {
		// Parameter not defined in constraints - accept any value
		return nil
	}

	// Validate based on parameter type
	switch constraint.Type {
	case "float":
		return validateFloatParameter(paramName, value, constraint)
	case "int":
		return validateIntParameter(paramName, value, constraint)
	case "string":
		return validateStringParameter(paramName, value, constraint)
	default:
		return fmt.Errorf("parameter '%s' has unknown constraint type '%s'", paramName, constraint.Type)
	}
}

func validateFloatParameter(paramName string, value interface{}, constraint ParameterConstraint) error {
	var floatVal float64

	// Handle different numeric types
	switch v := value.(type) {
	case float64:
		floatVal = v
	case float32:
		floatVal = float64(v)
	case int:
		floatVal = float64(v)
	case int32:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		return fmt.Errorf("parameter '%s' must be a numeric value, got %T", paramName, value)
	}

	// Check bounds
	if constraint.MinValue != nil && floatVal < *constraint.MinValue {
		return fmt.Errorf("parameter '%s' value %.2f must be >= %.2f", paramName, floatVal, *constraint.MinValue)
	}
	if constraint.MaxValue != nil && floatVal > *constraint.MaxValue {
		return fmt.Errorf("parameter '%s' value %.2f must be <= %.2f", paramName, floatVal, *constraint.MaxValue)
	}

	return nil
}

func validateIntParameter(paramName string, value interface{}, constraint ParameterConstraint) error {
	var intVal int64

	// Handle different numeric types
	switch v := value.(type) {
	case int:
		intVal = int64(v)
	case int32:
		intVal = int64(v)
	case int64:
		intVal = v
	case float64:
		// Allow float64 that represents whole numbers (common in JSON)
		if v != float64(int64(v)) {
			return fmt.Errorf("parameter '%s' must be an integer, got float value %.2f", paramName, v)
		}
		intVal = int64(v)
	case float32:
		// Allow float32 that represents whole numbers
		if v != float32(int64(v)) {
			return fmt.Errorf("parameter '%s' must be an integer, got float value %.2f", paramName, v)
		}
		intVal = int64(v)
	default:
		return fmt.Errorf("parameter '%s' must be an integer value, got %T", paramName, value)
	}

	// Check bounds
	if constraint.MinValue != nil && float64(intVal) < *constraint.MinValue {
		return fmt.Errorf("parameter '%s' value %d must be >= %.0f", paramName, intVal, *constraint.MinValue)
	}
	if constraint.MaxValue != nil && float64(intVal) > *constraint.MaxValue {
		return fmt.Errorf("parameter '%s' value %d must be <= %.0f", paramName, intVal, *constraint.MaxValue)
	}

	return nil
}

func validateStringParameter(paramName string, value interface{}, constraint ParameterConstraint) error {
	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("parameter '%s' must be a string, got %T", paramName, value)
	}

	// Check enum values if specified
	if len(constraint.EnumValues) > 0 {
		for _, allowedValue := range constraint.EnumValues {
			if strVal == allowedValue {
				return nil
			}
		}
		return fmt.Errorf("parameter '%s' value '%s' must be one of: %v", paramName, strVal, constraint.EnumValues)
	}

	return nil
}
