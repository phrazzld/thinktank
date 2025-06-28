package tokenizers

import (
	"fmt"
	"sync"
)

// tokenizerManagerImpl implements TokenizerManager with lazy loading and caching.
type tokenizerManagerImpl struct {
	// tokenizerCache stores initialized tokenizers by provider name
	tokenizerCache sync.Map

	// initMutex ensures only one tokenizer initialization per provider
	initMutex sync.Mutex
}

// NewTokenizerManager creates a new tokenizer manager with lazy loading.
func NewTokenizerManager() TokenizerManager {
	return &tokenizerManagerImpl{}
}

// GetTokenizer returns a cached tokenizer or creates a new one for the provider.
func (m *tokenizerManagerImpl) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	// Check cache first for fast path
	if cached, ok := m.tokenizerCache.Load(provider); ok {
		return cached.(AccurateTokenCounter), nil
	}

	// Use mutex to prevent duplicate initialization
	m.initMutex.Lock()
	defer m.initMutex.Unlock()

	// Double-check cache after acquiring lock
	if cached, ok := m.tokenizerCache.Load(provider); ok {
		return cached.(AccurateTokenCounter), nil
	}

	// Create new tokenizer based on provider
	var tokenizer AccurateTokenCounter
	switch provider {
	case "openai":
		tokenizer = NewOpenAITokenizer()
	case "gemini":
		// TODO: Implement Gemini tokenizer in Phase 2
		return nil, fmt.Errorf("gemini tokenizer not yet implemented")
	case "openrouter":
		// TODO: OpenRouter uses estimation or GPT-4o tokenizer
		return nil, fmt.Errorf("openrouter tokenizer not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	// Cache the tokenizer for future use
	m.tokenizerCache.Store(provider, tokenizer)
	return tokenizer, nil
}

// SupportsProvider returns true if accurate tokenization is available for the provider.
func (m *tokenizerManagerImpl) SupportsProvider(provider string) bool {
	switch provider {
	case "openai":
		return true
	case "gemini":
		// TODO: Return true when Gemini tokenizer is implemented
		return false
	case "openrouter":
		// TODO: Return true when OpenRouter strategy is implemented
		return false
	default:
		return false
	}
}

// ClearCache clears all cached tokenizers to free memory.
func (m *tokenizerManagerImpl) ClearCache() {
	m.tokenizerCache.Range(func(key, value interface{}) bool {
		// Clear individual tokenizer caches if they support it
		if tokenizer, ok := value.(*OpenAITokenizer); ok {
			tokenizer.ClearCache()
		}
		m.tokenizerCache.Delete(key)
		return true
	})
}
