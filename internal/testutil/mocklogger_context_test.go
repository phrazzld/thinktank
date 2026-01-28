package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/misty-step/thinktank/internal/auditlog"
	"github.com/misty-step/thinktank/internal/logutil"
)

// TestMockLoggerWithContext tests context handling in MockLogger
func TestMockLoggerWithContext(t *testing.T) {
	// Create a mock logger
	logger := NewMockLogger()

	// Create a context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background(), "test-correlation-id")

	// Log messages with context
	logger.DebugContext(ctx, "Debug message with context")
	logger.InfoContext(ctx, "Info message with context")
	logger.WarnContext(ctx, "Warn message with context")
	logger.ErrorContext(ctx, "Error message with context")

	// Get log entries by correlation ID
	entries := logger.GetLogEntriesByCorrelationID("test-correlation-id")

	// Verify entries were logged with correlation ID
	if len(entries) != 4 {
		t.Errorf("Expected 4 log entries with correlation ID, got %d", len(entries))
	}

	// Verify each entry has the correct correlation ID
	for i, entry := range entries {
		if entry.CorrelationID != "test-correlation-id" {
			t.Errorf("Entry %d has incorrect correlation ID: %s", i, entry.CorrelationID)
		}
	}

	// Verify all correlation IDs are tracked
	correlationIDs := logger.GetAllCorrelationIDs()
	if len(correlationIDs) != 1 {
		t.Errorf("Expected 1 correlation ID, got %d", len(correlationIDs))
	}
	if correlationIDs[0] != "test-correlation-id" {
		t.Errorf("Expected correlation ID 'test-correlation-id', got '%s'", correlationIDs[0])
	}
}

// TestMockLoggerBackgroundContext tests handling of background contexts
func TestMockLoggerBackgroundContext(t *testing.T) {
	// Create a mock logger
	logger := NewMockLogger()

	// Log with background context
	logger.DebugContext(context.TODO(), "Debug message with background context")
	logger.InfoContext(context.TODO(), "Info message with background context")
	logger.WarnContext(context.TODO(), "Warn message with background context")
	logger.ErrorContext(context.TODO(), "Error message with background context")

	// Get all messages
	messages := logger.GetMessages()

	// Verify messages were logged with background context
	if len(messages) != 4 {
		t.Errorf("Expected 4 log messages with background context, got %d", len(messages))
	}
}

// TestMockLoggerMultipleCorrelationIDs tests tracking of multiple correlation IDs
func TestMockLoggerMultipleCorrelationIDs(t *testing.T) {
	// Create a mock logger
	logger := NewMockLogger()

	// Create contexts with different correlation IDs
	ctx1 := logutil.WithCorrelationID(context.Background(), "correlation-id-1")
	ctx2 := logutil.WithCorrelationID(context.Background(), "correlation-id-2")
	ctx3 := logutil.WithCorrelationID(context.Background(), "correlation-id-3")

	// Log messages with different contexts
	logger.InfoContext(ctx1, "Message with correlation ID 1")
	logger.InfoContext(ctx2, "Message with correlation ID 2")
	logger.InfoContext(ctx3, "Message with correlation ID 3")
	logger.InfoContext(ctx1, "Another message with correlation ID 1")

	// Get all correlation IDs
	correlationIDs := logger.GetAllCorrelationIDs()

	// Verify all correlation IDs are tracked
	if len(correlationIDs) != 3 {
		t.Errorf("Expected 3 correlation IDs, got %d", len(correlationIDs))
	}

	// Check entries for specific correlation ID
	entries1 := logger.GetLogEntriesByCorrelationID("correlation-id-1")
	if len(entries1) != 2 {
		t.Errorf("Expected 2 entries with correlation ID 1, got %d", len(entries1))
	}

	entries2 := logger.GetLogEntriesByCorrelationID("correlation-id-2")
	if len(entries2) != 1 {
		t.Errorf("Expected 1 entry with correlation ID 2, got %d", len(entries2))
	}
}

// TestMockLoggerAuditLoggerWithContext tests AuditLogger interface with context
func TestMockLoggerAuditLoggerWithContext(t *testing.T) {
	// Create a mock logger
	logger := NewMockLogger()

	// Create a context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background(), "audit-correlation-id")

	// Log an audit entry
	entry := auditlog.AuditEntry{
		Operation: "TestOperation",
		Status:    "Success",
		Message:   "Test audit message",
	}
	err := logger.Log(ctx, entry)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Log an operation
	err = logger.LogOp(ctx, "TestOperation", "InProgress", map[string]interface{}{
		"param1": "value1",
	}, nil, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Get audit entries by correlation ID
	auditEntries := logger.GetAuditEntriesByCorrelationID("audit-correlation-id")
	if len(auditEntries) != 2 {
		t.Errorf("Expected 2 audit entries with correlation ID, got %d", len(auditEntries))
	}

	// Verify correlation ID is in inputs
	for i, auditEntry := range auditEntries {
		if auditEntry.Inputs == nil {
			t.Errorf("Audit entry %d has nil inputs", i)
			continue
		}
		if id, ok := auditEntry.Inputs["correlation_id"]; !ok || id != "audit-correlation-id" {
			t.Errorf("Audit entry %d missing or incorrect correlation ID", i)
		}
	}

	// Get LogOp calls by correlation ID
	logOpCalls := logger.GetLogOpCallsByCorrelationID("audit-correlation-id")
	if len(logOpCalls) != 1 {
		t.Errorf("Expected 1 LogOp call with correlation ID, got %d", len(logOpCalls))
	}

	// Verify LogOp call has correlation ID
	if logOpCalls[0].CorrelationID != "audit-correlation-id" {
		t.Errorf("LogOp call has incorrect correlation ID: %s", logOpCalls[0].CorrelationID)
	}
}

// TestMockLoggerAuditLoggerWithError tests error handling in AuditLogger interface
func TestMockLoggerAuditLoggerWithError(t *testing.T) {
	// Create a mock logger
	logger := NewMockLogger()

	// Set an error to be returned by audit logging methods
	testError := fmt.Errorf("test audit error")
	logger.SetLogError(testError)

	// Create a context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background(), "error-correlation-id")

	// Log an audit entry
	entry := auditlog.AuditEntry{
		Operation: "ErrorOperation",
		Status:    "Failure",
		Message:   "Test error audit message",
	}
	err := logger.Log(ctx, entry)
	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}

	// Log an operation
	err = logger.LogOp(ctx, "ErrorOperation", "Failure", nil, nil,
		fmt.Errorf("operation error"))
	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}

	// Verify that the calls were tracked despite errors
	logOpCalls := logger.GetLogOpCalls()
	if len(logOpCalls) != 1 {
		t.Errorf("Expected 1 LogOp call, got %d", len(logOpCalls))
	}

	// Verify correlation ID is still tracked
	if logOpCalls[0].CorrelationID != "error-correlation-id" {
		t.Errorf("LogOp call has incorrect correlation ID: %s", logOpCalls[0].CorrelationID)
	}

	// Clear the error
	logger.ClearLogError()

	// Try logging again (should succeed now)
	err = logger.Log(ctx, entry)
	if err != nil {
		t.Errorf("Expected no error after clearing, got %v", err)
	}
}
