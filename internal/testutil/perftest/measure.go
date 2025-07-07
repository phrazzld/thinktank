package perftest

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// ThroughputMeasurement represents the result of a throughput test
type ThroughputMeasurement struct {
	BytesProcessed int64
	Duration       time.Duration
	BytesPerSecond float64
}

// MemoryStats captures memory usage statistics
type MemoryStats struct {
	AllocBytes      uint64
	TotalAllocBytes uint64
	NumGC           uint32
	NumAllocs       uint64
}

// MeasureThroughput measures the throughput of a function processing data
func MeasureThroughput(t *testing.T, name string, fn func() (bytesProcessed int64, err error)) ThroughputMeasurement {
	t.Helper()

	start := time.Now()
	bytesProcessed, err := fn()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("%s failed: %v", name, err)
	}

	bytesPerSecond := float64(bytesProcessed) / duration.Seconds()

	measurement := ThroughputMeasurement{
		BytesProcessed: bytesProcessed,
		Duration:       duration,
		BytesPerSecond: bytesPerSecond,
	}

	t.Logf("%s: processed %d bytes in %v (%.2f KB/s)",
		name, bytesProcessed, duration, bytesPerSecond/1024)

	return measurement
}

// AssertThroughput checks if throughput meets the environment-adjusted minimum
func AssertThroughput(t *testing.T, measurement ThroughputMeasurement, minBytesPerSecond float64) {
	t.Helper()

	cfg := NewConfig()
	adjustedMin := cfg.AdjustThroughput(minBytesPerSecond)

	if measurement.BytesPerSecond < adjustedMin {
		t.Errorf("Throughput %.2f KB/s is below minimum %.2f KB/s (adjusted for %s environment)",
			measurement.BytesPerSecond/1024,
			adjustedMin/1024,
			cfg.Environment.RunnerType)
	}
}

// MeasureMemory captures memory statistics before and after a function execution
func MeasureMemory(t *testing.T, name string, fn func()) (before, after MemoryStats, delta MemoryStats) {
	t.Helper()

	// Force GC and capture initial state
	runtime.GC()
	runtime.GC() // Run twice to ensure finalizers are processed
	before = captureMemoryStats()

	// Run the function
	fn()

	// Capture final state
	after = captureMemoryStats()

	// Calculate deltas
	delta = MemoryStats{
		AllocBytes:      after.AllocBytes - before.AllocBytes,
		TotalAllocBytes: after.TotalAllocBytes - before.TotalAllocBytes,
		NumGC:           after.NumGC - before.NumGC,
		NumAllocs:       after.NumAllocs - before.NumAllocs,
	}

	t.Logf("%s memory usage: %s allocated, %d allocations, %d GCs",
		name,
		formatBytes(delta.TotalAllocBytes),
		delta.NumAllocs,
		delta.NumGC)

	return before, after, delta
}

// AssertConstantMemory verifies that memory usage doesn't grow with repeated operations
func AssertConstantMemory(t *testing.T, name string, iterations int, fn func()) {
	t.Helper()

	cfg := NewConfig()

	// Warm up
	for i := 0; i < 10; i++ {
		fn()
	}

	// Measure first batch
	_, _, delta1 := MeasureMemory(t, fmt.Sprintf("%s first batch", name), func() {
		for i := 0; i < iterations; i++ {
			fn()
		}
	})

	// Measure second batch
	_, _, delta2 := MeasureMemory(t, fmt.Sprintf("%s second batch", name), func() {
		for i := 0; i < iterations; i++ {
			fn()
		}
	})

	// Allow some variance but ensure no significant growth
	allowedGrowth := cfg.AdjustMemory(int64(delta1.TotalAllocBytes) / 10) // 10% tolerance

	if int64(delta2.TotalAllocBytes) > int64(delta1.TotalAllocBytes)+allowedGrowth {
		t.Errorf("Memory usage grew from %s to %s (more than allowed %s growth)",
			formatBytes(delta1.TotalAllocBytes),
			formatBytes(delta2.TotalAllocBytes),
			formatBytes(uint64(allowedGrowth)))
	}
}

// WithTimeout runs a test function with an environment-adjusted timeout
func WithTimeout(t *testing.T, baseTimeout time.Duration, fn func()) {
	t.Helper()

	cfg := NewConfig()
	timeout := cfg.AdjustTimeout(baseTimeout)

	done := make(chan struct{})
	var panicValue interface{}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicValue = r
			}
			close(done)
		}()
		fn()
	}()

	select {
	case <-done:
		if panicValue != nil {
			panic(panicValue)
		}
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v (base: %v, environment: %s)",
			timeout, baseTimeout, cfg.Environment.RunnerType)
	}
}

// Helper functions

func captureMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemoryStats{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		NumGC:           m.NumGC,
		NumAllocs:       m.Mallocs,
	}
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
