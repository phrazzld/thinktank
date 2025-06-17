// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// TestExecuteHappyPath tests the happy path of the Execute function
func TestExecuteHappyPath(t *testing.T) {
	// Set up test environment
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create instruction file
	instructionsContent := "Test instructions for plan generation"
	instructionsFile := createTestFile(t, filepath.Join(testDir, "instructions.md"), instructionsContent)

	// Set up output directory
	outputDir := filepath.Join(testDir, "output")

	// Create configuration
	cliConfig := &config.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        outputDir,
		ModelNames:       []string{"test-model"},
		APIKey:           "test-api-key",
		Paths:            []string{testDir},
		LogLevel:         logutil.InfoLevel,
	}

	// Create mocks
	mockLogger := NewMockLogger()
	mockAuditLogger := NewMockAuditLogger()
	mockLLMClient := NewMockLLMClient("test-model")
	mockAPIService := NewMockAPIService()
	mockAPIService.mockLLMClient = mockLLMClient
	mockOrchestrator := NewMockOrchestrator()

	// Save original constructor for orchestrator
	originalOrchestrator := orchestratorConstructor

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter) Orchestrator {
		return mockOrchestrator
	}

	// Restore original constructor when test finishes
	defer func() {
		orchestratorConstructor = originalOrchestrator
	}()

	// Execute the function - pass mockAPIService directly
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	if err != nil {
		t.Errorf("Execute returned an error: %v", err)
	}

	// Verify output directory was created
	// Create filesystem abstraction
	fs := testutil.NewRealFS()

	exists, _ := fs.Stat(outputDir)
	if !exists {
		t.Errorf("Output directory was not created at %s", outputDir)
	}

	// Verify audit log entries
	executeStartEntry := mockAuditLogger.FindEntry("ExecuteStart")
	if executeStartEntry == nil {
		t.Error("No ExecuteStart entry found in audit log")
		return
	}

	readInstructionsEntry := mockAuditLogger.FindEntry("ReadInstructions")
	if readInstructionsEntry == nil {
		t.Error("No ReadInstructions entry found in audit log")
		return
	}
	if readInstructionsEntry.Status != "Success" {
		t.Errorf("ReadInstructions status was %s, expected Success", readInstructionsEntry.Status)
	}

	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	if executeEndEntry == nil {
		t.Error("No ExecuteEnd entry found in audit log")
		return
	}
	if executeEndEntry.Status != "Success" {
		t.Errorf("ExecuteEnd status was %s, expected Success", executeEndEntry.Status)
	}
}

// TestExecuteDryRun tests the dry run mode
func TestExecuteDryRun(t *testing.T) {
	// Set up test environment
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create instruction file
	instructionsContent := "Test instructions for plan generation"
	instructionsFile := createTestFile(t, filepath.Join(testDir, "instructions.md"), instructionsContent)

	// Set up output directory
	outputDir := filepath.Join(testDir, "output")

	// Create configuration with dry run enabled
	cliConfig := &config.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        outputDir,
		ModelNames:       []string{"test-model"},
		APIKey:           "test-api-key",
		Paths:            []string{testDir},
		LogLevel:         logutil.InfoLevel,
		DryRun:           true, // Enable dry run mode
	}

	// Create mocks
	mockLogger := NewMockLogger()
	mockAuditLogger := NewMockAuditLogger()
	mockLLMClient := NewMockLLMClient("test-model")
	mockAPIService := NewMockAPIService()
	mockAPIService.mockLLMClient = mockLLMClient
	mockOrchestrator := NewMockOrchestrator()

	// Save original constructor for orchestrator
	originalOrchestrator := orchestratorConstructor

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter) Orchestrator {
		return mockOrchestrator
	}

	// Restore original constructor when test finishes
	defer func() {
		orchestratorConstructor = originalOrchestrator
	}()

	// Execute the function - pass mockAPIService directly
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	if err != nil {
		t.Errorf("Execute returned an error in dry run mode: %v", err)
	}

	// Verify output directory was created (even in dry run mode, we create the directory)
	// Create filesystem abstraction
	fs := testutil.NewRealFS()

	exists, _ := fs.Stat(outputDir)
	if !exists {
		t.Errorf("Output directory was not created at %s", outputDir)
	}

	// Verify audit log entries
	executeStartEntry := mockAuditLogger.FindEntry("ExecuteStart")
	if executeStartEntry == nil {
		t.Error("No ExecuteStart entry found in audit log")
		return
	}

	// Check if dry_run is true in the inputs
	foundDryRun := false
	if executeStartEntry.Inputs != nil {
		if dryRun, ok := executeStartEntry.Inputs["dry_run"].(bool); ok && dryRun {
			foundDryRun = true
		}
	}
	if !foundDryRun {
		t.Error("ExecuteStart entry doesn't show dry_run = true")
	}

	readInstructionsEntry := mockAuditLogger.FindEntry("ReadInstructions")
	if readInstructionsEntry == nil {
		t.Error("No ReadInstructions entry found in audit log")
		return
	}
	if readInstructionsEntry.Status != "Success" {
		t.Errorf("ReadInstructions status was %s, expected Success", readInstructionsEntry.Status)
	}

	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	if executeEndEntry == nil {
		t.Error("No ExecuteEnd entry found in audit log")
		return
	}
	if executeEndEntry.Status != "Success" {
		t.Errorf("ExecuteEnd status was %s, expected Success", executeEndEntry.Status)
	}
}
