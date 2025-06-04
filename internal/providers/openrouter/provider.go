// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"context"
	"fmt"
	"strings"

	"github.com/phrazzld/thinktank/internal/apikey"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers"
)

// Helper function to get minimum of two integers
// Reason: Maintained for upcoming stream token management functionality in OpenRouter client
func minInt(a, b int) int { //nolint:unused
	if a < b {
		return a
	}
	return b
}

// OpenRouterProvider implements the Provider interface for OpenRouter models.
type OpenRouterProvider struct {
	logger logutil.LoggerInterface
}

// NewProvider creates a new instance of OpenRouterProvider.
func NewProvider(logger logutil.LoggerInterface) providers.Provider {
	// If no logger provided, create a default one
	if logger == nil {
		logger = logutil.NewLogger(logutil.InfoLevel, nil, "[openrouter-provider] ")
	}

	return &OpenRouterProvider{
		logger: logger,
	}
}

// CreateClient implements the Provider interface.
func (p *OpenRouterProvider) CreateClient(
	ctx context.Context,
	apiKey string,
	modelID string,
	apiEndpoint string,
) (llm.LLMClient, error) {
	p.logger.Debug("Creating OpenRouter client for model: %s", modelID)

	// Use centralized API key resolver
	keyResolver := apikey.NewAPIKeyResolver(p.logger)
	keyResult, err := keyResolver.ResolveAPIKey(ctx, "openrouter", apiKey)
	if err != nil {
		return nil, err
	}

	// Validate the resolved API key with basic validation
	if err := keyResolver.ValidateAPIKey(ctx, "openrouter", keyResult.Key); err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	effectiveAPIKey := keyResult.Key

	// Additional OpenRouter-specific validation
	if !strings.HasPrefix(effectiveAPIKey, "sk-or") {
		p.logger.Warn("OpenRouter API key does not have the expected 'sk-or' prefix. This will cause authentication failures.")
		return nil, fmt.Errorf("invalid OpenRouter API key format: key should start with 'sk-or'. Please check your API key and ensure you're using an OpenRouter key, not another provider's key")
	}

	// Set default API endpoint if none provided
	effectiveAPIEndpoint := apiEndpoint
	if effectiveAPIEndpoint == "" {
		effectiveAPIEndpoint = "https://openrouter.ai/api/v1"
		p.logger.Debug("Using default OpenRouter API endpoint")
	} else {
		p.logger.Debug("Using %s", GetBaseURLLogInfo(effectiveAPIEndpoint))
	}

	// Validate modelID
	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}

	// Verify the modelID has the expected format (provider/model or provider/organization/model)
	if !strings.Contains(modelID, "/") {
		p.logger.Warn("OpenRouter model ID '%s' does not have expected format 'provider/model' or 'provider/organization/model'", modelID)
	}

	// Create the client using our OpenRouter implementation
	client, err := NewClient(effectiveAPIKey, modelID, effectiveAPIEndpoint, p.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter client: %w", err)
	}

	return client, nil
}
