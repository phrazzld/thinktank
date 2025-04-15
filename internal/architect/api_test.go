package architect

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
)

// mockAPILogger mocks the logger for testing
type mockAPILogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockAPILogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, formatLog(format, args...))
}

func (m *mockAPILogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, formatLog(format, args...))
}

func (m *mockAPILogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, formatLog(format, args...))
}

func (m *mockAPILogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, formatLog(format, args...))
}

func (m *mockAPILogger) Fatal(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, "FATAL: "+formatLog(format, args...))
}

func (m *mockAPILogger) Println(args ...interface{})               {}
func (m *mockAPILogger) Printf(format string, args ...interface{}) {}

func formatLog(format string, args ...interface{}) string {
	return format // Simplified for tests
}

// mockLLMClient mocks an LLM client for testing
type mockLLMClient struct {
	modelName     string
	tokenCount    int32
	responseText  string
	responseError error
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	if m.responseError != nil {
		return nil, m.responseError
	}
	return &llm.ProviderResult{
		Content:    m.responseText,
		TokenCount: 100,
	}, nil
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.responseError != nil {
		return nil, m.responseError
	}
	return &llm.ProviderTokenCount{
		Total: m.tokenCount,
	}, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.responseError != nil {
		return nil, m.responseError
	}
	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  32000,
		OutputTokenLimit: 8000,
	}, nil
}

func (m *mockLLMClient) GetModelName() string {
	return m.modelName
}

func (m *mockLLMClient) Close() error {
	return nil
}

// These functions (newGeminiClientWrapperForTest and newOpenAIClientWrapperForTest)
// are defined in api_test_helper.go

// TestNewAPIService tests the NewAPIService constructor
func TestNewAPIService(t *testing.T) {
	// Create a test logger
	logger := &mockAPILogger{}

	// Create a new APIService
	service := NewAPIService(logger)

	// Check that the service was created
	if service == nil {
		t.Error("Expected non-nil APIService")
	}

	// Check that the logger was correctly assigned
	if service.(*apiService).logger != logger {
		t.Error("Expected logger to be correctly assigned")
	}
}

// TestProcessLLMResponse tests the ProcessLLMResponse method of APIService
func TestProcessLLMResponse(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	tests := []struct {
		name          string
		result        *llm.ProviderResult
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
			result:        &llm.ProviderResult{Content: "", FinishReason: "stop"},
			expectError:   true,
			errorContains: "empty response",
			expectedText:  "",
		},
		{
			name: "Safety Blocked",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{Category: "HARM_CATEGORY_DANGEROUS", Blocked: true},
				},
				FinishReason: "safety",
			},
			expectError:   true,
			errorContains: "safety filters",
			expectedText:  "",
		},
		{
			name:          "Whitespace Only",
			result:        &llm.ProviderResult{Content: "   \n   \t  "},
			expectError:   true,
			errorContains: "empty output",
			expectedText:  "",
		},
		{
			name:          "Valid Content",
			result:        &llm.ProviderResult{Content: "This is valid content"},
			expectError:   false,
			errorContains: "",
			expectedText:  "This is valid content",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := apiService.ProcessLLMResponse(tc.result)

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
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}

			// Check content for non-error cases
			if !tc.expectError {
				if content != tc.expectedText {
					t.Errorf("Expected content '%s', got '%s'", tc.expectedText, content)
				}
			}
		})
	}
}

// TestInitLLMClient tests the InitLLMClient method of APIService
func TestInitLLMClient(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		modelName     string
		apiEndpoint   string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Missing API Key",
			apiKey:        "",
			modelName:     "gemini-pro",
			apiEndpoint:   "",
			expectError:   true,
			errorContains: "API key",
		},
		{
			name:          "Missing Model Name",
			apiKey:        "test-key",
			modelName:     "",
			apiEndpoint:   "",
			expectError:   true,
			errorContains: "model name",
		},
		{
			name:          "API Key Error",
			apiKey:        "error-key",
			modelName:     "gemini-pro",
			apiEndpoint:   "",
			expectError:   true,
			errorContains: "test API key error",
		},
		{
			name:          "Model Error",
			apiKey:        "test-key",
			modelName:     "error-model",
			apiEndpoint:   "",
			expectError:   true,
			errorContains: "test model error",
		},
		{
			name:          "Valid Gemini Model",
			apiKey:        "test-key",
			modelName:     "gemini-pro",
			apiEndpoint:   "",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "Valid OpenAI Model",
			apiKey:        "test-key",
			modelName:     "gpt-4",
			apiEndpoint:   "",
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create API service with custom client functions for testing
			logger := &mockAPILogger{}
			apiService := &apiService{
				logger:              logger,
				newGeminiClientFunc: newGeminiClientWrapperForTest,
				newOpenAIClientFunc: newOpenAIClientWrapperForTest,
			}

			// Call the method being tested
			client, err := apiService.InitLLMClient(context.Background(), tc.apiKey, tc.modelName, tc.apiEndpoint)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got: '%s'", tc.errorContains, err.Error())
				}
				if client != nil {
					t.Errorf("Expected client to be nil when error is returned")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if client == nil {
					t.Errorf("Expected client to be non-nil")
					return
				}

				// Check model name is correctly passed to client
				modelName := client.GetModelName()
				if modelName != tc.modelName {
					t.Errorf("Expected model name '%s', got '%s'", tc.modelName, modelName)
				}
			}
		})
	}
}

