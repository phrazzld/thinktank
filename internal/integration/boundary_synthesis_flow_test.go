// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestBoundarySynthesisFlow tests the complete flow with synthesis model using boundary mocks
// This test demonstrates the approach of mocking only external boundaries while using real
// internal implementations.
func TestBoundarySynthesisFlow(t *testing.T) {
	t.Skip("Temporarily skipping test due to filesystem mocking issues")
	// Use the real orchestrator logic with mocked external boundaries
	// Create test environment with mocked boundaries
	env := NewBoundaryTestEnv(t)

	// Configure test parameters
	instructions := "Test instructions for synthesis"
	modelNames := []string{"model1", "model2", "model3"}
	synthesisModel := "synthesis-model"

	// Set up models to use
	env.SetupModels(modelNames, synthesisModel)

	// Set up specific model outputs for testing
	env.SetupModelResponse("model1", "# Output from Model 1\n\nThis is test output from model1.")
	env.SetupModelResponse("model2", "# Output from Model 2\n\nThis is test output from model2.")
	env.SetupModelResponse("model3", "# Output from Model 3\n\nThis is test output from model3.")
	env.SetupModelResponse(synthesisModel, "# Synthesized Output\n\nThis content combines insights from all models.")

	// Configure instructions file
	instructionsPath := env.SetupInstructionsFile(instructions)

	// Verify that the instructions file was created
	fileExists, _ := env.Filesystem.Stat(instructionsPath)
	if !fileExists {
		t.Fatalf("Failed to create instructions file at %s", instructionsPath)
	}

	// Run the test with a context
	ctx := context.Background()
	err := env.Run(ctx, instructions)
	if err != nil {
		t.Fatalf("Failed to run orchestration: %v", err)
	}

	// Verify that synthesis output file was created
	outputDir := env.Config.OutputDir
	expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")

	exists, err := env.Filesystem.Stat(expectedSynthesisFile)
	if err != nil || !exists {
		t.Errorf("Expected synthesis output file %s not created", expectedSynthesisFile)
	} else {
		// Verify synthesis file content
		content, err := env.Filesystem.ReadFile(expectedSynthesisFile)
		if err != nil {
			t.Errorf("Failed to read synthesis output file %s: %v", expectedSynthesisFile, err)
		} else {
			expectedContent := "# Synthesized Output\n\nThis content combines insights from all models."
			if string(content) != expectedContent {
				t.Errorf("File content mismatch for synthesis file:\nExpected: %s\nActual: %s",
					expectedContent, string(content))
			} else {
				t.Logf("Verified content for synthesis file")
			}
		}
	}

	// Verify that individual model output files were also created
	for _, modelName := range modelNames {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")

		exists, err := env.Filesystem.Stat(expectedFilePath)
		if err != nil || !exists {
			t.Errorf("Expected output file %s not created", expectedFilePath)
		} else {
			// Verify file content
			content, err := env.Filesystem.ReadFile(expectedFilePath)
			if err != nil {
				t.Errorf("Failed to read output file %s: %v", expectedFilePath, err)
			} else {
				expectedContent := env.ModelOutputs[modelName]
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

// TestBoundarySynthesisWithPartialFailure tests synthesis with some failed models
func TestBoundarySynthesisWithPartialFailure(t *testing.T) {
	t.Skip("Temporarily skipping test due to filesystem mocking issues")
	// Use the real orchestrator logic with mocked external boundaries
	// Create test environment with mocked boundaries
	env := NewBoundaryTestEnv(t)

	// Configure test parameters
	instructions := "Test instructions for synthesis with partial failure"
	modelNames := []string{"model1", "model2", "model3"}
	synthesisModel := "synthesis-model"

	// Set up models to use
	env.SetupModels(modelNames, synthesisModel)

	// Set up specific model outputs for testing
	env.SetupModelResponse("model1", "# Output from Model 1\n\nThis is test output from model1.")

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

	env.SetupModelResponse("model3", "# Output from Model 3\n\nThis is test output from model3.")
	env.SetupModelResponse(synthesisModel, "# Synthesized Output with Partial Failure\n\nThis content combines insights from models 1 and 3 only.")

	// Configure instructions file
	instructionsPath := env.SetupInstructionsFile(instructions)

	// Verify that the instructions file was created
	fileExists, _ := env.Filesystem.Stat(instructionsPath)
	if !fileExists {
		t.Fatalf("Failed to create instructions file at %s", instructionsPath)
	}

	// Run the test with a context
	ctx := logutil.WithCorrelationID(context.Background())
	err := env.Run(ctx, instructions)

	// We expect an error due to the partial failure, but synthesis should still work
	if err == nil {
		t.Errorf("Expected error due to model2 failure, but got nil")
	} else if !containsSubstring(err.Error(), "processed 2/3 models successfully") {
		t.Errorf("Expected error message about partial success, got: %v", err)
	}

	// Verify that synthesis output file was created despite the partial failure
	outputDir := env.Config.OutputDir
	expectedSynthesisFile := filepath.Join(outputDir, synthesisModel+"-synthesis.md")

	exists, err := env.Filesystem.Stat(expectedSynthesisFile)
	if err != nil || !exists {
		t.Errorf("Expected synthesis output file %s not created despite partial failure", expectedSynthesisFile)
	} else {
		t.Logf("Synthesis file was correctly created despite partial failure")
	}

	// Verify that model1 and model3 output files were created, but not model2
	successfulModels := []string{"model1", "model3"}
	for _, modelName := range successfulModels {
		expectedFilePath := filepath.Join(outputDir, modelName+".md")

		exists, err := env.Filesystem.Stat(expectedFilePath)
		if err != nil || !exists {
			t.Errorf("Expected output file %s not created", expectedFilePath)
		} else {
			t.Logf("Verified file creation for model %s", modelName)
		}
	}

	// Verify model2 file was NOT created
	failedModelPath := filepath.Join(outputDir, "model2.md")
	exists, _ = env.Filesystem.Stat(failedModelPath)
	if exists {
		t.Errorf("File for failed model2 should not exist, but it does")
	} else {
		t.Logf("Correctly verified that failed model2 has no output file")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
