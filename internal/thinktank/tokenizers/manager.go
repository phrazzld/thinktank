// Package tokenizers provides accurate token counting implementations for various LLM providers.
// This package enables precise token counting to replace estimation-based model selection.
package tokenizers

import (
	"context"
	"sync"
	"time"
)

// tokenizerManagerImpl implements TokenizerManager with lazy loading and caching.
// This is the base implementation that provides core tokenizer management functionality.
type tokenizerManagerImpl struct {
	// tokenizerCache stores initialized tokenizers by provider name for fast access
	tokenizerCache sync.Map

	// initMutex ensures only one tokenizer initialization per provider to prevent races
	initMutex sync.Mutex
}

// NewTokenizerManager creates a new tokenizer manager with lazy loading.
// Tokenizers are initialized only when first requested to minimize startup overhead.
func NewTokenizerManager() TokenizerManager {
	return &tokenizerManagerImpl{}
}

// GetTokenizer returns a cached tokenizer or creates a new one for the provider.
// Uses double-checked locking pattern to ensure thread-safe lazy initialization.
func (m *tokenizerManagerImpl) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	// Check cache first for fast path
	if cached, ok := m.tokenizerCache.Load(provider); ok {
		return cached.(AccurateTokenCounter), nil
	}

	// Use mutex to prevent duplicate initialization
	m.initMutex.Lock()
	defer m.initMutex.Unlock()

	// Double-check cache after acquiring lock
	if cached, ok := m.tokenizerCache.Load(provider); ok {
		return cached.(AccurateTokenCounter), nil
	}

	// Create new tokenizer based on provider
	var tokenizer AccurateTokenCounter
	switch provider {
	case "openrouter":
		tokenizer = NewOpenRouterTokenizer()
	default:
		return nil, NewTokenizerErrorWithDetails(provider, "", "unsupported provider (only OpenRouter is supported after consolidation)", nil, "unknown")
	}

	// Cache the tokenizer for future use
	m.tokenizerCache.Store(provider, tokenizer)
	return tokenizer, nil
}

// SupportsProvider returns true if accurate tokenization is available for the provider.
func (m *tokenizerManagerImpl) SupportsProvider(provider string) bool {
	switch provider {
	case "openrouter":
		return true
	default:
		return false // Only OpenRouter is supported after provider consolidation
	}
}

// ClearCache clears all cached tokenizers to free memory.
func (m *tokenizerManagerImpl) ClearCache() {
	m.tokenizerCache.Range(func(key, value interface{}) bool {
		// Note: OpenRouter tokenizer wraps OpenAI tokenizer internally
		// The internal tokenizer will be garbage collected when the wrapper is deleted
		m.tokenizerCache.Delete(key)
		return true
	})
}

// tokenizerManagerWithCircuitBreaker extends tokenizerManagerImpl with circuit breaker functionality
type tokenizerManagerWithCircuitBreaker struct {
	tokenizerManagerImpl
	circuitBreakers map[string]*CircuitBreaker
	circuitMutex    sync.RWMutex
	mockTokenizers  map[string]AccurateTokenCounter // For testing
	mockMutex       sync.RWMutex
	currentTime     time.Time // For testing time advancement
	timeSource      func() time.Time
	usingMockTime   bool // Flag to track if we're using mock time
}

// NewTokenizerManagerWithCircuitBreaker creates a tokenizer manager with circuit breaker protection
func NewTokenizerManagerWithCircuitBreaker() *tokenizerManagerWithCircuitBreaker {
	return &tokenizerManagerWithCircuitBreaker{
		tokenizerManagerImpl: tokenizerManagerImpl{},
		circuitBreakers:      make(map[string]*CircuitBreaker),
		mockTokenizers:       make(map[string]AccurateTokenCounter),
		timeSource:           time.Now,
		usingMockTime:        false,
	}
}

// GetTokenizer returns a tokenizer with circuit breaker protection
func (m *tokenizerManagerWithCircuitBreaker) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	// Check if we have a mock tokenizer for testing
	m.mockMutex.RLock()
	if mockTokenizer, ok := m.mockTokenizers[provider]; ok {
		m.mockMutex.RUnlock()

		// Get circuit breaker for this provider
		circuitBreaker := m.getOrCreateCircuitBreaker(provider)

		// Check if circuit breaker allows execution
		if !circuitBreaker.CanExecute() {
			return nil, NewTokenizerErrorWithDetails(provider, "", "circuit breaker open", nil, getTokenizerType(provider))
		}

		// Test the mock tokenizer by calling CountTokens with empty input
		_, err := mockTokenizer.CountTokens(context.TODO(), "", "test")
		if err != nil {
			circuitBreaker.RecordFailure()
			return nil, NewTokenizerErrorWithDetails(provider, "", "tokenizer initialization failed", err, getTokenizerType(provider))
		}

		circuitBreaker.RecordSuccess()
		return mockTokenizer, nil
	}
	m.mockMutex.RUnlock()

	// Get circuit breaker for this provider
	circuitBreaker := m.getOrCreateCircuitBreaker(provider)

	// Check if circuit breaker allows execution
	if !circuitBreaker.CanExecute() {
		return nil, NewTokenizerErrorWithDetails(provider, "", "circuit breaker open", nil, getTokenizerType(provider))
	}

	// Call the original implementation
	tokenizer, err := m.tokenizerManagerImpl.GetTokenizer(provider)
	if err != nil {
		circuitBreaker.RecordFailure()
		return nil, err
	}

	circuitBreaker.RecordSuccess()
	return tokenizer, nil
}

