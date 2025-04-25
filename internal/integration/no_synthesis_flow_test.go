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
	"github.com/phrazzld/thinktank/internal/registry"
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
		GetModelParametersFunc: func(modelName string) (map[string]interface{}, error) {
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
		SaveToFileFunc: func(content, filePath string) error {
			// Actually save the files to verify they exist later
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return os.WriteFile(filePath, []byte(content), 0644)
		},
	}

	// Create audit logger
	auditLogger := &MockAuditLogger{
		LogFunc: func(entry auditlog.AuditEntry) error {
			return nil
		},
		CloseFunc: func() error {
			return nil
		},
	}

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(cfg.MaxConcurrentRequests, cfg.RateLimitRequestsPerMinute)

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		cfg,
		logger,
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

// MockAPIService is a mock implementation of the APIService interface
type MockAPIService struct {
	InitLLMClientFunc          func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	GetModelParametersFunc     func(modelName string) (map[string]interface{}, error)
	ValidateModelParameterFunc func(modelName, paramName string, value interface{}) (bool, error)
	GetModelDefinitionFunc     func(modelName string) (*registry.ModelDefinition, error)
	GetModelTokenLimitsFunc    func(modelName string) (contextWindow, maxOutputTokens int32, err error)
	ProcessLLMResponseFunc     func(result *llm.ProviderResult) (string, error)
	IsEmptyResponseErrorFunc   func(err error) bool
	IsSafetyBlockedErrorFunc   func(err error) bool
	GetErrorDetailsFunc        func(err error) string
}

func (m *MockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.InitLLMClientFunc != nil {
		return m.InitLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
	}
	return nil, nil
}

func (m *MockAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	if m.GetModelParametersFunc != nil {
		return m.GetModelParametersFunc(modelName)
	}
	return map[string]interface{}{}, nil
}

func (m *MockAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	if m.ValidateModelParameterFunc != nil {
		return m.ValidateModelParameterFunc(modelName, paramName, value)
	}
	return true, nil
}

func (m *MockAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	if m.GetModelDefinitionFunc != nil {
		return m.GetModelDefinitionFunc(modelName)
	}
	return nil, nil
}

func (m *MockAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if m.GetModelTokenLimitsFunc != nil {
		return m.GetModelTokenLimitsFunc(modelName)
	}
	return 0, 0, nil
}

func (m *MockAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.ProcessLLMResponseFunc != nil {
		return m.ProcessLLMResponseFunc(result)
	}
	return "", nil
}

func (m *MockAPIService) IsEmptyResponseError(err error) bool {
	if m.IsEmptyResponseErrorFunc != nil {
		return m.IsEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *MockAPIService) IsSafetyBlockedError(err error) bool {
	if m.IsSafetyBlockedErrorFunc != nil {
		return m.IsSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *MockAPIService) GetErrorDetails(err error) string {
	if m.GetErrorDetailsFunc != nil {
		return m.GetErrorDetailsFunc(err)
	}
	return ""
}

// MockContextGatherer is a mock implementation of the context gatherer
type MockContextGatherer struct {
	GatherContextFunc     func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error)
	DisplayDryRunInfoFunc func(ctx context.Context, stats *interfaces.ContextStats) error
}

// GatherContext implements the context gatherer interface
func (m *MockContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return m.GatherContextFunc(ctx, config)
}

// DisplayDryRunInfo implements the context gatherer interface
func (m *MockContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return m.DisplayDryRunInfoFunc(ctx, stats)
}

// MockFileWriter is a mock implementation of the file writer
type MockFileWriter struct {
	SaveToFileFunc func(content, filePath string) error
}

// SaveToFile implements the file writer interface
func (m *MockFileWriter) SaveToFile(content, filePath string) error {
	return m.SaveToFileFunc(content, filePath)
}

// MockAuditLogger is a mock implementation of the audit logger
type MockAuditLogger struct {
	LogFunc   func(entry auditlog.AuditEntry) error
	CloseFunc func() error
}

// Log implements the audit logger interface
func (m *MockAuditLogger) Log(entry auditlog.AuditEntry) error {
	if m.LogFunc != nil {
		return m.LogFunc(entry)
	}
	return nil
}

// Close implements the audit logger interface
func (m *MockAuditLogger) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
