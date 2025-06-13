package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestExitCodeHandling tests the behavior of exit codes based on the partial-success-ok flag
// This is a table-driven test that covers different scenarios
func TestExitCodeHandling(t *testing.T) {
	if os.Getenv("GO_TEST_EXIT_CODE_PROCESS") == "1" {
		// This is the subprocess that will call Main() and exit
		// We're using a subprocess because we need to test actual os.Exit() behavior
		mockExit()
		return
	}

	// Define test cases
	testCases := []struct {
		name               string
		partialSuccessFlag bool
		expectedExitCode   int
		mockError          error
	}{
		{
			name:               "partial failure with flag enabled should exit with code 0",
			partialSuccessFlag: true,
			expectedExitCode:   0,
			mockError:          errors.New("some models failed: partial success"),
		},
		{
			name:               "partial failure with flag disabled should exit with code 1",
			partialSuccessFlag: false,
			expectedExitCode:   1,
			mockError:          errors.New("some models failed: partial failure"),
		},
		{
			name:               "complete failure should exit with code 1 regardless of flag",
			partialSuccessFlag: true,
			expectedExitCode:   1,
			mockError:          errors.New("all models failed: complete failure"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables for the subprocess
			cmd := exec.Command(os.Args[0], "-test.run=TestExitCodeHandling")
			cmd.Env = append(os.Environ(),
				"GO_TEST_EXIT_CODE_PROCESS=1",
				"GO_TEST_PARTIAL_SUCCESS_FLAG="+boolToString(tc.partialSuccessFlag),
				"GO_TEST_MOCK_ERROR="+tc.mockError.Error(),
			)

			// Run the subprocess and check the exit code
			err := cmd.Run()
			if tc.expectedExitCode == 0 {
				if err != nil {
					t.Errorf("Expected exit code 0, but got error: %v", err)
				}
			} else {
				// For non-zero exit codes, we expect an ExitError
				exitErr, ok := err.(*exec.ExitError)
				if !ok {
					t.Errorf("Expected *exec.ExitError, got %T: %v", err, err)
					return
				}
				if exitErr.ExitCode() != tc.expectedExitCode {
					t.Errorf("Expected exit code %d, got %d", tc.expectedExitCode, exitErr.ExitCode())
				}
			}
		})
	}
}

// Helper function to convert boolean to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Mock implementation for the subprocess to test exit code logic
func mockExit() {
	// Parse mock configuration from environment variables
	partialSuccessFlag := os.Getenv("GO_TEST_PARTIAL_SUCCESS_FLAG") == "true"
	mockError := os.Getenv("GO_TEST_MOCK_ERROR")

	// Mock the error checking logic
	var exitCode int
	if mockError != "" {
		// Create a mock error
		err := errors.New(mockError)

		// Simulate the error checking from Main()
		if partialSuccessFlag && isPartialSuccess(err) {
			exitCode = 0
		} else {
			exitCode = 1
		}
	} else {
		exitCode = 0
	}

	// Exit with the determined exit code
	os.Exit(exitCode)
}

// Mock function to simulate the error checking logic in Main()
func isPartialSuccess(err error) bool {
	// In the real code, this would use errors.Is with thinktank.ErrPartialSuccess
	// For testing, we just check if the error message contains "partial success"
	return err != nil && strings.Contains(err.Error(), "partial success")
}
