package cli

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMemoryUsageProfile tests memory usage for different parsing scenarios
func TestMemoryUsageProfile(t *testing.T) {
	// Force garbage collection to get baseline
	runtime.GC()
	runtime.GC()

	// Get baseline memory stats
	var baselineStats runtime.MemStats
	runtime.ReadMemStats(&baselineStats)

	testCases := []struct {
		name          string
		args          []string
		maxAllocBytes uint64
		description   string
	}{
		{
			name:          "simple_parsing_basic",
			args:          []string{"thinktank", "instructions.txt", "src/"},
			maxAllocBytes: 1024, // 1KB target
			description:   "Basic simplified parsing should use <1KB",
		},
		{
			name:          "simple_parsing_with_flags",
			args:          []string{"thinktank", "instructions.txt", "src/", "--dry-run", "--verbose"},
			maxAllocBytes: 1024, // 1KB target
			description:   "Simplified parsing with flags should use <1KB",
		},
		{
			name:          "simple_parsing_all_flags",
			args:          []string{"thinktank", "instructions.txt", "src/", "--dry-run", "--verbose", "--model", "gpt-4.1", "--output-dir", "/tmp"},
			maxAllocBytes: 1536, // 1.5KB for complex case with model interning
			description:   "Complex simplified parsing should use <1.5KB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Force garbage collection before test
			runtime.GC()
			runtime.GC()

			// Get start memory stats
			var startStats runtime.MemStats
			runtime.ReadMemStats(&startStats)
			startAllocs := startStats.TotalAlloc

			// Parse arguments multiple times to get average
			const iterations = 100
			for i := 0; i < iterations; i++ {
				config, err := ParseSimpleArgsWithArgs(tc.args)
				if err == nil {
					_ = config // Use the config to prevent optimization
				}
			}

			// Force garbage collection after operations
			runtime.GC()
			runtime.GC()

			// Get end memory stats
			var endStats runtime.MemStats
			runtime.ReadMemStats(&endStats)
			endAllocs := endStats.TotalAlloc

			// Calculate allocated memory (accounting for baseline drift)
			totalAllocated := endAllocs - startAllocs
			avgPerParse := totalAllocated / iterations

			t.Logf("Memory usage for %s:", tc.name)
			t.Logf("  Total allocated: %d bytes", totalAllocated)
			t.Logf("  Average per parse: %d bytes", avgPerParse)
			t.Logf("  Target: %d bytes", tc.maxAllocBytes)

			// Verify memory usage is within target
			assert.LessOrEqual(t, avgPerParse, tc.maxAllocBytes,
				"Average memory per parse (%d bytes) should be <= %d bytes. %s",
				avgPerParse, tc.maxAllocBytes, tc.description)
		})
	}
}

// TestMemoryUsageWithPools tests that memory pools reduce allocations
func TestMemoryUsageWithPools(t *testing.T) {
	testCases := []struct {
		name        string
		operation   func()
		description string
		maxAllocs   int64
	}{
		{
			name: "string_builder_pool",
			operation: func() {
				sb := GetStringBuilder()
				sb.WriteString("thinktank instructions.txt target_path --model ")
				sb.WriteString("gpt-4.1")
				_ = sb.String()
				PutStringBuilder(sb)
			},
			description: "String builder pool should minimize allocations",
			maxAllocs:   2,
		},
		{
			name: "string_slice_pool",
			operation: func() {
				slice := GetStringSlice()
				slice = append(slice, "test1", "test2", "test3")
				PutStringSlice(slice)
			},
			description: "String slice pool should minimize allocations",
			maxAllocs:   1,
		},
		{
			name: "arguments_copy_pool",
			operation: func() {
				args := GetArgumentsCopy()
				args = append(args, "thinktank", "--model", "gpt-4.1", "instructions.txt", "target/")
				PutArgumentsCopy(args)
			},
			description: "Arguments copy pool should minimize allocations",
			maxAllocs:   1,
		},
		{
			name: "string_interning",
			operation: func() {
				_ = InternModelName("gpt-4.1")
				_ = InternModelName("gemini-2.5-pro")
				_ = InternModelName("gpt-4.1") // Should be cached
			},
			description: "String interning should have zero allocations for cached strings",
			maxAllocs:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Warm up (ensure pools are initialized and strings are interned)
			for i := 0; i < 10; i++ {
				tc.operation()
			}

			// Force garbage collection
			runtime.GC()
			runtime.GC()

			// Measure allocations
			var startStats runtime.MemStats
			runtime.ReadMemStats(&startStats)
			startMallocs := startStats.Mallocs

			// Run the operation
			const iterations = 100
			for i := 0; i < iterations; i++ {
				tc.operation()
			}

			// Get end stats
			var endStats runtime.MemStats
			runtime.ReadMemStats(&endStats)
			endMallocs := endStats.Mallocs

			// Calculate allocations per operation
			totalAllocs := int64(endMallocs - startMallocs)
			avgAllocsPerOp := totalAllocs / iterations

			t.Logf("Pool efficiency for %s:", tc.name)
			t.Logf("  Total mallocs: %d", totalAllocs)
			t.Logf("  Average mallocs per operation: %d", avgAllocsPerOp)
			t.Logf("  Target: <= %d mallocs", tc.maxAllocs)

			assert.LessOrEqual(t, avgAllocsPerOp, tc.maxAllocs,
				"Average mallocs per operation (%d) should be <= %d. %s",
				avgAllocsPerOp, tc.maxAllocs, tc.description)
		})
	}
}

