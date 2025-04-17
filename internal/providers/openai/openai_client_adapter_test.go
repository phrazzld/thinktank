// Package openai contains tests for the OpenAI client adapter
package openai

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMClientWithParamChecking is a specialized mock for tracking parameter-related calls
type MockLLMClientWithParamChecking struct {
	// Mock function implementations
	GenerateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	CountTokensFunc     func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error)
	GetModelInfoFunc    func(ctx context.Context) (*llm.ProviderModelInfo, error)
	GetModelNameFunc    func() string
	CloseFunc           func() error

	// Parameter tracking
	TemperatureSet       bool
	TemperatureValue     float32
	TopPSet              bool
	TopPValue            float32
	MaxTokensSet         bool
	MaxTokensValue       int32
	FreqPenaltySet       bool
	FreqPenaltyValue     float32
	PresencePenaltySet   bool
	PresencePenaltyValue float32
	GenerateParams       map[string]interface{}

	// Method call tracking
	GenerateContentCalled bool
	CountTokensCalled     bool
	GetModelInfoCalled    bool
	GetModelNameCalled    bool
	CloseCalled           bool
}

// SetTemperature implementation for parameter checking
func (m *MockLLMClientWithParamChecking) SetTemperature(temp float32) {
	m.TemperatureSet = true
	m.TemperatureValue = temp
}

// SetTopP implementation for parameter checking
func (m *MockLLMClientWithParamChecking) SetTopP(topP float32) {
	m.TopPSet = true
	m.TopPValue = topP
}

// SetMaxTokens implementation for parameter checking
func (m *MockLLMClientWithParamChecking) SetMaxTokens(tokens int32) {
	m.MaxTokensSet = true
	m.MaxTokensValue = tokens
}

// SetFrequencyPenalty implementation for parameter checking
func (m *MockLLMClientWithParamChecking) SetFrequencyPenalty(penalty float32) {
	m.FreqPenaltySet = true
	m.FreqPenaltyValue = penalty
}

// SetPresencePenalty implementation for parameter checking
func (m *MockLLMClientWithParamChecking) SetPresencePenalty(penalty float32) {
	m.PresencePenaltySet = true
	m.PresencePenaltyValue = penalty
}

// GenerateContent overrides the method to track calls
func (m *MockLLMClientWithParamChecking) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	m.GenerateContentCalled = true
	m.GenerateParams = params

	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{Content: "Mock response"}, nil
}

// CountTokens overrides the method to track calls
func (m *MockLLMClientWithParamChecking) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	m.CountTokensCalled = true

	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, prompt)
	}
	return &llm.ProviderTokenCount{Total: int32(len(prompt) / 4)}, nil
}

// GetModelInfo overrides the method to track calls
func (m *MockLLMClientWithParamChecking) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	m.GetModelInfoCalled = true

	if m.GetModelInfoFunc != nil {
		return m.GetModelInfoFunc(ctx)
	}
	return &llm.ProviderModelInfo{
		Name:             "gpt-4",
		InputTokenLimit:  8192,
		OutputTokenLimit: 2048,
	}, nil
}

// GetModelName overrides the method to track calls
func (m *MockLLMClientWithParamChecking) GetModelName() string {
	m.GetModelNameCalled = true

	if m.GetModelNameFunc != nil {
		return m.GetModelNameFunc()
	}
	return "gpt-4"
}

// Close overrides the method to track calls
func (m *MockLLMClientWithParamChecking) Close() error {
	m.CloseCalled = true

	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// TestNewOpenAIClientAdapter verifies that NewOpenAIClientAdapter correctly initializes
// the adapter with all required fields
func TestNewOpenAIClientAdapter(t *testing.T) {
	// Create mock client
	mockClient := &MockLLMClient{}

	// Create adapter
	adapter := NewOpenAIClientAdapter(mockClient)

	// Verify adapter is properly initialized
	require.NotNil(t, adapter, "Adapter should not be nil")
	assert.Equal(t, mockClient, adapter.client, "Adapter should store the provided client")
	assert.NotNil(t, adapter.params, "Adapter should initialize params map")
	assert.Empty(t, adapter.params, "Params map should be empty initially")

	// Verify adapter implements LLMClient interface
	var _ llm.LLMClient = adapter
}

// TestSetParameters verifies that SetParameters correctly updates the parameter map
func TestSetParameters(t *testing.T) {
	// Create adapter with mock client
	mockClient := &MockLLMClient{}
	adapter := NewOpenAIClientAdapter(mockClient)

	// Initial params should be empty
	assert.Empty(t, adapter.params, "Initial params should be empty")

	// Set parameters
	testParams := map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.8,
		"max_tokens":  100,
	}
	adapter.SetParameters(testParams)

	// Verify parameters are stored
	assert.Equal(t, testParams, adapter.params, "SetParameters should store the provided map")

	// Update parameters
	updatedParams := map[string]interface{}{
		"temperature":       0.5,
		"frequency_penalty": 0.2,
	}
	adapter.SetParameters(updatedParams)

	// Verify parameters are completely replaced, not merged
	assert.Equal(t, updatedParams, adapter.params, "SetParameters should replace the entire params map")
}

