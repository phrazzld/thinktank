// internal/fileutil/fileutil_test.go
package fileutil

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
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

func TestCalculateStatisticsWithTokenCounting(t *testing.T) {
	input := "Hello world, this is a test of the token counting system."
	ctx := context.Background()

	// Setup a mock client that will return a predefined token count
	mockClient := &gemini.MockClient{
		CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
			return &gemini.TokenCount{Total: 15}, nil
		},
	}

	// Create a mock logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ")

	// Test with mock client
	chars, lines, tokens := CalculateStatisticsWithTokenCounting(ctx, mockClient, input, logger)

	// Verify character count
	expectedChars := len(input)
	if chars != expectedChars {
		t.Errorf("Character count: got %d, want %d", chars, expectedChars)
	}

	// Verify line count
	expectedLines := 1
	if lines != expectedLines {
		t.Errorf("Line count: got %d, want %d", lines, expectedLines)
	}

	// Verify token count from mock client
	expectedTokens := 15 // From our mock
	if tokens != expectedTokens {
		t.Errorf("Token count: got %d, want %d", tokens, expectedTokens)
	}

	// Now test fallback behavior when client is nil
	_, _, tokens = CalculateStatisticsWithTokenCounting(ctx, nil, input, logger)

	// Should use estimation
	expectedTokensFallback := estimateTokenCount(input)
	if tokens != expectedTokensFallback {
		t.Errorf("Fallback token count: got %d, want %d", tokens, expectedTokensFallback)
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
