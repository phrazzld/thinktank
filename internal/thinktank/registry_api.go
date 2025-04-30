// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/phrazzld/thinktank/internal/gemini"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/openai"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// Define package-level error types for better error handling
var (
	// ErrEmptyResponse indicates the API returned an empty response
	ErrEmptyResponse = errors.New("received empty response from LLM")

	// ErrWhitespaceContent indicates the API returned only whitespace content
	ErrWhitespaceContent = errors.New("LLM returned an empty output text")

	// ErrSafetyBlocked indicates content was blocked by safety filters
	ErrSafetyBlocked = errors.New("content blocked by LLM safety filters")

	// ErrAPICall indicates a general API call error
	ErrAPICall = errors.New("error calling LLM API")

	// ErrClientInitialization indicates client initialization failed
	ErrClientInitialization = errors.New("error creating LLM client")

	// ErrUnsupportedModel indicates an unsupported model was requested
	ErrUnsupportedModel = errors.New("unsupported model type")

	// ErrModelNotFound indicates a model definition was not found in the registry
	ErrModelNotFound = errors.New("model definition not found in registry")
)

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// registryAPIService implements the APIService interface using the Registry
type registryAPIService struct {
	registry *registry.Registry
	logger   logutil.LoggerInterface
}

// NewRegistryAPIService creates a new Registry-based API service
// This implementation uses the registry to look up model and provider information,
// providing a more flexible and configurable approach than the legacy APIService.
func NewRegistryAPIService(registryManager *registry.Manager, logger logutil.LoggerInterface) interfaces.APIService {
	return &registryAPIService{
		registry: registryManager.GetRegistry(),
		logger:   logger,
	}
}

