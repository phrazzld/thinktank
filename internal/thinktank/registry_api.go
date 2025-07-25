// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/providers"
	openrouterprovider "github.com/phrazzld/thinktank/internal/providers/openrouter"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// registryAPIService implements the APIService interface using the models package
type registryAPIService struct {
	logger logutil.LoggerInterface
}

// NewRegistryAPIService creates a new models-based API service
// This implementation uses the models package for model and provider information,
// providing a simplified approach with hardcoded model definitions.
func NewRegistryAPIService(logger logutil.LoggerInterface) interfaces.APIService {
	return &registryAPIService{
		logger: logger,
	}
}

// InitLLMClient initializes and returns an LLM client based on the model name
// using the models package to look up model and provider information
func (s *registryAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// Validate required parameters
	if modelName == "" {
		return nil, fmt.Errorf("%w: model name is required", llm.ErrClientInitialization)
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrClientInitialization, ctx.Err())
	}

	// Log custom endpoint if provided (used for API endpoint override)
	if apiEndpoint != "" {
		s.logger.DebugContext(ctx, "Using custom API endpoint: %s", apiEndpoint)
	}

	// Special case for testing with error-model
	if modelName == "error-model" {
		return nil, errors.New("test model error")
	}

	// Look up the model in the models package
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		s.logger.DebugContext(ctx, "Model '%s' not found: %v", modelName, err)
		return nil, llm.Wrap(err, "", fmt.Sprintf("model '%s' not supported", modelName), llm.CategoryInvalidRequest)
	}

	// Get provider name from model info
	providerName := modelInfo.Provider
	s.logger.DebugContext(ctx, "Model '%s' uses provider '%s'", modelName, providerName)

	// Determine which API endpoint to use
	effectiveEndpoint := apiEndpoint
	if effectiveEndpoint == "" {
		// Set provider-specific base URLs
		switch providerName {
		case "openrouter":
			effectiveEndpoint = "https://openrouter.ai/api/v1"
			// OpenAI and Gemini use their default endpoints, no need to set explicitly
		}
	}

	// Get the provider implementation directly
	var providerImpl providers.Provider
	switch providerName {
	case "openrouter":
		providerImpl = openrouterprovider.NewProvider(s.logger)
	default:
		return nil, llm.Wrap(
			fmt.Errorf("unsupported provider: %s", providerName),
			"",
			fmt.Sprintf("provider '%s' is not supported - only OpenRouter is supported after consolidation", providerName),
			llm.CategoryInvalidRequest,
		)
	}

	// API Key Resolution Logic
	// ------------------------
	// The system follows this precedence order for API keys:
	// 1. Environment variables specific to each provider (highest priority)
	//    - For OpenRouter: OPENROUTER_API_KEY (only supported provider)
	// 2. Explicitly provided API key parameter (fallback only)
	//
	// After provider consolidation, only OpenRouter is supported. All models
	// (previously OpenAI and Gemini) now route through OpenRouter API.

	// Start with an empty key, which we'll populate from environment or passed parameter
	effectiveApiKey := ""

	// STEP 1: First try to get the key from environment variable based on provider
	// This is the recommended and preferred method for providing API keys
	envVar := models.GetAPIKeyEnvVar(providerName)
	if envVar != "" {
		envApiKey := os.Getenv(envVar) // TODO: Replace with cached lookup for startup performance
		if envApiKey != "" {
			effectiveApiKey = envApiKey
			s.logger.DebugContext(ctx, "Using API key from environment variable %s for provider '%s'",
				envVar, providerName)
		}
	}

	// STEP 2: Only fall back to the passed apiKey if environment variable is not set
	// This is discouraged for production use but supported for testing/development
	if effectiveApiKey == "" && apiKey != "" {
		effectiveApiKey = apiKey
		s.logger.DebugContext(ctx, "Environment variable not set or empty, using provided API key for provider '%s'",
			providerName)
	}

	// STEP 3: If no API key is available from either source, reject the request
	// API keys are required for all providers
	if effectiveApiKey == "" {
		envVarName := models.GetAPIKeyEnvVar(providerName)
		return nil, fmt.Errorf("%w: API key is required for model '%s' with provider '%s'. Please set the %s environment variable",
			llm.ErrClientInitialization, modelName, providerName, envVarName)
	}

	// Create the client using the provider implementation
	s.logger.DebugContext(ctx, "Creating LLM client for model '%s' using provider '%s'",
		modelName, providerName)

	// Verify the API key is non-empty before passing it to the provider
	if effectiveApiKey == "" {
		s.logger.ErrorContext(ctx, "Empty API key for provider '%s' - this will cause authentication failures", providerName)
	} else {
		// Log API key metadata only (NEVER log any portion of the key itself)
		s.logger.DebugContext(ctx, "Using API key for provider '%s' (length: %d, source: via environment variable)",
			providerName, len(effectiveApiKey))
	}

	// Since we're now using the providers.Provider type directly, we no longer need to do
	// a type assertion to get the CreateClient method - it's already part of the interface
	client, err := providerImpl.CreateClient(ctx, effectiveApiKey, modelInfo.APIModelID, effectiveEndpoint)
	if err != nil {
		// Check for OpenRouter errors (only supported provider after consolidation)
		if apiErr, ok := openrouterprovider.IsOpenRouterError(err); ok {
			return nil, fmt.Errorf("%w: %s", llm.ErrClientInitialization, apiErr.UserFacingError())
		}

		// Wrap the original error for any non-OpenRouter provider errors
		return nil, fmt.Errorf("%w: %v", llm.ErrClientInitialization, err)
	}

	return client, nil
}

