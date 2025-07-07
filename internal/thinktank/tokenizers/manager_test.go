package tokenizers

import (
	"testing"

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenizerManager_SupportsProvider(t *testing.T) {
	t.Parallel()

	manager := NewTokenizerManager()

	tests := []struct {
		provider  string
		supported bool
	}{
		{"openrouter", true}, // Only supported provider after consolidation
		{"openai", false},    // No longer supported after OpenRouter consolidation
		{"gemini", false},    // No longer supported after OpenRouter consolidation
		{"unknown", false},   // Never supported
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := manager.SupportsProvider(tt.provider)
			assert.Equal(t, tt.supported, result)
		})
	}
}

func TestTokenizerManager_GetTokenizer(t *testing.T) {
	t.Parallel()

	manager := NewTokenizerManager()

	// Test supported provider (OpenRouter only after consolidation)
	tokenizer, err := manager.GetTokenizer("openrouter")
	require.NoError(t, err)
	assert.NotNil(t, tokenizer)
	assert.IsType(t, &OpenRouterTokenizer{}, tokenizer)

	// Test that same tokenizer is returned (caching)
	tokenizer2, err := manager.GetTokenizer("openrouter")
	require.NoError(t, err)
	assert.Same(t, tokenizer, tokenizer2, "Should return cached tokenizer")

	// Test legacy providers are no longer supported
	legacyProviders := []string{"openai", "gemini"}
	for _, provider := range legacyProviders {
		t.Run("legacy_"+provider, func(t *testing.T) {
			_, err := manager.GetTokenizer(provider)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported provider")
			assert.Contains(t, err.Error(), "only OpenRouter is supported")
		})
	}

	// Test unknown provider
	_, err = manager.GetTokenizer("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestTokenizerManager_ClearCache(t *testing.T) {
	t.Parallel()

	manager := NewTokenizerManager()

	// Get tokenizer to populate cache (using OpenRouter after consolidation)
	tokenizer1, err := manager.GetTokenizer("openrouter")
	require.NoError(t, err)

	// Clear cache
	manager.ClearCache()

	// Get tokenizer again - should be new instance
	tokenizer2, err := manager.GetTokenizer("openrouter")
	require.NoError(t, err)

	// Should be different instances after cache clear
	assert.NotSame(t, tokenizer1, tokenizer2, "Should create new tokenizer after cache clear")
}

func TestTokenizerManager_ConcurrentAccess(t *testing.T) {
	// Test that concurrent access to tokenizer manager is safe
	manager := NewTokenizerManager()

	// Use supported provider after OpenRouter consolidation
	const testProvider = "openrouter"
	require.True(t, manager.SupportsProvider(testProvider), "Test setup error: %s should be supported", testProvider)

	// Run multiple goroutines trying to get the same tokenizer
	const numGoroutines = 10
	tokenizers := make([]AccurateTokenCounter, numGoroutines)
	errors := make([]error, numGoroutines)

	done := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			tokenizers[index], errors[index] = manager.GetTokenizer(testProvider)
			done <- index
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all succeeded and got the same tokenizer instance
	for i := 0; i < numGoroutines; i++ {
		require.NoError(t, errors[i], "Goroutine %d should not have error", i)
		require.NotNil(t, tokenizers[i], "Goroutine %d should get tokenizer", i)

		if i > 0 {
			assert.Same(t, tokenizers[0], tokenizers[i],
				"All goroutines should get the same cached tokenizer instance")
		}
	}
}

// Benchmark for manager overhead
func BenchmarkTokenizerManager_GetTokenizer(b *testing.B) {
	manager := NewTokenizerManager()

	// Warm up
	_, err := manager.GetTokenizer("openrouter")
	require.NoError(b, err)

	perftest.RunBenchmark(b, "TokenizerManager_GetTokenizer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.GetTokenizer("openrouter")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
