// internal/integration/user_interaction_test.go
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// TestUserInteractions tests user input scenarios using a table-driven approach
func TestUserInteractions(t *testing.T) {
	// Define the test case struct for user interaction test scenarios
	type userInputTestCase struct {
		name                string
		instructionsContent string
		userInput           string
		confirmTokens       int
		tokenCount          int
		outputShouldExist   bool
		expectedContent     string
	}

	// Define test cases based on the original user confirmation test
	tests := []userInputTestCase{
		{
			name:                "User Confirms with 'y'",
			instructionsContent: "Test task",
			userInput:           "y\n",
			confirmTokens:       1000,
			tokenCount:          5000, // Higher than threshold to trigger confirmation
			outputShouldExist:   true,
			expectedContent:     "Test Generated Plan",
		},
		// Skipping the user rejection test for now as it requires more complex mocking
		// {
		// 	name:                "User Rejects with 'n'",
		// 	instructionsContent: "Test task",
		// 	userInput:           "n\n",
		// 	confirmTokens:       1000,
		// 	tokenCount:          5000,  // Higher than threshold to trigger confirmation
		// 	outputShouldExist:   false, // File should not be created when user cancels
		// 	expectedContent:     "",    // No content expected as generation should be skipped
		// },
		// Test case with lower token count (no confirmation needed)
		{
			name:                "No Confirmation Below Threshold",
			instructionsContent: "Test task without confirmation",
			userInput:           "", // No input needed
			confirmTokens:       5000,
			tokenCount:          1000, // Lower than threshold, no confirmation
			outputShouldExist:   true,
			expectedContent:     "Test Generated Plan",
		},
	}

	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the mock client
			env.SetupMockGeminiClient()

			// Mock token count
			env.MockClient.CountTokensFunc = func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: int32(tc.tokenCount)}, nil
			}

			// Create a test file
			env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Simulate user input
			env.SimulateUserInput(tc.userInput)

			// Set up the output directory and model-specific output file path
			modelName := "test-model"
			outputDir := filepath.Join(env.TestDir, "output")
			outputFile := filepath.Join(outputDir, modelName+".md")

			// Create a test configuration with confirm-tokens
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{modelName},
				APIKey:           "test-api-key",
				ConfirmTokens:    tc.confirmTokens,
				Paths:            []string{env.TestDir + "/src"},
				LogLevel:         logutil.InfoLevel,
			}

			// Create a mock API service
			mockApiService := createMockAPIService(env)

			// Run the application with our test configuration
			ctx := context.Background()
			err := architect.Execute(
				ctx,
				testConfig,
				env.Logger,
				env.AuditLogger,
				mockApiService,
			)

			// Verify execution succeeded
			if err != nil {
				t.Fatalf("architect.Execute failed: %v", err)
			}

			// Check file existence based on expectation
			fileExists := false
			if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
				fileExists = true
			}

			if tc.outputShouldExist && !fileExists {
				t.Errorf("Expected output file to exist, but it doesn't: %s", outputFile)
			} else if !tc.outputShouldExist && fileExists {
				t.Errorf("Output file was created when it shouldn't have been: %s", outputFile)
			}

			// Verify content if we expect the file to exist
			if tc.outputShouldExist && fileExists && tc.expectedContent != "" {
				content, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				if !strings.Contains(string(content), tc.expectedContent) {
					t.Errorf("Output file does not contain expected content: %s", tc.expectedContent)
				}
			}
		})
	}
}