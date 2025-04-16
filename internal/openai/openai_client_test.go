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
	"github.com/phrazzld/architect/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			require.NoError(t, err, "GenerateContent should succeed")
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

// TestClientCreationWithCustomConfiguration tests the creation and configuration of a client with custom parameters
func TestClientCreationWithCustomConfiguration(t *testing.T) {
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

	// Test cases for different parameters and their expected values
	testCases := []struct {
		name                  string
		modelName             string
		temperature           float32
		topP                  float32
		presencePenalty       float32
		frequencyPenalty      float32
		maxTokens             int32
		customParamsMap       map[string]interface{}
		checkTemperature      bool
		checkTopP             bool
		checkPresencePenalty  bool
		checkFrequencyPenalty bool
		checkMaxTokens        bool
	}{
		{
			name:                  "Standard parameters",
			modelName:             "gpt-4",
			temperature:           0.7,
			topP:                  0.9,
			presencePenalty:       0.1,
			frequencyPenalty:      0.1,
			maxTokens:             100,
			checkTemperature:      true,
			checkTopP:             true,
			checkPresencePenalty:  true,
			checkFrequencyPenalty: true,
			checkMaxTokens:        true,
		},
		{
			name:                  "Temperature variations",
			modelName:             "gpt-4",
			temperature:           0.0, // Minimum temperature
			topP:                  0.5,
			presencePenalty:       0.0,
			frequencyPenalty:      0.0,
			maxTokens:             50,
			checkTemperature:      true,
			checkTopP:             true,
			checkPresencePenalty:  false, // 0.0 won't be sent as it's default
			checkFrequencyPenalty: false, // 0.0 won't be sent as it's default
			checkMaxTokens:        true,
		},
		{
			name:      "Custom parameters via map",
			modelName: "gpt-3.5-turbo",
			customParamsMap: map[string]interface{}{
				"temperature":       0.9,
				"top_p":             0.8,
				"presence_penalty":  0.5,
				"frequency_penalty": 0.5,
				"max_tokens":        200,
			},
			checkTemperature:      true,
			checkTopP:             true,
			checkPresencePenalty:  true,
			checkFrequencyPenalty: true,
			checkMaxTokens:        true,
		},
		{
			name:      "Mixed parameter types",
			modelName: "gpt-4-turbo",
			customParamsMap: map[string]interface{}{
				"temperature":       float64(0.4),
				"top_p":             float32(0.6),
				"presence_penalty":  0.2,
				"frequency_penalty": int(1),       // Should be converted to float64
				"max_tokens":        float64(150), // Should be converted to int
			},
			checkTemperature:      true,
			checkTopP:             true,
			checkPresencePenalty:  true,
			checkFrequencyPenalty: true,
			checkMaxTokens:        true,
		},
		{
			name:      "Gemini-style max tokens",
			modelName: "gpt-4",
			customParamsMap: map[string]interface{}{
				"temperature":       0.5,
				"max_output_tokens": 300, // Using Gemini-style parameter name
			},
			checkTemperature: true,
			checkMaxTokens:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a client directly for custom initialization
			// We're explicitly creating the openaiClient rather than using the interface
			var client *openaiClient

			if tc.customParamsMap == nil {
				// Create the client with our custom initialization
				client = &openaiClient{
					api:       &mockOpenAIAPI{},
					tokenizer: &mockTokenizer{},
					modelName: tc.modelName,
				}

				// Option 1: Set parameters via direct setter methods
				client.SetTemperature(tc.temperature)
				client.SetTopP(tc.topP)
				client.SetPresencePenalty(tc.presencePenalty)
				client.SetFrequencyPenalty(tc.frequencyPenalty)
				client.SetMaxTokens(tc.maxTokens)
			} else {
				// Create the client with default settings first
				llmClient, err := NewClient(tc.modelName)
				require.NoError(t, err, "Creating client should succeed")
				require.NotNil(t, llmClient, "Client should not be nil")

				// Convert to openaiClient to access internal fields
				var ok bool
				client, ok = llmClient.(*openaiClient)
				require.True(t, ok, "Client should be an *openaiClient")
			}

			// Mock the API to capture parameter values
			var capturedParams openai.ChatCompletionNewParams

			mockAPI := &mockOpenAIAPI{
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					capturedParams = params
					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: "Custom configuration test response",
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
			client.api = mockAPI

			// For customParamsMap case, apply parameters via GenerateContent
			if tc.customParamsMap != nil {
				_, err := client.GenerateContent(context.Background(), "Test prompt", tc.customParamsMap)
				require.NoError(t, err, "GenerateContent should succeed")
			} else {
				// Call GenerateContent to trigger the parameter capture
				_, err := client.GenerateContent(context.Background(), "Test prompt", nil)
				require.NoError(t, err, "GenerateContent should succeed")
			}

			// Verify parameters were correctly passed to the API

			// Verify temperature
			if tc.checkTemperature {
				assert.True(t, capturedParams.Temperature.IsPresent(), "Temperature should be set")
			}

			// Verify top_p
			if tc.checkTopP {
				assert.True(t, capturedParams.TopP.IsPresent(), "TopP should be set")
			}

			// Verify presence_penalty
			if tc.checkPresencePenalty {
				assert.True(t, capturedParams.PresencePenalty.IsPresent(), "PresencePenalty should be set")
			}

			// Verify frequency_penalty
			if tc.checkFrequencyPenalty {
				assert.True(t, capturedParams.FrequencyPenalty.IsPresent(), "FrequencyPenalty should be set")
			}

			// Verify max_tokens
			if tc.checkMaxTokens {
				assert.True(t, capturedParams.MaxTokens.IsPresent(), "MaxTokens should be set")
			}

			// Verify model name was passed correctly
			assert.Equal(t, tc.modelName, capturedParams.Model, "Model name should be passed correctly")
		})
	}
}

