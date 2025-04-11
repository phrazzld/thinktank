// internal/integration/xml_integration_test.go
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// TestXMLPromptStructure tests the new XML-structured prompt approach
func TestXMLPromptStructure(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test source files
	env.CreateTestFile(t, "src/main.go", `package main

func main() {
	println("Hello, world!")
}`)

	env.CreateTestFile(t, "src/helper.go", `package main

// Helper function with special characters < > to test XML escaping
func helper(a, b int) int {
	if a < b {
		return a
	}
	return b
}`)

	// Create an instructions file (not a task file)
	instructions := "Implement a feature to multiply two numbers"
	instructionsFile := env.CreateTestFile(t, "instructions.md", instructions)

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a variable to capture the prompt for verification
	var capturedPrompt string

	// Set up the mock client with custom implementation to capture the prompt
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// Save the prompt for verification
		capturedPrompt = prompt

		// Return a mock response
		return &gemini.GenerationResult{
			Content:    "# Generated Plan\n\nImplemented as requested.",
			TokenCount: 100,
		}, nil
	}

	// Create a test configuration using the new InstructionsFile
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputFile:       outputFile,
		ModelName:        "test-model",
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
	}

	// Run the application
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	// Check for errors
	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Verify that the output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Now verify the XML structure of the captured prompt
	if capturedPrompt == "" {
		t.Fatal("No prompt was captured")
	}

	// Check for the instructions tag
	if !strings.Contains(capturedPrompt, "<instructions>") || !strings.Contains(capturedPrompt, "</instructions>") {
		t.Errorf("Prompt is missing <instructions> tags: %s", capturedPrompt)
	}

	// Check that the instructions content is included
	if !strings.Contains(capturedPrompt, instructions) {
		t.Errorf("Prompt doesn't contain the expected instructions content: %s", capturedPrompt)
	}

	// Check for the context tag
	if !strings.Contains(capturedPrompt, "<context>") || !strings.Contains(capturedPrompt, "</context>") {
		t.Errorf("Prompt is missing <context> tags: %s", capturedPrompt)
	}

	// Check for file paths
	if !strings.Contains(capturedPrompt, "main.go") || !strings.Contains(capturedPrompt, "helper.go") {
		t.Errorf("Prompt doesn't contain expected file paths: %s", capturedPrompt)
	}

	// Check for XML escaping of special characters
	if !strings.Contains(capturedPrompt, "&lt;") || !strings.Contains(capturedPrompt, "&gt;") {
		t.Errorf("Prompt doesn't have properly escaped XML characters: %s", capturedPrompt)
	}

	// Read the output file to verify the content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify the output contains the expected content
	if !strings.Contains(string(content), "Generated Plan") {
		t.Errorf("Output file does not contain expected content")
	}
}

// TestEmptyInstructionsFile tests handling empty instructions files
func TestEmptyInstructionsFile(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test source file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create an empty instructions file
	instructionsFile := env.CreateTestFile(t, "empty_instructions.md", "")

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Create a variable to capture the prompt for verification
	var capturedPrompt string

	// Set up the mock client
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		capturedPrompt = prompt
		return &gemini.GenerationResult{
			Content:    "# Generated Plan\n\nPlan with empty instructions.",
			TokenCount: 100,
		}, nil
	}

	// Create a test configuration
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputFile:       outputFile,
		ModelName:        "test-model",
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
	}

	// Run the application
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	// Check for errors
	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Verify that the output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Verify the prompt structure with empty instructions
	if !strings.Contains(capturedPrompt, "<instructions>\n</instructions>") {
		t.Errorf("Empty instructions not handled correctly: %s", capturedPrompt)
	}
}

// TestComplexXMLStructure tests more complex XML structure scenarios
func TestComplexXMLStructure(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create source files with XML-like content to test escaping
	env.CreateTestFile(t, "src/component.jsx", `
function Component() {
  return (
    <div className="container">
      <h1>Hello, world!</h1>
      {items.map(item => (
        <div key={item.id}>
          {item.name}: {item.value > 10 ? "High" : "Low"}
        </div>
      ))}
    </div>
  );
}
`)

	env.CreateTestFile(t, "src/template.html", `
<!DOCTYPE html>
<html>
<head>
  <title>Test Template</title>
</head>
<body>
  <div id="app">
    <h1>{{ title }}</h1>
    <p>This is a test with <!-- comments --> and special characters: < > & " '</p>
  </div>
</body>
</html>
`)

	// Create instructions with tags that should be preserved (not escaped)
	instructions := `# Feature Request
Implement a new <Component> that displays data in a <table> format.
The component should handle the following edge cases:
1. When data.length === 0, show "No data available"
2. When data.length > 1000, paginate with <Pagination> component

## Requirements
- Use existing <DataTable> if possible
- Ensure all data is properly escaped before rendering
- Add proper error handling for API failures
`

	// Create the instructions file
	instructionsFile := env.CreateInstructionsFile(t, instructions)

	// Set up the output file path
	outputFile := filepath.Join(env.TestDir, "output.md")

	// Set up the XML validating client with expected content fragments
	env.SetupXMLValidatingClient(t, "component.jsx", "template.html", "&lt;div", "&lt;html&gt;")

	// Create a test configuration using the new InstructionsFile
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputFile:       outputFile,
		ModelName:        "test-model",
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
	}

	// Run the application
	ctx := context.Background()
	err := RunTestWithConfig(ctx, testConfig, env)

	// Check for errors
	if err != nil {
		t.Fatalf("RunTestWithConfig failed: %v", err)
	}

	// Verify that the output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputFile)
	}

	// Read the output file to verify the content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify the output contains the expected content
	if !strings.Contains(string(content), "Validated XML Structure Plan") {
		t.Errorf("Output file does not contain expected content: %s", string(content))
	}
}
