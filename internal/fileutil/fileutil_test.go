// internal/fileutil/fileutil_test.go
package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Single word",
			input:    "hello",
			expected: 1,
		},
		{
			name:     "Two words",
			input:    "hello world",
			expected: 2,
		},
		{
			name:     "Multiple spaces",
			input:    "hello  world",
			expected: 2,
		},
		{
			name:     "With newlines",
			input:    "hello\nworld",
			expected: 2,
		},
		{
			name:     "With tabs",
			input:    "hello\tworld",
			expected: 2,
		},
		{
			name:     "Mixed whitespace",
			input:    "hello\n \t world",
			expected: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := estimateTokenCount(test.input)
			if result != test.expected {
				t.Errorf("estimateTokenCount(%q) = %d, expected %d", test.input, result, test.expected)
			}
		})
	}
}

func TestCalculateStatistics(t *testing.T) {
	input := "Hello\nWorld\nThis is a test."
	expectedChars := len(input)
	expectedLines := 3
	expectedTokens := 6 // "Hello", "World", "This", "is", "a", "test."

	chars, lines, tokens := CalculateStatistics(input)

	if chars != expectedChars {
		t.Errorf("Character count: got %d, want %d", chars, expectedChars)
	}

	if lines != expectedLines {
		t.Errorf("Line count: got %d, want %d", lines, expectedLines)
	}

	if tokens != expectedTokens {
		t.Errorf("Token count: got %d, want %d", tokens, expectedTokens)
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "Empty content",
			content:  []byte{},
			expected: false,
		},
		{
			name:     "Text content",
			content:  []byte("This is a text file with some content.\nIt has multiple lines."),
			expected: false,
		},
		{
			name:     "Content with null byte",
			content:  []byte("This has a null byte.\x00Right there."),
			expected: true,
		},
		{
			name:     "Content with many non-printable characters",
			content:  []byte{0x01, 0x02, 0x03, 'H', 'e', 'l', 'l', 'o', 0x04, 0x05},
			expected: true,
		},
		{
			name:     "Content with few non-printable characters",
			content:  []byte("Hello\nWorld\tThis is a test with a bell sound: \a"),
			expected: false,
		},
		{
			name:     "Very long text content",
			content:  []byte(strings.Repeat("This is a test. ", 100)),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryFile(tt.content)
			if result != tt.expected {
				t.Errorf("isBinaryFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name         string
		verbose      bool
		include      string
		exclude      string
		excludeNames string
		format       string
		checkFunc    func(*testing.T, *Config)
	}{
		{
			name:         "Default config",
			verbose:      false,
			include:      "",
			exclude:      "",
			excludeNames: "",
			format:       "test-format",
			checkFunc: func(t *testing.T, c *Config) {
				if c.Verbose != false {
					t.Errorf("Expected Verbose to be false, got %v", c.Verbose)
				}
				if len(c.IncludeExts) != 0 {
					t.Errorf("Expected IncludeExts to be empty, got %v", c.IncludeExts)
				}
				if len(c.ExcludeExts) != 0 {
					t.Errorf("Expected ExcludeExts to be empty, got %v", c.ExcludeExts)
				}
				if len(c.ExcludeNames) != 0 {
					t.Errorf("Expected ExcludeNames to be empty, got %v", c.ExcludeNames)
				}
				if c.Format != "test-format" {
					t.Errorf("Expected Format to be 'test-format', got %v", c.Format)
				}
			},
		},
		{
			name:         "With include extensions",
			verbose:      true,
			include:      "go,md,txt",
			exclude:      "",
			excludeNames: "",
			format:       "format",
			checkFunc: func(t *testing.T, c *Config) {
				if !c.Verbose {
					t.Errorf("Expected Verbose to be true, got %v", c.Verbose)
				}
				if len(c.IncludeExts) != 3 {
					t.Errorf("Expected 3 include extensions, got %d", len(c.IncludeExts))
				}
				expected := []string{".go", ".md", ".txt"}
				for i, ext := range expected {
					if c.IncludeExts[i] != ext {
						t.Errorf("Expected include extension %d to be %s, got %s", i, ext, c.IncludeExts[i])
					}
				}
			},
		},
		{
			name:         "With exclude extensions",
			verbose:      false,
			include:      "",
			exclude:      "exe,bin,obj",
			excludeNames: "",
			format:       "format",
			checkFunc: func(t *testing.T, c *Config) {
				if len(c.ExcludeExts) != 3 {
					t.Errorf("Expected 3 exclude extensions, got %d", len(c.ExcludeExts))
				}
				expected := []string{".exe", ".bin", ".obj"}
				for i, ext := range expected {
					if c.ExcludeExts[i] != ext {
						t.Errorf("Expected exclude extension %d to be %s, got %s", i, ext, c.ExcludeExts[i])
					}
				}
			},
		},
		{
			name:         "With exclude names",
			verbose:      false,
			include:      "",
			exclude:      "",
			excludeNames: "node_modules,dist,build",
			format:       "format",
			checkFunc: func(t *testing.T, c *Config) {
				if len(c.ExcludeNames) != 3 {
					t.Errorf("Expected 3 exclude names, got %d", len(c.ExcludeNames))
				}
				expected := []string{"node_modules", "dist", "build"}
				for i, name := range expected {
					if c.ExcludeNames[i] != name {
						t.Errorf("Expected exclude name %d to be %s, got %s", i, name, c.ExcludeNames[i])
					}
				}
			},
		},
		{
			name:         "With all options",
			verbose:      true,
			include:      "go,js",
			exclude:      "exe,bin",
			excludeNames: "node_modules,dist",
			format:       "custom-format",
			checkFunc: func(t *testing.T, c *Config) {
				if !c.Verbose {
					t.Errorf("Expected Verbose to be true, got %v", c.Verbose)
				}
				if len(c.IncludeExts) != 2 {
					t.Errorf("Expected 2 include extensions, got %d", len(c.IncludeExts))
				}
				if len(c.ExcludeExts) != 2 {
					t.Errorf("Expected 2 exclude extensions, got %d", len(c.ExcludeExts))
				}
				if len(c.ExcludeNames) != 2 {
					t.Errorf("Expected 2 exclude names, got %d", len(c.ExcludeNames))
				}
				if c.Format != "custom-format" {
					t.Errorf("Expected Format to be 'custom-format', got %v", c.Format)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewMockLogger()
			config := NewConfig(tt.verbose, tt.include, tt.exclude, tt.excludeNames, tt.format, logger)
			tt.checkFunc(t, config)
		})
	}
}

// TestIsGitIgnored tests the isGitIgnored function with basic cases
func TestIsGitIgnored(t *testing.T) {
	logger := NewMockLogger()
	config := &Config{Logger: logger}

	// Basic cases that don't require git
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     ".git directory",
			path:     ".git",
			expected: true,
		},
		{
			name:     "File in .git directory",
			path:     filepath.Join(".git", "config"),
			expected: true,
		},
		{
			name:     "Hidden file",
			path:     ".gitignore",
			expected: true,
		},
		{
			name:     "Regular file",
			path:     "README.md",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGitIgnored(tt.path, config)
			// For .git and hidden files, we expect true
			// For other cases without git, the behavior depends on git availability
			if (tt.path == ".git" ||
				strings.Contains(tt.path, string(filepath.Separator)+".git"+string(filepath.Separator)) ||
				(strings.HasPrefix(filepath.Base(tt.path), ".") &&
					filepath.Base(tt.path) != "." &&
					filepath.Base(tt.path) != "..")) && !result {
				t.Errorf("isGitIgnored(%q) = %v, want %v", tt.path, result, true)
			}
		})
	}
}

