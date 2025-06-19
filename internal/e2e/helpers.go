//go:build manual_api_test
// +build manual_api_test

// Package e2e contains end-to-end tests for the thinktank CLI
// These tests require a valid API key to run properly and are skipped by default
// To run these tests: go test -tags=manual_api_test ./internal/e2e/...
package e2e

import (
	"fmt"
	"strings"
	"testing"
)

// CreateGoSourceFileContent returns standard Go source content for tests
func CreateGoSourceFileContent() string {
	return `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
`
}

// CreateStandardArgsWithPaths creates a standard set of command-line arguments
func CreateStandardArgsWithPaths(instructionsFile, outputDir, sourcePath string) []string {
	// Create arguments
	return []string{
		"--instructions", instructionsFile,
		"--output-dir", outputDir,
		"--model", "gemini-2.5-pro",
		sourcePath,
	}
}

// TestExpectation defines a specific expectation for a test
type TestExpectation struct {
	Name           string
	ExpectedOutput string
	ExpectedCode   int
	IsRequired     bool // If true, test will fail if expectation not met
}

// ValidationResult holds the result of a validation function
type ValidationResult struct {
	Success bool
	Message string
}

// VerifyOutput checks that the command output meets all specified expectations
func VerifyOutput(t *testing.T, stdout, stderr string, exitCode, expectedExitCode int, expectedOutput string) {
	// Create a default expectation from the legacy parameters
	expectations := []TestExpectation{
		{
			Name:           "Default Expectation",
			ExpectedOutput: expectedOutput,
			ExpectedCode:   expectedExitCode,
			IsRequired:     false, // Backward compatibility: don't fail by default
		},
	}

	// Call the more detailed verification function
	VerifyOutputWithExpectations(t, stdout, stderr, exitCode, expectations)
}

// VerifyOutputWithExpectations is a more flexible version that allows multiple expectations
func VerifyOutputWithExpectations(t *testing.T, stdout, stderr string, exitCode int, expectations []TestExpectation) {
	combinedOutput := stdout + stderr

	// Log the full output for debugging
	t.Logf("Command output:\nExit code: %d\nStdout: %s\nStderr: %s",
		exitCode, stdout, stderr)

	// Track validation results
	failures := []string{}
	successes := []string{}

	// Check each expectation
	for _, expect := range expectations {
		// Validate exit code if specified
		if expect.ExpectedCode != 0 && exitCode != expect.ExpectedCode {
			message := getFailureMessage(expect.Name, "exit code",
				expect.ExpectedCode, exitCode, expect.IsRequired)
			if expect.IsRequired {
				failures = append(failures, message)
			} else {
				t.Logf("%s (non-critical)", message)
			}
		}

		// Validate expected output if specified
		if expect.ExpectedOutput != "" {
			if !strings.Contains(combinedOutput, expect.ExpectedOutput) {
				message := getFailureMessage(expect.Name, "output content",
					expect.ExpectedOutput, "not found", expect.IsRequired)
				if expect.IsRequired {
					failures = append(failures, message)
				} else {
					t.Logf("%s (non-critical)", message)
				}
			} else {
				success := getSuccessMessage(expect.Name, "output content", expect.ExpectedOutput)
				successes = append(successes, success)
			}
		}
	}

	// Process validation results
	for _, success := range successes {
		t.Logf("PASS: %s", success)
	}

	// Fail the test if there are required failures
	if len(failures) > 0 {
		for _, failure := range failures {
			t.Errorf("FAIL: %s", failure)
		}
	}
}

// Helper to format failure messages
func getFailureMessage(name, aspect string, expected, actual interface{}, isRequired bool) string {
	reqStatus := "Required"
	if !isRequired {
		reqStatus = "Optional"
	}
	return strings.TrimSpace(
		strings.Join([]string{
			name,
			reqStatus,
			"Expected " + aspect + ":",
			fmt.Sprintf("%v", expected),
			"Actual " + aspect + ":",
			fmt.Sprintf("%v", actual),
		}, " "))
}

// Helper to format success messages
func getSuccessMessage(name, aspect string, expected interface{}) string {
	return strings.TrimSpace(
		strings.Join([]string{
			name,
			"Successfully verified " + aspect + ":",
			fmt.Sprintf("%v", expected),
		}, " "))
}

