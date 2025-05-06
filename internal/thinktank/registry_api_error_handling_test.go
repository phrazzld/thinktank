// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestProcessLLMResponse tests the ProcessLLMResponse method
func TestProcessLLMResponse(t *testing.T) {
	service, _, _ := setupTest(t)

	testCases := []struct {
		name            string
		result          *llm.ProviderResult
		expectedContent string
		expectError     bool
		expectedError   error
		errorSubstring  string // To verify specific error message content
	}{
		{
			name:            "nil result",
			result:          nil,
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
			errorSubstring:  "result is nil",
		},
		{
			name: "empty content with no details",
			result: &llm.ProviderResult{
				Content: "",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
		},
		{
			name: "empty content with finish reason",
			result: &llm.ProviderResult{
				Content:      "",
				FinishReason: "stop",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
			errorSubstring:  "Finish Reason: stop",
		},
		{
			name: "empty content with finish reason length",
			result: &llm.ProviderResult{
				Content:      "",
				FinishReason: "length",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
			errorSubstring:  "Finish Reason: length",
		},
		{
			name: "empty content with single safety block",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{
						Category: "harmful_content",
						Blocked:  true,
					},
				},
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrSafetyBlocked,
			errorSubstring:  "Blocked by Safety Category: harmful_content",
		},
		{
			name: "empty content with multiple safety blocks",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{
						Category: "violence",
						Blocked:  true,
					},
					{
						Category: "hate_speech",
						Blocked:  true,
					},
					{
						Category: "not_blocked_category",
						Blocked:  false,
					},
				},
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrSafetyBlocked,
			errorSubstring:  "Safety Blocking:",
		},
		{
			name: "empty content with finish reason and safety blocks",
			result: &llm.ProviderResult{
				Content:      "",
				FinishReason: "safety",
				SafetyInfo: []llm.Safety{
					{
						Category: "unsafe_content",
						Blocked:  true,
					},
				},
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrSafetyBlocked,
			errorSubstring:  "Finish Reason: safety",
		},
		{
			name: "non-blocking safety info",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{
						Category: "flagged_but_not_blocked",
						Blocked:  false,
					},
				},
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
		},
		{
			name: "whitespace only content",
			result: &llm.ProviderResult{
				Content: "   \n   ",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrWhitespaceContent,
		},
		{
			name: "whitespace with tabs and newlines",
			result: &llm.ProviderResult{
				Content: "\t\n \t\r\n",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrWhitespaceContent,
		},
		{
			name: "valid content",
			result: &llm.ProviderResult{
				Content: "This is a valid response",
			},
			expectedContent: "This is a valid response",
			expectError:     false,
		},
		{
			name: "valid content with safety info (not blocked)",
			result: &llm.ProviderResult{
				Content: "This is a valid response with safety info",
				SafetyInfo: []llm.Safety{
					{
						Category: "flagged_but_not_blocked",
						Blocked:  false,
					},
				},
			},
			expectedContent: "This is a valid response with safety info",
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := service.ProcessLLMResponse(tc.result)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}

				// Check that error is of expected type
				if !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error to be '%v', got '%v'", tc.expectedError, err)
				}

				// If an error substring is specified, check it's in the error message
				if tc.errorSubstring != "" && !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error message to contain '%s', got '%s'",
						tc.errorSubstring, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

				// Verify the content matches expected
				if content != tc.expectedContent {
					t.Errorf("Expected content to be '%s', got '%s'", tc.expectedContent, content)
				}
			}
		})
	}
}

