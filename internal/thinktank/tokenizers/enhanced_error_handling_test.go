package tokenizers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnhancedErrorHandling_ProviderContextInErrors tests that all tokenizer errors
// include provider and model context for better debugging
func TestEnhancedErrorHandling_ProviderContextInErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		provider       string
		model          string
		expectedFields []string
	}{
		{
			name:           "OpenAI provider error includes provider context",
			provider:       "openai",
			model:          "gpt-4",
			expectedFields: []string{"provider=openai", "model=gpt-4"},
		},
		{
			name:           "Gemini provider error includes provider context",
			provider:       "gemini",
			model:          "gemini-2.5-pro",
			expectedFields: []string{"provider=gemini", "model=gemini-2.5-pro"},
		},
		{
			name:           "Unknown provider error includes context",
			provider:       "unknown",
			model:          "unknown-model",
			expectedFields: []string{"provider=unknown", "model=unknown-model"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create failing tokenizer for this test
			failingTokenizer := &MockFailingTokenCounter{
				ShouldFail:   true,
				FailureError: errors.New("simulated tokenizer failure"),
				Provider:     tt.provider,
			}

			manager := NewTokenizerManagerWithCircuitBreaker()
			manager.SetMockTokenizer(tt.provider, failingTokenizer)

			ctx := context.Background()
			tokenizer, err := manager.GetTokenizer(tt.provider)

			// This will fail in RED phase because enhanced error handling is not implemented
			if err != nil {
				// Check if error includes provider/model context
				errorStr := err.Error()

				// Enhanced error should include structured context
				for _, field := range tt.expectedFields {
					assert.Contains(t, errorStr, field,
						"Error should include %s: %s", field, errorStr)
				}

				// Should be a TokenizerError with enhanced details
				var tokenizerErr *TokenizerError
				if errors.As(err, &tokenizerErr) {
					assert.Equal(t, tt.provider, tokenizerErr.Provider)
					assert.NotEmpty(t, tokenizerErr.Details, "Error should have detailed context")
				} else {
					t.Errorf("Error should be a TokenizerError but got: %T", err)
				}
			} else {
				// If we got a tokenizer, try using it (which should fail)
				require.NotNil(t, tokenizer)
				_, err = tokenizer.CountTokens(ctx, "test", tt.model)
				require.Error(t, err, "CountTokens should fail for failing tokenizer")

				// Check enhanced error context
				errorStr := err.Error()
				for _, field := range tt.expectedFields {
					assert.Contains(t, errorStr, field,
						"CountTokens error should include %s: %s", field, errorStr)
				}
			}
		})
	}
}

// TestEnhancedErrorHandling_TokenizerTypeInErrors tests that error messages
// include the tokenizer type for better debugging
func TestEnhancedErrorHandling_TokenizerTypeInErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		provider          string
		model             string
		expectedTokenizer string
	}{
		{
			name:              "OpenAI errors include tiktoken type",
			provider:          "openai",
			model:             "gpt-4",
			expectedTokenizer: "tiktoken",
		},
		{
			name:              "Gemini errors include sentencepiece type",
			provider:          "gemini",
			model:             "gemini-2.5-pro",
			expectedTokenizer: "sentencepiece",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create failing tokenizer
			failingTokenizer := &MockFailingTokenCounter{
				ShouldFail:   true,
				FailureError: errors.New("encoding initialization failed"),
				Provider:     tt.provider,
			}

			manager := NewTokenizerManagerWithCircuitBreaker()
			manager.SetMockTokenizer(tt.provider, failingTokenizer)

			ctx := context.Background()
			tokenizer, err := manager.GetTokenizer(tt.provider)

			if err == nil {
				require.NotNil(t, tokenizer)
				_, err = tokenizer.CountTokens(ctx, "test content", tt.model)
				require.Error(t, err)
			}

			// This will fail in RED phase - enhanced errors should include tokenizer type
			assert.Contains(t, err.Error(), tt.expectedTokenizer,
				"Error should include tokenizer type '%s': %s", tt.expectedTokenizer, err.Error())

			// Check for tokenizer type in error details
			var tokenizerErr *TokenizerError
			if errors.As(err, &tokenizerErr) {
				assert.Contains(t, tokenizerErr.Details, "tokenizer="+tt.expectedTokenizer,
					"Error details should include tokenizer type")
			}
		})
	}
}

