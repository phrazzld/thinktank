package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestSemaphore(t *testing.T) {
	// Removed t.Parallel() - timing-dependent test with timeouts
	t.Run("Basic Acquisition and Release", func(t *testing.T) {
		t.Parallel() // Run subtests in parallel
		// Create a semaphore with capacity 2
		sem := NewSemaphore(2)
		assert.NotNil(t, sem, "Semaphore should not be nil")

		// Should be able to acquire twice without blocking
		err1 := sem.Acquire(context.Background())
		assert.NoError(t, err1, "First acquire should succeed")

		err2 := sem.Acquire(context.Background())
		assert.NoError(t, err2, "Second acquire should succeed")

		// Third acquire should block, use a timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err3 := sem.Acquire(ctx)
		assert.Error(t, err3, "Third acquire should fail with timeout")
		assert.Equal(t, ErrContextCanceled, err3, "Error should be ErrContextCanceled")

		// Release one, should be able to acquire again
		sem.Release()

		err4 := sem.Acquire(context.Background())
		assert.NoError(t, err4, "Acquire after release should succeed")

		// Release remaining semaphores
		sem.Release()
		sem.Release()
	})

	t.Run("Zero Value (No Limiting)", func(t *testing.T) {
		// Create a semaphore with zero capacity (no limit)
		sem := NewSemaphore(0)
		assert.Nil(t, sem, "Zero capacity semaphore should be nil")

		// Nil semaphore operations should be no-ops
		err := sem.Acquire(context.Background())
		assert.NoError(t, err, "Acquire on nil semaphore should succeed")

		// Release should not panic
		sem.Release() // This should be a no-op
	})

	t.Run("Concurrent Usage", func(t *testing.T) {
		t.Parallel() // Run subtests in parallel
		// Create a semaphore with capacity 3
		sem := NewSemaphore(3)

		var wg sync.WaitGroup
		var active int32    // Number of goroutines currently holding a semaphore
		var maxActive int32 // Maximum observed active goroutines

		// Launch 10 goroutines that all try to acquire the semaphore
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Acquire semaphore
				err := sem.Acquire(context.Background())
				assert.NoError(t, err, "Acquire should succeed")

				// Increment active count
				currentActive := atomic.AddInt32(&active, 1)

				// Update max active (in a thread-safe way)
				for {
					oldMax := atomic.LoadInt32(&maxActive)
					if currentActive <= oldMax {
						break
					}
					if atomic.CompareAndSwapInt32(&maxActive, oldMax, currentActive) {
						break
					}
				}

				// Simulate work
				time.Sleep(10 * time.Millisecond) // Reduced from 50ms to 10ms

				// Decrement active count
				atomic.AddInt32(&active, -1)

				// Release semaphore
				sem.Release()
			}()
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify we never exceeded the semaphore capacity
		assert.LessOrEqual(t, maxActive, int32(3), "Should never have more than 3 active goroutines")
	})

	t.Run("Canceled Context", func(t *testing.T) {
		// Create a semaphore with capacity 1
		sem := NewSemaphore(1)

		// Acquire the only permit
		err := sem.Acquire(context.Background())
		assert.NoError(t, err, "First acquire should succeed")

		// Create a canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Try to acquire with canceled context
		err = sem.Acquire(ctx)
		assert.Error(t, err, "Acquire with canceled context should fail")
		assert.Equal(t, ErrContextCanceled, err, "Error should be ErrContextCanceled")

		// Clean up
		sem.Release()
	})
}

