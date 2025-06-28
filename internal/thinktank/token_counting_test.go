package thinktank

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
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

// Test error handling and fallback behavior for tokenizer failures
func TestTokenCountingService_CountTokensForModel_TokenizerFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		modelName       string
		expectAccurate  bool
		expectProvider  string
		expectTokenizer string
	}{
		{
			name:            "OpenAI model with accurate tokenization",
			modelName:       "gpt-4.1",
			expectAccurate:  true,
			expectProvider:  "openai",
			expectTokenizer: "tiktoken",
		},
		{
			name:            "Gemini model fallback to estimation (no SentencePiece yet)",
			modelName:       "gemini-2.5-pro",
			expectAccurate:  false,
			expectProvider:  "gemini",
			expectTokenizer: "estimation",
		},
		{
			name:            "OpenRouter model fallback to estimation",
			modelName:       "openrouter/deepseek/deepseek-chat-v3-0324",
			expectAccurate:  false,
			expectProvider:  "openrouter",
			expectTokenizer: "estimation",
		},
	}

	service := NewTokenCountingService()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := TokenCountingRequest{
				Instructions: "Analyze this code",
				Files: []FileContent{
					{Path: "test.go", Content: "package main\nfunc main() {}"},
				},
			}

			result, err := service.CountTokensForModel(ctx, req, tt.modelName)

			require.NoError(t, err)
			assert.Equal(t, tt.modelName, result.ModelName)
			assert.Equal(t, tt.expectProvider, result.Provider)
			assert.Equal(t, tt.expectAccurate, result.IsAccurate)
			assert.Equal(t, tt.expectTokenizer, result.TokenizerUsed)
			assert.Greater(t, result.TotalTokens, 0)
		})
	}
}

// Test edge cases for CountTokensForModel
func TestTokenCountingService_CountTokensForModel_EdgeCases(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	tests := []struct {
		name         string
		req          TokenCountingRequest
		modelName    string
		expectTokens int
	}{
		{
			name: "empty instructions and files",
			req: TokenCountingRequest{
				Instructions: "",
				Files:        []FileContent{},
			},
			modelName:    "gpt-4.1",
			expectTokens: 0,
		},
		{
			name: "only whitespace instructions",
			req: TokenCountingRequest{
				Instructions: "   \n\t  ",
				Files:        []FileContent{},
			},
			modelName:    "gpt-4.1",
			expectTokens: 503, // whitespace counted as content + overhead
		},
		{
			name: "empty file content",
			req: TokenCountingRequest{
				Instructions: "test",
				Files: []FileContent{
					{Path: "empty.txt", Content: ""},
					{Path: "another.txt", Content: ""},
				},
			},
			modelName:    "o4-mini",
			expectTokens: 501, // instruction tokens (1) + overhead (500), no file tokens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.CountTokensForModel(ctx, tt.req, tt.modelName)

			require.NoError(t, err)
			assert.InDelta(t, tt.expectTokens, result.TotalTokens, 10,
				"Expected ~%d tokens, got %d", tt.expectTokens, result.TotalTokens)
		})
	}
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

// Test the new GetCompatibleModels method with accurate token counting
func TestTokenCountingService_GetCompatibleModels_EmptyInput(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	result, err := service.GetCompatibleModels(ctx, TokenCountingRequest{}, []string{})

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestTokenCountingService_GetCompatibleModels_SingleCompatibleModel(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions: "short instruction",
		Files:        []FileContent{},
	}

	result, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)
	assert.NotEmpty(t, result, "Should return at least one model for openai provider")

	// Find a compatible model in the results
	var compatibleFound bool
	for _, model := range result {
		if model.IsCompatible {
			compatibleFound = true
			assert.NotEmpty(t, model.ModelName)
			assert.Equal(t, "openai", model.Provider)
			assert.Greater(t, model.TokenCount, 0)
			assert.Greater(t, model.ContextWindow, 0)
			assert.Greater(t, model.UsableContext, 0)
			break
		}
	}
	assert.True(t, compatibleFound, "At least one model should be compatible with short instruction")
}

