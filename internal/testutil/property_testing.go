package testutil

import (
	"reflect"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

// Property-based testing utilities using Gopter library.
// This file provides generators for common project types and testing patterns.

// ContentGenerator creates generators for various content strings
type ContentGenerator struct{}

// NewContentGenerator creates a new content generator instance
func NewContentGenerator() *ContentGenerator {
	return &ContentGenerator{}
}

// SimpleText generates basic text content
func (g *ContentGenerator) SimpleText() gopter.Gen {
	return gen.RegexMatch(`[a-zA-Z0-9\s.,!?]{1,100}`)
}

// Prompt generates realistic prompt text
func (g *ContentGenerator) Prompt() gopter.Gen {
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

	// Convert string slice to interface{} slice for OneConstOf
	ifaces := make([]interface{}, len(prompts))
	for i, p := range prompts {
		ifaces[i] = p
	}
	return gen.OneConstOf(ifaces...).FlatMap(func(prefix interface{}) gopter.Gen {
		return gen.RegexMatch(`[a-zA-Z0-9\s.,!?]{10,50}`).Map(func(suffix interface{}) interface{} {
			return prefix.(string) + " " + suffix.(string)
		})
	}, reflect.TypeOf(""))
}

// ModelName generates valid model names
func (g *ContentGenerator) ModelName() gopter.Gen {
	models := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4o",
		"gemini-pro",
		"gemini-pro-vision",
		"claude-3-opus",
		"claude-3-sonnet",
	}

	// Convert string slice to interface{} slice for OneConstOf
	ifaces := make([]interface{}, len(models))
	for i, m := range models {
		ifaces[i] = m
	}
	return gen.OneConstOf(ifaces...)
}

// ParameterGenerator creates generators for API parameters
type ParameterGenerator struct{}

// NewParameterGenerator creates a new parameter generator instance
func NewParameterGenerator() *ParameterGenerator {
	return &ParameterGenerator{}
}

// Temperature generates valid temperature values
func (g *ParameterGenerator) Temperature() gopter.Gen {
	return gen.Float64Range(0.0, 2.0)
}

// MaxTokens generates valid max token values
func (g *ParameterGenerator) MaxTokens() gopter.Gen {
	return gen.IntRange(1, 4096)
}

// TopP generates valid top-p values
func (g *ParameterGenerator) TopP() gopter.Gen {
	return gen.Float64Range(0.0, 1.0)
}

// PresencePenalty generates valid presence penalty values
func (g *ParameterGenerator) PresencePenalty() gopter.Gen {
	return gen.Float64Range(-2.0, 2.0)
}

// FrequencyPenalty generates valid frequency penalty values
func (g *ParameterGenerator) FrequencyPenalty() gopter.Gen {
	return gen.Float64Range(-2.0, 2.0)
}

// ValidParameterMap generates a map of valid API parameters
func (g *ParameterGenerator) ValidParameterMap() gopter.Gen {
	// Simplified approach: create generator that produces a map directly
	return gen.UInt8().Map(func(flags interface{}) interface{} {
		flagVal := flags.(uint8)
		params := make(map[string]interface{})

		// Use bit flags to determine which parameters to include
		if flagVal&1 != 0 {
			params["temperature"] = 1.0 // Fixed value for simplicity
		}
		if flagVal&2 != 0 {
			params["max_tokens"] = 1000
		}
		if flagVal&4 != 0 {
			params["top_p"] = 0.9
		}
		if flagVal&8 != 0 {
			params["presence_penalty"] = 0.0
		}
		if flagVal&16 != 0 {
			params["frequency_penalty"] = 0.0
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
func (g *ConfigGenerator) APIKey() gopter.Gen {
	return gen.RegexMatch(`test-[a-zA-Z0-9]{10,50}`)
}

// URL generates valid URLs
func (g *ConfigGenerator) URL() gopter.Gen {
	schemes := []string{"http://", "https://"}
	domains := []string{
		"api.openai.com",
		"generativelanguage.googleapis.com",
		"openrouter.ai",
		"localhost:8080",
		"test-server.local",
	}

	// Convert slices to interface{} slice for OneConstOf
	schemeIfaces := make([]interface{}, len(schemes))
	for i, s := range schemes {
		schemeIfaces[i] = s
	}
	domainIfaces := make([]interface{}, len(domains))
	for i, d := range domains {
		domainIfaces[i] = d
	}

	return gen.OneConstOf(schemeIfaces...).FlatMap(func(scheme interface{}) gopter.Gen {
		return gen.OneConstOf(domainIfaces...).FlatMap(func(domain interface{}) gopter.Gen {
			return gen.RegexMatch(`(/[a-zA-Z0-9_-]*)*`).Map(func(path interface{}) interface{} {
				return scheme.(string) + domain.(string) + path.(string)
			})
		}, reflect.TypeOf(""))
	}, reflect.TypeOf(""))
}

// TextProcessor provides utilities for testing text processing functions
type TextProcessor struct{}

// NewTextProcessor creates a new text processor instance
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{}
}

// NonEmptyText generates non-empty text strings
func (g *TextProcessor) NonEmptyText() gopter.Gen {
	return gen.RegexMatch(`[a-zA-Z0-9\s.,!?]{1,1000}`)
}

// BoundedText generates text within specified length bounds
func (g *TextProcessor) BoundedText(minLen, maxLen int) gopter.Gen {
	// Generate string within bounds using alphanumeric characters
	return gen.SliceOfN(minLen, gen.RuneRange('a', 'z')).Map(func(runes interface{}) interface{} {
		r := runes.([]rune)
		return string(r)
	})
}

// TokenCount generates realistic token counts
func (g *TextProcessor) TokenCount() gopter.Gen {
	return gen.IntRange(0, 100000)
}

// ErrorScenarioGenerator creates generators for error testing scenarios
type ErrorScenarioGenerator struct{}

// NewErrorScenarioGenerator creates a new error scenario generator instance
func NewErrorScenarioGenerator() *ErrorScenarioGenerator {
	return &ErrorScenarioGenerator{}
}

// HTTPStatusCode generates HTTP status codes for error testing
func (g *ErrorScenarioGenerator) HTTPStatusCode() gopter.Gen {
	// Common HTTP status codes for testing
	codes := []int{200, 400, 401, 403, 404, 429, 500, 502, 503, 504}
	// Convert int slice to interface{} slice for OneConstOf
	ifaces := make([]interface{}, len(codes))
	for i, c := range codes {
		ifaces[i] = c
	}
	return gen.OneConstOf(ifaces...)
}

// ErrorMessage generates realistic error messages
func (g *ErrorScenarioGenerator) ErrorMessage() gopter.Gen {
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
	// Convert string slice to interface{} slice for OneConstOf
	ifaces := make([]interface{}, len(messages))
	for i, m := range messages {
		ifaces[i] = m
	}
	return gen.OneConstOf(ifaces...)
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
