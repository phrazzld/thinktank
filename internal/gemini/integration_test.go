package gemini

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"

	genai "github.com/google/generative-ai-go/genai"
)

func TestNewLLMClientIntegration(t *testing.T) {
	t.Run("client creation with invalid API key", func(t *testing.T) {
		_, err := NewLLMClient(context.Background(), "", "gemini-pro", "")

		if err == nil {
			t.Fatal("Expected error for empty API key, got nil")
		}

		expectedMsg := "API key cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("client creation with invalid model name", func(t *testing.T) {
		_, err := NewLLMClient(context.Background(), "test-api-key", "", "")

		if err == nil {
			t.Fatal("Expected error for empty model name, got nil")
		}

		expectedMsg := "model name cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestGetModelNameIntegration(t *testing.T) {
	t.Run("returns correct model name", func(t *testing.T) {
		testCases := []struct {
			name      string
			modelName string
		}{
			{"gemini-pro", "gemini-pro"},
			{"gemini-1.5-pro", "gemini-1.5-pro"},
			{"custom-model", "custom-model"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create client directly to test GetModelName without SDK issues
				client := &geminiClient{
					modelName: tc.modelName,
					logger:    getTestLogger(),
				}

				modelName := client.GetModelName()
				if modelName != tc.modelName {
					t.Errorf("Expected model name %q, got %q", tc.modelName, modelName)
				}
			})
		}
	})
}

func TestCloseIntegration(t *testing.T) {
	t.Run("close succeeds on nil client", func(t *testing.T) {
		// Create client directly to test Close behavior
		client := &geminiClient{
			client:    nil, // Simulate closed or uninitialized client
			modelName: "gemini-pro",
			logger:    getTestLogger(),
		}

		// Close should succeed even with nil client
		err := client.Close()
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
		client := &geminiClient{
			client:    nil, // Simulate closed client
			modelName: "gemini-pro",
			logger:    getTestLogger(),
		}

		// Close the client
		err := client.Close()
		if err != nil {
			t.Fatalf("Close() failed: %v", err)
		}

		// GetModelName should still work (it doesn't use the underlying client)
		modelName := client.GetModelName()
		if modelName != "gemini-pro" {
			t.Errorf("GetModelName after close returned wrong value: %q", modelName)
		}
	})
}

func TestGenerateContentIntegrationDirect(t *testing.T) {
	t.Run("empty prompt validation", func(t *testing.T) {
		// Test the entry point validation directly
		client := &geminiClient{
			modelName: "gemini-pro",
			logger:    getTestLogger(),
		}

		result, err := client.GenerateContent(context.Background(), "", nil)

		if err == nil {
			t.Fatal("Expected error for empty prompt, got nil")
		}
		if result != nil {
			t.Error("Expected nil result for empty prompt")
		}

		// Verify it's a Gemini error with correct category
		apiErr, ok := IsGeminiError(err)
		if !ok {
			t.Fatalf("Expected Gemini error, got %T", err)
		}

		if apiErr.Category() != llm.CategoryInvalidRequest {
			t.Errorf("Expected CategoryInvalidRequest, got %v", apiErr.Category())
		}
	})
}

func TestMapSafetyRatingsUnitTest(t *testing.T) {
	t.Run("handles nil input", func(t *testing.T) {
		result := mapSafetyRatings(nil)
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %v", result)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := mapSafetyRatings([]*genai.SafetyRating{})
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("maps single safety rating", func(t *testing.T) {
		input := []*genai.SafetyRating{
			{
				Category:    genai.HarmCategoryHarassment,
				Probability: genai.HarmProbabilityLow,
				Blocked:     false,
			},
		}

		result := mapSafetyRatings(input)

		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		expected := SafetyRating{
			Category: string(genai.HarmCategoryHarassment),
			Blocked:  false,
			Score:    float32(genai.HarmProbabilityLow),
		}

		if result[0].Category != expected.Category {
			t.Errorf("Expected category %q, got %q", expected.Category, result[0].Category)
		}
		if result[0].Blocked != expected.Blocked {
			t.Errorf("Expected blocked %v, got %v", expected.Blocked, result[0].Blocked)
		}
		if result[0].Score != expected.Score {
			t.Errorf("Expected score %v, got %v", expected.Score, result[0].Score)
		}
	})

	t.Run("maps multiple safety ratings", func(t *testing.T) {
		input := []*genai.SafetyRating{
			{
				Category:    genai.HarmCategoryHarassment,
				Probability: genai.HarmProbabilityLow,
				Blocked:     false,
			},
			{
				Category:    genai.HarmCategoryDangerousContent,
				Probability: genai.HarmProbabilityHigh,
				Blocked:     true,
			},
		}

		result := mapSafetyRatings(input)

		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// Verify first rating
		if result[0].Category != string(genai.HarmCategoryHarassment) {
			t.Errorf("Expected first category %q, got %q", string(genai.HarmCategoryHarassment), result[0].Category)
		}
		if result[0].Blocked != false {
			t.Errorf("Expected first blocked %v, got %v", false, result[0].Blocked)
		}

		// Verify second rating
		if result[1].Category != string(genai.HarmCategoryDangerousContent) {
			t.Errorf("Expected second category %q, got %q", string(genai.HarmCategoryDangerousContent), result[1].Category)
		}
		if result[1].Blocked != true {
			t.Errorf("Expected second blocked %v, got %v", true, result[1].Blocked)
		}
	})
}

func TestToProviderSafetyUnitTest(t *testing.T) {
	t.Run("handles nil input", func(t *testing.T) {
		result := toProviderSafety(nil)
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %v", result)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := toProviderSafety([]SafetyRating{})
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("converts single safety rating", func(t *testing.T) {
		input := []SafetyRating{
			{
				Category: "HARM_CATEGORY_HARASSMENT",
				Blocked:  true,
				Score:    0.8,
			},
		}

		result := toProviderSafety(input)

		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		expected := llm.Safety{
			Category: "HARM_CATEGORY_HARASSMENT",
			Blocked:  true,
			Score:    0.8,
		}

		if result[0].Category != expected.Category {
			t.Errorf("Expected category %q, got %q", expected.Category, result[0].Category)
		}
		if result[0].Blocked != expected.Blocked {
			t.Errorf("Expected blocked %v, got %v", expected.Blocked, result[0].Blocked)
		}
		if result[0].Score != expected.Score {
			t.Errorf("Expected score %v, got %v", expected.Score, result[0].Score)
		}
	})

	t.Run("converts multiple safety ratings", func(t *testing.T) {
		input := []SafetyRating{
			{
				Category: "HARM_CATEGORY_HARASSMENT",
				Blocked:  false,
				Score:    0.1,
			},
			{
				Category: "HARM_CATEGORY_DANGEROUS_CONTENT",
				Blocked:  true,
				Score:    0.9,
			},
		}

		result := toProviderSafety(input)

		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// Verify first safety info
		if result[0].Category != "HARM_CATEGORY_HARASSMENT" {
			t.Errorf("Expected first category %q, got %q", "HARM_CATEGORY_HARASSMENT", result[0].Category)
		}
		if result[0].Blocked != false {
			t.Errorf("Expected first blocked %v, got %v", false, result[0].Blocked)
		}
		if result[0].Score != 0.1 {
			t.Errorf("Expected first score %v, got %v", 0.1, result[0].Score)
		}

		// Verify second safety info
		if result[1].Category != "HARM_CATEGORY_DANGEROUS_CONTENT" {
			t.Errorf("Expected second category %q, got %q", "HARM_CATEGORY_DANGEROUS_CONTENT", result[1].Category)
		}
		if result[1].Blocked != true {
			t.Errorf("Expected second blocked %v, got %v", true, result[1].Blocked)
		}
		if result[1].Score != 0.9 {
			t.Errorf("Expected second score %v, got %v", 0.9, result[1].Score)
		}
	})

	t.Run("preserves all data fields", func(t *testing.T) {
		// Test that the conversion preserves all the important data
		input := []SafetyRating{
			{
				Category: "CUSTOM_CATEGORY",
				Blocked:  true,
				Score:    0.456,
			},
		}

		result := toProviderSafety(input)

		if len(result) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(result))
		}

		// Ensure exact preservation of data
		if result[0].Category != input[0].Category {
			t.Errorf("Category not preserved: expected %q, got %q", input[0].Category, result[0].Category)
		}
		if result[0].Blocked != input[0].Blocked {
			t.Errorf("Blocked not preserved: expected %v, got %v", input[0].Blocked, result[0].Blocked)
		}
		if result[0].Score != input[0].Score {
			t.Errorf("Score not preserved: expected %v, got %v", input[0].Score, result[0].Score)
		}
	})
}

func TestClientParameterHandling(t *testing.T) {
	t.Run("handles empty parameters", func(t *testing.T) {
		// Test that the client can handle nil parameters without panic
		client := &geminiClient{
			modelName: "gemini-pro",
			logger:    getTestLogger(),
		}

		// This will fail at the genini client level, but we're testing that nil params don't panic
		result, err := client.GenerateContent(context.Background(), "Test prompt", nil)

		// We expect an error since we don't have a real genai client
		if err == nil {
			t.Fatal("Expected error due to nil client, got nil")
		}
		if result != nil {
			t.Error("Expected nil result due to error")
		}
	})
}

// Additional tests to increase coverage of entry points
func TestNewLLMClientValidation(t *testing.T) {
	t.Run("empty API key validation", func(t *testing.T) {
		client, err := NewLLMClient(context.Background(), "", "model", "")
		if err == nil {
			t.Fatal("Expected error for empty API key")
		}
		if client != nil {
			t.Error("Expected nil client for invalid input")
		}
		if err.Error() != "API key cannot be empty" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("empty model name validation", func(t *testing.T) {
		client, err := NewLLMClient(context.Background(), "key", "", "")
		if err == nil {
			t.Fatal("Expected error for empty model")
		}
		if client != nil {
			t.Error("Expected nil client for invalid input")
		}
		if err.Error() != "model name cannot be empty" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestClientMethodCoverage(t *testing.T) {
	t.Run("GetModelName returns stored model name", func(t *testing.T) {
		testCases := []string{"gemini-pro", "gemini-1.5-pro", "custom-model"}

		for _, modelName := range testCases {
			client := &geminiClient{
				modelName: modelName,
				logger:    getTestLogger(),
			}

			result := client.GetModelName()
			if result != modelName {
				t.Errorf("Expected %q, got %q", modelName, result)
			}
		}
	})

	t.Run("Close handles nil client gracefully", func(t *testing.T) {
		client := &geminiClient{
			client:    nil,
			modelName: "test",
			logger:    getTestLogger(),
		}

		err := client.Close()
		if err != nil {
			t.Errorf("Close should not error with nil client: %v", err)
		}

		// Should be idempotent
		err = client.Close()
		if err != nil {
			t.Errorf("Second close should not error: %v", err)
		}
	})
}

// Test error creation and formatting
func TestErrorHandling(t *testing.T) {
	t.Run("CreateAPIError with custom message", func(t *testing.T) {
		customMsg := "Custom error message"
		err := CreateAPIError(llm.CategoryAuth, customMsg, nil, "test details")

		if err.Provider != "gemini" {
			t.Errorf("Expected provider 'gemini', got %q", err.Provider)
		}
		if err.Message != customMsg {
			t.Errorf("Expected message %q, got %q", customMsg, err.Message)
		}
		if err.Details != "test details" {
			t.Errorf("Expected details 'test details', got %q", err.Details)
		}
		if err.Category() != llm.CategoryAuth {
			t.Errorf("Expected CategoryAuth, got %v", err.Category())
		}
	})

	t.Run("CreateAPIError with original error", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := CreateAPIError(llm.CategoryNetwork, "", originalErr, "")

		if err.Message != "Network error while connecting to the Gemini API" {
			t.Errorf("Unexpected message: %q", err.Message)
		}
		if unwrapped := err.Unwrap(); unwrapped != originalErr {
			t.Error("Original error not preserved")
		}
	})

	t.Run("FormatAPIError with structured error", func(t *testing.T) {
		originalErr := errors.New("rate limit exceeded")
		formatted := FormatAPIError(originalErr, 429)

		if formatted.Category() != llm.CategoryRateLimit {
			t.Errorf("Expected CategoryRateLimit, got %v", formatted.Category())
		}
	})
}

// Test additional error scenarios and coverage paths
func TestAdditionalCoverage(t *testing.T) {
	t.Run("getErrorCategory with different status codes", func(t *testing.T) {
		testCases := []struct {
			err        error
			statusCode int
			expected   llm.ErrorCategory
		}{
			{errors.New("unauthorized"), 401, llm.CategoryAuth},
			{errors.New("forbidden"), 403, llm.CategoryAuth},
			{errors.New("bad request"), 400, llm.CategoryInvalidRequest},
			{errors.New("not found"), 404, llm.CategoryNotFound},
			{errors.New("server error"), 500, llm.CategoryServer},
			{errors.New("safety filter triggered"), 0, llm.CategoryContentFiltered},
			{errors.New("token limit exceeded"), 0, llm.CategoryInputLimit},
			{errors.New("network connection failed"), 0, llm.CategoryNetwork},
			{errors.New("context cancelled"), 0, llm.CategoryCancelled},
		}

		for _, tc := range testCases {
			category := getErrorCategory(tc.err, tc.statusCode)
			if category != tc.expected {
				t.Errorf("Error %q (status %d): expected %v, got %v", tc.err.Error(), tc.statusCode, tc.expected, category)
			}
		}
	})

	t.Run("CreateAPIError with all categories", func(t *testing.T) {
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
			err := CreateAPIError(category, "", nil, "")
			if err.Category() != category {
				t.Errorf("Expected category %v, got %v", category, err.Category())
			}
			if err.Provider != "gemini" {
				t.Errorf("Expected provider 'gemini', got %q", err.Provider)
			}
			if err.Suggestion == "" {
				t.Errorf("Expected non-empty suggestion for category %v", category)
			}
		}
	})

	t.Run("FormatAPIError with nil error", func(t *testing.T) {
		result := FormatAPIError(nil, 200)
		if result != nil {
			t.Errorf("Expected nil result for nil error, got %v", result)
		}
	})

	t.Run("IsAPIError backward compatibility", func(t *testing.T) {
		geminiErr := CreateAPIError(llm.CategoryAuth, "test", nil, "")
		apiErr, ok := IsAPIError(geminiErr)
		if !ok {
			t.Fatal("Expected IsAPIError to return true for Gemini error")
		}
		if apiErr != geminiErr {
			t.Error("Expected same error to be returned")
		}
	})

	t.Run("DefaultModelConfig validation", func(t *testing.T) {
		config := DefaultModelConfig()
		if config.MaxOutputTokens != 8192 {
			t.Errorf("Expected MaxOutputTokens 8192, got %d", config.MaxOutputTokens)
		}
		if config.Temperature != 0.3 {
			t.Errorf("Expected Temperature 0.3, got %f", config.Temperature)
		}
		if config.TopP != 0.9 {
			t.Errorf("Expected TopP 0.9, got %f", config.TopP)
		}
	})

	t.Run("IsGeminiError with different error types", func(t *testing.T) {
		// Test with nil error
		_, ok := IsGeminiError(nil)
		if ok {
			t.Error("Expected false for nil error")
		}

		// Test with regular error
		regularErr := errors.New("regular error")
		_, ok = IsGeminiError(regularErr)
		if ok {
			t.Error("Expected false for regular error")
		}

		// Test with non-Gemini LLMError
		otherProviderErr := llm.New("openai", "", 0, "test", "", nil, llm.CategoryAuth)
		_, ok = IsGeminiError(otherProviderErr)
		if ok {
			t.Error("Expected false for non-Gemini LLMError")
		}

		// Test with Gemini error
		geminiErr := CreateAPIError(llm.CategoryAuth, "test", nil, "")
		llmErr, ok := IsGeminiError(geminiErr)
		if !ok {
			t.Fatal("Expected true for Gemini error")
		}
		if llmErr.Provider != "gemini" {
			t.Errorf("Expected gemini provider, got %q", llmErr.Provider)
		}
	})

	t.Run("GetErrorType helper function", func(t *testing.T) {
		testCases := []struct {
			category     llm.ErrorCategory
			expectedType int
		}{
			{llm.CategoryUnknown, ErrorTypeUnknown},
			{llm.CategoryAuth, ErrorTypeAuth},
			{llm.CategoryRateLimit, ErrorTypeRateLimit},
			{llm.CategoryInvalidRequest, ErrorTypeInvalidRequest},
			{llm.CategoryNotFound, ErrorTypeNotFound},
			{llm.CategoryServer, ErrorTypeServer},
			{llm.CategoryNetwork, ErrorTypeNetwork},
			{llm.CategoryCancelled, ErrorTypeCancelled},
			{llm.CategoryInputLimit, ErrorTypeInputLimit},
			{llm.CategoryContentFiltered, ErrorTypeContentFiltered},
		}

		for _, tc := range testCases {
			err := CreateAPIError(tc.category, "test", nil, "")
			errorType := GetErrorType(err)
			if errorType != tc.expectedType {
				t.Errorf("Category %v: expected type %d, got %d", tc.category, tc.expectedType, errorType)
			}
		}
	})
}

// Test helper function coverage
func TestHelperFunctionCoverage(t *testing.T) {
	t.Run("mapSafetyRatings comprehensive", func(t *testing.T) {
		// Test all harm categories and probabilities
		input := []*genai.SafetyRating{
			{
				Category:    genai.HarmCategoryHarassment,
				Probability: genai.HarmProbabilityNegligible,
				Blocked:     false,
			},
			{
				Category:    genai.HarmCategoryDangerousContent,
				Probability: genai.HarmProbabilityHigh,
				Blocked:     true,
			},
		}

		result := mapSafetyRatings(input)

		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// Check first rating
		if result[0].Category != string(genai.HarmCategoryHarassment) {
			t.Errorf("Expected category %q, got %q", string(genai.HarmCategoryHarassment), result[0].Category)
		}
		if result[0].Blocked != false {
			t.Errorf("Expected blocked false, got %v", result[0].Blocked)
		}

		// Check second rating
		if result[1].Category != string(genai.HarmCategoryDangerousContent) {
			t.Errorf("Expected category %q, got %q", string(genai.HarmCategoryDangerousContent), result[1].Category)
		}
		if result[1].Blocked != true {
			t.Errorf("Expected blocked true, got %v", result[1].Blocked)
		}
	})

	t.Run("toProviderSafety comprehensive", func(t *testing.T) {
		input := []SafetyRating{
			{
				Category: "HARM_CATEGORY_HARASSMENT",
				Blocked:  false,
				Score:    0.1,
			},
			{
				Category: "HARM_CATEGORY_DANGEROUS_CONTENT",
				Blocked:  true,
				Score:    0.9,
			},
		}

		result := toProviderSafety(input)

		if len(result) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(result))
		}

		// Verify conversion preserves all fields
		for i, expected := range input {
			if result[i].Category != expected.Category {
				t.Errorf("Category %d: expected %q, got %q", i, expected.Category, result[i].Category)
			}
			if result[i].Blocked != expected.Blocked {
				t.Errorf("Blocked %d: expected %v, got %v", i, expected.Blocked, result[i].Blocked)
			}
			if result[i].Score != expected.Score {
				t.Errorf("Score %d: expected %v, got %v", i, expected.Score, result[i].Score)
			}
		}
	})
}

// Test functions with low coverage
func TestLowCoverageFunctions(t *testing.T) {
	t.Run("WithLogger option", func(t *testing.T) {
		// Test the WithLogger client option
		logger := getTestLogger()
		opt := WithLogger(logger)

		// Apply option to a client to test it
		client := &geminiClient{
			modelName: "test",
			logger:    getTestLogger(),
		}

		opt(client)

		// Verify the logger was set
		if client.logger != logger {
			t.Error("WithLogger option did not set the logger correctly")
		}
	})

	t.Run("NewClient success path", func(t *testing.T) {
		// This should succeed with validation, creating a mock client for backward compatibility
		client, err := NewClient(context.Background(), "test-key", "test-model", "")

		// We expect success due to mock client creation
		if err != nil {
			t.Fatalf("Expected no error during client creation, got: %v", err)
		}
		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}
	})

	t.Run("GetErrorType with unknown category", func(t *testing.T) {
		// Create an error with a category that does not map to legacy types
		err := llm.New("gemini", "", 0, "test", "", nil, llm.ErrorCategory(999))
		errorType := GetErrorType(err)
		if errorType != ErrorTypeUnknown {
			t.Errorf("Expected ErrorTypeUnknown for unmapped category, got %d", errorType)
		}
	})
}
