package orchestrator

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/testutil"
)

// TestAggregateErrors tests the aggregateErrors method of the Orchestrator
func TestAggregateErrors(t *testing.T) {
	// Create a minimal orchestrator with just the required dependencies for these tests
	logger := testutil.NewMockLogger()
	config := &config.CliConfig{}
	o := &Orchestrator{
		logger: logger,
		config: config,
	}

	// Create some sample errors for testing
	err1 := errors.New("model1 failed")
	err2 := errors.New("model2 failed with different reason")
	err3 := errors.New("model3 connection timeout")

	tests := []struct {
		name          string
		errs          []error
		totalCount    int
		successCount  int
		expectError   bool
		expectedError string
	}{
		{
			name:          "empty error list",
			errs:          []error{},
			totalCount:    3,
			successCount:  3,
			expectError:   false,
			expectedError: "",
		},
		{
			name:          "all models failed",
			errs:          []error{err1, err2, err3},
			totalCount:    3,
			successCount:  0,
			expectError:   true,
			expectedError: "all models failed: model1 failed; model2 failed with different reason; model3 connection timeout",
		},
		{
			name:          "partial failure - some models succeeded",
			errs:          []error{err1, err2},
			totalCount:    3,
			successCount:  1,
			expectError:   true,
			expectedError: "processed 1/3 models successfully; 2 failed: model1 failed; model2 failed with different reason",
		},
		{
			name:          "one error",
			errs:          []error{err1},
			totalCount:    3,
			successCount:  2,
			expectError:   true,
			expectedError: "processed 2/3 models successfully; 1 failed: model1 failed",
		},
		{
			name:         "mixed nil and non-nil errors",
			errs:         []error{err1, nil, err3},
			totalCount:   4,
			successCount: 2,
			expectError:  true,
			// The error count reflects len(errs), even though nil errors are filtered in the message
			expectedError: "processed 2/4 models successfully; 3 failed: model1 failed; model3 connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.aggregateErrors(tt.errs, tt.totalCount, tt.successCount)

			if tt.expectError && result == nil {
				t.Fatalf("Expected error but got nil")
			}

			if !tt.expectError && result != nil {
				t.Fatalf("Expected nil error but got: %v", result)
			}

			if tt.expectError {
				// Check for sentinel errors and wrapped message content
				if !strings.Contains(result.Error(), tt.expectedError) {
					t.Errorf("Error message mismatch\nExpected to contain: %q\nActual: %q",
						tt.expectedError, result.Error())
				}

				// Verify that appropriate sentinel error is used
				if len(tt.errs) > 0 && tt.successCount == 0 {
					if !errors.Is(result, ErrAllProcessingFailed) {
						t.Errorf("Expected ErrAllProcessingFailed sentinel error, got: %v", result)
					}
				} else if len(tt.errs) > 0 {
					if !errors.Is(result, ErrPartialProcessingFailure) {
						t.Errorf("Expected ErrPartialProcessingFailure sentinel error, got: %v", result)
					}
				}
			}
		})
	}
}

// TestAggregateErrorMessages tests the aggregateErrorMessages helper function
func TestAggregateErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		errs     []error
		expected string
	}{
		{
			name:     "empty error list",
			errs:     []error{},
			expected: "",
		},
		{
			name:     "single error",
			errs:     []error{errors.New("error1")},
			expected: "error1",
		},
		{
			name: "multiple errors",
			errs: []error{
				errors.New("error1"),
				errors.New("error2"),
				errors.New("error3"),
			},
			expected: "error1; error2; error3",
		},
		{
			name: "with nil errors",
			errs: []error{
				errors.New("error1"),
				nil,
				errors.New("error3"),
				nil,
			},
			expected: "error1; error3",
		},
		{
			name:     "all nil errors",
			errs:     []error{nil, nil, nil},
			expected: "",
		},
		{
			name: "errors with semicolons in messages",
			errs: []error{
				errors.New("error1; part2"),
				errors.New("error2"),
			},
			expected: "error1; part2; error2",
		},
		{
			name: "formatted errors",
			errs: []error{
				fmt.Errorf("formatted error: %s", "details"),
				errors.New("plain error"),
			},
			expected: "formatted error: details; plain error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aggregateErrorMessages(tt.errs)

			if result != tt.expected {
				t.Errorf("Result mismatch\nExpected: %q\nActual:   %q",
					tt.expected, result)
			}
		})
	}
}

// TestAggregateErrorsWrappedErrors tests aggregateErrors with wrapped errors
func TestAggregateErrorsWrappedErrors(t *testing.T) {
	logger := testutil.NewMockLogger()
	config := &config.CliConfig{}
	o := &Orchestrator{
		logger: logger,
		config: config,
	}

	// Create some wrapped errors
	baseErr := errors.New("base error")
	wrappedErr1 := fmt.Errorf("model1 failed: %w", baseErr)
	wrappedErr2 := fmt.Errorf("model2 failed: %w", errors.New("internal error"))

	// Test that wrapped errors are properly included in the error message
	result := o.aggregateErrors(
		[]error{wrappedErr1, wrappedErr2},
		3,
		1,
	)

	expected := "processed 1/3 models successfully; 2 failed: model1 failed: base error; model2 failed: internal error"
	if !strings.Contains(result.Error(), expected) {
		t.Errorf("Wrapped errors not handled correctly\nExpected to contain: %q\nActual: %q",
			expected, result.Error())
	}

	// Check for sentinel error
	if !errors.Is(result, ErrPartialProcessingFailure) {
		t.Errorf("Expected ErrPartialProcessingFailure sentinel error, got: %v", result)
	}

	// Test that errors.Is works with the result
	if !strings.Contains(result.Error(), "base error") {
		t.Errorf("Expected wrapped error message to be included")
	}
}

