package openai

import (
	"context"
	"os"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	// Test with logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")
	provider := NewProvider(logger)
	assert.NotNil(t, provider, "Provider should not be nil")
	assert.IsType(t, &OpenAIProvider{}, provider, "Provider should be of type *OpenAIProvider")

	// Test without logger (should create default logger)
	provider = NewProvider(nil)
	assert.NotNil(t, provider, "Provider should not be nil")
	assert.IsType(t, &OpenAIProvider{}, provider, "Provider should be of type *OpenAIProvider")
}

func TestCreateClient(t *testing.T) {
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")

	// Save current environment
	origAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		// Restore environment
		if origAPIKey != "" {
			_ = os.Setenv("OPENAI_API_KEY", origAPIKey)
		} else {
			_ = os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	tests := []struct {
		name         string
		apiKey       string
		envKey       string
		modelID      string
		apiEndpoint  string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Valid API key as parameter",
			apiKey:      "sk-test12345",
			envKey:      "",
			modelID:     "gpt-3.5-turbo",
			apiEndpoint: "",
			expectError: false,
		},
		{
			name:        "Valid API key in environment",
			apiKey:      "",
			envKey:      "sk-envtest12345",
			modelID:     "gpt-4",
			apiEndpoint: "",
			expectError: false,
		},
		{
			name:         "No API key provided",
			apiKey:       "",
			envKey:       "",
			modelID:      "gpt-3.5-turbo",
			apiEndpoint:  "",
			expectError:  true,
			errorMessage: "no valid OpenAI API key provided",
		},
		{
			name:         "Non-standard API key format",
			apiKey:       "not-an-sk-key",
			envKey:       "",
			modelID:      "gpt-3.5-turbo",
			apiEndpoint:  "",
			expectError:  true, // It will error as it doesn't look like an OpenAI key
			errorMessage: "no valid OpenAI API key provided",
		},
		{
			name:        "Custom API endpoint",
			apiKey:      "sk-test12345",
			envKey:      "",
			modelID:     "gpt-3.5-turbo",
			apiEndpoint: "https://custom.openai.api/v1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envKey != "" {
				err := os.Setenv("OPENAI_API_KEY", tt.envKey)
				require.NoError(t, err, "Failed to set environment variable")
			} else {
				err := os.Unsetenv("OPENAI_API_KEY")
				require.NoError(t, err, "Failed to unset environment variable")
			}

			// Create provider
			provider := NewProvider(logger)
			client, err := provider.CreateClient(ctx, tt.apiKey, tt.modelID, tt.apiEndpoint)

			// Check results
			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage, "Error message doesn't match expected")
				}
				assert.Nil(t, client, "Client should be nil when there's an error")
			} else {
				// For a successful client creation, we should have a valid client
				// but we can't test the actual API calls without mocking the OpenAI SDK
				if !assert.NoError(t, err, "Unexpected error: %v", err) {
					return
				}
				assert.NotNil(t, client, "Client should not be nil")
				assert.IsType(t, &OpenAIClientAdapter{}, client, "Client should be of type *OpenAIClientAdapter")

				// Check model name
				modelName := client.GetModelName()
				assert.Equal(t, tt.modelID, modelName, "Model name doesn't match expected")

				// Test Close method
				err = client.Close()
				assert.NoError(t, err, "Close should not return an error")
			}
		})
	}
}

func TestAdapterParameterHandling(t *testing.T) {
	// Create a mock LLM client to verify parameter handling
	mockClient := &MockLLMClient{
		modelName: "gpt-3.5-turbo",
	}

	// Create adapter with mock client
	adapter := NewOpenAIClientAdapter(mockClient)

	// Test parameters
	testParams := map[string]interface{}{
		"temperature":       0.7,
		"top_p":             0.95,
		"max_tokens":        100,
		"frequency_penalty": 0.5,
		"presence_penalty":  0.2,
	}

	// Set parameters
	adapter.SetParameters(testParams)

	// Verify adapter stored parameters
	assert.Equal(t, testParams, adapter.params, "Parameters not stored correctly")

	// Test GenerateContent with parameters
	ctx := context.Background()
	result, err := adapter.GenerateContent(ctx, "Test prompt", nil)
	assert.NoError(t, err, "GenerateContent should not return an error")
	assert.NotNil(t, result, "Result should not be nil")

	// Verify the mock client was called with the right parameters
	assert.Equal(t, "Test prompt", mockClient.lastPrompt, "Prompt was not passed correctly")

	// Override parameters with direct call
	newParams := map[string]interface{}{
		"temperature": 0.5,
		"top_p":       0.8,
	}

	_, err = adapter.GenerateContent(ctx, "New prompt", newParams)
	assert.NoError(t, err, "GenerateContent should not return an error")

	// Verify parameters were overridden
	assert.Equal(t, "New prompt", mockClient.lastPrompt, "Prompt was not passed correctly")
	assert.Equal(t, newParams, adapter.params, "Parameters were not overridden correctly")
}

