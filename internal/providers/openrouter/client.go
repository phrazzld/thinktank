// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// openrouterClient implements the llm.LLMClient interface for OpenRouter
type openrouterClient struct {
	apiKey      string
	modelID     string
	apiEndpoint string
	httpClient  *http.Client
	logger      logutil.LoggerInterface

	// Optional request parameters
	temperature      *float32
	topP             *float32
	presencePenalty  *float32
	frequencyPenalty *float32
	maxTokens        *int32
}

// NewClient creates a new OpenRouter client that implements the llm.LLMClient interface
func NewClient(apiKey string, modelID string, apiEndpoint string, logger logutil.LoggerInterface) (*openrouterClient, error) {
	// Validate required parameters
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	if modelID == "" {
		return nil, fmt.Errorf("model ID cannot be empty")
	}

	// Verify the modelID has the expected format (provider/model or provider/organization/model)
	if !strings.Contains(modelID, "/") {
		if logger != nil {
			logger.Warn("OpenRouter model ID '%s' does not have expected format 'provider/model' or 'provider/organization/model'", modelID)
		}
	}

	// Set default API endpoint if not provided
	if apiEndpoint == "" {
		apiEndpoint = "https://openrouter.ai/api/v1"
	}

	// Create HTTP client with reasonable timeout
	httpClient := &http.Client{
		Timeout: 120 * time.Second, // 2 minute timeout for potentially long LLM generations
	}

	// Create and return client
	return &openrouterClient{
		apiKey:      apiKey,
		modelID:     modelID,
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
		logger:      logger,
	}, nil
}

// GenerateContent sends a prompt to the LLM and returns the generated content
func (c *openrouterClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// This will be implemented in T006
	return nil, fmt.Errorf("GenerateContent method not yet implemented")
}

// CountTokens counts the tokens in the given prompt
func (c *openrouterClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	// This will be implemented in T009
	return nil, fmt.Errorf("CountTokens method not yet implemented")
}

// GetModelInfo retrieves information about the current model
func (c *openrouterClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	// This will be implemented in T010
	return nil, fmt.Errorf("GetModelInfo method not yet implemented")
}

// GetModelName returns the name of the model being used
func (c *openrouterClient) GetModelName() string {
	// This will be implemented in T011
	return c.modelID
}

// Close releases resources used by the client
func (c *openrouterClient) Close() error {
	// This will be implemented in T012
	// For a simple HTTP client, this is typically a no-op
	return nil
}

// Helper methods for setting parameters

// SetTemperature sets the temperature parameter
func (c *openrouterClient) SetTemperature(temp float32) {
	c.temperature = &temp
}

// SetTopP sets the top_p parameter
func (c *openrouterClient) SetTopP(topP float32) {
	c.topP = &topP
}

// SetMaxTokens sets the max_tokens parameter
func (c *openrouterClient) SetMaxTokens(tokens int32) {
	c.maxTokens = &tokens
}

// SetPresencePenalty sets the presence_penalty parameter
func (c *openrouterClient) SetPresencePenalty(penalty float32) {
	c.presencePenalty = &penalty
}

// SetFrequencyPenalty sets the frequency_penalty parameter
func (c *openrouterClient) SetFrequencyPenalty(penalty float32) {
	c.frequencyPenalty = &penalty
}
