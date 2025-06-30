package tokenizers

import (
	"context"
	"io"
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

// TestCalculateStreamingTimeout_ReflectsCIPerformance validates that timeout calculations
// reflect realistic CI performance expectations based on empirical evidence
func TestCalculateStreamingTimeout_ReflectsCIPerformance(t *testing.T) {
	// Based on actual CI measurements with race detection:
	// - 1MB in 2.29s = 0.44 MB/s
	// - 10MB in 20.83s = 0.48 MB/s
	// - 20MB in 39.79s = 0.50 MB/s
	// Average performance is ~0.47 MB/s, so 0.5 MB/s is a realistic expectation

	tests := []struct {
		name               string
		inputSizeBytes     int
		maxExpectedTimeout time.Duration
		reason             string
	}{
		{
			name:               "20MB_should_use_realistic_0.5_MBps",
			inputSizeBytes:     20 * 1024 * 1024,
			maxExpectedTimeout: 75 * time.Second, // 20MB ÷ 0.5MB/s = 40s + 30s buffer = 70s (allow 75s)
			reason:             "With 0.5 MB/s realistic performance, 20MB should timeout at ~70s, not 80s+",
		},
		{
			name:               "50MB_should_use_realistic_0.5_MBps",
			inputSizeBytes:     50 * 1024 * 1024,
			maxExpectedTimeout: 135 * time.Second, // 50MB ÷ 0.5MB/s = 100s + 30s buffer = 130s (allow 135s)
			reason:             "With 0.5 MB/s realistic performance, 50MB should timeout at ~130s, not 155s+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := calculateStreamingTimeout(tt.inputSizeBytes)

			assert.LessOrEqual(t, timeout, tt.maxExpectedTimeout,
				"Timeout %v should be <= %v. %s", timeout, tt.maxExpectedTimeout, tt.reason)

			t.Logf("Input: %d bytes, Timeout: %v, Reason: %s",
				tt.inputSizeBytes, timeout, tt.reason)
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
			expectedMinimum: 65 * time.Second, // Should be ~70s (40s + 30s buffer)
			expectedMaximum: 75 * time.Second,
		},
		{
			name:            "Very_Large_50MB",
			inputSizeBytes:  50 * 1024 * 1024,
			expectedMinimum: 125 * time.Second, // Should be ~130s (100s + 30s buffer)
			expectedMaximum: 135 * time.Second,
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
	// Realistic performance expectation based on CI evidence:
	// - Local without race: ~0.5 MB/s
	// - CI with race detection: ~0.5 MB/s (measured 0.44-0.50 MB/s actual)
	// Using 0.5 MB/s based on empirical CI measurements
	const bytesPerSecond = 512 * 1024 // 0.5 MB/s (empirical CI performance)
	const bufferSeconds = 30          // Additional safety buffer
	const minTimeoutSeconds = 60      // Minimum timeout regardless of size

	expectedSeconds := inputSizeBytes / bytesPerSecond
	timeoutSeconds := expectedSeconds + bufferSeconds

	if timeoutSeconds < minTimeoutSeconds {
		timeoutSeconds = minTimeoutSeconds
	}

	return time.Duration(timeoutSeconds) * time.Second
}

// TestGetChunkSizeForInput_20MBBoundary tests the specific 20MB boundary behavior
// This test expects 20MB to use large chunks (64KB) not medium chunks (32KB)
func TestGetChunkSizeForInput_20MBBoundary(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
	require.NoError(t, err)

	// Cast to access chunk size method
	adaptiveTokenizer, ok := streamingTokenizer.(interface {
		GetChunkSizeForInput(inputSizeBytes int) int
	})
	require.True(t, ok, "Streaming tokenizer must implement adaptive chunking interface")

	// Test exactly at 20MB boundary - should use large chunks (64KB)
	// With current 25MB threshold, this will FAIL because 20MB < 25MB returns 32KB
	actualChunk := adaptiveTokenizer.GetChunkSizeForInput(20 * 1024 * 1024) // 20MB
	expectedChunk := 64 * 1024                                              // 64KB for large inputs

	assert.Equal(t, expectedChunk, actualChunk,
		"20MB should use 64KB chunks (large), not 32KB chunks (medium). "+
			"This indicates the threshold should be 20MB, not 25MB")
}

// TestStreamingTokenizer_AdaptsChunkSizeBasedOnInputSize tests that the streaming tokenizer
// uses appropriate chunk sizes based on input size for optimal performance
func TestStreamingTokenizer_AdaptsChunkSizeBasedOnInputSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inputSizeBytes int
		expectedChunk  int
	}{
		{
			name:           "Small_input_1MB_uses_8KB_chunks",
			inputSizeBytes: 1024 * 1024, // 1MB
			expectedChunk:  8 * 1024,    // 8KB
		},
		{
			name:           "Medium_input_10MB_uses_32KB_chunks",
			inputSizeBytes: 10 * 1024 * 1024, // 10MB
			expectedChunk:  32 * 1024,        // 32KB
		},
		{
			name:           "Large_input_20MB_uses_64KB_chunks",
			inputSizeBytes: 20 * 1024 * 1024, // 20MB (boundary case)
			expectedChunk:  64 * 1024,        // 64KB
		},
		{
			name:           "Large_input_50MB_uses_64KB_chunks",
			inputSizeBytes: 50 * 1024 * 1024, // 50MB
			expectedChunk:  64 * 1024,        // 64KB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will FAIL initially (RED phase) - need GetChunkSizeForInput method
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
			require.NoError(t, err)

			// Cast to implementation to access chunk size method
			if adaptiveTokenizer, ok := streamingTokenizer.(interface {
				GetChunkSizeForInput(inputSizeBytes int) int
			}); ok {
				actualChunk := adaptiveTokenizer.GetChunkSizeForInput(tt.inputSizeBytes)
				assert.Equal(t, tt.expectedChunk, actualChunk,
					"Input size %d bytes should use %d byte chunks, got %d",
					tt.inputSizeBytes, tt.expectedChunk, actualChunk)
			} else {
				t.Fatal("Streaming tokenizer does not implement adaptive chunking interface")
			}
		})
	}
}

