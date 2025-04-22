package gemini

import (
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/gemini"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
)

// TestErrorDebugInfo tests the debug information in errors
func TestErrorDebugInfo(t *testing.T) {
	// Test LLMError debug info for Gemini errors
	testCases := []struct {
		name            string
		category        llm.ErrorCategory
		message         string
		code            string
		statusCode      int
		requestID       string
		details         string
		suggestion      string
		expectedInfoStr []string
	}{
		{
			name:       "Auth error with full details",
			category:   llm.CategoryAuth,
			message:    "Authentication failed",
			code:       "invalid_api_key",
			statusCode: 401,
			requestID:  "req-12345",
			details:    "API key invalid",
			suggestion: "Check your API key",
			expectedInfoStr: []string{
				"Provider: gemini",
				"Error Category: Auth",
				"Message: Authentication failed",
				"Error Code: invalid_api_key",
				"Status Code: 401",
				"Request ID: req-12345",
				"Details: API key invalid",
				"Suggestion: Check your API key",
			},
		},
		{
			name:       "Rate limit error with minimal details",
			category:   llm.CategoryRateLimit,
			message:    "Rate limit exceeded",
			statusCode: 429,
			expectedInfoStr: []string{
				"Provider: gemini",
				"Error Category: RateLimit",
				"Message: Rate limit exceeded",
				"Status Code: 429",
			},
		},
		{
			name:       "Safety filter error",
			category:   llm.CategoryContentFiltered,
			message:    "Content filtered by safety settings",
			statusCode: 200,
			details:    "finishReason: SAFETY",
			suggestion: "Consider modifying your content",
			expectedInfoStr: []string{
				"Provider: gemini",
				"Error Category: ContentFiltered",
				"Message: Content filtered by safety settings",
				"Details: finishReason: SAFETY",
				"Suggestion: Consider modifying your content",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create LLMError
			llmErr := &llm.LLMError{
				Provider:      "gemini",
				Code:          tc.code,
				StatusCode:    tc.statusCode,
				Message:       tc.message,
				RequestID:     tc.requestID,
				Original:      fmt.Errorf("original error"),
				ErrorCategory: tc.category,
				Suggestion:    tc.suggestion,
				Details:       tc.details,
			}

			// Get debug info
			debugInfo := llmErr.DebugInfo()

			// Check all expected strings are in the debug info
			for _, expectedStr := range tc.expectedInfoStr {
				assert.Contains(t, debugInfo, expectedStr,
					"Expected debug info to contain %q, but it was not found", expectedStr)
			}
		})
	}
}

// TestUserFacingError tests the user-facing error messages
func TestUserFacingError(t *testing.T) {
	// Test LLMError user-facing messages
	testCases := []struct {
		name              string
		category          llm.ErrorCategory
		message           string
		suggestion        string
		expectedUserFaced []string
		notExpectedInUser []string
	}{
		{
			name:       "Auth error with user suggestion",
			category:   llm.CategoryAuth,
			message:    "Authentication failed: invalid API key",
			suggestion: "Check your API key and ensure it has not expired",
			expectedUserFaced: []string{
				"Authentication failed",
				"Check your API key",
			},
			notExpectedInUser: []string{
				"Debug", "Provider:", "Error Category:",
			},
		},
		{
			name:       "Safety filter error with user advice",
			category:   llm.CategoryContentFiltered,
			message:    "Content filtered by Gemini's safety settings",
			suggestion: "Try rephrasing your prompt to avoid triggering safety filters",
			expectedUserFaced: []string{
				"Content filtered",
				"Try rephrasing your prompt",
			},
			notExpectedInUser: []string{
				"Original Error:", "Status Code:", "Details:",
			},
		},
		{
			name:       "Network error with troubleshooting steps",
			category:   llm.CategoryNetwork,
			message:    "Network error: connection refused",
			suggestion: "Check your internet connection and verify that Gemini API is available",
			expectedUserFaced: []string{
				"Network error",
				"Check your internet connection",
			},
			notExpectedInUser: []string{
				"Original Error:", "Status Code:", "Debug:",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create error through API function
			llmErr := gemini.CreateAPIError(
				tc.category,
				tc.message,
				fmt.Errorf("original error"),
				"",
			)

			// Set suggestion directly to ensure we test with exact text
			llmErr.Suggestion = tc.suggestion

			// Get user-facing message
			userMsg := llmErr.UserFacingError()

			// Check all expected strings are in the user message
			for _, expectedStr := range tc.expectedUserFaced {
				assert.Contains(t, userMsg, expectedStr,
					"Expected user message to contain %q, but it was not found", expectedStr)
			}

			// Check that debug info is not included in user message
			for _, notExpectedStr := range tc.notExpectedInUser {
				assert.NotContains(t, userMsg, notExpectedStr,
					"User message should not contain %q, but it was found", notExpectedStr)
			}
		})
	}
}
