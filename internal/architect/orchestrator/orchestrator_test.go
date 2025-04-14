package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
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

// Integration Tests

// TestIntegration_BasicWorkflow tests the basic workflow of the orchestrator
func TestIntegration_BasicWorkflow(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()
	deps.setupGeminiClient()
	deps.setupMultiModelConfig([]string{"model1", "model2"})

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify no errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify workflow
	deps.verifyBasicWorkflow(t, []string{"model1", "model2"})

	// We're not verifying specific audit log entries since they come from the mock processor
	// and our test is checking the flow, not specific log messages

	// Verify log messages
	deps.Logger.VerifyLogContains(t, "Processing")
}

// TestIntegration_DryRunMode tests the orchestrator in dry run mode
func TestIntegration_DryRunMode(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()
	deps.setupDryRunConfig()

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify no errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify dry run workflow
	deps.verifyDryRunWorkflow(t)
}

// TestIntegration_EmptyModelNames tests that no error is returned when no models are specified
// but dry run is enabled, since dry run only displays context info
func TestIntegration_EmptyModelNames(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()

	// Create a config with empty model names but with dry run enabled
	deps.Config.ModelNames = []string{}
	deps.Config.DryRun = true

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify no error was returned in dry run mode
	if err != nil {
		t.Errorf("Expected no error with dry run enabled, got %v", err)
	}

	// Verify dry run workflow
	deps.verifyDryRunWorkflow(t)
}

// TestIntegration_ErrorPropagation tests error propagation from dependencies
func TestIntegration_ErrorPropagation(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()
	deps.setupMultiModelConfig([]string{"model1", "model2"})

	// Configure errors for different models
	expectedError1 := errors.New("error from model1")
	expectedError2 := errors.New("error from model2")
	deps.APIService.SetModelError("model1", expectedError1)
	deps.APIService.SetModelError("model2", expectedError2)

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify error
	if err == nil {
		t.Error("Expected error, got nil")
	} else {
		// Verify error contains both model errors
		if !strings.Contains(err.Error(), "model1") || !strings.Contains(err.Error(), "model2") {
			t.Errorf("Error does not contain expected model names: %v", err)
		}
		if !strings.Contains(err.Error(), expectedError1.Error()) || !strings.Contains(err.Error(), expectedError2.Error()) {
			t.Errorf("Error does not contain expected error messages: %v", err)
		}
	}
}

// TestIntegration_ContextCancellation tests context cancellation
// This test is skipped when running in short mode (like during pre-commit hooks)
func TestIntegration_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()
	deps.setupMultiModelConfig([]string{"model1", "model2", "model3"})

	// Configure API service to delay client initialization
	deps.APIService.initClientDelay = true

	// Ensure client is not nil to avoid dereference panic
	deps.APIService.client = &mockGeminiClient{}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after a short delay
	go func() {
		cancel() // Cancel immediately for test purposes
	}()

	// Run orchestrator with cancellable context
	err := deps.runOrchestrator(ctx, "test instructions")

	// Verify context cancellation error
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	} else if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected error to mention context cancellation, got: %v", err)
	}
}

// TestIntegration_GatherContextError tests error handling when gathering context fails
func TestIntegration_GatherContextError(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupMultiModelConfig([]string{"model1"})

	// Configure gather context to return an error
	expectedError := errors.New("gather context error")
	deps.ContextGatherer.SetGatherContextError(expectedError)

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify error
	if err == nil {
		t.Error("Expected error, got nil")
	} else if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to be %v, got: %v", expectedError, err)
	}
}

