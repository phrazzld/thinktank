package auditlog

import (
	"encoding/json"
	"testing"
	"time"
)

func TestErrorDetailsJSON(t *testing.T) {
	// Create a sample ErrorDetails
	details := ErrorDetails{
		Message: "Something went wrong",
		Type:    "APIError",
		Details: "Additional error context",
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(details)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorDetails to JSON: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON back to map: %v", err)
	}

	// Check fields
	if result["message"] != "Something went wrong" {
		t.Errorf("Expected message 'Something went wrong', got %v", result["message"])
	}
	if result["type"] != "APIError" {
		t.Errorf("Expected type 'APIError', got %v", result["type"])
	}
	if result["details"] != "Additional error context" {
		t.Errorf("Expected details 'Additional error context', got %v", result["details"])
	}

	// Test omitempty
	minimalDetails := ErrorDetails{
		Message: "Just a message",
	}
	jsonBytes, err = json.Marshal(minimalDetails)
	if err != nil {
		t.Fatalf("Failed to marshal minimal ErrorDetails to JSON: %v", err)
	}

	result = make(map[string]interface{})
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON back to map: %v", err)
	}

	if _, exists := result["type"]; exists {
		t.Error("Expected 'type' to be omitted when empty")
	}
	if _, exists := result["details"]; exists {
		t.Error("Expected 'details' to be omitted when empty")
	}
}

func TestAuditEventJSON(t *testing.T) {
	// Create a fixed timestamp for testing
	fixedTime := time.Date(2025, 4, 9, 15, 30, 0, 0, time.UTC)

	// Create a sample AuditEvent
	event := AuditEvent{
		Timestamp: fixedTime,
		Level:     "INFO",
		Operation: "TestOperation",
		Message:   "Test event message",
		Inputs: map[string]interface{}{
			"param1": "value1",
			"param2": 123,
		},
		Outputs: map[string]interface{}{
			"result": "success",
			"count":  42,
		},
		Metadata: map[string]interface{}{
			"duration_ms": 150,
		},
		Error: &ErrorDetails{
			Message: "Some error",
			Type:    "TestError",
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal AuditEvent to JSON: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON back to map: %v", err)
	}

	// Check core fields
	expectedTimestamp := "2025-04-09T15:30:00Z"
	if result["timestamp"] != expectedTimestamp {
		t.Errorf("Expected timestamp %q, got %v", expectedTimestamp, result["timestamp"])
	}
	if result["level"] != "INFO" {
		t.Errorf("Expected level 'INFO', got %v", result["level"])
	}
	if result["operation"] != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got %v", result["operation"])
	}
	if result["message"] != "Test event message" {
		t.Errorf("Expected message 'Test event message', got %v", result["message"])
	}

	// Check inputs
	inputs, ok := result["inputs"].(map[string]interface{})
	if !ok {
		t.Error("Expected inputs to be a map")
	} else {
		if inputs["param1"] != "value1" {
			t.Errorf("Expected inputs.param1 to be 'value1', got %v", inputs["param1"])
		}
		if int(inputs["param2"].(float64)) != 123 {
			t.Errorf("Expected inputs.param2 to be 123, got %v", inputs["param2"])
		}
	}

	// Test omitempty
	minimalEvent := AuditEvent{
		Timestamp: fixedTime,
		Level:     "ERROR",
		Operation: "MinimalOp",
		Message:   "Minimal message",
	}
	jsonBytes, err = json.Marshal(minimalEvent)
	if err != nil {
		t.Fatalf("Failed to marshal minimal AuditEvent to JSON: %v", err)
	}

	result = make(map[string]interface{})
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON back to map: %v", err)
	}

	// Check that optional fields are omitted
	optionalFields := []string{"inputs", "outputs", "metadata", "error"}
	for _, field := range optionalFields {
		if _, exists := result[field]; exists {
			t.Errorf("Expected '%s' to be omitted when empty", field)
		}
	}
}

