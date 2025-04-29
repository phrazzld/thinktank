// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// fileWriterMockLogger is a minimal logger for testing file writing
type fileWriterMockLogger struct {
	errorCalled bool
	infoCalled  bool
	errorMsg    string
	infoMsg     string
}

func (l *fileWriterMockLogger) Error(format string, args ...interface{}) {
	l.errorCalled = true
	l.errorMsg = format
}

func (l *fileWriterMockLogger) Debug(format string, args ...interface{}) {}
func (l *fileWriterMockLogger) Info(format string, args ...interface{}) {
	l.infoCalled = true
	l.infoMsg = format
}
func (l *fileWriterMockLogger) Warn(format string, args ...interface{})   {}
func (l *fileWriterMockLogger) Fatal(format string, args ...interface{})  {}
func (l *fileWriterMockLogger) Printf(format string, args ...interface{}) {}
func (l *fileWriterMockLogger) Println(v ...interface{})                  {}
func (l *fileWriterMockLogger) GetLevel() logutil.LogLevel                { return logutil.InfoLevel }

// Context-aware logging methods
func (l *fileWriterMockLogger) DebugContext(ctx context.Context, format string, args ...interface{}) {
}
func (l *fileWriterMockLogger) InfoContext(ctx context.Context, format string, args ...interface{}) {
	l.infoCalled = true
	l.infoMsg = format
}
func (l *fileWriterMockLogger) WarnContext(ctx context.Context, format string, args ...interface{}) {}
func (l *fileWriterMockLogger) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	l.errorCalled = true
	l.errorMsg = format
}
func (l *fileWriterMockLogger) FatalContext(ctx context.Context, format string, args ...interface{}) {
}
func (l *fileWriterMockLogger) WithContext(ctx context.Context) logutil.LoggerInterface { return l }

func TestNewFileWriter(t *testing.T) {
	logger := &fileWriterMockLogger{}
	fw := NewFileWriter(logger, 0750, 0640)

	if fw == nil {
		t.Errorf("NewFileWriter() returned nil")
	}
}

func TestSaveToFile(t *testing.T) {
	// Create a filesystem abstraction for testing
	fs := testutil.NewRealFS()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "filewriter-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = fs.RemoveAll(tempDir) }()

	tests := []struct {
		name           string
		content        string
		outputFilePath string
		preparePath    func(string) error
		expectError    bool
	}{
		{
			name:           "Successful write to new file with absolute path",
			content:        "Test content",
			outputFilePath: filepath.Join(tempDir, "test1.txt"),
			preparePath:    nil,
			expectError:    false,
		},
		{
			name:           "Successful write to new file with relative path",
			content:        "Test content with relative path",
			outputFilePath: "test_relative.txt", // Will be relative to current working directory
			preparePath:    nil,
			expectError:    false,
		},
		{
			name:           "Successful overwrite of existing file",
			content:        "New content that should overwrite existing content",
			outputFilePath: filepath.Join(tempDir, "test_overwrite.txt"),
			preparePath: func(path string) error {
				return fs.WriteFile(path, []byte("Original content"), 0640)
			},
			expectError: false,
		},
		{
			name:           "Successfully creates directories in path",
			content:        "Test content in nested directory",
			outputFilePath: filepath.Join(tempDir, "nested", "dirs", "test.txt"),
			preparePath:    nil,
			expectError:    false,
		},
		{
			name:           "Error on unwritable directory",
			content:        "This should fail",
			outputFilePath: "/root/test_should_fail.txt", // Most users won't have write permission here
			preparePath:    nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare any test-specific setup
			if tt.preparePath != nil {
				err := tt.preparePath(tt.outputFilePath)
				if err != nil {
					t.Fatalf("Failed to prepare test path: %v", err)
				}
			}

			// Create a new logger for each test
			logger := &fileWriterMockLogger{}
			fw := NewFileWriter(logger, 0750, 0640)

			// Capture the current working directory before the test
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current working directory: %v", err)
			}

			// Execute the SaveToFile method
			err = fw.SaveToFile(tt.content, tt.outputFilePath)

			// Check if error matches expectation
			if (err != nil) != tt.expectError {
				t.Errorf("SaveToFile() error = %v, expectError %v", err, tt.expectError)
			}

			// For error cases, verify the logger was called and return early
			if tt.expectError {
				if !logger.errorCalled {
					t.Errorf("Expected error to be logged when SaveToFile returns error")
				}
				return
			}

			// For success cases, verify file was written correctly
			var filePath string
			if filepath.IsAbs(tt.outputFilePath) {
				filePath = tt.outputFilePath
			} else {
				// For relative paths, the file should be in the current working directory
				filePath = filepath.Join(cwd, tt.outputFilePath)
				// Clean up relative path files at the end of the test
				defer func() { _ = fs.RemoveAll(filePath) }()
			}

			// Read the file content using FilesystemIO
			content, err := fs.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// Verify content matches
			if string(content) != tt.content {
				t.Errorf("File content = %q, want %q", string(content), tt.content)
			}

			// Verify logger was called with success message
			if !logger.infoCalled {
				t.Errorf("Expected info to be logged on successful write")
			}
			if !strings.Contains(logger.infoMsg, "Successfully") {
				t.Errorf("Expected success message, got: %q", logger.infoMsg)
			}
		})
	}
}

func TestSaveToFile_ErrorConditions(t *testing.T) {
	// Create a filesystem abstraction for testing
	fs := testutil.NewRealFS()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "filewriter-error-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = fs.RemoveAll(tempDir) }()

	// Test error with current working directory
	t.Run("Error getting current working directory", func(t *testing.T) {
		// Create a mock logger to capture errors
		logger := &fileWriterMockLogger{}
		fw := NewFileWriter(logger, 0750, 0640)

		// Create a test file with invalid permissions for directory creation
		invalidPath := filepath.Join(tempDir, "invalid-dir")
		if err := os.Mkdir(invalidPath, 0444); err != nil { // Read-only directory
			t.Fatalf("Failed to create test directory: %v", err)
		}
		// Try to write to a path inside the read-only directory
		err := fw.SaveToFile("test content", filepath.Join(invalidPath, "subdir", "test.txt"))

		// Verify error was returned
		if err == nil {
			t.Errorf("Expected error when trying to create directory in read-only path, got nil")
		}

		// Verify error was logged
		if !logger.errorCalled {
			t.Errorf("Expected error to be logged")
		}

		// Verify file doesn't exist using FilesystemIO
		exists, _ := fs.Stat(filepath.Join(invalidPath, "subdir", "test.txt"))
		if exists {
			t.Errorf("File should not exist after error")
		}
	})
}
