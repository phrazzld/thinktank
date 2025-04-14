// Package architect contains the core application logic for the architect tool.
// This file tests the API service adapter.
package architect

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
)

// TestAPIServiceAdapter_InitClient tests the InitClient method of the APIServiceAdapter
func TestAPIServiceAdapter_InitClient(t *testing.T) {
	// Test constants
	const (
		testAPIKey      = "test-api-key"
		testModelName   = "test-model"
		testAPIEndpoint = "https://test-api-endpoint.example.com"
	)

	// Create test context
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name          string
		mockSetup     func(mock *MockAPIServiceForAdapter)
		expectedError bool
		expectedMsg   string // For error message validation
	}{
		{
			name: "success case - passes arguments correctly and returns client",
			mockSetup: func(mock *MockAPIServiceForAdapter) {
				// Setup to verify arguments and return a mock client
				var capturedAPIKey, capturedModelName, capturedAPIEndpoint string

				mock.InitClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
					// Capture the arguments for later verification
					capturedAPIKey = apiKey
					capturedModelName = modelName
					capturedAPIEndpoint = apiEndpoint

					// Return a mock client
					return &gemini.MockClient{}, nil
				}

				// Verify after the function call that arguments were passed through
				t.Cleanup(func() {
					if capturedAPIKey != testAPIKey {
						t.Errorf("Expected apiKey: %s, got: %s", testAPIKey, capturedAPIKey)
					}
					if capturedModelName != testModelName {
						t.Errorf("Expected modelName: %s, got: %s", testModelName, capturedModelName)
					}
					if capturedAPIEndpoint != testAPIEndpoint {
						t.Errorf("Expected apiEndpoint: %s, got: %s", testAPIEndpoint, capturedAPIEndpoint)
					}
				})
			},
			expectedError: false,
		},
		{
			name: "error case - returns error from underlying service",
			mockSetup: func(mock *MockAPIServiceForAdapter) {
				// Setup to return an error
				mock.InitClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
					return nil, errors.New("test error from APIService")
				}
			},
			expectedError: true,
			expectedMsg:   "test error from APIService",
		},
		{
			name: "nil APIService - returns error",
			mockSetup: func(mock *MockAPIServiceForAdapter) {
				// No setup needed - we'll use a nil APIService
			},
			expectedError: true,
			expectedMsg:   "nil APIService", // Expected error due to nil pointer dereference
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *APIServiceAdapter

			// For the nil APIService test
			if tc.name == "nil APIService - returns error" {
				// Create an adapter with nil APIService - should panic
				adapter = &APIServiceAdapter{
					APIService: nil,
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
				_, _ = adapter.InitClient(ctx, testAPIKey, testModelName, testAPIEndpoint)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			client, err := adapter.InitClient(ctx, testAPIKey, testModelName, testAPIEndpoint)

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected an error but got nil")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check error message if applicable
			if tc.expectedError && err != nil && tc.expectedMsg != "" {
				if !strings.Contains(err.Error(), tc.expectedMsg) {
					t.Errorf("Expected error message to contain '%s', got: '%s'", tc.expectedMsg, err.Error())
				}
			}

			// For success case, verify non-nil client
			if !tc.expectedError {
				if client == nil {
					t.Error("Expected a non-nil client but got nil")
				}
			}
		})
	}
}

