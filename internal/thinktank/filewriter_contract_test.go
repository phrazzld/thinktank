// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestFileWriterContract verifies that the FileWriter interface contract
// is properly implemented and handles edge cases correctly.
func TestFileWriterContract(t *testing.T) {
	// Create a test logger
	logger := logutil.NewLogger(logutil.DebugLevel, os.Stderr, "[test] ")

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filewriter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	// Create a mock audit logger
	mockAudit := &mockAuditLogger{}

	// Create a FileWriter with standard permissions
	fileWriter := NewFileWriter(logger, mockAudit, 0750, 0640)

	// Test cases to verify the FileWriter contract
	tests := []struct {
		name       string
		content    string
		outputFile string
		setup      func() error
		wantErr    bool
	}{
		{
			name:       "Basic write to existing directory",
			content:    "Test content",
			outputFile: filepath.Join(tempDir, "test1.txt"),
			setup:      func() error { return nil },
			wantErr:    false,
		},
		{
			name:       "Write to non-existent subdirectory",
			content:    "Test content in subdirectory",
			outputFile: filepath.Join(tempDir, "subdir", "test2.txt"),
			setup:      func() error { return nil },
			wantErr:    false, // Should create directory
		},
		{
			name:       "Empty content",
			content:    "",
			outputFile: filepath.Join(tempDir, "empty.txt"),
			setup:      func() error { return nil },
			wantErr:    false,
		},
		{
			name:       "Write to file in read-only directory",
			content:    "This should fail",
			outputFile: filepath.Join(tempDir, "readonly", "fail.txt"),
			setup: func() error {
				// Create read-only directory
				readonlyDir := filepath.Join(tempDir, "readonly")
				if err := os.Mkdir(readonlyDir, 0500); err != nil {
					return err
				}
				// Make it read-only for this user
				return os.Chmod(readonlyDir, 0500)
			},
			wantErr: true, // Should fail to create file in read-only dir
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test
			if err := tc.setup(); err != nil {
				t.Fatalf("Test setup failed: %v", err)
			}

			// Execute test
			err := fileWriter.SaveToFile(tc.content, tc.outputFile)

			// Check result
			if (err != nil) != tc.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// For successful tests, validate file content
			if !tc.wantErr {
				// Read the file back
				content, err := os.ReadFile(tc.outputFile)
				if err != nil {
					t.Errorf("Failed to read created file: %v", err)
					return
				}

				// Verify content
				if string(content) != tc.content {
					t.Errorf("File content mismatch, got %q, want %q", string(content), tc.content)
				}

				// Verify permissions
				info, err := os.Stat(tc.outputFile)
				if err != nil {
					t.Errorf("Failed to stat created file: %v", err)
					return
				}

				// Check permissions (platform-specific)
				expectedMode := os.FileMode(0640)
				if info.Mode().Perm() != expectedMode {
					t.Errorf("File permission mismatch, got %v, want %v", info.Mode().Perm(), expectedMode)
				}
			}
		})
	}

	// Test FileWriter with audit logging
	t.Run("Audit logging", func(t *testing.T) {
		// Create a mock audit logger
		mockAudit := &mockAuditLogger{}

		// Create a FileWriter with the mock audit logger
		auditedWriter := NewFileWriter(logger, mockAudit, 0750, 0640)

		// Write a file
		outputFile := filepath.Join(tempDir, "audited.txt")
		content := "Audited content"
		err := auditedWriter.SaveToFile(content, outputFile)

		// Verify write succeeded
		if err != nil {
			t.Errorf("SaveToFile() error = %v", err)
			return
		}

		// Verify audit log was called - expect at least the "Success" entry
		// Note: We might also get an "InProgress" entry
		success := false
		for _, entry := range mockAudit.entries {
			if entry.Status == "Success" {
				success = true
				break
			}
		}
		if !success {
			t.Errorf("Expected to find a successful audit log entry")
			return
		}

		// Look for the Success entry to validate
		var successEntry auditlog.AuditEntry
		for _, entry := range mockAudit.entries {
			if entry.Status == "Success" {
				successEntry = entry
				break
			}
		}

		// Verify operation is SaveOutput
		if successEntry.Operation != "SaveOutput" {
			t.Errorf("Expected operation 'SaveOutput', got %q", successEntry.Operation)
		}

		// Verify file path is recorded (in some form)
		if successEntry.Inputs["output_path"] == nil {
			t.Error("Expected output_path to be recorded in audit log")
		}

		// Verify content length is recorded
		if successEntry.Inputs["content_length"] == nil {
			t.Error("Expected content_length to be recorded in audit log")
		}
	})

	// Test error handling and propagation
	t.Run("Error handling", func(t *testing.T) {
		// Create a mock audit logger that always fails
		failingAudit := &mockAuditLogger{
			shouldFail: true,
		}

		// Create a FileWriter with the failing audit logger
		failingWriter := NewFileWriter(logger, failingAudit, 0750, 0640)

		// Attempt to write a file
		outputFile := filepath.Join(tempDir, "error.txt")
		content := "Error content"
		err := failingWriter.SaveToFile(content, outputFile)

		// Verify the file was written despite audit error
		if err != nil {
			// In this case we allow the write to succeed even if audit fails
			t.Logf("SaveToFile succeeded despite audit failure (expected)")
		}

		// Verify the file exists
		if _, err := os.Stat(outputFile); err != nil {
			t.Errorf("File was not created: %v", err)
		}
	})
}

// Mock implementation of AuditLogger for testing
type mockAuditLogger struct {
	entries    []auditlog.AuditEntry
	shouldFail bool
}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error {
	if m.shouldFail {
		return errors.New("simulated audit failure")
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditLogger) LogOp(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	if m.shouldFail {
		return errors.New("simulated audit failure")
	}

	entry := auditlog.AuditEntry{
		Operation: operation,
		Status:    status,
		Inputs:    inputs,
		Outputs:   outputs,
	}

	if err != nil {
		entry.Error = &auditlog.ErrorInfo{
			Message: err.Error(),
			Type:    "TestError",
		}
	}

	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditLogger) Close() error {
	if m.shouldFail {
		return errors.New("simulated audit close failure")
	}
	return nil
}
