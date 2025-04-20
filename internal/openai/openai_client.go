// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"context"
	"errors"
	"fmt"

	"github.com/phrazzld/architect/internal/llm"
)

// openaiClient implements the llm.LLMClient interface for OpenAI
type openaiClient struct {
	modelName string
	// Parameters for the OpenAI API
	temperature      *float32
	topP             *float32
	presencePenalty  *float32
	frequencyPenalty *float32
	maxTokens        *int32
	reasoningEffort  *string // For O-series model reasoning parameter
}

// NewClient creates a new OpenAI client that implements the llm.LLMClient interface
// This implementation is simplified after removing token functionality
func NewClient(modelName string) (llm.LLMClient, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	// API key validation is now handled by the provider, not the client
	// Client only accepts modelName since we've removed token-related functionality

	return &openaiClient{
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

	// This is a simplified implementation due to removal of token functionality
	// In a real implementation, this would call the OpenAI API

	// Apply parameters from request
	if params != nil {
		applyOpenAIParameters(c, params)
	}

	// In a real implementation, this would call the OpenAI API
	// For now, just return a mock response after token functionality was removed
	content := fmt.Sprintf("OpenAI API response for prompt: %s", getShortSummary(prompt))

	// Build result
	result := &llm.ProviderResult{
		Content:      content,
		FinishReason: "stop",
		Truncated:    false,
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

// Helper function to get a short summary of a prompt for logging
func getShortSummary(text string) string {
	if len(text) <= 30 {
		return text
	}
	return text[:27] + "..."
}

// Apply parameters from map to OpenAI client
func applyOpenAIParameters(client *openaiClient, params map[string]interface{}) {
	// Apply standard parameters
	if temp, ok := params["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			floatVal := float32(v)
			client.temperature = &floatVal
		} else if v, ok := temp.(float32); ok {
			client.temperature = &v
		}
	}

	if topP, ok := params["top_p"]; ok {
		if v, ok := topP.(float64); ok {
			floatVal := float32(v)
			client.topP = &floatVal
		} else if v, ok := topP.(float32); ok {
			client.topP = &v
		}
	}

	if presencePenalty, ok := params["presence_penalty"]; ok {
		if v, ok := presencePenalty.(float64); ok {
			floatVal := float32(v)
			client.presencePenalty = &floatVal
		} else if v, ok := presencePenalty.(float32); ok {
			client.presencePenalty = &v
		}
	}

	if frequencyPenalty, ok := params["frequency_penalty"]; ok {
		if v, ok := frequencyPenalty.(float64); ok {
			floatVal := float32(v)
			client.frequencyPenalty = &floatVal
		} else if v, ok := frequencyPenalty.(float32); ok {
			client.frequencyPenalty = &v
		}
	}

	if maxTokens, ok := params["max_tokens"]; ok {
		if v, ok := maxTokens.(int); ok {
			intVal := int32(v)
			client.maxTokens = &intVal
		} else if v, ok := maxTokens.(int32); ok {
			client.maxTokens = &v
		} else if v, ok := maxTokens.(float64); ok {
			intVal := int32(v)
			client.maxTokens = &intVal
		}
	}

	// Handle special parameters for newer models
	if reasoningEffort, ok := params["reasoning_effort"]; ok {
		if v, ok := reasoningEffort.(string); ok {
			client.reasoningEffort = &v
		}
	}
}
