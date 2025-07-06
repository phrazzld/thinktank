# CI Testing Guidelines

This document establishes patterns and best practices for writing tests that work reliably across different environments (local development and various CI runners).

## Table of Contents
- [Environment Detection](#environment-detection)
- [Performance Testing](#performance-testing)
- [Handling Flaky Tests](#handling-flaky-tests)
- [Environment-Dependent Tests](#environment-dependent-tests)
- [API Key Management](#api-key-management)
- [Test Categories](#test-categories)
- [CI-Specific Patterns](#ci-specific-patterns)

## Environment Detection

### Using the Performance Testing Framework

For new performance tests, use the CI-aware framework in `internal/testutil/perftest`:

```go
import "github.com/phrazzld/thinktank/internal/testutil/perftest"

func TestPerformance(t *testing.T) {
    // Automatically adjusts thresholds based on environment
    measurement := perftest.MeasureThroughput(t, "Operation", func() (int64, error) {
        // Your operation here
        return bytesProcessed, nil
    })

    // 100 MB/s baseline, automatically adjusted for CI
    perftest.AssertThroughput(t, measurement, 100*1024*1024)
}
```

### Manual Environment Detection

For tests that need custom environment handling:

```go
isCI := os.Getenv("CI") == "true"
if isCI {
    // Adjust expectations for CI
    timeout = 2 * time.Minute
} else {
    // Local development settings
    timeout = 30 * time.Second
}
```

## Performance Testing

### Key Principles

1. **Use relative comparisons over absolute thresholds**
   - CI runners have variable performance
   - Compare against baseline from same environment
   - Use statistical analysis (benchstat) for regression detection

2. **Adjust expectations by environment**
   - CI: 70% of local throughput baseline
   - With race detector: Additional 50% reduction
   - Memory: Allow 20% more in CI

3. **Example migration to framework**

Before:
```go
func TestThroughput(t *testing.T) {
    start := time.Now()
    processData()
    duration := time.Since(start)

    // Hard-coded threshold fails in CI
    assert.Less(t, duration, 100*time.Millisecond)
}
```

After:
```go
func TestThroughput(t *testing.T) {
    measurement := perftest.MeasureThroughput(t, "ProcessData", func() (int64, error) {
        data := generateData(1024*1024) // 1MB
        processData(data)
        return int64(len(data)), nil
    })

    // Baseline adjusted for environment
    perftest.AssertThroughput(t, measurement, 10*1024*1024) // 10 MB/s baseline
}
```

## Handling Flaky Tests

### Common Causes and Solutions

1. **Timing-dependent tests**
   ```go
   // Bad: Fixed sleep times
   time.Sleep(100 * time.Millisecond)
   assert.True(t, isComplete)

   // Good: Use retry with timeout
   require.Eventually(t, func() bool {
       return isComplete
   }, 5*time.Second, 100*time.Millisecond)
   ```

2. **Resource contention**
   ```go
   // Use t.Parallel() carefully
   func TestResourceIntensive(t *testing.T) {
       if testing.Short() {
           t.Skip("Skipping resource-intensive test in short mode")
       }
       // Don't use t.Parallel() for resource-intensive tests
   }
   ```

3. **Network-dependent tests**
   ```go
   // Add retries for transient failures
   var lastErr error
   for i := 0; i < 3; i++ {
       if err := networkOperation(); err != nil {
           lastErr = err
           time.Sleep(time.Second * time.Duration(i+1))
           continue
       }
       return // Success
   }
   t.Fatalf("Failed after 3 attempts: %v", lastErr)
   ```

## Environment-Dependent Tests

### Test Categories

Mark tests with categories to control execution:

```go
func TestHeavyComputation(t *testing.T) {
    cfg := perftest.NewConfig()
    if skip, reason := cfg.ShouldSkip("heavy-cpu"); skip {
        t.Skip(reason)
    }
    // Test implementation
}
```

Supported categories:
- `heavy-cpu`: Requires ≥4 CPUs
- `race-sensitive`: Incompatible with race detector
- `local-only`: Skip in CI environments

### Build Tags for Test Separation

```go
//go:build integration

package mypackage_test

func TestIntegration(t *testing.T) {
    // This test only runs with: go test -tags=integration
}
```

## API Key Management

### Post-OpenRouter Consolidation Pattern

All tests now use a single API key pattern:

```go
func TestOpenRouterIntegration(t *testing.T) {
    apiKey := os.Getenv("OPENROUTER_API_KEY")
    if apiKey == "" {
        t.Skip("OPENROUTER_API_KEY not set")
    }
    // Use API key for testing
}
```

### CI Configuration

Ensure CI workflow exposes required secrets:

```yaml
# .github/workflows/go.yml
- name: Run Tests
  env:
    OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
  run: go test ./...
```

## Test Categories

### Unit Tests
- Run on every commit
- No external dependencies
- Use mocks/stubs for isolation
- Target: <100ms per test

### Integration Tests
- Separate CI job with retries
- May use external services
- Longer timeouts allowed
- Skip if dependencies unavailable

### Performance Tests
- Run in dedicated CI job
- Compare against baselines
- Use `perftest` framework
- Non-blocking for PRs (informational)

## CI-Specific Patterns

### Timeout Handling

```go
func TestWithTimeout(t *testing.T) {
    // Use framework for automatic adjustment
    perftest.WithTimeout(t, 30*time.Second, func() {
        // Operation that might be slower in CI
    })
}
```

### Logging for Debugging

Always include environment information in test failures:

```go
cfg := perftest.NewConfig()
t.Logf("Environment: %+v", cfg.Environment)
t.Errorf("Test failed in %s environment", cfg.Environment.RunnerType)
```

### Parallel Test Execution

```go
func TestParallel(t *testing.T) {
    // Check if parallel execution is safe
    if os.Getenv("CI") == "true" && runtime.NumCPU() < 4 {
        // Run serially in resource-constrained CI
    } else {
        t.Parallel()
    }
}
```

### CI Detection Patterns

The framework detects these CI environments:
- GitHub Actions (`GITHUB_ACTIONS=true`)
- GitLab CI (`GITLAB_CI` present)
- CircleCI (`CIRCLECI=true`)
- Generic CI (`CI=true`)

### Race Detector Awareness

Tests automatically detect race detector through:
- `RACE_ENABLED=true` environment variable
- `-race` flag in command line
- Performance heuristics (operations take >5x longer)

## Best Practices Summary

1. **Always use the `perftest` framework for performance tests**
2. **Set appropriate timeouts with environment adjustment**
3. **Use test categories to control execution**
4. **Add retries for network operations**
5. **Log environment details on failures**
6. **Compare performance against baselines, not absolute values**
7. **Handle missing API keys gracefully with skips**
8. **Separate heavy tests with build tags**
9. **Document why a test might be flaky**
10. **Monitor test reliability metrics in CI dashboard**

## Migration Checklist

When updating existing tests for CI compatibility:

- [ ] Replace hard-coded timeouts with `perftest.WithTimeout`
- [ ] Convert absolute performance thresholds to `perftest.AssertThroughput`
- [ ] Add environment detection for CI-specific behavior
- [ ] Implement retries for flaky operations
- [ ] Add test categories for conditional execution
- [ ] Ensure API keys are handled with proper skips
- [ ] Add detailed logging for CI debugging
- [ ] Consider parallel execution impact

## References

- Performance Testing Framework: `/internal/testutil/perftest/`
- Example Usage: `/internal/testutil/perftest/example_test.go`
- Migration Example: `/internal/thinktank/tokenizers/streaming_performance_test.go`
- CI Workflow: `/.github/workflows/go.yml`
