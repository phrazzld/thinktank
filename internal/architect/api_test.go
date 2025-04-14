package architect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
)

// mockAPILogger is a logger implementation for API tests
type mockAPILogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockAPILogger) Debug(format string, v ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, v...))
}

func (m *mockAPILogger) Info(format string, v ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, v...))
}

func (m *mockAPILogger) Warn(format string, v ...interface{}) {
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, v...))
}

func (m *mockAPILogger) Error(format string, v ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, v...))
}

func (m *mockAPILogger) Fatal(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

func (m *mockAPILogger) Println(v ...interface{}) {}

func (m *mockAPILogger) Printf(format string, v ...interface{}) {}

// TestProcessResponse tests the ProcessResponse method of APIService
func TestProcessResponse(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	tests := []struct {
		name          string
		result        *gemini.GenerationResult
		expectError   bool
		errorContains string
		expectedText  string
	}{
		{
			name:          "Nil Result",
			result:        nil,
			expectError:   true,
			errorContains: "empty response",
			expectedText:  "",
		},
		{
			name:          "Empty Content",
			result:        &gemini.GenerationResult{Content: "", FinishReason: "STOP"},
			expectError:   true,
			errorContains: "empty response",
			expectedText:  "",
		},
		{
			name: "Safety Blocked",
			result: &gemini.GenerationResult{
				Content: "",
				SafetyRatings: []gemini.SafetyRating{
					{Category: "HARM_CATEGORY_DANGEROUS", Blocked: true},
				},
				FinishReason: "SAFETY",
			},
			expectError:   true,
			errorContains: "safety filters",
			expectedText:  "",
		},
		{
			name:          "Whitespace Only",
			result:        &gemini.GenerationResult{Content: "   \n   \t  "},
			expectError:   true,
			errorContains: "empty output",
			expectedText:  "",
		},
		{
			name:          "Valid Content",
			result:        &gemini.GenerationResult{Content: "This is valid content"},
			expectError:   false,
			errorContains: "",
			expectedText:  "This is valid content",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := apiService.ProcessResponse(tc.result)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got: '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
			}

			// Check content
			if content != tc.expectedText {
				t.Errorf("Expected content '%s', got '%s'", tc.expectedText, content)
			}
		})
	}
}

