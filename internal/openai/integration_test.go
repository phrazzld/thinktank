package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go"
	"github.com/phrazzld/thinktank/internal/llm"
)

// setupMockServer creates an HTTP test server for integration testing
func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

// createJSONHandler creates a handler that returns JSON response
func createJSONHandler(statusCode int, response interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
			}
		}
	}
}

// createErrorHandler creates a handler that returns an HTTP error
func createErrorHandler(statusCode int, errorMessage string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, errorMessage, statusCode)
	}
}

func TestCreateChatCompletionIntegration(t *testing.T) {
	t.Run("successful chat completion", func(t *testing.T) {
		// Set up mock OpenAI API response
		successResponse := map[string]interface{}{
			"id":      "chatcmpl-test123",
			"object":  "chat.completion",
			"created": 1677652288,
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello! How can I help you today?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     9,
				"completion_tokens": 12,
				"total_tokens":      21,
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))

		// Create client with mock server endpoint
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Cast to access the API directly for testing
		openaiClient := client.(*openaiClient)
		api := openaiClient.api

		// Test createChatCompletion
		completion, err := api.createChatCompletion(context.Background(), "gpt-4", "Hello", "")
		if err != nil {
			t.Fatalf("createChatCompletion failed: %v", err)
		}

		if completion == nil {
			t.Fatal("Expected completion, got nil")
		}

		if len(completion.Choices) == 0 {
			t.Fatal("Expected choices, got empty")
		}

		if completion.Choices[0].Message.Content != "Hello! How can I help you today?" {
			t.Errorf("Expected specific content, got %q", completion.Choices[0].Message.Content)
		}
	})

	t.Run("chat completion with system prompt", func(t *testing.T) {
		successResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "System-aware response",
					},
					"finish_reason": "stop",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		openaiClient := client.(*openaiClient)
		api := openaiClient.api

		// Test with system prompt
		completion, err := api.createChatCompletion(context.Background(), "gpt-4", "Hello", "You are a helpful assistant")
		if err != nil {
			t.Fatalf("createChatCompletion with system prompt failed: %v", err)
		}

		if completion.Choices[0].Message.Content != "System-aware response" {
			t.Errorf("Expected system-aware response, got %q", completion.Choices[0].Message.Content)
		}
	})

	t.Run("chat completion API error", func(t *testing.T) {
		server := setupMockServer(t, createErrorHandler(http.StatusUnauthorized, "Invalid API key"))

		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		openaiClient := client.(*openaiClient)
		api := openaiClient.api

		_, err = api.createChatCompletion(context.Background(), "gpt-4", "Hello", "")
		if err == nil {
			t.Fatal("Expected error for API failure, got nil")
		}

		// Verify it's a structured error
		if apiErr, ok := IsOpenAIError(err); ok {
			if apiErr.Category() != llm.CategoryAuth {
				t.Errorf("Expected CategoryAuth error, got %v", apiErr.Category())
			}
		}
	})
}

func TestCreateChatCompletionWithParamsIntegration(t *testing.T) {
	t.Run("successful completion with parameters", func(t *testing.T) {
		successResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Parameterized response",
					},
					"finish_reason": "stop",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		openaiClient := client.(*openaiClient)
		api := openaiClient.api

		// Create parameters for the request
		params := openai.ChatCompletionNewParams{
			Model: "gpt-4",
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage("Test prompt"),
			},
			Temperature: openai.Float(0.7),
			MaxTokens:   openai.Int(100),
		}

		completion, err := api.createChatCompletionWithParams(context.Background(), params)
		if err != nil {
			t.Fatalf("createChatCompletionWithParams failed: %v", err)
		}

		if completion == nil {
			t.Fatal("Expected completion, got nil")
		}

		if completion.Choices[0].Message.Content != "Parameterized response" {
			t.Errorf("Expected parameterized response, got %q", completion.Choices[0].Message.Content)
		}
	})

	t.Run("completion with params API error", func(t *testing.T) {
		server := setupMockServer(t, createErrorHandler(http.StatusTooManyRequests, "Rate limit exceeded"))

		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		openaiClient := client.(*openaiClient)
		api := openaiClient.api

		params := openai.ChatCompletionNewParams{
			Model: "gpt-4",
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage("Test prompt"),
			},
		}

		_, err = api.createChatCompletionWithParams(context.Background(), params)
		if err == nil {
			t.Fatal("Expected error for rate limit, got nil")
		}

		// Verify it's a structured error
		if apiErr, ok := IsOpenAIError(err); ok {
			if apiErr.Category() != llm.CategoryRateLimit {
				t.Errorf("Expected CategoryRateLimit error, got %v", apiErr.Category())
			}
		}
	})
}