// TestMemoryLeakDetection tests for memory leaks in repeated parsing
func TestMemoryLeakDetection(t *testing.T) {
	// Get baseline memory usage
	runtime.GC()
	runtime.GC()

	var baselineStats runtime.MemStats
	runtime.ReadMemStats(&baselineStats)
	baselineHeap := baselineStats.HeapAlloc

	// Run many parsing operations
	const iterations = 1000
	args := []string{"thinktank", "instructions.txt", "src/", "--dry-run", "--verbose"}

	for i := 0; i < iterations; i++ {
		config, err := ParseSimpleArgsWithArgs(args)
		if err == nil {
			_ = config // Use the config
		}

		// Also test migration guide generation (high allocation operation)
		if i%100 == 0 {
			telemetry := NewDeprecationTelemetry()
			telemetry.RecordUsagePattern("--model", []string{"--model", "gpt-4.1"})
			generator := NewMigrationGuideGenerator()
			guide, err := generator.GenerateFromTelemetry(telemetry)
			if err == nil {
				_ = guide // Use the guide
			}
		}
	}

	// Force garbage collection to clean up
	runtime.GC()
	runtime.GC()

	// Check final memory usage
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)
	finalHeap := finalStats.HeapAlloc

	// Calculate heap growth
	heapGrowth := int64(finalHeap) - int64(baselineHeap)
	growthPerIteration := heapGrowth / iterations

	t.Logf("Memory leak detection results:")
	t.Logf("  Baseline heap: %d bytes", baselineHeap)
	t.Logf("  Final heap: %d bytes", finalHeap)
	t.Logf("  Heap growth: %d bytes", heapGrowth)
	t.Logf("  Growth per iteration: %d bytes", growthPerIteration)

	// Allow some growth but detect significant leaks
	// Healthy growth should be < 10 bytes per iteration
	const maxGrowthPerIteration = 50 // bytes

	assert.LessOrEqual(t, growthPerIteration, int64(maxGrowthPerIteration),
		"Heap growth per iteration (%d bytes) suggests memory leak. Should be <= %d bytes",
		growthPerIteration, maxGrowthPerIteration)

	// Also check that we're not holding onto too much memory
	const maxTotalGrowth = 50 * 1024 // 50KB
	assert.LessOrEqual(t, heapGrowth, int64(maxTotalGrowth),
		"Total heap growth (%d bytes) is too large. Should be <= %d bytes",
		heapGrowth, maxTotalGrowth)
}

// BenchmarkMemoryOptimizedParsing benchmarks the optimized parsing with memory pools
func BenchmarkMemoryOptimizedParsing(b *testing.B) {
	args := []string{"thinktank", "instructions.txt", "src/", "--dry-run", "--verbose", "--model", "gpt-4.1"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		config, err := ParseSimpleArgsWithArgs(args)
		if err == nil {
			_ = config
		}
	}
}

// BenchmarkMemoryOptimizedMigrationGuide benchmarks migration guide with pools
func BenchmarkMemoryOptimizedMigrationGuide(b *testing.B) {
	telemetry := NewDeprecationTelemetry()
	generator := NewMigrationGuideGenerator()

	// Pre-populate with some patterns
	patterns := [][]string{
		{"--model", "gpt-4.1"},
		{"--instructions", "file.txt"},
		{"--output-dir", "/tmp"},
		{"--verbose"},
		{"--dry-run"},
	}

	for _, pattern := range patterns {
		telemetry.RecordUsagePattern(pattern[0], pattern)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		guide, err := generator.GenerateFromTelemetry(telemetry)
		if err == nil {
			_ = guide
		}
	}
}
