package openai

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidationErrors tests error handling for request validation failures
func TestValidationErrors(t *testing.T) {
	t.Parallel() // Pure CPU-bound parameter validation test
	tests := []struct {
		name                string
		statusCode          int
		responseBody        []byte
		expectErrorContains string
		expectErrorCategory llm.ErrorCategory
	}{
		{
			name:                "Invalid model",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid model"),
			expectErrorContains: "Invalid model",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
		{
			name:                "Input token limit exceeded",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "context_length_exceeded", "This model's maximum context length is 4097 tokens"),
			expectErrorContains: "context length",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
		{
			name:                "Invalid temperature parameter",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid temperature: must be between 0 and 2"),
			expectErrorContains: "temperature",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
		{
			name:                "Invalid top_p parameter",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid top_p: must be between 0 and 1"),
			expectErrorContains: "top_p",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
		{
			name:                "Invalid max_tokens parameter",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid max_tokens: must be positive"),
			expectErrorContains: "max_tokens",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
		{
			name:                "Invalid frequency_penalty parameter",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid frequency_penalty: must be between -2 and 2"),
			expectErrorContains: "frequency_penalty",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
		{
			name:                "Invalid presence_penalty parameter",
			statusCode:          400,
			responseBody:        makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid presence_penalty: must be between -2 and 2"),
			expectErrorContains: "presence_penalty",
			expectErrorCategory: llm.CategoryInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock transport for testing, which we'll use in a future implementation
			// that allows injecting HTTP clients directly
			_ = &ErrorMockRoundTripper{
				statusCode:   tt.statusCode,
				responseBody: tt.responseBody,
			}

			// Since we can't fully inject the HTTP client yet, we'll test error handling functions directly
			// In a complete implementation, we would use setupTestClient and call client.GenerateContent

			var errorMsg string
			if len(tt.responseBody) > 0 {
				var respMap map[string]interface{}
				if jsonErr := json.Unmarshal(tt.responseBody, &respMap); jsonErr == nil {
					if errObj, hasError := respMap["error"].(map[string]interface{}); hasError {
						if msg, hasMsg := errObj["message"].(string); hasMsg {
							errorMsg = msg
						}
					}
				} else {
					errorMsg = "Failed to parse error response"
				}
			} else {
				errorMsg = "Empty response from API"
			}
			err := openai.FormatAPIError(errors.New(errorMsg), tt.statusCode)

			// Assert that the error is not nil
			require.NotNil(t, err, "Expected an error but got nil")

			// Check error message contains expected text
			assert.Contains(t, err.Error(), tt.expectErrorContains,
				"Expected error message to contain %q, got %q", tt.expectErrorContains, err.Error())

			// Check error category
			var llmErr *llm.LLMError
			if errors.As(err, &llmErr) {
				assert.Equal(t, "openai", llmErr.Provider)
				assert.Equal(t, tt.expectErrorCategory, llmErr.Category(),
					"Expected error category to be %v, got %v", tt.expectErrorCategory, llmErr.Category())
				assert.NotEmpty(t, llmErr.Suggestion, "Expected non-empty suggestion for error")
			} else {
				t.Fatalf("Expected error to be of type *llm.LLMError")
			}
		})
	}
}
