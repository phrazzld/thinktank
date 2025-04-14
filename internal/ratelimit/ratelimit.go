// Package ratelimit provides concurrency control and rate limiting functionality
package ratelimit

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/time/rate"
)

var (
	// ErrContextCanceled is returned when the context is canceled during acquisition
	ErrContextCanceled = errors.New("context canceled while waiting for resource")
)

// Semaphore provides a simple mechanism for limiting concurrent operations
type Semaphore struct {
	tickets chan struct{}
}

// NewSemaphore creates a new semaphore with the given capacity
// If maxConcurrent is <= 0, returns nil (no limiting)
func NewSemaphore(maxConcurrent int) *Semaphore {
	if maxConcurrent <= 0 {
		return nil // No limit
	}
	return &Semaphore{
		tickets: make(chan struct{}, maxConcurrent),
	}
}

// Acquire gets a ticket from the semaphore, blocking if none are available
// Returns nil if successful, or error if the context is canceled
// Does nothing if semaphore is nil (no limiting)
func (s *Semaphore) Acquire(ctx context.Context) error {
	if s == nil {
		return nil // No limiting
	}

	select {
	case s.tickets <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrContextCanceled
	}
}

// Release returns a ticket to the semaphore
// Does nothing if semaphore is nil (no limiting)
func (s *Semaphore) Release() {
	if s == nil {
		return // No limiting
	}
	select {
	case <-s.tickets:
		// Successfully released a ticket
	default:
		// This should never happen in correct usage
		// but prevents deadlock if Release is called without Acquire
	}
}

// TokenBucket manages rate limiting using a token bucket algorithm
type TokenBucket struct {
	// Map of model names to limiters
	limiters   map[string]*rate.Limiter
	mutex      sync.RWMutex
	ratePerMin int
	limit      rate.Limit
	burst      int
}

// NewTokenBucket creates a new token bucket rate limiter
// If ratePerMin is <= 0, returns nil (no limiting)
func NewTokenBucket(ratePerMin, maxBurst int) *TokenBucket {
	if ratePerMin <= 0 {
		return nil // No limit
	}

	// Convert from per-minute to per-second (which is what rate.Limit uses)
	rps := rate.Limit(float64(ratePerMin) / 60.0)

	// Set burst size, defaulting to 1/10 of rate (min 1, max 10)
	if maxBurst <= 0 {
		maxBurst = min(max(1, ratePerMin/10), 10)
	}

	return &TokenBucket{
		limiters:   make(map[string]*rate.Limiter),
		ratePerMin: ratePerMin,
		limit:      rps,
		burst:      maxBurst,
	}
}

// getLimiter returns the rate limiter for a specific model, creating it if needed
func (tb *TokenBucket) getLimiter(modelName string) *rate.Limiter {
	// For no limit case
	if tb == nil {
		return nil
	}

	// Check if we already have a limiter for this model
	tb.mutex.RLock()
	limiter, exists := tb.limiters[modelName]
	tb.mutex.RUnlock()

	if exists {
		return limiter
	}

	// Create a new limiter for this model
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Double-check in case another goroutine created it
	if limiter, exists = tb.limiters[modelName]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(tb.limit, tb.burst)
	tb.limiters[modelName] = limiter
	return limiter
}

// Acquire waits to acquire a token, returns error if context canceled
func (tb *TokenBucket) Acquire(ctx context.Context, modelName string) error {
	if tb == nil {
		return nil // No limiting
	}

	limiter := tb.getLimiter(modelName)
	if limiter.Allow() {
		// Fast path - if we can get a token without waiting
		return nil
	}

	// Slow path - wait for a token to become available
	return limiter.Wait(ctx)
}

// RateLimiter combines semaphore and token bucket limiters
// BUGFIX: Remove unnecessary mutex causing deadlocks in concurrent Acquire/Release calls.
// CAUSE: Holding rl.mu across blocking calls (semaphore.Acquire, tokenBucket.Acquire)
//
//	prevented Release calls (which also needed rl.mu) from freeing resources,
//	leading to deadlock when resources were contended.
//
// FIX: Removed rl.mu entirely. Acquire semaphore then token bucket sequentially.
//
//	Release semaphore immediately if token bucket acquisition fails.
//	The underlying Semaphore and TokenBucket handle their own concurrency.
type RateLimiter struct {
	semaphore   *Semaphore
	tokenBucket *TokenBucket
}

// NewRateLimiter creates a new combined rate limiter
// By default, uses a burst size of 1 for the token bucket to make rate limiting more strict
func NewRateLimiter(maxConcurrent, ratePerMin int) *RateLimiter {
	return &RateLimiter{
		semaphore:   NewSemaphore(maxConcurrent),
		tokenBucket: NewTokenBucket(ratePerMin, 1), // Use explicit burst size of 1 for stricter rate limiting
	}
}

// Acquire waits to acquire both semaphore and rate limit permissions
func (rl *RateLimiter) Acquire(ctx context.Context, modelName string) error {
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
	rl.semaphore.Release()
	// No explicit release needed for token bucket
}

// Helper functions for min/max (Go 1.21+ has these in standard math package)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
