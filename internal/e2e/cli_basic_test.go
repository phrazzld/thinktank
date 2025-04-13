package e2e

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestBasicExecution tests the most basic execution of the architect CLI
func TestBasicExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Implement a new feature to multiply two numbers")
	env.CreateTestFile("src/main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

	// Set up the output directory
	outputDir := filepath.Join(env.TempDir, "output")
	modelName := "test-model"
	outputFile := filepath.Join(outputDir, modelName+".md")

	// Construct arguments
	args := []string{
		"--instructions", instructionsFile,
		"--output-dir", outputDir,
		"--model", modelName,
		env.TempDir + "/src",
	}

	// Run the architect binary
	stdout, stderr, exitCode, err := env.RunArchitect(args, nil)
	if err != nil {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify output file was created
	if !env.FileExists(filepath.Join("output", modelName+".md")) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Verify output file content
	content, err := env.ReadFile(filepath.Join("output", modelName+".md"))
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(content, "Test Generated Plan") {
		t.Errorf("Output file does not contain expected content")
	}
}

// TestDryRunMode tests the dry run mode
func TestDryRunMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	env.CreateTestFile("src/main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

	// Set up the output directory
	outputDir := filepath.Join(env.TempDir, "output")
	modelName := "test-model"
	outputFile := filepath.Join(outputDir, modelName+".md")

	// Set up flags for dry run
	flags := env.DefaultFlags
	flags.DryRun = true
	flags.Instructions = "Test instructions"

	// Run the architect binary
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{env.TempDir + "/src"})
	if err != nil {
		t.Fatalf("Failed to run architect in dry run mode: %v", err)
	}

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify output file was NOT created
	if env.FileExists(filepath.Join("output", modelName+".md")) {
		t.Errorf("Output file was created in dry run mode at %s", outputFile)
	}

	// Verify stdout contains file statistics
	if !strings.Contains(stdout, "Files:") && !strings.Contains(stderr, "Files:") {
		t.Errorf("Dry run output does not contain file statistics")
	}
}

// TestMissingRequiredFlags tests error handling for missing required flags
func TestMissingRequiredFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	testCases := []struct {
		name             string
		args             []string
		expectedExitCode int
		expectedError    string
	}{
		{
			name:             "Missing instructions file and paths",
			args:             []string{},
			expectedExitCode: 1,
			expectedError:    "no paths specified",
		},
		{
			name:             "Missing paths",
			args:             []string{"--instructions", "instructions.md"},
			expectedExitCode: 1,
			expectedError:    "no paths specified",
		},
		{
			name:             "Missing instructions file",
			args:             []string{env.TempDir + "/src"},
			expectedExitCode: 1,
			expectedError:    "missing required --instructions",
		},
		{
			name:             "Dry run allows missing instructions file",
			args:             []string{"--dry-run", env.TempDir + "/src"},
			expectedExitCode: 0,
			expectedError:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the architect binary
			stdout, stderr, exitCode, err := env.RunArchitect(tc.args, nil)
			if err != nil && err.Error() != "exit status 1" {
				t.Fatalf("Failed to run architect: %v", err)
			}

			// Verify exit code
			if exitCode != tc.expectedExitCode {
				t.Errorf("Expected exit code %d, got %d", tc.expectedExitCode, exitCode)
				t.Logf("Stdout: %s", stdout)
				t.Logf("Stderr: %s", stderr)
			}

			// If we expect an error, verify error message
			if tc.expectedError != "" {
				combinedOutput := stdout + stderr
				if !strings.Contains(combinedOutput, tc.expectedError) {
					t.Errorf("Expected error message to contain %q, but got: %s", tc.expectedError, combinedOutput)
				}
			}
		})
	}
}

// TestAPIKeyError tests error handling for missing API key
func TestAPIKeyError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Test instructions")
	env.CreateTestFile("src/main.go", `package main

func main() {}`)

	// Set up the output directory
	outputDir := filepath.Join(env.TempDir, "output")

	// Construct arguments
	args := []string{
		"--instructions", instructionsFile,
		"--output-dir", outputDir,
		env.TempDir + "/src",
	}

	// Run the architect binary without API key in environment
	stdout, stderr, exitCode, err := env.RunArchitect(args, nil)
	if err != nil && err.Error() != "exit status 1" {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// Verify exit code
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify error message
	combinedOutput := stdout + stderr
	if !strings.Contains(combinedOutput, "API key not set") {
		t.Errorf("Expected error message to contain 'API key not set', but got: %s", combinedOutput)
	}
}

// TestVerboseFlagAndLogLevel tests the verbose flag and log level
func TestVerboseFlagAndLogLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	testCases := []struct {
		name          string
		verbose       bool
		logLevel      string
		expectedLevel string
	}{
		{
			name:          "Default log level",
			verbose:       false,
			logLevel:      "",
			expectedLevel: "INFO",
		},
		{
			name:          "Verbose flag",
			verbose:       true,
			logLevel:      "",
			expectedLevel: "DEBUG",
		},
		{
			name:          "Error log level",
			verbose:       false,
			logLevel:      "error",
			expectedLevel: "ERROR",
		},
		{
			name:          "Verbose overrides log level",
			verbose:       true,
			logLevel:      "error",
			expectedLevel: "DEBUG",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test files
			instructionsFile := env.CreateTestFile(fmt.Sprintf("instructions-%s.md", tc.name), "Test instructions")
			srcDir := fmt.Sprintf("src-%s", tc.name)
			env.CreateTestFile(filepath.Join(srcDir, "main.go"), `package main

func main() {}`)

			// Set up flags
			flags := env.DefaultFlags
			flags.Instructions = instructionsFile
			flags.Verbose = tc.verbose
			if tc.logLevel != "" {
				flags.LogLevel = tc.logLevel
			}
			flags.DryRun = true // Use dry run to make tests faster

			// Run the architect binary
			stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, srcDir)})
			if err != nil {
				t.Fatalf("Failed to run architect: %v", err)
			}

			// Verify exit code
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
				t.Logf("Stdout: %s", stdout)
				t.Logf("Stderr: %s", stderr)
			}

			// Verify log level
			combinedOutput := stdout + stderr
			levelFound := false
			// Look for the expected log level indicator
			if strings.Contains(combinedOutput, "["+tc.expectedLevel+"]") {
				levelFound = true
			}

			if !levelFound {
				t.Errorf("Expected log level %s not found in output", tc.expectedLevel)
				t.Logf("Stdout: %s", stdout)
				t.Logf("Stderr: %s", stderr)
			}
		})
	}
}
