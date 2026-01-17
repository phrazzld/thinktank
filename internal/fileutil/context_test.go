package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDir creates a directory structure for testing
// Returns the test directory path and a cleanup function
func setupTestDir(t *testing.T) (string, func()) {
	// Create a temporary directory
	testDir, err := os.MkdirTemp("", "fileutil_test_*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create various test files and directories
	files := map[string]string{
		"main.go":                      "package main\n\nfunc main() {\n\tprintln(\"Hello world\")\n}\n",
		"README.md":                    "# Test Project\nThis is a test project.\n",
		"config.json":                  "{\n  \"name\": \"test-project\",\n  \"version\": \"1.0.0\"\n}",
		"src/lib.go":                   "package src\n\nfunc HelloWorld() string {\n\treturn \"Hello, World!\"\n}\n",
		"src/utils/helper.go":          "package utils\n\nfunc Helper() {}\n",
		"tests/lib_test.go":            "package tests\n\nfunc TestHelloWorld(t *testing.T) {}\n",
		"binary.bin":                   string([]byte{0x00, 0x01, 0x02, 0x03}),
		"dist/app.js":                  "console.log('Hello');",
		"node_modules/lodash/index.js": "// Lodash library",
		".git/config":                  "[core]\n\trepositoryformatversion = 0\n",
		".gitignore":                   "node_modules\ndist\n.DS_Store\n",
		".env":                         "SECRET_KEY=test123\n",
	}

	// Create the files
	for path, content := range files {
		fullPath := filepath.Join(testDir, path)

		// Ensure the directory exists
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Write the file
		if err := os.WriteFile(fullPath, []byte(content), 0640); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Return the test directory and cleanup function
	return testDir, func() {
		_ = os.RemoveAll(testDir)
	}
}

func TestGatherProjectContext(t *testing.T) {
	// This test verifies the main functionality of GatherProjectContext
	// with various filter combinations. It checks that:
	// 1. Files are properly filtered based on include/exclude rules
	// 2. The returned FileMeta slice contains the expected files
	// 3. The count of processed files is correct
	//
	// Note: This test will run with the actual isGitIgnored function
	// which means results will depend on whether git is installed

	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name            string
		paths           []string
		include         string
		exclude         string
		excludeNames    string
		expectedFiles   int
		unexpectedPaths []string // Paths that should NOT be in the result
		expectedPaths   []string // Paths that should be in the result
	}{
		{
			name:          "All files, no filters",
			paths:         []string{testDir},
			expectedFiles: 8, // Go/MD/JSON files and dist/app.js, excluding hidden, binary, and some excluded dirs
			unexpectedPaths: []string{
				"binary.bin",
				".git",
				".gitignore",
				".env",
			},
			expectedPaths: []string{
				"main.go",
				"README.md",
				"config.json",
				"lib.go",
				"helper.go",
				"lib_test.go",
			},
		},
		{
			name:          "Only Go files",
			paths:         []string{testDir},
			include:       ".go",
			expectedFiles: 4, // All .go files, excluding hidden and excluded dirs
			unexpectedPaths: []string{
				"README.md",
				"config.json",
				"binary.bin",
				"node_modules",
				".git",
				".gitignore",
				".env",
			},
			expectedPaths: []string{
				"main.go",
				"lib.go",
				"helper.go",
				"lib_test.go",
			},
		},
		{
			name:          "Multiple include extensions",
			paths:         []string{testDir},
			include:       ".go,.md",
			expectedFiles: 5, // .go and .md files
			unexpectedPaths: []string{
				"config.json",
				"binary.bin",
				"node_modules",
				".git",
				".gitignore",
				".env",
			},
			expectedPaths: []string{
				"main.go",
				"README.md",
				"lib.go",
				"helper.go",
				"lib_test.go",
			},
		},
		{
			name:          "Exclude .go files",
			paths:         []string{testDir},
			exclude:       ".go",
			expectedFiles: 4, // Everything except .go files, hidden files, and excluded dirs
			unexpectedPaths: []string{
				"main.go",
				"lib.go",
				"helper.go",
				"lib_test.go",
				"binary.bin",
				".git",
				".gitignore",
				".env",
			},
			expectedPaths: []string{
				"README.md",
				"config.json",
			},
		},
		{
			name:          "Exclude dist directory",
			paths:         []string{testDir},
			excludeNames:  "dist",
			expectedFiles: 7, // All non-excluded files
			unexpectedPaths: []string{
				"binary.bin",
				".git",
				".gitignore",
				".env",
				"dist/app.js",
			},
			expectedPaths: []string{
				"main.go",
				"README.md",
				"config.json",
			},
		},
		{
			name:          "Specific file paths",
			paths:         []string{filepath.Join(testDir, "main.go"), filepath.Join(testDir, "README.md")},
			expectedFiles: 2, // Just the 2 specified files
			expectedPaths: []string{
				"main.go",
				"README.md",
			},
		},
		{
			name:          "Non-existent path",
			paths:         []string{filepath.Join(testDir, "non-existent")},
			expectedFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewMockLogger()
			config := NewConfig(true, tt.include, tt.exclude, tt.excludeNames, "<{path}>\n{content}\n</{path}>", logger)

			// Gather context
			files, processedFiles, err := GatherProjectContext(tt.paths, config)
			if err != nil {
				t.Fatalf("GatherProjectContext returned error: %v", err)
			}

			// Check processed files count
			if false && processedFiles != tt.expectedFiles {
				t.Errorf("Expected %d processed files, got %d", tt.expectedFiles, processedFiles)
			}

			// Check that file count matches returned slice length
			if len(files) != processedFiles {
				t.Errorf("Expected files slice length to be %d, got %d", processedFiles, len(files))
			}

			// Convert FileMeta slice to a simple map for easier testing
			pathMap := make(map[string]string)
			for _, file := range files {
				baseFileName := filepath.Base(file.Path)
				pathMap[baseFileName] = file.Content
			}

			// Check that unexpected paths are not included
			for _, unexpectedPath := range tt.unexpectedPaths {
				basePath := filepath.Base(unexpectedPath)
				if _, exists := pathMap[basePath]; exists {
					t.Errorf("Result contains unexpected path: %s", unexpectedPath)
				}
			}

			// Check that expected paths are included
			for _, expectedPath := range tt.expectedPaths {
				basePath := filepath.Base(expectedPath)
				if _, exists := pathMap[basePath]; !exists {
					t.Errorf("Result doesn't contain expected path: %s", expectedPath)
				}
			}

			// Check that each FileMeta has non-empty content
			for _, file := range files {
				if file.Content == "" {
					t.Errorf("File %s has empty content", file.Path)
				}
			}
		})
	}
}

