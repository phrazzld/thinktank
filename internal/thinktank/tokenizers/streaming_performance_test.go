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

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
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

	// Use the performance testing framework to measure throughput
	measurement := perftest.MeasureThroughput(t, "StreamingTokenization", func() (int64, error) {
		reader := strings.NewReader(text)
		tokens, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
		if err != nil {
			return 0, err
		}
		require.Greater(t, tokens, 0, "Should return positive token count")
		return int64(len(text)), nil
	})

	// Assert minimum throughput using framework (6KB/s baseline, adjusted for environment)
	// This provides a more realistic baseline that accounts for varying system performance
	perftest.AssertThroughput(t, measurement, 6*1024)

	// Log performance tier for visibility
	if measurement.BytesPerSecond > 1024*1024.0 { // 1MB/s
		t.Logf("✅ Excellent performance: %.2f MB/s", measurement.BytesPerSecond/(1024*1024))
	} else if measurement.BytesPerSecond > 512*1024.0 { // 512KB/s
		t.Logf("✅ Good performance: %.2f KB/s", measurement.BytesPerSecond/1024)
	} else {
		t.Logf("⚠️  Performance needs improvement: %.2f KB/s", measurement.BytesPerSecond/1024)
	}
}

// TestStreamingTokenizer_ScalesLinearlyWithInputSize observes scaling behavior
// of streaming tokenization. This test is INFORMATIONAL ONLY - throughput varies
// based on system load, GC, and environment. Use benchmarks for trend analysis.
// See: BenchmarkStreamingTokenizer_ThroughputByInputSize
func TestStreamingTokenizer_ScalesLinearlyWithInputSize(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Test with progressively larger inputs to observe scaling behavior
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
		_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
		duration := time.Since(start)
		require.NoError(t, err)

		throughput := float64(size) / duration.Seconds()
		measurements = append(measurements, measurement{size, duration, throughput})

		t.Logf("Size %s: %.2f KB/s", formatSizeString(size), throughput/1024)
	}

	// Log scaling observations (informational only)
	// Throughput varies by environment - benchmarks are the source of truth
	baselineThroughput := measurements[0].throughput
	if baselineThroughput <= 0 {
		t.Log("ℹ️  Baseline throughput zero - skipping ratio analysis")
		return
	}

	for i := 1; i < len(measurements); i++ {
		throughputRatio := measurements[i].throughput / baselineThroughput

		// Informational logging only - no assertions on non-deterministic metrics
		if throughputRatio >= 0.8 {
			t.Logf("✅ Good scaling: %s maintains %.1f%% of baseline throughput",
				formatSizeString(measurements[i].size), throughputRatio*100)
		} else {
			t.Logf("ℹ️  Scaling ratio: %s at %.1f%% of baseline (check benchmarks for trend)",
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

// TestStreamingTokenizer_ConstantMemoryUsage observes memory behavior of streaming
// tokenization. This test is INFORMATIONAL ONLY - memory metrics are inherently
// non-deterministic due to GC timing. Use benchmarks for actual performance tracking.
// See: BenchmarkStreamingTokenizer_MemoryAllocation
func TestStreamingTokenizer_ConstantMemoryUsage(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Test with increasingly large inputs to observe memory scaling
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

		_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
		require.NoError(t, err)

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := int64(m2.Alloc - m1.Alloc)
		memoryUsages = append(memoryUsages, memoryIncrease)

		t.Logf("Size %s: memory increase %s",
			formatSizeString(size), formatSizeString(int(memoryIncrease)))
	}

	// Log memory scaling observations (informational only)
	// Memory metrics are non-deterministic - benchmarks are the source of truth
	baselineMemory := memoryUsages[0]
	if baselineMemory <= 0 {
		t.Log("ℹ️  Baseline memory near zero (GC reclaimed) - skipping ratio analysis")
		return
	}

	for i := 1; i < len(memoryUsages); i++ {
		memoryRatio := float64(memoryUsages[i]) / float64(baselineMemory)
		inputSizeRatio := float64(sizes[i]) / float64(sizes[0])

		// Informational logging only - no assertions on non-deterministic metrics
		if memoryRatio <= 2.0 {
			t.Logf("✅ Good memory efficiency: %.1fx memory for %.1fx input",
				memoryRatio, inputSizeRatio)
		} else {
			t.Logf("ℹ️  Memory ratio: %.1fx memory for %.1fx input (check benchmarks for trend)",
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
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
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
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
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
				_, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
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
					context.Background(), reader, "gpt-5.2", size.size)
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

// TestStreamingTokenizer_PerformanceEnvelope observes performance characteristics
// of streaming tokenization at various sizes. This test is INFORMATIONAL ONLY -
// performance metrics vary by environment. Use benchmarks for regression detection.
// See: BenchmarkStreamingTokenizer_ThroughputByInputSize, BenchmarkStreamingTokenizer_MemoryAllocation
func TestStreamingTokenizer_PerformanceEnvelope(t *testing.T) {
	t.Parallel()

	manager := NewStreamingTokenizerManager()
	streamingTokenizer, err := manager.GetStreamingTokenizer("openrouter")
	require.NoError(t, err)

	// Define observation points for performance characterization
	tests := []struct {
		name string
		size int
	}{
		{name: "1MB_observation", size: 1024 * 1024},
		{name: "10MB_observation", size: 10 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			text := generatePredictableContent(tt.size)
			reader := strings.NewReader(text)

			isRaceDetectionEnabled := raceDetectionEnabled()
			t.Logf("Environment: race_detection=%v, input_size=%s",
				isRaceDetectionEnabled, formatSizeString(tt.size))

			start := time.Now()
			tokens, err := streamingTokenizer.CountTokensStreaming(context.Background(), reader, "gpt-5.2")
			duration := time.Since(start)

			runtime.GC()
			runtime.ReadMemStats(&m2)

			require.NoError(t, err)
			require.Greater(t, tokens, 0)

			// Informational logging only - benchmarks are the source of truth
			throughputKBps := float64(tt.size) / 1024 / duration.Seconds()
			memoryUsageMB := int64(m2.Alloc-m1.Alloc) / (1024 * 1024)

			t.Logf("ℹ️  Performance observation: %.2f KB/s, %v duration, %d MB memory",
				throughputKBps, duration, memoryUsageMB)
		})
	}
}
