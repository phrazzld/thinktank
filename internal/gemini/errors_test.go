// Package gemini provides a client for interacting with Google's Gemini API
package gemini

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// testError is a simple error for testing purposes
type testError struct {
	message string
}

func (e testError) Error() string {
	return e.message
}

// newTestError creates a new test error with the given message
func newTestError(message string) error {
	return testError{message: message}
}

// Test setup helper to create an APIError for testing
func createTestAPIError() *APIError {
	return &APIError{
		Original:   newTestError("original error"),
		Type:       ErrorTypeInvalidRequest,
		Message:    "Test error message",
		StatusCode: http.StatusBadRequest,
		Suggestion: "Test suggestion",
		Details:    "Test details",
	}
}

// TestAPIError_Error tests the Error method of APIError
func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		want     string
	}{
		{
			name:     "with original error",
			apiError: createTestAPIError(),
			want:     "Test error message: original error",
		},
		{
			name: "without original error",
			apiError: &APIError{
				Message: "Error without original",
			},
			want: "Error without original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Error()
			if got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAPIError_Unwrap tests the Unwrap method of APIError
func TestAPIError_Unwrap(t *testing.T) {
	originalErr := newTestError("original error")
	tests := []struct {
		name     string
		apiError *APIError
		want     error
	}{
		{
			name: "with original error",
			apiError: &APIError{
				Original: originalErr,
			},
			want: originalErr,
		},
		{
			name: "without original error",
			apiError: &APIError{
				Original: nil,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Unwrap()
			if !errors.Is(got, tt.want) {
				t.Errorf("APIError.Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAPIError_UserFacingError tests the UserFacingError method of APIError
func TestAPIError_UserFacingError(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		want     string
	}{
		{
			name:     "with suggestion",
			apiError: createTestAPIError(),
			want:     "Test error message\n\nSuggestion: Test suggestion",
		},
		{
			name: "without suggestion",
			apiError: &APIError{
				Message: "Error without suggestion",
			},
			want: "Error without suggestion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.UserFacingError()
			if got != tt.want {
				t.Errorf("APIError.UserFacingError() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAPIError_DebugInfo tests the DebugInfo method of APIError
func TestAPIError_DebugInfo(t *testing.T) {
	t.Run("full debug info", func(t *testing.T) {
		apiErr := createTestAPIError()
		info := apiErr.DebugInfo()

		// Check that all fields are included in the debug info
		expectedParts := []string{
			"Error Type: 3",
			"Message: Test error message",
			"Status Code: 400",
			"Original Error: original error",
			"Details: Test details",
			"Suggestion: Test suggestion",
		}

		for _, part := range expectedParts {
			if !strings.Contains(info, part) {
				t.Errorf("APIError.DebugInfo() missing %q", part)
			}
		}
	})

	t.Run("with partial fields", func(t *testing.T) {
		apiErr := &APIError{
			Type:    ErrorTypeAuth,
			Message: "Partial debug info",
		}
		info := apiErr.DebugInfo()

		// These should be included
		shouldContain := []string{
			"Error Type: 1",
			"Message: Partial debug info",
		}

		// These should NOT be included
		shouldNotContain := []string{
			"Status Code:",
			"Original Error:",
			"Details:",
			"Suggestion:",
		}

		for _, part := range shouldContain {
			if !strings.Contains(info, part) {
				t.Errorf("APIError.DebugInfo() missing %q", part)
			}
		}

		for _, part := range shouldNotContain {
			if strings.Contains(info, part) {
				t.Errorf("APIError.DebugInfo() should not contain %q", part)
			}
		}
	})
}

// TestIsAPIError tests the IsAPIError function
func TestIsAPIError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantErr   *APIError
		wantIsAPI bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantErr:   nil,
			wantIsAPI: false,
		},
		{
			name:      "regular error",
			err:       newTestError("regular error"),
			wantErr:   nil,
			wantIsAPI: false,
		},
		{
			name:      "APIError",
			err:       createTestAPIError(),
			wantErr:   createTestAPIError(),
			wantIsAPI: true,
		},
		{
			name: "wrapped APIError",
			err: fmt.Errorf("wrapped: %w", &APIError{
				Type:    ErrorTypeAuth,
				Message: "wrapped APIError",
			}),
			wantIsAPI: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr, gotIsAPI := IsAPIError(tt.err)

			if gotIsAPI != tt.wantIsAPI {
				t.Errorf("IsAPIError() isAPI = %v, want %v", gotIsAPI, tt.wantIsAPI)
			}

			if tt.wantIsAPI {
				if gotErr == nil {
					t.Errorf("IsAPIError() returned nil APIError when expected non-nil")
				}
			} else {
				if gotErr != nil {
					t.Errorf("IsAPIError() returned non-nil APIError when expected nil")
				}
			}
		})
	}
}

// TestGetErrorType tests the GetErrorType function
func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		want       ErrorType
	}{
		{
			name:       "nil error",
			err:        nil,
			statusCode: 0,
			want:       ErrorTypeUnknown,
		},
		// Status code based classification
		{
			name:       "unauthorized status",
			err:        errors.New("any error"),
			statusCode: http.StatusUnauthorized,
			want:       ErrorTypeAuth,
		},
		{
			name:       "forbidden status",
			err:        errors.New("any error"),
			statusCode: http.StatusForbidden,
			want:       ErrorTypeAuth,
		},
		{
			name:       "too many requests status",
			err:        errors.New("any error"),
			statusCode: http.StatusTooManyRequests,
			want:       ErrorTypeRateLimit,
		},
		{
			name:       "bad request status",
			err:        errors.New("any error"),
			statusCode: http.StatusBadRequest,
			want:       ErrorTypeInvalidRequest,
		},
		{
			name:       "not found status",
			err:        errors.New("any error"),
			statusCode: http.StatusNotFound,
			want:       ErrorTypeNotFound,
		},
		{
			name:       "server error status",
			err:        errors.New("any error"),
			statusCode: http.StatusInternalServerError,
			want:       ErrorTypeServer,
		},

		// Message based classification
		{
			name:       "rate limit in message",
			err:        errors.New("rate limit exceeded"),
			statusCode: 0,
			want:       ErrorTypeRateLimit,
		},
		{
			name:       "quota in message",
			err:        errors.New("quota exceeded"),
			statusCode: 0,
			want:       ErrorTypeRateLimit,
		},
		{
			name:       "safety in message",
			err:        errors.New("safety setting triggered"),
			statusCode: 0,
			want:       ErrorTypeContentFiltered,
		},
		{
			name:       "blocked in message",
			err:        errors.New("content blocked"),
			statusCode: 0,
			want:       ErrorTypeContentFiltered,
		},
		{
			name:       "filtered in message",
			err:        errors.New("content filtered"),
			statusCode: 0,
			want:       ErrorTypeContentFiltered,
		},
		{
			name:       "token limit in message",
			err:        errors.New("token limit exceeded"),
			statusCode: 0,
			want:       ErrorTypeInputLimit,
		},
		{
			name:       "tokens exceeds in message",
			err:        errors.New("tokens exceeds limit"),
			statusCode: 0,
			want:       ErrorTypeInputLimit,
		},
		{
			name:       "network in message",
			err:        errors.New("network error"),
			statusCode: 0,
			want:       ErrorTypeNetwork,
		},
		{
			name:       "connection in message",
			err:        errors.New("connection failed"),
			statusCode: 0,
			want:       ErrorTypeNetwork,
		},
		{
			name:       "timeout in message",
			err:        errors.New("timeout occurred"),
			statusCode: 0,
			want:       ErrorTypeNetwork,
		},
		{
			name:       "canceled in message",
			err:        errors.New("request canceled"),
			statusCode: 0,
			want:       ErrorTypeCancelled,
		},
		{
			name:       "cancelled in message (UK spelling)",
			err:        errors.New("request cancelled"),
			statusCode: 0,
			want:       ErrorTypeCancelled,
		},
		{
			name:       "deadline exceeded in message",
			err:        errors.New("deadline exceeded"),
			statusCode: 0,
			want:       ErrorTypeCancelled,
		},
		{
			name:       "unknown error",
			err:        errors.New("unknown error"),
			statusCode: 0,
			want:       ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorType(tt.err, tt.statusCode)
			if got != tt.want {
				t.Errorf("GetErrorType() = %v, want %v", got, tt.want)
			}
		})
	}
}
