// Package architect contains the core application logic for the architect tool.
// This file contains shared mock definitions for adapter tests.
package architect

import (
	"context"
	"errors"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// MockAPIServiceForAdapter is a testing mock for the APIService interface, specifically for adapter tests
type MockAPIServiceForAdapter struct {
	InitClientFunc           func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error)
	ProcessResponseFunc      func(result *gemini.GenerationResult) (string, error)
	IsEmptyResponseErrorFunc func(err error) bool
	IsSafetyBlockedErrorFunc func(err error) bool
	GetErrorDetailsFunc      func(err error) string
}

func (m *MockAPIServiceForAdapter) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	if m.InitClientFunc != nil {
		return m.InitClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, errors.New("InitClient not implemented")
}

func (m *MockAPIServiceForAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	if m.ProcessResponseFunc != nil {
		return m.ProcessResponseFunc(result)
	}
	return "", errors.New("ProcessResponse not implemented")
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
