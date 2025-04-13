package modelproc_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// Mock implementations
type mockAPIService struct {
	initClientFunc           func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error)
	processResponseFunc      func(result *gemini.GenerationResult) (string, error)
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	return m.initClientFunc(ctx, apiKey, modelName, apiEndpoint)
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return m.processResponseFunc(result)
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

type mockTokenManager struct {
	checkTokenLimitFunc       func(ctx context.Context, prompt string) error
	getTokenInfoFunc          func(ctx context.Context, prompt string) (*modelproc.TokenResult, error)
	promptForConfirmationFunc func(tokenCount int32, threshold int) bool
	getTokenInfoCalled        bool
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	if m.checkTokenLimitFunc != nil {
		return m.checkTokenLimitFunc(ctx, prompt)
	}
	return nil
}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
	m.getTokenInfoCalled = true
	if m.getTokenInfoFunc != nil {
		return m.getTokenInfoFunc(ctx, prompt)
	}
	return &modelproc.TokenResult{
		TokenCount:   100,
		InputLimit:   1000,
		ExceedsLimit: false,
		Percentage:   10.0,
	}, nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	if m.promptForConfirmationFunc != nil {
		return m.promptForConfirmationFunc(tokenCount, threshold)
	}
	return true
}

type mockFileWriter struct {
	saveToFileFunc func(content, outputFile string) error
}

func (m *mockFileWriter) SaveToFile(content, outputFile string) error {
	return m.saveToFileFunc(content, outputFile)
}

type mockAuditLogger struct {
	logFunc   func(entry auditlog.AuditEntry) error
	closeFunc func() error
}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error {
	if m.logFunc != nil {
		return m.logFunc(entry)
	}
	return nil
}

func (m *mockAuditLogger) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type mockClient struct {
	generateContentFunc    func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	countTokensFunc        func(ctx context.Context, text string) (*gemini.TokenCount, error)
	getModelInfoFunc       func(ctx context.Context) (*gemini.ModelInfo, error)
	getModelNameFunc       func() string
	getTemperatureFunc     func() float32
	getMaxOutputTokensFunc func() int32
	getTopPFunc            func() float32
	closeFunc              func() error
}

func (m *mockClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	return m.generateContentFunc(ctx, prompt)
}

func (m *mockClient) CountTokens(ctx context.Context, text string) (*gemini.TokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, text)
	}
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *mockClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &gemini.ModelInfo{InputTokenLimit: 1000}, nil
}

func (m *mockClient) GetModelName() string {
	if m.getModelNameFunc != nil {
		return m.getModelNameFunc()
	}
	return "test-model"
}

func (m *mockClient) GetTemperature() float32 {
	if m.getTemperatureFunc != nil {
		return m.getTemperatureFunc()
	}
	return 0.7
}

func (m *mockClient) GetMaxOutputTokens() int32 {
	if m.getMaxOutputTokensFunc != nil {
		return m.getMaxOutputTokensFunc()
	}
	return 2048
}

func (m *mockClient) GetTopP() float32 {
	if m.getTopPFunc != nil {
		return m.getTopPFunc()
	}
	return 0.9
}

func (m *mockClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// Create a no-op logger for testing
type noOpLogger struct{}

func (l *noOpLogger) Println(v ...interface{})               {}
func (l *noOpLogger) Printf(format string, v ...interface{}) {}
func (l *noOpLogger) Debug(format string, v ...interface{})  {}
func (l *noOpLogger) Info(format string, v ...interface{})   {}
func (l *noOpLogger) Warn(format string, v ...interface{})   {}
func (l *noOpLogger) Error(format string, v ...interface{})  {}
func (l *noOpLogger) Fatal(format string, v ...interface{})  {}

func newNoOpLogger() logutil.LoggerInterface {
	return &noOpLogger{}
}

func TestModelProcessor_Process_Success(t *testing.T) {
	// Setup mocks
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return &gemini.GenerationResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			return result.Content, nil
		},
	}

	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Track if SaveToFile was called
	saveToFileCalled := false
	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			saveToFileCalled = true
			// Verify the content and output path are what we expect
			if content != "Generated content" {
				t.Errorf("Expected content 'Generated content', got '%s'", content)
			}
			if filepath.Base(outputFile) != "test-model.md" {
				t.Errorf("Expected output file 'test-model.md', got '%s'", filepath.Base(outputFile))
			}
			return nil
		},
	}

	// Track the audit log entries
	auditEntries := make([]auditlog.AuditEntry, 0)
	mockAudit := &mockAuditLogger{
		logFunc: func(entry auditlog.AuditEntry) error {
			auditEntries = append(auditEntries, entry)
			return nil
		},
	}

	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !saveToFileCalled {
		t.Errorf("Expected SaveToFile to be called, but it wasn't")
	}

	// Verify audit log entries
	expectedOperations := []string{
		"CheckTokensStart",
		"CheckTokens",
		"GenerateContentStart",
		"GenerateContentEnd",
		"SaveOutputStart",
		"SaveOutputEnd",
	}

	if len(auditEntries) != len(expectedOperations) {
		t.Errorf("Expected %d audit entries, got %d", len(expectedOperations), len(auditEntries))
	} else {
		for i, operation := range expectedOperations {
			if auditEntries[i].Operation != operation {
				t.Errorf("Expected audit operation '%s', got '%s'", operation, auditEntries[i].Operation)
			}
		}
	}
}

