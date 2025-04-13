package e2e

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestFileFiltering tests the file filtering flags
func TestFileFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files of different types
	env.CreateTestFile("src/main.go", `package main
func main() {}`)

	env.CreateTestFile("src/utils.go", `package main
func add(a, b int) int { return a + b }`)

	env.CreateTestFile("src/README.md", "# Test Project")

	env.CreateTestFile("src/data.json", `{"key": "value"}`)

	env.CreateTestFile("src/ignored.tmp", "Temporary file to be ignored")

	// Create an ignored directory with some files
	env.CreateTestFile("src/node_modules/package.json", `{"name": "test"}`)

	testCases := []struct {
		name                string
		include             string
		exclude             string
		excludeNames        string
		expectedInOutput    []string
		notExpectedInOutput []string
		dryRun              bool
	}{
		{
			name:                "Include Go files only",
			include:             "*.go",
			exclude:             "",
			excludeNames:        "",
			expectedInOutput:    []string{"main.go", "utils.go"},
			notExpectedInOutput: []string{"README.md", "data.json", "ignored.tmp", "node_modules"},
			dryRun:              true,
		},
		{
			name:                "Include Go and Markdown files",
			include:             "*.go,*.md",
			exclude:             "",
			excludeNames:        "",
			expectedInOutput:    []string{"main.go", "utils.go", "README.md"},
			notExpectedInOutput: []string{"data.json", "ignored.tmp", "node_modules"},
			dryRun:              true,
		},
		{
			name:                "Exclude temporary files",
			include:             "",
			exclude:             "*.tmp",
			excludeNames:        "",
			expectedInOutput:    []string{"main.go", "utils.go", "README.md", "data.json"},
			notExpectedInOutput: []string{"ignored.tmp"},
			dryRun:              true,
		},
		{
			name:                "Exclude directory by name",
			include:             "",
			exclude:             "",
			excludeNames:        "node_modules",
			expectedInOutput:    []string{"main.go", "utils.go", "README.md", "data.json", "ignored.tmp"},
			notExpectedInOutput: []string{"node_modules/package.json"},
			dryRun:              true,
		},
		{
			name:                "Combine include and exclude",
			include:             "*.go,*.md",
			exclude:             "utils*",
			excludeNames:        "node_modules",
			expectedInOutput:    []string{"main.go", "README.md"},
			notExpectedInOutput: []string{"utils.go", "data.json", "ignored.tmp", "node_modules"},
			dryRun:              true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create instructions file for this test case
			instructionsFile := env.CreateTestFile("instructions-"+tc.name+".md", "Test filtering: "+tc.name)

			// Set up flags
			flags := env.DefaultFlags
			flags.Instructions = instructionsFile
			flags.Include = tc.include
			flags.Exclude = tc.exclude
			flags.ExcludeNames = tc.excludeNames
			flags.DryRun = tc.dryRun

			// Run the architect binary
			stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
			if err != nil {
				t.Fatalf("Failed to run architect: %v", err)
			}

			// Verify exit code
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
				t.Logf("Stdout: %s", stdout)
				t.Logf("Stderr: %s", stderr)
			}

			// Combine stdout and stderr for checking
			combinedOutput := stdout + stderr

			// Check for expected files in output
			for _, expectedFile := range tc.expectedInOutput {
				if !strings.Contains(combinedOutput, expectedFile) {
					t.Errorf("Expected file %s not found in output", expectedFile)
				}
			}

			// Check for files that should not be in output
			for _, notExpectedFile := range tc.notExpectedInOutput {
				if strings.Contains(combinedOutput, notExpectedFile) {
					t.Errorf("Unexpected file %s found in output", notExpectedFile)
				}
			}
		})
	}
}

// TestMultipleDirectories tests filtering with multiple input directories
func TestMultipleDirectories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files in multiple directories
	env.CreateTestFile("dir1/file1.go", `package main
func main() {}`)

	env.CreateTestFile("dir2/file2.go", `package main
func add(a, b int) int { return a + b }`)

	env.CreateTestFile("dir3/file3.md", "# Test Project")

	// Create instructions file
	instructionsFile := env.CreateTestFile("instructions.md", "Test multiple directories")

	// Set up flags
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.DryRun = true

	// Run the architect binary with multiple directories
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
		filepath.Join(env.TempDir, "dir1"),
		filepath.Join(env.TempDir, "dir2"),
	})
	if err != nil {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Combine stdout and stderr for checking
	combinedOutput := stdout + stderr

	// Check for expected files in output
	expectedFiles := []string{"file1.go", "file2.go"}
	for _, expectedFile := range expectedFiles {
		if !strings.Contains(combinedOutput, expectedFile) {
			t.Errorf("Expected file %s not found in output", expectedFile)
		}
	}

	// Check that file3.md is not included (it's in dir3 which wasn't specified)
	if strings.Contains(combinedOutput, "file3.md") {
		t.Errorf("Unexpected file file3.md found in output")
	}
}
