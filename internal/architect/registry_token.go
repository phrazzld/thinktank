// Package architect contains the core application logic for the architect tool
package architect

import (
	"context"
	"fmt"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
)

// registryTokenManager extends the base token manager with registry awareness
type registryTokenManager struct {
	*tokenManager // Embed the base token manager
	registry      *registry.Registry
}

// NewRegistryTokenManager creates a new token manager that uses registry information
func NewRegistryTokenManager(
	logger logutil.LoggerInterface,
	auditLogger auditlog.AuditLogger,
	client llm.LLMClient,
	reg *registry.Registry,
) (TokenManager, error) {
	// Create the base token manager
	baseTM, err := NewTokenManager(logger, auditLogger, client)
	if err != nil {
		return nil, err
	}

	// Cast to access the embedded instance
	baseImpl, ok := baseTM.(*tokenManager)
	if !ok {
		return nil, fmt.Errorf("unexpected token manager implementation type")
	}

	// Return the registry-aware version
	return &registryTokenManager{
		tokenManager: baseImpl,
		registry:     reg,
	}, nil
}

// CheckTokenLimit is overridden to use registry information when available
func (rtm *registryTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	// Get the model name from the client
	modelName := rtm.client.GetModelName()

	// First, try to get model info from the registry
	modelDef, err := rtm.registry.GetModel(modelName)
	if err == nil && modelDef != nil {
		// We found the model in the registry, use its context window
		rtm.logger.Debug("Using token limits from registry for model %s: %d tokens",
			modelName, modelDef.ContextWindow)

		// Count tokens in the prompt
		tokenResult, err := rtm.client.CountTokens(ctx, prompt)
		if err != nil {
			return fmt.Errorf("failed to count tokens for token limit check: %w", err)
		}

		// Check if exceeds limit
		if tokenResult.Total > modelDef.ContextWindow {
			return fmt.Errorf("prompt exceeds token limit (%d tokens > %d token limit)",
				tokenResult.Total, modelDef.ContextWindow)
		}

		// Log token usage information
		percentage := float64(tokenResult.Total) / float64(modelDef.ContextWindow) * 100
		rtm.logger.Debug("Token usage (registry): %d / %d (%.1f%%)",
			tokenResult.Total, modelDef.ContextWindow, percentage)

		return nil
	}

	// If not in registry, fall back to the base implementation
	rtm.logger.Debug("Model %s not found in registry, falling back to client-provided token limits", modelName)
	return rtm.tokenManager.CheckTokenLimit(ctx, prompt)
}

// GetTokenInfo is overridden to use registry information when available
func (rtm *registryTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	// Get the model name from the client
	modelName := rtm.client.GetModelName()

	// First, try to get model info from the registry
	modelDef, err := rtm.registry.GetModel(modelName)
	if err == nil && modelDef != nil {
		// We found the model in the registry
		rtm.logger.Debug("Using token limits from registry for model %s: %d tokens",
			modelName, modelDef.ContextWindow)

		// Count tokens in the prompt
		tokenResult, err := rtm.client.CountTokens(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to count tokens for token limit check: %w", err)
		}

		// Create result structure
		result := &TokenResult{
			TokenCount:   tokenResult.Total,
			InputLimit:   modelDef.ContextWindow,
			ExceedsLimit: false,
		}

		// Calculate percentage of limit
		result.Percentage = float64(result.TokenCount) / float64(result.InputLimit) * 100

		// Log token usage information
		rtm.logger.Debug("Token usage (registry): %d / %d (%.1f%%)",
			result.TokenCount, result.InputLimit, result.Percentage)

		// Check if the prompt exceeds the token limit
		if result.TokenCount > result.InputLimit {
			result.ExceedsLimit = true
			result.LimitError = fmt.Sprintf("prompt exceeds token limit (%d tokens > %d token limit)",
				result.TokenCount, result.InputLimit)
		}

		return result, nil
	}

	// If not in registry, fall back to the base implementation
	rtm.logger.Debug("Model %s not found in registry, falling back to client-provided token limits", modelName)
	return rtm.tokenManager.GetTokenInfo(ctx, prompt)
}

// PromptForConfirmation is directly inherited from the base tokenManager
// No need to override it as it doesn't depend on token limit sources
