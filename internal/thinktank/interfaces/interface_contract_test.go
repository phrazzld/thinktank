// Package interfaces contains shared interfaces for the thinktank application
package interfaces

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

// TestAPIServiceContractVerification ensures that a correct implementation
// of the APIService interface meets all contractual expectations.
func TestAPIServiceContractVerification(t *testing.T) {
	// Create a mock implementation of the APIService interface
	mockAPI := &MockAPIService{
		// InitLLMClient implementation for testing
		InitLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			if modelName == "" {
				return nil, errors.New("model name required")
			}
			if apiKey == "" {
				return nil, errors.New("API key required")
			}
			return &MockLLMClient{
				modelName: modelName,
			}, nil
		},

		// ProcessLLMResponse implementation for testing
		ProcessLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			if result == nil {
				return "", errors.New("result is nil")
			}
			if result.Content == "" {
				return "", errors.New("empty content")
			}
			return result.Content, nil
		},

		// Error checking functions for testing
		IsEmptyResponseErrorFunc: func(err error) bool {
			return err != nil && err.Error() == "empty content"
		},

		IsSafetyBlockedErrorFunc: func(err error) bool {
			return err != nil && err.Error() == "safety blocked"
		},

		GetErrorDetailsFunc: func(err error) string {
			if err == nil {
				return "no error"
			}
			return err.Error()
		},

		// Model parameter functions for testing
		GetModelParametersFunc: func(modelName string) (map[string]interface{}, error) {
			if modelName == "" {
				return nil, errors.New("model name required")
			}
			return map[string]interface{}{
				"temperature": 0.7,
			}, nil
		},

		ValidateModelParameterFunc: func(modelName, paramName string, value interface{}) (bool, error) {
			if modelName == "" || paramName == "" {
				return false, errors.New("model and parameter names required")
			}
			// Simple validation for temperature parameter
			if paramName == "temperature" {
				temp, ok := value.(float64)
				if !ok {
					return false, errors.New("temperature must be a float")
				}
				if temp < 0.0 || temp > 1.0 {
					return false, errors.New("temperature must be between 0.0 and 1.0")
				}
				return true, nil
			}
			return false, errors.New("unknown parameter")
		},
	}

	// Verify InitLLMClient contract
	t.Run("InitLLMClient", func(t *testing.T) {
		// Test valid case
		client, err := mockAPI.InitLLMClient(context.Background(), "test-key", "test-model", "")
		if err != nil {
			t.Errorf("Expected no error for valid inputs, got: %v", err)
		}
		if client == nil {
			t.Errorf("Expected client, got nil")
		} else if name := client.GetModelName(); name != "test-model" {
			t.Errorf("Expected model name %q, got %q", "test-model", name)
		}

		// Test empty model name (invalid)
		client, err = mockAPI.InitLLMClient(context.Background(), "test-key", "", "")
		if err == nil {
			t.Errorf("Expected error for empty model name, got nil")
		}
		if client != nil {
			t.Errorf("Expected nil client for error case, got: %v", client)
		}

		// Test empty API key (invalid)
		client, err = mockAPI.InitLLMClient(context.Background(), "", "test-model", "")
		if err == nil {
			t.Errorf("Expected error for empty API key, got nil")
		}
		if client != nil {
			t.Errorf("Expected nil client for error case, got: %v", client)
		}
	})

	// Verify ProcessLLMResponse contract
	t.Run("ProcessLLMResponse", func(t *testing.T) {
		// Test valid case
		result := &llm.ProviderResult{
			Content: "test content",
		}
		content, err := mockAPI.ProcessLLMResponse(result)
		if err != nil {
			t.Errorf("Expected no error for valid result, got: %v", err)
		}
		if content != "test content" {
			t.Errorf("Expected content %q, got %q", "test content", content)
		}

		// Test nil result (invalid)
		content, err = mockAPI.ProcessLLMResponse(nil)
		if err == nil {
			t.Errorf("Expected error for nil result, got nil")
		}
		if content != "" {
			t.Errorf("Expected empty content for error case, got: %q", content)
		}

		// Test empty content (invalid)
		result = &llm.ProviderResult{
			Content: "",
		}
		content, err = mockAPI.ProcessLLMResponse(result)
		if err == nil {
			t.Errorf("Expected error for empty content, got nil")
		}
		if content != "" {
			t.Errorf("Expected empty content for error case, got: %q", content)
		}
	})

	// Verify error classification contracts
	t.Run("ErrorClassification", func(t *testing.T) {
		emptyErr := errors.New("empty content")
		safetyErr := errors.New("safety blocked")
		otherErr := errors.New("other error")

		// Test IsEmptyResponseError
		if !mockAPI.IsEmptyResponseError(emptyErr) {
			t.Errorf("Expected IsEmptyResponseError to return true for %q", emptyErr)
		}
		if mockAPI.IsEmptyResponseError(safetyErr) {
			t.Errorf("Expected IsEmptyResponseError to return false for %q", safetyErr)
		}
		if mockAPI.IsEmptyResponseError(otherErr) {
			t.Errorf("Expected IsEmptyResponseError to return false for %q", otherErr)
		}

		// Test IsSafetyBlockedError
		if !mockAPI.IsSafetyBlockedError(safetyErr) {
			t.Errorf("Expected IsSafetyBlockedError to return true for %q", safetyErr)
		}
		if mockAPI.IsSafetyBlockedError(emptyErr) {
			t.Errorf("Expected IsSafetyBlockedError to return false for %q", emptyErr)
		}
		if mockAPI.IsSafetyBlockedError(otherErr) {
			t.Errorf("Expected IsSafetyBlockedError to return false for %q", otherErr)
		}

		// Test GetErrorDetails
		if details := mockAPI.GetErrorDetails(emptyErr); details != "empty content" {
			t.Errorf("Expected GetErrorDetails to return %q, got %q", "empty content", details)
		}
		if details := mockAPI.GetErrorDetails(nil); details != "no error" {
			t.Errorf("Expected GetErrorDetails to return %q for nil error, got %q", "no error", details)
		}
	})

	// Verify parameter-related contracts
	t.Run("ModelParameters", func(t *testing.T) {
		// Test GetModelParameters
		params, err := mockAPI.GetModelParameters("test-model")
		if err != nil {
			t.Errorf("Expected no error for valid model, got: %v", err)
		}
		if temp, ok := params["temperature"]; !ok || temp != 0.7 {
			t.Errorf("Expected temperature=0.7, got %v", temp)
		}

		// Test GetModelParameters with empty model name
		_, err = mockAPI.GetModelParameters("")
		if err == nil {
			t.Errorf("Expected error for empty model name, got nil")
		}

		// Test ValidateModelParameter
		valid, err := mockAPI.ValidateModelParameter("test-model", "temperature", 0.5)
		if err != nil {
			t.Errorf("Expected no error for valid parameter, got: %v", err)
		}
		if !valid {
			t.Errorf("Expected parameter to be valid")
		}

		// Test ValidateModelParameter with invalid value
		valid, err = mockAPI.ValidateModelParameter("test-model", "temperature", 1.5)
		if err == nil {
			t.Errorf("Expected error for invalid parameter value, got nil")
		}
		if valid {
			t.Errorf("Expected parameter to be invalid")
		}

		// Test ValidateModelParameter with invalid type
		valid, err = mockAPI.ValidateModelParameter("test-model", "temperature", "not a float")
		if err == nil {
			t.Errorf("Expected error for invalid parameter type, got nil")
		}
		if valid {
			t.Errorf("Expected parameter to be invalid")
		}
	})
}