func TestModelProcessor_Process_TokenLimitExceeded(t *testing.T) {
	// Setup mocks
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
					return &gemini.ModelInfo{
						Name:             "test-model",
						InputTokenLimit:  4000,
						OutputTokenLimit: 1000,
					}, nil
				},
				countTokensFunc: func(ctx context.Context, text string) (*gemini.TokenCount, error) {
					return &gemini.TokenCount{Total: 5000}, nil
				},
				getModelNameFunc: func() string {
					return "test-model"
				},
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					// This should not be called in the token exceeded case
					return nil, errors.New("should not be called")
				},
			}, nil
		},
	}
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   5000,
				InputLimit:   4000,
				ExceedsLimit: true,
				LimitError:   "prompt exceeds token limit (5000 tokens > 4000 token limit)",
				Percentage:   125.0,
			}, nil
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for token limit exceeded, got nil")
	} else if err.Error() != "token limit exceeded for model test-model: prompt exceeds token limit (5000 tokens > 4000 token limit)" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestModelProcessor_Process_ClientInitError(t *testing.T) {
	// Setup mocks
	expectedErr := errors.New("client init error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return nil, expectedErr
		},
	}

	mockToken := &mockTokenManager{}
	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for client initialization, got nil")
	} else if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestModelProcessor_Process_GenerationError(t *testing.T) {
	// Setup mocks
	expectedErr := errors.New("generation error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, expectedErr
				},
			}, nil
		},
	}

	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for generation failure, got nil")
	} else if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestModelProcessor_Process_SaveError(t *testing.T) {
	// Setup mocks
	expectedErr := errors.New("save error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return &gemini.GenerationResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			return result.Content, nil
		},
	}

	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			return expectedErr
		},
	}

	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for save failure, got nil")
	} else if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestModelProcessor_Process_UserCancellation(t *testing.T) {
	// Create a fake implementation of NewTokenManagerWithClient to control the confirmation
	originalNewTokenManagerWithClient := modelproc.NewTokenManagerWithClient

	// Store the original implementation to restore it after the test
	defer func() {
		modelproc.NewTokenManagerWithClient = originalNewTokenManagerWithClient
	}()

	// Replace with our mock that will return user cancellation
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client gemini.Client) modelproc.TokenManager {
		return &mockTokenManager{
			getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
				return &modelproc.TokenResult{
					TokenCount:   100,
					InputLimit:   1000,
					ExceedsLimit: false,
					Percentage:   10.0,
				}, nil
			},
			promptForConfirmationFunc: func(tokenCount int32, threshold int) bool {
				// Return false to simulate user cancellation
				return false
			},
		}
	}

	// Setup mocks
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
					return &gemini.ModelInfo{
						Name:             "test-model",
						InputTokenLimit:  1000,
						OutputTokenLimit: 1000,
					}, nil
				},
				countTokensFunc: func(ctx context.Context, text string) (*gemini.TokenCount, error) {
					return &gemini.TokenCount{Total: 100}, nil
				},
				getModelNameFunc: func() string {
					return "test-model"
				},
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					// Should not be called in this test
					t.Error("GenerateContent should not be called when user cancels")
					return nil, errors.New("should not be called")
				},
			}, nil
		},
	}

	// Create a token manager - this won't be used directly but needed for the NewProcessor call
	mockToken := &mockTokenManager{}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	// The operation will be logged when tokenInfo is fetched
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"
	cfg.ConfirmTokens = 50 // Set a threshold that will trigger confirmation

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results - if generateContent was called, the test would fail with t.Error in the mock
	if err != nil {
		t.Errorf("Expected no error on user cancellation, got %v", err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gemini-1.0-pro", "gemini-1.0-pro"},
		{"gemini/1.0-pro", "gemini-1.0-pro"},
		{"gemini\\1.0:pro", "gemini-1.0-pro"},
		{"gemini-1.0-pro*?\"<>|", "gemini-1.0-pro------"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			// We're calling the unexported sanitizeFilename indirectly through Process
			// by creating a mock setup that lets us check the file path
			mockAPI := &mockAPIService{
				initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
					return &mockClient{
						generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
							return &gemini.GenerationResult{
								Content: "Generated content",
							}, nil
						},
					}, nil
				},
				processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
					return result.Content, nil
				},
			}

			mockToken := &mockTokenManager{
				getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
					return &modelproc.TokenResult{
						TokenCount:   100,
						InputLimit:   1000,
						ExceedsLimit: false,
						Percentage:   10.0,
					}, nil
				},
			}

			// Check if the sanitized filename is used in the output path
			mockWriter := &mockFileWriter{
				saveToFileFunc: func(content, outputFile string) error {
					fileName := filepath.Base(outputFile)
					expectedFileName := test.expected + ".md"
					if fileName != expectedFileName {
						t.Errorf("Expected filename '%s', got '%s'", expectedFileName, fileName)
					}
					return nil
				},
			}

			mockAudit := &mockAuditLogger{}
			mockLogger := newNoOpLogger()

			// Setup config
			cfg := config.NewDefaultCliConfig()
			cfg.APIKey = "test-api-key"
			cfg.OutputDir = "/tmp/test-output"

			// Create processor
			processor := modelproc.NewProcessor(
				mockAPI,
				mockToken,
				mockWriter,
				mockAudit,
				mockLogger,
				cfg,
			)

			// Run Process with the test input as the model name
			_ = processor.Process(
				context.Background(),
				test.input,
				"Test prompt",
			)
		})
	}
}