// TestGenerateContentDelegation verifies that GenerateContent correctly delegates to the
// underlying client and applies parameters
func TestGenerateContentDelegation(t *testing.T) {
	tests := []struct {
		name                string
		adapterParams       map[string]interface{}
		requestParams       map[string]interface{}
		expectedTemperature float32
		expectedTopP        float32
		expectedMaxTokens   int32
		expectedFreqPenalty float32
		expectedPresPenalty float32
		mockResponse        *llm.ProviderResult
		mockError           error
	}{
		{
			name:                "No parameters",
			adapterParams:       nil,
			requestParams:       nil,
			expectedTemperature: 0,
			expectedTopP:        0,
			expectedMaxTokens:   0,
			expectedFreqPenalty: 0,
			expectedPresPenalty: 0,
			mockResponse:        &llm.ProviderResult{Content: "Response with no parameters"},
			mockError:           nil,
		},
		{
			name: "Adapter-level parameters",
			adapterParams: map[string]interface{}{
				"temperature":       0.7,
				"top_p":             0.8,
				"max_tokens":        100,
				"frequency_penalty": 0.2,
				"presence_penalty":  0.3,
			},
			requestParams:       nil,
			expectedTemperature: 0.7,
			expectedTopP:        0.8,
			expectedMaxTokens:   100,
			expectedFreqPenalty: 0.2,
			expectedPresPenalty: 0.3,
			mockResponse:        &llm.ProviderResult{Content: "Response with adapter parameters"},
			mockError:           nil,
		},
		{
			name: "Request-level parameters override adapter parameters",
			adapterParams: map[string]interface{}{
				"temperature":       0.7,
				"top_p":             0.8,
				"max_tokens":        100,
				"frequency_penalty": 0.2,
				"presence_penalty":  0.3,
			},
			requestParams: map[string]interface{}{
				"temperature":      0.5,
				"max_tokens":       200,
				"presence_penalty": 0.1,
			},
			expectedTemperature: 0.5,
			expectedTopP:        0, // Not checking this parameter in this test
			expectedMaxTokens:   200,
			expectedFreqPenalty: 0, // Not checking this parameter in this test
			expectedPresPenalty: 0.1,
			mockResponse:        &llm.ProviderResult{Content: "Response with overridden parameters"},
			mockError:           nil,
		},
		{
			name:                "Error case",
			adapterParams:       nil,
			requestParams:       nil,
			expectedTemperature: 0,
			expectedTopP:        0,
			expectedMaxTokens:   0,
			expectedFreqPenalty: 0,
			expectedPresPenalty: 0,
			mockResponse:        nil,
			mockError:           errors.New("mock API error"),
		},
		{
			name:          "Alternative parameter names (Gemini-style)",
			adapterParams: nil,
			requestParams: map[string]interface{}{
				"max_output_tokens": 150, // Gemini-style parameter
			},
			expectedTemperature: 0,
			expectedTopP:        0,
			expectedMaxTokens:   150, // Should be converted to max_tokens
			expectedFreqPenalty: 0,
			expectedPresPenalty: 0,
			mockResponse:        &llm.ProviderResult{Content: "Response with Gemini-style parameters"},
			mockError:           nil,
		},
		{
			name:          "Type conversion test",
			adapterParams: nil,
			requestParams: map[string]interface{}{
				"temperature":       float32(0.6),
				"top_p":             int(1),
				"max_tokens":        float64(300),
				"frequency_penalty": int32(2),
				"presence_penalty":  int64(1),
			},
			expectedTemperature: 0.6,
			expectedTopP:        1.0,
			expectedMaxTokens:   300,
			expectedFreqPenalty: 2.0,
			expectedPresPenalty: 1.0,
			mockResponse:        &llm.ProviderResult{Content: "Response with type-converted parameters"},
			mockError:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client with parameter checking
			mockClient := &MockLLMClientWithParamChecking{}
			mockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
				return tt.mockResponse, tt.mockError
			}

			// Create adapter
			adapter := NewOpenAIClientAdapter(mockClient)

			// Set adapter parameters if provided
			if tt.adapterParams != nil {
				adapter.SetParameters(tt.adapterParams)
			}

			// Call GenerateContent
			prompt := "Test prompt"
			result, err := adapter.GenerateContent(context.Background(), prompt, tt.requestParams)

			// Verify error behavior
			if tt.mockError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.mockError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse, result)
			}

			// Verify GenerateContent was called on the underlying client
			assert.True(t, mockClient.GenerateContentCalled, "GenerateContent should be called on the underlying client")

			// Verify parameters were properly set on the client
			if tt.expectedTemperature > 0 {
				assert.True(t, mockClient.TemperatureSet, "Temperature should be set on the client")
				assert.Equal(t, tt.expectedTemperature, mockClient.TemperatureValue)
			}

			if tt.expectedTopP > 0 {
				assert.True(t, mockClient.TopPSet, "TopP should be set on the client")
				assert.Equal(t, tt.expectedTopP, mockClient.TopPValue)
			}

			if tt.expectedMaxTokens > 0 {
				assert.True(t, mockClient.MaxTokensSet, "MaxTokens should be set on the client")
				assert.Equal(t, tt.expectedMaxTokens, mockClient.MaxTokensValue)
			}

			if tt.expectedFreqPenalty > 0 {
				assert.True(t, mockClient.FreqPenaltySet, "FrequencyPenalty should be set on the client")
				assert.Equal(t, tt.expectedFreqPenalty, mockClient.FreqPenaltyValue)
			}

			if tt.expectedPresPenalty > 0 {
				assert.True(t, mockClient.PresencePenaltySet, "PresencePenalty should be set on the client")
				assert.Equal(t, tt.expectedPresPenalty, mockClient.PresencePenaltyValue)
			}

			// Verify params were passed to the underlying client
			if tt.requestParams != nil {
				// The adapter should replace its params with the request params
				assert.Equal(t, tt.requestParams, mockClient.GenerateParams)
			} else if tt.adapterParams != nil {
				// The adapter should use its pre-set params
				assert.Equal(t, tt.adapterParams, mockClient.GenerateParams)
			}
		})
	}
}

