// Package thinktank contains the core application logic for the thinktank tool.
// This file contains shared mock definitions and tests for adapter functionality.
package thinktank

import (
	"context"
	"errors"

	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// MockAPIServiceForAdapter is a testing mock for the APIService interface, specifically for adapter tests
type MockAPIServiceForAdapter struct {
	InitLLMClientFunc          func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	ProcessLLMResponseFunc     func(result *llm.ProviderResult) (string, error)
	IsEmptyResponseErrorFunc   func(err error) bool
	IsSafetyBlockedErrorFunc   func(err error) bool
	GetErrorDetailsFunc        func(err error) string
	GetModelParametersFunc     func(ctx context.Context, modelName string) (map[string]interface{}, error)
	ValidateModelParameterFunc func(ctx context.Context, modelName, paramName string, value interface{}) (bool, error)
	GetModelDefinitionFunc     func(ctx context.Context, modelName string) (*models.ModelInfo, error)
	GetModelTokenLimitsFunc    func(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error)

	// Call tracking fields
	InitLLMClientCalls          []InitLLMClientCall
	ProcessLLMResponseCalls     []ProcessLLMResponseCall
	GetErrorDetailsCalls        []GetErrorDetailsCall
	IsEmptyResponseErrorCalls   []IsEmptyResponseErrorCall
	IsSafetyBlockedErrorCalls   []IsSafetyBlockedErrorCall
	GetModelParametersCalls     []GetModelParametersCall
	ValidateModelParameterCalls []ValidateModelParameterCall
	GetModelDefinitionCalls     []GetModelDefinitionCall
	GetModelTokenLimitsCalls    []GetModelTokenLimitsCall
}

// Call record structs
type InitLLMClientCall struct {
	Ctx         context.Context
	APIKey      string
	ModelName   string
	APIEndpoint string
}

type ProcessLLMResponseCall struct {
	Result *llm.ProviderResult
}

type GetErrorDetailsCall struct {
	Err error
}

type IsEmptyResponseErrorCall struct {
	Err error
}

type IsSafetyBlockedErrorCall struct {
	Err error
}

type GetModelParametersCall struct {
	Ctx       context.Context
	ModelName string
}

type ValidateModelParameterCall struct {
	Ctx       context.Context
	ModelName string
	ParamName string
	Value     interface{}
}

type GetModelDefinitionCall struct {
	Ctx       context.Context
	ModelName string
}

type GetModelTokenLimitsCall struct {
	Ctx       context.Context
	ModelName string
}

func (m *MockAPIServiceForAdapter) IsEmptyResponseError(err error) bool {
	m.IsEmptyResponseErrorCalls = append(m.IsEmptyResponseErrorCalls, IsEmptyResponseErrorCall{
		Err: err,
	})
	if m.IsEmptyResponseErrorFunc != nil {
		return m.IsEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *MockAPIServiceForAdapter) IsSafetyBlockedError(err error) bool {
	m.IsSafetyBlockedErrorCalls = append(m.IsSafetyBlockedErrorCalls, IsSafetyBlockedErrorCall{
		Err: err,
	})
	if m.IsSafetyBlockedErrorFunc != nil {
		return m.IsSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *MockAPIServiceForAdapter) GetErrorDetails(err error) string {
	m.GetErrorDetailsCalls = append(m.GetErrorDetailsCalls, GetErrorDetailsCall{
		Err: err,
	})
	if m.GetErrorDetailsFunc != nil {
		return m.GetErrorDetailsFunc(err)
	}
	return "Error details not implemented"
}

func (m *MockAPIServiceForAdapter) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	m.InitLLMClientCalls = append(m.InitLLMClientCalls, InitLLMClientCall{
		Ctx:         ctx,
		APIKey:      apiKey,
		ModelName:   modelName,
		APIEndpoint: apiEndpoint,
	})
	if m.InitLLMClientFunc != nil {
		return m.InitLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, errors.New("InitLLMClient not implemented")
}

func (m *MockAPIServiceForAdapter) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	m.ProcessLLMResponseCalls = append(m.ProcessLLMResponseCalls, ProcessLLMResponseCall{
		Result: result,
	})
	if m.ProcessLLMResponseFunc != nil {
		return m.ProcessLLMResponseFunc(result)
	}
	return "", errors.New("ProcessLLMResponse not implemented")
}

func (m *MockAPIServiceForAdapter) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	m.GetModelParametersCalls = append(m.GetModelParametersCalls, GetModelParametersCall{
		Ctx:       ctx,
		ModelName: modelName,
	})
	if m.GetModelParametersFunc != nil {
		return m.GetModelParametersFunc(ctx, modelName)
	}
	return make(map[string]interface{}), nil
}

func (m *MockAPIServiceForAdapter) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	m.ValidateModelParameterCalls = append(m.ValidateModelParameterCalls, ValidateModelParameterCall{
		Ctx:       ctx,
		ModelName: modelName,
		ParamName: paramName,
		Value:     value,
	})
	if m.ValidateModelParameterFunc != nil {
		return m.ValidateModelParameterFunc(ctx, modelName, paramName, value)
	}
	return true, nil
}

