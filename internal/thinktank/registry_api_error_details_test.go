// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"errors"
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/gemini"
	"github.com/phrazzld/thinktank/internal/llm"
	// We need openai package for the error detection in GetErrorDetails,
	// but we don't call its functions directly in the test
	_ "github.com/phrazzld/thinktank/internal/openai"
)

// TestGetErrorDetails tests the GetErrorDetails method
func TestGetErrorDetails(t *testing.T) {
	service, _, _ := setupTest(t)

	// Create test errors for Gemini
	geminiErrWithSuggestion := llm.New("gemini", "code1", 400, "Gemini API error", "req123", nil, llm.CategoryAuth)
	geminiErrWithSuggestion.Suggestion = "Check your Gemini API key and permissions"

	geminiErrWithoutSuggestion := llm.New("gemini", "code2", 429, "Rate limit exceeded", "req456", nil, llm.CategoryRateLimit)

	// Create test errors for OpenAI
	openaiErrWithSuggestion := llm.New("openai", "code3", 403, "OpenAI authentication error", "req789", nil, llm.CategoryAuth)
	openaiErrWithSuggestion.Suggestion = "Verify your OpenAI API key is correct and active"

	openaiErrWithoutSuggestion := llm.New("openai", "code4", 500, "OpenAI server error", "req101", nil, llm.CategoryServer)

	// Create wrapped errors
	wrappedGeminiErr := fmt.Errorf("wrapper: %w", geminiErrWithSuggestion)
	wrappedOpenaiErr := fmt.Errorf("wrapper: %w", openaiErrWithSuggestion)

	// Standard errors
	standardErr := errors.New("standard error message")
	wrappedStandardErr := fmt.Errorf("wrapper: %w", standardErr)

	// Complex error with nested original
	nestedErr := errors.New("nested error")
	complexErr := llm.New("gemini", "complex", 400, "Complex error", "req999", nestedErr, llm.CategoryInvalidRequest)
	complexErr.Suggestion = "Try simplifying your request"

	// Define test cases
	testCases := []struct {
		name           string
		err            error
		expectedOutput string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedOutput: "no error",
		},
		{
			name:           "gemini error with suggestion",
			err:            geminiErrWithSuggestion,
			expectedOutput: "Gemini API error\n\nSuggestion: Check your Gemini API key and permissions",
		},
		{
			name:           "gemini error without suggestion",
			err:            geminiErrWithoutSuggestion,
			expectedOutput: "Rate limit exceeded",
		},
		{
			name:           "openai error with suggestion",
			err:            openaiErrWithSuggestion,
			expectedOutput: "OpenAI authentication error\n\nSuggestion: Verify your OpenAI API key is correct and active",
		},
		{
			name:           "openai error without suggestion",
			err:            openaiErrWithoutSuggestion,
			expectedOutput: "OpenAI server error",
		},
		{
			name:           "wrapped gemini error",
			err:            wrappedGeminiErr,
			expectedOutput: "Gemini API error\n\nSuggestion: Check your Gemini API key and permissions",
		},
		{
			name:           "wrapped openai error",
			err:            wrappedOpenaiErr,
			expectedOutput: "OpenAI authentication error\n\nSuggestion: Verify your OpenAI API key is correct and active",
		},
		{
			name:           "standard error",
			err:            standardErr,
			expectedOutput: "standard error message",
		},
		{
			name:           "wrapped standard error",
			err:            wrappedStandardErr,
			expectedOutput: "wrapper: standard error message",
		},
		{
			name:           "complex error with nested original",
			err:            complexErr,
			expectedOutput: "Complex error: nested error\n\nSuggestion: Try simplifying your request",
		},
	}

	// Explicitly register the errors with the gemini and openai packages
	// This ensures the error type assertion works correctly in the GetErrorDetails method
	_ = gemini.FormatAPIError(geminiErrWithSuggestion, 400)
	_ = gemini.FormatAPIError(geminiErrWithoutSuggestion, 429)
	_ = gemini.FormatAPIError(complexErr, 400)

	// For OpenAI errors, we would use the FormatAPIError function as well
	// However, since we're using pre-created llm.LLMError instances with the openai provider,
	// the IsOpenAIError function should be able to detect them correctly

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method being tested
			details := service.GetErrorDetails(tc.err)

			// Verify output matches expected
			if details != tc.expectedOutput {
				t.Errorf("Expected error details '%s', got '%s'", tc.expectedOutput, details)
			}
		})
	}
}
