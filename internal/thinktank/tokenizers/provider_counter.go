// Package tokenizers provides accurate token counting implementations for various LLM providers.
package tokenizers

import (
	"context"
	"fmt"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
)

// ProviderTokenCounter provides unified provider-aware token counting.
// This struct centralizes provider detection and routing logic for all tokenizers.
// It automatically selects the best tokenizer based on the model's provider and
// gracefully falls back to estimation when accurate tokenizers are unavailable.
// Thread-safe for concurrent use across multiple goroutines.
type ProviderTokenCounter struct {
	tiktoken      AccurateTokenCounter   // For OpenAI models
	sentencePiece AccurateTokenCounter   // For Gemini models
	openrouter    AccurateTokenCounter   // For OpenRouter models (o200k_base normalization)
	fallback      EstimationTokenCounter // For unsupported models
	logger        logutil.LoggerInterface
}

// EstimationTokenCounter provides fallback token counting using character estimation.
// This interface allows for dependency injection and testing of the fallback mechanism.
// Used when provider-specific tokenizers are unavailable or fail to initialize.
// Provides universal compatibility at the cost of reduced accuracy (~75% vs 90%+).
type EstimationTokenCounter interface {
	// CountTokens estimates tokens using character-based calculation.
	CountTokens(ctx context.Context, text string, modelName string) (int, error)

	// SupportsModel returns true for all models (estimation works universally).
	SupportsModel(modelName string) bool

	// GetEncoding returns "estimation" for all models.
	GetEncoding(modelName string) (string, error)
}

// estimationTokenCounterImpl implements EstimationTokenCounter using pure character-based logic.
// Uses 0.75 tokens per character ratio without the instruction overhead added by models.EstimateTokensFromText.
// This provides clean text tokenization for the tokenizers package.
type estimationTokenCounterImpl struct{}

// NewEstimationTokenCounter creates a new estimation-based token counter.
// This counter works universally with all models and provides consistent fallback behavior.
// Used internally by ProviderTokenCounter when accurate tokenizers are unavailable.
func NewEstimationTokenCounter() EstimationTokenCounter {
	return &estimationTokenCounterImpl{}
}

// CountTokens implements estimation-based token counting.
func (e *estimationTokenCounterImpl) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if text == "" {
		return 0, nil
	}
	// Use the same estimation logic as models.EstimateTokensFromText
	charCount := len(text)
	tokens := int(float64(charCount) * 0.75)
	return tokens, nil
}

// SupportsModel returns true for all models (estimation works universally).
func (e *estimationTokenCounterImpl) SupportsModel(modelName string) bool {
	return true
}

// GetEncoding returns "estimation" for all models.
func (e *estimationTokenCounterImpl) GetEncoding(modelName string) (string, error) {
	return "estimation", nil
}

// NewProviderTokenCounter creates a new provider-aware token counter.
// Uses lazy loading for tokenizers - they will be initialized on first use.
// Logger is optional (can be nil) but recommended for debugging tokenizer selection decisions.
// This is the recommended entry point for production token counting.
func NewProviderTokenCounter(logger logutil.LoggerInterface) *ProviderTokenCounter {
	return &ProviderTokenCounter{
		tiktoken:      nil, // Lazy loaded on first use
		sentencePiece: nil, // Lazy loaded on first use
		openrouter:    nil, // Lazy loaded on first use
		fallback:      NewEstimationTokenCounter(),
		logger:        logger,
	}
}

// CountTokens routes to the appropriate tokenizer based on model provider.
// Implements provider detection logic using models.GetModelInfo().
func (p *ProviderTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if text == "" {
		return 0, nil
	}

	// Detect provider for the model
	provider, err := p.detectProvider(modelName)
	if err != nil {
		if p.logger != nil {
			p.logger.WarnContext(ctx, "Provider detection failed for model, using estimation",
				"model", modelName, "error", err.Error())
		}
		return p.fallback.CountTokens(ctx, text, modelName)
	}

	// Route to appropriate tokenizer based on provider
	tokenizer, tokenizerType, err := p.getTokenizerForProvider(provider)
	if err != nil {
		if p.logger != nil {
			p.logger.WarnContext(ctx, "Failed to get tokenizer for provider, using estimation",
				"model", modelName, "provider", provider, "error", err.Error())
		}
		return p.fallback.CountTokens(ctx, text, modelName)
	}

	// Log tokenizer selection decision
	if p.logger != nil {
		p.logger.DebugContext(ctx, "Selected tokenizer for model",
			"model", modelName,
			"provider", provider,
			"tokenizer", tokenizerType)
	}

	// Use the selected tokenizer
	tokens, err := tokenizer.CountTokens(ctx, text, modelName)
	if err != nil {
		if p.logger != nil {
			p.logger.WarnContext(ctx, "Accurate tokenization failed, falling back to estimation",
				"model", modelName, "provider", provider, "tokenizer", tokenizerType, "error", err.Error())
		}
		return p.fallback.CountTokens(ctx, text, modelName)
	}

	return tokens, nil
}