// TestAPIServiceAdapter_ProcessResponse tests the ProcessResponse method of the APIServiceAdapter
func TestAPIServiceAdapter_ProcessResponse(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		inputResult   *gemini.GenerationResult
		mockSetup     func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult)
		expectedValue string
		expectedError bool
		expectedMsg   string // For error message validation
	}{
		{
			name: "success case - passes result correctly and returns content",
			inputResult: &gemini.GenerationResult{
				Content: "This is a test response",
			},
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// Setup to verify arguments and return content
				var capturedResult *gemini.GenerationResult

				mock.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					// Capture the input for later verification
					capturedResult = result

					// Return the expected content
					return "This is a test response", nil
				}

				// Verify after the function call that arguments were passed through
				t.Cleanup(func() {
					if capturedResult != inputResult {
						t.Errorf("Expected the same input result instance to be passed through")
					}
				})
			},
			expectedValue: "This is a test response",
			expectedError: false,
		},
		{
			name: "error case - returns error from underlying service",
			inputResult: &gemini.GenerationResult{
				Content: "",
			},
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// Setup to return an error
				mock.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					return "", errors.New("empty response error")
				}
			},
			expectedValue: "",
			expectedError: true,
			expectedMsg:   "empty response error",
		},
		{
			name: "nil APIService - returns error",
			inputResult: &gemini.GenerationResult{
				Content: "This will panic",
			},
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// No setup needed - we'll use a nil APIService
			},
			expectedValue: "",
			expectedError: true,
			expectedMsg:   "nil APIService", // Expected error due to nil pointer dereference
		},
		{
			name:        "nil result - returns error",
			inputResult: nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// Setup to handle nil result
				mock.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					if result == nil {
						return "", errors.New("nil result")
					}
					return result.Content, nil
				}
			},
			expectedValue: "",
			expectedError: true,
			expectedMsg:   "nil result",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *APIServiceAdapter

			// For the nil APIService test
			if tc.name == "nil APIService - returns error" {
				// Create an adapter with nil APIService - should panic
				adapter = &APIServiceAdapter{
					APIService: nil,
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
				_, _ = adapter.ProcessResponse(tc.inputResult)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.inputResult)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			content, err := adapter.ProcessResponse(tc.inputResult)

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected an error but got nil")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check error message if applicable
			if tc.expectedError && err != nil && tc.expectedMsg != "" {
				if !strings.Contains(err.Error(), tc.expectedMsg) {
					t.Errorf("Expected error message to contain '%s', got: '%s'", tc.expectedMsg, err.Error())
				}
			}

			// For success case, verify content
			if !tc.expectedError {
				if content != tc.expectedValue {
					t.Errorf("Expected content: '%s', got: '%s'", tc.expectedValue, content)
				}
			}
		})
	}
}

// TestAPIServiceAdapter_IsEmptyResponseError tests the IsEmptyResponseError method of the APIServiceAdapter
func TestAPIServiceAdapter_IsEmptyResponseError(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		err         error
		mockSetup   func(mock *MockAPIServiceForAdapter, err error)
		expectedVal bool
	}{
		{
			name: "true case - passthrough to underlying implementation",
			err:  errors.New("empty response error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup mock to return true and verify passthrough
				var capturedErr error

				mock.IsEmptyResponseErrorFunc = func(err error) bool {
					// Capture the input for later verification
					capturedErr = err

					// Return true to indicate it's an empty response error
					return true
				}

				// Verify after the function call that error was passed through
				t.Cleanup(func() {
					if capturedErr != err {
						t.Errorf("Expected error to be passed through to underlying implementation")
					}
				})
			},
			expectedVal: true,
		},
		{
			name: "false case - passthrough to underlying implementation",
			err:  errors.New("some other error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup mock to return false and verify passthrough
				var capturedErr error

				mock.IsEmptyResponseErrorFunc = func(err error) bool {
					// Capture the input for later verification
					capturedErr = err

					// Return false to indicate it's not an empty response error
					return false
				}

				// Verify after the function call that error was passed through
				t.Cleanup(func() {
					if capturedErr != err {
						t.Errorf("Expected error to be passed through to underlying implementation")
					}
				})
			},
			expectedVal: false,
		},
		{
			name: "nil APIService - should panic",
			err:  errors.New("this will cause panic"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// No setup needed - we'll use a nil APIService
			},
			expectedVal: false, // Not used in this test case
		},
		{
			name: "nil error - should handle gracefully",
			err:  nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to handle nil error
				mock.IsEmptyResponseErrorFunc = func(err error) bool {
					if err == nil {
						return false
					}
					return true
				}
			},
			expectedVal: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *APIServiceAdapter

			// For the nil APIService test
			if tc.name == "nil APIService - should panic" {
				// Create an adapter with nil APIService - should panic
				adapter = &APIServiceAdapter{
					APIService: nil,
				}

				// Call should panic, recover and mark as pass
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.IsEmptyResponseError(tc.err)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.err)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			result := adapter.IsEmptyResponseError(tc.err)

			// Verify expected result except for panic case
			if result != tc.expectedVal {
				t.Errorf("Expected result: %v, got: %v", tc.expectedVal, result)
			}
		})
	}
}