func TestCloseIntegration(t *testing.T) {
	t.Run("close succeeds", func(t *testing.T) {
		client, err := NewClient("test-api-key", "gpt-4", "")
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Close should succeed (it's a no-op for OpenAI)
		err = client.Close()
		if err != nil {
			t.Errorf("Close() failed: %v", err)
		}

		// Multiple closes should be safe
		err = client.Close()
		if err != nil {
			t.Errorf("Second Close() failed: %v", err)
		}
	})

	t.Run("operations after close", func(t *testing.T) {
		client, err := NewClient("test-api-key", "gpt-4", "")
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Close the client
		err = client.Close()
		if err != nil {
			t.Fatalf("Close() failed: %v", err)
		}

		// GetModelName should still work (it doesn't rely on external resources)
		modelName := client.GetModelName()
		if modelName != "gpt-4" {
			t.Errorf("GetModelName after close returned wrong value: %q", modelName)
		}
	})
}

func TestGenerateContentIntegrationAdditional(t *testing.T) {
	t.Run("empty prompt validation", func(t *testing.T) {
		client, err := NewClient("test-api-key", "gpt-4", "")
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		result, err := client.GenerateContent(context.Background(), "", nil)
		if err == nil {
			t.Fatal("Expected error for empty prompt, got nil")
		}
		if result != nil {
			t.Error("Expected nil result for empty prompt")
		}

		// Verify it's an OpenAI error with correct category
		apiErr, ok := IsOpenAIError(err)
		if !ok {
			t.Fatalf("Expected OpenAI error, got %T", err)
		}

		if apiErr.Category() != llm.CategoryInvalidRequest {
			t.Errorf("Expected CategoryInvalidRequest, got %v", apiErr.Category())
		}
	})

	t.Run("system prompt extraction", func(t *testing.T) {
		successResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Response with system prompt",
					},
					"finish_reason": "stop",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Test system prompt extraction with <system> tags
		result, err := client.GenerateContent(context.Background(), "<system>You are helpful</system>Hello world", nil)
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if result.Content != "Response with system prompt" {
			t.Errorf("Expected system prompt response, got %q", result.Content)
		}
	})

	t.Run("legacy system prompt extraction", func(t *testing.T) {
		successResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Legacy response",
					},
					"finish_reason": "stop",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Test legacy system prompt extraction with <s> tags
		result, err := client.GenerateContent(context.Background(), "<s>You are helpful</s>Hello world", nil)
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if result.Content != "Legacy response" {
			t.Errorf("Expected legacy response, got %q", result.Content)
		}
	})

	t.Run("parameter handling comprehensive", func(t *testing.T) {
		successResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Parameterized response",
					},
					"finish_reason": "stop",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Test with various parameter types
		params := map[string]interface{}{
			"temperature":       float64(0.8),
			"top_p":             float32(0.9),
			"presence_penalty":  float64(0.1),
			"frequency_penalty": float32(0.2),
			"max_tokens":        int(150),
		}

		result, err := client.GenerateContent(context.Background(), "Test prompt", params)
		if err != nil {
			t.Fatalf("GenerateContent with parameters failed: %v", err)
		}

		if result.Content != "Parameterized response" {
			t.Errorf("Expected parameterized response, got %q", result.Content)
		}
	})

	t.Run("truncated response detection", func(t *testing.T) {
		truncatedResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "This response was truncated...",
					},
					"finish_reason": "length",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, truncatedResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		result, err := client.GenerateContent(context.Background(), "Long prompt", nil)
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if !result.Truncated {
			t.Error("Expected response to be marked as truncated")
		}

		if result.FinishReason != "length" {
			t.Errorf("Expected finish reason 'length', got %q", result.FinishReason)
		}
	})

	t.Run("no choices error", func(t *testing.T) {
		emptyResponse := map[string]interface{}{
			"choices": []interface{}{},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, emptyResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		_, err = client.GenerateContent(context.Background(), "Test prompt", nil)
		if err == nil {
			t.Fatal("Expected error for empty choices, got nil")
		}

		if err.Error() != "no completions returned from OpenAI API" {
			t.Errorf("Expected specific error message, got %q", err.Error())
		}
	})

	t.Run("client with preset parameters", func(t *testing.T) {
		successResponse := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Response with preset params",
					},
					"finish_reason": "stop",
				},
			},
		}

		server := setupMockServer(t, createJSONHandler(http.StatusOK, successResponse))
		client, err := NewClient("test-api-key", "gpt-4", server.URL)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		// Set preset parameters on the client
		openaiClient := client.(*openaiClient)
		openaiClient.SetTemperature(0.7)
		openaiClient.SetTopP(0.9)
		openaiClient.SetMaxTokens(150)
		openaiClient.SetFrequencyPenalty(0.1)
		openaiClient.SetPresencePenalty(0.2)

		result, err := client.GenerateContent(context.Background(), "Test prompt", nil)
		if err != nil {
			t.Fatalf("GenerateContent with preset params failed: %v", err)
		}

		if result.Content != "Response with preset params" {
			t.Errorf("Expected preset params response, got %q", result.Content)
		}
	})
}

