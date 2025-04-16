// Package openai provides shared test utilities for OpenAI client tests
package openai

import (
	"context"

	"github.com/openai/openai-go"
)

// mockOpenAIAPI is a test double implementing openaiAPI for capturing or simulating API calls
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
	return m.createChatCompletionFunc(ctx, params.Messages, params.Model)
}

// mockAPIWithError creates a mockOpenAIAPI that returns the specified error for all API calls
func mockAPIWithError(err error) *mockOpenAIAPI {
	return &mockOpenAIAPI{
		createChatCompletionFunc: func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
			return nil, err
		},
		createChatCompletionWithParamsFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return nil, err
		},
	}
}

// mockTokenizer is a test double implementing tokenizerAPI for controlling token counts
type mockTokenizer struct {
	countTokensFunc func(text string, model string) (int, error)
}

func (m *mockTokenizer) countTokens(text string, model string) (int, error) {
	return m.countTokensFunc(text, model)
}

// mockModelInfoProvider implements model info retrieval for testing
type mockModelInfoProvider struct {
	getModelInfoFunc func(ctx context.Context, modelName string) (*modelInfo, error)
}

// getModelInfo retrieves model information using the provided mock function
func (m *mockModelInfoProvider) getModelInfo(ctx context.Context, modelName string) (*modelInfo, error) {
	return m.getModelInfoFunc(ctx, modelName)
}

// Helper function to convert to a pointer
func toPtr[T any](v T) *T {
	return &v
}

// MockAPIErrorResponse creates a mock error response with specifics about the error
func MockAPIErrorResponse(errorType ErrorType, statusCode int, message string, details string) *APIError {
	suggestion := ""

	// Set appropriate suggestion based on error type
	switch errorType {
	case ErrorTypeAuth:
		suggestion = "Check that your API key is valid and has not expired. Ensure environment variables are set correctly."
	case ErrorTypeRateLimit:
		suggestion = "Wait and try again later. Consider adjusting the --max-concurrent and --rate-limit flags to limit request rate."
	case ErrorTypeInvalidRequest:
		suggestion = "Check the prompt format and parameters. Ensure they comply with the API requirements."
	case ErrorTypeNotFound:
		suggestion = "Verify that the model name is correct and that the model is available in your region."
	case ErrorTypeServer:
		suggestion = "This is typically a temporary issue. Wait a few moments and try again."
	case ErrorTypeNetwork:
		suggestion = "Check your internet connection and try again."
	case ErrorTypeInputLimit:
		suggestion = "Reduce the input size or use a model with higher context limits."
	case ErrorTypeContentFiltered:
		suggestion = "Your prompt or content may have triggered safety filters. Review and modify your input to comply with content policies."
	}

	return &APIError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
		Suggestion: suggestion,
		Details:    details,
	}
}

// MockModelInfo creates a mock model info provider that returns a fixed model info for any model
func MockModelInfo(inputTokenLimit, outputTokenLimit int32, errorToReturn error) *mockModelInfoProvider {
	return &mockModelInfoProvider{
		getModelInfoFunc: func(ctx context.Context, modelName string) (*modelInfo, error) {
			if errorToReturn != nil {
				return nil, errorToReturn
			}
			return &modelInfo{
				inputTokenLimit:  inputTokenLimit,
				outputTokenLimit: outputTokenLimit,
			}, nil
		},
	}
}

// MockModelSpecificInfo creates a mock model info provider that returns different info for specific models
func MockModelSpecificInfo(modelInfoMap map[string]*modelInfo, defaultInfo *modelInfo, errorToReturn error) *mockModelInfoProvider {
	return &mockModelInfoProvider{
		getModelInfoFunc: func(ctx context.Context, modelName string) (*modelInfo, error) {
			if errorToReturn != nil {
				return nil, errorToReturn
			}

			// Check if we have specific info for this model
			if info, ok := modelInfoMap[modelName]; ok {
				return info, nil
			}

			// Return default info if no specific info is found
			return defaultInfo, nil
		},
	}
}

// MockModelInfoWithErrors creates a mock model info provider that returns errors for specific models
func MockModelInfoWithErrors(errorModels map[string]error, defaultInfo *modelInfo) *mockModelInfoProvider {
	return &mockModelInfoProvider{
		getModelInfoFunc: func(ctx context.Context, modelName string) (*modelInfo, error) {
			// Check if this model should return an error
			if err, ok := errorModels[modelName]; ok {
				return nil, err
			}

			// Return default info for non-error models
			return defaultInfo, nil
		},
	}
}

// MockTokenCounter creates a mock token counter with predictable token counts
func MockTokenCounter(fixedTokenCount int, errorToReturn error) *mockTokenizer {
	return &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			if errorToReturn != nil {
				return 0, errorToReturn
			}
			return fixedTokenCount, nil
		},
	}
}

// MockDynamicTokenCounter creates a mock token counter that returns token counts based on text length
func MockDynamicTokenCounter(tokensPerChar float64, errorToReturn error) *mockTokenizer {
	return &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			if errorToReturn != nil {
				return 0, errorToReturn
			}
			// Calculate tokens based on text length - simple approximation
			return int(float64(len(text)) * tokensPerChar), nil
		},
	}
}

// MockModelAwareTokenCounter creates a mock token counter that returns different counts based on the model
func MockModelAwareTokenCounter(modelTokenCounts map[string]int, defaultCount int, errorToReturn error) *mockTokenizer {
	return &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			if errorToReturn != nil {
				return 0, errorToReturn
			}

			// If we have a specific count for this model, return it
			if count, ok := modelTokenCounts[model]; ok {
				return count, nil
			}

			// Otherwise, return the default count
			return defaultCount, nil
		},
	}
}

// MockPredictableTokenCounter creates a token counter that returns predictable results for specific inputs
func MockPredictableTokenCounter(textToTokenMap map[string]int, defaultCount int, errorToReturn error) *mockTokenizer {
	return &mockTokenizer{
		countTokensFunc: func(text string, model string) (int, error) {
			if errorToReturn != nil {
				return 0, errorToReturn
			}

			// If we have a specific count for this text, return it
			if count, ok := textToTokenMap[text]; ok {
				return count, nil
			}

			// Otherwise, return the default count
			return defaultCount, nil
		},
	}
}
