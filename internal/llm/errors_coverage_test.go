package llm

import (
	"errors"
	"strings"
	"testing"
)

// Test Error method with nil Original error
func TestLLMError_Error_NilOriginal(t *testing.T) {
	t.Parallel()
	llmErr := &LLMError{
		Provider:      "test-provider",
		Message:       "Error message with no original",
		ErrorCategory: CategoryUnknown,
	}

	expected := "Error message with no original"
	if result := llmErr.Error(); result != expected {
		t.Errorf("Expected Error() = %q, got %q", expected, result)
	}
}

// Additional test for UserFacingError with nil Original error
func TestLLMError_UserFacingError_NilOriginal(t *testing.T) {
	t.Parallel()
	llmErr := &LLMError{
		Provider:      "test-provider",
		Message:       "User message with no original",
		ErrorCategory: CategoryUnknown,
		Suggestion:    "Try this instead",
	}

	expected := "User message with no original\n\nSuggestion: Try this instead"
	if result := llmErr.UserFacingError(); result != expected {
		t.Errorf("Expected UserFacingError() = %q, got %q", expected, result)
	}
}

// Additional test for UserFacingError with no suggestion
func TestLLMError_UserFacingError_NoSuggestion(t *testing.T) {
	t.Parallel()
	origErr := errors.New("original error")
	llmErr := &LLMError{
		Provider:      "test-provider",
		Message:       "Error message",
		Original:      origErr,
		ErrorCategory: CategoryUnknown,
		// No suggestion provided
	}

	expected := "Error message: original error"
	if result := llmErr.UserFacingError(); result != expected {
		t.Errorf("Expected UserFacingError() = %q, got %q", expected, result)
	}
}

// Comprehensive test for CreateStandardErrorWithMessage
func TestCreateStandardErrorWithMessage_Comprehensive(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		provider       string
		category       ErrorCategory
		details        string
		expectContains []string
	}{
		{
			name:     "invalid request",
			provider: "test-provider",
			category: CategoryInvalidRequest,
			details:  "invalid parameters",
			expectContains: []string{
				"Invalid request",
				"test-provider",
				"invalid parameters",
				"Check the prompt format",
			},
		},
		{
			name:     "not found",
			provider: "test-provider",
			category: CategoryNotFound,
			details:  "",
			expectContains: []string{
				"not found",
				"Verify that the model name is correct",
			},
		},
		{
			name:     "server error",
			provider: "test-provider",
			category: CategoryServer,
			details:  "internal server error",
			expectContains: []string{
				"server error",
				"test-provider",
				"internal server error",
				"temporary issue",
			},
		},
		{
			name:     "network error",
			provider: "test-provider",
			category: CategoryNetwork,
			details:  "timeout",
			expectContains: []string{
				"Network error",
				"test-provider",
				"timeout",
				"Check your internet connection",
			},
		},
		{
			name:     "cancelled",
			provider: "test-provider",
			category: CategoryCancelled,
			details:  "deadline exceeded",
			expectContains: []string{
				"cancelled",
				"test-provider",
				"deadline exceeded",
				"longer timeout",
			},
		},
		{
			name:     "input limit",
			provider: "test-provider",
			category: CategoryInputLimit,
			details:  "token count 10000",
			expectContains: []string{
				"token limit",
				"token count 10000",
				"Reduce the input size",
			},
		},
		{
			name:     "content filtered",
			provider: "test-provider",
			category: CategoryContentFiltered,
			details:  "unsafe content detected",
			expectContains: []string{
				"Content was filtered",
				"unsafe content detected",
				"safety filters",
			},
		},
	}

	origErr := errors.New("original error")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llmErr := CreateStandardErrorWithMessage(tc.provider, tc.category, origErr, tc.details)

			// Verify basic fields
			if llmErr.Provider != tc.provider {
				t.Errorf("Expected Provider = %q, got %q", tc.provider, llmErr.Provider)
			}
			if llmErr.ErrorCategory != tc.category {
				t.Errorf("Expected ErrorCategory = %v, got %v", tc.category, llmErr.ErrorCategory)
			}
			if llmErr.Original != origErr {
				t.Errorf("Expected Original = %v, got %v", origErr, llmErr.Original)
			}
			if tc.details != "" && llmErr.Details != tc.details {
				t.Errorf("Expected Details = %q, got %q", tc.details, llmErr.Details)
			}

			// Check for expected content in message and suggestion
			for _, expected := range tc.expectContains {
				msgContains := false
				suggContains := false

				if llmErr.Message != "" && llmErr.Message != tc.provider {
					msgContains = contains(llmErr.Message, expected)
				}
				if llmErr.Suggestion != "" {
					suggContains = contains(llmErr.Suggestion, expected)
				}

				if !msgContains && !suggContains {
					t.Errorf("Neither Message nor Suggestion contains %q\nMessage: %q\nSuggestion: %q",
						expected, llmErr.Message, llmErr.Suggestion)
				}
			}
		})
	}
}

// Test MockError implementation
func TestMockError(t *testing.T) {
	t.Parallel()
	mockErr := &MockError{
		Message:       "mock error",
		ErrorCategory: CategoryAuth,
	}

	// Test Error() method
	expected := "mock error"
	if result := mockErr.Error(); result != expected {
		t.Errorf("Expected Error() = %q, got %q", expected, result)
	}

	// Test Category() method
	if cat := mockErr.Category(); cat != CategoryAuth {
		t.Errorf("Expected Category() = %v, got %v", CategoryAuth, cat)
	}

	// Verify MockError implements CategorizedError
	var catError CategorizedError = mockErr
	if catError.Category() != CategoryAuth {
		t.Errorf("MockError does not properly implement CategorizedError")
	}
}

// Helper function to check if a string contains another string (case-insensitive)
func contains(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	// Convert both to lowercase for case-insensitive comparison
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
