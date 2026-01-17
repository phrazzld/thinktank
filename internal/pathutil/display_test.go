package pathutil

import (
	"path/filepath"
	"testing"
)

func TestSanitizePathWithCwd(t *testing.T) {
	// Use a consistent fake cwd for testing
	fakeCwd := "/home/user/project"

	tests := []struct {
		name     string
		absPath  string
		cwd      string
		expected string
	}{
		{
			name:     "empty path",
			absPath:  "",
			cwd:      fakeCwd,
			expected: "",
		},
		{
			name:     "path equals cwd",
			absPath:  "/home/user/project",
			cwd:      fakeCwd,
			expected: ".",
		},
		{
			name:     "path within cwd",
			absPath:  "/home/user/project/output",
			cwd:      fakeCwd,
			expected: "./output",
		},
		{
			name:     "nested path within cwd",
			absPath:  "/home/user/project/output/subdir/file.txt",
			cwd:      fakeCwd,
			expected: "./output/subdir/file.txt",
		},
		{
			name:     "path outside cwd - sibling",
			absPath:  "/home/user/other-project/output",
			cwd:      fakeCwd,
			expected: "./output",
		},
		{
			name:     "path outside cwd - parent",
			absPath:  "/home/user",
			cwd:      fakeCwd,
			expected: "./user",
		},
		{
			name:     "path outside cwd - different tree",
			absPath:  "/var/log/app.log",
			cwd:      fakeCwd,
			expected: "./app.log",
		},
		{
			name:     "path with trailing slash",
			absPath:  "/home/user/project/output/",
			cwd:      fakeCwd,
			expected: "./output",
		},
		{
			name:     "path with redundant components",
			absPath:  "/home/user/project/./output/../output",
			cwd:      fakeCwd,
			expected: "./output",
		},
		{
			name:     "cwd with trailing slash",
			absPath:  "/home/user/project/output",
			cwd:      "/home/user/project/",
			expected: "./output",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizePathWithCwd(tc.absPath, tc.cwd)
			if result != tc.expected {
				t.Errorf("sanitizePathWithCwd(%q, %q) = %q; want %q",
					tc.absPath, tc.cwd, result, tc.expected)
			}
		})
	}
}

func TestSanitizePathWithCwdWindows(t *testing.T) {
	// Test Windows-style paths (will work on all platforms via filepath.Clean)
	if filepath.Separator != '\\' {
		t.Skip("Skipping Windows path tests on non-Windows platform")
	}

	tests := []struct {
		name     string
		absPath  string
		cwd      string
		expected string
	}{
		{
			name:     "windows path within cwd",
			absPath:  `C:\Users\dev\project\output`,
			cwd:      `C:\Users\dev\project`,
			expected: `.\output`,
		},
		{
			name:     "windows path outside cwd",
			absPath:  `D:\other\path`,
			cwd:      `C:\Users\dev\project`,
			expected: `.\path`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizePathWithCwd(tc.absPath, tc.cwd)
			if result != tc.expected {
				t.Errorf("sanitizePathWithCwd(%q, %q) = %q; want %q",
					tc.absPath, tc.cwd, result, tc.expected)
			}
		})
	}
}

func TestSanitizePathForDisplay_Integration(t *testing.T) {
	// This test verifies the function works with actual os.Getwd()
	// It's more of an integration test to ensure the function doesn't panic

	tests := []struct {
		name    string
		absPath string
	}{
		{"empty", ""},
		{"relative already", "output"},
		{"absolute temp", "/tmp/test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			result := SanitizePathForDisplay(tc.absPath)

			// Empty input should return empty output
			if tc.absPath == "" && result != "" {
				t.Errorf("SanitizePathForDisplay(%q) = %q; want empty", tc.absPath, result)
			}
		})
	}
}
