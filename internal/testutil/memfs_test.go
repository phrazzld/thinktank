package testutil

import (
	"context"
	"testing"
)

// TestMemFSBasicOperations tests basic file operations with MemFS
func TestMemFSBasicOperations(t *testing.T) {
	fs := NewMemFS()

	// Test directory creation
	err := fs.MkdirAll("test/dir1", 0750)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Verify directory exists
	exists, err := fs.Stat("test/dir1")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if !exists {
		t.Fatal("Directory should exist")
	}

	// Test file writing
	testData := []byte("Hello, world!")
	err = fs.WriteFile("test/dir1/file1.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	exists, err = fs.Stat("test/dir1/file1.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if !exists {
		t.Fatal("File should exist")
	}

	// Test file reading
	data, err := fs.ReadFile("test/dir1/file1.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "Hello, world!" {
		t.Fatalf("Expected 'Hello, world!', got '%s'", string(data))
	}

	// Test error case: reading non-existent file
	_, err = fs.ReadFile("test/dir1/nonexistent.txt")
	if err == nil {
		t.Fatal("Expected error when reading non-existent file, got nil")
	}

	// Test error case: writing to non-existent directory
	err = fs.WriteFile("nonexistent/dir/file.txt", []byte("data"), 0640)
	if err == nil {
		t.Fatal("Expected error when writing to non-existent directory, got nil")
	}
}

// TestMemFSRemoveAll tests the RemoveAll functionality
func TestMemFSRemoveAll(t *testing.T) {
	fs := NewMemFS()

	// Create nested directory structure
	err := fs.MkdirAll("test/dir1/subdir", 0750)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Create files
	testData := []byte("test data")
	err = fs.WriteFile("test/dir1/file1.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/dir1/subdir/file2.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// RemoveAll test/dir1
	err = fs.RemoveAll("test/dir1")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify directory is gone
	exists, _ := fs.Stat("test/dir1")
	if exists {
		t.Fatal("Directory should not exist after RemoveAll")
	}

	// Verify parent still exists
	exists, _ = fs.Stat("test")
	if !exists {
		t.Fatal("Parent directory should still exist")
	}

	// Verify files are gone
	exists, _ = fs.Stat("test/dir1/file1.txt")
	if exists {
		t.Fatal("File should not exist after RemoveAll")
	}
	exists, _ = fs.Stat("test/dir1/subdir/file2.txt")
	if exists {
		t.Fatal("File in subdirectory should not exist after RemoveAll")
	}
}

// TestMemFSList tests the List functionality
func TestMemFSList(t *testing.T) {
	fs := NewMemFS()

	// Create directory structure
	err := fs.MkdirAll("test/dir1", 0750)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	err = fs.MkdirAll("test/dir2", 0750)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Create files
	testData := []byte("test data")
	err = fs.WriteFile("test/file1.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/file2.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/dir1/file3.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// List test directory
	entries, err := fs.List("test")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify we have the expected entries
	expectedEntries := map[string]bool{
		"dir1":      true,
		"dir2":      true,
		"file1.txt": true,
		"file2.txt": true,
	}

	if len(entries) != len(expectedEntries) {
		t.Fatalf("Expected %d entries, got %d", len(expectedEntries), len(entries))
	}

	for _, entry := range entries {
		if !expectedEntries[entry] {
			t.Fatalf("Unexpected entry: %s", entry)
		}
	}

	// List subdirectory
	entries, err = fs.List("test/dir1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 1 || entries[0] != "file3.txt" {
		t.Fatalf("Expected [file3.txt], got %v", entries)
	}

	// Error case: list non-existent directory
	_, err = fs.List("nonexistent")
	if err == nil {
		t.Fatal("Expected error when listing non-existent directory, got nil")
	}
}

// TestMemFSSearchFiles tests the SearchFiles functionality
func TestMemFSSearchFiles(t *testing.T) {
	fs := NewMemFS()

	// Create files
	testData := []byte("test data")
	err := fs.MkdirAll("test/dir1", 0750)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	err = fs.WriteFile("test/file1.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/file2.log", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/dir1/file3.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Search for *.txt files
	matches, err := fs.SearchFiles("test/*.txt")
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	if len(matches) != 1 || matches[0] != "test/file1.txt" {
		t.Fatalf("Expected [test/file1.txt], got %v", matches)
	}

	// Search for all files in test directory
	matches, err = fs.SearchFiles("test/*")
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(matches))
	}
}

// TestMemFSFindInFiles tests the FindInFiles functionality
func TestMemFSFindInFiles(t *testing.T) {
	fs := NewMemFS()

	// Create test files with content
	err := fs.MkdirAll("test", 0750)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Each file contains the word "test" in different formats and positions
	fileContent1 := "Line 1\nTest line 2\nLine 3 test\n"
	fileContent2 := "Other content\nNo match here\n"
	fileContent3 := "Test on line 1\nTest on line 2\n"

	err = fs.WriteFile("test/file1.txt", []byte(fileContent1), 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/file2.txt", []byte(fileContent2), 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	err = fs.WriteFile("test/file3.log", []byte(fileContent3), 0640)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Use lowercase "test" to find all instances
	results, err := fs.FindInFiles("test")
	if err != nil {
		t.Fatalf("FindInFiles failed: %v", err)
	}

	// Should match both file1.txt (line 3) and file3.log (which contains "Test")
	if len(results) != 2 {
		t.Fatalf("Expected 2 files with matches, got %d", len(results))
	}

	// Check file1.txt matches
	matches, ok := results["test/file1.txt"]
	if !ok {
		t.Fatal("Expected matches in test/file1.txt")
	}
	if !containsLine(matches, 3) { // "Line 3 test" is on line 3
		t.Fatalf("Expected match on line 3 in file1.txt, got %v", matches)
	}

	// Check file3.log matches
	matches, ok = results["test/file3.log"]
	if !ok {
		t.Fatal("Expected matches in test/file3.log")
	}
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches in file3.log, got %d", len(matches))
	}

	// Find "Test" in only .txt files
	results, err = fs.FindInFiles("Test", "*.txt")
	if err != nil {
		t.Fatalf("FindInFiles failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 .txt file with matches, got %d", len(results))
	}

	// Check file1.txt matches
	matches, ok = results["test/file1.txt"]
	if !ok {
		t.Fatal("Expected matches in test/file1.txt")
	}
	if !containsLine(matches, 2) { // "Test line 2" is on line 2
		t.Fatalf("Expected match on line 2 in file1.txt, got %v", matches)
	}
}

// TestMemFSPathOperations tests the path utility methods
func TestMemFSPathOperations(t *testing.T) {
	fs := NewMemFS()

	// Test Join
	path := fs.Join("test", "dir", "file.txt")
	expected := "test/dir/file.txt"
	if path != expected {
		t.Fatalf("Expected Join to return '%s', got '%s'", expected, path)
	}

	// Test Base
	base := fs.Base("test/dir/file.txt")
	if base != "file.txt" {
		t.Fatalf("Expected Base to return 'file.txt', got '%s'", base)
	}

	// Test Dir
	dir := fs.Dir("test/dir/file.txt")
	if dir != "test/dir" {
		t.Fatalf("Expected Dir to return 'test/dir', got '%s'", dir)
	}
}

// TestMemFSFilesystemIOImplementation tests that MemFS implements FilesystemIO
func TestMemFSFilesystemIOImplementation(t *testing.T) {
	// Get a type that satisfies FilesystemIO from the integration package
	// We do this indirectly since we can't import the integration package here
	var fs interface{} = NewMemFS()

	// Check if the essential non-context methods exist
	_, readFileOk := fs.(interface {
		ReadFile(path string) ([]byte, error)
	})
	_, writeFileOk := fs.(interface {
		WriteFile(path string, data []byte, perm int) error
	})
	_, mkdirAllOk := fs.(interface {
		MkdirAll(path string, perm int) error
	})
	_, removeAllOk := fs.(interface{ RemoveAll(path string) error })
	_, statOk := fs.(interface {
		Stat(path string) (bool, error)
	})

	// Check if the essential context-aware methods exist
	_, readFileWithContextOk := fs.(interface {
		ReadFileWithContext(ctx context.Context, path string) ([]byte, error)
	})
	_, writeFileWithContextOk := fs.(interface {
		WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error
	})
	_, mkdirAllWithContextOk := fs.(interface {
		MkdirAllWithContext(ctx context.Context, path string, perm int) error
	})
	_, removeAllWithContextOk := fs.(interface {
		RemoveAllWithContext(ctx context.Context, path string) error
	})
	_, statWithContextOk := fs.(interface {
		StatWithContext(ctx context.Context, path string) (bool, error)
	})

	// Check non-context methods
	if !readFileOk {
		t.Fatal("MemFS does not implement ReadFile method")
	}
	if !writeFileOk {
		t.Fatal("MemFS does not implement WriteFile method")
	}
	if !mkdirAllOk {
		t.Fatal("MemFS does not implement MkdirAll method")
	}
	if !removeAllOk {
		t.Fatal("MemFS does not implement RemoveAll method")
	}
	if !statOk {
		t.Fatal("MemFS does not implement Stat method")
	}

	// Check context-aware methods
	if !readFileWithContextOk {
		t.Fatal("MemFS does not implement ReadFileWithContext method")
	}
	if !writeFileWithContextOk {
		t.Fatal("MemFS does not implement WriteFileWithContext method")
	}
	if !mkdirAllWithContextOk {
		t.Fatal("MemFS does not implement MkdirAllWithContext method")
	}
	if !removeAllWithContextOk {
		t.Fatal("MemFS does not implement RemoveAllWithContext method")
	}
	if !statWithContextOk {
		t.Fatal("MemFS does not implement StatWithContext method")
	}
}

// TestMemFSContextOperations tests the context-aware file operations with MemFS
func TestMemFSContextOperations(t *testing.T) {
	fs := NewMemFS()
	ctx := context.Background()

	// Test context-aware directory creation
	err := fs.MkdirAllWithContext(ctx, "ctx-test/dir1", 0750)
	if err != nil {
		t.Fatalf("MkdirAllWithContext failed: %v", err)
	}

	// Verify directory exists using context-aware method
	exists, err := fs.StatWithContext(ctx, "ctx-test/dir1")
	if err != nil {
		t.Fatalf("StatWithContext failed: %v", err)
	}
	if !exists {
		t.Fatal("Directory should exist")
	}

	// Test context-aware file writing
	testData := []byte("Hello, context world!")
	err = fs.WriteFileWithContext(ctx, "ctx-test/dir1/file1.txt", testData, 0640)
	if err != nil {
		t.Fatalf("WriteFileWithContext failed: %v", err)
	}

	// Test context-aware file reading
	data, err := fs.ReadFileWithContext(ctx, "ctx-test/dir1/file1.txt")
	if err != nil {
		t.Fatalf("ReadFileWithContext failed: %v", err)
	}
	if string(data) != "Hello, context world!" {
		t.Fatalf("Expected 'Hello, context world!', got '%s'", string(data))
	}

	// Test context-aware error case: reading non-existent file
	_, err = fs.ReadFileWithContext(ctx, "ctx-test/dir1/nonexistent.txt")
	if err == nil {
		t.Fatal("Expected error when reading non-existent file with context, got nil")
	}

	// Test context cancellation
	// Note: we're not actually using the canceled context yet
	_, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context immediately

	// The MemFS implementation should check for context cancellation
	// but since it's in-memory and operations are fast, we may need
	// to add explicit context checks to make this test meaningful.
	// For now, we're just ensuring the methods exist and work with a valid context.

	// Test context-aware removal
	err = fs.RemoveAllWithContext(ctx, "ctx-test/dir1")
	if err != nil {
		t.Fatalf("RemoveAllWithContext failed: %v", err)
	}

	// Verify directory is gone using context-aware method
	exists, _ = fs.StatWithContext(ctx, "ctx-test/dir1")
	if exists {
		t.Fatal("Directory should not exist after RemoveAllWithContext")
	}
}

// Helper function to check if a slice of line numbers contains a specific line
func containsLine(lines []int, line int) bool {
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}
