package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestUserFacingError tests the UserFacingError method of LLMError
func TestUserFacingError(t *testing.T) {
	tests := []struct {
		name          string
		llmErr        *llm.LLMError
		shouldContain []string
	}{
		{
			name: "Error with original error",
			llmErr: &llm.LLMError{
				Message:    "Authentication failed",
				Original:   errors.New("API key expired"),
				Suggestion: "Renew your API key",
			},
			shouldContain: []string{
				"Authentication failed: API key expired",
				"Suggestion: Renew your API key",
			},
		},
		{
			name: "Error without original error",
			llmErr: &llm.LLMError{
				Message:    "Rate limit exceeded",
				Suggestion: "Try again later",
			},
			shouldContain: []string{
				"Rate limit exceeded",
				"Suggestion: Try again later",
			},
		},
		{
			name: "Error without suggestion",
			llmErr: &llm.LLMError{
				Message:  "Network error",
				Original: errors.New("connection timeout"),
			},
			shouldContain: []string{
				"Network error: connection timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.llmErr.UserFacingError()

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected UserFacingError() to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

// TestDebugInfo tests that LLMError.DebugInfo returns all expected details
func TestDebugInfo(t *testing.T) {
	t.Parallel()
	originalErr := errors.New("original error")
	llmErr := &llm.LLMError{
		Provider:      "test-provider",
		Code:          "TEST_ERR_CODE",
		StatusCode:    429,
		Message:       "Rate limit exceeded",
		RequestID:     "req-12345",
		Original:      originalErr,
		ErrorCategory: llm.CategoryRateLimit,
		Suggestion:    "Try again later",
		Details:       "Additional details about the error",
	}

	debugInfo := llmErr.DebugInfo()

	// Check that all fields are included in the debug info
	expectedFields := []string{
		"Provider: test-provider",
		"Error Category: RateLimit",
		"Message: Rate limit exceeded",
		"Error Code: TEST_ERR_CODE",
		"Status Code: 429",
		"Request ID: req-12345",
		"Original Error: original error",
		"Details: Additional details about the error",
		"Suggestion: Try again later",
	}

	for _, field := range expectedFields {
		if !strings.Contains(debugInfo, field) {
			t.Errorf("Expected debug info to contain %q, but it didn't.\nDebug info: %s", field, debugInfo)
		}
	}
}
