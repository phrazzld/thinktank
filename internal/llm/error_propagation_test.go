// Package llm provides comprehensive tests for contextual error wrapping
// These tests verify error context propagation through the complete call chain
// from API client → model processor → orchestrator → CLI
package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextualErrorPropagation tests error context propagation through all layers
func TestContextualErrorPropagation(t *testing.T) {
	t.Parallel()
	t.Run("Complete 4-layer error chain propagation", func(t *testing.T) {
		// Red phase: Start with a failing test that defines what we want

		// 1. API Client layer creates an LLMError
		apiError := &LLMError{
			Provider:      "openai",
			Message:       "Authentication failed",
			ErrorCategory: CategoryAuth,
			Original:      errors.New("invalid API key"),
			RequestID:     "req_12345",
		}

		// 2. Model Processor layer wraps with processing context
		modelProcessorError := fmt.Errorf("model processing failed for gpt-4: %w", apiError)

		// 3. Orchestrator layer wraps with coordination context
		orchestratorError := fmt.Errorf("orchestration error during model execution: %w", modelProcessorError)

		// 4. CLI layer wraps with user-facing context
		cliError := fmt.Errorf("CLI execution failed: %w", orchestratorError)

		// Verify that the original structured error is still accessible
		var originalLLMError *LLMError
		require.True(t, errors.As(cliError, &originalLLMError), "Should find original LLMError in chain")

		// Verify category preservation
		assert.Equal(t, CategoryAuth, originalLLMError.Category(), "Category should be preserved through chain")

		// Verify request ID preservation
		assert.Equal(t, "req_12345", originalLLMError.RequestID, "Request ID should be preserved")

		// Verify all context is present in error message
		errorMsg := cliError.Error()
		assert.Contains(t, errorMsg, "CLI execution failed", "Should contain CLI context")
		assert.Contains(t, errorMsg, "orchestration error", "Should contain orchestrator context")
		assert.Contains(t, errorMsg, "model processing failed", "Should contain model processor context")
		assert.Contains(t, errorMsg, "Authentication failed", "Should contain original API error")
	})

	t.Run("Error category detection through chain", func(t *testing.T) {
		testCases := []struct {
			name          string
			originalError *LLMError
			wantCategory  ErrorCategory
		}{
			{
				name: "auth_error_propagation",
				originalError: &LLMError{
					Provider:      "gemini",
					Message:       "Invalid API key",
					ErrorCategory: CategoryAuth,
				},
				wantCategory: CategoryAuth,
			},
			{
				name: "rate_limit_propagation",
				originalError: &LLMError{
					Provider:      "openai",
					Message:       "Rate limit exceeded",
					ErrorCategory: CategoryRateLimit,
				},
				wantCategory: CategoryRateLimit,
			},
			{
				name: "network_error_propagation",
				originalError: &LLMError{
					Provider:      "anthropic",
					Message:       "Connection timeout",
					ErrorCategory: CategoryNetwork,
				},
				wantCategory: CategoryNetwork,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate 4-layer error wrapping
				wrappedError := simulateFullChainPropagation(tc.originalError)

				// Verify category is preserved
				assert.True(t, IsCategory(wrappedError, tc.wantCategory),
					"Category %v should be detectable in wrapped error", tc.wantCategory)

				// Verify categorized error interface works
				categorizedErr, ok := IsCategorizedError(wrappedError)
				require.True(t, ok, "Should implement CategorizedError interface")
				assert.Equal(t, tc.wantCategory, categorizedErr.Category(),
					"Category should match original error")
			})
		}
	})

	t.Run("Context cancellation propagation", func(t *testing.T) {
		// Test context cancellation through the error chain
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Create an error that includes context cancellation
		apiError := &LLMError{
			Provider:      "openai",
			Message:       "Request cancelled",
			ErrorCategory: CategoryCancelled,
			Original:      ctx.Err(),
		}

		// Wrap through layers
		fullChainError := simulateFullChainPropagation(apiError)

		// Verify context cancellation is detectable
		assert.True(t, errors.Is(fullChainError, context.Canceled),
			"Should detect context cancellation in error chain")

		// Verify category is preserved
		assert.True(t, IsCategory(fullChainError, CategoryCancelled),
			"Should detect cancellation category")
	})

	t.Run("Error information preservation", func(t *testing.T) {
		// Create a detailed LLMError with all fields populated
		originalError := &LLMError{
			Provider:      "openai",
			Code:          "auth_error",
			StatusCode:    401,
			Message:       "API key validation failed",
			RequestID:     "req_67890",
			Original:      errors.New("HTTP 401: Unauthorized"),
			ErrorCategory: CategoryAuth,
			Suggestion:    "Check your API key configuration",
			Details:       "API key format is invalid",
		}

		// Wrap through all layers
		wrappedError := simulateFullChainPropagation(originalError)

		// Extract the original error
		var extractedError *LLMError
		require.True(t, errors.As(wrappedError, &extractedError),
			"Should extract original LLMError")

		// Verify all fields are preserved
		assert.Equal(t, "openai", extractedError.Provider)
		assert.Equal(t, "auth_error", extractedError.Code)
		assert.Equal(t, 401, extractedError.StatusCode)
		assert.Equal(t, "API key validation failed", extractedError.Message)
		assert.Equal(t, "req_67890", extractedError.RequestID)
		assert.Equal(t, CategoryAuth, extractedError.ErrorCategory)
		assert.Equal(t, "Check your API key configuration", extractedError.Suggestion)
		assert.Equal(t, "API key format is invalid", extractedError.Details)
		assert.NotNil(t, extractedError.Original)
	})
}

