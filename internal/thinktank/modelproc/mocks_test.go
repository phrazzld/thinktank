package modelproc_test

import (
	"context"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

// Mock implementations
type mockAPIService struct {
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string
	initLLMClientFunc        func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	processLLMResponseFunc   func(result *llm.ProviderResult) (string, error)
	getModelParametersFunc   func(ctx context.Context, modelName string) (map[string]interface{}, error)
	getModelDefinitionFunc   func(ctx context.Context, modelName string) (*registry.ModelDefinition, error)
	getModelTokenLimitsFunc  func(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error)
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
	if m.getErrorDetailsFunc != nil {
		return m.getErrorDetailsFunc(err)
	}
	return err.Error()
}

func (m *mockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.initLLMClientFunc != nil {
		return m.initLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	// Default implementation uses a mock LLM client
	return &mockLLMClient{}, nil
}

func (m *mockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.processLLMResponseFunc != nil {
		return m.processLLMResponseFunc(result)
	}
	// Default implementation returns the content
	return result.Content, nil
}

// Implement the new registry methods
func (m *mockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	if m.getModelParametersFunc != nil {
		return m.getModelParametersFunc(ctx, modelName)
	}
	// Default implementation returns empty map
	return make(map[string]interface{}), nil
}

func (m *mockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*registry.ModelDefinition, error) {
	if m.getModelDefinitionFunc != nil {
		return m.getModelDefinitionFunc(ctx, modelName)
	}
	// Default implementation returns nil and error
	return nil, nil
}

func (m *mockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.getModelTokenLimitsFunc != nil {
		return m.getModelTokenLimitsFunc(ctx, modelName)
	}
	// Default implementation returns zeros with no error
	return 0, 0, nil
}

// ValidateModelParameter validates a parameter value against its constraints
func (m *mockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	// Default implementation returns true with no error
	return true, nil
}

// Note: TokenManager implementation and related code has been removed
// as token management is no longer part of the production code

type mockFileWriter struct {
	writeFileFunc  func(path string, content string) error
	saveToFileFunc func(ctx context.Context, content, outputFile string) error
}

func (m *mockFileWriter) WriteFile(path string, content string) error {
	if m.writeFileFunc != nil {
		return m.writeFileFunc(path, content)
	}
	return nil
}

func (m *mockFileWriter) SaveToFile(ctx context.Context, content, outputFile string) error {
	if m.saveToFileFunc != nil {
		return m.saveToFileFunc(ctx, content, outputFile)
	}
	return nil
}

type mockLLMClient struct {
	generateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	getModelNameFunc    func() string
	closeFunc           func() error
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{Content: "mock content"}, nil
}

func (m *mockLLMClient) GetModelName() string {
	if m.getModelNameFunc != nil {
		return m.getModelNameFunc()
	}
	return "mock-model"
}

func (m *mockLLMClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type mockAuditLogger struct {
	logFunc         func(ctx context.Context, entry auditlog.AuditEntry) error
	logLegacyFunc   func(entry auditlog.AuditEntry) error
	logOpFunc       func(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error
	logOpLegacyFunc func(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error
	closeFunc       func() error
}

func (m *mockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	if m.logFunc != nil {
		return m.logFunc(ctx, entry)
	}

	// A simple default implementation
	return nil
}

func (m *mockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	if m.logLegacyFunc != nil {
		return m.logLegacyFunc(entry)
	}
	return m.Log(context.Background(), entry)
}

func (m *mockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	if m.logOpFunc != nil {
		return m.logOpFunc(ctx, operation, status, inputs, outputs, err)
	}
	return nil
}

func (m *mockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	if m.logOpLegacyFunc != nil {
		return m.logOpLegacyFunc(operation, status, inputs, outputs, err)
	}
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

func (m *mockAuditLogger) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}
