// Package llm provides tests for the WrapWithCorrelationID function
// Following TDD principles to verify the simple correlation ID integration
package llm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrapWithCorrelationID tests the simple correlation ID wrapping function
func TestWrapWithCorrelationID(t *testing.T) {
	t.Parallel()
	t.Run("WrapWithCorrelationID with new error", func(t *testing.T) {
		// Test wrapping a plain error with correlation ID
		originalErr := errors.New("test error message")
		correlationID := "test-correlation-123"

		wrappedErr := WrapWithCorrelationID(originalErr, "test-provider", "Test message", CategoryAuth, correlationID)

		// Verify LLMError creation
		require.NotNil(t, wrappedErr)
		assert.Equal(t, "test-provider", wrappedErr.Provider)
		assert.Equal(t, "Test message", wrappedErr.Message)
		assert.Equal(t, CategoryAuth, wrappedErr.ErrorCategory)
		assert.Equal(t, originalErr, wrappedErr.Original)
		assert.Equal(t, correlationID, wrappedErr.RequestID) // Correlation ID stored in RequestID
	})

	t.Run("WrapWithCorrelationID with existing LLMError", func(t *testing.T) {
		// Test wrapping an existing LLMError (should update it)
		existingErr := &LLMError{
			Provider:      "old-provider",
			Message:       "old message",
			ErrorCategory: CategoryUnknown,
			RequestID:     "old-correlation",
		}

		newCorrelationID := "new-correlation-456"
		wrappedErr := WrapWithCorrelationID(existingErr, "new-provider", "Updated message", CategoryRateLimit, newCorrelationID)

		// Should get back the same LLMError but updated
		require.NotNil(t, wrappedErr)
		assert.Equal(t, "new-provider", wrappedErr.Provider)
		assert.Equal(t, "Updated message", wrappedErr.Message)
		assert.Equal(t, CategoryRateLimit, wrappedErr.ErrorCategory)
		assert.Equal(t, newCorrelationID, wrappedErr.RequestID)
	})

	t.Run("WrapWithCorrelationID with empty correlation ID", func(t *testing.T) {
		// Test that empty correlation ID doesn't override existing RequestID
		originalErr := errors.New("test error")

		wrappedErr := WrapWithCorrelationID(originalErr, "provider", "message", CategoryNetwork, "")

		require.NotNil(t, wrappedErr)
		assert.Equal(t, "", wrappedErr.RequestID) // Should remain empty
	})

	t.Run("WrapWithCorrelationID with nil error", func(t *testing.T) {
		// Test that nil error returns nil
		wrappedErr := WrapWithCorrelationID(nil, "provider", "message", CategoryAuth, "correlation")
		assert.Nil(t, wrappedErr)
	})

	t.Run("ExtractCorrelationID works with WrapWithCorrelationID result", func(t *testing.T) {
		// Test that ExtractCorrelationID can extract correlation ID from wrapped error
		originalErr := errors.New("test error")
		correlationID := "extraction-test-789"

		wrappedErr := WrapWithCorrelationID(originalErr, "provider", "message", CategoryServer, correlationID)

		extractedID := ExtractCorrelationID(wrappedErr)
		assert.Equal(t, correlationID, extractedID)
	})
}

// TestCorrelationIDIntegrationWithExistingSystem tests integration with existing error handling
func TestCorrelationIDIntegrationWithExistingSystem(t *testing.T) {
	t.Parallel()
	t.Run("correlation ID preserves through error chains", func(t *testing.T) {
		// Create an error with correlation ID
		originalErr := errors.New("API failure")
		correlationID := "chain-test-101112"

		// Wrap with correlation ID
		apiErr := WrapWithCorrelationID(originalErr, "openai", "API call failed", CategoryRateLimit, correlationID)

		// Further wrap using standard Wrap function
		orchestratorErr := Wrap(apiErr, "orchestrator", "Processing failed", CategoryRateLimit)

		// Correlation ID should still be extractable
		extractedID := ExtractCorrelationID(orchestratorErr)
		assert.Equal(t, correlationID, extractedID)

		// Verify the error is still categorized correctly
		assert.True(t, IsCategory(orchestratorErr, CategoryRateLimit))

		// Verify the original error is still accessible
		var originalLLMErr *LLMError
		require.True(t, errors.As(orchestratorErr, &originalLLMErr))
		assert.Equal(t, correlationID, originalLLMErr.RequestID)
	})

	t.Run("correlation ID works with error categorization", func(t *testing.T) {
		// Test that correlation ID doesn't interfere with error categorization
		testCases := []struct {
			category     ErrorCategory
			testFunction func(error) bool
		}{
			{CategoryAuth, IsAuth},
			{CategoryRateLimit, IsRateLimit},
			{CategoryNetwork, IsNetwork},
			{CategoryServer, IsServer},
		}

		for _, tc := range testCases {
			err := WrapWithCorrelationID(errors.New("test"), "provider", "message", tc.category, "test-corr")
			assert.True(t, tc.testFunction(err), "Category %v should be detectable", tc.category)
		}
	})

	t.Run("correlation ID in complex error chains", func(t *testing.T) {
		// Test correlation ID in more complex error scenarios
		originalErr := errors.New("network timeout")
		correlationID := "complex-chain-131415"

		// Layer 1: API client error with correlation
		apiErr := WrapWithCorrelationID(originalErr, "gemini", "API timeout", CategoryNetwork, correlationID)

		// Layer 2: Model processor error (standard wrap)
		modelErr := Wrap(apiErr, "model-processor", "Model processing timeout", CategoryNetwork)

		// Layer 3: Orchestrator error (standard wrap)
		orchestratorErr := Wrap(modelErr, "orchestrator", "Orchestration timeout", CategoryNetwork)

		// Verify correlation ID propagates through the entire chain
		extractedID := ExtractCorrelationID(orchestratorErr)
		assert.Equal(t, correlationID, extractedID)

		// Verify error information is preserved
		errorMsg := orchestratorErr.Error()
		assert.Contains(t, errorMsg, "Orchestration timeout")
		assert.Contains(t, errorMsg, "network timeout") // Original error message should be included
	})
}