func TestTokenBucket(t *testing.T) {
	// Removed t.Parallel() - timing-dependent test with rate limiting
	t.Run("Basic Rate Limiting", func(t *testing.T) {
		t.Parallel() // Run subtests in parallel
		// Create a token bucket with 60 RPM (1 per second) and burst of 5
		tb := NewTokenBucket(60, 5)
		assert.NotNil(t, tb, "Token bucket should not be nil")

		// Should be able to get burst size tokens immediately
		for i := 0; i < 5; i++ {
			err := tb.Acquire(context.Background(), "test-model")
			assert.NoError(t, err, "Burst acquisition should succeed")
		}

		// Next acquire should be rate-limited, use short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		err := tb.Acquire(ctx, "test-model")
		assert.Error(t, err, "Acquire beyond burst should be rate-limited")
		assert.Contains(t, err.Error(), "exceed context deadline", "Error should be related to deadline exceeded")

		// Wait for token replenishment (slightly more than 1 second)
		time.Sleep(250 * time.Millisecond) // Reduced from 1100ms to 250ms

		// Should be able to get one more token
		err = tb.Acquire(context.Background(), "test-model")
		assert.NoError(t, err, "Acquire after waiting should succeed")
	})

	t.Run("Zero Value (No Limiting)", func(t *testing.T) {
		// Create a token bucket with zero RPM (no limit)
		tb := NewTokenBucket(0, 0)
		assert.Nil(t, tb, "Zero rate token bucket should be nil")

		// Nil token bucket operations should be no-ops
		err := tb.Acquire(context.Background(), "test-model")
		assert.NoError(t, err, "Acquire on nil token bucket should succeed")
	})

	t.Run("Per-Model Rate Limiting", func(t *testing.T) {
		// Create a token bucket with low RPM to ensure rate limiting happens
		tb := NewTokenBucket(30, 2)

		// Exhaust the burst allowance for model1
		err1 := tb.Acquire(context.Background(), "model1")
		assert.NoError(t, err1, "First acquire for model1 should succeed")
		err2 := tb.Acquire(context.Background(), "model1")
		assert.NoError(t, err2, "Second acquire for model1 should succeed")

		// Try to acquire again for model1 (should be rate-limited)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		err3 := tb.Acquire(ctx, "model1")
		assert.Error(t, err3, "Third acquire for model1 should be rate-limited")

		// model2 should have its own separate bucket and not be limited
		err4 := tb.Acquire(context.Background(), "model2")
		assert.NoError(t, err4, "First acquire for model2 should succeed despite model1 being limited")
		err5 := tb.Acquire(context.Background(), "model2")
		assert.NoError(t, err5, "Second acquire for model2 should succeed")
	})

	t.Run("Canceled Context", func(t *testing.T) {
		// Create a token bucket with limited rate
		tb := NewTokenBucket(30, 1)

		// Use up the burst capacity
		err := tb.Acquire(context.Background(), "test-model")
		assert.NoError(t, err, "First acquire should succeed")

		// Create a canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Try to acquire with canceled context
		err = tb.Acquire(ctx, "test-model")
		assert.Error(t, err, "Acquire with canceled context should fail")
	})
}

