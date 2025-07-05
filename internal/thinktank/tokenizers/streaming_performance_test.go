package tokenizers

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamingTokenizer_MeasuresBasicThroughput tests the fundamental
// performance behavior of streaming tokenization
func TestStreamingTokenizer_MeasuresBasicThroughput(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Test the smallest meaningful performance unit
	text := strings.Repeat("word ", 1000) // 5KB of predictable content
	reader := strings.NewReader(text)

	start := time.Now()
	tokens, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
	duration := time.Since(start)

	require.NoError(t, err)
	require.Greater(t, tokens, 0, "Should return positive token count")

	// The behavior we're driving: streaming should be measurably fast
	throughputBytesPerSec := float64(len(text)) / duration.Seconds()

	// This establishes our performance contract - streaming should be reasonably fast
	// Use a conservative threshold that works with and without race detection
	minThroughput := 10 * 1024.0 // 10KB/s minimum (conservative for all environments)

	assert.Greater(t, throughputBytesPerSec, minThroughput,
		"Streaming tokenizer should process at least %.0f KB/s, got %.2f bytes/s",
		minThroughput/1024, throughputBytesPerSec)

	// If we're achieving good performance already, ensure it's meaningfully fast
	if throughputBytesPerSec > 1024*1024.0 { // 1MB/s
		t.Logf("✅ Excellent performance: %.2f MB/s", throughputBytesPerSec/(1024*1024))
	} else if throughputBytesPerSec > 512*1024.0 { // 512KB/s
		t.Logf("✅ Good performance: %.2f KB/s", throughputBytesPerSec/1024)
	} else {
		t.Logf("⚠️  Performance needs improvement: %.2f KB/s", throughputBytesPerSec/1024)
	}

	t.Logf("Throughput: %.2f MB/s for %d bytes",
		throughputBytesPerSec/(1024*1024), len(text))
}

// TestStreamingTokenizer_ScalesLinearlyWithInputSize tests that streaming
// performance scales predictably with input size
func TestStreamingTokenizer_ScalesLinearlyWithInputSize(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Test with progressively larger inputs to understand scaling behavior
	sizes := []int{
		5 * 1024,   // 5KB
		50 * 1024,  // 50KB
		500 * 1024, // 500KB
	}

	type measurement struct {
		size       int
		duration   time.Duration
		throughput float64
	}

	var measurements []measurement

	for _, size := range sizes {
		text := strings.Repeat("a", size)
		reader := strings.NewReader(text)

		start := time.Now()
		_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
		duration := time.Since(start)
		require.NoError(t, err)

		throughput := float64(size) / duration.Seconds()
		measurements = append(measurements, measurement{size, duration, throughput})

		t.Logf("Size %s: %.2f KB/s", formatSizeString(size), throughput/1024)
	}

	// The behavior: throughput should remain roughly stable (linear scaling)
	// If throughput degrades significantly, that indicates scaling problems
	baselineThroughput := measurements[0].throughput

	for i := 1; i < len(measurements); i++ {
		throughputRatio := measurements[i].throughput / baselineThroughput

		// Allow some variance but flag major degradation
		// This will FAIL if streaming doesn't scale linearly
		assert.GreaterOrEqual(t, throughputRatio, 0.3, // Allow up to 70% degradation
			"Throughput degraded significantly. Size %s vs %s: %.2fx ratio (%.2f KB/s vs %.2f KB/s)",
			formatSizeString(measurements[i].size), formatSizeString(measurements[0].size),
			throughputRatio, measurements[i].throughput/1024, measurements[0].throughput/1024)

		if throughputRatio >= 0.8 {
			t.Logf("✅ Good scaling: %s maintains %.1f%% of baseline throughput",
				formatSizeString(measurements[i].size), throughputRatio*100)
		} else {
			t.Logf("⚠️  Scaling concern: %s has %.1f%% of baseline throughput",
				formatSizeString(measurements[i].size), throughputRatio*100)
		}
	}
}

// formatSizeString helper function for test output
func formatSizeString(bytes int) string {
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.0fKB", float64(bytes)/1024)
}

