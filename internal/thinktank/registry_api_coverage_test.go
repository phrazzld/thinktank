package thinktank

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestRegistryAPICoverageBoost adds tests for functions that were missing coverage
func TestRegistryAPICoverageBoost(t *testing.T) {
	logger := testutil.NewMockLogger()
	service := NewRegistryAPIService(logger)
	ctx := context.Background()

	t.Run("GetModelDefinition", func(t *testing.T) {
		// Test valid model
		info, err := service.GetModelDefinition(ctx, "gpt-4.1")
		if err != nil {
			t.Errorf("Expected no error for valid model, got: %v", err)
		}
		if info == nil {
			t.Error("Expected model info, got nil")
		}

		// Test invalid model
		_, err = service.GetModelDefinition(ctx, "invalid-model")
		if err == nil {
			t.Error("Expected error for invalid model")
		}
	})

	t.Run("ProcessLLMResponse", func(t *testing.T) {
		// Test nil result
		_, err := service.ProcessLLMResponse(nil)
		if err == nil {
			t.Error("Expected error for nil result")
		}

		// Test empty content
		result := &llm.ProviderResult{Content: ""}
		_, err = service.ProcessLLMResponse(result)
		if err == nil {
			t.Error("Expected error for empty content")
		}

		// Test valid content
		result = &llm.ProviderResult{Content: "test response"}
		content, err := service.ProcessLLMResponse(result)
		if err != nil {
			t.Errorf("Expected no error for valid result, got: %v", err)
		}
		if content != "test response" {
			t.Errorf("Expected 'test response', got: %v", content)
		}
	})

	t.Run("IsEmptyResponseError", func(t *testing.T) {
		// Test with empty response error
		if !service.IsEmptyResponseError(llm.ErrEmptyResponse) {
			t.Error("Expected true for empty response error")
		}

		// Test with other error
		if service.IsEmptyResponseError(errors.New("some other error")) {
			t.Error("Expected false for non-empty-response error")
		}

		// Test with nil
		if service.IsEmptyResponseError(nil) {
			t.Error("Expected false for nil error")
		}
	})

	t.Run("IsSafetyBlockedError", func(t *testing.T) {
		// Test with safety blocked error
		if !service.IsSafetyBlockedError(llm.ErrSafetyBlocked) {
			t.Error("Expected true for safety blocked error")
		}

		// Test with other error
		if service.IsSafetyBlockedError(llm.ErrEmptyResponse) {
			t.Error("Expected false for non-safety error")
		}

		// Test with nil
		if service.IsSafetyBlockedError(nil) {
			t.Error("Expected false for nil error")
		}
	})

	t.Run("GetErrorDetails", func(t *testing.T) {
		// Test with nil error
		details := service.GetErrorDetails(nil)
		if details != "no error" {
			t.Errorf("Expected 'no error' for nil, got: %v", details)
		}

		// Test with regular error
		err := errors.New("test error")
		details = service.GetErrorDetails(err)
		if details == "" {
			t.Error("Expected non-empty details for error")
		}
	})

	t.Run("GetModelParameters", func(t *testing.T) {
		// Test valid model
		params, err := service.GetModelParameters(ctx, "gpt-4.1")
		if err != nil {
			t.Errorf("Expected no error for valid model, got: %v", err)
		}
		if params == nil {
			t.Error("Expected parameters, got nil")
		}

		// Test invalid model
		_, err = service.GetModelParameters(ctx, "invalid-model")
		if err == nil {
			t.Error("Expected error for invalid model")
		}
	})

	t.Run("GetModelTokenLimits", func(t *testing.T) {
		// Test valid model
		contextWindow, maxOutput, err := service.GetModelTokenLimits(ctx, "gpt-4.1")
		if err != nil {
			t.Errorf("Expected no error for valid model, got: %v", err)
		}
		if contextWindow <= 0 || maxOutput <= 0 {
			t.Errorf("Expected positive token limits, got context=%d, output=%d", contextWindow, maxOutput)
		}

		// Test invalid model
		_, _, err = service.GetModelTokenLimits(ctx, "invalid-model")
		if err == nil {
			t.Error("Expected error for invalid model")
		}
	})
}

// TestOrchestratorConstructor tests the dependency injection helpers
func TestOrchestratorConstructor(t *testing.T) {
	// Test getting orchestrator constructor
	original := GetOrchestratorConstructor()
	if original == nil {
		t.Error("Expected non-nil orchestrator constructor")
	}

	// Test setting the constructor (just exercise the function)
	SetOrchestratorConstructor(original)

	// Verify we can get it back
	current := GetOrchestratorConstructor()
	if current == nil {
		t.Error("Expected constructor to be retrievable")
	}
}
