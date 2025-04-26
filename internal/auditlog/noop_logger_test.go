package auditlog

import (
	"fmt"
	"testing"
)

// TestNoOpAuditLogger_Log tests the Log method of NoOpAuditLogger
func TestNoOpAuditLogger_Log(t *testing.T) {
	// Create a NoOpAuditLogger
	logger := NewNoOpAuditLogger()

	// Test Log with various entries
	testCases := []struct {
		name  string
		entry AuditEntry
	}{
		{
			name: "Complete Entry",
			entry: AuditEntry{
				Operation: "TestOperation",
				Status:    "Success",
				Message:   "Test message",
				Inputs:    map[string]interface{}{"param": "value"},
				Outputs:   map[string]interface{}{"result": "success"},
			},
		},
		{
			name: "Entry With Error",
			entry: AuditEntry{
				Operation: "ErrorOperation",
				Status:    "Failure",
				Message:   "Error message",
				Error: &ErrorInfo{
					Message: "Something went wrong",
					Type:    "TestError",
				},
			},
		},
		{
			name: "Minimal Entry",
			entry: AuditEntry{
				Operation: "MinimalOp",
				Status:    "Minimal",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NoOpAuditLogger should never return an error
			err := logger.Log(tc.entry)
			if err != nil {
				t.Errorf("NoOpAuditLogger.Log returned error: %v", err)
			}
		})
	}

	// Verify interface implementation
	var _ AuditLogger = (*NoOpAuditLogger)(nil)
}

// TestNoOpAuditLogger_Close tests the Close method of NoOpAuditLogger
func TestNoOpAuditLogger_Close(t *testing.T) {
	// Create a NoOpAuditLogger
	logger := NewNoOpAuditLogger()

	// Close should never return an error
	err := logger.Close()
	if err != nil {
		t.Errorf("NoOpAuditLogger.Close returned error: %v", err)
	}

	// Should be able to log after close
	err = logger.Log(AuditEntry{
		Operation: "AfterClose",
		Status:    "Test",
	})
	if err != nil {
		t.Errorf("NoOpAuditLogger.Log after close returned error: %v", err)
	}

	// Multiple closes should be fine
	err = logger.Close()
	if err != nil {
		t.Errorf("NoOpAuditLogger.Close (second time) returned error: %v", err)
	}

	// Test behavior consistency
	// After closing, logging and other operations should still work
	noopTests := []func() error{
		func() error {
			return logger.Log(AuditEntry{Operation: "Test1", Status: "Test"})
		},
		func() error {
			return logger.LogOp("Test2", "Test", nil, nil, nil)
		},
		func() error {
			return logger.LogOp("Test3", "Test", nil, nil, fmt.Errorf("test error"))
		},
		func() error {
			return logger.Close()
		},
	}

	for i, test := range noopTests {
		if err := test(); err != nil {
			t.Errorf("NoOpAuditLogger test %d returned error: %v", i, err)
		}
	}
}
