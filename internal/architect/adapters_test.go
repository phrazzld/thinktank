// Package architect contains the core application logic for the architect tool.
// This file tests the adapters that implement interfaces for various services.
package architect

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
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
			name:        "nil result - returns error",
			inputResult: nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// Setup to return an error for nil result
				mock.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					if result == nil {
						return "", errors.New("nil result error")
					}
					return "this should not be reached", nil
				}
			},
			expectedValue: "",
			expectedError: true,
			expectedMsg:   "nil result error",
		},
		{
			name: "safety blocked - returns error",
			inputResult: &gemini.GenerationResult{
				Content: "",
				SafetyRatings: []gemini.SafetyRating{
					{
						Category: "HARMFUL_CATEGORY",
						Blocked:  true,
					},
				},
			},
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// Setup to return a safety error
				mock.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					return "", errors.New("content blocked by safety filters")
				}
			},
			expectedValue: "",
			expectedError: true,
			expectedMsg:   "safety filters",
		},
		{
			name: "whitespace response - returns error",
			inputResult: &gemini.GenerationResult{
				Content: "   \n   ",
			},
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// Setup to return a whitespace error
				mock.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					return "", errors.New("whitespace content error")
				}
			},
			expectedValue: "",
			expectedError: true,
			expectedMsg:   "whitespace content error",
		},
		{
			name: "nil APIService - returns error",
			inputResult: &gemini.GenerationResult{
				Content: "This is a test response",
			},
			mockSetup: func(mock *MockAPIServiceForAdapter, inputResult *gemini.GenerationResult) {
				// No setup needed - we'll use a nil APIService
			},
			expectedValue: "",
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
					t.Errorf("Expected content '%s', got: '%s'", tc.expectedValue, content)
				}
			}
		})
	}
}

// TestAPIServiceAdapter_IsEmptyResponseError tests the IsEmptyResponseError method of the APIServiceAdapter
func TestAPIServiceAdapter_IsEmptyResponseError(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		testError      error
		mockSetup      func(mock *MockAPIServiceForAdapter, err error)
		expectedResult bool
	}{
		{
			name:      "should delegate to APIService and return true for empty response error",
			testError: errors.New("empty response error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify argument and return true
				var capturedError error

				mock.IsEmptyResponseErrorFunc = func(e error) bool {
					// Capture the input for later verification
					capturedError = e

					// Return true as if it's an empty response error
					return true
				}

				// Verify after the function call that the error was passed through
				t.Cleanup(func() {
					if capturedError != err {
						t.Errorf("Expected the same error instance to be passed through")
					}
				})
			},
			expectedResult: true,
		},
		{
			name:      "should delegate to APIService and return false for non-empty response error",
			testError: errors.New("some other error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify argument and return false
				var capturedError error

				mock.IsEmptyResponseErrorFunc = func(e error) bool {
					// Capture the input for later verification
					capturedError = e

					// Return false as if it's not an empty response error
					return false
				}

				// Verify after the function call that the error was passed through
				t.Cleanup(func() {
					if capturedError != err {
						t.Errorf("Expected the same error instance to be passed through")
					}
				})
			},
			expectedResult: false,
		},
		{
			name:      "should handle nil error",
			testError: nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify nil error is passed and return false
				mock.IsEmptyResponseErrorFunc = func(e error) bool {
					if e != nil {
						t.Errorf("Expected nil error to be passed, got: %v", e)
					}
					return false
				}
			},
			expectedResult: false,
		},
		{
			name:      "nil APIService - should panic",
			testError: errors.New("test error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// No setup needed - we'll use a nil APIService
			},
			expectedResult: false, // Not used in this case as we expect a panic
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

				// Call should panic, recover and mark as success
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.IsEmptyResponseError(tc.testError)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.testError)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			result := adapter.IsEmptyResponseError(tc.testError)

			// Verify the result
			if result != tc.expectedResult {
				t.Errorf("Expected result %v, got: %v", tc.expectedResult, result)
			}
		})
	}
}