func TestTokenCountingService_GetCompatibleModels_TableDriven(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	tests := []struct {
		name             string
		req              TokenCountingRequest
		providers        []string
		expectCompatible bool
		expectResults    bool
		expectError      bool
		minResults       int
		maxResults       int
	}{
		{
			name: "empty providers",
			req: TokenCountingRequest{
				Instructions: "test",
				Files:        []FileContent{},
			},
			providers:        []string{},
			expectCompatible: false,
			expectResults:    false,
		},
		{
			name: "empty request",
			req: TokenCountingRequest{
				Instructions: "",
				Files:        []FileContent{},
			},
			providers:        []string{"openai"},
			expectCompatible: false,
			expectResults:    false,
		},
		{
			name: "short instruction with openai",
			req: TokenCountingRequest{
				Instructions: "Analyze this code",
				Files:        []FileContent{},
			},
			providers:        []string{"openai"},
			expectCompatible: true,
			expectResults:    true,
			minResults:       1,
			maxResults:       10, // reasonable upper bound
		},
		{
			name: "multiple providers",
			req: TokenCountingRequest{
				Instructions: "Review this",
				Files:        []FileContent{},
			},
			providers:        []string{"openai", "gemini"},
			expectCompatible: true,
			expectResults:    true,
			minResults:       2, // Should get models from both providers
			maxResults:       20,
		},
		{
			name: "with file content",
			req: TokenCountingRequest{
				Instructions: "Analyze",
				Files: []FileContent{
					{Path: "test.go", Content: "package main\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"},
				},
			},
			providers:        []string{"openai"},
			expectCompatible: true,
			expectResults:    true,
			minResults:       1,
			maxResults:       10,
		},
		{
			name: "unknown provider",
			req: TokenCountingRequest{
				Instructions: "test",
				Files:        []FileContent{},
			},
			providers:        []string{"unknown-provider"},
			expectCompatible: false,
			expectResults:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := service.GetCompatibleModels(ctx, tt.req, tt.providers)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if !tt.expectResults {
				assert.Empty(t, result)
				return
			}

			// Check result count bounds
			if tt.minResults > 0 {
				assert.GreaterOrEqual(t, len(result), tt.minResults,
					"Should have at least %d results", tt.minResults)
			}
			if tt.maxResults > 0 {
				assert.LessOrEqual(t, len(result), tt.maxResults,
					"Should have at most %d results", tt.maxResults)
			}

			// Check if we found compatible models
			if tt.expectCompatible {
				var compatibleFound bool
				for _, model := range result {
					if model.IsCompatible {
						compatibleFound = true
						// Validate structure of compatible model
						assert.NotEmpty(t, model.ModelName)
						assert.NotEmpty(t, model.Provider)
						assert.Greater(t, model.TokenCount, 0)
						assert.Greater(t, model.ContextWindow, 0)
						assert.Greater(t, model.UsableContext, 0)
						assert.NotEmpty(t, model.TokenizerUsed)
						assert.Empty(t, model.Reason) // Compatible models shouldn't have failure reason
						break
					}
				}
				assert.True(t, compatibleFound, "Should find at least one compatible model")
			}

			// Check that results are sorted (compatible first, then by context window)
			if len(result) > 1 {
				for i := 1; i < len(result); i++ {
					prev := result[i-1]
					curr := result[i]

					// If compatibility differs, compatible should come first
					if prev.IsCompatible != curr.IsCompatible {
						assert.True(t, prev.IsCompatible, "Compatible models should come first")
					} else if prev.IsCompatible == curr.IsCompatible {
						// Within same compatibility, larger context windows should come first
						assert.GreaterOrEqual(t, prev.ContextWindow, curr.ContextWindow,
							"Within same compatibility, larger context windows should come first")
					}
				}
			}
		})
	}
}

// Test NewTokenCountingServiceWithManager constructor
func TestNewTokenCountingServiceWithManager(t *testing.T) {
	t.Parallel()

	mockManager := &testutil.MockTokenizerManager{}
	service := NewTokenCountingServiceWithManager(mockManager)

	assert.NotNil(t, service, "Service should be created with custom manager")

	// Verify service works with custom manager
	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	}

	result, err := service.CountTokens(ctx, req)
	require.NoError(t, err)
	assert.Greater(t, result.TotalTokens, 0)
}

// RED: Write the smallest failing test first - logging integration
func TestTokenCountingService_LogsModelSelectionStart(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)
	assert.True(t, mockLogger.ContainsMessage("Starting model compatibility check"))
}

// RED: Test correlation ID propagation through logging
func TestTokenCountingService_LogsWithCorrelationID(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	correlationID := "test-correlation-123"
	ctx := logutil.WithCorrelationID(context.Background(), correlationID)

	req := TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)
	entries := mockLogger.GetLogEntriesByCorrelationID(correlationID)
	assert.NotEmpty(t, entries, "Should log with correlation ID")
}

// RED: Test detailed model evaluation logging per TODO.md requirements
func TestTokenCountingService_LogsModelEvaluationDetails(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "Analyze this code",
		Files: []FileContent{
			{Path: "test.go", Content: "package main\nfunc main() {}"},
		},
	}

	results, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// Should log each model evaluation with details as per TODO.md:
	// "Model {name} ({provider}, context: {window}) - {COMPATIBLE|SKIPPED}: {reason}"
	assert.True(t, mockLogger.ContainsMessage("Model evaluation:"))

	// Should log final summary as per TODO.md:
	// "Selected {count} compatible models: {names} (accuracy: {accurateCount} accurate, {estimatedCount} estimated)"
	assert.True(t, mockLogger.ContainsMessage("Model compatibility check completed"))
}

