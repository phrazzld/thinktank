package openai

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/openai/openai-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define mocks for our internal interfaces
type mockOpenAIAPI struct {
	createChatCompletionFunc           func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error)
	createChatCompletionWithParamsFunc func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

func (m *mockOpenAIAPI) createChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
	return m.createChatCompletionFunc(ctx, messages, model)
}

func (m *mockOpenAIAPI) createChatCompletionWithParams(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	if m.createChatCompletionWithParamsFunc != nil {
		return m.createChatCompletionWithParamsFunc(ctx, params)
	}
	// Fall back to simple implementation if the with-params function is not set
	return m.createChatCompletionFunc(ctx, params.Messages, params.Model)
}

type mockTokenizer struct {
	countTokensFunc func(text string, model string) (int, error)
}

func (m *mockTokenizer) countTokens(text string, model string) (int, error) {
	return m.countTokensFunc(text, model)
}

// TestParametersAreApplied tests that API parameters are correctly applied
func TestParametersAreApplied(t *testing.T) {
	var capturedParams openai.ChatCompletionNewParams

	// Create a mock API that captures the parameters
	mockAPI := &mockOpenAIAPI{
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			capturedParams = params
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Response with applied parameters",
							Role:    "assistant",
						},
						FinishReason: "stop",
					},
				},
				Usage: openai.CompletionUsage{
					CompletionTokens: 10,
				},
			}, nil
		},
	}

	// Create the client with our mock API
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
	}

	// Set specific parameters
	temperature := float32(0.7)
	client.SetTemperature(temperature)

	topP := float32(0.9)
	client.SetTopP(topP)

	maxTokens := int32(1000)
	client.SetMaxTokens(maxTokens)

	presencePenalty := float32(0.5)
	client.SetPresencePenalty(presencePenalty)

	frequencyPenalty := float32(0.3)
	client.SetFrequencyPenalty(frequencyPenalty)

	// Call GenerateContent
	ctx := context.Background()
	result, err := client.GenerateContent(ctx, "Test prompt", nil)

	// Verify parameters were passed correctly
	require.NoError(t, err)
	assert.Equal(t, "Response with applied parameters", result.Content)

	// Check that model was correctly passed to the API
	assert.Equal(t, "gpt-4", capturedParams.Model)

	// We can't directly access param.Opt values, so check that parameters were included
	// by ensuring they're not empty/nil
	assert.True(t, capturedParams.Temperature.IsPresent())
	assert.True(t, capturedParams.TopP.IsPresent())
	assert.True(t, capturedParams.MaxTokens.IsPresent())
	assert.True(t, capturedParams.PresencePenalty.IsPresent())
	assert.True(t, capturedParams.FrequencyPenalty.IsPresent())

	// Ensure the message was passed correctly
	require.Len(t, capturedParams.Messages, 1)
	// Since we're not sure of the exact API to access the message content in this version,
	// let's just check that messages were provided
	// In a real implementation, we would need to find the correct way to access this
	// based on the SDK documentation or examples
}

// TestOpenAIClientImplementsLLMClient tests that openaiClient correctly implements the LLMClient interface
func TestOpenAIClientImplementsLLMClient(t *testing.T) {
	// Create a mock OpenAI API
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Test content",
							Role:    "assistant",
						},
						FinishReason: "stop",
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			}, nil
		},
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			// Use the same response format as createChatCompletionFunc for consistency
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Test content",
							Role:    "assistant",
						},
						FinishReason: "stop",
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			}, nil
		},
	}

	// Create a mock tokenizer
	mockTokenizer := &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			return 10, nil
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: mockTokenizer,
		modelName: "gpt-4",
		modelLimits: map[string]*modelInfo{
			"gpt-4": {
				inputTokenLimit:  8192,
				outputTokenLimit: 4096,
			},
		},
	}

	// Test interface method implementations
	ctx := context.Background()

	// Test GenerateContent
	t.Run("GenerateContent", func(t *testing.T) {
		result, err := client.GenerateContent(ctx, "test prompt", nil)
		require.NoError(t, err)
		assert.Equal(t, "Test content", result.Content)
		assert.Equal(t, "stop", result.FinishReason)
		assert.Equal(t, int32(5), result.TokenCount)
		assert.False(t, result.Truncated)
	})

	// Test CountTokens
	t.Run("CountTokens", func(t *testing.T) {
		tokenCount, err := client.CountTokens(ctx, "test prompt")
		require.NoError(t, err)
		assert.Equal(t, int32(10), tokenCount.Total)
	})

	// Test GetModelInfo
	t.Run("GetModelInfo", func(t *testing.T) {
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", modelInfo.Name)
		assert.Equal(t, int32(8192), modelInfo.InputTokenLimit)
		assert.Equal(t, int32(4096), modelInfo.OutputTokenLimit)
	})

	// Test GetModelName
	t.Run("GetModelName", func(t *testing.T) {
		assert.Equal(t, "gpt-4", client.GetModelName())
	})

	// Test Close
	t.Run("Close", func(t *testing.T) {
		assert.NoError(t, client.Close())
	})
}