// TestInitLLMClientWithCustomEndpoint tests the InitLLMClient method with custom endpoint
func TestInitLLMClientWithCustomEndpoint(t *testing.T) {
	// Test passing custom endpoint
	t.Run("Custom Endpoint Logging", func(t *testing.T) {
		logger := &mockAPILogger{}
		customEndpoint := "https://custom-endpoint.com"

		// Create API service
		apiService := &apiService{
			logger:              logger,
			newGeminiClientFunc: newGeminiClientWrapperForTest,
			newOpenAIClientFunc: newOpenAIClientWrapperForTest,
		}

		// Call InitLLMClient with custom endpoint
		client, err := apiService.InitLLMClient(context.Background(), "test-key", "gemini-pro", customEndpoint)

		// Check for successful client creation
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
			return
		}

		if client == nil {
			t.Errorf("Expected non-nil client")
			return
		}

		// Get the model name which should contain the endpoint info with our test method
		modelName := client.GetModelName()
		if !strings.Contains(modelName, customEndpoint) {
			t.Errorf("Expected model name to contain endpoint info, got: %s", modelName)
		}
	})

	// Test without custom endpoint
	t.Run("No Custom Endpoint", func(t *testing.T) {
		logger := &mockAPILogger{}

		// Create API service
		apiService := &apiService{
			logger:              logger,
			newGeminiClientFunc: newGeminiClientWrapperForTest,
			newOpenAIClientFunc: newOpenAIClientWrapperForTest,
		}

		// Call InitLLMClient without custom endpoint
		client, err := apiService.InitLLMClient(context.Background(), "test-key", "gemini-pro", "")

		// Check for successful client creation
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
			return
		}

		if client == nil {
			t.Errorf("Expected non-nil client")
			return
		}

		// Get the model name
		modelName := client.GetModelName()
		if modelName != "gemini-pro" {
			t.Errorf("Expected model name 'gemini-pro', got: %s", modelName)
		}
	})
}

// TestProviderDetection tests that the InitLLMClient method correctly detects providers from model names
func TestProviderDetection(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}

	// Create a custom API service with test wrappers
	apiService := &apiService{
		logger:              logger,
		newGeminiClientFunc: newGeminiClientWrapperForTest,
		newOpenAIClientFunc: newOpenAIClientWrapperForTest,
	}

	tests := []struct {
		name        string
		modelName   string
		expectError bool
		provider    string
	}{
		{
			name:        "Gemini Model",
			modelName:   "gemini-pro",
			expectError: false,
			provider:    "Gemini",
		},
		{
			name:        "OpenAI Model",
			modelName:   "gpt-4",
			expectError: false,
			provider:    "OpenAI",
		},
		{
			name:        "Unknown Provider",
			modelName:   "unknown-model",
			expectError: true,
			provider:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset logger for each test
			logger.debugMessages = []string{}

			client, err := apiService.InitLLMClient(context.Background(), "test-key", tc.modelName, "")

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}

				// Verify model name
				if client.GetModelName() != tc.modelName {
					t.Errorf("Expected model name %s, got %s", tc.modelName, client.GetModelName())
				}
			}
		})
	}
}

// TestIsEmptyResponseError tests the IsEmptyResponseError method
func TestIsEmptyResponseError(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	// Test cases
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Empty Response Error",
			err:      errors.New("API returned empty response"),
			expected: true,
		},
		{
			name:     "Empty Content Error",
			err:      errors.New("empty content in response"),
			expected: true,
		},
		{
			name:     "Empty Output Error",
			err:      errors.New("empty output after processing"),
			expected: true,
		},
		{
			name:     "Non-Empty Error",
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
				t.Errorf("Expected IsEmptyResponseError to return %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestIsSafetyBlockedError tests the IsSafetyBlockedError method
func TestIsSafetyBlockedError(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	// Test cases
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Safety Filter Error",
			err:      errors.New("content blocked by safety filters"),
			expected: true,
		},
		{
			name:     "Content Policy Error",
			err:      errors.New("violates content policy"),
			expected: true,
		},
		{
			name:     "Safety Threshold Error",
			err:      errors.New("safety threshold exceeded"),
			expected: true,
		},
		{
			name:     "Non-Safety Error",
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
				t.Errorf("Expected IsSafetyBlockedError to return %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestGetErrorDetails tests the GetErrorDetails method
func TestGetErrorDetails(t *testing.T) {
	// Create API service
	logger := &mockAPILogger{}
	apiService := NewAPIService(logger)

	// Test cases
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "With Error",
			err:      errors.New("test error"),
			contains: "test error",
		},
		{
			name:     "Nil Error",
			err:      nil,
			contains: "no error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			details := apiService.GetErrorDetails(tc.err)
			if tc.contains != "" && !strings.Contains(strings.ToLower(details), tc.contains) {
				t.Errorf("Expected error details to contain '%s', got: '%s'", tc.contains, details)
			}
		})
	}
}
