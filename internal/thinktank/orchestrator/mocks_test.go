package orchestrator

import (
	"context"
	"errors"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// Mock implementations needed for testing

// MockLogger implements a minimal logger for testing
type MockLogger struct{}

func (m *MockLogger) Println(v ...interface{})                 {}
func (m *MockLogger) Printf(format string, v ...interface{})   {}
func (m *MockLogger) Debug(format string, args ...interface{}) {}
func (m *MockLogger) Info(format string, args ...interface{})  {}
func (m *MockLogger) Warn(format string, args ...interface{})  {}
func (m *MockLogger) Error(format string, args ...interface{}) {}
func (m *MockLogger) Fatal(format string, args ...interface{}) {}

// Context-aware logging methods
func (m *MockLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {}
func (m *MockLogger) InfoContext(ctx context.Context, format string, args ...interface{})  {}
func (m *MockLogger) WarnContext(ctx context.Context, format string, args ...interface{})  {}
func (m *MockLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {}
func (m *MockLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {}

// WithContext returns the logger with context information
func (m *MockLogger) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
}

// MockContextGatherer provides a mock implementation for testing
type MockContextGatherer struct{}

// GatherContext is a mock implementation
func (m *MockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return []fileutil.FileMeta{}, &interfaces.ContextStats{}, nil
}

// DisplayDryRunInfo is a mock implementation
func (m *MockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return nil
}

// MockAuditLogger provides a mock implementation for testing
type MockAuditLogger struct{}

// Log is a mock implementation
func (m *MockAuditLogger) Log(entry auditlog.AuditEntry) error {
	return nil
}

// LogOp is a mock implementation
func (m *MockAuditLogger) LogOp(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return nil
}

// Close is a mock implementation
func (m *MockAuditLogger) Close() error {
	return nil
}

// MockAPIService provides a mock implementation for testing
type MockAPIService struct{}

// InitLLMClient is a mock implementation
func (m *MockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return &MockLLMClient{}, nil
}

// ProcessLLMResponse is a mock implementation
func (m *MockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if result == nil {
		return "", errors.New("nil result")
	}
	return result.Content, nil
}

// IsEmptyResponseError is a mock implementation
func (m *MockAPIService) IsEmptyResponseError(err error) bool {
	return false
}

// IsSafetyBlockedError is a mock implementation
func (m *MockAPIService) IsSafetyBlockedError(err error) bool {
	return false
}

// GetErrorDetails is a mock implementation
func (m *MockAPIService) GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// GetModelParameters is a mock implementation
func (m *MockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// GetModelDefinition is a mock implementation
func (m *MockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	return &registry.ModelDefinition{}, nil
}

// GetModelTokenLimits is a mock implementation
func (m *MockAPIService) GetModelTokenLimits(modelName string) (int32, int32, error) {
	return 8000, 1000, nil
}

// ValidateModelParameter is a mock implementation
func (m *MockAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	return true, nil
}

// MockLLMClient is a mock implementation of llm.LLMClient for testing
type MockLLMClient struct{}

// GenerateContent is a mock implementation
func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:      "Mock content",
		FinishReason: "mock",
	}, nil
}

// GetModelName is a mock implementation
func (m *MockLLMClient) GetModelName() string {
	return "mock-model"
}

// Close is a mock implementation
func (m *MockLLMClient) Close() error {
	return nil
}

// MockFileWriter provides a minimal implementation for testing
type MockFileWriter struct {
	savedFiles map[string]string
	saveError  error
}

// SaveToFile is a mock implementation
func (m *MockFileWriter) SaveToFile(content, outputFile string) error {
	if m.saveError != nil {
		return m.saveError
	}

	if m.savedFiles == nil {
		m.savedFiles = make(map[string]string)
	}

	m.savedFiles[outputFile] = content
	return nil
}
