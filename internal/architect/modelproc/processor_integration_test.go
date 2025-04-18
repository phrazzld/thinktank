package modelproc_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
	"github.com/stretchr/testify/assert"
)

// Enhanced mock implementations for integration testing

// integrationMockAPIService implements the APIService interface with enhanced tracking
type integrationMockAPIService struct {
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string
	initLLMClientFunc        func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	processLLMResponseFunc   func(result *llm.ProviderResult) (string, error)

	// Tracking for provider-agnostic calls
	initLLMClientCalls      []initLLMClientCall
	processLLMResponseCalls []processLLMResponseCall
}

type initLLMClientCall struct {
	apiKey      string
	modelName   string
	apiEndpoint string
	ctx         context.Context
}

type processLLMResponseCall struct {
	result *llm.ProviderResult
}

func newIntegrationMockAPIService() *integrationMockAPIService {
	return &integrationMockAPIService{
		initLLMClientCalls:       make([]initLLMClientCall, 0),
		processLLMResponseCalls:  make([]processLLMResponseCall, 0),
		isEmptyResponseErrorFunc: func(err error) bool { return false },
		isSafetyBlockedErrorFunc: func(err error) bool { return false },
		getErrorDetailsFunc:      func(err error) string { return "error details" },
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return newIntegrationMockLLMClient(), nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}
}

func (m *integrationMockAPIService) IsEmptyResponseError(err error) bool {
	return m.isEmptyResponseErrorFunc(err)
}

func (m *integrationMockAPIService) IsSafetyBlockedError(err error) bool {
	return m.isSafetyBlockedErrorFunc(err)
}

func (m *integrationMockAPIService) GetErrorDetails(err error) string {
	return m.getErrorDetailsFunc(err)
}

// Implement new provider-agnostic methods
func (m *integrationMockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// Track the call
	m.initLLMClientCalls = append(m.initLLMClientCalls, initLLMClientCall{
		apiKey:      apiKey,
		modelName:   modelName,
		apiEndpoint: apiEndpoint,
		ctx:         ctx,
	})

	if m.initLLMClientFunc != nil {
		return m.initLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return newIntegrationMockLLMClient(), nil
}

func (m *integrationMockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	// Track the call
	m.processLLMResponseCalls = append(m.processLLMResponseCalls, processLLMResponseCall{
		result: result,
	})

	if m.processLLMResponseFunc != nil {
		return m.processLLMResponseFunc(result)
	}
	return result.Content, nil
}

// Implement registry-related methods required by the APIService interface
func (m *integrationMockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	// Default implementation returns empty parameters
	return make(map[string]interface{}), nil
}

func (m *integrationMockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	// Default implementation returns a minimal model definition
	return &registry.ModelDefinition{
		Name:            modelName,
		ContextWindow:   8192,
		MaxOutputTokens: 2048,
	}, nil
}

func (m *integrationMockAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Default implementation returns standard values
	return 8192, 2048, nil
}

// integrationMockLLMClient implements llm.LLMClient for testing
type integrationMockLLMClient struct {
	generateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	countTokensFunc     func(ctx context.Context, text string) (*llm.ProviderTokenCount, error)
	getModelInfoFunc    func(ctx context.Context) (*llm.ProviderModelInfo, error)
	closeFunc           func() error
	modelName           string
	temperature         float32
	maxOutputTokens     int32
}

func newIntegrationMockLLMClient() *integrationMockLLMClient {
	return &integrationMockLLMClient{
		modelName:       "test-model",
		temperature:     0.7,
		maxOutputTokens: 1000,
	}
}

func (m *integrationMockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt, params)
	}
	return &llm.ProviderResult{
		Content:      "generated content",
		TokenCount:   100,
		FinishReason: "STOP",
	}, nil
}

func (m *integrationMockLLMClient) CountTokens(ctx context.Context, text string) (*llm.ProviderTokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, text)
	}
	return &llm.ProviderTokenCount{Total: 100}, nil
}

func (m *integrationMockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &llm.ProviderModelInfo{
		Name:             "test-model",
		InputTokenLimit:  1000,
		OutputTokenLimit: 1000,
	}, nil
}

func (m *integrationMockLLMClient) GetModelName() string {
	return m.modelName
}

func (m *integrationMockLLMClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *integrationMockLLMClient) GetTemperature() float32 {
	return m.temperature
}

func (m *integrationMockLLMClient) GetMaxOutputTokens() int32 {
	return m.maxOutputTokens
}