// TestGetUserFriendlyErrorMessage tests the getUserFriendlyErrorMessage method
func TestGetUserFriendlyErrorMessage(t *testing.T) {
	logger := testutil.NewMockLogger()
	config := &config.CliConfig{}
	o := &Orchestrator{
		logger: logger,
		config: config,
	}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "LLMError with CategoryAuth",
			err:      &llm.LLMError{Original: errors.New("auth failed"), ErrorCategory: llm.CategoryAuth},
			expected: "authentication failed",
		},
		{
			name:     "LLMError with CategoryRateLimit",
			err:      &llm.LLMError{Original: errors.New("rate limit exceeded"), ErrorCategory: llm.CategoryRateLimit},
			expected: "rate limited",
		},
		{
			name:     "LLMError with CategoryInvalidRequest and context keyword",
			err:      &llm.LLMError{Original: errors.New("context window exceeded"), ErrorCategory: llm.CategoryInvalidRequest},
			expected: "input too large",
		},
		{
			name:     "LLMError with CategoryInvalidRequest and token keyword",
			err:      &llm.LLMError{Original: errors.New("token limit exceeded"), ErrorCategory: llm.CategoryInvalidRequest},
			expected: "input too large",
		},
		{
			name:     "LLMError with CategoryInvalidRequest and length keyword",
			err:      &llm.LLMError{Original: errors.New("length too long"), ErrorCategory: llm.CategoryInvalidRequest},
			expected: "input too large",
		},
		{
			name:     "LLMError with CategoryInvalidRequest without size keywords",
			err:      &llm.LLMError{Original: errors.New("invalid parameter"), ErrorCategory: llm.CategoryInvalidRequest},
			expected: "invalid request",
		},
		{
			name:     "LLMError with CategoryInsufficientCredits",
			err:      &llm.LLMError{Original: errors.New("insufficient credits"), ErrorCategory: llm.CategoryInsufficientCredits},
			expected: "insufficient credits",
		},
		{
			name:     "LLMError with CategoryServer (default case)",
			err:      &llm.LLMError{Original: errors.New("server error"), ErrorCategory: llm.CategoryServer},
			expected: "error",
		},
		{
			name:     "LLMError with CategoryUnknown (default case)",
			err:      &llm.LLMError{Original: errors.New("unknown error"), ErrorCategory: llm.CategoryUnknown},
			expected: "error",
		},
		{
			name:     "Non-LLMError",
			err:      errors.New("regular error"),
			expected: "error",
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.getUserFriendlyErrorMessage(tt.err, "test-model")
			if result != tt.expected {
				t.Errorf("getUserFriendlyErrorMessage() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestCategorizeOrchestratorError tests the CategorizeOrchestratorError function
func TestCategorizeOrchestratorError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected llm.ErrorCategory
	}{
		{
			name:     "ErrInvalidSynthesisModel",
			err:      ErrInvalidSynthesisModel,
			expected: llm.CategoryInvalidRequest,
		},
		{
			name:     "ErrNoValidModels",
			err:      ErrNoValidModels,
			expected: llm.CategoryInvalidRequest,
		},
		{
			name:     "ErrPartialProcessingFailure",
			err:      ErrPartialProcessingFailure,
			expected: llm.CategoryServer,
		},
		{
			name:     "ErrAllProcessingFailed",
			err:      ErrAllProcessingFailed,
			expected: llm.CategoryServer,
		},
		{
			name:     "ErrSynthesisFailed",
			err:      ErrSynthesisFailed,
			expected: llm.CategoryServer,
		},
		{
			name:     "ErrOutputFileSaveFailed",
			err:      ErrOutputFileSaveFailed,
			expected: llm.CategoryServer,
		},
		{
			name:     "ErrModelProcessingCancelled",
			err:      ErrModelProcessingCancelled,
			expected: llm.CategoryCancelled,
		},
		{
			name:     "Wrapped ErrInvalidSynthesisModel",
			err:      fmt.Errorf("wrapper: %w", ErrInvalidSynthesisModel),
			expected: llm.CategoryInvalidRequest,
		},
		{
			name:     "Wrapped ErrPartialProcessingFailure",
			err:      fmt.Errorf("wrapper: %w", ErrPartialProcessingFailure),
			expected: llm.CategoryServer,
		},
		{
			name:     "Already categorized LLMError",
			err:      &llm.LLMError{Original: errors.New("auth failed"), ErrorCategory: llm.CategoryAuth},
			expected: llm.CategoryAuth,
		},
		{
			name:     "Already categorized LLMError with different category",
			err:      &llm.LLMError{Original: errors.New("rate limit"), ErrorCategory: llm.CategoryRateLimit},
			expected: llm.CategoryRateLimit,
		},
		{
			name:     "Unknown error",
			err:      errors.New("unknown error"),
			expected: llm.CategoryUnknown,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: llm.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeOrchestratorError(tt.err)
			if result != tt.expected {
				t.Errorf("CategorizeOrchestratorError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
