# Debug Analysis for Rate Limiter Deadlock

## Root Cause Analysis
All three Gemini models identified the same root cause for the deadlock:

The deadlock is caused by an **overly broad locking strategy** in the `RateLimiter` implementation. Specifically:

1. The `RateLimiter` holds its primary mutex (`rl.mu`) across potentially long-blocking operations:
   - `rl.semaphore.Acquire(ctx)` - blocks when no semaphore tickets are available
   - `rl.tokenBucket.Acquire(ctx, modelName)` - blocks when rate limit is hit

2. This creates a classic deadlock scenario:
   - Goroutine A calls `Acquire`, locks `rl.mu`, then blocks inside `semaphore.Acquire` waiting for a ticket
   - Goroutine B finishes its task and calls `Release` to free a ticket, but can't proceed because it's trying to acquire `rl.mu` (held by Goroutine A)
   - Goroutine A can't proceed without Goroutine B releasing a ticket, but Goroutine B can't proceed without Goroutine A releasing the mutex

3. The mutex `rl.mu` is unnecessary because:
   - The underlying `Semaphore` (channel-based) and `TokenBucket` (using its own `RWMutex`) already safely handle their own concurrency
   - The critical operation is acquiring these resources sequentially, not locking around both operations

## Recommended Fix
All models recommend the same fix: removing the unnecessary mutex entirely from the `RateLimiter`:

```diff
// RateLimiter combines semaphore and token bucket limiters
type RateLimiter struct {
-	mu          sync.Mutex // Mutex to protect concurrent access
	semaphore   *Semaphore
	tokenBucket *TokenBucket
}

// Acquire waits to acquire both semaphore and rate limit permissions
func (rl *RateLimiter) Acquire(ctx context.Context, modelName string) error {
-	rl.mu.Lock()
-	defer rl.mu.Unlock()

	// First try to acquire the semaphore
	if err := rl.semaphore.Acquire(ctx); err != nil {
		return err
	}

	// If we got the semaphore but fail to get the rate limit, release the semaphore
	if err := rl.tokenBucket.Acquire(ctx, modelName); err != nil {
		rl.semaphore.Release()
		return err
	}

	return nil
}

// Release releases the semaphore (token bucket doesn't need explicit release)
func (rl *RateLimiter) Release() {
-	rl.mu.Lock()
-	defer rl.mu.Unlock()

	rl.semaphore.Release()
	// No explicit release needed for token bucket
}
```

Inline comments should explain:
```go
// BUGFIX: Remove unnecessary mutex causing deadlocks in concurrent Acquire/Release calls.
// CAUSE: Holding rl.mu across blocking calls (semaphore.Acquire, tokenBucket.Acquire)
//        prevented Release calls (which also needed rl.mu) from freeing resources,
//        leading to deadlock when resources were contended.
// FIX: Removed rl.mu entirely. Acquire semaphore then token bucket sequentially.
//      Release semaphore immediately if token bucket acquisition fails.
//      The underlying Semaphore and TokenBucket handle their own concurrency.
```

## Verification
To verify the fix:
1. Apply the changes to `internal/ratelimit/ratelimit.go`
2. Run the tests focusing on the rate limit features:
   ```bash
   go test ./internal/integration/... -v -count=1 -race -run TestRateLimitFeatures
   ```
3. Run the full test suite to ensure no regressions:
   ```bash
   go test ./... -v -count=1 -race
   ```