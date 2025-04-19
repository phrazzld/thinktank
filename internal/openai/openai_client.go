// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/phrazzld/architect/internal/llm"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

// openaiAPI defines the operations we need from the OpenAI client
type openaiAPI interface {
	createChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error)
	createChatCompletionWithParams(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

// tokenizerAPI defines the operations we need for token counting
type tokenizerAPI interface {
	countTokens(text string, model string) (int, error)
}

// openaiClient implements the llm.LLMClient interface for OpenAI
type openaiClient struct {
	api         openaiAPI
	tokenizer   tokenizerAPI
	modelName   string
	modelLimits map[string]*modelInfo // Cache of model token limits
	// Parameters for the OpenAI API
	temperature      *float64
	topP             *float64
	presencePenalty  *float64
	frequencyPenalty *float64
	maxTokens        *int
	reasoningEffort  *string // For O-series model reasoning parameter
}

// Internal model info struct
type modelInfo struct {
	inputTokenLimit  int32
	outputTokenLimit int32
}

// realOpenAIAPI implements openaiAPI using the real OpenAI client
type realOpenAIAPI struct {
	client openai.Client
}

// createChatCompletion implements the openaiAPI interface method using the real OpenAI client
func (api *realOpenAIAPI) createChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
	completion, err := api.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	})
	if err != nil {
		// Format the error using our error handling
		return nil, FormatAPIError(err, 0)
	}
	return completion, nil
}

// createChatCompletionWithParams makes an API call with specific parameters
func (api *realOpenAIAPI) createChatCompletionWithParams(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	completion, err := api.client.Chat.Completions.New(ctx, params)
	if err != nil {
		// Format the error using our error handling
		return nil, FormatAPIError(err, 0)
	}
	return completion, nil
}

// realTokenizer implements tokenizerAPI
type realTokenizer struct{}

// countTokens implements the tokenizerAPI interface using tiktoken
func (t *realTokenizer) countTokens(text string, model string) (int, error) {
	encoding := getEncodingForModel(model)

	tokenizer, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return 0, CreateAPIError(
			llm.CategoryInvalidRequest,
			fmt.Sprintf("Failed to get encoding for model %s: %v", model, err),
			err,
			"",
		)
	}

	tokens := tokenizer.Encode(text, nil, nil)
	return len(tokens), nil
}

// getEncodingForModel returns the appropriate encoding name for a given OpenAI model
func getEncodingForModel(model string) string {
	// For all modern models, use cl100k_base
	// This will be replaced by registry-based encoding determination in future
	return "cl100k_base"
}

// NewClient creates a new OpenAI client that implements the llm.LLMClient interface
func NewClient(modelName string) (llm.LLMClient, error) {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Trim whitespace from API key
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable not set")
	}

	// Initialize OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Initialize tokenizer
	tokenizer := &realTokenizer{}

	// Create and return client
	return &openaiClient{
		api:         &realOpenAIAPI{client: client},
		tokenizer:   tokenizer,
		modelName:   modelName,
		modelLimits: initializeModelLimits(),
	}, nil
}

// Initialize model limits for common models
func initializeModelLimits() map[string]*modelInfo {
	// Hardcoded limits for common models
	// These values should be verified and may be updated in the future
	return map[string]*modelInfo{
		// GPT-4 series
		"gpt-4": {
			inputTokenLimit:  8192,
			outputTokenLimit: 2048,
		},
		"gpt-4-32k": {
			inputTokenLimit:  32768,
			outputTokenLimit: 4096,
		},
		"gpt-4-turbo": {
			inputTokenLimit:  128000,
			outputTokenLimit: 4096,
		},
		"gpt-4-turbo-2024-04-09": {
			inputTokenLimit:  128000,
			outputTokenLimit: 4096,
		},
		"gpt-4o": {
			inputTokenLimit:  128000,
			outputTokenLimit: 4096,
		},
		"gpt-4.1-mini": {
			inputTokenLimit:  1000000, // 1M tokens
			outputTokenLimit: 32768,
		},
		"gpt-4.1": {
			inputTokenLimit:  1000000, // 1M tokens
			outputTokenLimit: 32768,
		},
		"gpt-4.1-preview": {
			inputTokenLimit:  1000000, // 1M tokens
			outputTokenLimit: 32768,
		},
		"o4-mini": {
			inputTokenLimit:  1000000, // 1M tokens
			outputTokenLimit: 32768,
		},
		"o4": {
			inputTokenLimit:  1000000, // 1M tokens
			outputTokenLimit: 32768,
		},

		// GPT-3.5 series
		"gpt-3.5-turbo": {
			inputTokenLimit:  16385,
			outputTokenLimit: 4096,
		},
		"gpt-3.5-turbo-16k": {
			inputTokenLimit:  16385,
			outputTokenLimit: 4096,
		},
		"gpt-3.5-turbo-0125": {
			inputTokenLimit:  16385,
			outputTokenLimit: 4096,
		},
		"gpt-3.5-turbo-instruct": {
			inputTokenLimit:  16385,
			outputTokenLimit: 4096,
		},
	}
}

