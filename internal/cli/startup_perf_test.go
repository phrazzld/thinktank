package cli

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
)

// TestStartupPerformanceBaseline establishes the current performance baseline
// This test will initially fail and drive our optimization work
func TestStartupPerformanceBaseline(t *testing.T) {
	// Target: <10ms cold start for simplified CLI path
	target := 10 * time.Millisecond

	measurements := make([]time.Duration, 10)

	for i := 0; i < 10; i++ {
		// Force garbage collection to get consistent measurements
		runtime.GC()
		runtime.GC()

		start := time.Now()

		// Simulate minimal startup path for simplified CLI
		result := simulateStartupPath()

		elapsed := time.Since(start)
		measurements[i] = elapsed

		// Verify we got a valid result
		assert.NotNil(t, result, "Startup should return valid result")
	}

	// Calculate average startup time
	var total time.Duration
	for _, m := range measurements {
		total += m
	}
	average := total / time.Duration(len(measurements))

	t.Logf("Startup time measurements:")
	for i, m := range measurements {
		t.Logf("  Run %d: %v", i+1, m)
	}
	t.Logf("Average: %v (target: %v)", average, target)

	// This test captures our performance requirement
	assert.Less(t, average, target,
		"Average startup time (%v) should be less than target (%v)",
		average, target)
}

// simulateStartupPath simulates the critical path for CLI startup
func simulateStartupPath() *config.CliConfig {
	// This represents the minimum work needed for CLI to be usable
	// For startup performance, we focus on CPU/memory work, not I/O validation

	// Create a temporary instructions file for testing
	tempDir := "/tmp"
	instructionsFile := "/tmp/test_instructions.txt"

	// Create the temporary file (minimal I/O for test setup)
	if _, err := os.Stat(instructionsFile); os.IsNotExist(err) {
		file, err := os.Create(instructionsFile)
		if err == nil {
			if _, writeErr := file.WriteString("Test instructions for performance testing"); writeErr != nil {
				// Note: This function is called from tests, so we can't use b.Logf here
				fmt.Printf("Failed to write to instructions file: %v\n", writeErr)
			}
			if closeErr := file.Close(); closeErr != nil {
				fmt.Printf("Failed to close instructions file: %v\n", closeErr)
			}
		}
	}

	// 0. Initialize startup optimizations (environment cache prewarming)
	PrewarmEnvCache()

	// 1. Parse arguments (fast mode) - focus on parsing logic without I/O validation
	args := []string{"thinktank", instructionsFile, tempDir, "--dry-run"}
	simpleConfig, err := ParseSimpleArgsWithArgsFast(args, SkipValidationMode)
	if err != nil {
		// Debug: print the error to understand what's failing
		fmt.Printf("Parse error: %v\n", err)
		return nil
	}

	// 2. Convert to complex config (triggers various lookups) with fast mode
	complexConfig := simpleConfig.ToCliConfigFast()

	// 3. Basic validation (but not full validation which is slower)
	if complexConfig.InstructionsFile == "" {
		// Debug: check what we got
		fmt.Printf("Instructions file empty, got: %+v\n", complexConfig)
		return nil
	}

	return complexConfig
}

// BenchmarkStartupTime provides detailed startup performance benchmarks
func BenchmarkStartupTime(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := simulateStartupPath()
		if result == nil {
			b.Fatal("Startup simulation failed")
		}
	}
}

// BenchmarkStartupComponents benchmarks individual startup components
func BenchmarkStartupComponents(b *testing.B) {
	components := []struct {
		name string
		fn   func() error
	}{
		{
			name: "argument_parsing",
			fn: func() error {
				_, err := ParseSimpleArgsWithArgs([]string{"thinktank", "instructions.txt", "src/"})
				return err
			},
		},
		{
			name: "config_conversion",
			fn: func() error {
				config, _ := ParseSimpleArgsWithArgs([]string{"thinktank", "instructions.txt", "src/"})
				_ = config.ToCliConfig()
				return nil
			},
		},
		{
			name: "logger_creation",
			fn: func() error {
				_ = logutil.NewSlogLoggerFromLogLevel(nil, logutil.InfoLevel)
				return nil
			},
		},
	}

	for _, comp := range components {
		b.Run(comp.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := comp.fn(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestStartupMemoryUsage validates memory usage during startup
func TestStartupMemoryUsage(t *testing.T) {
	// Target: <5MB memory usage for startup
	targetMemory := int64(5 * 1024 * 1024) // 5MB in bytes

	runtime.GC()
	runtime.GC()

	var beforeStats runtime.MemStats
	runtime.ReadMemStats(&beforeStats)
	beforeAlloc := beforeStats.Alloc

	// Run startup simulation
	result := simulateStartupPath()
	assert.NotNil(t, result, "Startup should succeed")

	runtime.GC()
	runtime.GC()

	var afterStats runtime.MemStats
	runtime.ReadMemStats(&afterStats)
	afterAlloc := afterStats.Alloc

	memoryUsed := int64(afterAlloc - beforeAlloc)

	t.Logf("Memory usage during startup: %d bytes (%.2f MB)",
		memoryUsed, float64(memoryUsed)/(1024*1024))
	t.Logf("Target: %d bytes (%.2f MB)",
		targetMemory, float64(targetMemory)/(1024*1024))

	assert.Less(t, memoryUsed, targetMemory,
		"Startup memory usage (%d bytes) should be less than target (%d bytes)",
		memoryUsed, targetMemory)
}

// TestStartupTimePercentiles measures startup time distribution
func TestStartupTimePercentiles(t *testing.T) {
	// Take many measurements to understand distribution
	measurements := make([]time.Duration, 100)

	for i := 0; i < len(measurements); i++ {
		runtime.GC()

		start := time.Now()
		result := simulateStartupPath()
		elapsed := time.Since(start)

		if result == nil {
			t.Fatal("Startup simulation failed")
		}

		measurements[i] = elapsed
	}

	// Sort for percentile calculation
	for i := 0; i < len(measurements); i++ {
		for j := i + 1; j < len(measurements); j++ {
			if measurements[i] > measurements[j] {
				measurements[i], measurements[j] = measurements[j], measurements[i]
			}
		}
	}

	p50 := measurements[50]
	p95 := measurements[95]
	p99 := measurements[99]

	t.Logf("Startup time percentiles:")
	t.Logf("  P50: %v", p50)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)

	// P95 should be under target (most users have good experience)
	target := 10 * time.Millisecond
	assert.Less(t, p95, target,
		"P95 startup time (%v) should be under target (%v)", p95, target)
}

// BenchmarkColdStartVsWarmStart compares cold vs warm startup
func BenchmarkColdStartVsWarmStart(b *testing.B) {
	b.Run("cold_start", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate cold start by resetting any caches
			// (We'll implement cache reset as we add caching)
			result := simulateStartupPath()
			if result == nil {
				b.Fatal("Cold start failed")
			}
		}
	})

	b.Run("warm_start", func(b *testing.B) {
		// Warm up caches first
		_ = simulateStartupPath()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := simulateStartupPath()
			if result == nil {
				b.Fatal("Warm start failed")
			}
		}
	})
}
