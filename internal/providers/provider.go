// Package providers contains common interfaces and utilities for LLM providers.
package providers

import (
	"context"

	"github.com/phrazzld/thinktank/internal/llm"
)

// Provider defines the interface for LLM provider implementations.
// Each provider is responsible for creating a client that implements the llm.LLMClient interface.
type Provider interface {
	// CreateClient creates a new LLM client for a specific model.
	// Parameters:
	//   - ctx: The context for the operation
	//   - apiKey: Authentication key for the provider's API
	//   - modelID: The ID of the model to use with this provider
	//   - apiEndpoint: Optional custom API endpoint (if empty, provider's default is used)
	// Returns:
	//   - An LLM client that can generate content, count tokens, and provide model info
	//   - An error if client creation fails
	CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error)
}