// TestAPIServiceAdapter_IsSafetyBlockedError tests the IsSafetyBlockedError method of the APIServiceAdapter
func TestAPIServiceAdapter_IsSafetyBlockedError(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		testError      error
		mockSetup      func(mock *MockAPIServiceForAdapter, err error)
		expectedResult bool
	}{
		{
			name:      "should delegate to APIService and return true for safety blocked error",
			testError: errors.New("safety blocked error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify argument and return true
				var capturedError error

				mock.IsSafetyBlockedErrorFunc = func(e error) bool {
					// Capture the input for later verification
					capturedError = e

					// Return true as if it's a safety blocked error
					return true
				}

				// Verify after the function call that the error was passed through
				t.Cleanup(func() {
					if capturedError != err {
						t.Errorf("Expected the same error instance to be passed through")
					}
				})
			},
			expectedResult: true,
		},
		{
			name:      "should delegate to APIService and return false for non-safety blocked error",
			testError: errors.New("some other error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify argument and return false
				var capturedError error

				mock.IsSafetyBlockedErrorFunc = func(e error) bool {
					// Capture the input for later verification
					capturedError = e

					// Return false as if it's not a safety blocked error
					return false
				}

				// Verify after the function call that the error was passed through
				t.Cleanup(func() {
					if capturedError != err {
						t.Errorf("Expected the same error instance to be passed through")
					}
				})
			},
			expectedResult: false,
		},
		{
			name:      "should handle nil error",
			testError: nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify nil error is passed and return false
				mock.IsSafetyBlockedErrorFunc = func(e error) bool {
					if e != nil {
						t.Errorf("Expected nil error to be passed, got: %v", e)
					}
					return false
				}
			},
			expectedResult: false,
		},
		{
			name:      "nil APIService - should panic",
			testError: errors.New("test error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// No setup needed - we'll use a nil APIService
			},
			expectedResult: false, // Not used in this case as we expect a panic
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

				// Call should panic, recover and mark as success
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.IsSafetyBlockedError(tc.testError)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.testError)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			result := adapter.IsSafetyBlockedError(tc.testError)

			// Verify the result
			if result != tc.expectedResult {
				t.Errorf("Expected result %v, got: %v", tc.expectedResult, result)
			}
		})
	}
}

// TestAPIServiceAdapter_GetErrorDetails tests the GetErrorDetails method of the APIServiceAdapter
func TestAPIServiceAdapter_GetErrorDetails(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		testError      error
		mockSetup      func(mock *MockAPIServiceForAdapter, err error)
		expectedResult string
	}{
		{
			name:      "should delegate to APIService and return detailed error message",
			testError: errors.New("test error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify argument and return a detailed message
				var capturedError error

				mock.GetErrorDetailsFunc = func(e error) string {
					// Capture the input for later verification
					capturedError = e

					// Return a detailed error message
					return "Detailed error message for test error"
				}

				// Verify after the function call that the error was passed through
				t.Cleanup(func() {
					if capturedError != err {
						t.Errorf("Expected the same error instance to be passed through")
					}
				})
			},
			expectedResult: "Detailed error message for test error",
		},
		{
			name:      "should handle nil error",
			testError: nil,
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to verify nil error is passed and return empty string
				mock.GetErrorDetailsFunc = func(e error) string {
					if e != nil {
						t.Errorf("Expected nil error to be passed, got: %v", e)
					}
					return ""
				}
			},
			expectedResult: "",
		},
		{
			name:      "should handle API error with user-facing message",
			testError: &gemini.APIError{Message: "API Error", Type: gemini.ErrorTypeRateLimit, Suggestion: "Try again later"},
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// Setup to return a user-facing message for API errors
				mock.GetErrorDetailsFunc = func(e error) string {
					// For API errors, return a user-facing message
					return "Error: API Error. Suggestion: Try again later"
				}
			},
			expectedResult: "Error: API Error. Suggestion: Try again later",
		},
		{
			name:      "nil APIService - should panic",
			testError: errors.New("test error"),
			mockSetup: func(mock *MockAPIServiceForAdapter, err error) {
				// No setup needed - we'll use a nil APIService
			},
			expectedResult: "", // Not used in this case as we expect a panic
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

				// Call should panic, recover and mark as success
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.GetErrorDetails(tc.testError)
				return
			}

			// Create a mock APIService for non-nil test cases
			mockAPIService := &MockAPIServiceForAdapter{}

			// Setup the mock
			tc.mockSetup(mockAPIService, tc.testError)

			// Create adapter with mock
			adapter = &APIServiceAdapter{
				APIService: mockAPIService,
			}

			// Call the method being tested
			result := adapter.GetErrorDetails(tc.testError)

			// Verify the result
			if result != tc.expectedResult {
				t.Errorf("Expected result '%s', got: '%s'", tc.expectedResult, result)
			}
		})
	}
}