// TestErrorChainUtilities tests utility functions for error chain management
func TestErrorChainUtilities(t *testing.T) {
	t.Parallel()
	t.Run("IsCategorizedError utility", func(t *testing.T) {
		// Test with LLMError
		llmErr := &LLMError{
			Provider:      "gemini",
			ErrorCategory: CategoryRateLimit,
		}
		wrappedErr := fmt.Errorf("wrapped: %w", llmErr)

		categorizedErr, ok := IsCategorizedError(wrappedErr)
		assert.True(t, ok, "Should identify categorized error in chain")
		assert.Equal(t, CategoryRateLimit, categorizedErr.Category())

		// Test with non-categorized error
		plainErr := errors.New("plain error")
		_, ok = IsCategorizedError(plainErr)
		assert.False(t, ok, "Should not identify plain error as categorized")

		// Test with nil error
		_, ok = IsCategorizedError(nil)
		assert.False(t, ok, "Should handle nil error gracefully")
	})

	t.Run("IsCategory utility functions", func(t *testing.T) {
		// Create errors for each category
		testCases := []struct {
			category    ErrorCategory
			testFunc    func(error) bool
			description string
		}{
			{CategoryAuth, IsAuth, "auth error detection"},
			{CategoryRateLimit, IsRateLimit, "rate limit detection"},
			{CategoryInvalidRequest, IsInvalidRequest, "invalid request detection"},
			{CategoryNotFound, IsNotFound, "not found detection"},
			{CategoryServer, IsServer, "server error detection"},
			{CategoryNetwork, IsNetwork, "network error detection"},
			{CategoryCancelled, IsCancelled, "cancellation detection"},
			{CategoryInputLimit, IsInputLimit, "input limit detection"},
			{CategoryContentFiltered, IsContentFiltered, "content filtered detection"},
			{CategoryInsufficientCredits, IsInsufficientCredits, "insufficient credits detection"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Create LLMError with specific category
				llmErr := &LLMError{
					Provider:      "test",
					ErrorCategory: tc.category,
				}

				// Wrap multiple times
				wrappedErr := fmt.Errorf("layer3: %w",
					fmt.Errorf("layer2: %w",
						fmt.Errorf("layer1: %w", llmErr)))

				// Test category detection function
				assert.True(t, tc.testFunc(wrappedErr),
					"Should detect %v in wrapped error", tc.category)

				// Test negative case
				otherErr := &LLMError{
					Provider:      "test",
					ErrorCategory: CategoryUnknown,
				}
				assert.False(t, tc.testFunc(otherErr),
					"Should not detect %v in different category error", tc.category)
			})
		}
	})
}

