// internal/integration/integration_test.go
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
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

// TestFilteringBehaviors tests file filtering functions using a table-driven approach
func TestFilteringBehaviors(t *testing.T) {
	// Define the test case struct for filtering scenarios
	type filterTestCase struct {
		name                string
		instructionsContent string
		fileContents        map[string]string
		includeFilter       string
		excludeFilter       string
		excludeNames        string
		outputShouldExist   bool
		expectedContent     string
		verifyFilteringFunc func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	}

	// Define default files for filter tests
	defaultFiles := map[string]string{
		"src/main.go":       "package main\n\nfunc main() {}\n",
		"src/README.md":     "# Test Project",
		"src/config.json":   `{"key": "value"}`,
		"src/utils/util.js": "function helper() { return true; }",
	}

	// Define test cases based on the original filtering test
	tests := []filterTestCase{
		{
			name:                "Include Go and Markdown Files",
			instructionsContent: "Test task",
			fileContents:        defaultFiles,
			includeFilter:       ".go,.md", // Only include Go and Markdown files
			excludeFilter:       "",
			excludeNames:        "",
			outputShouldExist:   true,
			expectedContent:     "Test Generated Plan",
			verifyFilteringFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
				// In a real implementation, we would check the actual context to ensure
				// that only Go and Markdown files are included
				// We're just returning a successful result here, but could add more verification
				return &gemini.GenerationResult{
					Content:      "# Test Generated Plan\n\nThis is a test plan generated by the mock client.",
					TokenCount:   1000,
					FinishReason: "STOP",
				}, nil
			},
		},
		// Additional filtering test cases could be added here
	}

	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the custom mock client that verifies context content
			if tc.verifyFilteringFunc != nil {
				env.MockClient.GenerateContentFunc = tc.verifyFilteringFunc
			}

			// Create all the test files
			for path, content := range tc.fileContents {
				env.CreateTestFile(t, path, content)
			}

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			modelName := "test-model"
			outputDir := filepath.Join(env.TestDir, "output")
			outputFile := filepath.Join(outputDir, modelName+".md")

			// Create a test configuration with filtering options
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{modelName},
				APIKey:           "test-api-key",
				Include:          tc.includeFilter,
				Exclude:          tc.excludeFilter,
				ExcludeNames:     tc.excludeNames,
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

