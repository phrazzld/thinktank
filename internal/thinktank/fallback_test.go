package thinktank

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/thinktank/tokenizers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenCountingService_LogsFallbackEvents tests that all fallback scenarios are properly logged
// This is the real gap - current fallbacks work but aren't logged with structured messages
func TestTokenCountingService_LogsFallbackEvents(t *testing.T) {
	t.Parallel()

	// Create a mock logger to capture fallback events
	mockLogger := NewMockLogger()

	// Create a service with failing tokenizer to force fallback
	mockTokenizer := &MockAccurateTokenCounterWithFailure{
		CountTokensError: errors.New("tiktoken encoding failed"),
	}

	mockManager := &MockTokenizerManagerSuccess{
		TokenizerInstance: mockTokenizer,
	}

	// Create a service that combines the logger and failing manager
	service := &tokenCountingServiceImpl{
		tokenizerManager: mockManager,
		logger:           mockLogger,
	}

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions:        "Analyze this code for performance issues",
		Files:               []FileContent{{Path: "test.go", Content: "package main\nfunc main() {}"}},
		SafetyMarginPercent: 20,
	}

	// This should trigger the tokenizer failure and subsequent fallback
	result, err := service.CountTokensForModel(ctx, req, "gpt-4.1")

	require.NoError(t, err, "Service should handle tokenization failure gracefully")
	assert.Equal(t, "estimation", result.TokenizerUsed, "Should fallback to estimation")

	// This assertion should FAIL initially - we don't currently log fallback events
	fallbackLogged := len(mockLogger.warnMessages) > 0
	for _, msg := range mockLogger.warnMessages {
		if strings.Contains(msg, "fallback") || strings.Contains(msg, "estimation") {
			fallbackLogged = true
			break
		}
	}

	// This should drive us to implement structured fallback logging
	assert.True(t, fallbackLogged, "Should log fallback event when tokenization fails")
}

// TestTokenCountingService_FallbackOnSpecificTokenizationFailure tests when tokenizer
// initializes successfully but fails during actual tokenization
func TestTokenCountingService_FallbackOnSpecificTokenizationFailure(t *testing.T) {
	t.Parallel()

	// Create a tokenizer that fails during CountTokens()
	mockTokenizer := &MockAccurateTokenCounterWithFailure{
		CountTokensError: errors.New("tokenization failed due to malformed input"),
	}

	mockManager := &MockTokenizerManagerSuccess{
		TokenizerInstance: mockTokenizer,
	}

	service := NewTokenCountingServiceWithManager(mockManager)
	ctx := context.Background()

	req := TokenCountingRequest{
		Instructions:        "Process this malformed input",
		Files:               []FileContent{{Path: "bad.txt", Content: "malformed content"}},
		SafetyMarginPercent: 20,
	}

	result, err := service.CountTokensForModel(ctx, req, "gpt-4.1")

	require.NoError(t, err, "Service should handle tokenization failure gracefully")
	assert.Equal(t, "estimation", result.TokenizerUsed, "Should fallback to estimation")
	assert.False(t, result.IsAccurate, "Result should be marked as not accurate")
	assert.Greater(t, result.TotalTokens, 0, "Should return estimation-based token count")
}

// TestTokenCountingService_FallbackWithEmptyContext tests fallback when context gathering fails
func TestTokenCountingService_FallbackWithEmptyContext(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	ctx := context.Background()

	// Test with completely empty context
	req := TokenCountingRequest{
		Instructions:        "",
		Files:               []FileContent{},
		SafetyMarginPercent: 20,
	}

	result, err := service.CountTokensForModel(ctx, req, "gpt-4.1")

	require.NoError(t, err, "Service should handle empty context gracefully")
	assert.Equal(t, 0, result.TotalTokens, "Empty context should result in 0 tokens")
	assert.Equal(t, 0, result.InstructionTokens, "Empty instructions should result in 0 tokens")
	assert.Equal(t, 0, result.FileTokens, "Empty files should result in 0 tokens")
}

// Mock implementations for testing

type MockTokenizerManagerWithFailure struct {
	GetTokenizerError error
}

func (m *MockTokenizerManagerWithFailure) SupportsProvider(provider string) bool {
	return true // Claim to support provider but fail during GetTokenizer
}

func (m *MockTokenizerManagerWithFailure) GetTokenizer(provider string) (tokenizers.AccurateTokenCounter, error) {
	return nil, m.GetTokenizerError
}

func (m *MockTokenizerManagerWithFailure) ClearCache() {
	// No-op for mock
}

type MockTokenizerManagerSuccess struct {
	TokenizerInstance tokenizers.AccurateTokenCounter
}

func (m *MockTokenizerManagerSuccess) SupportsProvider(provider string) bool {
	return true
}

func (m *MockTokenizerManagerSuccess) GetTokenizer(provider string) (tokenizers.AccurateTokenCounter, error) {
	return m.TokenizerInstance, nil
}

func (m *MockTokenizerManagerSuccess) ClearCache() {
	// No-op for mock
}

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

// Use existing MockLogger from app_test_mocks.go
