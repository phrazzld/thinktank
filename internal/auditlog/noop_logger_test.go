package auditlog

import (
	"context"
	"fmt"
	"testing"
)

// TestNoOpAuditLogger_Log tests the Log method of NoOpAuditLogger
func TestNoOpAuditLogger_Log(t *testing.T) {
	t.Parallel(
	// Create a NoOpAuditLogger
	)

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
			// Use a background context for the test
			ctx := context.Background()
			err := logger.Log(ctx, tc.entry)
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
	t.Parallel(
	// Create a NoOpAuditLogger
	)

	logger := NewNoOpAuditLogger()

	// Close should never return an error
	err := logger.Close()
	if err != nil {
		t.Errorf("NoOpAuditLogger.Close returned error: %v", err)
	}

	// Should be able to log after close
	ctx := context.Background()
	err = logger.Log(ctx, AuditEntry{
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
	ctxForTests := context.Background()
	noopTests := []func() error{
		func() error {
			return logger.Log(ctxForTests, AuditEntry{Operation: "Test1", Status: "Test"})
		},
		func() error {
			return logger.LogOp(ctxForTests, "Test2", "Test", nil, nil, nil)
		},
		func() error {
			return logger.LogOp(ctxForTests, "Test3", "Test", nil, nil, fmt.Errorf("test error"))
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

// TestNoOpAuditLogger_LogLegacy tests the LogLegacy method of NoOpAuditLogger
func TestNoOpAuditLogger_LogLegacy(t *testing.T) {
	t.Parallel()

	logger := NewNoOpAuditLogger()

	// Test LogLegacy with various entries
	testCases := []struct {
		name  string
		entry AuditEntry
	}{
		{
			name: "Complete Legacy Entry",
			entry: AuditEntry{
				Operation: "LegacyOperation",
				Status:    "Success",
				Message:   "Legacy test message",
				Inputs:    map[string]interface{}{"legacy_param": "legacy_value"},
				Outputs:   map[string]interface{}{"legacy_result": "legacy_success"},
			},
		},
		{
			name: "Legacy Entry With Error",
			entry: AuditEntry{
				Operation: "LegacyErrorOperation",
				Status:    "Failure",
				Message:   "Legacy error message",
				Error: &ErrorInfo{
					Message: "Legacy error occurred",
					Type:    "LegacyError",
				},
			},
		},
		{
			name: "Minimal Legacy Entry",
			entry: AuditEntry{
				Operation: "MinimalLegacyOp",
				Status:    "Minimal",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NoOpAuditLogger should never return an error
			err := logger.LogLegacy(tc.entry)
			if err != nil {
				t.Errorf("NoOpAuditLogger.LogLegacy returned error: %v", err)
			}
		})
	}
}

// TestNoOpAuditLogger_LogOpLegacy tests the LogOpLegacy method of NoOpAuditLogger
func TestNoOpAuditLogger_LogOpLegacy(t *testing.T) {
	t.Parallel()

	logger := NewNoOpAuditLogger()

	// Test LogOpLegacy with various parameters
	testCases := []struct {
		name      string
		operation string
		status    string
		inputs    map[string]interface{}
		outputs   map[string]interface{}
		err       error
	}{
		{
			name:      "Complete Legacy Operation",
			operation: "LegacyTestOp",
			status:    "Success",
			inputs:    map[string]interface{}{"input_param": "input_value"},
			outputs:   map[string]interface{}{"output_result": "output_success"},
			err:       nil,
		},
		{
			name:      "Legacy Operation With Error",
			operation: "LegacyErrorOp",
			status:    "Failure",
			inputs:    map[string]interface{}{"error_param": "error_value"},
			outputs:   map[string]interface{}{"error_result": "error_failure"},
			err:       fmt.Errorf("legacy operation failed"),
		},
		{
			name:      "Minimal Legacy Operation",
			operation: "MinimalLegacyOp",
			status:    "Minimal",
			inputs:    nil,
			outputs:   nil,
			err:       nil,
		},
		{
			name:      "Legacy Operation With Empty Maps",
			operation: "EmptyMapsOp",
			status:    "Empty",
			inputs:    map[string]interface{}{},
			outputs:   map[string]interface{}{},
			err:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NoOpAuditLogger should never return an error
			err := logger.LogOpLegacy(tc.operation, tc.status, tc.inputs, tc.outputs, tc.err)
			if err != nil {
				t.Errorf("NoOpAuditLogger.LogOpLegacy returned error: %v", err)
			}
		})
	}
}