// TestIntegration_RateLimiting tests rate limiting during model processing
// This test is skipped when running in short mode (like during pre-commit hooks)
// because it intentionally uses real time delays to test rate limiting behavior
func TestIntegration_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limiting test in short mode")
	}
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()

	// Set up multiple models
	modelCount := 5
	modelNames := make([]string, modelCount)
	for i := 0; i < modelCount; i++ {
		modelNames[i] = fmt.Sprintf("model%d", i+1)
	}
	deps.setupMultiModelConfig(modelNames)

	// Configure rate limiter with strict limits
	deps.RateLimiter = ratelimit.NewRateLimiter(2, 2) // Max 2 concurrent, 2 per minute

	// Configure API service with delay to ensure rate limiting kicks in
	deps.APIService.initClientDelay = true

	// Ensure client is not nil to avoid dereference panic
	deps.APIService.client = &mockGeminiClient{}

	// Run orchestrator
	startTime := time.Now()
	err := deps.runOrchestrator(context.Background(), "test instructions")
	duration := time.Since(startTime)

	// Verify no errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify workflow
	deps.verifyBasicWorkflow(t, modelNames)

	// Verify rate limiting occurred by checking duration
	// Note: This is a rough approximation, not an exact check
	if duration < 50*time.Millisecond*time.Duration(modelCount/2) {
		t.Errorf("Expected rate limiting to slow execution, but duration was only %v", duration)
	}
}

// TestIntegration_ModelProcessingError tests error handling when model processing fails
func TestIntegration_ModelProcessingError(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()
	deps.setupMultiModelConfig([]string{"model1"})

	// Configure a global client initialization error
	expectedError := errors.New("model processing error")
	deps.APIService.initClientErr = expectedError

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify the error was returned
	if err == nil {
		t.Error("Expected error, got nil")
	} else if !strings.Contains(err.Error(), expectedError.Error()) {
		t.Errorf("Expected error to contain %v, got: %v", expectedError, err)
	}
}

// TestRun_ContextCancellation tests that processing is interrupted when context is cancelled
// This test is skipped when running in short mode (like during pre-commit hooks)
func TestRun_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}
	// Create mock dependencies with a delay in client initialization
	mockAPIService := newMockAPIService()
	mockAPIService.initClientDelay = true       // This will simulate a slow operation that should be cancelled
	mockAPIService.client = &mockGeminiClient{} // Ensure client is not nil to avoid dereference panic

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
	mockAPIService := newMockAPIService()
	mockGatherer := newMockContextGatherer()
	mockTokenManager := newMockTokenManager()
	mockFileWriter := newMockFileWriter()
	mockAuditLogger := newMockAuditLogger()
	mockLogger := newMockLogger()
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

// Integration Test Helper Functions

// orchestratorTestDeps contains all mock dependencies for integration tests
type orchestratorTestDeps struct {
	APIService      *mockAPIService
	ContextGatherer *mockContextGatherer
	TokenManager    *mockTokenManager
	FileWriter      *mockFileWriter
	AuditLogger     *mockAuditLogger
	Logger          *mockLogger
	RateLimiter     *ratelimit.RateLimiter
	Config          *config.CliConfig
}

// newTestDeps creates a new set of test dependencies
func newTestDeps() *orchestratorTestDeps {
	return &orchestratorTestDeps{
		APIService:      newMockAPIService(),
		ContextGatherer: newMockContextGatherer(),
		TokenManager:    newMockTokenManager(),
		FileWriter:      newMockFileWriter(),
		AuditLogger:     newMockAuditLogger(),
		Logger:          newMockLogger(),
		RateLimiter:     ratelimit.NewRateLimiter(5, 60), // Default reasonable limits
		Config:          config.NewDefaultCliConfig(),
	}
}

// createOrchestrator creates an orchestrator with the current dependencies
func (d *orchestratorTestDeps) createOrchestrator() *Orchestrator {
	return NewOrchestrator(
		d.APIService,
		d.ContextGatherer,
		d.TokenManager,
		d.FileWriter,
		d.AuditLogger,
		d.RateLimiter,
		d.Config,
		d.Logger,
	)
}

// setupBasicContext sets up basic context files and stats
func (d *orchestratorTestDeps) setupBasicContext() {
	d.ContextGatherer.SetContextFiles([]fileutil.FileMeta{
		{Path: "file1.go", Content: "package main\nfunc main() {}"},
		{Path: "file2.go", Content: "package test\nfunc Test() {}"},
	})

	d.ContextGatherer.SetContextStats(&interfaces.ContextStats{
		ProcessedFilesCount: 2,
		CharCount:           50,
		LineCount:           4,
		TokenCount:          10,
	})
}

