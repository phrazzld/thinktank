// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	tiktoken "github.com/pkoukk/tiktoken-go"
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

	// Registry will be injected later if needed
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
	// Create local variables for request parameters instead of modifying receiver fields
	var temperature, topP, presencePenalty, frequencyPenalty *float32
	var maxTokens *int32

	// Copy receiver defaults to local variables if they exist
	if c.temperature != nil {
		temp := *c.temperature
		temperature = &temp
	}
	if c.topP != nil {
		tp := *c.topP
		topP = &tp
	}
	if c.presencePenalty != nil {
		pp := *c.presencePenalty
		presencePenalty = &pp
	}
	if c.frequencyPenalty != nil {
		fp := *c.frequencyPenalty
		frequencyPenalty = &fp
	}
	if c.maxTokens != nil {
		mt := *c.maxTokens
		maxTokens = &mt
	}

	// Apply parameters from method argument if provided
	if params != nil {
		// Temperature
		if temp, ok := params["temperature"]; ok {
			switch v := temp.(type) {
			case float32:
				temperature = &v
			case float64:
				tempFloat32 := float32(v)
				temperature = &tempFloat32
			case int:
				tempFloat32 := float32(v)
				temperature = &tempFloat32
			}
		}

		// Top P
		if tp, ok := params["top_p"]; ok {
			switch v := tp.(type) {
			case float32:
				topP = &v
			case float64:
				topPFloat32 := float32(v)
				topP = &topPFloat32
			case int:
				topPFloat32 := float32(v)
				topP = &topPFloat32
			}
		}

		// Presence Penalty
		if penalty, ok := params["presence_penalty"]; ok {
			switch v := penalty.(type) {
			case float32:
				presencePenalty = &v
			case float64:
				penaltyFloat32 := float32(v)
				presencePenalty = &penaltyFloat32
			case int:
				penaltyFloat32 := float32(v)
				presencePenalty = &penaltyFloat32
			}
		}

		// Frequency Penalty
		if penalty, ok := params["frequency_penalty"]; ok {
			switch v := penalty.(type) {
			case float32:
				frequencyPenalty = &v
			case float64:
				penaltyFloat32 := float32(v)
				frequencyPenalty = &penaltyFloat32
			case int:
				penaltyFloat32 := float32(v)
				frequencyPenalty = &penaltyFloat32
			}
		}

		// Max Tokens - try both OpenAI-style and Gemini-style parameter names
		if mt, ok := params["max_tokens"]; ok {
			switch v := mt.(type) {
			case int32:
				maxTokens = &v
			case int:
				maxInt32 := int32(v)
				maxTokens = &maxInt32
			case int64:
				maxInt32 := int32(v)
				maxTokens = &maxInt32
			case float64:
				maxInt32 := int32(v)
				maxTokens = &maxInt32
			}
		} else if mt, ok := params["max_output_tokens"]; ok {
			// Try the Gemini-style parameter name as a fallback
			switch v := mt.(type) {
			case int32:
				maxTokens = &v
			case int:
				maxInt32 := int32(v)
				maxTokens = &maxInt32
			case int64:
				maxInt32 := int32(v)
				maxTokens = &maxInt32
			case float64:
				maxInt32 := int32(v)
				maxTokens = &maxInt32
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

	// Build the request body using local variables instead of receiver fields
	requestBody := ChatCompletionRequest{
		Model:            c.modelID,
		Messages:         messages,
		Temperature:      temperature,
		TopP:             topP,
		FrequencyPenalty: frequencyPenalty,
		PresencePenalty:  presencePenalty,
		MaxTokens:        maxTokens,
		Stream:           false, // Non-streaming implementation for initial version
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		// Create a categorized error for request marshaling failures
		return nil, CreateAPIError(
			llm.CategoryInvalidRequest,
			"Failed to prepare request to OpenRouter API",
			err,
			fmt.Sprintf("JSON marshal error: %v", err),
		)
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("%s/chat/completions", c.apiEndpoint)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// Create a categorized error for HTTP request creation failures
		return nil, CreateAPIError(
			llm.CategoryNetwork,
			"Failed to create HTTP request to OpenRouter API",
			err,
			fmt.Sprintf("Request creation error: %v", err),
		)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Execute the request
	if c.logger != nil {
		c.logger.Debug("Sending request to OpenRouter API: %s", SanitizeURL(apiURL))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check for context cancellation
		category := llm.CategoryNetwork
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			category = llm.CategoryCancelled
		}

		// Create a categorized error for HTTP execution failures
		return nil, CreateAPIError(
			category,
			"Failed to connect to OpenRouter API",
			err,
			fmt.Sprintf("HTTP error: %v", err),
		)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && c.logger != nil {
			c.logger.Warn("Failed to close response body: %v", closeErr)
		}
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, CreateAPIError(
			llm.CategoryNetwork,
			"Failed to read response from OpenRouter API",
			err,
			fmt.Sprintf("Response read error: %v", err),
		)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		// Create a categorized API error using the FormatAPIError function
		apiErr := FormatAPIError(
			fmt.Errorf("OpenRouter API returned non-200 status code: %d", resp.StatusCode),
			resp.StatusCode,
			body,
		)

		// Try to parse the response for any additional information
		var usageInfo *ChatCompletionUsage

		// Attempt to extract usage information and possibly finish reason if available
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			// Check if there's usage information
			if usage, ok := errorResponse["usage"].(map[string]interface{}); ok {
				usageInfo = &ChatCompletionUsage{}
				if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
					usageInfo.PromptTokens = int(promptTokens)
				}
				if completionTokens, ok := usage["completion_tokens"].(float64); ok {
					usageInfo.CompletionTokens = int(completionTokens)
				}
				if totalTokens, ok := usage["total_tokens"].(float64); ok {
					usageInfo.TotalTokens = int(totalTokens)
				}
			}

			// Check if there's a finish reason and add to error details if found
			if choices, ok := errorResponse["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if reason, ok := choice["finish_reason"].(string); ok && apiErr != nil {
						apiErr.Details += fmt.Sprintf(" (Finish reason: %s)", reason)
					}
				}
			}
		}

		// If we have token usage information, include it in debug details
		if usageInfo != nil && apiErr != nil {
			apiErr.Details += fmt.Sprintf(" (Token usage: %d prompt, %d completion, %d total)",
				usageInfo.PromptTokens, usageInfo.CompletionTokens, usageInfo.TotalTokens)
		}

		return nil, apiErr
	}

	// Parse the response
	var completionResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &completionResponse); err != nil {
		return nil, CreateAPIError(
			llm.CategoryServer,
			"Failed to parse response from OpenRouter API",
			err,
			fmt.Sprintf("JSON unmarshal error: %v, Body: %s", err, truncateString(string(body), 200)),
		)
	}

	// Validate response structure
	if len(completionResponse.Choices) == 0 {
		return nil, CreateAPIError(
			llm.CategoryServer,
			"OpenRouter API returned an empty response",
			fmt.Errorf("no completion choices in response"),
			fmt.Sprintf("Response contained zero choices: %s", truncateString(string(body), 200)),
		)
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

// CountTokens counts the tokens in the given prompt using tiktoken
func (c *openrouterClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	// Always use cl100k_base for modern models
	encoding := "cl100k_base"

	// Get the tokenizer for the appropriate encoding
	tokenizer, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, CreateAPIError(
			llm.CategoryInvalidRequest,
			fmt.Sprintf("Failed to get encoding for model %s: %v", c.modelID, err),
			err,
			fmt.Sprintf("Tiktoken error: %v", err),
		)
	}

	// Encode the text to get the token count
	tokens := tokenizer.Encode(prompt, nil, nil)
	count := len(tokens)

	if c.logger != nil {
		c.logger.Debug("Counted %d tokens for model %s using encoding %s", count, c.modelID, encoding)
	}

	// Return the token count as a ProviderTokenCount struct
	return &llm.ProviderTokenCount{
		Total: int32(count),
	}, nil
}

