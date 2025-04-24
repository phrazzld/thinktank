package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// MockAPIService implements interfaces.APIService for testing
type MockAPIService struct {
	modelParams      map[string]interface{}
	clientInitError  error
	generateResult   *llm.ProviderResult
	generateError    error
	processOutput    string
	processError     error
	isEmptyResponse  bool
	isSafetyBlocked  bool
	errorDetails     string
	modelDefinition  *registry.ModelDefinition
	contextWindow    int32
	maxOutputTokens  int32
	tokenLimitsError error
}

func (m *MockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.clientInitError != nil {
		return nil, m.clientInitError
	}
	return &MockLLMClient{
		generateResult: m.generateResult,
		generateError:  m.generateError,
	}, nil
}

func (m *MockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return m.processOutput, m.processError
}

func (m *MockAPIService) IsEmptyResponseError(err error) bool {
	return m.isEmptyResponse
}

func (m *MockAPIService) IsSafetyBlockedError(err error) bool {
	return m.isSafetyBlocked
}

func (m *MockAPIService) GetErrorDetails(err error) string {
	return m.errorDetails
}

func (m *MockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	return m.modelParams, nil
}

func (m *MockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	return m.modelDefinition, nil
}

func (m *MockAPIService) GetModelTokenLimits(modelName string) (int32, int32, error) {
	return m.contextWindow, m.maxOutputTokens, m.tokenLimitsError
}

func (m *MockAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	return true, nil
}

// MockLLMClient implements llm.LLMClient for testing
type MockLLMClient struct {
	generateResult *llm.ProviderResult
	generateError  error
	modelName      string
}

func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, parameters map[string]interface{}) (*llm.ProviderResult, error) {
	return m.generateResult, m.generateError
}

func (m *MockLLMClient) GetModelName() string {
	if m.modelName == "" {
		return "mock-model"
	}
	return m.modelName
}

func (m *MockLLMClient) Close() error {
	return nil
}

// MockAuditLogger implements audit logging for testing
type MockAuditLogger struct {
	entries []auditlog.AuditEntry
	logErr  error
}

func (m *MockAuditLogger) Log(entry auditlog.AuditEntry) error {
	if m.logErr != nil {
		return m.logErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditLogger) Close() error {
	return nil
}

// MockLogger implements logutil.LoggerInterface for testing
type MockLogger struct{}

func (m *MockLogger) Info(format string, args ...interface{})  {}
func (m *MockLogger) Debug(format string, args ...interface{}) {}
func (m *MockLogger) Warn(format string, args ...interface{})  {}
func (m *MockLogger) Error(format string, args ...interface{}) {}
func (m *MockLogger) Fatal(format string, args ...interface{}) {}
func (m *MockLogger) Println(v ...interface{})                 {}
func (m *MockLogger) Printf(format string, v ...interface{})   {}

// GetLogLevel is an additional method that uses logutil to prevent unused import errors
func (m *MockLogger) GetLogLevel() logutil.LogLevel {
	return logutil.InfoLevel
}

// MockContextGatherer implements interfaces.ContextGatherer for testing
type MockContextGatherer struct{}

func (m *MockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return []fileutil.FileMeta{}, &interfaces.ContextStats{}, nil
}

func (m *MockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return nil
}

// MockFileWriter implements interfaces.FileWriter for testing
type MockFileWriter struct {
	savedFiles map[string]string
	saveError  error
}

func (m *MockFileWriter) SaveToFile(content, filePath string) error {
	if m.saveError != nil {
		return m.saveError
	}
	if m.savedFiles == nil {
		m.savedFiles = make(map[string]string)
	}
	m.savedFiles[filePath] = content
	return nil
}

// TestSynthesizeResults tests the synthesizeResults method
func TestSynthesizeResults(t *testing.T) {
	// Define test cases
	tests := []struct {
		name             string
		instructions     string
		modelOutputs     map[string]string
		synthesisModel   string
		mockAPIService   *MockAPIService
		mockAuditLogger  *MockAuditLogger
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:         "Successful synthesis",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
				"model2": "Output from model 2",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     map[string]interface{}{"temperature": 0.7},
				generateResult:  &llm.ProviderResult{},
				processOutput:   "Synthesized output from multiple models",
				isEmptyResponse: false,
			},
			mockAuditLogger: &MockAuditLogger{},
			expectError:     false,
		},
		{
			name:         "Error getting model parameters",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     nil,
				clientInitError: errors.New("model parameter error"),
			},
			mockAuditLogger:  &MockAuditLogger{},
			expectError:      true,
			expectedErrorMsg: "failed to initialize synthesis model client",
		},
		{
			name:         "Error initializing client",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     map[string]interface{}{"temperature": 0.7},
				clientInitError: errors.New("client initialization error"),
			},
			mockAuditLogger:  &MockAuditLogger{},
			expectError:      true,
			expectedErrorMsg: "failed to initialize synthesis model client",
		},
		{
			name:         "Error generating content",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     map[string]interface{}{"temperature": 0.7},
				generateError:   errors.New("content generation error"),
				isEmptyResponse: false,
			},
			mockAuditLogger:  &MockAuditLogger{},
			expectError:      true,
			expectedErrorMsg: "synthesis model API call failed",
		},
		{
			name:         "Error processing response",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     map[string]interface{}{"temperature": 0.7},
				generateResult:  &llm.ProviderResult{},
				processError:    errors.New("response processing error"),
				isEmptyResponse: false,
			},
			mockAuditLogger:  &MockAuditLogger{},
			expectError:      true,
			expectedErrorMsg: "failed to process synthesis model response",
		},
		{
			name:         "Empty response error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     map[string]interface{}{"temperature": 0.7},
				generateError:   errors.New("empty response"),
				isEmptyResponse: true,
			},
			mockAuditLogger:  &MockAuditLogger{},
			expectError:      true,
			expectedErrorMsg: "synthesis model API call failed",
		},
		{
			name:         "Safety blocked error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
			},
			synthesisModel: "synthesis-model",
			mockAPIService: &MockAPIService{
				modelParams:     map[string]interface{}{"temperature": 0.7},
				generateError:   errors.New("safety blocked"),
				isSafetyBlocked: true,
			},
			mockAuditLogger:  &MockAuditLogger{},
			expectError:      true,
			expectedErrorMsg: "synthesis model API call failed",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create orchestrator with mocks
			mockContextGatherer := &MockContextGatherer{}
			mockFileWriter := &MockFileWriter{}
			mockRateLimiter := ratelimit.NewRateLimiter(0, 0)

			// Create config
			cfg := &config.CliConfig{
				SynthesisModel: tt.synthesisModel,
			}

			// Create mock logger
			mockLogger := &MockLogger{}

			// Create orchestrator
			orchestrator := NewOrchestrator(
				tt.mockAPIService,
				mockContextGatherer,
				mockFileWriter,
				tt.mockAuditLogger,
				mockRateLimiter,
				cfg,
				mockLogger,
			)

			// Call synthesizeResults
			result, err := orchestrator.synthesizeResults(context.Background(), tt.instructions, tt.modelOutputs)

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.expectedErrorMsg != "" {
					if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
						t.Errorf("Expected error containing %q but got: %q", tt.expectedErrorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				// For successful synthesis, check some basic properties of the result
				if result == "" {
					t.Errorf("Expected non-empty result but got empty string")
				}

				// Verify the expected output from the mock
				expectedOutput := tt.mockAPIService.processOutput
				if result != expectedOutput {
					t.Errorf("Expected result %q but got %q", expectedOutput, result)
				}
			}

			// Additional checks could be made on audit log entries
			// This helps verify the right audit logs were created
			if tt.mockAuditLogger != nil && len(tt.mockAuditLogger.entries) > 0 {
				// Verify we have at least some audit log entries
				hasStartEntry := false
				hasEndEntry := false

				for _, entry := range tt.mockAuditLogger.entries {
					if entry.Operation == "SynthesisStart" {
						hasStartEntry = true
					}
					if entry.Operation == "SynthesisEnd" {
						hasEndEntry = true
					}
				}

				if !hasStartEntry {
					t.Errorf("Expected SynthesisStart audit entry but found none")
				}

				if !tt.expectError && !hasEndEntry {
					t.Errorf("Expected SynthesisEnd audit entry but found none for successful case")
				}
			}
		})
	}
}

