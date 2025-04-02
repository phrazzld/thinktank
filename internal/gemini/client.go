// internal/gemini/client.go
package gemini

import (
	"context"
)

// ModelConfig holds Gemini model configuration parameters
type ModelConfig struct {
	MaxOutputTokens int32
	Temperature     float32
	TopP            float32
}

// ModelInfo holds model capabilities and limits
type ModelInfo struct {
	Name             string
	InputTokenLimit  int32
	OutputTokenLimit int32
}

// SafetyRating represents a content safety evaluation
type SafetyRating struct {
	Category string
	Blocked  bool
	Score    float32
}

// GenerationResult holds the response from a generate content call
type GenerationResult struct {
	Content       string
	FinishReason  string
	SafetyRatings []SafetyRating
	TokenCount    int32
	Truncated     bool
}

// TokenCount holds the result of a token counting operation
type TokenCount struct {
	Total int32
}

// Client defines the interface for interacting with Gemini API
type Client interface {
	// GenerateContent sends a text prompt to Gemini and returns the generated content
	GenerateContent(ctx context.Context, prompt string) (*GenerationResult, error)

	// CountTokens counts the tokens in a given prompt
	CountTokens(ctx context.Context, prompt string) (*TokenCount, error)

	// GetModelInfo retrieves information about a model (for future implementation)
	GetModelInfo(ctx context.Context) (*ModelInfo, error)

	// Close releases resources used by the client
	Close() error
}

// NewClient creates a new Gemini client with the given API key and model name
func NewClient(ctx context.Context, apiKey, modelName string) (Client, error) {
	return newGeminiClient(ctx, apiKey, modelName)
}

// DefaultModelConfig returns a reasonable default model configuration
func DefaultModelConfig() ModelConfig {
	return ModelConfig{
		MaxOutputTokens: 8192, // High limit for plan generation
		Temperature:     0.3,  // Lower temperature for deterministic output
		TopP:            0.9,  // Allow some creativity
	}
}