// This file contains tests for adapter implementations including:
// - APIServiceAdapter: Adapts internal APIService to interfaces.APIService
// - TokenResultAdapter: Adapts TokenResult to modelproc.TokenResult
// - TokenManagerAdapter: Adapts internal TokenManager to interfaces.TokenManager
// - ContextGathererAdapter: Adapts internal ContextGatherer to interfaces.ContextGatherer
// - FileWriterAdapter: Adapts internal FileWriter to interfaces.FileWriter

// MockAPIServiceForAdapter is a testing mock for the APIService interface, specifically for adapter tests
type MockAPIServiceForAdapter struct {
	InitClientFunc           func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error)
	ProcessResponseFunc      func(result *gemini.GenerationResult) (string, error)
	IsEmptyResponseErrorFunc func(err error) bool
	IsSafetyBlockedErrorFunc func(err error) bool
	GetErrorDetailsFunc      func(err error) string
}

func (m *MockAPIServiceForAdapter) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	if m.InitClientFunc != nil {
		return m.InitClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, errors.New("InitClient not implemented")
}

func (m *MockAPIServiceForAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	if m.ProcessResponseFunc != nil {
		return m.ProcessResponseFunc(result)
	}
	return "", errors.New("ProcessResponse not implemented")
}

func (m *MockAPIServiceForAdapter) IsEmptyResponseError(err error) bool {
	if m.IsEmptyResponseErrorFunc != nil {
		return m.IsEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *MockAPIServiceForAdapter) IsSafetyBlockedError(err error) bool {
	if m.IsSafetyBlockedErrorFunc != nil {
		return m.IsSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *MockAPIServiceForAdapter) GetErrorDetails(err error) string {
	if m.GetErrorDetailsFunc != nil {
		return m.GetErrorDetailsFunc(err)
	}
	return "Error details not implemented"
}

// MockContextGathererForAdapter is a testing mock for the ContextGatherer interface, specifically for adapter tests
type MockContextGathererForAdapter struct {
	GatherContextFunc     func(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)
	DisplayDryRunInfoFunc func(ctx context.Context, stats *ContextStats) error
}

func (m *MockContextGathererForAdapter) GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
	if m.GatherContextFunc != nil {
		return m.GatherContextFunc(ctx, config)
	}
	return nil, nil, errors.New("GatherContext not implemented")
}

func (m *MockContextGathererForAdapter) DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error {
	if m.DisplayDryRunInfoFunc != nil {
		return m.DisplayDryRunInfoFunc(ctx, stats)
	}
	return errors.New("DisplayDryRunInfo not implemented")
}

// MockFileWriterForAdapter is a testing mock for the FileWriter interface, specifically for adapter tests
type MockFileWriterForAdapter struct {
	SaveToFileFunc func(content, outputFile string) error
}

func (m *MockFileWriterForAdapter) SaveToFile(content, outputFile string) error {
	if m.SaveToFileFunc != nil {
		return m.SaveToFileFunc(content, outputFile)
	}
	return errors.New("SaveToFile not implemented")
}

// GetAdapterTestLogger returns a logger for adapter tests
func GetAdapterTestLogger() logutil.LoggerInterface {
	return logutil.NewLogger(logutil.DebugLevel, nil, "[adapter-test] ")
}

// The wrapper for gemini.NewClient is now implemented directly in the API service file
