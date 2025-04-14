// internal/fileutil/config_test.go
package fileutil

import (
	"testing"
)

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
