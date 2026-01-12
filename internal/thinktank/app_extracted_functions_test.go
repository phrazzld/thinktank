// Package thinktank contains tests for extracted pure functions from app.go
package thinktank

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// TestGatherProjectFiles tests the gatherProjectFiles function with table-driven tests
func TestGatherProjectFiles(t *testing.T) {
	tests := []struct {
		name        string
		cliConfig   *config.CliConfig
		auditLogErr error
		wantErr     bool
		expectedErr error
		validateLog func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger)
	}{
		{
			name: "successful setup with valid output directory",
			cliConfig: &config.CliConfig{
				OutputDir:        "test-output",
				DirPermissions:   0755,
				InstructionsFile: "test.md",
				Paths:            []string{"src/"},
				Include:          "*.go",
				Exclude:          "vendor/",
				ExcludeNames:     "test_",
				DryRun:           false,
				Verbose:          true,
				ModelNames:       []string{"gemini-3-flash"},
				LogLevel:         logutil.InfoLevel,
			},
			auditLogErr: nil,
			wantErr:     false,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				// Check that the expected log messages were recorded
				assert.Contains(t, logger.infoMessages, "Using output directory: test-output")
				// Note: No debug message for setup is logged in the actual implementation

				// Check audit log entry
				entry := auditLogger.FindEntry("ExecuteStart")
				assert.NotNil(t, entry)
				assert.Equal(t, "InProgress", entry.Status)
			},
		},
		{
			name: "successful setup in dry-run mode",
			cliConfig: &config.CliConfig{
				OutputDir:        "",
				InstructionsFile: "test.md",
				Paths:            []string{"src/"},
				DryRun:           true,
				ModelNames:       []string{"gemini-3-flash"},
				LogLevel:         logutil.InfoLevel,
			},
			auditLogErr: nil,
			wantErr:     false,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				// In dry-run mode with empty OutputDir, it generates and uses a directory
				// Check for "Generated output directory" and "Using output directory" messages
				found := false
				for _, msg := range logger.infoMessages {
					if containsSubstring(msg, "Generated output directory:") || containsSubstring(msg, "Using output directory:") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected output directory setup messages in dry-run mode")

				// Check audit log entry
				entry := auditLogger.FindEntry("ExecuteStart")
				assert.NotNil(t, entry)
				assert.Equal(t, "InProgress", entry.Status)
			},
		},
		{
			name: "audit logger error is logged but doesn't fail operation",
			cliConfig: &config.CliConfig{
				OutputDir:        "test-output",
				DirPermissions:   0755,
				InstructionsFile: "test.md",
				Paths:            []string{"src/"},
				DryRun:           false,
				ModelNames:       []string{"gemini-3-flash"},
				LogLevel:         logutil.InfoLevel,
			},
			auditLogErr: errors.New("audit log write failed"),
			wantErr:     false,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				// Should have logged the audit error
				found := false
				for _, msg := range logger.errorMessages {
					if msg == "Failed to write audit log: audit log write failed" {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected audit log error message not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockLogger := NewMockLogger()
			mockAuditLogger := NewMockAuditLogger()

			// Set up audit logger error if specified
			if tt.auditLogErr != nil {
				mockAuditLogger.SetLogError(tt.auditLogErr)
			}

			// Create context
			ctx := context.Background()

			// Call the function
			err := gatherProjectFiles(ctx, tt.cliConfig, mockLogger, mockAuditLogger)

			// Verify expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// Validate log messages if provided
			if tt.validateLog != nil {
				tt.validateLog(t, mockLogger, mockAuditLogger)
			}
		})
	}
}

