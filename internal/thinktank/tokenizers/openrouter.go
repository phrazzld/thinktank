// Package tokenizers provides accurate token counting implementations for various LLM providers.
package tokenizers

import (
	"context"
)

// OpenRouterTokenizer provides accurate token counting for OpenRouter models.
// Since OpenRouter normalizes all API responses to GPT-4o's o200k_base encoding,
// this tokenizer wraps the existing OpenAI tokenizer with o200k_base.
type OpenRouterTokenizer struct {
	openaiTokenizer *OpenAITokenizer
}

// NewOpenRouterTokenizer creates a new OpenRouter tokenizer.
// Uses composition to leverage the existing OpenAI tokenizer since OpenRouter
// normalizes all responses to o200k_base encoding (used by GPT-4o family).
func NewOpenRouterTokenizer() *OpenRouterTokenizer {
	return &OpenRouterTokenizer{
		openaiTokenizer: NewOpenAITokenizer(),
	}
}

// CountTokens implements AccurateTokenCounter for OpenRouter models.
// All OpenRouter models are normalized to o200k_base encoding, so we delegate
// to the OpenAI tokenizer using a model that uses o200k_base.
func (t *OpenRouterTokenizer) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if text == "" {
		return 0, nil
	}

	// All OpenRouter models are normalized to o200k_base encoding.
	// Use o4-mini which uses o200k_base encoding for consistent tokenization.
	return t.openaiTokenizer.CountTokens(ctx, text, "o4-mini")
}

// SupportsModel returns true if the model name appears to be an OpenRouter model.
// OpenRouter models typically have the format "openrouter/provider/model".
func (t *OpenRouterTokenizer) SupportsModel(modelName string) bool {
	// For now, assume all models are supported since OpenRouter normalizes everything
	// In the future, we could check against the models registry
	return true
}

// GetEncoding returns the tokenizer encoding name for OpenRouter models.
// All OpenRouter models are normalized to o200k_base encoding.
func (t *OpenRouterTokenizer) GetEncoding(modelName string) (string, error) {
	if !t.SupportsModel(modelName) {
		return "", NewTokenizerErrorWithDetails("openrouter", modelName, "unsupported model", nil, "tiktoken-o200k")
	}
	return "o200k_base", nil
}