func TestShouldProcess(t *testing.T) {
	// This test uses the real isGitIgnored function,
	// so it might behave differently depending on git being installed

	tests := []struct {
		name         string
		path         string
		config       *Config
		expected     bool
		setupFunc    func(*Config)
		cleanupFunc  func()
		checkLogging func(*testing.T, *MockLogger)
	}{
		{
			name:     "Simple file, no filters",
			path:     "test.txt",
			expected: true,
			setupFunc: func(c *Config) {
				// No filters to set up
			},
		},
		{
			name:     "Excluded by extension",
			path:     "test.exe",
			expected: false,
			setupFunc: func(c *Config) {
				c.ExcludeExts = []string{".exe", ".bin"}
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping excluded extension") {
					t.Errorf("Expected log about excluded extension")
				}
			},
		},
		{
			name:     "Explicitly included extension",
			path:     "test.go",
			expected: true,
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go", ".md"}
			},
		},
		{
			name:     "Not in include list",
			path:     "test.js",
			expected: false,
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go", ".md"}
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping non-included extension") {
					t.Errorf("Expected log about non-included extension")
				}
			},
		},
		{
			name:     "Excluded by name",
			path:     "node_modules",
			expected: false,
			setupFunc: func(c *Config) {
				c.ExcludeNames = []string{"node_modules", "dist"}
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping excluded name") {
					t.Errorf("Expected log about excluded name")
				}
			},
		},
		{
			name:     "Hidden file",
			path:     ".gitignore",
			expected: false,
			setupFunc: func(c *Config) {
				// No special setup needed for hidden files
			},
		},
		{
			name:     "Complex filters - should process",
			path:     "src/main.go",
			expected: true,
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go", ".md"}
				c.ExcludeExts = []string{".exe", ".bin"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up a mock logger to capture logs
			logger := NewMockLogger()
			logger.SetVerbose(true)

			// Create a config
			config := &Config{
				Logger: logger,
			}

			// Apply any test-specific setup
			if tt.setupFunc != nil {
				tt.setupFunc(config)
			}

			// Clean up after the test
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc()
			}

			// Run the test
			result := shouldProcess(tt.path, config)

			// Check the result
			if result != tt.expected {
				t.Errorf("shouldProcess(%q) = %v, want %v", tt.path, result, tt.expected)
			}

			// Check logs if needed
			if tt.checkLogging != nil {
				tt.checkLogging(t, logger)
			}
		})
	}
}