// SetMockTokenizer sets a mock tokenizer for testing
func (m *tokenizerManagerWithCircuitBreaker) SetMockTokenizer(provider string, tokenizer AccurateTokenCounter) {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()
	m.mockTokenizers[provider] = tokenizer
}

// IsCircuitOpen returns true if the circuit breaker is open for the given provider
func (m *tokenizerManagerWithCircuitBreaker) IsCircuitOpen(provider string) bool {
	m.circuitMutex.RLock()
	defer m.circuitMutex.RUnlock()

	if cb, ok := m.circuitBreakers[provider]; ok {
		return cb.IsOpen()
	}
	return false
}

// AdvanceTime advances the mock time for testing (simulates time passage)
func (m *tokenizerManagerWithCircuitBreaker) AdvanceTime(duration time.Duration) {
	// Initialize current time if not set
	if m.currentTime.IsZero() {
		m.currentTime = time.Now()
	}

	m.currentTime = m.currentTime.Add(duration)
	m.usingMockTime = true

	// Update time source for all circuit breakers
	m.circuitMutex.Lock()
	defer m.circuitMutex.Unlock()

	currentTimeCopy := m.currentTime // Capture current time for closure
	for _, cb := range m.circuitBreakers {
		cb.SetTimeSource(func() time.Time { return currentTimeCopy })
	}
}

// getOrCreateCircuitBreaker gets or creates a circuit breaker for the provider
func (m *tokenizerManagerWithCircuitBreaker) getOrCreateCircuitBreaker(provider string) *CircuitBreaker {
	m.circuitMutex.RLock()
	if cb, ok := m.circuitBreakers[provider]; ok {
		m.circuitMutex.RUnlock()
		return cb
	}
	m.circuitMutex.RUnlock()

	m.circuitMutex.Lock()
	defer m.circuitMutex.Unlock()

	// Double-check pattern
	if cb, ok := m.circuitBreakers[provider]; ok {
		return cb
	}

	cb := NewCircuitBreaker()
	if m.usingMockTime {
		cb.SetTimeSource(m.timeSource)
	}
	m.circuitBreakers[provider] = cb
	return cb
}

// PerformanceMetrics tracks performance statistics for a tokenizer provider
type PerformanceMetrics struct {
	RequestCount int
	SuccessCount int
	FailureCount int
	AvgLatency   time.Duration
	SuccessRate  float64
	totalLatency time.Duration
}

// tokenizerManagerWithPerformanceMonitoring extends tokenizerManagerImpl with performance tracking
type tokenizerManagerWithPerformanceMonitoring struct {
	tokenizerManagerImpl
	metrics        map[string]*PerformanceMetrics
	metricsMutex   sync.RWMutex
	mockTokenizers map[string]AccurateTokenCounter // For testing
	mockMutex      sync.RWMutex
}

// NewTokenizerManagerWithPerformanceMonitoring creates a tokenizer manager with performance monitoring
func NewTokenizerManagerWithPerformanceMonitoring() *tokenizerManagerWithPerformanceMonitoring {
	return &tokenizerManagerWithPerformanceMonitoring{
		tokenizerManagerImpl: tokenizerManagerImpl{},
		metrics:              make(map[string]*PerformanceMetrics),
		mockTokenizers:       make(map[string]AccurateTokenCounter),
	}
}

// GetTokenizer wraps the original method with performance tracking
func (m *tokenizerManagerWithPerformanceMonitoring) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	// Check if we have a mock tokenizer for testing
	m.mockMutex.RLock()
	if mockTokenizer, ok := m.mockTokenizers[provider]; ok {
		m.mockMutex.RUnlock()
		return &performanceTrackingTokenizer{
			underlying: mockTokenizer,
			provider:   provider,
			manager:    m,
		}, nil
	}
	m.mockMutex.RUnlock()

	// Call the original implementation and wrap with performance tracking
	tokenizer, err := m.tokenizerManagerImpl.GetTokenizer(provider)
	if err != nil {
		return nil, err
	}

	return &performanceTrackingTokenizer{
		underlying: tokenizer,
		provider:   provider,
		manager:    m,
	}, nil
}