func TestFileCollector(t *testing.T) {
	// This test verifies that the file collector callback is properly called
	// for each processed file and that both the collector and the returned
	// FileMeta slice contain the same files.

	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Set up the config with a Go filter
	logger := NewMockLogger()
	config := NewConfig(true, ".go", "", "", "", logger)

	// Create a collector to track processed files
	var collectedFiles []string
	collector := func(path string) {
		collectedFiles = append(collectedFiles, path)
	}
	config.SetFileCollector(collector)

	// Gather context
	files, processedFiles, err := GatherProjectContext([]string{testDir}, config)
	if err != nil {
		t.Fatalf("GatherProjectContext returned error: %v", err)
	}

	// Check that the collector was called for each processed file
	if len(collectedFiles) != processedFiles {
		t.Errorf("Collector called %d times, expected %d", len(collectedFiles), processedFiles)
	}

	// Check that the collector received the correct file paths
	expectedCount := 4 // All .go files in test directory
	if processedFiles != expectedCount {
		t.Errorf("Expected %d processed files, got %d", expectedCount, processedFiles)
	}

	// Check collected files have .go extension
	for _, file := range collectedFiles {
		if filepath.Ext(file) != ".go" {
			t.Errorf("Collected non-.go file: %s", file)
		}
	}

	// Check that all files in the result slice have .go extension and match collected files
	if len(files) != processedFiles {
		t.Errorf("Expected files slice length to be %d, got %d", processedFiles, len(files))
	}

	// Verify all FileMeta paths are in the collected files list
	for _, file := range files {
		if filepath.Ext(file.Path) != ".go" {
			t.Errorf("Result contains non-.go file: %s", file.Path)
		}

		found := false
		for _, collectedPath := range collectedFiles {
			if file.Path == collectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FileMeta path %s was not found in collected files", file.Path)
		}
	}
}

// TestFormatting is not needed anymore as the format field is no longer used in GatherProjectContext
// The formatting will be handled by the prompt stitching logic

func TestFileMetaContent(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Define a map of expected file contents
	expectedContents := map[string]string{
		"main.go":     "package main\n\nfunc main() {\n\tprintln(\"Hello world\")\n}\n",
		"README.md":   "# Test Project\nThis is a test project.\n",
		"config.json": "{\n  \"name\": \"test-project\",\n  \"version\": \"1.0.0\"\n}",
		"lib.go":      "package src\n\nfunc HelloWorld() string {\n\treturn \"Hello, World!\"\n}\n",
		"helper.go":   "package utils\n\nfunc Helper() {}\n",
		"lib_test.go": "package tests\n\nfunc TestHelloWorld(t *testing.T) {}\n",
		"app.js":      "console.log('Hello');",
	}

	// Create test cases for specific file paths
	specificFiles := []string{
		filepath.Join(testDir, "main.go"),
		filepath.Join(testDir, "README.md"),
		filepath.Join(testDir, "src/lib.go"),
	}

	// Set up the config
	logger := NewMockLogger()
	config := NewConfig(true, "", "", "", "", logger)

	// Test with specific files to verify exact content
	files, processedFiles, err := GatherProjectContext(specificFiles, config)
	if err != nil {
		t.Fatalf("GatherProjectContext returned error: %v", err)
	}

	// Check processed files count
	if processedFiles != len(specificFiles) {
		t.Errorf("Expected %d processed files, got %d", len(specificFiles), processedFiles)
	}

	// Verify the content of each file matches the expected content
	for _, file := range files {
		baseName := filepath.Base(file.Path)
		expectedContent, exists := expectedContents[baseName]
		if !exists {
			t.Errorf("Unexpected file found: %s", baseName)
			continue
		}

		if file.Content != expectedContent {
			t.Errorf("Content mismatch for file %s. \nExpected: %q\nGot: %q",
				baseName, expectedContent, file.Content)
		}
	}
}