// TestStreamingTokenizer_ConstantMemoryUsage tests that streaming
// tokenization uses constant memory regardless of input size
func TestStreamingTokenizer_ConstantMemoryUsage(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Test with increasingly large inputs to check memory scaling
	sizes := []int{
		100 * 1024,      // 100KB
		1024 * 1024,     // 1MB
		5 * 1024 * 1024, // 5MB
	}

	var memoryUsages []int64

	for _, size := range sizes {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		text := strings.Repeat("content ", size/8)
		reader := strings.NewReader(text)

		_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
		require.NoError(t, err)

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := int64(m2.Alloc - m1.Alloc)
		memoryUsages = append(memoryUsages, memoryIncrease)

		t.Logf("Size %s: memory increase %s",
			formatSizeString(size), formatSizeString(int(memoryIncrease)))
	}

	// The behavior: memory usage should NOT scale linearly with input size
	// Streaming should use constant memory regardless of input size
	baselineMemory := memoryUsages[0]

	for i := 1; i < len(memoryUsages); i++ {
		memoryRatio := float64(memoryUsages[i]) / float64(baselineMemory)
		inputSizeRatio := float64(sizes[i]) / float64(sizes[0])

		// Memory growth should be much less than input size growth
		// This will FAIL if we're loading entire input into memory
		assert.Less(t, memoryRatio, inputSizeRatio*0.5,
			"Memory usage should not scale with input size. Got %.1fx memory increase for %.1fx input increase",
			memoryRatio, inputSizeRatio)

		if memoryRatio <= 2.0 {
			t.Logf("✅ Good memory efficiency: %.1fx memory for %.1fx input",
				memoryRatio, inputSizeRatio)
		} else {
			t.Logf("⚠️  Memory scaling concern: %.1fx memory for %.1fx input",
				memoryRatio, inputSizeRatio)
		}
	}
}

