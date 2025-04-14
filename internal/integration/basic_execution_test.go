// internal/integration/basic_execution_test.go
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

// createMockAPIService creates a mock API service from a TestEnv
func createMockAPIService(env *TestEnv) *mockIntAPIService {
	return &mockIntAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}
}

// TestBasicExecutionFlows tests various basic execution flows of the application
// using a table-driven approach to reduce repetitive test code.
func TestBasicExecutionFlows(t *testing.T) {
	// Define the test case struct for basic execution test scenarios
	type basicExecutionTestCase struct {
		name                string
		instructionsContent string
		srcFiles            map[string]string
		modelName           string
		expectedContent     string
	}

	// Define test cases based on the original individual tests
	tests := []basicExecutionTestCase{
		{
			name:                "Basic Plan Generation",
			instructionsContent: "Implement a new feature to multiply two numbers",
			srcFiles: map[string]string{
				"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
				"utils.go": `package main

func add(a, b int) int {
	return a + b
}`,
			},
			modelName:       "test-model",
			expectedContent: "Test Generated Plan",
		},
		{
			name:                "Instructions File Input",
			instructionsContent: "Implement a new feature to multiply two numbers",
			srcFiles: map[string]string{
				"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
			},
			modelName:       "test-model",
			expectedContent: "Test Generated Plan",
		},
		{
			name:                "Task Execution",
			instructionsContent: "Implement a new feature",
			srcFiles: map[string]string{
				"main.go": `package main

func main() {}`,
			},
			modelName:       "test-model",
			expectedContent: "Test Generated Plan",
		},
	}

	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the mock client with default responses
			env.SetupMockGeminiClient()

			// Create source files from the map
			for filename, content := range tc.srcFiles {
				env.CreateTestFile(t, "src/"+filename, content)
			}

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			outputDir := filepath.Join(env.TestDir, "output")
			outputFile := filepath.Join(outputDir, tc.modelName+".md")

			// Create a test configuration
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{tc.modelName},
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

			// Verify execution succeeded
			if err != nil {
				t.Fatalf("architect.Execute failed: %v", err)
			}

			// Check that the model-specific output file exists
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Errorf("Model-specific output file was not created at %s", outputFile)
			}

			// Verify content
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read model-specific output file: %v", err)
			}

			// Check that content includes expected text
			if !strings.Contains(string(content), tc.expectedContent) {
				t.Errorf("Model-specific output file does not contain expected content: %s", tc.expectedContent)
			}
		})
	}
}

// TestModeVariations tests different execution modes of the application
// using a table-driven approach to reduce repetitive test code.
func TestModeVariations(t *testing.T) {
	// Define the test case struct for mode variation test scenarios
	type modeTestCase struct {
		name                string
		instructionsContent string
		srcFiles            map[string]string
		modelName           string
		dryRun              bool
		outputShouldExist   bool
		expectedContent     string
	}

	// Define test cases based on the original tests
	tests := []modeTestCase{
		{
			name:                "Dry Run Mode",
			instructionsContent: "Test instructions",
			srcFiles: map[string]string{
				"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
			},
			modelName:         "test-model",
			dryRun:            true,
			outputShouldExist: false,
			expectedContent:   "", // Not needed for dry run
		},
		// Add more mode variations here if needed
	}

	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the mock client with default responses
			env.SetupMockGeminiClient()

			// Create source files from the map
			for filename, content := range tc.srcFiles {
				env.CreateTestFile(t, "src/"+filename, content)
			}

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			outputDir := filepath.Join(env.TestDir, "output")
			outputFile := filepath.Join(outputDir, tc.modelName+".md")

			// Create a test configuration
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{tc.modelName},
				APIKey:           "test-api-key",
				DryRun:           tc.dryRun,
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

			// Check if output file exists or not based on expectation
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

// TestPromptFileTemplateHandling tests the different types of template files and their processing
func TestPromptFileTemplateHandling(t *testing.T) {
	// This test might need significant modifications to run with the new architecture
	// For now, we're skipping it but will implement properly in a future update
	t.Skip("Template handling tests need to be reimplemented with the new architecture")

	/* Template for future implementation:

	t.Run("RegularTextFile", func(t *testing.T) {
		// Set up the test environment
		env := NewTestEnv(t)
		defer env.Cleanup()
		env.SetupMockGeminiClient()

		// Create a regular text file without template variables
		instructionsContent := "This is a regular task description without any template variables."
		instructionsFile := env.CreateTestFile(t, "instructions.md", instructionsContent)

		// Set up the output file path
		outputDir := filepath.Join(env.TestDir, "output")
			modelName := "test-model"
			outputFile := filepath.Join(outputDir, modelName+".md")

		// Create a test configuration
		testConfig := &config.CliConfig{
			InstructionsFile: instructionsFile,
			OutputDir: outputDir,
			ModelNames:  []string{"test-model"},
			APIKey:     "test-api-key",
			Paths:      []string{env.TestDir},
			LogLevel:   logutil.InfoLevel,
		}

		// Run the application with our test configuration
		ctx := context.Background()
		err := RunTestWithConfig(ctx, testConfig, env)

		if err != nil {
			t.Fatalf("RunTestWithConfig failed: %v", err)
		}

		// Check that the output file exists
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Model-specific output file was not created at %s", outputFile)
		}

		// Verify the content indicates that the file was NOT processed as a template
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if !strings.Contains(string(content), "TEMPLATE_PROCESSED: NO") {
			t.Errorf("Regular text file was incorrectly processed as a template: %s", string(content))
		}
	})
	*/
}
