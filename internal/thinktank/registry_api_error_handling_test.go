// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"errors"
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
	}{
		{
			name:            "nil result",
			result:          nil,
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
		},
		{
			name: "empty content",
			result: &llm.ProviderResult{
				Content:      "",
				FinishReason: "stop",
			},
			expectedContent: "",
			expectError:     true,
			expectedError:   llm.ErrEmptyResponse,
		},
		{
			name: "safety blocked",
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
			name: "valid content",
			result: &llm.ProviderResult{
				Content: "This is a valid response",
			},
			expectedContent: "This is a valid response",
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

				if !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error to be '%v', got '%v'", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

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
