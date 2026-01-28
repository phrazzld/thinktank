// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/logutil"
)

func TestRunDryRun(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test_dry_run")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test files
	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instructionsFile := filepath.Join(tempDir, "instructions.txt")
	err = os.WriteFile(instructionsFile, []byte("Test instructions"), 0644)
	if err != nil {
		t.Fatalf("Failed to create instructions file: %v", err)
	}

	cfg := &config.MinimalConfig{
		InstructionsFile: instructionsFile,
		TargetPaths:      []string{tempDir},
		ModelNames:       []string{"gemini-1.5-flash"},
		OutputDir:        tempDir,
		Format:           config.DefaultFormat,
		Exclude:          config.DefaultExcludes,
		ExcludeNames:     config.DefaultExcludeNames,
	}

	logger := logutil.NewSlogLoggerFromLogLevel(nil, logutil.InfoLevel)
	ctx := context.Background()
	instructions := "Test instructions for dry run"

	err = runDryRun(ctx, cfg, instructions, logger)
	if err != nil {
		t.Errorf("runDryRun() failed: %v", err)
	}
}
