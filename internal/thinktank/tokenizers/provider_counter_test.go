package tokenizers

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newNoOpLogger creates a simple no-op logger for testing
func newNoOpLogger() logutil.LoggerInterface {
	return logutil.NewLogger(logutil.DebugLevel, io.Discard, "")
}

func TestProviderTokenCounter_CountTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		text           string
		modelName      string
		expectedTokens int
		expectError    bool
		expectedMethod string
	}{
		{
			name:           "empty text returns zero tokens",
			text:           "",
			modelName:      "gpt-4.1",
			expectedTokens: 0,
			expectError:    false,
			expectedMethod: "tiktoken",
		},
		{
			name:           "openai model uses tiktoken",
			text:           "Hello world, this is a test message",
			modelName:      "gpt-4.1",
			expectedTokens: 9, // Approximate expected tokens for this text
			expectError:    false,
			expectedMethod: "tiktoken",
		},
		{
			name:           "gemini model uses sentencepiece",
			text:           "Hello world, this is a test message",
			modelName:      "gemini-2.5-pro",
			expectedTokens: 13, // Approximate expected tokens (different tokenization)
			expectError:    false,
			expectedMethod: "sentencepiece",
		},
		{
			name:           "openrouter model uses tiktoken-o200k",
			text:           "Hello world, this is a test message",
			modelName:      "openrouter/meta-llama/llama-4-maverick",
			expectedTokens: 8, // Accurate tokenization with o200k_base
			expectError:    false,
			expectedMethod: "tiktoken-o200k",
		},
		{
			name:           "unknown model falls back to estimation",
			text:           "Hello world, this is a test message",
			modelName:      "unknown-model",
			expectedTokens: 26, // 35 chars * 0.75 = ~26 tokens
			expectError:    false,
			expectedMethod: "estimation",
		},
		{
			name:           "o3 model uses tiktoken",
			text:           "Hello world",
			modelName:      "o3",
			expectedTokens: 2, // "Hello" and "world" as separate tokens
			expectError:    false,
			expectedMethod: "tiktoken",
		},
		{
			name:           "o4-mini model uses tiktoken",
			text:           "Hello world",
			modelName:      "o4-mini",
			expectedTokens: 2, // "Hello" and "world" as separate tokens
			expectError:    false,
			expectedMethod: "tiktoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newNoOpLogger()
			counter := NewProviderTokenCounter(logger)

			tokens, err := counter.CountTokens(context.Background(), tt.text, tt.modelName)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// For estimation, check exact match. For accurate tokenizers, allow some variance
				if tt.expectedMethod == "estimation" {
					assert.Equal(t, tt.expectedTokens, tokens, "Expected exact token count for estimation")
				} else {
					// Accurate tokenizers may vary slightly, allow reasonable range
					assert.InDelta(t, tt.expectedTokens, tokens, float64(tt.expectedTokens)*0.3,
						"Token count should be within 30%% of expected for accurate tokenizers")
				}
			}
		})
	}
}

func TestProviderTokenCounter_SupportsModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		modelName string
		expected  bool
	}{
		{
			name:      "openai model is supported",
			modelName: "gpt-4.1",
			expected:  true,
		},
		{
			name:      "gemini model is supported",
			modelName: "gemini-2.5-pro",
			expected:  true,
		},
		{
			name:      "openrouter model is supported (uses tiktoken-o200k)",
			modelName: "openrouter/meta-llama/llama-4-maverick",
			expected:  true,
		},
		{
			name:      "unknown model is not supported",
			modelName: "unknown-model",
			expected:  false,
		},
		{
			name:      "o3 model is supported",
			modelName: "o3",
			expected:  true,
		},
		{
			name:      "o4-mini model is supported",
			modelName: "o4-mini",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newNoOpLogger()
			counter := NewProviderTokenCounter(logger)

			result := counter.SupportsModel(tt.modelName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProviderTokenCounter_GetEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		modelName     string
		expectedStart string // Check prefix since exact encoding may vary
	}{
		{
			name:          "openai model returns tiktoken encoding",
			modelName:     "gpt-4.1",
			expectedStart: "tiktoken:",
		},
		{
			name:          "gemini model returns sentencepiece encoding",
			modelName:     "gemini-2.5-pro",
			expectedStart: "sentencepiece:",
		},
		{
			name:          "openrouter model returns o200k_base",
			modelName:     "openrouter/meta-llama/llama-4-maverick",
			expectedStart: "o200k_base",
		},
		{
			name:          "unknown model returns estimation",
			modelName:     "unknown-model",
			expectedStart: "estimation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newNoOpLogger()
			counter := NewProviderTokenCounter(logger)

			encoding, err := counter.GetEncoding(tt.modelName)
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(encoding, tt.expectedStart),
				"Expected encoding to start with %s, got %s", tt.expectedStart, encoding)
		})
	}
}