// TestErrorWrappingBestPractices tests error wrapping best practices
func TestErrorWrappingBestPractices(t *testing.T) {
	t.Parallel()
	t.Run("Error wrapping preserves unwrapping", func(t *testing.T) {
		original := errors.New("original error")
		llmErr := &LLMError{
			Provider:      "test",
			Message:       "LLM error",
			Original:      original,
			ErrorCategory: CategoryUnknown,
		}

		// Wrap multiple times
		wrapped := fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", llmErr))

		// Should be able to unwrap to original
		assert.True(t, errors.Is(wrapped, original),
			"Should be able to unwrap to original error")
	})

	t.Run("Error message composition", func(t *testing.T) {
		apiErr := &LLMError{
			Provider:      "openai",
			Message:       "API error occurred",
			ErrorCategory: CategoryServer,
		}

		// Build realistic error chain
		modelErr := fmt.Errorf("model processing failed: %w", apiErr)
		orchestratorErr := fmt.Errorf("orchestrator failed to execute models: %w", modelErr)
		cliErr := fmt.Errorf("CLI command failed: %w", orchestratorErr)

		errorMessage := cliErr.Error()

		// Verify message composition maintains readability
		expectedParts := []string{
			"CLI command failed",
			"orchestrator failed to execute models",
			"model processing failed",
			"API error occurred",
		}

		for _, part := range expectedParts {
			assert.Contains(t, errorMessage, part,
				"Error message should contain context from each layer")
		}

		// Verify the message reads logically
		assert.True(t, strings.Index(errorMessage, "CLI command failed") <
			strings.Index(errorMessage, "orchestrator failed"),
			"CLI context should appear before orchestrator context")
	})

	t.Run("Error categorization accuracy", func(t *testing.T) {
		// Test error categorization from different status codes
		testCases := []struct {
			statusCode   int
			wantCategory ErrorCategory
		}{
			{401, CategoryAuth},
			{403, CategoryAuth},
			{402, CategoryInsufficientCredits},
			{429, CategoryRateLimit},
			{400, CategoryInvalidRequest},
			{404, CategoryNotFound},
			{500, CategoryServer},
			{502, CategoryServer},
			{503, CategoryServer},
			{0, CategoryUnknown},
			{200, CategoryUnknown},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("status_%d", tc.statusCode), func(t *testing.T) {
				category := GetErrorCategoryFromStatusCode(tc.statusCode)
				assert.Equal(t, tc.wantCategory, category,
					"Status code %d should map to category %v", tc.statusCode, tc.wantCategory)
			})
		}
	})
}

// TestConcurrentErrorHandling tests error handling in concurrent scenarios
func TestConcurrentErrorHandling(t *testing.T) {
	t.Parallel()
	t.Run("Concurrent error wrapping safety", func(t *testing.T) {
		// Create an error that might be accessed concurrently
		baseErr := &LLMError{
			Provider:      "openai",
			Message:       "Base error",
			ErrorCategory: CategoryServer,
		}

		const goroutines = 10
		const iterations = 100

		// Channel to collect errors from goroutines
		errorChan := make(chan error, goroutines*iterations)

		// Spawn multiple goroutines that wrap the same base error
		for g := 0; g < goroutines; g++ {
			go func(id int) {
				for i := 0; i < iterations; i++ {
					// Each goroutine wraps the error with unique context
					wrapped := fmt.Errorf("goroutine_%d_iteration_%d: %w", id, i, baseErr)
					errorChan <- wrapped
				}
			}(g)
		}

		// Collect all wrapped errors
		var collectedErrors []error
		for i := 0; i < goroutines*iterations; i++ {
			err := <-errorChan
			collectedErrors = append(collectedErrors, err)
		}

		// Verify all errors maintain the original category
		for i, err := range collectedErrors {
			assert.True(t, IsCategory(err, CategoryServer),
				"Error %d should maintain server category", i)

			// Verify original error is still accessible
			var originalErr *LLMError
			assert.True(t, errors.As(err, &originalErr),
				"Error %d should contain original LLMError", i)
		}
	})
}

