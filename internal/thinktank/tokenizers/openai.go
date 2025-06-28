package tokenizers

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

// OpenAITokenizer provides accurate token counting for OpenAI models using tiktoken.
type OpenAITokenizer struct {
	// encoderCache stores initialized tiktoken encoders by encoding name
	encoderCache sync.Map

	// modelEncodings maps OpenAI model names to their tiktoken encoding names
	modelEncodings map[string]string
}

// NewOpenAITokenizer creates a new OpenAI tokenizer with model encoding mappings.
func NewOpenAITokenizer() *OpenAITokenizer {
	return &OpenAITokenizer{
		modelEncodings: map[string]string{
			// GPT-4 models use cl100k_base encoding
			"gpt-4.1": "cl100k_base",

			// o4-mini and o3 use o200k_base encoding (GPT-4o family)
			"o4-mini": "o200k_base",
			"o3":      "o200k_base",

			// Add more OpenAI models as needed
			"gpt-4":       "cl100k_base",
			"gpt-4o":      "o200k_base",
			"gpt-4o-mini": "o200k_base",
		},
	}
}

// CountTokens implements AccurateTokenCounter for OpenAI models.
func (t *OpenAITokenizer) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if text == "" {
		return 0, nil
	}

	encoding, err := t.GetEncoding(modelName)
	if err != nil {
		return 0, NewTokenizerError("openai", modelName, "failed to get encoding", err)
	}

	encoder, err := t.getEncoder(encoding)
	if err != nil {
		return 0, NewTokenizerError("openai", modelName, "failed to get encoder", err)
	}

	// Use tiktoken to get accurate token count
	tokens := encoder.Encode(text, nil, nil)
	return len(tokens), nil
}

// SupportsModel returns true if the model is supported by OpenAI tokenizer.
func (t *OpenAITokenizer) SupportsModel(modelName string) bool {
	_, supported := t.modelEncodings[modelName]
	return supported
}

// GetEncoding returns the tiktoken encoding name for the given OpenAI model.
func (t *OpenAITokenizer) GetEncoding(modelName string) (string, error) {
	encoding, ok := t.modelEncodings[modelName]
	if !ok {
		return "", fmt.Errorf("unsupported OpenAI model: %s", modelName)
	}
	return encoding, nil
}

// getEncoder returns a cached tiktoken encoder or creates a new one.
// Uses lazy loading to avoid 4MB vocabulary initialization at startup.
func (t *OpenAITokenizer) getEncoder(encoding string) (*tiktoken.Tiktoken, error) {
	// Check cache first
	if cached, ok := t.encoderCache.Load(encoding); ok {
		return cached.(*tiktoken.Tiktoken), nil
	}

	// Create new encoder with lazy loading
	encoder, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tiktoken encoding %s: %w", encoding, err)
	}

	// Cache the encoder for future use
	t.encoderCache.Store(encoding, encoder)
	return encoder, nil
}

// ClearCache clears all cached encoders to free memory.
func (t *OpenAITokenizer) ClearCache() {
	t.encoderCache.Range(func(key, value interface{}) bool {
		t.encoderCache.Delete(key)
		return true
	})
}
