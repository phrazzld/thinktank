package tokenizers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamingTokenization_HandlesLargeInputs tests that streaming tokenization
// can handle very large inputs efficiently without loading them fully into memory
func TestStreamingTokenization_HandlesLargeInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputSize   int
		chunkSize   int
		expectError bool
	}{
		{
			name:        "Medium input 1MB in 64KB chunks",
			inputSize:   1024 * 1024, // 1MB
			chunkSize:   64 * 1024,   // 64KB chunks
			expectError: false,
		},
		{
			name:        "Large input 10MB in 256KB chunks",
			inputSize:   10 * 1024 * 1024, // 10MB
			chunkSize:   256 * 1024,       // 256KB chunks
			expectError: false,
		},
		{
			name:        "Large input 20MB in 1MB chunks",
			inputSize:   20 * 1024 * 1024, // 20MB (realistic large file size)
			chunkSize:   1024 * 1024,      // 1MB chunks
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create streaming tokenizer manager - this will FAIL (RED phase)
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
			require.NoError(t, err)

			// Create a large text stream
			text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", tt.inputSize/46)
			reader := strings.NewReader(text)

			// Set timeout based on realistic CI performance expectations
			timeout := calculateStreamingTimeout(tt.inputSize)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			start := time.Now()
			tokens, err := streamingTokenizer.CountTokensStreaming(ctx, reader, "gpt-4")
			duration := time.Since(start)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "Streaming tokenization should handle large inputs")
				assert.Greater(t, tokens, 0, "Should return positive token count")
				t.Logf("Processed %d bytes in %v (%.2f MB/s)",
					tt.inputSize, duration, float64(tt.inputSize)/(1024*1024)/duration.Seconds())
			}
		})
	}
}

// TestStreamingTokenization_MatchesInMemoryResults tests that streaming tokenization
// produces the same results as in-memory tokenization for smaller inputs
func TestStreamingTokenization_MatchesInMemoryResults(t *testing.T) {
	t.Parallel()

	testInputs := []string{
		"Short text",
		"Medium length text with multiple sentences. This should be tokenized consistently.",
		strings.Repeat("Longer repeated text pattern. ", 100),
	}

	for i, input := range testInputs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// This will FAIL initially (RED phase) - need streaming implementation
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
			require.NoError(t, err)

			// Test streaming approach
			reader := strings.NewReader(input)
			streamingTokens, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4")
			require.NoError(t, err)

			// Test in-memory approach
			regularTokenizer, err := manager.GetTokenizer("openai")
			require.NoError(t, err)
			inMemoryTokens, err := regularTokenizer.CountTokens(context.Background(), input, "gpt-4")
			require.NoError(t, err)

			// Results should match
			assert.Equal(t, inMemoryTokens, streamingTokens,
				"Streaming and in-memory tokenization should produce identical results")
		})
	}
}

// TestStreamingTokenization_RespectsContextCancellation tests that streaming
// tokenization properly handles context cancellation for long-running operations
func TestStreamingTokenization_RespectsContextCancellation(t *testing.T) {
	t.Parallel()

	// This will FAIL initially (RED phase) - need streaming implementation
	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
	require.NoError(t, err)

	// Create a large text stream that would take a while to process
	largeText := strings.Repeat("Text that takes time to process. ", 10000) // ~300KB (smaller for faster cancellation)
	reader := strings.NewReader(largeText)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = streamingTokenizer.CountTokensStreaming(ctx, reader, "gpt-4")
	duration := time.Since(start)

	assert.Error(t, err, "Should return error when context is cancelled")
	assert.LessOrEqual(t, duration, 200*time.Millisecond, "Should cancel reasonably quickly")
	assert.Equal(t, context.DeadlineExceeded, err, "Should return deadline exceeded error")
}

