package thinktank_test

import (
	"context"
	"os"
	"testing"

	"github.com/misty-step/thinktank/internal/auditlog"
	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/fileutil"
	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/models"
	"github.com/misty-step/thinktank/internal/ratelimit"
	"github.com/misty-step/thinktank/internal/thinktank"
	"github.com/misty-step/thinktank/internal/thinktank/interfaces"
)

// Mock implementations for testing

type mockAPIService struct{}

func (m *mockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return nil, nil
}

func (m *mockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	return true, nil
}

func (m *mockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	return &models.ModelInfo{}, nil
}

func (m *mockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	return 4000, 1000, nil
}

func (m *mockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return "test response", nil
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

type mockContextGatherer struct{}

func (m *mockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return []fileutil.FileMeta{}, &interfaces.ContextStats{}, nil
}

func (m *mockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return nil
}

type mockFileWriter struct{}

func (m *mockFileWriter) SaveToFile(ctx context.Context, content, outputPath string) error {
	return nil
}

type mockTokenCountingService struct{}

func (m *mockTokenCountingService) CountTokens(ctx context.Context, req interfaces.TokenCountingRequest) (interfaces.TokenCountingResult, error) {
	return interfaces.TokenCountingResult{TotalTokens: 100}, nil
}

func (m *mockTokenCountingService) CountTokensForModel(ctx context.Context, req interfaces.TokenCountingRequest, modelName string) (interfaces.ModelTokenCountingResult, error) {
	return interfaces.ModelTokenCountingResult{
		TokenCountingResult: interfaces.TokenCountingResult{TotalTokens: 100},
		ModelName:           modelName,
	}, nil
}

func (m *mockTokenCountingService) GetCompatibleModels(ctx context.Context, req interfaces.TokenCountingRequest, availableProviders []string) ([]interfaces.ModelCompatibility, error) {
	return []interfaces.ModelCompatibility{}, nil
}

func TestNewOrchestrator(t *testing.T) {
	// Create temporary directory for test isolation
	tempDir, err := os.MkdirTemp("", "orchestrator_test_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Setup test dependencies
	apiService := &mockAPIService{}
	contextGatherer := &mockContextGatherer{}
	fileWriter := &mockFileWriter{}
	auditLogger := auditlog.NewNoOpAuditLogger()
	rateLimiter := ratelimit.NewRateLimiter(5, 60) // 5 concurrent, 60 per minute
	config := &config.CliConfig{
		ModelNames:       []string{"test-model"},
		InstructionsFile: "test instructions",
		OutputDir:        tempDir,
		DryRun:           false,
		Verbose:          false,
		Include:          "",
		Exclude:          "",
		ExcludeNames:     "",
		Format:           "markdown",
		APIKey:           "test-key",
		APIEndpoint:      "",
		Paths:            []string{},
		LogLevel:         logutil.InfoLevel,
		AuditLogFile:     "",
		SynthesisModel:   "",
	}
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	// Test creating orchestrator
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	tokenCountingService := &mockTokenCountingService{}
	orchestrator := thinktank.NewOrchestrator(thinktank.OrchestratorDeps{
		APIService:           apiService,
		ContextGatherer:      contextGatherer,
		FileWriter:           fileWriter,
		AuditLogger:          auditLogger,
		RateLimiter:          rateLimiter,
		Config:               config,
		Logger:               logger,
		ConsoleWriter:        consoleWriter,
		TokenCountingService: tokenCountingService,
	})

	// Verify orchestrator was created successfully
	if orchestrator == nil {
		t.Fatal("NewOrchestrator returned nil")
	}

	// NewOrchestrator already returns Orchestrator type, no need to assert
}

// Note: Nil dependency validation tests are in orchestrator/orchestrator_deps_test.go
// which comprehensively tests that NewOrchestrator panics on nil dependencies.
