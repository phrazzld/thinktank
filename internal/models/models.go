package models

import (
	"fmt"
	"os"
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

	// RateLimitRPM overrides the provider-specific default rate limit for this model (optional)
	// If nil, uses provider-specific default rate limits. If set, enforces per-model rate limit.
	RateLimitRPM *int `json:"rate_limit_rpm,omitempty"`

	// RequiresBYOK indicates if this model requires users to bring their own API key
	// When true, users must provide their provider-specific API key (e.g., OpenAI key for o3)
	RequiresBYOK bool `json:"requires_byok,omitempty"`
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
		Provider:        "openrouter",
		APIModelID:      "openai/gpt-4.1",
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
		Provider:        "openrouter",
		APIModelID:      "openai/o4-mini",
		ContextWindow:   200000,
		MaxOutputTokens: 200000,
		DefaultParams: map[string]interface{}{
			"temperature":       1.0,
			"top_p":             1.0,
			"frequency_penalty": 0.0,
			"presence_penalty":  0.0,
			"reasoning_effort":  "high",
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 200000),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
			"reasoning_effort":  {Type: "string", EnumValues: []string{"low", "medium", "high"}},
		},
	},

	// Gemini Models
	"gemini-2.5-pro": {
		Provider:        "openrouter",
		APIModelID:      "google/gemini-2.5-pro",
		ContextWindow:   1000000,
		MaxOutputTokens: 65000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
			"top_k":       40,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"top_k":       intConstraint(1, 100),
			"max_tokens":  intConstraint(1, 65000),
		},
	},
	"gemini-2.5-flash": {
		Provider:        "openrouter",
		APIModelID:      "google/gemini-2.5-flash",
		ContextWindow:   1000000,
		MaxOutputTokens: 65000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
			"top_k":       40,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"top_k":       intConstraint(1, 100),
			"max_tokens":  intConstraint(1, 65000),
		},
	},

	"o3": {
		Provider:        "openrouter",
		APIModelID:      "openai/o3",
		ContextWindow:   200000,
		MaxOutputTokens: 200000,
		DefaultParams: map[string]interface{}{
			"temperature":       1.0,
			"top_p":             1.0,
			"frequency_penalty": 0.0,
			"presence_penalty":  0.0,
			"reasoning_effort":  "high", // Maximum reasoning effort for complex tasks
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 200000),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
			"reasoning_effort":  {Type: "string", EnumValues: []string{"low", "medium", "high"}}, // String enum values
		},
		RequiresBYOK: true, // o3 requires users to bring their own OpenAI API key
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
		RateLimitRPM:          &[]int{5}[0], // Very low rate limit for this specific model
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
		RateLimitRPM:          &[]int{3}[0], // Very low rate limit for free tier
	},
	"openrouter/meta-llama/llama-3.3-70b-instruct": {
		Provider:        "openrouter",
		APIModelID:      "meta/llama-3.3-70b-instruct",
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
		APIModelID:      "meta/llama-4-maverick",
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
		APIModelID:      "meta/llama-4-scout",
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
	"grok-4": {
		Provider:        "openrouter",
		APIModelID:      "x-ai/grok-4",
		ContextWindow:   256000,
		MaxOutputTokens: 256000, // Using context window as max since OpenRouter page doesn't specify
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.95,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 256000),
		},
	},

	// Inception Models
	"mercury": {
		Provider:        "openrouter",
		APIModelID:      "inception/mercury",
		ContextWindow:   32000,
		MaxOutputTokens: 16000,
		DefaultParams: map[string]interface{}{
			"temperature":       0.7,
			"top_p":             1.0,
			"frequency_penalty": 0.0,
			"presence_penalty":  0.0,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature":       floatConstraint(0.0, 2.0),
			"top_p":             floatConstraint(0.0, 1.0),
			"max_tokens":        intConstraint(1, 16000),
			"frequency_penalty": floatConstraint(-2.0, 2.0),
			"presence_penalty":  floatConstraint(-2.0, 2.0),
		},
	},

	// Test models for integration testing only
	// These models are used by integration tests to simulate various scenarios
	"model1": {
		Provider:        "test",
		APIModelID:      "test-model-1",
		ContextWindow:   10000,
		MaxOutputTokens: 5000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       1.0,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 5000),
		},
	},
	"model2": {
		Provider:        "test",
		APIModelID:      "test-model-2",
		ContextWindow:   10000,
		MaxOutputTokens: 5000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       1.0,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 5000),
		},
	},
	"model3": {
		Provider:        "test",
		APIModelID:      "test-model-3",
		ContextWindow:   10000,
		MaxOutputTokens: 5000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       1.0,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 5000),
		},
	},
	"synthesis-model": {
		Provider:        "test",
		APIModelID:      "test-synthesis-model",
		ContextWindow:   2000,
		MaxOutputTokens: 1000,
		DefaultParams: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       1.0,
		},
		ParameterConstraints: map[string]ParameterConstraint{
			"temperature": floatConstraint(0.0, 2.0),
			"top_p":       floatConstraint(0.0, 1.0),
			"max_tokens":  intConstraint(1, 1000),
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

// EstimateTokensFromText provides a rough estimation of token count from text.
// Uses a conservative approximation where 1 token ≈ 1.33 characters.
// This estimation works reasonably well across different tokenizers.
func EstimateTokensFromText(text string) int {
	if text == "" {
		return 0
	}
	// Conservative estimation: 1 token per 1.33 characters (or 0.75 tokens per character)
	charCount := len(text)
	estimatedTokens := int(float64(charCount) * 0.75)

	// Add a small buffer for typical instruction overhead
	const instructionOverhead = 1000
	return estimatedTokens + instructionOverhead
}

// EstimateTokensFromStats estimates tokens from ContextStats.
// Includes the character count plus estimated instruction and formatting overhead.
func EstimateTokensFromStats(charCount int, instructionsText string) int {
	// Estimate tokens from the content
	contentTokens := int(float64(charCount) * 0.75)

	// Add instruction tokens
	instructionTokens := EstimateTokensFromText(instructionsText)

	// Add formatting overhead (file paths, markdown formatting, etc.)
	const formatOverhead = 500

	return contentTokens + instructionTokens + formatOverhead
}

// GetModelsWithMinContextWindow returns models that have at least the specified context window.
// Results are sorted by context window size in descending order (largest first).
func GetModelsWithMinContextWindow(minTokens int) []string {
	type modelWithContext struct {
		name          string
		contextWindow int
	}

	var candidates []modelWithContext
	for name, info := range modelDefinitions {
		if info.ContextWindow >= minTokens {
			candidates = append(candidates, modelWithContext{
				name:          name,
				contextWindow: info.ContextWindow,
			})
		}
	}

	// Sort by context window size (largest first), then by name for deterministic ordering
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].contextWindow == candidates[j].contextWindow {
			// Secondary sort by name for deterministic ordering when context windows are equal
			return candidates[i].name < candidates[j].name
		}
		return candidates[i].contextWindow > candidates[j].contextWindow
	})

	// Extract model names
	result := make([]string, len(candidates))
	for i, candidate := range candidates {
		result[i] = candidate.name
	}

	return result
}

