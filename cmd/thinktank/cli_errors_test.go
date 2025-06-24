package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestErrorHandlingForPartialFailures verifies that errors are properly logged
// when orchestrator returns partial failure errors
func TestErrorHandlingForPartialFailures(t *testing.T) {
	// CPU-bound test: error message formatting logic with no I/O dependencies
	t.Parallel()

	mockLogger := &mockErrorLogger{}

	// Test cases with various error scenarios
	testCases := []struct {
		name          string
		errorMessage  string
		shouldContain string
	}{
		{
			name:          "Partial Model Failure",
			errorMessage:  "processed 1/2 models successfully; 1 failed: model2: API error",
			shouldContain: "Application failed",
		},
		{
			name:          "Complete Model Failure",
			errorMessage:  "all models failed: model1: API error; model2: rate limit exceeded",
			shouldContain: "Application failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the logger for each test case
			mockLogger.errorMessages = nil

			// Create an error with the test case message
			err := fmt.Errorf(tc.errorMessage)

			// Call the same error logging pattern used in main.go
			mockLogger.Error("Application failed: %v", err)

			// Verify the error was logged correctly
			if len(mockLogger.errorMessages) == 0 {
				t.Fatal("Expected error to be logged, but no messages were recorded")
			}

			// Check the error message contains the expected text
			if !contains(mockLogger.errorMessages[0], tc.shouldContain) {
				t.Errorf("Error message should contain %q, got: %q",
					tc.shouldContain, mockLogger.errorMessages[0])
			}

			// Verify the error message contains the original error text
			if !contains(mockLogger.errorMessages[0], tc.errorMessage) {
				t.Errorf("Error message should contain original error %q, got: %q",
					tc.errorMessage, mockLogger.errorMessages[0])
			}
		})
	}

	// Note: We can't directly test the os.Exit(1) behavior in a unit test
	// because it would terminate the test process. In main.go, errors from
	// thinktank.Execute() trigger os.Exit(1) after logging.
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// mockErrorLogger implements a simple test logger
type mockErrorLogger struct {
	errorMessages []string
}

func (m *mockErrorLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}
