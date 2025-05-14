package testutil

import (
	"context"
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

// TestContextFilesystemOperations tests the context-aware operations for both filesystem implementations
func TestContextFilesystemOperations(t *testing.T) {
	// Create both filesystem implementations
	memFS := NewMemFS()
	realFS := NewRealFS()
	ctx := context.Background()

	// Create a temp directory for real filesystem tests
	tempDir, err := os.MkdirTemp("", "fs-context-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Test content for both filesystems
	testContent := []byte("Context-aware test content")

	// Test context-aware operations on MemFS
	if err := memFS.MkdirAllWithContext(ctx, "tmp/memfs-ctx", 0755); err != nil {
		t.Errorf("MemFS.MkdirAllWithContext failed: %v", err)
	}

	if err := memFS.WriteFileWithContext(ctx, "tmp/memfs-ctx/test.txt", testContent, 0640); err != nil {
		t.Errorf("MemFS.WriteFileWithContext failed: %v", err)
	}

	content, err := memFS.ReadFileWithContext(ctx, "tmp/memfs-ctx/test.txt")
	if err != nil {
		t.Errorf("MemFS.ReadFileWithContext failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("MemFS context content mismatch: got %q, want %q", string(content), string(testContent))
	}

	// Test context-aware operations on RealFS
	realCtxDir := filepath.Join(tempDir, "realfs-ctx")
	if err := realFS.MkdirAllWithContext(ctx, realCtxDir, 0755); err != nil {
		t.Errorf("RealFS.MkdirAllWithContext failed: %v", err)
	}

	realCtxTestPath := filepath.Join(realCtxDir, "test.txt")
	if err := realFS.WriteFileWithContext(ctx, realCtxTestPath, testContent, 0640); err != nil {
		t.Errorf("RealFS.WriteFileWithContext failed: %v", err)
	}

	content, err = realFS.ReadFileWithContext(ctx, realCtxTestPath)
	if err != nil {
		t.Errorf("RealFS.ReadFileWithContext failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("RealFS context content mismatch: got %q, want %q", string(content), string(testContent))
	}

	// Test context cancellation
	// Note: we're not actually using the canceled context yet
	_, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context immediately

	// The following operations should respect the canceled context
	// but since they're typically fast (especially for MemFS), we're mostly
	// testing the interface compatibility rather than actual cancellation behavior

	// Test Stat with context
	exists, err := memFS.StatWithContext(ctx, "tmp/memfs-ctx")
	if err != nil {
		t.Errorf("MemFS.StatWithContext failed: %v", err)
	}
	if !exists {
		t.Errorf("MemFS.StatWithContext should return true for existing directory")
	}

	exists, err = realFS.StatWithContext(ctx, realCtxDir)
	if err != nil {
		t.Errorf("RealFS.StatWithContext failed: %v", err)
	}
	if !exists {
		t.Errorf("RealFS.StatWithContext should return true for existing directory")
	}

	// Test RemoveAll with context
	if err := memFS.RemoveAllWithContext(ctx, "tmp/memfs-ctx"); err != nil {
		t.Errorf("MemFS.RemoveAllWithContext failed: %v", err)
	}

	exists, _ = memFS.StatWithContext(ctx, "tmp/memfs-ctx")
	if exists {
		t.Errorf("Directory should be removed after MemFS.RemoveAllWithContext")
	}

	if err := realFS.RemoveAllWithContext(ctx, realCtxDir); err != nil {
		t.Errorf("RealFS.RemoveAllWithContext failed: %v", err)
	}

	exists, _ = realFS.StatWithContext(ctx, realCtxDir)
	if exists {
		t.Errorf("Directory should be removed after RealFS.RemoveAllWithContext")
	}
}