// TestProcessFiles tests the processFiles function with table-driven tests
func TestProcessFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create test instruction files
	validInstructionsPath := filepath.Join(tempDir, "valid_instructions.md")
	err := os.WriteFile(validInstructionsPath, []byte("Test instructions content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		cliConfig            *config.CliConfig
		auditLogErr          error
		expectedInstructions string
		wantErr              bool
		expectedErr          error
		validateLog          func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger)
	}{
		{
			name: "successfully read instructions file",
			cliConfig: &config.CliConfig{
				InstructionsFile: validInstructionsPath,
				DryRun:           false,
			},
			expectedInstructions: "Test instructions content",
			wantErr:              false,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				assert.Contains(t, logger.infoMessages, "Successfully read instructions from "+validInstructionsPath)

				// Check audit log entry
				entry := auditLogger.FindEntry("ReadInstructions")
				assert.NotNil(t, entry)
				assert.Equal(t, "Success", entry.Status)
				assert.Equal(t, 25, entry.Outputs["content_length"]) // "Test instructions content" is 25 chars
			},
		},
		{
			name: "dry-run mode without instructions file",
			cliConfig: &config.CliConfig{
				InstructionsFile: "",
				DryRun:           true,
			},
			expectedInstructions: "Dry run mode - no instructions provided",
			wantErr:              false,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				assert.Contains(t, logger.infoMessages, "Dry run mode: proceeding without instructions file")
			},
		},
		{
			name: "fail to read non-existent instructions file",
			cliConfig: &config.CliConfig{
				InstructionsFile: filepath.Join(tempDir, "non_existent.md"),
				DryRun:           false,
			},
			wantErr:     true,
			expectedErr: ErrInvalidInstructions,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				// Should have logged the error
				found := false
				for _, msg := range logger.errorMessages {
					if contains := "Failed to read instructions file"; contains != "" && containsSubstring(msg, contains) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error message about failing to read instructions file")

				// Check audit log entry for failure
				entry := auditLogger.FindEntry("ReadInstructions")
				assert.NotNil(t, entry)
				assert.Equal(t, "Failure", entry.Status)
			},
		},
		{
			name: "error when instructions file required but not provided",
			cliConfig: &config.CliConfig{
				InstructionsFile: "",
				DryRun:           false,
			},
			wantErr:     true,
			expectedErr: ErrInvalidInstructions,
		},
		{
			name: "audit log error doesn't fail the operation",
			cliConfig: &config.CliConfig{
				InstructionsFile: validInstructionsPath,
				DryRun:           false,
			},
			auditLogErr:          errors.New("audit log write failed"),
			expectedInstructions: "Test instructions content",
			wantErr:              false,
			validateLog: func(t *testing.T, logger *MockLogger, auditLogger *MockAuditLogger) {
				// Should have logged the audit error
				found := false
				for _, msg := range logger.errorMessages {
					if msg == "Failed to write audit log: audit log write failed" {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected audit log error message not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockLogger := NewMockLogger()
			mockAuditLogger := NewMockAuditLogger()

			// Set up audit logger error if specified
			if tt.auditLogErr != nil {
				mockAuditLogger.SetLogError(tt.auditLogErr)
			}

			// Create context
			ctx := context.Background()

			// Call the function
			instructions, err := processFiles(ctx, tt.cliConfig, mockLogger, mockAuditLogger)

			// Verify expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInstructions, instructions)
			}

			// Validate log messages if provided
			if tt.validateLog != nil {
				tt.validateLog(t, mockLogger, mockAuditLogger)
			}
		})
	}
}

