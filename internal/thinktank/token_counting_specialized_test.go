package thinktank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test tokenizer basic functionality with various models.
// Note: All models use tiktoken o200k_base via OpenRouter normalization.
func TestTokenCountingService_ModelTokenizer_BasicEnglish(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	req := TokenCountingRequest{
		Instructions: "Analyze this English text",
		Files: []FileContent{
			{
				Path:    "english.txt",
				Content: "This is basic English text that should be tokenized accurately.",
			},
		},
	}

	result, err := service.CountTokensForModel(context.Background(), req, "gemini-3-flash")

	require.NoError(t, err)
	assert.Greater(t, result.TotalTokens, 0, "Should count tokens for English text")
}

// Test accuracy comparison between different models.
// All models use tiktoken o200k_base via OpenRouter, so counts should be similar.
func TestTokenCountingService_ModelComparison(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	req := TokenCountingRequest{
		Instructions: "Compare tokenization accuracy between different methods",
		Files: []FileContent{
			{
				Path: "comparison.txt",
				Content: `
This is a comprehensive test of tokenization accuracy across different approaches.
We're comparing token counts across different model providers.

Key points to analyze:
1. Subword tokenization efficiency
2. Handling of technical terms like "tokenization", "subword"
3. Performance with code-like content: func main() { fmt.Println("hello") }
4. Mixed content with punctuation, numbers (123, 456), and symbols (@#$%)

All OpenRouter models are normalized to tiktoken o200k_base encoding,
so token counts should be consistent across providers.
`,
			},
		},
	}

	// Get result for one model
	result1, err := service.CountTokensForModel(context.Background(), req, "gemini-3-flash")
	require.NoError(t, err)
	assert.Greater(t, result1.TotalTokens, 0, "Should count tokens")

	t.Logf("Model 1 (gemini-3-flash) tokens: %d", result1.TotalTokens)

	// Test with a different model - should use same tiktoken o200k_base
	result2, err := service.CountTokensForModel(context.Background(), req, "gpt-5.2")
	require.NoError(t, err)
	assert.Greater(t, result2.TotalTokens, 0, "Should count tokens")

	t.Logf("Model 2 (gpt-5.2) tokens: %d", result2.TotalTokens)

	// Both use tiktoken o200k_base via OpenRouter, counts should be very close
	// (may differ slightly due to model-specific overhead calculations)
}

// Test non-English content to verify Unicode handling.
// All models use tiktoken o200k_base which handles Unicode well.
func TestTokenCountingService_NonEnglish_TokenCharacterRatioBreakdown(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	tests := []struct {
		name        string
		language    string
		content     string
		description string
	}{
		{
			name:        "Japanese text",
			language:    "japanese",
			content:     "これはテストです。日本語のテキストを分析しています。",
			description: "Japanese with hiragana and kanji",
		},
		{
			name:        "Chinese text",
			language:    "chinese",
			content:     "这是一个测试。我们正在分析中文文本的标记化。",
			description: "Simplified Chinese characters",
		},
		{
			name:        "Arabic text",
			language:    "arabic",
			content:     "هذا اختبار. نحن نحلل نص عربي للترميز.",
			description: "Arabic text with RTL script",
		},
		{
			name:        "Mixed Unicode",
			language:    "mixed",
			content:     "Hello 世界! 🌍 Testing émojis and ñoñ-ASCII characters. مرحبا!",
			description: "Mixed languages with emoji and special characters",
		},
		{
			name:        "Code with Unicode",
			language:    "code",
			content:     `func main() { fmt.Println("Hello 世界! 🚀") }`,
			description: "Go code with Unicode strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := TokenCountingRequest{
				Instructions: "Analyze this " + tt.language + " content for tokenization",
				Files: []FileContent{
					{
						Path:    tt.name + ".txt",
						Content: tt.content,
					},
				},
			}

			// Test with one model (all use tiktoken o200k_base via OpenRouter)
			result1, err := service.CountTokensForModel(context.Background(), req, "gemini-3-flash")
			require.NoError(t, err, "Should handle %s", tt.description)

			// Test with another model
			result2, err := service.CountTokensForModel(context.Background(), req, "gpt-5.2")
			require.NoError(t, err, "Should handle %s", tt.description)

			// Both should produce token counts
			assert.Greater(t, result1.TotalTokens, 0, "Should count tokens for %s", tt.description)
			assert.Greater(t, result2.TotalTokens, 0, "Should count tokens for %s", tt.description)

			// Calculate character count for comparison
			charCount := len([]rune(req.Instructions + tt.content))

			t.Logf("%s:", tt.description)
			t.Logf("  Characters: %d", charCount)
			t.Logf("  Model 1 tokens: %d (ratio: %.2f tokens/char)",
				result1.TotalTokens, float64(result1.TotalTokens)/float64(charCount))
			t.Logf("  Model 2 tokens: %d (ratio: %.2f tokens/char)",
				result2.TotalTokens, float64(result2.TotalTokens)/float64(charCount))
		})
	}
}
