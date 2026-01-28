package auditlog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
)

// TestFileAuditLogger_LogOp tests the LogOp method of FileAuditLogger
func TestFileAuditLogger_LogOp(t *testing.T) {
	t.Parallel(
	// Setup a temporary file for testing
	)

	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	// Create a mock logger
	mockLog := newMockLogger()

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

	// Standard inputs and outputs for test cases
	inputs := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}
	outputs := map[string]interface{}{
		"result": "success",
		"code":   200,
	}

	// Define test cases
	testCases := []struct {
		name           string
		operation      string
		status         string
		inputs         map[string]interface{}
		outputs        map[string]interface{}
		err            error
		expectedMsg    string
		expectedErrMsg string
		expectedErrTyp string
	}{
		{
			name:        "Success Status",
			operation:   "TestOperation",
			status:      "Success",
			inputs:      inputs,
			outputs:     outputs,
			err:         nil,
			expectedMsg: "TestOperation completed successfully",
		},
		{
			name:        "InProgress Status",
			operation:   "StartOperation",
			status:      "InProgress",
			inputs:      inputs,
			outputs:     nil,
			err:         nil,
			expectedMsg: "StartOperation started",
		},
		{
			name:           "Failure Status",
			operation:      "FailOperation",
			status:         "Failure",
			inputs:         inputs,
			outputs:        nil,
			err:            fmt.Errorf("test error"),
			expectedMsg:    "FailOperation failed",
			expectedErrMsg: "test error",
			expectedErrTyp: "GeneralError",
		},
		{
			name:        "Custom Status",
			operation:   "CustomOperation",
			status:      "CustomStatus",
			inputs:      inputs,
			outputs:     nil,
			err:         nil,
			expectedMsg: "CustomOperation - CustomStatus",
		},
		{
			name:           "Categorized Error",
			operation:      "SafetyOperation",
			status:         "Failure",
			inputs:         inputs,
			outputs:        nil,
			err:            &mockCategorizedError{msg: "content safety error", category: llm.CategoryContentFiltered},
			expectedMsg:    "SafetyOperation failed",
			expectedErrMsg: "content safety error",
			expectedErrTyp: fmt.Sprintf("Error:%s", llm.CategoryContentFiltered.String()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear the file for this test case
			if err := logger.file.Truncate(0); err != nil {
				t.Fatalf("Failed to truncate log file: %v", err)
			}
			if _, err := logger.file.Seek(0, 0); err != nil {
				t.Fatalf("Failed to seek in log file: %v", err)
			}

			// Execute LogOp with the test case parameters
			// Use a background context for the test
			ctx := context.Background()
			err = logger.LogOp(ctx, tc.operation, tc.status, tc.inputs, tc.outputs, tc.err)
			if err != nil {
				t.Fatalf("Failed to log operation: %v", err)
			}

			// Read the log file
			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			// Parse the JSON line
			var parsedEntry AuditEntry
			if err := json.Unmarshal(content, &parsedEntry); err != nil {
				t.Fatalf("Failed to parse JSON: %v\nContent: %s", err, content)
			}

			// Verify the entry was logged correctly
			if parsedEntry.Operation != tc.operation {
				t.Errorf("Expected Operation %s, got %s", tc.operation, parsedEntry.Operation)
			}
			if parsedEntry.Status != tc.status {
				t.Errorf("Expected Status %s, got %s", tc.status, parsedEntry.Status)
			}
			if parsedEntry.Timestamp.IsZero() {
				t.Error("Expected Timestamp to be set")
			}
			if parsedEntry.Message != tc.expectedMsg {
				t.Errorf("Expected Message %q, got %q", tc.expectedMsg, parsedEntry.Message)
			}

			// Verify inputs (if provided)
			if tc.inputs != nil {
				for k, v := range tc.inputs {
					if k == "param2" {
						// JSON unmarshals numbers to float64
						if val, ok := parsedEntry.Inputs[k]; !ok || val != float64(v.(int)) {
							t.Errorf("Expected Inputs to contain %s=%v, got %v", k, v, parsedEntry.Inputs[k])
						}
					} else {
						if val, ok := parsedEntry.Inputs[k]; !ok || val != v {
							t.Errorf("Expected Inputs to contain %s=%v, got %v", k, v, parsedEntry.Inputs[k])
						}
					}
				}
			}

			// Verify outputs (if provided)
			if tc.outputs != nil {
				for k, v := range tc.outputs {
					if k == "code" {
						// JSON unmarshals numbers to float64
						if val, ok := parsedEntry.Outputs[k]; !ok || val != float64(v.(int)) {
							t.Errorf("Expected Outputs to contain %s=%v, got %v", k, v, parsedEntry.Outputs[k])
						}
					} else {
						if val, ok := parsedEntry.Outputs[k]; !ok || val != v {
							t.Errorf("Expected Outputs to contain %s=%v, got %v", k, v, parsedEntry.Outputs[k])
						}
					}
				}
			}

			// Verify error (if expected)
			if tc.expectedErrMsg != "" {
				if parsedEntry.Error == nil {
					t.Fatal("Expected Error to be set")
				}
				if parsedEntry.Error.Message != tc.expectedErrMsg {
					t.Errorf("Expected Error.Message %q, got %q", tc.expectedErrMsg, parsedEntry.Error.Message)
				}
				if parsedEntry.Error.Type != tc.expectedErrTyp {
					t.Errorf("Expected Error.Type %q, got %q", tc.expectedErrTyp, parsedEntry.Error.Type)
				}
			} else if parsedEntry.Error != nil {
				t.Errorf("Expected no error, but got %v", parsedEntry.Error)
			}
		})
	}
}

// mockCategorizedError implements llm.CategorizedError for testing
type mockCategorizedError struct {
	msg      string
	category llm.ErrorCategory
}

func (e *mockCategorizedError) Error() string {
	return e.msg
}

func (e *mockCategorizedError) Category() llm.ErrorCategory {
	return e.category
}

// TestNoOpAuditLogger_LogOp tests the LogOp method of NoOpAuditLogger
func TestNoOpAuditLogger_LogOp(t *testing.T) {
	t.Parallel(
	// Create a NoOpAuditLogger
	)

	logger := NewNoOpAuditLogger()

	// Test LogOp with various parameters
	testCases := []struct {
		name      string
		operation string
		status    string
		inputs    map[string]interface{}
		outputs   map[string]interface{}
		err       error
	}{
		{
			name:      "Success",
			operation: "TestOperation",
			status:    "Success",
			inputs:    map[string]interface{}{"param": "value"},
			outputs:   map[string]interface{}{"result": "success"},
			err:       nil,
		},
		{
			name:      "Failure",
			operation: "FailOperation",
			status:    "Failure",
			inputs:    map[string]interface{}{"param": "value"},
			outputs:   nil,
			err:       fmt.Errorf("test error"),
		},
		{
			name:      "InProgress",
			operation: "StartOperation",
			status:    "InProgress",
			inputs:    nil,
			outputs:   nil,
			err:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NoOpAuditLogger should never return an error
			// Use a background context for the test
			ctx := context.Background()
			err := logger.LogOp(ctx, tc.operation, tc.status, tc.inputs, tc.outputs, tc.err)
			if err != nil {
				t.Errorf("NoOpAuditLogger.LogOp returned error: %v", err)
			}
		})
	}
}
