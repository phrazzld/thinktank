package auditlog

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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