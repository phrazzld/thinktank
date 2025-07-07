package tokenizers

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRouterTokenizer_CountTokens(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenRouterTokenizer()
	ctx := context.Background()

	tests := []struct {
		name        string
		text        string
		model       string
		minTokens   int
		maxTokens   int
		expectError bool
	}{
		{
			name:      "basic english text",
			text:      "Hello, world!",
			model:     "openrouter/anthropic/claude-3-sonnet",
			minTokens: 2,
			maxTokens: 4,
		},
		{
			name:      "technical content",
			text:      "func main() { fmt.Println(\"test\") }",
			model:     "openrouter/deepseek/deepseek-chat",
			minTokens: 8,
			maxTokens: 15,
		},
		{
			name:      "unicode content",
			text:      "Hello 世界 🌍",
			model:     "openrouter/google/gemini-pro",
			minTokens: 3,
			maxTokens: 8,
		},
		{
			name:      "empty text",
			text:      "",
			model:     "openrouter/openai/gpt-4o",
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:      "long technical text",
			text:      "package main\n\nimport \"fmt\"\n\nfunc fibonacci(n int) int {\n\tif n <= 1 {\n\t\treturn n\n\t}\n\treturn fibonacci(n-1) + fibonacci(n-2)\n}\n\nfunc main() {\n\tfmt.Println(fibonacci(10))\n}",
			model:     "openrouter/meta-llama/llama-3.3-70b-instruct",
			minTokens: 40,
			maxTokens: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tokenizer.CountTokens(ctx, tt.text, tt.model)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, count, tt.minTokens, "Token count should be at least %d", tt.minTokens)
			assert.LessOrEqual(t, count, tt.maxTokens, "Token count should be at most %d", tt.maxTokens)
		})
	}
}

func TestOpenRouterTokenizer_SupportsModel(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenRouterTokenizer()

	tests := []struct {
		name      string
		modelName string
		expected  bool
	}{
		{
			name:      "openrouter anthropic model",
			modelName: "openrouter/anthropic/claude-3-sonnet",
			expected:  true,
		},
		{
			name:      "openrouter deepseek model",
			modelName: "openrouter/deepseek/deepseek-chat",
			expected:  true,
		},
		{
			name:      "openrouter openai model",
			modelName: "openrouter/openai/gpt-4o",
			expected:  true,
		},
		{
			name:      "generic model name",
			modelName: "some-model",
			expected:  true, // OpenRouter normalizes everything, so we support all
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenizer.SupportsModel(tt.modelName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenRouterTokenizer_GetEncoding(t *testing.T) {
	t.Parallel()

	tokenizer := NewOpenRouterTokenizer()

	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{
			name:     "openrouter model returns o200k_base",
			model:    "openrouter/anthropic/claude-3-sonnet",
			expected: "o200k_base",
		},
		{
			name:     "any model returns o200k_base",
			model:    "openrouter/meta-llama/llama-3.3-70b-instruct",
			expected: "o200k_base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoding, err := tokenizer.GetEncoding(tt.model)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, encoding)
		})
	}
}

func TestOpenRouterTokenizer_ConsistencyWithOpenAI(t *testing.T) {
	t.Parallel()

	openrouterTokenizer := NewOpenRouterTokenizer()
	openaiTokenizer := NewOpenAITokenizer()
	ctx := context.Background()

	testText := "The quick brown fox jumps over the lazy dog."

	// OpenRouter should give same count as OpenAI o4-mini (o200k_base)
	openrouterCount, err := openrouterTokenizer.CountTokens(ctx, testText, "openrouter/anthropic/claude-3-sonnet")
	require.NoError(t, err)

	openaiCount, err := openaiTokenizer.CountTokens(ctx, testText, "o4-mini")
	require.NoError(t, err)

	assert.Equal(t, openaiCount, openrouterCount, "OpenRouter tokenization should match OpenAI o4-mini (o200k_base)")
}

// Benchmark tests to validate performance characteristics
func BenchmarkOpenRouterTokenizer_CountTokens(b *testing.B) {
	perftest.RunBenchmark(b, "OpenRouterTokenizer_CountTokens", func(b *testing.B) {
		tokenizer := NewOpenRouterTokenizer()
		ctx := context.Background()
		text := "The quick brown fox jumps over the lazy dog. This is a performance test for OpenRouter tokenization."
		model := "openrouter/anthropic/claude-3-sonnet"

		// Warm up
		_, err := tokenizer.CountTokens(ctx, text, model)
		require.NoError(b, err)

		b.ResetTimer()
		perftest.ReportAllocs(b)

		for i := 0; i < b.N; i++ {
			_, err := tokenizer.CountTokens(ctx, text, model)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkOpenRouterTokenizer_vs_OpenAI(b *testing.B) {
	// Compare wrapper overhead vs direct OpenAI usage
	openaiTokenizer := NewOpenAITokenizer()
	openrouterTokenizer := NewOpenRouterTokenizer()

	text := "Performance comparison test content for tokenization overhead measurement"
	ctx := context.Background()

	b.Run("Direct_OpenAI", func(b *testing.B) {
		perftest.RunBenchmark(b, "OpenRouterTokenizer_vs_OpenAI_Direct", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_, err := openaiTokenizer.CountTokens(ctx, text, "o4-mini")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("OpenRouter_Wrapper", func(b *testing.B) {
		perftest.RunBenchmark(b, "OpenRouterTokenizer_vs_OpenAI_Wrapper", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_, err := openrouterTokenizer.CountTokens(ctx, text, "openrouter/openai/gpt-4o")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