// TestWriteResults tests the writeResults function with table-driven tests
func TestWriteResults(t *testing.T) {
	tests := []struct {
		name         string
		orchestrator func() *MockOrchestrator
		instructions string
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "successful orchestrator execution",
			orchestrator: func() *MockOrchestrator {
				return &MockOrchestrator{runErr: nil}
			},
			instructions: "test instructions",
			wantErr:      false,
		},
		{
			name: "orchestrator returns generic error",
			orchestrator: func() *MockOrchestrator {
				return &MockOrchestrator{runErr: errors.New("generic error")}
			},
			instructions: "test instructions",
			wantErr:      true,
		},
		{
			name: "orchestrator returns partial success error",
			orchestrator: func() *MockOrchestrator {
				// Simulate an error that contains the partial processing failure message
				partialErr := errors.New("some models failed during processing")
				return &MockOrchestrator{runErr: partialErr}
			},
			instructions: "test instructions",
			wantErr:      true,
			expectedErr:  ErrPartialSuccess,
		},
		{
			name: "orchestrator returns already wrapped partial success",
			orchestrator: func() *MockOrchestrator {
				return &MockOrchestrator{runErr: ErrPartialSuccess}
			},
			instructions: "test instructions",
			wantErr:      true,
			expectedErr:  ErrPartialSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context
			ctx := context.Background()

			// Get orchestrator
			orch := tt.orchestrator()

			// Call the function
			err := writeResults(ctx, orch, tt.instructions)

			// Verify expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGenerateOutput tests the generateOutput function with table-driven tests
func TestGenerateOutput(t *testing.T) {
	tests := []struct {
		name        string
		cliConfig   *config.CliConfig
		apiService  func() *MockAPIService
		wantErr     bool
		expectedErr error
		validateLog func(t *testing.T, logger *MockLogger)
	}{
		{
			name: "successful orchestrator creation in dry-run mode",
			cliConfig: &config.CliConfig{
				DryRun:                     true,
				ModelNames:                 []string{"gemini-3-flash"},
				MaxConcurrentRequests:      5,
				RateLimitRequestsPerMinute: 60,
				DirPermissions:             0755,
				FilePermissions:            0644,
			},
			apiService: func() *MockAPIService {
				// In dry-run mode, no LLM client initialization is needed
				return &MockAPIService{}
			},
			wantErr: false,
		},
		{
			name: "successful orchestrator creation with LLM client",
			cliConfig: &config.CliConfig{
				DryRun:                     false,
				ModelNames:                 []string{"gemini-3-flash"},
				APIEndpoint:                "https://api.openrouter.ai",
				MaxConcurrentRequests:      5,
				RateLimitRequestsPerMinute: 60,
				DirPermissions:             0755,
				FilePermissions:            0644,
			},
			apiService: func() *MockAPIService {
				mockService := &MockAPIService{
					mockLLMClient: &MockLLMClient{},
				}
				return mockService
			},
			wantErr: false,
		},
		{
			name: "LLM client initialization fails with auth error",
			cliConfig: &config.CliConfig{
				DryRun:      false,
				ModelNames:  []string{"invalid-model"},
				APIEndpoint: "https://api.openrouter.ai",
			},
			apiService: func() *MockAPIService {
				// Create a categorized error for auth failure
				authErr := &MockCategorizedError{
					category: llm.CategoryAuth,
					message:  "authentication failed",
				}
				mockService := &MockAPIService{
					initLLMClientErr: authErr,
				}
				return mockService
			},
			wantErr:     true,
			expectedErr: ErrInvalidAPIKey,
			validateLog: func(t *testing.T, logger *MockLogger) {
				// Should have logged the error with category
				found := false
				for _, msg := range logger.errorMessages {
					if containsSubstring(msg, "Failed to initialize reference client for context gathering") &&
						containsSubstring(msg, "(category: Auth)") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected categorized error log message not found")
			},
		},
		{
			name: "LLM client initialization fails with rate limit error",
			cliConfig: &config.CliConfig{
				DryRun:      false,
				ModelNames:  []string{"gemini-3-flash"},
				APIEndpoint: "https://api.openrouter.ai",
			},
			apiService: func() *MockAPIService {
				// Create a categorized error for rate limit
				rateLimitErr := &MockCategorizedError{
					category: llm.CategoryRateLimit,
					message:  "rate limit exceeded",
				}
				mockService := &MockAPIService{
					initLLMClientErr: rateLimitErr,
				}
				return mockService
			},
			wantErr:     true,
			expectedErr: ErrInvalidModelName,
			validateLog: func(t *testing.T, logger *MockLogger) {
				// Should have logged the error with category
				found := false
				for _, msg := range logger.errorMessages {
					if containsSubstring(msg, "Failed to initialize reference client for context gathering") &&
						containsSubstring(msg, "(category: RateLimit)") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected categorized error log message not found")
			},
		},
		{
			name: "LLM client initialization fails with generic error",
			cliConfig: &config.CliConfig{
				DryRun:      false,
				ModelNames:  []string{"gemini-3-flash"},
				APIEndpoint: "https://api.openrouter.ai",
			},
			apiService: func() *MockAPIService {
				mockService := &MockAPIService{
					initLLMClientErr: errors.New("connection failed"),
				}
				return mockService
			},
			wantErr:     true,
			expectedErr: ErrContextGatheringFailed,
			validateLog: func(t *testing.T, logger *MockLogger) {
				// Should have logged the generic error
				found := false
				for _, msg := range logger.errorMessages {
					if containsSubstring(msg, "Failed to initialize reference client for context gathering") &&
						containsSubstring(msg, "connection failed") {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error log message not found")
			},
		},
	}

	// Set the orchestrator constructor to return a mock for testing
	oldConstructor := orchestratorConstructor
	defer func() { orchestratorConstructor = oldConstructor }()

	orchestratorConstructor = func(
		apiService interfaces.APIService,
		contextGatherer interfaces.ContextGatherer,
		fileWriter interfaces.FileWriter,
		auditLogger auditlog.AuditLogger,
		rateLimiter *ratelimit.RateLimiter,
		cliConfig *config.CliConfig,
		logger logutil.LoggerInterface,
		consoleWriter logutil.ConsoleWriter,
		tokenCountingService interfaces.TokenCountingService,
	) Orchestrator {
		return &MockOrchestrator{}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockLogger := NewMockLogger()
			mockAuditLogger := NewMockAuditLogger()
			mockAPIService := tt.apiService()
			mockConsoleWriter := &MockConsoleWriter{}

			// Create context
			ctx := context.Background()

			// Call the function
			orch, err := generateOutput(ctx, tt.cliConfig, mockLogger, mockAuditLogger, mockAPIService, mockConsoleWriter)

			// Verify expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, orch)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, orch)
			}

			// Validate log messages if provided
			if tt.validateLog != nil {
				tt.validateLog(t, mockLogger)
			}
		})
	}
}

// MockConsoleWriter is a minimal mock implementation of ConsoleWriter for testing
type MockConsoleWriter struct{}

// Progress Reporting Methods
func (m *MockConsoleWriter) StartProcessing(modelCount int)          {}
func (m *MockConsoleWriter) StartStatusTracking(modelNames []string) {}
func (m *MockConsoleWriter) UpdateModelStatus(modelName string, status logutil.ModelStatus, duration time.Duration, errorMsg string) {
}
func (m *MockConsoleWriter) UpdateModelRateLimited(modelName string, retryAfter time.Duration) {}
func (m *MockConsoleWriter) RefreshStatusDisplay()                                             {}
func (m *MockConsoleWriter) FinishStatusTracking()                                             {}
func (m *MockConsoleWriter) ModelQueued(modelName string, index int)                           {}
func (m *MockConsoleWriter) ModelStarted(modelIndex, totalModels int, modelName string)        {}
func (m *MockConsoleWriter) ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration) {
}
func (m *MockConsoleWriter) ModelFailed(modelIndex, totalModels int, modelName string, reason string) {
}
func (m *MockConsoleWriter) ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration) {
}

// Modern Clean Output Methods
func (m *MockConsoleWriter) ShowProcessingLine(modelName string)                  {}
func (m *MockConsoleWriter) UpdateProcessingLine(modelName string, status string) {}
func (m *MockConsoleWriter) ShowFileOperations(message string)                    {}
func (m *MockConsoleWriter) ShowSummarySection(summary logutil.SummaryData)       {}
func (m *MockConsoleWriter) ShowOutputFiles(files []logutil.OutputFile)           {}
func (m *MockConsoleWriter) ShowFailedModels(failed []logutil.FailedModel)        {}

// Status Update Methods
func (m *MockConsoleWriter) SynthesisStarted()                    {}
func (m *MockConsoleWriter) SynthesisCompleted(outputPath string) {}
func (m *MockConsoleWriter) StatusMessage(message string)         {}

// Control Methods
func (m *MockConsoleWriter) SetQuiet(quiet bool)                 {}
func (m *MockConsoleWriter) SetNoProgress(noProgress bool)       {}
func (m *MockConsoleWriter) IsInteractive() bool                 { return false }
func (m *MockConsoleWriter) GetTerminalWidth() int               { return 80 }
func (m *MockConsoleWriter) FormatMessage(message string) string { return message }
func (m *MockConsoleWriter) ErrorMessage(message string)         {}
func (m *MockConsoleWriter) WarningMessage(message string)       {}
func (m *MockConsoleWriter) SuccessMessage(message string)       {}

// Mock implementation of CategorizedError for testing
type MockCategorizedError struct {
	category llm.ErrorCategory
	message  string
}

func (e *MockCategorizedError) Error() string {
	return e.message
}

func (e *MockCategorizedError) Category() llm.ErrorCategory {
	return e.category
}

func (e *MockCategorizedError) Unwrap() error {
	return nil
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexString(s, substr) >= 0)
}

// indexString returns the index of substr in s, or -1 if not found
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
