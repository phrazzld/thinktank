// Package architect contains the core application logic for the architect tool.
// This file contains shared mock definitions for adapter tests.
package architect

import (
	"context"
	"errors"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
)

// MockAPIServiceForAdapter is a testing mock for the APIService interface, specifically for adapter tests
type MockAPIServiceForAdapter struct {
	InitLLMClientFunc          func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	ProcessLLMResponseFunc     func(result *llm.ProviderResult) (string, error)
	IsEmptyResponseErrorFunc   func(err error) bool
	IsSafetyBlockedErrorFunc   func(err error) bool
	GetErrorDetailsFunc        func(err error) string
	GetModelParametersFunc     func(modelName string) (map[string]interface{}, error)
	ValidateModelParameterFunc func(modelName, paramName string, value interface{}) (bool, error)
	GetModelDefinitionFunc     func(modelName string) (*registry.ModelDefinition, error)
	GetModelTokenLimitsFunc    func(modelName string) (contextWindow, maxOutputTokens int32, err error)
}

func (m *MockAPIServiceForAdapter) IsEmptyResponseError(err error) bool {
	if m.IsEmptyResponseErrorFunc != nil {
		return m.IsEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *MockAPIServiceForAdapter) IsSafetyBlockedError(err error) bool {
	if m.IsSafetyBlockedErrorFunc != nil {
		return m.IsSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *MockAPIServiceForAdapter) GetErrorDetails(err error) string {
	if m.GetErrorDetailsFunc != nil {
		return m.GetErrorDetailsFunc(err)
	}
	return "Error details not implemented"
}

func (m *MockAPIServiceForAdapter) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.InitLLMClientFunc != nil {
		return m.InitLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, errors.New("InitLLMClient not implemented")
}

func (m *MockAPIServiceForAdapter) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.ProcessLLMResponseFunc != nil {
		return m.ProcessLLMResponseFunc(result)
	}
	return "", errors.New("ProcessLLMResponse not implemented")
}

func (m *MockAPIServiceForAdapter) GetModelParameters(modelName string) (map[string]interface{}, error) {
	if m.GetModelParametersFunc != nil {
		return m.GetModelParametersFunc(modelName)
	}
	return make(map[string]interface{}), nil
}

func (m *MockAPIServiceForAdapter) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	if m.ValidateModelParameterFunc != nil {
		return m.ValidateModelParameterFunc(modelName, paramName, value)
	}
	return true, nil
}

func (m *MockAPIServiceForAdapter) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	if m.GetModelDefinitionFunc != nil {
		return m.GetModelDefinitionFunc(modelName)
	}
	return nil, errors.New("GetModelDefinition not implemented")
}

func (m *MockAPIServiceForAdapter) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.GetModelTokenLimitsFunc != nil {
		return m.GetModelTokenLimitsFunc(modelName)
	}
	return 0, 0, errors.New("GetModelTokenLimits not implemented")
}

// MockTokenManagerForAdapter is a testing mock for the TokenManager interface, specifically for adapter tests
type MockTokenManagerForAdapter struct {
	CheckTokenLimitFunc       func(ctx context.Context, prompt string) error
	GetTokenInfoFunc          func(ctx context.Context, prompt string) (*TokenResult, error)
	PromptForConfirmationFunc func(tokenCount int32, threshold int) bool
}

func (m *MockTokenManagerForAdapter) CheckTokenLimit(ctx context.Context, prompt string) error {
	if m.CheckTokenLimitFunc != nil {
		return m.CheckTokenLimitFunc(ctx, prompt)
	}
	return errors.New("CheckTokenLimit not implemented")
}

func (m *MockTokenManagerForAdapter) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	if m.GetTokenInfoFunc != nil {
		return m.GetTokenInfoFunc(ctx, prompt)
	}
	return nil, errors.New("GetTokenInfo not implemented")
}

func (m *MockTokenManagerForAdapter) PromptForConfirmation(tokenCount int32, threshold int) bool {
	if m.PromptForConfirmationFunc != nil {
		return m.PromptForConfirmationFunc(tokenCount, threshold)
	}
	return false
}

// GetAdapterTestLogger returns a logger for adapter tests
func GetAdapterTestLogger() logutil.LoggerInterface {
	return logutil.NewLogger(logutil.DebugLevel, nil, "[adapter-test] ")
}
