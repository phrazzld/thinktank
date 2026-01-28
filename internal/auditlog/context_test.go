// Package auditlog provides structured logging for audit purposes
package auditlog

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/misty-step/thinktank/internal/logutil"
)

// TestAuditLogger_Context tests the context-aware methods of AuditLogger
func TestAuditLogger_Context(t *testing.T) {
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

	// Create a context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background(), "test-correlation-id")

	// Test logging with complete audit entry
	entry := AuditEntry{
		Operation: "ContextOperation",
		Status:    "Success",
		Message:   "Context test message",
		Inputs: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
		Outputs: map[string]interface{}{
			"result": "success",
			"code":   200,
		},
	}

	// Log the entry with context
	err = logger.Log(ctx, entry)
	if err != nil {
		t.Fatalf("Failed to log audit entry with context: %v", err)
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
	if parsedEntry.Operation != "ContextOperation" {
		t.Errorf("Expected Operation ContextOperation, got %s", parsedEntry.Operation)
	}
	if parsedEntry.Status != "Success" {
		t.Errorf("Expected Status Success, got %s", parsedEntry.Status)
	}
	if parsedEntry.Message != "Context test message" {
		t.Errorf("Expected Message 'Context test message', got %s", parsedEntry.Message)
	}
	if parsedEntry.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}

	// Verify input parameters
	if val, ok := parsedEntry.Inputs["param1"]; !ok || val != "value1" {
		t.Errorf("Expected Inputs to contain param1=value1, got %v", parsedEntry.Inputs)
	}
	if val, ok := parsedEntry.Inputs["param2"]; !ok || val != float64(42) { // JSON unmarshals to float64
		t.Errorf("Expected Inputs to contain param2=42, got %v", parsedEntry.Inputs)
	}

	// Verify correlation ID was added from context
	if val, ok := parsedEntry.Inputs["correlation_id"]; !ok || val != "test-correlation-id" {
		t.Errorf("Expected Inputs to contain correlation_id=test-correlation-id, got %v", parsedEntry.Inputs)
	}
}

// TestLogOp_Context tests the context-aware LogOp method
func TestLogOp_Context(t *testing.T) {
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

	// Create a context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background(), "logop-correlation-id")

	// Test LogOp with context
	inputs := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}
	outputs := map[string]interface{}{
		"result": "success",
		"code":   200,
	}

	// Call LogOp with context
	err = logger.LogOp(ctx, "ContextLogOp", "Success", inputs, outputs, nil)
	if err != nil {
		t.Fatalf("Failed to log operation with context: %v", err)
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
	if parsedEntry.Operation != "ContextLogOp" {
		t.Errorf("Expected Operation ContextLogOp, got %s", parsedEntry.Operation)
	}
	if parsedEntry.Status != "Success" {
		t.Errorf("Expected Status Success, got %s", parsedEntry.Status)
	}
	if parsedEntry.Message != "ContextLogOp completed successfully" {
		t.Errorf("Expected Message 'ContextLogOp completed successfully', got %s", parsedEntry.Message)
	}

	// Verify correlation ID was added from context
	if val, ok := parsedEntry.Inputs["correlation_id"]; !ok || val != "logop-correlation-id" {
		t.Errorf("Expected Inputs to contain correlation_id=logop-correlation-id, got %v", parsedEntry.Inputs)
	}
}

// TestNoOpAuditLogger_Context tests the context-aware methods of NoOpAuditLogger
func TestNoOpAuditLogger_Context(t *testing.T) {
	t.Parallel(
	// Create a NoOpAuditLogger
	)

	logger := NewNoOpAuditLogger()

	// Create a context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background(), "noop-correlation-id")

	// Test context-aware methods
	entry := AuditEntry{
		Operation: "NoOpOperation",
		Status:    "Success",
	}

	// Verify methods don't return errors
	if err := logger.Log(ctx, entry); err != nil {
		t.Errorf("NoOpAuditLogger.Log with context returned error: %v", err)
	}

	if err := logger.LogOp(ctx, "NoOpOp", "Success", nil, nil, nil); err != nil {
		t.Errorf("NoOpAuditLogger.LogOp with context returned error: %v", err)
	}
}

// TestLogLegacy verifies that the legacy methods work correctly
func TestLogLegacy(t *testing.T) {
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

	// Test logging with the legacy method
	entry := AuditEntry{
		Operation: "LegacyOp",
		Status:    "Legacy",
		Message:   "Legacy test message",
		Inputs: map[string]interface{}{
			"legacy": true,
		},
	}

	// Log the entry using the legacy method
	err = logger.LogLegacy(entry)
	if err != nil {
		t.Fatalf("Failed to log audit entry with legacy method: %v", err)
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
	if parsedEntry.Operation != "LegacyOp" {
		t.Errorf("Expected Operation LegacyOp, got %s", parsedEntry.Operation)
	}
	if parsedEntry.Status != "Legacy" {
		t.Errorf("Expected Status Legacy, got %s", parsedEntry.Status)
	}

	// Verify the inputs were preserved
	if val, ok := parsedEntry.Inputs["legacy"]; !ok || val != true {
		t.Errorf("Expected Inputs to contain legacy=true, got %v", parsedEntry.Inputs)
	}
}

// TestLogOpLegacy verifies that the legacy LogOp method works correctly
func TestLogOpLegacy(t *testing.T) {
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

	// Test logging with the legacy LogOp method
	inputs := map[string]interface{}{
		"legacy": true,
	}
	outputs := map[string]interface{}{
		"result": "legacy success",
	}

	// Call the legacy LogOp method
	err = logger.LogOpLegacy("LegacyLogOp", "Success", inputs, outputs, nil)
	if err != nil {
		t.Fatalf("Failed to log operation with legacy method: %v", err)
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
	if parsedEntry.Operation != "LegacyLogOp" {
		t.Errorf("Expected Operation LegacyLogOp, got %s", parsedEntry.Operation)
	}
	if parsedEntry.Status != "Success" {
		t.Errorf("Expected Status Success, got %s", parsedEntry.Status)
	}
	if parsedEntry.Message != "LegacyLogOp completed successfully" {
		t.Errorf("Expected Message 'LegacyLogOp completed successfully', got %s", parsedEntry.Message)
	}

	// Verify the inputs and outputs were preserved
	if val, ok := parsedEntry.Inputs["legacy"]; !ok || val != true {
		t.Errorf("Expected Inputs to contain legacy=true, got %v", parsedEntry.Inputs)
	}
	if val, ok := parsedEntry.Outputs["result"]; !ok || val != "legacy success" {
		t.Errorf("Expected Outputs to contain result=legacy success, got %v", parsedEntry.Outputs)
	}
}
