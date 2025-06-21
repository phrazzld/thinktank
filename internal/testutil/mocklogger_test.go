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

	// Test LogOp with context
	ctx := context.Background()
	err := mockLogger.LogOp(
		ctx,
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

	err = mockLogger.LogOp(ctx, "ErrorOperation", "Failure", nil, nil, nil)
	if err != testErr {
		t.Errorf("Expected error '%v', got '%v'", testErr, err)
	}

	// Clear error and verify error is gone
	mockLogger.ClearLogError()
	err = mockLogger.LogOp(ctx, "AnotherOperation", "Success", nil, nil, nil)
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

	err = mockLogger.Log(ctx, entry)
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

// TestMockLoggerZeroCoverageFunctions tests the functions that currently have 0% coverage
func TestMockLoggerZeroCoverageFunctions(t *testing.T) {
	mockLogger := NewMockLogger()

	// Test Println function
	t.Run("Println", func(t *testing.T) {
		mockLogger.ClearMessages()
		mockLogger.Println("test println message")
		// Println should add message to info level
		infoMessages := mockLogger.GetInfoMessages()
		if len(infoMessages) == 0 {
			t.Fatal("Println should add message to info messages")
		}
		if !contains(infoMessages[len(infoMessages)-1], "test println message") {
			t.Fatalf("Expected info message to contain 'test println message', got '%s'", infoMessages[len(infoMessages)-1])
		}

		// Also should add to general messages
		allMessages := mockLogger.GetMessages()
		if len(allMessages) == 0 {
			t.Fatal("Println should add message to general messages")
		}
	})

	// Test Printf function
	t.Run("Printf", func(t *testing.T) {
		mockLogger.ClearMessages()
		mockLogger.Printf("test printf %s %d", "message", 42)
		// Printf should add to general messages (not to specific level messages)
		allMessages := mockLogger.GetMessages()
		if len(allMessages) == 0 {
			t.Fatal("Printf should add message to general messages")
		}
		expected := "test printf message 42"
		if !contains(allMessages[0], expected) {
			t.Fatalf("Expected message to contain '%s', got '%s'", expected, allMessages[0])
		}
	})

	// Test WithContext function
	t.Run("WithContext", func(t *testing.T) {
		ctx := context.Background()
		contextLogger := mockLogger.WithContext(ctx)

		// WithContext should return the same logger instance for MockLogger
		if contextLogger != mockLogger {
			t.Fatal("WithContext should return the same MockLogger instance")
		}
	})

	// Test SetLevel and GetLevel functions
	t.Run("SetLevel and GetLevel", func(t *testing.T) {
		// Test setting different log levels
		originalLevel := mockLogger.GetLevel()

		mockLogger.SetLevel(logutil.DebugLevel)
		if mockLogger.GetLevel() != logutil.DebugLevel {
			t.Fatalf("Expected level DebugLevel, got %v", mockLogger.GetLevel())
		}

		mockLogger.SetLevel(logutil.InfoLevel)
		if mockLogger.GetLevel() != logutil.InfoLevel {
			t.Fatalf("Expected level InfoLevel, got %v", mockLogger.GetLevel())
		}

		mockLogger.SetLevel(logutil.WarnLevel)
		if mockLogger.GetLevel() != logutil.WarnLevel {
			t.Fatalf("Expected level WarnLevel, got %v", mockLogger.GetLevel())
		}

		mockLogger.SetLevel(logutil.ErrorLevel)
		if mockLogger.GetLevel() != logutil.ErrorLevel {
			t.Fatalf("Expected level ErrorLevel, got %v", mockLogger.GetLevel())
		}

		// Restore original level
		mockLogger.SetLevel(originalLevel)
	})

	// Test SetVerbose function
	t.Run("SetVerbose", func(t *testing.T) {
		// Test setting verbose to true
		mockLogger.SetVerbose(true)

		// Test setting verbose to false
		mockLogger.SetVerbose(false)

		// SetVerbose doesn't return anything to verify, but we're testing coverage
	})

	// Test LogLegacy function
	t.Run("LogLegacy", func(t *testing.T) {
		entry := auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "LegacyOperation",
			Status:    "Success",
			Message:   "Test legacy entry",
		}

		err := mockLogger.LogLegacy(entry)
		if err != nil {
			t.Fatalf("LogLegacy should not return error, got: %v", err)
		}

		// Verify entry was recorded
		entries := mockLogger.GetAuditEntries()
		legacyEntryFound := false
		for _, e := range entries {
			if e.Operation == "LegacyOperation" && e.Status == "Success" {
				legacyEntryFound = true
				break
			}
		}
		if !legacyEntryFound {
			t.Error("Legacy audit entry was not recorded correctly")
		}
	})

	// Test LogOpLegacy function
	t.Run("LogOpLegacy", func(t *testing.T) {
		err := mockLogger.LogOpLegacy(
			"LegacyOpOperation",
			"InProgress",
			map[string]interface{}{"legacy_input": "value"},
			map[string]interface{}{"legacy_output": "result"},
			nil,
		)

		if err != nil {
			t.Fatalf("LogOpLegacy should not return error, got: %v", err)
		}

		// Verify LogOp call was recorded
		logOpCalls := mockLogger.GetLogOpCalls()
		legacyOpFound := false
		for _, call := range logOpCalls {
			if call.Operation == "LegacyOpOperation" && call.Status == "InProgress" {
				legacyOpFound = true
				break
			}
		}
		if !legacyOpFound {
			t.Error("Legacy LogOp call was not recorded correctly")
		}
	})

	// Test Close function
	t.Run("Close", func(t *testing.T) {
		err := mockLogger.Close()
		if err != nil {
			t.Fatalf("Close should not return error, got: %v", err)
		}
	})

	// Test ClearAuditRecords function
	t.Run("ClearAuditRecords", func(t *testing.T) {
		// Add some audit entries first
		_ = mockLogger.LogOp(context.Background(), "TestOp1", "Success", nil, nil, nil)
		_ = mockLogger.LogOp(context.Background(), "TestOp2", "Success", nil, nil, nil)

		// Verify entries exist
		if len(mockLogger.GetAuditEntries()) == 0 {
			t.Fatal("Expected audit entries before clearing")
		}
		if len(mockLogger.GetLogOpCalls()) == 0 {
			t.Fatal("Expected LogOp calls before clearing")
		}

		// Clear audit records
		mockLogger.ClearAuditRecords()

		// Verify entries are cleared
		if len(mockLogger.GetAuditEntries()) != 0 {
			t.Fatalf("Expected 0 audit entries after clearing, got %d", len(mockLogger.GetAuditEntries()))
		}
		if len(mockLogger.GetLogOpCalls()) != 0 {
			t.Fatalf("Expected 0 LogOp calls after clearing, got %d", len(mockLogger.GetLogOpCalls()))
		}
	})

	// Test ContainsMessage function
	t.Run("ContainsMessage", func(t *testing.T) {
		mockLogger.ClearMessages()

		// Add some test messages
		mockLogger.Debug("unique debug message")
		mockLogger.Info("unique info message")
		mockLogger.Warn("unique warn message")
		mockLogger.Error("unique error message")

		// Test existing messages
		if !mockLogger.ContainsMessage("unique debug message") {
			t.Fatal("ContainsMessage should return true for existing debug message")
		}
		if !mockLogger.ContainsMessage("unique info message") {
			t.Fatal("ContainsMessage should return true for existing info message")
		}
		if !mockLogger.ContainsMessage("unique warn message") {
			t.Fatal("ContainsMessage should return true for existing warn message")
		}
		if !mockLogger.ContainsMessage("unique error message") {
			t.Fatal("ContainsMessage should return true for existing error message")
		}

		// Test partial matching
		if !mockLogger.ContainsMessage("unique debug") {
			t.Fatal("ContainsMessage should return true for partial match")
		}

		// Test non-existing message
		if mockLogger.ContainsMessage("non-existing message") {
			t.Fatal("ContainsMessage should return false for non-existing message")
		}
	})

	// Test GetLogEntries function
	t.Run("GetLogEntries", func(t *testing.T) {
		mockLogger.ClearMessages()

		// Add some log entries
		ctx := logutil.WithCustomCorrelationID(context.Background(), "test-corr-id")
		mockLogger.DebugContext(ctx, "debug entry")
		mockLogger.InfoContext(ctx, "info entry")
		mockLogger.ErrorContext(ctx, "error entry")

		// Get log entries
		entries := mockLogger.GetLogEntries()

		// Should have entries for each logged message
		if len(entries) < 3 {
			t.Fatalf("Expected at least 3 log entries, got %d", len(entries))
		}

		// Verify entries contain expected information
		foundDebug := false
		foundInfo := false
		foundError := false

		for _, entry := range entries {
			if contains(entry.Message, "debug entry") && entry.Level == "debug" {
				foundDebug = true
			}
			if contains(entry.Message, "info entry") && entry.Level == "info" {
				foundInfo = true
			}
			if contains(entry.Message, "error entry") && entry.Level == "error" {
				foundError = true
			}
		}

		if !foundDebug {
			t.Fatal("Expected to find debug log entry")
		}
		if !foundInfo {
			t.Fatal("Expected to find info log entry")
		}
		if !foundError {
			t.Fatal("Expected to find error log entry")
		}
	})
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
