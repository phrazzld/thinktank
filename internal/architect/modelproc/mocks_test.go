package modelproc_test

import (
	"context"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
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

// mockGeminiClient is a mock implementation of gemini.Client for testing
type mockGeminiClient struct {
	generateContentFunc    func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	countTokensFunc        func(ctx context.Context, prompt string) (*gemini.TokenCount, error)
	getModelInfoFunc       func(ctx context.Context) (*gemini.ModelInfo, error)
	getModelNameFunc       func() string
	getTemperatureFunc     func() float32
	getMaxOutputTokensFunc func() int32
	getTopPFunc            func() float32
	closeFunc              func() error
}

func (m *mockGeminiClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt)
	}
	return &gemini.GenerationResult{Content: "mock content"}, nil
}

func (m *mockGeminiClient) CountTokens(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, prompt)
	}
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *mockGeminiClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &gemini.ModelInfo{InputTokenLimit: 1000}, nil
}

func (m *mockGeminiClient) GetModelName() string {
	if m.getModelNameFunc != nil {
		return m.getModelNameFunc()
	}
	return "test-model"
}

func (m *mockGeminiClient) GetTemperature() float32 {
	if m.getTemperatureFunc != nil {
		return m.getTemperatureFunc()
	}
	return 0.7
}

func (m *mockGeminiClient) GetMaxOutputTokens() int32 {
	if m.getMaxOutputTokensFunc != nil {
		return m.getMaxOutputTokensFunc()
	}
	return 1000
}

func (m *mockGeminiClient) GetTopP() float32 {
	if m.getTopPFunc != nil {
		return m.getTopPFunc()
	}
	return 0.9
}

func (m *mockGeminiClient) Close() error {
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