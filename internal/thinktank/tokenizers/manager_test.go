package tokenizers

import (
	"testing"

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
		{"openai", true},
		{"gemini", true},     // Adding SentencePiece support
		{"openrouter", true}, // OpenRouter normalization to o200k_base
		{"unknown", false},
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

	// Test supported provider
	tokenizer, err := manager.GetTokenizer("openai")
	require.NoError(t, err)
	assert.NotNil(t, tokenizer)
	assert.IsType(t, &OpenAITokenizer{}, tokenizer)

	// Test that same tokenizer is returned (caching)
	tokenizer2, err := manager.GetTokenizer("openai")
	require.NoError(t, err)
	assert.Same(t, tokenizer, tokenizer2, "Should return cached tokenizer")

	// Test Gemini provider (now supported)
	geminiTokenizer, err := manager.GetTokenizer("gemini")
	require.NoError(t, err)
	assert.NotNil(t, geminiTokenizer)
	assert.IsType(t, &GeminiTokenizer{}, geminiTokenizer)

	// Test OpenRouter provider (should return OpenRouter tokenizer)
	openrouterTokenizer, err := manager.GetTokenizer("openrouter")
	require.NoError(t, err)
	assert.NotNil(t, openrouterTokenizer)
	assert.IsType(t, &OpenRouterTokenizer{}, openrouterTokenizer)

	// Test unknown provider
	_, err = manager.GetTokenizer("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestTokenizerManager_ClearCache(t *testing.T) {
	t.Parallel()

	manager := NewTokenizerManager()

	// Get tokenizer to populate cache
	tokenizer1, err := manager.GetTokenizer("openai")
	require.NoError(t, err)

	// Clear cache
	manager.ClearCache()

	// Get tokenizer again - should be new instance
	tokenizer2, err := manager.GetTokenizer("openai")
	require.NoError(t, err)

	// Should be different instances after cache clear
	assert.NotSame(t, tokenizer1, tokenizer2, "Should create new tokenizer after cache clear")
}

func TestTokenizerManager_ConcurrentAccess(t *testing.T) {
	// Test that concurrent access to tokenizer manager is safe
	manager := NewTokenizerManager()

	// Run multiple goroutines trying to get the same tokenizer
	const numGoroutines = 10
	tokenizers := make([]AccurateTokenCounter, numGoroutines)
	errors := make([]error, numGoroutines)

	done := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			tokenizers[index], errors[index] = manager.GetTokenizer("openai")
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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.GetTokenizer("openrouter")
		if err != nil {
			b.Fatal(err)
		}
	}
}
