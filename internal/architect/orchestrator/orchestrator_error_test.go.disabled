package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/fileutil"
)

// TestRun_GatherContextError tests handling of errors from context gathering
func TestRun_GatherContextError(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	modelNames := []string{"model1"}
	deps.setupMultiModelConfig(modelNames)

	// Setup a context gatherer that returns an error
	expectedErr := errors.New("gather context error")
	// Use proper imports for interfaces package and fileutil
	deps.contextGatherer.GatherContextFunc = func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
		return nil, nil, expectedErr
	}

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)

	// Verify that the error is returned
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	// The error should contain the original error message
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr.Error(), err)
	}

	// Verify no calls were made to initialize clients or write files
	if len(deps.apiService.InitLLMClientCalls) > 0 {
		t.Errorf("Expected no InitLLMClient calls, got %d", len(deps.apiService.InitLLMClientCalls))
	}
	if len(deps.fileWriter.SaveToFileCalls) > 0 {
		t.Errorf("Expected no SaveToFile calls, got %d", len(deps.fileWriter.SaveToFileCalls))
	}
}

// TestAggregateAndFormatErrors tests error aggregation and formatting
func TestAggregateAndFormatErrors(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		modelErrors   []error
		expectedParts []string // Strings that should appear in the result
	}{
		{
			name:          "no errors",
			modelErrors:   []error{},
			expectedParts: []string{},
		},
		{
			name: "single error",
			modelErrors: []error{
				errors.New("test error"),
			},
			expectedParts: []string{
				"test error",
			},
		},
		{
			name: "multiple errors",
			modelErrors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			},
			expectedParts: []string{
				"error 1",
				"error 2",
				"error 3",
			},
		},
		{
			name: "API safety error",
			modelErrors: []error{
				errors.New("safety blocked content"),
			},
			expectedParts: []string{
				"safety blocked content",
			},
		},
		{
			name: "empty response error",
			modelErrors: []error{
				errors.New("empty response received"),
			},
			expectedParts: []string{
				"empty response received",
			},
		},
		{
			name: "rate limit error",
			modelErrors: []error{
				errors.New("rate limit exceeded"),
			},
			expectedParts: []string{
				"rate limit",
			},
		},
	}

	// Setup a mock API service for error testing
	apiSvc := &mockAPIService{
		IsSafetyBlockedErrorFunc: func(err error) bool {
			return err != nil && strings.Contains(err.Error(), "safety")
		},
		IsEmptyResponseErrorFunc: func(err error) bool {
			return err != nil && strings.Contains(err.Error(), "empty")
		},
		GetErrorDetailsFunc: func(err error) string {
			if err == nil {
				return ""
			}
			return "Details: " + err.Error()
		},
	}

	// Create orchestrator with the mock API service
	logger := &mockLogger{}
	orch := &Orchestrator{
		apiService: apiSvc,
		logger:     logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// If no errors, should have no result
			if len(tc.modelErrors) == 0 {
				if result := orch.aggregateAndFormatErrors(tc.modelErrors); result != nil {
					t.Errorf("Expected nil result for no errors, got: %v", result)
				}
				return
			}

			// Otherwise, check the formatted error
			result := orch.aggregateAndFormatErrors(tc.modelErrors)
			if result == nil {
				t.Fatal("Expected non-nil error result, got nil")
			}

			// Check that the result contains expected parts
			errStr := result.Error()
			for _, part := range tc.expectedParts {
				if !strings.Contains(errStr, part) {
					t.Errorf("Expected error to contain '%s', but it doesn't. Error: %s", part, errStr)
				}
			}
		})
	}
}
