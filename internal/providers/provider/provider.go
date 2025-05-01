// Package provider defines common interfaces, types, and utilities for LLM providers
package provider

import (
	"context"

	"github.com/phrazzld/thinktank/internal/llm"
)

// Provider defines the interface for LLM provider implementations.
// Each provider (e.g. OpenAI, Gemini) will implement this interface.
type Provider interface {
	// CreateClient creates a new LLM client for a specific model.
	// Parameters:
	//   - ctx: Context for the client creation (e.g. for cancellation)
	//   - apiKey: The API key for the provider
	//   - modelID: The ID of the model to use (as known by the provider)
	//   - apiEndpoint: Optional custom API endpoint (if empty, default is used)
	// Returns:
	//   - An initialized LLM client
	//   - An error if client creation fails
	CreateClient(ctx context.Context, apiKey string, modelID string, apiEndpoint string) (llm.LLMClient, error)
}