// SupportsModel returns true if accurate tokenization is available for the model.
func (p *ProviderTokenCounter) SupportsModel(modelName string) bool {
	provider, err := p.detectProvider(modelName)
	if err != nil {
		return false // Unknown provider, only estimation available
	}

	tokenizer, _, err := p.getTokenizerForProvider(provider)
	if err != nil {
		return false // Provider not supported for accurate tokenization
	}

	return tokenizer.SupportsModel(modelName)
}

// GetEncoding returns the tokenizer encoding name for the given model.
func (p *ProviderTokenCounter) GetEncoding(modelName string) (string, error) {
	provider, err := p.detectProvider(modelName)
	if err != nil {
		return "estimation", nil
	}

	tokenizer, tokenizerType, err := p.getTokenizerForProvider(provider)
	if err != nil {
		return "estimation", nil
	}

	encoding, err := tokenizer.GetEncoding(modelName)
	if err != nil {
		return "estimation", nil
	}

	// Return tokenizer type for better identification
	switch tokenizerType {
	case "tiktoken":
		return fmt.Sprintf("tiktoken:%s", encoding), nil
	case "sentencepiece":
		return fmt.Sprintf("sentencepiece:%s", encoding), nil
	default:
		return encoding, nil
	}
}

// detectProvider determines the provider for a given model name.
// Uses existing models.GetModelInfo() to get provider information.
func (p *ProviderTokenCounter) detectProvider(modelName string) (string, error) {
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		return "", NewTokenizerErrorWithDetails("unknown", modelName, "failed to detect provider", err, "unknown")
	}
	return modelInfo.Provider, nil
}

// getTokenizerForProvider returns the appropriate tokenizer for a provider.
// Implements lazy loading to avoid initialization overhead at startup.
func (p *ProviderTokenCounter) getTokenizerForProvider(provider string) (AccurateTokenCounter, string, error) {
	switch provider {
	case "openai":
		if p.tiktoken == nil {
			p.tiktoken = NewOpenAITokenizer()
		}
		return p.tiktoken, "tiktoken", nil

	case "gemini":
		if p.sentencePiece == nil {
			p.sentencePiece = NewGeminiTokenizer()
		}
		return p.sentencePiece, "sentencepiece", nil

	case "openrouter":
		if p.openrouter == nil {
			p.openrouter = NewOpenRouterTokenizer()
		}
		return p.openrouter, "tiktoken-o200k", nil

	default:
		return nil, "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// GetTokenizerType returns the tokenizer type that would be used for a model.
// Useful for logging and debugging without actually performing tokenization.
func (p *ProviderTokenCounter) GetTokenizerType(modelName string) string {
	provider, err := p.detectProvider(modelName)
	if err != nil {
		return "estimation"
	}

	switch provider {
	case "openai":
		return "tiktoken"
	case "gemini":
		return "sentencepiece"
	case "openrouter":
		return "tiktoken-o200k"
	default:
		return "estimation"
	}
}

// IsAccurate returns true if accurate tokenization is available for the model.
// This is a convenience method for determining tokenization accuracy.
func (p *ProviderTokenCounter) IsAccurate(modelName string) bool {
	provider, err := p.detectProvider(modelName)
	if err != nil {
		return false
	}

	// OpenAI, Gemini, and OpenRouter providers have accurate tokenization
	return provider == "openai" || provider == "gemini" || provider == "openrouter"
}

// ClearCache clears all cached tokenizers to free memory.
// Implements the cache management interface for memory optimization.
func (p *ProviderTokenCounter) ClearCache() {
	if p.tiktoken != nil {
		if openaiTokenizer, ok := p.tiktoken.(*OpenAITokenizer); ok {
			openaiTokenizer.ClearCache()
		}
		p.tiktoken = nil
	}

	if p.sentencePiece != nil {
		if geminiTokenizer, ok := p.sentencePiece.(*GeminiTokenizer); ok {
			geminiTokenizer.ClearCache()
		}
		p.sentencePiece = nil
	}

	if p.openrouter != nil {
		// OpenRouter wraps OpenAI tokenizer, so no special cache clearing needed
		p.openrouter = nil
	}
}