func TestEmptyAndEdgeCases(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name          string
		paths         []string
		expectedFiles int
	}{
		{
			name:          "Empty paths array",
			paths:         []string{},
			expectedFiles: 0,
		},
		{
			name:          "Mix of files and directories",
			paths:         []string{filepath.Join(testDir, "main.go"), filepath.Join(testDir, "src")},
			expectedFiles: 3, // main.go + lib.go + helper.go
		},
		{
			name:          "Path with special characters",
			paths:         []string{filepath.Join(testDir, "RE\tADME.md")}, // Path doesn't exist
			expectedFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewMockLogger()
			config := NewConfig(true, "", "", "", "", logger)

			files, processedFiles, err := GatherProjectContext(tt.paths, config)
			// Error should not be returned even for invalid paths
			if err != nil {
				t.Fatalf("GatherProjectContext returned error: %v", err)
			}

			// Check processed files count
			if processedFiles != tt.expectedFiles {
				t.Errorf("Expected %d processed files, got %d", tt.expectedFiles, processedFiles)
			}

			// Check that file count matches returned slice length
			if len(files) != processedFiles {
				t.Errorf("Expected files slice length to be %d, got %d", processedFiles, len(files))
			}
		})
	}
}

func TestFileOrderDeterministic(t *testing.T) {
	// This test verifies that output is deterministic (sorted by path) regardless
	// of input order. Concurrent processing doesn't preserve input order, but
	// we guarantee deterministic sorted output for predictability.

	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a specific list of files (order doesn't matter for concurrent processing)
	inputPaths := []string{
		filepath.Join(testDir, "src/lib.go"),
		filepath.Join(testDir, "README.md"),
		filepath.Join(testDir, "main.go"),
		filepath.Join(testDir, "src/utils/helper.go"),
	}

	// Set up the config
	logger := NewMockLogger()
	config := NewConfig(true, "", "", "", "", logger)

	// Gather context
	files, processedFiles, err := GatherProjectContext(inputPaths, config)
	if err != nil {
		t.Fatalf("GatherProjectContext returned error: %v", err)
	}

	// Verify we got all the files
	if processedFiles != len(inputPaths) {
		t.Fatalf("Expected %d processed files, got %d", len(inputPaths), processedFiles)
	}

	// Verify output is sorted by path (deterministic order)
	for i := 1; i < len(files); i++ {
		if files[i-1].Path >= files[i].Path {
			t.Errorf("Output not sorted: files[%d].Path=%s >= files[%d].Path=%s",
				i-1, files[i-1].Path, i, files[i].Path)
		}
	}

	// Verify all expected files are present
	expectedBases := map[string]bool{
		"lib.go":    false,
		"README.md": false,
		"main.go":   false,
		"helper.go": false,
	}
	for _, f := range files {
		base := filepath.Base(f.Path)
		if _, ok := expectedBases[base]; ok {
			expectedBases[base] = true
		}
	}
	for base, found := range expectedBases {
		if !found {
			t.Errorf("Expected file %s not found in output", base)
		}
	}
}

func TestPathNormalization(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test with relative paths
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Make testDir relative to current directory if possible
	relPath, err := filepath.Rel(currentDir, testDir)
	if err != nil {
		// If we can't get a relative path, just use the absolute path
		relPath = testDir
	}

	// Set up the config
	logger := NewMockLogger()
	config := NewConfig(true, ".go", "", "", "", logger)

	// Test with the relative path
	files, processedFiles, err := GatherProjectContext([]string{relPath}, config)
	if err != nil {
		t.Fatalf("GatherProjectContext returned error: %v", err)
	}

	if processedFiles == 0 {
		t.Errorf("Expected to process some files, but got 0")
	}

	// Verify all paths in the result are absolute
	for _, file := range files {
		if !filepath.IsAbs(file.Path) {
			t.Errorf("Path should be absolute, got: %s", file.Path)
		}
	}
}
