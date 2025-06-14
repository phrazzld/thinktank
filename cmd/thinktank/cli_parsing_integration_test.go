// Package main provides integration tests for CLI parsing functionality
package main

import (
	"os"
	"testing"
)

// TestParseFlags tests the main ParseFlags function with real command line arguments
// Note: We can only test one case since ParseFlags uses the global flag.CommandLine
func TestParseFlags(t *testing.T) {
	// Save original os.Args to restore later
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test a basic valid case - this exercises the real ParseFlags() function
	os.Args = []string{"thinktank", "--instructions", "test.txt", "--dry-run", "src/"}

	config, err := ParseFlags()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the basic integration works
	if config.InstructionsFile != "test.txt" {
		t.Errorf("InstructionsFile: expected %q, got %q", "test.txt", config.InstructionsFile)
	}
	if !config.DryRun {
		t.Errorf("DryRun: expected true, got false")
	}
	if len(config.Paths) != 1 || config.Paths[0] != "src/" {
		t.Errorf("Paths: expected [\"src/\"], got %v", config.Paths)
	}
}

// Note: Additional ParseFlags tests cannot be added here due to flag redefinition issues
// when using the global flag.CommandLine. The existing test coverage comes from
// cli_args_test.go which tests ParseFlagsWithEnv extensively with custom flag sets.
