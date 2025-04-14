// internal/gemini/client.go
package gemini

import (
	"context"

	"github.com/phrazzld/architect/internal/logutil"
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

	// GetModelName returns the name of the model being used
	GetModelName() string

	// GetTemperature returns the temperature setting for the model
	GetTemperature() float32

	// GetMaxOutputTokens returns the max output tokens setting for the model
	GetMaxOutputTokens() int32

	// GetTopP returns the topP setting for the model
	GetTopP() float32

	// Close releases resources used by the client
	Close() error
}

// ClientOption defines a function that modifies a client
type ClientOption func(*geminiClient)

// WithHTTPClient allows injecting a custom HTTP client
func WithHTTPClient(client HTTPClient) ClientOption {
	return func(gc *geminiClient) {
		gc.httpClient = client
	}
}

// WithLogger allows injecting a custom logger
func WithLogger(logger logutil.LoggerInterface) ClientOption {
	return func(gc *geminiClient) {
		gc.logger = logger
	}
}

// NewClient creates a new Gemini client with the given API key, model name, and optional API endpoint
func NewClient(ctx context.Context, apiKey, modelName, apiEndpoint string, opts ...ClientOption) (Client, error) {
	// Convert public options to internal options
	var internalOpts []geminiClientOption
	for _, opt := range opts {
		// Conversion is safe as both function types have the same signature
		internalOpts = append(internalOpts, geminiClientOption(opt))
	}

	return newGeminiClient(ctx, apiKey, modelName, apiEndpoint, internalOpts...)
}

// DefaultModelConfig returns a reasonable default model configuration
func DefaultModelConfig() ModelConfig {
	return ModelConfig{
		MaxOutputTokens: 8192, // High limit for plan generation
		Temperature:     0.3,  // Lower temperature for deterministic output
		TopP:            0.9,  // Allow some creativity
	}
}