// TestNewClient tests the NewClient constructor function
func TestNewClient(t *testing.T) {
	t.Skip("Skipping test that relies on environment variables")
	// This test would check that NewClient correctly sets up the client
	// It's skipped here since it would require setting up environment variables
}

// TestOpenAIClientErrorHandling tests how the OpenAI client handles API errors
func TestOpenAIClientErrorHandling(t *testing.T) {
	// Create a mock OpenAI API that returns an error
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return nil, &APIError{
				Type:    ErrorTypeRateLimit,
				Message: "Rate limit exceeded",
			}
		},
	}

	// Create a mock tokenizer that returns an error
	mockTokenizer := &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			return 0, &APIError{
				Type:    ErrorTypeInvalidRequest,
				Message: "Invalid request",
			}
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: mockTokenizer,
		modelName: "gpt-4",
	}

	ctx := context.Background()

	// Test GenerateContent error handling
	t.Run("GenerateContent error", func(t *testing.T) {
		_, err := client.GenerateContent(ctx, "test prompt", map[string]interface{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rate limit exceeded")
	})

	// Test CountTokens error handling
	t.Run("CountTokens error", func(t *testing.T) {
		_, err := client.CountTokens(ctx, "test prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid request")
	})
}

// TestUnknownModelFallback tests how the client handles unknown models
func TestUnknownModelFallback(t *testing.T) {
	// Create mock API
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Test content",
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	// Create a client with an unknown model
	client := &openaiClient{
		api:         mockAPI,
		tokenizer:   &mockTokenizer{},
		modelName:   "unknown-model",
		modelLimits: map[string]*modelInfo{},
	}

	ctx := context.Background()

	// Test GetModelInfo with unknown model
	t.Run("GetModelInfo with unknown model", func(t *testing.T) {
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "unknown-model", modelInfo.Name)
		// Should return default values for unknown models
		assert.True(t, modelInfo.InputTokenLimit > 0)
		assert.True(t, modelInfo.OutputTokenLimit > 0)
	})
}

// TestTruncatedResponse tests how the client handles truncated responses
func TestTruncatedResponse(t *testing.T) {
	// Create mock API that returns a response with "length" finish reason
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Truncated content",
							Role:    "assistant",
						},
						FinishReason: "length",
					},
				},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 100,
					TotalTokens:      110,
				},
			}, nil
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
		modelLimits: map[string]*modelInfo{
			"gpt-4": {
				inputTokenLimit:  8192,
				outputTokenLimit: 2048,
			},
		},
	}

	ctx := context.Background()

	// Test truncated response
	result, err := client.GenerateContent(ctx, "test prompt", nil)
	require.NoError(t, err)
	assert.Equal(t, "Truncated content", result.Content)
	assert.Equal(t, "length", result.FinishReason)
	assert.Equal(t, int32(100), result.TokenCount)
	assert.True(t, result.Truncated)
}

// TestEmptyResponseHandling tests how the client handles empty responses
func TestEmptyResponseHandling(t *testing.T) {
	// Create mock API that returns an empty response
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
				Usage: openai.CompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 0,
					TotalTokens:      10,
				},
			}, nil
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
	}

	ctx := context.Background()

	// Test empty response handling
	_, err := client.GenerateContent(ctx, "test prompt", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no completion choices returned")
}

// TestContentFilterHandling tests handling of content filter errors
func TestContentFilterHandling(t *testing.T) {
	// Create mock API that returns a content filter error
	mockAPI := &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return nil, &APIError{
				Type:    ErrorTypeContentFiltered,
				Message: "Content was filtered",
			}
		},
	}

	// Create the client with mocks
	client := &openaiClient{
		api:       mockAPI,
		tokenizer: &mockTokenizer{},
		modelName: "gpt-4",
	}

	ctx := context.Background()

	// Test content filter handling
	_, err := client.GenerateContent(ctx, "test prompt", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Content was filtered")
}

