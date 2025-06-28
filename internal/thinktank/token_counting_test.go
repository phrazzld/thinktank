package thinktank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Write the smallest failing test first
func TestTokenCountingService_CountsEmptyContext(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "",
		Files:        []FileContent{},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalTokens)
}

// Second RED: Add test for single file counting
func TestTokenCountingService_CountsSingleFile(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "analyze this",
		Files: []FileContent{
			{Path: "test.go", Content: "package main\nfunc main() {}"},
		},
	})

	require.NoError(t, err)
	assert.Greater(t, result.TotalTokens, 0)
	assert.Greater(t, result.InstructionTokens, 0)
	assert.Greater(t, result.FileTokens, 0)

	// Verify components sum to total
	expected := result.InstructionTokens + result.FileTokens + result.Overhead
	assert.Equal(t, expected, result.TotalTokens)
}

// Third RED: Add table-driven test for comprehensive scenarios
func TestTokenCountingService_TableDriven(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()

	tests := []struct {
		name         string
		instructions string
		files        []FileContent
		expectTokens int
	}{
		{
			name:         "empty context",
			instructions: "",
			files:        []FileContent{},
			expectTokens: 0,
		},
		{
			name:         "instructions only",
			instructions: "test", // 4 chars * 0.75 + 1000 overhead = 1003
			files:        []FileContent{},
			expectTokens: 1503, // instruction tokens (1003) + overhead (500)
		},
		{
			name:         "files only",
			instructions: "",
			files: []FileContent{
				{Path: "test.go", Content: "test"}, // 4 chars * 0.75 = 3
			},
			expectTokens: 503, // 0 instruction tokens + 3 file tokens + 500 overhead
		},
		{
			name:         "instructions and files",
			instructions: "analyze", // 7 chars * 0.75 + 1000 = 1005.25 -> 1005
			files: []FileContent{
				{Path: "main.go", Content: "package main"}, // 12 chars * 0.75 = 9
			},
			expectTokens: 1519, // 1005 + 9 + 500 + 5 (rounding)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.CountTokens(context.Background(), TokenCountingRequest{
				Instructions: tt.instructions,
				Files:        tt.files,
			})

			require.NoError(t, err)
			// Allow small variance due to rounding in token estimation
			assert.InDelta(t, tt.expectTokens, result.TotalTokens, 10,
				"Expected ~%d tokens, got %d", tt.expectTokens, result.TotalTokens)
		})
	}
}

// Test the new CountTokensForModel method with accurate tokenization
func TestTokenCountingService_CountTokensForModel(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	tests := []struct {
		name           string
		instructions   string
		files          []FileContent
		modelName      string
		expectAccurate bool
		expectProvider string
	}{
		{
			name:         "OpenAI model with accurate tokenization",
			instructions: "Analyze this code",
			files: []FileContent{
				{Path: "test.go", Content: "package main\nfunc main() {}"},
			},
			modelName:      "gpt-4.1",
			expectAccurate: true,
			expectProvider: "openai",
		},
		{
			name:           "OpenAI o4-mini model with accurate tokenization",
			instructions:   "Review this",
			files:          []FileContent{{Path: "test.go", Content: "fmt.Println(\"hello\")"}},
			modelName:      "o4-mini",
			expectAccurate: true,
			expectProvider: "openai",
		},
		{
			name:           "Gemini model falls back to estimation",
			instructions:   "Analyze this",
			files:          []FileContent{{Path: "test.go", Content: "test content"}},
			modelName:      "gemini-2.5-pro",
			expectAccurate: false,
			expectProvider: "gemini",
		},
		{
			name:           "OpenRouter model falls back to estimation",
			instructions:   "Review",
			files:          []FileContent{{Path: "test.py", Content: "print('hello')"}},
			modelName:      "openrouter/deepseek/deepseek-chat-v3-0324",
			expectAccurate: false,
			expectProvider: "openrouter",
		},
		{
			name:           "Empty context with OpenAI model",
			instructions:   "",
			files:          []FileContent{},
			modelName:      "gpt-4.1",
			expectAccurate: true,
			expectProvider: "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.CountTokensForModel(ctx, TokenCountingRequest{
				Instructions: tt.instructions,
				Files:        tt.files,
			}, tt.modelName)

			require.NoError(t, err)
			assert.Equal(t, tt.modelName, result.ModelName)
			assert.Equal(t, tt.expectProvider, result.Provider)
			assert.Equal(t, tt.expectAccurate, result.IsAccurate)

			if tt.expectAccurate {
				switch tt.expectProvider {
				case "openai":
					assert.Equal(t, "tiktoken", result.TokenizerUsed)
				case "gemini":
					assert.Equal(t, "sentencepiece", result.TokenizerUsed)
				}
			} else {
				assert.Equal(t, "estimation", result.TokenizerUsed)
			}

			// Verify token counts are reasonable
			if tt.instructions == "" && len(tt.files) == 0 {
				assert.Equal(t, 0, result.TotalTokens)
			} else {
				assert.Greater(t, result.TotalTokens, 0)
			}

			// Verify breakdown consistency
			expectedTotal := result.InstructionTokens + result.FileTokens + result.Overhead
			assert.Equal(t, expectedTotal, result.TotalTokens)
		})
	}
}

