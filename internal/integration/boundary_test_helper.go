// Package integration provides integration tests for the thinktank package
package integration

import (
	"path/filepath"
	"testing"
)

// TestWithBoundaries runs an integration test with proper boundary mocks
// This helper function makes it easier to refactor tests to use boundary mocks
func TestWithBoundaries(t *testing.T, testFunc func(env *BoundaryTestEnv)) {
	// Create test environment with mocked boundaries
	env := NewBoundaryTestEnv(t)

	// Run the test function with the environment
	testFunc(env)
}

// VerifyFileContent verifies a file exists in the mock filesystem and has the expected content
func VerifyFileContent(t *testing.T, env *BoundaryTestEnv, filePath, expectedContent string) {
	t.Helper()
	exists, err := env.Filesystem.Stat(filePath)
	if err != nil || !exists {
		t.Errorf("Expected file %s not created", filePath)
		return
	}

	content, err := env.Filesystem.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", filePath, err)
		return
	}

	if string(content) != expectedContent {
		t.Errorf("File content mismatch for %s:\nExpected: %s\nActual: %s",
			filePath, expectedContent, string(content))
	} else {
		t.Logf("Verified content for file %s", filepath.Base(filePath))
	}
}

// SetupStandardTestEnvironment configures a standard test environment with the given models
func SetupStandardTestEnvironment(t *testing.T, env *BoundaryTestEnv, instructions string,
	modelNames []string, synthesisModel string, modelOutputs map[string]string) string {
	t.Helper()
	// Set up models to use
	env.SetupModels(modelNames, synthesisModel)

	// Set up mock responses for each model
	for model, output := range modelOutputs {
		env.SetupModelResponse(model, output)
	}

	// Configure instructions file (this creates the file in the mock filesystem)
	instructionsPath := env.SetupInstructionsFile(instructions)

	// Verify the file exists in the mock filesystem
	fileExists, _ := env.Filesystem.Stat(instructionsPath)
	if !fileExists {
		t.Fatalf("Failed to create instructions file at %s", instructionsPath)
	}

	// Explicitly set the instructions file in the config
	env.Config.InstructionsFile = instructionsPath

	return instructionsPath
}