// SetMockTokenizer sets a mock tokenizer for testing
func (m *tokenizerManagerWithPerformanceMonitoring) SetMockTokenizer(provider string, tokenizer AccurateTokenCounter) {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()
	m.mockTokenizers[provider] = tokenizer
}

// GetMetrics returns performance metrics for a provider
func (m *tokenizerManagerWithPerformanceMonitoring) GetMetrics(provider string) *PerformanceMetrics {
	m.metricsMutex.RLock()
	defer m.metricsMutex.RUnlock()

	if metrics, ok := m.metrics[provider]; ok {
		// Return a copy to avoid race conditions
		return &PerformanceMetrics{
			RequestCount: metrics.RequestCount,
			SuccessCount: metrics.SuccessCount,
			FailureCount: metrics.FailureCount,
			AvgLatency:   metrics.AvgLatency,
			SuccessRate:  metrics.SuccessRate,
		}
	}

	return &PerformanceMetrics{}
}

// recordMetrics updates performance metrics for a provider
func (m *tokenizerManagerWithPerformanceMonitoring) recordMetrics(provider string, latency time.Duration, success bool) {
	m.metricsMutex.Lock()
	defer m.metricsMutex.Unlock()

	if m.metrics[provider] == nil {
		m.metrics[provider] = &PerformanceMetrics{}
	}

	metrics := m.metrics[provider]
	metrics.RequestCount++
	metrics.totalLatency += latency
	metrics.AvgLatency = metrics.totalLatency / time.Duration(metrics.RequestCount)

	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	if metrics.RequestCount > 0 {
		metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100.0
	}
}

// performanceTrackingTokenizer wraps an AccurateTokenCounter with performance tracking
type performanceTrackingTokenizer struct {
	underlying AccurateTokenCounter
	provider   string
	manager    *tokenizerManagerWithPerformanceMonitoring
}

// CountTokens tracks performance metrics while delegating to underlying tokenizer
func (p *performanceTrackingTokenizer) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	start := time.Now()
	tokens, err := p.underlying.CountTokens(ctx, text, modelName)
	latency := time.Since(start)

	p.manager.recordMetrics(p.provider, latency, err == nil)
	return tokens, err
}

// SupportsModel delegates to underlying tokenizer
func (p *performanceTrackingTokenizer) SupportsModel(modelName string) bool {
	return p.underlying.SupportsModel(modelName)
}

// GetEncoding delegates to underlying tokenizer
func (p *performanceTrackingTokenizer) GetEncoding(modelName string) (string, error) {
	return p.underlying.GetEncoding(modelName)
}

// tokenizerManagerWithTimeout extends tokenizerManagerImpl with timeout protection
type tokenizerManagerWithTimeout struct {
	tokenizerManagerImpl
	timeout        time.Duration
	mockTokenizers map[string]AccurateTokenCounter // For testing
	mockMutex      sync.RWMutex
}

// NewTokenizerManagerWithTimeout creates a tokenizer manager with timeout protection
func NewTokenizerManagerWithTimeout(timeout time.Duration) *tokenizerManagerWithTimeout {
	return &tokenizerManagerWithTimeout{
		tokenizerManagerImpl: tokenizerManagerImpl{},
		timeout:              timeout,
		mockTokenizers:       make(map[string]AccurateTokenCounter),
	}
}

// GetTokenizer wraps the original method with timeout protection
func (m *tokenizerManagerWithTimeout) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	// Check if we have a mock tokenizer for testing
	m.mockMutex.RLock()
	if mockTokenizer, ok := m.mockTokenizers[provider]; ok {
		m.mockMutex.RUnlock()
		return &timeoutTokenizer{
			underlying: mockTokenizer,
			timeout:    m.timeout,
		}, nil
	}
	m.mockMutex.RUnlock()

	// Call the original implementation and wrap with timeout protection
	tokenizer, err := m.tokenizerManagerImpl.GetTokenizer(provider)
	if err != nil {
		return nil, err
	}

	return &timeoutTokenizer{
		underlying: tokenizer,
		timeout:    m.timeout,
	}, nil
}

// SetMockTokenizer sets a mock tokenizer for testing
func (m *tokenizerManagerWithTimeout) SetMockTokenizer(provider string, tokenizer AccurateTokenCounter) {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()
	m.mockTokenizers[provider] = tokenizer
}

// timeoutTokenizer wraps an AccurateTokenCounter with timeout protection
type timeoutTokenizer struct {
	underlying AccurateTokenCounter
	timeout    time.Duration
}

// CountTokens adds timeout protection while delegating to underlying tokenizer
func (t *timeoutTokenizer) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	return t.underlying.CountTokens(ctx, text, modelName)
}

