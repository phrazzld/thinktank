// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
)

// ----- Mock Implementations -----

// MockAuditLogger mocks the auditlog.AuditLogger interface for testing
type MockAuditLogger struct {
	mu      sync.Mutex
	entries []auditlog.AuditEntry
	logErr  error
}

func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		entries: []auditlog.AuditEntry{},
		logErr:  nil,
	}
}

func (m *MockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add correlation ID from context if not already present
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		if entry.Inputs == nil {
			entry.Inputs = make(map[string]interface{})
		}
		// Only add if not already present
		if _, exists := entry.Inputs["correlation_id"]; !exists {
			entry.Inputs["correlation_id"] = correlationID
		}
	}

	m.entries = append(m.entries, entry)
	return m.logErr
}

func (m *MockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	return m.Log(context.Background(), entry)
}

func (m *MockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy of inputs to avoid modifying the original map
	inputsCopy := make(map[string]interface{})
	for k, v := range inputs {
		inputsCopy[k] = v
	}

	// Add correlation ID from context if not already present
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		// Only add if not already present
		if _, exists := inputsCopy["correlation_id"]; !exists {
			inputsCopy["correlation_id"] = correlationID
		}
	}

	// Create an AuditEntry from the parameters
	entry := auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: operation,
		Status:    status,
		Inputs:    inputsCopy,
		Outputs:   outputs,
		Message:   fmt.Sprintf("%s - %s", operation, status),
	}

	// Add error info if provided
	if err != nil {
		entry.Error = &auditlog.ErrorInfo{
			Message: err.Error(),
			Type:    "GeneralError",
		}
	}

	m.entries = append(m.entries, entry)
	return m.logErr
}

func (m *MockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

func (m *MockAuditLogger) Close() error {
	return nil
}

func (m *MockAuditLogger) GetEntries() []auditlog.AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]auditlog.AuditEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

func (m *MockAuditLogger) FindEntry(operation string) *auditlog.AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.entries) - 1; i >= 0; i-- {
		if m.entries[i].Operation == operation {
			return &m.entries[i]
		}
	}
	return nil
}

func (m *MockAuditLogger) SetLogError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logErr = err
}

// MockLogger mocks the logutil.LoggerInterface for testing
type MockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		debugMessages: []string{},
		infoMessages:  []string{},
		warnMessages:  []string{},
		errorMessages: []string{},
	}
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Fatal(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, "FATAL: "+fmt.Sprintf(format, args...))
}

func (m *MockLogger) Println(args ...interface{}) {}

func (m *MockLogger) Printf(format string, args ...interface{}) {}

// Context-aware logging methods
func (m *MockLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) WarnContext(ctx context.Context, format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, "FATAL: "+fmt.Sprintf(format, args...))
}

func (m *MockLogger) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
}

// MockAPIService mocks the APIService interface
type MockAPIService struct {
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string
	initLLMClientErr         error
	mockLLMClient            llm.LLMClient
	processLLMResponseErr    error
	processedContent         string
}

func NewMockAPIService() *MockAPIService {
	return &MockAPIService{
		processedContent:         "Test Generated Plan",
		initLLMClientErr:         nil,
		mockLLMClient:            nil,
		processLLMResponseErr:    nil,
		isEmptyResponseErrorFunc: nil,
		isSafetyBlockedErrorFunc: nil,
		getErrorDetailsFunc:      nil,
	}
}

func (m *MockAPIService) IsEmptyResponseError(err error) bool {
	if m.isEmptyResponseErrorFunc != nil {
		return m.isEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *MockAPIService) IsSafetyBlockedError(err error) bool {
	if m.isSafetyBlockedErrorFunc != nil {
		return m.isSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *MockAPIService) GetErrorDetails(err error) string {
	if m.getErrorDetailsFunc != nil {
		return m.getErrorDetailsFunc(err)
	}
	return err.Error()
}

func (m *MockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.initLLMClientErr != nil {
		return nil, m.initLLMClientErr
	}
	if m.mockLLMClient != nil {
		return m.mockLLMClient, nil
	}
	return &llm.MockLLMClient{}, nil
}

func (m *MockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.processLLMResponseErr != nil {
		return "", m.processLLMResponseErr
	}
	if result == nil {
		return "", errors.New("nil result")
	}
	return m.processedContent, nil
}

func (m *MockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *MockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	return true, nil
}

func (m *MockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	return 8192, 8192, nil
}

// MockOrchestrator mocks the orchestrator for testing
type MockOrchestrator struct {
	runErr error
}

func NewMockOrchestrator() *MockOrchestrator {
	return &MockOrchestrator{
		runErr: nil,
	}
}

func (m *MockOrchestrator) Run(ctx context.Context, instructions string) error {
	return m.runErr
}

// MockLLMClient implements the LLMClient interface for testing
type MockLLMClient struct {
	modelName       string
	generationErr   error
	generatedOutput string
}

func NewMockLLMClient(modelName string) *MockLLMClient {
	return &MockLLMClient{
		modelName:       modelName,
		generatedOutput: "Test Generated Plan",
	}
}

func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.generationErr != nil {
		return nil, m.generationErr
	}
	return &llm.ProviderResult{
		Content:      m.generatedOutput,
		FinishReason: "STOP",
	}, nil
}

func (m *MockLLMClient) Close() error {
	return nil
}

func (m *MockLLMClient) GetModelName() string {
	return m.modelName
}
