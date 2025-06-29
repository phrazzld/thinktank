package tokenizers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeminiTokenizer_SupportsModel(t *testing.T) {
	t.Parallel()

	tokenizer := NewGeminiTokenizer()

	tests := []struct {
		model     string
		supported bool
	}{
		{"gemini-2.5-pro", true},
		{"gemini-2.5-flash", true},
		{"gemini-1.5-pro", true},
		{"gemma-3-27b-it", true},
		{"gpt-4.1", false},
		{"o4-mini", false},
		{"claude-3", false},
		{"unknown-model", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := tokenizer.SupportsModel(tt.model)
			assert.Equal(t, tt.supported, result)
		})
	}
}

func TestGeminiTokenizer_GetEncoding(t *testing.T) {
	t.Parallel()

	tokenizer := NewGeminiTokenizer()

	tests := []struct {
		model            string
		expectedEncoding string
		expectError      bool
	}{
		{"gemini-2.5-pro", "sentencepiece", false},
		{"gemini-2.5-flash", "sentencepiece", false},
		{"gemma-3-27b-it", "sentencepiece", false},
		{"gpt-4.1", "", true},
		{"unsupported-model", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			encoding, err := tokenizer.GetEncoding(tt.model)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, encoding)
				assert.IsType(t, &TokenizerError{}, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEncoding, encoding)
			}
		})
	}
}

func TestGeminiTokenizer_CountTokens_BasicFunctionality(t *testing.T) {
	t.Parallel()

	tokenizer := NewGeminiTokenizer()
	ctx := context.Background()

	// Test that CountTokens now works with basic approximation
	count, err := tokenizer.CountTokens(ctx, "Hello, world!", "gemini-2.5-pro")
	assert.NoError(t, err)
	assert.Greater(t, count, 0)

	// Test unsupported model still returns error
	_, err = tokenizer.CountTokens(ctx, "Hello, world!", "gpt-4.1")
	assert.Error(t, err)
	assert.IsType(t, &TokenizerError{}, err)
	assert.Contains(t, err.Error(), "unsupported model")

	// Test empty text
	count, err = tokenizer.CountTokens(ctx, "", "gemini-2.5-pro")
	assert.NoError(t, err) // Empty text should return 0 without error
	assert.Equal(t, 0, count)
}

func TestGeminiTokenizer_ClearCache(t *testing.T) {
	t.Parallel()

	tokenizer := NewGeminiTokenizer()

	// Clear cache should not panic even when empty
	tokenizer.ClearCache()

	// This should work without issues
	assert.NotPanics(t, func() {
		tokenizer.ClearCache()
	})
}

// Benchmark for when actual implementation is added
func BenchmarkGeminiTokenizer_GetEncoding(b *testing.B) {
	tokenizer := NewGeminiTokenizer()
	model := "gemini-2.5-pro"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tokenizer.GetEncoding(model)
		if err != nil {
			b.Fatal(err)
		}
	}
}