// setupGeminiClient sets up a basic mock Gemini client
func (d *orchestratorTestDeps) setupGeminiClient() *mockGeminiClient {
	client := &mockGeminiClient{}
	d.APIService.SetClient(client)
	return client
}

// setupMultiModelConfig sets up a configuration with multiple models
func (d *orchestratorTestDeps) setupMultiModelConfig(modelNames []string) {
	d.Config.ModelNames = modelNames
	d.Config.DryRun = false
	d.Config.OutputDir = "/tmp/test-output"
	d.Config.APIKey = "test-api-key"
}

// setupDryRunConfig sets up a configuration for dry run mode
func (d *orchestratorTestDeps) setupDryRunConfig() {
	d.Config.DryRun = true
	d.Config.ModelNames = []string{"model1"} // Not used in dry run, but needed for test validation
}

// runOrchestrator runs the orchestrator with the given instructions
func (d *orchestratorTestDeps) runOrchestrator(ctx context.Context, instructions string) error {
	orch := d.createOrchestrator()
	return orch.Run(ctx, instructions)
}

// verifyBasicWorkflow verifies that the basic workflow was executed correctly
func (d *orchestratorTestDeps) verifyBasicWorkflow(t *testing.T, expectedModelNames []string) {
	// Verify that context gathering was called
	if !d.ContextGatherer.gatherContextCalled {
		t.Error("Expected GatherContext to be called")
	}

	// Verify that each model was initialized
	d.APIService.VerifyInitClientCalled(t, expectedModelNames)

	// Verify that files were written (one per model)
	if d.FileWriter.GetSaveToFileCallCount() != len(expectedModelNames) {
		t.Errorf("Expected %d files to be written, got %d", len(expectedModelNames), d.FileWriter.GetSaveToFileCallCount())
	}
}

// verifyDryRunWorkflow verifies that the dry run workflow was executed correctly
func (d *orchestratorTestDeps) verifyDryRunWorkflow(t *testing.T) {
	// Verify context gathering was called
	if !d.ContextGatherer.gatherContextCalled {
		t.Error("Expected GatherContext to be called")
	}

	// Verify dry run display was called
	if !d.ContextGatherer.displayDryRunCalled {
		t.Error("Expected DisplayDryRunInfo to be called")
	}

	// Verify no API client was initialized (short-circuit in dry run)
	if d.APIService.GetInitClientCallCount() > 0 {
		t.Error("Expected no API client initialization in dry run mode")
	}

	// Verify no files were written
	if d.FileWriter.GetSaveToFileCallCount() > 0 {
		t.Error("Expected no files to be written in dry run mode")
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

// mockAPIService implements architect.APIService with enhanced tracking for integration tests
type mockAPIService struct {
	client          gemini.Client
	initClientErr   error
	initClientDelay bool
	processErr      error

	// Tracking fields for integration testing
	initClientCalls   []initClientCall
	processResponses  []processResponseCall
	mu                sync.Mutex       // Protects the tracking fields for concurrent calls
	errorResponses    map[string]error // Model-specific errors
	isEmptyResponseFn func(err error) bool
	isSafetyBlockedFn func(err error) bool
	getErrorDetailsFn func(err error) string
}

// initClientCall tracks a call to InitClient
type initClientCall struct {
	apiKey      string
	modelName   string
	apiEndpoint string
}

// processResponseCall tracks a call to ProcessResponse
type processResponseCall struct {
	result *gemini.GenerationResult
}

func newMockAPIService() *mockAPIService {
	mock := &mockAPIService{
		errorResponses: make(map[string]error),
		// Initialize with a default client to avoid nil pointer dereference
		client: &mockGeminiClient{},
	}

	// Set default handler functions
	mock.isEmptyResponseFn = func(err error) bool { return false }
	mock.isSafetyBlockedFn = func(err error) bool { return false }
	mock.getErrorDetailsFn = func(err error) string { return "error details" }

	return mock
}

// SetModelError configures a specific error for a model
func (m *mockAPIService) SetModelError(modelName string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorResponses[modelName] = err
}

// SetClient sets the mock client to return
func (m *mockAPIService) SetClient(client gemini.Client) {
	m.client = client
}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	m.mu.Lock()
	m.initClientCalls = append(m.initClientCalls, initClientCall{
		apiKey:      apiKey,
		modelName:   modelName,
		apiEndpoint: apiEndpoint,
	})

	// Check for model-specific error
	if err, ok := m.errorResponses[modelName]; ok {
		m.mu.Unlock()
		return nil, err
	}
	m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processResponses = append(m.processResponses, processResponseCall{
		result: result,
	})

	if m.processErr != nil {
		return "", m.processErr
	}
	return result.Content, nil
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return m.isEmptyResponseFn(err)
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return m.isSafetyBlockedFn(err)
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return m.getErrorDetailsFn(err)
}

// GetInitClientCallCount returns the number of InitClient calls
func (m *mockAPIService) GetInitClientCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.initClientCalls)
}

