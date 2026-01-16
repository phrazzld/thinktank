// internal/fileutil/git_operations_test.go
package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	// Note: git check-ignore now uses "--" separator, so filename is $6
	mockGitScript := `#!/bin/sh
if [ "$1" = "-C" ] && [ "$3" = "rev-parse" ] && [ "$4" = "--is-inside-work-tree" ]; then
  # Simulate being in a git repo
  exit 0
elif [ "$1" = "-C" ] && [ "$3" = "check-ignore" ] && [ "$4" = "-q" ] && [ "$5" = "--" ]; then
  # Check if the file should be ignored (filename is after "--")
  filename="$6"

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
	// Note: git check-ignore now uses "--" separator, so filename is %6
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
      if "%5"=="--" (
        set filename=%6
        if "%filename%"=="ignored.txt" exit /b 0
        if "%filename%"=="build.log" exit /b 0
        if "%filename%"=="node_modules" exit /b 0
        exit /b 1
      )
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
		_ = os.Setenv("PATH", origPath) // Restore original PATH
	})
	_ = os.Setenv("PATH", tempDir+string(filepath.ListSeparator)+origPath)

	// Create a config with git available and GitChecker
	config := &Config{
		Logger:       logger,
		GitAvailable: true,
		GitChecker:   NewGitChecker(),
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
	// Note: git check-ignore now uses "--" separator, so filename is $6
	mockGitScript := `#!/bin/sh
if [ "$1" = "-C" ] && [ "$3" = "rev-parse" ] && [ "$4" = "--is-inside-work-tree" ]; then
  # Simulate not being in a git repo
  if [ "$2" = "not-a-repo" ]; then
    exit 128
  else
    exit 0  # Is a git repo
  fi
elif [ "$1" = "-C" ] && [ "$3" = "check-ignore" ] && [ "$4" = "-q" ] && [ "$5" = "--" ]; then
  # Simulate various check-ignore errors (filename is after "--")
  if [ "$2" = "error-repo" ]; then
    exit 128  # Fatal error
  elif [ "$6" = "error-file.txt" ]; then
    exit 2    # Other error code
  elif [ "$6" = "ignored.txt" ]; then
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
	// Note: git check-ignore now uses "--" separator, so filename is %6
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
      if "%5"=="--" (
        if "%2"=="error-repo" (
          exit /b 128
        )
        if "%6"=="error-file.txt" (
          exit /b 2
        )
        if "%6"=="ignored.txt" (
          exit /b 0
        )
        exit /b 1
      )
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
		_ = os.Setenv("PATH", origPath) // Restore original PATH
	})
	_ = os.Setenv("PATH", tempDir+string(filepath.ListSeparator)+origPath)

	// Create a config with git available and GitChecker
	config := &Config{
		Logger:       logger,
		GitAvailable: true,
		GitChecker:   NewGitChecker(),
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