// TestDelegationMethods verifies that the adapter correctly delegates
// all LLMClient interface methods to the underlying client
func TestDelegationMethods(t *testing.T) {
	// Create mock responses for each method
	mockTokenCount := &llm.ProviderTokenCount{Total: 42}
	mockModelInfo := &llm.ProviderModelInfo{
		Name:             "test-model",
		InputTokenLimit:  10000,
		OutputTokenLimit: 2000,
	}
	mockModelName := "custom-gpt-model"

	// Create error responses for each method
	tokenCountError := errors.New("token count error")
	modelInfoError := errors.New("model info error")
	closeError := errors.New("close error")

	// Test successful delegation
	t.Run("Successful delegation", func(t *testing.T) {
		// Create mock client with predefined responses
		mockClient := &MockLLMClientWithParamChecking{}
		mockClient.CountTokensFunc = func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
			return mockTokenCount, nil
		}
		mockClient.GetModelInfoFunc = func(ctx context.Context) (*llm.ProviderModelInfo, error) {
			return mockModelInfo, nil
		}
		mockClient.GetModelNameFunc = func() string {
			return mockModelName
		}
		mockClient.CloseFunc = func() error {
			return nil
		}

		// Create adapter
		adapter := NewOpenAIClientAdapter(mockClient)

		// Test CountTokens delegation
		tokenCount, err := adapter.CountTokens(context.Background(), "test prompt")
		assert.NoError(t, err)
		assert.Equal(t, mockTokenCount, tokenCount)
		assert.True(t, mockClient.CountTokensCalled)

		// Test GetModelInfo delegation
		modelInfo, err := adapter.GetModelInfo(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, mockModelInfo, modelInfo)
		assert.True(t, mockClient.GetModelInfoCalled)

		// Test GetModelName delegation
		modelName := adapter.GetModelName()
		assert.Equal(t, mockModelName, modelName)
		assert.True(t, mockClient.GetModelNameCalled)

		// Test Close delegation
		err = adapter.Close()
		assert.NoError(t, err)
		assert.True(t, mockClient.CloseCalled)
	})

	// Test error delegation
	t.Run("Error delegation", func(t *testing.T) {
		// Create mock client with error responses
		mockClient := &MockLLMClientWithParamChecking{}
		mockClient.CountTokensFunc = func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
			return nil, tokenCountError
		}
		mockClient.GetModelInfoFunc = func(ctx context.Context) (*llm.ProviderModelInfo, error) {
			return nil, modelInfoError
		}
		mockClient.CloseFunc = func() error {
			return closeError
		}

		// Create adapter
		adapter := NewOpenAIClientAdapter(mockClient)

		// Test CountTokens error delegation
		tokenCount, err := adapter.CountTokens(context.Background(), "test prompt")
		assert.Error(t, err)
		assert.Equal(t, tokenCountError, err)
		assert.Nil(t, tokenCount)

		// Test GetModelInfo error delegation
		modelInfo, err := adapter.GetModelInfo(context.Background())
		assert.Error(t, err)
		assert.Equal(t, modelInfoError, err)
		assert.Nil(t, modelInfo)

		// Test Close error delegation
		err = adapter.Close()
		assert.Error(t, err)
		assert.Equal(t, closeError, err)
	})
}