// GetProcessResponseCallCount returns the number of ProcessResponse calls
func (m *mockAPIService) GetProcessResponseCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.processResponses)
}

// VerifyInitClientCalled checks if InitClient was called with the expected model names
func (m *mockAPIService) VerifyInitClientCalled(t *testing.T, expectedModelNames []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a map for O(1) lookup
	modelMap := make(map[string]bool)
	for _, call := range m.initClientCalls {
		modelMap[call.modelName] = true
	}

	// Verify all expected models were called
	for _, expectedModel := range expectedModelNames {
		if !modelMap[expectedModel] {
			t.Errorf("Expected InitClient to be called for model %s, but it wasn't", expectedModel)
		}
	}
}

// mockContextGatherer implements interfaces.ContextGatherer with enhanced tracking
type mockContextGatherer struct {
	contextFiles        []fileutil.FileMeta
	contextStats        *interfaces.ContextStats
	gatherContextError  error
	gatherContextCalled bool
	displayDryRunCalled bool

	// Enhanced tracking for integration tests
	lastConfig  interfaces.GatherConfig
	dryRunStats *interfaces.ContextStats
	dryRunError error
	mu          sync.Mutex // Protects the tracking fields
}

func newMockContextGatherer() *mockContextGatherer {
	return &mockContextGatherer{
		contextFiles: []fileutil.FileMeta{},
		contextStats: &interfaces.ContextStats{},
	}
}

// SetContextFiles configures the mock to return specific context files
func (m *mockContextGatherer) SetContextFiles(files []fileutil.FileMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contextFiles = files
}

// SetContextStats configures the mock to return specific context stats
func (m *mockContextGatherer) SetContextStats(stats *interfaces.ContextStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contextStats = stats
}

// SetGatherContextError configures the mock to return a specific error
func (m *mockContextGatherer) SetGatherContextError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gatherContextError = err
}

// SetDryRunError configures the mock to return a specific error for DisplayDryRunInfo
func (m *mockContextGatherer) SetDryRunError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dryRunError = err
}

func (m *mockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gatherContextCalled = true
	m.lastConfig = config

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	if m.gatherContextError != nil {
		return nil, nil, m.gatherContextError
	}

	return m.contextFiles, m.contextStats, nil
}

func (m *mockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.displayDryRunCalled = true
	m.dryRunStats = stats

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return m.dryRunError
}

