package prompt_test

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/thinktank/prompt"
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
			expected: "func test() { if (x < y) { return } }",
		},
		{
			name:     "Single greater-than character",
			input:    "func test() { if (x > y) { return } }",
			expected: "func test() { if (x > y) { return } }",
		},
		{
			name:     "Both comparison operators",
			input:    "func validate(x, y, z int) bool { return x < y && y > z }",
			expected: "func validate(x, y, z int) bool { return x < y && y > z }",
		},
		{
			name:     "HTML/XML tags",
			input:    "<div>Some HTML content</div>",
			expected: "<div>Some HTML content</div>",
		},
		{
			name:     "Nested tags",
			input:    "<outer><inner>Nested tags</inner></outer>",
			expected: "<outer><inner>Nested tags</inner></outer>",
		},
		{
			name:     "Multiple occurrences mixed with code",
			input:    "if (x < 10 && y > 20) { return <r>value</r> }",
			expected: "if (x < 10 && y > 20) { return <r>value</r> }",
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
			result := prompt.EscapeContent(tt.input)
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
				// Check for no XML escaping (content should be preserved as-is)
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<r>value</r>") {
						t.Error("XML tags in content were not properly preserved")
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
					if !strings.Contains(result, "<context>") && strings.Contains(result, "</context>") {
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
			result := prompt.StitchPrompt(tt.instructions, tt.contextFiles)

			// Run all checks for this test case
			for _, check := range tt.checks {
				check(t, result)
			}
		})
	}
}