// BenchmarkStreamingTokenizer_ThroughputByInputSize provides comprehensive
// throughput benchmarks across different input sizes for regression detection
func BenchmarkStreamingTokenizer_ThroughputByInputSize(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
			require.NoError(b, err)

			// Generate predictable test content
			text := generatePredictableContent(size.size)

			b.SetBytes(int64(size.size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(text)
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkStreamingTokenizer_MemoryAllocation tracks memory allocation
// patterns to detect memory leaks and optimization opportunities
func BenchmarkStreamingTokenizer_MemoryAllocation(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_10KB", 10 * 1024},
		{"Medium_1MB", 1024 * 1024},
		{"Large_10MB", 10 * 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
			require.NoError(b, err)

			text := generatePredictableContent(size.size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(text)
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkStreamingTokenizer_AdaptiveVsRegular compares adaptive chunking
// performance against regular streaming to validate optimization benefits
func BenchmarkStreamingTokenizer_AdaptiveVsRegular(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"5MB", 5 * 1024 * 1024},
		{"20MB", 20 * 1024 * 1024},
		{"50MB", 50 * 1024 * 1024},
	}

	for _, size := range sizes {
		// Regular streaming benchmark
		b.Run("Regular_"+size.name, func(b *testing.B) {
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
			require.NoError(b, err)

			text := generatePredictableContent(size.size)

			b.SetBytes(int64(size.size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(text)
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// Adaptive chunking benchmark
		b.Run("Adaptive_"+size.name, func(b *testing.B) {
			manager := NewStreamingTokenizerManager()
			streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
			require.NoError(b, err)

			// Cast to access adaptive chunking functionality
			adaptiveTokenizer, ok := streamingTokenizer.(interface {
				CountTokensStreamingWithAdaptiveChunking(ctx context.Context, reader io.Reader, modelName string, inputSizeBytes int) (int, error)
			})
			if !ok {
				b.Skip("Adaptive chunking not available")
			}

			text := generatePredictableContent(size.size)

			b.SetBytes(int64(size.size))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(text)
				_, err := adaptiveTokenizer.CountTokensStreamingWithAdaptiveChunking(
					context.Background(), reader, "gpt-4.1", size.size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// generatePredictableContent creates test content of the specified size
// with predictable tokenization characteristics for consistent benchmarking
func generatePredictableContent(sizeBytes int) string {
	// Use a repeating pattern that tokenizes predictably
	pattern := "The quick brown fox jumps over the lazy dog. This is test content for streaming tokenizer benchmarks. "
	patternSize := len(pattern)

	if sizeBytes <= patternSize {
		return pattern[:sizeBytes]
	}

	repeats := sizeBytes / patternSize
	remaining := sizeBytes % patternSize

	result := strings.Repeat(pattern, repeats)
	if remaining > 0 {
		result += pattern[:remaining]
	}

	return result
}

// raceDetectionEnabled detects if Go race detection is enabled
// This helps provide better context in test logging for CI debugging
func raceDetectionEnabled() bool {
	// Detection strategy: Use safe methods that don't create races

	// Method 1: Environment variable check (can be set by CI)
	if val, exists := os.LookupEnv("RACE_ENABLED"); exists && val == "true" {
		return true
	}

	// Method 2: Check command line arguments
	for _, arg := range os.Args {
		if arg == "-test.race" || arg == "-race" {
			return true
		}
	}

	// Method 3: Simple performance heuristic without creating actual races
	// Race detection typically slows down all operations significantly
	start := time.Now()

	// Perform CPU-intensive work that race detector will slow down
	sum := 0
	for i := 0; i < 100000; i++ {
		sum += i * i
	}

	duration := time.Since(start)

	// Race detector typically makes operations 2-10x slower
	// This is a conservative threshold based on empirical observation
	isUnusuallySlowExecution := duration > 50*time.Millisecond

	// Method 4: Check for race-specific environment hints
	_, hasGoRaceEnv := os.LookupEnv("GORACE")

	return isUnusuallySlowExecution || hasGoRaceEnv
}

// TestStreamingTokenizer_PerformanceEnvelope validates that streaming tokenizer
// performance stays within acceptable bounds for regression detection
func TestStreamingTokenizer_PerformanceEnvelope(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Define performance envelope - minimum acceptable performance levels
	// Uses environment-aware timeouts to account for race detection overhead
	tests := []struct {
		name              string
		size              int
		minThroughputKBps float64 // KB/s
		maxMemoryMB       int64   // MB
	}{
		{
			name:              "1MB_performance_envelope",
			size:              1024 * 1024,
			minThroughputKBps: 100, // 100 KB/s minimum
			maxMemoryMB:       200, // 200MB max memory
		},
		{
			name:              "10MB_performance_envelope",
			size:              10 * 1024 * 1024,
			minThroughputKBps: 200, // 200 KB/s minimum
			maxMemoryMB:       500, // 500MB max memory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			text := generatePredictableContent(tt.size)
			reader := strings.NewReader(text)

			// Use environment-aware timeout that accounts for race detection overhead
			// This replaces the hardcoded timeouts that were causing CI failures
			maxDuration := calculateStreamingTimeout(tt.size)
			isRaceDetectionEnabled := raceDetectionEnabled()

			t.Logf("Environment context: race_detection=%v, max_duration=%v, input_size=%s",
				isRaceDetectionEnabled, maxDuration, formatSizeString(tt.size))

			start := time.Now()
			tokens, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-4.1")
			duration := time.Since(start)

			runtime.GC()
			runtime.ReadMemStats(&m2)

			require.NoError(t, err)
			require.Greater(t, tokens, 0)

			// Performance envelope validation
			throughputKBps := float64(tt.size) / 1024 / duration.Seconds()
			memoryUsageMB := int64(m2.Alloc-m1.Alloc) / (1024 * 1024)

			// Duration check - now uses realistic environment-aware timeout
			assert.LessOrEqual(t, duration, maxDuration,
				"Processing took %v, exceeded environment-aware max %v (race_detection=%v)",
				duration, maxDuration, isRaceDetectionEnabled)

			// Throughput check
			assert.GreaterOrEqual(t, throughputKBps, tt.minThroughputKBps,
				"Throughput %.2f KB/s below minimum %.2f KB/s",
				throughputKBps, tt.minThroughputKBps)

			// Memory check
			assert.LessOrEqual(t, memoryUsageMB, tt.maxMemoryMB,
				"Memory usage %d MB exceeded max %d MB",
				memoryUsageMB, tt.maxMemoryMB)

			t.Logf("✅ Performance within envelope: %.2f KB/s, %v duration, %d MB memory (expected max: %v)",
				throughputKBps, duration, memoryUsageMB, maxDuration)
		})
	}
}