// TestAPIServiceAdapter_IsSafetyBlockedError tests the IsSafetyBlockedError method of the APIServiceAdapter
func TestAPIServiceAdapter_IsSafetyBlockedError(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		err         error
		mockSetup   func(mock *MockAPIServiceForAdapter, err error)
		expectedVal bool
	}{
		{
			name: "true case - passthrough to underlying implementation",
			err:  errors.New("safety blocked error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup mock to return true and verify passthrough
				var capturedErr error

				mock.IsSafetyBlockedErrorFunc = func(err error) bool {
					// Capture the input for later verification
					capturedErr = err

					// Return true to indicate it's a safety blocked error
					return true
				}

				// Verify after the function call that error was passed through
				t.Cleanup(func() {
					if capturedErr != err {
						t.Errorf("Expected error to be passed through to underlying implementation")
					}
				})
			},
			expectedVal: true,
		},
		{
			name: "false case - passthrough to underlying implementation",
			err:  errors.New("some other error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup mock to return false and verify passthrough
				var capturedErr error

				mock.IsSafetyBlockedErrorFunc = func(err error) bool {
					// Capture the input for later verification
					capturedErr = err

					// Return false to indicate it's not a safety blocked error
					return false
				}

				// Verify after the function call that error was passed through
				t.Cleanup(func() {
					if capturedErr != err {
						t.Errorf("Expected error to be passed through to underlying implementation")
					}
				})
			},
			expectedVal: false,
		},
		{
			name: "nil APIService - should panic",
			err:  errors.New("this will cause panic"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// No setup needed - we'll use a nil APIService
			},
			expectedVal: false, // Not used in this test case
		},
		{
			name: "nil error - should handle gracefully",
			err:  nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to handle nil error
				mock.IsSafetyBlockedErrorFunc = func(err error) bool {
					if err == nil {
						return false
					}
					return true
				}
			},
			expectedVal: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *APIServiceAdapter

			// For the nil APIService test
			if tc.name == "nil APIService - should panic" {
				// Create an adapter with nil APIService - should panic
				adapter = &APIServiceAdapter{
					APIService: nil,
				}

				// Call should panic, recover and mark as pass
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.IsSafetyBlockedError(tc.err)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.err)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			result := adapter.IsSafetyBlockedError(tc.err)

			// Verify expected result except for panic case
			if result != tc.expectedVal {
				t.Errorf("Expected result: %v, got: %v", tc.expectedVal, result)
			}
		})
	}
}

// TestAPIServiceAdapter_GetErrorDetails tests the GetErrorDetails method of the APIServiceAdapter
func TestAPIServiceAdapter_GetErrorDetails(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		err         error
		mockSetup   func(mock *MockAPIServiceForAdapter, err error)
		expectedVal string
	}{
		{
			name: "normal case - passthrough to underlying implementation",
			err:  errors.New("test error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup mock to return details and verify passthrough
				var capturedErr error

				mock.GetErrorDetailsFunc = func(err error) string {
					// Capture the input for later verification
					capturedErr = err

					// Return detailed error message
					return "Detailed error: test error caused by X"
				}

				// Verify after the function call that error was passed through
				t.Cleanup(func() {
					if capturedErr != err {
						t.Errorf("Expected error to be passed through to underlying implementation")
					}
				})
			},
			expectedVal: "Detailed error: test error caused by X",
		},
		{
			name: "nil APIService - should panic",
			err:  errors.New("this will cause panic"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// No setup needed - we'll use a nil APIService
			},
			expectedVal: "", // Not used in this test case
		},
		{
			name: "nil error - should handle gracefully",
			err:  nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to handle nil error
				mock.GetErrorDetailsFunc = func(err error) string {
					if err == nil {
						return "No error"
					}
					return "Some error"
				}
			},
			expectedVal: "No error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *APIServiceAdapter

			// For the nil APIService test
			if tc.name == "nil APIService - should panic" {
				// Create an adapter with nil APIService - should panic
				adapter = &APIServiceAdapter{
					APIService: nil,
				}

				// Call should panic, recover and mark as pass
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.GetErrorDetails(tc.err)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.err)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			result := adapter.GetErrorDetails(tc.err)

			// Verify expected result except for panic case
			if result != tc.expectedVal {
				t.Errorf("Expected result: '%s', got: '%s'", tc.expectedVal, result)
			}
		})
	}
}
