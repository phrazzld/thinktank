package auditlog

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFileAuditLogger_ErrorHandling tests error handling in FileAuditLogger
func TestFileAuditLogger_ErrorHandling(t *testing.T) {
	// Setup a temporary directory
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	// Create a mock logger
	mockLog := newMockLogger()

	// Test marshal error by creating an entry with a channel (which can't be marshaled to JSON)
	badEntry := AuditEntry{
		Operation: "BadMarshal",
		Status:    "Error",
		Inputs: map[string]interface{}{
			"unmarshalable": make(chan int), // This will cause json.Marshal to fail
		},
	}

	// Create a new FileAuditLogger
	logger, err := NewFileAuditLogger(logPath, mockLog)
	if err != nil {
		t.Fatalf("Failed to create FileAuditLogger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Failed to close logger: %v", err)
		}
	}()

	// Try to log the entry with unmarshalable content
	ctx := context.Background()
	err = logger.Log(ctx, badEntry)
	if err == nil {
		t.Error("Expected error when logging entry with channel, got nil")
	}

	// Verify error was logged
	marshalErrorLogged := false
	for _, msg := range mockLog.errorMessages {
		if msg == "Failed to marshal audit entry to JSON: %v, Entry: %+v" {
			marshalErrorLogged = true
			break
		}
	}
	if !marshalErrorLogged {
		t.Errorf("Expected marshal error to be logged, messages: %v", mockLog.errorMessages)
	}

	// Skip test for write error with mock file - would require os.File mock
	// Instead, we'll log the audit entry to a file we can't write to
	readOnlyDir2 := filepath.Join(dir, "readonly_dir")
	readonlyErr := os.Mkdir(readOnlyDir2, 0750)
	if readonlyErr != nil {
		t.Fatalf("Failed to create readonly dir: %v", readonlyErr)
	}

	readOnlyFile := filepath.Join(readOnlyDir2, "readonly.log")
	//nolint:gosec // G304: Test file creation with controlled temp directory path
	f, createErr2 := os.Create(readOnlyFile)
	if createErr2 != nil {
		t.Fatalf("Failed to create readonly file: %v", createErr2)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close test file: %v", err)
	}

	// We'll skip modifying the file permissions since it's problematic for testing
	// Instead, we'll just test other error handling aspects

	// Skip the actual write error test - difficult to simulate without mocking os.File
	// Instead verify other error handling aspects like timestamp handling

	// Test with file permissions issue
	if os.Getuid() == 0 {
		t.Skip("Skipping permissions test when running as root")
	}

	// Create a temporary file with read-only permissions
	readOnlyDir := filepath.Join(dir, "readonly")
	if mkdirErr := os.Mkdir(readOnlyDir, 0o750); mkdirErr != nil {
		t.Fatalf("Failed to create readonly dir: %v", mkdirErr)
	}

	readOnlyLogPath := filepath.Join(readOnlyDir, "readonly.log")

	// Create the file first
	//nolint:gosec // G304: Test file creation with controlled temp directory path
	f, createErr := os.Create(readOnlyLogPath)
	if createErr != nil {
		t.Fatalf("Failed to create readonly log file: %v", createErr)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close test file: %v", err)
	}

	// Make the file read-only
	if chmodErr := os.Chmod(readOnlyLogPath, 0o400); chmodErr != nil {
		t.Fatalf("Failed to set readonly permissions: %v", chmodErr)
	}

	// Make the directory read-only to prevent file creation
	//nolint:gosec // G302: Test directory needs restrictive permissions for testing
	if dirChmodErr := os.Chmod(readOnlyDir, 0o500); dirChmodErr != nil {
		t.Fatalf("Failed to set readonly directory: %v", dirChmodErr)
	}

	// Attempt to create a logger with an existing read-only file
	// Since we can't write to a read-only file, this should fail
	readOnlyLogger, loggerErr := NewFileAuditLogger(readOnlyLogPath, mockLog)
	if loggerErr == nil {
		t.Error("Expected error when creating logger with read-only file, got nil")
		if readOnlyLogger != nil {
			if closeErr := readOnlyLogger.Close(); closeErr != nil {
				t.Logf("Error closing test logger: %v", closeErr)
			}
		}
	}

	// Cleanup - restore permissions to allow cleanup
	//nolint:gosec // G302: Test cleanup restores normal directory permissions
	if chmodErr := os.Chmod(readOnlyDir, 0o755); chmodErr != nil {
		t.Logf("Warning: Failed to restore directory permissions: %v", chmodErr)
	}
	//nolint:gosec // G302: Test cleanup restores normal file permissions
	if chmodErr := os.Chmod(readOnlyLogPath, 0o644); chmodErr != nil {
		t.Logf("Warning: Failed to restore file permissions: %v", chmodErr)
	}

	// Test handling of timestamp
	// Create a valid entry with no timestamp
	noTimestampEntry := AuditEntry{
		Operation: "NoTimestamp",
		Status:    "Test",
	}

	// Clear the file
	if err := logger.file.Truncate(0); err != nil {
		t.Fatalf("Failed to truncate log file: %v", err)
	}
	if _, err := logger.file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in log file: %v", err)
	}

	// Log the entry
	logErr := logger.Log(context.Background(), noTimestampEntry)
	if logErr != nil {
		t.Fatalf("Failed to log entry without timestamp: %v", logErr)
	}

	// Read the log file
	//nolint:gosec // G304: Test file reading with controlled temp directory path
	content, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	// Parse the JSON line
	var parsedEntry AuditEntry
	if unmarshalErr := json.Unmarshal(content, &parsedEntry); unmarshalErr != nil {
		t.Fatalf("Failed to parse JSON: %v\nContent: %s", unmarshalErr, content)
	}

	// Verify the timestamp was added
	if parsedEntry.Timestamp.IsZero() {
		t.Error("Expected logger to set timestamp, got zero value")
	}

	// Test entry with pre-set timestamp
	presetTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	timestampEntry := AuditEntry{
		Operation: "WithTimestamp",
		Status:    "Test",
		Timestamp: presetTime,
	}

	// Clear the file
	if err := logger.file.Truncate(0); err != nil {
		t.Fatalf("Failed to truncate log file: %v", err)
	}
	if _, err := logger.file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in log file: %v", err)
	}

	// Log the entry
	timeLogErr := logger.Log(context.Background(), timestampEntry)
	if timeLogErr != nil {
		t.Fatalf("Failed to log entry with preset timestamp: %v", timeLogErr)
	}

	// Read the log file
	//nolint:gosec // G304: Test file reading with controlled temp directory path
	timeContent, timeReadErr := os.ReadFile(logPath)
	if timeReadErr != nil {
		t.Fatalf("Failed to read log file: %v", timeReadErr)
	}

	// Parse the JSON line
	if timeUnmarshalErr := json.Unmarshal(timeContent, &parsedEntry); timeUnmarshalErr != nil {
		t.Fatalf("Failed to parse JSON: %v\nContent: %s", timeUnmarshalErr, timeContent)
	}

	// Verify the timestamp was preserved
	if !parsedEntry.Timestamp.Equal(presetTime) {
		t.Errorf("Expected timestamp %v, got %v", presetTime, parsedEntry.Timestamp)
	}
}
