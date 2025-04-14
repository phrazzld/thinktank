// internal/llm/client.go
package llm

import (
	"context"
)

// ProviderResult holds the response from a content generation call
type ProviderResult struct {
	Content      string   // The generated content
	FinishReason string   // Why generation stopped, e.g., "stop", "length", "safety"
	TokenCount   int32    // Number of tokens in the response
	Truncated    bool     // Whether the response was truncated
	SafetyInfo   []Safety // Optional safety information
}

// Safety represents content safety evaluation information
type Safety struct {
	Category string  // Safety category name
	Blocked  bool    // Whether content was blocked due to this category
	Score    float32 // Severity score (provider-specific scale)
}

// ProviderTokenCount holds the result of a token counting operation
type ProviderTokenCount struct {
	Total int32 // Total token count
}

// ProviderModelInfo holds model capabilities and limits
type ProviderModelInfo struct {
	Name             string // Model name
	InputTokenLimit  int32  // Maximum input tokens allowed
	OutputTokenLimit int32  // Maximum output tokens allowed
}

// LLMClient defines the interface for interacting with any LLM provider
type LLMClient interface {
	// GenerateContent sends a text prompt to the LLM and returns the generated content
	GenerateContent(ctx context.Context, prompt string) (*ProviderResult, error)

	// CountTokens counts the tokens in a given prompt
	CountTokens(ctx context.Context, prompt string) (*ProviderTokenCount, error)

	// GetModelInfo retrieves information about the current model
	GetModelInfo(ctx context.Context) (*ProviderModelInfo, error)

	// GetModelName returns the name of the model being used
	GetModelName() string

	// Close releases resources used by the client
	Close() error
}
