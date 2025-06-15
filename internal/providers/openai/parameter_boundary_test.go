package openai

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateContentParameterBoundaries tests parameter boundary validation for OpenAI provider
func TestGenerateContentParameterBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		prompt        string
		parameters    map[string]interface{}
		expectError   bool
		errorCategory llm.ErrorCategory
		errorContains string
	}{
		// Temperature boundary tests
		{
			name:   "valid temperature lower bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 0.0,
			},
			expectError: false,
		},
		{
			name:   "valid temperature middle range",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 1.0,
			},
			expectError: false,
		},
		{
			name:   "valid temperature upper bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 2.0,
			},
			expectError: false,
		},
		{
			name:   "invalid temperature below minimum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": -0.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "temperature",
		},
		{
			name:   "invalid temperature above maximum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 2.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "temperature",
		},

		// TopP boundary tests
		{
			name:   "valid top_p lower bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": 0.0,
			},
			expectError: false,
		},
		{
			name:   "valid top_p middle range",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": 0.5,
			},
			expectError: false,
		},
		{
			name:   "valid top_p upper bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": 1.0,
			},
			expectError: false,
		},
		{
			name:   "invalid top_p below minimum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": -0.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "top_p",
		},
		{
			name:   "invalid top_p above maximum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": 1.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "top_p",
		},

		// MaxTokens boundary tests
		{
			name:   "valid max_tokens minimum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens": 1,
			},
			expectError: false,
		},
		{
			name:   "valid max_tokens middle range",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens": 1000,
			},
			expectError: false,
		},
		{
			name:   "valid max_tokens large value",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens": 4096,
			},
			expectError: false,
		},
		{
			name:   "invalid max_tokens zero",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens": 0,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "max_tokens",
		},
		{
			name:   "invalid max_tokens negative",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens": -1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "max_tokens",
		},

		// FrequencyPenalty boundary tests
		{
			name:   "valid frequency_penalty lower bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"frequency_penalty": -2.0,
			},
			expectError: false,
		},
		{
			name:   "valid frequency_penalty middle range",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"frequency_penalty": 0.0,
			},
			expectError: false,
		},
		{
			name:   "valid frequency_penalty upper bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"frequency_penalty": 2.0,
			},
			expectError: false,
		},
		{
			name:   "invalid frequency_penalty below minimum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"frequency_penalty": -2.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "frequency_penalty",
		},
		{
			name:   "invalid frequency_penalty above maximum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"frequency_penalty": 2.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "frequency_penalty",
		},

		// PresencePenalty boundary tests
		{
			name:   "valid presence_penalty lower bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"presence_penalty": -2.0,
			},
			expectError: false,
		},
		{
			name:   "valid presence_penalty middle range",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"presence_penalty": 0.0,
			},
			expectError: false,
		},
		{
			name:   "valid presence_penalty upper bound",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"presence_penalty": 2.0,
			},
			expectError: false,
		},
		{
			name:   "invalid presence_penalty below minimum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"presence_penalty": -2.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "presence_penalty",
		},
		{
			name:   "invalid presence_penalty above maximum",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"presence_penalty": 2.1,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "presence_penalty",
		},

		// Multiple parameter combinations
		{
			name:   "valid multiple parameters",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature":       1.0,
				"top_p":             0.9,
				"max_tokens":        2048,
				"frequency_penalty": 0.5,
				"presence_penalty":  0.5,
			},
			expectError: false,
		},
		{
			name:   "multiple invalid parameters",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature":       -0.5,
				"top_p":             1.5,
				"max_tokens":        -100,
				"frequency_penalty": 3.0,
				"presence_penalty":  -3.0,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "parameter",
		},

		// Edge cases with different type formats
		{
			name:   "temperature as int (valid)",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 1,
			},
			expectError: false,
		},
		{
			name:   "temperature as float64 (valid)",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": float64(0.7),
			},
			expectError: false,
		},
		{
			name:   "temperature as float32 (valid)",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": float32(0.8),
			},
			expectError: false,
		},
		{
			name:   "max_tokens as float64 (valid)",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens": float64(1024),
			},
			expectError: false,
		},

		// Empty prompt test
		{
			name:          "empty prompt with valid parameters",
			prompt:        "",
			parameters:    map[string]interface{}{"temperature": 0.7},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "prompt",
		},

		// Boundary edge cases
		{
			name:   "temperature exactly zero",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 0.0,
			},
			expectError: false,
		},
		{
			name:   "temperature exactly two",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"temperature": 2.0,
			},
			expectError: false,
		},
		{
			name:   "top_p exactly zero",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": 0.0,
			},
			expectError: false,
		},
		{
			name:   "top_p exactly one",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"top_p": 1.0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client adapter with mock client
			mockClient := &llm.MockLLMClient{
				GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					// Mock client that always returns success for testing parameter validation
					return &llm.ProviderResult{
						Content:      "mock response",
						FinishReason: "stop",
						Truncated:    false,
					}, nil
				},
			}

			adapter := NewOpenAIClientAdapter(mockClient)

			// Test parameter validation by calling GenerateContent
			ctx := context.Background()
			result, err := adapter.GenerateContent(ctx, tt.prompt, tt.parameters)

			if tt.expectError {
				// We expect an error
				assert.Error(t, err, "Expected error for test case: %s", tt.name)

				if tt.errorCategory != llm.CategoryUnknown {
					// Check if error is the expected LLM error type
					var llmErr *llm.LLMError
					if assert.ErrorAs(t, err, &llmErr) {
						assert.Equal(t, tt.errorCategory, llmErr.Category(),
							"Expected error category %v, got %v", tt.errorCategory, llmErr.Category())
					}
				}

				if tt.errorContains != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorContains,
						"Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				// We don't expect an error
				if err != nil {
					t.Errorf("Unexpected error for test case %s: %v", tt.name, err)
				}

				// If no error, result should be valid
				assert.NotNil(t, result, "Expected non-nil result for successful case")
				if result != nil {
					assert.NotEmpty(t, result.Content, "Expected non-empty content")
				}
			}
		})
	}
}

