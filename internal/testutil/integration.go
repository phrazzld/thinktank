package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTempDir creates a temporary directory for testing and ensures it's cleaned up
// when the test completes. The directory name will have the given prefix.
func SetupTempDir(t testing.TB, prefix string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Failed to clean up temporary directory %s: %v", dir, err)
		}
	})

	return dir
}

// SetupTempFile creates a temporary file for testing and ensures it's cleaned up
// when the test completes. The file name will have the given prefix and suffix.
func SetupTempFile(t testing.TB, prefix, suffix string) (string, *os.File) {
	t.Helper()

	file, err := os.CreateTemp("", prefix+"*"+suffix)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	t.Cleanup(func() {
		// Close the file if it's still open
		if err := file.Close(); err != nil {
			t.Logf("Failed to close temporary file %s: %v", file.Name(), err)
		}
		// Remove the file
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("Failed to remove temporary file %s: %v", file.Name(), err)
		}
	})

	return file.Name(), file
}

// SetupTempFileInDir creates a temporary file in the specified directory for testing
// and ensures it's cleaned up when the test completes.
func SetupTempFileInDir(t testing.TB, dir, prefix, suffix string) (string, *os.File) {
	t.Helper()

	file, err := os.CreateTemp(dir, prefix+"*"+suffix)
	if err != nil {
		t.Fatalf("Failed to create temporary file in directory %s: %v", dir, err)
	}

	t.Cleanup(func() {
		// Close the file if it's still open
		if err := file.Close(); err != nil {
			t.Logf("Failed to close temporary file %s: %v", file.Name(), err)
		}
		// Remove the file
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("Failed to remove temporary file %s: %v", file.Name(), err)
		}
	})

	return file.Name(), file
}

// CreateTestFile creates a file with the given content in the specified directory.
// The file is automatically cleaned up when the test completes.
func CreateTestFile(t testing.TB, dir, filename string, content []byte) string {
	t.Helper()

	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", fullPath, err)
	}

	t.Cleanup(func() {
		if err := os.Remove(fullPath); err != nil {
			t.Logf("Failed to remove test file %s: %v", fullPath, err)
		}
	})

	return fullPath
}

// CreateTestDir creates a directory at the specified path and ensures it's cleaned up
// when the test completes.
func CreateTestDir(t testing.TB, dirPath string) string {
	t.Helper()

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory %s: %v", dirPath, err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dirPath); err != nil {
			t.Logf("Failed to remove test directory %s: %v", dirPath, err)
		}
	})

	return dirPath
}

// CreateTestFiles creates multiple files with given content in the specified directory.
// All files are automatically cleaned up when the test completes.
func CreateTestFiles(t testing.TB, dir string, files map[string][]byte) []string {
	t.Helper()

	var createdFiles []string

	for filename, content := range files {
		fullPath := CreateTestFile(t, dir, filename, content)
		createdFiles = append(createdFiles, fullPath)
	}

	return createdFiles
}

// WithTempDir executes a function with a temporary directory, ensuring cleanup.
// This is useful for table-driven tests or when you need to scope the directory usage.
func WithTempDir(t testing.TB, prefix string, fn func(dir string)) {
	t.Helper()

	dir := SetupTempDir(t, prefix)
	fn(dir)
}

// CreateNestedTestDirs creates a nested directory structure for testing.
// The structure is defined as a slice of relative paths.
func CreateNestedTestDirs(t testing.TB, baseDir string, paths []string) []string {
	t.Helper()

	var createdDirs []string

	for _, path := range paths {
		fullPath := filepath.Join(baseDir, path)
		createdDir := CreateTestDir(t, fullPath)
		createdDirs = append(createdDirs, createdDir)
	}

	return createdDirs
}
