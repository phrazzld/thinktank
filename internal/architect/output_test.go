// Package architect_test is used for testing the internal/architect package
package architect_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/logutil"
)

// TestEscapeContent tests the XML escaping helper function
func TestEscapeContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No escaping needed",
			input:    "Normal content without special characters",
			expected: "Normal content without special characters",
		},
		{
			name:     "Single less-than character",
			input:    "func test() { if (x < y) { return } }",
			expected: "func test() { if (x &lt; y) { return } }",
		},
		{
			name:     "Single greater-than character",
			input:    "func test() { if (x > y) { return } }",
			expected: "func test() { if (x &gt; y) { return } }",
		},
		{
			name:     "Both comparison operators",
			input:    "func validate(x, y, z int) bool { return x < y && y > z }",
			expected: "func validate(x, y, z int) bool { return x &lt; y && y &gt; z }",
		},
		{
			name:     "HTML/XML tags",
			input:    "<div>Some HTML content</div>",
			expected: "&lt;div&gt;Some HTML content&lt;/div&gt;",
		},
		{
			name:     "Nested tags",
			input:    "<outer><inner>Nested tags</inner></outer>",
			expected: "&lt;outer&gt;&lt;inner&gt;Nested tags&lt;/inner&gt;&lt;/outer&gt;",
		},
		{
			name:     "Multiple occurrences mixed with code",
			input:    "if (x < 10 && y > 20) { return <r>value</r> }",
			expected: "if (x &lt; 10 && y &gt; 20) { return &lt;r&gt;value&lt;/r&gt; }",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Already escaped content",
			input:    "function with &lt;param&gt;",
			expected: "function with &lt;param&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := architect.EscapeContent(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestStitchPrompt tests the main prompt stitching function
func TestStitchPrompt(t *testing.T) {
	// Create common test files
	standardFile := fileutil.FileMeta{
		Path:    "/standard/file.go",
		Content: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, world!\")\n}",
	}

	fileWithTags := fileutil.FileMeta{
		Path:    "/tagged/file.go",
		Content: "func test() { if (a < b || c > d) { return <r>value</r> } }",
	}

	emptyFile := fileutil.FileMeta{
		Path:    "/empty/file.go",
		Content: "",
	}

	// Define test cases
	tests := []struct {
		name         string
		instructions string
		contextFiles []fileutil.FileMeta
		checks       []func(t *testing.T, result string)
	}{
		{
			name:         "Standard case - instructions with multiple files",
			instructions: "Standard test instructions for multiple files",
			contextFiles: []fileutil.FileMeta{standardFile, fileWithTags},
			checks: []func(t *testing.T, result string){
				// Check instructions block
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<instructions>\nStandard test instructions for multiple files\n</instructions>") {
						t.Error("Instructions section not formatted correctly")
					}
				},
				// Check context block
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<context>") || !strings.Contains(result, "</context>") {
						t.Error("Missing context section tags")
					}
				},
				// Check file paths
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<path>/standard/file.go</path>") {
						t.Error("Missing or incorrectly formatted standard file path tag")
					}
					if !strings.Contains(result, "<path>/tagged/file.go</path>") {
						t.Error("Missing or incorrectly formatted tagged file path tag")
					}
				},
				// Check file content
				func(t *testing.T, result string) {
					if !strings.Contains(result, "package main") {
						t.Error("Missing content from standard file")
					}
					if !strings.Contains(result, "Hello, world!") {
						t.Error("Missing specific content from standard file")
					}
				},
				// Check XML escaping
				func(t *testing.T, result string) {
					if !strings.Contains(result, "&lt;") || !strings.Contains(result, "&gt;") {
						t.Error("XML special characters were not escaped properly")
					}
					if !strings.Contains(result, "&lt;r&gt;value&lt;/r&gt;") {
						t.Error("XML tags in content were not properly escaped")
					}
				},
			},
		},
		{
			name:         "Empty instructions",
			instructions: "",
			contextFiles: []fileutil.FileMeta{standardFile},
			checks: []func(t *testing.T, result string){
				func(t *testing.T, result string) {
					// Check that instructions tags exist even with empty content
					if !strings.Contains(result, "<instructions>\n</instructions>") {
						t.Error("Empty instructions not properly formatted")
					}
				},
				func(t *testing.T, result string) {
					// Check that file content is still included
					if !strings.Contains(result, "package main") {
						t.Error("Missing file content despite empty instructions")
					}
				},
			},
		},
		{
			name:         "Empty context files list",
			instructions: "Instructions with empty context",
			contextFiles: []fileutil.FileMeta{},
			checks: []func(t *testing.T, result string){
				func(t *testing.T, result string) {
					// Verify instructions are included
					if !strings.Contains(result, "Instructions with empty context") {
						t.Error("Missing instructions content")
					}
				},
				func(t *testing.T, result string) {
					// Check that context tags exist even with no files
					if !strings.Contains(result, "<context>\n</context>") {
						t.Error("Empty context not properly formatted")
					}
				},
			},
		},
		{
			name:         "Empty file content",
			instructions: "Instructions with empty file content",
			contextFiles: []fileutil.FileMeta{emptyFile},
			checks: []func(t *testing.T, result string){
				func(t *testing.T, result string) {
					// Verify file path is included
					if !strings.Contains(result, "<path>/empty/file.go</path>") {
						t.Error("Missing path tag for empty file")
					}
				},
				func(t *testing.T, result string) {
					// Ensure there's an appropriate representation of empty content
					if !strings.Contains(result, "<path>/empty/file.go</path>\n\n\n") {
						t.Error("Empty file content not properly handled")
					}
				},
			},
		},
		{
			name:         "Instructions with XML-like tags",
			instructions: "Instructions with <tags> that should NOT be escaped",
			contextFiles: []fileutil.FileMeta{standardFile},
			checks: []func(t *testing.T, result string){
				func(t *testing.T, result string) {
					// Verify instructions are preserved as-is (not escaped)
					if !strings.Contains(result, "Instructions with <tags> that should NOT be escaped") {
						t.Error("Instructions content was incorrectly modified")
					}
				},
				func(t *testing.T, result string) {
					// Double-check that the file content is still properly escaped
					if !strings.Contains(result, "fmt.Println") {
						t.Error("Missing expected file content")
					}
				},
			},
		},
		{
			name:         "File with special characters in path",
			instructions: "Testing path with special characters",
			contextFiles: []fileutil.FileMeta{
				{Path: "/path/with/<special>/chars.go", Content: "special path content"},
			},
			checks: []func(t *testing.T, result string){
				func(t *testing.T, result string) {
					// Verify path is correctly included without escaping
					if !strings.Contains(result, "<path>/path/with/<special>/chars.go</path>") {
						t.Error("Path with special characters not properly handled")
					}
				},
				func(t *testing.T, result string) {
					// Check content is included
					if !strings.Contains(result, "special path content") {
						t.Error("Missing content from file with special path")
					}
				},
			},
		},
	}

	// Run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the stitched prompt result
			result := architect.StitchPrompt(tt.instructions, tt.contextFiles)

			// Debug output for troubleshooting (commented out for production)
			// t.Logf("Result: %s", result)

			// Run all checks for this test case
			for _, check := range tt.checks {
				check(t, result)
			}
		})
	}
}

