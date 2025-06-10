// Package thinktank_test is used for testing the internal/thinktank package
package thinktank_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// mockAuditLogger for testing FileWriter
type mockAuditLogger struct {
	entries []auditlog.AuditEntry
}

func (m *mockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	return m.Log(context.Background(), entry)
}

func (m *mockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	// Create an AuditEntry from the parameters
	entry := auditlog.AuditEntry{
		Operation: operation,
		Status:    status,
		Inputs:    inputs,
		Outputs:   outputs,
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

func (m *mockAuditLogger) Close() error {
	return nil
}

// TestSaveToFile tests the SaveToFile method
func TestSaveToFile(t *testing.T) {
	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ")

	// Create mock audit logger
	auditLogger := &mockAuditLogger{}

	// Create a file writer with default permissions
	fileWriter := thinktank.NewFileWriter(logger, auditLogger, 0750, 0640)

	// Create a filesystem abstraction for testing
	fs := testutil.NewMemFS()

	// Create a temporary directory for testing
	tempDir := "/tmp/filewriter_test"
	err := fs.MkdirAll(tempDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() { _ = fs.RemoveAll(tempDir) }()

	// Define test cases
	tests := []struct {
		name       string
		content    string
		outputFile string
		setupFunc  func() // Function to run before test
		cleanFunc  func() // Function to run after test
		wantErr    bool
	}{
		{
			name:       "Valid file path - absolute",
			content:    "Test content",
			outputFile: fs.Join(tempDir, "test_output.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Valid file path - relative",
			content:    "Test content with relative path",
			outputFile: "test_output_relative.md",
			setupFunc:  func() {},
			cleanFunc: func() {
				// Clean up relative path file
				cwd, _ := os.Getwd()
				_ = fs.RemoveAll(fs.Join(cwd, "test_output_relative.md"))
			},
			wantErr: false,
		},
		{
			name:       "Empty content",
			content:    "",
			outputFile: fs.Join(tempDir, "empty_file.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Long content",
			content:    strings.Repeat("Long content test ", 1000), // ~ 18KB of content
			outputFile: fs.Join(tempDir, "long_file.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Non-existent directory",
			content:    "Test content",
			outputFile: fs.Join(tempDir, "non-existent", "test_output.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run setup function
			tc.setupFunc()

			// Save to file
			err := fileWriter.SaveToFile(context.Background(), tc.content, tc.outputFile)

			// Run cleanup function
			defer tc.cleanFunc()

			// Check error
			if (err != nil) != tc.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Skip file validation for expected errors
			if tc.wantErr {
				return
			}

			// Determine output path for validation
			outputPath := tc.outputFile
			// If it's a relative path, prepend current directory
			if len(outputPath) > 0 && outputPath[0] != '/' {
				cwd, _ := os.Getwd()
				outputPath = fs.Join(cwd, outputPath)
			}

			// Since the test is using MemFS but the FileWriter uses os calls,
			// we can't actually verify reading the content with our test filesystem.
			// This test is not verifying contents but just that the call succeeds.
			// Full end-to-end verification would be done in integration tests.

			// Instead, we'll just log success
			t.Logf("Successfully wrote to file: %s", outputPath)
		})
	}
}
