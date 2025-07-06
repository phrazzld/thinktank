// Package perftest provides a CI-aware performance testing framework for Go.
//
// The framework automatically adjusts performance expectations based on the
// testing environment (local development vs various CI runners) and provides
// utilities for consistent performance measurement across different environments.
//
// # Key Features
//
//   - Automatic CI environment detection
//   - Dynamic threshold adjustment based on environment
//   - Race detector awareness
//   - Consistent timeout calculation
//   - Memory usage tracking
//   - Integration with Go's standard testing package
//
// # Basic Usage
//
// The framework provides several levels of integration:
//
// ## 1. Simple Throughput Testing
//
//	func TestThroughput(t *testing.T) {
//	    measurement := perftest.MeasureThroughput(t, "DataProcessing", func() (int64, error) {
//	        data := generateTestData(1024 * 1024) // 1MB
//	        processed, err := processData(data)
//	        return int64(len(processed)), err
//	    })
//
//	    // Assert minimum 10 MB/s locally, automatically adjusted for CI
//	    perftest.AssertThroughput(t, measurement, 10*1024*1024)
//	}
//
// ## 2. Memory Usage Testing
//
//	func TestMemoryUsage(t *testing.T) {
//	    // Verify memory usage doesn't grow with repeated operations
//	    perftest.AssertConstantMemory(t, "CacheOperation", 1000, func() {
//	        cache.Store("key", "value")
//	        cache.Delete("key")
//	    })
//	}
//
// ## 3. CI-Aware Benchmarks
//
//	func BenchmarkOperation(b *testing.B) {
//	    perftest.RunBenchmark(b, "Operation", func(b *testing.B) {
//	        perftest.ReportAllocs(b)
//
//	        data := make([]byte, 1024)
//	        b.SetBytes(int64(len(data)))
//
//	        b.ResetTimer()
//	        for i := 0; i < b.N; i++ {
//	            process(data)
//	        }
//	    })
//	}
//
// ## 4. Manual Configuration
//
//	func TestWithCustomThresholds(t *testing.T) {
//	    cfg := perftest.NewConfig()
//
//	    // Check if we should skip this test
//	    if skip, reason := cfg.ShouldSkip("heavy-cpu"); skip {
//	        t.Skip(reason)
//	    }
//
//	    // Adjust timeout for environment
//	    timeout := cfg.AdjustTimeout(30 * time.Second)
//
//	    perftest.WithTimeout(t, timeout, func() {
//	        // Long-running operation
//	    })
//	}
//
// # Environment Detection
//
// The framework detects the following environments:
//   - GitHub Actions
//   - GitLab CI
//   - CircleCI
//   - Generic CI (via CI environment variable)
//   - Local development
//
// It also detects:
//   - Race detector enabled
//   - CPU count
//   - OS/Architecture
//
// # Performance Adjustments
//
// Default adjustments for CI environments:
//   - Throughput: 70% of baseline (0.7x multiplier)
//   - Timeouts: 200% of baseline (2x multiplier)
//   - Memory: 120% of baseline (1.2x multiplier)
//
// With race detector enabled:
//   - Additional 50% throughput reduction
//   - Additional 2x timeout increase
//   - Additional 3x memory allowance
//
// # Integration with CI/CD
//
// For regression detection in CI:
//
// 1. Run benchmarks on main branch and save results:
//
//	SAVE_BASELINE=true BASELINE_FILE=baseline.txt go test -bench=. ./...
//
// 2. Run benchmarks on PR branch and compare:
//
//	go test -bench=. ./... > pr.txt
//	benchstat baseline.txt pr.txt
//
// # Best Practices
//
// 1. Use relative comparisons over absolute thresholds when possible
// 2. Always log environment information in test output
// 3. Consider using test categories (heavy-cpu, race-sensitive, etc.)
// 4. Run performance tests in dedicated CI jobs
// 5. Use benchstat for statistical comparison of results
package perftest