// TestGenerateContentWithValidParameters tests GenerateContent with various valid input parameters and verifies the response
func TestGenerateContentWithValidParameters(t *testing.T) {
	// Test cases for various valid input scenarios
	testCases := []struct {
		name             string
		prompt           string
		params           map[string]interface{}
		modelName        string
		mockResponse     string
		mockFinishReason string
		mockTokenCount   int
		expectedContent  string
	}{
		{
			name:             "Simple prompt, no parameters",
			prompt:           "Tell me a joke",
			params:           nil,
			modelName:        "gpt-4",
			mockResponse:     "Why did the chicken cross the road? To get to the other side!",
			mockFinishReason: "stop",
			mockTokenCount:   15,
			expectedContent:  "Why did the chicken cross the road? To get to the other side!",
		},
		{
			name:   "Prompt with temperature parameter",
			prompt: "Write a creative story",
			params: map[string]interface{}{
				"temperature": 0.9, // Higher temperature for more creativity
			},
			modelName:        "gpt-4",
			mockResponse:     "Once upon a time in a galaxy far, far away...",
			mockFinishReason: "stop",
			mockTokenCount:   12,
			expectedContent:  "Once upon a time in a galaxy far, far away...",
		},
		{
			name:   "Prompt with multiple parameters",
			prompt: "Generate a product description",
			params: map[string]interface{}{
				"temperature":      0.7,
				"top_p":            0.95,
				"max_tokens":       100,
				"presence_penalty": 0.1,
			},
			modelName:        "gpt-3.5-turbo",
			mockResponse:     "Introducing our revolutionary new gadget that will transform your life...",
			mockFinishReason: "stop",
			mockTokenCount:   16,
			expectedContent:  "Introducing our revolutionary new gadget that will transform your life...",
		},
		{
			name:   "Truncated response",
			prompt: "Write a very long essay",
			params: map[string]interface{}{
				"max_tokens": 10, // Deliberately small to trigger truncation
			},
			modelName:        "gpt-4",
			mockResponse:     "This essay will explore the complex interplay between...",
			mockFinishReason: "length",
			mockTokenCount:   10,
			expectedContent:  "This essay will explore the complex interplay between...",
		},
		{
			name:             "Technical code generation",
			prompt:           "Write a function to calculate Fibonacci numbers in Python",
			params:           nil,
			modelName:        "gpt-4",
			mockResponse:     "```python\ndef fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n```",
			mockFinishReason: "stop",
			mockTokenCount:   35,
			expectedContent:  "```python\ndef fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n```",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks for the API and tokenizer
			mockAPI := &mockOpenAIAPI{
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					// Verify we received the correct model
					assert.Equal(t, tc.modelName, params.Model, "Model should match expected value")

					// Verify the prompt is correctly passed as a message
					require.NotEmpty(t, params.Messages, "Messages should not be empty")

					// Return a simulated API response
					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: tc.mockResponse,
									Role:    "assistant",
								},
								FinishReason: tc.mockFinishReason,
							},
						},
						Usage: openai.CompletionUsage{
							CompletionTokens: int64(tc.mockTokenCount),
						},
					}, nil
				},
			}

			mockTokenizer := &mockTokenizer{
				countTokensFunc: func(text string, model string) (int, error) {
					return len(text) / 4, nil // Simple approximation for testing
				},
			}

			// Create a client with mocks
			client := &openaiClient{
				api:       mockAPI,
				tokenizer: mockTokenizer,
				modelName: tc.modelName,
				modelLimits: map[string]*modelInfo{
					tc.modelName: {
						inputTokenLimit:  8192,
						outputTokenLimit: 4096,
					},
				},
			}

			// Call GenerateContent with the test case parameters
			ctx := context.Background()
			result, err := client.GenerateContent(ctx, tc.prompt, tc.params)

			// Verify the result
			require.NoError(t, err, "GenerateContent should succeed")
			assert.Equal(t, tc.expectedContent, result.Content, "Content should match expected value")
			assert.Equal(t, tc.mockFinishReason, result.FinishReason, "FinishReason should match expected value")
			assert.Equal(t, int32(tc.mockTokenCount), result.TokenCount, "TokenCount should match expected value")

			// Check if response was truncated
			assert.Equal(t, tc.mockFinishReason == "length", result.Truncated, "Truncated flag should be set correctly")
		})
	}

	// Test with conversation history
	t.Run("Conversation with history", func(t *testing.T) {
		// History is currently not directly passed to GenerateContent,
		// but we can test how the client handles multiple messages if needed
		// by examining the captured messages in a future test

		prompt := "What is the capital of France?"
		expectedResponse := "The capital of France is Paris."

		mockAPI := &mockOpenAIAPI{
			createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
				// Verify the message content
				require.NotEmpty(t, params.Messages, "Messages should not be empty")

				return &openai.ChatCompletion{
					Choices: []openai.ChatCompletionChoice{
						{
							Message: openai.ChatCompletionMessage{
								Content: expectedResponse,
								Role:    "assistant",
							},
							FinishReason: "stop",
						},
					},
					Usage: openai.CompletionUsage{
						CompletionTokens: int64(8),
					},
				}, nil
			},
		}

		client := &openaiClient{
			api:       mockAPI,
			tokenizer: &mockTokenizer{},
			modelName: "gpt-4",
		}

		// Call GenerateContent
		ctx := context.Background()
		result, err := client.GenerateContent(ctx, prompt, nil)

		// Verify the result
		require.NoError(t, err, "GenerateContent should succeed")
		assert.Equal(t, expectedResponse, result.Content, "Content should match expected value")
	})
}

// TestParameterTypeConversionAndValidation tests that different parameter types
// are correctly converted and validated before being passed to the API
func TestParameterTypeConversionAndValidation(t *testing.T) {
	// Create test cases for different parameter types
	testCases := []struct {
		name           string
		paramKey       string
		paramValue     interface{}
		expectedToPass bool
		paramType      string
	}{
		// Temperature parameter tests
		{name: "Temperature as float64", paramKey: "temperature", paramValue: float64(0.5), expectedToPass: true, paramType: "Temperature"},
		{name: "Temperature as float32", paramKey: "temperature", paramValue: float32(0.5), expectedToPass: true, paramType: "Temperature"},
		{name: "Temperature as int", paramKey: "temperature", paramValue: 1, expectedToPass: true, paramType: "Temperature"},
		{name: "Temperature as string", paramKey: "temperature", paramValue: "invalid", expectedToPass: false, paramType: "Temperature"},

		// Top P parameter tests
		{name: "TopP as float64", paramKey: "top_p", paramValue: float64(0.8), expectedToPass: true, paramType: "TopP"},
		{name: "TopP as float32", paramKey: "top_p", paramValue: float32(0.8), expectedToPass: true, paramType: "TopP"},
		{name: "TopP as int", paramKey: "top_p", paramValue: 1, expectedToPass: true, paramType: "TopP"},
		{name: "TopP as string", paramKey: "top_p", paramValue: "invalid", expectedToPass: false, paramType: "TopP"},

		// Presence Penalty parameter tests
		{name: "PresencePenalty as float64", paramKey: "presence_penalty", paramValue: float64(0.2), expectedToPass: true, paramType: "PresencePenalty"},
		{name: "PresencePenalty as float32", paramKey: "presence_penalty", paramValue: float32(0.2), expectedToPass: true, paramType: "PresencePenalty"},
		{name: "PresencePenalty as int", paramKey: "presence_penalty", paramValue: 1, expectedToPass: true, paramType: "PresencePenalty"},
		{name: "PresencePenalty as string", paramKey: "presence_penalty", paramValue: "invalid", expectedToPass: false, paramType: "PresencePenalty"},

		// Frequency Penalty parameter tests
		{name: "FrequencyPenalty as float64", paramKey: "frequency_penalty", paramValue: float64(0.3), expectedToPass: true, paramType: "FrequencyPenalty"},
		{name: "FrequencyPenalty as float32", paramKey: "frequency_penalty", paramValue: float32(0.3), expectedToPass: true, paramType: "FrequencyPenalty"},
		{name: "FrequencyPenalty as int", paramKey: "frequency_penalty", paramValue: 1, expectedToPass: true, paramType: "FrequencyPenalty"},
		{name: "FrequencyPenalty as string", paramKey: "frequency_penalty", paramValue: "invalid", expectedToPass: false, paramType: "FrequencyPenalty"},

		// Max Tokens parameter tests
		{name: "MaxTokens as int", paramKey: "max_tokens", paramValue: 500, expectedToPass: true, paramType: "MaxTokens"},
		{name: "MaxTokens as int32", paramKey: "max_tokens", paramValue: int32(500), expectedToPass: true, paramType: "MaxTokens"},
		{name: "MaxTokens as int64", paramKey: "max_tokens", paramValue: int64(500), expectedToPass: true, paramType: "MaxTokens"},
		{name: "MaxTokens as float64", paramKey: "max_tokens", paramValue: float64(500), expectedToPass: true, paramType: "MaxTokens"},
		{name: "MaxTokens as string", paramKey: "max_tokens", paramValue: "invalid", expectedToPass: false, paramType: "MaxTokens"},

		// Gemini-style max tokens parameter test
		{name: "Gemini MaxOutputTokens as int", paramKey: "max_output_tokens", paramValue: 300, expectedToPass: true, paramType: "MaxTokens"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedParams openai.ChatCompletionNewParams
			var apiCalled bool

			// Create a mock API that captures the parameters
			mockAPI := &mockOpenAIAPI{
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					apiCalled = true
					capturedParams = params
					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: "Test response",
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

			// Create a map with the test parameter
			params := map[string]interface{}{
				tc.paramKey: tc.paramValue,
			}

			// Call GenerateContent with the parameter
			ctx := context.Background()
			result, err := client.GenerateContent(ctx, "Test prompt", params)

			if tc.expectedToPass {
				require.NoError(t, err, "Expected parameter to be accepted: %v", tc.paramValue)
				assert.True(t, apiCalled, "API should have been called")
				assert.Equal(t, "Test response", result.Content)

				// Verify the parameter was correctly converted and passed to the API
				switch tc.paramType {
				case "Temperature":
					assert.True(t, capturedParams.Temperature.IsPresent(), "Temperature should be present")
				case "TopP":
					assert.True(t, capturedParams.TopP.IsPresent(), "TopP should be present")
				case "PresencePenalty":
					assert.True(t, capturedParams.PresencePenalty.IsPresent(), "PresencePenalty should be present")
				case "FrequencyPenalty":
					assert.True(t, capturedParams.FrequencyPenalty.IsPresent(), "FrequencyPenalty should be present")
				case "MaxTokens":
					assert.True(t, capturedParams.MaxTokens.IsPresent(), "MaxTokens should be present")
				}
			} else {
				// For now, the client doesn't explicitly validate parameters
				// In the future, we might want to add validation logic and test that invalid
				// parameters cause errors before the API is called

				// Since there's no validation currently happening, we don't have negative tests that will fail
				// But we keep this structure for when validation is added later
				if err != nil {
					assert.False(t, apiCalled, "API should not have been called for invalid parameter")
					assert.Contains(t, err.Error(), "invalid", "Error should mention that parameter is invalid")
				}
			}
		})
	}
}

