package llm

import (
	"errors"
	"fmt"
	"testing"
)

// Test the String method of ErrorCategory
func TestErrorCategory_String(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected string
	}{
		{CategoryUnknown, "Unknown"},
		{CategoryAuth, "Auth"},
		{CategoryRateLimit, "RateLimit"},
		{CategoryInvalidRequest, "InvalidRequest"},
		{CategoryNotFound, "NotFound"},
		{CategoryServer, "Server"},
		{CategoryNetwork, "Network"},
		{CategoryCancelled, "Cancelled"},
		{CategoryInputLimit, "InputLimit"},
		{CategoryContentFiltered, "ContentFiltered"},
		{ErrorCategory(99), "Unknown"}, // Unknown value should return "Unknown"
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.category.String()
			if result != tc.expected {
				t.Errorf("Expected ErrorCategory(%d).String() to be %q, got %q", tc.category, tc.expected, result)
			}
		})
	}
}

// Simple error type implementing CategorizedError for testing
type testCategorizedError struct {
	msg      string
	category ErrorCategory
}

func (e testCategorizedError) Error() string {
	return e.msg
}

func (e testCategorizedError) Category() ErrorCategory {
	return e.category
}

// Test that IsCategorizedError correctly identifies a CategorizedError
func TestIsCategorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		isCatErr bool
		category ErrorCategory
	}{
		{
			name:     "nil error",
			err:      nil,
			isCatErr: false,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			isCatErr: false,
		},
		{
			name:     "categorized error",
			err:      testCategorizedError{msg: "test error", category: CategoryRateLimit},
			isCatErr: true,
			category: CategoryRateLimit,
		},
		{
			name:     "wrapped categorized error",
			err:      fmt.Errorf("wrapped: %w", testCategorizedError{msg: "inner error", category: CategoryAuth}),
			isCatErr: true,
			category: CategoryAuth,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			catErr, ok := IsCategorizedError(tc.err)
			if ok != tc.isCatErr {
				t.Errorf("IsCategorizedError(%v) = %v, want %v", tc.err, ok, tc.isCatErr)
			}

			if ok && catErr.Category() != tc.category {
				t.Errorf("Category() = %v, want %v", catErr.Category(), tc.category)
			}
		})
	}
}