// TestSaveToFile tests the SaveToFile method
func TestSaveToFile(t *testing.T) {
	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")

	// Create a file writer
	fileWriter := architect.NewFileWriter(logger)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define test cases
	tests := []struct {
		name       string
		content    string
		outputFile string
		setupFunc  func() // Function to run before test
		cleanFunc  func() // Function to run after test
		wantErr    bool
	}{
		{
			name:       "Valid file path - absolute",
			content:    "Test content",
			outputFile: filepath.Join(tempDir, "test_output.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Valid file path - relative",
			content:    "Test content with relative path",
			outputFile: "test_output_relative.md",
			setupFunc:  func() {},
			cleanFunc: func() {
				// Clean up relative path file
				cwd, _ := os.Getwd()
				os.Remove(filepath.Join(cwd, "test_output_relative.md"))
			},
			wantErr: false,
		},
		{
			name:       "Empty content",
			content:    "",
			outputFile: filepath.Join(tempDir, "empty_file.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Long content",
			content:    strings.Repeat("Long content test ", 1000), // ~ 18KB of content
			outputFile: filepath.Join(tempDir, "long_file.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Non-existent directory",
			content:    "Test content",
			outputFile: filepath.Join(tempDir, "non-existent", "test_output.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run setup function
			tc.setupFunc()

			// Save to file
			err := fileWriter.SaveToFile(tc.content, tc.outputFile)

			// Run cleanup function
			defer tc.cleanFunc()

			// Check error
			if (err != nil) != tc.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Skip file validation for expected errors
			if tc.wantErr {
				return
			}

			// Determine output path for validation
			outputPath := tc.outputFile
			if !filepath.IsAbs(outputPath) {
				cwd, _ := os.Getwd()
				outputPath = filepath.Join(cwd, outputPath)
			}

			// Verify file was created and content matches
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			if string(content) != tc.content {
				t.Errorf("File content = %v, want %v", string(content), tc.content)
			}
		})
	}
}
