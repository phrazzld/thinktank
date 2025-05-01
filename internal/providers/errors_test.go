// Package providers defines the interfaces and implementations for LLM provider adapters
package providers

import (
	"errors"
	"testing"
)

// TestProviderErrors verifies that the sentinel errors are defined correctly
func TestProviderErrors(t *testing.T) {
	// Check that all sentinel errors are defined
	testCases := []struct {
		name  string
		err   error
		check func(error) bool
	}{
		{
			name: "ErrProviderNotFound",
			err:  ErrProviderNotFound,
			check: func(err error) bool {
				return errors.Is(err, ErrProviderNotFound)
			},
		},
		{
			name: "ErrInvalidAPIKey",
			err:  ErrInvalidAPIKey,
			check: func(err error) bool {
				return errors.Is(err, ErrInvalidAPIKey)
			},
		},
		{
			name: "ErrInvalidModelID",
			err:  ErrInvalidModelID,
			check: func(err error) bool {
				return errors.Is(err, ErrInvalidModelID)
			},
		},
		{
			name: "ErrInvalidEndpoint",
			err:  ErrInvalidEndpoint,
			check: func(err error) bool {
				return errors.Is(err, ErrInvalidEndpoint)
			},
		},
		{
			name: "ErrClientCreation",
			err:  ErrClientCreation,
			check: func(err error) bool {
				return errors.Is(err, ErrClientCreation)
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check that error is defined
			if tc.err == nil {
				t.Errorf("Error %s is nil", tc.name)
			}

			// Check that error message is not empty
			if tc.err.Error() == "" {
				t.Errorf("Error %s has empty message", tc.name)
			}

			// Test wrapping the error
			wrappedWithIs := errors.New("wrapped with is")
			wrappedErr := errors.Join(wrappedWithIs, tc.err)

			// Check that errors.Is works with the wrapped error
			if !tc.check(wrappedErr) {
				t.Errorf("errors.Is check failed for wrapped %s", tc.name)
			}
		})
	}
}
