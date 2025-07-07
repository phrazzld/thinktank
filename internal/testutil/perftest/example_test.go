package perftest_test

import (
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
)

// Example 1: Basic throughput test
func TestExample_Throughput(t *testing.T) {
	// Measure throughput of a data processing operation
	measurement := perftest.MeasureThroughput(t, "SHA256 Hashing", func() (int64, error) {
		data := []byte(strings.Repeat("test data ", 10000))
		hash := sha256.Sum256(data)
		_ = hash
		return int64(len(data)), nil
	})

	// Assert minimum 50 MB/s locally (automatically adjusted for CI)
	perftest.AssertThroughput(t, measurement, 50*1024*1024)
}

// Example 2: Memory usage test
func TestExample_MemoryUsage(t *testing.T) {
	// Test that repeated operations don't leak memory
	perftest.AssertConstantMemory(t, "Map Operations", 1000, func() {
		m := make(map[string]string)
		for i := 0; i < 100; i++ {
			m[string(rune(i))] = "value"
		}
		// Map goes out of scope and should be GC'd
	})
}

// Example 3: Timeout-aware test
func TestExample_WithTimeout(t *testing.T) {
	perftest.WithTimeout(t, 5*time.Second, func() {
		// Simulate work that might take longer in CI
		time.Sleep(100 * time.Millisecond)
		// In CI, this gets a 10-second timeout instead of 5
	})
}

// Example 4: Environment-aware test skipping
func TestExample_HeavyComputation(t *testing.T) {
	cfg := perftest.NewConfig()

	// Skip if environment doesn't meet requirements
	if skip, reason := cfg.ShouldSkip("heavy-cpu"); skip {
		t.Skip(reason)
	}

	// Heavy computation here...
}

// Example 5: CI-aware benchmark
func BenchmarkExample_DataProcessing(b *testing.B) {
	perftest.RunBenchmark(b, "DataProcessing", func(b *testing.B) {
		perftest.ReportAllocs(b)
		perftest.SkipIfShortMode(b)

		// Setup
		data := make([]byte, 1024)
		b.SetBytes(int64(len(data)))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Process data
			result := sha256.Sum256(data)
			_ = result
		}
	})
}

// Example 6: Benchmark with adjusted iterations
func BenchmarkExample_ComplexOperation(b *testing.B) {
	perftest.RunBenchmark(b, "ComplexOperation", func(b *testing.B) {
		// Adjust iterations for CI (1000 locally, 500 in CI, 250 with race detector)
		iterations := perftest.SetBenchmarkIterations(b, 1000)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sum := 0
			for j := 0; j < iterations; j++ {
				// Simulate complex work
				sum += j * j
			}
			_ = sum
		}
	})
}

// Example 7: Detailed memory measurement
func TestExample_DetailedMemory(t *testing.T) {
	before, after, delta := perftest.MeasureMemory(t, "Large Allocation", func() {
		// Allocate 10MB
		data := make([]byte, 10*1024*1024)
		// Do something with it
		for i := range data {
			data[i] = byte(i % 256)
		}
	})

	t.Logf("Memory before: %+v", before)
	t.Logf("Memory after: %+v", after)
	t.Logf("Memory delta: %+v", delta)
}

// Example 8: Performance test with multiple measurements
func TestExample_MultipleMetrics(t *testing.T) {
	cfg := perftest.NewConfig()
	t.Logf("Testing in %s environment", cfg.Environment.RunnerType)

	// Measure both throughput and memory
	measurement := perftest.MeasureThroughput(t, "Processing", func() (int64, error) {
		data := make([]byte, 1024*1024) // 1MB

		// Also check memory within the operation
		_, _, delta := perftest.MeasureMemory(t, "Inner allocation", func() {
			processed := processData(data)
			_ = processed
		})

		// Ensure we're not allocating too much
		maxAllowed := cfg.AdjustMemory(2 * 1024 * 1024) // 2MB baseline
		if delta.TotalAllocBytes > uint64(maxAllowed) {
			t.Errorf("Allocated too much memory: %d bytes (max: %d)",
				delta.TotalAllocBytes, maxAllowed)
		}

		return int64(len(data)), nil
	})

	// Check throughput
	perftest.AssertThroughput(t, measurement, 100*1024*1024) // 100 MB/s baseline
}

// Helper function for examples
func processData(data []byte) []byte {
	result := make([]byte, len(data))
	copy(result, data)
	return result
}
