package llm

import (
	"errors"
	"strings"
	"testing"
)

// TestCreateStandardErrorWithMessage_DetailHandling tests how details are added to error messages
func TestCreateStandardErrorWithMessage_DetailHandling(t *testing.T) {
	t.Parallel() // Pure CPU-bound error message detail handling test
	original := errors.New("test error")
	provider := "test-provider"

	testCases := []struct {
		name     string
		category ErrorCategory
		details  string
	}{
		{
			name:     "empty details with auth error",
			category: CategoryAuth,
			details:  "",
		},
		{
			name:     "with details for auth error",
			category: CategoryAuth,
			details:  "invalid API key",
		},
		{
			name:     "with details for rate limit error",
			category: CategoryRateLimit,
			details:  "unique-detail-text-xyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llmErr := CreateStandardErrorWithMessage(provider, tc.category, original, tc.details)

			// Verify details field is set correctly
			if llmErr.Details != tc.details {
				t.Errorf("Expected Details = %q, got %q", tc.details, llmErr.Details)
			}

			// If details are provided, verify they're included in the message
			// Either directly or in parentheses
			if tc.details != "" {
				if !strings.Contains(llmErr.Message, tc.details) {
					t.Errorf("Expected message to contain details %q somewhere in %q", tc.details, llmErr.Message)
				}
			}
		})
	}
}