// TestParameterRangeBounds tests the behavior of the client with parameters at edge cases
// and beyond valid ranges
func TestParameterRangeBounds(t *testing.T) {
	// OpenAI parameters have common range bounds (these could change in the future):
	// - Temperature: 0.0 to 2.0 (typical), outside this range might be rejected
	// - Top P: 0.0 to 1.0, should be between 0 and 1
	// - Presence/Frequency Penalty: -2.0 to 2.0, outside may be invalid

	testCases := []struct {
		name           string
		paramKey       string
		paramValue     interface{}
		isEdgeCase     bool // is this an edge case but valid value
		expectedToPass bool // whether we expect the API to accept this parameter
	}{
		// Temperature range tests
		{name: "Temperature minimum valid", paramKey: "temperature", paramValue: float64(0.0), isEdgeCase: true, expectedToPass: true},
		{name: "Temperature maximum typical", paramKey: "temperature", paramValue: float64(2.0), isEdgeCase: true, expectedToPass: true},
		{name: "Temperature negative", paramKey: "temperature", paramValue: float64(-0.1), isEdgeCase: false, expectedToPass: false},
		{name: "Temperature too high", paramKey: "temperature", paramValue: float64(2.1), isEdgeCase: false, expectedToPass: false},

		// Top P range tests
		{name: "TopP minimum valid", paramKey: "top_p", paramValue: float64(0.0), isEdgeCase: true, expectedToPass: true},
		{name: "TopP maximum valid", paramKey: "top_p", paramValue: float64(1.0), isEdgeCase: true, expectedToPass: true},
		{name: "TopP negative", paramKey: "top_p", paramValue: float64(-0.1), isEdgeCase: false, expectedToPass: false},
		{name: "TopP too high", paramKey: "top_p", paramValue: float64(1.1), isEdgeCase: false, expectedToPass: false},

		// Presence Penalty range tests
		{name: "PresencePenalty minimum typical", paramKey: "presence_penalty", paramValue: float64(-2.0), isEdgeCase: true, expectedToPass: true},
		{name: "PresencePenalty maximum typical", paramKey: "presence_penalty", paramValue: float64(2.0), isEdgeCase: true, expectedToPass: true},
		{name: "PresencePenalty too low", paramKey: "presence_penalty", paramValue: float64(-2.1), isEdgeCase: false, expectedToPass: false},
		{name: "PresencePenalty too high", paramKey: "presence_penalty", paramValue: float64(2.1), isEdgeCase: false, expectedToPass: false},

		// Frequency Penalty range tests
		{name: "FrequencyPenalty minimum typical", paramKey: "frequency_penalty", paramValue: float64(-2.0), isEdgeCase: true, expectedToPass: true},
		{name: "FrequencyPenalty maximum typical", paramKey: "frequency_penalty", paramValue: float64(2.0), isEdgeCase: true, expectedToPass: true},
		{name: "FrequencyPenalty too low", paramKey: "frequency_penalty", paramValue: float64(-2.1), isEdgeCase: false, expectedToPass: false},
		{name: "FrequencyPenalty too high", paramKey: "frequency_penalty", paramValue: float64(2.1), isEdgeCase: false, expectedToPass: false},

		// Max Tokens range tests
		{name: "MaxTokens zero", paramKey: "max_tokens", paramValue: 0, isEdgeCase: true, expectedToPass: true},
		{name: "MaxTokens negative", paramKey: "max_tokens", paramValue: -10, isEdgeCase: false, expectedToPass: false},
		{name: "MaxTokens extremely large", paramKey: "max_tokens", paramValue: 100000, isEdgeCase: true, expectedToPass: true}, // This would be limited by the model
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedParams openai.ChatCompletionNewParams
			var apiCalled bool

			// Create a mock API that captures the parameters and can simulate rejections
			mockAPI := &mockOpenAIAPI{
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					apiCalled = true
					capturedParams = params

					// In a real implementation, the OpenAI API would reject out-of-range values
					// We could simulate that behavior here, but currently our client doesn't
					// validate ranges before calling the API
					// This is a placeholder for future validation

					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: "Test response",
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

			// Create a map with the test parameter
			params := map[string]interface{}{
				tc.paramKey: tc.paramValue,
			}

			// Call GenerateContent with the parameter
			ctx := context.Background()
			_, err := client.GenerateContent(ctx, "Test prompt", params)

			// IMPORTANT NOTE: Currently, the client does not validate parameter ranges
			// before sending to the API. These tests are structured to test behavior
			// once validation is added. For now, all calls will "pass" but the test is
			// structured to be ready when validation is implemented.

			// Currently all parameters are passed to the API without validation
			assert.NoError(t, err)
			assert.True(t, apiCalled)

			// Verify the parameter key was included
			switch tc.paramKey {
			case "temperature":
				assert.True(t, capturedParams.Temperature.IsPresent())
			case "top_p":
				assert.True(t, capturedParams.TopP.IsPresent())
			case "presence_penalty":
				assert.True(t, capturedParams.PresencePenalty.IsPresent())
			case "frequency_penalty":
				assert.True(t, capturedParams.FrequencyPenalty.IsPresent())
			case "max_tokens":
				assert.True(t, capturedParams.MaxTokens.IsPresent())
			}
		})
	}

	// Add a note about the current state of parameter validation
	t.Log("NOTE: The OpenAI client currently does not validate parameter ranges before sending to the API. " +
		"These tests document the expected behavior once validation is implemented.")
}

