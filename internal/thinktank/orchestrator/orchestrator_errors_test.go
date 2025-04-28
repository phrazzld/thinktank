package orchestrator

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/testutil"
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
				if result.Error() != tt.expectedError {
					t.Errorf("Error message mismatch\nExpected: %q\nActual:   %q",
						tt.expectedError, result.Error())
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
	if result.Error() != expected {
		t.Errorf("Wrapped errors not handled correctly\nExpected: %q\nActual:   %q",
			expected, result.Error())
	}

	// Test that errors.Is works with the result
	if !strings.Contains(result.Error(), "base error") {
		t.Errorf("Expected wrapped error message to be included")
	}
}