// BenchmarkStreamingVsInMemory benchmarks streaming vs in-memory tokenization
// to validate that streaming doesn't have excessive overhead for medium inputs
func BenchmarkStreamingVsInMemory(b *testing.B) {
	sizes := []int{
		100 * 1024,       // 100KB
		1024 * 1024,      // 1MB
		10 * 1024 * 1024, // 10MB
	}

	for _, size := range sizes {
		sizeStr := formatSize(size)
		text := strings.Repeat("Benchmark text content. ", size/24)

		b.Run("InMemory_"+sizeStr, func(b *testing.B) {
			manager := NewTokenizerManager()
			tokenizer, err := manager.GetTokenizer("openai")
			require.NoError(b, err)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := tokenizer.CountTokens(context.Background(), text, "gpt-4")
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("Streaming_"+sizeStr, func(b *testing.B) {
			// This will FAIL initially (RED phase) - need streaming implementation
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
			require.NoError(b, err)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(text)
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestCalculateStreamingTimeout validates timeout calculation for various input sizes
func TestCalculateStreamingTimeout(t *testing.T) {
	tests := []struct {
		name            string
		inputSizeBytes  int
		expectedMinimum time.Duration
		expectedMaximum time.Duration
	}{
		{
			name:            "Small_1MB",
			inputSizeBytes:  1024 * 1024,
			expectedMinimum: 60 * time.Second, // Should use minimum
			expectedMaximum: 60 * time.Second,
		},
		{
			name:            "Medium_10MB",
			inputSizeBytes:  10 * 1024 * 1024,
			expectedMinimum: 60 * time.Second, // Should use minimum (25s + 30s buffer = 55s < 60s)
			expectedMaximum: 60 * time.Second,
		},
		{
			name:            "Large_20MB",
			inputSizeBytes:  20 * 1024 * 1024,
			expectedMinimum: 75 * time.Second, // Should be ~80s (50s + 30s buffer)
			expectedMaximum: 85 * time.Second,
		},
		{
			name:            "Very_Large_50MB",
			inputSizeBytes:  50 * 1024 * 1024,
			expectedMinimum: 155 * time.Second, // Should be ~158s (125s + 30s buffer)
			expectedMaximum: 165 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := calculateStreamingTimeout(tt.inputSizeBytes)

			assert.GreaterOrEqual(t, timeout, tt.expectedMinimum,
				"Timeout %v should be at least %v for %d bytes",
				timeout, tt.expectedMinimum, tt.inputSizeBytes)

			assert.LessOrEqual(t, timeout, tt.expectedMaximum,
				"Timeout %v should be at most %v for %d bytes",
				timeout, tt.expectedMaximum, tt.inputSizeBytes)

			t.Logf("%d bytes → %v timeout", tt.inputSizeBytes, timeout)
		})
	}
}

// calculateStreamingTimeout calculates realistic timeout for streaming tokenization
// Based on empirical performance data from CI with race detection enabled
func calculateStreamingTimeout(inputSizeBytes int) time.Duration {
	// Conservative performance expectation based on CI evidence:
	// - Local without race: ~0.5 MB/s
	// - CI with race detection: ~0.4 MB/s
	// Using 0.4 MB/s to handle worst-case CI environment
	const bytesPerSecond = 400 * 1024 // 0.4 MB/s (empirical CI performance)
	const bufferSeconds = 30          // Additional safety buffer
	const minTimeoutSeconds = 60      // Minimum timeout regardless of size

	expectedSeconds := inputSizeBytes / bytesPerSecond
	timeoutSeconds := expectedSeconds + bufferSeconds

	if timeoutSeconds < minTimeoutSeconds {
		timeoutSeconds = minTimeoutSeconds
	}

	return time.Duration(timeoutSeconds) * time.Second
}

// Helper function to format byte sizes for benchmark names
func formatSize(bytes int) string {
	if bytes >= 1024*1024 {
		return string(rune('0'+bytes/(1024*1024))) + "MB"
	}
	return string(rune('0'+bytes/1024)) + "KB"
}