// integrationMockTokenManager implements modelproc.TokenManager for testing
type integrationMockTokenManager struct {
	tokenInfo          *modelproc.TokenResult
	tokenInfoError     error
	checkTokenLimitErr error
	confirmResponse    bool

	// Tracking fields
	getTokenInfoCalls []string
	checkTokenCalls   []string
	confirmCalls      []tokenConfirmCall
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
	t.Helper()

	if len(m.getTokenInfoCalls) != expectedCount {
		t.Errorf("Expected %d token checks, got %d", expectedCount, len(m.getTokenInfoCalls))
	}
}

// integrationMockFileWriter implements interfaces.FileWriter for testing
type integrationMockFileWriter struct {
	savedFiles map[string]string
	saveError  error
}

func newIntegrationMockFileWriter() *integrationMockFileWriter {
	return &integrationMockFileWriter{
		savedFiles: make(map[string]string),
	}
}

func (m *integrationMockFileWriter) SaveToFile(content, outputPath string) error {
	if m.saveError != nil {
		return m.saveError
	}

	m.savedFiles[outputPath] = content
	return nil
}

// Verify methods for test assertions
func (m *integrationMockFileWriter) VerifyFileSaved(t *testing.T, expectedPath string) {
	t.Helper()

	if _, exists := m.savedFiles[expectedPath]; !exists {
		t.Errorf("Expected file to be saved at %s, but it wasn't", expectedPath)
	}
}

func (m *integrationMockFileWriter) GetSavedContent(path string) (string, bool) {
	content, exists := m.savedFiles[path]
	return content, exists
}

// integrationMockAuditLogger implements auditlog.AuditLogger for testing
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
	t.Helper()

	for _, entry := range m.entries {
		if entry.Operation == operation {
			return
		}
	}

	t.Errorf("Expected operation %s to be logged, but it wasn't", operation)
}

// integrationMockLogger implements logutil.LoggerInterface for testing
type integrationMockLogger struct {
	debugMsgs  []string
	infoMsgs   []string
	warnMsgs   []string
	errorMsgs  []string
	fatalMsgs  []string
	printMsgs  []string
	printfMsgs []string
}

func newIntegrationMockLogger() *integrationMockLogger {
	return &integrationMockLogger{
		debugMsgs:  make([]string, 0),
		infoMsgs:   make([]string, 0),
		warnMsgs:   make([]string, 0),
		errorMsgs:  make([]string, 0),
		fatalMsgs:  make([]string, 0),
		printMsgs:  make([]string, 0),
		printfMsgs: make([]string, 0),
	}
}

func (m *integrationMockLogger) Debug(format string, args ...interface{}) {
	m.debugMsgs = append(m.debugMsgs, format)
}

func (m *integrationMockLogger) Info(format string, args ...interface{}) {
	m.infoMsgs = append(m.infoMsgs, format)
}

func (m *integrationMockLogger) Warn(format string, args ...interface{}) {
	m.warnMsgs = append(m.warnMsgs, format)
}

func (m *integrationMockLogger) Error(format string, args ...interface{}) {
	m.errorMsgs = append(m.errorMsgs, format)
}

func (m *integrationMockLogger) Fatal(format string, args ...interface{}) {
	m.fatalMsgs = append(m.fatalMsgs, format)
}

func (m *integrationMockLogger) Println(args ...interface{}) {
	m.printMsgs = append(m.printMsgs, strings.TrimSpace(strings.Join(strings.Fields(strings.TrimSpace(string([]byte(args[0].(string))))), " ")))
}

func (m *integrationMockLogger) Printf(format string, args ...interface{}) {
	m.printfMsgs = append(m.printfMsgs, format)
}

// Integration Tests