// TestPerformanceCharacteristics tests error handling performance
func TestPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("Error chain traversal performance", func(t *testing.T) {
		// Create a realistic 4-layer error chain
		apiErr := &LLMError{
			Provider:      "openai",
			Message:       "API error",
			ErrorCategory: CategoryRateLimit,
		}

		wrapped := simulateFullChainPropagation(apiErr)

		// Benchmark category detection
		start := time.Now()
		const iterations = 10000

		for i := 0; i < iterations; i++ {
			_ = IsCategory(wrapped, CategoryRateLimit)
		}

		duration := time.Since(start)
		avgDuration := duration / iterations

		// Category detection should be fast (sub-10-microsecond)
		assert.True(t, avgDuration < 10*time.Microsecond,
			"Category detection should be fast: %v per operation", avgDuration)

		t.Logf("Error category detection: %v per operation (%d iterations)",
			avgDuration, iterations)
	})
}

// Helper functions for testing

// simulateFullChainPropagation simulates error propagation through all 4 layers
func simulateFullChainPropagation(originalError *LLMError) error {
	// Layer 1: API Client → Model Processor
	modelProcessorError := fmt.Errorf("model processing failed: %w", originalError)

	// Layer 2: Model Processor → Orchestrator
	orchestratorError := fmt.Errorf("orchestrator execution failed: %w", modelProcessorError)

	// Layer 3: Orchestrator → CLI
	cliError := fmt.Errorf("CLI command execution failed: %w", orchestratorError)

	return cliError
}

