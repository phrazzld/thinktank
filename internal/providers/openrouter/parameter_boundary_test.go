package openrouter

import (
	"context"
	"net/http"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockClient creates a test client with a mock HTTP transport
func createMockClient(t *testing.T) *openrouterClient {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	// Create mock transport that returns default success responses
	mockTransport := &MockRoundTripper{}

	// Create HTTP client with mock transport
	httpClient := &http.Client{
		Transport: mockTransport,
	}

	// Create OpenRouter client with mock HTTP client
	client, err := NewClient("sk-or-test-key", "test/model", "http://mock-endpoint", logger, WithHTTPClient(httpClient))
	require.NoError(t, err)

	return client
}

// TestGenerateContentParameterBoundaries tests parameter boundary validation for OpenRouter provider
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
			name:   "valid max_output_tokens (alternative name)",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_output_tokens": 2048,
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
		{
			name:   "invalid max_output_tokens zero",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_output_tokens": 0,
			},
			expectError:   true,
			errorCategory: llm.CategoryInvalidRequest,
			errorContains: "max_output_tokens",
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

		// Test both max_tokens and max_output_tokens parameter names
		{
			name:   "both max_tokens and max_output_tokens (should prioritize max_tokens)",
			prompt: "test prompt",
			parameters: map[string]interface{}{
				"max_tokens":        1024,
				"max_output_tokens": 2048,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with mock HTTP transport
			client := createMockClient(t)
			ctx := context.Background()

			// Test parameter validation by calling GenerateContent with mocked HTTP client
			_, err := client.GenerateContent(ctx, tt.prompt, tt.parameters)

			if tt.expectError {
				// We expect an error from parameter validation
				assert.Error(t, err, "Expected error for test case: %s", tt.name)

				// Check that it's the right type of error
				if tt.errorCategory != llm.CategoryUnknown {
					assert.True(t, llm.IsCategory(err, tt.errorCategory),
						"Expected error category %s for test case: %s, got error: %v",
						tt.errorCategory, tt.name, err)
				}

				// Check error message contains expected content
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains,
						"Expected error to contain '%s' for test case: %s", tt.errorContains, tt.name)
				}
			} else {
				// We don't expect a validation error
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
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
			name: "max_output_tokens as int",
			parameters: map[string]interface{}{
				"max_output_tokens": int(1024),
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
			// Create client with mock HTTP transport
			client := createMockClient(t)
			ctx := context.Background()

			// Test parameter type conversion by calling GenerateContent with mocked HTTP client
			_, err := client.GenerateContent(ctx, "test prompt", tt.parameters)

			if tt.expectError {
				require.Error(t, err)
			} else {
				// With proper mocking, we should not get network errors
				// Parameter type conversion should work without panicking
				assert.NoError(t, err, "Parameter type conversion should work for test case: %s", tt.name)
			}
		})
	}
}

// TestParameterValidationLogic tests the internal parameter validation logic
func TestParameterValidationLogic(t *testing.T) {
	tests := []struct {
		name          string
		parameters    map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "valid parameters should not error in processing",
			parameters: map[string]interface{}{
				"temperature": 0.7,
				"top_p":       0.9,
				"max_tokens":  1024,
			},
			expectError: false,
		},
		{
			name:        "empty parameters should not error",
			parameters:  map[string]interface{}{},
			expectError: false,
		},
		{
			name:        "nil parameters should not error",
			parameters:  nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with mock HTTP transport
			client := createMockClient(t)

			// Test that parameter processing doesn't cause panics or validation errors
			// We're testing the internal logic, not the network calls
			ctx := context.Background()

			// This tests parameter conversion and processing logic with mocked HTTP client
			_, err := client.GenerateContent(ctx, "test prompt", tt.parameters)

			// With proper mocking, parameter processing should work without errors or panics
			// The absence of panics indicates successful parameter type conversion
			assert.NoError(t, err, "Parameter processing should work without errors for test case: %s", tt.name)
		})
	}
}