// TestErrorClassificationMethods tests the error classification methods
func TestErrorClassificationMethods(t *testing.T) {
	service, _, _ := setupTest(t)

	// Test IsEmptyResponseError
	t.Run("IsEmptyResponseError", func(t *testing.T) {
		testCases := []struct {
			name     string
			err      error
			expected bool
		}{
			// Common error cases
			{
				name:     "nil error",
				err:      nil,
				expected: false,
			},
			{
				name:     "unrelated error",
				err:      errors.New("some other error"),
				expected: false,
			},

			// Specific error types
			{
				name:     "empty response error sentinel",
				err:      llm.ErrEmptyResponse,
				expected: true,
			},
			{
				name:     "whitespace content error sentinel",
				err:      llm.ErrWhitespaceContent,
				expected: true,
			},
			{
				name:     "wrapped empty response error",
				err:      fmt.Errorf("API call failed: %w", llm.ErrEmptyResponse),
				expected: true,
			},
			{
				name:     "wrapped whitespace content error",
				err:      fmt.Errorf("API call failed: %w", llm.ErrWhitespaceContent),
				expected: true,
			},
			{
				name:     "double wrapped empty response error",
				err:      fmt.Errorf("outer error: %w", fmt.Errorf("inner error: %w", llm.ErrEmptyResponse)),
				expected: true,
			},

			// Common empty response phrases
			{
				name:     "message contains 'empty response'",
				err:      errors.New("received empty response from API"),
				expected: true,
			},
			{
				name:     "message contains 'empty content'",
				err:      errors.New("API returned empty content"),
				expected: true,
			},
			{
				name:     "message contains 'empty output'",
				err:      errors.New("model generated empty output"),
				expected: true,
			},
			{
				name:     "message contains 'empty result'",
				err:      errors.New("got empty result from LLM API"),
				expected: true,
			},

			// Case sensitivity testing
			{
				name:     "case insensitive - EMPTY RESPONSE",
				err:      errors.New("RECEIVED EMPTY RESPONSE FROM API"),
				expected: true,
			},
			{
				name:     "case insensitive - Mixed Case",
				err:      errors.New("API Returned Empty Content"),
				expected: true,
			},

			// Provider-specific patterns
			{
				name:     "message contains 'zero candidates'",
				err:      errors.New("model returned zero candidates"),
				expected: true,
			},
			{
				name:     "message contains 'empty candidates'",
				err:      errors.New("response contained empty candidates"),
				expected: true,
			},
			{
				name:     "message contains 'no output'",
				err:      errors.New("model generated no output"),
				expected: true,
			},

			// Partial matches (should not match)
			{
				name:     "partial match - not at word boundary",
				err:      errors.New("someemptyresponse text"),
				expected: false,
			},
			{
				name:     "error about non-empty content",
				err:      errors.New("content is not empty but invalid"),
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := service.IsEmptyResponseError(tc.err)
				if result != tc.expected {
					t.Errorf("Expected IsEmptyResponseError to return %v, got %v", tc.expected, result)
				}
			})
		}
	})

	// Test IsSafetyBlockedError
	t.Run("IsSafetyBlockedError", func(t *testing.T) {
		testCases := []struct {
			name     string
			err      error
			expected bool
		}{
			// Common error cases
			{
				name:     "nil error",
				err:      nil,
				expected: false,
			},
			{
				name:     "unrelated error",
				err:      errors.New("some other error"),
				expected: false,
			},

			// Specific error types
			{
				name:     "safety blocked error sentinel",
				err:      llm.ErrSafetyBlocked,
				expected: true,
			},
			{
				name:     "wrapped safety blocked error",
				err:      fmt.Errorf("API call failed: %w", llm.ErrSafetyBlocked),
				expected: true,
			},
			{
				name:     "double wrapped safety blocked error",
				err:      fmt.Errorf("outer error: %w", fmt.Errorf("inner error: %w", llm.ErrSafetyBlocked)),
				expected: true,
			},

			// Common safety-related phrases
			{
				name:     "message contains 'safety'",
				err:      errors.New("content blocked by safety filters"),
				expected: true,
			},
			{
				name:     "message contains 'content policy'",
				err:      errors.New("response rejected due to content policy violation"),
				expected: true,
			},
			{
				name:     "message contains 'content filter'",
				err:      errors.New("request failed because of content filter"),
				expected: true,
			},
			{
				name:     "message contains 'content_filter'",
				err:      errors.New("error: content_filter triggered"),
				expected: true,
			},

			// Case sensitivity testing
			{
				name:     "case insensitive - SAFETY",
				err:      errors.New("CONTENT BLOCKED BY SAFETY FILTERS"),
				expected: true,
			},
			{
				name:     "case insensitive - Mixed Case",
				err:      errors.New("Content Policy Violation"),
				expected: true,
			},

			// Provider-specific moderation terminology
			{
				name:     "message contains 'moderation'",
				err:      errors.New("failed moderation check"),
				expected: true,
			},
			{
				name:     "message contains 'blocked'",
				err:      errors.New("content was blocked"),
				expected: true,
			},
			{
				name:     "message contains 'filtered'",
				err:      errors.New("response was filtered"),
				expected: true,
			},
			{
				name:     "message contains 'harm_category'",
				err:      errors.New("harm_category: HARASSMENT"),
				expected: true,
			},

			// Words that don't contain any of the key phrases
			{
				name:     "error without any safety-related terms",
				err:      errors.New("this error has no matches"),
				expected: false,
			},
			{
				name:     "partial term match with different meaning",
				err:      errors.New("this request is unfiltered"),
				expected: true, // Will match "filtered" as a substring
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := service.IsSafetyBlockedError(tc.err)
				if result != tc.expected {
					t.Errorf("Expected IsSafetyBlockedError to return %v, got %v", tc.expected, result)
				}
			})
		}
	})
}
