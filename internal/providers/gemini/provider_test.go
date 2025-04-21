// internal/providers/gemini/provider_test.go
package gemini

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMClient implements the llm.LLMClient interface for testing
type MockLLMClient struct {
	GenerateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	GetModelNameFunc    func() string
	CloseFunc           func() error
	// For parameter setting
	temperature     float32
	topP            float32
	topK            int32
	maxOutputTokens int32
	lastPrompt      string
	lastParams      map[string]interface{}
}

// GenerateContent implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	m.lastPrompt = prompt
	m.lastParams = params

	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{Content: "Test response"}, nil
}

// GetModelName implements the llm.LLMClient interface for testing
func (m *MockLLMClient) GetModelName() string {
	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "test-model"
}

// Close implements the llm.LLMClient interface for testing
func (m *MockLLMClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// SetTemperature implements the temperature setter interface for testing
func (m *MockLLMClient) SetTemperature(temp float32) {
	m.temperature = temp
}

// SetTopP implements the topP setter interface for testing
func (m *MockLLMClient) SetTopP(topP float32) {
	m.topP = topP
}

// SetTopK implements the topK setter interface for testing
func (m *MockLLMClient) SetTopK(topK int32) {
	m.topK = topK
}

// SetMaxOutputTokens implements the maxOutputTokens setter interface for testing
func (m *MockLLMClient) SetMaxOutputTokens(tokens int32) {
	m.maxOutputTokens = tokens
}

// TestGeminiProviderImplementsProviderInterface verifies that GeminiProvider implements the Provider interface
func TestGeminiProviderImplementsProviderInterface(t *testing.T) {
	// This is a compile-time check to ensure GeminiProvider implements Provider
	var _ providers.Provider = (*GeminiProvider)(nil)
}

// TestNewProvider verifies that NewProvider creates a valid GeminiProvider
func TestNewProvider(t *testing.T) {
	// Create a provider with no logger (should use default)
	provider := NewProvider(nil)

	// Verify it's not nil
	assert.NotNil(t, provider, "Provider should not be nil")
	assert.IsType(t, &GeminiProvider{}, provider, "Provider should be of type *GeminiProvider")

	// Create a provider with custom logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")
	provider = NewProvider(logger)

	// Verify it's not nil
	assert.NotNil(t, provider, "Provider should not be nil")
	assert.IsType(t, &GeminiProvider{}, provider, "Provider should be of type *GeminiProvider")
}

// TestCreateClient tests the provider's CreateClient method
func TestCreateClient(t *testing.T) {
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")

	// Save current environment
	origAPIKey := os.Getenv("GOOGLE_API_KEY")
	defer func() {
		// Restore environment
		if origAPIKey != "" {
			_ = os.Setenv("GOOGLE_API_KEY", origAPIKey)
		} else {
			_ = os.Unsetenv("GOOGLE_API_KEY")
		}
	}()

	// Test no API key provided
	err := os.Unsetenv("GOOGLE_API_KEY")
	require.NoError(t, err, "Failed to unset environment variable")

	provider := NewProvider(logger)
	client, err := provider.CreateClient(ctx, "", "gemini-1.5-pro", "")

	assert.Error(t, err, "Expected error when no API key provided")
	assert.Contains(t, err.Error(), "no API key provided", "Error message should mention missing API key")
	assert.Nil(t, client, "Client should be nil when there's an error")

	// Other test cases would need mocking of the gemini.NewLLMClient function
	// which is not easily done without refactoring the code for better testability
}

// TestGeminiClientAdapter verifies that GeminiClientAdapter correctly wraps a client
func TestGeminiClientAdapter(t *testing.T) {
	// Create a mock client
	mockClient := &MockLLMClient{}

	// Create an adapter
	adapter := NewGeminiClientAdapter(mockClient)

	// Test that adapter implements llm.LLMClient
	var _ llm.LLMClient = adapter

	// Test SetParameters
	params := map[string]interface{}{
		"temperature":       0.7,
		"top_p":             0.9,
		"top_k":             40,
		"max_output_tokens": 100,
	}
	adapter.SetParameters(params)

	// Verify parameters were stored
	assert.Equal(t, params, adapter.params, "Parameters were not stored correctly")

	// Test GenerateContent pass-through
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{Content: "Mock response"}, nil
	}
	result, err := adapter.GenerateContent(context.Background(), "test prompt", nil)
	assert.NoError(t, err, "Expected no error from GenerateContent")
	assert.Equal(t, "Mock response", result.Content, "Unexpected response content")
	assert.Equal(t, "test prompt", mockClient.lastPrompt, "Prompt was not passed correctly")

	// Test error case
	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return nil, errors.New("mock error")
	}
	_, err = adapter.GenerateContent(context.Background(), "test prompt", nil)
	assert.Error(t, err, "Expected error from GenerateContent")
	assert.Contains(t, err.Error(), "mock error", "Error message should contain original error")

	// Test parameter overrides in GenerateContent
	newParams := map[string]interface{}{
		"temperature": 0.5,
		"top_p":       0.8,
	}

	mockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{Content: "New response"}, nil
	}

	result, err = adapter.GenerateContent(context.Background(), "another prompt", newParams)
	assert.NoError(t, err, "Expected no error from GenerateContent with new params")
	assert.Equal(t, "New response", result.Content, "Unexpected response content")
	assert.Equal(t, "another prompt", mockClient.lastPrompt, "Prompt was not passed correctly")
	assert.Equal(t, newParams, adapter.params, "Parameters were not overridden correctly")

	// Test GetModelName pass-through
	mockClient.GetModelNameFunc = func() string {
		return "gemini-1.5-pro"
	}
	modelName := adapter.GetModelName()
	assert.Equal(t, "gemini-1.5-pro", modelName, "GetModelName did not pass through correctly")

	// Test Close pass-through
	closeCalled := false
	mockClient.CloseFunc = func() error {
		closeCalled = true
		return nil
	}
	err = adapter.Close()
	assert.NoError(t, err, "Close should not return an error")
	assert.True(t, closeCalled, "Close was not called on the underlying client")
}