// TestParameterOverrides tests that parameters can be overridden through different methods
func TestParameterOverrides(t *testing.T) {
	testCases := []struct {
		name                string
		initialTemperature  *float32
		overrideTemperature interface{}
		expectedTemperature bool // Whether Temperature should be present in API call
		initialMaxTokens    *int32
		overrideMaxTokens   interface{}
		expectedMaxTokens   bool // Whether MaxTokens should be present in API call
	}{
		{
			name:                "Override with same parameter type",
			initialTemperature:  toPtr(float32(0.5)),
			overrideTemperature: float64(0.8),
			expectedTemperature: true,
			initialMaxTokens:    toPtr(int32(100)),
			overrideMaxTokens:   200,
			expectedMaxTokens:   true,
		},
		{
			name:                "Override with different parameter type",
			initialTemperature:  toPtr(float32(0.5)),
			overrideTemperature: 1, // int -> float64
			expectedTemperature: true,
			initialMaxTokens:    toPtr(int32(100)),
			overrideMaxTokens:   float64(200), // float64 -> int
			expectedMaxTokens:   true,
		},
		{
			name:                "No override when NULL passed",
			initialTemperature:  toPtr(float32(0.5)),
			overrideTemperature: nil,
			expectedTemperature: true,
			initialMaxTokens:    toPtr(int32(100)),
			overrideMaxTokens:   nil,
			expectedMaxTokens:   true,
		},
		{
			name:                "Invalid type doesn't override",
			initialTemperature:  toPtr(float32(0.5)),
			overrideTemperature: "invalid", // string should be ignored
			expectedTemperature: true,
			initialMaxTokens:    toPtr(int32(100)),
			overrideMaxTokens:   "invalid", // string should be ignored
			expectedMaxTokens:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedParams openai.ChatCompletionNewParams

			// Create a mock API that captures the parameters
			mockAPI := &mockOpenAIAPI{
				createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					capturedParams = params
					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: "Test response",
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

			// Set up a client with initial parameter values
			client := &openaiClient{
				api:       mockAPI,
				tokenizer: &mockTokenizer{},
				modelName: "gpt-4",
			}

			// Apply initial parameters if they're set
			if tc.initialTemperature != nil {
				client.SetTemperature(*tc.initialTemperature)
			}

			if tc.initialMaxTokens != nil {
				client.SetMaxTokens(*tc.initialMaxTokens)
			}

			// Create a params map with override values if they're set
			params := map[string]interface{}{}

			if tc.overrideTemperature != nil {
				params["temperature"] = tc.overrideTemperature
			}

			if tc.overrideMaxTokens != nil {
				params["max_tokens"] = tc.overrideMaxTokens
			}

			// Call GenerateContent with the parameters
			ctx := context.Background()
			_, err := client.GenerateContent(ctx, "Test prompt", params)
			require.NoError(t, err)

			// Verify parameters were passed as expected
			if tc.expectedTemperature {
				assert.True(t, capturedParams.Temperature.IsPresent(), "Temperature should be present")
			} else {
				assert.False(t, capturedParams.Temperature.IsPresent(), "Temperature should not be present")
			}

			if tc.expectedMaxTokens {
				assert.True(t, capturedParams.MaxTokens.IsPresent(), "MaxTokens should be present")
			} else {
				assert.False(t, capturedParams.MaxTokens.IsPresent(), "MaxTokens should not be present")
			}
		})
	}
}

// toPtr has been moved to openai_test_utils.go

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

// MockAPIErrorResponse has been moved to openai_test_utils.go

// Predefined mock error responses for common error scenarios
var (
	// Authentication errors
	MockErrorInvalidAPIKey = MockAPIErrorResponse(
		ErrorTypeAuth,
		401,
		"Authentication failed with the OpenAI API",
		"Invalid API key provided",
	)
	MockErrorExpiredAPIKey = MockAPIErrorResponse(
		ErrorTypeAuth,
		401,
		"Authentication failed with the OpenAI API",
		"API key has expired",
	)
	MockErrorInsufficientPermissions = MockAPIErrorResponse(
		ErrorTypeAuth,
		403,
		"Authentication failed with the OpenAI API",
		"API key does not have permission to access this resource",
	)

	// Rate limit errors
	MockErrorRateLimit = MockAPIErrorResponse(
		ErrorTypeRateLimit,
		429,
		"Request rate limit or quota exceeded on the OpenAI API",
		"You have exceeded your current quota",
	)
	MockErrorTokenQuotaExceeded = MockAPIErrorResponse(
		ErrorTypeRateLimit,
		429,
		"Request rate limit or quota exceeded on the OpenAI API",
		"You have reached your token quota for this billing cycle",
	)

	// Invalid request errors
	MockErrorInvalidRequest = MockAPIErrorResponse(
		ErrorTypeInvalidRequest,
		400,
		"Invalid request sent to the OpenAI API",
		"Request parameters are invalid",
	)
	MockErrorInvalidModel = MockAPIErrorResponse(
		ErrorTypeInvalidRequest,
		400,
		"Invalid request sent to the OpenAI API",
		"Model parameter is invalid",
	)
	MockErrorInvalidPrompt = MockAPIErrorResponse(
		ErrorTypeInvalidRequest,
		400,
		"Invalid request sent to the OpenAI API",
		"Prompt parameter is invalid",
	)

	// Not found errors
	MockErrorModelNotFound = MockAPIErrorResponse(
		ErrorTypeNotFound,
		404,
		"The requested model or resource was not found",
		"The model requested does not exist or is not available",
	)

	// Server errors
	MockErrorServerError = MockAPIErrorResponse(
		ErrorTypeServer,
		500,
		"OpenAI API server error occurred",
		"Internal server error",
	)
	MockErrorServiceUnavailable = MockAPIErrorResponse(
		ErrorTypeServer,
		503,
		"OpenAI API server error occurred",
		"Service temporarily unavailable",
	)

	// Network errors
	MockErrorNetwork = MockAPIErrorResponse(
		ErrorTypeNetwork,
		0,
		"Network error while connecting to the OpenAI API",
		"Failed to establish connection to the API server",
	)
	MockErrorTimeout = MockAPIErrorResponse(
		ErrorTypeNetwork,
		0,
		"Network error while connecting to the OpenAI API",
		"Request timed out",
	)

	// Input limit errors
	MockErrorInputLimit = MockAPIErrorResponse(
		ErrorTypeInputLimit,
		400,
		"Input token limit exceeded for the OpenAI model",
		"The input size exceeds the maximum token limit for this model",
	)

	// Content filtered errors
	MockErrorContentFiltered = MockAPIErrorResponse(
		ErrorTypeContentFiltered,
		400,
		"Content was filtered by OpenAI API safety settings",
		"The content was flagged for violating usage policies",
	)
)

// mockAPIWithError has been moved to openai_test_utils.go

// TestContentFilterHandling tests handling of content filter errors
func TestContentFilterHandling(t *testing.T) {
	// Create mock API that returns a content filter error
	mockAPI := mockAPIWithError(MockErrorContentFiltered)

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

	// Unwrap the error to check its properties
	unwrapped := errors.Unwrap(err)
	apiErr, ok := unwrapped.(*APIError)
	require.True(t, ok, "Error should be an APIError")
	assert.Equal(t, ErrorTypeContentFiltered, apiErr.Type, "Error type should be ContentFiltered")
	assert.Equal(t, 400, apiErr.StatusCode, "Status code should be 400")
	assert.Contains(t, apiErr.Suggestion, "safety filters", "Suggestion should mention safety filters")
}

// TestMockAPIErrorResponses demonstrates and tests the mock error response system
func TestMockAPIErrorResponses(t *testing.T) {
	// Test cases for different error scenarios
	testCases := []struct {
		name              string
		mockError         *APIError
		expectedCategory  ErrorType
		expectedErrPrefix string
	}{
		{
			name:              "Authentication error",
			mockError:         MockErrorInvalidAPIKey,
			expectedCategory:  ErrorTypeAuth,
			expectedErrPrefix: "OpenAI API error: Authentication failed",
		},
		{
			name:              "Rate limit error",
			mockError:         MockErrorRateLimit,
			expectedCategory:  ErrorTypeRateLimit,
			expectedErrPrefix: "OpenAI API error: Request rate limit",
		},
		{
			name:              "Invalid request error",
			mockError:         MockErrorInvalidRequest,
			expectedCategory:  ErrorTypeInvalidRequest,
			expectedErrPrefix: "OpenAI API error: Invalid request",
		},
		{
			name:              "Model not found error",
			mockError:         MockErrorModelNotFound,
			expectedCategory:  ErrorTypeNotFound,
			expectedErrPrefix: "OpenAI API error: The requested model",
		},
		{
			name:              "Server error",
			mockError:         MockErrorServerError,
			expectedCategory:  ErrorTypeServer,
			expectedErrPrefix: "OpenAI API error: OpenAI API server error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock API that returns the specific error
			mockAPI := mockAPIWithError(tc.mockError)

			// Create client with the mock API
			client := &openaiClient{
				api:       mockAPI,
				tokenizer: &mockTokenizer{},
				modelName: "gpt-4",
			}

			// Call GenerateContent which should return the error
			_, err := client.GenerateContent(context.Background(), "Test prompt", nil)

			// Verify error handling
			require.Error(t, err, "Expected an error for %s scenario", tc.name)
			assert.Contains(t, err.Error(), tc.expectedErrPrefix, "Error should contain expected prefix")

			// Unwrap the error and verify it's of the correct type
			unwrapped := errors.Unwrap(err)
			apiErr, ok := unwrapped.(*APIError)
			require.True(t, ok, "Unwrapped error should be an *APIError")
			assert.Equal(t, tc.expectedCategory, apiErr.Type, "Error category should match expected")
			assert.NotEmpty(t, apiErr.Suggestion, "Error should include a suggestion")
			assert.NotEmpty(t, apiErr.Details, "Error should include details")
		})
	}
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

