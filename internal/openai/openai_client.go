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

// realTokenizer implements tokenizerAPI
type realTokenizer struct{}

// countTokens implements the tokenizerAPI interface using tiktoken
func (t *realTokenizer) countTokens(text string, model string) (int, error) {
	encoding := getEncodingForModel(model)

	tokenizer, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return 0, &APIError{
			Original: err,
			Type:     ErrorTypeInvalidRequest,
			Message:  fmt.Sprintf("Failed to get encoding for model %s: %v", model, err),
		}
	}

	tokens := tokenizer.Encode(text, nil, nil)
	return len(tokens), nil
}

// getEncodingForModel returns the appropriate encoding name for a given OpenAI model
func getEncodingForModel(model string) string {
	// Map model names to tiktoken encoding names
	// Default to cl100k_base for newer models (including gpt-4 and gpt-3.5-turbo series)
	model = strings.ToLower(model)

	if strings.HasPrefix(model, "gpt-4") {
		return "cl100k_base"
	}

	if strings.HasPrefix(model, "gpt-3.5-turbo") {
		return "cl100k_base"
	}

	if strings.HasPrefix(model, "text-embedding-ada-002") {
		return "cl100k_base"
	}

	// Fallback to p50k_base encoding for other/older models
	return "p50k_base"
}

// NewClient creates a new OpenAI client that implements the llm.LLMClient interface
func NewClient(modelName string) (llm.LLMClient, error) {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
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
		"gpt-4o": {
			inputTokenLimit:  128000,
			outputTokenLimit: 4096,
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
	}
}

// GenerateContent sends a text prompt to OpenAI and returns the generated content
func (c *openaiClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	// Create chat completion request with user prompt
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}

	// Call OpenAI API
	completion, err := c.api.createChatCompletion(ctx, messages, c.modelName)
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

// CountTokens counts the tokens in a given prompt
func (c *openaiClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	count, err := c.tokenizer.countTokens(prompt, c.modelName)
	if err != nil {
		return nil, fmt.Errorf("token counting error: %w", err)
	}

	return &llm.ProviderTokenCount{
		Total: int32(count),
	}, nil
}

// GetModelInfo retrieves information about the current model
func (c *openaiClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	// Get model info from cache or use defaults
	info, ok := c.modelLimits[c.modelName]
	if !ok {
		// Use conservative defaults for unknown models
		info = &modelInfo{
			inputTokenLimit:  4096, // Conservative default
			outputTokenLimit: 2048, // Conservative default
		}
	}

	return &llm.ProviderModelInfo{
		Name:             c.modelName,
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
