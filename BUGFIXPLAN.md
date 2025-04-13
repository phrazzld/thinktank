# Bug Fix Plan

## Bug Description:
The CI workflow is failing with a deadlock in the rate limiting tests. The error shows several goroutines locked in a mutex deadlock state, with the following key lines in the stack trace:

1. `goroutine 33 [sync.Mutex.Lock]:`
2. `github.com/phrazzld/architect/internal/ratelimit.(*RateLimiter).Release(0xc0000c7a20?)`
3. `panic({0xb3b7a0?, 0x12b1580?})`
4. Test timed out after 10 minutes on `TestRateLimitFeatures`

Multiple goroutines are deadlocked waiting on mutex locks, particularly around the `RateLimiter.Acquire` and `RateLimiter.Release` methods.

## Reproduction Steps:
Running the tests in the CI workflow, particularly the rate limit tests:
```
go test -v ./internal/integration/...
```

## Expected Behavior:
The tests should complete successfully without any deadlocks or timeouts.

## Actual Behavior:
Tests are deadlocking, with multiple goroutines stuck in mutex locks. The test eventually times out after 10 minutes, producing a panic and stack trace showing:
1. Some goroutines are waiting on `sync.Mutex.Lock`
2. Specifically, goroutines are stuck in either `RateLimiter.Acquire` or `RateLimiter.Release`
3. The `TestRateLimitFeatures` test is timing out

## Key Components/Files Mentioned:
- `internal/ratelimit/ratelimit.go`: Contains the rate limiter implementation with the `Acquire` and `Release` methods that are deadlocking
- `internal/architect/orchestrator/orchestrator.go`: The orchestrator handling multiple models, using rate limiting
- `internal/integration/rate_limit_test.go`: Contains the `TestRateLimitFeatures` test that's timing out

## Hypotheses:
1. **Nested Lock Acquisition in RateLimiter**: The deadlock may be caused by nested lock acquisition in the `RateLimiter` implementation. Examining the code, `Acquire` and `Release` methods both lock the `RateLimiter.mu` mutex, and within these methods, other locks might be acquired (like in `TokenBucket.getLimiter`).
   - Reasoning: The stack trace shows goroutines waiting on mutex locks in both `Acquire` and `Release` methods, suggesting a circular lock dependency.
   - Validation: Examine the code path to identify any circular lock dependencies, and consider adding debug logging to track lock acquisition order.

2. **Panic During Release Causing Deadlock**: The stack trace shows a panic occurring within the `Release` method. This could leave locks in an inconsistent state, causing other goroutines to deadlock.
   - Reasoning: The presence of "panic" in the stack trace near the `Release` method suggests an unexpected exception occurred during cleanup.
   - Validation: Modify the `Release` method to use a recover() to safely handle any panics and ensure locks are always released.

3. **Concurrent Modification of Shared Resources**: The issue might be related to the test environment where multiple goroutines are modifying shared resources like the rate limiter or mock client.
   - Reasoning: Tests like `TestRateLimitFeatures` use a common pattern where goroutines access shared mock resources.
   - Validation: Review the test setup for proper synchronization of access to shared test data and mocks.

4. **Mutex Not Released in Error Path**: There might be an error path in the code where a mutex is acquired but not released, causing other goroutines to deadlock.
   - Reasoning: The `Acquire` method in `RateLimiter` has error-handling logic that involves releasing the semaphore, but might miss releasing a mutex in some error cases.
   - Validation: Review all error paths in `Acquire` and `Release` to ensure proper cleanup in all cases.

## Test Log:
### Test 1: Code Review Analysis
**Hypothesis Tested:** Nested Lock Acquisition in RateLimiter (H1) and the other hypotheses
**Test Description:** Detailed code review and analysis of the RateLimiter implementation
**Execution Method:** Manual examination of the code in `internal/ratelimit/ratelimit.go` and related stack traces
**Expected Result (if true):** Identification of lock acquisition patterns that could lead to deadlock
**Expected Result (if false):** No problematic lock patterns found, suggesting another cause
**Actual Result:** Confirmed Hypothesis 1. The `RateLimiter` holds its primary mutex (`rl.mu`) across potentially long-blocking operations (`semaphore.Acquire` and `tokenBucket.Acquire`). This creates a classic deadlock scenario where:
- Goroutine A calls `Acquire`, locks `rl.mu`, then blocks inside `semaphore.Acquire` waiting for a ticket
- Goroutine B finishes its task and calls `Release` to free a ticket, but can't proceed because it's trying to acquire `rl.mu` (held by Goroutine A)
- Goroutine A can't proceed without Goroutine B releasing a ticket, but Goroutine B can't proceed without Goroutine A releasing the mutex

The mutex `rl.mu` is also unnecessary because the underlying `Semaphore` (channel-based) and `TokenBucket` (using its own `RWMutex`) already handle their own concurrency safely.

## Root Cause:
The deadlock is caused by an **overly broad locking strategy** in the `RateLimiter` implementation. The `RateLimiter` class uses coarse-grained locking by acquiring its mutex (`rl.mu`) at the beginning of its `Acquire` and `Release` methods and holding it throughout the execution of these methods.

Within the `Acquire` method, while holding the lock, it calls potentially blocking operations:
1. `rl.semaphore.Acquire(ctx)` - this blocks when no semaphore tickets are available
2. `rl.tokenBucket.Acquire(ctx, modelName)` - this blocks when the rate limit is hit

When resources are exhausted (e.g., no semaphore tickets available), goroutines calling `Acquire` will hold the `rl.mu` lock while waiting for resources to become available. However, the resources can only be freed when other goroutines call `Release`, which also tries to acquire the same `rl.mu` lock. This creates a classic deadlock situation.

This mutex is actually unnecessary as the underlying `Semaphore` and `TokenBucket` components already handle their own concurrency correctly.

## Fix Description:
Remove the unnecessary mutex entirely from the `RateLimiter` struct and its usage in the `Acquire` and `Release` methods. The underlying components (`Semaphore` and `TokenBucket`) already handle their own concurrency safely.

**Code Changes:**

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

**Inline Comments:**

```go
// BUGFIX: Remove unnecessary mutex causing deadlocks in concurrent Acquire/Release calls.
// CAUSE: Holding rl.mu across blocking calls (semaphore.Acquire, tokenBucket.Acquire)
//        prevented Release calls (which also needed rl.mu) from freeing resources,
//        leading to deadlock when resources were contended.
// FIX: Removed rl.mu entirely. Acquire semaphore then token bucket sequentially.
//      Release semaphore immediately if token bucket acquisition fails.
//      The underlying Semaphore and TokenBucket handle their own concurrency.
```

## Status: Root cause identified, fix proposed
The bug investigation is complete. We have identified that the core issue is an unnecessary mutex causing a deadlock pattern. The fix is to remove this mutex entirely as the underlying semaphore and token bucket components already handle their own synchronization correctly. The next step is to implement and verify this fix.