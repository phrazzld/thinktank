package tokenizers

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenizerManagerWithPerformanceMonitoring_TracksLatency tests that the tokenizer manager
// tracks latency and request count for performance monitoring
func TestTokenizerManagerWithPerformanceMonitoring_TracksLatency(t *testing.T) {
	t.Parallel()

	// Create a slow tokenizer that takes 50ms
	slowTokenizer := &MockSlowTokenCounter{
		LatencyMs: 50,
	}

	// Create manager with performance monitoring - this will fail (RED phase)
	manager := NewTokenizerManagerWithPerformanceMonitoring()
	manager.SetMockTokenizer("openai", slowTokenizer)

	ctx := context.Background()
	tokenizer, err := manager.GetTokenizer("openai")
	require.NoError(t, err)

	// Call tokenizer multiple times to track metrics
	for i := 0; i < 3; i++ {
		_, err := tokenizer.CountTokens(ctx, "test text", "gpt-4")
		require.NoError(t, err)
	}

	// Get performance metrics - this will fail (RED phase)
	metrics := manager.GetMetrics("openai")

	assert.Equal(t, 3, metrics.RequestCount, "Should track request count")
	assert.GreaterOrEqual(t, metrics.AvgLatency, 45*time.Millisecond, "Should track latency")
	// Use CI-aware threshold - more lenient in CI environments
	expectedMaxLatency := 60 * time.Millisecond
	if os.Getenv("CI") != "" {
		expectedMaxLatency = 100 * time.Millisecond // More lenient for CI
	}
	assert.LessOrEqual(t, metrics.AvgLatency, expectedMaxLatency, "Latency should be in expected range")
	assert.Equal(t, 3, metrics.SuccessCount, "Should track success count")
	assert.Equal(t, 0, metrics.FailureCount, "Should track failure count")
}

// TestTokenizerManagerWithPerformanceMonitoring_TracksFailures tests that the manager
// correctly tracks both successes and failures
func TestTokenizerManagerWithPerformanceMonitoring_TracksFailures(t *testing.T) {
	t.Parallel()

	// Create a tokenizer that fails every other call
	intermittentTokenizer := &MockIntermittentTokenCounter{
		FailurePattern: []bool{false, true, false, true, false}, // success, fail, success, fail, success
	}

	manager := NewTokenizerManagerWithPerformanceMonitoring()
	manager.SetMockTokenizer("gemini", intermittentTokenizer)

	ctx := context.Background()
	tokenizer, err := manager.GetTokenizer("gemini")
	require.NoError(t, err)

	// Make 5 calls with alternating success/failure pattern
	for i := 0; i < 5; i++ {
		_, _ = tokenizer.CountTokens(ctx, "test text", "gemini-3-flash") // Ignore errors
	}

	metrics := manager.GetMetrics("gemini")

	assert.Equal(t, 5, metrics.RequestCount, "Should track all requests")
	assert.Equal(t, 3, metrics.SuccessCount, "Should track successes")
	assert.Equal(t, 2, metrics.FailureCount, "Should track failures")
	assert.InDelta(t, 60.0, metrics.SuccessRate, 5.0, "Success rate should be 60% (3/5)")
}

// TestTokenizerManagerWithTimeout_RespectsContextTimeout tests that tokenizer operations
// respect context timeouts and fail fast
func TestTokenizerManagerWithTimeout_RespectsContextTimeout(t *testing.T) {
	t.Parallel()

	// Create a very slow tokenizer that takes 200ms
	verySlowTokenizer := &MockSlowTokenCounter{
		LatencyMs: 200,
	}

	// Create manager with 100ms timeout - this will fail (RED phase)
	manager := NewTokenizerManagerWithTimeout(100 * time.Millisecond)
	manager.SetMockTokenizer("openai", verySlowTokenizer)

	ctx := context.Background()
	tokenizer, err := manager.GetTokenizer("openai")
	require.NoError(t, err)

	start := time.Now()
	_, err = tokenizer.CountTokens(ctx, "test text", "gpt-4")
	duration := time.Since(start)

	assert.Error(t, err, "Should timeout and return error")
	assert.True(t,
		err.Error() == "context deadline exceeded" ||
			strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "deadline"),
		"Error should indicate timeout or deadline exceeded, got: %s", err.Error())
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond, "Should wait at least timeout duration")
	assert.LessOrEqual(t, duration, 150*time.Millisecond, "Should not wait much longer than timeout")
}

// TestTokenizerManagerWithTimeout_RecordsTimeoutAsFailure tests that timeouts
// are properly recorded as failures in metrics and circuit breaker
func TestTokenizerManagerWithTimeout_RecordsTimeoutAsFailure(t *testing.T) {
	t.Parallel()

	verySlowTokenizer := &MockSlowTokenCounter{
		LatencyMs: 200,
	}

	// Create manager with both timeout and circuit breaker - this will fail (RED phase)
	manager := NewTokenizerManagerWithTimeoutAndCircuitBreaker(50 * time.Millisecond)
	manager.SetMockTokenizer("openai", verySlowTokenizer)

	ctx := context.Background()
	tokenizer, err := manager.GetTokenizer("openai")
	require.NoError(t, err)

	// Cause multiple timeouts to trigger circuit breaker
	for i := 0; i < 6; i++ {
		_, _ = tokenizer.CountTokens(ctx, "test text", "gpt-4") // Ignore timeout errors
	}

	// Circuit breaker should be open due to repeated timeouts
	assert.True(t, manager.IsCircuitOpen("openai"), "Circuit breaker should open after repeated timeouts")

	// Metrics should track timeout failures
	metrics := manager.GetMetrics("openai")
	assert.Equal(t, 6, metrics.RequestCount, "Should track all timeout attempts")
	assert.Equal(t, 0, metrics.SuccessCount, "Should have no successes")
	assert.Equal(t, 6, metrics.FailureCount, "Should track timeouts as failures")
}

// Mock implementations for performance testing

type MockSlowTokenCounter struct {
	LatencyMs int
}

func (m *MockSlowTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	// Handle nil context (used by circuit breaker initialization test)
	if ctx == nil {
		time.Sleep(time.Duration(m.LatencyMs) * time.Millisecond)
		return len(text), nil
	}

	// Simulate the latency
	select {
	case <-time.After(time.Duration(m.LatencyMs) * time.Millisecond):
		return len(text), nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (m *MockSlowTokenCounter) SupportsModel(modelName string) bool {
	return true
}

func (m *MockSlowTokenCounter) GetEncoding(modelName string) (string, error) {
	return "test-encoding", nil
}

type MockIntermittentTokenCounter struct {
	FailurePattern []bool
	callCount      int
}

func (m *MockIntermittentTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	defer func() { m.callCount++ }()

	if m.callCount < len(m.FailurePattern) && m.FailurePattern[m.callCount] {
		return 0, errors.New("intermittent failure")
	}
	return len(text), nil
}

func (m *MockIntermittentTokenCounter) SupportsModel(modelName string) bool {
	return true
}

func (m *MockIntermittentTokenCounter) GetEncoding(modelName string) (string, error) {
	return "test-encoding", nil
}
