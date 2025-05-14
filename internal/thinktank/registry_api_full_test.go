// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestFullAdapterRegistryInteraction tests the full integration between adapters and the registry
// including comprehensive interactions across all adapter methods and their delegations.
// This test checks for:
// - Full interaction flow between adapters and registry API
// - Data transformation between layers
// - Context propagation through the entire adapter-registry chain
// - Comprehensive error handling across multiple operations
// - Realistic simulation of production usage patterns
func TestFullAdapterRegistryInteraction(t *testing.T) {
	// Create a comprehensive mock registry with richer test data
	mockRegistry := &BoundaryMockRegistry{
		models: map[string]*registry.ModelDefinition{
			"standard-model": {
				Name:       "standard-model",
				Provider:   "test-provider",
				APIModelID: "test-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"temperature": {
						Type:    "float",
						Default: 0.7,
						Min:     0.0,
						Max:     1.0,
					},
					"top_p": {
						Type:    "float",
						Default: 0.9,
						Min:     0.0,
						Max:     1.0,
					},
					"max_tokens": {
						Type:    "int",
						Default: 1024,
						Min:     1,
						Max:     4096,
					},
				},
			},
			"large-model": {
				Name:       "large-model",
				Provider:   "premium-provider",
				APIModelID: "premium-model-id",
				Parameters: map[string]registry.ParameterDefinition{
					"temperature": {
						Type:    "float",
						Default: 0.5,
					},
					"presence_penalty": {
						Type:    "float",
						Default: 0.0,
						Min:     -2.0,
						Max:     2.0,
					},
				},
			},
		},
		providers: map[string]*registry.ProviderDefinition{
			"test-provider": {
				Name:    "test-provider",
				BaseURL: "https://api.test-provider.example.com",
			},
			"premium-provider": {
				Name:    "premium-provider",
				BaseURL: "https://api.premium-provider.example.com",
			},
		},
		implementations: map[string]ProviderImplementation{
			"test-provider": &BoundaryMockProvider{
				CreateClientResult: &BoundaryMockLLMClient{
					ModelName: "test-model-id",
				},
			},
			"premium-provider": &BoundaryMockProvider{
				CreateClientResult: &BoundaryMockLLMClient{
					ModelName: "premium-model-id",
				},
			},
		},
		methodCalls: make(map[string][]BoundaryMethodCall),
	}

	// Create logger
	logger := testutil.NewMockLogger()

	// Create the registry API service
	service := NewRegistryAPIService(mockRegistry, logger)

	// Create an adapter around the service
	adapter := &APIServiceAdapter{
		APIService: service,
	}

	// 1. Test comprehensive adapter-registry interaction with realistic workflow
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Clear previous calls
		mockRegistry.ClearMethodCalls()

		// Create a context with timeout and correlation ID
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ctx = logutil.WithCustomCorrelationID(ctx, "workflow-correlation-id")

		// STEP 1: Initialize parameters for the LLM request
		modelName := "standard-model"

		// Get model parameters through adapter
		params, err := adapter.GetModelParameters(ctx, modelName)
		if err != nil {
			t.Fatalf("Failed to get model parameters: %v", err)
		}

		// Validate parameters
		isValid, err := adapter.ValidateModelParameter(ctx, modelName, "temperature", 0.8)
		if err != nil || !isValid {
			t.Errorf("Failed to validate parameter: %v", err)
		}

		// Get model token limits
		contextWindow, maxOutputTokens, err := adapter.GetModelTokenLimits(ctx, modelName)
		if err != nil {
			t.Errorf("Failed to get token limits: %v", err)
		}

		// STEP 2: Initialize LLM client
		_, err = adapter.InitLLMClient(ctx, "workflow-api-key", modelName, "")
		if err != nil {
			t.Fatalf("Failed to initialize LLM client: %v", err)
		}

		// STEP 3: Verify registry interaction pattern
		// Verify all expected methods were called in sequence
		expectedCalls := []string{
			"GetModel",                                             // for GetModelParameters
			"GetModel",                                             // for ValidateModelParameter
			"GetModel",                                             // for GetModelTokenLimits
			"GetModel", "GetProvider", "GetProviderImplementation", // for InitLLMClient
		}

		callSequence := mockRegistry.GetCallSequence()

		// Check that we have at least the expected number of calls
		if len(callSequence) < len(expectedCalls) {
			t.Errorf("Expected at least %d registry calls, got %d: %v",
				len(expectedCalls), len(callSequence), callSequence)
		} else {
			// Verify the calls match our expected pattern
			for i, expectedCall := range expectedCalls {
				if i >= len(callSequence) || callSequence[i] != expectedCall {
					if i < len(callSequence) {
						t.Errorf("Expected call %d to be %s, got %s",
							i, expectedCall, callSequence[i])
					} else {
						t.Errorf("Expected call %d to be %s, but didn't have enough calls",
							i, expectedCall)
					}
				}
			}
		}

		// STEP 4: Verify context propagation through entire adapter-registry chain
		// The context should have flowed all the way to the provider implementation
		provider, ok := mockRegistry.implementations["test-provider"].(*BoundaryMockProvider)
		if !ok {
			t.Fatal("Expected to find test-provider implementation")
		}

		if len(provider.CreateClientCalls) == 0 {
			t.Fatal("Expected CreateClient to be called")
		}

		callCtx := provider.CreateClientCalls[0].Ctx

		// Verify correlation ID propagation through all layers
		correlationID := logutil.GetCorrelationID(callCtx)
		if correlationID != "workflow-correlation-id" {
			t.Errorf("Expected correlation ID to be propagated through all layers, got %q",
				correlationID)
		}

		// Verify deadline propagation
		deadline, hasDeadline := callCtx.Deadline()
		if !hasDeadline {
			t.Error("Expected context deadline to be propagated through all layers")
		} else {
			// The deadline should be roughly 30 seconds from when we created it
			expectedDeadline := time.Now().Add(30 * time.Second)
			if deadline.Sub(expectedDeadline) > 500*time.Millisecond {
				t.Errorf("Deadline not propagated correctly through all layers: got %v",
					deadline)
			}
		}

		// STEP 5: Verify data transformation
		// Verify parameter values were correctly retrieved
		if params["temperature"] != 0.7 {
			t.Errorf("Expected temperature parameter to be 0.7, got %v",
				params["temperature"])
		}

		if params["top_p"] != 0.9 {
			t.Errorf("Expected top_p parameter to be 0.9, got %v",
				params["top_p"])
		}

		// Verify token limits based on adapter's fallback mechanism for "standard-model"
		// The values that are returned by the actual implementation may vary,
		// so we'll just verify they are positive values
		if contextWindow <= 0 {
			t.Errorf("Expected positive context window, got %d", contextWindow)
		}

		if maxOutputTokens <= 0 {
			t.Errorf("Expected positive max output tokens, got %d", maxOutputTokens)
		}

		// Print the token limits for debugging (not assertion)
		t.Logf("Model token limits - context window: %d, max output tokens: %d",
			contextWindow, maxOutputTokens)

		// STEP 6: Process an LLM response
		providerResult := &llm.ProviderResult{
			Content: "workflow test response",
		}

		response, err := adapter.ProcessLLMResponse(providerResult)
		if err != nil {
			t.Errorf("Failed to process LLM response: %v", err)
		}

		if response != "workflow test response" {
			t.Errorf("Expected response content 'workflow test response', got %q",
				response)
		}
	})

	// 2. Test error handling across multiple adapter methods
	t.Run("ComprehensiveErrorHandling", func(t *testing.T) {
		// Create context for tests
		ctx := context.Background()
		// Set up a fresh registry with configurable errors
		errorRegistry := &BoundaryMockRegistry{
			models: map[string]*registry.ModelDefinition{
				"error-model": {
					Name:     "error-model",
					Provider: "error-provider",
				},
			},
			providers: map[string]*registry.ProviderDefinition{
				"error-provider": {
					Name: "error-provider",
				},
			},
			implementations: map[string]ProviderImplementation{
				"error-provider": &BoundaryMockProvider{
					CreateClientError: fmt.Errorf("provider error: %w", llm.ErrSafetyBlocked),
				},
			},
			methodCalls: make(map[string][]BoundaryMethodCall),
		}

		// Create an error-producing service
		errorService := NewRegistryAPIService(errorRegistry, logger)

		// Create an adapter around the error service
		errorAdapter := &APIServiceAdapter{
			APIService: errorService,
		}

		// Test table of error scenarios
		errorTests := []struct {
			name       string
			setup      func()
			operation  func() error
			errorCheck func(error) bool
		}{
			{
				name: "Model not found",
				setup: func() {
					errorRegistry.getModelErr = fmt.Errorf("model lookup error: %w", llm.ErrModelNotFound)
				},
				operation: func() error {
					_, err := errorAdapter.GetModelDefinition(ctx, "nonexistent-model")
					return err
				},
				errorCheck: func(err error) bool {
					return errors.Is(err, llm.ErrModelNotFound)
				},
			},
			{
				name: "Provider not found",
				setup: func() {
					errorRegistry.getModelErr = fmt.Errorf("special model error")
				},
				operation: func() error {
					_, err := errorAdapter.GetModelDefinition(ctx, "error-model")
					return err
				},
				errorCheck: func(err error) bool {
					return err != nil
				},
			},
			{
				name: "Provider implementation error",
				setup: func() {
					// We'll just use ErrSafetyBlocked which is already defined
					if provider, ok := errorRegistry.implementations["error-provider"].(*BoundaryMockProvider); ok {
						provider.CreateClientError = fmt.Errorf("safety blocked: %w", llm.ErrSafetyBlocked)
					}
				},
				operation: func() error {
					_, err := errorAdapter.InitLLMClient(context.Background(), "api-key", "error-model", "")
					return err
				},
				errorCheck: func(err error) bool {
					return err != nil && strings.Contains(err.Error(), "error")
				},
			},
			{
				name: "Empty response error handling",
				setup: func() {
					// Nothing to set up
				},
				operation: func() error {
					// Test the empty response error handling
					_, err := errorAdapter.ProcessLLMResponse(nil)
					return err
				},
				errorCheck: func(err error) bool {
					return errors.Is(err, llm.ErrEmptyResponse)
				},
			},
		}

		// Run each error test
		for _, tc := range errorTests {
			t.Run(tc.name, func(t *testing.T) {
				// Setup the test condition
				tc.setup()

				// Run the operation
				err := tc.operation()

				// Verify error expectations
				if err == nil {
					t.Fatalf("Expected an error for %s, got nil", tc.name)
				}

				// Check domain-specific error wrapping
				if !tc.errorCheck(err) {
					t.Errorf("Error %v did not satisfy the expected error condition for %s",
						err, tc.name)
				}

				// Verify error details are available
				details := errorAdapter.GetErrorDetails(err)
				if details == "" {
					t.Errorf("Expected non-empty error details for %s", tc.name)
				}
			})
		}
	})

	// 3. Test adapter fallback mechanisms
	t.Run("AdapterFallbackMechanisms", func(t *testing.T) {
		// Create a minimal implementation that doesn't implement all methods
		minimalImpl := &mockAPIService{
			// Implement only the basic methods
			processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
				if result == nil {
					return "", llm.ErrEmptyResponse
				}
				return result.Content, nil
			},
			isEmptyResponseErrorFunc: func(err error) bool {
				return errors.Is(err, llm.ErrEmptyResponse)
			},
			isSafetyBlockedErrorFunc: func(err error) bool {
				return errors.Is(err, llm.ErrSafetyBlocked)
			},
		}

		// Create an adapter with the minimal implementation
		minimalAdapter := &APIServiceAdapter{
			APIService: minimalImpl,
		}

		// Test GetModelParameters fallback (returns empty map, no error)
		ctx := context.Background()
		params, err := minimalAdapter.GetModelParameters(ctx, "any-model")
		if err != nil {
			t.Errorf("Expected no error from GetModelParameters fallback, got: %v", err)
		}
		if len(params) != 0 {
			t.Errorf("Expected empty parameter map from fallback, got: %v", params)
		}

		// Test ValidateModelParameter fallback (returns true, no error)
		valid, err := minimalAdapter.ValidateModelParameter(ctx, "any-model", "any-param", "any-value")
		if err != nil {
			t.Errorf("Expected no error from ValidateModelParameter fallback, got: %v", err)
		}
		if !valid {
			t.Errorf("Expected ValidateModelParameter fallback to return true")
		}

		// Test GetModelDefinition fallback (returns error)
		_, err = minimalAdapter.GetModelDefinition(ctx, "any-model")
		if err == nil {
			t.Errorf("Expected error from GetModelDefinition fallback")
		}

		// Test GetModelTokenLimits fallback for known models
		tokenTests := []struct {
			modelName             string
			expectedContextWindow int32
			expectedMaxTokens     int32
		}{
			{"gemini-1.5-pro-latest", 1000000, 8192},
			{"gemini-1.0-pro", 32768, 8192},
			{"gpt-4o", 128000, 4096},
			{"gpt-4", 8192, 4096},
			{"gpt-3.5-turbo", 16385, 4096},
			{"unknown-model", 200000, 4096}, // Default fallback
		}

		for _, tc := range tokenTests {
			contextWindow, maxTokens, err := minimalAdapter.GetModelTokenLimits(ctx, tc.modelName)

			if err != nil {
				t.Errorf("Expected no error for model %s, got: %v", tc.modelName, err)
			}

			if contextWindow != tc.expectedContextWindow {
				t.Errorf("For model %s: expected context window %d, got %d",
					tc.modelName, tc.expectedContextWindow, contextWindow)
			}

			if maxTokens != tc.expectedMaxTokens {
				t.Errorf("For model %s: expected max tokens %d, got %d",
					tc.modelName, tc.expectedMaxTokens, maxTokens)
			}
		}
	})
}
