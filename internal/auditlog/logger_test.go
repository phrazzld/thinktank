package auditlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockLogger is a simple in-memory implementation of StructuredLogger for testing
type mockLogger struct {
	events []AuditEvent
	closed bool
	err    error
}

// Log records the event in memory
func (m *mockLogger) Log(event AuditEvent) {
	m.events = append(m.events, event)
}

// Close marks the logger as closed and returns the predefined error
func (m *mockLogger) Close() error {
	m.closed = true
	return m.err
}

// newMockLogger creates a new mock logger with optional error on close
func newMockLogger(err error) *mockLogger {
	return &mockLogger{
		events: []AuditEvent{},
		closed: false,
		err:    err,
	}
}

// TestStructuredLoggerInterface verifies that the interface is correctly defined
// by using a mock implementation
func TestStructuredLoggerInterface(t *testing.T) {
	// Create a mock logger
	mockLog := newMockLogger(nil)
	
	// Create a variable of StructuredLogger type and assign the mock
	var logger StructuredLogger = mockLog
	
	// Create a test event
	testEvent := AuditEvent{
		Timestamp: time.Now().UTC(),
		Level:     "INFO",
		Operation: "TestOperation",
		Message:   "Test message",
	}
	
	// Test Log method
	logger.Log(testEvent)
	
	// Verify the event was recorded
	if len(mockLog.events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(mockLog.events))
	}
	
	// Check the event fields
	if mockLog.events[0].Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", mockLog.events[0].Message)
	}
	
	// Test Close method
	err := logger.Close()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	
	// Verify logger was closed
	if !mockLog.closed {
		t.Error("Expected logger to be closed, but it wasn't")
	}
}

// TestStructuredLoggerCloseError verifies that Close errors are properly returned
func TestStructuredLoggerCloseError(t *testing.T) {
	// Create a mock logger with an error
	expectedErr := errors.New("close error")
	mockLog := newMockLogger(expectedErr)
	
	// Create a variable of StructuredLogger type and assign the mock
	var logger StructuredLogger = mockLog
	
	// Test Close method with error
	err := logger.Close()
	
	// Verify the error
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestFileLoggerCreation tests the basic creation of a FileLogger
func TestFileLoggerCreation(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path
	logFilePath := filepath.Join(tempDir, "audit.log")
	
	// Create a FileLogger
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger: %v", err)
	}
	defer logger.Close() // Ensure file is closed
	
	// Check that the logger is not nil
	if logger == nil {
		t.Fatal("Expected FileLogger to be non-nil")
	}
	
	// Check that the file was created
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logFilePath)
	}
	
	// Verify the logger implements StructuredLogger
	var _ StructuredLogger = logger
}

// TestFileLoggerWithNestedDirectories tests creating a logger with a nested directory structure
func TestFileLoggerWithNestedDirectories(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path with nested directories that don't exist yet
	nestedDir := filepath.Join(tempDir, "a", "b", "c")
	logFilePath := filepath.Join(nestedDir, "audit.log")
	
	// Create a FileLogger - this should create all the nested directories
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger with nested directories: %v", err)
	}
	defer logger.Close()
	
	// Check that the directories were created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Errorf("Nested directory structure was not created at %s", nestedDir)
	}
	
	// Check that the file was created
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logFilePath)
	}
}

// TestFileLoggerEmptyPath tests that an error is returned for an empty path
func TestFileLoggerEmptyPath(t *testing.T) {
	// Try to create a FileLogger with an empty path
	logger, err := NewFileLogger("")
	
	// Expect an error
	if err == nil {
		t.Error("Expected error for empty path, but got nil")
		if logger != nil {
			logger.Close() // Clean up if somehow created
		}
	}
	
	// Check that the error message is clear
	if err != nil && !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error message to mention 'empty', got: %v", err)
	}
}

