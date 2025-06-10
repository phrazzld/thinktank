// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// mockLogger for testing
type mockAPILogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

// Ensure mockAPILogger implements logutil.LoggerInterface
var _ logutil.LoggerInterface = (*mockAPILogger)(nil)

func (m *mockAPILogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockAPILogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockAPILogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, format)
}

func (m *mockAPILogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *mockAPILogger) Fatal(format string, args ...interface{}) {
	// Don't actually exit in tests
}

func (m *mockAPILogger) Println(v ...interface{}) {
	// No-op for tests
}

func (m *mockAPILogger) Printf(format string, v ...interface{}) {
	// No-op for tests
}

// Context-aware logging methods
func (m *mockAPILogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockAPILogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockAPILogger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, format)
}

func (m *mockAPILogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *mockAPILogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	// Don't actually exit in tests
}

func (m *mockAPILogger) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
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
	var _ = service // This is a compile-time check
}

// Since we can no longer access internal fields, we'll depend on the
// public interface behavior to verify correctness
func TestInitLLMClient(t *testing.T) {
	// Define test cases that don't require modifying internals
	testCases := []struct {
		name      string
		apiKey    string
		modelName string
		setupCtx  func() (context.Context, context.CancelFunc)
		wantErr   bool
	}{
		{
			name:      "Empty API Key",
			apiKey:    "",
			modelName: "fake-model",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:   true,
		},
		{
			name:      "Empty Model Name",
			apiKey:    "fake-api-key",
			modelName: "",
			setupCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:   true,
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
			wantErr: true,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &mockAPILogger{}
			api := NewAPIService(logger)

			// Setup context
			ctx, cancel := tc.setupCtx()
			defer cancel()

			// Call the method being tested
			client, err := api.InitLLMClient(ctx, tc.apiKey, tc.modelName, "")

			// Check error expectations
			if tc.wantErr && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tc.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// For cases expecting errors, client should be nil
			if tc.wantErr && client != nil {
				t.Errorf("Expected nil client when error occurs, got non-nil client")
			}
		})
	}
}

// TestProcessLLMResponse tests the ProcessLLMResponse method of APIService
func TestProcessLLMResponse(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		result      *llm.ProviderResult
		wantContent string
		wantErr     bool
	}{
		{
			name: "Successful Response",
			result: &llm.ProviderResult{
				Content:      "This is valid content",
				FinishReason: "STOP",
			},
			wantContent: "This is valid content",
			wantErr:     false,
		},
		{
			name:        "Nil Result",
			result:      nil,
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "Empty Content with Finish Reason",
			result: &llm.ProviderResult{
				Content:      "",
				FinishReason: "SAFETY",
			},
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "Whitespace-only Content",
			result: &llm.ProviderResult{
				Content:      "   \n\t   ",
				FinishReason: "STOP",
			},
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "Safety Blocked",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{
						Category: "HARM_CATEGORY_DANGEROUS",
						Blocked:  true,
					},
				},
			},
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "Multiple Safety Categories",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
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
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "Safety Ratings but Not Blocked",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{
						Category: "CATEGORY_1",
						Blocked:  false,
					},
				},
			},
			wantContent: "",
			wantErr:     true, // Should still error because content is empty
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &mockAPILogger{}
			api := NewAPIService(logger)

			// Call method being tested
			content, err := api.ProcessLLMResponse(tc.result)

			// Check error expectation
			if tc.wantErr && err == nil {
				t.Errorf("Expected error, got nil")
			} else if !tc.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Check content
			if content != tc.wantContent {
				t.Errorf("Expected content %q, got %q", tc.wantContent, content)
			}

			// For safety blocked cases, verify error type
			if tc.result != nil && len(tc.result.SafetyInfo) > 0 {
				for _, safety := range tc.result.SafetyInfo {
					if safety.Blocked && !api.IsSafetyBlockedError(err) {
						t.Errorf("Expected safety blocked error for blocked content")
					}
				}
			}
		})
	}
}

// TestErrorHelperMethods tests the error helper methods
func TestErrorHelperMethods(t *testing.T) {
	// Create new service
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
				err:      fmt.Errorf("%w: details", ErrEmptyResponse),
				expected: true,
			},
			{
				name:     "ErrWhitespaceContent",
				err:      ErrWhitespaceContent,
				expected: true,
			},
			{
				name:     "Error with empty response message",
				err:      errors.New("received empty response from API"),
				expected: true,
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
				err:      errors.New("some generic error"),
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
			name     string
			err      error
			contains string
		}{
			{
				name:     "Simple Error",
				err:      errors.New("simple error"),
				contains: "simple error",
			},
			{
				name:     "Nil Error",
				err:      nil,
				contains: "no error",
			},
			{
				name:     "ErrEmptyResponse",
				err:      ErrEmptyResponse,
				contains: "empty response",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				details := apiService.GetErrorDetails(tc.err)
				if !strings.Contains(details, tc.contains) {
					t.Errorf("Expected error details to contain %q, got %q", tc.contains, details)
				}
			})
		}
	})
}