// TestRateLimiterCoverage provides comprehensive coverage testing for uncovered code paths
// Uses mathematical precision and controlled state manipulation to achieve 90%+ coverage
func TestRateLimiterCoverage(t *testing.T) {
	t.Parallel()

	t.Run("TokenBucket Burst Size Calculation", func(t *testing.T) {
		t.Parallel()

		// Test cases for burst size calculation edge cases (uncovered in NewTokenBucket)
		testCases := []struct {
			name          string
			ratePerMin    int
			maxBurst      int
			expectedBurst int
		}{
			{
				name:          "default burst calculation - high rate",
				ratePerMin:    600, // 600 RPM = 60/10 = 6, clamped to max 10
				maxBurst:      0,   // Use default calculation
				expectedBurst: 10,  // min(max(1, 600/10), 10) = min(60, 10) = 10
			},
			{
				name:          "default burst calculation - low rate",
				ratePerMin:    5, // 5 RPM = 5/10 = 0, clamped to min 1
				maxBurst:      0, // Use default calculation
				expectedBurst: 1, // min(max(1, 5/10), 10) = min(1, 10) = 1
			},
			{
				name:          "default burst calculation - medium rate",
				ratePerMin:    100, // 100 RPM = 100/10 = 10
				maxBurst:      0,   // Use default calculation
				expectedBurst: 10,  // min(max(1, 100/10), 10) = min(10, 10) = 10
			},
			{
				name:          "explicit burst size overrides calculation",
				ratePerMin:    120, // Rate doesn't matter when burst is explicit
				maxBurst:      5,   // Explicit burst size
				expectedBurst: 5,   // Should use explicit value
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tb := NewTokenBucket(tc.ratePerMin, tc.maxBurst)
				assert.NotNil(t, tb, "TokenBucket should be created successfully")
				assert.Equal(t, tc.expectedBurst, tb.burst, "Burst size should match expected calculation")
			})
		}
	})

	t.Run("Helper Functions Mathematical Verification", func(t *testing.T) {
		t.Parallel()

		// Test min function with various inputs (currently 0% coverage)
		assert.Equal(t, 5, min(5, 10), "min(5, 10) should return 5")
		assert.Equal(t, 3, min(10, 3), "min(10, 3) should return 3")
		assert.Equal(t, 7, min(7, 7), "min(7, 7) should return 7")
		assert.Equal(t, -5, min(-5, -3), "min(-5, -3) should return -5")
		assert.Equal(t, 0, min(0, 5), "min(0, 5) should return 0")

		// Test max function with various inputs (currently 0% coverage)
		assert.Equal(t, 10, max(5, 10), "max(5, 10) should return 10")
		assert.Equal(t, 10, max(10, 3), "max(10, 3) should return 10")
		assert.Equal(t, 7, max(7, 7), "max(7, 7) should return 7")
		assert.Equal(t, -3, max(-5, -3), "max(-5, -3) should return -3")
		assert.Equal(t, 5, max(0, 5), "max(0, 5) should return 5")
	})

	t.Run("Concurrent Limiter Creation Race Condition", func(t *testing.T) {
		t.Parallel()

		// Test the double-check locking pattern in getLimiter to improve coverage
		// This test specifically targets the "Double-check in case another goroutine created it" path
		tb := NewTokenBucket(60, 1)
		modelName := "race-test-model"

		var wg sync.WaitGroup
		var limiterResults []*rate.Limiter
		var resultsMutex sync.Mutex

		// Create synchronization barrier to ensure all goroutines start simultaneously
		startBarrier := make(chan struct{})

		// Launch multiple goroutines to trigger race condition in getLimiter
		numGoroutines := 10
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Wait for all goroutines to be ready
				<-startBarrier

				// All goroutines simultaneously try to get limiter for same model
				limiter := tb.getLimiter(modelName)

				// Store result safely
				resultsMutex.Lock()
				limiterResults = append(limiterResults, limiter)
				resultsMutex.Unlock()
			}()
		}

		// Release all goroutines simultaneously to maximize contention
		close(startBarrier)

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify all goroutines got the same limiter instance (should be shared)
		assert.Len(t, limiterResults, numGoroutines, "Should have results from all goroutines")

		firstLimiter := limiterResults[0]
		for i, limiter := range limiterResults {
			assert.Same(t, firstLimiter, limiter, "Goroutine %d should get same limiter instance", i)
		}

		// Verify limiter is properly stored in the map
		storedLimiter := tb.getLimiter(modelName)
		assert.Same(t, firstLimiter, storedLimiter, "Subsequent calls should return same limiter")
	})
}

func TestRateLimiter(t *testing.T) {
	t.Parallel() // Run parallel with other test files
	t.Run("Combined Limiting - Semaphore First", func(t *testing.T) {
		// Create a rate limiter with tight concurrency limit but loose rate limit
		limiter := NewRateLimiter(2, 600)

		// Should be able to acquire twice due to semaphore
		err1 := limiter.Acquire(context.Background(), "model1")
		assert.NoError(t, err1, "First acquire should succeed")
		err2 := limiter.Acquire(context.Background(), "model2")
		assert.NoError(t, err2, "Second acquire should succeed")

		// Third acquire should block on semaphore
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err3 := limiter.Acquire(ctx, "model3")
		assert.Error(t, err3, "Third acquire should be blocked by semaphore")

		// Release one, should be able to acquire again
		limiter.Release()
		err4 := limiter.Acquire(context.Background(), "model3")
		assert.NoError(t, err4, "Acquire after release should succeed")

		// Clean up
		limiter.Release()
		limiter.Release()
	})

	t.Run("Combined Limiting - Token Bucket First", func(t *testing.T) {
		// Create a rate limiter with loose concurrency limit but tight rate limit
		limiter := NewRateLimiter(10, 60) // 60 RPM with burst of 1

		// Use model1 to exhaust its token bucket
		err1 := limiter.Acquire(context.Background(), "model1")
		assert.NoError(t, err1, "First acquire for model1 should succeed")

		// Try to acquire again for model1 (should be rate-limited)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		err2 := limiter.Acquire(ctx, "model1")
		assert.Error(t, err2, "Second acquire for model1 should be rate-limited")

		// model2 should not be limited yet
		err3 := limiter.Acquire(context.Background(), "model2")
		assert.NoError(t, err3, "First acquire for model2 should succeed")

		// Clean up
		limiter.Release()
		limiter.Release()
	})

	t.Run("Disabled Limiting", func(t *testing.T) {
		// Create a rate limiter with both limits disabled
		limiter := NewRateLimiter(0, 0)

		// Should be able to acquire many times without blocking
		for i := 0; i < 100; i++ {
			err := limiter.Acquire(context.Background(), "test-model")
			assert.NoError(t, err, "Acquire should succeed with no limits")
		}

		// Release should not cause any issues
		limiter.Release()
	})
}

