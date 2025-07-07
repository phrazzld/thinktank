package perftest

import (
	"os"
	"strings"
	"testing"
)

// BenchmarkConfig provides CI-aware configuration for benchmarks
type BenchmarkConfig struct {
	*Config
	// SaveBaseline indicates if benchmark results should be saved as a baseline
	SaveBaseline bool
	// BaselineFile is the path to save/load baseline results
	BaselineFile string
	// CompareBaseline indicates if results should be compared to a baseline
	CompareBaseline bool
}

// NewBenchmarkConfig creates a benchmark configuration for the current environment
func NewBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		Config:       NewConfig(),
		SaveBaseline: os.Getenv("SAVE_BASELINE") == "true",
		BaselineFile: os.Getenv("BASELINE_FILE"),
	}
}

// RunBenchmark runs a benchmark with CI-aware adjustments
func RunBenchmark(b *testing.B, name string, fn func(b *testing.B)) {
	cfg := NewBenchmarkConfig()

	// Skip if necessary
	if skip, reason := cfg.ShouldSkip(extractTestType(name)); skip {
		b.Skipf("Skipping %s: %s", name, reason)
		return
	}

	// Log environment info
	b.Logf("Running benchmark in %s environment (CPU: %d, Race: %v)",
		cfg.Environment.RunnerType,
		cfg.Environment.CPUCount,
		cfg.Environment.IsRaceEnabled)

	// Run the benchmark
	fn(b)

	// Save results if requested
	if cfg.SaveBaseline && cfg.BaselineFile != "" {
		saveBaselineResult(b, name, cfg.BaselineFile)
	}
}

// SetBenchmarkIterations adjusts b.N based on the environment to ensure
// meaningful results without excessive runtime in CI
func SetBenchmarkIterations(b *testing.B, baseIterations int) int {
	cfg := NewConfig()

	iterations := baseIterations
	if cfg.Environment.IsCI {
		// Reduce iterations in CI to save time
		iterations = iterations / 2
		if iterations < 1 {
			iterations = 1
		}
	}

	if cfg.Environment.IsRaceEnabled {
		// Further reduce for race detector
		iterations = iterations / 2
		if iterations < 1 {
			iterations = 1
		}
	}

	return iterations
}

// ReportAllocs enables allocation reporting with environment context
func ReportAllocs(b *testing.B) {
	b.ReportAllocs()

	cfg := NewConfig()
	if cfg.Environment.IsCI {
		b.Logf("Note: Running in CI environment, allocation counts may vary")
	}
}

// SkipIfShortMode skips benchmarks in short mode with helpful context
func SkipIfShortMode(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode. Remove -short flag to run performance tests.")
	}
}

// BenchmarkResult represents a single benchmark result for comparison
type BenchmarkResult struct {
	Name        string
	NsPerOp     float64
	AllocsPerOp uint64
	BytesPerOp  uint64
	MBPerSecond float64
	Environment Environment
}

// CompareBenchmarks compares current results against a baseline
// This is a simplified version - in production, use benchstat
func CompareBenchmarks(b *testing.B, current BenchmarkResult, baseline BenchmarkResult) {
	b.Helper()

	// Calculate percentage changes
	timeChange := ((current.NsPerOp - baseline.NsPerOp) / baseline.NsPerOp) * 100
	allocChange := float64(current.AllocsPerOp-baseline.AllocsPerOp) / float64(baseline.AllocsPerOp) * 100
	bytesChange := float64(current.BytesPerOp-baseline.BytesPerOp) / float64(baseline.BytesPerOp) * 100

	b.Logf("Benchmark comparison for %s:", current.Name)
	b.Logf("  Time:   %.2f ns/op (%.1f%% change)", current.NsPerOp, timeChange)
	b.Logf("  Allocs: %d allocs/op (%.1f%% change)", current.AllocsPerOp, allocChange)
	b.Logf("  Bytes:  %d B/op (%.1f%% change)", current.BytesPerOp, bytesChange)

	// Define acceptable regression thresholds
	// These are more lenient in CI environments
	cfg := NewConfig()
	timeThreshold := 10.0 // 10% regression allowed
	allocThreshold := 5.0 // 5% regression allowed

	if cfg.Environment.IsCI {
		timeThreshold = 20.0 // More lenient in CI
		allocThreshold = 10.0
	}

	// Check for regressions
	if timeChange > timeThreshold {
		b.Errorf("Performance regression: %.1f%% slower (threshold: %.1f%%)", timeChange, timeThreshold)
	}

	if allocChange > allocThreshold {
		b.Errorf("Allocation regression: %.1f%% more allocations (threshold: %.1f%%)", allocChange, allocThreshold)
	}
}

// Helper functions

func extractTestType(benchmarkName string) string {
	name := strings.ToLower(benchmarkName)
	switch {
	case strings.Contains(name, "heavy"):
		return "heavy-cpu"
	case strings.Contains(name, "race"):
		return "race-sensitive"
	case strings.Contains(name, "local"):
		return "local-only"
	default:
		return ""
	}
}

func saveBaselineResult(b *testing.B, name, file string) {
	// This is a placeholder - in production, integrate with go test -bench output
	// and benchstat for proper baseline management
	b.Logf("Would save baseline for %s to %s", name, file)
}

// Example usage patterns that can be documented:

// Pattern 1: Basic CI-aware benchmark
func ExampleBasicBenchmark(b *testing.B) {
	RunBenchmark(b, "ExampleOperation", func(b *testing.B) {
		ReportAllocs(b)

		data := make([]byte, 1024*1024) // 1MB
		b.SetBytes(int64(len(data)))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Benchmark operation here
			_ = processData(data)
		}
	})
}

// Pattern 2: Benchmark with environment-specific iterations
func ExampleIterationAdjustedBenchmark(b *testing.B) {
	RunBenchmark(b, "ComplexOperation", func(b *testing.B) {
		iterations := SetBenchmarkIterations(b, 1000)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Run with adjusted iterations
			for j := 0; j < iterations; j++ {
				// Complex operation
			}
		}
	})
}

// Placeholder function for example
func processData(data []byte) []byte {
	return data
}
