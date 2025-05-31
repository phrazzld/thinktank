package orchestrator

import (
	"context"
	"errors"
	"sync"

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

// LogCall represents a single call to LogOp
type LogCall struct {
	Operation     string
	Status        string
	Inputs        map[string]interface{}
	Outputs       map[string]interface{}
	Error         error
	CorrelationID string // Track correlation ID from context
}

// MockAuditLogger provides a mock implementation for testing
type MockAuditLogger struct {
	LogCalls []LogCall
	LogError error      // To simulate logging errors for testing error handling
	mutex    sync.Mutex // Mutex for thread-safe access to LogCalls
}

// NewMockAuditLogger creates a new instance of MockAuditLogger
func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		LogCalls: make([]LogCall, 0),
		LogError: nil,
	}
}

// Log is a mock implementation with context
func (m *MockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	// Extract correlation ID for testing
	correlationID := logutil.GetCorrelationID(ctx)

	// Add correlation ID to entry inputs if needed
	if correlationID != "" {
		if entry.Inputs == nil {
			entry.Inputs = make(map[string]interface{})
		}
		if _, exists := entry.Inputs["correlation_id"]; !exists {
			entry.Inputs["correlation_id"] = correlationID
		}
	}

	// Lock mutex for thread-safe access to LogError
	m.mutex.Lock()
	err := m.LogError
	m.mutex.Unlock()

	return err
}

// LogLegacy implements the backward-compatible AuditLogger.LogLegacy method
func (m *MockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	return m.Log(context.Background(), entry)
}

// LogOp is a mock implementation with context
func (m *MockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	// Lock the mutex to prevent concurrent access to LogCalls
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Extract correlation ID for testing
	correlationID := logutil.GetCorrelationID(ctx)

	// Make a copy of inputs to avoid modifying the original map
	inputsCopy := make(map[string]interface{})
	for k, v := range inputs {
		inputsCopy[k] = v
	}

	// Add correlation ID to inputs if needed
	if correlationID != "" {
		if _, exists := inputsCopy["correlation_id"]; !exists {
			inputsCopy["correlation_id"] = correlationID
		}
	}

	// Record the call parameters
	m.LogCalls = append(m.LogCalls, LogCall{
		Operation:     operation,
		Status:        status,
		Inputs:        inputsCopy,
		Outputs:       outputs,
		Error:         err,
		CorrelationID: correlationID,
	})

	// Return configured error (nil by default)
	return m.LogError
}

// LogOpLegacy implements the backward-compatible AuditLogger.LogOpLegacy method
func (m *MockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
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
func (m *MockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// GetModelDefinition is a mock implementation
func (m *MockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*registry.ModelDefinition, error) {
	return &registry.ModelDefinition{}, nil
}

// GetModelTokenLimits is a mock implementation
func (m *MockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (int32, int32, error) {
	return 8000, 1000, nil
}

// ValidateModelParameter is a mock implementation
func (m *MockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
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
	mutex      sync.Mutex // Add mutex to protect against concurrent access
}

// NewMockFileWriter creates a new instance of MockFileWriter
func NewMockFileWriter() *MockFileWriter {
	return &MockFileWriter{
		savedFiles: make(map[string]string),
		saveError:  nil,
	}
}

// SaveToFile is a mock implementation
func (m *MockFileWriter) SaveToFile(content, outputFile string) error {
	// Lock the mutex to prevent concurrent access to savedFiles and saveError
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.saveError != nil {
		return m.saveError
	}

	// Initialize map if nil (but this shouldn't happen if NewMockFileWriter is used)
	if m.savedFiles == nil {
		m.savedFiles = make(map[string]string)
	}

	m.savedFiles[outputFile] = content
	return nil
}

// SetSaveError sets the save error in a thread-safe manner
func (m *MockFileWriter) SetSaveError(err error) {
	m.mutex.Lock()
	m.saveError = err
	m.mutex.Unlock()
}

// GetSavedFiles returns a copy of saved files in a thread-safe manner
func (m *MockFileWriter) GetSavedFiles() map[string]string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a copy to avoid data races
	result := make(map[string]string)
	for k, v := range m.savedFiles {
		result[k] = v
	}
	return result
}
