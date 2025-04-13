package e2e

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditLogging tests the audit logging functionality
func TestAuditLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Implement a new feature")
	env.CreateTestFile("src/main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

	// Set up the output directory and audit log file
	outputDir := filepath.Join(env.TempDir, "output")
	auditLogFile := filepath.Join(env.TempDir, "audit.log")

	// Set up flags
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.OutputDir = outputDir
	flags.AuditLogFile = auditLogFile

	// Run the architect binary
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
	if err != nil {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify output file creation
	if !env.FileExists(filepath.Join("output", "test-model.md")) {
		t.Errorf("Output file was not created")
	}

	// Verify audit log file creation
	if !env.FileExists("audit.log") {
		t.Errorf("Audit log file was not created at %s", auditLogFile)
		return
	}

	// Read the audit log file
	auditContent, err := env.ReadFile("audit.log")
	if err != nil {
		t.Fatalf("Failed to read audit log file: %v", err)
	}

	// Expected operations in the audit log
	expectedOperations := []string{
		"ExecuteStart",
		"ReadInstructions",
		"GatherContextStart",
		"GatherContextEnd",
		"CheckTokens",
		"GenerateContentStart",
		"GenerateContentEnd",
		"SaveOutputStart",
		"SaveOutputEnd",
		"ExecuteEnd",
	}

	// Parse the audit log entries
	auditEntries := parseAuditLog(t, auditContent)
	if len(auditEntries) == 0 {
		t.Fatalf("No audit log entries found")
	}

	// Create a map to track which operations were found
	foundOperations := make(map[string]bool)
	for _, entry := range auditEntries {
		operation, ok := entry["operation"].(string)
		if ok {
			foundOperations[operation] = true
		}
	}

	// Verify all expected operations are present
	for _, op := range expectedOperations {
		if !foundOperations[op] {
			t.Errorf("Expected operation %q not found in audit log", op)
		}
	}

	// Verify specific details of the ExecuteStart entry
	verifyExecuteStartEntry(t, auditEntries)
}

// TestAuditLogInvalidPath tests behavior when the audit log file cannot be created
func TestAuditLogInvalidPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Implement a new feature")
	env.CreateTestFile("src/main.go", `package main

func main() {}`)

	// Set up the output directory and an invalid audit log file path
	outputDir := filepath.Join(env.TempDir, "output")
	invalidDir := filepath.Join(env.TempDir, "nonexistent-dir")
	invalidAuditFile := filepath.Join(invalidDir, "audit.log")

	// Set up flags
	flags := env.DefaultFlags
	flags.Instructions = instructionsFile
	flags.OutputDir = outputDir
	flags.AuditLogFile = invalidAuditFile

	// Run the architect binary
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{filepath.Join(env.TempDir, "src")})
	if err != nil {
		t.Fatalf("Failed to run architect: %v", err)
	}

	// The application should still run successfully, with a warning about the audit log
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)
	}

	// Verify output file creation (should still be created despite audit log issue)
	if !env.FileExists(filepath.Join("output", "test-model.md")) {
		t.Errorf("Output file was not created")
	}

	// Verify the stderr output contains an error about the audit log
	combinedOutput := stdout + stderr
	if !strings.Contains(combinedOutput, "audit log") || !strings.Contains(combinedOutput, "failed") {
		t.Errorf("Expected error message about audit log failure, none found in output")
	}
}

// parseAuditLog parses the audit log content into a slice of map entries
func parseAuditLog(t *testing.T, content string) []map[string]interface{} {
	t.Helper()

	var entries []map[string]interface{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Failed to parse audit log line: %v", err)
			t.Errorf("Line content: %s", line)
			continue
		}

		entries = append(entries, entry)
	}

	return entries
}

// verifyExecuteStartEntry checks that the ExecuteStart entry has the expected fields
func verifyExecuteStartEntry(t *testing.T, entries []map[string]interface{}) {
	t.Helper()

	for _, entry := range entries {
		operation, ok := entry["operation"].(string)
		if !ok || operation != "ExecuteStart" {
			continue
		}

		// Check status
		status, ok := entry["status"].(string)
		if !ok || status != "InProgress" {
			t.Errorf("Expected ExecuteStart status to be 'InProgress', got %v", status)
		}

		// Check timestamp exists
		if _, hasTimestamp := entry["timestamp"]; !hasTimestamp {
			t.Error("ExecuteStart entry missing timestamp")
		}

		// Check inputs field exists and contains expected fields
		inputs, ok := entry["inputs"].(map[string]interface{})
		if !ok {
			t.Error("ExecuteStart entry missing inputs field")
			continue
		}

		// Check for model_names field in inputs
		modelNames, ok := inputs["model_names"]
		if !ok {
			t.Error("ExecuteStart inputs missing model_names field")
		} else {
			// Verify the model names array contains at least one entry
			modelNamesArr, ok := modelNames.([]interface{})
			if !ok || len(modelNamesArr) == 0 {
				t.Errorf("ExecuteStart inputs has invalid model_names field: %v", modelNames)
			}
		}

		return // Found and validated the entry
	}

	t.Error("ExecuteStart entry not found in audit log")
}