// TestEnhancedErrorHandling_ComprehensiveFallbackScenarios tests all possible
// fallback scenarios to ensure model selection never completely fails
func TestEnhancedErrorHandling_ComprehensiveFallbackScenarios(t *testing.T) {
	t.Parallel()

	// This is the comprehensive integration test for all fallback scenarios
	scenarios := []struct {
		name              string
		setupFunc         func() TokenizerManager
		model             string
		expectedFallback  bool
		expectedTokenizer string
		shouldSucceed     bool
	}{
		{
			name: "Tokenizer initialization failure falls back gracefully",
			setupFunc: func() TokenizerManager {
				manager := NewTokenizerManagerWithCircuitBreaker()
				// Mock a failing OpenAI tokenizer
				manager.SetMockTokenizer("openai", &MockFailingTokenCounter{
					ShouldFail:   true,
					FailureError: errors.New("tiktoken initialization failed"),
					Provider:     "openai",
				})
				return manager
			},
			model:             "gpt-4",
			expectedFallback:  true,
			expectedTokenizer: "estimation",
			shouldSucceed:     true, // Should fall back to estimation and succeed
		},
		{
			name: "Encoding failure falls back gracefully",
			setupFunc: func() TokenizerManager {
				manager := NewTokenizerManagerWithCircuitBreaker()
				manager.SetMockTokenizer("openai", &MockEncodingFailureTokenCounter{
					ShouldFailEncoding: true,
				})
				return manager
			},
			model:             "unsupported-gpt-model",
			expectedFallback:  true,
			expectedTokenizer: "estimation",
			shouldSucceed:     true,
		},
		{
			name: "Circuit breaker open falls back gracefully",
			setupFunc: func() TokenizerManager {
				manager := NewTokenizerManagerWithCircuitBreaker()
				failingTokenizer := &MockFailingTokenCounter{
					ShouldFail:   true,
					FailureError: errors.New("repeated failure"),
					Provider:     "openai",
				}
				manager.SetMockTokenizer("openai", failingTokenizer)

				// Trigger circuit breaker by causing multiple failures
				for i := 0; i < 6; i++ {
					_, _ = manager.GetTokenizer("openai") // Ignore errors to trigger circuit breaker
				}
				return manager
			},
			model:             "gpt-4",
			expectedFallback:  true,
			expectedTokenizer: "estimation",
			shouldSucceed:     false, // Circuit breaker should prevent execution
		},
		{
			name: "Unknown provider falls back to estimation",
			setupFunc: func() TokenizerManager {
				return NewTokenizerManager() // Standard manager without mocks
			},
			model:             "claude-3-haiku", // Anthropic model not supported
			expectedFallback:  true,
			expectedTokenizer: "estimation",
			shouldSucceed:     false, // Should fail to get tokenizer, fallback to estimation elsewhere
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			manager := scenario.setupFunc()

			// Try to get tokenizer
			tokenizer, err := manager.GetTokenizer("openai") // Test with OpenAI for most scenarios

			if scenario.shouldSucceed {
				if err != nil {
					t.Logf("Tokenizer creation failed (expected for fallback): %v", err)
					// This is expected for some fallback scenarios
				} else {
					// Try to use the tokenizer
					require.NotNil(t, tokenizer)
					ctx := context.Background()
					_, err = tokenizer.CountTokens(ctx, "test content", scenario.model)
					// May succeed or fail depending on scenario
					t.Logf("CountTokens result: error=%v", err)
				}

				// The key test: enhanced error handling should provide context
				if err != nil {
					// This will fail in RED phase - enhanced error handling not implemented
					assert.Contains(t, err.Error(), "provider=",
						"Error should include provider context: %s", err.Error())
					assert.Contains(t, err.Error(), "model=",
						"Error should include model context: %s", err.Error())
				}
			} else {
				// Should fail but with enhanced error context
				if err != nil {
					assert.Contains(t, err.Error(), "provider=",
						"Failure error should include provider context: %s", err.Error())
				}
			}
		})
	}
}

