// Package architect contains the core application logic for the architect tool.
// This file tests the token-related adapters.
package architect

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
)

// TestTokenResultAdapter tests the TokenResultAdapter function that maps modelproc.TokenResult to TokenResult
func TestTokenResultAdapter(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		input          *TokenResult
		expectedOutput *modelproc.TokenResult
	}{
		{
			name: "normal case - maps all fields correctly",
			input: &TokenResult{
				TokenCount:   1000,
				InputLimit:   2000,
				ExceedsLimit: false,
				LimitError:   "",
				Percentage:   50.0,
			},
			expectedOutput: &modelproc.TokenResult{
				TokenCount:   1000,
				InputLimit:   2000,
				ExceedsLimit: false,
				LimitError:   "",
				Percentage:   50.0,
			},
		},
		{
			name: "exceeded limit case - maps all fields including error",
			input: &TokenResult{
				TokenCount:   3000,
				InputLimit:   2000,
				ExceedsLimit: true,
				LimitError:   "Token limit exceeded (3000 > 2000)",
				Percentage:   150.0,
			},
			expectedOutput: &modelproc.TokenResult{
				TokenCount:   3000,
				InputLimit:   2000,
				ExceedsLimit: true,
				LimitError:   "Token limit exceeded (3000 > 2000)",
				Percentage:   150.0,
			},
		},
		{
			name:           "nil input - returns nil",
			input:          nil,
			expectedOutput: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the adapter function
			result := TokenResultAdapter(tc.input)

			// For nil case, just check that the result is nil
			if tc.input == nil {
				if result != nil {
					t.Errorf("Expected nil result for nil input, got: %+v", result)
				}
				return
			}

			// For non-nil cases, verify all fields were mapped correctly
			if result == nil {
				t.Fatalf("Got nil result for non-nil input")
			}

			// Verify all fields
			if result.TokenCount != tc.expectedOutput.TokenCount {
				t.Errorf("TokenCount mismatch: expected %d, got %d",
					tc.expectedOutput.TokenCount, result.TokenCount)
			}
			if result.InputLimit != tc.expectedOutput.InputLimit {
				t.Errorf("InputLimit mismatch: expected %d, got %d",
					tc.expectedOutput.InputLimit, result.InputLimit)
			}
			if result.ExceedsLimit != tc.expectedOutput.ExceedsLimit {
				t.Errorf("ExceedsLimit mismatch: expected %v, got %v",
					tc.expectedOutput.ExceedsLimit, result.ExceedsLimit)
			}
			if result.LimitError != tc.expectedOutput.LimitError {
				t.Errorf("LimitError mismatch: expected %q, got %q",
					tc.expectedOutput.LimitError, result.LimitError)
			}
			if result.Percentage != tc.expectedOutput.Percentage {
				t.Errorf("Percentage mismatch: expected %.2f, got %.2f",
					tc.expectedOutput.Percentage, result.Percentage)
			}
		})
	}
}