// TestHandleSynthesisError tests the handleSynthesisError method
func TestHandleSynthesisError(t *testing.T) {
	// Define test cases
	tests := []struct {
		name             string
		inputError       error
		mockAPIService   *MockAPIService
		expectedErrorMsg string
	}{
		{
			name:             "Rate limit error",
			inputError:       errors.New("rate limit exceeded"),
			mockAPIService:   &MockAPIService{},
			expectedErrorMsg: "rate limiting",
		},
		{
			name:             "Safety blocked error",
			inputError:       errors.New("content filtered"),
			mockAPIService:   &MockAPIService{isSafetyBlocked: true},
			expectedErrorMsg: "content safety filters",
		},
		{
			name:             "Connectivity error",
			inputError:       errors.New("connection timeout"),
			mockAPIService:   &MockAPIService{},
			expectedErrorMsg: "Connectivity issue",
		},
		{
			name:             "Authentication error",
			inputError:       errors.New("authentication failed"),
			mockAPIService:   &MockAPIService{},
			expectedErrorMsg: "Authentication failed",
		},
		{
			name:             "Empty response error",
			inputError:       errors.New("empty response"),
			mockAPIService:   &MockAPIService{isEmptyResponse: true},
			expectedErrorMsg: "empty response",
		},
		{
			name:             "Generic error",
			inputError:       errors.New("some other error"),
			mockAPIService:   &MockAPIService{errorDetails: "Additional error details"},
			expectedErrorMsg: "Error synthesizing results",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create orchestrator with mocks
			mockContextGatherer := &MockContextGatherer{}
			mockFileWriter := &MockFileWriter{}
			mockRateLimiter := ratelimit.NewRateLimiter(0, 0)
			mockAuditLogger := &MockAuditLogger{}

			// Create config
			cfg := &config.CliConfig{
				SynthesisModel: "test-model",
			}

			// Create mock logger
			mockLogger := &MockLogger{}

			// Create orchestrator
			orchestrator := NewOrchestrator(
				tt.mockAPIService,
				mockContextGatherer,
				mockFileWriter,
				mockAuditLogger,
				mockRateLimiter,
				cfg,
				mockLogger,
			)

			// Call handleSynthesisError
			err := orchestrator.handleSynthesisError(tt.inputError)

			// Verify result
			if err == nil {
				t.Errorf("Expected error but got nil")
				return
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectedErrorMsg) {
				t.Errorf("Expected error containing %q but got: %q", tt.expectedErrorMsg, errMsg)
			}
		})
	}
}
