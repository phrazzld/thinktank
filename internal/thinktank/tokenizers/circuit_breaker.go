package tokenizers

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitClosed - normal operation, requests are allowed
	CircuitClosed CircuitBreakerState = iota
	// CircuitOpen - circuit is open, requests are blocked
	CircuitOpen
	// CircuitHalfOpen - limited requests allowed to test if service is back
	CircuitHalfOpen
)

// String returns a string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern for tokenizer fault tolerance
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitBreakerState
	failureCount     int
	lastFailureTime  time.Time
	nextRetryTime    time.Time
	failureThreshold int
	cooldownDuration time.Duration
	timeSource       func() time.Time // For testing
}

// NewCircuitBreaker creates a new circuit breaker with default settings
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: 5,                // Default: open after 5 failures
		cooldownDuration: 30 * time.Second, // Default: 30 second cooldown
		timeSource:       time.Now,         // Default: use real time
	}
}

// NewCircuitBreakerWithConfig creates a circuit breaker with custom settings
func NewCircuitBreakerWithConfig(failureThreshold int, cooldownDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		cooldownDuration: cooldownDuration,
		timeSource:       time.Now,
	}
}

// CanExecute returns true if the circuit breaker allows the request to proceed
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if enough time has passed to attempt a retry
		if cb.timeSource().After(cb.nextRetryTime) {
			// Transition to half-open to allow one test request
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation and may close the circuit
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Reset failure count on any success
	cb.failureCount = 0

	// If circuit is half-open, close it completely
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
}

// RecordFailure records a failed operation and may open the circuit
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = cb.timeSource()

	if cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
		cb.nextRetryTime = cb.timeSource().Add(cb.cooldownDuration)
	}
}

// GetState returns the current state of the circuit breaker (thread-safe)
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count (thread-safe)
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.GetState() == CircuitOpen
}

// SetTimeSource allows injection of time source for testing
func (cb *CircuitBreaker) SetTimeSource(timeSource func() time.Time) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.timeSource = timeSource
}
