package modelproc_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/registry"
	"github.com/stretchr/testify/assert"
)

// Enhanced mock implementations for integration testing
type integrationMockAPIService struct {
	initLLMClientFunc        func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	processLLMResponseFunc   func(result *llm.ProviderResult) (string, error)
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string
	getModelParametersFunc   func(modelName string) (map[string]interface{}, error)
	getModelDefinitionFunc   func(modelName string) (*registry.ModelDefinition, error)
	getModelTokenLimitsFunc  func(modelName string) (contextWindow, maxOutputTokens int32, err error)
}

func newIntegrationMockAPIService() *integrationMockAPIService {
	return &integrationMockAPIService{
		getErrorDetailsFunc: func(err error) string {
			return err.Error()
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
		getModelParametersFunc: func(modelName string) (map[string]interface{}, error) {
			return make(map[string]interface{}), nil
		},
	}
}

func (m *integrationMockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.initLLMClientFunc != nil {
		return m.initLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	// Default implementation returns a mock LLM client
	return newIntegrationMockLLMClient(), nil
}

func (m *integrationMockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.processLLMResponseFunc != nil {
		return m.processLLMResponseFunc(result)
	}
	return result.Content, nil
}

func (m *integrationMockAPIService) IsEmptyResponseError(err error) bool {
	if m.isEmptyResponseErrorFunc != nil {
		return m.isEmptyResponseErrorFunc(err)
	}
	return strings.Contains(err.Error(), "empty")
}

func (m *integrationMockAPIService) IsSafetyBlockedError(err error) bool {
	if m.isSafetyBlockedErrorFunc != nil {
		return m.isSafetyBlockedErrorFunc(err)
	}
	return strings.Contains(err.Error(), "safety") || strings.Contains(err.Error(), "blocked")
}

func (m *integrationMockAPIService) GetErrorDetails(err error) string {
	if m.getErrorDetailsFunc != nil {
		return m.getErrorDetailsFunc(err)
	}
	return err.Error()
}

func (m *integrationMockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	if m.getModelParametersFunc != nil {
		return m.getModelParametersFunc(modelName)
	}
	return make(map[string]interface{}), nil
}

func (m *integrationMockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	if m.getModelDefinitionFunc != nil {
		return m.getModelDefinitionFunc(modelName)
	}
	return &registry.ModelDefinition{
		Name: modelName,
	}, nil
}

func (m *integrationMockAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.getModelTokenLimitsFunc != nil {
		return m.getModelTokenLimitsFunc(modelName)
	}
	return 4000, 1000, nil
}

type tokenConfirmCall struct {
	tokenCount int32
	threshold  int
}

func newIntegrationMockTokenManager() *integrationMockTokenManager {
	return &integrationMockTokenManager{
		tokenInfo: &modelproc.TokenResult{
			TokenCount:   100,
			InputLimit:   1000,
			ExceedsLimit: false,
			LimitError:   "",
			Percentage:   10.0,
		},
		confirmResponse:   true,
		getTokenInfoCalls: make([]string, 0),
		checkTokenCalls:   make([]string, 0),
		confirmCalls:      make([]tokenConfirmCall, 0),
	}
}

func (m *integrationMockTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
	m.getTokenInfoCalls = append(m.getTokenInfoCalls, prompt)

	if m.tokenInfoError != nil {
		return nil, m.tokenInfoError
	}

	return m.tokenInfo, nil
}

func (m *integrationMockTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	m.checkTokenCalls = append(m.checkTokenCalls, prompt)

	return m.checkTokenLimitErr
}

func (m *integrationMockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	m.confirmCalls = append(m.confirmCalls, tokenConfirmCall{
		tokenCount: tokenCount,
		threshold:  threshold,
	})

	return m.confirmResponse
}

// Verify methods for test assertions
func (m *integrationMockTokenManager) VerifyTokenChecks(t *testing.T, expectedCount int) {
	assert.Equal(t, expectedCount, len(m.checkTokenCalls), "Expected %d token limit checks, got %d", expectedCount, len(m.checkTokenCalls))
}

type integrationMockTokenManager struct {
	tokenInfo          *modelproc.TokenResult
	tokenInfoError     error
	checkTokenLimitErr error
	confirmResponse    bool

	getTokenInfoCalls []string
	checkTokenCalls   []string
	confirmCalls      []tokenConfirmCall
}

type integrationMockFileWriter struct {
	savedFiles    map[string]string
	saveError     error
	pathValidator func(path string) bool
}

func newIntegrationMockFileWriter() *integrationMockFileWriter {
	return &integrationMockFileWriter{
		savedFiles: make(map[string]string),
		pathValidator: func(path string) bool {
			// Default validator - only fail for specific test paths
			return !strings.Contains(path, "error-model")
		},
	}
}

func (m *integrationMockFileWriter) SaveToFile(content, outputFile string) error {
	// Check if we should simulate an error for this path
	if !m.pathValidator(outputFile) || m.saveError != nil {
		if m.saveError != nil {
			return m.saveError
		}
		return errors.New("simulated error for test path")
	}

	// Save the content
	m.savedFiles[outputFile] = content
	return nil
}

// Test helper methods for verification
func (m *integrationMockFileWriter) VerifyFileSaved(t *testing.T, path string) {
	_, exists := m.savedFiles[path]
	assert.True(t, exists, "Expected file to be saved at path: %s", path)
}

func (m *integrationMockFileWriter) GetSavedContent(path string) (string, bool) {
	content, exists := m.savedFiles[path]
	return content, exists
}

type integrationMockLLMClient struct {
	generateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	countTokensFunc     func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error)
	getModelInfoFunc    func(ctx context.Context) (*llm.ProviderModelInfo, error)
	getModelNameFunc    func() string
	closeFunc           func() error
}

func newIntegrationMockLLMClient() *integrationMockLLMClient {
	return &integrationMockLLMClient{
		generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
			return &llm.ProviderResult{
				Content:      "Generated model output for test",
				TokenCount:   123,
				FinishReason: "STOP",
			}, nil
		},
		countTokensFunc: func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
			return &llm.ProviderTokenCount{Total: int32(len(prompt) / 4)}, nil
		},
		getModelInfoFunc: func(ctx context.Context) (*llm.ProviderModelInfo, error) {
			return &llm.ProviderModelInfo{
				Name:             "test-model",
				InputTokenLimit:  4000,
				OutputTokenLimit: 1000,
			}, nil
		},
		getModelNameFunc: func() string {
			return "test-model"
		},
		closeFunc: func() error {
			return nil
		},
	}
}

