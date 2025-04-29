package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileSystemInterface verifies that both MemFS and RealFS implement the same interface
func TestFileSystemInterface(t *testing.T) {
	// Create both filesystem implementations
	memFS := NewMemFS()
	realFS := NewRealFS()

	// Check common methods
	if err := memFS.MkdirAll("/test/path", 0755); err != nil {
		t.Errorf("memFS.MkdirAll failed: %v", err)
	}
	_ = realFS.Join("test", "path")
	_ = memFS.Base("/test/path/file.txt")
	_ = memFS.Dir("/test/path/file.txt")

	// Create a small file in both filesystems
	testContent := []byte("Test content")

	// Create a temp directory for real filesystem tests
	tempDir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Test in-memory filesystem operations
	if err := memFS.MkdirAll("tmp/memfs", 0755); err != nil {
		t.Errorf("MemFS.MkdirAll failed: %v", err)
	}

	if err := memFS.WriteFile("tmp/memfs/test.txt", testContent, 0640); err != nil {
		t.Errorf("MemFS.WriteFile failed: %v", err)
	}

	content, err := memFS.ReadFile("tmp/memfs/test.txt")
	if err != nil {
		t.Errorf("MemFS.ReadFile failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("MemFS content mismatch: got %q, want %q", string(content), string(testContent))
	}

	// Test real filesystem operations
	realTestPath := filepath.Join(tempDir, "test.txt")
	if err := realFS.WriteFile(realTestPath, testContent, 0640); err != nil {
		t.Errorf("RealFS.WriteFile failed: %v", err)
	}

	content, err = realFS.ReadFile(realTestPath)
	if err != nil {
		t.Errorf("RealFS.ReadFile failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("RealFS content mismatch: got %q, want %q", string(content), string(testContent))
	}
}
