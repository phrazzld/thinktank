// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"errors"
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
			{
				name:     "nil error",
				err:      nil,
				expected: false,
			},
			{
				name:     "empty response error",
				err:      llm.ErrEmptyResponse,
				expected: true,
			},
			{
				name:     "whitespace content error",
				err:      llm.ErrWhitespaceContent,
				expected: true,
			},
			{
				name:     "message contains empty response",
				err:      errors.New("received empty response from API"),
				expected: true,
			},
			{
				name:     "unrelated error",
				err:      errors.New("some other error"),
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
			{
				name:     "nil error",
				err:      nil,
				expected: false,
			},
			{
				name:     "safety blocked error",
				err:      llm.ErrSafetyBlocked,
				expected: true,
			},
			{
				name:     "message contains safety",
				err:      errors.New("content blocked by safety filters"),
				expected: true,
			},
			{
				name:     "unrelated error",
				err:      errors.New("some other error"),
				expected: false,
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