// GetAvailableProviders returns a list of providers for which API keys are available.
// Checks environment variables for each provider's API key.
// Returns providers in a deterministic order.
// Also provides helpful error messages if obsolete API keys are detected.
func GetAvailableProviders() []string {
	var providers []string

	// Check for OpenRouter API key first
	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey != "" {
		providers = append(providers, "openrouter")
	} else {
		// Only show migration warnings if OpenRouter key is not set
		if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
			fmt.Fprintf(os.Stderr, "Warning: OPENAI_API_KEY detected but no longer used.\n")
			fmt.Fprintf(os.Stderr, "Please set OPENROUTER_API_KEY instead.\n")
			fmt.Fprintf(os.Stderr, "Get your key at: https://openrouter.ai/keys\n")
		}

		if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
			fmt.Fprintf(os.Stderr, "Warning: GEMINI_API_KEY detected but no longer used.\n")
			fmt.Fprintf(os.Stderr, "Please set OPENROUTER_API_KEY instead.\n")
			fmt.Fprintf(os.Stderr, "Get your key at: https://openrouter.ai/keys\n")
		}
	}

	// Test provider is only available when explicitly enabled for development/testing
	if os.Getenv("THINKTANK_ENABLE_TEST_MODELS") == "true" {
		providers = append(providers, "test")
	}

	return providers
}

// SelectModelsForInput intelligently selects models based on estimated input size and available API keys.
// Returns models that can handle the input size, filtered by available providers.
// Results are sorted by context window size (largest first).
func SelectModelsForInput(estimatedTokens int, availableProviders []string) []string {
	// Add safety margin - reserve 20% of context window for output and overhead
	safetyFactor := 1.25
	requiredContextWindow := int(float64(estimatedTokens) * safetyFactor)

	// Get all models with sufficient context window
	candidateModels := GetModelsWithMinContextWindow(requiredContextWindow)

	// Filter by available providers
	availableProviderSet := make(map[string]bool)
	for _, provider := range availableProviders {
		availableProviderSet[provider] = true
	}

	var filteredModels []string
	for _, modelName := range candidateModels {
		if info, exists := modelDefinitions[modelName]; exists {
			if availableProviderSet[info.Provider] {
				filteredModels = append(filteredModels, modelName)
			}
		}
	}

	return filteredModels
}

// GetLargestContextModel returns the model with the largest context window from the given list.
// Returns empty string if the list is empty or no models are found.
func GetLargestContextModel(modelNames []string) string {
	if len(modelNames) == 0 {
		return ""
	}

	largestModel := ""
	largestContext := 0

	for _, modelName := range modelNames {
		if info, exists := modelDefinitions[modelName]; exists {
			if info.ContextWindow > largestContext {
				largestContext = info.ContextWindow
				largestModel = modelName
			}
		}
	}

	return largestModel
}

// GetAPIKeyEnvVar returns the environment variable name for the given provider's API key.
// Returns an empty string for unknown providers.
// The "test" provider is used for integration testing and doesn't require an API key.
func GetAPIKeyEnvVar(provider string) string {
	switch provider {
	case "openrouter":
		return "OPENROUTER_API_KEY"
	case "test":
		return "" // Test provider doesn't require API key
	default:
		return "" // Obsolete providers (openai, gemini) no longer supported
	}
}

// GetProviderDefaultRateLimit returns the default rate limit (requests per minute) for a given provider.
// These defaults are based on typical provider capabilities and can be overridden via CLI flags.
func GetProviderDefaultRateLimit(provider string) int {
	switch provider {
	case "openrouter":
		return 20 // OpenRouter varies by model, conservative default
	case "test":
		return 1000 // Test provider has high limits for testing
	default:
		return 60 // Conservative fallback for unknown providers
	}
}

// GetModelRateLimit returns the effective rate limit for a specific model.
// Priority: model-specific override > provider default
func GetModelRateLimit(modelName string) (int, error) {
	modelInfo, err := GetModelInfo(modelName)
	if err != nil {
		return 0, err
	}

	// If model has a specific rate limit override, use it
	if modelInfo.RateLimitRPM != nil {
		return *modelInfo.RateLimitRPM, nil
	}

	// Otherwise, use provider default
	return GetProviderDefaultRateLimit(modelInfo.Provider), nil
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
