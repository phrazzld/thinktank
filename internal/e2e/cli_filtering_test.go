//go:build manual_api_test
// +build manual_api_test

// Package e2e contains end-to-end tests for the architect CLI
// These tests require a valid API key to run properly and are skipped by default
// To run these tests: go test -tags=manual_api_test ./internal/e2e/...
package e2e

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestFileFiltering tests the file filtering flags (simplified to one representative test case)
func TestFileFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files of different types
	env.CreateTestFile("src/main.go", CreateGoSourceFileContent())
	env.CreateTestFile("src/utils.go", `package main
func add(a, b int) int { return a + b }`)
	env.CreateTestFile("src/README.md", "# Test Project")
	env.CreateTestFile("src/data.json", `{"key": "value"}`)
	env.CreateTestFile("src/ignored.tmp", "Temporary file to be ignored")
	env.CreateTestFile("src/node_modules/package.json", `{"name": "test"}`)

	// Create instructions file
	instructionsFile := env.CreateTestFile("instructions.md", "Test filtering")

	// Set up flags - test the most common filtering scenario
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.Include = "*.go,*.md"
	flags.Exclude = "utils*"
	flags.ExcludeNames = "node_modules"
	flags.DryRun = true

	// Run the architect binary
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
	if err != nil {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// Verify exit code
	VerifyOutput(t, stdout, stderr, exitCode, 0, "")

	// Combine stdout and stderr for checking
	combinedOutput := stdout + stderr

	// Check for expected files in output
	expectedFiles := []string{"main.go", "README.md"}
	for _, expectedFile := range expectedFiles {
		if !strings.Contains(combinedOutput, expectedFile) {
			t.Errorf("Expected file %s not found in output", expectedFile)
		}
	}

	// Check for files that should not be in output
	notExpectedFiles := []string{"utils.go", "data.json", "ignored.tmp", "node_modules"}
	for _, notExpectedFile := range notExpectedFiles {
		if strings.Contains(combinedOutput, notExpectedFile) {
			t.Errorf("Unexpected file %s found in output", notExpectedFile)
		}
	}
}