// TestFileLoggerInvalidPath tests that an error is returned for an invalid path
func TestFileLoggerInvalidPath(t *testing.T) {
	// Try to create a FileLogger with an invalid path (assuming we can find one)
	// This is platform-dependent, but we'll try a path with invalid characters
	invalidPath := filepath.Join(string([]byte{0}), "audit.log")
	
	// Try to create a FileLogger with an invalid path
	logger, err := NewFileLogger(invalidPath)
	
	// Expect an error
	if err == nil {
		t.Error("Expected error for invalid path, but got nil")
		if logger != nil {
			logger.Close() // Clean up if somehow created
		}
	}
}

// TestFileLoggerPermissionDenied tests that an error is returned when permission is denied
func TestFileLoggerPermissionDenied(t *testing.T) {
	// Skip on Windows as permissions work differently
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Make the directory read-only
	if err := os.Chmod(tempDir, 0500); err != nil { // r-x permissions
		t.Fatalf("Failed to set directory permissions: %v", err)
	}
	
	// Try to create a FileLogger in the read-only directory
	logFilePath := filepath.Join(tempDir, "audit.log")
	logger, err := NewFileLogger(logFilePath)
	
	// Restore permissions to ensure cleanup can succeed
	os.Chmod(tempDir, 0700)
	
	// Expect an error
	if err == nil {
		t.Error("Expected error for permission denied, but got nil")
		if logger != nil {
			logger.Close() // Clean up if somehow created
		}
	}
}

// TestFileLoggerLog tests the basic logging functionality
func TestFileLoggerLog(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path
	logFilePath := filepath.Join(tempDir, "audit.log")
	
	// Create a FileLogger
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger: %v", err)
	}
	
	// Create a test event
	testEvent := AuditEvent{
		Timestamp: time.Date(2025, 4, 9, 12, 0, 0, 0, time.UTC),
		Level:     "INFO",
		Operation: "TestOperation",
		Message:   "Test log message",
	}
	
	// Log the event
	logger.Log(testEvent)
	
	// Close the logger to ensure the file is written
	err = logger.Close()
	if err != nil {
		t.Fatalf("Failed to close FileLogger: %v", err)
	}
	
	// Read the file contents
	fileContent, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Check that the file contains JSON
	var parsedEvent map[string]interface{}
	err = json.Unmarshal(fileContent, &parsedEvent)
	if err != nil {
		t.Fatalf("File does not contain valid JSON: %v", err)
	}
	
	// Verify the event was logged correctly
	if parsedEvent["level"] != "INFO" {
		t.Errorf("Expected level 'INFO', got %v", parsedEvent["level"])
	}
	if parsedEvent["operation"] != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got %v", parsedEvent["operation"])
	}
	if parsedEvent["message"] != "Test log message" {
		t.Errorf("Expected message 'Test log message', got %v", parsedEvent["message"])
	}
}

// TestConcurrentLogging tests that the Log method is safe for concurrent use
func TestConcurrentLogging(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path
	logFilePath := filepath.Join(tempDir, "concurrent.log")
	
	// Create a FileLogger
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger: %v", err)
	}
	defer logger.Close()
	
	// Number of concurrent goroutines
	numGoroutines := 100
	// Number of log events per goroutine
	eventsPerGoroutine := 10
	
	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Start multiple goroutines that log events concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < eventsPerGoroutine; j++ {
				event := NewAuditEvent(
					"INFO",
					fmt.Sprintf("Operation-%d-%d", id, j),
					fmt.Sprintf("Message from goroutine %d, event %d", id, j),
				)
				logger.Log(event)
			}
		}(i)
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Close the logger to ensure all data is flushed
	logger.Close()
	
	// Verify the log file exists
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logFilePath)
	}
	
	// Read the file contents
	fileContent, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Split by newlines to get individual JSON lines
	lines := strings.Split(string(fileContent), "\n")
	// Remove last empty line if present
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	
	// Check that we have the expected number of log entries
	expectedLines := numGoroutines * eventsPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log entries, got %d", expectedLines, len(lines))
	}
	
	// Verify each line is valid JSON
	for i, line := range lines {
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

// TestLogWithNilFile tests that the Log method handles a nil file gracefully
func TestLogWithNilFile(t *testing.T) {
	// Create a logger with a nil file
	logger := &FileLogger{file: nil}
	
	// Create a test event
	testEvent := NewAuditEvent("INFO", "TestOperation", "Test message")
	
	// Log the event - this should not panic
	logger.Log(testEvent)
	
	// If we reach here without panicking, the test passes
}

// TestLogAfterClose tests that the Log method handles logging after close
func TestLogAfterClose(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path
	logFilePath := filepath.Join(tempDir, "closed.log")
	
	// Create a FileLogger
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger: %v", err)
	}
	
	// Close the logger
	logger.Close()
	
	// Try to log after closing - should not panic
	testEvent := NewAuditEvent("INFO", "TestOperation", "Test message")
	logger.Log(testEvent)
	
	// If we reach here without panicking, the test passes
}

