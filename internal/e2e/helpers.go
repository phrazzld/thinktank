// Package e2e contains end-to-end tests for the architect CLI
package e2e

import (
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
		"--model", "test-model",
		sourcePath,
	}
}

// VerifyOutput checks standard output parameters
func VerifyOutput(t *testing.T, stdout, stderr string, exitCode, expectedExitCode int, expectedOutput string) {
	// Check exit code
	if exitCode != expectedExitCode {
		t.Errorf("Expected exit code %d, got %d", expectedExitCode, exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Check output if specified
	if expectedOutput != "" {
		combinedOutput := stdout + stderr
		if !strings.Contains(combinedOutput, expectedOutput) {
			t.Errorf("Expected output to contain %q, but got: %s", expectedOutput, combinedOutput)
		}
	}
}