// The remaining methods are carried over from the existing APIService implementation
// since they don't depend on the provider initialization logic

// GetModelParameters retrieves parameter values for a given model
// It returns a map of parameter name to parameter value, applying defaults from the model definition
func (s *registryAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	// Look up the model in the models package
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		s.logger.DebugContext(ctx, "Model '%s' not found: %v", modelName, err)
		return make(map[string]interface{}), llm.Wrap(err, "", fmt.Sprintf("model '%s' not supported", modelName), llm.CategoryInvalidRequest)
	}

	// Return the default parameters from the model info
	if modelInfo.DefaultParams == nil {
		return make(map[string]interface{}), nil
	}

	// Create a copy of the default parameters map to avoid modifying the original
	params := make(map[string]interface{})
	for key, value := range modelInfo.DefaultParams {
		params[key] = value
	}

	return params, nil
}

// ValidateModelParameter validates a parameter value against its constraints
// It returns true if the parameter is valid, false otherwise
func (s *registryAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	// First, verify the model exists
	_, err := models.GetModelInfo(modelName)
	if err != nil {
		s.logger.DebugContext(ctx, "Model '%s' not found: %v", modelName, err)
		return false, llm.Wrap(err, "", fmt.Sprintf("model '%s' not supported", modelName), llm.CategoryInvalidRequest)
	}

	// Use the comprehensive parameter validation from models package
	if err := models.ValidateParameter(modelName, paramName, value); err != nil {
		s.logger.DebugContext(ctx, "Parameter validation failed for model '%s', parameter '%s': %v", modelName, paramName, err)
		return false, fmt.Errorf("parameter validation failed: %w", err)
	}

	return true, nil
}

// GetModelDefinition retrieves the full model definition
func (s *registryAPIService) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	// Look up the model in the models package
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		s.logger.DebugContext(ctx, "Model '%s' not found: %v", modelName, err)
		return nil, llm.Wrap(err, "", fmt.Sprintf("model '%s' not supported", modelName), llm.CategoryInvalidRequest)
	}

	return &modelInfo, nil
}

// GetModelTokenLimits retrieves token limits from the models package for a given model
// This method now returns the actual model token limits when available,
// and very high default values when not specified (to prevent truncation issues)
func (s *registryAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Look up the model in the models package
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		s.logger.DebugContext(ctx, "Model '%s' not found: %v", modelName, err)
		return 0, 0, llm.Wrap(err, "", fmt.Sprintf("model '%s' not supported", modelName), llm.CategoryInvalidRequest)
	}

	// Use values from model definition
	contextWindow = int32(modelInfo.ContextWindow)
	maxOutputTokens = int32(modelInfo.MaxOutputTokens)

	s.logger.DebugContext(ctx, "Using token limits for model '%s': context window=%d, max output tokens=%d",
		modelName, contextWindow, maxOutputTokens)

	return contextWindow, maxOutputTokens, nil
}

