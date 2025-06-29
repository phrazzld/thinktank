package thinktank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Gemini tokenizer basic functionality
func TestTokenCountingService_GeminiTokenizer_BasicEnglish(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	req := TokenCountingRequest{
		Instructions: "Analyze this English text",
		Files: []FileContent{
			{
				Path:    "english.txt",
				Content: "This is basic English text that should be tokenized using SentencePiece.",
			},
		},
	}

	result, err := service.CountTokensForModel(context.Background(), req, "gemini-2.5-pro")

	require.NoError(t, err)
	assert.Greater(t, result.TotalTokens, 0, "Should count tokens for English text with Gemini")
	// Note: FileCount is not part of the TokenCountingResult interface
	// This was removed in the interface design
}

// Test accuracy comparison between Gemini and estimation
func TestTokenCountingService_GeminiVsEstimation_Comparison(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	// Use content that should show clear differences between tokenizers
	req := TokenCountingRequest{
		Instructions: "Compare tokenization accuracy between different methods",
		Files: []FileContent{
			{
				Path: "comparison.txt",
				Content: `
This is a comprehensive test of tokenization accuracy across different approaches.
We're comparing SentencePiece (Gemini) tokenization with character-based estimation.

Key points to analyze:
1. Subword tokenization efficiency
2. Handling of technical terms like "tokenization", "SentencePiece", "subword"
3. Performance with code-like content: func main() { fmt.Println("hello") }
4. Mixed content with punctuation, numbers (123, 456), and symbols (@#$%)

The expectation is that SentencePiece should provide more accurate token counts
for this type of mixed content compared to simple character estimation.
`,
			},
		},
	}

	// Get Gemini result (should use SentencePiece)
	geminiResult, err := service.CountTokensForModel(context.Background(), req, "gemini-2.5-pro")
	require.NoError(t, err)

	// Verify we got a reasonable result
	assert.Greater(t, geminiResult.TotalTokens, 0, "Gemini should count some tokens")
	// Note: FileCount is not part of the TokenCountingResult interface
	// This was removed in the interface design

	t.Logf("Gemini (SentencePiece) tokens: %d", geminiResult.TotalTokens)

	// Test with an unknown model (should fall back to estimation)
	estimationResult, err := service.CountTokensForModel(context.Background(), req, "unknown-estimation-model")
	require.NoError(t, err)

	assert.Greater(t, estimationResult.TotalTokens, 0, "Estimation should count some tokens")

	t.Logf("Estimation tokens: %d", estimationResult.TotalTokens)

	// Both should be in reasonable ranges, but might be different
	// This test documents the behavior rather than asserting specific relationships
}

// Test non-English content to demonstrate tokenizer differences
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

			// Test with Gemini (SentencePiece)
			geminiResult, err := service.CountTokensForModel(context.Background(), req, "gemini-2.5-pro")
			require.NoError(t, err, "Gemini should handle %s", tt.description)

			// Test with estimation fallback
			estimationResult, err := service.CountTokensForModel(context.Background(), req, "unknown-model")
			require.NoError(t, err, "Estimation should handle %s", tt.description)

			// Both should produce token counts
			assert.Greater(t, geminiResult.TotalTokens, 0, "Gemini should count tokens for %s", tt.description)
			assert.Greater(t, estimationResult.TotalTokens, 0, "Estimation should count tokens for %s", tt.description)

			// Calculate character count for comparison
			charCount := len([]rune(req.Instructions + tt.content))

			t.Logf("%s:", tt.description)
			t.Logf("  Characters: %d", charCount)
			t.Logf("  Gemini tokens: %d (ratio: %.2f tokens/char)",
				geminiResult.TotalTokens, float64(geminiResult.TotalTokens)/float64(charCount))
			t.Logf("  Estimation tokens: %d (ratio: %.2f tokens/char)",
				estimationResult.TotalTokens, float64(estimationResult.TotalTokens)/float64(charCount))

			// Document the differences - this test is primarily for observing behavior
			// rather than asserting specific relationships since different tokenizers
			// legitimately produce different results
		})
	}
}