// TestIsEmptyResponseError tests the IsEmptyResponseError method of APIService
func TestIsEmptyResponseError(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrEmptyResponse",
			err:      ErrEmptyResponse,
			expected: true,
		},
		{
			name:     "ErrWhitespaceContent",
			err:      ErrWhitespaceContent,
			expected: true,
		},
		{
			name:     "Wrapped ErrEmptyResponse",
			err:      errors.New("something went wrong: " + ErrEmptyResponse.Error()),
			expected: false, // Simple string wrapping doesn't work with errors.Is
		},
		{
			name:     "Properly Wrapped ErrEmptyResponse",
			err:      fmt.Errorf("wrapped: %w", ErrEmptyResponse),
			expected: true,
		},
		{
			name:     "ErrSafetyBlocked",
			err:      ErrSafetyBlocked,
			expected: false,
		},
		{
			name:     "Other Error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "Nil Error",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := apiService.IsEmptyResponseError(tc.err)
			if result != tc.expected {
				t.Errorf("IsEmptyResponseError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

// TestIsSafetyBlockedError tests the IsSafetyBlockedError method of APIService
func TestIsSafetyBlockedError(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrSafetyBlocked",
			err:      ErrSafetyBlocked,
			expected: true,
		},
		{
			name:     "Wrapped ErrSafetyBlocked",
			err:      errors.New("something went wrong: " + ErrSafetyBlocked.Error()),
			expected: false, // Simple string wrapping doesn't work with errors.Is
		},
		{
			name:     "Properly Wrapped ErrSafetyBlocked",
			err:      fmt.Errorf("wrapped: %w", ErrSafetyBlocked),
			expected: true,
		},
		{
			name:     "ErrEmptyResponse",
			err:      ErrEmptyResponse,
			expected: false,
		},
		{
			name:     "Other Error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "Nil Error",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := apiService.IsSafetyBlockedError(tc.err)
			if result != tc.expected {
				t.Errorf("IsSafetyBlockedError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

// TestGetErrorDetails tests the GetErrorDetails method of APIService
func TestGetErrorDetails(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	// Since we can't directly override the package function, we'll test with direct error types

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Regular Error",
			err:      errors.New("regular error"),
			expected: "regular error",
		},
		{
			name:     "Nil Error",
			err:      nil,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := apiService.GetErrorDetails(tc.err)
			if result != tc.expected {
				t.Errorf("GetErrorDetails(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

// Mock Gemini client implementation for testing
type mockAPIClient struct {
	generateContentFunc    func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	countTokensFunc        func(ctx context.Context, text string) (*gemini.TokenCount, error)
	getModelInfoFunc       func(ctx context.Context) (*gemini.ModelInfo, error)
	getModelNameFunc       func() string
	getTemperatureFunc     func() float32
	getMaxOutputTokensFunc func() int32
	getTopPFunc            func() float32
	closeFunc              func() error
}

func (m *mockAPIClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt)
	}
	return &gemini.GenerationResult{Content: "mock content"}, nil
}

func (m *mockAPIClient) CountTokens(ctx context.Context, text string) (*gemini.TokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, text)
	}
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *mockAPIClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &gemini.ModelInfo{InputTokenLimit: 1000}, nil
}

func (m *mockAPIClient) GetModelName() string {
	if m.getModelNameFunc != nil {
		return m.getModelNameFunc()
	}
	return "test-model"
}

func (m *mockAPIClient) GetTemperature() float32 {
	if m.getTemperatureFunc != nil {
		return m.getTemperatureFunc()
	}
	return 0.7
}

func (m *mockAPIClient) GetMaxOutputTokens() int32 {
	if m.getMaxOutputTokensFunc != nil {
		return m.getMaxOutputTokensFunc()
	}
	return 2048
}

func (m *mockAPIClient) GetTopP() float32 {
	if m.getTopPFunc != nil {
		return m.getTopPFunc()
	}
	return 0.9
}

func (m *mockAPIClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// TestInitClient tests the InitClient method of APIService with basic validation
func TestInitClient(t *testing.T) {
	// Test empty API key
	t.Run("Empty API Key", func(t *testing.T) {
		logger := &mockAPILogger{}
		apiService := NewAPIService(logger)

		client, err := apiService.InitClient(context.Background(), "", "test-model", "")

		if err == nil {
			t.Error("Expected error for empty API key, got nil")
		} else if !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Expected error to mention API key requirement, got: %v", err)
		}
		if client != nil {
			t.Error("Expected nil client for error case, got non-nil client")
		}
	})

	// Test empty model name
	t.Run("Empty Model Name", func(t *testing.T) {
		logger := &mockAPILogger{}
		apiService := NewAPIService(logger)

		client, err := apiService.InitClient(context.Background(), "test-key", "", "")

		if err == nil {
			t.Error("Expected error for empty model name, got nil")
		} else if !strings.Contains(err.Error(), "model name is required") {
			t.Errorf("Expected error to mention model name requirement, got: %v", err)
		}
		if client != nil {
			t.Error("Expected nil client for error case, got non-nil client")
		}
	})

	// Test logging custom endpoint
	t.Run("Custom Endpoint Logging", func(t *testing.T) {
		logger := &mockAPILogger{}

		// Create API service with a mock client creation function
		apiService := &apiService{
			logger: logger,
			newClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
				return &mockAPIClient{}, nil
			},
		}

		customEndpoint := "https://custom-endpoint.com"
		_, err := apiService.InitClient(context.Background(), "test-key", "test-model", customEndpoint)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check if custom endpoint was logged
		found := false
		for _, msg := range logger.debugMessages {
			if strings.Contains(msg, customEndpoint) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected custom endpoint '%s' to be logged, but it wasn't", customEndpoint)
		}
	})
}
