// internal/integration/integration_test_mocks.go
package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// Deprecated: MockAPIService directly mocks an internal implementation
// This will be removed in favor of boundary-based mocks
// Use BoundaryAPIService from boundary_test_adapter.go instead
type MockAPIService struct {
	InitLLMClientFunc          func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	GetModelParametersFunc     func(ctx context.Context, modelName string) (map[string]interface{}, error)
	ValidateModelParameterFunc func(ctx context.Context, modelName, paramName string, value interface{}) (bool, error)
	GetModelDefinitionFunc     func(ctx context.Context, modelName string) (*registry.ModelDefinition, error)
	GetModelTokenLimitsFunc    func(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error)
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

func (m *MockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	if m.GetModelParametersFunc != nil {
		return m.GetModelParametersFunc(ctx, modelName)
	}
	return map[string]interface{}{}, nil
}

func (m *MockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	if m.ValidateModelParameterFunc != nil {
		return m.ValidateModelParameterFunc(ctx, modelName, paramName, value)
	}
	return true, nil
}

func (m *MockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*registry.ModelDefinition, error) {
	if m.GetModelDefinitionFunc != nil {
		return m.GetModelDefinitionFunc(ctx, modelName)
	}
	return nil, nil
}

func (m *MockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.GetModelTokenLimitsFunc != nil {
		return m.GetModelTokenLimitsFunc(ctx, modelName)
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

// Deprecated: MockContextGatherer directly mocks an internal implementation
// This will be removed in favor of boundary-based mocks
// Use BoundaryContextGatherer from boundary_test_adapter.go instead
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

// Deprecated: MockFileWriter directly mocks an internal implementation
// This will be removed in favor of boundary-based mocks
// Use BoundaryFileWriter from boundary_test_adapter.go instead
type MockFileWriter struct {
	SaveToFileFunc func(content, filePath string) error
}

// SaveToFile implements the file writer interface
func (m *MockFileWriter) SaveToFile(content, filePath string) error {
	return m.SaveToFileFunc(content, filePath)
}

// Deprecated: MockAuditLogger directly mocks an internal implementation
// This will be removed in favor of boundary-based mocks
// Use BoundaryAuditLogger from boundary_test_adapter.go instead
type MockAuditLogger struct {
	LogFunc         func(ctx context.Context, entry auditlog.AuditEntry) error
	LogLegacyFunc   func(entry auditlog.AuditEntry) error
	LogOpFunc       func(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error
	LogOpLegacyFunc func(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error
	CloseFunc       func() error
}

// Log implements the audit logger interface with context
func (m *MockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	if m.LogFunc != nil {
		return m.LogFunc(ctx, entry)
	}

	// Default implementation: Add correlation ID from context
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		if entry.Inputs == nil {
			entry.Inputs = make(map[string]interface{})
		}
		if _, exists := entry.Inputs["correlation_id"]; !exists {
			entry.Inputs["correlation_id"] = correlationID
		}
	}
	return nil
}

// LogLegacy implements the backward-compatible AuditLogger.LogLegacy method
func (m *MockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	if m.LogLegacyFunc != nil {
		return m.LogLegacyFunc(entry)
	}
	return m.Log(context.Background(), entry)
}

// LogOp implements the audit logger interface with context
func (m *MockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	if m.LogOpFunc != nil {
		return m.LogOpFunc(ctx, operation, status, inputs, outputs, err)
	}

	// Default implementation: Add correlation ID and call Log
	inputsCopy := make(map[string]interface{})
	for k, v := range inputs {
		inputsCopy[k] = v
	}

	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		if _, exists := inputsCopy["correlation_id"]; !exists {
			inputsCopy["correlation_id"] = correlationID
		}
	}

	entry := auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: operation,
		Status:    status,
		Inputs:    inputsCopy,
		Outputs:   outputs,
		Message:   fmt.Sprintf("%s - %s", operation, status),
	}

	if err != nil {
		entry.Error = &auditlog.ErrorInfo{
			Message: err.Error(),
			Type:    "GeneralError",
		}
	}

	return m.Log(ctx, entry)
}

// LogOpLegacy implements the backward-compatible AuditLogger.LogOpLegacy method
func (m *MockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	if m.LogOpLegacyFunc != nil {
		return m.LogOpLegacyFunc(operation, status, inputs, outputs, err)
	}
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

// Close implements the audit logger interface
func (m *MockAuditLogger) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
