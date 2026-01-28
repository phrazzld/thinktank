// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/providers"
)

// Helper function to get minimum of two integers
// Reason: Maintained for upcoming stream token management functionality in OpenRouter client
func min(a, b int) int { //nolint:unused
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

	// Initialize API key from provided argument or environment variable
	effectiveAPIKey := apiKey
	if effectiveAPIKey != "" {
		p.logger.Debug("Using provided API key (length: %d)", len(effectiveAPIKey))
	} else {
		// Fall back to the OPENROUTER_API_KEY environment variable
		effectiveAPIKey = os.Getenv("OPENROUTER_API_KEY")
		p.logger.Debug("OPENROUTER_API_KEY environment variable length: %d", len(effectiveAPIKey))
		if effectiveAPIKey == "" {
			return nil, fmt.Errorf("no OpenRouter API key provided and OPENROUTER_API_KEY environment variable not set")
		}
		p.logger.Debug("Using API key from OPENROUTER_API_KEY environment variable")
	}

	// Validate the API key format
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
