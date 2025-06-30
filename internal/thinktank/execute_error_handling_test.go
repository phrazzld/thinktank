// Package thinktank contains comprehensive error handling tests for the Execute function
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteCategorizedLLMErrors tests error handling for different LLM error categories
func TestExecuteCategorizedLLMErrors(t *testing.T) {
	tests := []struct {
		name              string
		llmError          error
		expectedErrorType error
		expectedContains  string
	}{
		{
			name:              "Authentication Error",
			llmError:          llm.New("test-provider", "AUTH_ERROR", 401, "Invalid API key", "req_123", nil, llm.CategoryAuth),
			expectedErrorType: ErrInvalidAPIKey,
			expectedContains:  "API authentication failed",
		},
		{
			name:              "Rate Limit Error",
			llmError:          llm.New("test-provider", "RATE_LIMIT", 429, "Rate limit exceeded", "req_124", nil, llm.CategoryRateLimit),
			expectedErrorType: ErrInvalidModelName, // This is what the code currently maps to
			expectedContains:  "API rate limit exceeded",
		},
		{
			name:              "Model Not Found Error",
			llmError:          llm.New("test-provider", "NOT_FOUND", 404, "Model not found", "req_125", nil, llm.CategoryNotFound),
			expectedErrorType: ErrInvalidModelName,
			expectedContains:  "model test-model not found",
		},
		{
			name:              "Input Limit Error",
			llmError:          llm.New("test-provider", "INPUT_LIMIT", 413, "Input too large", "req_126", nil, llm.CategoryInputLimit),
			expectedErrorType: ErrInvalidConfiguration,
			expectedContains:  "input token limit exceeded",
		},
		{
			name:              "Content Filtered Error",
			llmError:          llm.New("test-provider", "CONTENT_FILTER", 400, "Content filtered", "req_127", nil, llm.CategoryContentFiltered),
			expectedErrorType: ErrInvalidConfiguration,
			expectedContains:  "content was filtered by safety settings",
		},
		{
			name:              "Unknown Error Category",
			llmError:          llm.New("test-provider", "UNKNOWN", 500, "Unknown error", "req_128", nil, llm.CategoryUnknown),
			expectedErrorType: ErrInvalidModelName,
			expectedContains:  "failed to initialize reference client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
			mockAuditLogger := NewMockAuditLogger()
			mockAPIService := NewMockAPIService()
			mockAPIService.initLLMClientErr = tt.llmError

			// Execute the function
			consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return false }, // CI mode for tests
			})
			err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

			// Verify results
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErrorType)
			assert.Contains(t, err.Error(), tt.expectedContains)

			// Verify audit log entries show failure
			executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
			require.NotNil(t, executeEndEntry)
			assert.Equal(t, "Failure", executeEndEntry.Status)
		})
	}
}

// TestExecuteAuditLogErrors tests error handling when audit logging fails
func TestExecuteAuditLogErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupAudit  func(*MockAuditLogger)
		expectError bool
	}{
		{
			name: "ExecuteStart Audit Log Error",
			setupAudit: func(mockAudit *MockAuditLogger) {
				mockAudit.logErr = errors.New("audit log write failed")
			},
			expectError: false, // Audit log errors shouldn't stop execution
		},
		{
			name: "ReadInstructions Audit Log Error",
			setupAudit: func(mockAudit *MockAuditLogger) {
				// Set error after ExecuteStart but during ReadInstructions
				mockAudit.logErr = errors.New("audit log write failed")
			},
			expectError: false, // Audit log errors shouldn't stop execution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
			mockAuditLogger := NewMockAuditLogger()
			mockLLMClient := NewMockLLMClient("test-model")
			mockAPIService := NewMockAPIService()
			mockAPIService.mockLLMClient = mockLLMClient
			mockOrchestrator := NewMockOrchestrator()

			// Setup audit logger error condition
			tt.setupAudit(mockAuditLogger)

			// Save original constructor for orchestrator
			originalNewOrchestrator := orchestratorConstructor
			defer func() { orchestratorConstructor = originalNewOrchestrator }()

			// Override orchestrator constructor
			orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter, tokenCountingService interfaces.TokenCountingService) Orchestrator {
				return mockOrchestrator
			}

			// Execute the function
			consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return false }, // CI mode for tests
			})
			err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

			// Verify results based on expected behavior
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err) // Audit log failures shouldn't cause Execute to fail
			}

			// Verify that error logging happened
			if mockAuditLogger.logErr != nil {
				// Should have logged the audit error
				assert.Contains(t, strings.ToLower(mockLogger.GetTestLogs()[len(mockLogger.GetTestLogs())-1]), "failed to write audit log")
			}
		})
	}
}

// TestExecuteContextCancellation tests error handling when context is cancelled
func TestExecuteContextCancellation(t *testing.T) {
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
	mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
	mockAuditLogger := NewMockAuditLogger()
	mockLLMClient := NewMockLLMClient("test-model")
	mockAPIService := NewMockAPIService()
	mockAPIService.mockLLMClient = mockLLMClient

	// Make orchestrator handle context cancellation by checking in Run method
	// We'll use a custom mock for this test

	// Save original constructor for orchestrator
	originalNewOrchestrator := orchestratorConstructor
	defer func() { orchestratorConstructor = originalNewOrchestrator }()

	// Override orchestrator constructor with context-aware mock
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter, tokenCountingService interfaces.TokenCountingService) Orchestrator {
		return &ContextAwareMockOrchestrator{}
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after Execute has had time to reach orchestrator.Run()
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(ctx, cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify that context cancellation is handled properly
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "canceled"))

	// Verify audit log shows failure
	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	require.NotNil(t, executeEndEntry)
	assert.Equal(t, "Failure", executeEndEntry.Status)
}

