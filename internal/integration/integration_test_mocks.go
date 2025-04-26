// internal/integration/integration_test_mocks.go
package integration

import (
	"context"
	"fmt"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// MockAPIService is a mock implementation of the APIService interface
type MockAPIService struct {
	InitLLMClientFunc          func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	GetModelParametersFunc     func(modelName string) (map[string]interface{}, error)
	ValidateModelParameterFunc func(modelName, paramName string, value interface{}) (bool, error)
	GetModelDefinitionFunc     func(modelName string) (*registry.ModelDefinition, error)
	GetModelTokenLimitsFunc    func(modelName string) (contextWindow, maxOutputTokens int32, err error)
	ProcessLLMResponseFunc     func(result *llm.ProviderResult) (string, error)
	IsEmptyResponseErrorFunc   func(err error) bool
	IsSafetyBlockedErrorFunc   func(err error) bool
	GetErrorDetailsFunc        func(err error) string
}

func (m *MockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.InitLLMClientFunc != nil {
		return m.InitLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, nil
}

func (m *MockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	if m.GetModelParametersFunc != nil {
		return m.GetModelParametersFunc(modelName)
	}
	return map[string]interface{}{}, nil
}

func (m *MockAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	if m.ValidateModelParameterFunc != nil {
		return m.ValidateModelParameterFunc(modelName, paramName, value)
	}
	return true, nil
}

func (m *MockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	if m.GetModelDefinitionFunc != nil {
		return m.GetModelDefinitionFunc(modelName)
	}
	return nil, nil
}

func (m *MockAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.GetModelTokenLimitsFunc != nil {
		return m.GetModelTokenLimitsFunc(modelName)
	}
	return 0, 0, nil
}

func (m *MockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.ProcessLLMResponseFunc != nil {
		return m.ProcessLLMResponseFunc(result)
	}
	return "", nil
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
	return ""
}

// MockContextGatherer is a mock implementation of the context gatherer
type MockContextGatherer struct {
	GatherContextFunc     func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error)
	DisplayDryRunInfoFunc func(ctx context.Context, stats *interfaces.ContextStats) error
}

// GatherContext implements the context gatherer interface
func (m *MockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return m.GatherContextFunc(ctx, config)
}

// DisplayDryRunInfo implements the context gatherer interface
func (m *MockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return m.DisplayDryRunInfoFunc(ctx, stats)
}

// MockFileWriter is a mock implementation of the file writer
type MockFileWriter struct {
	SaveToFileFunc func(content, filePath string) error
}

// SaveToFile implements the file writer interface
func (m *MockFileWriter) SaveToFile(content, filePath string) error {
	return m.SaveToFileFunc(content, filePath)
}

// MockAuditLogger is a mock implementation of the audit logger
type MockAuditLogger struct {
	LogFunc   func(entry auditlog.AuditEntry) error
	LogOpFunc func(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error
	CloseFunc func() error
}

// Log implements the audit logger interface
func (m *MockAuditLogger) Log(entry auditlog.AuditEntry) error {
	if m.LogFunc != nil {
		return m.LogFunc(entry)
	}
	return nil
}

// LogOp implements the audit logger interface
func (m *MockAuditLogger) LogOp(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	if m.LogOpFunc != nil {
		return m.LogOpFunc(operation, status, inputs, outputs, err)
	}
	// Create an AuditEntry from the parameters and pass to Log if LogFunc is defined
	if m.LogFunc != nil {
		entry := auditlog.AuditEntry{
			Operation: operation,
			Status:    status,
			Inputs:    inputs,
			Outputs:   outputs,
			Message:   fmt.Sprintf("%s - %s", operation, status),
		}
		if err != nil {
			entry.Error = &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "GeneralError",
			}
		}
		return m.LogFunc(entry)
	}
	return nil
}

// Close implements the audit logger interface
func (m *MockAuditLogger) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
