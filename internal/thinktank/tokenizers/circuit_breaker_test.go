package tokenizers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenizerManagerWithCircuitBreaker_TracksFailures tests that the tokenizer manager
// integrates with circuit breaker to track consecutive failures per provider
func TestTokenizerManagerWithCircuitBreaker_TracksFailures(t *testing.T) {
	t.Parallel()

	// Create a failing tokenizer that always returns errors
	failingTokenizer := &MockAccurateTokenCounterWithFailure{
		CountTokensError: errors.New("tiktoken encoding failed"),
	}

	// Create manager with circuit breaker integration - this will fail (RED phase)
	manager := NewTokenizerManagerWithCircuitBreaker()

	// Mock the failing tokenizer for openai provider - this will fail (RED phase)
	manager.SetMockTokenizer("openai", failingTokenizer)

	// First few failures should keep circuit closed
	for i := 0; i < 4; i++ {
		_, err := manager.GetTokenizer("openai")
		assert.Error(t, err, "Should get tokenizer initialization error")
		assert.False(t, manager.IsCircuitOpen("openai"), "Circuit should remain closed for first 4 failures")
	}

	// Fifth failure should open the circuit (default threshold is 5)
	_, err := manager.GetTokenizer("openai")
	assert.Error(t, err, "Should get tokenizer initialization error")
	assert.True(t, manager.IsCircuitOpen("openai"), "Circuit should be open after 5 consecutive failures")

	// Subsequent attempts should fail fast due to open circuit
	start := time.Now()
	_, err = manager.GetTokenizer("openai")
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker open", "Error should indicate circuit breaker is open")
	assert.Less(t, duration, 10*time.Millisecond, "Should fail fast when circuit is open")
}

// TestTokenizerManagerWithCircuitBreaker_RecoveryAfterSuccess tests that
// the circuit breaker recovers after successful operations
func TestTokenizerManagerWithCircuitBreaker_RecoveryAfterSuccess(t *testing.T) {
	t.Parallel()

	// Create a tokenizer that initially fails, then succeeds
	switchingTokenizer := &MockAccurateTokenCounterWithSwitching{
		FailureCount:     5, // Will fail first 5 times, then succeed
		CountTokensError: errors.New("temporary tiktoken failure"),
	}

	manager := NewTokenizerManagerWithCircuitBreaker()
	manager.SetMockTokenizer("openai", switchingTokenizer)

	// Cause circuit to open with 5 failures
	for i := 0; i < 5; i++ {
		_, err := manager.GetTokenizer("openai")
		assert.Error(t, err)
	}
	assert.True(t, manager.IsCircuitOpen("openai"), "Circuit should be open after failures")

	// Wait for cooldown period (mocked time advancement)
	manager.AdvanceTime(35 * time.Second) // Beyond 30 second cooldown

	// Now the tokenizer will succeed - circuit should close
	tokenizer, err := manager.GetTokenizer("openai")
	require.NoError(t, err, "Should succeed after cooldown and recovery")
	assert.NotNil(t, tokenizer)
	assert.False(t, manager.IsCircuitOpen("openai"), "Circuit should be closed after successful recovery")
}

// TestTokenizerManagerWithCircuitBreaker_ProviderIsolation tests that
// circuit breaker failures in one provider don't affect other providers
func TestTokenizerManagerWithCircuitBreaker_ProviderIsolation(t *testing.T) {
	t.Parallel()

	failingTokenizer := &MockAccurateTokenCounterWithFailure{
		CountTokensError: errors.New("openai tiktoken failure"),
	}
	workingTokenizer := &MockAccurateTokenCounterSuccess{}

	manager := NewTokenizerManagerWithCircuitBreaker()
	manager.SetMockTokenizer("openai", failingTokenizer)
	manager.SetMockTokenizer("gemini", workingTokenizer)

	// Fail openai provider 5 times to open its circuit
	for i := 0; i < 5; i++ {
		_, err := manager.GetTokenizer("openai")
		assert.Error(t, err)
	}
	assert.True(t, manager.IsCircuitOpen("openai"), "OpenAI circuit should be open")

	// Gemini should still work normally
	tokenizer, err := manager.GetTokenizer("gemini")
	require.NoError(t, err, "Gemini should work despite OpenAI failure")
	assert.NotNil(t, tokenizer)
	assert.False(t, manager.IsCircuitOpen("gemini"), "Gemini circuit should remain closed")
}

// Mock implementations for testing

type MockAccurateTokenCounterWithFailure struct {
	CountTokensError error
}

func (m *MockAccurateTokenCounterWithFailure) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	return 0, m.CountTokensError
}

func (m *MockAccurateTokenCounterWithFailure) SupportsModel(modelName string) bool {
	return true
}

func (m *MockAccurateTokenCounterWithFailure) GetEncoding(modelName string) (string, error) {
	return "test-encoding", nil
}

type MockAccurateTokenCounterWithSwitching struct {
	FailureCount     int
	CallCount        int
	CountTokensError error
}

func (m *MockAccurateTokenCounterWithSwitching) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	m.CallCount++
	if m.CallCount <= m.FailureCount {
		return 0, m.CountTokensError
	}
	return len(text), nil
}

func (m *MockAccurateTokenCounterWithSwitching) SupportsModel(modelName string) bool {
	return true
}

func (m *MockAccurateTokenCounterWithSwitching) GetEncoding(modelName string) (string, error) {
	return "test-encoding", nil
}

type MockAccurateTokenCounterSuccess struct{}

func (m *MockAccurateTokenCounterSuccess) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	return len(text), nil
}

func (m *MockAccurateTokenCounterSuccess) SupportsModel(modelName string) bool {
	return true
}

func (m *MockAccurateTokenCounterSuccess) GetEncoding(modelName string) (string, error) {
	return "test-encoding", nil
}
