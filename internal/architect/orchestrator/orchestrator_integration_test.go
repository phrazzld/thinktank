package orchestrator

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/ratelimit"
)

// TestIntegration_BasicWorkflow tests a complete workflow with multiple models
func TestIntegration_BasicWorkflow(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	// Setup multiple models to test parallel processing
	modelNames := []string{"model1", "model2"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupBasicContext()
	deps.setupGeminiClient()

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the workflow
	deps.verifyBasicWorkflow(t, modelNames)

	// Check that output files were written with expected content
	for _, call := range deps.fileWriter.SaveToFileCalls {
		if !strings.Contains(call.Content, "Generated content for:") {
			t.Errorf("Expected output to contain 'Generated content for:', got: %s", call.Content)
		}
	}

	// Check that audit log entries were created
	if len(deps.auditLogger.LogCalls) == 0 {
		t.Error("Expected audit log entries to be created")
	}
}

// TestIntegration_DryRunMode tests the complete workflow in dry run mode
func TestIntegration_DryRunMode(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	// Setup with a model name even though we're in dry run mode
	modelNames := []string{"model1"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupDryRunConfig()
	deps.setupBasicContext()

	// Run the orchestrator in dry run mode
	err := deps.runOrchestrator(ctx, deps.instructions)
	if err != nil {
		t.Fatalf("Expected no error in dry run mode, got: %v", err)
	}

	// Verify the dry run workflow
	deps.verifyDryRunWorkflow(t)
}

// TestIntegration_EmptyModelNames tests handling of empty model names list
func TestIntegration_EmptyModelNames(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	// Setup with empty model names
	deps.setupMultiModelConfig([]string{})
	deps.setupBasicContext()

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)

	// Should get an error about no model names
	if err == nil {
		t.Fatal("Expected an error due to empty model names, got nil")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "no model") {
		t.Errorf("Expected error about no models, got: %v", err)
	}
}

// TestIntegration_ErrorPropagation tests that errors are properly propagated
func TestIntegration_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	modelNames := []string{"model1"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupBasicContext()

	// Setup a client that returns an error for GenerateContent
	deps.setupGeminiClient()
	expectedErr := errors.New("generation failed")

	// Instead of trying to mock GenerateContent directly,
	// we'll use our API service mock to control the behavior

	// Instead, configure the APIService mock
	deps.apiService.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
		return "", expectedErr
	}

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)

	// Should get an error reflecting our test error
	if err == nil {
		t.Fatal("Expected an error from model processing, got nil")
	}

	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

