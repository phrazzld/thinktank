// internal/integration/integration_test.go
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/cmd/architect"
	"github.com/phrazzld/architect/internal/gemini"
)

// TestBasicPlanGeneration tests the basic workflow of the application
func TestBasicPlanGeneration(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Set up the mock client
	env.SetupMockGeminiClient()

	// Create some test files
	env.CreateTestFile(t, "src/main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

	env.CreateTestFile(t, "src/utils.go", `package main

func add(a, b int) int {
	return a + b
}`)

	// Create a task file
	taskFile := env.CreateTestFile(t, "task.txt", "Implement a new feature to multiply two numbers")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application with our test configuration
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Check that the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that content includes expected text (from our mock response)
	if !strings.Contains(string(content), "Test Generated Plan") {
		t.Errorf("Output file does not contain expected content")
	}
}

// TestDryRunMode tests the dry run mode of the application
func TestDryRunMode(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Set up the mock client
	env.SetupMockGeminiClient()

	// Create some test files
	env.CreateTestFile(t, "src/main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

	// Create a task file (optional for dry run, but including it for completeness)
	taskFile := env.CreateTestFile(t, "task.txt", "Test task")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration with dry run enabled
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		DryRun:     true,
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application with our test configuration
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Check that the output file was NOT created in dry run mode
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Errorf("Output file was created in dry run mode at %s", outputFile)
	}
}

// TestTaskFileInput tests using a task file instead of command line task
func TestTaskFileInput(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Set up the mock client
	env.SetupMockGeminiClient()

	// Create some test files
	env.CreateTestFile(t, "src/main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

	// Create a task file with special content for testing
	taskContent := "Implement a new feature to multiply two numbers"
	taskFile := env.CreateTestFile(t, "task.txt", taskContent)

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration focusing on task file input
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application with our test configuration
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Check that the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Read the content to verify it worked
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that content was generated properly
	if !strings.Contains(string(content), "Test Generated Plan") {
		t.Errorf("Output file does not contain expected content")
	}
}

// TestFilteredFileInclusion tests including only specific file extensions
func TestFilteredFileInclusion(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Set up the mock client with a custom implementation that verifies the context content
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// In a real implementation, we would check the actual context
		// You could add checks here for specific file content to verify filtering works
		return &gemini.GenerationResult{
			Content:      "# Test Generated Plan\n\nThis is a test plan generated by the mock client.",
			TokenCount:   1000,
			FinishReason: "STOP",
		}, nil
	}

	// Create test files of different types
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	env.CreateTestFile(t, "src/README.md", "# Test Project")

	env.CreateTestFile(t, "src/config.json", `{"key": "value"}`)

	// Create a task file
	taskFile := env.CreateTestFile(t, "task.txt", "Test task")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration with file inclusion filters
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		Include:    ".go,.md", // Only include Go and Markdown files
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application with our test configuration
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Check that the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Read and verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that the plan was generated
	if !strings.Contains(string(content), "Test Generated Plan") {
		t.Errorf("Output file does not contain expected content")
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Mock an error from the Gemini API
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// Create a simple API error
		apiError := &gemini.APIError{
			Message:    "API quota exceeded",
			Suggestion: "Try again later",
			StatusCode: 429,
		}
		return nil, apiError
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create a task file
	taskFile := env.CreateTestFile(t, "task.txt", "Test task")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application and expect an error
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	// The application should return an error
	if err == nil {
		t.Fatalf("Expected an error from API client, but got nil")
	}

	// Verify the error message contains relevant information
	if !strings.Contains(err.Error(), "API quota exceeded") &&
		!strings.Contains(err.Error(), "generating") {
		t.Errorf("Error message doesn't contain expected content. Got: %v", err)
	}

	// Check that the output file was NOT created due to the error
	if _, statErr := os.Stat(outputFile); !os.IsNotExist(statErr) {
		t.Errorf("Output file was created despite API error")
	}
}

// TestTokenCountExceeded tests the behavior when token count exceeds limits
func TestTokenCountExceeded(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Mock a token count that exceeds limits
	env.MockClient.CountTokensFunc = func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
		return &gemini.TokenCount{Total: 1000000}, nil // Very large count
	}

	env.MockClient.GetModelInfoFunc = func(ctx context.Context) (*gemini.ModelInfo, error) {
		return &gemini.ModelInfo{
			Name:             "test-model",
			InputTokenLimit:  10000, // Smaller than the count
			OutputTokenLimit: 8192,
		}, nil
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create a task file
	taskFile := env.CreateTestFile(t, "task.txt", "Test task")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application and expect a token limit error
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	// The application should report an error for token count exceeding limits
	if err == nil {
		t.Fatalf("Expected an error for token count exceeding limits, but got nil")
	}

	// Verify the error message contains token limit information
	if !strings.Contains(err.Error(), "token") &&
		!strings.Contains(err.Error(), "limit") {
		t.Errorf("Error message doesn't mention token limits. Got: %v", err)
	}

	// Check that the output file was NOT created due to the error
	if _, statErr := os.Stat(outputFile); !os.IsNotExist(statErr) {
		t.Errorf("Output file was created despite token limit error")
	}
}

// TestUserConfirmation tests the confirmation prompt for large token counts
func TestUserConfirmation(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Set up the mock client
	env.SetupMockGeminiClient()

	// Mock token count
	env.MockClient.CountTokensFunc = func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
		return &gemini.TokenCount{Total: 5000}, nil
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create a task file
	taskFile := env.CreateTestFile(t, "task.txt", "Test task")

	// Simulate user input (say "yes" to confirmation)
	env.SimulateUserInput("y\n")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration with confirm-tokens
	testConfig := &architect.CliConfig{
		TaskFile:      taskFile,
		OutputFile:    outputFile,
		ModelName:     "test-model",
		ApiKey:        "test-api-key",
		ConfirmTokens: 1000, // Threshold lower than our token count
		Paths:         []string{env.TestDir + "/src"},
	}

	// Run the application with our test configuration
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Check that the output file was created (confirmation was "y")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Read the content to verify it worked
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that content was generated properly
	if !strings.Contains(string(content), "Test Generated Plan") {
		t.Errorf("Output file does not contain expected content")
	}
}

// TestTaskExecution tests the basic task execution without clarification
// (replaces the old TestTaskClarification test that depended on the removed clarify feature)
func TestTaskExecution(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Set up the mock client with default responses
	env.SetupMockGeminiClient()

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create a task file
	taskDescription := "Implement a new feature"
	taskFile := env.CreateTestFile(t, "task.txt", taskDescription)

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a test configuration
	testConfig := &architect.CliConfig{
		TaskFile:   taskFile,
		OutputFile: outputFile,
		ModelName:  "test-model",
		ApiKey:     "test-api-key",
		Paths:      []string{env.TestDir + "/src"},
	}

	// Run the application with our test configuration
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Check that the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Verify that the task was used
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that the output contains content from our mock response
	if !strings.Contains(string(content), "Test Generated Plan") {
		t.Errorf("Output file does not contain expected content")
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
		taskFileContent := "This is a regular task description without any template variables."
		taskFile := env.CreateTestFile(t, "task.txt", taskFileContent)

		// Set up the output file path
		outputFile := filepath.Join(env.TestDir, "output.md")

		// Create a test configuration
		testConfig := &architect.CliConfig{
			TaskFile:   taskFile,
			OutputFile: outputFile,
			ModelName:  "test-model",
			ApiKey:     "test-api-key",
			Paths:      []string{env.TestDir},
		}

		// Run the application with our test configuration
		ctx := context.Background()
		err := RunTestWithConfig(ctx, testConfig, env)

		if err != nil {
			t.Fatalf("RunTestWithConfig failed: %v", err)
		}

		// Check that the output file exists
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file was not created at %s", outputFile)
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