// GetModelInfo retrieves information about the current model
func (c *openrouterClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	// Create base model info with the model name
	modelInfo := &llm.ProviderModelInfo{
		Name: c.modelID,
	}

	// Get provider and model name from the OpenRouter model ID format
	// which is typically "provider/model" or "provider/org/model"
	parts := strings.Split(strings.ToLower(c.modelID), "/")

	if len(parts) < 2 {
		// If model ID is not in the expected format, use conservative defaults
		modelInfo.InputTokenLimit = 4096
		modelInfo.OutputTokenLimit = 2048

		if c.logger != nil {
			c.logger.Warn("Model ID '%s' does not have expected format 'provider/model' or 'provider/organization/model'. Using conservative token limits.", c.modelID)
		}

		return modelInfo, nil
	}

	provider := parts[0]

	// Default values for most models
	modelInfo.InputTokenLimit = 8192  // 8K input as a default
	modelInfo.OutputTokenLimit = 2048 // 2K output as a default

	// Set token limits based on known provider and model patterns
	switch provider {
	case "anthropic":
		// Claude models
		if strings.Contains(c.modelID, "claude-3-opus") || strings.Contains(c.modelID, "claude-3-sonnet") {
			modelInfo.InputTokenLimit = 200000 // 200K for Claude 3 Opus/Sonnet
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "claude-3-haiku") {
			modelInfo.InputTokenLimit = 200000 // 200K for Claude 3 Haiku
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "claude-2") {
			modelInfo.InputTokenLimit = 100000 // 100K for Claude 2
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "claude-instant") {
			modelInfo.InputTokenLimit = 100000 // 100K for Claude Instant
			modelInfo.OutputTokenLimit = 4096
		} else {
			// Default for other Claude models
			modelInfo.InputTokenLimit = 100000
			modelInfo.OutputTokenLimit = 4096
		}

	case "openai":
		// OpenAI models
		if strings.Contains(c.modelID, "gpt-4o") || strings.Contains(c.modelID, "o4") {
			modelInfo.InputTokenLimit = 128000 // 128K for GPT-4o
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "gpt-4-turbo") {
			modelInfo.InputTokenLimit = 128000 // 128K for GPT-4 Turbo
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "gpt-4.1") {
			modelInfo.InputTokenLimit = 1000000 // 1M tokens for GPT-4.1 series
			modelInfo.OutputTokenLimit = 32768
		} else if strings.Contains(c.modelID, "gpt-4-32k") {
			modelInfo.InputTokenLimit = 32768 // 32K for GPT-4-32k
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "gpt-4") {
			modelInfo.InputTokenLimit = 8192 // 8K for GPT-4
			modelInfo.OutputTokenLimit = 2048
		} else if strings.Contains(c.modelID, "gpt-3.5-turbo-16k") {
			modelInfo.InputTokenLimit = 16385 // 16K for GPT-3.5 Turbo 16K
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "gpt-3.5-turbo") {
			modelInfo.InputTokenLimit = 16385 // 16K for GPT-3.5 Turbo
			modelInfo.OutputTokenLimit = 4096
		} else {
			// Default for other OpenAI models
			modelInfo.InputTokenLimit = 8192
			modelInfo.OutputTokenLimit = 2048
		}

	case "google":
		// Google/Gemini models
		if strings.Contains(c.modelID, "gemini-1.5") {
			modelInfo.InputTokenLimit = 1000000 // 1M tokens for Gemini 1.5
			modelInfo.OutputTokenLimit = 8192
		} else if strings.Contains(c.modelID, "gemini-1.0") {
			modelInfo.InputTokenLimit = 32768 // 32K for Gemini 1.0
			modelInfo.OutputTokenLimit = 8192
		} else if strings.Contains(c.modelID, "palm") {
			modelInfo.InputTokenLimit = 8192
			modelInfo.OutputTokenLimit = 1024
		} else {
			// Default for other Google models
			modelInfo.InputTokenLimit = 32768
			modelInfo.OutputTokenLimit = 4096
		}
	}

	if c.logger != nil {
		c.logger.Debug("GetModelInfo for '%s': using provider '%s', input limit = %d, output limit = %d",
			c.modelID, provider, modelInfo.InputTokenLimit, modelInfo.OutputTokenLimit)
	}

	return modelInfo, nil
}