// TestContextualErrorHandling tests the comprehensive contextual error system
func TestContextualErrorHandling(t *testing.T) {
	t.Parallel()
	t.Run("ContextualError basic functionality", func(t *testing.T) {
		originalError := errors.New("original error")
		context := LayerContext{
			Layer:         "api-client",
			Operation:     "makeRequest",
			Timestamp:     time.Now(),
			Details:       map[string]interface{}{"endpoint": "/api/test"},
			CorrelationID: "test-correlation-id",
		}

		contextualErr := &ContextualError{
			Original: originalError,
			Context:  context,
		}

		// Test Error() method
		errorMsg := contextualErr.Error()
		assert.Contains(t, errorMsg, "[api-client:makeRequest]")
		assert.Contains(t, errorMsg, "original error")

		// Test Unwrap() method
		unwrapped := contextualErr.Unwrap()
		assert.Equal(t, originalError, unwrapped)

		// Test Category() method - should return CategoryUnknown for standard error
		category := contextualErr.Category()
		assert.Equal(t, CategoryUnknown, category)
	})

	t.Run("ContextualError with categorized original error", func(t *testing.T) {
		originalLLMError := &LLMError{
			Provider:      "openai",
			Message:       "Rate limit exceeded",
			ErrorCategory: CategoryRateLimit,
		}

		contextualErr := &ContextualError{
			Original: originalLLMError,
			Context: LayerContext{
				Layer:     "api-client",
				Operation: "generateContent",
			},
		}

		// Test that Category() delegates to the original error
		category := contextualErr.Category()
		assert.Equal(t, CategoryRateLimit, category)

		// Test that the contextual error is itself categorized
		categorizedErr, ok := IsCategorizedError(contextualErr)
		assert.True(t, ok)
		assert.Equal(t, CategoryRateLimit, categorizedErr.Category())
	})

	t.Run("WrapWithContext functionality", func(t *testing.T) {
		originalError := errors.New("database connection failed")
		details := map[string]interface{}{
			"database": "postgres",
			"host":     "localhost",
		}

		// Test wrapping with context
		wrappedErr := WrapWithContext(originalError, "data-layer", "connect", details, "corr-123")
		assert.NotNil(t, wrappedErr)

		// Verify it's a ContextualError
		contextualErr, ok := wrappedErr.(*ContextualError)
		require.True(t, ok)

		// Verify context information
		assert.Equal(t, "data-layer", contextualErr.Context.Layer)
		assert.Equal(t, "connect", contextualErr.Context.Operation)
		assert.Equal(t, "corr-123", contextualErr.Context.CorrelationID)
		assert.Equal(t, details, contextualErr.Context.Details)
		assert.Equal(t, originalError, contextualErr.Original)

		// Test that nil error returns nil
		nilWrapped := WrapWithContext(nil, "layer", "op", nil, "corr")
		assert.Nil(t, nilWrapped)
	})

	t.Run("Layer-specific wrapping functions", func(t *testing.T) {
		originalError := errors.New("test error")
		correlationID := "test-correlation"

		// Test WrapAPIClientError
		apiErr := WrapAPIClientError(originalError, "openai", "generateContent",
			map[string]interface{}{"model": "gpt-4"}, correlationID)
		apiContextual, ok := apiErr.(*ContextualError)
		require.True(t, ok)
		assert.Equal(t, "api-client", apiContextual.Context.Layer)
		assert.Equal(t, "generateContent", apiContextual.Context.Operation)
		assert.Equal(t, "openai", apiContextual.Context.Details["provider"])
		assert.Equal(t, "gpt-4", apiContextual.Context.Details["model"])

		// Test WrapModelProcessorError
		modelErr := WrapModelProcessorError(originalError, "gpt-4", "processResponse",
			map[string]interface{}{"tokens": 1000}, correlationID)
		modelContextual, ok := modelErr.(*ContextualError)
		require.True(t, ok)
		assert.Equal(t, "model-processor", modelContextual.Context.Layer)
		assert.Equal(t, "processResponse", modelContextual.Context.Operation)
		assert.Equal(t, "gpt-4", modelContextual.Context.Details["model"])

		// Test WrapOrchestratorError
		orchestratorErr := WrapOrchestratorError(originalError, "orchestrate", "model-execution",
			map[string]interface{}{"models": []string{"gpt-4", "gemini"}}, correlationID)
		orchestratorContextual, ok := orchestratorErr.(*ContextualError)
		require.True(t, ok)
		assert.Equal(t, "orchestrator", orchestratorContextual.Context.Layer)
		assert.Equal(t, "orchestrate", orchestratorContextual.Context.Operation)
		assert.Equal(t, "model-execution", orchestratorContextual.Context.Details["workflow_stage"])

		// Test WrapCLIError
		cliErr := WrapCLIError(originalError, "thinktank", "run", []string{"--model", "gpt-4"},
			map[string]interface{}{"dry_run": false}, correlationID)
		cliContextual, ok := cliErr.(*ContextualError)
		require.True(t, ok)
		assert.Equal(t, "cli", cliContextual.Context.Layer)
		assert.Equal(t, "run", cliContextual.Context.Operation)
		assert.Equal(t, "thinktank", cliContextual.Context.Details["command"])
		assert.Equal(t, []string{"--model", "gpt-4"}, cliContextual.Context.Details["args"])

		// Test layer wrappers with nil details
		nilDetailsErr := WrapAPIClientError(originalError, "openai", "test", nil, correlationID)
		nilDetailsContextual, ok := nilDetailsErr.(*ContextualError)
		require.True(t, ok)
		assert.NotNil(t, nilDetailsContextual.Context.Details)
		assert.Equal(t, "openai", nilDetailsContextual.Context.Details["provider"])
	})

	t.Run("ExtractLayerContext functionality", func(t *testing.T) {
		// Create a multi-layer error chain
		originalError := &LLMError{
			Provider:      "openai",
			Message:       "Rate limit exceeded",
			ErrorCategory: CategoryRateLimit,
		}

		// Wrap with API client context
		apiErr := WrapAPIClientError(originalError, "openai", "generateContent",
			map[string]interface{}{"model": "gpt-4"}, "corr-123")

		// Wrap with model processor context
		modelErr := WrapModelProcessorError(apiErr, "gpt-4", "processModel",
			map[string]interface{}{"tokens": 2000}, "corr-123")

		// Extract API client context
		apiContext, found := ExtractLayerContext(modelErr, "api-client")
		assert.True(t, found)
		assert.Equal(t, "api-client", apiContext.Layer)
		assert.Equal(t, "generateContent", apiContext.Operation)
		assert.Equal(t, "corr-123", apiContext.CorrelationID)
		assert.Equal(t, "openai", apiContext.Details["provider"])

		// Extract model processor context
		modelContext, found := ExtractLayerContext(modelErr, "model-processor")
		assert.True(t, found)
		assert.Equal(t, "model-processor", modelContext.Layer)
		assert.Equal(t, "processModel", modelContext.Operation)

		// Try to extract non-existent layer
		_, found = ExtractLayerContext(modelErr, "non-existent-layer")
		assert.False(t, found)

		// Test with nil error
		_, found = ExtractLayerContext(nil, "any-layer")
		assert.False(t, found)

		// Test with non-contextual error
		plainErr := errors.New("plain error")
		_, found = ExtractLayerContext(plainErr, "any-layer")
		assert.False(t, found)
	})

	t.Run("ExtractCorrelationID functionality", func(t *testing.T) {
		// Test with contextual error
		originalError := errors.New("test error")
		wrappedErr := WrapWithContext(originalError, "test-layer", "test-op", nil, "test-correlation-id")

		correlationID := ExtractCorrelationID(wrappedErr)
		assert.Equal(t, "test-correlation-id", correlationID)

		// Test with multi-layer error (should find the first correlation ID)
		layeredErr := WrapWithContext(wrappedErr, "outer-layer", "outer-op", nil, "outer-correlation")
		correlationID = ExtractCorrelationID(layeredErr)
		assert.Equal(t, "outer-correlation", correlationID)

		// Test with nil error
		correlationID = ExtractCorrelationID(nil)
		assert.Equal(t, "", correlationID)

		// Test with non-contextual error
		plainErr := errors.New("plain error")
		correlationID = ExtractCorrelationID(plainErr)
		assert.Equal(t, "", correlationID)

		// Test with contextual error but empty correlation ID
		emptyCorrelationErr := WrapWithContext(originalError, "test-layer", "test-op", nil, "")
		correlationID = ExtractCorrelationID(emptyCorrelationErr)
		assert.Equal(t, "", correlationID)
	})

	t.Run("ExtractRecoveryInformation functionality", func(t *testing.T) {
		// Test with nil error
		recovery := ExtractRecoveryInformation(nil)
		assert.Equal(t, RecoveryInformation{}, recovery)

		// Test with rate limit error
		rateLimitErr := &LLMError{
			Provider:      "openai",
			Message:       "Rate limit exceeded",
			ErrorCategory: CategoryRateLimit,
		}
		wrappedRateLimit := WrapWithContext(rateLimitErr, "api-client", "request", nil, "rate-limit-corr")

		recovery = ExtractRecoveryInformation(wrappedRateLimit)
		assert.Equal(t, "rate-limit-corr", recovery.CorrelationID)
		assert.Contains(t, recovery.UserFacingMessage, "Request rate limit exceeded")
		assert.Contains(t, recovery.DeveloperDetails, "OpenAI API rate limit hit")
		assert.True(t, recovery.RetryPossible)
		assert.Equal(t, time.Second*60, recovery.EstimatedWaitTime)
		assert.Len(t, recovery.SuggestedActions, 4) // Should have 4 suggested actions

		// Verify suggested actions contain expected layers
		actionLayers := make(map[string]bool)
		for _, action := range recovery.SuggestedActions {
			actionLayers[action.Layer] = true
		}
		assert.True(t, actionLayers["cli"])
		assert.True(t, actionLayers["orchestrator"])
		assert.True(t, actionLayers["model-processor"])
		assert.True(t, actionLayers["api-client"])

		// Test with auth error
		authErr := &LLMError{
			Provider:      "openai",
			Message:       "Invalid API key",
			ErrorCategory: CategoryAuth,
		}
		recovery = ExtractRecoveryInformation(authErr)
		assert.Contains(t, recovery.UserFacingMessage, "Authentication failed")
		assert.False(t, recovery.RetryPossible)
		assert.Len(t, recovery.SuggestedActions, 2) // Auth errors have 2 suggested actions

		// Test with network error
		networkErr := &LLMError{
			Provider:      "openai",
			Message:       "Connection timeout",
			ErrorCategory: CategoryNetwork,
		}
		recovery = ExtractRecoveryInformation(networkErr)
		assert.Contains(t, recovery.UserFacingMessage, "Network error occurred")
		assert.True(t, recovery.RetryPossible)
		assert.Equal(t, time.Second*30, recovery.EstimatedWaitTime)

		// Test with unknown error category
		unknownErr := &LLMError{
			Provider:      "openai",
			Message:       "Unknown error",
			ErrorCategory: CategoryUnknown,
		}
		recovery = ExtractRecoveryInformation(unknownErr)
		assert.Contains(t, recovery.UserFacingMessage, "An error occurred")
		assert.True(t, recovery.RetryPossible)

		// Test with non-categorized error
		plainErr := errors.New("plain error")
		recovery = ExtractRecoveryInformation(plainErr)
		assert.Equal(t, "", recovery.UserFacingMessage)
		assert.Equal(t, "", recovery.CorrelationID)
	})

	t.Run("GetUserFriendlyErrorMessage functionality", func(t *testing.T) {
		// Test with nil error
		msg := GetUserFriendlyErrorMessage(nil)
		assert.Equal(t, "", msg)

		// Test with categorized error that has recovery information
		rateLimitErr := &LLMError{
			Provider:      "openai",
			Message:       "Rate limit exceeded",
			ErrorCategory: CategoryRateLimit,
		}
		msg = GetUserFriendlyErrorMessage(rateLimitErr)
		assert.Contains(t, msg, "Request rate limit exceeded")

		// Test with plain error (should fall back to error message)
		plainErr := errors.New("plain error message")
		msg = GetUserFriendlyErrorMessage(plainErr)
		assert.Equal(t, "plain error message", msg)
	})

	t.Run("GetDeveloperDebugInfo functionality", func(t *testing.T) {
		// Test with nil error
		debugInfo := GetDeveloperDebugInfo(nil)
		assert.Nil(t, debugInfo)

		// Create a complex multi-layer error chain
		originalErr := &LLMError{
			Provider:      "openai",
			Message:       "Rate limit exceeded",
			ErrorCategory: CategoryRateLimit,
		}

		// Build a 4-layer error chain with contextual information
		apiErr := WrapAPIClientError(originalErr, "openai", "generateContent",
			map[string]interface{}{"model": "gpt-4", "endpoint": "/v1/chat/completions"}, "debug-corr-123")

		modelErr := WrapModelProcessorError(apiErr, "gpt-4", "processResponse",
			map[string]interface{}{"tokens": 2000, "temperature": 0.7}, "debug-corr-123")

		orchestratorErr := WrapOrchestratorError(modelErr, "executeModels", "parallel-processing",
			map[string]interface{}{"total_models": 3, "current_model": 1}, "debug-corr-123")

		cliErr := WrapCLIError(orchestratorErr, "thinktank", "generate",
			[]string{"--model", "gpt-4", "--instructions", "test.md"},
			map[string]interface{}{"dry_run": false, "verbose": true}, "debug-corr-123")

		debugInfo = GetDeveloperDebugInfo(cliErr)
		assert.NotNil(t, debugInfo)

		// Check correlation ID
		assert.Equal(t, "debug-corr-123", debugInfo["correlation_id"])

		// Check error category
		assert.Equal(t, "RateLimit", debugInfo["error_category"])

		// Check error message
		errorMsg, ok := debugInfo["error_message"].(string)
		assert.True(t, ok)
		assert.Contains(t, errorMsg, "Rate limit exceeded")

		// Check layer information
		cliInfo, ok := debugInfo["cli"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "generate", cliInfo["operation"])
		assert.NotNil(t, cliInfo["timestamp"])
		cliDetails, ok := cliInfo["details"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "thinktank", cliDetails["command"])

		orchestratorInfo, ok := debugInfo["orchestrator"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "executeModels", orchestratorInfo["operation"])

		modelProcessorInfo, ok := debugInfo["model-processor"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "processResponse", modelProcessorInfo["operation"])

		apiClientInfo, ok := debugInfo["api-client"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "generateContent", apiClientInfo["operation"])

		// Check recovery information
		recoveryInfo, ok := debugInfo["recovery_info"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, recoveryInfo["developer_details"], "OpenAI API rate limit hit")
		assert.Equal(t, true, recoveryInfo["retry_possible"])
		assert.NotNil(t, recoveryInfo["suggested_actions"])

		// Test with simpler error using api-client layer (one of the expected layers)
		simpleErr := WrapAPIClientError(errors.New("simple error"), "test-provider", "test-op",
			map[string]interface{}{"key": "value"}, "simple-corr")

		simpleDebugInfo := GetDeveloperDebugInfo(simpleErr)
		assert.Equal(t, "simple-corr", simpleDebugInfo["correlation_id"])
		// The error message will include the contextual wrapper formatting
		simpleErrorMsg, ok := simpleDebugInfo["error_message"].(string)
		assert.True(t, ok)
		assert.Contains(t, simpleErrorMsg, "simple error")

		apiLayerInfo, ok := simpleDebugInfo["api-client"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test-op", apiLayerInfo["operation"])
	})
}

// BenchmarkErrorChainTraversal benchmarks error chain operations
func BenchmarkErrorChainTraversal(b *testing.B) {
	b.ResetTimer() // Reset timer before benchmarking
	// Create a 4-layer error chain

	apiErr := &LLMError{
		Provider:      "openai",
		Message:       "API error",
		ErrorCategory: CategoryRateLimit,
	}
	chainedErr := simulateFullChainPropagation(apiErr)

	b.Run("IsCategorizedError", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = IsCategorizedError(chainedErr)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		var llmErr *LLMError
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = errors.As(chainedErr, &llmErr)
		}
	})

	b.Run("error.Error()", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = chainedErr.Error()
		}
	})

	// Add benchmarks for contextual error functions
	contextualErr := WrapWithContext(apiErr, "test-layer", "test-op",
		map[string]interface{}{"key": "value"}, "test-corr")

	b.Run("ExtractLayerContext", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ExtractLayerContext(contextualErr, "test-layer")
		}
	})

	b.Run("ExtractCorrelationID", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ExtractCorrelationID(contextualErr)
		}
	})

	b.Run("ExtractRecoveryInformation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ExtractRecoveryInformation(contextualErr)
		}
	})
}