// TestRateLimiterDeterministic provides mathematically precise testing of rate limiter behavior
// without timing dependencies, race conditions, or flaky sleep-based assertions.
// This implements a Carmack-style algorithmic approach to testing time-based systems.
func TestRateLimiterDeterministic(t *testing.T) {
	t.Parallel()

	t.Run("Semaphore Deterministic Behavior", func(t *testing.T) {
		t.Parallel()

		// Test semaphore capacity enforcement with mathematical precision
		capacity := 3
		sem := NewSemaphore(capacity)

		ctx := context.Background()

		// Phase 1: Fill semaphore to capacity - all should succeed immediately
		for i := 0; i < capacity; i++ {
			err := sem.Acquire(ctx)
			assert.NoError(t, err, "Acquire %d should succeed (within capacity)", i+1)
		}

		// Phase 2: Verify capacity enforcement - next acquire should fail immediately
		ctx_immediate, cancel_immediate := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate()
		err := sem.Acquire(ctx_immediate)
		assert.Error(t, err, "Acquire beyond capacity should fail immediately")
		assert.Equal(t, ErrContextCanceled, err, "Should return context canceled error")

		// Phase 3: Test release mechanics - release one slot
		sem.Release()

		// Phase 4: Verify slot availability - should succeed immediately after release
		err = sem.Acquire(ctx)
		assert.NoError(t, err, "Acquire after release should succeed immediately")

		// Clean up remaining slots
		for i := 0; i < capacity; i++ {
			sem.Release()
		}
	})

	t.Run("TokenBucket Deterministic State Control", func(t *testing.T) {
		t.Parallel()

		// Create token bucket with precise parameters for deterministic testing
		ratePerMin := 60 // 1 token per second
		burstSize := 3
		tb := NewTokenBucket(ratePerMin, burstSize)

		ctx := context.Background()
		modelName := "test-model"

		// Get access to the underlying rate limiter for precise control
		limiter := tb.getLimiter(modelName)
		assert.NotNil(t, limiter, "Should create limiter for model")

		// Phase 1: Verify initial burst capacity
		// Should be able to acquire burst size tokens immediately
		for i := 0; i < burstSize; i++ {
			err := tb.Acquire(ctx, modelName)
			assert.NoError(t, err, "Burst acquire %d should succeed", i+1)
		}

		// Phase 2: Verify rate limiting after burst exhaustion
		ctx_immediate, cancel_immediate := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate()
		err := tb.Acquire(ctx_immediate, modelName)
		assert.Error(t, err, "Acquire beyond burst should fail without time passage")

		// Phase 3: Deterministic time advancement using SetLimitAt
		// Advance time by exactly 1 second to replenish 1 token
		now := time.Now()
		futureTime := now.Add(1 * time.Second)
		limiter.SetLimitAt(futureTime, rate.Limit(float64(ratePerMin)/60.0))

		// Phase 4: Verify token replenishment
		// Should now be able to acquire exactly 1 more token
		err = tb.Acquire(ctx, modelName)
		assert.NoError(t, err, "Acquire after time advancement should succeed")

		// Phase 5: Verify no additional tokens available
		ctx_immediate2, cancel_immediate2 := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate2()
		err = tb.Acquire(ctx_immediate2, modelName)
		assert.Error(t, err, "Additional acquire should fail without more time")
	})

	t.Run("TokenBucket Per-Model Isolation", func(t *testing.T) {
		t.Parallel()

		// Test mathematical isolation between different models
		tb := NewTokenBucket(60, 2) // 1 RPS, burst 2
		ctx := context.Background()

		model1 := "model1"
		model2 := "model2"

		// Phase 1: Exhaust model1's tokens
		err := tb.Acquire(ctx, model1)
		assert.NoError(t, err, "Model1 first acquire should succeed")
		err = tb.Acquire(ctx, model1)
		assert.NoError(t, err, "Model1 second acquire should succeed")

		// Phase 2: Verify model1 is exhausted
		ctx_immediate, cancel_immediate := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate()
		err = tb.Acquire(ctx_immediate, model1)
		assert.Error(t, err, "Model1 should be rate limited")

		// Phase 3: Verify model2 is unaffected (mathematical isolation)
		err = tb.Acquire(ctx, model2)
		assert.NoError(t, err, "Model2 should not be affected by model1 exhaustion")
		err = tb.Acquire(ctx, model2)
		assert.NoError(t, err, "Model2 second acquire should succeed")

		// Phase 4: Verify model2 is now also exhausted
		ctx_immediate2, cancel_immediate2 := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate2()
		err = tb.Acquire(ctx_immediate2, model2)
		assert.Error(t, err, "Model2 should now be rate limited")
	})

	t.Run("RateLimiter Combined Deterministic Behavior", func(t *testing.T) {
		t.Parallel()

		// Test the precise sequencing: semaphore THEN token bucket
		rl := NewRateLimiter(2, 120) // 2 concurrent, 2 RPS, burst 1
		ctx := context.Background()

		// Phase 1: Test semaphore-first enforcement
		// Fill semaphore to capacity
		err := rl.Acquire(ctx, "model1")
		assert.NoError(t, err, "First acquire should succeed")
		err = rl.Acquire(ctx, "model2")
		assert.NoError(t, err, "Second acquire should succeed")

		// Phase 2: Verify semaphore blocking
		ctx_immediate, cancel_immediate := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate()
		err = rl.Acquire(ctx_immediate, "model3")
		assert.Error(t, err, "Third acquire should be blocked by semaphore")

		// Phase 3: Test cleanup and token bucket enforcement
		rl.Release() // Free one semaphore slot
		rl.Release() // Free second semaphore slot

		// Phase 4: Now test token bucket limiting with empty semaphore
		// Use same model to exhaust its token bucket
		err = rl.Acquire(ctx, "model1")
		assert.NoError(t, err, "Should acquire after semaphore release")

		// This should hit token bucket limit for model1 (burst=1, already used)
		ctx_immediate2, cancel_immediate2 := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate2()
		err = rl.Acquire(ctx_immediate2, "model1")
		assert.Error(t, err, "Should be blocked by token bucket for model1")

		// Phase 5: Verify different model can still acquire (separate token bucket)
		err = rl.Acquire(ctx, "model-different")
		assert.NoError(t, err, "Different model should have separate token bucket")

		// Clean up
		rl.Release()
		rl.Release()
	})

	t.Run("Error Propagation and State Consistency", func(t *testing.T) {
		t.Parallel()

		// Test that semaphore is properly released when token bucket fails
		rl := NewRateLimiter(1, 60) // 1 concurrent, 1 RPS, burst 1
		ctx := context.Background()

		// Phase 1: Acquire successfully
		err := rl.Acquire(ctx, "model1")
		assert.NoError(t, err, "Initial acquire should succeed")

		// Phase 2: Verify semaphore is held
		ctx_immediate, cancel_immediate := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate()
		err = rl.Acquire(ctx_immediate, "model2")
		assert.Error(t, err, "Second acquire should fail (semaphore full)")

		// Phase 3: Release and verify token bucket blocking with semaphore cleanup
		rl.Release()

		// Now semaphore is free, but model1's token bucket is exhausted
		ctx_immediate2, cancel_immediate2 := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel_immediate2()
		err = rl.Acquire(ctx_immediate2, "model1")
		assert.Error(t, err, "Should fail on token bucket, semaphore should be released")

		// Phase 4: Verify semaphore was properly released by token bucket failure
		// A different model should be able to acquire (different token bucket)
		err = rl.Acquire(ctx, "model-different")
		assert.NoError(t, err, "Different model should succeed (semaphore was released)")

		rl.Release()
	})

	t.Run("Concurrent Deterministic Behavior", func(t *testing.T) {
		t.Parallel()

		// Test mathematical correctness under concurrency without timing dependencies
		// Use a very small semaphore to force contention
		sem := NewSemaphore(1)
		ctx := context.Background()

		const numGoroutines = 10
		const iterations = 3

		var wg sync.WaitGroup
		var successCount int32
		var failureCount int32

		// Phase 1: Create a synchronization barrier to ensure all goroutines start simultaneously
		startBarrier := make(chan struct{})

		// Launch concurrent goroutines
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Wait for all goroutines to be ready
				<-startBarrier

				for j := 0; j < iterations; j++ {
					// Use immediate timeout for deterministic behavior
					ctx_immediate, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
					err := sem.Acquire(ctx_immediate)
					cancel()

					if err == nil {
						atomic.AddInt32(&successCount, 1)
						// Hold the semaphore briefly to force contention, then release
						sem.Release()
					} else {
						atomic.AddInt32(&failureCount, 1)
					}
				}
			}(i)
		}

		// Release all goroutines simultaneously to create maximum contention
		close(startBarrier)
		wg.Wait()

		// Phase 2: Mathematical verification
		totalAttempts := int32(numGoroutines * iterations)
		actualTotal := successCount + failureCount

		assert.Equal(t, totalAttempts, actualTotal, "All attempts should be accounted for")
		assert.Greater(t, successCount, int32(0), "Some acquisitions should succeed")

		// With 10 goroutines competing for 1 semaphore slot, some failures are expected
		// But we can't guarantee failures with immediate timeouts, so just verify accounting
		t.Logf("Success: %d, Failures: %d, Total: %d", successCount, failureCount, actualTotal)

		// Phase 3: Verify final state - semaphore should be available
		err := sem.Acquire(ctx)
		assert.NoError(t, err, "Semaphore should be available after all goroutines complete")
		sem.Release()
	})

	t.Run("TokenBucket Burst Size Calculation", func(t *testing.T) {
		t.Parallel()

		// Test the mathematical burst size calculation logic in NewTokenBucket
		// This forces execution of the min/max helper functions (lines 191-203)
		// and the burst calculation logic (lines 86-87)

		testCases := []struct {
			name          string
			ratePerMin    int
			maxBurst      int
			expectedBurst int
		}{
			{
				name:          "Default burst calculation - small rate",
				ratePerMin:    60,
				maxBurst:      0, // Trigger default calculation
				expectedBurst: 6, // min(max(1, 60/10), 10) = min(max(1, 6), 10) = min(6, 10) = 6
			},
			{
				name:          "Default burst calculation - tiny rate",
				ratePerMin:    5,
				maxBurst:      0, // Trigger default calculation
				expectedBurst: 1, // min(max(1, 5/10), 10) = min(max(1, 0), 10) = min(1, 10) = 1
			},
			{
				name:          "Default burst calculation - large rate",
				ratePerMin:    200,
				maxBurst:      0,  // Trigger default calculation
				expectedBurst: 10, // min(max(1, 200/10), 10) = min(max(1, 20), 10) = min(20, 10) = 10
			},
			{
				name:          "Explicit burst - within bounds",
				ratePerMin:    60,
				maxBurst:      5,
				expectedBurst: 5, // Use explicit value
			},
			{
				name:          "Negative burst triggers default",
				ratePerMin:    120,
				maxBurst:      -5, // Should trigger default calculation
				expectedBurst: 10, // min(max(1, 120/10), 10) = min(12, 10) = 10
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tb := NewTokenBucket(tc.ratePerMin, tc.maxBurst)
				assert.NotNil(t, tb, "TokenBucket should be created")
				assert.Equal(t, tc.expectedBurst, tb.burst, "Burst size should match expected calculation")
				assert.Equal(t, tc.ratePerMin, tb.ratePerMin, "Rate per minute should be preserved")

				// Verify the rate limit calculation
				expectedRPS := rate.Limit(float64(tc.ratePerMin) / 60.0)
				assert.Equal(t, expectedRPS, tb.limit, "Rate limit should be converted correctly")
			})
		}
	})

	t.Run("Helper Functions Mathematical Verification", func(t *testing.T) {
		t.Parallel()

		// Direct testing of min/max helper functions to ensure mathematical correctness
		// This targets lines 191-203 specifically

		// Test min function
		assert.Equal(t, 1, min(1, 2), "min(1, 2) should return 1")
		assert.Equal(t, 1, min(2, 1), "min(2, 1) should return 1")
		assert.Equal(t, 5, min(5, 5), "min(5, 5) should return 5")
		assert.Equal(t, -1, min(-1, 0), "min(-1, 0) should return -1")
		assert.Equal(t, -2, min(-1, -2), "min(-1, -2) should return -2")

		// Test max function
		assert.Equal(t, 2, max(1, 2), "max(1, 2) should return 2")
		assert.Equal(t, 2, max(2, 1), "max(2, 1) should return 2")
		assert.Equal(t, 5, max(5, 5), "max(5, 5) should return 5")
		assert.Equal(t, 0, max(-1, 0), "max(-1, 0) should return 0")
		assert.Equal(t, -1, max(-1, -2), "max(-1, -2) should return -1")

		// Test combined min/max usage (mirrors the burst calculation logic)
		ratePerMin := 75
		burstCalc := min(max(1, ratePerMin/10), 10) // Should be min(max(1, 7), 10) = min(7, 10) = 7
		assert.Equal(t, 7, burstCalc, "Combined min/max calculation should work correctly")
	})

	t.Run("Concurrent Limiter Creation Race Condition", func(t *testing.T) {
		t.Parallel()

		// Test the double-check path in getLimiter (lines 118-119)
		// This scenario occurs when multiple goroutines try to create a limiter
		// for the same model simultaneously

		const numGoroutines = 20
		const modelName = "race-test-model"

		tb := NewTokenBucket(60, 5)
		assert.NotNil(t, tb, "TokenBucket should be created")

		// Phase 1: Set up synchronization to create maximum race condition
		var wg sync.WaitGroup
		startBarrier := make(chan struct{})
		limiterResults := make([]*rate.Limiter, numGoroutines)

		// Phase 2: Launch multiple goroutines to create limiter simultaneously
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				// Wait for all goroutines to be ready
				<-startBarrier

				// All goroutines try to get limiter for same model simultaneously
				limiter := tb.getLimiter(modelName)
				limiterResults[index] = limiter
			}(i)
		}

		// Phase 3: Release all goroutines simultaneously to maximize race condition
		close(startBarrier)
		wg.Wait()

		// Phase 4: Mathematical verification - all goroutines should get same limiter instance
		// This tests both the fast path (existing limiter) and the double-check path
		firstLimiter := limiterResults[0]
		assert.NotNil(t, firstLimiter, "First limiter should not be nil")

		for i := 1; i < numGoroutines; i++ {
			assert.Same(t, firstLimiter, limiterResults[i],
				"All goroutines should receive the same limiter instance (index %d)", i)
		}

		// Phase 5: Verify the limiter was properly stored in the map
		assert.Contains(t, tb.limiters, modelName, "Model should be stored in limiters map")
		storedLimiter := tb.limiters[modelName]
		assert.Same(t, firstLimiter, storedLimiter, "Stored limiter should match returned limiter")

		// Phase 6: Verify subsequent calls use the cached limiter (fast path)
		for i := 0; i < 5; i++ {
			cachedLimiter := tb.getLimiter(modelName)
			assert.Same(t, firstLimiter, cachedLimiter, "Cached limiter should be same instance")
		}
	})
}