func (m *integrationMockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{Content: "mock content"}, nil
}

func (m *integrationMockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, prompt)
	}
	return &llm.ProviderTokenCount{Total: 100}, nil
}

func (m *integrationMockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &llm.ProviderModelInfo{
		Name:             "test-model",
		InputTokenLimit:  4000,
		OutputTokenLimit: 1000,
	}, nil
}

func (m *integrationMockLLMClient) GetModelName() string {
	if m.getModelNameFunc != nil {
		return m.getModelNameFunc()
	}
	return "test-model"
}

func (m *integrationMockLLMClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// integrationMockLogger implements logutil.LoggerInterface for testing
type integrationMockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
	fatalMessages []string
}

func newIntegrationMockLogger() *integrationMockLogger {
	return &integrationMockLogger{
		debugMessages: make([]string, 0),
		infoMessages:  make([]string, 0),
		warnMessages:  make([]string, 0),
		errorMessages: make([]string, 0),
		fatalMessages: make([]string, 0),
	}
}

func (m *integrationMockLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}

func (m *integrationMockLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

func (m *integrationMockLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, args...))
}

func (m *integrationMockLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

func (m *integrationMockLogger) Fatal(format string, args ...interface{}) {
	m.fatalMessages = append(m.fatalMessages, fmt.Sprintf(format, args...))
}

func (m *integrationMockLogger) Println(args ...interface{}) {
	// Not tracked in integration tests
}

func (m *integrationMockLogger) Printf(format string, args ...interface{}) {
	// Not tracked in integration tests
}

type integrationMockAuditLogger struct {
	entries []auditlog.AuditEntry
}

func newIntegrationMockAuditLogger() *integrationMockAuditLogger {
	return &integrationMockAuditLogger{
		entries: make([]auditlog.AuditEntry, 0),
	}
}

func (m *integrationMockAuditLogger) Log(entry auditlog.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *integrationMockAuditLogger) Close() error {
	return nil
}

// Verify methods for test assertions
func (m *integrationMockAuditLogger) VerifyOperationLogged(t *testing.T, operation string) {
	found := false
	for _, entry := range m.entries {
		if entry.Operation == operation {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected audit log entry with operation %s", operation)
}

// TestIntegration_ModelProcessor_BasicWorkflow tests the basic workflow of the ModelProcessor
func TestIntegration_ModelProcessor_BasicWorkflow(t *testing.T) {
	// Set up test dependencies
	mockAPI := newIntegrationMockAPIService()
	mockFileWriter := newIntegrationMockFileWriter()
	mockAudit := newIntegrationMockAuditLogger()
	mockLogger := newIntegrationMockLogger()

	// Configure API service to return specific content
	mockLLMClient := newIntegrationMockLLMClient()
	mockLLMClient.generateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{
			Content:    "Generated model output",
			TokenCount: 50,
		}, nil
	}

	mockAPI.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return mockLLMClient, nil
	}

	// Create configuration
	cfg := &config.CliConfig{
		APIKey:    "test-api-key",
		OutputDir: "/tmp/test-output",
	}

	// Create model processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockFileWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Test input prompt
	prompt := "Test prompt"

	// Process the model
	err := processor.Process(context.Background(), "test-model", prompt)

	// Verify no errors occurred
	assert.NoError(t, err, "Process should not return an error")

	// Verify file writing
	expectedOutputPath := "/tmp/test-output/test-model.md"
	mockFileWriter.VerifyFileSaved(t, expectedOutputPath)

	// Verify file content
	content, exists := mockFileWriter.GetSavedContent(expectedOutputPath)
	assert.True(t, exists, "File content should exist")
	assert.Equal(t, "Generated model output", content, "File content should match model output")

	// Verify audit logging
	mockAudit.VerifyOperationLogged(t, "GenerateContentStart")
	mockAudit.VerifyOperationLogged(t, "GenerateContentEnd")
	mockAudit.VerifyOperationLogged(t, "SaveOutputStart")
	mockAudit.VerifyOperationLogged(t, "SaveOutputEnd")
}

// TestIntegration_ModelProcessor_TokenLimit tests the handling of token limit errors from provider
// Note: As of T032B, TokenManager has been removed from ModelProcessor
func TestIntegration_ModelProcessor_TokenLimit(t *testing.T) {
	// Set up test dependencies
	mockAPI := newIntegrationMockAPIService()
	mockFileWriter := newIntegrationMockFileWriter()
	mockAudit := newIntegrationMockAuditLogger()
	mockLogger := newIntegrationMockLogger()

	// Configure API mock to return an error for GenerateContent
	// simulating a provider token limit error
	mockAPI.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		mockClient := newIntegrationMockLLMClient()
		mockClient.generateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
			return nil, errors.New("token limit exceeded from provider")
		}
		return mockClient, nil
	}

	// Create configuration
	cfg := &config.CliConfig{
		APIKey:                "test-api-key",
		ModelNames:            []string{"test-model"},
		OutputDir:             "/tmp/test-output",
		ConfirmTokens:         0, // This flag is now unused but still in the config
		MaxConcurrentRequests: 1,
	}

	// Create model processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockFileWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Test input prompt
	prompt := "Test prompt exceeding token limit"

	// Process the model - should fail due to token limit
	err := processor.Process(context.Background(), "test-model", prompt)

	// Verify error is from the provider API call
	assert.Error(t, err, "Process should return an error from the provider")
	assert.Contains(t, err.Error(), "token limit exceeded from provider", "Error should be from the provider")

	// Verify no file was written
	assert.Empty(t, mockFileWriter.savedFiles, "No files should be saved when token limit is exceeded")
}