// InitLLMClient initializes and returns an LLM client based on the model name
// using the Registry to look up model and provider information
func (s *registryAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// Validate required parameters
	if modelName == "" {
		return nil, fmt.Errorf("%w: model name is required", ErrClientInitialization)
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInitialization, ctx.Err())
	}

	// Log custom endpoint if provided (used for API endpoint override)
	if apiEndpoint != "" {
		s.logger.Debug("Using custom API endpoint: %s", apiEndpoint)
	}

	// Special case for testing with error-model
	if modelName == "error-model" {
		return nil, errors.New("test model error")
	}

	// Look up the model in the registry
	modelDef, err := s.registry.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		// If the model isn't found in the registry, try the fallback strategy
		return s.createLLMClientFallback(ctx, apiKey, modelName, apiEndpoint)
	}

	// Get the provider info from the registry
	providerDef, err := s.registry.GetProvider(modelDef.Provider)
	if err != nil {
		s.logger.Debug("Provider '%s' not found in registry: %v", modelDef.Provider, err)
		return nil, fmt.Errorf("%w: provider for model '%s' not found: %v",
			ErrClientInitialization, modelName, err)
	}

	// Determine which API endpoint to use
	effectiveEndpoint := apiEndpoint
	if effectiveEndpoint == "" && providerDef.BaseURL != "" {
		effectiveEndpoint = providerDef.BaseURL
	}

	// Get the provider implementation from the registry
	providerImpl, err := s.registry.GetProviderImplementation(modelDef.Provider)
	if err != nil {
		s.logger.Debug("Provider implementation '%s' not found: %v", modelDef.Provider, err)
		return nil, fmt.Errorf("%w: provider implementation for '%s' not registered",
			ErrClientInitialization, modelDef.Provider)
	}

	// API Key Resolution Logic
	// ------------------------
	// The system follows this precedence order for API keys:
	// 1. Environment variables specific to each provider (highest priority)
	//    - For OpenAI: OPENAI_API_KEY
	//    - For Gemini: GEMINI_API_KEY
	//    - For OpenRouter: OPENROUTER_API_KEY
	//    These mappings are defined in ~/.config/thinktank/models.yaml
	// 2. Explicitly provided API key parameter (fallback only)
	//
	// This ensures proper isolation of API keys between different providers,
	// preventing issues like using an OpenAI key for OpenRouter requests.
	// Each provider requires its own specific API key format.

	// Start with an empty key, which we'll populate from environment or passed parameter
	effectiveApiKey := ""

	// STEP 1: First try to get the key from environment variable based on provider
	// This is the recommended and preferred method for providing API keys
	configLoader := registry.NewConfigLoader()
	modelConfig, err := configLoader.Load()
	if err == nil && modelConfig != nil && modelConfig.APIKeySources != nil {
		if envVar, ok := modelConfig.APIKeySources[modelDef.Provider]; ok && envVar != "" {
			envApiKey := os.Getenv(envVar)
			if envApiKey != "" {
				effectiveApiKey = envApiKey
				s.logger.Debug("Using API key from environment variable %s for provider '%s'",
					envVar, modelDef.Provider)
			}
		}
	}

	// STEP 2: Only fall back to the passed apiKey if environment variable is not set
	// This is discouraged for production use but supported for testing/development
	if effectiveApiKey == "" && apiKey != "" {
		effectiveApiKey = apiKey
		s.logger.Debug("Environment variable not set or empty, using provided API key for provider '%s'",
			modelDef.Provider)
	}

	// STEP 3: If no API key is available from either source, reject the request
	// API keys are required for all providers
	if effectiveApiKey == "" {
		envVarName := getEnvVarNameForProvider(modelDef.Provider, modelConfig)
		return nil, fmt.Errorf("%w: API key is required for model '%s' with provider '%s'. Please set the %s environment variable",
			ErrClientInitialization, modelName, modelDef.Provider, envVarName)
	}

	// Create the client using the provider implementation
	s.logger.Debug("Creating LLM client for model '%s' using provider '%s'",
		modelName, modelDef.Provider)

	// Verify the API key is non-empty before passing it to the provider
	if effectiveApiKey == "" {
		s.logger.Error("Empty API key for provider '%s' - this will cause authentication failures", modelDef.Provider)
	} else {
		// Log a prefix of the API key for easier debugging while maintaining security
		keyPrefix := ""
		if len(effectiveApiKey) >= 5 {
			keyPrefix = effectiveApiKey[:5]
		} else if len(effectiveApiKey) > 0 {
			keyPrefix = effectiveApiKey[:]
		}
		s.logger.Debug("Using API key for provider '%s' (length: %d, starts with: %s)",
			modelDef.Provider, len(effectiveApiKey), keyPrefix)
	}

	client, err := providerImpl.CreateClient(ctx, effectiveApiKey, modelDef.APIModelID, effectiveEndpoint)
	if err != nil {
		// Check if it's already an API error with enhanced details from Gemini
		if apiErr, ok := gemini.IsAPIError(err); ok {
			return nil, fmt.Errorf("%w: %s", ErrClientInitialization, apiErr.UserFacingError())
		}

		// Check if it's an OpenAI API error
		if apiErr, ok := openai.IsOpenAIError(err); ok {
			return nil, fmt.Errorf("%w: %s", ErrClientInitialization, apiErr.UserFacingError())
		}

		// Wrap the original error
		return nil, fmt.Errorf("%w: %v", ErrClientInitialization, err)
	}

	return client, nil
}

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	// ProviderGemini represents the Gemini provider
	ProviderGemini ProviderType = "gemini"
	// ProviderOpenAI represents the OpenAI provider
	ProviderOpenAI ProviderType = "openai"
	// ProviderUnknown represents an unknown provider
	ProviderUnknown ProviderType = "unknown"
)

// detectProviderFromModelName is an internal function that detects the provider type from model name
// This is used only as a fallback when a model isn't found in the registry
func detectProviderFromModelName(modelName string) ProviderType {
	if modelName == "" {
		return ProviderUnknown
	}

	// Check for Gemini models
	if len(modelName) >= 6 && modelName[:6] == "gemini" {
		return ProviderGemini
	}

	// Check for OpenAI GPT models
	if len(modelName) >= 3 && modelName[:3] == "gpt" {
		return ProviderOpenAI
	}

	// Check for other OpenAI models
	otherOpenAIModels := []string{
		"text-davinci",
		"davinci",
		"curie",
		"babbage",
		"ada",
		"text-embedding",
		"text-moderation",
		"whisper",
	}

	for _, prefix := range otherOpenAIModels {
		if len(modelName) >= len(prefix) && modelName[:len(prefix)] == prefix {
			return ProviderOpenAI
		}
	}

	// Unknown model type
	return ProviderUnknown
}

