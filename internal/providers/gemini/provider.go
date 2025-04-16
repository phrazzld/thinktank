// internal/providers/gemini/provider.go
package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/providers"
)

// GeminiProvider implements the Provider interface for Google's Gemini models.
type GeminiProvider struct {
	logger logutil.LoggerInterface
}

// NewProvider creates a new instance of GeminiProvider.
func NewProvider(logger logutil.LoggerInterface) providers.Provider {
	// If no logger provided, create a default one
	if logger == nil {
		logger = logutil.NewLogger(logutil.InfoLevel, nil, "[gemini-provider] ")
	}

	return &GeminiProvider{
		logger: logger,
	}
}

// CreateClient implements the Provider interface.
func (p *GeminiProvider) CreateClient(
	ctx context.Context,
	apiKey string,
	modelID string,
	apiEndpoint string,
) (llm.LLMClient, error) {
	p.logger.Debug("Creating Gemini client for model: %s", modelID)

	// Create a list of client options
	var clientOpts []gemini.ClientOption

	// Add custom logger to client options
	clientOpts = append(clientOpts, gemini.WithLogger(p.logger))

	// Create the LLM client using the existing Gemini implementation
	client, err := gemini.NewLLMClient(ctx, apiKey, modelID, apiEndpoint, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Wrap the client in our parameter-aware adapter
	return NewGeminiClientAdapter(client), nil
}

// GeminiClientAdapter wraps the standard Gemini client to provide
// parameter-aware capabilities
type GeminiClientAdapter struct {
	client llm.LLMClient
	params map[string]interface{}
}

// NewGeminiClientAdapter creates a new adapter for the Gemini client
func NewGeminiClientAdapter(client llm.LLMClient) *GeminiClientAdapter {
	return &GeminiClientAdapter{
		client: client,
		params: make(map[string]interface{}),
	}
}

// SetParameters sets the parameters to use for API calls
func (a *GeminiClientAdapter) SetParameters(params map[string]interface{}) {
	a.params = params
}

// GenerateContent implements the llm.LLMClient interface and applies parameters
func (a *GeminiClientAdapter) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	// Apply parameters to underlying gemini client (when we evolve the interface)
	// For now, we just pass through to the existing client

	// TODO: When llm.LLMClient interface evolves to support parameters directly,
	// we'll add parameter handling here.

	return a.client.GenerateContent(ctx, prompt)
}

// CountTokens implements the llm.LLMClient interface
func (a *GeminiClientAdapter) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return a.client.CountTokens(ctx, prompt)
}

// GetModelInfo implements the llm.LLMClient interface and uses configuration data
func (a *GeminiClientAdapter) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	// First try to get the model info from the underlying client
	modelInfo, err := a.client.GetModelInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get the model information from registry if needed
	// Registry integration is now complete, but we'll keep the underlying
	// client call as a fallback for when registry data is unavailable
	if modelInfo.InputTokenLimit == 0 {
		// This indicates we should try using registry data, as zero is an invalid limit
		// The TokenManager should automatically use registry data via registry_token.go
		// when available, so we're covered for client usage inside TokenManager.
		modelName := a.client.GetModelName()

		// For now we use placeholder values that should be overridden by
		// the registry-aware token manager if that's being used
		if strings.HasPrefix(modelName, "gemini-1.5-") {
			modelInfo.InputTokenLimit = 1000000 // 1M tokens for Gemini 1.5 models
		} else if strings.HasPrefix(modelName, "gemini-1.0-") {
			modelInfo.InputTokenLimit = 32768 // 32K for Gemini 1.0 models
		} else {
			// Default fallback for unknown models
			modelInfo.InputTokenLimit = 32768
		}
	}

	return modelInfo, nil
}

// GetModelName implements the llm.LLMClient interface
func (a *GeminiClientAdapter) GetModelName() string {
	return a.client.GetModelName()
}

// Close implements the llm.LLMClient interface
func (a *GeminiClientAdapter) Close() error {
	return a.client.Close()
}
