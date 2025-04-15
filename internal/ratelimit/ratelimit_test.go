package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSemaphore(t *testing.T) {
	t.Parallel() // Run parallel with other test files
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
	t.Parallel() // Run parallel with other test files
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
