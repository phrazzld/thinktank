// Package tokenizers provides accurate token counting implementations for various LLM providers.
// This package enables precise token counting to replace estimation-based model selection.
package tokenizers

import (
	"sync"
)

// tokenizerManagerImpl implements TokenizerManager with lazy loading and caching.
// This is the base implementation that provides core tokenizer management functionality.
type tokenizerManagerImpl struct {
	// tokenizerCache stores initialized tokenizers by provider name for fast access
	tokenizerCache sync.Map

	// initMutex ensures only one tokenizer initialization per provider to prevent races
	initMutex sync.Mutex
}

// NewTokenizerManager creates a new tokenizer manager with lazy loading.
// Tokenizers are initialized only when first requested to minimize startup overhead.
func NewTokenizerManager() TokenizerManager {
	return &tokenizerManagerImpl{}
}

// GetTokenizer returns a cached tokenizer or creates a new one for the provider.
// Uses double-checked locking pattern to ensure thread-safe lazy initialization.
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
	case "openrouter":
		tokenizer = NewOpenRouterTokenizer()
	default:
		return nil, NewTokenizerErrorWithDetails(provider, "", "unsupported provider (only OpenRouter is supported after consolidation)", nil, "unknown")
	}

	// Cache the tokenizer for future use
	m.tokenizerCache.Store(provider, tokenizer)
	return tokenizer, nil
}

// SupportsProvider returns true if accurate tokenization is available for the provider.
func (m *tokenizerManagerImpl) SupportsProvider(provider string) bool {
	switch provider {
	case "openrouter":
		return true
	default:
		return false // Only OpenRouter is supported after provider consolidation
	}
}

// ClearCache clears all cached tokenizers to free memory.
func (m *tokenizerManagerImpl) ClearCache() {
	m.tokenizerCache.Range(func(key, value interface{}) bool {
		// Note: OpenRouter tokenizer wraps OpenAI tokenizer internally
		// The internal tokenizer will be garbage collected when the wrapper is deleted
		m.tokenizerCache.Delete(key)
		return true
	})
}