// GetModelName returns the name of the model being used
func (c *openrouterClient) GetModelName() string {
	return c.modelID
}

// Close releases resources used by the client
func (c *openrouterClient) Close() error {
	// For a standard HTTP client, no explicit cleanup is required
	return nil
}

// Helper methods for parameter configuration

// SetTemperature sets the default temperature parameter
// Note: This setting only applies as a default. Request-specific parameters
// passed to GenerateContent will override this value without modifying it.
func (c *openrouterClient) SetTemperature(temp float32) {
	localTemp := temp
	c.temperature = &localTemp
}

// SetTopP sets the default top_p parameter
// Note: This setting only applies as a default. Request-specific parameters
// passed to GenerateContent will override this value without modifying it.
func (c *openrouterClient) SetTopP(topP float32) {
	localTopP := topP
	c.topP = &localTopP
}

// SetMaxTokens sets the default max_tokens parameter
// Note: This setting only applies as a default. Request-specific parameters
// passed to GenerateContent will override this value without modifying it.
func (c *openrouterClient) SetMaxTokens(tokens int32) {
	localTokens := tokens
	c.maxTokens = &localTokens
}

// SetPresencePenalty sets the default presence_penalty parameter
// Note: This setting only applies as a default. Request-specific parameters
// passed to GenerateContent will override this value without modifying it.
func (c *openrouterClient) SetPresencePenalty(penalty float32) {
	localPenalty := penalty
	c.presencePenalty = &localPenalty
}

// SetFrequencyPenalty sets the default frequency_penalty parameter
// Note: This setting only applies as a default. Request-specific parameters
// passed to GenerateContent will override this value without modifying it.
func (c *openrouterClient) SetFrequencyPenalty(penalty float32) {
	localPenalty := penalty
	c.frequencyPenalty = &localPenalty
}

// truncateString truncates a string to the specified length and adds an ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