// createLLMClientFallback is used when a model isn't found in the registry
// It uses basic string pattern matching for backward compatibility
func (s *registryAPIService) createLLMClientFallback(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	s.logger.Warn("Using fallback provider detection for model '%s'. Please add it to your models.yaml configuration.", modelName)

	// Detect provider type from model name using string pattern matching
	providerType := detectProviderFromModelName(modelName)

	// API Key Resolution Logic for Legacy Provider Detection
	// --------------------------------------------------
	// This follows the same precedence logic as the main path:
	// 1. Environment variables specific to each provider (highest priority)
	// 2. Explicitly provided API key parameter (fallback only)
	//
	// Note: This is a legacy fallback path for models not found in the registry

	// Start with an empty key, which we'll populate from environment or passed parameter
	effectiveApiKey := ""

	// STEP 1: First, try to get the key from environment variables based on provider type
	// Create a new config loader to get the API key sources from config
	configLoader := registry.NewConfigLoader()
	modelConfig, configErr := configLoader.Load()

	// If we have a config, use its API key mappings from models.yaml
	if configErr == nil && modelConfig != nil && modelConfig.APIKeySources != nil {
		var envVar string
		var ok bool

		// Map the provider type to the provider name used in the config
		var providerName string
		switch providerType {
		case ProviderGemini:
			providerName = "gemini"
		case ProviderOpenAI:
			providerName = "openai"
		}

		// Look up the environment variable name for this provider
		if providerName != "" {
			if envVar, ok = modelConfig.APIKeySources[providerName]; ok && envVar != "" {
				envApiKey := os.Getenv(envVar)
				if envApiKey != "" {
					effectiveApiKey = envApiKey
					s.logger.Debug("Using API key from environment variable %s for provider %s",
						envVar, providerName)

					// Log the API key prefix for debugging
					if len(effectiveApiKey) > 0 {
						keyPrefix := ""
						if len(effectiveApiKey) >= 5 {
							keyPrefix = effectiveApiKey[:5]
						} else {
							keyPrefix = effectiveApiKey[:]
						}
						s.logger.Debug("API key for provider %s (length: %d, starts with: %s)",
							providerName, len(effectiveApiKey), keyPrefix)
					}
				}
			}
		}
	} else {
		// Fallback to hardcoded defaults if no config is available
		// This path is rarely used but provided as a safety net
		s.logger.Warn("No models.yaml configuration found. Using hardcoded environment variable names.")

		switch providerType {
		case ProviderGemini:
			envApiKey := os.Getenv("GEMINI_API_KEY")
			if envApiKey != "" {
				effectiveApiKey = envApiKey
				s.logger.Debug("No config found. Using API key from GEMINI_API_KEY environment variable")
				if len(effectiveApiKey) > 0 {
					s.logger.Debug("API key length: %d, starts with: %s",
						len(effectiveApiKey), effectiveApiKey[:min(5, len(effectiveApiKey))])
				}
			}
		case ProviderOpenAI:
			envApiKey := os.Getenv("OPENAI_API_KEY")
			if envApiKey != "" {
				effectiveApiKey = envApiKey
				s.logger.Debug("No config found. Using API key from OPENAI_API_KEY environment variable")
				if len(effectiveApiKey) > 0 {
					s.logger.Debug("API key length: %d, starts with: %s",
						len(effectiveApiKey), effectiveApiKey[:min(5, len(effectiveApiKey))])
				}
			}
		}
	}

	// STEP 2: Only fall back to the passed apiKey if environment variable is not set
	// This is discouraged for production use but supported for testing/development
	if effectiveApiKey == "" && apiKey != "" {
		effectiveApiKey = apiKey
		s.logger.Debug("Environment variable not set or empty, using provided API key for provider type %v",
			providerType)
	}

	// STEP 3: If no API key is available from either source, reject the request
	// API keys are required for all providers
	if effectiveApiKey == "" {
		var envVarName string
		switch providerType {
		case ProviderGemini:
			envVarName = "GEMINI_API_KEY"
		case ProviderOpenAI:
			envVarName = "OPENAI_API_KEY"
		default:
			envVarName = "API_KEY for this provider"
		}

		return nil, fmt.Errorf("%w: API key is required for provider %s. Please set the %s environment variable",
			ErrClientInitialization, providerType, envVarName)
	}

	// Initialize the appropriate client based on provider type
	var client llm.LLMClient
	var err error

	switch providerType {
	case ProviderGemini:
		s.logger.Debug("Using Gemini provider for model %s (legacy detection)", modelName)
		client, err = gemini.NewLLMClient(ctx, effectiveApiKey, modelName, apiEndpoint)
	case ProviderOpenAI:
		s.logger.Debug("Using OpenAI provider for model %s (legacy detection)", modelName)

		client, err = openai.NewClient(effectiveApiKey, modelName, apiEndpoint)
	case ProviderUnknown:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedModel, modelName)
	}

	// Handle client creation error
	if err != nil {
		// Check if it's already an API error with enhanced details from Gemini
		if apiErr, ok := gemini.IsAPIError(err); ok {
			return nil, fmt.Errorf("%w: %s", ErrClientInitialization, apiErr.UserFacingError())
		}

		// Check if it's an OpenAI API error
		if apiErr, ok := openai.IsOpenAIError(err); ok {
			return nil, fmt.Errorf("%w: %s", ErrClientInitialization, apiErr.UserFacingError())
		}

		// Wrap the original error
		return nil, fmt.Errorf("%w: %v", ErrClientInitialization, err)
	}

	return client, nil
}

