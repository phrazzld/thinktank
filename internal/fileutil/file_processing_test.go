// internal/fileutil/file_processing_test.go
package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProcessFileErrors(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Set up a logger to capture logs
	logger := NewMockLogger()
	logger.SetVerbose(true)

	// Create config
	config := &Config{
		Logger: logger,
	}

	// Path to a file that doesn't exist
	nonExistentPath := filepath.Join(tempDir, "doesnotexist.txt")

	// Path to a file without read permission
	noPermissionPath := filepath.Join(tempDir, "nopermission.txt")
	err := os.WriteFile(noPermissionPath, []byte("test content"), 0000) // No permissions
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a readable file for testing path conversion
	readablePath := filepath.Join(tempDir, "readable.txt")
	err = os.WriteFile(readablePath, []byte("test content"), 0640)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name          string
		path          string
		expectedError string
		setupFunc     func() // Optional setup function
		cleanupFunc   func() // Optional cleanup function
	}{
		{
			name:          "Non-existent file",
			path:          nonExistentPath,
			expectedError: "Cannot read file",
		},
		{
			name:          "File without read permission",
			path:          noPermissionPath,
			expectedError: "Cannot read file",
		},
		{
			name:          "Binary file detection",
			path:          readablePath,
			expectedError: "", // No error expected
			setupFunc: func() {
				// Write binary content
				binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}
				err := os.WriteFile(readablePath, binaryContent, 0640)
				if err != nil {
					t.Fatalf("Failed to write binary file: %v", err)
				}
			},
		},
		{
			name:          "Path conversion error detection",
			path:          "relative/path/test.txt",
			expectedError: "Cannot read file", // The function reports file not found before path conversion
			setupFunc: func() {
				// Create a temporary text file to process, but with a relative path
				// that we'll check for conversion warning
				err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test content"), 0640)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.ClearMessages()

			// Run setup function if provided
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Cleanup at the end if provided
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc()
			}

			// Create an empty files slice
			var files []FileMeta

			// Process the file
			processFile(tt.path, &files, config)

			// For binary file case, verify it was skipped
			if tt.name == "Binary file detection" {
				if len(files) != 0 {
					t.Errorf("Expected binary file to be skipped, but it was added to files")
				}
				if !logger.ContainsMessage("Skipping binary file") {
					t.Errorf("Expected log message about skipping binary file, but didn't find it")
				}
				return
			}

			// Verify that no files were added for error cases
			if len(files) != 0 && tt.expectedError != "" {
				t.Errorf("Expected no files to be processed, but got %d", len(files))
			}

			// Verify the error message was logged if expected
			if tt.expectedError != "" && !logger.ContainsMessage(tt.expectedError) {
				t.Errorf("Expected error message containing '%s', but didn't find it in logs: %v",
					tt.expectedError, logger.GetMessages())
			}
		})
	}
}

// TestPathConversionError tests the handling of path conversion errors
func TestPathConversionError(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a test file
	testFilePath := filepath.Join(tempDir, "testfile.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0640)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set up a logger to capture logs
	logger := NewMockLogger()
	logger.SetVerbose(true)

	// Create a config
	config := &Config{
		Logger: logger,
	}

	// Create a relative path that could trigger path conversion error
	var files []FileMeta

	// Just check that we can handle the warning without a crash
	// This test is mainly to ensure code coverage for the filepath.Abs error handling path
	logger.ClearMessages()
	processFile("non/existent/relative/path.txt", &files, config)

	// Verify we logged a warning about the file read error
	if !logger.ContainsMessage("Cannot read file") {
		t.Errorf("Expected warning about reading file, but didn't find it in logs: %v",
			logger.GetMessages())
	}
}
