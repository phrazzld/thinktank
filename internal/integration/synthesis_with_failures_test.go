// internal/integration/synthesis_with_failures_test.go
package integration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/orchestrator"
)

// TestSynthesisWithModelFailuresFlow tests the synthesis flow with one or more model failures
// This test verifies that synthesis proceeds with available outputs when some models fail
// but at least one model succeeds
func TestSynthesisWithModelFailuresFlow(t *testing.T) {
	// Create logger for the test
	logger := logutil.NewTestLogger(t)

	// Create filesystem abstraction for testing
	fs := testutil.NewRealFS()

	// Create temp directory for outputs
	tempDir, err := os.MkdirTemp("", "thinktank-synthesis-failures-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := fs.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}()

	outputDir := filepath.Join(tempDir, "output")
	if err := fs.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Set up test instructions
	instructions := "Test instructions for synthesis with failures"

	// Set up multiple model names and synthesis model
	// We'll make model2 fail
	modelNames := []string{"model1", "model2", "model3"}
	failingModel := "model2"
	synthesisModel := "synthesis-model"

	// Parse log level
	logLevel, _ := logutil.ParseLogLevel("debug")

	// Create the CLI config with synthesis model
	cfg := &config.CliConfig{
		ModelNames:                 modelNames,
		OutputDir:                  outputDir,
		Verbose:                    true,
		LogLevel:                   logLevel,
		Format:                     "markdown",
		AuditLogFile:               filepath.Join(tempDir, "audit.log"),
		MaxConcurrentRequests:      2,
		RateLimitRequestsPerMinute: 60,
		SynthesisModel:             synthesisModel, // Set synthesis model
	}

	// Create test adapter with mock content for successful models
	mockOutputs := map[string]string{
		"model1": "# Output from Model 1\n\nThis is test output from model1.",
		"model3": "# Output from Model 3\n\nThis is test output from model3.",
	}

	// Expected synthesis output - should only contain model1 and model3 content
	synthesisOutput := "# Synthesized Output\n\nThis content combines insights from successful models only."

	// Track which models were called and which ones succeeded or failed
	calledModels := make(map[string]bool)
	modelSucceeded := make(map[string]bool)
	modelFailed := make(map[string]bool)
	var modelsMutex sync.Mutex

	// Create mock API service
	apiService := &MockAPIService{
		ProcessLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			// For the synthesis model, we'll detect it from a specific content marker
			if result.Content == "synthesized content" {
				return synthesisOutput, nil
			}
			// For regular models, extract model name from the result and return mock output
			return mockOutputs[result.Content], nil
		},
		InitLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			// Add model to called models with mutex protection
			modelsMutex.Lock()
			calledModels[modelName] = true
			modelsMutex.Unlock()

			// Return appropriate mock client based on whether it's a synthesis model or one of the regular models
			switch modelName {
			case synthesisModel:
				return &llm.MockLLMClient{
					GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
						// Synthesis model should return synthesized content
						return &llm.ProviderResult{
							Content:      "synthesized content",
							FinishReason: "stop",
						}, nil
					},
					GetModelNameFunc: func() string {
						return modelName
					},
					CloseFunc: func() error {
						return nil
					},
				}, nil
			case failingModel:
				// Failing model client
				modelsMutex.Lock()
				modelFailed[modelName] = true
				modelsMutex.Unlock()
				return &llm.MockLLMClient{
					GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
						// Return an error for the failing model
						return nil, errors.New("model service unavailable: test simulated failure")
					},
					GetModelNameFunc: func() string {
						return modelName
					},
					CloseFunc: func() error {
						return nil
					},
				}, nil
			default:
				// Regular working model client
				modelsMutex.Lock()
				modelSucceeded[modelName] = true
				modelsMutex.Unlock()
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
					CloseFunc: func() error {
						return nil
					},
				}, nil
			}
		},
		GetModelParametersFunc: func(ctx context.Context, modelName string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
		GetModelDefinitionFunc: func(ctx context.Context, modelName string) (*registry.ModelDefinition, error) {
			return &registry.ModelDefinition{
				Name:     modelName,
				Provider: "test-provider",
			}, nil
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

	// Create file writer that records what files are written
	savedFiles := make(map[string]string)
	var filesMutex sync.Mutex
	fileWriter := &MockFileWriter{
		SaveToFileFunc: func(content, filePath string) error {
			// Store the file content for verification with mutex protection
			filesMutex.Lock()
			savedFiles[filePath] = content
			filesMutex.Unlock()

			// Actually save the files to verify they exist later
			dir := filepath.Dir(filePath)
			if err := fs.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return fs.WriteFile(filePath, []byte(content), 0640)
		},
	}

	// Track audit log entries for verification
	var auditEntries []auditlog.AuditEntry
	var auditMutex sync.Mutex

	// Create audit logger
	auditLogger := &MockAuditLogger{
		LogFunc: func(ctx context.Context, entry auditlog.AuditEntry) error {
			auditMutex.Lock()
			auditEntries = append(auditEntries, entry)
			auditMutex.Unlock()
			return nil
		},
		LogLegacyFunc: func(entry auditlog.AuditEntry) error {
			auditMutex.Lock()
			auditEntries = append(auditEntries, entry)
			auditMutex.Unlock()
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

	// Verify that all models were called (even the failing one)
	modelsMutex.Lock()
	for _, modelName := range modelNames {
		if !calledModels[modelName] {
			t.Errorf("Expected model %s to be called, but it wasn't", modelName)
		}
	}

	// Verify the specific model that was configured to fail actually failed
	if !modelFailed[failingModel] {
		t.Errorf("Expected model %s to fail but it didn't", failingModel)
	}

	// Verify the non-failing models succeeded
	// Note: The non-failing models should still be processed even if one fails
	for _, modelName := range modelNames {
		if modelName != failingModel {
			if !modelSucceeded[modelName] {
				t.Errorf("Expected model %s to succeed, but it failed or wasn't called", modelName)
			}
		}
	}
	modelsMutex.Unlock()

	// Verify files for successful models were created
	modelsMutex.Lock()
	for modelName := range modelSucceeded {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")
		exists, modelStatErr := fs.Stat(expectedFilePath)
		if !exists {
			t.Errorf("Expected output file %s for successful model not created", expectedFilePath)
		} else if modelStatErr != nil {
			t.Errorf("Error checking file %s: %v", expectedFilePath, modelStatErr)
		} else {
			// Verify file content
			content, modelReadErr := fs.ReadFile(expectedFilePath)
			if modelReadErr != nil {
				t.Errorf("Failed to read output file %s: %v", expectedFilePath, modelReadErr)
			} else {
				expectedContent := mockOutputs[modelName]
				if string(content) != expectedContent {
					t.Errorf("File content mismatch for %s:\nExpected: %s\nActual: %s",
						modelName, expectedContent, string(content))
				} else {
					t.Logf("Verified content for successful model %s", modelName)
				}
			}
		}
	}
	modelsMutex.Unlock()

	// Verify that the orchestrator returns a descriptive error when some models fail
	// But still processes the synthesis with successful outputs
	if err == nil {
		t.Errorf("Expected Orchestrator.Run to return an error detailing partial failures")
	} else if !errors.Is(err, orchestrator.ErrPartialProcessingFailure) {
		t.Errorf("Expected error to be ErrPartialProcessingFailure, but got: %v", err)
	} else {
		t.Logf("Orchestrator correctly returned partial failure error: %v", err)
	}

	// Verify that synthesis model WAS called despite some model failures
	// The new orchestrator implementation continues as long as at least one model succeeds
	modelsMutex.Lock()
	if !calledModels[synthesisModel] {
		t.Errorf("Expected synthesis model %s to be called despite partial model failures, but it wasn't", synthesisModel)
	} else {
		t.Logf("Synthesis model was correctly called despite partial model failures")
	}
	modelsMutex.Unlock()

	// Verify that synthesis output file WAS created despite some model failures
	expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")
	exists, statErr := fs.Stat(expectedSynthesisFile)
	if !exists {
		t.Errorf("Expected synthesis output file %s to be created despite partial model failures, but it wasn't", expectedSynthesisFile)
	} else if statErr != nil {
		t.Errorf("Error checking synthesis file: %v", statErr)
	} else {
		// Verify synthesis file content
		content, synthReadErr := fs.ReadFile(expectedSynthesisFile)
		if synthReadErr != nil {
			t.Errorf("Failed to read synthesis output file: %v", synthReadErr)
		} else if string(content) != synthesisOutput {
			t.Errorf("Synthesis content mismatch:\nExpected: %s\nActual: %s",
				synthesisOutput, string(content))
		} else {
			t.Logf("Verified correct synthesis content despite partial model failures")
		}
	}

	t.Logf("Test verified correct behavior: orchestrator continues with synthesis when some models succeed")
}
