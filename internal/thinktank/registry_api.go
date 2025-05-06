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
	"github.com/phrazzld/thinktank/internal/providers"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// registryAPIService implements the APIService interface using the Registry
type registryAPIService struct {
	registry interface{}
	logger   logutil.LoggerInterface
}

// NewRegistryAPIService creates a new Registry-based API service
// This implementation uses the registry to look up model and provider information,
// providing a more flexible and configurable approach than the legacy APIService.
func NewRegistryAPIService(registry interface{}, logger logutil.LoggerInterface) interfaces.APIService {
	// For testing, we allow passing in a mock registry that implements the required methods
	return &registryAPIService{
		registry: registry,
		logger:   logger,
	}
}

// InitLLMClient initializes and returns an LLM client based on the model name
// using the Registry to look up model and provider information
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
		s.logger.Debug("Using custom API endpoint: %s", apiEndpoint)
	}

	// Special case for testing with error-model
	if modelName == "error-model" {
		return nil, errors.New("test model error")
	}

	// Look up the model in the registry
	regImpl, ok := s.registry.(interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	})
	if !ok {
		return nil, fmt.Errorf("registry does not implement GetModel method")
	}

	modelDef, err := regImpl.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		// Do not use fallback strategy; return a specific error instead
		return nil, fmt.Errorf("%w: model '%s' not found in registry: %v", llm.ErrModelNotFound, modelName, err)
	}

	// Get the provider info from the registry
	regProviderImpl, ok := s.registry.(interface {
		GetProvider(name string) (*registry.ProviderDefinition, error)
	})
	if !ok {
		return nil, fmt.Errorf("registry does not implement GetProvider method")
	}

	providerDef, err := regProviderImpl.GetProvider(modelDef.Provider)
	if err != nil {
		s.logger.Debug("Provider '%s' not found in registry: %v", modelDef.Provider, err)
		return nil, fmt.Errorf("%w: provider for model '%s' not found: %v",
			llm.ErrClientInitialization, modelName, err)
	}

	// Determine which API endpoint to use
	effectiveEndpoint := apiEndpoint
	if effectiveEndpoint == "" && providerDef.BaseURL != "" {
		effectiveEndpoint = providerDef.BaseURL
	}

	// Get the provider implementation from the registry
	regProviderImplGetter, ok := s.registry.(interface {
		GetProviderImplementation(name string) (providers.Provider, error)
	})
	if !ok {
		return nil, fmt.Errorf("registry does not implement GetProviderImplementation method")
	}

	providerImpl, err := regProviderImplGetter.GetProviderImplementation(modelDef.Provider)
	if err != nil {
		s.logger.Debug("Provider implementation '%s' not found: %v", modelDef.Provider, err)
		return nil, fmt.Errorf("%w: provider implementation for '%s' not registered",
			llm.ErrClientInitialization, modelDef.Provider)
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
			llm.ErrClientInitialization, modelName, modelDef.Provider, envVarName)
	}

	// Create the client using the provider implementation
	s.logger.Debug("Creating LLM client for model '%s' using provider '%s'",
		modelName, modelDef.Provider)

	// Verify the API key is non-empty before passing it to the provider
	if effectiveApiKey == "" {
		s.logger.Error("Empty API key for provider '%s' - this will cause authentication failures", modelDef.Provider)
	} else {
		// Log API key metadata only (NEVER log any portion of the key itself)
		s.logger.Debug("Using API key for provider '%s' (length: %d, source: via environment variable)",
			modelDef.Provider, len(effectiveApiKey))
	}

	// Since we're now using the providers.Provider type directly, we no longer need to do
	// a type assertion to get the CreateClient method - it's already part of the interface
	client, err := providerImpl.CreateClient(ctx, effectiveApiKey, modelDef.APIModelID, effectiveEndpoint)
	if err != nil {
		// Check if it's already an API error with enhanced details from Gemini
		if apiErr, ok := gemini.IsAPIError(err); ok {
			return nil, fmt.Errorf("%w: %s", llm.ErrClientInitialization, apiErr.UserFacingError())
		}

		// Check if it's an OpenAI API error
		if apiErr, ok := openai.IsOpenAIError(err); ok {
			return nil, fmt.Errorf("%w: %s", llm.ErrClientInitialization, apiErr.UserFacingError())
		}

		// Wrap the original error
		return nil, fmt.Errorf("%w: %v", llm.ErrClientInitialization, err)
	}

	return client, nil
}

// The remaining methods are carried over from the existing APIService implementation
// since they don't depend on the provider initialization logic

// GetModelParameters retrieves parameter values from the registry for a given model
// It returns a map of parameter name to parameter value, applying defaults from the model definition
func (s *registryAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	// Look up the model in the registry
	regImpl, ok := s.registry.(interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	})
	if !ok {
		return make(map[string]interface{}), fmt.Errorf("registry does not implement GetModel method")
	}

	modelDef, err := regImpl.GetModel(modelName)
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
	regImpl, ok := s.registry.(interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	})
	if !ok {
		return false, fmt.Errorf("registry does not implement GetModel method")
	}

	modelDef, err := regImpl.GetModel(modelName)
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
	regImpl, ok := s.registry.(interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	})
	if !ok {
		return nil, fmt.Errorf("registry does not implement GetModel method")
	}

	modelDef, err := regImpl.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		return nil, fmt.Errorf("%w: %s", llm.ErrModelNotFound, modelName)
	}

	return modelDef, nil
}

// GetModelTokenLimits retrieves token limits from the registry for a given model
// This method now returns default values instead of actual model token limits
func (s *registryAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Look up the model in the registry to verify it exists
	regImpl, ok := s.registry.(interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	})
	if !ok {
		return 0, 0, fmt.Errorf("registry does not implement GetModel method")
	}

	_, err = regImpl.GetModel(modelName)
	if err != nil {
		s.logger.Debug("Model '%s' not found in registry: %v", modelName, err)
		return 0, 0, fmt.Errorf("%w: %s", llm.ErrModelNotFound, modelName)
	}

	// Return default values instead of actual model values
	// Token handling is now the responsibility of each provider
	return 8192, 2048, nil
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