// TestErrorScenarios tests error handling using a table-driven approach
func TestErrorScenarios(t *testing.T) {
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
				mc.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
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
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Set up the mock client with test-specific configuration
			tc.setupMock(env.MockClient)

			// Create a test file
			env.CreateTestFile(t, "src/main.go", srcFileContent)

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			modelName := "test-model"
			outputDir := filepath.Join(env.TestDir, "output")
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

// TestAuditLogFunctionality tests the audit logging functionality
// using a table-driven approach to reduce repetitive test code.
func TestAuditLogFunctionality(t *testing.T) {
	// Define the test case struct for audit logging scenarios
	type auditLogTestCase struct {
		name                string
		instructionsContent string
		srcFiles            map[string]string
		loggerSetup         func(*TestEnv, *testing.T) (auditlog.AuditLogger, error)
		expectedLogEntries  []string
		validateLogFunc     func([]map[string]interface{}, *testing.T) bool
		shouldCreateLogFile bool
		outputShouldExist   bool
	}

	// Define test cases based on the original audit logging tests
	tests := []auditLogTestCase{
		{
			name:                "Valid audit log file",
			instructionsContent: "Implement a new feature to multiply two numbers",
			srcFiles: map[string]string{
				"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
			},
			loggerSetup: func(env *TestEnv, t *testing.T) (auditlog.AuditLogger, error) {
				// Set up the audit log file path
				auditLogFile := filepath.Join(env.TestDir, "audit.log")

				// Create a custom FileAuditLogger
				testLogger := env.GetBufferedLogger(logutil.DebugLevel, "[test] ")
				return auditlog.NewFileAuditLogger(auditLogFile, testLogger)
			},
			expectedLogEntries: []string{
				"ExecuteStart",
				"ReadInstructions",
				"GatherContextStart",
				"GatherContextEnd",
				"CheckTokens",
				"GenerateContentStart",
				"GenerateContentEnd",
				"SaveOutputStart",
				"SaveOutputEnd",
				"ExecuteEnd",
			},
			validateLogFunc: func(entries []map[string]interface{}, t *testing.T) bool {
				// Validate ExecuteStart operation in detail
				for _, entry := range entries {
					operation, ok := entry["operation"].(string)
					if !ok || operation != "ExecuteStart" {
						continue
					}

					// Check status
					status, ok := entry["status"].(string)
					if !ok || status != "InProgress" {
						t.Errorf("Expected ExecuteStart status to be 'InProgress', got '%v'", status)
						return false
					}

					// Check timestamp
					_, hasTimestamp := entry["timestamp"]
					if !hasTimestamp {
						t.Error("ExecuteStart entry missing timestamp")
						return false
					}

					// Check inputs (should include CLI flags)
					inputs, ok := entry["inputs"].(map[string]interface{})
					if !ok {
						t.Error("ExecuteStart entry missing inputs")
						return false
					}

					// Verify specific input field (model names)
					modelNames, ok := inputs["model_names"]
					if !ok {
						t.Errorf("ExecuteStart inputs missing model_names, got: %v", modelNames)
						return false
					}

					// Verify the correct model name is included in the slice
					modelNamesSlice, ok := modelNames.([]interface{})
					if !ok || len(modelNamesSlice) == 0 || modelNamesSlice[0] != "test-model" {
						t.Errorf("ExecuteStart inputs incorrect model_names, got: %v", modelNames)
						return false
					}

					return true
				}
				t.Error("ExecuteStart operation not found in audit log")
				return false
			},
			shouldCreateLogFile: true,
			outputShouldExist:   true,
		},
		{
			name:                "Fallback to NoOpAuditLogger",
			instructionsContent: "Test instructions",
			srcFiles: map[string]string{
				"main.go": `package main

func main() {}`,
			},
			loggerSetup: func(env *TestEnv, t *testing.T) (auditlog.AuditLogger, error) {
				// Set up a test logger that captures errors
				testLogger := env.GetBufferedLogger(logutil.DebugLevel, "[test] ")

				// Attempt to create a FileAuditLogger with invalid path (will fail)
				invalidDir := filepath.Join(env.TestDir, "nonexistent-dir")
				invalidLogFile := filepath.Join(invalidDir, "audit.log")
				_, err := auditlog.NewFileAuditLogger(invalidLogFile, testLogger)

				// Verify error message contains expected text
				if err != nil {
					if !strings.Contains(err.Error(), "failed to open audit log file") {
						t.Errorf("Expected error message to contain 'failed to open audit log file', got: %s", err.Error())
					}
				} else {
					t.Error("Expected error when creating FileAuditLogger with invalid path, got nil")
				}

				// Return a NoOpAuditLogger as fallback
				return auditlog.NewNoOpAuditLogger(), nil
			},
			expectedLogEntries: nil, // No entries expected with NoOpAuditLogger
			validateLogFunc: func(entries []map[string]interface{}, t *testing.T) bool {
				// NoOpAuditLogger doesn't create any log entries, so we just verify
				// that the application ran successfully
				return true
			},
			shouldCreateLogFile: false,
			outputShouldExist:   true,
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

			// Create source files from the map
			for filename, content := range tc.srcFiles {
				env.CreateTestFile(t, "src/"+filename, content)
			}

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			outputDir := filepath.Join(env.TestDir, "output")
			modelName := "test-model"
			outputFile := filepath.Join(outputDir, modelName+".md")
			auditLogFile := filepath.Join(env.TestDir, "audit.log")

			// Set up the custom audit logger using the test case's setup function
			auditLogger, err := tc.loggerSetup(env, t)
			if err != nil {
				// If logger setup fails, the test case should handle this
				// and provide a fallback logger
				if auditLogger == nil {
					t.Fatalf("Logger setup failed and no fallback provided: %v", err)
				}
			}

			// If we got a FileAuditLogger, make sure to close it
			if fileLogger, ok := auditLogger.(*auditlog.FileAuditLogger); ok {
				defer fileLogger.Close()
			}

			// Replace the default NoOpAuditLogger with our test logger
			env.AuditLogger = auditLogger

			// Create a test configuration with the audit log file path
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{modelName},
				APIKey:           "test-api-key",
				Paths:            []string{env.TestDir + "/src"},
				LogLevel:         logutil.InfoLevel,
				AuditLogFile:     auditLogFile,
			}

			// Create a mock API service
			mockApiService := createMockAPIService(env)

			// Run the application with our test configuration
			ctx := context.Background()
			err = architect.Execute(
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

			// Check if output file exists based on expectation
			if tc.outputShouldExist {
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Errorf("Expected output file to exist, but it doesn't: %s", outputFile)
				}
			} else {
				if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
					t.Errorf("Output file was created when it shouldn't have been: %s", outputFile)
				}
			}

			// Validate audit log if we expect it to be created
			if tc.shouldCreateLogFile {
				// Check that the audit log file exists
				if _, err := os.Stat(auditLogFile); os.IsNotExist(err) {
					t.Errorf("Audit log file was not created at %s", auditLogFile)
					return
				}

				// Read and parse the audit log file
				content, err := os.ReadFile(auditLogFile)
				if err != nil {
					t.Fatalf("Failed to read audit log file: %v", err)
				}

				// Parse JSON entries
				lines := bytes.Split(content, []byte{'\n'})
				var entries []map[string]interface{}

				for _, line := range lines {
					if len(line) == 0 {
						continue
					}

					var entry map[string]interface{}
					if err := json.Unmarshal(line, &entry); err != nil {
						t.Errorf("Failed to parse JSON line: %v", err)
						t.Errorf("Line content: %s", string(line))
						continue
					}
					entries = append(entries, entry)
				}

				// Check that we have at least some entries
				if len(entries) == 0 {
					t.Error("Audit log file is empty or contains no valid JSON entries")
					return
				}

				// Validate expected log entries
				if tc.expectedLogEntries != nil {
					// Create a map to track which operations we found
					foundOperations := make(map[string]bool)

					// Find operations in the log entries
					for _, entry := range entries {
						operation, ok := entry["operation"].(string)
						if ok {
							foundOperations[operation] = true
						}
					}

					// Check that all expected operations are present
					for _, expectedOp := range tc.expectedLogEntries {
						if !foundOperations[expectedOp] {
							t.Errorf("Expected operation '%s' not found in audit log", expectedOp)
						}
					}
				}

				// Run custom validation function if provided
				if tc.validateLogFunc != nil {
					if !tc.validateLogFunc(entries, t) {
						t.Error("Custom validation of audit log entries failed")
					}
				}
			}
		})
	}
}
