package architect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockLogger for testing
type mockAPILogger struct {
	logutil.LoggerInterface
	debugMessages []string
	infoMessages  []string
	errorMessages []string
}

func (m *mockAPILogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockAPILogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockAPILogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

// mockAPIService is a test implementation of APIService
type mockAPIService struct {
	logger     logutil.LoggerInterface
	initError  error
	mockClient gemini.Client
}

func newMockAPIService(logger logutil.LoggerInterface, initError error, mockClient gemini.Client) APIService {
	return &mockAPIService{
		logger:     logger,
		initError:  initError,
		mockClient: mockClient,
	}
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	if m.initError != nil {
		return nil, m.initError
	}
	return m.mockClient, nil
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return "", errors.New("not implemented in mock")
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return false
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return err.Error()
}

// TestNewAPIService tests the creation of a new APIService
func TestNewAPIService(t *testing.T) {
	logger := &mockAPILogger{}

	// Create a new APIService
	service := NewAPIService(logger)

	// Check that service is not nil
	if service == nil {
		t.Error("Expected non-nil APIService, got nil")
	}

	// Check that it implements the APIService interface
	var _ APIService = service // This is a compile-time check
}

// TestInitClient tests the InitClient method with table-driven tests
func TestInitClient(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name      string
		apiKey    string
		modelName string
		setupCtx  func() (context.Context, context.CancelFunc)
		mockError error  // Error to inject into the mock gemini.NewClient
		wantErr   error  // Expected error type to match with errors.Is
		wantMsg   string // Expected error message substring
	}{
		{
			name:      "Empty API Key",
			apiKey:    "",
			modelName: "fake-model",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:   ErrClientInitialization,
			wantMsg:   "API key is required",
		},
		{
			name:      "Empty Model Name",
			apiKey:    "fake-api-key",
			modelName: "",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:   ErrClientInitialization,
			wantMsg:   "model name is required",
		},
		{
			name:      "Cancelled Context",
			apiKey:    "fake-api-key",
			modelName: "fake-model",
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			wantErr: ErrClientInitialization,
			wantMsg: "context",
		},
		{
			name:      "Generic Error From NewClient",
			apiKey:    "fake-api-key",
			modelName: "fake-model",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			mockError: errors.New("generic client error"),
			wantErr:   ErrClientInitialization,
			wantMsg:   "generic client error",
		},
		{
			name:      "API Error From NewClient",
			apiKey:    "fake-api-key",
			modelName: "fake-model",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			mockError: &gemini.APIError{
				Message:    "API authentication failed",
				Suggestion: "Check your API key",
			},
			wantErr: ErrClientInitialization,
			wantMsg: "API authentication failed",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &mockAPILogger{}
			api := NewAPIService(logger).(*apiService) // Type assertion to access internal fields
			
			// Override the newClientFunc for tests with mockError
			if tc.mockError != nil {
				api.newClientFunc = func(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
					return nil, tc.mockError
				}
			}

			// Setup context
			ctx, cancel := tc.setupCtx()
			defer cancel()

			// Call the method being tested
			client, err := api.InitClient(ctx, tc.apiKey, tc.modelName)

			// Check error expectations
			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else {
					if !errors.Is(err, tc.wantErr) {
						t.Errorf("Expected error type %v, got %v", tc.wantErr, err)
					}

					if !strings.Contains(err.Error(), tc.wantMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tc.wantMsg, err.Error())
					}
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// For cases expecting errors, client should be nil
			if tc.wantErr != nil && client != nil {
				t.Errorf("Expected nil client when error occurs, got non-nil client")
			}
		})
	}
}

// TestProcessResponse tests the ProcessResponse method of APIService
func TestProcessResponse(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		result        *gemini.GenerationResult
		wantContent   string
		wantErr       error  // Expected error type
		wantErrSubstr string // Expected substring in error message
	}{
		{
			name: "Successful Response",
			result: &gemini.GenerationResult{
				Content:      "This is valid content",
				FinishReason: "STOP",
			},
			wantContent: "This is valid content",
			wantErr:     nil,
		},
		{
			name:          "Nil Result",
			result:        nil,
			wantContent:   "",
			wantErr:       ErrEmptyResponse,
			wantErrSubstr: "result is nil",
		},
		{
			name: "Empty Content with Finish Reason",
			result: &gemini.GenerationResult{
				Content:      "",
				FinishReason: "SAFETY",
			},
			wantContent:   "",
			wantErr:       ErrEmptyResponse,
			wantErrSubstr: "SAFETY",
		},
		{
			name: "Whitespace-only Content",
			result: &gemini.GenerationResult{
				Content:      "   \n\t   ",
				FinishReason: "STOP",
			},
			wantContent:   "",
			wantErr:       ErrWhitespaceContent,
			wantErrSubstr: "empty plan text",
		},
		{
			name: "Safety Blocked",
			result: &gemini.GenerationResult{
				Content: "",
				SafetyRatings: []gemini.SafetyRating{
					{
						Category: "HARM_CATEGORY_DANGEROUS",
						Blocked:  true,
					},
				},
			},
			wantContent:   "",
			wantErr:       ErrSafetyBlocked,
			wantErrSubstr: "HARM_CATEGORY_DANGEROUS",
		},
		{
			name: "Multiple Safety Categories",
			result: &gemini.GenerationResult{
				Content: "",
				SafetyRatings: []gemini.SafetyRating{
					{
						Category: "CATEGORY_1",
						Blocked:  true,
					},
					{
						Category: "CATEGORY_2",
						Blocked:  true,
					},
				},
			},
			wantContent:   "",
			wantErr:       ErrSafetyBlocked,
			wantErrSubstr: "Safety Blocking",
		},
		{
			name: "Safety Ratings but Not Blocked",
			result: &gemini.GenerationResult{
				Content: "",
				SafetyRatings: []gemini.SafetyRating{
					{
						Category: "CATEGORY_1",
						Blocked:  false,
					},
				},
			},
			wantContent:   "",
			wantErr:       ErrEmptyResponse,
			wantErrSubstr: "empty response",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &mockAPILogger{}
			apiService := NewAPIService(logger)

			// Process the response
			content, err := apiService.ProcessResponse(tc.result)

			// Verify error expectations
			if tc.wantErr != nil {
				if err == nil {
					t.Error("Expected error, got nil")
				} else {
					// Check error type
					if !errors.Is(err, tc.wantErr) {
						t.Errorf("Expected error type %v, got %v", tc.wantErr, err)
					}

					// Check error message contains expected substring
					if tc.wantErrSubstr != "" && !strings.Contains(err.Error(), tc.wantErrSubstr) {
						t.Errorf("Expected error message to contain %q, got %q", tc.wantErrSubstr, err.Error())
					}
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Verify content matches expectation
			if content != tc.wantContent {
				t.Errorf("Expected content %q, got %q", tc.wantContent, content)
			}
		})
	}
}

