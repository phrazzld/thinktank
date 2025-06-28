package tokenizers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAITokenizer_SupportsModel(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenAITokenizer()

	tests := []struct {
		model     string
		supported bool
	}{
		{"gpt-4.1", true},
		{"o4-mini", true},
		{"o3", true},
		{"gpt-4", true},
		{"gpt-4o", true},
		{"gpt-4o-mini", true},
		{"claude-3", false},
		{"gemini-2.5-pro", false},
		{"unknown-model", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := tokenizer.SupportsModel(tt.model)
			assert.Equal(t, tt.supported, result)
		})
	}
}

func TestOpenAITokenizer_GetEncoding(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenAITokenizer()

	tests := []struct {
		model            string
		expectedEncoding string
		expectError      bool
	}{
		{"gpt-4.1", "cl100k_base", false},
		{"o4-mini", "o200k_base", false},
		{"o3", "o200k_base", false},
		{"gpt-4", "cl100k_base", false},
		{"gpt-4o", "o200k_base", false},
		{"gpt-4o-mini", "o200k_base", false},
		{"unsupported-model", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			encoding, err := tokenizer.GetEncoding(tt.model)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, encoding)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEncoding, encoding)
			}
		})
	}
}

func TestOpenAITokenizer_CountTokens(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenAITokenizer()
	ctx := context.Background()

	tests := []struct {
		name        string
		text        string
		model       string
		minTokens   int // Minimum expected tokens (tiktoken may vary slightly)
		maxTokens   int // Maximum expected tokens
		expectError bool
	}{
		{
			name:        "empty text",
			text:        "",
			model:       "gpt-4.1",
			minTokens:   0,
			maxTokens:   0,
			expectError: false,
		},
		{
			name:        "simple text with gpt-4.1",
			text:        "Hello, world!",
			model:       "gpt-4.1",
			minTokens:   2, // Usually tokenizes to 2-4 tokens
			maxTokens:   4,
			expectError: false,
		},
		{
			name:        "simple text with o4-mini",
			text:        "Hello, world!",
			model:       "o4-mini",
			minTokens:   2, // Different encoding may have different counts
			maxTokens:   4,
			expectError: false,
		},
		{
			name:        "longer text with gpt-4.1",
			text:        "This is a longer piece of text that should tokenize into multiple tokens.",
			model:       "gpt-4.1",
			minTokens:   10, // Should be more than 10 tokens
			maxTokens:   25, // But less than 25
			expectError: false,
		},
		{
			name:        "code snippet with gpt-4.1",
			text:        "func main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			model:       "gpt-4.1",
			minTokens:   8, // Code typically has more tokens
			maxTokens:   20,
			expectError: false,
		},
		{
			name:        "unicode text with gpt-4.1",
			text:        "Hello 世界 🌍",
			model:       "gpt-4.1",
			minTokens:   3, // Unicode handling
			maxTokens:   8,
			expectError: false,
		},
		{
			name:        "unsupported model",
			text:        "Hello, world!",
			model:       "claude-3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tokenizer.CountTokens(ctx, tt.text, tt.model)

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, &TokenizerError{}, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, count, tt.minTokens)
				assert.LessOrEqual(t, count, tt.maxTokens)
			}
		})
	}
}

func TestOpenAITokenizer_CountTokens_Accuracy(t *testing.T) {
	// This test validates that tiktoken produces reasonable token counts
	// compared to the 0.75 tokens/character estimation

	tokenizer := NewOpenAITokenizer()
	ctx := context.Background()

	tests := []struct {
		name  string
		text  string
		model string
	}{
		{
			name:  "english sentence",
			text:  "The quick brown fox jumps over the lazy dog.",
			model: "gpt-4.1",
		},
		{
			name:  "technical text",
			text:  "The TCP/IP protocol stack includes application, transport, network, and data link layers.",
			model: "gpt-4.1",
		},
		{
			name:  "code example",
			text:  "function calculateSum(a, b) { return a + b; }",
			model: "gpt-4.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualTokens, err := tokenizer.CountTokens(ctx, tt.text, tt.model)
			require.NoError(t, err)

			// Calculate estimation using same method as models package
			estimatedTokens := int(float64(len(tt.text)) * 0.75)

			// Log for manual inspection of accuracy
			t.Logf("Text: %q", tt.text)
			t.Logf("Length: %d characters", len(tt.text))
			t.Logf("Estimated tokens: %d", estimatedTokens)
			t.Logf("Actual tokens: %d", actualTokens)
			t.Logf("Ratio: %.2f tokens/char", float64(actualTokens)/float64(len(tt.text)))

			// Sanity checks - actual should be reasonable
			assert.Greater(t, actualTokens, 0, "Should have at least some tokens")
			assert.Less(t, actualTokens, len(tt.text), "Should have fewer tokens than characters")
		})
	}
}

func TestOpenAITokenizer_ClearCache(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenAITokenizer()
	ctx := context.Background()

	// Trigger encoder initialization
	_, err := tokenizer.CountTokens(ctx, "test", "gpt-4.1")
	require.NoError(t, err)

	// Verify cache has content (indirect test)
	encoding, err := tokenizer.GetEncoding("gpt-4.1")
	require.NoError(t, err)

	encoder, err := tokenizer.getEncoder(encoding)
	require.NoError(t, err)
	assert.NotNil(t, encoder)

	// Clear cache
	tokenizer.ClearCache()

	// Should still work after cache clear (will re-initialize)
	count, err := tokenizer.CountTokens(ctx, "test", "gpt-4.1")
	require.NoError(t, err)
	assert.Greater(t, count, 0)
}

// Benchmark tests for performance validation
func BenchmarkOpenAITokenizer_CountTokens(b *testing.B) {
	tokenizer := NewOpenAITokenizer()
	ctx := context.Background()
	text := "The quick brown fox jumps over the lazy dog."
	model := "gpt-4.1"

	// Warm up the tokenizer
	_, err := tokenizer.CountTokens(ctx, text, model)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := tokenizer.CountTokens(ctx, text, model)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenAITokenizer_InitializationCost(b *testing.B) {
	ctx := context.Background()
	text := "test"
	model := "gpt-4.1"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create new tokenizer each time to measure initialization cost
		tokenizer := NewOpenAITokenizer()
		_, err := tokenizer.CountTokens(ctx, text, model)
		if err != nil {
			b.Fatal(err)
		}
	}
}
