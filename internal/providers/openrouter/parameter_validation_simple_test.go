package openrouter

import (
	"context"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParameterValidationDirect tests parameter validation without making HTTP calls
func TestParameterValidationDirect(t *testing.T) {
	tests := []struct {
		name          string
		parameters    map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "valid parameters",
			parameters: map[string]interface{}{
				"temperature": 0.7,
				"top_p":       0.9,
				"max_tokens":  1024,
			},
			expectError: false,
		},
		{
			name: "invalid temperature below minimum",
			parameters: map[string]interface{}{
				"temperature": -0.1,
			},
			expectError:   true,
			errorContains: "temperature must be between 0.0 and 2.0",
		},
		{
			name: "invalid temperature above maximum",
			parameters: map[string]interface{}{
				"temperature": 2.1,
			},
			expectError:   true,
			errorContains: "temperature must be between 0.0 and 2.0",
		},
		{
			name: "invalid top_p below minimum",
			parameters: map[string]interface{}{
				"top_p": -0.1,
			},
			expectError:   true,
			errorContains: "top_p must be between 0.0 and 1.0",
		},
		{
			name: "invalid top_p above maximum",
			parameters: map[string]interface{}{
				"top_p": 1.1,
			},
			expectError:   true,
			errorContains: "top_p must be between 0.0 and 1.0",
		},
		{
			name: "invalid max_tokens zero",
			parameters: map[string]interface{}{
				"max_tokens": 0,
			},
			expectError:   true,
			errorContains: "max_tokens must be positive",
		},
		{
			name: "invalid max_tokens negative",
			parameters: map[string]interface{}{
				"max_tokens": -1,
			},
			expectError:   true,
			errorContains: "max_tokens must be positive",
		},
		{
			name: "invalid frequency_penalty below minimum",
			parameters: map[string]interface{}{
				"frequency_penalty": -2.1,
			},
			expectError:   true,
			errorContains: "frequency_penalty must be between -2.0 and 2.0",
		},
		{
			name: "invalid presence_penalty above maximum",
			parameters: map[string]interface{}{
				"presence_penalty": 2.1,
			},
			expectError:   true,
			errorContains: "presence_penalty must be between -2.0 and 2.0",
		},
		{
			name: "empty prompt",
			parameters: map[string]interface{}{
				"temperature": 0.7,
			},
			expectError:   true,
			errorContains: "Empty prompt provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
			client, err := NewClient("sk-or-test-key", "test/model", "http://mock-endpoint", logger)
			require.NoError(t, err)

			ctx := context.Background()
			prompt := "test prompt"

			// For the empty prompt test, use empty prompt
			if tt.name == "empty prompt" {
				prompt = ""
			}

			// Test parameter validation by calling validateParameters directly
			if tt.name != "empty prompt" {
				err = client.validateParameters(tt.parameters)
			} else {
				// For empty prompt test, call GenerateContent which will validate the prompt
				_, err = client.GenerateContent(ctx, prompt, tt.parameters)
			}

			if tt.expectError {
				require.Error(t, err, "Expected error for test case: %s", tt.name)

				// Check if error is the expected LLM error type
				var llmErr *llm.LLMError
				if assert.ErrorAs(t, err, &llmErr) {
					assert.Equal(t, llm.CategoryInvalidRequest, llmErr.Category())
				}

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains,
						"Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
			}
		})
	}
}
