// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/thinktank/modelproc"
)

// TestBoundarySynthesisFlow tests the complete flow with synthesis model using boundary mocks
// This test demonstrates the approach of mocking only external boundaries while using real
// internal implementations.
func TestBoundarySynthesisFlow(t *testing.T) {
	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// Configure test parameters
		instructions := "Test instructions for synthesis"
		modelNames := []string{"moonshotai/kimi-k2.5", "gpt-5.2", "gemini-3-flash"}
		synthesisModel := "gpt-5.2"

		// Set up mock responses for each model
		mockOutputs := map[string]string{
			"moonshotai/kimi-k2.5": "# Output from Model 1\n\nThis is test output from gpt-4o-mini.",
			"gpt-5.2":              "# Output from Model 2\n\nThis is test output from gpt-4o.",
			"gemini-3-flash":       "# Output from Model 3\n\nThis is test output from gemini-3-flash.",
			synthesisModel:         "# Synthesized Output\n\nThis content combines insights from all models.",
		}

		// Setup the test environment using the standard helper
		SetupStandardTestEnvironment(t, env, instructions, modelNames, synthesisModel, mockOutputs)

		// Run the test with a context
		ctx := context.Background()
		err := env.Run(ctx, instructions)
		if err != nil {
			t.Fatalf("Failed to run orchestration: %v", err)
		}

		// Verify that synthesis output file was created
		outputDir := env.Config.OutputDir
		expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")
		expectedSynthesisContent := mockOutputs[synthesisModel]
		VerifyFileContent(t, env, expectedSynthesisFile, expectedSynthesisContent)

		// Verify that individual model output files were also created
		for _, modelName := range modelNames {
			sanitizedModelName := modelproc.SanitizeFilename(modelName)
			expectedFilePath := filepath.Join(outputDir, sanitizedModelName+".md")
			expectedContent := mockOutputs[modelName]
			VerifyFileContent(t, env, expectedFilePath, expectedContent)
		}
	})
}

// TestBoundarySynthesisWithPartialFailure tests synthesis with some failed models
func TestBoundarySynthesisWithPartialFailure(t *testing.T) {
	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// Configure test parameters
		instructions := "Test instructions for synthesis with partial failure"
		modelNames := []string{"model1", "model2", "model3"}
		synthesisModel := "synthesis-model"

		// Set up mock responses for each model
		mockOutputs := map[string]string{
			"model1":       "# Output from Model 1\n\nThis is test output from model1.",
			"model3":       "# Output from Model 3\n\nThis is test output from model3.",
			synthesisModel: "# Synthesized Output with Partial Failure\n\nThis content combines insights from models 1 and 3 only.",
		}

		// Setup the test environment using the standard helper
		SetupStandardTestEnvironment(t, env, instructions, modelNames, synthesisModel, mockOutputs)

		// Declare expected error patterns for model2 failure (this is part of the test scenario)
		// Include both legacy error messages from modelproc and new detailed error structure
		env.ExpectError("Generation failed for model model2")         // Legacy from modelproc/processor.go:148
		env.ExpectError("Error generating content with model model2") // Legacy from modelproc/processor.go:152
		env.ExpectError("output generation failed for model model2")  // New detailed error from orchestrator
		env.ExpectError("API rate limit exceeded")                    // Part of detailed error chain
		env.ExpectError("Completed with model errors")                // Final error summary

		// Configure model2 to fail
		mockAPICaller := env.APICaller.(*MockExternalAPICaller)
		originalCallFunc := mockAPICaller.CallLLMAPIFunc
		mockAPICaller.CallLLMAPIFunc = func(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
			if modelName == "model2" {
				return nil, &llm.MockError{
					Message:       "API rate limit exceeded",
					ErrorCategory: llm.CategoryRateLimit,
				}
			}
			return originalCallFunc(ctx, modelName, prompt, params)
		}

		// Run the test with a context that includes correlation ID
		ctx := logutil.WithCorrelationID(context.Background())
		err := env.Run(ctx, instructions)

		// We expect an error due to the partial failure, but synthesis should still work
		if err == nil {
			t.Errorf("Expected error due to model2 failure, but got nil")
		} else if !strings.Contains(strings.ToLower(err.Error()), "processed") {
			t.Errorf("Expected error message about partial success, got: %v", err)
		} else {
			t.Logf("Got expected partial failure error: %v", err)
		}

		// Verify that synthesis output file was created despite the partial failure
		outputDir := env.Config.OutputDir
		expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")
		expectedSynthesisContent := mockOutputs[synthesisModel]
		VerifyFileContent(t, env, expectedSynthesisFile, expectedSynthesisContent)

		// Verify that model1 and model3 output files were created, but not model2
		successfulModels := []string{"model1", "model3"}
		for _, modelName := range successfulModels {
			sanitizedModelName := modelproc.SanitizeFilename(modelName)
			expectedFilePath := filepath.Join(outputDir, sanitizedModelName+".md")
			expectedContent := mockOutputs[modelName]
			VerifyFileContent(t, env, expectedFilePath, expectedContent)
		}

		// Verify model2 file was NOT created
		failedModelPath := filepath.Join(outputDir, "model2.md")
		exists, _ := env.Filesystem.Stat(failedModelPath)
		if exists {
			t.Errorf("File for failed model2 should not exist, but it does")
		} else {
			t.Logf("Correctly verified that failed model2 has no output file")
		}
	})
}

// Remove unused function
