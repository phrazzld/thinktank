package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

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

// TestRun_EmptyModelNames tests that no error is returned when no models are specified
// but dry run is enabled, since dry run only displays context info
func TestRun_EmptyModelNames(t *testing.T) {
	// Create mock dependencies
	mockClient := &mockGeminiClient{}
	mockAPIService := &mockAPIService{
		client: mockClient,
	}

	mockGatherer := &mockContextGatherer{
		contextFiles: []fileutil.FileMeta{
			{Path: "file1.go", Content: "content1"},
		},
		contextStats: &interfaces.ContextStats{TokenCount: 100},
	}

	mockAuditLogger := &mockAuditLogger{}
	mockLogger := &mockLogger{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}

	// Create a config with empty model names but with dry run enabled
	cfg := &config.CliConfig{
		ModelNames: []string{},
		DryRun:     true, // Enable dry run mode
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

	// Verify no error was returned in dry run mode
	if err != nil {
		t.Errorf("Expected no error with dry run enabled, got %v", err)
	}
}

// TestRun_ModelProcessingError tests error handling when model processing fails
func TestRun_ModelProcessingError(t *testing.T) {
	// Create a mock client that returns an error for InitClient
	expectedError := errors.New("model processing error")
	mockAPIService := &mockAPIService{
		initClientErr: expectedError,
	}

	mockGatherer := &mockContextGatherer{
		contextFiles: []fileutil.FileMeta{
			{Path: "file1.go", Content: "content1"},
		},
		contextStats: &interfaces.ContextStats{TokenCount: 100},
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
	}
}

// TestRun_ContextCancellation tests that processing is interrupted when context is cancelled
func TestRun_ContextCancellation(t *testing.T) {
	// Create mock dependencies with a delay in client initialization
	mockAPIService := &mockAPIService{
		initClientDelay: true, // This will simulate a slow operation that should be cancelled
	}

	mockGatherer := &mockContextGatherer{
		contextFiles: []fileutil.FileMeta{
			{Path: "file1.go", Content: "content1"},
		},
		contextStats: &interfaces.ContextStats{TokenCount: 100},
	}

	mockAuditLogger := &mockAuditLogger{}
	mockLogger := &mockLogger{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}

	cfg := &config.CliConfig{
		ModelNames: []string{"model1", "model2", "model3"}, // Multiple models to ensure we have time to cancel
	}

	limiter := ratelimit.NewRateLimiter(1, 1) // Slow rate limit to ensure cancellation happens

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

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after a small delay
	go func() {
		cancel() // Cancel immediately for test purposes
	}()

	// Run the orchestrator with the cancellable context
	err := orch.Run(ctx, "test instructions")

	// Verify the error indicates context cancellation (either directly or in the aggregated error)
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	} else if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected error to mention context cancellation, got: %v", err)
	}
}

// TestBuildPrompt tests the buildPrompt method to ensure it properly combines
// instructions with context files
func TestBuildPrompt(t *testing.T) {
	// Create some test data
	instructions := "# Test Instructions\nThis is a test"
	contextFiles := []fileutil.FileMeta{
		{Path: "file1.go", Content: "package main\n\nfunc main() {}"},
		{Path: "file2.go", Content: "package test\n\nfunc Test() {}"},
	}

	// Create an orchestrator instance
	mockAPIService := &mockAPIService{}
	mockGatherer := &mockContextGatherer{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}
	mockAuditLogger := &mockAuditLogger{}
	mockLogger := &mockLogger{}
	cfg := &config.CliConfig{}
	limiter := ratelimit.NewRateLimiter(1, 1)

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

	// Call the buildPrompt method
	result := orch.buildPrompt(instructions, contextFiles)

	// Verify that the result contains the instructions
	if !strings.Contains(result, instructions) {
		t.Errorf("Expected prompt to contain instructions, but it doesn't")
	}

	// Verify that the result contains both file paths
	if !strings.Contains(result, "file1.go") || !strings.Contains(result, "file2.go") {
		t.Errorf("Expected prompt to contain file paths, but it doesn't")
	}

	// Verify that the result contains both file contents
	if !strings.Contains(result, "package main") || !strings.Contains(result, "package test") {
		t.Errorf("Expected prompt to contain file contents, but it doesn't")
	}
}

// TestAggregateAndFormatErrors tests the error aggregation functionality
func TestAggregateAndFormatErrors(t *testing.T) {
	// Create an orchestrator instance
	mockAPIService := &mockAPIService{}
	mockGatherer := &mockContextGatherer{}
	mockTokenManager := &mockTokenManager{}
	mockFileWriter := &mockFileWriter{}
	mockAuditLogger := &mockAuditLogger{}
	mockLogger := &mockLogger{}
	cfg := &config.CliConfig{}
	limiter := ratelimit.NewRateLimiter(1, 1)

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

	// Test with no errors - the implementation actually always returns an error message
	// even for empty slices, so we check that it contains the standard prefix
	result := orch.aggregateAndFormatErrors([]error{})
	if result == nil {
		t.Error("Expected non-nil error even for empty slice")
	} else if !strings.Contains(result.Error(), "errors occurred during model processing") {
		t.Errorf("Unexpected error format: %v", result.Error())
	}

	// Test with single error
	modelErrors := []error{
		fmt.Errorf("model1: %w", errors.New("test error")),
	}
	result = orch.aggregateAndFormatErrors(modelErrors)
	if result == nil {
		t.Error("Expected non-nil error for non-empty errors slice")
	} else if !strings.Contains(result.Error(), "model1") || !strings.Contains(result.Error(), "test error") {
		t.Errorf("Error message doesn't contain expected content: %v", result)
	}

	// Test with multiple errors
	modelErrors = []error{
		fmt.Errorf("model1: %w", errors.New("test error 1")),
		fmt.Errorf("model2: %w", errors.New("test error 2")),
	}
	result = orch.aggregateAndFormatErrors(modelErrors)
	if result == nil {
		t.Error("Expected non-nil error for non-empty errors slice")
	} else if !strings.Contains(result.Error(), "model1") ||
		!strings.Contains(result.Error(), "model2") ||
		!strings.Contains(result.Error(), "test error 1") ||
		!strings.Contains(result.Error(), "test error 2") {
		t.Errorf("Error message doesn't contain expected content: %v", result)
	}

	// Test with rate limit errors
	modelErrors = []error{
		fmt.Errorf("model3 rate limit: %w", errors.New("rate limit exceeded")),
	}
	result = orch.aggregateAndFormatErrors(modelErrors)
	if result == nil {
		t.Error("Expected non-nil error for rate limit error")
	} else if !strings.Contains(result.Error(), "rate limit") {
		t.Errorf("Error message doesn't contain rate limit information: %v", result)
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
	client          gemini.Client
	initClientErr   error
	initClientDelay bool
	processErr      error
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Simulate a delay if requested
	if m.initClientDelay {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			// Continue after delay
		}
	}

	if m.initClientErr != nil {
		return nil, m.initClientErr
	}
	return m.client, nil
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	if m.processErr != nil {
		return "", m.processErr
	}
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

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*interfaces.TokenResult, error) {
	// Return a non-nil token result for the test
	return &interfaces.TokenResult{
		TokenCount:   100,
		InputLimit:   1000,
		ExceedsLimit: false,
		LimitError:   "",
		Percentage:   10.0,
	}, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
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