// TestGatherProjectContextFiltering tests the filtering behavior of GatherProjectContext directly.
// This test verifies that the file filtering functionality works correctly without
// running the entire application flow through architect.Execute.
func TestGatherProjectContextFiltering(t *testing.T) {
	type filterTestCase struct {
		name                    string
		fileContents            map[string]string // Map of relative paths to file contents
		includeFilter           string
		excludeFilter           string
		excludeNames            string
		expectedIncludedFiles   []string
		expectedExcludedFiles   []string
		expectedProcessedCount  int
		expectedFilteredMessage string
	}

	tests := []filterTestCase{
		{
			name: "Include Go and Markdown Files",
			fileContents: map[string]string{
				"main.go":       "package main\n\nfunc main() {}\n",
				"README.md":     "# Test Project",
				"config.json":   `{"key": "value"}`,
				"utils/util.js": "function helper() { return true; }",
			},
			includeFilter: ".go,.md",
			excludeFilter: "",
			excludeNames:  "",
			expectedIncludedFiles: []string{
				"main.go",
				"README.md",
			},
			expectedExcludedFiles: []string{
				"config.json",
				"utils/util.js",
			},
			expectedProcessedCount: 2, // Should only process the .go and .md files
		},
		{
			name: "Exclude JSON Files",
			fileContents: map[string]string{
				"main.go":       "package main\n\nfunc main() {}\n",
				"README.md":     "# Test Project",
				"config.json":   `{"key": "value"}`,
				"data.json":     `{"data": [1,2,3]}`,
				"utils/util.js": "function helper() { return true; }",
			},
			includeFilter: "",
			excludeFilter: ".json",
			excludeNames:  "",
			expectedIncludedFiles: []string{
				"main.go",
				"README.md",
				"utils/util.js",
			},
			expectedExcludedFiles: []string{
				"config.json",
				"data.json",
			},
			expectedProcessedCount: 3, // Should process everything except the .json files
		},
		{
			name: "Exclude Directory Names",
			fileContents: map[string]string{
				"main.go":               "package main\n\nfunc main() {}\n",
				"README.md":             "# Test Project",
				"node_modules/index.js": "module.exports = {};",
				"dist/bundle.js":        "console.log('bundled');",
				"utils/util.js":         "function helper() { return true; }",
			},
			includeFilter: "",
			excludeFilter: "",
			excludeNames:  "node_modules,dist",
			expectedIncludedFiles: []string{
				"main.go",
				"README.md",
				"utils/util.js",
			},
			expectedExcludedFiles: []string{
				"node_modules/index.js",
				"dist/bundle.js",
			},
			expectedProcessedCount: 3, // Should skip the excluded directory names
		},
		{
			name: "Complex Filter Combination",
			fileContents: map[string]string{
				"src/main.go":              "package main\n\nfunc main() {}\n",
				"src/util.go":              "package main\n\nfunc util() {}\n",
				"docs/README.md":           "# Test Project",
				"docs/USAGE.md":            "## Usage Instructions",
				"config.json":              `{"key": "value"}`,
				"node_modules/index.js":    "module.exports = {};",
				"src/vendor/lib.js":        "function lib() { return true; }",
				"build/output.exe":         "binary content",
				"src/tests/test_main.go":   "package test\n\nfunc TestMain() {}\n",
				"src/tests/test_helper.go": "package test\n\nfunc TestHelper() {}\n",
			},
			includeFilter: ".go,.md",
			excludeFilter: ".exe",
			excludeNames:  "node_modules,vendor",
			expectedIncludedFiles: []string{
				"src/main.go",
				"src/util.go",
				"docs/README.md",
				"docs/USAGE.md",
				"src/tests/test_main.go",
				"src/tests/test_helper.go",
			},
			expectedExcludedFiles: []string{
				"config.json",
				"node_modules/index.js",
				"src/vendor/lib.js",
				"build/output.exe",
			},
			expectedProcessedCount: 6, // Should match only the included extensions while respecting the exclusions
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for each test
			tempDir := t.TempDir()

			// Set up a logger to capture logs
			logger := NewMockLogger()
			logger.SetVerbose(true)

			// Create files based on the test case
			for relativePath, content := range tc.fileContents {
				fullPath := filepath.Join(tempDir, relativePath)

				// Ensure the parent directory exists
				err := os.MkdirAll(filepath.Dir(fullPath), 0755)
				if err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}

				// Create the file
				err = os.WriteFile(fullPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			}

			// Create config for the test
			config := NewConfig(true, tc.includeFilter, tc.excludeFilter, tc.excludeNames, "format", logger)

			// Call GatherProjectContext directly
			files, processedCount, err := GatherProjectContext([]string{tempDir}, config)
			if err != nil {
				t.Fatalf("GatherProjectContext returned an error: %v", err)
			}

			// Verify processed count
			if processedCount != tc.expectedProcessedCount {
				t.Errorf("Processed count: got %d, want %d", processedCount, tc.expectedProcessedCount)
			}

			// Get the list of actual file paths (relative to the temp dir for easier comparison)
			var includedFiles []string
			for _, file := range files {
				// Convert absolute path to a path relative to the temp dir
				relativePath, err := filepath.Rel(tempDir, file.Path)
				if err != nil {
					t.Fatalf("Failed to get relative path: %v", err)
				}
				includedFiles = append(includedFiles, relativePath)
			}

			// Check that all expected files are included
			for _, expected := range tc.expectedIncludedFiles {
				found := false
				for _, actual := range includedFiles {
					if filepath.ToSlash(actual) == filepath.ToSlash(expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s to be included, but it wasn't", expected)
				}
			}

			// Check that expected excluded files are not included
			for _, excluded := range tc.expectedExcludedFiles {
				for _, actual := range includedFiles {
					if filepath.ToSlash(actual) == filepath.ToSlash(excluded) {
						t.Errorf("Expected file %s to be excluded, but it was included", excluded)
					}
				}
			}

			// Check for expected filter messages in the logs
			if tc.expectedFilteredMessage != "" && !logger.ContainsMessage(tc.expectedFilteredMessage) {
				t.Errorf("Expected log message about %s but didn't find it", tc.expectedFilteredMessage)
			}
		})
	}
}
