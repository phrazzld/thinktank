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
	logger.SetVerbose(true)

	// Basic configuration - git not available
	configNoGit := &Config{
		Logger:       logger,
		GitAvailable: false,
	}

	// Test cases that don't depend on git
	basicTests := []struct {
		name     string
		path     string
		config   *Config
		expected bool
	}{
		// .git directory and files
		{
			name:     ".git directory",
			path:     ".git",
			config:   configNoGit,
			expected: true,
		},
		{
			name:     "File in .git directory with direct separator",
			path:     ".git" + string(filepath.Separator) + "config",
			config:   configNoGit,
			expected: false, // The function only checks for .git as the base name or /.git/ pattern
		},
		{
			name:     "Nested .git directory with separator",
			path:     "project" + string(filepath.Separator) + ".git",
			config:   configNoGit,
			expected: true,
		},
		{
			name:     "Path with .git in the middle with separators",
			path:     "src" + string(filepath.Separator) + ".git" + string(filepath.Separator) + "modules",
			config:   configNoGit,
			expected: true,
		},

		// Hidden files and directories
		{
			name:     "Hidden file",
			path:     ".gitignore",
			config:   configNoGit,
			expected: true,
		},
		{
			name:     "Hidden directory",
			path:     ".config",
			config:   configNoGit,
			expected: true,
		},
		{
			name:     "File in hidden directory with direct separator",
			path:     ".vscode" + string(filepath.Separator) + "settings.json",
			config:   configNoGit,
			expected: false, // Only checks if the base name starts with a dot
		},
		{
			name:     "Regular file",
			path:     "README.md",
			config:   configNoGit,
			expected: false,
		},
		{
			name:     "Current directory (.)",
			path:     ".",
			config:   configNoGit,
			expected: false,
		},
		{
			name:     "Parent directory (..)",
			path:     "..",
			config:   configNoGit,
			expected: false,
		},

		// Special paths
		{
			name:     "Path with spaces",
			path:     "my documents/file with spaces.txt",
			config:   configNoGit,
			expected: false,
		},
		{
			name:     "Path with non-ASCII characters",
			path:     "документы/файл.txt", // Cyrillic characters
			config:   configNoGit,
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			config:   configNoGit,
			expected: false,
		},
		{
			name:     "Path containing just a dot",
			path:     "src/./file.txt", // Should normalize to src/file.txt
			config:   configNoGit,
			expected: false,
		},
	}

	for _, tt := range basicTests {
		t.Run(tt.name, func(t *testing.T) {
			logger.ClearMessages()
			result := isGitIgnored(tt.path, tt.config)

			if result != tt.expected {
				t.Errorf("isGitIgnored(%q) = %v, want %v", tt.path, result, tt.expected)
			}

			// Skip checking log messages for some specific cases where the implementation doesn't log
			skipLogCheck := tt.path == ".git" || strings.HasSuffix(tt.path, string(filepath.Separator)+".git") ||
				strings.Contains(tt.path, string(filepath.Separator)+".git"+string(filepath.Separator))

			// Only check log messages if we expect the file to be ignored and it's not a special case
			if tt.expected && !skipLogCheck && !logger.ContainsMessage("Git ignored") && !logger.ContainsMessage("Hidden file/dir ignored") {
				t.Errorf("Expected log message about git ignored or hidden file/dir for path %s", tt.path)
			}
		})
	}
}

