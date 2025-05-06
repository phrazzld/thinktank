package thinktank

import (
	"errors"
	"fmt"
	"testing"
)

// TestWrapOrchestratorErrors tests the wrapOrchestratorErrors function
// which converts orchestrator-specific errors to thinktank package errors
func TestWrapOrchestratorErrors(t *testing.T) {
	testCases := []struct {
		name            string
		inputError      error
		expectedWrapped bool
		expectedError   error
	}{
		{
			name:            "nil error should remain nil",
			inputError:      nil,
			expectedWrapped: false,
			expectedError:   nil,
		},
		{
			name:            "regular error should remain unchanged",
			inputError:      errors.New("generic error"),
			expectedWrapped: false,
			expectedError:   errors.New("generic error"),
		},
		{
			name:            "ErrPartialSuccess should remain unchanged",
			inputError:      fmt.Errorf("wrapped: %w", ErrPartialSuccess),
			expectedWrapped: true,
			expectedError:   ErrPartialSuccess,
		},
		{
			name:            "error with 'some models failed' text should be wrapped with ErrPartialSuccess",
			inputError:      errors.New("some models failed during processing: error details"),
			expectedWrapped: true,
			expectedError:   ErrPartialSuccess,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := wrapOrchestratorErrors(tc.inputError)

			// Check if result is nil when expected
			if tc.inputError == nil && result != nil {
				t.Errorf("Expected nil error, but got: %v", result)
				return
			}

			// For non-nil errors, check the wrapping
			if tc.inputError != nil {
				if tc.expectedWrapped {
					if !errors.Is(result, tc.expectedError) {
						t.Errorf("Expected error to be wrapped with %v, but it wasn't: %v", tc.expectedError, result)
					}
				} else {
					// For non-wrapped errors, just check the error message
					if result.Error() != tc.inputError.Error() {
						t.Errorf("Expected error message '%v', got '%v'", tc.inputError.Error(), result.Error())
					}
				}
			}
		})
	}
}