// VerifyGatherContextConfig verifies that GatherContext was called with the expected config
func (m *mockContextGatherer) VerifyGatherContextConfig(t *testing.T, expectedConfig interfaces.GatherConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.gatherContextCalled {
		t.Error("Expected GatherContext to be called, but it wasn't")
		return
	}

	// Verify key fields of the config
	if !reflect.DeepEqual(m.lastConfig.Paths, expectedConfig.Paths) {
		t.Errorf("Expected Paths %v, got %v", expectedConfig.Paths, m.lastConfig.Paths)
	}

	if !reflect.DeepEqual(m.lastConfig.Include, expectedConfig.Include) {
		t.Errorf("Expected Include %v, got %v", expectedConfig.Include, m.lastConfig.Include)
	}

	if !reflect.DeepEqual(m.lastConfig.Exclude, expectedConfig.Exclude) {
		t.Errorf("Expected Exclude %v, got %v", expectedConfig.Exclude, m.lastConfig.Exclude)
	}

	if !reflect.DeepEqual(m.lastConfig.ExcludeNames, expectedConfig.ExcludeNames) {
		t.Errorf("Expected ExcludeNames %v, got %v", expectedConfig.ExcludeNames, m.lastConfig.ExcludeNames)
	}
}

// mockTokenManager implements interfaces.TokenManager with enhanced tracking
type mockTokenManager struct {
	// Base behavior configuration
	tokenInfo       *interfaces.TokenResult
	tokenInfoError  error
	tokenLimitError error
	confirmResponse bool

	// Enhanced tracking for integration tests
	getTokenInfoCalls []getTokenInfoCall
	checkTokenCalls   []checkTokenCall
	confirmationCalls []confirmationCall
	mu                sync.Mutex // Protects the tracking fields
}

// getTokenInfoCall tracks a call to GetTokenInfo
type getTokenInfoCall struct {
	prompt string
}

// checkTokenCall tracks a call to CheckTokenLimit
type checkTokenCall struct {
	prompt string
}

// confirmationCall tracks a call to PromptForConfirmation
type confirmationCall struct {
	tokenCount int32
	threshold  int
}

func newMockTokenManager() *mockTokenManager {
	return &mockTokenManager{
		tokenInfo: &interfaces.TokenResult{
			TokenCount:   100,
			InputLimit:   1000,
			ExceedsLimit: false,
			LimitError:   "",
			Percentage:   10.0,
		},
		confirmResponse:   true,
		getTokenInfoCalls: []getTokenInfoCall{},
		checkTokenCalls:   []checkTokenCall{},
		confirmationCalls: []confirmationCall{},
	}
}

// SetTokenInfo configures the mock to return specific token info
func (m *mockTokenManager) SetTokenInfo(info *interfaces.TokenResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenInfo = info
}

// SetTokenInfoError configures the mock to return a specific error for GetTokenInfo
func (m *mockTokenManager) SetTokenInfoError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenInfoError = err
}

// SetTokenLimitError configures the mock to return a specific error for CheckTokenLimit
func (m *mockTokenManager) SetTokenLimitError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenLimitError = err
}

// SetConfirmResponse configures the mock to return a specific response for PromptForConfirmation
func (m *mockTokenManager) SetConfirmResponse(confirm bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.confirmResponse = confirm
}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*interfaces.TokenResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getTokenInfoCalls = append(m.getTokenInfoCalls, getTokenInfoCall{
		prompt: prompt,
	})

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if m.tokenInfoError != nil {
		return nil, m.tokenInfoError
	}

	return m.tokenInfo, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkTokenCalls = append(m.checkTokenCalls, checkTokenCall{
		prompt: prompt,
	})

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return m.tokenLimitError
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.confirmationCalls = append(m.confirmationCalls, confirmationCall{
		tokenCount: tokenCount,
		threshold:  threshold,
	})

	return m.confirmResponse
}

// GetTokenInfoCallCount returns the number of GetTokenInfo calls
func (m *mockTokenManager) GetTokenInfoCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.getTokenInfoCalls)
}

// GetCheckTokenCallCount returns the number of CheckTokenLimit calls
func (m *mockTokenManager) GetCheckTokenCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.checkTokenCalls)
}

// GetConfirmationCallCount returns the number of PromptForConfirmation calls
func (m *mockTokenManager) GetConfirmationCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.confirmationCalls)
}

