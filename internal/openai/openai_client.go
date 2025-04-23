// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/phrazzld/thinktank/internal/llm"
)

// openaiAPI defines the operations we need from the OpenAI client
type openaiAPI interface {
	createChatCompletion(ctx context.Context, model string, prompt string, systemPrompt string) (*openai.ChatCompletion, error)
	createChatCompletionWithParams(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

// openaiClient implements the llm.LLMClient interface for OpenAI
type openaiClient struct {
	api       openaiAPI
	modelName string
	// Parameters for the OpenAI API
	temperature      *float32
	topP             *float32
	presencePenalty  *float32
	frequencyPenalty *float32
	maxTokens        *int32
}

// realOpenAIAPI implements openaiAPI using the real OpenAI client
type realOpenAIAPI struct {
	client openai.Client
}

// createChatCompletion implements the openaiAPI interface method using the real OpenAI client
func (api *realOpenAIAPI) createChatCompletion(ctx context.Context, model string, prompt string, systemPrompt string) (*openai.ChatCompletion, error) {
	messages := []openai.ChatCompletionMessageParamUnion{}

	// Add system message if provided
	if systemPrompt != "" {
		messages = append(messages, openai.SystemMessage(systemPrompt))
	}

	// Add user message
	messages = append(messages, openai.UserMessage(prompt))

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	}

	completion, err := api.client.Chat.Completions.New(ctx, params)
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

// NewClient creates a new OpenAI client that implements the llm.LLMClient interface
func NewClient(modelName string) (llm.LLMClient, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	// Create the OpenAI client - API key is set in environment variable
	client := openai.NewClient()

	// Create the real API implementation
	api := &realOpenAIAPI{
		client: client,
	}

	// Create the client
	return &openaiClient{
		api:       api,
		modelName: modelName,
	}, nil
}

// GenerateContent implements the LLMClient interface
func (c *openaiClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if prompt == "" {
		return nil, CreateAPIError(
			llm.CategoryInvalidRequest,
			"Empty prompt",
			errors.New("prompt cannot be empty"),
			"Please provide a valid prompt",
		)
	}

	// Extract system prompt if present
	systemPrompt := ""
	if strings.Contains(prompt, "<system>") && strings.Contains(prompt, "</system>") {
		systemPrompt = extractTag(prompt, "system")
		if systemPrompt != "" {
			// Remove system content from prompt
			prompt = removeTag(prompt, "system")
		}
	} else if strings.Contains(prompt, "<s>") && strings.Contains(prompt, "</s>") {
		// For backward compatibility with <s> tags
		systemPrompt = extractTag(prompt, "s")
		if systemPrompt != "" {
			// Remove system content from prompt
			prompt = removeTag(prompt, "s")
		}
	}

	// Build the messages array
	messages := []openai.ChatCompletionMessageParamUnion{}

	// Add system message if present
	if systemPrompt != "" {
		messages = append(messages, openai.SystemMessage(systemPrompt))
	}

	// Add user message
	messages = append(messages, openai.UserMessage(prompt))

	// Build request parameters
	requestParams := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    c.modelName,
	}

	// Apply fixed parameters from client
	if c.temperature != nil {
		temp := float64(*c.temperature)
		requestParams.Temperature = openai.Float(temp)
	}
	if c.topP != nil {
		topP := float64(*c.topP)
		requestParams.TopP = openai.Float(topP)
	}
	if c.presencePenalty != nil {
		penalty := float64(*c.presencePenalty)
		requestParams.PresencePenalty = openai.Float(penalty)
	}
	if c.frequencyPenalty != nil {
		penalty := float64(*c.frequencyPenalty)
		requestParams.FrequencyPenalty = openai.Float(penalty)
	}
	if c.maxTokens != nil {
		tokens := int64(*c.maxTokens)
		requestParams.MaxTokens = openai.Int(tokens)
	}
	// Skip reasoning effort parameter since we need to understand the API better

	// Override with parameters from request
	if params != nil {
		applyOpenAIParameters(&requestParams, params)
	}

	// Make the API call
	completion, err := c.api.createChatCompletionWithParams(ctx, requestParams)
	if err != nil {
		return nil, err
	}

	// Handle empty choices
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no completions returned from OpenAI API")
	}

	// Extract content
	choice := completion.Choices[0]
	content := choice.Message.Content

	// Get finish reason
	finishReason := choice.FinishReason

	// Build result
	result := &llm.ProviderResult{
		Content:      content,
		FinishReason: finishReason,
		Truncated:    finishReason == "length",
	}

	return result, nil
}