// TestErrorHelperMethods tests the error helper methods of APIService
func TestErrorHelperMethods(t *testing.T) {
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	// Test IsEmptyResponseError
	t.Run("IsEmptyResponseError", func(t *testing.T) {
		testCases := []struct {
			name     string
			err      error
			expected bool
		}{
			{
				name:     "Direct ErrEmptyResponse",
				err:      ErrEmptyResponse,
				expected: true,
			},
			{
				name:     "Wrapped ErrEmptyResponse",
				err:      fmt.Errorf("%w: some details", ErrEmptyResponse),
				expected: true,
			},
			{
				name:     "Direct ErrWhitespaceContent",
				err:      ErrWhitespaceContent,
				expected: true,
			},
			{
				name:     "Wrapped ErrWhitespaceContent",
				err:      fmt.Errorf("%w: whitespace details", ErrWhitespaceContent),
				expected: true,
			},
			{
				name:     "Deeply Wrapped ErrEmptyResponse",
				err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrEmptyResponse)),
				expected: true,
			},
			{
				name:     "ErrSafetyBlocked",
				err:      ErrSafetyBlocked,
				expected: false,
			},
			{
				name:     "Generic Error",
				err:      errors.New("some other error"),
				expected: false,
			},
			{
				name:     "Nil Error",
				err:      nil,
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := apiService.IsEmptyResponseError(tc.err)
				if result != tc.expected {
					t.Errorf("Expected IsEmptyResponseError to return %v for %v, got %v",
						tc.expected, tc.err, result)
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
				name:     "Direct ErrSafetyBlocked",
				err:      ErrSafetyBlocked,
				expected: true,
			},
			{
				name:     "Wrapped ErrSafetyBlocked",
				err:      fmt.Errorf("%w: safety details", ErrSafetyBlocked),
				expected: true,
			},
			{
				name:     "Deeply Wrapped ErrSafetyBlocked",
				err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrSafetyBlocked)),
				expected: true,
			},
			{
				name:     "ErrEmptyResponse",
				err:      ErrEmptyResponse,
				expected: false,
			},
			{
				name:     "ErrWhitespaceContent",
				err:      ErrWhitespaceContent,
				expected: false,
			},
			{
				name:     "Generic Error",
				err:      errors.New("some safety error"),
				expected: false,
			},
			{
				name:     "Nil Error",
				err:      nil,
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := apiService.IsSafetyBlockedError(tc.err)
				if result != tc.expected {
					t.Errorf("Expected IsSafetyBlockedError to return %v for %v, got %v",
						tc.expected, tc.err, result)
				}
			})
		}
	})

	// Test GetErrorDetails
	t.Run("GetErrorDetails", func(t *testing.T) {
		testCases := []struct {
			name           string
			err            error
			expectedResult string
		}{
			{
				name:           "Regular Error",
				err:            errors.New("regular error"),
				expectedResult: "regular error",
			},
			{
				name:           "Wrapped Error",
				err:            fmt.Errorf("outer: %w", errors.New("inner error")),
				expectedResult: "outer: inner error",
			},
			{
				name:           "Nil Error",
				err:            nil,
				expectedResult: "",
			},
			// We can create a gemini.APIError for testing, since it's exported
			{
				name: "API Error with Suggestion",
				err: &gemini.APIError{
					Message:    "API error message",
					Suggestion: "Try something else",
				},
				expectedResult: "API error message\n\nSuggestion: Try something else",
			},
			{
				name: "API Error without Suggestion",
				err: &gemini.APIError{
					Message: "API error message",
				},
				expectedResult: "API error message",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := apiService.GetErrorDetails(tc.err)
				// Remove special case handling as it's now handled in the implementation
				if result != tc.expectedResult {
					t.Errorf("Expected error details %q, got %q",
						tc.expectedResult, result)
				}
			})
		}
	})
}
