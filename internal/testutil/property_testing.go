package testutil

import (
	"pgregory.net/rapid"
)

// Property-based testing utilities using Rapid library.
// This file provides generators for common project types and testing patterns.

// ContentGenerator creates generators for various content strings
type ContentGenerator struct{}

// NewContentGenerator creates a new content generator instance
func NewContentGenerator() *ContentGenerator {
	return &ContentGenerator{}
}

// SimpleText generates basic text content
func (g *ContentGenerator) SimpleText() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z0-9\s.,!?]{1,100}`)
}

// Prompt generates realistic prompt text
func (g *ContentGenerator) Prompt() *rapid.Generator[string] {
	prompts := []string{
		"Write a function that",
		"Explain how to",
		"Create a plan for",
		"Analyze the following",
		"Generate code that",
		"Describe the process of",
		"What are the benefits of",
		"How would you implement",
	}

	return rapid.Custom(func(t *rapid.T) string {
		prefix := rapid.SampledFrom(prompts).Draw(t, "prefix")
		suffix := rapid.StringMatching(`[a-zA-Z0-9\s.,!?]{10,50}`).Draw(t, "suffix")
		return prefix + " " + suffix
	})
}

// ModelName generates valid model names
func (g *ContentGenerator) ModelName() *rapid.Generator[string] {
	models := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4o",
		"gemini-pro",
		"gemini-pro-vision",
		"claude-3-opus",
		"claude-3-sonnet",
	}

	return rapid.SampledFrom(models)
}

// ParameterGenerator creates generators for API parameters
type ParameterGenerator struct{}

// NewParameterGenerator creates a new parameter generator instance
func NewParameterGenerator() *ParameterGenerator {
	return &ParameterGenerator{}
}

// Temperature generates valid temperature values
func (g *ParameterGenerator) Temperature() *rapid.Generator[float64] {
	return rapid.Float64Range(0.0, 2.0)
}

// MaxTokens generates valid max token values
func (g *ParameterGenerator) MaxTokens() *rapid.Generator[int] {
	return rapid.IntRange(1, 4096)
}

// TopP generates valid top-p values
func (g *ParameterGenerator) TopP() *rapid.Generator[float64] {
	return rapid.Float64Range(0.0, 1.0)
}

// PresencePenalty generates valid presence penalty values
func (g *ParameterGenerator) PresencePenalty() *rapid.Generator[float64] {
	return rapid.Float64Range(-2.0, 2.0)
}

// FrequencyPenalty generates valid frequency penalty values
func (g *ParameterGenerator) FrequencyPenalty() *rapid.Generator[float64] {
	return rapid.Float64Range(-2.0, 2.0)
}

// ValidParameterMap generates a map of valid API parameters
func (g *ParameterGenerator) ValidParameterMap() *rapid.Generator[map[string]interface{}] {
	return rapid.Custom(func(t *rapid.T) map[string]interface{} {
		params := make(map[string]interface{})

		// Optionally include each parameter
		if rapid.Bool().Draw(t, "include_temperature") {
			params["temperature"] = g.Temperature().Draw(t, "temperature")
		}
		if rapid.Bool().Draw(t, "include_max_tokens") {
			params["max_tokens"] = g.MaxTokens().Draw(t, "max_tokens")
		}
		if rapid.Bool().Draw(t, "include_top_p") {
			params["top_p"] = g.TopP().Draw(t, "top_p")
		}
		if rapid.Bool().Draw(t, "include_presence_penalty") {
			params["presence_penalty"] = g.PresencePenalty().Draw(t, "presence_penalty")
		}
		if rapid.Bool().Draw(t, "include_frequency_penalty") {
			params["frequency_penalty"] = g.FrequencyPenalty().Draw(t, "frequency_penalty")
		}

		return params
	})
}

// ConfigGenerator creates generators for configuration structures
type ConfigGenerator struct{}

// NewConfigGenerator creates a new config generator instance
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// APIKey generates test API keys with proper prefixes
func (g *ConfigGenerator) APIKey() *rapid.Generator[string] {
	return rapid.StringMatching(`test-[a-zA-Z0-9]{10,50}`)
}

// URL generates valid URLs
func (g *ConfigGenerator) URL() *rapid.Generator[string] {
	schemes := []string{"http://", "https://"}
	domains := []string{
		"api.openai.com",
		"generativelanguage.googleapis.com",
		"openrouter.ai",
		"localhost:8080",
		"test-server.local",
	}

	return rapid.Custom(func(t *rapid.T) string {
		scheme := rapid.SampledFrom(schemes).Draw(t, "scheme")
		domain := rapid.SampledFrom(domains).Draw(t, "domain")
		path := rapid.StringMatching(`(/[a-zA-Z0-9_-]*)*`).Draw(t, "path")
		return scheme + domain + path
	})
}

// TextProcessor provides utilities for testing text processing functions
type TextProcessor struct{}

// NewTextProcessor creates a new text processor instance
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{}
}

// NonEmptyText generates non-empty text strings
func (g *TextProcessor) NonEmptyText() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z0-9\s.,!?]{1,1000}`)
}

