// main_flags_test.go
package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestClarifyFlagRemoved verifies that the clarify flag has been removed from the CLI
func TestClarifyFlagRemoved(t *testing.T) {
	// Skip this test if running in a normal Go test environment
	// since we need to invoke the binary directly
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping test that requires RUN_INTEGRATION_TESTS=true")
	}

	// Build a temporary binary specifically for this test
	tempBin := buildTempBinary(t)
	defer os.Remove(tempBin)

	// Run the binary with --help flag
	cmd := exec.Command(tempBin, "--help")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	// Get the output
	output := stdout.String()

	// Check that the output does not contain the clarify flag
	if strings.Contains(output, "--clarify") {
		t.Error("Help output still contains --clarify flag, which should have been removed")
	}

	if strings.Contains(output, "Enable interactive task clarification") {
		t.Error("Help output still contains task clarification description, which should have been removed")
	}
}

// TestConvertConfigNoClarity verifies that convertConfigToMap() does not include clarify_task
func TestConvertConfigNoClarity(t *testing.T) {
	// Create a minimal config for testing
	config := &Configuration{
		TaskDescription: "Test task",
		OutputFile:      "output.md",
	}

	// Convert to map
	configMap := convertConfigToMap(config)

	// Check that clarify_task is not in the map
	_, hasClarifyTask := configMap["clarify_task"]
	if hasClarifyTask {
		t.Error("convertConfigToMap() should not include clarify_task key")
	}
}

// Helper function to build a temporary binary for testing
func buildTempBinary(t *testing.T) string {
	tempBin := "architect-test-binary"
	cmd := exec.Command("go", "build", "-o", tempBin, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	return tempBin
}
