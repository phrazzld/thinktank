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
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Return the test directory and cleanup function
	return testDir, func() {
		os.RemoveAll(testDir)
	}
}

func TestGatherProjectContext(t *testing.T) {
	// This test will run with the actual isGitIgnored function
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
		unexpectedPaths []string
		expectedPaths   []string
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
				"app.js",
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
			if processedFiles != tt.expectedFiles {
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
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Set up the config
	logger := NewMockLogger()
	config := NewConfig(true, ".go", "", "", "<{path}>\n{content}\n</{path}>", logger)

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

	// Check that all files in the result slice have .go extension
	if len(files) != processedFiles {
		t.Errorf("Expected files slice length to be %d, got %d", processedFiles, len(files))
	}

	for _, file := range files {
		if filepath.Ext(file.Path) != ".go" {
			t.Errorf("Result contains non-.go file: %s", file.Path)
		}
	}
}

// TestFormatting is not needed anymore as the format field is no longer used in GatherProjectContext
// The formatting will be handled by the prompt stitching logic
