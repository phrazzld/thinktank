package modelproc_test

import (
	"context"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
)

// Mock implementations
type mockAPIService struct {
	initClientFunc           func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error)
	processResponseFunc      func(result *gemini.GenerationResult) (string, error)
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string

	// New methods for provider-agnostic interface
	initLLMClientFunc      func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	processLLMResponseFunc func(result *llm.ProviderResult) (string, error)
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

// Implement new interface methods for provider-agnostic API
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
	return &modelproc.TokenResult{TokenCount: 100, InputLimit: 1000, ExceedsLimit: false}, nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	if m.promptForConfirmationFunc != nil {
		return m.promptForConfirmationFunc(tokenCount, threshold)
	}
	return true
}

type mockFileWriter struct {
	writeFileFunc  func(path string, content string) error
	saveToFileFunc func(content, outputFile string) error
}

func (m *mockFileWriter) WriteFile(path string, content string) error {
	if m.writeFileFunc != nil {
		return m.writeFileFunc(path, content)
	}
	return nil
}

func (m *mockFileWriter) SaveToFile(content, outputFile string) error {
	if m.saveToFileFunc != nil {
		return m.saveToFileFunc(content, outputFile)
	}
	return nil
}

type mockLLMClient struct {
	generateContentFunc func(ctx context.Context, prompt string) (*llm.ProviderResult, error)
	countTokensFunc     func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error)
	getModelInfoFunc    func(ctx context.Context) (*llm.ProviderModelInfo, error)
	getModelNameFunc    func() string
	closeFunc           func() error
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt)
	}
	return &llm.ProviderResult{Content: "mock content"}, nil
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, prompt)
	}
	return &llm.ProviderTokenCount{Total: 100}, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &llm.ProviderModelInfo{
		Name:             "mock-model",
		InputTokenLimit:  32000,
		OutputTokenLimit: 8000,
	}, nil
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

type mockLogger struct {
	infoFunc    func(format string, args ...interface{})
	debugFunc   func(format string, args ...interface{})
	warnFunc    func(format string, args ...interface{})
	errorFunc   func(format string, args ...interface{})
	printlnFunc func(args ...interface{})
	printfFunc  func(format string, args ...interface{})
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	if m.debugFunc != nil {
		m.debugFunc(format, args...)
	}
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	if m.infoFunc != nil {
		m.infoFunc(format, args...)
	}
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	if m.warnFunc != nil {
		m.warnFunc(format, args...)
	}
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	if m.errorFunc != nil {
		m.errorFunc(format, args...)
	}
}

func (m *mockLogger) Fatal(format string, args ...interface{}) {
	// In tests, we don't want to exit, so just log
	if m.errorFunc != nil {
		m.errorFunc("FATAL: "+format, args...)
	}
}

func (m *mockLogger) Println(args ...interface{}) {
	if m.printlnFunc != nil {
		m.printlnFunc(args...)
	}
}

func (m *mockLogger) Printf(format string, args ...interface{}) {
	if m.printfFunc != nil {
		m.printfFunc(format, args...)
	}
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