// TestParameterTypeConversion tests that different parameter types are handled correctly
func TestParameterTypeConversion(t *testing.T) {
	tests := []struct {
		name        string
		parameters  map[string]interface{}
		expectError bool
	}{
		{
			name: "temperature as different numeric types",
			parameters: map[string]interface{}{
				"temperature": int(1),
			},
			expectError: false,
		},
		{
			name: "temperature as float64",
			parameters: map[string]interface{}{
				"temperature": float64(0.7),
			},
			expectError: false,
		},
		{
			name: "temperature as float32",
			parameters: map[string]interface{}{
				"temperature": float32(0.8),
			},
			expectError: false,
		},
		{
			name: "max_tokens as different integer types",
			parameters: map[string]interface{}{
				"max_tokens": int32(1024),
			},
			expectError: false,
		},
		{
			name: "max_tokens as int64",
			parameters: map[string]interface{}{
				"max_tokens": int64(2048),
			},
			expectError: false,
		},
		{
			name: "frequency_penalty as float32",
			parameters: map[string]interface{}{
				"frequency_penalty": float32(0.5),
			},
			expectError: false,
		},
		{
			name: "presence_penalty as float64",
			parameters: map[string]interface{}{
				"presence_penalty": float64(0.5),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client adapter with mock client
			mockClient := &llm.MockLLMClient{
				GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return &llm.ProviderResult{
						Content:      "mock response",
						FinishReason: "stop",
						Truncated:    false,
					}, nil
				},
			}

			adapter := NewOpenAIClientAdapter(mockClient)

			ctx := context.Background()
			result, err := adapter.GenerateContent(ctx, "test prompt", tt.parameters)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}
