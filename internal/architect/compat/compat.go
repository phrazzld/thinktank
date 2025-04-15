// Package compat provides backward compatibility shims for deprecated APIs
// within the architect internal package.
//
// These functions and types are intended for temporary use during the transition
// period as the application moves towards provider-agnostic interfaces. They
// allow existing code that relies on the older, Gemini-specific interfaces
// to continue functioning while migration occurs.
//
// Deprecation Timeline:
//   - Phase 1 (Completed): Deprecation notices added to original methods.
//   - Phase 2 (Current): Methods moved into this 'compat' package.
//   - Phase 3 (Target: v0.8.0 / ~Q4 2024): Removal of this package and all
//     associated functions and types.
//
// Users of these compatibility functions should migrate to the primary,
// provider-agnostic methods (e.g., InitLLMClient, ProcessLLMResponse) before
// the removal date/version to avoid breaking changes.
package compat

import (
	"context"
	"errors"
	"fmt"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// InitLLMClientFunc defines the signature for a function that initializes an LLM client.
// This is used to inject the core client initialization logic from the main apiService
// into the compatibility layer without creating tight coupling or circular dependencies.
type InitLLMClientFunc func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)

// ProcessLLMResponseFunc defines the signature for a function that processes an LLM result.
// This allows injecting the core response processing logic into the compatibility layer.
type ProcessLLMResponseFunc func(result *llm.ProviderResult) (string, error)

// InitClient initializes and returns a Gemini client (for backward compatibility).
//
// Deprecated: This function is part of the compatibility layer scheduled for removal
// by v0.8.0 / ~Q4 2024. Use the provider-agnostic `InitLLMClient` function from the
// `internal/architect` package instead. See package documentation for the deprecation timeline.
func InitClient(
	ctx context.Context,
	apiKey, modelName, apiEndpoint string,
	initLLMClient InitLLMClientFunc,
	logger logutil.LoggerInterface,
) (gemini.Client, error) {
	// For backward compatibility, strictly enforce that only Gemini models are used here.
	// Simple check for Gemini models (duplicated from DetectProviderFromModel to avoid import cycle)
	isGemini := len(modelName) >= 6 && modelName[:6] == "gemini"
	if !isGemini {
		// Create a local error instead of using architect.ErrUnsupportedModel
		errUnsupportedModel := errors.New("unsupported model type")
		return nil, fmt.Errorf("%w: InitClient only supports Gemini models, use InitLLMClient instead", errUnsupportedModel)
	}

	// Use the injected function to create the underlying LLM client.
	// This delegates the actual client creation (and associated error handling like API key checks)
	// to the main, up-to-date implementation.
	llmClient, err := initLLMClient(ctx, apiKey, modelName, apiEndpoint)
	if err != nil {
		// The error returned by initLLMClient should already be appropriately wrapped.
		return nil, err
	}

	// Wrap the provider-agnostic LLMClient with the adapter to satisfy the gemini.Client interface.
	return NewLLMToGeminiClientAdapter(llmClient, logger), nil
}

// ProcessResponse processes the Gemini API response and extracts content.
//
// Deprecated: This function is part of the compatibility layer scheduled for removal
// by v0.8.0 / ~Q4 2024. Use the provider-agnostic `ProcessLLMResponse` function from the
// `internal/architect` package instead. See package documentation for the deprecation timeline.
func ProcessResponse(
	result *gemini.GenerationResult,
	processLLMResponse ProcessLLMResponseFunc,
) (string, error) {
	// Handle nil input gracefully.
	if result == nil {
		errEmptyResponse := errors.New("received empty response from LLM")
		return "", fmt.Errorf("%w: result is nil", errEmptyResponse)
	}

	// Convert safety ratings
	var safetyInfo []llm.Safety
	if result.SafetyRatings != nil {
		safetyInfo = make([]llm.Safety, len(result.SafetyRatings))
		for i, rating := range result.SafetyRatings {
			safetyInfo[i] = llm.Safety{
				Category: rating.Category,
				Blocked:  rating.Blocked,
				Score:    rating.Score,
			}
		}
	}

	// Create provider-agnostic result
	providerResult := &llm.ProviderResult{
		Content:      result.Content,
		FinishReason: result.FinishReason,
		TokenCount:   result.TokenCount,
		Truncated:    result.Truncated,
		SafetyInfo:   safetyInfo,
	}

	// Use the injected provider-agnostic function for processing
	return processLLMResponse(providerResult)
}

// llmToGeminiClientAdapter adapts an LLMClient to a gemini.Client for backward compatibility.
//
// Deprecated: This adapter is part of the compatibility layer scheduled for removal
// by v0.8.0 / ~Q4 2024. Code should be updated to use the llm.LLMClient interface
// directly. See package documentation for the deprecation timeline.
type llmToGeminiClientAdapter struct {
	llmClient llm.LLMClient
	logger    logutil.LoggerInterface
}

// NewLLMToGeminiClientAdapter creates a new adapter to convert from LLMClient to gemini.Client.
// This is exported to allow proper testing of the adapter.
func NewLLMToGeminiClientAdapter(client llm.LLMClient, logger logutil.LoggerInterface) gemini.Client {
	return &llmToGeminiClientAdapter{
		llmClient: client,
		logger:    logger,
	}
}

// GenerateContent calls the underlying LLMClient and converts the result
func (a *llmToGeminiClientAdapter) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	// Call the LLMClient implementation
	result, err := a.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Convert to Gemini format
	return &gemini.GenerationResult{
		Content:      result.Content,
		FinishReason: result.FinishReason,
		TokenCount:   result.TokenCount,
		Truncated:    result.Truncated,
		// We don't convert safety info back since it's not typically used by consumers
	}, nil
}

// CountTokens delegates to the LLMClient
func (a *llmToGeminiClientAdapter) CountTokens(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
	result, err := a.llmClient.CountTokens(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &gemini.TokenCount{
		Total: result.Total,
	}, nil
}

// GetModelInfo delegates to the LLMClient
func (a *llmToGeminiClientAdapter) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	result, err := a.llmClient.GetModelInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &gemini.ModelInfo{
		Name:             result.Name,
		InputTokenLimit:  result.InputTokenLimit,
		OutputTokenLimit: result.OutputTokenLimit,
	}, nil
}

// GetModelName returns the name of the model
func (a *llmToGeminiClientAdapter) GetModelName() string {
	return a.llmClient.GetModelName()
}

// GetTemperature returns a default temperature (not used in tests)
func (a *llmToGeminiClientAdapter) GetTemperature() float32 {
	a.logger.Debug("compat.llmToGeminiClientAdapter.GetTemperature() called - returning default 0.7")
	return 0.7 // Default value
}

// GetMaxOutputTokens returns a default max output tokens (not used in tests)
func (a *llmToGeminiClientAdapter) GetMaxOutputTokens() int32 {
	a.logger.Debug("compat.llmToGeminiClientAdapter.GetMaxOutputTokens() called - returning default 1024")
	return 1024 // Default value
}

// GetTopP returns a default topP (not used in tests)
func (a *llmToGeminiClientAdapter) GetTopP() float32 {
	a.logger.Debug("compat.llmToGeminiClientAdapter.GetTopP() called - returning default 0.95")
	return 0.95 // Default value
}

// Close delegates to the LLMClient
func (a *llmToGeminiClientAdapter) Close() error {
	return a.llmClient.Close()
}