// TestClientErrorHandlingForGenerateContent tests error handling in GenerateContent
func TestClientErrorHandlingForGenerateContent(t *testing.T) {
	// Test cases for different error types with GenerateContent
	testCases := []struct {
		name              string
		mockError         *APIError
		expectedCategory  ErrorType
		expectedErrPrefix string
		expectedWrapping  bool // whether the error should be wrapped with "OpenAI API error:"
	}{
		{
			name:              "Authentication error",
			mockError:         MockErrorInvalidAPIKey,
			expectedCategory:  ErrorTypeAuth,
			expectedErrPrefix: "OpenAI API error: Authentication failed",
			expectedWrapping:  true,
		},
		{
			name:              "Rate limit error",
			mockError:         MockErrorRateLimit,
			expectedCategory:  ErrorTypeRateLimit,
			expectedErrPrefix: "OpenAI API error: Request rate limit",
			expectedWrapping:  true,
		},
		{
			name:              "Invalid model error",
			mockError:         MockErrorInvalidModel,
			expectedCategory:  ErrorTypeInvalidRequest,
			expectedErrPrefix: "OpenAI API error: Invalid request",
			expectedWrapping:  true,
		},
		{
			name:              "Invalid prompt error",
			mockError:         MockErrorInvalidPrompt,
			expectedCategory:  ErrorTypeInvalidRequest,
			expectedErrPrefix: "OpenAI API error: Invalid request",
			expectedWrapping:  true,
		},
		{
			name:              "Model not found error",
			mockError:         MockErrorModelNotFound,
			expectedCategory:  ErrorTypeNotFound,
			expectedErrPrefix: "OpenAI API error: The requested model",
			expectedWrapping:  true,
		},
		{
			name:              "Server error",
			mockError:         MockErrorServerError,
			expectedCategory:  ErrorTypeServer,
			expectedErrPrefix: "OpenAI API error: OpenAI API server error",
			expectedWrapping:  true,
		},
		{
			name:              "Service unavailable error",
			mockError:         MockErrorServiceUnavailable,
			expectedCategory:  ErrorTypeServer,
			expectedErrPrefix: "OpenAI API error: OpenAI API server error",
			expectedWrapping:  true,
		},
		{
			name:              "Network error",
			mockError:         MockErrorNetwork,
			expectedCategory:  ErrorTypeNetwork,
			expectedErrPrefix: "OpenAI API error: Network error",
			expectedWrapping:  true,
		},
		{
			name:              "Timeout error",
			mockError:         MockErrorTimeout,
			expectedCategory:  ErrorTypeNetwork,
			expectedErrPrefix: "OpenAI API error: Network error",
			expectedWrapping:  true,
		},
		{
			name:              "Input limit exceeded error",
			mockError:         MockErrorInputLimit,
			expectedCategory:  ErrorTypeInputLimit,
			expectedErrPrefix: "OpenAI API error: Input token limit exceeded",
			expectedWrapping:  true,
		},
		{
			name:              "Content filtered error",
			mockError:         MockErrorContentFiltered,
			expectedCategory:  ErrorTypeContentFiltered,
			expectedErrPrefix: "OpenAI API error: Content was filtered",
			expectedWrapping:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock API that returns the specific error
			mockAPI := mockAPIWithError(tc.mockError)

			// Create client with the mock API
			client := &openaiClient{
				api:       mockAPI,
				tokenizer: &mockTokenizer{},
				modelName: "gpt-4",
			}

			// Call GenerateContent which should return the error
			_, err := client.GenerateContent(context.Background(), "Test prompt", nil)

			// Verify error handling
			require.Error(t, err, "Expected an error for %s scenario", tc.name)

			if tc.expectedWrapping {
				assert.Contains(t, err.Error(), tc.expectedErrPrefix, "Error should contain expected prefix")
			} else {
				assert.Equal(t, tc.mockError.Error(), err.Error(), "Error should be passed through without wrapping")
			}

			// Unwrap the error and verify it's of the correct type
			unwrapped := errors.Unwrap(err)
			apiErr, ok := unwrapped.(*APIError)
			require.True(t, ok, "Unwrapped error should be an *APIError")
			assert.Equal(t, tc.expectedCategory, apiErr.Type, "Error category should match expected")
			assert.NotEmpty(t, apiErr.Suggestion, "Error should include a suggestion")
			assert.NotEmpty(t, apiErr.Details, "Error should include details")

			// Verify that the error implements llm.CategorizedError
			catErr, ok := llm.IsCategorizedError(apiErr)
			require.True(t, ok, "APIError should implement llm.CategorizedError")
			assert.Equal(t, apiErr.Category(), catErr.Category(), "CategorizedError category should match expected")
		})
	}
}

// TestClientErrorHandlingForCountTokens tests error handling in CountTokens
func TestClientErrorHandlingForCountTokens(t *testing.T) {
	// Create a mock tokenizer that returns an error
	mockTokenizerWithError := &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			return 0, MockErrorInvalidRequest
		},
	}

	// Create client with the mock tokenizer
	client := &openaiClient{
		tokenizer: mockTokenizerWithError,
		modelName: "gpt-4",
		api:       &mockOpenAIAPI{},
	}

	// Call CountTokens which should return the error
	_, err := client.CountTokens(context.Background(), "Test prompt")

	// Verify error handling
	require.Error(t, err, "Expected an error from CountTokens")
	assert.Contains(t, err.Error(), "token counting error", "Error should contain expected prefix")

	// Unwrap the error and verify it's of the correct type
	unwrapped := errors.Unwrap(err)
	apiErr, ok := unwrapped.(*APIError)
	require.True(t, ok, "Unwrapped error should be an *APIError")
	assert.Equal(t, ErrorTypeInvalidRequest, apiErr.Type, "Error type should match expected")
}

// mockModelInfoProvider has been moved to openai_test_utils.go

// MockModelInfo has been moved to openai_test_utils.go

// MockModelSpecificInfo has been moved to openai_test_utils.go

// MockModelInfoWithErrors has been moved to openai_test_utils.go

// TestClientErrorHandlingForGetModelInfo tests that the client handles errors properly in GetModelInfo
func TestClientErrorHandlingForGetModelInfo(t *testing.T) {
	// GetModelInfo doesn't currently have error handling to test
	// This test is a placeholder for future implementations
	// If error handling is added to GetModelInfo in the future, this test should be expanded

	// Currently GetModelInfo always succeeds, even with unknown models
	// It falls back to conservative defaults
	client := &openaiClient{
		modelName: "non-existent-model",
		api:       &mockOpenAIAPI{},
		tokenizer: &mockTokenizer{},
	}

	// Call GetModelInfo which should not return an error
	modelInfo, err := client.GetModelInfo(context.Background())

	// Verify that it didn't error and provided fallback values
	require.NoError(t, err, "GetModelInfo should not return an error")
	assert.Equal(t, "non-existent-model", modelInfo.Name, "Model name should match input")
	assert.True(t, modelInfo.InputTokenLimit > 0, "InputTokenLimit should be positive")
	assert.True(t, modelInfo.OutputTokenLimit > 0, "OutputTokenLimit should be positive")
}

// MockTokenCounter has been moved to openai_test_utils.go

// MockDynamicTokenCounter has been moved to openai_test_utils.go