// TestParamTypeConversion tests the parameter type conversion functions
func TestParamTypeConversion(t *testing.T) {
	adapter := &GeminiClientAdapter{
		params: map[string]interface{}{
			"float64_param": float64(1.5),
			"float32_param": float32(2.5),
			"int_param":     42,
			"int32_param":   int32(43),
			"int64_param":   int64(44),
			"string_param":  "not a number",
		},
	}

	// Test getFloatParam
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
		t.Run("Float:"+tt.name, func(t *testing.T) {
			value, ok := adapter.getFloatParam(tt.paramName)
			assert.Equal(t, tt.expectedOk, ok, "Expected ok=%v but got %v", tt.expectedOk, ok)
			assert.Equal(t, tt.expectedValue, value, "Expected value=%v but got %v", tt.expectedValue, value)
		})
	}

	// Test getIntParam
	intTests := []struct {
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

	for _, tt := range intTests {
		t.Run("Int:"+tt.name, func(t *testing.T) {
			value, ok := adapter.getIntParam(tt.paramName)
			assert.Equal(t, tt.expectedOk, ok, "Expected ok=%v but got %v", tt.expectedOk, ok)
			assert.Equal(t, tt.expectedValue, value, "Expected value=%v but got %v", tt.expectedValue, value)
		})
	}

	// Test with nil params
	adapter.params = nil
	floatVal, floatOk := adapter.getFloatParam("any_param")
	assert.False(t, floatOk, "Expected ok=false with nil params for float")
	assert.Equal(t, float32(0), floatVal, "Expected value=0 with nil params for float")

	intVal, intOk := adapter.getIntParam("any_param")
	assert.False(t, intOk, "Expected ok=false with nil params for int")
	assert.Equal(t, int32(0), intVal, "Expected value=0 with nil params for int")
}

// TestParameterApplication tests that the adapter correctly applies parameters to the client
func TestParameterApplication(t *testing.T) {
	// Create a mock client
	mockClient := &MockLLMClient{}

	// Create an adapter
	adapter := NewGeminiClientAdapter(mockClient)

	// Setup parameters
	params := map[string]interface{}{
		"temperature":       0.7,
		"top_p":             0.9,
		"top_k":             40,
		"max_output_tokens": 100,
	}

	// Apply parameters and call GenerateContent
	adapter.SetParameters(params)
	_, err := adapter.GenerateContent(context.Background(), "test prompt", nil)
	assert.NoError(t, err, "GenerateContent should not return an error")

	// Verify parameters were applied to the client
	assert.Equal(t, float32(0.7), mockClient.temperature, "Temperature was not set correctly")
	assert.Equal(t, float32(0.9), mockClient.topP, "TopP was not set correctly")
	assert.Equal(t, int32(40), mockClient.topK, "TopK was not set correctly")
	assert.Equal(t, int32(100), mockClient.maxOutputTokens, "MaxOutputTokens was not set correctly")

	// Test with OpenAI-style parameter names
	params = map[string]interface{}{
		"temperature": 0.5,
		"top_p":       0.8,
		"max_tokens":  200, // OpenAI style instead of max_output_tokens
	}

	// Apply parameters and call GenerateContent
	_, err = adapter.GenerateContent(context.Background(), "test prompt", params)
	assert.NoError(t, err, "GenerateContent should not return an error")

	// Verify parameters were applied to the client
	assert.Equal(t, float32(0.5), mockClient.temperature, "Temperature was not set correctly")
	assert.Equal(t, float32(0.8), mockClient.topP, "TopP was not set correctly")
	assert.Equal(t, int32(200), mockClient.maxOutputTokens, "MaxOutputTokens was not set correctly with OpenAI-style param name")
}
