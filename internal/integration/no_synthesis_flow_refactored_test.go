// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"path/filepath"
	"testing"
)

// TestNoSynthesisFlowRefactored tests the complete flow without synthesis model specified
// This test verifies that multiple output files are created, one for each model
func TestNoSynthesisFlowRefactored(t *testing.T) {
	IntegrationTestWithBoundaries(t, func(env *BoundaryTestEnv) {
		// Set up test parameters
		instructions := "Test instructions for multiple models"
		modelNames := []string{"model1", "model2", "model3"}

		// Set up models without synthesis model
		env.Config.ModelNames = modelNames
		env.Config.SynthesisModel = "" // Explicitly set to empty to ensure no synthesis

		// Set up mock responses for each model
		mockOutputs := map[string]string{
			"model1": "# Output from Model 1\n\nThis is test output from model1.",
			"model2": "# Output from Model 2\n\nThis is test output from model2.",
			"model3": "# Output from Model 3\n\nThis is test output from model3.",
		}

		// Setup the test environment
		SetupStandardTestEnvironment(t, env, instructions, modelNames, "", mockOutputs)

		// Run the orchestrator
		ctx := context.Background()
		err := env.Run(ctx, instructions)
		if err != nil {
			t.Fatalf("Failed to run orchestration: %v", err)
		}

		// Verify that individual model output files were created (no synthesis file)
		outputDir := env.Config.OutputDir
		for _, modelName := range modelNames {
			expectedFilePath := filepath.Join(outputDir, modelName+".md")
			expectedContent := mockOutputs[modelName]
			VerifyFileContent(t, env, expectedFilePath, expectedContent)
		}

		// Verify no synthesis file was created
		// This ensures when SynthesisModel is empty, no synthesis is performed
		synthesisFilePath := filepath.Join(outputDir, "synthesis.md")
		exists, _ := env.Filesystem.Stat(synthesisFilePath)
		if exists {
			t.Errorf("Synthesis file %s should not exist when SynthesisModel is empty", synthesisFilePath)
		} else {
			t.Logf("Correctly verified no synthesis file was created")
		}
	})
}
