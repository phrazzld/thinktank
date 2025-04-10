package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestReadTaskFromFileErrorCases tests all error conditions in readTaskFromFile
func TestReadTaskFromFileErrorCases(t *testing.T) {
	// Create a test directory structure
	tempDir, err := os.MkdirTemp("", "task-file-tests-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test logger
	logger := newTestLogger()

	// Test all error cases using subtests
	t.Run("FileNotFound", func(t *testing.T) {
		// Test with a non-existent file
		nonExistentFilePath := filepath.Join(tempDir, "non-existent-file.txt")

		// Call readTaskFromFile with a non-existent file
		_, err := readTaskFromFile(nonExistentFilePath, logger)

		// Check that it returns the expected error
		if err == nil {
			t.Fatal("Expected an error for non-existent file, got nil")
		}

		// Check that the error is wrapped with the correct sentinel error
		if !errors.Is(err, ErrTaskFileNotFound) {
			t.Errorf("Error %v is not ErrTaskFileNotFound", err)
		}
	})

	t.Run("FileIsDirectory", func(t *testing.T) {
		// Create a test directory
		dirPath := filepath.Join(tempDir, "directory-not-file")
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Call readTaskFromFile with a directory path
		_, err = readTaskFromFile(dirPath, logger)

		// Check that it returns the expected error
		if err == nil {
			t.Fatal("Expected an error for directory path, got nil")
		}

		// Check that the error is wrapped with the correct sentinel error
		if !errors.Is(err, ErrTaskFileIsDir) {
			t.Errorf("Error %v is not ErrTaskFileIsDir", err)
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		// Create an empty file
		emptyFilePath := filepath.Join(tempDir, "empty-file.txt")
		emptyFile, err := os.Create(emptyFilePath)
		if err != nil {
			t.Fatalf("Failed to create empty test file: %v", err)
		}
		emptyFile.Close()

		// Call readTaskFromFile with an empty file
		_, err = readTaskFromFile(emptyFilePath, logger)

		// Check that it returns the expected error
		if err == nil {
			t.Fatal("Expected an error for empty file, got nil")
		}

		// Check that the error is wrapped with the correct sentinel error
		if !errors.Is(err, ErrTaskFileEmpty) {
			t.Errorf("Error %v is not ErrTaskFileEmpty", err)
		}
	})

	t.Run("WhitespaceOnlyFile", func(t *testing.T) {
		// Create a file with only whitespace
		whitespaceFilePath := filepath.Join(tempDir, "whitespace-file.txt")
		whitespaceFile, err := os.Create(whitespaceFilePath)
		if err != nil {
			t.Fatalf("Failed to create whitespace test file: %v", err)
		}
		// Write whitespace to the file
		_, err = whitespaceFile.WriteString("   \n\t  \n")
		if err != nil {
			t.Fatalf("Failed to write to whitespace test file: %v", err)
		}
		whitespaceFile.Close()

		// Call readTaskFromFile with a whitespace-only file
		_, err = readTaskFromFile(whitespaceFilePath, logger)

		// Check that it returns the expected error
		if err == nil {
			t.Fatal("Expected an error for whitespace-only file, got nil")
		}

		// Check that the error is wrapped with the correct sentinel error
		if !errors.Is(err, ErrTaskFileEmpty) {
			t.Errorf("Error %v is not ErrTaskFileEmpty", err)
		}
	})

	// Note: Testing permission errors is challenging in a cross-platform way
	// We'll check if we can change permissions on the temp file system
	t.Run("PermissionDenied", func(t *testing.T) {
		// Create a file
		permFilePath := filepath.Join(tempDir, "no-read-permission.txt")
		permFile, err := os.Create(permFilePath)
		if err != nil {
			t.Fatalf("Failed to create permission test file: %v", err)
		}

		// Write some content
		_, err = permFile.WriteString("Test content")
		if err != nil {
			t.Fatalf("Failed to write to permission test file: %v", err)
		}
		permFile.Close()

		// Try to make it unreadable
		err = os.Chmod(permFilePath, 0)
		if err != nil {
			// If chmod fails, skip this test as we can't reliably test this condition
			t.Skip("Unable to change file permissions for testing permission denied")
		}

		// Call readTaskFromFile with a file without read permissions
		_, err = readTaskFromFile(permFilePath, logger)

		// Check that it returns an error
		if err == nil {
			// If no error, it might be because the file system or user permissions
			// allow reading regardless of file permissions (e.g., if running as admin/root)
			t.Skip("No error returned for permission test - skipping as this may be due to elevated privileges")
		}

		// Check that the error is wrapped with the correct sentinel error
		if !errors.Is(err, ErrTaskFileReadPermission) {
			t.Errorf("Error %v is not ErrTaskFileReadPermission", err)
		}
	})
}

// TestReadTaskFromFileSuccess tests the successful case
func TestReadTaskFromFileSuccess(t *testing.T) {
	// Create a temporary file with content
	tempFile, err := os.CreateTemp("", "task-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary test file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	expectedContent := "This is the task description content"
	_, err = tempFile.WriteString(expectedContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary test file: %v", err)
	}
	tempFile.Close()

	// Create a test logger
	logger := newTestLogger()

	// Call readTaskFromFile with a valid file
	content, err := readTaskFromFile(tempFile.Name(), logger)

	// Check that it doesn't return an error
	if err != nil {
		t.Fatalf("Unexpected error for valid file: %v", err)
	}

	// Check that the content is correct
	if content != expectedContent {
		t.Errorf("Content doesn't match. Expected %q, got %q", expectedContent, content)
	}
}

// TestRelativePathHandling tests handling of relative paths
func TestRelativePathHandling(t *testing.T) {
	// Create a temporary file in the current directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create a subdirectory for the test file
	testDir := filepath.Join(cwd, "test-relative-path")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a test file in the subdirectory
	relativePath := filepath.Join("test-relative-path", "relative-task.txt")
	absolutePath := filepath.Join(cwd, relativePath)

	// Create file with content
	err = os.WriteFile(absolutePath, []byte("Relative path test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create relative path test file: %v", err)
	}

	// Create a test logger
	logger := newTestLogger()

	// Call readTaskFromFile with the relative path
	content, err := readTaskFromFile(relativePath, logger)

	// Check that it doesn't return an error
	if err != nil {
		t.Fatalf("Unexpected error for relative path: %v", err)
	}

	// Check that the content is correct
	if content != "Relative path test content" {
		t.Errorf("Content doesn't match. Expected %q, got %q",
			"Relative path test content", content)
	}
}