// RED: Test enhanced logging for initial model selection context as per TODO.md requirements
func TestTokenCountingService_LogsDetailedStartContext(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "Analyze this code carefully",
		Files: []FileContent{
			{Path: "main.go", Content: "package main\nfunc main() { fmt.Println(\"hello\") }"},
			{Path: "utils.go", Content: "package main\nfunc helper() {}"},
		},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai", "gemini"})

	require.NoError(t, err)

	// Should log enhanced start details as per TODO.md:
	// "Starting model selection with X accurate tokens from Y files using {tokenizer}"
	assert.True(t, mockLogger.ContainsMessage("Starting model compatibility check"))

	// Should include structured context information in the log message
	entries := mockLogger.GetLogEntries()
	startEntry := findLogEntryContaining(entries, "Starting model compatibility check")
	require.NotNil(t, startEntry, "Should have start log entry")

	// Should log with contextual information (since MockLogger formats args into message)
	message := startEntry.Message
	assert.Contains(t, message, "provider_count")
	assert.Contains(t, message, "file_count")
	assert.Contains(t, message, "has_instructions")
}

// Test checkModelCompatibility error paths to improve coverage
func TestTokenCountingService_CheckModelCompatibility_ErrorPaths(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	tests := []struct {
		name             string
		req              TokenCountingRequest
		modelName        string
		expectCompatible bool
		expectReason     bool
	}{
		{
			name: "model with insufficient context",
			req: TokenCountingRequest{
				Instructions: strings.Repeat("This is a very long instruction that will exceed the model's context window. ", 100000),
				Files: []FileContent{
					{Path: "large.txt", Content: strings.Repeat("Large file content. ", 100000)},
				},
			},
			modelName:        "gpt-4.1", // Has 1M context, but this should exceed it with safety margin
			expectCompatible: false,
			expectReason:     true,
		},
		{
			name: "model with sufficient context",
			req: TokenCountingRequest{
				Instructions: "Short instruction",
				Files: []FileContent{
					{Path: "small.txt", Content: "Small file"},
				},
			},
			modelName:        "gpt-4.1",
			expectCompatible: true,
			expectReason:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			results, err := service.GetCompatibleModels(ctx, tt.req, []string{"openai"})

			require.NoError(t, err)
			assert.NotEmpty(t, results)

			// Find the specific model in results
			var found bool
			for _, result := range results {
				if result.ModelName == tt.modelName {
					found = true
					assert.Equal(t, tt.expectCompatible, result.IsCompatible)
					if tt.expectReason {
						assert.NotEmpty(t, result.Reason, "Should have reason for incompatibility")
					} else {
						assert.Empty(t, result.Reason, "Compatible model should not have reason")
					}
					break
				}
			}
			assert.True(t, found, "Should find model %s in results", tt.modelName)
		})
	}
}

// Performance benchmarks as per TODO.md Phase 8.2 requirements
func BenchmarkTokenCountingService_CountTokens(b *testing.B) {
	service := NewTokenCountingService()
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions: "Analyze this large codebase and provide comprehensive optimization suggestions.",
		Files: []FileContent{
			{Path: "main.go", Content: strings.Repeat("package main\nfunc main() { fmt.Println(\"test\") }\n", 1000)},
			{Path: "utils.go", Content: strings.Repeat("func helper() { return }\n", 1000)},
			{Path: "api.go", Content: strings.Repeat("type API struct {}\n", 1000)},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.CountTokens(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenCountingService_GetCompatibleModels(b *testing.B) {
	service := NewTokenCountingService()
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions: "Analyze and optimize this code",
		Files: []FileContent{
			{Path: "example.go", Content: strings.Repeat("func example() {}\n", 500)},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetCompatibleModels(ctx, req, []string{"openai", "gemini"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenCountingService_CountTokensForModel_Accurate(b *testing.B) {
	service := NewTokenCountingService()
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions: "Perform detailed code analysis with comprehensive reporting",
		Files: []FileContent{
			{Path: "complex.go", Content: strings.Repeat("// Complex algorithm implementation\nfunc complexFunc() { /* implementation */ }\n", 200)},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.CountTokensForModel(ctx, req, "gpt-4.1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenCountingService_CountTokensForModel_Estimation(b *testing.B) {
	service := NewTokenCountingService()
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions: "Perform detailed code analysis with comprehensive reporting",
		Files: []FileContent{
			{Path: "complex.go", Content: strings.Repeat("// Complex algorithm implementation\nfunc complexFunc() { /* implementation */ }\n", 200)},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.CountTokensForModel(ctx, req, "gemini-2.5-pro")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Stress test with large file sets as per TODO.md requirements (>100 files, >1MB total)
func BenchmarkTokenCountingService_LargeFileSet(b *testing.B) {
	service := NewTokenCountingService()
	ctx := context.Background()

	// Create 150 files with varying sizes to exceed 1MB total
	files := make([]FileContent, 150)
	for i := 0; i < 150; i++ {
		files[i] = FileContent{
			Path:    fmt.Sprintf("file%d.go", i),
			Content: strings.Repeat(fmt.Sprintf("// File %d content\npackage main\nfunc file%dFunc() {}\n", i, i), 100),
		}
	}

	req := TokenCountingRequest{
		Instructions: "Analyze this large codebase comprehensively",
		Files:        files,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.CountTokens(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to find log entry containing specific message
func findLogEntryContaining(entries []testutil.LogEntry, message string) *testutil.LogEntry {
	for _, entry := range entries {
		if strings.Contains(entry.Message, message) {
			return &entry
		}
	}
	return nil
}
