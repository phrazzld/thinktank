package cli

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	// Clear cache before test
	ClearEnvCache()

	// Set a test environment variable
	testKey := "TEST_ENV_VAR"
	testValue := "test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		t.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	// First call should cache the value
	result1 := GetEnv(testKey)
	assert.Equal(t, testValue, result1)

	// Second call should use cached value (verify cache is working)
	result2 := GetEnv(testKey)
	assert.Equal(t, testValue, result2)

	// Test non-existent key
	result3 := GetEnv("NON_EXISTENT_KEY")
	assert.Equal(t, "", result3)
}

func TestGetEnvWithDefault(t *testing.T) {
	ClearEnvCache()

	// Test with existing environment variable
	testKey := "TEST_ENV_WITH_DEFAULT"
	testValue := "existing_value"
	defaultValue := "default_value"

	if err := os.Setenv(testKey, testValue); err != nil {
		t.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	result1 := GetEnvWithDefault(testKey, defaultValue)
	assert.Equal(t, testValue, result1)

	// Test with non-existent key (should return default)
	result2 := GetEnvWithDefault("NON_EXISTENT_KEY", defaultValue)
	assert.Equal(t, defaultValue, result2)

	// Test with empty value (should return default)
	if err := os.Setenv("EMPTY_KEY", ""); err != nil {
		t.Fatalf("Failed to set empty environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("EMPTY_KEY"); err != nil {
			t.Logf("Failed to unset empty environment variable: %v", err)
		}
	}()
	result3 := GetEnvWithDefault("EMPTY_KEY", defaultValue)
	assert.Equal(t, defaultValue, result3)
}

func TestPrewarmEnvCache(t *testing.T) {
	ClearEnvCache()

	// Set some test environment variables
	if err := os.Setenv("OPENAI_API_KEY", "test_openai_key"); err != nil {
		t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
	}
	if err := os.Setenv("GEMINI_API_KEY", "test_gemini_key"); err != nil {
		t.Fatalf("Failed to set GEMINI_API_KEY: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
			t.Logf("Failed to unset OPENAI_API_KEY: %v", err)
		}
		if err := os.Unsetenv("GEMINI_API_KEY"); err != nil {
			t.Logf("Failed to unset GEMINI_API_KEY: %v", err)
		}
	}()

	// Prewarm the cache
	PrewarmEnvCache()

	// Verify that common environment variables are cached
	// We can't directly inspect the cache, but we can verify the behavior
	result1 := GetEnv("OPENAI_API_KEY")
	assert.Equal(t, "test_openai_key", result1)

	result2 := GetEnv("GEMINI_API_KEY")
	assert.Equal(t, "test_gemini_key", result2)

	// Calling PrewarmEnvCache again should be safe (sync.Once behavior)
	PrewarmEnvCache()
}

func TestClearEnvCache(t *testing.T) {
	// Set up some cached values
	testKey := "TEST_CLEAR_CACHE"
	testValue := "test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		t.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	// Cache the value
	result1 := GetEnv(testKey)
	assert.Equal(t, testValue, result1)

	// Clear the cache
	ClearEnvCache()

	// Change the environment variable
	newValue := "new_value"
	if err := os.Setenv(testKey, newValue); err != nil {
		t.Fatalf("Failed to set test environment variable to new value: %v", err)
	}

	// Should get the new value (cache was cleared)
	result2 := GetEnv(testKey)
	assert.Equal(t, newValue, result2)
}

func TestEnvCachePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ClearEnvCache()

	// Set up test environment variable
	testKey := "PERF_TEST_KEY"
	testValue := "performance_test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		t.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	// Warm up the cache
	GetEnv(testKey)

	// Benchmark cached lookups vs direct os.Getenv calls
	const iterations = 10000

	// Test cached lookups
	start := time.Now()
	for i := 0; i < iterations; i++ {
		GetEnv(testKey)
	}
	cachedDuration := time.Since(start)

	// Clear cache and test direct os.Getenv calls
	ClearEnvCache()
	start = time.Now()
	for i := 0; i < iterations; i++ {
		os.Getenv(testKey)
	}
	directDuration := time.Since(start)

	t.Logf("Cached lookups (%d iterations): %v", iterations, cachedDuration)
	t.Logf("Direct lookups (%d iterations): %v", iterations, directDuration)
	t.Logf("Speedup: %.2fx", float64(directDuration)/float64(cachedDuration))

	// Cached lookups should be significantly faster
	// Allow some variance but expect at least 2x improvement
	assert.Less(t, cachedDuration*2, directDuration,
		"Cached lookups should be at least 2x faster than direct lookups")
}

func TestEnvCacheConcurrency(t *testing.T) {
	ClearEnvCache()

	// Set up test environment variable
	testKey := "CONCURRENCY_TEST_KEY"
	testValue := "concurrency_test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		t.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	// Test concurrent access
	const numGoroutines = 100
	const iterations = 100
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				result := GetEnv(testKey)
				if result != testValue {
					done <- assert.AnError
					return
				}
			}
			done <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err, "Concurrent access should not cause errors")
	}
}

// BenchmarkGetEnvCached benchmarks cached environment variable lookups
func BenchmarkGetEnvCached(b *testing.B) {
	ClearEnvCache()

	testKey := "BENCH_TEST_KEY"
	testValue := "benchmark_test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		b.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			b.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	// Warm up the cache
	GetEnv(testKey)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		GetEnv(testKey)
	}
}

// BenchmarkGetEnvDirect benchmarks direct os.Getenv calls for comparison
func BenchmarkGetEnvDirect(b *testing.B) {
	testKey := "BENCH_TEST_KEY"
	testValue := "benchmark_test_value"
	if err := os.Setenv(testKey, testValue); err != nil {
		b.Fatalf("Failed to set test environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			b.Logf("Failed to unset test environment variable: %v", err)
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		os.Getenv(testKey)
	}
}