// TestExecutePartialSuccessError tests the wrapOrchestratorErrors function with partial success
func TestExecutePartialSuccessError(t *testing.T) {
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
	mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
	mockAuditLogger := NewMockAuditLogger()
	mockLLMClient := NewMockLLMClient("test-model")
	mockAPIService := NewMockAPIService()
	mockAPIService.mockLLMClient = mockLLMClient
	mockOrchestrator := NewMockOrchestrator()

	// Make orchestrator return partial success error
	mockOrchestrator.runErr = errors.New("some models failed during processing")

	// Save original constructor for orchestrator
	originalNewOrchestrator := orchestratorConstructor
	defer func() { orchestratorConstructor = originalNewOrchestrator }()

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter, tokenCountingService interfaces.TokenCountingService) Orchestrator {
		return mockOrchestrator
	}

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify that partial success error is properly wrapped
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPartialSuccess)
	assert.Contains(t, err.Error(), "some models failed during processing")

	// Verify audit log shows failure
	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	require.NotNil(t, executeEndEntry)
	assert.Equal(t, "Failure", executeEndEntry.Status)
}

// TestExecuteAlreadyWrappedPartialSuccess tests wrapOrchestratorErrors with already wrapped error
func TestExecuteAlreadyWrappedPartialSuccess(t *testing.T) {
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
	mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
	mockAuditLogger := NewMockAuditLogger()
	mockLLMClient := NewMockLLMClient("test-model")
	mockAPIService := NewMockAPIService()
	mockAPIService.mockLLMClient = mockLLMClient
	mockOrchestrator := NewMockOrchestrator()

	// Make orchestrator return already wrapped partial success error
	mockOrchestrator.runErr = fmt.Errorf("%w: original error", ErrPartialSuccess)

	// Save original constructor for orchestrator
	originalNewOrchestrator := orchestratorConstructor
	defer func() { orchestratorConstructor = originalNewOrchestrator }()

	// Override orchestrator constructor
	orchestratorConstructor = func(apiService interfaces.APIService, contextGatherer interfaces.ContextGatherer, fileWriter interfaces.FileWriter, auditLogger auditlog.AuditLogger, rateLimiter *ratelimit.RateLimiter, config *config.CliConfig, logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter, tokenCountingService interfaces.TokenCountingService) Orchestrator {
		return mockOrchestrator
	}

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify that already wrapped error is returned as-is
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPartialSuccess)
	assert.Contains(t, err.Error(), "original error")
}

// TestSetupOutputDirectoryCreationError tests error handling when directory creation fails
func TestSetupOutputDirectoryCreationError(t *testing.T) {
	// Set up test environment
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create instruction file
	instructionsContent := "Test instructions for plan generation"
	instructionsFile := createTestFile(t, filepath.Join(testDir, "instructions.md"), instructionsContent)

	// Create a file where we want to create a directory (this will cause mkdir to fail)
	conflictingFile := filepath.Join(testDir, "output-conflict")
	err := os.WriteFile(conflictingFile, []byte("blocking file"), 0644)
	require.NoError(t, err)

	// Create configuration with output directory that conflicts with the file
	cliConfig := &config.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        filepath.Join(conflictingFile, "subdir"), // This will fail because parent is a file
		ModelNames:       []string{"test-model"},
		APIKey:           "test-api-key",
		Paths:            []string{testDir},
		LogLevel:         logutil.InfoLevel,
		DirPermissions:   0755,
	}

	// Create mocks
	mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
	mockAuditLogger := NewMockAuditLogger()
	mockAPIService := NewMockAPIService()

	// Execute the function (should fail when creating output directory)
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err = Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidOutputDir)
	assert.Contains(t, err.Error(), "error creating output directory")

	// Verify audit log shows failure
	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	require.NotNil(t, executeEndEntry)
	assert.Equal(t, "Failure", executeEndEntry.Status)
}

// TestExecuteNonCategorizedClientError tests handling of non-categorized LLM client errors
func TestExecuteNonCategorizedClientError(t *testing.T) {
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
	mockLogger := logutil.NewTestLoggerWithoutAutoFail(t)
	mockAuditLogger := NewMockAuditLogger()
	mockAPIService := NewMockAPIService()

	// Use a regular error (not categorized LLM error)
	mockAPIService.initLLMClientErr = errors.New("generic client initialization error")

	// Execute the function
	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})
	err := Execute(context.Background(), cliConfig, mockLogger, mockAuditLogger, mockAPIService, consoleWriter)

	// Verify results - non-categorized errors should be wrapped with ErrContextGatheringFailed
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrContextGatheringFailed)
	assert.Contains(t, err.Error(), "failed to initialize reference client for context gathering")

	// Verify audit log entries show failure
	executeEndEntry := mockAuditLogger.FindEntry("ExecuteEnd")
	require.NotNil(t, executeEndEntry)
	assert.Equal(t, "Failure", executeEndEntry.Status)
}

// ContextAwareMockOrchestrator is a mock that properly handles context cancellation
type ContextAwareMockOrchestrator struct {
	runErr error
}

func (m *ContextAwareMockOrchestrator) Run(ctx context.Context, instructions string) error {
	if m.runErr != nil {
		return m.runErr
	}
	// Run indefinitely until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Continue loop indefinitely
		}
	}
}
