// internal/integration/xml_integration_test.go
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// xmlTestCase defines a test case for XML prompt structure tests
type xmlTestCase struct {
	name                      string
	sourceFiles               map[string]string
	instructionsContent       string
	expectedPromptContains    []string
	expectedPromptNotContains []string
	setupCustomMock           func(t *testing.T, env *TestEnv) (string, error)
	expectedOutputContains    string
	expectedError             bool
}

// TestXMLPromptFeatures tests various aspects of XML prompt structure
func TestXMLPromptFeatures(t *testing.T) {
	// Define common helper function source code
	helperGoSrc := `package main

// Helper function with special characters < > to test XML escaping
func helper(a, b int) int {
	if a < b {
		return a
	}
	return b
}`

	mainGoSrc := `package main

func main() {
	println("Hello, world!")
}`

	componentJSX := `
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
}`

	templateHTML := `
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
</html>`

	complexInstructions := `# Feature Request
Implement a new <Component> that displays data in a <table> format.
The component should handle the following edge cases:
1. When data.length === 0, show "No data available"
2. When data.length > 1000, paginate with <Pagination> component

## Requirements
- Use existing <DataTable> if possible
- Ensure all data is properly escaped before rendering
- Add proper error handling for API failures
`

	// Define test cases
	testCases := []xmlTestCase{
		{
			name: "Basic XML Structure",
			sourceFiles: map[string]string{
				"src/main.go":   mainGoSrc,
				"src/helper.go": helperGoSrc,
			},
			instructionsContent: "Implement a feature to multiply two numbers",
			expectedPromptContains: []string{
				"<instructions>", "</instructions>",
				"<context>", "</context>",
				"main.go", "helper.go",
				"// No longer checking for XML escaping",
			},
			setupCustomMock: func(t *testing.T, env *TestEnv) (string, error) {
				var capturedPrompt string
				env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					capturedPrompt = prompt
					return &gemini.GenerationResult{
						Content:    "# Generated Plan\n\nImplemented as requested.",
						TokenCount: 100,
					}, nil
				}
				return capturedPrompt, nil
			},
			expectedOutputContains: "Generated Plan",
		},
		{
			name: "Empty Instructions",
			sourceFiles: map[string]string{
				"src/main.go": `package main

func main() {}`,
			},
			instructionsContent: "",
			expectedPromptContains: []string{
				"<instructions>\n</instructions>",
			},
			setupCustomMock: func(t *testing.T, env *TestEnv) (string, error) {
				var capturedPrompt string
				env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					capturedPrompt = prompt
					return &gemini.GenerationResult{
						Content:    "# Generated Plan\n\nPlan with empty instructions.",
						TokenCount: 100,
					}, nil
				}
				return capturedPrompt, nil
			},
			expectedOutputContains: "Generated Plan",
		},
		{
			name: "Complex XML Structure",
			sourceFiles: map[string]string{
				"src/component.jsx": componentJSX,
				"src/template.html": templateHTML,
			},
			instructionsContent:    complexInstructions,
			expectedPromptContains: []string{
				// XML validation is handled by SetupXMLValidatingClient
			},
			setupCustomMock: func(t *testing.T, env *TestEnv) (string, error) {
				// Set up the XML validating client with expected content fragments
				env.SetupXMLValidatingClient(t, "component.jsx", "template.html")
				return "", nil
			},
			expectedOutputContains: "Validated XML Structure Plan",
		},
	}

	// Execute each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Create source files
			for path, content := range tc.sourceFiles {
				env.CreateTestFile(t, path, content)
			}

			// Create instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory and model-specific output file path
			modelName := "test-model"
			outputDir := filepath.Join(env.TestDir, "output")
			outputFile := filepath.Join(outputDir, modelName+".md")

			// Set up mock client
			var capturedPrompt string
			var setupErr error
			if tc.setupCustomMock != nil {
				capturedPrompt, setupErr = tc.setupCustomMock(t, env)
				if setupErr != nil && !tc.expectedError {
					t.Fatalf("Failed to set up custom mock: %v", setupErr)
				}
			}

			// Create a test configuration
			testConfig := &config.CliConfig{
				InstructionsFile: instructionsFile,
				OutputDir:        outputDir,
				ModelNames:       []string{modelName},
				APIKey:           "test-api-key",
				Paths:            []string{env.TestDir + "/src"},
				LogLevel:         logutil.InfoLevel,
			}

			// Run the application
			ctx := context.Background()
			err := RunTestWithConfig(ctx, testConfig, env)

			// Check for expected errors
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			} else if err != nil {
				t.Fatalf("RunTestWithConfig failed: %v", err)
			}

			// Verify the output file was created
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Errorf("Output file was not created at %s", outputFile)
				return
			}

			// Check the prompt structure if a prompt was captured
			if capturedPrompt != "" {
				for _, expected := range tc.expectedPromptContains {
					if !strings.Contains(capturedPrompt, expected) {
						t.Errorf("Prompt doesn't contain expected content: %s", expected)
					}
				}

				for _, unexpected := range tc.expectedPromptNotContains {
					if strings.Contains(capturedPrompt, unexpected) {
						t.Errorf("Prompt contains unexpected content: %s", unexpected)
					}
				}
			}

			// Verify the output content
			if tc.expectedOutputContains != "" {
				content, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				if !strings.Contains(string(content), tc.expectedOutputContains) {
					t.Errorf("Output file does not contain expected content: %s", tc.expectedOutputContains)
				}
			}
		})
	}
}
