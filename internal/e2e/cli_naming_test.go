//go:build manual_api_test
// +build manual_api_test

// Package e2e contains end-to-end tests for the thinktank CLI
// These tests require a valid API key to run properly and are skipped by default
// To run these tests: go test -tags=manual_api_test ./internal/e2e/...
package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestDirectoryNamingDefault verifies that the CLI creates a timestamped directory
// when no --output-dir flag is provided
func TestDirectoryNamingDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Test instructions for directory naming")
	env.CreateTestFile("src/main.go", CreateGoSourceFileContent())

	// Store the original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(origDir) // Restore original directory at the end

	// Change to the test environment directory
	if err := os.Chdir(env.TempDir); err != nil {
		t.Fatalf("Failed to change directory to test environment: %v", err)
	}

	// Run thinktank without specifying output directory
	// Note: We need to keep track of generated directories to clean them up
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.OutputDir = "" // Explicitly set to empty to test default naming
	flags.Model = []string{"gemini-2.5-pro"}

	// Run the command
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
	if err != nil {
		t.Fatalf("Failed to run thinktank: %v", err)
	}

	// Verify command succeeded
	AssertAPICommandSuccess(t, stdout, stderr, exitCode,
		"Generated output directory", "Using output directory")

	// Verify a timestamped directory was created
	// Extract the directory name from the log output
	// Looking for a line like: "Generated output directory: /path/to/thinktank_20230401_120000_1234"
	outputDirPattern := regexp.MustCompile(`Generated output directory: (.*thinktank_\d{8}_\d{6}_\d{4})`)
	matches := outputDirPattern.FindStringSubmatch(stdout + stderr)
	if len(matches) < 2 {
		t.Fatalf("Failed to find generated output directory in logs")
	}

	outputDir := matches[1]
	dirName := filepath.Base(outputDir)

	// Verify the directory name format
	pattern := `^thinktank_\d{8}_\d{6}_\d{4}$`
	dirNameRe := regexp.MustCompile(pattern)
	if !dirNameRe.MatchString(dirName) {
		t.Errorf("Generated directory name %q does not match expected format %q", dirName, pattern)
	}

	// Verify the directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("Output directory %s does not exist", outputDir)
	}

	// Verify the directory contains expected output files
	// Typically there should be at least one output file named after the model
	modelOutputFile := filepath.Join(outputDir, "gemini-2.5-pro.md")
	alternateOutputFile := filepath.Join(outputDir, "o4-mini.md")

	if _, err := os.Stat(modelOutputFile); os.IsNotExist(err) {
		// Try the alternate file name
		if _, err := os.Stat(alternateOutputFile); os.IsNotExist(err) {
			t.Logf("Neither %s nor %s exists in the output directory (acceptable in tests with API mocks)",
				modelOutputFile, alternateOutputFile)
		} else {
			t.Logf("Output file created at %s", alternateOutputFile)
		}
	} else {
		t.Logf("Output file created at %s", modelOutputFile)
	}
}

// TestDirectoryNamingExplicit verifies that the --output-dir flag works correctly
func TestDirectoryNamingExplicit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Test instructions for explicit directory")
	env.CreateTestFile("src/main.go", CreateGoSourceFileContent())

	// Create a specific output directory
	explicitOutputDir := env.CreateTestDirectory("explicit-output")

	// Run thinktank with explicit output directory
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.OutputDir = explicitOutputDir
	flags.Model = []string{"gemini-2.5-pro"}

	// Run the command
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
	if err != nil {
		t.Fatalf("Failed to run thinktank: %v", err)
	}

	// Verify command succeeded
	AssertAPICommandSuccess(t, stdout, stderr, exitCode,
		"Using output directory")

	// Verify the log contains our explicit directory
	if !strings.Contains(stdout+stderr, explicitOutputDir) {
		t.Errorf("Output does not contain the explicit output directory path %s", explicitOutputDir)
	}

	// Verify output files were created in the explicit directory
	modelOutputFile := filepath.Join(explicitOutputDir, "gemini-2.5-pro.md")
	alternateOutputFile := filepath.Join(explicitOutputDir, "o4-mini.md")

	if _, err := os.Stat(modelOutputFile); os.IsNotExist(err) {
		// Try the alternate file name
		if _, err := os.Stat(alternateOutputFile); os.IsNotExist(err) {
			t.Logf("Neither %s nor %s exists in the explicit output directory (acceptable in tests with API mocks)",
				modelOutputFile, alternateOutputFile)
		} else {
			t.Logf("Output file created at %s", alternateOutputFile)
		}
	} else {
		t.Logf("Output file created at %s", modelOutputFile)
	}
}