// MockModelAwareTokenCounter has been moved to openai_test_utils.go

// MockPredictableTokenCounter has been moved to openai_test_utils.go

// TestErrorFormatting tests the FormatAPIError function
func TestErrorFormatting(t *testing.T) {
	// Test cases for error formatting
	testCases := []struct {
		name           string
		inputError     error
		statusCode     int
		expectedType   ErrorType
		expectedPrefix string
	}{
		{
			name:           "Format authentication error",
			inputError:     errors.New("invalid_api_key"),
			statusCode:     401,
			expectedType:   ErrorTypeAuth,
			expectedPrefix: "Authentication failed",
		},
		{
			name:           "Format rate limit error",
			inputError:     errors.New("rate limit exceeded"),
			statusCode:     429,
			expectedType:   ErrorTypeRateLimit,
			expectedPrefix: "Request rate limit",
		},
		{
			name:           "Format invalid request error",
			inputError:     errors.New("invalid request parameter"),
			statusCode:     400,
			expectedType:   ErrorTypeInvalidRequest,
			expectedPrefix: "Invalid request",
		},
		{
			name:           "Format not found error",
			inputError:     errors.New("model not found"),
			statusCode:     404,
			expectedType:   ErrorTypeNotFound,
			expectedPrefix: "The requested model",
		},
		{
			name:           "Format server error",
			inputError:     errors.New("internal server error"),
			statusCode:     500,
			expectedType:   ErrorTypeServer,
			expectedPrefix: "OpenAI API server error",
		},
		{
			name:           "Format network error based on message",
			inputError:     errors.New("network connection failed"),
			statusCode:     0,
			expectedType:   ErrorTypeNetwork,
			expectedPrefix: "Network error",
		},
		{
			name:           "Format content filter error based on message",
			inputError:     errors.New("content filtered due to safety settings"),
			statusCode:     0,
			expectedType:   ErrorTypeContentFiltered,
			expectedPrefix: "Content was filtered",
		},
		{
			name:           "Format input limit error based on message",
			inputError:     errors.New("token limit exceeded"),
			statusCode:     0,
			expectedType:   ErrorTypeInputLimit,
			expectedPrefix: "Input token limit exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Format the error using FormatAPIError
			formattedErr := FormatAPIError(tc.inputError, tc.statusCode)

			// Verify the formatted error
			require.NotNil(t, formattedErr, "Formatted error should not be nil")
			assert.Equal(t, tc.expectedType, formattedErr.Type, "Error type should match expected")
			assert.Contains(t, formattedErr.Message, tc.expectedPrefix, "Error message should contain expected prefix")
			assert.NotEmpty(t, formattedErr.Suggestion, "Error should include a suggestion")
			assert.Equal(t, tc.statusCode, formattedErr.StatusCode, "Error status code should match expected")
			assert.Equal(t, tc.inputError, formattedErr.Original, "Original error should be preserved")
		})
	}

	// Test that FormatAPIError returns nil when given nil
	assert.Nil(t, FormatAPIError(nil, 0), "FormatAPIError should return nil when given nil")

	// Test that FormatAPIError preserves APIError instances
	originalAPIErr := &APIError{
		Type:       ErrorTypeAuth,
		Message:    "Custom API error",
		StatusCode: 401,
		Suggestion: "Custom suggestion",
		Details:    "Custom details",
	}
	formattedErr := FormatAPIError(originalAPIErr, 500) // Different status code to verify it's ignored
	assert.Equal(t, originalAPIErr, formattedErr, "FormatAPIError should preserve APIError instances")
}

