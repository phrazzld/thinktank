package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/ratelimit"
)

// TestNewOrchestrator verifies that the constructor creates an Orchestrator instance
// with the provided dependencies.
func TestNewOrchestrator(t *testing.T) {
	// Create mock dependencies
	apiSvc := &mockAPIService{}
	gatherer := &mockContextGatherer{}
	tokenMgr := &mockTokenManager{}
	writer := &mockFileWriter{}
	auditor := &mockAuditLogger{}
	limiter := ratelimit.NewRateLimiter(1, 1)
	cfg := &config.CliConfig{}
	logger := &mockLogger{}

	// Create the orchestrator with the mock dependencies
	orch := NewOrchestrator(
		apiSvc,
		gatherer,
		tokenMgr,
		writer,
		auditor,
		limiter,
		cfg,
		logger,
	)

	// Verify the orchestrator is not nil
	if orch == nil {
		t.Fatal("Expected non-nil Orchestrator")
	}

	// Verify the dependencies were properly assigned
	// Note: We can't directly compare the interfaces, so we just check for nil
	if orch.apiService == nil {
		t.Errorf("Expected apiService to be non-nil")
	}
	if orch.contextGatherer == nil {
		t.Errorf("Expected contextGatherer to be non-nil")
	}
	if orch.tokenManager == nil {
		t.Errorf("Expected tokenManager to be non-nil")
	}
	if orch.fileWriter == nil {
		t.Errorf("Expected fileWriter to be non-nil")
	}
	if orch.auditLogger == nil {
		t.Errorf("Expected auditLogger to be non-nil")
	}
	if orch.rateLimiter != limiter {
		t.Errorf("Expected rateLimiter to be set correctly")
	}
	if orch.config != cfg {
		t.Errorf("Expected config to be set correctly")
	}
	if orch.logger == nil {
		t.Errorf("Expected logger to be non-nil")
	}
}

// TestRun_DryRun tests the Run method when in dry run mode
func TestRun_DryRun(t *testing.T) {
	// Create mock dependencies
	mockClient := &mockGeminiClient{}
	mockAPIService := &mockAPIService{
		client: mockClient,
	}

	mockContextStats := &interfaces.ContextStats{
		ProcessedFilesCount: 5,
		CharCount:           1000,
		LineCount:           50,
		TokenCount:          100,
	}

	mockGatherer := &mockContextGatherer{
		contextFiles: []fileutil.FileMeta{
			{Path: "file1.go", Content: "content1"},
			{Path: "file2.go", Content: "content2"},
		},
		contextStats: mockContextStats,
	}

	mockAuditLogger := &mockAuditLogger{}
	mockLogger := &mockLogger{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}

	// Create a config with dry run enabled
	cfg := &config.CliConfig{
		DryRun:     true,
		ModelNames: []string{"model1"},
	}

	limiter := ratelimit.NewRateLimiter(1, 1)

	// Create the orchestrator
	orch := NewOrchestrator(
		mockAPIService,
		mockGatherer,
		mockTokenManager,
		mockFileWriter,
		mockAuditLogger,
		limiter,
		cfg,
		mockLogger,
	)

	// Run the orchestrator
	err := orch.Run(context.Background(), "test instructions")

	// Verify there was no error
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the dry run info was displayed
	if !mockGatherer.displayDryRunCalled {
		t.Error("Expected DisplayDryRunInfo to be called")
	}

	// Verify GatherContext was called
	if !mockGatherer.gatherContextCalled {
		t.Error("Expected GatherContext to be called")
	}
}

// TestRun_ModelProcessing tests the Run method with multiple model processing
func TestRun_ModelProcessing(t *testing.T) {
	// Create mock dependencies
	mockClient := &mockGeminiClient{}
	mockAPIService := &mockAPIService{
		client: mockClient,
	}

	mockContextStats := &interfaces.ContextStats{
		ProcessedFilesCount: 5,
		CharCount:           1000,
		LineCount:           50,
		TokenCount:          100,
	}

	mockGatherer := &mockContextGatherer{
		contextFiles: []fileutil.FileMeta{
			{Path: "file1.go", Content: "content1"},
			{Path: "file2.go", Content: "content2"},
		},
		contextStats: mockContextStats,
	}

	mockAuditLogger := &mockAuditLogger{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}
	mockLogger := &mockLogger{}

	// Create a config with multiple models
	cfg := &config.CliConfig{
		DryRun:     false,
		ModelNames: []string{"model1", "model2"},
	}

	limiter := ratelimit.NewRateLimiter(2, 10)

	// Create the orchestrator
	orch := NewOrchestrator(
		mockAPIService,
		mockGatherer,
		mockTokenManager,
		mockFileWriter,
		mockAuditLogger,
		limiter,
		cfg,
		mockLogger,
	)

	// Run the orchestrator
	// This tests that the method will run without panicking, since we can't
	// easily check all internal behavior without a lot more mocking
	err := orch.Run(context.Background(), "test instructions")

	// Verify there was no error
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify GatherContext was called
	if !mockGatherer.gatherContextCalled {
		t.Error("Expected GatherContext to be called")
	}

	// Verify DisplayDryRunInfo was NOT called (since dry run is disabled)
	if mockGatherer.displayDryRunCalled {
		t.Error("DisplayDryRunInfo should not be called when dry run is disabled")
	}
}

