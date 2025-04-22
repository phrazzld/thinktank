package openai

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorDebugInfo tests that errors include useful debug information
func TestErrorDebugInfo(t *testing.T) {
	tests := []struct {
		name               string
		statusCode         int
		responseBody       []byte
		transportErr       error
		expectDebugStrings []string
	}{
		{
			name:               "Authentication error debug info",
			statusCode:         401,
			responseBody:       makeOpenAIErrorResponse(401, "auth_error", "Invalid API key"),
			transportErr:       nil,
			expectDebugStrings: []string{"auth", "API key"},
		},
		{
			name:               "Rate limit error debug info",
			statusCode:         429,
			responseBody:       makeOpenAIErrorResponse(429, "rate_limit_exceeded", "Rate limit exceeded"),
			transportErr:       nil,
			expectDebugStrings: []string{"rate limit", "again"},
		},
		{
			name:               "Server error debug info",
			statusCode:         500,
			responseBody:       makeOpenAIErrorResponse(500, "server_error", "Internal server error"),
			transportErr:       nil,
			expectDebugStrings: []string{"server error", "try again"},
		},
		{
			name:               "Network error debug info",
			statusCode:         0,
			responseBody:       nil,
			transportErr:       errors.New("network error: connection refused"),
			expectDebugStrings: []string{"network", "connection", "try again"},
		},
		{
			name:               "Invalid request error debug info",
			statusCode:         400,
			responseBody:       makeOpenAIErrorResponse(400, "invalid_request_error", "Invalid model"),
			transportErr:       nil,
			expectDebugStrings: []string{"invalid request", "model"},
		},
		{
			name:               "Context length error debug info",
			statusCode:         400,
			responseBody:       makeOpenAIErrorResponse(400, "context_length_exceeded", "This model's maximum context length is 4097 tokens"),
			transportErr:       nil,
			expectDebugStrings: []string{"context", "token", "parameters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock transport for testing, which we'll use in a future implementation
			// that allows injecting HTTP clients directly
			_ = &ErrorMockRoundTripper{
				statusCode:   tt.statusCode,
				responseBody: tt.responseBody,
				err:          tt.transportErr,
			}

			// Since we can't fully inject the HTTP client yet, we'll test error handling functions directly
			// In a complete implementation, we would use setupTestClient and call client.GenerateContent

			var err error
			if tt.transportErr != nil {
				// Simulate transport errors
				err = openai.FormatAPIError(tt.transportErr, 0)
			} else {
				// Simulate HTTP response errors
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
				err = openai.FormatAPIError(errors.New(errorMsg), tt.statusCode)
			}

			// Assert that the error is not nil
			require.NotNil(t, err, "Expected an error but got nil")

			// Check for debug info
			var llmErr *llm.LLMError
			if errors.As(err, &llmErr) {
				debugInfo := llmErr.Error() + " " + llmErr.Suggestion
				debugInfo = strings.ToLower(debugInfo)
				for _, expectedStr := range tt.expectDebugStrings {
					assert.Contains(t, debugInfo, strings.ToLower(expectedStr),
						"Expected debug info to contain %q", expectedStr)
				}

				// Check debug field existence
				assert.NotEmpty(t, llmErr.DebugInfo, "Expected non-empty debug info")

				// Check provider field
				assert.Equal(t, "openai", llmErr.Provider, "Expected provider to be 'openai'")
			} else {
				t.Fatalf("Expected error to be of type *llm.LLMError")
			}
		})
	}
}
