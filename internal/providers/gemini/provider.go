// internal/providers/gemini/provider.go
package gemini

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/phrazzld/thinktank/internal/gemini"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/providers"
)

// GeminiProvider implements the Provider interface for Gemini models.
type GeminiProvider struct {
	logger logutil.LoggerInterface
	apiKey string
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

	// Use provided API key first
	effectiveAPIKey := apiKey

	// If none provided, try environment variable
	if effectiveAPIKey == "" {
		effectiveAPIKey = os.Getenv("GOOGLE_API_KEY")
		if effectiveAPIKey == "" {
			return nil, fmt.Errorf("no API key provided and GOOGLE_API_KEY environment variable not set")
		}
		p.logger.Debug("Using API key from GOOGLE_API_KEY environment variable")
	} else {
		p.logger.Debug("Using provided API key")
	}

	// Store API key for later use
	p.apiKey = effectiveAPIKey

	// Create the client using the Gemini implementation
	baseClient, err := gemini.NewLLMClient(ctx, effectiveAPIKey, modelID, apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Wrap the client in our parameter-aware adapter
	return NewGeminiClientAdapter(baseClient), nil
}

// GeminiClientAdapter wraps the standard Gemini client to provide
// parameter-aware capabilities
type GeminiClientAdapter struct {
	client llm.LLMClient
	params map[string]interface{}
	mu     sync.RWMutex
}

// NewGeminiClientAdapter creates a new adapter for the Gemini client
func NewGeminiClientAdapter(client llm.LLMClient) *GeminiClientAdapter {
	return &GeminiClientAdapter{
		client: client,
		params: make(map[string]interface{}),
		mu:     sync.RWMutex{},
	}
}

// SetParameters sets the parameters to use for API calls
func (a *GeminiClientAdapter) SetParameters(params map[string]interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.params = params
}

// GenerateContent implements the llm.LLMClient interface and applies parameters
func (a *GeminiClientAdapter) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Validate prompt
	if prompt == "" {
		return nil, gemini.CreateAPIError(
			llm.CategoryInvalidRequest,
			"Empty prompt provided",
			nil,
			"Please provide a non-empty prompt",
		)
	}

	// Apply parameters if provided
	if len(params) > 0 {
		a.params = params
	}

	// Validate parameters before applying them
	if err := a.validateParameters(); err != nil {
		return nil, err
	}

	// We need to convert the generic parameters from the models package to Gemini-specific settings
	// We'll try to apply common parameters to the underlying client if it supports them

	// For Gemini adapter, we use safer type assertions with specific interfaces
	// Since we don't control the underlying Gemini client implementation directly

	// Apply temperature parameter if client supports it
	if client, ok := a.client.(interface{ SetTemperature(temp float32) }); ok && a.params != nil {
		if temp, ok := a.getFloatParam("temperature"); ok {
			client.SetTemperature(temp)
		}
	}

	// Apply top_p parameter if client supports it
	if client, ok := a.client.(interface{ SetTopP(topP float32) }); ok && a.params != nil {
		if topP, ok := a.getFloatParam("top_p"); ok {
			client.SetTopP(topP)
		}
	}

	// Apply top_k parameter if client supports it
	if client, ok := a.client.(interface{ SetTopK(topK int32) }); ok && a.params != nil {
		if topK, ok := a.getIntParam("top_k"); ok {
			client.SetTopK(topK)
		}
	}

	// Apply max_tokens parameter if client supports it
	if client, ok := a.client.(interface{ SetMaxOutputTokens(tokens int32) }); ok && a.params != nil {
		if maxTokens, ok := a.getIntParam("max_tokens"); ok {
			client.SetMaxOutputTokens(maxTokens)
		} else if maxTokens, ok := a.getIntParam("max_output_tokens"); ok {
			// Try the Gemini-style parameter name as a fallback
			client.SetMaxOutputTokens(maxTokens)
		}
	}

	// Call the underlying client's implementation with combined parameters
	return a.client.GenerateContent(ctx, prompt, a.params)
}

// getFloatParam safely extracts a float parameter
func (a *GeminiClientAdapter) getFloatParam(name string) (float32, bool) {
	if a.params == nil {
		return 0, false
	}

	if val, ok := a.params[name]; ok {
		switch v := val.(type) {
		case float64:
			return float32(v), true
		case float32:
			return v, true
		case int:
			return float32(v), true
		case int32:
			return float32(v), true
		case int64:
			return float32(v), true
		}
	}
	return 0, false
}

// getIntParam safely extracts an integer parameter
func (a *GeminiClientAdapter) getIntParam(name string) (int32, bool) {
	if a.params == nil {
		return 0, false
	}

	if val, ok := a.params[name]; ok {
		switch v := val.(type) {
		case int:
			return int32(v), true
		case int32:
			return v, true
		case int64:
			return int32(v), true
		case float64:
			return int32(v), true
		case float32:
			return int32(v), true
		}
	}
	return 0, false
}

// GetModelName implements the llm.LLMClient interface
func (a *GeminiClientAdapter) GetModelName() string {
	return a.client.GetModelName()
}

// Close implements the llm.LLMClient interface
func (a *GeminiClientAdapter) Close() error {
	return a.client.Close()
}

// validateParameters validates parameter values according to Gemini API requirements
func (a *GeminiClientAdapter) validateParameters() error {
	if a.params == nil {
		return nil
	}

	var validationErrors []string

	// Validate temperature (0.0 to 2.0)
	if temp, exists := a.params["temperature"]; exists {
		if tempVal, ok := a.getFloatParam("temperature"); ok {
			if tempVal < 0.0 || tempVal > 2.0 {
				validationErrors = append(validationErrors, fmt.Sprintf("temperature must be between 0.0 and 2.0, got %v", temp))
			}
		}
	}

	// Validate top_p (0.0 to 1.0)
	if topP, exists := a.params["top_p"]; exists {
		if topPVal, ok := a.getFloatParam("top_p"); ok {
			if topPVal < 0.0 || topPVal > 1.0 {
				validationErrors = append(validationErrors, fmt.Sprintf("top_p must be between 0.0 and 1.0, got %v", topP))
			}
		}
	}

	// Validate top_k (must be positive)
	if topK, exists := a.params["top_k"]; exists {
		if topKVal, ok := a.getIntParam("top_k"); ok {
			if topKVal <= 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("top_k must be positive, got %v", topK))
			}
		}
	}

	// Validate max_output_tokens (must be positive)
	if maxTokens, exists := a.params["max_output_tokens"]; exists {
		if maxTokensVal, ok := a.getIntParam("max_output_tokens"); ok {
			if maxTokensVal <= 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("max_output_tokens must be positive, got %v", maxTokens))
			}
		}
	}

	// If there are validation errors, return them
	if len(validationErrors) > 0 {
		if len(validationErrors) == 1 {
			return gemini.CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Invalid parameter: %s", validationErrors[0]),
				nil,
				"Please check the parameter values and ensure they are within valid ranges",
			)
		} else {
			return gemini.CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Multiple invalid parameters: %s", strings.Join(validationErrors, "; ")),
				nil,
				"Please check the parameter values and ensure they are within valid ranges",
			)
		}
	}

	return nil
}