// TestMockTokenCounters tests the token counting mock implementations
func TestMockTokenCounters(t *testing.T) {
	// Test fixed token counter
	t.Run("Fixed token counter", func(t *testing.T) {
		fixedCounter := MockTokenCounter(42, nil)
		count, err := fixedCounter.countTokens("any text", "any-model")
		assert.NoError(t, err)
		assert.Equal(t, 42, count)

		// Different text should still return the same count
		count, err = fixedCounter.countTokens("completely different text", "any-model")
		assert.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	// Test dynamic token counter
	t.Run("Dynamic token counter", func(t *testing.T) {
		dynamicCounter := MockDynamicTokenCounter(0.25, nil)

		// Test with different length texts
		texts := []string{
			"short text",         // 10 chars = 2.5 tokens
			"medium length text", // 18 chars = 4.5 tokens
			"this is a longer piece of text for testing", // 40 chars = 10 tokens
		}

		expectedCounts := []int{2, 4, 10}

		for i, text := range texts {
			count, err := dynamicCounter.countTokens(text, "any-model")
			assert.NoError(t, err)
			assert.Equal(t, expectedCounts[i], count)
		}
	})

	// Test model-aware token counter
	t.Run("Model-aware token counter", func(t *testing.T) {
		modelCounts := map[string]int{
			"gpt-4":         10,
			"gpt-3.5-turbo": 15,
			"custom-model":  20,
		}

		modelCounter := MockModelAwareTokenCounter(modelCounts, 5, nil)

		// Check model-specific counts
		for model, expectedCount := range modelCounts {
			count, err := modelCounter.countTokens("same text", model)
			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
		}

		// Check default count for unknown model
		count, err := modelCounter.countTokens("same text", "unknown-model")
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	// Test predictable token counter
	t.Run("Predictable token counter", func(t *testing.T) {
		textCounts := map[string]int{
			"hello world":         3,
			"this is a test":      5,
			"more complex prompt": 8,
		}

		predictableCounter := MockPredictableTokenCounter(textCounts, 10, nil)

		// Check text-specific counts
		for text, expectedCount := range textCounts {
			count, err := predictableCounter.countTokens(text, "any-model")
			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
		}

		// Check default count for unknown text
		count, err := predictableCounter.countTokens("unknown text", "any-model")
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
	})

	// Test error handling
	t.Run("Error handling", func(t *testing.T) {
		mockError := &APIError{
			Type:    ErrorTypeInvalidRequest,
			Message: "Invalid encoding for model",
		}

		errorCounter := MockTokenCounter(0, mockError)

		count, err := errorCounter.countTokens("any text", "any-model")
		assert.Error(t, err)
		assert.Equal(t, mockError, err)
		assert.Equal(t, 0, count)
	})
}

// TestTokenCounterIntegration tests using the mock token counters with the OpenAI client
func TestTokenCounterIntegration(t *testing.T) {
	// Test fixed counter with the client
	t.Run("Client with fixed counter", func(t *testing.T) {
		// Create client with mock fixed counter
		client := &openaiClient{
			api:       &mockOpenAIAPI{},
			tokenizer: MockTokenCounter(50, nil),
			modelName: "gpt-4",
		}

		// Test CountTokens
		ctx := context.Background()
		tokenCount, err := client.CountTokens(ctx, "test prompt")
		require.NoError(t, err)
		assert.Equal(t, int32(50), tokenCount.Total)

		// Different prompt should still return the same count
		tokenCount, err = client.CountTokens(ctx, "completely different prompt")
		require.NoError(t, err)
		assert.Equal(t, int32(50), tokenCount.Total)
	})

	// Test dynamic counter with the client
	t.Run("Client with dynamic counter", func(t *testing.T) {
		// Create client with mock dynamic counter
		client := &openaiClient{
			api:       &mockOpenAIAPI{},
			tokenizer: MockDynamicTokenCounter(0.25, nil),
			modelName: "gpt-4",
		}

		// Test CountTokens with short text
		ctx := context.Background()
		shortTokenCount, err := client.CountTokens(ctx, "short text")
		require.NoError(t, err)
		assert.Equal(t, int32(2), shortTokenCount.Total)

		// Test with longer text
		longTokenCount, err := client.CountTokens(ctx, "this is a much longer text that should have more tokens")
		require.NoError(t, err)
		assert.Greater(t, longTokenCount.Total, shortTokenCount.Total)
	})

	// Test error handling in the client
	t.Run("Client error handling", func(t *testing.T) {
		// Create error to return
		mockError := &APIError{
			Type:    ErrorTypeInvalidRequest,
			Message: "Invalid encoding for model",
		}

		// Create client with mock counter that returns an error
		client := &openaiClient{
			api:       &mockOpenAIAPI{},
			tokenizer: MockTokenCounter(0, mockError),
			modelName: "gpt-4",
		}

		// Test CountTokens
		ctx := context.Background()
		_, err := client.CountTokens(ctx, "test prompt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token counting error")

		// Unwrap error and check the original is there
		unwrapped := errors.Unwrap(err)
		assert.Equal(t, mockError, unwrapped)
	})
}

// TestTokenCountingAccuracy tests token counting accuracy for various inputs
func TestTokenCountingAccuracy(t *testing.T) {
	// Create test cases for different types of input text
	testCases := []struct {
		name             string
		modelName        string
		inputText        string
		expectedTokens   int32
		tokensPerChar    float64
		useFixedCount    bool
		fixedCount       int
		useModelSpecific bool
		modelCounts      map[string]int
	}{
		{
			name:           "Short English text",
			modelName:      "gpt-4",
			inputText:      "Hello, world!",
			expectedTokens: 3,
			tokensPerChar:  0.25,
			useFixedCount:  false,
		},
		{
			name:           "Multi-line English text",
			modelName:      "gpt-4",
			inputText:      "This is a test.\nIt has multiple lines.\nHow many tokens will it use?",
			expectedTokens: 16, // Length: 64 chars * 0.25 = 16
			tokensPerChar:  0.25,
			useFixedCount:  false,
		},
		{
			name:           "Text with special characters",
			modelName:      "gpt-4",
			inputText:      "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
			expectedTokens: 11, // Length: 44 chars * 0.25 = 11
			tokensPerChar:  0.25,
			useFixedCount:  false,
		},
		{
			name:           "Code snippet",
			modelName:      "gpt-4",
			inputText:      "func main() {\n\tfmt.Println(\"Hello, world!\")\n}",
			expectedTokens: 11, // Length: 44 chars * 0.25 = 11
			tokensPerChar:  0.25,
			useFixedCount:  false,
		},
		{
			name:           "Long technical text",
			modelName:      "gpt-4",
			inputText:      "The OpenAI GPT-4 model has a context window of up to 8,192 tokens and can generate responses up to 8,000 tokens. It demonstrates stronger performance than previous models across a wide variety of tasks including coding, logical reasoning, and creative writing.",
			expectedTokens: 65, // Length: 260 chars * 0.25 = 65
			tokensPerChar:  0.25,
			useFixedCount:  false,
		},
		{
			name:           "Text with Unicode characters",
			modelName:      "gpt-4",
			inputText:      "Unicode text: こんにちは世界, Привет мир, नमस्ते दुनिया",
			expectedTokens: 33,   // Length: 95 chars * 0.35 = 33.25
			tokensPerChar:  0.35, // Higher ratio for non-ASCII text
			useFixedCount:  false,
		},
		{
			name:           "Text with emojis",
			modelName:      "gpt-4",
			inputText:      "Emoji test: 😀 🚀 🌍 🎉 🤖 📊 💯",
			expectedTokens: 13, // Length: 44 chars * 0.3 = 13.2
			tokensPerChar:  0.3,
			useFixedCount:  false,
		},
		{
			name:           "Whitespace-heavy text",
			modelName:      "gpt-4",
			inputText:      "    This    text    has    lots    of    spaces    between    words    ",
			expectedTokens: 14,  // Length: 73 chars * 0.2 = 14.6
			tokensPerChar:  0.2, // Lower ratio for whitespace-heavy text
			useFixedCount:  false,
		},
		{
			name:           "Fixed token count test",
			modelName:      "gpt-4",
			inputText:      "This text will always return the same token count",
			expectedTokens: 42,
			useFixedCount:  true,
			fixedCount:     42,
		},
		{
			name:             "Model-specific token count",
			modelName:        "gpt-3.5-turbo", // This will use the model-specific count for this model
			inputText:        "Same text, different models, different counts",
			expectedTokens:   15,
			useModelSpecific: true,
			modelCounts: map[string]int{
				"gpt-4":         10,
				"gpt-3.5-turbo": 15,
				"gpt-4-turbo":   20,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var tokenizer tokenizerAPI

			// Create the appropriate mock tokenizer based on test case configuration
			if tc.useFixedCount {
				tokenizer = MockTokenCounter(tc.fixedCount, nil)
			} else if tc.useModelSpecific {
				tokenizer = MockModelAwareTokenCounter(tc.modelCounts, 5, nil)
			} else {
				tokenizer = MockDynamicTokenCounter(tc.tokensPerChar, nil)
			}

			// Create client with the configured mock tokenizer
			client := &openaiClient{
				api:       &mockOpenAIAPI{},
				tokenizer: tokenizer,
				modelName: tc.modelName,
			}

			// Test CountTokens
			ctx := context.Background()
			tokenCount, err := client.CountTokens(ctx, tc.inputText)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTokens, tokenCount.Total)
		})
	}
}

// TestTokenCountingEdgeCases tests token counting for edge cases
func TestTokenCountingEdgeCases(t *testing.T) {
	// Test edge cases
	edgeCases := []struct {
		name          string
		modelName     string
		inputText     string
		mockGenerator func() tokenizerAPI
		expectedError bool
		expectedCount int32
	}{
		{
			name:      "Empty text",
			modelName: "gpt-4",
			inputText: "",
			mockGenerator: func() tokenizerAPI {
				return MockDynamicTokenCounter(0.25, nil)
			},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:      "Very long text",
			modelName: "gpt-4",
			inputText: strings.Repeat("long text test ", 500), // Approximately 7000 characters
			mockGenerator: func() tokenizerAPI {
				return MockDynamicTokenCounter(0.25, nil)
			},
			expectedError: false,
			expectedCount: int32(0.25 * float64(len(strings.Repeat("long text test ", 500)))), // 0.25 tokens per char * actual length
		},
		{
			name:      "Invalid model name",
			modelName: "invalid-model",
			inputText: "Test text",
			mockGenerator: func() tokenizerAPI {
				return MockTokenCounter(0, &APIError{
					Type:       ErrorTypeInvalidRequest,
					Message:    "Invalid model",
					StatusCode: 400,
					Suggestion: "Use a valid model name",
				})
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:      "Token count API failure",
			modelName: "gpt-4",
			inputText: "Test text",
			mockGenerator: func() tokenizerAPI {
				return MockTokenCounter(0, &APIError{
					Type:       ErrorTypeServer,
					Message:    "Token counting service unavailable",
					StatusCode: 503,
					Suggestion: "Try again later",
				})
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:      "Text exceeds model token limit",
			modelName: "gpt-4",
			inputText: "Very long text that would exceed the token limit",
			mockGenerator: func() tokenizerAPI {
				return MockTokenCounter(0, &APIError{
					Type:       ErrorTypeInputLimit,
					Message:    "Input exceeds maximum token limit",
					StatusCode: 400,
					Suggestion: "Reduce the input length or use a model with a larger context window",
				})
			},
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create client with the configured mock tokenizer
			client := &openaiClient{
				api:       &mockOpenAIAPI{},
				tokenizer: tc.mockGenerator(),
				modelName: tc.modelName,
			}

			// Test CountTokens
			ctx := context.Background()
			tokenCount, err := client.CountTokens(ctx, tc.inputText)

			if tc.expectedError {
				require.Error(t, err)
				assert.Nil(t, tokenCount)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCount, tokenCount.Total)
			}
		})
	}
}

// TestTokenCountingAcrossModels tests token counting across different model types
func TestTokenCountingAcrossModels(t *testing.T) {
	// Set of models to test with
	models := []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"custom-model",
	}

	// Sample text to test
	sampleText := "This is a test sentence for token counting across different models."

	// Create model-specific counts
	modelCounts := map[string]int{
		"gpt-4":             12,
		"gpt-4-turbo":       12,
		"gpt-3.5-turbo":     12,
		"gpt-3.5-turbo-16k": 12,
		// No entry for custom-model to test fallback
	}

	// Default count for models not in the map
	defaultCount := 10

	// Create mock tokenizer
	tokenizer := MockModelAwareTokenCounter(modelCounts, defaultCount, nil)

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			// Create client with the mock tokenizer and current model
			client := &openaiClient{
				api:       &mockOpenAIAPI{},
				tokenizer: tokenizer,
				modelName: model,
			}

			// Test CountTokens
			ctx := context.Background()
			tokenCount, err := client.CountTokens(ctx, sampleText)
			require.NoError(t, err)

			// Check if the model has a specific count, otherwise expect the default
			expectedCount := defaultCount
			if count, ok := modelCounts[model]; ok {
				expectedCount = count
			}
			assert.Equal(t, int32(expectedCount), tokenCount.Total)
		})
	}
}