func TestGetFloatParam(t *testing.T) {
	adapter := &OpenAIClientAdapter{
		params: map[string]interface{}{
			"float64_param": float64(1.5),
			"float32_param": float32(2.5),
			"int_param":     42,
			"int32_param":   int32(43),
			"int64_param":   int64(44),
			"string_param":  "not a number",
		},
	}

	tests := []struct {
		name          string
		paramName     string
		expectedValue float32
		expectedOk    bool
	}{
		{
			name:          "float64 parameter",
			paramName:     "float64_param",
			expectedValue: 1.5,
			expectedOk:    true,
		},
		{
			name:          "float32 parameter",
			paramName:     "float32_param",
			expectedValue: 2.5,
			expectedOk:    true,
		},
		{
			name:          "int parameter",
			paramName:     "int_param",
			expectedValue: 42.0,
			expectedOk:    true,
		},
		{
			name:          "int32 parameter",
			paramName:     "int32_param",
			expectedValue: 43.0,
			expectedOk:    true,
		},
		{
			name:          "int64 parameter",
			paramName:     "int64_param",
			expectedValue: 44.0,
			expectedOk:    true,
		},
		{
			name:          "string parameter (invalid type)",
			paramName:     "string_param",
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "missing parameter",
			paramName:     "missing_param",
			expectedValue: 0,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := adapter.getFloatParam(tt.paramName)
			assert.Equal(t, tt.expectedOk, ok, "Expected ok=%v but got %v", tt.expectedOk, ok)
			assert.Equal(t, tt.expectedValue, value, "Expected value=%v but got %v", tt.expectedValue, value)
		})
	}

	// Test with nil params
	adapter.params = nil
	value, ok := adapter.getFloatParam("any_param")
	assert.False(t, ok, "Expected ok=false with nil params")
	assert.Equal(t, float32(0), value, "Expected value=0 with nil params")
}

func TestGetIntParam(t *testing.T) {
	adapter := &OpenAIClientAdapter{
		params: map[string]interface{}{
			"float64_param": float64(1.5),
			"float32_param": float32(2.5),
			"int_param":     42,
			"int32_param":   int32(43),
			"int64_param":   int64(44),
			"string_param":  "not a number",
		},
	}

	tests := []struct {
		name          string
		paramName     string
		expectedValue int32
		expectedOk    bool
	}{
		{
			name:          "float64 parameter",
			paramName:     "float64_param",
			expectedValue: 1,
			expectedOk:    true,
		},
		{
			name:          "float32 parameter",
			paramName:     "float32_param",
			expectedValue: 2,
			expectedOk:    true,
		},
		{
			name:          "int parameter",
			paramName:     "int_param",
			expectedValue: 42,
			expectedOk:    true,
		},
		{
			name:          "int32 parameter",
			paramName:     "int32_param",
			expectedValue: 43,
			expectedOk:    true,
		},
		{
			name:          "int64 parameter",
			paramName:     "int64_param",
			expectedValue: 44,
			expectedOk:    true,
		},
		{
			name:          "string parameter (invalid type)",
			paramName:     "string_param",
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "missing parameter",
			paramName:     "missing_param",
			expectedValue: 0,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := adapter.getIntParam(tt.paramName)
			assert.Equal(t, tt.expectedOk, ok, "Expected ok=%v but got %v", tt.expectedOk, ok)
			assert.Equal(t, tt.expectedValue, value, "Expected value=%v but got %v", tt.expectedValue, value)
		})
	}

	// Test with nil params
	adapter.params = nil
	value, ok := adapter.getIntParam("any_param")
	assert.False(t, ok, "Expected ok=false with nil params")
	assert.Equal(t, int32(0), value, "Expected value=0 with nil params")
}

// MockLLMClient implements the llm.LLMClient interface for testing
type MockLLMClient struct {
	modelName       string
	lastPrompt      string
	lastParams      map[string]interface{}
	temperature     float32
	topP            float32
	maxTokens       int32
	freqPenalty     float32
	presencePenalty float32
	shouldError     bool
	mockResult      *llm.ProviderResult
}

// GenerateContent implements the llm.LLMClient interface
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	m.lastPrompt = prompt
	m.lastParams = params

	if m.shouldError {
		return nil, assert.AnError
	}

	if m.mockResult != nil {
		return m.mockResult, nil
	}

	// Default mock response
	return &llm.ProviderResult{
		Content: "Mock response for: " + prompt,
	}, nil
}

// GetModelName implements the llm.LLMClient interface
func (m *MockLLMClient) GetModelName() string {
	return m.modelName
}

// Close implements the llm.LLMClient interface
func (m *MockLLMClient) Close() error {
	return nil
}

// SetTemperature implements the temperature setter interface
func (m *MockLLMClient) SetTemperature(temp float32) {
	m.temperature = temp
}

// SetTopP implements the topP setter interface
func (m *MockLLMClient) SetTopP(topP float32) {
	m.topP = topP
}

// SetMaxTokens implements the maxTokens setter interface
func (m *MockLLMClient) SetMaxTokens(tokens int32) {
	m.maxTokens = tokens
}

// SetFrequencyPenalty implements the frequency penalty setter interface
func (m *MockLLMClient) SetFrequencyPenalty(penalty float32) {
	m.freqPenalty = penalty
}

// SetPresencePenalty implements the presence penalty setter interface
func (m *MockLLMClient) SetPresencePenalty(penalty float32) {
	m.presencePenalty = penalty
}