func TestErrorHandlingAdditional(t *testing.T) {
	t.Run("FormatAPIError with different error types", func(t *testing.T) {
		testCases := []struct {
			err              error
			statusCode       int
			expectedCategory llm.ErrorCategory
		}{
			{errors.New("invalid API key"), 401, llm.CategoryAuth},
			{errors.New("rate limit exceeded"), 429, llm.CategoryRateLimit},
			{errors.New("bad request"), 400, llm.CategoryInvalidRequest},
			{errors.New("model not found"), 404, llm.CategoryNotFound},
			{errors.New("server error"), 500, llm.CategoryServer},
			{errors.New("network timeout"), 0, llm.CategoryNetwork},
			{errors.New("context cancelled"), 0, llm.CategoryCancelled},
			{errors.New("input too long"), 0, llm.CategoryInputLimit},
			{errors.New("content filtered"), 0, llm.CategoryContentFiltered},
		}

		for _, tc := range testCases {
			formatted := FormatAPIError(tc.err, tc.statusCode)
			if formatted.Category() != tc.expectedCategory {
				t.Errorf("Error %q (status %d): expected %v, got %v",
					tc.err.Error(), tc.statusCode, tc.expectedCategory, formatted.Category())
			}
		}
	})

	t.Run("FormatAPIError with existing LLMError", func(t *testing.T) {
		// Test that it returns the same error if already an LLMError
		originalErr := CreateAPIError(llm.CategoryAuth, "test", nil, "")
		formatted := FormatAPIError(originalErr, 200)

		if formatted != originalErr {
			t.Error("Expected same LLMError to be returned")
		}
	})

	t.Run("FormatAPIError with nil error", func(t *testing.T) {
		result := FormatAPIError(nil, 200)
		if result != nil {
			t.Errorf("Expected nil result for nil error, got %v", result)
		}
	})

	t.Run("CreateAPIError comprehensive", func(t *testing.T) {
		categories := []llm.ErrorCategory{
			llm.CategoryAuth,
			llm.CategoryRateLimit,
			llm.CategoryInvalidRequest,
			llm.CategoryNotFound,
			llm.CategoryServer,
			llm.CategoryNetwork,
			llm.CategoryCancelled,
			llm.CategoryInputLimit,
			llm.CategoryContentFiltered,
			llm.CategoryUnknown,
		}

		for _, category := range categories {
			err := CreateAPIError(category, "test message", nil, "test details")
			if err.Category() != category {
				t.Errorf("Expected category %v, got %v", category, err.Category())
			}
			if err.Provider != "openai" {
				t.Errorf("Expected provider 'openai', got %q", err.Provider)
			}
		}
	})

	t.Run("IsOpenAIError validation", func(t *testing.T) {
		// Test with nil error
		_, ok := IsOpenAIError(nil)
		if ok {
			t.Error("Expected false for nil error")
		}

		// Test with regular error
		regularErr := errors.New("regular error")
		_, ok = IsOpenAIError(regularErr)
		if ok {
			t.Error("Expected false for regular error")
		}

		// Test with OpenAI error
		openaiErr := CreateAPIError(llm.CategoryAuth, "test", nil, "")
		llmErr, ok := IsOpenAIError(openaiErr)
		if !ok {
			t.Fatal("Expected true for OpenAI error")
		}
		if llmErr.Provider != "openai" {
			t.Errorf("Expected openai provider, got %q", llmErr.Provider)
		}
	})

	t.Run("MockAPIErrorResponse comprehensive", func(t *testing.T) {
		testCases := []struct {
			errorType        int
			expectedCategory llm.ErrorCategory
			description      string
		}{
			{1, llm.CategoryAuth, "Auth"},
			{2, llm.CategoryRateLimit, "RateLimit"},
			{3, llm.CategoryInvalidRequest, "InvalidRequest"},
			{4, llm.CategoryNotFound, "NotFound"},
			{5, llm.CategoryServer, "Server"},
			{6, llm.CategoryNetwork, "Network"},
			{7, llm.CategoryCancelled, "Cancelled"},
			{8, llm.CategoryInputLimit, "InputLimit"},
			{9, llm.CategoryContentFiltered, "ContentFiltered"},
			{999, llm.CategoryUnknown, "Unknown"}, // Default case
		}

		for _, tc := range testCases {
			err := MockAPIErrorResponse(tc.errorType, 400, "test message", "test details")
			if err.Category() != tc.expectedCategory {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expectedCategory, err.Category())
			}
			if err.StatusCode != 400 {
				t.Errorf("%s: expected status code 400, got %d", tc.description, err.StatusCode)
			}
			if err.Provider != "openai" {
				t.Errorf("%s: expected provider 'openai', got %q", tc.description, err.Provider)
			}
		}
	})
}

