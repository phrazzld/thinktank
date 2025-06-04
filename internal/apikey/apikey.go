// Package apikey provides centralized API key resolution functionality.
package apikey

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// APIKeySource represents the source of an API key
type APIKeySource int

const (
	// APIKeySourceNone indicates no API key was found
	APIKeySourceNone APIKeySource = iota
	// APIKeySourceEnvironment indicates the key came from an environment variable
	APIKeySourceEnvironment
	// APIKeySourceParameter indicates the key came from a function parameter
	APIKeySourceParameter
)

// APIKeyResult contains the resolved API key and metadata about its source
type APIKeyResult struct {
	// Key is the resolved API key value
	Key string
	// Source indicates where the key came from
	Source APIKeySource
	// EnvironmentVariable is the name of the environment variable used (if applicable)
	EnvironmentVariable string
	// Provider is the name of the provider this key is for
	Provider string
}

// APIKeyResolver handles API key resolution with a clear precedence order
type APIKeyResolver struct {
	logger             logutil.LoggerInterface
	apiKeySources      map[string]string // provider -> env var mapping
}

// NewAPIKeyResolver creates a new API key resolver
func NewAPIKeyResolver(logger logutil.LoggerInterface) *APIKeyResolver {
	if logger == nil {
		logger = logutil.NewLogger(logutil.InfoLevel, nil, "[apikey] ")
	}
	
	return &APIKeyResolver{
		logger:        logger,
		apiKeySources: make(map[string]string),
	}
}

// NewAPIKeyResolverWithConfig creates a new API key resolver with custom API key sources
func NewAPIKeyResolverWithConfig(logger logutil.LoggerInterface, apiKeySources map[string]string) *APIKeyResolver {
	if logger == nil {
		logger = logutil.NewLogger(logutil.InfoLevel, nil, "[apikey] ")
	}
	
	return &APIKeyResolver{
		logger:        logger,
		apiKeySources: apiKeySources,
	}
}

// ResolveAPIKey resolves an API key for a given provider following a clear precedence order:
// 1. Environment variables specific to each provider (highest priority)
//    - For OpenAI: OPENAI_API_KEY
//    - For Gemini: GEMINI_API_KEY  
//    - For OpenRouter: OPENROUTER_API_KEY
//    These mappings can be customized via NewAPIKeyResolverWithConfig
// 2. Explicitly provided API key parameter (fallback only)
//
// This ensures proper isolation of API keys between different providers,
// preventing issues like using an OpenAI key for OpenRouter requests.
func (r *APIKeyResolver) ResolveAPIKey(ctx context.Context, providerName, providedKey string) (*APIKeyResult, error) {
	result := &APIKeyResult{
		Provider: providerName,
		Source:   APIKeySourceNone,
	}

	// STEP 1: First try to get the key from environment variable based on provider
	// This is the recommended and preferred method for providing API keys
	envVarName := r.getEnvironmentVariableName(providerName)
	if envVarName != "" {
		envAPIKey := os.Getenv(envVarName)
		if envAPIKey != "" {
			result.Key = envAPIKey
			result.Source = APIKeySourceEnvironment
			result.EnvironmentVariable = envVarName
			r.logger.DebugContext(ctx, "Using API key from environment variable %s for provider '%s'",
				envVarName, providerName)
			return result, nil
		}
		r.logger.DebugContext(ctx, "Environment variable %s not set or empty for provider '%s'",
			envVarName, providerName)
	}

	// STEP 2: Only fall back to the passed apiKey if environment variable is not set
	// This is discouraged for production use but supported for testing/development
	if providedKey != "" {
		result.Key = providedKey
		result.Source = APIKeySourceParameter
		r.logger.DebugContext(ctx, "Environment variable not set, using provided API key for provider '%s'",
			providerName)
		return result, nil
	}

	// STEP 3: If no API key is available from either source, return an error
	// API keys are required for all providers
	return nil, r.createMissingKeyError(providerName, envVarName)
}

// ValidateAPIKey performs basic validation on an API key
func (r *APIKeyResolver) ValidateAPIKey(ctx context.Context, providerName, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Provider-specific validation
	switch strings.ToLower(providerName) {
	case "openai":
		// OpenAI keys should start with "sk-"
		if !strings.HasPrefix(apiKey, "sk-") {
			r.logger.WarnContext(ctx, "OpenAI API key does not have expected format (should start with 'sk-')")
		}
	case "gemini":
		// Gemini keys are typically longer alphanumeric strings
		if len(apiKey) < 20 {
			r.logger.WarnContext(ctx, "Gemini API key appears unusually short")
		}
	case "openrouter":
		// OpenRouter keys can have various formats
		// No specific validation at this time
	}

	// Log key metadata (never log the actual key)
	r.logger.DebugContext(ctx, "Validated API key for provider '%s' (length: %d)",
		providerName, len(apiKey))

	return nil
}

// getEnvironmentVariableName returns the environment variable name for a given provider
func (r *APIKeyResolver) getEnvironmentVariableName(providerName string) string {
	// Try to get the env var name from the configured sources
	if r.apiKeySources != nil {
		if envVar, ok := r.apiKeySources[providerName]; ok && envVar != "" {
			return envVar
		}
	}

	// Fallback to hard-coded defaults if not configured
	switch strings.ToLower(providerName) {
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

// createMissingKeyError creates a descriptive error for missing API keys
func (r *APIKeyResolver) createMissingKeyError(providerName, envVarName string) error {
	if envVarName == "" {
		envVarName = r.getEnvironmentVariableName(providerName)
	}

	return llm.Wrap(
		fmt.Errorf("API key required but not found"),
		"",
		fmt.Sprintf("API key is required for provider '%s'. Please set the %s environment variable",
			providerName, envVarName),
		llm.CategoryInvalidRequest,
	)
}

// GetAPIKeyPrecedenceDocumentation returns human-readable documentation about API key precedence
func (r *APIKeyResolver) GetAPIKeyPrecedenceDocumentation() string {
	return `API Key Resolution Precedence:

1. Environment Variables (Highest Priority)
   - OpenAI: OPENAI_API_KEY
   - Gemini: GEMINI_API_KEY
   - OpenRouter: OPENROUTER_API_KEY
   - Custom providers: Check ~/.config/thinktank/models.yaml

2. Explicitly Provided API Key Parameter (Fallback)
   - Used only if environment variable is not set
   - Primarily for testing and development

Note: Each provider requires its own specific API key. Never use one provider's
key for another provider as this will result in authentication failures.`
}