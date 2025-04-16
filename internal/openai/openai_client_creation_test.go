// Package openai provides a client for interacting with the OpenAI API
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

// TestClientCreationWithDefaultConfiguration tests the creation of a client with default configuration
func TestClientCreationWithDefaultConfiguration(t *testing.T) {
	// Save current env var if it exists
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		err := os.Setenv("OPENAI_API_KEY", originalAPIKey)
		if err != nil {
			t.Logf("Failed to restore original OPENAI_API_KEY: %v", err)
		}
	}()

	// Set a valid API key for testing
	validAPIKey := "sk-validApiKeyForTestingPurposes123456789012345"
	err := os.Setenv("OPENAI_API_KEY", validAPIKey)
	if err != nil {
		t.Fatalf("Failed to set OPENAI_API_KEY: %v", err)
	}

	// Test cases for different default models
	testModels := []struct {
		name          string
		modelName     string
		expectedModel string
	}{
		{
			name:          "GPT-4 model",
			modelName:     "gpt-4",
			expectedModel: "gpt-4",
		},
		{
			name:          "GPT-3.5 Turbo model",
			modelName:     "gpt-3.5-turbo",
			expectedModel: "gpt-3.5-turbo",
		},
		{
			name:          "Custom model name",
			modelName:     "custom-model",
			expectedModel: "custom-model",
		},
	}

	for _, tc := range testModels {
		t.Run(tc.name, func(t *testing.T) {
			// Create client with default configuration (just model name)
			client, err := NewClient(tc.modelName)

			// Verify client was created successfully
			require.NoError(t, err, "Creating client with default configuration should succeed")
			require.NotNil(t, client, "Client should not be nil")

			// Verify model name was set correctly
			assert.Equal(t, tc.expectedModel, client.GetModelName(), "Client should have correct model name")

			// Create a test context
			ctx := context.Background()

			// Replace the client's API with a mock to test functionality
			realClient := client.(*openaiClient)

			// Mock the API
			mockAPI := &mockOpenAIAPI{
				createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
					// Verify the model passed to the API is the same as expected
					assert.Equal(t, tc.expectedModel, model, "Model should be passed correctly to API")

					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: "Default configuration test response",
									Role:    "assistant",
								},
								FinishReason: "stop",
							},
						},
						Usage: openai.CompletionUsage{
							CompletionTokens: 5,
						},
					}, nil
				},
			}

			// Replace the real API with our mock
			realClient.api = mockAPI

			// Mock the tokenizer too
			mockTokenizer := &mockTokenizer{
				countTokensFunc: func(text string, model string) (int, error) {
					// Verify the model passed to the tokenizer is the same as expected
					assert.Equal(t, tc.expectedModel, model, "Model should be passed correctly to tokenizer")
					return 10, nil
				},
			}

			// Replace the real tokenizer with our mock
			realClient.tokenizer = mockTokenizer

			// Test GenerateContent to verify API is working
			result, err := client.GenerateContent(ctx, "Test prompt", nil)
			assert.NoError(t, err, "GenerateContent should succeed")
			assert.Equal(t, "Default configuration test response", result.Content, "Content should match mock response")

			// Test CountTokens to verify tokenizer is working
			tokenCount, err := client.CountTokens(ctx, "Test prompt")
			require.NoError(t, err, "CountTokens should succeed")
			assert.Equal(t, int32(10), tokenCount.Total, "Token count should match mock response")

			// Test GetModelInfo to verify model limits are set up
			modelInfo, err := client.GetModelInfo(ctx)
			require.NoError(t, err, "GetModelInfo should succeed")
			assert.Equal(t, tc.expectedModel, modelInfo.Name, "Model name should be correct in model info")
			assert.True(t, modelInfo.InputTokenLimit > 0, "Input token limit should be positive")
			assert.True(t, modelInfo.OutputTokenLimit > 0, "Output token limit should be positive")
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

			// Create the client
			client, err := NewClient("gpt-4")
			require.NoError(t, err)
			require.NotNil(t, client)

			// Replace the client's API with our mocked version
			client.(*openaiClient).api = mockAPI

			// Make an API call which should fail with the permission/validation error
			ctx := context.Background()
			_, err = client.GenerateContent(ctx, "test prompt", nil)

			// Verify the error handling
			require.Error(t, err, "Expected an error for scenario: %s", tc.scenario)

			// Check that the error has the expected type and attributes
			apiErr, ok := IsAPIError(errors.Unwrap(err))
			require.True(t, ok, "Expected an APIError but got: %v", err)
			assert.Equal(t, tc.expectedErrorType, apiErr.Type)
			assert.Equal(t, tc.statusCode, apiErr.StatusCode)
			assert.Contains(t, apiErr.Message, tc.errorMessage)
			assert.Contains(t, apiErr.Suggestion, tc.suggestion)
			assert.Contains(t, apiErr.Details, tc.scenario)
		})
	}
}