// The remaining methods are carried over from the existing APIService implementation
// since they don't depend on the provider initialization logic

// GetModelParameters retrieves parameter values from the registry for a given model
// It returns a map of parameter name to parameter value, applying defaults from the model definition
func (s *registryAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	// Look up the model in the registry
	modelDef, err := s.registry.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		// Return an empty map if model not found
		return make(map[string]interface{}), nil
	}

	// Prepare the parameters map
	params := make(map[string]interface{})

	// Add all parameters from the model definition with their default values
	for paramName, paramDef := range modelDef.Parameters {
		// Apply the default value from the definition
		if paramDef.Default != nil {
			params[paramName] = paramDef.Default
		}
	}

	return params, nil
}

// ValidateModelParameter validates a parameter value against its constraints
// It returns true if the parameter is valid, false otherwise
func (s *registryAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	// Look up the model in the registry
	modelDef, err := s.registry.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		// Can't validate if model is not found
		return false, fmt.Errorf("model '%s' not found: %v", modelName, err)
	}

	// Look up the parameter definition
	paramDef, ok := modelDef.Parameters[paramName]
	if !ok {
		s.logger.Debug("Parameter '%s' not found for model '%s'", paramName, modelName)
		return false, fmt.Errorf("parameter '%s' not defined for model '%s'", paramName, modelName)
	}

	// Type check and constraints validation
	switch paramDef.Type {
	case "float":
		// Check if value is a float
		floatVal, ok := value.(float64)
		if !ok {
			return false, fmt.Errorf("parameter '%s' must be a float", paramName)
		}

		// Check minimum value if defined
		if paramDef.Min != nil {
			minVal, ok := paramDef.Min.(float64)
			if ok && floatVal < minVal {
				return false, fmt.Errorf("parameter '%s' value %.2f is below minimum %.2f",
					paramName, floatVal, minVal)
			}
		}

		// Check maximum value if defined
		if paramDef.Max != nil {
			maxVal, ok := paramDef.Max.(float64)
			if ok && floatVal > maxVal {
				return false, fmt.Errorf("parameter '%s' value %.2f exceeds maximum %.2f",
					paramName, floatVal, maxVal)
			}
		}

	case "int":
		// Check if value is an int
		intVal, ok := value.(int)
		if !ok {
			return false, fmt.Errorf("parameter '%s' must be an integer", paramName)
		}

		// Check minimum value if defined
		if paramDef.Min != nil {
			minVal, ok := paramDef.Min.(int)
			if ok && intVal < minVal {
				return false, fmt.Errorf("parameter '%s' value %d is below minimum %d",
					paramName, intVal, minVal)
			}
		}

		// Check maximum value if defined
		if paramDef.Max != nil {
			maxVal, ok := paramDef.Max.(int)
			if ok && intVal > maxVal {
				return false, fmt.Errorf("parameter '%s' value %d exceeds maximum %d",
					paramName, intVal, maxVal)
			}
		}

	case "string":
		// Check if value is a string
		strVal, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("parameter '%s' must be a string", paramName)
		}

		// Check enum values if defined
		if len(paramDef.EnumValues) > 0 {
			valid := false
			for _, enumVal := range paramDef.EnumValues {
				if strVal == enumVal {
					valid = true
					break
				}
			}
			if !valid {
				return false, fmt.Errorf("parameter '%s' value '%s' is not in allowed values: %v",
					paramName, strVal, paramDef.EnumValues)
			}
		}

	default:
		// Unknown parameter type
		s.logger.Warn("Unknown parameter type '%s' for parameter '%s'", paramDef.Type, paramName)
	}

	// If we reached here, the parameter is valid
	return true, nil
}

