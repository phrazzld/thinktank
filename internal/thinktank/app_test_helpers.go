// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/misty-step/thinktank/internal/testutil"
)

// ----- Test Helper Functions -----

// setupTestEnvironment creates a temporary directory for testing
func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create filesystem abstraction
	fs := testutil.NewRealFS()

	testDir, err := os.MkdirTemp("", "thinktank-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	cleanup := func() {
		err := fs.RemoveAll(testDir)
		if err != nil {
			t.Logf("Warning: Failed to clean up test directory: %v", err)
		}
	}

	return testDir, cleanup
}

// createTestFile creates a test file with the given content
func createTestFile(t *testing.T, path, content string) string {
	// Create filesystem abstraction
	fs := testutil.NewRealFS()

	err := fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("Failed to create directory for test file: %v", err)
	}

	err = fs.WriteFile(path, []byte(content), 0640)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return path
}
