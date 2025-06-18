package thinktank_test

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
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

func TestNewOrchestrator(t *testing.T) {
	// Setup test dependencies
	apiService := &mockAPIService{}
	contextGatherer := &mockContextGatherer{}
	fileWriter := &mockFileWriter{}
	auditLogger := auditlog.NewNoOpAuditLogger()
	rateLimiter := ratelimit.NewRateLimiter(5, 60) // 5 concurrent, 60 per minute
	config := &config.CliConfig{
		ModelNames:       []string{"test-model"},
		InstructionsFile: "test instructions",
		OutputDir:        "/tmp/test",
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
	orchestrator := thinktank.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		config,
		logger,
		consoleWriter,
	)

	// Verify orchestrator was created successfully
	if orchestrator == nil {
		t.Fatal("NewOrchestrator returned nil")
	}

	// NewOrchestrator already returns Orchestrator type, no need to assert
}

func TestNewOrchestratorWithNilDependencies(t *testing.T) {
	// Test with various nil dependencies to ensure no panics
	testCases := []struct {
		name            string
		apiService      interfaces.APIService
		contextGatherer interfaces.ContextGatherer
		fileWriter      interfaces.FileWriter
		auditLogger     auditlog.AuditLogger
		rateLimiter     *ratelimit.RateLimiter
		config          *config.CliConfig
		logger          logutil.LoggerInterface
		expectPanic     bool
	}{
		{
			name:            "all valid dependencies",
			apiService:      &mockAPIService{},
			contextGatherer: &mockContextGatherer{},
			fileWriter:      &mockFileWriter{},
			auditLogger:     auditlog.NewNoOpAuditLogger(),
			rateLimiter:     ratelimit.NewRateLimiter(5, 60),
			config:          &config.CliConfig{},
			logger:          logutil.NewLogger(logutil.InfoLevel, nil, "[test] "),
			expectPanic:     false,
		},
		{
			name:            "nil logger",
			apiService:      &mockAPIService{},
			contextGatherer: &mockContextGatherer{},
			fileWriter:      &mockFileWriter{},
			auditLogger:     auditlog.NewNoOpAuditLogger(),
			rateLimiter:     ratelimit.NewRateLimiter(5, 60),
			config:          &config.CliConfig{},
			logger:          nil,
			expectPanic:     false, // Should handle nil gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.expectPanic {
						t.Fatalf("NewOrchestrator panicked unexpectedly: %v", r)
					}
				} else if tc.expectPanic {
					t.Fatal("NewOrchestrator was expected to panic but didn't")
				}
			}()

			consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return false }, // CI mode for tests
			})
			orchestrator := thinktank.NewOrchestrator(
				tc.apiService,
				tc.contextGatherer,
				tc.fileWriter,
				tc.auditLogger,
				tc.rateLimiter,
				tc.config,
				tc.logger,
				consoleWriter,
			)

			if !tc.expectPanic && orchestrator == nil {
				t.Fatal("NewOrchestrator returned nil when it shouldn't have")
			}
		})
	}
}