// TestLogWithInvalidEvent tests that the Log method handles invalid events gracefully
func TestLogWithInvalidEvent(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path
	logFilePath := filepath.Join(tempDir, "invalid.log")
	
	// Create a FileLogger
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger: %v", err)
	}
	defer logger.Close()
	
	// Create a test event with a value that cannot be marshaled to JSON
	// For this test, we'll create a circular reference which will cause
	// the JSON marshaler to fail
	invalidEvent := NewAuditEvent("INFO", "InvalidEvent", "This event has an invalid field")
	
	// Create a circular reference
	circular := make(map[string]interface{})
	circular["self"] = circular
	invalidEvent.Inputs = circular
	
	// Log the event - this should not panic despite marshaling failing
	logger.Log(invalidEvent)
	
	// If we reach here without panicking, the test passes
}

// TestFileLoggerClose tests the Close method
func TestFileLoggerClose(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filelogger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up
	
	// Test file path
	logFilePath := filepath.Join(tempDir, "audit.log")
	
	// Create a FileLogger
	logger, err := NewFileLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to create FileLogger: %v", err)
	}
	
	// Close the logger
	err = logger.Close()
	if err != nil {
		t.Errorf("Failed to close FileLogger: %v", err)
	}
	
	// Closing again should not cause an error (idempotent)
	err = logger.Close()
	if err != nil {
		t.Errorf("Second close caused an error: %v", err)
	}
}

// TestNoopLoggerCreation tests the creation of a NoopLogger
func TestNoopLoggerCreation(t *testing.T) {
	// Create a NoopLogger
	logger := NewNoopLogger()
	
	// Check that the logger is not nil
	if logger == nil {
		t.Fatal("Expected NoopLogger to be non-nil")
	}
	
	// Verify the logger implements StructuredLogger
	var _ StructuredLogger = logger
}

// TestNoopLoggerLog tests that the Log method doesn't cause errors
func TestNoopLoggerLog(t *testing.T) {
	// Create a NoopLogger
	logger := NewNoopLogger()
	
	// Create a test event
	testEvent := AuditEvent{
		Timestamp: time.Now().UTC(),
		Level:     "INFO",
		Operation: "TestOperation",
		Message:   "Test message",
	}
	
	// Log the event - this should do nothing but not cause errors
	logger.Log(testEvent)
	
	// If we reached here without panics, the test passes
}

// TestNoopLoggerClose tests the Close method
func TestNoopLoggerClose(t *testing.T) {
	// Create a NoopLogger
	logger := NewNoopLogger()
	
	// Close the logger - should return nil
	err := logger.Close()
	if err != nil {
		t.Errorf("Expected nil error from NoopLogger.Close(), got %v", err)
	}
	
	// Closing again should also work fine
	err = logger.Close()
	if err != nil {
		t.Errorf("Second close caused an error: %v", err)
	}
}