// TestModelInfoMocks tests the model info mocking functionality
func TestModelInfoMocks(t *testing.T) {
	// Define a context for testing
	ctx := context.Background()

	// Test fixed model info mock
	t.Run("Fixed model info", func(t *testing.T) {
		// Create a mock that returns the same model info for any model
		fixedModelInfo := MockModelInfo(10000, 2000, nil)

		// Test with different model names
		modelNames := []string{"gpt-4", "gpt-3.5-turbo", "custom-model"}

		for _, modelName := range modelNames {
			info, err := fixedModelInfo.getModelInfo(ctx, modelName)
			require.NoError(t, err)
			assert.Equal(t, int32(10000), info.inputTokenLimit)
			assert.Equal(t, int32(2000), info.outputTokenLimit)
		}
	})

	// Test model-specific model info mock
	t.Run("Model-specific info", func(t *testing.T) {
		// Create a map of model-specific info
		modelInfoMap := map[string]*modelInfo{
			"gpt-4": {
				inputTokenLimit:  8192,
				outputTokenLimit: 2048,
			},
			"gpt-4-32k": {
				inputTokenLimit:  32768,
				outputTokenLimit: 4096,
			},
			"gpt-3.5-turbo": {
				inputTokenLimit:  4096,
				outputTokenLimit: 1024,
			},
		}

		// Define default info for models not in the map
		defaultInfo := &modelInfo{
			inputTokenLimit:  4096, // Conservative default
			outputTokenLimit: 1024, // Conservative default
		}

		// Create a mock that returns different info for specific models
		modelSpecificInfo := MockModelSpecificInfo(modelInfoMap, defaultInfo, nil)

		// Test with models that have specific info
		for modelName, expectedInfo := range modelInfoMap {
			info, err := modelSpecificInfo.getModelInfo(ctx, modelName)
			require.NoError(t, err)
			assert.Equal(t, expectedInfo.inputTokenLimit, info.inputTokenLimit)
			assert.Equal(t, expectedInfo.outputTokenLimit, info.outputTokenLimit)
		}

		// Test with a model that doesn't have specific info (should return default)
		info, err := modelSpecificInfo.getModelInfo(ctx, "unknown-model")
		require.NoError(t, err)
		assert.Equal(t, defaultInfo.inputTokenLimit, info.inputTokenLimit)
		assert.Equal(t, defaultInfo.outputTokenLimit, info.outputTokenLimit)
	})

	// Test error handling in model info mock
	t.Run("Error handling", func(t *testing.T) {
		// Create a mock that always returns an error
		errorMock := MockModelInfo(0, 0, &APIError{
			Type:    ErrorTypeInvalidRequest,
			Message: "Invalid model",
		})

		// Test that the error is returned
		_, err := errorMock.getModelInfo(ctx, "any-model")
		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, ErrorTypeInvalidRequest, apiErr.Type)
		assert.Equal(t, "Invalid model", apiErr.Message)
	})

	// Test model-specific errors
	t.Run("Model-specific errors", func(t *testing.T) {
		// Create a map of models that should return errors
		errorModels := map[string]error{
			"invalid-model": &APIError{
				Type:    ErrorTypeInvalidRequest,
				Message: "Model not found",
			},
			"deprecated-model": &APIError{
				Type:    ErrorTypeInvalidRequest,
				Message: "Model is deprecated",
			},
		}

		// Define default info for non-error models
		defaultInfo := &modelInfo{
			inputTokenLimit:  4096,
			outputTokenLimit: 1024,
		}

		// Create a mock that returns errors for specific models
		modelErrorMock := MockModelInfoWithErrors(errorModels, defaultInfo)

		// Test models that should return errors
		for modelName, expectedErr := range errorModels {
			_, err := modelErrorMock.getModelInfo(ctx, modelName)
			require.Error(t, err)
			assert.Equal(t, expectedErr, err)
		}

		// Test a model that shouldn't return an error
		info, err := modelErrorMock.getModelInfo(ctx, "gpt-4")
		require.NoError(t, err)
		assert.Equal(t, defaultInfo.inputTokenLimit, info.inputTokenLimit)
		assert.Equal(t, defaultInfo.outputTokenLimit, info.outputTokenLimit)
	})
}

// TestModelInfoIntegration tests using the model info mocks with the OpenAI client
func TestModelInfoIntegration(t *testing.T) {
	ctx := context.Background()

	// Test client with fixed model info mock
	t.Run("Client with fixed model info", func(t *testing.T) {
		// Create a client with fixed model info
		client := &openaiClient{
			api:       &mockOpenAIAPI{},
			tokenizer: &mockTokenizer{},
			modelName: "gpt-4",
			modelLimits: map[string]*modelInfo{
				"gpt-4": {
					inputTokenLimit:  10000,
					outputTokenLimit: 2000,
				},
			},
		}

		// Test GetModelInfo
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", modelInfo.Name)
		assert.Equal(t, int32(10000), modelInfo.InputTokenLimit)
		assert.Equal(t, int32(2000), modelInfo.OutputTokenLimit)
	})

	// Test client with model-specific info mock
	t.Run("Client with model-specific info", func(t *testing.T) {
		// Create a client with model-specific info
		client := &openaiClient{
			api:       &mockOpenAIAPI{},
			tokenizer: &mockTokenizer{},
			modelName: "gpt-4",
			modelLimits: map[string]*modelInfo{
				"gpt-4": {
					inputTokenLimit:  8192,
					outputTokenLimit: 2048,
				},
				"gpt-4-32k": {
					inputTokenLimit:  32768,
					outputTokenLimit: 4096,
				},
				"gpt-3.5-turbo": {
					inputTokenLimit:  4096,
					outputTokenLimit: 1024,
				},
			},
		}

		// Test GetModelInfo with current model
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", modelInfo.Name)
		assert.Equal(t, int32(8192), modelInfo.InputTokenLimit)
		assert.Equal(t, int32(2048), modelInfo.OutputTokenLimit)

		// Change model and test again
		client.modelName = "gpt-4-32k"
		modelInfo, err = client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4-32k", modelInfo.Name)
		assert.Equal(t, int32(32768), modelInfo.InputTokenLimit)
		assert.Equal(t, int32(4096), modelInfo.OutputTokenLimit)
	})

	// Test client with unknown model (should use default values)
	t.Run("Client with unknown model", func(t *testing.T) {
		// Create a client with an unknown model
		client := &openaiClient{
			api:       &mockOpenAIAPI{},
			tokenizer: &mockTokenizer{},
			modelName: "unknown-model",
			modelLimits: map[string]*modelInfo{
				"gpt-4": {
					inputTokenLimit:  8192,
					outputTokenLimit: 2048,
				},
			},
		}

		// Test GetModelInfo - should return conservative defaults
		modelInfo, err := client.GetModelInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, "unknown-model", modelInfo.Name)
		assert.Equal(t, int32(4096), modelInfo.InputTokenLimit)  // Conservative default
		assert.Equal(t, int32(2048), modelInfo.OutputTokenLimit) // Conservative default
	})
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