// GetModelName returns the model name
func (c *openaiClient) GetModelName() string {
	return c.modelName
}

// Close cleans up any resources (nothing to clean up for OpenAI)
func (c *openaiClient) Close() error {
	return nil
}

// SetTemperature sets the temperature parameter
func (c *openaiClient) SetTemperature(temp float32) {
	c.temperature = &temp
}

// SetTopP sets the top_p parameter
func (c *openaiClient) SetTopP(topP float32) {
	c.topP = &topP
}

// SetMaxTokens sets the max_tokens parameter
func (c *openaiClient) SetMaxTokens(maxTokens int32) {
	c.maxTokens = &maxTokens
}

// SetFrequencyPenalty sets the frequency_penalty parameter
func (c *openaiClient) SetFrequencyPenalty(penalty float32) {
	c.frequencyPenalty = &penalty
}

// SetPresencePenalty sets the presence_penalty parameter
func (c *openaiClient) SetPresencePenalty(penalty float32) {
	c.presencePenalty = &penalty
}

// Helper function to extract content between XML tags
func extractTag(text, tag string) string {
	startTag := fmt.Sprintf("<%s>", tag)
	endTag := fmt.Sprintf("</%s>", tag)

	startIdx := strings.Index(text, startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)

	endIdx := strings.Index(text[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return text[startIdx : startIdx+endIdx]
}

// Helper function to remove content between XML tags including the tags
func removeTag(text, tag string) string {
	startTag := fmt.Sprintf("<%s>", tag)
	endTag := fmt.Sprintf("</%s>", tag)

	startIdx := strings.Index(text, startTag)
	if startIdx == -1 {
		return text
	}

	endIdx := strings.Index(text[startIdx:], endTag)
	if endIdx == -1 {
		return text
	}
	endIdx = startIdx + endIdx + len(endTag)

	return text[:startIdx] + text[endIdx:]
}

// Apply parameters from map to OpenAI request parameters
func applyOpenAIParameters(params *openai.ChatCompletionNewParams, customParams map[string]interface{}) {
	// Apply standard parameters
	if temp, ok := customParams["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			params.Temperature = openai.Float(v)
		} else if v, ok := temp.(float32); ok {
			params.Temperature = openai.Float(float64(v))
		}
	}

	if topP, ok := customParams["top_p"]; ok {
		if v, ok := topP.(float64); ok {
			params.TopP = openai.Float(v)
		} else if v, ok := topP.(float32); ok {
			params.TopP = openai.Float(float64(v))
		}
	}

	if presencePenalty, ok := customParams["presence_penalty"]; ok {
		if v, ok := presencePenalty.(float64); ok {
			params.PresencePenalty = openai.Float(v)
		} else if v, ok := presencePenalty.(float32); ok {
			params.PresencePenalty = openai.Float(float64(v))
		}
	}

	if frequencyPenalty, ok := customParams["frequency_penalty"]; ok {
		if v, ok := frequencyPenalty.(float64); ok {
			params.FrequencyPenalty = openai.Float(v)
		} else if v, ok := frequencyPenalty.(float32); ok {
			params.FrequencyPenalty = openai.Float(float64(v))
		}
	}

	if maxTokens, ok := customParams["max_tokens"]; ok {
		if v, ok := maxTokens.(int); ok {
			params.MaxTokens = openai.Int(int64(v))
		} else if v, ok := maxTokens.(int32); ok {
			params.MaxTokens = openai.Int(int64(v))
		} else if v, ok := maxTokens.(float64); ok {
			params.MaxTokens = openai.Int(int64(v))
		}
	}

	// Skip reasoning effort for now
	// if reasoningEffort, ok := customParams["reasoning_effort"]; ok {
	// 	if v, ok := reasoningEffort.(string); ok {
	// 		// TODO: Figure out the right way to set this
	// 	}
	// }
}