// TestIntegration_ModelProcessor_FileWriter tests the interaction between
// ModelProcessor and FileWriter for output saving
func TestIntegration_ModelProcessor_FileWriter(t *testing.T) {
	// Set up test dependencies
	mockAPI := newIntegrationMockAPIService()
	mockFileWriter := newIntegrationMockFileWriter()
	mockAudit := newIntegrationMockAuditLogger()
	mockLogger := newIntegrationMockLogger()

	// Configure API service to return specific content
	expectedContent := "This is the specific generated content that should be saved to file"
	mockLLMClient := newIntegrationMockLLMClient()
	mockLLMClient.generateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{
			Content:      expectedContent,
			TokenCount:   100,
			FinishReason: "STOP",
		}, nil
	}

	mockAPI.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return mockLLMClient, nil
	}

	// Simulate a file writing error
	mockFileWriter.saveError = errors.New("simulated file write error")

	// Create configuration
	cfg := &config.CliConfig{
		APIKey:                "test-api-key",
		ModelNames:            []string{"test-model", "error-model"},
		OutputDir:             "/tmp/test-output",
		ConfirmTokens:         0,
		MaxConcurrentRequests: 1,
	}

	// Create model processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockFileWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Test with filename that will cause an error
	err := processor.Process(context.Background(), "error-model", "Test prompt")

	// Verify error related to file writing
	assert.Error(t, err, "Process should return an error when file writing fails")
	assert.Contains(t, err.Error(), "simulated file write error", "Error should contain the original error message")

	// Clear the error and test with a normal model name
	mockFileWriter.saveError = nil

	// Process with a normal model name
	err = processor.Process(context.Background(), "test-model", "Test prompt")

	// Verify no error
	assert.NoError(t, err, "Process should not return an error for normal model")

	// Verify file writing
	expectedOutputPath := "/tmp/test-output/test-model.md"
	mockFileWriter.VerifyFileSaved(t, expectedOutputPath)

	// Verify file content
	content, exists := mockFileWriter.GetSavedContent(expectedOutputPath)
	assert.True(t, exists, "File content should exist")
	assert.Equal(t, expectedContent, content, "File content should match expected output")

	// Verify audit logging for file writing
	mockAudit.VerifyOperationLogged(t, "SaveOutputStart")
	mockAudit.VerifyOperationLogged(t, "SaveOutputEnd")
}

