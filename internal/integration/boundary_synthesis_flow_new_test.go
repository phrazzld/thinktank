// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// TestBoundarySynthesisFlowNew tests the complete flow with synthesis model using boundary mocks
// This test demonstrates the approach of mocking only external boundaries while using real
// internal implementations.
func TestBoundarySynthesisFlowNew(t *testing.T) {
	// Create temp directory for outputs
	tempDir, err := os.MkdirTemp("", "thinktank-synthesis-test")
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

	// Set up multiple model names and synthesis model
	modelNames := []string{"model1", "model2", "model3"}
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

	// Create logger
	logger := logutil.NewTestLogger(t)

	// Create mock API caller (boundary mock)
	mockAPICaller := &MockExternalAPICaller{}

	// Expected model outputs
	mockOutputs := map[string]string{
		"model1": "# Output from Model 1\n\nThis is test output from model1.",
		"model2": "# Output from Model 2\n\nThis is test output from model2.",
		"model3": "# Output from Model 3\n\nThis is test output from model3.",
	}

	// Expected synthesis output
	synthesisOutput := "# Synthesized Output\n\nThis content combines insights from all models."

	// Configure mock API caller to return appropriate responses
	mockAPICaller.CallLLMAPIFunc = func(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		// For synthesis model
		if modelName == synthesisModel {
			return &llm.ProviderResult{
				Content:      synthesisOutput,
				FinishReason: "stop",
			}, nil
		}

		// For regular models
		if content, ok := mockOutputs[modelName]; ok {
			return &llm.ProviderResult{
				Content:      content,
				FinishReason: "stop",
			}, nil
		}

		// Default response
		return &llm.ProviderResult{
			Content:      fmt.Sprintf("Default response for %s", modelName),
			FinishReason: "stop",
		}, nil
	}

	// Create environment provider (boundary mock)
	envProvider := NewMockEnvironmentProvider()
	envProvider.EnvVars["OPENAI_API_KEY"] = "dummy-api-key"

	// Create filesystem (real for simplicity)
	filesystem := &RealFilesystemIO{}

	// Create API service using the boundary mocks
	apiService := NewBoundaryAPIService(mockAPICaller, envProvider, logger)

	// Create context gatherer (boundary-based)
	contextGatherer := &BoundaryContextGatherer{
		filesystem: filesystem,
		logger:     logger,
	}

	// Create file writer (boundary-based)
	fileWriter := &BoundaryFileWriter{
		filesystem: filesystem,
		logger:     logger,
	}

	// Create audit logger (boundary-based)
	auditLogger := NewBoundaryAuditLogger(filesystem, logger)

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(cfg.MaxConcurrentRequests, cfg.RateLimitRequestsPerMinute)

	// Create orchestrator using the real implementation with boundary mocks
	orch := thinktank.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		cfg,
		logger,
	)

	// Run the orchestrator
	err = orch.Run(context.Background(), instructions)
	if err != nil {
		t.Fatalf("Orchestrator.Run failed: %v", err)
	}

	// Verify that synthesis output file was created
	expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")
	_, statErr := os.Stat(expectedSynthesisFile)
	if os.IsNotExist(statErr) {
		t.Errorf("Expected synthesis output file %s not created", expectedSynthesisFile)
	} else {
		// Verify synthesis file content
		content, readErr := os.ReadFile(expectedSynthesisFile)
		if readErr != nil {
			t.Errorf("Failed to read synthesis output file %s: %v", expectedSynthesisFile, readErr)
		} else {
			if string(content) != synthesisOutput {
				t.Errorf("File content mismatch for synthesis file:\nExpected: %s\nActual: %s",
					synthesisOutput, string(content))
			} else {
				t.Logf("Verified content for synthesis file")
			}
		}
	}

	// Verify that individual model output files were also created
	for _, modelName := range modelNames {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")
		_, modelStatErr := os.Stat(expectedFilePath)
		if os.IsNotExist(modelStatErr) {
			t.Errorf("Expected output file %s not created", expectedFilePath)
		} else {
			// Verify file content
			content, modelReadErr := os.ReadFile(expectedFilePath)
			if modelReadErr != nil {
				t.Errorf("Failed to read output file %s: %v", expectedFilePath, modelReadErr)
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

// TestBoundarySynthesisWithFailures tests synthesis with some failed models
func TestBoundarySynthesisWithFailures(t *testing.T) {
	// Create temp directory for outputs
	tempDir, err := os.MkdirTemp("", "thinktank-synthesis-failures-test")
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
	instructions := "Test instructions for synthesis with failures"

	// Set up multiple model names and synthesis model
	modelNames := []string{"model1", "model2", "model3"}
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

	// Create logger
	logger := logutil.NewTestLogger(t)

	// Create mock API caller (boundary mock)
	mockAPICaller := &MockExternalAPICaller{}

	// Expected model outputs
	mockOutputs := map[string]string{
		"model1": "# Output from Model 1\n\nThis is test output from model1.",
		"model3": "# Output from Model 3\n\nThis is test output from model3.",
	}

	// Expected synthesis output
	synthesisOutput := "# Synthesized Output\n\nThis content combines insights from models 1 and 3."

	// Configure mock API caller to return appropriate responses
	mockAPICaller.CallLLMAPIFunc = func(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		// Simulate model2 failing
		if modelName == "model2" {
			return nil, &llm.MockError{
				Message:       "API rate limit exceeded",
				ErrorCategory: llm.CategoryRateLimit,
			}
		}

		// For synthesis model
		if modelName == synthesisModel {
			return &llm.ProviderResult{
				Content:      synthesisOutput,
				FinishReason: "stop",
			}, nil
		}

		// For regular models
		if content, ok := mockOutputs[modelName]; ok {
			return &llm.ProviderResult{
				Content:      content,
				FinishReason: "stop",
			}, nil
		}

		// Default response
		return &llm.ProviderResult{
			Content:      fmt.Sprintf("Default response for %s", modelName),
			FinishReason: "stop",
		}, nil
	}

	// Create environment provider (boundary mock)
	envProvider := NewMockEnvironmentProvider()
	envProvider.EnvVars["OPENAI_API_KEY"] = "dummy-api-key"

	// Create filesystem (real for simplicity)
	filesystem := &RealFilesystemIO{}

	// Create API service using the boundary mocks
	apiService := NewBoundaryAPIService(mockAPICaller, envProvider, logger)

	// Create context gatherer (boundary-based)
	contextGatherer := &BoundaryContextGatherer{
		filesystem: filesystem,
		logger:     logger,
	}

	// Create file writer (boundary-based)
	fileWriter := &BoundaryFileWriter{
		filesystem: filesystem,
		logger:     logger,
	}

	// Create audit logger (boundary-based)
	auditLogger := NewBoundaryAuditLogger(filesystem, logger)

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(cfg.MaxConcurrentRequests, cfg.RateLimitRequestsPerMinute)

	// Create orchestrator using the real implementation with boundary mocks
	orch := thinktank.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		cfg,
		logger,
	)

	// Run the orchestrator (expecting partial failure)
	err = orch.Run(context.Background(), instructions)

	// We expect an error due to the model2 failure
	if err == nil {
		t.Errorf("Expected error due to model2 failure, but got nil")
	} else if !strings.Contains(strings.ToLower(err.Error()), "processed") {
		t.Errorf("Expected error to mention partial processing, got: %v", err)
	} else {
		t.Logf("Got expected partial failure error: %v", err)
	}

	// Verify that synthesis output file was created despite the failure
	expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")
	_, statErr := os.Stat(expectedSynthesisFile)
	if os.IsNotExist(statErr) {
		t.Errorf("Expected synthesis output file %s not created despite the failure", expectedSynthesisFile)
	} else {
		t.Logf("Synthesis file was correctly created despite the failure")
	}

	// Verify that model1 and model3 output files were created
	successfulModels := []string{"model1", "model3"}
	for _, modelName := range successfulModels {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")
		_, modelStatErr := os.Stat(expectedFilePath)
		if os.IsNotExist(modelStatErr) {
			t.Errorf("Expected output file %s not created", expectedFilePath)
		} else {
			t.Logf("Verified file creation for model %s", modelName)
		}
	}

	// Verify model2 file was NOT created
	failedModelPath := filepath.Join(outputDir, "model2.md")
	_, modelStatErr := os.Stat(failedModelPath)
	if !os.IsNotExist(modelStatErr) {
		t.Errorf("File for failed model2 should not exist, but it does")
	} else {
		t.Logf("Correctly verified that failed model2 has no output file")
	}
}
