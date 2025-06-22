// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// TestMockAPIService provides a simple mock for APIService interface
type TestMockAPIService struct{}

func (m *TestMockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return &TestMockLLMClient{}, nil
}
func (m *TestMockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if result == nil {
		return "", nil
	}
	return result.Content, nil
}
func (m *TestMockAPIService) IsEmptyResponseError(err error) bool { return false }
func (m *TestMockAPIService) IsSafetyBlockedError(err error) bool { return false }
func (m *TestMockAPIService) GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
func (m *TestMockAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *TestMockAPIService) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	return &models.ModelInfo{}, nil
}
func (m *TestMockAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (int32, int32, error) {
	return 8000, 1000, nil
}
func (m *TestMockAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	return true, nil
}

// TestMockLLMClient provides a simple mock for LLMClient interface
type TestMockLLMClient struct{}

func (m *TestMockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:      "Mock content",
		FinishReason: "mock",
	}, nil
}
func (m *TestMockLLMClient) GetModelName() string { return "mock-model" }
func (m *TestMockLLMClient) Close() error         { return nil }

// TestMockConsoleWriter provides a simple mock for ConsoleWriter interface
type TestMockConsoleWriter struct {
	quiet      bool
	noProgress bool
}

func (m *TestMockConsoleWriter) StartProcessing(modelCount int)                             {}
func (m *TestMockConsoleWriter) ModelQueued(modelName string, index int)                    {}
func (m *TestMockConsoleWriter) ModelStarted(modelIndex, totalModels int, modelName string) {}
func (m *TestMockConsoleWriter) ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration) {
}
func (m *TestMockConsoleWriter) ModelFailed(modelIndex, totalModels int, modelName string, reason string) {
}
func (m *TestMockConsoleWriter) ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration) {
}
func (m *TestMockConsoleWriter) ShowProcessingLine(modelName string)                  {}
func (m *TestMockConsoleWriter) UpdateProcessingLine(modelName string, status string) {}
func (m *TestMockConsoleWriter) ShowFileOperations(message string)                    {}
func (m *TestMockConsoleWriter) ShowSummarySection(summary logutil.SummaryData)       {}
func (m *TestMockConsoleWriter) ShowOutputFiles(files []logutil.OutputFile)           {}
func (m *TestMockConsoleWriter) ShowFailedModels(failed []logutil.FailedModel)        {}
func (m *TestMockConsoleWriter) SynthesisStarted()                                    {}
func (m *TestMockConsoleWriter) SynthesisCompleted(outputPath string)                 {}
func (m *TestMockConsoleWriter) StatusMessage(message string)                         {}
func (m *TestMockConsoleWriter) SetQuiet(quiet bool)                                  { m.quiet = quiet }
func (m *TestMockConsoleWriter) SetNoProgress(noProgress bool)                        { m.noProgress = noProgress }
func (m *TestMockConsoleWriter) IsInteractive() bool                                  { return false }
func (m *TestMockConsoleWriter) GetTerminalWidth() int                                { return 80 }
func (m *TestMockConsoleWriter) FormatMessage(message string) string                  { return message }
func (m *TestMockConsoleWriter) ErrorMessage(message string)                          {}
func (m *TestMockConsoleWriter) WarningMessage(message string)                        {}
func (m *TestMockConsoleWriter) SuccessMessage(message string)                        {}

// TestRunDryRunSuccess tests the Run() function with dry-run mode enabled
// This replaces the subprocess test "TestMainDryRun/main_dry_run_success"
func TestRunDryRunSuccess(t *testing.T) {
	// Create test context
	ctx := context.Background()
	ctx = logutil.WithCorrelationID(ctx, "test-correlation-id")

	// Create mock dependencies
	mockLogger := testutil.NewMockLogger()
	mockAuditLogger := auditlog.NewNoOpAuditLogger()
	mockAPIService := &TestMockAPIService{}
	mockConsoleWriter := &TestMockConsoleWriter{}
	mockFileSystem := NewMockFileSystem()
	mockExitHandler := NewMockExitHandler()

	// Create a real temporary instructions file (since thinktank.Execute reads it directly)
	tmpFile, err := os.CreateTemp("", "test_instructions_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary instructions file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	instructionsContent := "Test instructions for the LLM dry run"
	if _, err := tmpFile.WriteString(instructionsContent); err != nil {
		t.Fatalf("Failed to write instructions content: %v", err)
	}

	// Create test configuration for dry-run mode
	config := &config.CliConfig{
		DryRun:           true,
		InstructionsFile: tmpFile.Name(),
		OutputDir:        "./test_output",
		Paths:            []string{"test.go"},
		ModelNames:       []string{"gemini-2.5-pro"},
		Timeout:          10 * time.Minute,
		LogLevel:         logutil.InfoLevel,
	}

	// Create RunConfig with injected dependencies
	runConfig := &RunConfig{
		Context:       ctx,
		Config:        config,
		Logger:        mockLogger,
		AuditLogger:   mockAuditLogger,
		APIService:    mockAPIService,
		ConsoleWriter: mockConsoleWriter,
		FileSystem:    mockFileSystem,
		ExitHandler:   mockExitHandler,
	}

	// Execute the Run function
	result := Run(runConfig)

	// Verify the result
	assert.Equal(t, ExitCodeSuccess, result.ExitCode, "Expected success exit code for dry-run")
	assert.NoError(t, result.Error, "Expected no error for dry-run")
	assert.NotNil(t, result.Stats, "Expected stats to be populated")

	// Verify mock interactions
	assert.False(t, mockExitHandler.WasCalled(), "Exit handler should not be called for successful dry-run")

	// Verify logging occurred
	logEntries := mockLogger.GetLogEntries()
	assert.True(t, len(logEntries) > 0, "Expected log entries to be written")
}
