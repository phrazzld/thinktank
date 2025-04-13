package modelproc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// Mock implementations
type mockAPIService struct {
	initClientFunc      func(ctx context.Context, apiKey, modelName string) (gemini.Client, error)
	processResponseFunc func(result *gemini.GenerationResult) (string, error)
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	return m.initClientFunc(ctx, apiKey, modelName)
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return m.processResponseFunc(result)
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return false
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return err.Error()
}

type mockTokenManager struct {
	checkTokenLimitFunc       func(ctx context.Context, client gemini.Client, prompt string) error
	getTokenInfoFunc          func(ctx context.Context, client gemini.Client, prompt string) (*architect.TokenResult, error)
	promptForConfirmationFunc func(tokenCount int32, threshold int) bool
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return m.checkTokenLimitFunc(ctx, client, prompt)
}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*architect.TokenResult, error) {
	return m.getTokenInfoFunc(ctx, client, prompt)
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return m.promptForConfirmationFunc(tokenCount, threshold)
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
	return m.logFunc(entry)
}

func (m *mockAuditLogger) Close() error {
	return m.closeFunc()
}

type mockClient struct {
	generateContentFunc func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
}

func (m *mockClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	return m.generateContentFunc(ctx, prompt)
}

func (m *mockClient) CountTokens(ctx context.Context, text string) (*gemini.TokenCount, error) {
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *mockClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	return &gemini.ModelInfo{InputTokenLimit: 1000}, nil
}

func (m *mockClient) GetModelName() string {
	return "test-model"
}

func (m *mockClient) GetTemperature() float32 {
	return 0.7
}

func (m *mockClient) GetMaxOutputTokens() int32 {
	return 2048
}

func (m *mockClient) GetTopP() float32 {
	return 0.9
}

func (m *mockClient) Close() error {
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

func TestModelProcessor_ProcessModel_Success(t *testing.T) {
	// Setup mocks
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
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
		checkTokenLimitFunc: func(ctx context.Context, client gemini.Client, prompt string) error {
			return nil
		},
		getTokenInfoFunc: func(ctx context.Context, client gemini.Client, prompt string) (*architect.TokenResult, error) {
			return &architect.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
		promptForConfirmationFunc: func(tokenCount int32, threshold int) bool {
			return true
		},
	}

	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			return nil
		},
	}

	mockAudit := &mockAuditLogger{
		logFunc: func(entry auditlog.AuditEntry) error {
			return nil
		},
		closeFunc: func() error {
			return nil
		},
	}

	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.ApiKey = "test-api-key"

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
	content, err := processor.ProcessModel(
		context.Background(),
		"test-model",
		"Test prompt",
		"output.md",
	)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if content != "Generated content" {
		t.Errorf("Expected 'Generated content', got '%s'", content)
	}
}

func TestModelProcessor_ProcessModel_ClientInitError(t *testing.T) {
	// Setup mocks
	expectedErr := errors.New("client init error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
			return nil, expectedErr
		},
	}

	mockToken := &mockTokenManager{}
	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.ApiKey = "test-api-key"

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
	_, err := processor.ProcessModel(
		context.Background(),
		"test-model",
		"Test prompt",
		"output.md",
	)

	// Verify results
	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}
