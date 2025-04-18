// internal/gemini/client.go
package gemini

import (
	"context"

	"github.com/phrazzld/architect/internal/llm"
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
// This interface is maintained for backward compatibility
// New code should use the provider-agnostic llm.LLMClient interface instead
type Client interface {
	// GenerateContent sends a text prompt to Gemini and returns the generated content
	// If params is provided, these parameters will override the default model parameters
	GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*GenerationResult, error)

	// CountTokens counts the tokens in a given prompt
	CountTokens(ctx context.Context, prompt string) (*TokenCount, error)

	// GetModelInfo retrieves information about a model
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
type ClientOption func(interface{})

// WithHTTPClient allows injecting a custom HTTP client
func WithHTTPClient(client HTTPClient) ClientOption {
	return func(c interface{}) {
		if gc, ok := c.(*geminiClient); ok {
			gc.httpClient = client
		}
	}
}

// WithLogger allows injecting a custom logger
func WithLogger(logger logutil.LoggerInterface) ClientOption {
	return func(c interface{}) {
		if gc, ok := c.(*geminiClient); ok {
			gc.logger = logger
		}
	}
}

// NewClient creates a new Gemini client implementing the legacy Client interface
// For new code, consider using NewLLMClient which returns the provider-agnostic llm.LLMClient directly
func NewClient(ctx context.Context, apiKey, modelName, apiEndpoint string, opts ...ClientOption) (Client, error) {
	// Create internalOpts that wrap the public options
	var internalOpts []geminiClientOption
	for _, opt := range opts {
		// Create a wrapper that calls the public option with the geminiClient
		internalOpts = append(internalOpts, func(gc *geminiClient) {
			opt(gc)
		})
	}

	// Create the provider-agnostic LLM client
	llmClient, err := newGeminiClient(ctx, apiKey, modelName, apiEndpoint, internalOpts...)
	if err != nil {
		return nil, err
	}

	// Wrap it in a ClientAdapter for backward compatibility
	return NewClientAdapter(llmClient), nil
}

// NewLLMClient creates a new Gemini client implementing the provider-agnostic llm.LLMClient interface
// This is the preferred method for new code
func NewLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string, opts ...ClientOption) (llm.LLMClient, error) {
	// Create internalOpts that wrap the public options
	var internalOpts []geminiClientOption
	for _, opt := range opts {
		// Create a wrapper that calls the public option with the geminiClient
		internalOpts = append(internalOpts, func(gc *geminiClient) {
			opt(gc)
		})
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