func TestNewAuditEvent(t *testing.T) {
	// Test creation with basic parameters
	event := NewAuditEvent("INFO", "TestOp", "Test message")

	// Check required fields
	if event.Level != "INFO" {
		t.Errorf("Expected level 'INFO', got %v", event.Level)
	}
	if event.Operation != "TestOp" {
		t.Errorf("Expected operation 'TestOp', got %v", event.Operation)
	}
	if event.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %v", event.Message)
	}

	// Check timestamp was set
	if event.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set, but it was zero")
	}

	// Check optional maps are nil
	if event.Inputs != nil {
		t.Error("Expected Inputs to be nil for new event")
	}
	if event.Outputs != nil {
		t.Error("Expected Outputs to be nil for new event")
	}
	if event.Metadata != nil {
		t.Error("Expected Metadata to be nil for new event")
	}
	if event.Error != nil {
		t.Error("Expected Error to be nil for new event")
	}
}

func TestNewErrorDetails(t *testing.T) {
	// Test with just the message
	error := NewErrorDetails("An error occurred")
	if error.Message != "An error occurred" {
		t.Errorf("Expected message 'An error occurred', got %v", error.Message)
	}
	if error.Type != "" {
		t.Errorf("Expected type to be empty, got %v", error.Type)
	}
	if error.Details != "" {
		t.Errorf("Expected details to be empty, got %v", error.Details)
	}

	// Test with all fields
	error = NewErrorDetails("An error occurred", "TestError", "More details")
	if error.Message != "An error occurred" {
		t.Errorf("Expected message 'An error occurred', got %v", error.Message)
	}
	if error.Type != "TestError" {
		t.Errorf("Expected type 'TestError', got %v", error.Type)
	}
	if error.Details != "More details" {
		t.Errorf("Expected details 'More details', got %v", error.Details)
	}
}

func TestAuditEventWithMethods(t *testing.T) {
	// Create event using builder pattern methods
	event := NewAuditEvent("INFO", "TestOp", "Test message").
		WithInput("param1", "value1").
		WithInput("param2", 123).
		WithOutput("result", "success").
		WithMetadata("duration_ms", 150).
		WithError(NewErrorDetails("Some error", "TestError", ""))

	// Check required fields
	if event.Level != "INFO" {
		t.Errorf("Expected level 'INFO', got %v", event.Level)
	}

	// Check inputs were set correctly
	if event.Inputs == nil || len(event.Inputs) != 2 {
		t.Fatalf("Expected Inputs to have 2 items, got %v", event.Inputs)
	}
	if event.Inputs["param1"] != "value1" {
		t.Errorf("Expected inputs.param1 to be 'value1', got %v", event.Inputs["param1"])
	}
	if event.Inputs["param2"] != 123 {
		t.Errorf("Expected inputs.param2 to be 123, got %v", event.Inputs["param2"])
	}

	// Check outputs were set correctly
	if event.Outputs == nil || len(event.Outputs) != 1 {
		t.Fatalf("Expected Outputs to have 1 item, got %v", event.Outputs)
	}
	if event.Outputs["result"] != "success" {
		t.Errorf("Expected outputs.result to be 'success', got %v", event.Outputs["result"])
	}

	// Check metadata was set correctly
	if event.Metadata == nil || len(event.Metadata) != 1 {
		t.Fatalf("Expected Metadata to have 1 item, got %v", event.Metadata)
	}
	if event.Metadata["duration_ms"] != 150 {
		t.Errorf("Expected metadata.duration_ms to be 150, got %v", event.Metadata["duration_ms"])
	}

	// Check error was set correctly
	if event.Error == nil {
		t.Fatal("Expected Error to be set, but it was nil")
	}
	if event.Error.Message != "Some error" {
		t.Errorf("Expected error.message to be 'Some error', got %v", event.Error.Message)
	}
	if event.Error.Type != "TestError" {
		t.Errorf("Expected error.type to be 'TestError', got %v", event.Error.Type)
	}
}

func TestWithErrorFromGoError(t *testing.T) {
	// Create a standard Go error
	err := &testError{msg: "Go error"}

	// Create event with the error
	event := NewAuditEvent("ERROR", "TestOp", "Test error").
		WithErrorFromGoError(err)

	// Check error was set correctly
	if event.Error == nil {
		t.Fatal("Expected Error to be set, but it was nil")
	}
	if event.Error.Message != "Go error" {
		t.Errorf("Expected error.message to be 'Go error', got %v", event.Error.Message)
	}
	if event.Error.Type != "testError" {
		t.Errorf("Expected error.type to be 'testError', got %v", event.Error.Type)
	}
}

// Helper for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