// mockFileWriter implements interfaces.FileWriter with enhanced tracking
type mockFileWriter struct {
	// Base behavior configuration
	saveToFileError error

	// Enhanced tracking for integration tests
	saveToFileCalls []saveToFileCall
	mu              sync.Mutex // Protects the tracking fields
}

// saveToFileCall tracks a call to SaveToFile
type saveToFileCall struct {
	content    string
	outputFile string
}

func newMockFileWriter() *mockFileWriter {
	return &mockFileWriter{
		saveToFileCalls: []saveToFileCall{},
	}
}

// SetSaveToFileError configures the mock to return a specific error for SaveToFile
func (m *mockFileWriter) SetSaveToFileError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveToFileError = err
}

func (m *mockFileWriter) SaveToFile(content, outputFile string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.saveToFileCalls = append(m.saveToFileCalls, saveToFileCall{
		content:    content,
		outputFile: outputFile,
	})

	return m.saveToFileError
}

// GetSaveToFileCallCount returns the number of SaveToFile calls
func (m *mockFileWriter) GetSaveToFileCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.saveToFileCalls)
}

// VerifySaveToFileCalled checks if SaveToFile was called with the expected output file
func (m *mockFileWriter) VerifySaveToFileCalled(t *testing.T, expectedOutputFile string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.saveToFileCalls {
		if call.outputFile == expectedOutputFile {
			return
		}
	}

	t.Errorf("Expected SaveToFile to be called with output file %s, but it wasn't", expectedOutputFile)
}

// GetSavedContent returns the content that was saved to a specific output file
func (m *mockFileWriter) GetSavedContent(outputFile string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.saveToFileCalls {
		if call.outputFile == outputFile {
			return call.content, true
		}
	}

	return "", false
}

// mockAuditLogger implements auditlog.AuditLogger with enhanced tracking
type mockAuditLogger struct {
	// Base behavior configuration
	logError   error
	closeError error

	// Enhanced tracking for integration tests
	entries []auditlog.AuditEntry
	closed  bool
	mu      sync.Mutex // Protects the tracking fields
}

func newMockAuditLogger() *mockAuditLogger {
	return &mockAuditLogger{
		entries: []auditlog.AuditEntry{},
	}
}

// SetLogError configures the mock to return a specific error for Log
func (m *mockAuditLogger) SetLogError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logError = err
}

// SetCloseError configures the mock to return a specific error for Close
func (m *mockAuditLogger) SetCloseError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeError = err
}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = append(m.entries, entry)
	return m.logError
}

func (m *mockAuditLogger) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return m.closeError
}

// GetLogEntryCount returns the number of log entries
func (m *mockAuditLogger) GetLogEntryCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

// VerifyOperationLogged checks if a specific operation was logged
func (m *mockAuditLogger) VerifyOperationLogged(t *testing.T, operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, entry := range m.entries {
		if entry.Operation == operation {
			return
		}
	}

	t.Errorf("Expected operation %s to be logged, but it wasn't", operation)
}

// GetLogEntriesByOperation returns all log entries for a specific operation
func (m *mockAuditLogger) GetLogEntriesByOperation(operation string) []auditlog.AuditEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []auditlog.AuditEntry
	for _, entry := range m.entries {
		if entry.Operation == operation {
			result = append(result, entry)
		}
	}

	return result
}

// WasClosed returns whether Close was called
func (m *mockAuditLogger) WasClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// mockLogger implements logutil.LoggerInterface with enhanced tracking
type mockLogger struct {
	// Enhanced tracking for integration tests
	debugMessages   []string
	infoMessages    []string
	warnMessages    []string
	errorMessages   []string
	fatalMessages   []string
	printlnMessages []string
	printfMessages  []string
	mu              sync.Mutex // Protects the tracking fields
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages:   []string{},
		infoMessages:    []string{},
		warnMessages:    []string{},
		errorMessages:   []string{},
		fatalMessages:   []string{},
		printlnMessages: []string{},
		printfMessages:  []string{},
	}
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Warn(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Fatal(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fatalMessages = append(m.fatalMessages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Println(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.printlnMessages = append(m.printlnMessages, fmt.Sprint(args...))
}

func (m *mockLogger) Printf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.printfMessages = append(m.printfMessages, fmt.Sprintf(format, args...))
}

// GetDebugMessages returns all debug messages
func (m *mockLogger) GetDebugMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.debugMessages
}

// GetInfoMessages returns all info messages
func (m *mockLogger) GetInfoMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.infoMessages
}

