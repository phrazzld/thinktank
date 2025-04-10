package main

import (
	"os"
	"testing"
)

// TestTaskFileRequirementSimple confirms that the application requires the --task-file flag
func TestTaskFileRequirementSimple(t *testing.T) {
	// This test verifies that a Configuration with a non-empty task file passes validation,
	// while one without a task file will cause an error (in normal mode, not dry run)
	
	// Create a test task file
	tempFile, err := os.CreateTemp("", "task-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary task file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	_, err = tempFile.WriteString("Test task content")
	if err != nil {
		t.Fatalf("Failed to write to temporary task file: %v", err)
	}
	tempFile.Close()
	
	// Test with a valid task file
	config := &Configuration{
		TaskFile: tempFile.Name(),
		Paths:    []string{"testfile"},
		ApiKey:   "test-key",
	}
	
	// Create a custom logger that tracks error calls
	errLogger := &errorTrackingLogger{}
	
	// Run validation with a valid task file
	result := doValidateInputs(config, errLogger)
	
	// Validation should pass with a valid task file
	if !result.Valid {
		t.Errorf("Validation failed with a valid task file: %s", result.ErrorMessage)
	}
	
	// Reset for next test
	errLogger.reset()
	
	// Test with no task file and not in dry run mode
	configNoFile := &Configuration{
		TaskFile: "",
		DryRun:   false,
		Paths:    []string{"testfile"},
		ApiKey:   "test-key",
	}
	
	// Run validation with no task file
	result = doValidateInputs(configNoFile, errLogger)
	
	// Validation should fail without a task file
	if result.Valid {
		t.Error("Validation passed with no task file, expected failure")
	}
	
	// Should log an error about missing task file
	if !errLogger.errorCalled {
		t.Error("No error was logged for missing task file")
	}
	
	// Reset for next test
	errLogger.reset()
	
	// Test with no task file but in dry run mode
	configDryRun := &Configuration{
		TaskFile: "",
		DryRun:   true,
		Paths:    []string{"testfile"},
		ApiKey:   "test-key",
	}
	
	// Run validation in dry run mode
	result = doValidateInputs(configDryRun, errLogger)
	
	// Validation should pass in dry run mode even without a task file
	if !result.Valid {
		t.Errorf("Validation failed in dry run mode without a task file: %s", result.ErrorMessage)
	}
}

// errorTrackingLogger is a minimal logger that tracks if error methods were called
type errorTrackingLogger struct {
	errorCalled bool
}

func (l *errorTrackingLogger) Error(format string, args ...interface{}) {
	l.errorCalled = true
}

func (l *errorTrackingLogger) reset() {
	l.errorCalled = false
}

func (l *errorTrackingLogger) Debug(format string, args ...interface{}) {}
func (l *errorTrackingLogger) Info(format string, args ...interface{})  {}
func (l *errorTrackingLogger) Warn(format string, args ...interface{})  {}
func (l *errorTrackingLogger) Fatal(format string, args ...interface{}) {}
func (l *errorTrackingLogger) Printf(format string, args ...interface{}) {}
func (l *errorTrackingLogger) Println(v ...interface{})                  {}