// internal/integration/no_synthesis_flow_test.go
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/orchestrator"
)

// TestNoSynthesisFlow tests the complete flow without synthesis model specified
// This test verifies that multiple output files are created, one for each model
func TestNoSynthesisFlow(t *testing.T) {
	// Create logger for the test
	logger := logutil.NewTestLogger(t)

	// Create temp directory for outputs
	tempDir, err := os.MkdirTemp("", "thinktank-nosynthesis-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}()

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Set up test instructions
	instructions := "Test instructions for multiple models"

	// Set up multiple model names
	modelNames := []string{"model1", "model2", "model3"}

	// Parse log level
	logLevel, _ := logutil.ParseLogLevel("debug")

	// Create the CLI config
	cfg := &config.CliConfig{
		ModelNames:                 modelNames,
		OutputDir:                  outputDir,
		Verbose:                    true,
		LogLevel:                   logLevel,
		Format:                     "markdown",
		AuditLogFile:               filepath.Join(tempDir, "audit.log"),
		MaxConcurrentRequests:      2,
		RateLimitRequestsPerMinute: 60,
		SynthesisModel:             "", // Explicitly set to empty to ensure no synthesis
	}

	// Create test adapter with mock content for each model
	mockOutputs := map[string]string{
		"model1": "# Output from Model 1\n\nThis is test output from model1.",
		"model2": "# Output from Model 2\n\nThis is test output from model2.",
		"model3": "# Output from Model 3\n\nThis is test output from model3.",
	}

	// Create mock API service
	apiService := &MockAPIService{
		ProcessLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			// Extract model name from the result and return the mock output
			return mockOutputs[result.Content], nil
		},
		InitLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			// Return a mock client that returns the model name as the result
			return &llm.MockLLMClient{
				GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					// Store the model name in the content field so we can identify it in ProcessLLMResponse
					return &llm.ProviderResult{
						Content:      modelName,
						FinishReason: "stop",
					}, nil
				},
				GetModelNameFunc: func() string {
					return modelName
				},
			}, nil
		},
		GetModelParametersFunc: func(ctx context.Context, modelName string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	// Create context gatherer
	contextGatherer := &MockContextGatherer{
		GatherContextFunc: func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
			// Return empty context files for this test
			return []fileutil.FileMeta{}, &interfaces.ContextStats{
				ProcessedFilesCount: 0,
				CharCount:           0,
			}, nil
		},
		DisplayDryRunInfoFunc: func(ctx context.Context, stats *interfaces.ContextStats) error {
			return nil
		},
	}

	// Create file writer
	fileWriter := &MockFileWriter{
		SaveToFileFunc: func(ctx context.Context, content, filePath string) error {
			// Actually save the files to verify they exist later
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return os.WriteFile(filePath, []byte(content), 0640)
		},
	}

	// Create audit logger
	auditLogger := &MockAuditLogger{
		LogFunc: func(ctx context.Context, entry auditlog.AuditEntry) error {
			return nil
		},
		LogLegacyFunc: func(entry auditlog.AuditEntry) error {
			return nil
		},
		CloseFunc: func() error {
			return nil
		},
	}

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(cfg.MaxConcurrentRequests, cfg.RateLimitRequestsPerMinute)

	// Create orchestrator
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	orch := orchestrator.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		cfg,
		logger,
		consoleWriter,
	)

	// Execute the orchestrator
	err = orch.Run(context.Background(), instructions)
	if err != nil {
		t.Fatalf("Orchestrator.Run failed: %v", err)
	}

	// Verify that multiple output files were created
	for _, modelName := range modelNames {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")
		if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s not created", expectedFilePath)
		} else {
			// Verify file content
			content, err := os.ReadFile(expectedFilePath)
			if err != nil {
				t.Errorf("Failed to read output file %s: %v", expectedFilePath, err)
			} else {
				expectedContent := mockOutputs[modelName]
				if string(content) != expectedContent {
					t.Errorf("File content mismatch for %s:\nExpected: %s\nActual: %s",
						modelName, expectedContent, string(content))
				} else {
					t.Logf("Verified content for model %s", modelName)
				}
			}
		}
	}
}

// Mock implementations moved to integration_test_mocks.go
