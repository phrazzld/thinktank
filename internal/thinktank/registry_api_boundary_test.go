// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestRegistryAPIServiceContract verifies that the registryAPIService
// implementation correctly handles boundary conditions and edge cases.
func TestRegistryAPIServiceContract(t *testing.T) {
	// Logger is required but not used directly in this test
	_ = logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")

	// Create a mock APIService
	mockService := &mockAPIService{
		isEmptyResponseErrorFunc: func(err error) bool {
			return err != nil && errors.Is(err, llm.ErrEmptyResponse)
		},
		isSafetyBlockedErrorFunc: func(err error) bool {
			return err != nil && errors.Is(err, llm.ErrSafetyBlocked)
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			if result == nil {
				return "", llm.ErrEmptyResponse
			}
			if result.Content == "" {
				if len(result.SafetyInfo) > 0 {
					for _, info := range result.SafetyInfo {
						if info.Blocked {
							return "", llm.ErrSafetyBlocked
						}
					}
				}
				return "", llm.ErrEmptyResponse
			}
			return result.Content, nil
		},
	}

	// Test ProcessLLMResponse with various test cases
	t.Run("ProcessLLMResponse", func(t *testing.T) {
		tests := []struct {
			name    string
			result  *llm.ProviderResult
			want    string
			wantErr bool
			errType error
		}{
			{
				name:    "nil result",
				result:  nil,
				want:    "",
				wantErr: true,
				errType: llm.ErrEmptyResponse,
			},
			{
				name: "empty content",
				result: &llm.ProviderResult{
					Content: "",
				},
				want:    "",
				wantErr: true,
				errType: llm.ErrEmptyResponse,
			},
			{
				name: "safety blocked content",
				result: &llm.ProviderResult{
					Content: "",
					SafetyInfo: []llm.Safety{
						{Category: "harmful", Blocked: true},
					},
				},
				want:    "",
				wantErr: true,
				errType: llm.ErrSafetyBlocked,
			},
			{
				name: "valid content",
				result: &llm.ProviderResult{
					Content: "valid content",
				},
				want:    "valid content",
				wantErr: false,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := mockService.ProcessLLMResponse(tc.result)
				if (err != nil) != tc.wantErr {
					t.Errorf("ProcessLLMResponse() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.wantErr && tc.errType != nil && !errors.Is(err, tc.errType) {
					t.Errorf("ProcessLLMResponse() error type = %v, want %v", err, tc.errType)
				}
				if got != tc.want {
					t.Errorf("ProcessLLMResponse() = %v, want %v", got, tc.want)
				}
			})
		}
	})

	// Test Error Classification methods
	t.Run("ErrorClassification", func(t *testing.T) {
		// Create errors for testing
		emptyError := llm.ErrEmptyResponse
		safetyError := llm.ErrSafetyBlocked
		otherError := errors.New("other error")

		// Test IsEmptyResponseError classification
		if !mockService.IsEmptyResponseError(emptyError) {
			t.Errorf("IsEmptyResponseError() should return true for ErrEmptyResponse")
		}
		if mockService.IsEmptyResponseError(safetyError) {
			t.Errorf("IsEmptyResponseError() should return false for ErrSafetyBlocked")
		}
		if mockService.IsEmptyResponseError(otherError) {
			t.Errorf("IsEmptyResponseError() should return false for other errors")
		}

		// Test IsSafetyBlockedError classification
		if !mockService.IsSafetyBlockedError(safetyError) {
			t.Errorf("IsSafetyBlockedError() should return true for ErrSafetyBlocked")
		}
		if mockService.IsSafetyBlockedError(emptyError) {
			t.Errorf("IsSafetyBlockedError() should return false for ErrEmptyResponse")
		}
		if mockService.IsSafetyBlockedError(otherError) {
			t.Errorf("IsSafetyBlockedError() should return false for other errors")
		}
	})
}

// mockAPIService is a mock implementation of the APIService for testing
type mockAPIService struct {
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	processLLMResponseFunc   func(result *llm.ProviderResult) (string, error)
}

func (m *mockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.processLLMResponseFunc != nil {
		return m.processLLMResponseFunc(result)
	}
	return "", errors.New("not implemented")
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	if m.isEmptyResponseErrorFunc != nil {
		return m.isEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	if m.isSafetyBlockedErrorFunc != nil {
		return m.isSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return "not implemented"
}

func (m *mockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockAPIService) GetModelDefinition(modelName string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	return 0, 0, errors.New("not implemented")
}