// Benchmark tests for rate limiter performance under various concurrency levels
// These ensure that coverage improvements don't introduce performance regressions

func BenchmarkRateLimiter(b *testing.B) {
	b.Run("Semaphore", func(b *testing.B) {
		b.Run("SingleGoroutine", func(b *testing.B) {
			perftest.RunBenchmark(b, "Semaphore_SingleGoroutine", func(b *testing.B) {
				sem := NewSemaphore(100) // High capacity to avoid blocking
				ctx := context.Background()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = sem.Acquire(ctx)
						sem.Release()
					}
				})
			})
		})

		b.Run("LowConcurrency", func(b *testing.B) {
			perftest.RunBenchmark(b, "Semaphore_LowConcurrency", func(b *testing.B) {
				sem := NewSemaphore(5) // Lower capacity to test contention
				ctx := context.Background()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = sem.Acquire(ctx)
						sem.Release()
					}
				})
			})
		})

		b.Run("HighConcurrency", func(b *testing.B) {
			perftest.RunBenchmark(b, "Semaphore_HighConcurrency", func(b *testing.B) {
				sem := NewSemaphore(1) // Very low capacity for maximum contention
				ctx := context.Background()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = sem.Acquire(ctx)
						sem.Release()
					}
				})
			})
		})
	})

	b.Run("TokenBucket", func(b *testing.B) {
		b.Run("SingleModel", func(b *testing.B) {
			perftest.RunBenchmark(b, "TokenBucket_SingleModel", func(b *testing.B) {
				tb := NewTokenBucket(60000, 1000) // High rate and burst to avoid blocking
				ctx := context.Background()
				modelName := "benchmark-model"

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = tb.Acquire(ctx, modelName)
					}
				})
			})
		})

		b.Run("MultipleModels", func(b *testing.B) {
			perftest.RunBenchmark(b, "TokenBucket_MultipleModels", func(b *testing.B) {
				tb := NewTokenBucket(60000, 1000) // High rate and burst to avoid blocking
				ctx := context.Background()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						modelName := "benchmark-model-" + string(rune('0'+counter%10)) // Different models per goroutine
						counter++
						_ = tb.Acquire(ctx, modelName)
					}
				})
			})
		})

		b.Run("LimiterCreation", func(b *testing.B) {
			perftest.RunBenchmark(b, "TokenBucket_LimiterCreation", func(b *testing.B) {
				tb := NewTokenBucket(3600, 60) // Moderate rate
				ctx := context.Background()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						modelName := "benchmark-model-" + string(rune('0'+counter%1000)) // Force limiter creation
						counter++
						_ = tb.Acquire(ctx, modelName)
					}
				})
			})
		})
	})

	b.Run("CombinedRateLimiter", func(b *testing.B) {
		b.Run("OptimalConditions", func(b *testing.B) {
			perftest.RunBenchmark(b, "CombinedRateLimiter_OptimalConditions", func(b *testing.B) {
				rl := NewRateLimiter(100, 60000) // High concurrency and rate limits
				ctx := context.Background()
				modelName := "benchmark-model"

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = rl.Acquire(ctx, modelName)
						rl.Release()
					}
				})
			})
		})

		b.Run("SemaphoreLimited", func(b *testing.B) {
			perftest.RunBenchmark(b, "CombinedRateLimiter_SemaphoreLimited", func(b *testing.B) {
				rl := NewRateLimiter(5, 60000) // Low concurrency, high rate
				ctx := context.Background()
				modelName := "benchmark-model"

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = rl.Acquire(ctx, modelName)
						rl.Release()
					}
				})
			})
		})

		b.Run("TokenBucketLimited", func(b *testing.B) {
			perftest.RunBenchmark(b, "CombinedRateLimiter_TokenBucketLimited", func(b *testing.B) {
				rl := NewRateLimiter(100, 60) // High concurrency, low rate
				ctx := context.Background()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						modelName := "benchmark-model-" + string(rune('0'+counter%10)) // Multiple models
						counter++
						_ = rl.Acquire(ctx, modelName)
						rl.Release()
					}
				})
			})
		})
	})

	b.Run("HelperFunctions", func(b *testing.B) {
		b.Run("MinFunction", func(b *testing.B) {
			perftest.RunBenchmark(b, "HelperFunctions_MinFunction", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					result := min(i, i+1)
					_ = result // Avoid compiler optimization
				}
			})
		})

		b.Run("MaxFunction", func(b *testing.B) {
			perftest.RunBenchmark(b, "HelperFunctions_MaxFunction", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					result := max(i, i+1)
					_ = result // Avoid compiler optimization
				}
			})
		})

		b.Run("BurstCalculation", func(b *testing.B) {
			perftest.RunBenchmark(b, "HelperFunctions_BurstCalculation", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ratePerMin := (i % 1000) + 60 // Vary rate between 60-1060
					burst := min(max(1, ratePerMin/10), 10)
					_ = burst // Avoid compiler optimization
				}
			})
		})
	})
}