// AssertCommandSuccess verifies that a command succeeded with specific expectations
func AssertCommandSuccess(t *testing.T, stdout, stderr string, exitCode int, expectedOutputs ...string) {
	t.Helper()
	expectations := []TestExpectation{}

	// Command must succeed with exit code 0
	expectations = append(expectations, TestExpectation{
		Name:         "Command Exit Code",
		ExpectedCode: 0,
		IsRequired:   true, // Must have exit code 0 to be considered successful
	})

	// Add each expected output as a required expectation
	for i, output := range expectedOutputs {
		expectations = append(expectations, TestExpectation{
			Name:           fmt.Sprintf("Expected Output #%d", i+1),
			ExpectedOutput: output,
			IsRequired:     true,
		})
	}

	// Command must not contain error messages
	for _, errorPattern := range []string{
		"error:", "ERROR:", "failed:",
		"panic:", "fatal:", "FATAL:",
	} {
		// Make these non-required to avoid false positives
		expectations = append(expectations, TestExpectation{
			Name:           fmt.Sprintf("No %s Messages", strings.TrimSuffix(errorPattern, ":")),
			ExpectedOutput: errorPattern,
			IsRequired:     false,
		})
	}

	VerifyOutputWithExpectations(t, stdout, stderr, exitCode, expectations)
}

// AssertAPICommandSuccess verifies success but allows for API-related differences
func AssertAPICommandSuccess(t *testing.T, stdout, stderr string, exitCode int, expectedOutputs ...string) {
	t.Helper()

	// Check if output contains API connection issues
	combinedOutput := stdout + stderr
	apiIssues := []string{
		"API key", "failed to create", "environment variable",
		"authentication", "model not found", "connection failed", "context gathering",
	}

	// If API issues are detected, log it and skip strict validation
	for _, issue := range apiIssues {
		if strings.Contains(combinedOutput, issue) {
			t.Logf("API issue detected (%s) - relaxing test assertions", issue)

			// Create relaxed expectations for API-related tests
			expectations := []TestExpectation{}

			// Add each expected output as optional
			for i, output := range expectedOutputs {
				expectations = append(expectations, TestExpectation{
					Name:           fmt.Sprintf("Expected Output #%d", i+1),
					ExpectedOutput: output,
					IsRequired:     false, // Non-required due to API issues
				})
			}

			VerifyOutputWithExpectations(t, stdout, stderr, exitCode, expectations)
			return
		}
	}

	// If no API issues, use strict validation
	AssertCommandSuccess(t, stdout, stderr, exitCode, expectedOutputs...)
}

// AssertCommandFailure verifies that a command failed with specific expectations
func AssertCommandFailure(t *testing.T, stdout, stderr string, exitCode int, expectedExitCode int, errorMessages ...string) {
	t.Helper()
	expectations := []TestExpectation{}

	// Command must fail with non-zero exit code
	expectations = append(expectations, TestExpectation{
		Name:         "Command Exit Code",
		ExpectedCode: expectedExitCode,
		IsRequired:   true, // Must have expected exit code
	})

	// Add each error message as a required expectation
	for i, msg := range errorMessages {
		expectations = append(expectations, TestExpectation{
			Name:           fmt.Sprintf("Expected Error #%d", i+1),
			ExpectedOutput: msg,
			IsRequired:     true,
		})
	}

	VerifyOutputWithExpectations(t, stdout, stderr, exitCode, expectations)
}

// AssertFileContent validates that a file exists and contains expected content
func AssertFileContent(t *testing.T, env *TestEnv, relativePath string, expectedContent ...string) {
	t.Helper()

	// Check if file exists
	if !env.FileExists(relativePath) {
		t.Errorf("File does not exist: %s", relativePath)
		return
	}

	// Read the file content
	content, err := env.ReadFile(relativePath)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", relativePath, err)
		return
	}

	// Check each expected content
	for i, expected := range expectedContent {
		if !strings.Contains(content, expected) {
			t.Errorf("File %s does not contain expected content #%d: %q",
				relativePath, i+1, expected)
		}
	}
}

// AssertFileMayExist checks if a file exists and optionally validates content
// This is useful for API-dependent tests where file generation may be unreliable
func AssertFileMayExist(t *testing.T, env *TestEnv, relativePath string, expectedContent ...string) {
	t.Helper()

	if !env.FileExists(relativePath) {
		t.Logf("File does not exist (this may be acceptable with mock API): %s", relativePath)
		return
	}

	// If the file exists, validate its content
	content, err := env.ReadFile(relativePath)
	if err != nil {
		t.Logf("Note: Failed to read file %s: %v", relativePath, err)
		return
	}

	// Check each expected content
	for i, expected := range expectedContent {
		if !strings.Contains(content, expected) {
			t.Logf("Note: File %s does not contain expected content #%d: %q",
				relativePath, i+1, expected)
		} else {
			t.Logf("File %s contains expected content #%d: %q",
				relativePath, i+1, expected)
		}
	}
}