// MockAPIService is a mock implementation of the APIService interface for testing
type MockAPIService struct {
	InitLLMClientFunc          func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	ProcessLLMResponseFunc     func(result *llm.ProviderResult) (string, error)
	IsEmptyResponseErrorFunc   func(err error) bool
	IsSafetyBlockedErrorFunc   func(err error) bool
	GetErrorDetailsFunc        func(err error) string
	GetModelParametersFunc     func(modelName string) (map[string]interface{}, error)
	ValidateModelParameterFunc func(modelName, paramName string, value interface{}) (bool, error)
	GetModelDefinitionFunc     func(modelName string) (interface{}, error)
	GetModelTokenLimitsFunc    func(modelName string) (int32, int32, error)
}

// Implement APIService interface

func (m *MockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.InitLLMClientFunc != nil {
		return m.InitLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.ProcessLLMResponseFunc != nil {
		return m.ProcessLLMResponseFunc(result)
	}
	return "", errors.New("not implemented")
}

func (m *MockAPIService) IsEmptyResponseError(err error) bool {
	if m.IsEmptyResponseErrorFunc != nil {
		return m.IsEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *MockAPIService) IsSafetyBlockedError(err error) bool {
	if m.IsSafetyBlockedErrorFunc != nil {
		return m.IsSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *MockAPIService) GetErrorDetails(err error) string {
	if m.GetErrorDetailsFunc != nil {
		return m.GetErrorDetailsFunc(err)
	}
	return "not implemented"
}

func (m *MockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	if m.GetModelParametersFunc != nil {
		return m.GetModelParametersFunc(modelName)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	if m.ValidateModelParameterFunc != nil {
		return m.ValidateModelParameterFunc(modelName, paramName, value)
	}
	return false, errors.New("not implemented")
}

func (m *MockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.GetModelTokenLimitsFunc != nil {
		return m.GetModelTokenLimitsFunc(modelName)
	}
	return 0, 0, errors.New("not implemented")
}

// MockLLMClient is a mock implementation of the llm.LLMClient interface for testing
type MockLLMClient struct {
	modelName           string
	generateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	closeFunc           func() error
}

func (c *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if c.generateContentFunc != nil {
		return c.generateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{
		Content: "mock response",
	}, nil
}

func (c *MockLLMClient) GetModelName() string {
	if c.modelName != "" {
		return c.modelName
	}
	return "mock-model"
}

func (c *MockLLMClient) Close() error {
	if c.closeFunc != nil {
		return c.closeFunc()
	}
	return nil
}