func TestProviderTokenCounter_GetTokenizerType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		modelName    string
		expectedType string
	}{
		{
			name:         "openai model returns tiktoken",
			modelName:    "gpt-4.1",
			expectedType: "tiktoken",
		},
		{
			name:         "gemini model returns sentencepiece",
			modelName:    "gemini-2.5-pro",
			expectedType: "sentencepiece",
		},
		{
			name:         "openrouter model returns tiktoken-o200k",
			modelName:    "openrouter/meta-llama/llama-4-maverick",
			expectedType: "tiktoken-o200k",
		},
		{
			name:         "unknown model returns estimation",
			modelName:    "unknown-model",
			expectedType: "estimation",
		},
		{
			name:         "o3 model returns tiktoken",
			modelName:    "o3",
			expectedType: "tiktoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newNoOpLogger()
			counter := NewProviderTokenCounter(logger)

			result := counter.GetTokenizerType(tt.modelName)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

func TestProviderTokenCounter_IsAccurate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		modelName    string
		expectedType bool
	}{
		{
			name:         "openai model is accurate",
			modelName:    "gpt-4.1",
			expectedType: true,
		},
		{
			name:         "gemini model is accurate",
			modelName:    "gemini-2.5-pro",
			expectedType: true,
		},
		{
			name:         "openrouter model is accurate",
			modelName:    "openrouter/meta-llama/llama-4-maverick",
			expectedType: true,
		},
		{
			name:         "unknown model is not accurate",
			modelName:    "unknown-model",
			expectedType: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newNoOpLogger()
			counter := NewProviderTokenCounter(logger)

			result := counter.IsAccurate(tt.modelName)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

func TestProviderTokenCounter_LazyLoading(t *testing.T) {
	t.Parallel()

	logger := newNoOpLogger()
	counter := NewProviderTokenCounter(logger)

	// Initially, tokenizers should be nil (lazy loading)
	assert.Nil(t, counter.tiktoken)
	assert.Nil(t, counter.sentencePiece)
	assert.NotNil(t, counter.fallback)

	// After using OpenAI model, tiktoken should be initialized
	_, err := counter.CountTokens(context.Background(), "test", "gpt-4.1")
	require.NoError(t, err)
	assert.NotNil(t, counter.tiktoken)
	assert.Nil(t, counter.sentencePiece) // Still not initialized

	// After using Gemini model, sentencepiece should be initialized
	_, err = counter.CountTokens(context.Background(), "test", "gemini-2.5-pro")
	require.NoError(t, err)
	assert.NotNil(t, counter.tiktoken)
	assert.NotNil(t, counter.sentencePiece) // Now initialized
}

func TestProviderTokenCounter_ClearCache(t *testing.T) {
	t.Parallel()

	logger := newNoOpLogger()
	counter := NewProviderTokenCounter(logger)

	// Initialize tokenizers by using them
	_, err := counter.CountTokens(context.Background(), "test", "gpt-4.1")
	require.NoError(t, err)
	_, err = counter.CountTokens(context.Background(), "test", "gemini-2.5-pro")
	require.NoError(t, err)

	// Verify tokenizers are initialized
	assert.NotNil(t, counter.tiktoken)
	assert.NotNil(t, counter.sentencePiece)

	// Clear cache
	counter.ClearCache()

	// Verify tokenizers are cleared
	assert.Nil(t, counter.tiktoken)
	assert.Nil(t, counter.sentencePiece)
}

func TestEstimationTokenCounter(t *testing.T) {
	t.Parallel()

	counter := NewEstimationTokenCounter()

	tests := []struct {
		name           string
		text           string
		modelName      string
		expectedTokens int
	}{
		{
			name:           "empty text returns zero",
			text:           "",
			modelName:      "any-model",
			expectedTokens: 0,
		},
		{
			name:           "simple text calculation",
			text:           "Hello world", // 11 chars * 0.75 = 8.25 → 8
			modelName:      "any-model",
			expectedTokens: 8,
		},
		{
			name:           "longer text calculation",
			text:           "This is a longer test message with more characters", // 50 chars * 0.75 = 37.5 → 37
			modelName:      "any-model",
			expectedTokens: 37,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := counter.CountTokens(context.Background(), tt.text, tt.modelName)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTokens, tokens)
		})
	}

	// Test that it supports all models
	assert.True(t, counter.SupportsModel("any-model"))
	assert.True(t, counter.SupportsModel("unknown-model"))

	// Test encoding
	encoding, err := counter.GetEncoding("any-model")
	require.NoError(t, err)
	assert.Equal(t, "estimation", encoding)
}

func TestProviderTokenCounter_Logging(t *testing.T) {
	t.Parallel()

	// Create a test logger to capture log messages
	testLogger := logutil.NewTestLoggerWithoutAutoFail(t)

	counter := NewProviderTokenCounter(testLogger)

	// Test successful tokenization logging
	_, err := counter.CountTokens(context.Background(), "test message", "gpt-4.1")
	require.NoError(t, err)

	// Should have logged tokenizer selection (debug level)
	logs := testLogger.GetTestLogs()
	assert.Greater(t, len(logs), 0, "Should have debug messages for tokenizer selection")

	// Test fallback logging with unknown model
	_, err = counter.CountTokens(context.Background(), "test message", "unknown-model")
	require.NoError(t, err)

	// Should have logged provider detection failure (warn level)
	logs = testLogger.GetTestLogs()
	foundWarnLog := false
	for _, log := range logs {
		if strings.Contains(log, "[WARN]") && strings.Contains(log, "Provider detection failed") {
			foundWarnLog = true
			break
		}
	}
	assert.True(t, foundWarnLog, "Should have warn messages for provider detection failure")
}
