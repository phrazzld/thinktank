package auditlog

import (
	"errors"
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