// GenerateContent sends a text prompt to OpenAI and returns the generated content
func (c *openaiClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// Apply parameters if provided
	if params != nil {
		// Temperature
		if temp, ok := params["temperature"]; ok {
			switch v := temp.(type) {
			case float64:
				c.temperature = &v
			case float32:
				tempFloat64 := float64(v)
				c.temperature = &tempFloat64
			case int:
				tempFloat64 := float64(v)
				c.temperature = &tempFloat64
			}
		}

		// Top P
		if topP, ok := params["top_p"]; ok {
			switch v := topP.(type) {
			case float64:
				c.topP = &v
			case float32:
				topPFloat64 := float64(v)
				c.topP = &topPFloat64
			case int:
				topPFloat64 := float64(v)
				c.topP = &topPFloat64
			}
		}

		// Presence Penalty
		if penalty, ok := params["presence_penalty"]; ok {
			switch v := penalty.(type) {
			case float64:
				c.presencePenalty = &v
			case float32:
				penaltyFloat64 := float64(v)
				c.presencePenalty = &penaltyFloat64
			case int:
				penaltyFloat64 := float64(v)
				c.presencePenalty = &penaltyFloat64
			}
		}

		// Frequency Penalty
		if penalty, ok := params["frequency_penalty"]; ok {
			switch v := penalty.(type) {
			case float64:
				c.frequencyPenalty = &v
			case float32:
				penaltyFloat64 := float64(v)
				c.frequencyPenalty = &penaltyFloat64
			case int:
				penaltyFloat64 := float64(v)
				c.frequencyPenalty = &penaltyFloat64
			}
		}

		// Max Tokens - try both OpenAI-style and Gemini-style parameter names
		if maxTokens, ok := params["max_tokens"]; ok {
			switch v := maxTokens.(type) {
			case int:
				c.maxTokens = &v
			case int32:
				maxInt := int(v)
				c.maxTokens = &maxInt
			case int64:
				maxInt := int(v)
				c.maxTokens = &maxInt
			case float64:
				maxInt := int(v)
				c.maxTokens = &maxInt
			}
		} else if maxTokens, ok := params["max_output_tokens"]; ok {
			// Try the Gemini-style parameter name as a fallback
			switch v := maxTokens.(type) {
			case int:
				c.maxTokens = &v
			case int32:
				maxInt := int(v)
				c.maxTokens = &maxInt
			case int64:
				maxInt := int(v)
				c.maxTokens = &maxInt
			case float64:
				maxInt := int(v)
				c.maxTokens = &maxInt
			}
		}

		// Handle reasoning parameter for O-series models
		if reasoning, ok := params["reasoning"]; ok {
			// Handle map with effort field
			if reasoningMap, ok := reasoning.(map[string]interface{}); ok {
				if effort, ok := reasoningMap["effort"].(string); ok {
					c.reasoningEffort = &effort
				}
			} else if reasoningMap, ok := reasoning.(map[string]string); ok {
				if effort, ok := reasoningMap["effort"]; ok {
					c.reasoningEffort = &effort
				}
			} else if effortStr, ok := reasoning.(string); ok {
				// Direct string value
				c.reasoningEffort = &effortStr
			}
		}
	}

	// Set default reasoning effort if needed
	// Note: Model-specific defaults will be provided by the registry in the future
	if c.reasoningEffort == nil {
		defaultEffort := "high"
		c.reasoningEffort = &defaultEffort
	}

	// Create chat completion request with user prompt
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}

	// Create a custom implementation that passes the parameters to the OpenAI API
	// Call OpenAI API with the parameters through our customized realOpenAIAPI.createChatCompletion method
	completion, err := c.createChatCompletionWithParams(ctx, messages)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	// Map response to ProviderResult
	if len(completion.Choices) == 0 {
		return nil, errors.New("no completion choices returned")
	}

	content := completion.Choices[0].Message.Content
	finishReason := string(completion.Choices[0].FinishReason)

	// Build safetyInfo if present
	var safetyInfo []llm.Safety

	// Note: OpenAI does not provide safety ratings in the same format as Gemini
	// We could potentially parse any content filter flags here if needed

	return &llm.ProviderResult{
		Content:      content,
		FinishReason: finishReason,
		TokenCount:   int32(completion.Usage.CompletionTokens),
		Truncated:    finishReason == "length",
		SafetyInfo:   safetyInfo,
	}, nil
}

