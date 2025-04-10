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
