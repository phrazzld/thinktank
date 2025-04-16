// internal/integration/error_handling_test.go
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

// TestErrorScenarios tests error handling using a table-driven approach
func TestErrorScenarios(t *testing.T) {
	t.Parallel() // Add parallelization
	// Define the test case struct for error handling scenarios
	type errorTestCase struct {
		name                 string
		instructionsContent  string
		setupMock            func(*gemini.MockClient)
		expectedErrorContent string
		outputShouldExist    bool
	}

	// Define test cases based on the original error tests
	tests := []errorTestCase{
		{
			name:                "API Error Handling",
			instructionsContent: "Test task",
			setupMock: func(mc *gemini.MockClient) {
				mc.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
					// Create a simple API error
					apiError := &gemini.APIError{
						Message:    "API quota exceeded",
						Suggestion: "Try again later",
						StatusCode: 429,
					}
					return nil, apiError
				}
			},
			expectedErrorContent: "API quota exceeded",
			outputShouldExist:    false,
		},
		{
			name:                "Token Count Exceeded",
			instructionsContent: "Test task",
			setupMock: func(mc *gemini.MockClient) {
				mc.CountTokensFunc = func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
					return &gemini.TokenCount{Total: 1000000}, nil // Very large count
				}
				mc.GetModelInfoFunc = func(ctx context.Context) (*gemini.ModelInfo, error) {
					return &gemini.ModelInfo{
						Name:             "test-model",
						InputTokenLimit:  10000, // Smaller than the count
						OutputTokenLimit: 8192,
					}, nil
				}
			},
			expectedErrorContent: "token",
			outputShouldExist:    false,
		},
	}

	// Basic source file content for all tests
	srcFileContent := `package main

func main() {}`

	// Execute each test case
	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the mock client with test-specific configuration
			tc.setupMock(env.MockClient)

			// Create a test file
			env.CreateTestFile(t, "src/main.go", srcFileContent)

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up a unique output directory for test isolation using t.TempDir()
			modelName := "test-model"
			outputDir := t.TempDir()
			outputFile := filepath.Join(outputDir, modelName+".md")

			// Create a test configuration
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{modelName},
				APIKey:           "test-api-key",
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

			// Verify that an error was returned
			if err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			// Verify the error message contains expected content
			if !strings.Contains(err.Error(), tc.expectedErrorContent) {
				t.Errorf("Error message doesn't contain expected content %q. Got: %v",
					tc.expectedErrorContent, err)
			}

			// Check file existence based on expectation
			fileExists := false
			if _, statErr := os.Stat(outputFile); !os.IsNotExist(statErr) {
				fileExists = true
			}

			if tc.outputShouldExist && !fileExists {
				t.Errorf("Expected output file to exist, but it doesn't: %s", outputFile)
			} else if !tc.outputShouldExist && fileExists {
				t.Errorf("Output file was created when it shouldn't have been: %s", outputFile)
			}
		})
	}
}