// TestDirectTokenInfoCall tests the GetTokenInfo method directly
func TestDirectTokenInfoCall(t *testing.T) {
	// Set up a channel to receive a signal when GetTokenInfo is called
	called := make(chan bool, 1)

	// Create a mock token manager with the channel
	tokenManager := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			// Signal that this function was called
			called <- true
			return &modelproc.TokenResult{
				TokenCount:   500,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   50.0,
			}, nil
		},
	}

	// Call GetTokenInfo directly
	result, err := tokenManager.GetTokenInfo(context.Background(), "Test prompt")

	// Verify no errors occurred
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify result values
	if result.TokenCount != 500 {
		t.Errorf("Expected token count 500, got %d", result.TokenCount)
	}

	// Verify the function was called by checking if we received a signal
	select {
	case <-called:
		// Function was called, which is expected
	default:
		t.Error("GetTokenInfo function was not called")
	}
}

// TestProcess_ProcessResponseError tests error handling in response processing
func TestProcess_ProcessResponseError(t *testing.T) {
	// Create expected error
	expectedError := errors.New("response processing error")

	// Create mock API service that returns an error from ProcessResponse
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return &gemini.GenerationResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			return "", expectedError
		},
	}

	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify the error was returned
	if err == nil {
		t.Errorf("Expected error for response processing, got nil")
	} else if !errors.Is(err, expectedError) {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
}

// TestProcess_EmptyResponseError tests handling of empty response errors
func TestProcess_EmptyResponseError(t *testing.T) {
	// Create expected error
	generationErr := errors.New("empty response")

	// Create mock API service that identifies empty response errors
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, generationErr
				},
			}, nil
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return true // Report that this is an empty response error
		},
		getErrorDetailsFunc: func(err error) string {
			return "Empty response error details"
		},
	}

	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify the error was returned and contains appropriate context
	if err == nil {
		t.Errorf("Expected error for empty response, got nil")
	} else if !strings.Contains(err.Error(), "empty response") {
		t.Errorf("Expected error to mention empty response, got: %v", err)
	}
}

// TestProcess_SafetyBlockedError tests handling of safety blocked errors
func TestProcess_SafetyBlockedError(t *testing.T) {
	// Create expected error
	generationErr := errors.New("safety blocked")

	// Create mock API service that identifies safety blocked errors
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, generationErr
				},
			}, nil
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return false
		},
		isSafetyBlockedErrorFunc: func(err error) bool {
			return true // Report that this is a safety blocked error
		},
		getErrorDetailsFunc: func(err error) string {
			return "Content blocked due to safety concerns"
		},
	}

	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify the error was returned and contains appropriate context
	if err == nil {
		t.Errorf("Expected error for safety blocked, got nil")
	} else if !strings.Contains(err.Error(), "safety") {
		t.Errorf("Expected error to mention safety concerns, got: %v", err)
	}
}