// SupportsModel delegates to underlying tokenizer
func (t *timeoutTokenizer) SupportsModel(modelName string) bool {
	return t.underlying.SupportsModel(modelName)
}

// GetEncoding delegates to underlying tokenizer
func (t *timeoutTokenizer) GetEncoding(modelName string) (string, error) {
	return t.underlying.GetEncoding(modelName)
}

// tokenizerManagerWithTimeoutAndCircuitBreaker combines timeout and circuit breaker functionality
type tokenizerManagerWithTimeoutAndCircuitBreaker struct {
	tokenizerManagerWithCircuitBreaker
	timeout      time.Duration
	metrics      map[string]*PerformanceMetrics
	metricsMutex sync.RWMutex
}

// NewTokenizerManagerWithTimeoutAndCircuitBreaker creates a manager with both timeout and circuit breaker
func NewTokenizerManagerWithTimeoutAndCircuitBreaker(timeout time.Duration) *tokenizerManagerWithTimeoutAndCircuitBreaker {
	return &tokenizerManagerWithTimeoutAndCircuitBreaker{
		tokenizerManagerWithCircuitBreaker: *NewTokenizerManagerWithCircuitBreaker(),
		timeout:                            timeout,
		metrics:                            make(map[string]*PerformanceMetrics),
	}
}

// GetTokenizer wraps the circuit breaker method with timeout protection
func (m *tokenizerManagerWithTimeoutAndCircuitBreaker) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	tokenizer, err := m.tokenizerManagerWithCircuitBreaker.GetTokenizer(provider)
	if err != nil {
		return nil, err
	}

	return &timeoutAndCircuitBreakerTokenizer{
		underlying:     tokenizer,
		timeout:        m.timeout,
		provider:       provider,
		circuitManager: &m.tokenizerManagerWithCircuitBreaker,
		metricsManager: m,
	}, nil
}

// GetMetrics returns performance metrics for a provider
func (m *tokenizerManagerWithTimeoutAndCircuitBreaker) GetMetrics(provider string) *PerformanceMetrics {
	m.metricsMutex.RLock()
	defer m.metricsMutex.RUnlock()

	if metrics, ok := m.metrics[provider]; ok {
		// Return a copy to avoid race conditions
		return &PerformanceMetrics{
			RequestCount: metrics.RequestCount,
			SuccessCount: metrics.SuccessCount,
			FailureCount: metrics.FailureCount,
			AvgLatency:   metrics.AvgLatency,
			SuccessRate:  metrics.SuccessRate,
		}
	}

	return &PerformanceMetrics{}
}

// recordMetrics updates performance metrics for a provider
func (m *tokenizerManagerWithTimeoutAndCircuitBreaker) recordMetrics(provider string, latency time.Duration, success bool) {
	m.metricsMutex.Lock()
	defer m.metricsMutex.Unlock()

	if m.metrics[provider] == nil {
		m.metrics[provider] = &PerformanceMetrics{}
	}

	metrics := m.metrics[provider]
	metrics.RequestCount++
	metrics.totalLatency += latency
	metrics.AvgLatency = metrics.totalLatency / time.Duration(metrics.RequestCount)

	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	if metrics.RequestCount > 0 {
		metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100.0
	}
}

// timeoutAndCircuitBreakerTokenizer combines timeout protection with circuit breaker integration
type timeoutAndCircuitBreakerTokenizer struct {
	underlying     AccurateTokenCounter
	timeout        time.Duration
	provider       string
	circuitManager *tokenizerManagerWithCircuitBreaker
	metricsManager *tokenizerManagerWithTimeoutAndCircuitBreaker
}

// CountTokens adds timeout protection and records timeout failures to circuit breaker
func (t *timeoutAndCircuitBreakerTokenizer) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	tokens, err := t.underlying.CountTokens(ctx, text, modelName)
	latency := time.Since(start)

	// Record metrics regardless of success/failure
	success := err == nil
	if t.metricsManager != nil {
		t.metricsManager.recordMetrics(t.provider, latency, success)
	}

	// If timeout occurred, record it as a circuit breaker failure
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		if cb := t.circuitManager.getOrCreateCircuitBreaker(t.provider); cb != nil {
			cb.RecordFailure()
		}
		return 0, NewTokenizerErrorWithDetails(t.provider, modelName, "timeout", ctx.Err(), getTokenizerType(t.provider))
	}

	return tokens, err
}

// SupportsModel delegates to underlying tokenizer
func (t *timeoutAndCircuitBreakerTokenizer) SupportsModel(modelName string) bool {
	return t.underlying.SupportsModel(modelName)
}

// GetEncoding delegates to underlying tokenizer
func (t *timeoutAndCircuitBreakerTokenizer) GetEncoding(modelName string) (string, error) {
	return t.underlying.GetEncoding(modelName)
}
