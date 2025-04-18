// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// ChatCompletionMessage represents a message in the OpenRouter chat API format
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request structure for the OpenRouter chat API
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	Temperature      *float32                `json:"temperature,omitempty"`
	TopP             *float32                `json:"top_p,omitempty"`
	FrequencyPenalty *float32                `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32                `json:"presence_penalty,omitempty"`
	MaxTokens        *int32                  `json:"max_tokens,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
}

// ChatCompletionChoice represents a choice in the OpenRouter chat completion response
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

// ChatCompletionUsage represents token usage information
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse represents the response structure from the OpenRouter chat API
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   ChatCompletionUsage    `json:"usage"`
}

// GenerateContent sends a prompt to the LLM and returns the generated content
func (c *openrouterClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// Apply parameters from method argument if provided
	if params != nil {
		// Temperature
		if temp, ok := params["temperature"]; ok {
			switch v := temp.(type) {
			case float32:
				c.temperature = &v
			case float64:
				tempFloat32 := float32(v)
				c.temperature = &tempFloat32
			case int:
				tempFloat32 := float32(v)
				c.temperature = &tempFloat32
			}
		}

		// Top P
		if topP, ok := params["top_p"]; ok {
			switch v := topP.(type) {
			case float32:
				c.topP = &v
			case float64:
				topPFloat32 := float32(v)
				c.topP = &topPFloat32
			case int:
				topPFloat32 := float32(v)
				c.topP = &topPFloat32
			}
		}

		// Presence Penalty
		if penalty, ok := params["presence_penalty"]; ok {
			switch v := penalty.(type) {
			case float32:
				c.presencePenalty = &v
			case float64:
				penaltyFloat32 := float32(v)
				c.presencePenalty = &penaltyFloat32
			case int:
				penaltyFloat32 := float32(v)
				c.presencePenalty = &penaltyFloat32
			}
		}

		// Frequency Penalty
		if penalty, ok := params["frequency_penalty"]; ok {
			switch v := penalty.(type) {
			case float32:
				c.frequencyPenalty = &v
			case float64:
				penaltyFloat32 := float32(v)
				c.frequencyPenalty = &penaltyFloat32
			case int:
				penaltyFloat32 := float32(v)
				c.frequencyPenalty = &penaltyFloat32
			}
		}

		// Max Tokens - try both OpenAI-style and Gemini-style parameter names
		if maxTokens, ok := params["max_tokens"]; ok {
			switch v := maxTokens.(type) {
			case int32:
				c.maxTokens = &v
			case int:
				maxInt32 := int32(v)
				c.maxTokens = &maxInt32
			case int64:
				maxInt32 := int32(v)
				c.maxTokens = &maxInt32
			case float64:
				maxInt32 := int32(v)
				c.maxTokens = &maxInt32
			}
		} else if maxTokens, ok := params["max_output_tokens"]; ok {
			// Try the Gemini-style parameter name as a fallback
			switch v := maxTokens.(type) {
			case int32:
				c.maxTokens = &v
			case int:
				maxInt32 := int32(v)
				c.maxTokens = &maxInt32
			case int64:
				maxInt32 := int32(v)
				c.maxTokens = &maxInt32
			case float64:
				maxInt32 := int32(v)
				c.maxTokens = &maxInt32
			}
		}
	}

	// Create chat completion request
	messages := []ChatCompletionMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Build the request body
	requestBody := ChatCompletionRequest{
		Model:            c.modelID,
		Messages:         messages,
		Temperature:      c.temperature,
		TopP:             c.topP,
		FrequencyPenalty: c.frequencyPenalty,
		PresencePenalty:  c.presencePenalty,
		MaxTokens:        c.maxTokens,
		Stream:           false, // Non-streaming implementation for initial version
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("%s/chat/completions", c.apiEndpoint)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Execute the request
	if c.logger != nil {
		c.logger.Debug("Sending request to OpenRouter API: %s", apiURL)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && c.logger != nil {
			c.logger.Warn("Failed to close response body: %v", closeErr)
		}
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		// Error handling will be implemented in T008
		// For now, just return a basic error with the status code
		return nil, fmt.Errorf("OpenRouter API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var completionResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &completionResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Validate response structure
	if len(completionResponse.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned in the response")
	}

	// Extract the content and other fields
	content := completionResponse.Choices[0].Message.Content
	finishReason := completionResponse.Choices[0].FinishReason

	// Build and return the result
	return &llm.ProviderResult{
		Content:      content,
		FinishReason: finishReason,
		TokenCount:   int32(completionResponse.Usage.CompletionTokens),
		Truncated:    finishReason == "length",
		// OpenRouter doesn't provide safety info in the same format as Gemini,
		// so we leave SafetyInfo empty for now
		SafetyInfo: []llm.Safety{},
	}, nil
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
