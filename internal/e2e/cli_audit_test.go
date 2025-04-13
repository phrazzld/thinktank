package e2e

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditLogging tests the basic functionality of audit logging
// Simplified from the original to focus on file creation and basic content
func TestAuditLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a new test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files
	instructionsFile := env.CreateTestFile("instructions.md", "Implement a new feature")
	env.CreateTestFile("src/main.go", CreateGoSourceFileContent())

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
	VerifyOutput(t, stdout, stderr, exitCode, 0, "")

	// Verify both output and audit log files were created
	if !env.FileExists(filepath.Join("output", "test-model.md")) {
		t.Errorf("Output file was not created")
	}

	if !env.FileExists("audit.log") {
		t.Errorf("Audit log file was not created at %s", auditLogFile)
		return
	}

	// Read the audit log file
	auditContent, err := env.ReadFile("audit.log")
	if err != nil {
		t.Fatalf("Failed to read audit log file: %v", err)
	}

	// Parse the audit log entries
	auditEntries := parseAuditLog(t, auditContent)
	if len(auditEntries) == 0 {
		t.Fatalf("No audit log entries found")
	}

	// Verify at least the key operations are present
	keyOperations := []string{"ExecuteStart", "ExecuteEnd"}
	for _, op := range keyOperations {
		foundOp := false
		for _, entry := range auditEntries {
			operation, ok := entry["operation"].(string)
			if ok && operation == op {
				foundOp = true
				break
			}
		}
		if !foundOp {
			t.Errorf("Expected operation %q not found in audit log", op)
		}
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