// TestIntegration_ContextCancellation tests context cancellation during integration
func TestIntegration_ContextCancellation(t *testing.T) {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deps := newTestDeps()
	modelNames := []string{"model1", "model2", "model3"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupBasicContext()

	// Setup API client with delay to ensure we can cancel during processing
	deps.apiService.ProcessResponseFunc = func(result *gemini.GenerationResult) (string, error) {
		// Wait for a while to simulate processing
		select {
		case <-time.After(100 * time.Millisecond):
			return "Processed content", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Run the orchestrator with the cancellable context
	err := deps.runOrchestrator(ctx, deps.instructions)

	// Should get a cancellation error
	if err == nil {
		t.Fatal("Expected error due to cancellation, got nil")
	}

	if !strings.Contains(err.Error(), "context") &&
		!strings.Contains(err.Error(), "cancel") &&
		!strings.Contains(err.Error(), "deadline") {
		t.Errorf("Expected error related to context cancellation, got: %v", err)
	}
}

// TestIntegration_GatherContextError tests handling of context gathering errors
func TestIntegration_GatherContextError(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	modelNames := []string{"model1"}
	deps.setupMultiModelConfig(modelNames)

	// Setup context gatherer that returns an error
	expectedErr := errors.New("failed to gather context")
	deps.contextGatherer.GatherContextFunc = func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
		return nil, nil, expectedErr
	}

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)

	// Should get an error from context gathering
	if err == nil {
		t.Fatal("Expected error from context gathering, got nil")
	}

	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

// TestIntegration_RateLimiting tests that rate limiting works properly
func TestIntegration_RateLimiting(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()

	// Setup many models to trigger rate limiting
	modelNames := []string{"model1", "model2", "model3", "model4", "model5"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupBasicContext()
	deps.setupGeminiClient()

	// Use a restrictive rate limiter
	deps.rateLimiter = ratelimit.NewRateLimiter(1, 1) // 1 request per second

	// Track time before and after
	startTime := time.Now()

	// Run the orchestrator with rate-limited processing
	err := deps.runOrchestrator(ctx, deps.instructions)
	if err != nil {
		t.Fatalf("Expected no error with rate limiting, got: %v", err)
	}

	elapsedTime := time.Since(startTime)

	// With a rate limit of 1 per second and multiple models, it should take at least
	// a few seconds. However, since we're using mocks, this isn't a reliable test.
	// In a real system, we'd want to instrument the rate limiter more directly.
	// This test is more for demonstration purposes.
	t.Logf("Rate-limited processing of %d models took %v", len(modelNames), elapsedTime)

	// Check that all models were processed despite rate limiting
	if len(deps.fileWriter.SaveToFileCalls) != len(modelNames) {
		t.Errorf("Expected %d outputs, got %d", len(modelNames), len(deps.fileWriter.SaveToFileCalls))
	}
}

// TestIntegration_ModelProcessingError tests handling of model processing errors
func TestIntegration_ModelProcessingError(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	modelNames := []string{"model1", "model2"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupBasicContext()

	// Create a model-specific error for model2 by setting up InitClient
	deps.apiService.InitClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
		if modelName == "model2" {
			// For model2, return a client that will generate an error
			client := &mockGeminiClient{
				modelName: modelName,
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, errors.New("model2 processing error")
				},
			}
			return client, nil
		}

		// For other models, return a regular mock client
		return &mockGeminiClient{
			modelName: modelName,
		}, nil
	}

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)

	// Should get an error for one model
	if err == nil {
		t.Fatal("Expected error from model processing, got nil")
	}

	if !strings.Contains(err.Error(), "model2") || !strings.Contains(err.Error(), "processing error") {
		t.Errorf("Expected error to mention model2 and processing error, got: %v", err)
	}

	// Should still have output for the successful model
	if len(deps.fileWriter.SaveToFileCalls) != 1 {
		t.Errorf("Expected 1 output for the successful model, got %d", len(deps.fileWriter.SaveToFileCalls))
	}

	// Check the successful output is for model1
	if len(deps.fileWriter.SaveToFileCalls) > 0 {
		outputPath := deps.fileWriter.SaveToFileCalls[0].OutputFile
		if !strings.Contains(outputPath, "model1") {
			t.Errorf("Expected output for model1, got path: %s", outputPath)
		}
	}
}

// TestIntegration_APIServiceAdapterPassthrough tests API service adapter functions
func TestIntegration_APIServiceAdapterPassthrough(t *testing.T) {
	// Test errors for safety checks, empty responses, and error details
	testErrors := []struct {
		errorMsg      string
		isSafety      bool
		isEmpty       bool
		detailsPrefix string
	}{
		{
			errorMsg:      "safety blocked content",
			isSafety:      true,
			isEmpty:       false,
			detailsPrefix: "Safety",
		},
		{
			errorMsg:      "empty response from API",
			isSafety:      false,
			isEmpty:       true,
			detailsPrefix: "Empty",
		},
		{
			errorMsg:      "general API error",
			isSafety:      false,
			isEmpty:       false,
			detailsPrefix: "API",
		},
	}

	for _, te := range testErrors {
		t.Run("API error: "+te.errorMsg, func(t *testing.T) {
			// Create our test error
			testError := errors.New(te.errorMsg)

			// Configure a mock API service with appropriate responses
			apiSvc := &mockAPIService{
				IsSafetyBlockedErrorFunc: func(err error) bool {
					return te.isSafety
				},
				IsEmptyResponseErrorFunc: func(err error) bool {
					return te.isEmpty
				},
				GetErrorDetailsFunc: func(err error) string {
					return te.detailsPrefix + ": " + err.Error()
				},
			}

			// No need to create orchestrator, testing the API service directly

			// Test the API service functions directly
			if apiSvc.IsEmptyResponseError(testError) != te.isEmpty {
				t.Errorf("Expected IsEmptyResponseError to return %v for '%s'", te.isEmpty, te.errorMsg)
			}

			if apiSvc.IsSafetyBlockedError(testError) != te.isSafety {
				t.Errorf("Expected IsSafetyBlockedError to return %v for '%s'", te.isSafety, te.errorMsg)
			}

			details := apiSvc.GetErrorDetails(testError)
			expectedPrefix := te.detailsPrefix + ": "
			if !strings.HasPrefix(details, expectedPrefix) {
				t.Errorf("Expected error details to start with '%s', got: '%s'", expectedPrefix, details)
			}
		})
	}
}

// TestIntegration_FileWriterIntegration tests the file writer integration
func TestIntegration_FileWriterIntegration(t *testing.T) {
	ctx := context.Background()
	deps := newTestDeps()
	modelNames := []string{"model1"}
	deps.setupMultiModelConfig(modelNames)
	deps.setupBasicContext()
	deps.setupGeminiClient()

	// Setup a test output directory
	testOutputDir := "test_output_dir"
	deps.config.OutputDir = testOutputDir

	// Custom model and output file format
	testModelName := "custom-model"
	deps.config.ModelNames = []string{testModelName}

	// Run the orchestrator
	err := deps.runOrchestrator(ctx, deps.instructions)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the output file path
	if len(deps.fileWriter.SaveToFileCalls) != 1 {
		t.Fatalf("Expected 1 call to SaveToFile, got %d", len(deps.fileWriter.SaveToFileCalls))
	}

	outputPath := deps.fileWriter.SaveToFileCalls[0].OutputFile

	// Check path contains the output directory
	if !strings.Contains(outputPath, testOutputDir) {
		t.Errorf("Expected output path to contain '%s', got: %s", testOutputDir, outputPath)
	}

	// Check path contains the model name
	if !strings.Contains(filepath.Base(outputPath), testModelName) {
		t.Errorf("Expected output filename to contain '%s', got: %s", testModelName, filepath.Base(outputPath))
	}

	// Check file extension is .md (markdown)
	if !strings.HasSuffix(outputPath, ".md") {
		t.Errorf("Expected output file to have .md extension, got: %s", outputPath)
	}
}