func (m *MockAPIServiceForAdapter) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	m.GetModelDefinitionCalls = append(m.GetModelDefinitionCalls, GetModelDefinitionCall{
		Ctx:       ctx,
		ModelName: modelName,
	})
	if m.GetModelDefinitionFunc != nil {
		return m.GetModelDefinitionFunc(ctx, modelName)
	}
	return nil, errors.New("GetModelDefinition not implemented")
}

func (m *MockAPIServiceForAdapter) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	m.GetModelTokenLimitsCalls = append(m.GetModelTokenLimitsCalls, GetModelTokenLimitsCall{
		Ctx:       ctx,
		ModelName: modelName,
	})
	if m.GetModelTokenLimitsFunc != nil {
		return m.GetModelTokenLimitsFunc(ctx, modelName)
	}
	return 0, 0, errors.New("GetModelTokenLimits not implemented")
}

// TestTokenResult represents a token count result structure for testing only
// This replaces the removed production TokenResult
type TestTokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// MockTokenManagerForAdapter is a testing mock for the TokenManager interface, specifically for adapter tests
type MockTokenManagerForAdapter struct {
	CheckTokenLimitFunc       func(ctx context.Context, prompt string) error
	GetTokenInfoFunc          func(ctx context.Context, prompt string) (*TestTokenResult, error)
	PromptForConfirmationFunc func(tokenCount int32, threshold int) bool
}

func (m *MockTokenManagerForAdapter) CheckTokenLimit(ctx context.Context, prompt string) error {
	if m.CheckTokenLimitFunc != nil {
		return m.CheckTokenLimitFunc(ctx, prompt)
	}
	return errors.New("CheckTokenLimit not implemented")
}

func (m *MockTokenManagerForAdapter) GetTokenInfo(ctx context.Context, prompt string) (*TestTokenResult, error) {
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

// MockAPIServiceWithoutExtensions implements the full interfaces.APIService interface
// but its extended methods are meant to be skipped by the adapter's interface assertions
// This is important for testing the fallback logic in adapters
type MockAPIServiceWithoutExtensions struct {
	// Function fields
	InitLLMClientFunc        func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	ProcessLLMResponseFunc   func(result *llm.ProviderResult) (string, error)
	GetErrorDetailsFunc      func(err error) string
	IsEmptyResponseErrorFunc func(err error) bool
	IsSafetyBlockedErrorFunc func(err error) bool
}

func (m *MockAPIServiceWithoutExtensions) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return m.InitLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
}

func (m *MockAPIServiceWithoutExtensions) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return m.ProcessLLMResponseFunc(result)
}

func (m *MockAPIServiceWithoutExtensions) GetErrorDetails(err error) string {
	return m.GetErrorDetailsFunc(err)
}

func (m *MockAPIServiceWithoutExtensions) IsEmptyResponseError(err error) bool {
	return m.IsEmptyResponseErrorFunc(err)
}

func (m *MockAPIServiceWithoutExtensions) IsSafetyBlockedError(err error) bool {
	return m.IsSafetyBlockedErrorFunc(err)
}