// Test accuracy comparison between estimation and tiktoken
func TestTokenCountingService_AccuracyComparison(t *testing.T) {
	service := NewTokenCountingService()
	ctx := context.Background()

	testTexts := []struct {
		name         string
		instructions string
		files        []FileContent
	}{
		{
			name:         "English text",
			instructions: "Analyze the following code and provide suggestions for improvement.",
			files: []FileContent{
				{Path: "example.go", Content: "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"},
			},
		},
		{
			name:         "Technical documentation",
			instructions: "Review this API documentation for clarity and completeness.",
			files: []FileContent{
				{Path: "api.md", Content: "# API Documentation\n\n## Endpoints\n\n### GET /users\n\nReturns a list of users.\n\n#### Parameters\n\n- `limit`: Maximum number of users to return\n- `offset`: Number of users to skip"},
			},
		},
		{
			name:         "Code with comments",
			instructions: "Optimize this algorithm for better performance.",
			files: []FileContent{
				{Path: "algorithm.js", Content: "// Bubble sort implementation\nfunction bubbleSort(arr) {\n  // Outer loop for number of passes\n  for (let i = 0; i < arr.length - 1; i++) {\n    // Inner loop for comparisons\n    for (let j = 0; j < arr.length - i - 1; j++) {\n      if (arr[j] > arr[j + 1]) {\n        // Swap elements\n        [arr[j], arr[j + 1]] = [arr[j + 1], arr[j]];\n      }\n    }\n  }\n  return arr;\n}"},
			},
		},
	}

	for _, tt := range testTexts {
		t.Run(tt.name, func(t *testing.T) {
			// Get estimation-based count
			estimationResult, err := service.CountTokens(ctx, TokenCountingRequest{
				Instructions: tt.instructions,
				Files:        tt.files,
			})
			require.NoError(t, err)

			// Get accurate count for OpenAI model
			accurateResult, err := service.CountTokensForModel(ctx, TokenCountingRequest{
				Instructions: tt.instructions,
				Files:        tt.files,
			}, "gpt-4.1")
			require.NoError(t, err)

			// Log results for manual inspection
			t.Logf("Test case: %s", tt.name)
			t.Logf("Estimation tokens: %d", estimationResult.TotalTokens)
			t.Logf("Accurate tokens: %d", accurateResult.TotalTokens)
			t.Logf("Ratio: %.2f (accurate/estimation)", float64(accurateResult.TotalTokens)/float64(estimationResult.TotalTokens))
			t.Logf("Tokenizer used: %s", accurateResult.TokenizerUsed)
			t.Logf("Is accurate: %t", accurateResult.IsAccurate)

			// Verify accurate count is reasonable (should be different from estimation)
			assert.True(t, accurateResult.IsAccurate, "Should use accurate tokenization for OpenAI model")
			assert.Equal(t, "tiktoken", accurateResult.TokenizerUsed)
			assert.Greater(t, accurateResult.TotalTokens, 0, "Should have positive token count")

			// The accurate count should typically be lower than estimation for normal text
			// but we don't enforce this as an absolute rule since it depends on content
		})
	}
}

// Test error handling for invalid models
func TestTokenCountingService_CountTokensForModel_InvalidModel(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	_, err := service.CountTokensForModel(ctx, TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{{Path: "test.txt", Content: "content"}},
	}, "invalid-model")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown model")
}

// Benchmark comparing estimation vs accurate tokenization
func BenchmarkTokenCountingService_CompareTokenization(b *testing.B) {
	service := NewTokenCountingService()
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions: "Analyze this code and provide suggestions for improvement and optimization.",
		Files: []FileContent{
			{Path: "main.go", Content: "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tif len(os.Args) < 2 {\n\t\tfmt.Println(\"Usage: program <name>\")\n\t\treturn\n\t}\n\tfmt.Printf(\"Hello, %s!\\n\", os.Args[1])\n}"},
		},
	}

	b.Run("Estimation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := service.CountTokens(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Accurate_Tiktoken", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := service.CountTokensForModel(ctx, req, "gpt-4.1")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