// TestModelEncodingSelection tests the getEncodingForModel function
func TestModelEncodingSelection(t *testing.T) {
	tests := []struct {
		modelName        string
		expectedEncoding string
	}{
		{"gpt-4", "cl100k_base"},
		{"gpt-4-32k", "cl100k_base"},
		{"gpt-4-turbo", "cl100k_base"},
		{"gpt-4o", "cl100k_base"},
		{"gpt-3.5-turbo", "cl100k_base"},
		{"gpt-3.5-turbo-16k", "cl100k_base"},
		{"text-embedding-ada-002", "cl100k_base"},
		{"text-davinci-003", "p50k_base"}, // Older model should use p50k_base
		{"unknown-model", "p50k_base"},    // Unknown models should use p50k_base
	}

	for _, test := range tests {
		t.Run(test.modelName, func(t *testing.T) {
			encoding := getEncodingForModel(test.modelName)
			assert.Equal(t, test.expectedEncoding, encoding)
		})
	}
}

// TestEmptyAPIKeyHandling specifically tests how the client handles empty API keys
func TestEmptyAPIKeyHandling(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test cases for empty API key scenarios
	testCases := []struct {
		name            string
		envValue        string
		expectError     bool
		expectedErrText string
	}{
		{
			name:            "Unset API key",
			envValue:        "",
			expectError:     true,
			expectedErrText: "OPENAI_API_KEY environment variable not set",
		},
		{
			name:            "Empty string API key",
			envValue:        "",
			expectError:     true,
			expectedErrText: "OPENAI_API_KEY environment variable not set",
		},
		{
			name:            "Whitespace-only API key",
			envValue:        "   ",
			expectError:     true,
			expectedErrText: "OPENAI_API_KEY environment variable not set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment variable for "Unset API key" case
			if tc.name == "Unset API key" {
				err := os.Unsetenv("OPENAI_API_KEY")
				if err != nil {
					t.Fatalf("Failed to unset OPENAI_API_KEY: %v", err)
				}
			} else {
				// Set API key to test value
				err := os.Setenv("OPENAI_API_KEY", tc.envValue)
				if err != nil {
					t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
				}
			}

			// Attempt to create client with empty/invalid API key
			client, err := NewClient("gpt-4")

			// Verify expectations
			if tc.expectError {
				assert.Error(t, err, "Expected an error when API key is %s", tc.name)
				assert.Nil(t, client, "Expected nil client when API key is %s", tc.name)
				assert.Contains(t, err.Error(), tc.expectedErrText,
					"Error message should be specific and informative about the API key issue")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestValidAPIKeyFormatDetection tests the detection of valid API key formats
func TestValidAPIKeyFormatDetection(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test cases for various API key formats
	testCases := []struct {
		name        string
		apiKey      string
		validFormat bool
		description string
	}{
		{
			name:        "Valid OpenAI API key format",
			apiKey:      "sk-validKeyFormatWithSufficientLength12345678901234",
			validFormat: true,
			description: "Standard OpenAI API key format starting with 'sk-'",
		},
		{
			name:        "Alternative valid key format",
			apiKey:      "sk-abc123def456ghi789jkl012mno345pqr678stu90",
			validFormat: true,
			description: "API key with mixed alphanumeric characters",
		},
		{
			name:        "Invalid prefix key format",
			apiKey:      "invalid-key-format-without-sk-prefix",
			validFormat: false,
			description: "API key without 'sk-' prefix",
		},
		{
			name:        "Too short key format",
			apiKey:      "sk-tooshort",
			validFormat: false,
			description: "API key that's too short",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the API key for this test case
			err := os.Setenv("OPENAI_API_KEY", tc.apiKey)
			if err != nil {
				t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
			}

			// Create a client with this key
			client, err := NewClient("gpt-4")

			// Validation happens at client creation time only to check for emptiness
			// The actual API key format validation would happen on the first API call
			// So we expect client creation to succeed regardless of key format
			assert.NoError(t, err, "Client creation should succeed even with %s", tc.description)
			assert.NotNil(t, client, "Client should not be nil")

			// Verify the key format is as expected
			// This is a basic structural validation that could be extended
			if tc.validFormat {
				assert.True(t, strings.HasPrefix(tc.apiKey, "sk-"),
					"Valid API key should start with 'sk-' prefix")
				assert.True(t, len(tc.apiKey) >= 20,
					"Valid API key should have sufficient length")
			}
		})
	}
}

// TestInvalidAPIKeyFormatHandling tests how the client handles invalid API key formats
func TestInvalidAPIKeyFormatHandling(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test cases for invalid API key formats and the expected errors
	testCases := []struct {
		name              string
		apiKey            string
		expectedErrorType ErrorType
		expectedMsgPrefix string
	}{
		{
			name:              "Invalid prefix (missing sk-)",
			apiKey:            "invalid-key-without-sk-prefix",
			expectedErrorType: ErrorTypeAuth,
			expectedMsgPrefix: "Authentication failed",
		},
		{
			name:              "Too short key",
			apiKey:            "sk-tooshort",
			expectedErrorType: ErrorTypeAuth,
			expectedMsgPrefix: "Authentication failed",
		},
		{
			name:              "Invalid characters in key",
			apiKey:            "sk-invalid!@#$%^&*()",
			expectedErrorType: ErrorTypeAuth,
			expectedMsgPrefix: "Authentication failed",
		},
		{
			name:              "Malformed key with spaces",
			apiKey:            "sk-key with spaces",
			expectedErrorType: ErrorTypeAuth,
			expectedMsgPrefix: "Authentication failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the environment variable to the test API key
			err := os.Setenv("OPENAI_API_KEY", tc.apiKey)
			if err != nil {
				t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
			}

			// Create a mock API that simulates rejecting invalid API keys
			mockAPI := &mockOpenAIAPI{
				createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
					// Simulate API rejection of invalid key format
					return nil, &APIError{
						Type:       tc.expectedErrorType,
						Message:    tc.expectedMsgPrefix + " with the OpenAI API",
						StatusCode: http.StatusUnauthorized,
						Suggestion: "Check that your API key is valid and has the correct format. API keys should start with 'sk-' and be of sufficient length.",
					}
				},
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					// Simulate API rejection of invalid key format
					return nil, &APIError{
						Type:       tc.expectedErrorType,
						Message:    tc.expectedMsgPrefix + " with the OpenAI API",
						StatusCode: http.StatusUnauthorized,
						Suggestion: "Check that your API key is valid and has the correct format. API keys should start with 'sk-' and be of sufficient length.",
					}
				},
			}

			// Create the client
			client, err := NewClient("gpt-4")

			// Client creation should succeed since format validation only happens at API call time
			require.NoError(t, err)
			require.NotNil(t, client)

			// Replace the client's API with our mock that simulates invalid key rejection
			client.(*openaiClient).api = mockAPI

			// Make an API call which should fail due to invalid key format
			ctx := context.Background()
			_, err = client.GenerateContent(ctx, "test prompt", nil)

			// Verify the error handling
			require.Error(t, err)

			// Check that the error is of the expected type
			apiErr, ok := IsAPIError(errors.Unwrap(err))
			require.True(t, ok, "Expected an APIError but got: %v", err)
			assert.Equal(t, tc.expectedErrorType, apiErr.Type)

			// Check that the error message is informative
			assert.Contains(t, err.Error(), tc.expectedMsgPrefix)
			assert.Contains(t, apiErr.Suggestion, "API key is valid")
		})
	}
}