// The following methods implement the extended interfaces.APIService interface

// GetModelParameters implements interfaces.APIService.GetModelParameters
func (m *MockAPIServiceWithoutExtensions) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	// Return default values - this method should not be called by adapter tests
	return make(map[string]interface{}), nil
}

// ValidateModelParameter implements interfaces.APIService.ValidateModelParameter
func (m *MockAPIServiceWithoutExtensions) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	// Return default values - this method should not be called by adapter tests
	return true, nil
}

// GetModelDefinition implements interfaces.APIService.GetModelDefinition
func (m *MockAPIServiceWithoutExtensions) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	// Return default values - this method should not be called by adapter tests
	return nil, errors.New("model definition not available")
}

// GetModelTokenLimits returns fallback values that won't actually be used by the adapter
func (m *MockAPIServiceWithoutExtensions) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Return default values
	return 0, 0, nil
}

// NewMockAPIServiceWithoutExtensions creates a new MockAPIServiceWithoutExtensions with default implementations
func NewMockAPIServiceWithoutExtensions() *MockAPIServiceWithoutExtensions {
	return &MockAPIServiceWithoutExtensions{
		InitLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return nil, nil
		},
		ProcessLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return "", nil
		},
		GetErrorDetailsFunc: func(err error) string {
			return ""
		},
		IsEmptyResponseErrorFunc: func(err error) bool {
			return false
		},
		IsSafetyBlockedErrorFunc: func(err error) bool {
			return false
		},
	}
}

// MockContextGatherer implements interfaces.ContextGatherer for testing
type MockContextGatherer struct {
	// Function fields
	GatherContextFunc     func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error)
	DisplayDryRunInfoFunc func(ctx context.Context, stats *interfaces.ContextStats) error

	// Call tracking fields
	GatherContextCalls     []GatherContextCall
	DisplayDryRunInfoCalls []DisplayDryRunInfoCall
}

// Call record structs
type GatherContextCall struct {
	Ctx    context.Context
	Config interfaces.GatherConfig
}

type DisplayDryRunInfoCall struct {
	Ctx   context.Context
	Stats *interfaces.ContextStats
}

// MockContextGatherer method implementations
func (m *MockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	m.GatherContextCalls = append(m.GatherContextCalls, GatherContextCall{
		Ctx:    ctx,
		Config: config,
	})
	return m.GatherContextFunc(ctx, config)
}

func (m *MockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	m.DisplayDryRunInfoCalls = append(m.DisplayDryRunInfoCalls, DisplayDryRunInfoCall{
		Ctx:   ctx,
		Stats: stats,
	})
	return m.DisplayDryRunInfoFunc(ctx, stats)
}

// NewMockContextGatherer creates a new MockContextGatherer with default implementations
func NewMockContextGatherer() *MockContextGatherer {
	return &MockContextGatherer{
		GatherContextFunc: func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
			return nil, nil, nil
		},
		DisplayDryRunInfoFunc: func(ctx context.Context, stats *interfaces.ContextStats) error {
			return nil
		},
	}
}

// MockFileWriter implements FileWriter for testing
type MockFileWriter struct {
	// Function fields
	SaveToFileFunc func(ctx context.Context, content, outputFile string) error

	// Call tracking fields
	SaveToFileCalls []SaveToFileCall
}

// Call record structs
type SaveToFileCall struct {
	Ctx        context.Context
	Content    string
	OutputFile string
}

// MockFileWriter method implementations
func (m *MockFileWriter) SaveToFile(ctx context.Context, content, outputFile string) error {
	m.SaveToFileCalls = append(m.SaveToFileCalls, SaveToFileCall{
		Ctx:        ctx,
		Content:    content,
		OutputFile: outputFile,
	})
	return m.SaveToFileFunc(ctx, content, outputFile)
}

// NewMockFileWriter creates a new MockFileWriter with default implementations
func NewMockFileWriter() *MockFileWriter {
	return &MockFileWriter{
		SaveToFileFunc: func(ctx context.Context, content, outputFile string) error {
			return nil
		},
	}
}