// GetWarnMessages returns all warn messages
func (m *mockLogger) GetWarnMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.warnMessages
}

// GetErrorMessages returns all error messages
func (m *mockLogger) GetErrorMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errorMessages
}

// VerifyLogContains checks if any log message contains the specified text
func (m *mockLogger) VerifyLogContains(t *testing.T, text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, msg := range m.debugMessages {
		if strings.Contains(msg, text) {
			return
		}
	}

	for _, msg := range m.infoMessages {
		if strings.Contains(msg, text) {
			return
		}
	}

	for _, msg := range m.warnMessages {
		if strings.Contains(msg, text) {
			return
		}
	}

	for _, msg := range m.errorMessages {
		if strings.Contains(msg, text) {
			return
		}
	}

	for _, msg := range m.fatalMessages {
		if strings.Contains(msg, text) {
			return
		}
	}

	t.Errorf("Expected log to contain text '%s', but it wasn't found", text)
}

// Additional Integration Tests

// TestIntegration_APIServiceAdapterPassthrough tests the APIServiceAdapter passes through to the underlying APIService
func TestIntegration_APIServiceAdapterPassthrough(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()

	// Create a client and set specific return values
	mockClient := &mockGeminiClient{}
	deps.APIService.SetClient(mockClient)

	// Create an adapter
	adapter := &APIServiceAdapter{APIService: deps.APIService}

	// Test InitClient
	client, err := adapter.InitClient(context.Background(), "test-api-key", "test-model", "")
	if err != nil {
		t.Errorf("Expected no error from InitClient, got: %v", err)
	}
	if client != mockClient {
		t.Error("Expected adapter to return the client from the APIService")
	}

	// Test ProcessResponse
	result := &gemini.GenerationResult{Content: "test content"}
	content, err := adapter.ProcessResponse(result)
	if err != nil {
		t.Errorf("Expected no error from ProcessResponse, got: %v", err)
	}
	if content != result.Content {
		t.Errorf("Expected content to be %q, got %q", result.Content, content)
	}

	// Test error helper methods
	testErr := errors.New("test error")
	adapter.IsEmptyResponseError(testErr) // Just verifying pass-through
	adapter.IsSafetyBlockedError(testErr) // Just verifying pass-through

	details := adapter.GetErrorDetails(testErr)
	if details != "error details" {
		t.Errorf("Expected error details to be 'error details', got %q", details)
	}
}

// TestIntegration_FileWriterIntegration tests the integration with FileWriter
func TestIntegration_FileWriterIntegration(t *testing.T) {
	// Setup test dependencies
	deps := newTestDeps()
	deps.setupBasicContext()
	deps.setupGeminiClient()
	deps.setupMultiModelConfig([]string{"model1"})

	// Configure client settings

	// Configure client to return expected content
	client := &mockGeminiClient{}
	deps.APIService.SetClient(client)

	// Run orchestrator
	err := deps.runOrchestrator(context.Background(), "test instructions")

	// Verify no errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify file writing
	if deps.FileWriter.GetSaveToFileCallCount() != 1 {
		t.Errorf("Expected 1 file to be written, got %d", deps.FileWriter.GetSaveToFileCallCount())
	}

	// Verify output directory path is used
	// We use model name from config (model1) instead of hardcoded "test-model"
	outputPath := filepath.Join(deps.Config.OutputDir, "model1.md")
	deps.FileWriter.VerifySaveToFileCalled(t, outputPath)
}