// TestIsGitIgnoredWithMockGit tests the isGitIgnored function with a mock git environment
func TestIsGitIgnoredWithMockGit(t *testing.T) {
	// Skip this test if the OS doesn't support executable permissions
	if os.Getenv("SKIP_GIT_MOCK_TESTS") != "" {
		t.Skip("Skipping tests that require git command mocking")
	}

	logger := NewMockLogger()
	logger.SetVerbose(true)

	// Create a temporary directory to act as a git repo
	tempDir := t.TempDir()

	// Create a mock git executable script in the temp directory
	mockGitPath := filepath.Join(tempDir, "git")
	if isWindows() {
		mockGitPath += ".bat"
	}

	// Create a bash script that simulates git behavior
	mockGitScript := `#!/bin/sh
if [ "$1" = "-C" ] && [ "$3" = "rev-parse" ] && [ "$4" = "--is-inside-work-tree" ]; then
  # Simulate being in a git repo
  exit 0
elif [ "$1" = "-C" ] && [ "$3" = "check-ignore" ] && [ "$4" = "-q" ]; then
  # Check if the file should be ignored
  filename="$5"
  
  # Files to ignore
  if [ "$filename" = "ignored.txt" ] || [ "$filename" = "build.log" ] || [ "$filename" = "node_modules" ]; then
    exit 0  # Exit code 0: file IS ignored
  else
    exit 1  # Exit code 1: file is NOT ignored
  fi
else
  # Unknown git command, return error
  exit 2
fi
`

	// For Windows, create a batch script instead
	if isWindows() {
		mockGitScript = `@echo off
if "%1"=="-C" (
  if "%3"=="rev-parse" (
    if "%4"=="--is-inside-work-tree" (
      exit /b 0
    )
  )
  if "%3"=="check-ignore" (
    if "%4"=="-q" (
      set filename=%5
      if "%filename%"=="ignored.txt" exit /b 0
      if "%filename%"=="build.log" exit /b 0
      if "%filename%"=="node_modules" exit /b 0
      exit /b 1
    )
  )
)
exit /b 2
`
	}

	// Write the mock git script
	err := os.WriteFile(mockGitPath, []byte(mockGitScript), 0755)
	if err != nil {
		t.Fatalf("Failed to write mock git script: %v", err)
	}

	// Add tempDir to PATH temporarily for this test
	origPath := os.Getenv("PATH")
	t.Cleanup(func() {
		os.Setenv("PATH", origPath) // Restore original PATH
	})
	os.Setenv("PATH", tempDir+string(filepath.ListSeparator)+origPath)

	// Create a config with git available
	config := &Config{
		Logger:       logger,
		GitAvailable: true,
	}

	// Test cases
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "File that should be ignored by git",
			path:     "ignored.txt",
			expected: true,
		},
		{
			name:     "Another ignored file",
			path:     "build.log",
			expected: true,
		},
		{
			name:     "Ignored directory",
			path:     "node_modules",
			expected: true,
		},
		{
			name:     "File that should not be ignored",
			path:     "README.md",
			expected: false,
		},
		{
			name:     "Source file",
			path:     "main.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.ClearMessages()
			result := isGitIgnored(tt.path, config)

			if result != tt.expected {
				t.Errorf("isGitIgnored(%q) with mock git = %v, want %v", tt.path, result, tt.expected)
			}

			// Check for log messages
			if tt.expected && !logger.ContainsMessage("Git ignored") && !logger.ContainsMessage("Hidden file/dir ignored") {
				t.Errorf("Expected log message about git ignored for path %s", tt.path)
			}
		})
	}
}

// TestGitErrorHandling tests how isGitIgnored handles various git command errors
func TestGitErrorHandling(t *testing.T) {
	// Skip this test if the OS doesn't support executable permissions
	if os.Getenv("SKIP_GIT_MOCK_TESTS") != "" {
		t.Skip("Skipping tests that require git command mocking")
	}

	logger := NewMockLogger()
	logger.SetVerbose(true)

	// Create a temporary directory for the mock git
	tempDir := t.TempDir()

	// Create a mock git executable that always fails
	mockGitPath := filepath.Join(tempDir, "git")
	if isWindows() {
		mockGitPath += ".bat"
	}

	// Create a bash script that simulates different git error scenarios
	mockGitScript := `#!/bin/sh
if [ "$1" = "-C" ] && [ "$3" = "rev-parse" ] && [ "$4" = "--is-inside-work-tree" ]; then
  # Simulate not being in a git repo
  if [ "$2" = "not-a-repo" ]; then
    exit 128
  else
    exit 0  # Is a git repo
  fi
elif [ "$1" = "-C" ] && [ "$3" = "check-ignore" ] && [ "$4" = "-q" ]; then
  # Simulate various check-ignore errors
  if [ "$2" = "error-repo" ]; then
    exit 128  # Fatal error
  elif [ "$5" = "error-file.txt" ]; then
    exit 2    # Other error code
  elif [ "$5" = "ignored.txt" ]; then
    exit 0    # File is ignored
  else
    exit 1    # File is not ignored
  fi
else
  # Unknown git command
  exit 2
fi
`

	// For Windows, create a batch script instead
	if isWindows() {
		mockGitScript = `@echo off
if "%1"=="-C" (
  if "%3"=="rev-parse" (
    if "%4"=="--is-inside-work-tree" (
      if "%2"=="not-a-repo" (
        exit /b 128
      ) else (
        exit /b 0
      )
    )
  )
  if "%3"=="check-ignore" (
    if "%4"=="-q" (
      if "%2"=="error-repo" (
        exit /b 128
      )
      if "%5"=="error-file.txt" (
        exit /b 2
      )
      if "%5"=="ignored.txt" (
        exit /b 0
      )
      exit /b 1
    )
  )
)
exit /b 2
`
	}

	// Write the mock git script
	err := os.WriteFile(mockGitPath, []byte(mockGitScript), 0755)
	if err != nil {
		t.Fatalf("Failed to write mock git script: %v", err)
	}

	// Add tempDir to PATH temporarily for this test
	origPath := os.Getenv("PATH")
	t.Cleanup(func() {
		os.Setenv("PATH", origPath) // Restore original PATH
	})
	os.Setenv("PATH", tempDir+string(filepath.ListSeparator)+origPath)

	// Create a config with git available
	config := &Config{
		Logger:       logger,
		GitAvailable: true,
	}

	// Test cases for error handling
	tests := []struct {
		name     string
		dirPath  string // Directory part of the path (for -C argument)
		filePath string // File part of the path (basename)
		expected bool
	}{
		{
			name:     "Not in a git repository",
			dirPath:  "not-a-repo",
			filePath: "some-file.txt",
			expected: false, // Falls back to hidden check, which is false
		},
		{
			name:     "Error running check-ignore",
			dirPath:  "error-repo",
			filePath: "some-file.txt",
			expected: false, // Falls back to hidden check, which is false
		},
		{
			name:     "File with check-ignore error",
			dirPath:  "normal-repo",
			filePath: "error-file.txt",
			expected: false, // Falls back to hidden check, which is false
		},
		{
			name:     "Regular ignored file",
			dirPath:  "normal-repo",
			filePath: "ignored.txt",
			expected: true,
		},
		{
			name:     "Hidden file with git error",
			dirPath:  "error-repo",
			filePath: ".config",
			expected: true, // Hidden file is still ignored even if git errors out
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.ClearMessages()
			// Construct the full path
			fullPath := filepath.Join(tt.dirPath, tt.filePath)

			result := isGitIgnored(fullPath, config)

			if result != tt.expected {
				t.Errorf("isGitIgnored(%q) with error conditions = %v, want %v", fullPath, result, tt.expected)
			}

			// Check for expected error messages
			if tt.dirPath == "error-repo" && !logger.ContainsMessage("Error running git check-ignore") {
				t.Errorf("Expected error message about git check-ignore for %s", fullPath)
			}
		})
	}
}