// TestStreamingTokenization_UsesBetterPerformanceWithAdaptiveChunking tests that
// adaptive chunking provides better performance for large inputs
func TestStreamingTokenization_UsesBetterPerformanceWithAdaptiveChunking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inputSize      int
		expectedFaster bool // Whether adaptive chunking should be faster
	}{
		{
			name:           "Large_input_20MB_benefits_from_adaptive_chunking",
			inputSize:      20 * 1024 * 1024, // 20MB
			expectedFaster: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
			require.NoError(t, err)

			// Create large text for testing
			text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", tt.inputSize/46)

			// Test regular streaming tokenization
			reader1 := strings.NewReader(text)
			ctx1, cancel1 := context.WithTimeout(context.Background(), calculateStreamingTimeout(tt.inputSize))
			defer cancel1()

			start1 := time.Now()
			tokens1, err := streamingTokenizer.CountTokensStreaming(ctx1, reader1, "gpt-4")
			duration1 := time.Since(start1)
			require.NoError(t, err)

			// Test adaptive chunking method
			if adaptiveTokenizer, ok := streamingTokenizer.(interface {
				CountTokensStreamingWithAdaptiveChunking(ctx context.Context, reader io.Reader, modelName string, inputSizeBytes int) (int, error)
			}); ok {
				reader2 := strings.NewReader(text)
				ctx2, cancel2 := context.WithTimeout(context.Background(), calculateStreamingTimeout(tt.inputSize))
				defer cancel2()

				start2 := time.Now()
				tokens2, err := adaptiveTokenizer.CountTokensStreamingWithAdaptiveChunking(ctx2, reader2, "gpt-4", tt.inputSize)
				duration2 := time.Since(start2)
				require.NoError(t, err)

				// Verify token counts are reasonably close (chunking boundaries may cause small differences)
				// Allow up to 1% difference due to tokenization boundary effects
				tokenDiff := float64(abs(tokens1-tokens2)) / float64(tokens1)
				assert.LessOrEqual(t, tokenDiff, 0.01,
					"Token counts should be within 1%% (got %d vs %d, diff: %.2f%%)",
					tokens1, tokens2, tokenDiff*100)
				t.Logf("Token counts: regular=%d, adaptive=%d, diff=%.2f%%", tokens1, tokens2, tokenDiff*100)

				// Log performance comparison
				t.Logf("Regular streaming: %v (%.2f MB/s)", duration1, float64(tt.inputSize)/(1024*1024)/duration1.Seconds())
				t.Logf("Adaptive chunking: %v (%.2f MB/s)", duration2, float64(tt.inputSize)/(1024*1024)/duration2.Seconds())

				if tt.expectedFaster {
					// Adaptive chunking should be faster or at least not significantly slower
					// Allow for some variance in timing, but adaptive should be within 120% of regular
					performanceRatio := float64(duration2) / float64(duration1)
					assert.LessOrEqual(t, performanceRatio, 1.2,
						"Adaptive chunking should not be significantly slower (ratio: %.2f)", performanceRatio)
					t.Logf("Performance ratio (adaptive/regular): %.2f", performanceRatio)
				}
			} else {
				t.Fatal("Streaming tokenizer does not implement adaptive chunking interface")
			}
		})
	}
}

// BenchmarkAdaptiveChunkingPerformance benchmarks adaptive chunking vs regular streaming
func BenchmarkAdaptiveChunkingPerformance(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"5MB", 5 * 1024 * 1024},
		{"20MB", 20 * 1024 * 1024},
		{"50MB", 50 * 1024 * 1024},
	}

	for _, size := range sizes {
		text := strings.Repeat("Benchmark content for adaptive chunking performance testing. ", size.size/64)

		b.Run("Regular_"+size.name, func(b *testing.B) {
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

		b.Run("Adaptive_"+size.name, func(b *testing.B) {
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openai")
			require.NoError(b, err)

			adaptiveTokenizer := streamingTokenizer.(interface {
				CountTokensStreamingWithAdaptiveChunking(ctx context.Context, reader io.Reader, modelName string, inputSizeBytes int) (int, error)
			})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(text)
				_, err := adaptiveTokenizer.CountTokensStreamingWithAdaptiveChunking(context.Background(), reader, "gpt-4", size.size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Helper function to format byte sizes for benchmark names
func formatSize(bytes int) string {
	if bytes >= 1024*1024 {
		return string(rune('0'+bytes/(1024*1024))) + "MB"
	}
	return string(rune('0'+bytes/1024)) + "KB"
}

// Helper function to calculate absolute difference
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
