// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestSynthesisFlowRefactored tests the complete flow with synthesis model specified
// This test verifies that a single synthesis output file is created
func TestSynthesisFlowRefactored(t *testing.T) {
	t.Skip("Temporarily skipping test due to filesystem mocking issues")
	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// Set up test parameters
		instructions := "Test instructions for synthesis"
		modelNames := []string{"model1", "model2", "model3"}
		synthesisModel := "synthesis-model"

		// Set up mock responses for each model
		mockOutputs := map[string]string{
			"model1":       "# Output from Model 1\n\nThis is test output from model1.",
			"model2":       "# Output from Model 2\n\nThis is test output from model2.",
			"model3":       "# Output from Model 3\n\nThis is test output from model3.",
			synthesisModel: "# Synthesized Output\n\nThis content combines insights from all models.",
		}

		// Setup the test environment
		SetupStandardTestEnvironment(t, env, instructions, modelNames, synthesisModel, mockOutputs)

		// Run the orchestrator
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
			expectedFilePath := filepath.Join(outputDir, modelName+".md")
			expectedContent := mockOutputs[modelName]
			VerifyFileContent(t, env, expectedFilePath, expectedContent)
		}
	})
}

// TestSynthesisWithPartialFailureRefactored tests synthesis flow with some models failing
func TestSynthesisWithPartialFailureRefactored(t *testing.T) {
	t.Skip("Temporarily skipping test due to filesystem mocking issues")
	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// Set up test parameters
		instructions := "Test instructions for synthesis with partial failure"
		modelNames := []string{"model1", "model2", "model3"}
		synthesisModel := "synthesis-model"

		// Set up mock responses for each model
		mockOutputs := map[string]string{
			"model1":       "# Output from Model 1\n\nThis is test output from model1.",
			"model3":       "# Output from Model 3\n\nThis is test output from model3.",
			synthesisModel: "# Synthesized Output with Partial Failure\n\nThis content combines insights from models 1 and 3 only.",
		}

		// Setup the test environment
		SetupStandardTestEnvironment(t, env, instructions, modelNames, synthesisModel, mockOutputs)

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

		// Run the orchestrator, expecting partial failure
		ctx := context.Background()
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
			expectedFilePath := filepath.Join(outputDir, modelName+".md")
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