// BoundedText generates text within specified length bounds
func (g *TextProcessor) BoundedText(minLen, maxLen int) *rapid.Generator[string] {
	charset := rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 .,!?-"))
	return rapid.StringOfN(charset, minLen, maxLen, maxLen)
}

// TokenCount generates realistic token counts
func (g *TextProcessor) TokenCount() *rapid.Generator[int] {
	return rapid.IntRange(0, 100000)
}

// ErrorScenarioGenerator creates generators for error testing scenarios
type ErrorScenarioGenerator struct{}

// NewErrorScenarioGenerator creates a new error scenario generator instance
func NewErrorScenarioGenerator() *ErrorScenarioGenerator {
	return &ErrorScenarioGenerator{}
}

// HTTPStatusCode generates HTTP status codes for error testing
func (g *ErrorScenarioGenerator) HTTPStatusCode() *rapid.Generator[int] {
	// Common HTTP status codes for testing
	codes := []int{200, 400, 401, 403, 404, 429, 500, 502, 503, 504}
	return rapid.SampledFrom(codes)
}

// ErrorMessage generates realistic error messages
func (g *ErrorScenarioGenerator) ErrorMessage() *rapid.Generator[string] {
	messages := []string{
		"Invalid API key",
		"Rate limit exceeded",
		"Model not found",
		"Invalid request",
		"Server error",
		"Timeout occurred",
		"Network error",
		"Authentication failed",
		"Insufficient credits",
		"Content filtered",
	}
	return rapid.SampledFrom(messages)
}

// Global generators for convenience
var (
	// Content generators
	ContentGen   = NewContentGenerator()
	ParameterGen = NewParameterGenerator()
	ConfigGen    = NewConfigGenerator()
	TextProc     = NewTextProcessor()
	ErrorScenGen = NewErrorScenarioGenerator()
)

// Common generator shortcuts for frequently used patterns
var (
	// Text and content
	SimpleText   = ContentGen.SimpleText()
	PromptText   = ContentGen.Prompt()
	ModelName    = ContentGen.ModelName()
	NonEmptyText = TextProc.NonEmptyText()

	// Parameters
	Temperature = ParameterGen.Temperature()
	MaxTokens   = ParameterGen.MaxTokens()
	ValidParams = ParameterGen.ValidParameterMap()

	// Configuration
	TestAPIKey = ConfigGen.APIKey()
	TestURL    = ConfigGen.URL()

	// Error scenarios
	HTTPStatus = ErrorScenGen.HTTPStatusCode()
	ErrorMsg   = ErrorScenGen.ErrorMessage()

	// Numeric values
	TokenCount = TextProc.TokenCount()
)
