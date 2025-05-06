// Package testutil provides utilities for testing in the thinktank project
package testutil

import (
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

// Standard model parameter definitions used in tests
var (
	// TemperatureParam is a standard temperature parameter definition
	TemperatureParam = registry.ParameterDefinition{
		Type:    "float",
		Default: 0.7,
		Min:     0.0,
		Max:     1.0,
	}

	// TopPParam is a standard top-p parameter definition
	TopPParam = registry.ParameterDefinition{
		Type:    "float",
		Default: 0.95,
		Min:     0.0,
		Max:     1.0,
	}

	// MaxTokensParam is a standard max tokens parameter definition
	MaxTokensParam = registry.ParameterDefinition{
		Type:    "int",
		Default: 1024,
		Min:     1,
		Max:     8192,
	}

	// ReasoningEffortParam is a standard reasoning effort parameter definition
	ReasoningEffortParam = registry.ParameterDefinition{
		Type:    "float",
		Default: 0.5,
		Min:     0.0,
		Max:     1.0,
	}

	// FrequencyPenaltyParam is a standard frequency penalty parameter
	FrequencyPenaltyParam = registry.ParameterDefinition{
		Type:    "float",
		Default: 0.0,
		Min:     -2.0,
		Max:     2.0,
	}

	// PresencePenaltyParam is a standard presence penalty parameter
	PresencePenaltyParam = registry.ParameterDefinition{
		Type:    "float",
		Default: 0.0,
		Min:     -2.0,
		Max:     2.0,
	}

	// ModelParam is a standard model parameter definition
	ModelParam = registry.ParameterDefinition{
		Type:    "string",
		Default: "gpt-4",
	}

	// StandardParameters contains common parameters used by most models
	StandardParameters = map[string]registry.ParameterDefinition{
		"temperature":       TemperatureParam,
		"top_p":             TopPParam,
		"max_tokens":        MaxTokensParam,
		"frequency_penalty": FrequencyPenaltyParam,
		"presence_penalty":  PresencePenaltyParam,
	}

	// GeminiParameters contains parameters specific to Gemini models
	GeminiParameters = map[string]registry.ParameterDefinition{
		"temperature":      TemperatureParam,
		"top_p":            TopPParam,
		"max_tokens":       MaxTokensParam,
		"reasoning_effort": ReasoningEffortParam,
	}
)

// Standard provider definitions
var (
	// OpenAIProvider is a standard OpenAI provider definition
	OpenAIProvider = registry.ProviderDefinition{
		Name:    "openai",
		BaseURL: "",
	}

	// CustomOpenAIProvider is an OpenAI provider with custom base URL
	CustomOpenAIProvider = registry.ProviderDefinition{
		Name:    "openai-custom",
		BaseURL: "https://custom-openai.example.com/v1",
	}

	// GeminiProvider is a standard Gemini provider definition
	GeminiProvider = registry.ProviderDefinition{
		Name:    "gemini",
		BaseURL: "",
	}

	// OpenRouterProvider is a standard OpenRouter provider definition
	OpenRouterProvider = registry.ProviderDefinition{
		Name:    "openrouter",
		BaseURL: "",
	}

	// CustomOpenRouterProvider is an OpenRouter provider with custom base URL
	CustomOpenRouterProvider = registry.ProviderDefinition{
		Name:    "openrouter-custom",
		BaseURL: "https://custom-openrouter.example.com/api",
	}
)

// Standard model definitions
var (
	// GPT4Model is a standard GPT-4 model definition
	GPT4Model = registry.ModelDefinition{
		Name:       "gpt-4",
		Provider:   "openai",
		APIModelID: "gpt-4",
		Parameters: StandardParameters,
	}

	// GPT4Turbo is a standard GPT-4 Turbo model definition
	GPT4Turbo = registry.ModelDefinition{
		Name:       "gpt-4-turbo",
		Provider:   "openai",
		APIModelID: "gpt-4-turbo-preview",
		Parameters: StandardParameters,
	}

	// GPT4Vision is a standard GPT-4 Vision model definition
	GPT4Vision = registry.ModelDefinition{
		Name:       "gpt-4-vision",
		Provider:   "openai",
		APIModelID: "gpt-4-vision-preview",
		Parameters: StandardParameters,
	}

	// GPT35Turbo is a standard GPT-3.5 Turbo model definition
	GPT35Turbo = registry.ModelDefinition{
		Name:       "gpt-3.5-turbo",
		Provider:   "openai",
		APIModelID: "gpt-3.5-turbo",
		Parameters: StandardParameters,
	}

	// Gemini1Pro is a standard Gemini 1.0 Pro model definition
	Gemini1Pro = registry.ModelDefinition{
		Name:       "gemini-1.0-pro",
		Provider:   "gemini",
		APIModelID: "gemini-1.0-pro",
		Parameters: GeminiParameters,
	}

	// Gemini15Pro is a standard Gemini 1.5 Pro model definition
	Gemini15Pro = registry.ModelDefinition{
		Name:       "gemini-1.5-pro",
		Provider:   "gemini",
		APIModelID: "gemini-1.5-pro-latest",
		Parameters: GeminiParameters,
	}

	// Claude3Opus is a standard Claude 3 Opus model definition via OpenRouter
	Claude3Opus = registry.ModelDefinition{
		Name:       "claude-3-opus",
		Provider:   "openrouter",
		APIModelID: "anthropic/claude-3-opus",
		Parameters: StandardParameters,
	}

	// Claude3Sonnet is a standard Claude 3 Sonnet model definition via OpenRouter
	Claude3Sonnet = registry.ModelDefinition{
		Name:       "claude-3-sonnet",
		Provider:   "openrouter",
		APIModelID: "anthropic/claude-3-sonnet",
		Parameters: StandardParameters,
	}
)

// CreateTestModels returns a map of model names to model definitions
// suitable for testing. This is useful for mocking a registry.
func CreateTestModels() map[string]registry.ModelDefinition {
	return map[string]registry.ModelDefinition{
		"gpt-4":           GPT4Model,
		"gpt-4-turbo":     GPT4Turbo,
		"gpt-4-vision":    GPT4Vision,
		"gpt-3.5-turbo":   GPT35Turbo,
		"gemini-1.0-pro":  Gemini1Pro,
		"gemini-1.5-pro":  Gemini15Pro,
		"claude-3-opus":   Claude3Opus,
		"claude-3-sonnet": Claude3Sonnet,
		"test-model": {
			Name:       "test-model",
			Provider:   "test-provider",
			APIModelID: "test-model-id",
			Parameters: StandardParameters,
		},
	}
}

// CreateTestProviders returns a map of provider names to provider definitions
// suitable for testing. This is useful for mocking a registry.
func CreateTestProviders() map[string]registry.ProviderDefinition {
	return map[string]registry.ProviderDefinition{
		"openai":            OpenAIProvider,
		"openai-custom":     CustomOpenAIProvider,
		"gemini":            GeminiProvider,
		"openrouter":        OpenRouterProvider,
		"openrouter-custom": CustomOpenRouterProvider,
		"test-provider": {
			Name:    "test-provider",
			BaseURL: "https://test.example.com",
		},
	}
}

// Standard LLM responses
var (
	// BasicSuccessResponse is a simple success response with content
	BasicSuccessResponse = &llm.ProviderResult{
		Content:      "This is a test response from the LLM",
		FinishReason: "stop",
		Truncated:    false,
	}

	// TruncatedResponse is a response truncated due to length
	TruncatedResponse = &llm.ProviderResult{
		Content:      "This response was truncated due to length limitations...",
		FinishReason: "length",
		Truncated:    true,
	}

	// SafetyBlockedResponse is a response blocked by safety filters
	SafetyBlockedResponse = &llm.ProviderResult{
		Content:      "",
		FinishReason: "safety",
		Truncated:    false,
		SafetyInfo: []llm.Safety{
			{
				Category: "harmful",
				Blocked:  true,
				Score:    0.9,
			},
		},
	}

	// PartialSafetyBlockedResponse is a response that triggered safety filters but wasn't completely blocked
	PartialSafetyBlockedResponse = &llm.ProviderResult{
		Content:      "This is a partial response with some content filtered.",
		FinishReason: "safety",
		Truncated:    true,
		SafetyInfo: []llm.Safety{
			{
				Category: "harmful",
				Blocked:  false,
				Score:    0.6,
			},
		},
	}

	// EmptyResponse is a response with no content
	EmptyResponse = &llm.ProviderResult{
		Content:      "",
		FinishReason: "stop",
		Truncated:    false,
	}

	// WhitespaceResponse is a response with only whitespace
	WhitespaceResponse = &llm.ProviderResult{
		Content:      "   \n\t   ",
		FinishReason: "stop",
		Truncated:    false,
	}

	// JSONResponse is a response containing JSON
	JSONResponse = &llm.ProviderResult{
		Content:      `{"name": "Test", "value": 123, "items": ["a", "b", "c"]}`,
		FinishReason: "stop",
		Truncated:    false,
	}

	// CodeResponse is a response containing code
	CodeResponse = &llm.ProviderResult{
		Content: `
// Here's an example function
function hello() {
    console.log("Hello, world!");
    return 42;
}`,
		FinishReason: "stop",
		Truncated:    false,
	}
)

// CreateSuccessResponse creates a success response with the given content
func CreateSuccessResponse(content string) *llm.ProviderResult {
	return &llm.ProviderResult{
		Content:      content,
		FinishReason: "stop",
		Truncated:    false,
	}
}

// CreateTruncatedResponse creates a truncated response with the given content
func CreateTruncatedResponse(content string) *llm.ProviderResult {
	return &llm.ProviderResult{
		Content:      content,
		FinishReason: "length",
		Truncated:    true,
	}
}

// CreateSafetyBlockedResponse creates a safety-blocked response with details
func CreateSafetyBlockedResponse(category string, score float32) *llm.ProviderResult {
	return &llm.ProviderResult{
		Content:      "",
		FinishReason: "safety",
		Truncated:    false,
		SafetyInfo: []llm.Safety{
			{
				Category: category,
				Blocked:  true,
				Score:    score,
			},
		},
	}
}

// Standard errors used in tests
var (
	// CreateAuthError creates an authentication error
	CreateAuthError = func(provider string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			provider,
			llm.CategoryAuth,
			llm.ErrAPICall,
			"Invalid API key",
		)
	}

	// CreateRateLimitError creates a rate limit error
	CreateRateLimitError = func(provider string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			provider,
			llm.CategoryRateLimit,
			llm.ErrAPICall,
			"Rate limit exceeded",
		)
	}

	// CreateSafetyError creates a safety filter error
	CreateSafetyError = func(provider string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			provider,
			llm.CategoryContentFiltered,
			llm.ErrSafetyBlocked,
			"Content blocked by safety filters",
		)
	}

	// CreateInputLimitError creates an input token limit error
	CreateInputLimitError = func(provider string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			provider,
			llm.CategoryInputLimit,
			llm.ErrAPICall,
			"Input token limit exceeded",
		)
	}

	// CreateNetworkError creates a network error
	CreateNetworkError = func(provider string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			provider,
			llm.CategoryNetwork,
			llm.ErrAPICall,
			"Network connection error",
		)
	}

	// CreateServerError creates a server error
	CreateServerError = func(provider string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			provider,
			llm.CategoryServer,
			llm.ErrAPICall,
			"Server error occurred",
		)
	}

	// CreateModelNotFoundError creates a model not found error
	CreateModelNotFoundError = func(modelName string) *llm.LLMError {
		return llm.CreateStandardErrorWithMessage(
			"registry",
			llm.CategoryNotFound,
			llm.ErrModelNotFound,
			"Model '"+modelName+"' not found",
		)
	}
)