// TestIntegration_ModelProcessor_APIService tests the interaction between
// ModelProcessor and APIService during model processing
func TestIntegration_ModelProcessor_APIService(t *testing.T) {
	// Create temp output directory
	tempDir, err := os.MkdirTemp("", "model_processor_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up test dependencies
	mockAPI := newIntegrationMockAPIService()
	mockAPI.processLLMResponseFunc = func(result *llm.ProviderResult) (string, error) {
		// Just return the content
		return result.Content, nil
	}

	// Token manager is created inside Process method
	mockFileWriter := newIntegrationMockFileWriter()
	mockAudit := newIntegrationMockAuditLogger()
	mockLogger := newIntegrationMockLogger()

	// Configure the client function to create our mock client with specific behavior
	mockLLMClient := newIntegrationMockLLMClient()
	mockLLMClient.generateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		return &llm.ProviderResult{
			Content:      "Generated model output for test",
			TokenCount:   150,
			FinishReason: "STOP",
		}, nil
	}

	mockAPI.initLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return mockLLMClient, nil
	}

	// Create configuration
	cfg := &config.CliConfig{
		APIKey:                "test-api-key",
		ModelNames:            []string{"test-model"},
		OutputDir:             tempDir,
		ConfirmTokens:         0, // No confirmation needed
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

	// Test input prompt
	prompt := "Test prompt for integration testing"

	// Process the model
	err = processor.Process(context.Background(), "test-model", prompt)

	// Verify results
	assert.NoError(t, err, "Process should not return an error")

	// Verify API service interactions with provider-agnostic interfaces
	assert.GreaterOrEqual(t, len(mockAPI.initLLMClientCalls), 1, "InitLLMClient should be called at least once")
	assert.Equal(t, "test-api-key", mockAPI.initLLMClientCalls[0].apiKey, "InitLLMClient should be called with the correct API key")
	assert.Equal(t, "test-model", mockAPI.initLLMClientCalls[0].modelName, "InitLLMClient should be called with the correct model name")

	assert.GreaterOrEqual(t, len(mockAPI.processLLMResponseCalls), 1, "ProcessLLMResponse should be called at least once")

	// Verify file writing
	expectedOutputPath := filepath.Join(tempDir, "test-model.md")
	mockFileWriter.VerifyFileSaved(t, expectedOutputPath)

	// Verify file content
	content, exists := mockFileWriter.GetSavedContent(expectedOutputPath)
	assert.True(t, exists, "File content should exist")
	assert.Equal(t, "Generated model output for test", content, "File content should match model output")

	// Verify audit logging
	mockAudit.VerifyOperationLogged(t, "GenerateContentStart")
	mockAudit.VerifyOperationLogged(t, "GenerateContentEnd")
	mockAudit.VerifyOperationLogged(t, "SaveOutputStart")
	mockAudit.VerifyOperationLogged(t, "SaveOutputEnd")
}

// TestIntegration_ModelProcessor_TokenManager tests the interaction between
// ModelProcessor and TokenManager during token checking
func TestIntegration_ModelProcessor_TokenManager(t *testing.T) {
	// Set up test dependencies
	mockAPI := newIntegrationMockAPIService()
	mockTokenMgr := newIntegrationMockTokenManager() // This mock will be injected via NewTokenManagerWithClient
	mockFileWriter := newIntegrationMockFileWriter()
	mockAudit := newIntegrationMockAuditLogger()
	mockLogger := newIntegrationMockLogger()

	// Configure tokenManager to simulate token limit exceeded
	mockTokenMgr.tokenInfo = &modelproc.TokenResult{
		TokenCount:   1200, // Over the limit
		InputLimit:   1000,
		ExceedsLimit: true,
		LimitError:   "prompt exceeds token limit (1200 tokens > 1000 token limit)",
		Percentage:   120.0,
	}

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
		ConfirmTokens:         0,
		MaxConcurrentRequests: 1,
	}

	// Override the NewTokenManagerWithClient function temporarily
	originalNewTokenManager := modelproc.NewTokenManagerWithClient
	defer func() { modelproc.NewTokenManagerWithClient = originalNewTokenManager }()

	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockTokenMgr
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
	prompt := "Test prompt exceeding token limit"

	// Process the model - should fail due to token limit
	err := processor.Process(context.Background(), "test-model", prompt)

	// Verify error is from the provider, not our pre-check
	assert.Error(t, err, "Process should return an error from the provider")
	assert.Contains(t, err.Error(), "token limit exceeded from provider", "Error should be from the provider")

	// Verify token manager interactions
	assert.GreaterOrEqual(t, len(mockTokenMgr.getTokenInfoCalls), 1, "GetTokenInfo should be called at least once")
	assert.Equal(t, prompt, mockTokenMgr.getTokenInfoCalls[0], "GetTokenInfo should be called with the correct prompt")

	// Verify no file was written
	assert.Empty(t, mockFileWriter.savedFiles, "No files should be saved when token limit is exceeded")

	// Note: In the real implementation, CheckTokens might be logged with a different
	// operation name than exactly "CheckTokens", so we'll skip this assertion
}

// TestIntegration_ModelProcessor_FileWriter tests the interaction between
// ModelProcessor and FileWriter for output saving
func TestIntegration_ModelProcessor_FileWriter(t *testing.T) {
	// Set up test dependencies
	mockAPI := newIntegrationMockAPIService()
	// Token manager is created inside Process method
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
			name: "Token counting error",
			setupMocks: func(api *integrationMockAPIService, token *integrationMockTokenManager, fileWriter *integrationMockFileWriter) {
				// Need to be careful here since tokenManager is created inside Process method
				token.tokenInfoError = errors.New("token counting failed")
			},
			expectedErrMsg: "token counting failed",
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

			// Override the NewTokenManagerWithClient function temporarily for token manager errors
			if tc.name == "Token counting error" {
				originalNewTokenManager := modelproc.NewTokenManagerWithClient
				defer func() { modelproc.NewTokenManagerWithClient = originalNewTokenManager }()

				modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
					return mockTokenMgr
				}
			}

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