// TestTokenManagerAdapter_GetTokenInfo tests the GetTokenInfo method of the TokenManagerAdapter
func TestTokenManagerAdapter_GetTokenInfo(t *testing.T) {
	// Test constants
	const testPrompt = "This is a test prompt"

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		mockSetup     func(mock *MockTokenManagerForAdapter)
		expectedError bool
		expectedMsg   string // For error message validation
	}{
		{
			name: "success case - passes arguments correctly and adapts return value",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// Setup to verify arguments and return a token result
				var capturedCtx context.Context
				var capturedPrompt string

				mock.GetTokenInfoFunc = func(ctx context.Context, prompt string) (*TokenResult, error) {
					// Capture the arguments for later verification
					capturedCtx = ctx
					capturedPrompt = prompt

					// Return a mock result
					return &TokenResult{
						TokenCount:   1000,
						InputLimit:   2000,
						ExceedsLimit: false,
						LimitError:   "",
						Percentage:   50.0,
					}, nil
				}

				// Verify after the function call that arguments were passed through
				t.Cleanup(func() {
					if capturedCtx != ctx {
						t.Errorf("Expected context to be passed through")
					}
					if capturedPrompt != testPrompt {
						t.Errorf("Expected prompt: %q, got: %q", testPrompt, capturedPrompt)
					}
				})
			},
			expectedError: false,
		},
		{
			name: "error case - returns error from underlying service",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// Setup to return an error
				mock.GetTokenInfoFunc = func(ctx context.Context, prompt string) (*TokenResult, error) {
					return nil, errors.New("token counting error")
				}
			},
			expectedError: true,
			expectedMsg:   "token counting error",
		},
		{
			name: "nil TokenManager - should panic",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// No setup needed - we'll use a nil TokenManager
			},
			expectedError: true,
			expectedMsg:   "nil TokenManager", // Should never reach here due to panic
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *TokenManagerAdapter

			// For the nil TokenManager test
			if tc.name == "nil TokenManager - should panic" {
				// Create an adapter with nil TokenManager - should panic
				adapter = &TokenManagerAdapter{
					TokenManager: nil,
				}

				// Call should panic, recover and mark as error
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_, _ = adapter.GetTokenInfo(ctx, testPrompt)
				return
			}

			// Create a mock TokenManager for non-nil test cases
			mockTokenManager := &MockTokenManagerForAdapter{}

			// Setup the mock
			tc.mockSetup(mockTokenManager)

			// Create adapter with mock
			adapter = &TokenManagerAdapter{
				TokenManager: mockTokenManager,
			}

			// Call the method being tested
			result, err := adapter.GetTokenInfo(ctx, testPrompt)

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected an error but got nil")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check error message if applicable
			if tc.expectedError && err != nil && tc.expectedMsg != "" {
				if err.Error() != tc.expectedMsg {
					t.Errorf("Expected error message '%s', got: '%s'", tc.expectedMsg, err.Error())
				}
			}

			// For success case, verify non-nil result with adapted values
			if !tc.expectedError {
				if result == nil {
					t.Error("Expected a non-nil result but got nil")
				} else {
					// Verify specific fields from the adapted result
					if result.TokenCount != 1000 || result.InputLimit != 2000 ||
						result.ExceedsLimit != false || result.Percentage != 50.0 {
						t.Errorf("Result adaptation incorrect: %+v", result)
					}
				}
			}
		})
	}
}

// TestTokenManagerAdapter_CheckTokenLimit tests the CheckTokenLimit method of the TokenManagerAdapter
func TestTokenManagerAdapter_CheckTokenLimit(t *testing.T) {
	// Test constants
	const testPrompt = "This is a test prompt"

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		mockSetup     func(mock *MockTokenManagerForAdapter)
		expectedError bool
		expectedMsg   string // For error message validation
	}{
		{
			name: "success case - passes arguments correctly",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// Setup to verify arguments
				var capturedCtx context.Context
				var capturedPrompt string

				mock.CheckTokenLimitFunc = func(ctx context.Context, prompt string) error {
					// Capture the arguments for later verification
					capturedCtx = ctx
					capturedPrompt = prompt

					// Return success
					return nil
				}

				// Verify after the function call that arguments were passed through
				t.Cleanup(func() {
					if capturedCtx != ctx {
						t.Errorf("Expected context to be passed through")
					}
					if capturedPrompt != testPrompt {
						t.Errorf("Expected prompt: %q, got: %q", testPrompt, capturedPrompt)
					}
				})
			},
			expectedError: false,
		},
		{
			name: "error case - returns error from underlying service",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// Setup to return an error
				mock.CheckTokenLimitFunc = func(ctx context.Context, prompt string) error {
					return errors.New("token limit exceeded")
				}
			},
			expectedError: true,
			expectedMsg:   "token limit exceeded",
		},
		{
			name: "nil TokenManager - should panic",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// No setup needed - we'll use a nil TokenManager
			},
			expectedError: true,
			expectedMsg:   "nil TokenManager", // Should never reach here due to panic
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *TokenManagerAdapter

			// For the nil TokenManager test
			if tc.name == "nil TokenManager - should panic" {
				// Create an adapter with nil TokenManager - should panic
				adapter = &TokenManagerAdapter{
					TokenManager: nil,
				}

				// Call should panic, recover and mark as error
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.CheckTokenLimit(ctx, testPrompt)
				return
			}

			// Create a mock TokenManager for non-nil test cases
			mockTokenManager := &MockTokenManagerForAdapter{}

			// Setup the mock
			tc.mockSetup(mockTokenManager)

			// Create adapter with mock
			adapter = &TokenManagerAdapter{
				TokenManager: mockTokenManager,
			}

			// Call the method being tested
			err := adapter.CheckTokenLimit(ctx, testPrompt)

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected an error but got nil")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check error message if applicable
			if tc.expectedError && err != nil && tc.expectedMsg != "" {
				if err.Error() != tc.expectedMsg {
					t.Errorf("Expected error message '%s', got: '%s'", tc.expectedMsg, err.Error())
				}
			}
		})
	}
}

