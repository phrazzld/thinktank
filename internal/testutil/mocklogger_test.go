package testutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestMockLoggerImplementsInterfaces tests that MockLogger implements the required interfaces
func TestMockLoggerImplementsInterfaces(t *testing.T) {
	// Create a new instance of MockLogger
	mockLogger := NewMockLogger()

	// Verify that it implements logutil.LoggerInterface
	var _ logutil.LoggerInterface = mockLogger

	// Verify that it implements auditlog.AuditLogger
	var _ auditlog.AuditLogger = mockLogger
}

// TestMockLoggerLogging tests the basic logging functionality
func TestMockLoggerLogging(t *testing.T) {
	mockLogger := NewMockLogger()

	// Test basic logging methods
	mockLogger.Debug("Debug message")
	mockLogger.Info("Info message")
	mockLogger.Warn("Warning message")
	mockLogger.Error("Error message")
	mockLogger.Fatal("Fatal message")

	// Verify messages were recorded
	if len(mockLogger.GetDebugMessages()) != 1 || mockLogger.GetDebugMessages()[0] != "Debug message" {
		t.Error("Debug message not recorded correctly")
	}
	if len(mockLogger.GetInfoMessages()) != 1 || mockLogger.GetInfoMessages()[0] != "Info message" {
		t.Error("Info message not recorded correctly")
	}
	if len(mockLogger.GetWarnMessages()) != 1 || mockLogger.GetWarnMessages()[0] != "Warning message" {
		t.Error("Warning message not recorded correctly")
	}
	if len(mockLogger.GetErrorMessages()) != 1 || mockLogger.GetErrorMessages()[0] != "Error message" {
		t.Error("Error message not recorded correctly")
	}
	if len(mockLogger.GetFatalMessages()) != 1 || mockLogger.GetFatalMessages()[0] != "Fatal message" {
		t.Error("Fatal message not recorded correctly")
	}

	// Test context-aware logging methods
	ctx := logutil.WithCustomCorrelationID(context.Background(), "test-correlation-id")
	mockLogger.ClearMessages()

	mockLogger.DebugContext(ctx, "Debug with context")
	mockLogger.InfoContext(ctx, "Info with context")
	mockLogger.WarnContext(ctx, "Warning with context")
	mockLogger.ErrorContext(ctx, "Error with context")
	mockLogger.FatalContext(ctx, "Fatal with context")

	// Verify messages were recorded (correlation ID will be stored separately in our implementation)
	if len(mockLogger.GetDebugMessages()) != 1 ||
		!contains(mockLogger.GetDebugMessages()[0], "Debug with context") {
		t.Error("Context-aware debug message not recorded correctly")
	}
}

// TestMockLoggerAuditLogging tests the audit logging functionality
func TestMockLoggerAuditLogging(t *testing.T) {
	mockLogger := NewMockLogger()

	// Test LogOp
	err := mockLogger.LogOp(
		"TestOperation",
		"Success",
		map[string]interface{}{"input1": "value1"},
		map[string]interface{}{"output1": "result1"},
		nil,
	)

	// Verify no error was returned
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify LogOp call was recorded
	logOpCalls := mockLogger.GetLogOpCalls()
	if len(logOpCalls) != 1 {
		t.Fatalf("Expected 1 LogOp call, got %d", len(logOpCalls))
	}
	if logOpCalls[0].Operation != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got '%s'", logOpCalls[0].Operation)
	}
	if logOpCalls[0].Status != "Success" {
		t.Errorf("Expected status 'Success', got '%s'", logOpCalls[0].Status)
	}

	// Verify audit entry was created
	entries := mockLogger.GetAuditEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Operation != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got '%s'", entries[0].Operation)
	}
	if entries[0].Status != "Success" {
		t.Errorf("Expected status 'Success', got '%s'", entries[0].Status)
	}

	// Test error simulation
	testErr := errors.New("test audit log error")
	mockLogger.SetLogError(testErr)

	err = mockLogger.LogOp("ErrorOperation", "Failure", nil, nil, nil)
	if err != testErr {
		t.Errorf("Expected error '%v', got '%v'", testErr, err)
	}

	// Clear error and verify error is gone
	mockLogger.ClearLogError()
	err = mockLogger.LogOp("AnotherOperation", "Success", nil, nil, nil)
	if err != nil {
		t.Errorf("Expected no error after clearing, got: %v", err)
	}

	// Test Log method
	entry := auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "DirectEntry",
		Status:    "InProgress",
		Message:   "Test direct entry",
	}

	err = mockLogger.Log(entry)
	if err != nil {
		t.Errorf("Expected no error from Log, got: %v", err)
	}

	// Verify entry was recorded
	entries = mockLogger.GetAuditEntries()
	directEntryFound := false
	for _, e := range entries {
		if e.Operation == "DirectEntry" && e.Status == "InProgress" {
			directEntryFound = true
			break
		}
	}
	if !directEntryFound {
		t.Error("Direct audit entry was not recorded correctly")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
