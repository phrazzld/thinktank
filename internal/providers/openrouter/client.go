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
		// Create a categorized error for request marshaling failures
		return nil, &APIError{
			Original:   err,
			Type:       ErrorTypeInvalidRequest,
			Message:    "Failed to prepare request to OpenRouter API",
			Suggestion: "This is likely an internal error. Please check your input parameters for any invalid values.",
			Details:    fmt.Sprintf("JSON marshal error: %v", err),
		}
	}

	// Construct the API URL
	apiURL := fmt.Sprintf("%s/chat/completions", c.apiEndpoint)

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// Create a categorized error for HTTP request creation failures
		return nil, &APIError{
			Original:   err,
			Type:       ErrorTypeNetwork,
			Message:    "Failed to create HTTP request to OpenRouter API",
			Suggestion: "This could be due to an invalid API endpoint or network configuration issue.",
			Details:    fmt.Sprintf("Request creation error: %v", err),
		}
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
		// Determine error type based on the actual error
		errType := ErrorTypeNetwork

		// Check for context cancellation
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			errType = ErrorTypeCancelled
		}

		// Create a categorized error for HTTP execution failures
		return nil, &APIError{
			Original:   err,
			Type:       errType,
			Message:    "Failed to connect to OpenRouter API",
			Suggestion: "Check your internet connection. If the issue persists, the OpenRouter service may be experiencing issues.",
			Details:    fmt.Sprintf("HTTP error: %v", err),
		}
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && c.logger != nil {
			c.logger.Warn("Failed to close response body: %v", closeErr)
		}
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			Original:   err,
			Type:       ErrorTypeNetwork,
			Message:    "Failed to read response from OpenRouter API",
			Suggestion: "This is likely a temporary network issue. Try again later.",
			Details:    fmt.Sprintf("Response read error: %v", err),
		}
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
		return nil, &APIError{
			Original:   err,
			Type:       ErrorTypeServer,
			Message:    "Failed to parse response from OpenRouter API",
			Suggestion: "The API returned an unexpected response format. This might be due to an API change or temporary issue with OpenRouter.",
			Details:    fmt.Sprintf("JSON unmarshal error: %v, Body: %s", err, truncateString(string(body), 200)),
		}
	}

	// Validate response structure
	if len(completionResponse.Choices) == 0 {
		return nil, &APIError{
			Original:   fmt.Errorf("no completion choices in response"),
			Type:       ErrorTypeServer,
			Message:    "OpenRouter API returned an empty response",
			Suggestion: "This could be a temporary issue with the OpenRouter service or the underlying model provider. Try again later.",
			Details:    fmt.Sprintf("Response contained zero choices: %s", truncateString(string(body), 200)),
		}
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
	// Use tiktoken to count tokens
	encoding := getEncodingForModelID(c.modelID)

	// Get the tokenizer for the appropriate encoding
	tokenizer, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, &APIError{
			Original:   err,
			Type:       ErrorTypeInvalidRequest,
			Message:    fmt.Sprintf("Failed to get encoding for model %s: %v", c.modelID, err),
			Suggestion: "This could be due to an unsupported model. Try using a different model.",
			Details:    fmt.Sprintf("Tiktoken error: %v", err),
		}
	}

	// Encode the text to get the token count
	tokens := tokenizer.Encode(prompt, nil, nil)
	count := len(tokens)

	if c.logger != nil {
		c.logger.Debug("Counted %d tokens for model %s", count, c.modelID)
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

	// The registry is the source of truth for token limits
	// This function provides reasonable defaults for when registry data is unavailable
	// or for direct client usage outside of the TokenManager

	// Determine appropriate token limits based on the model ID
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

	case "meta":
		// Meta models (Llama)
		if strings.Contains(c.modelID, "llama-3-70b") {
			modelInfo.InputTokenLimit = 8192
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "llama-3") {
			modelInfo.InputTokenLimit = 8192
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "llama-2-70b") {
			modelInfo.InputTokenLimit = 4096
			modelInfo.OutputTokenLimit = 4096
		} else {
			// Default for other Meta models
			modelInfo.InputTokenLimit = 4096
			modelInfo.OutputTokenLimit = 4096
		}

	case "mistral":
		// Mistral models
		if strings.Contains(c.modelID, "mixtral") || strings.Contains(c.modelID, "mixtral-8x7b") {
			modelInfo.InputTokenLimit = 32768
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "mistral-medium") {
			modelInfo.InputTokenLimit = 32768
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "mistral-small") {
			modelInfo.InputTokenLimit = 32768
			modelInfo.OutputTokenLimit = 4096
		} else if strings.Contains(c.modelID, "mistral-tiny") {
			modelInfo.InputTokenLimit = 32768
			modelInfo.OutputTokenLimit = 4096
		} else {
			// Default for other Mistral models
			modelInfo.InputTokenLimit = 32768
			modelInfo.OutputTokenLimit = 4096
		}

	case "cohere":
		// Cohere models
		modelInfo.InputTokenLimit = 32768
		modelInfo.OutputTokenLimit = 4096

	case "deepseek":
		// DeepSeek models
		modelInfo.InputTokenLimit = 32768
		modelInfo.OutputTokenLimit = 4096

	case "perplexity":
		// Perplexity models
		modelInfo.InputTokenLimit = 32768
		modelInfo.OutputTokenLimit = 4096

	default:
		// For unknown providers, use more conservative defaults
		modelInfo.InputTokenLimit = 8192  // 8K input as a safe default
		modelInfo.OutputTokenLimit = 2048 // 2K output as a safe default

		if c.logger != nil {
			c.logger.Debug("Unknown provider '%s' in model ID. Using conservative token limits.", provider)
		}
	}

	if c.logger != nil {
		c.logger.Debug("GetModelInfo for '%s': input limit = %d, output limit = %d",
			c.modelID, modelInfo.InputTokenLimit, modelInfo.OutputTokenLimit)
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

// truncateString truncates a string to the specified length and adds an ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getEncodingForModelID determines the appropriate tiktoken encoding
// for a given OpenRouter model ID (e.g., "anthropic/claude-3-sonnet-20240229")
func getEncodingForModelID(modelID string) string {
	// Convert to lowercase for consistent matching
	modelID = strings.ToLower(modelID)

	// Extract provider and model name from the OpenRouter model ID format
	// which is typically "provider/model" or "provider/org/model"
	parts := strings.Split(modelID, "/")
	if len(parts) < 2 {
		// If not in the expected format, use most modern encoding as fallback
		return "cl100k_base"
	}

	provider := parts[0]

	// Match by provider
	switch provider {
	case "openai", "openai-compatible", "perplexity", "mistral", "cohere":
		// Most modern models use cl100k_base
		return "cl100k_base"
	case "anthropic":
		// Claude models use cl100k_base
		return "cl100k_base"
	case "google":
		// Gemini models use cl100k_base
		return "cl100k_base"
	case "meta":
		// Llama models use cl100k_base
		return "cl100k_base"
	}

	// For specific model patterns across providers
	if strings.Contains(modelID, "claude") ||
		strings.Contains(modelID, "gpt") ||
		strings.Contains(modelID, "gpt-4") ||
		strings.Contains(modelID, "gpt-3.5") ||
		strings.Contains(modelID, "llama") ||
		strings.Contains(modelID, "gemini") ||
		strings.Contains(modelID, "mixtral") ||
		strings.Contains(modelID, "text-embedding") {
		return "cl100k_base"
	}

	// For older OpenAI models
	if strings.Contains(modelID, "davinci") ||
		strings.Contains(modelID, "curie") ||
		strings.Contains(modelID, "babbage") ||
		strings.Contains(modelID, "ada") {
		return "p50k_base"
	}

	// Default to the most modern encoding for unknown models
	return "cl100k_base"
}