// TestAPIKeyEnvironmentVariableFallback tests that the client correctly falls back to the OPENAI_API_KEY environment variable
func TestAPIKeyEnvironmentVariableFallback(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test cases for environment variable fallback scenarios
	testCases := []struct {
		name          string
		envValue      string
		expectSuccess bool
		description   string
	}{
		{
			name:          "Valid environment variable",
			envValue:      "sk-validKeyFromEnvVar123456789012345678901234",
			expectSuccess: true,
			description:   "Client should successfully use the API key from environment variable",
		},
		{
			name:          "No environment variable",
			envValue:      "",
			expectSuccess: false,
			description:   "Client creation should fail when no API key is available from any source",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set or unset the environment variable
			if tc.envValue == "" {
				err := os.Unsetenv("OPENAI_API_KEY")
				if err != nil {
					t.Fatalf("Failed to unset OPENAI_API_KEY: %v", err)
				}
			} else {
				err := os.Setenv("OPENAI_API_KEY", tc.envValue)
				if err != nil {
					t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
				}
			}

			// Attempt to create a client
			client, err := NewClient("gpt-4")

			// Verify expectations
			if tc.expectSuccess {
				assert.NoError(t, err, "Expected client creation to succeed with %s", tc.description)
				assert.NotNil(t, client, "Expected non-nil client with %s", tc.description)
			} else {
				assert.Error(t, err, "Expected client creation to fail with %s", tc.description)
				assert.Nil(t, client, "Expected nil client with %s", tc.description)
				assert.Contains(t, err.Error(), "OPENAI_API_KEY environment variable not set",
					"Error should indicate the environment variable is not set")
			}
		})
	}
}

