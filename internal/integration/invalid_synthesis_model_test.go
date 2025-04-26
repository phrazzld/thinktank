// internal/integration/invalid_synthesis_model_test.go
package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// TestInvalidSynthesisModel tests that validation correctly rejects an invalid synthesis model
func TestInvalidSynthesisModel(t *testing.T) {
	// Create logger for the test
	logger := logutil.NewTestLogger(t)

	// Create temp directory for outputs
	tempDir, err := os.MkdirTemp("", "thinktank-invalid-synthesis-test")
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
	instructions := "Test instructions for synthesis"

	// Set up multiple model names and an invalid synthesis model
	modelNames := []string{"model1", "model2"}
	invalidSynthesisModel := "non-existent-model"

	// Parse log level
	logLevel, _ := logutil.ParseLogLevel("debug")

	// Create the CLI config with invalid synthesis model
	cfg := &config.CliConfig{
		ModelNames:                 modelNames,
		OutputDir:                  outputDir,
		Verbose:                    true,
		LogLevel:                   logLevel,
		Format:                     "markdown",
		AuditLogFile:               filepath.Join(tempDir, "audit.log"),
		MaxConcurrentRequests:      2,
		RateLimitRequestsPerMinute: 60,
		SynthesisModel:             invalidSynthesisModel, // Set invalid synthesis model
	}

	// Track whether files were written
	filesWritten := make(map[string]bool)
	var filesMutex sync.Mutex

	// Create mock API service that returns error for the invalid model
	apiService := &MockAPIService{
		InitLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			// All regular models should initialize successfully
			if modelName == "model1" || modelName == "model2" {
				return &llm.MockLLMClient{
					GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
						return &llm.ProviderResult{
							Content:      "Output from " + modelName,
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

			// Invalid synthesis model should fail to initialize
			if modelName == invalidSynthesisModel {
				return nil, fmt.Errorf("model '%s' not found or not supported", modelName)
			}

			return nil, errors.New("unexpected model")
		},
		GetModelDefinitionFunc: func(modelName string) (*registry.ModelDefinition, error) {
			// Return error for invalid synthesis model
			if modelName == invalidSynthesisModel {
				return nil, fmt.Errorf("model '%s' not found in registry", modelName)
			}

			// Return definition for valid models
			return &registry.ModelDefinition{
				Name:     modelName,
				Provider: "test-provider",
			}, nil
		},
		GetModelParametersFunc: func(modelName string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
		ProcessLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
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

	// Create file writer that tracks written files
	fileWriter := &MockFileWriter{
		SaveToFileFunc: func(content, filePath string) error {
			// Record that this file was written with mutex protection
			filesMutex.Lock()
			filesWritten[filePath] = true
			filesMutex.Unlock()

			// Actually save the file to test file existence later
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

	// Execute the orchestrator - expect it to fail due to invalid synthesis model
	err = orch.Run(context.Background(), instructions)

	// Verify that the orchestrator returned an error referencing the invalid model
	if err == nil {
		t.Errorf("Expected error due to invalid synthesis model, but got nil")
	} else {
		// Check that the error message mentions the invalid model
		if !containsString(err.Error(), invalidSynthesisModel) {
			t.Errorf("Expected error to mention invalid model '%s', but got: %v",
				invalidSynthesisModel, err)
		} else {
			t.Logf("Received expected error for invalid synthesis model: %v", err)
		}
	}

	// Verify that individual model output files were written
	// The orchestrator processes regular models first, then attempts synthesis
	for _, modelName := range modelNames {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")
		_, err := os.Stat(expectedFilePath)
		if os.IsNotExist(err) {
			t.Errorf("Expected output file %s not created", expectedFilePath)
		} else {
			t.Logf("Verified regular model output file exists: %s", expectedFilePath)
		}
	}

	// Verify that synthesis output file was NOT created
	synthesisFilePath := filepath.Join(outputDir, invalidSynthesisModel+"-synthesis.md")
	_, err = os.Stat(synthesisFilePath)
	if !os.IsNotExist(err) {
		t.Errorf("Synthesis output file %s was unexpectedly created", synthesisFilePath)
	} else {
		t.Logf("Verified synthesis output file was not created, as expected")
	}
}

// containsString is a helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return s != "" && substr != "" && len(s) > 0 && len(substr) > 0 && s != substr && strings.Contains(s, substr)
}