// TestGetModelInfoTokenLimitOverrides verifies the adapter's token limit handling logic
func TestGetModelInfoTokenLimitOverrides(t *testing.T) {
	testCases := []struct {
		name           string
		modelName      string
		clientResponse *llm.ProviderModelInfo
		clientError    error
		expectError    bool
		expectedLimits *llm.ProviderModelInfo
	}{
		{
			name:      "Client response is used when valid",
			modelName: "gpt-4",
			clientResponse: &llm.ProviderModelInfo{
				Name:             "gpt-4",
				InputTokenLimit:  8192,
				OutputTokenLimit: 2048,
			},
			clientError: nil,
			expectError: false,
			expectedLimits: &llm.ProviderModelInfo{
				Name:             "gpt-4",
				InputTokenLimit:  8192,
				OutputTokenLimit: 2048,
			},
		},
		{
			name:           "Error from client is propagated",
			modelName:      "gpt-4",
			clientResponse: nil,
			clientError:    errors.New("model info error"),
			expectError:    true,
			expectedLimits: nil,
		},
		{
			name:      "Zero limits are replaced with defaults for known models (gpt-4)",
			modelName: "gpt-4",
			clientResponse: &llm.ProviderModelInfo{
				Name:             "gpt-4",
				InputTokenLimit:  0, // Invalid limit
				OutputTokenLimit: 0, // Invalid limit
			},
			clientError: nil,
			expectError: false,
			expectedLimits: &llm.ProviderModelInfo{
				Name:             "gpt-4",
				InputTokenLimit:  8192, // Default for gpt-4
				OutputTokenLimit: 2048, // Default for gpt-4
			},
		},
		{
			name:      "Zero limits are replaced with defaults for known models (gpt-4-turbo)",
			modelName: "gpt-4-turbo",
			clientResponse: &llm.ProviderModelInfo{
				Name:             "gpt-4-turbo",
				InputTokenLimit:  0, // Invalid limit
				OutputTokenLimit: 0, // Invalid limit
			},
			clientError: nil,
			expectError: false,
			expectedLimits: &llm.ProviderModelInfo{
				Name:             "gpt-4-turbo",
				InputTokenLimit:  128000, // Default for gpt-4-turbo
				OutputTokenLimit: 4096,   // Default for gpt-4-turbo
			},
		},
		{
			name:      "Zero limits are replaced with defaults for known models (gpt-3.5-turbo)",
			modelName: "gpt-3.5-turbo",
			clientResponse: &llm.ProviderModelInfo{
				Name:             "gpt-3.5-turbo",
				InputTokenLimit:  0, // Invalid limit
				OutputTokenLimit: 0, // Invalid limit
			},
			clientError: nil,
			expectError: false,
			expectedLimits: &llm.ProviderModelInfo{
				Name:             "gpt-3.5-turbo",
				InputTokenLimit:  16385, // Default for gpt-3.5-turbo
				OutputTokenLimit: 4096,  // Default for gpt-3.5-turbo
			},
		},
		{
			name:      "Zero limits are replaced with defaults for unknown models",
			modelName: "unknown-model",
			clientResponse: &llm.ProviderModelInfo{
				Name:             "unknown-model",
				InputTokenLimit:  0, // Invalid limit
				OutputTokenLimit: 0, // Invalid limit
			},
			clientError: nil,
			expectError: false,
			expectedLimits: &llm.ProviderModelInfo{
				Name:             "unknown-model",
				InputTokenLimit:  4096, // Default for unknown models
				OutputTokenLimit: 2048, // Default for unknown models
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockLLMClientWithParamChecking{}
			mockClient.GetModelInfoFunc = func(ctx context.Context) (*llm.ProviderModelInfo, error) {
				return tc.clientResponse, tc.clientError
			}
			mockClient.GetModelNameFunc = func() string {
				return tc.modelName
			}

			// Create adapter
			adapter := NewOpenAIClientAdapter(mockClient)

			// Call GetModelInfo
			result, err := adapter.GetModelInfo(context.Background())

			// Verify error behavior
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify limits
			assert.Equal(t, tc.expectedLimits.Name, result.Name)
			assert.Equal(t, tc.expectedLimits.InputTokenLimit, result.InputTokenLimit)
			assert.Equal(t, tc.expectedLimits.OutputTokenLimit, result.OutputTokenLimit)
		})
	}
}