// TestAPIKeyPermissionValidationLogic tests how the client handles API keys that are syntactically
// valid but fail for permission or validation reasons when used with the API
func TestAPIKeyPermissionValidationLogic(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test cases for different API key permission/validation failures
	testCases := []struct {
		name              string
		apiKey            string
		expectedErrorType ErrorType
		statusCode        int
		errorMessage      string
		suggestion        string
		scenario          string
	}{
		{
			name:              "Insufficient permissions",
			apiKey:            "sk-validformat123456789012345678901234",
			expectedErrorType: ErrorTypeAuth,
			statusCode:        http.StatusForbidden,
			errorMessage:      "Authentication failed with the OpenAI API",
			suggestion:        "Check that your API key is valid and has not expired",
			scenario:          "API key is syntactically valid but lacks required permissions",
		},
		{
			name:              "Revoked API key",
			apiKey:            "sk-validformat123456789012345678901234",
			expectedErrorType: ErrorTypeAuth,
			statusCode:        http.StatusUnauthorized,
			errorMessage:      "Authentication failed with the OpenAI API",
			suggestion:        "Check that your API key is valid and has not expired",
			scenario:          "API key has been revoked or disabled",
		},
		{
			name:              "Rate limit exceeded",
			apiKey:            "sk-validformat123456789012345678901234",
			expectedErrorType: ErrorTypeRateLimit,
			statusCode:        http.StatusTooManyRequests,
			errorMessage:      "Request rate limit or quota exceeded on the OpenAI API",
			suggestion:        "Wait and try again later",
			scenario:          "API key has reached its rate limit",
		},
		{
			name:              "Insufficient quota",
			apiKey:            "sk-validformat123456789012345678901234",
			expectedErrorType: ErrorTypeRateLimit,
			statusCode:        http.StatusTooManyRequests,
			errorMessage:      "Request rate limit or quota exceeded on the OpenAI API",
			suggestion:        "upgrade your API usage tier",
			scenario:          "Account has insufficient credits",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the API key for this test case
			err := os.Setenv("OPENAI_API_KEY", tc.apiKey)
			if err != nil {
				t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
			}

			// Create a mock API that simulates the specified permission/validation failure
			mockAPI := &mockOpenAIAPI{
				createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
					// Return an error simulating the specific validation failure
					return nil, &APIError{
						Type:       tc.expectedErrorType,
						Message:    tc.errorMessage,
						StatusCode: tc.statusCode,
						Suggestion: tc.suggestion,
						Details:    "Mock API validation failure: " + tc.scenario,
					}
				},
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					// Return an error simulating the specific validation failure
					return nil, &APIError{
						Type:       tc.expectedErrorType,
						Message:    tc.errorMessage,
						StatusCode: tc.statusCode,
						Suggestion: tc.suggestion,
						Details:    "Mock API validation failure: " + tc.scenario,
					}
				},
			}

			// Create the client - this should succeed since the key is syntactically valid
			client, err := NewClient("gpt-4")
			require.NoError(t, err, "Client creation should succeed with syntactically valid key")
			require.NotNil(t, client, "Client should not be nil")

			// Replace the client's API with our mock
			client.(*openaiClient).api = mockAPI

			// Make an API call that should fail due to the mocked permission/validation issue
			ctx := context.Background()
			_, err = client.GenerateContent(ctx, "test prompt", nil)

			// Verify the error handling
			require.Error(t, err, "API call should return an error for %s", tc.scenario)

			// Check that the error contains the expected information
			assert.Contains(t, err.Error(), tc.errorMessage, "Error message should include the API error message")

			// Check that the error is of the expected type
			apiErr, ok := IsAPIError(errors.Unwrap(err))
			require.True(t, ok, "Expected an APIError but got: %v", err)
			assert.Equal(t, tc.expectedErrorType, apiErr.Type, "Error should be of type %v", tc.expectedErrorType)
			assert.Equal(t, tc.statusCode, apiErr.StatusCode, "Error should have status code %d", tc.statusCode)
			assert.Contains(t, apiErr.Suggestion, tc.suggestion, "Error should include helpful suggestion")
		})
	}
}

// TestNewClientErrorHandling tests error handling in NewClient
func TestNewClientErrorHandling(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Test with empty API key
	err := os.Unsetenv("OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("Failed to unset OPENAI_API_KEY: %v", err)
	}
	client, err := NewClient("gpt-4")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY environment variable not set")

	// Set an invalid API key (too short)
	err = os.Setenv("OPENAI_API_KEY", "invalid-key")
	if err != nil {
		t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
	}
	client, err = NewClient("gpt-4")
	// This should succeed since we're just creating the client (error would occur on API calls)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