// GetModelDefinition retrieves the full model definition from the registry
func (s *registryAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	// Look up the model in the registry
	modelDef, err := s.registry.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		return nil, fmt.Errorf("%w: %s", ErrModelNotFound, modelName)
	}

	return modelDef, nil
}

// GetModelTokenLimits retrieves token limits from the registry for a given model
// Note: This method is kept for backward compatibility but now returns default values
// Token handling has been removed as part of T036C
func (s *registryAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Look up the model in the registry to verify it exists
	_, err = s.registry.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		return 0, 0, fmt.Errorf("%w: %s", ErrModelNotFound, modelName)
	}

	// Return default values instead of actual model values
	// Token handling is now the responsibility of each provider
	return 8192, 2048, nil
}

// ProcessLLMResponse processes a provider-agnostic API response and extracts content
func (s *registryAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	// Check for nil result
	if result == nil {
		return "", fmt.Errorf("%w: result is nil", ErrEmptyResponse)
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
				return "", fmt.Errorf("%w%s", ErrSafetyBlocked, errDetails.String())
			}
		}

		// If we don't have safety blocks, use the generic empty response error
		return "", fmt.Errorf("%w%s", ErrEmptyResponse, errDetails.String())
	}

	// Check for whitespace-only content
	if strings.TrimSpace(result.Content) == "" {
		return "", ErrWhitespaceContent
	}

	return result.Content, nil
}

// IsEmptyResponseError checks if an error is related to empty API responses
func (s *registryAPIService) IsEmptyResponseError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types using errors.Is
	if errors.Is(err, ErrEmptyResponse) || errors.Is(err, ErrWhitespaceContent) {
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
	if errors.Is(err, ErrSafetyBlocked) {
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

// getEnvVarNameForProvider returns the appropriate environment variable name for a given provider
// This helper function is used to provide better error messages to users
func getEnvVarNameForProvider(providerName string, modelConfig *registry.ModelsConfig) string {
	// Try to get the env var name from the config
	if modelConfig != nil && modelConfig.APIKeySources != nil {
		if envVar, ok := modelConfig.APIKeySources[providerName]; ok && envVar != "" {
			return envVar
		}
	}

	// Fallback to hard-coded defaults if not in config
	switch providerName {
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	default:
		// Use a generic format for unknown providers
		return strings.ToUpper(providerName) + "_API_KEY"
	}
}

// GetErrorDetails extracts detailed information from an error
func (s *registryAPIService) GetErrorDetails(err error) string {
	// Handle nil error case
	if err == nil {
		return "no error"
	}

	// Check if it's a Gemini API error with enhanced details
	if apiErr, ok := gemini.IsAPIError(err); ok {
		return apiErr.UserFacingError()
	}

	// Check if it's an OpenAI API error with enhanced details
	if apiErr, ok := openai.IsOpenAIError(err); ok {
		return apiErr.UserFacingError()
	}

	// Return the error string for other error types
	return err.Error()
}
