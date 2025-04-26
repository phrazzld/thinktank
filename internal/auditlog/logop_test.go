package auditlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestFileAuditLogger_LogOp tests the LogOp method of FileAuditLogger
func TestFileAuditLogger_LogOp(t *testing.T) {
	// Setup a temporary file for testing
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

	// Test LogOp with success status
	inputs := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}
	outputs := map[string]interface{}{
		"result": "success",
		"code":   200,
	}

	// Log success operation
	err = logger.LogOp("TestOperation", "Success", inputs, outputs, nil)
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
	if parsedEntry.Operation != "TestOperation" {
		t.Errorf("Expected Operation TestOperation, got %s", parsedEntry.Operation)
	}
	if parsedEntry.Status != "Success" {
		t.Errorf("Expected Status Success, got %s", parsedEntry.Status)
	}
	if parsedEntry.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
	expectedMessage := "TestOperation completed successfully"
	if parsedEntry.Message != expectedMessage {
		t.Errorf("Expected Message %q, got %q", expectedMessage, parsedEntry.Message)
	}

	// Verify input parameters
	if val, ok := parsedEntry.Inputs["param1"]; !ok || val != "value1" {
		t.Errorf("Expected Inputs to contain param1=value1, got %v", parsedEntry.Inputs)
	}
	if val, ok := parsedEntry.Inputs["param2"]; !ok || val != float64(42) { // JSON unmarshals to float64
		t.Errorf("Expected Inputs to contain param2=42, got %v", parsedEntry.Inputs)
	}

	// Verify output parameters
	if val, ok := parsedEntry.Outputs["result"]; !ok || val != "success" {
		t.Errorf("Expected Outputs to contain result=success, got %v", parsedEntry.Outputs)
	}
	if val, ok := parsedEntry.Outputs["code"]; !ok || val != float64(200) { // JSON unmarshals to float64
		t.Errorf("Expected Outputs to contain code=200, got %v", parsedEntry.Outputs)
	}

	// Test LogOp with in-progress status
	if err := logger.file.Truncate(0); err != nil { // Clear the file for new test
		t.Fatalf("Failed to truncate log file: %v", err)
	}
	if _, err := logger.file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in log file: %v", err)
	}

	err = logger.LogOp("StartOperation", "InProgress", inputs, nil, nil)
	if err != nil {
		t.Fatalf("Failed to log in-progress operation: %v", err)
	}

	// Read the log file
	content, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse the JSON
	if err := json.Unmarshal(content, &parsedEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nContent: %s", err, content)
	}

	// Verify in-progress message
	expectedInProgressMessage := "StartOperation started"
	if parsedEntry.Message != expectedInProgressMessage {
		t.Errorf("Expected Message %q, got %q", expectedInProgressMessage, parsedEntry.Message)
	}

	// Test LogOp with failure status and error
	if err := logger.file.Truncate(0); err != nil { // Clear the file for new test
		t.Fatalf("Failed to truncate log file: %v", err)
	}
	if _, err := logger.file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in log file: %v", err)
	}

	testError := fmt.Errorf("test error")
	err = logger.LogOp("FailOperation", "Failure", inputs, nil, testError)
	if err != nil {
		t.Fatalf("Failed to log failure operation: %v", err)
	}

	// Read the log file
	content, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse the JSON
	if err := json.Unmarshal(content, &parsedEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nContent: %s", err, content)
	}

	// Verify failure message and error
	expectedFailureMessage := "FailOperation failed"
	if parsedEntry.Message != expectedFailureMessage {
		t.Errorf("Expected Message %q, got %q", expectedFailureMessage, parsedEntry.Message)
	}
	if parsedEntry.Error == nil {
		t.Fatal("Expected Error to be set for failure operation")
	}
	if parsedEntry.Error.Message != "test error" {
		t.Errorf("Expected Error.Message %q, got %q", "test error", parsedEntry.Error.Message)
	}
	if parsedEntry.Error.Type != "GeneralError" {
		t.Errorf("Expected Error.Type %q, got %q", "GeneralError", parsedEntry.Error.Type)
	}

	// Test LogOp with custom status
	if err := logger.file.Truncate(0); err != nil { // Clear the file for new test
		t.Fatalf("Failed to truncate log file: %v", err)
	}
	if _, err := logger.file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in log file: %v", err)
	}

	err = logger.LogOp("CustomOperation", "CustomStatus", inputs, nil, nil)
	if err != nil {
		t.Fatalf("Failed to log custom status operation: %v", err)
	}

	// Read the log file
	content, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse the JSON
	if err := json.Unmarshal(content, &parsedEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nContent: %s", err, content)
	}

	// Verify custom status message
	expectedCustomMessage := "CustomOperation - CustomStatus"
	if parsedEntry.Message != expectedCustomMessage {
		t.Errorf("Expected Message %q, got %q", expectedCustomMessage, parsedEntry.Message)
	}

	// Test with categorized error
	if err := logger.file.Truncate(0); err != nil { // Clear the file for new test
		t.Fatalf("Failed to truncate log file: %v", err)
	}
	if _, err := logger.file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek in log file: %v", err)
	}

	// Create a mock categorized error
	categorizedErr := &mockCategorizedError{
		msg:      "content safety error",
		category: llm.CategoryContentFiltered,
	}

	err = logger.LogOp("SafetyOperation", "Failure", inputs, nil, categorizedErr)
	if err != nil {
		t.Fatalf("Failed to log operation with categorized error: %v", err)
	}

	// Read the log file
	content, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse the JSON
	if err := json.Unmarshal(content, &parsedEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nContent: %s", err, content)
	}

	// Verify error type for categorized error
	if parsedEntry.Error == nil {
		t.Fatal("Expected Error to be set for operation with categorized error")
	}
	if parsedEntry.Error.Message != "content safety error" {
		t.Errorf("Expected Error.Message %q, got %q", "content safety error", parsedEntry.Error.Message)
	}
	expectedErrorType := fmt.Sprintf("Error:%s", llm.CategoryContentFiltered.String())
	if parsedEntry.Error.Type != expectedErrorType {
		t.Errorf("Expected Error.Type %q, got %q", expectedErrorType, parsedEntry.Error.Type)
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
	// Create a NoOpAuditLogger
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
			err := logger.LogOp(tc.operation, tc.status, tc.inputs, tc.outputs, tc.err)
			if err != nil {
				t.Errorf("NoOpAuditLogger.LogOp returned error: %v", err)
			}
		})
	}
}
