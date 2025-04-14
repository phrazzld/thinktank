// internal/fileutil/filtering_test.go
package fileutil

import (
	"testing"
)

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