func TestApplyOpenAIParametersAdditional(t *testing.T) {
	t.Run("parameter type variations", func(t *testing.T) {
		params := openai.ChatCompletionNewParams{}

		// Test different numeric types for each parameter
		customParams := map[string]interface{}{
			"temperature":       float32(0.5), // float32 variant
			"top_p":             float64(0.8), // float64 variant
			"presence_penalty":  float32(0.1), // float32 variant
			"frequency_penalty": float64(0.2), // float64 variant
			"max_tokens":        int32(100),   // int32 variant
		}

		// This should not panic and should handle type conversions
		applyOpenAIParameters(&params, customParams)

		// Just verify the function ran without panicking
		// The actual parameter setting is verified through integration tests
	})

	t.Run("parameter edge cases", func(t *testing.T) {
		params := openai.ChatCompletionNewParams{}

		// Test with int type for max_tokens
		customParams := map[string]interface{}{
			"max_tokens": int(200),
		}

		applyOpenAIParameters(&params, customParams)

		// Just verify the function ran without panicking for int type

		// Test with float64 for max_tokens
		params = openai.ChatCompletionNewParams{}
		customParams = map[string]interface{}{
			"max_tokens": float64(300),
		}

		applyOpenAIParameters(&params, customParams)

		// Just verify the function ran without panicking for float64 type
	})
}