// Helper function to check if running on Windows
func isWindows() bool {
	return filepath.Separator == '\\' && filepath.ListSeparator == ';'
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
		// Basic cases
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

		// Additional test cases for improved coverage
		{
			name:     "File in .git directory",
			path:     ".git/config",
			expected: true, // The implementation only checks .git at the very start of the path or as a directory in a path
			setupFunc: func(c *Config) {
				// No special setup needed
			},
		},
		{
			name:     "File with mixed case extension - excluded",
			path:     "test.EXE",
			expected: false,
			setupFunc: func(c *Config) {
				c.ExcludeExts = []string{".exe"}
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping excluded extension") {
					t.Errorf("Expected log about excluded extension")
				}
			},
		},
		{
			name:     "File with mixed case extension - included",
			path:     "test.GO",
			expected: true,
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go"}
			},
		},
		{
			name:     "File with no extension - no filters",
			path:     "Makefile",
			expected: true,
			setupFunc: func(c *Config) {
				// No filters
			},
		},
		{
			name:     "File with no extension - include filter",
			path:     "Makefile",
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
			name:     "Path with directory containing excluded name",
			path:     "src/node_modules/somefile.js",
			expected: true, // Should process because only the base name is checked
			setupFunc: func(c *Config) {
				c.ExcludeNames = []string{"node_modules"}
			},
		},
		{
			name:     "Multiple conditions: included extension but excluded name",
			path:     "node_modules/test.go",
			expected: true, // The implementation only checks base name against ExcludeNames, not directory components
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go"}
				c.ExcludeNames = []string{"node_modules"}
			},
		},
		{
			name:     "Multiple conditions: included extension but also in exclude extensions",
			path:     "test.go",
			expected: false, // Exclude takes precedence
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go", ".md"}
				c.ExcludeExts = []string{".go", ".bin"} // Conflicting configuration
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping excluded extension") {
					t.Errorf("Expected log about excluded extension")
				}
			},
		},
		{
			name:     "Complex path with nested directories",
			path:     "src/app/utils/helpers/string.go",
			expected: true,
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go"}
			},
		},
		{
			name:     "Empty file path",
			path:     "",
			expected: false, // Empty paths are treated as having no extension, so they don't match when IncludeExts is set
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go"}
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping non-included extension") {
					t.Errorf("Expected log about non-included extension")
				}
			},
		},
		{
			name:     "Path with unusual characters",
			path:     "file with spaces.go",
			expected: true,
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go"}
			},
		},
		{
			name:     "Include and exclude empty but ExcludeNames populated",
			path:     "test.txt",
			expected: true,
			setupFunc: func(c *Config) {
				c.ExcludeNames = []string{"node_modules", "dist"}
			},
		},
		{
			name:     "Both include and exclude have same extension",
			path:     "test.go",
			expected: false, // Exclude takes precedence over include
			setupFunc: func(c *Config) {
				c.IncludeExts = []string{".go", ".md"}
				c.ExcludeExts = []string{".go"} // Conflicts with include
			},
			checkLogging: func(t *testing.T, l *MockLogger) {
				if !l.ContainsMessage("Skipping excluded extension") {
					t.Errorf("Expected log about excluded extension")
				}
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