// CountTokens implements the llm.LLMClient interface by counting tokens for a text
func (c *openaiClient) CountTokens(ctx context.Context, text string) (int32, error) {
	count, err := c.tokenizer.countTokens(text, c.modelName)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

// GetModelLimits implements the llm.LLMClient interface
func (c *openaiClient) GetModelLimits(ctx context.Context) (*llm.ModelLimits, error) {
	// Get model limits from cached info
	info, ok := c.modelLimits[c.modelName]
	if !ok {
		// Use default limits if not found
		info = &modelInfo{
			inputTokenLimit:  8192, // Conservative default
			outputTokenLimit: 2048, // Conservative default
		}
	}

	return &llm.ModelLimits{
		InputTokenLimit:  info.inputTokenLimit,
		OutputTokenLimit: info.outputTokenLimit,
	}, nil
}

// GetModelName returns the name of the model being used
func (c *openaiClient) GetModelName() string {
	return c.modelName
}

// Close releases resources used by the client
func (c *openaiClient) Close() error {
	// The OpenAI Go client doesn't require explicit closing
	return nil
}

// createChatCompletionWithParams builds a parameter-aware OpenAI API request
func (c *openaiClient) createChatCompletionWithParams(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	// Create the base params
	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    c.modelName,
	}

	// Validate and apply all optional parameters if they have been set
	if c.temperature != nil {
		// Temperature should be between 0.0 and 2.0 for all models
		// Note: Model-specific restrictions will be handled by the registry in the future
		if *c.temperature < 0.0 || *c.temperature > 2.0 {
			return nil, CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Temperature must be between 0.0 and 2.0, got %f", *c.temperature),
				nil,
				"Set temperature to a value between 0.0 and 2.0",
			)
		}
		params.Temperature = openai.Float(*c.temperature)
	}

	if c.topP != nil {
		// Top_p should be between 0.0 and 1.0
		if *c.topP < 0.0 || *c.topP > 1.0 {
			return nil, CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Top_p must be between 0.0 and 1.0, got %f", *c.topP),
				nil,
				"Set top_p to a value between 0.0 and 1.0",
			)
		}
		params.TopP = openai.Float(*c.topP)
	}

	if c.maxTokens != nil {
		// Max tokens should be positive
		if *c.maxTokens <= 0 {
			return nil, CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Max tokens must be positive, got %d", *c.maxTokens),
				nil,
				"Set max_tokens to a positive value",
			)
		}
		params.MaxTokens = openai.Int(int64(*c.maxTokens))
	}

	if c.frequencyPenalty != nil {
		// Frequency penalty should be between -2.0 and 2.0
		if *c.frequencyPenalty < -2.0 || *c.frequencyPenalty > 2.0 {
			return nil, CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Frequency penalty must be between -2.0 and 2.0, got %f", *c.frequencyPenalty),
				nil,
				"Set frequency_penalty to a value between -2.0 and 2.0",
			)
		}
		params.FrequencyPenalty = openai.Float(*c.frequencyPenalty)
	}

	if c.presencePenalty != nil {
		// Presence penalty should be between -2.0 and 2.0
		if *c.presencePenalty < -2.0 || *c.presencePenalty > 2.0 {
			return nil, CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Presence penalty must be between -2.0 and 2.0, got %f", *c.presencePenalty),
				nil,
				"Set presence_penalty to a value between -2.0 and 2.0",
			)
		}
		params.PresencePenalty = openai.Float(*c.presencePenalty)
	}

	// Apply reasoning parameter if provided
	// Note: Model-specific requirements will be handled by the registry in the future
	if c.reasoningEffort != nil {
		// Validate reasoning effort value
		effort := strings.ToLower(*c.reasoningEffort)
		if effort != "low" && effort != "medium" && effort != "high" {
			return nil, CreateAPIError(
				llm.CategoryInvalidRequest,
				fmt.Sprintf("Reasoning effort must be 'low', 'medium', or 'high', got '%s'", *c.reasoningEffort),
				nil,
				"Set reasoning.effort to 'low', 'medium', or 'high'",
			)
		}

		// Set the reasoning_effort parameter
		params.ReasoningEffort = openai.ReasoningEffort(effort)
	}

	// Call the API with all parameters
	return c.api.createChatCompletionWithParams(ctx, params)
}

// SetTemperature sets the temperature parameter for generations
func (c *openaiClient) SetTemperature(temp float32) {
	tempFloat64 := float64(temp)
	c.temperature = &tempFloat64
}

// SetTopP sets the top_p parameter for generations
func (c *openaiClient) SetTopP(topP float32) {
	topPFloat64 := float64(topP)
	c.topP = &topPFloat64
}

// SetPresencePenalty sets the presence_penalty parameter for generations
func (c *openaiClient) SetPresencePenalty(penalty float32) {
	penaltyFloat64 := float64(penalty)
	c.presencePenalty = &penaltyFloat64
}

// SetFrequencyPenalty sets the frequency_penalty parameter for generations
func (c *openaiClient) SetFrequencyPenalty(penalty float32) {
	penaltyFloat64 := float64(penalty)
	c.frequencyPenalty = &penaltyFloat64
}

// SetMaxTokens sets the max_tokens parameter for generations
func (c *openaiClient) SetMaxTokens(tokens int32) {
	maxInt := int(tokens)
	c.maxTokens = &maxInt
}