// ProcessLLMResponse processes a provider-agnostic API response and extracts content
func (s *registryAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	// Check for nil result
	if result == nil {
		return "", fmt.Errorf("%w: result is nil", llm.ErrEmptyResponse)
	}

	// Check for empty content
	if result.Content == "" {
		var errDetails strings.Builder

		// Add finish reason if available
		if result.FinishReason != "" {
			fmt.Fprintf(&errDetails, " (Finish Reason: %s)", result.FinishReason)
		}

		// Check for safety blocks
		if len(result.SafetyInfo) > 0 {
			blocked := false
			safetyInfo := ""
			for _, safety := range result.SafetyInfo {
				if safety.Blocked {
					blocked = true
					safetyInfo += fmt.Sprintf(" Blocked by Safety Category: %s;", safety.Category)
				}
			}

			if blocked {
				if errDetails.Len() > 0 {
					errDetails.WriteString(" ")
				}
				errDetails.WriteString("Safety Blocking:")
				errDetails.WriteString(safetyInfo)

				// If we have safety blocks, use the specific safety error
				return "", fmt.Errorf("%w%s", llm.ErrSafetyBlocked, errDetails.String())
			}
		}

		// If we don't have safety blocks, use the generic empty response error
		return "", fmt.Errorf("%w%s", llm.ErrEmptyResponse, errDetails.String())
	}

	// Check for whitespace-only content
	if strings.TrimSpace(result.Content) == "" {
		return "", llm.ErrWhitespaceContent
	}

	return result.Content, nil
}

// IsEmptyResponseError checks if an error is related to empty API responses
func (s *registryAPIService) IsEmptyResponseError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types using errors.Is
	if errors.Is(err, llm.ErrEmptyResponse) || errors.Is(err, llm.ErrWhitespaceContent) {
		return true
	}

	// Convert the error message to lowercase for case-insensitive matching
	errMsg := strings.ToLower(err.Error())

	// Check for common empty response phrases
	if strings.Contains(errMsg, "empty response") ||
		strings.Contains(errMsg, "empty content") ||
		strings.Contains(errMsg, "empty output") ||
		strings.Contains(errMsg, "empty result") {
		return true
	}

	// Check for provider-specific empty response patterns
	if strings.Contains(errMsg, "zero candidates") ||
		strings.Contains(errMsg, "empty candidates") ||
		strings.Contains(errMsg, "no output") {
		return true
	}

	return false
}

// IsSafetyBlockedError checks if an error is related to safety filters
func (s *registryAPIService) IsSafetyBlockedError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types using errors.Is
	if errors.Is(err, llm.ErrSafetyBlocked) {
		return true
	}

	// Convert to lowercase for case-insensitive matching
	errMsg := strings.ToLower(err.Error())

	// Check for common safety-related phrases
	if strings.Contains(errMsg, "safety") ||
		strings.Contains(errMsg, "content policy") ||
		strings.Contains(errMsg, "content filter") ||
		strings.Contains(errMsg, "content_filter") {
		return true
	}

	// Check for provider-specific moderation terminology
	if strings.Contains(errMsg, "moderation") ||
		strings.Contains(errMsg, "blocked") ||
		strings.Contains(errMsg, "filtered") ||
		strings.Contains(errMsg, "harm_category") {
		return true
	}

	return false
}

// GetErrorDetails extracts detailed information from an error
func (s *registryAPIService) GetErrorDetails(err error) string {
	// Handle nil error case
	if err == nil {
		return "no error"
	}

	// Check for OpenRouter errors (only supported provider after consolidation)
	if apiErr, ok := openrouterprovider.IsOpenRouterError(err); ok {
		return apiErr.UserFacingError()
	}

	// Return the error string for other error types
	return err.Error()
}