// TestIntegration_ModelProcessor_ErrorHandling tests how ModelProcessor handles
// various error scenarios from its dependencies
func TestIntegration_ModelProcessor_ErrorHandling(t *testing.T) {
	// Define test cases for different error scenarios
	testCases := []struct {
		name           string
		setupMocks     func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter)
		expectedErrMsg string
	}{
		{
			name: "API client initialization error",
			setupMocks: func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter) {
				api.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
					return nil, errors.New("API client initialization failed")
				}
			},
			expectedErrMsg: "API client initialization failed",
		},
		{
			name: "Token limit error from provider",
			setupMocks: func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter) {
				// Configure API client to return a token limit error
				client := newIntegrationMockLLMClient()
				client.generateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return nil, errors.New("token limit exceeded")
				}
				api.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
					return client, nil
				}
			},
			expectedErrMsg: "token limit exceeded",
		},
		{
			name: "Content generation error",
			setupMocks: func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter) {
				client := newIntegrationMockLLMClient()
				client.generateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return nil, errors.New("content generation failed")
				}
				api.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
					return client, nil
				}
			},
			expectedErrMsg: "content generation failed",
		},
		{
			name: "Response processing error",
			setupMocks: func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter) {
				client := newIntegrationMockLLMClient()
				api.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
					return client, nil
				}
				api.processLLMResponseFunc = func(result *llm.ProviderResult) (string, error) {
					return "", errors.New("response processing failed")
				}
			},
			expectedErrMsg: "response processing failed",
		},
		{
			name: "File writing error",
			setupMocks: func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter) {
				fileWriter.saveError = errors.New("file writing failed")
			},
			expectedErrMsg: "file writing failed",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test dependencies
			mockAPI := newIntegrationMockAPIService()
			mockTokenMgr := newIntegrationMockTokenManager()
			mockFileWriter := newIntegrationMockFileWriter()
			mockAudit := newIntegrationMockAuditLogger()
			mockLogger := newIntegrationMockLogger()

			// Setup the specific test scenario
			tc.setupMocks(mockAPI, mockTokenMgr, mockFileWriter)

			// Create configuration
			cfg := &config.CliConfig{
				APIKey:                "test-api-key",
				ModelNames:            []string{"test-model"},
				OutputDir:             "/tmp/test-output",
				ConfirmTokens:         0,
				MaxConcurrentRequests: 1,
			}

			// Token validation has been removed as part of T032B

			// Create model processor with updated constructor signature
			processor := modelproc.NewProcessor(
				mockAPI,
				mockFileWriter,
				mockAudit,
				mockLogger,
				cfg,
			)

			// Process the model
			err := processor.Process(context.Background(), "test-model", "Test prompt")

			// Verify error
			assert.Error(t, err, "Process should return an error for %s", tc.name)
			assert.Contains(t, err.Error(), tc.expectedErrMsg, "Error should contain expected message for %s", tc.name)
		})
	}
}

// TestIntegration_ModelProcessor_OutputFilePath tests the file path generation
// logic in the ModelProcessor
func TestIntegration_ModelProcessor_OutputFilePath(t *testing.T) {
	// Skip the complex path validation test for now
	t.Skip("Temporarily skipping output path validation test until T032B is completed")
}
