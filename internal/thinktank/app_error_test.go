// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// TestExecuteInstructionsFileError tests error handling when instructions file can't be read
func TestExecuteInstructionsFileError(t *testing.T) {
	// Set up test environment
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up a non-existent instructions file
	instructionsFile := filepath.Join(testDir, "nonexistent-instructions.md")

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
	originalNewOrchestrator := orchestratorConstructor

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter) Orchestrator {
		return mockOrchestrator
	}

	// Restore original constructor when test finishes
	defer func() {
		orchestratorConstructor = originalNewOrchestrator
	}()

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	if err == nil {
		t.Error("Execute did not return an error for nonexistent instructions file")
	}
	if !strings.Contains(err.Error(), "failed to read instructions file") {
		t.Errorf("Unexpected error message: %v", err)
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
	if readInstructionsEntry.Status != "Failure" {
		t.Errorf("ReadInstructions status was %s, expected Failure", readInstructionsEntry.Status)
	}

	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	if executeEndEntry == nil {
		t.Error("No ExecuteEnd entry found in audit log")
		return
	}
	if executeEndEntry.Status != "Failure" {
		t.Errorf("ExecuteEnd status was %s, expected Failure", executeEndEntry.Status)
	}
}

// TestExecuteClientInitializationError tests error handling when API client initialization fails
func TestExecuteClientInitializationError(t *testing.T) {
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
	mockAPIService := NewMockAPIService()
	mockAPIService.initLLMClientErr = errors.New("API client initialization error")
	mockOrchestrator := NewMockOrchestrator()

	// Save original constructor for orchestrator
	originalNewOrchestrator := orchestratorConstructor

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter) Orchestrator {
		return mockOrchestrator
	}

	// Restore original constructor when test finishes
	defer func() {
		orchestratorConstructor = originalNewOrchestrator
	}()

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	if err == nil {
		t.Error("Execute did not return an error for API client initialization failure")
	}
	if !strings.Contains(err.Error(), "failed to initialize reference client") {
		t.Errorf("Unexpected error message: %v", err)
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
	if executeEndEntry.Status != "Failure" {
		t.Errorf("ExecuteEnd status was %s, expected Failure", executeEndEntry.Status)
	}
}

// TestExecuteOrchestratorError tests error handling when orchestrator.Run returns an error
func TestExecuteOrchestratorError(t *testing.T) {
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
	mockOrchestrator.runErr = errors.New("orchestrator run error")

	// Save original constructor for orchestrator
	originalNewOrchestrator := orchestratorConstructor

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter) Orchestrator {
		return mockOrchestrator
	}

	// Restore original constructor when test finishes
	defer func() {
		orchestratorConstructor = originalNewOrchestrator
	}()

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	if err == nil {
		t.Error("Execute did not return an error when orchestrator.Run failed")
	}
	if !strings.Contains(err.Error(), "orchestrator run error") {
		t.Errorf("Unexpected error message: %v", err)
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
	if executeEndEntry.Status != "Failure" {
		t.Errorf("ExecuteEnd status was %s, expected Failure", executeEndEntry.Status)
	}
}

// TestSetupOutputDirectoryError tests error handling when output directory creation fails
func TestSetupOutputDirectoryError(t *testing.T) {
	// Create filesystem abstraction
	fs := testutil.NewRealFS()

	// Create a temporary test directory
	parentDir, err := os.MkdirTemp("", "thinktank-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() { _ = fs.RemoveAll(parentDir) }()

	// Create a file with the same name where we will try to create a directory
	invalidDirPath := filepath.Join(parentDir, "cannot-be-dir")
	err = fs.WriteFile(invalidDirPath, []byte("this is a file"), 0640)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a valid instructions file
	instructionsContent := "Test instructions for plan generation"
	instructionsFile := filepath.Join(parentDir, "instructions.md")
	err = fs.WriteFile(instructionsFile, []byte(instructionsContent), 0640)
	if err != nil {
		t.Fatalf("Failed to create instructions file: %v", err)
	}

	// Create configuration with the problematic output directory
	cliConfig := &config.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        filepath.Join(invalidDirPath, "subdir"), // This will fail because parent is a file
		ModelNames:       []string{"test-model"},
		APIKey:           "test-api-key",
		Paths:            []string{parentDir},
		LogLevel:         logutil.InfoLevel,
	}

	// Create mocks
	mockLogger := NewMockLogger()
	mockAuditLogger := NewMockAuditLogger()
	mockLLMClient := NewMockLLMClient("test-model")
	mockAPIService := NewMockAPIService()
	mockAPIService.mockLLMClient = mockLLMClient

	// No need to override any constructors here, since we're passing mockAPIService directly

	// Execute the function (should fail when creating output directory)
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err = Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	if err == nil {
		t.Error("Execute did not return an error for output directory creation failure")
	}
	if !strings.Contains(err.Error(), "error creating output directory") {
		t.Errorf("Unexpected error message: %v", err)
	}

	// We're mostly concerned with the error return in this test
	// The specific audit log entries might vary depending on exactly when the error occurs
	// So we don't verify them in detail here
}