// TestRun_GatherContextError tests error handling when context gathering fails
func TestRun_GatherContextError(t *testing.T) {
	// Create mock dependencies with an error in GatherContext
	mockClient := &mockGeminiClient{}
	mockAPIService := &mockAPIService{
		client: mockClient,
	}

	expectedError := errors.New("gather context error")
	mockGatherer := &mockContextGatherer{
		gatherContextError: expectedError,
	}

	mockAuditLogger := &mockAuditLogger{}
	mockLogger := &mockLogger{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}

	cfg := &config.CliConfig{
		ModelNames: []string{"model1"},
	}

	limiter := ratelimit.NewRateLimiter(1, 1)

	// Create the orchestrator
	orch := NewOrchestrator(
		mockAPIService,
		mockGatherer,
		mockTokenManager,
		mockFileWriter,
		mockAuditLogger,
		limiter,
		cfg,
		mockLogger,
	)

	// Run the orchestrator
	err := orch.Run(context.Background(), "test instructions")

	// Verify the error was returned
	if err == nil {
		t.Error("Expected error, got nil")
	} else if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to contain %v, got: %v", expectedError, err)
	}
}

// Mock implementations of the required interfaces

// mockGeminiClient implements gemini.Client
type mockGeminiClient struct{}

func (m *mockGeminiClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	return &gemini.GenerationResult{
		Content:      "generated content",
		TokenCount:   50,
		FinishReason: "STOP",
	}, nil
}

func (m *mockGeminiClient) CountTokens(ctx context.Context, text string) (*gemini.TokenCount, error) {
	return &gemini.TokenCount{
		Total: 100,
	}, nil
}

func (m *mockGeminiClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	return &gemini.ModelInfo{
		InputTokenLimit:  1000,
		OutputTokenLimit: 1000,
	}, nil
}

func (m *mockGeminiClient) Close() error {
	return nil
}

func (m *mockGeminiClient) GetModelName() string {
	return "test-model"
}

func (m *mockGeminiClient) GetTemperature() float32 {
	return 0.7
}

func (m *mockGeminiClient) GetMaxOutputTokens() int32 {
	return 1000
}

func (m *mockGeminiClient) GetTopP() float32 {
	return 0.9
}

// mockAPIService implements architect.APIService
type mockAPIService struct {
	client gemini.Client
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	return m.client, nil
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return result.Content, nil
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return false
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return "error details"
}

// mockContextGatherer implements interfaces.ContextGatherer
type mockContextGatherer struct {
	contextFiles        []fileutil.FileMeta
	contextStats        *interfaces.ContextStats
	gatherContextError  error
	gatherContextCalled bool
	displayDryRunCalled bool
}

func (m *mockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	m.gatherContextCalled = true
	if m.gatherContextError != nil {
		return nil, nil, m.gatherContextError
	}
	return m.contextFiles, m.contextStats, nil
}

func (m *mockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	m.displayDryRunCalled = true
	return nil
}

// mockTokenManager implements interfaces.TokenManager
type mockTokenManager struct{}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*interfaces.TokenResult, error) {
	// Return a non-nil token result for the test
	return &interfaces.TokenResult{
		TokenCount:   100,
		InputLimit:   1000,
		ExceedsLimit: false,
		LimitError:   "",
		Percentage:   10.0,
	}, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return true
}

// mockFileWriter implements interfaces.FileWriter
type mockFileWriter struct{}

func (m *mockFileWriter) SaveToFile(content, outputFile string) error {
	return nil
}

// mockAuditLogger implements auditlog.AuditLogger
type mockAuditLogger struct{}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error {
	return nil
}

func (m *mockAuditLogger) Close() error {
	return nil
}

// mockLogger implements logutil.LoggerInterface
type mockLogger struct{}

func (m *mockLogger) Debug(format string, args ...interface{})  {}
func (m *mockLogger) Info(format string, args ...interface{})   {}
func (m *mockLogger) Warn(format string, args ...interface{})   {}
func (m *mockLogger) Error(format string, args ...interface{})  {}
func (m *mockLogger) Fatal(format string, args ...interface{})  {}
func (m *mockLogger) Println(args ...interface{})               {}
func (m *mockLogger) Printf(format string, args ...interface{}) {}