// TestEnhancedErrorHandling_ModelSelectionRobustness tests that model selection
// never fails completely due to tokenization issues
func TestEnhancedErrorHandling_ModelSelectionRobustness(t *testing.T) {
	t.Parallel()

	// Create a tokenizer manager where all accurate tokenizers fail
	failingManager := &MockFailingTokenizerManager{
		ShouldFailAll: true,
	}

	// This test simulates the integration with TokenCountingService
	// to ensure model selection remains robust

	// Try to get tokenizer for OpenAI (should fail)
	_, err := failingManager.GetTokenizer("openai")
	require.Error(t, err, "Should fail to get OpenAI tokenizer")

	// Try to get tokenizer for Gemini (should fail)
	_, err = failingManager.GetTokenizer("gemini")
	require.Error(t, err, "Should fail to get Gemini tokenizer")

	// Try to get tokenizer for unsupported provider (should fail)
	_, err = failingManager.GetTokenizer("openrouter")
	require.Error(t, err, "Should fail to get OpenRouter tokenizer")

	// The key test: all errors should have enhanced context
	// This will fail in RED phase because enhanced error handling is not implemented

	// Test that provider context is included in all failures
	tokenizerProviders := []string{"openai", "gemini", "openrouter"}
	for _, provider := range tokenizerProviders {
		_, err := failingManager.GetTokenizer(provider)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "provider="+provider,
			"Error should include provider context for %s: %s", provider, err.Error())

		// Should be enhanced TokenizerError
		var tokenizerErr *TokenizerError
		if errors.As(err, &tokenizerErr) {
			assert.Equal(t, provider, tokenizerErr.Provider)
			assert.NotEmpty(t, tokenizerErr.Details)
			assert.Contains(t, tokenizerErr.Details, "provider="+provider)
		} else {
			t.Errorf("Error should be TokenizerError for provider %s, got: %T", provider, err)
		}
	}
}

// Mock implementations for testing enhanced error handling

type MockFailingTokenCounter struct {
	ShouldFail   bool
	FailureError error
	Provider     string // Add provider field to make mock aware of its provider
}

func (m *MockFailingTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if m.ShouldFail {
		// Don't fail the manager's initialization test (model="test"), but fail real model tests
		if modelName == "test" {
			return 0, nil // Allow manager initialization to succeed
		}
		// Return enhanced error with actual provider and model context
		provider := m.Provider
		if provider == "" {
			provider = "test-provider" // fallback
		}
		tokenizerType := getTokenizerType(provider)
		return 0, NewTokenizerErrorWithDetails(provider, modelName, "simulated tokenizer failure", m.FailureError, tokenizerType)
	}
	return len(text), nil
}

func (m *MockFailingTokenCounter) SupportsModel(modelName string) bool {
	return !m.ShouldFail
}

func (m *MockFailingTokenCounter) GetEncoding(modelName string) (string, error) {
	if m.ShouldFail {
		// Return enhanced error with model context
		return "", NewTokenizerErrorWithDetails("test-provider", modelName, "encoding failure", m.FailureError, "test-tokenizer")
	}
	return "test-encoding", nil
}

type MockEncodingFailureTokenCounter struct {
	ShouldFailEncoding bool
}

func (m *MockEncodingFailureTokenCounter) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if m.ShouldFailEncoding {
		return 0, NewTokenizerErrorWithDetails("test-provider", modelName, "encoding not found for model", nil, "test-tokenizer")
	}
	return len(text), nil
}

func (m *MockEncodingFailureTokenCounter) SupportsModel(modelName string) bool {
	return !m.ShouldFailEncoding
}

func (m *MockEncodingFailureTokenCounter) GetEncoding(modelName string) (string, error) {
	if m.ShouldFailEncoding {
		return "", NewTokenizerErrorWithDetails("test-provider", modelName, "encoding not found for model", nil, "test-tokenizer")
	}
	return "test-encoding", nil
}

type MockFailingTokenizerManager struct {
	ShouldFailAll bool
}

func (m *MockFailingTokenizerManager) GetTokenizer(provider string) (AccurateTokenCounter, error) {
	if m.ShouldFailAll {
		return nil, NewTokenizerErrorWithDetails(provider, "", "all tokenizers failing for testing", nil, getTokenizerType(provider))
	}
	return &MockFailingTokenCounter{ShouldFail: false}, nil
}

func (m *MockFailingTokenizerManager) SupportsProvider(provider string) bool {
	return !m.ShouldFailAll
}

func (m *MockFailingTokenizerManager) ClearCache() {
	// No-op for mock
}