// TestTokenManagerAdapter_PromptForConfirmation tests the PromptForConfirmation method of the TokenManagerAdapter
func TestTokenManagerAdapter_PromptForConfirmation(t *testing.T) {
	// Test constants
	const (
		testTokenCount = int32(5000)
		testThreshold  = 1000
	)

	// Test cases
	tests := []struct {
		name          string
		mockSetup     func(mock *MockTokenManagerForAdapter)
		expectedValue bool
	}{
		{
			name: "true case - passes arguments correctly and returns true",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// Setup to verify arguments and return true
				var capturedTokenCount int32
				var capturedThreshold int

				mock.PromptForConfirmationFunc = func(tokenCount int32, threshold int) bool {
					// Capture the arguments for later verification
					capturedTokenCount = tokenCount
					capturedThreshold = threshold

					// Return true to indicate confirmation
					return true
				}

				// Verify after the function call that arguments were passed through
				t.Cleanup(func() {
					if capturedTokenCount != testTokenCount {
						t.Errorf("Expected tokenCount: %d, got: %d", testTokenCount, capturedTokenCount)
					}
					if capturedThreshold != testThreshold {
						t.Errorf("Expected threshold: %d, got: %d", testThreshold, capturedThreshold)
					}
				})
			},
			expectedValue: true,
		},
		{
			name: "false case - passes arguments correctly and returns false",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// Setup to verify arguments and return false
				var capturedTokenCount int32
				var capturedThreshold int

				mock.PromptForConfirmationFunc = func(tokenCount int32, threshold int) bool {
					// Capture the arguments for later verification
					capturedTokenCount = tokenCount
					capturedThreshold = threshold

					// Return false to indicate no confirmation
					return false
				}

				// Verify after the function call that arguments were passed through
				t.Cleanup(func() {
					if capturedTokenCount != testTokenCount {
						t.Errorf("Expected tokenCount: %d, got: %d", testTokenCount, capturedTokenCount)
					}
					if capturedThreshold != testThreshold {
						t.Errorf("Expected threshold: %d, got: %d", testThreshold, capturedThreshold)
					}
				})
			},
			expectedValue: false,
		},
		{
			name: "nil TokenManager - should panic",
			mockSetup: func(mock *MockTokenManagerForAdapter) {
				// No setup needed - we'll use a nil TokenManager
			},
			expectedValue: false, // Not used in this test case
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *TokenManagerAdapter

			// For the nil TokenManager test
			if tc.name == "nil TokenManager - should panic" {
				// Create an adapter with nil TokenManager - should panic
				adapter = &TokenManagerAdapter{
					TokenManager: nil,
				}

				// Call should panic, recover and mark as error
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.PromptForConfirmation(testTokenCount, testThreshold)
				return
			}

			// Create a mock TokenManager for non-nil test cases
			mockTokenManager := &MockTokenManagerForAdapter{}

			// Setup the mock
			tc.mockSetup(mockTokenManager)

			// Create adapter with mock
			adapter = &TokenManagerAdapter{
				TokenManager: mockTokenManager,
			}

			// Call the method being tested
			result := adapter.PromptForConfirmation(testTokenCount, testThreshold)

			// Verify expected result except for panic case
			if result != tc.expectedValue {
				t.Errorf("Expected result: %v, got: %v", tc.expectedValue, result)
			}
		})
	}
}
