package e2e

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

// TestTokenLimit tests behavior when token count exceeds limits
func TestTokenLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Implement a new feature")
	env.CreateTestFile("src/main.go", `package main

func main() {}`)

	// Set up the output directory
	outputDir := filepath.Join(env.TempDir, "output")

	// Configure the mock server to return a token count that exceeds the model's limit
	originalHandleTokenCount := env.MockConfig.HandleTokenCount
	originalHandleModelInfo := env.MockConfig.HandleModelInfo

	// Set token count to 100,000 (very large)
	env.MockConfig.HandleTokenCount = func(req *http.Request) (int, error) {
		return 100000, nil
	}

	// Set model input token limit to 50,000 (less than the token count)
	env.MockConfig.HandleModelInfo = func(req *http.Request) (string, int, int, error) {
		return "test-model", 50000, 8192, nil
	}

	// Restore original handlers when done
	defer func() {
		env.MockConfig.HandleTokenCount = originalHandleTokenCount
		env.MockConfig.HandleModelInfo = originalHandleModelInfo
	}()

	// Set up flags
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.OutputDir = outputDir

	// Run the architect binary
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
	if err != nil {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// Verify exit code (should be non-zero due to token limit error)
	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code, got 0")
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify error message contains token limit information
	combinedOutput := stdout + stderr
	if !strings.Contains(combinedOutput, "token") || !strings.Contains(combinedOutput, "limit") {
		t.Errorf("Output does not contain token limit error information")
	}

	// Verify output file is NOT created
	if env.FileExists(filepath.Join("output", "test-model.md")) {
		t.Errorf("Output file was created despite token limit error")
	}
}

// TestUserConfirmation tests the confirmation prompt for large token counts
func TestUserConfirmation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Implement a new feature")
	env.CreateTestFile("src/main.go", `package main

func main() {}`)

	// Set up the output directory
	outputDir := filepath.Join(env.TempDir, "output")

	// Configure the mock server to return a token count that will trigger confirmation
	originalHandleTokenCount := env.MockConfig.HandleTokenCount

	// Set token count to 5,000
	env.MockConfig.HandleTokenCount = func(req *http.Request) (int, error) {
		return 5000, nil
	}

	// Restore original handler when done
	defer func() {
		env.MockConfig.HandleTokenCount = originalHandleTokenCount
	}()

	testCases := []struct {
		name           string
		confirmTokens  int
		userInput      string
		expectedExit   int
		shouldGenerate bool
	}{
		{
			name:           "Confirm with 'y'",
			confirmTokens:  1000, // Lower than token count to trigger confirmation
			userInput:      "y\n",
			expectedExit:   0,
			shouldGenerate: true,
		},
		{
			name:           "Reject with 'n'",
			confirmTokens:  1000, // Lower than token count to trigger confirmation
			userInput:      "n\n",
			expectedExit:   1,
			shouldGenerate: false,
		},
		{
			name:           "No confirmation needed",
			confirmTokens:  10000, // Higher than token count, no confirmation needed
			userInput:      "",    // No input needed
			expectedExit:   0,
			shouldGenerate: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up flags
			flags := env.DefaultFlags
			flags.Instructions = instructionsFile
			flags.OutputDir = outputDir
			flags.ConfirmTokens = tc.confirmTokens

			// Create a reader for user input
			stdin := SimulateUserInput(tc.userInput)

			// Construct arguments
			args := []string{
				"--instructions", instructionsFile,
				"--output-dir", outputDir,
				"--confirm-tokens", fmt.Sprintf("%d", tc.confirmTokens),
				filepath.Join(env.TempDir, "src"),
			}

			// Run the architect binary
			stdout, stderr, exitCode, err := env.RunArchitect(args, stdin)
			if err != nil && tc.expectedExit == 0 {
				t.Fatalf("Failed to run architect: %v", err)
			}

			// Verify exit code
			if exitCode != tc.expectedExit {
				t.Errorf("Expected exit code %d, got %d", tc.expectedExit, exitCode)
				t.Logf("Stdout: %s", stdout)
				t.Logf("Stderr: %s", stderr)
			}

			// Verify prompt in output
			combinedOutput := stdout + stderr
			if tc.confirmTokens < 5000 && !strings.Contains(combinedOutput, "confirm") {
				t.Errorf("Expected confirmation prompt, but none found in output")
			}

			// Verify output file creation
			hasOutputFile := env.FileExists(filepath.Join("output", "test-model.md"))
			if tc.shouldGenerate && !hasOutputFile {
				t.Errorf("Expected output file to be created, but it wasn't")
			} else if !tc.shouldGenerate && hasOutputFile {
				t.Errorf("Expected no output file to be created, but it was")
			}
		})
	}
}
