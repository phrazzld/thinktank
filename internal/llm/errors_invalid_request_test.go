package llm

import (
	"errors"
	"strings"
	"testing"
)

// TestCreateStandardErrorWithMessage_InvalidRequest specifically tests the InvalidRequest case
func TestCreateStandardErrorWithMessage_InvalidRequest(t *testing.T) {
	original := errors.New("test error")
	provider := "test-provider"
	details := "test details"
	category := CategoryInvalidRequest

	llmErr := CreateStandardErrorWithMessage(provider, category, original, details)

	// Verify basic properties
	if llmErr.Provider != provider {
		t.Errorf("Expected Provider = %q, got %q", provider, llmErr.Provider)
	}
	if llmErr.Original != original {
		t.Errorf("Expected Original = %v, got %v", original, llmErr.Original)
	}
	if llmErr.ErrorCategory != category {
		t.Errorf("Expected ErrorCategory = %v, got %v", category, llmErr.ErrorCategory)
	}
	if llmErr.Details != details {
		t.Errorf("Expected Details = %q, got %q", details, llmErr.Details)
	}

	// Verify message content for this specific category
	expectedMsgPart := "Invalid request"
	if !strings.Contains(llmErr.Message, expectedMsgPart) {
		t.Errorf("Expected message to contain %q, got %q", expectedMsgPart, llmErr.Message)
	}

	// Verify suggestion content for this specific category
	expectedSuggestionPart := "Check the prompt format"
	if !strings.Contains(llmErr.Suggestion, expectedSuggestionPart) {
		t.Errorf("Expected suggestion to contain %q, got %q", expectedSuggestionPart, llmErr.Suggestion)
	}
}
