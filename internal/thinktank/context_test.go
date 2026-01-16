// Package thinktank contains tests for context gathering functionality
package thinktank

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// mockConsoleWriter is a simple mock implementation of ConsoleWriter for testing
type mockConsoleWriter struct {
	messages []string
}

func (m *mockConsoleWriter) StartProcessing(modelCount int)                             {}
func (m *mockConsoleWriter) ModelQueued(modelName string, index int)                    {}
func (m *mockConsoleWriter) ModelStarted(modelIndex, totalModels int, modelName string) {}
func (m *mockConsoleWriter) ModelCompleted(modelIndex, totalModels int, modelName string, duration time.Duration) {
}
func (m *mockConsoleWriter) ModelFailed(modelIndex, totalModels int, modelName string, reason string) {
}
func (m *mockConsoleWriter) ModelRateLimited(modelIndex, totalModels int, modelName string, retryAfter time.Duration) {
}
func (m *mockConsoleWriter) ShowProcessingLine(modelName string)                  {}
func (m *mockConsoleWriter) UpdateProcessingLine(modelName string, status string) {}
func (m *mockConsoleWriter) ShowFileOperations(message string)                    {}
func (m *mockConsoleWriter) ShowSummarySection(summary logutil.SummaryData)       {}
func (m *mockConsoleWriter) ShowOutputFiles(files []logutil.OutputFile)           {}
func (m *mockConsoleWriter) ShowFailedModels(failed []logutil.FailedModel)        {}
func (m *mockConsoleWriter) SynthesisStarted()                                    {}
func (m *mockConsoleWriter) SynthesisCompleted(outputPath string)                 {}
func (m *mockConsoleWriter) StatusMessage(message string) {
	m.messages = append(m.messages, message)
}
func (m *mockConsoleWriter) SetQuiet(quiet bool)                 {}
func (m *mockConsoleWriter) SetNoProgress(noProgress bool)       {}
func (m *mockConsoleWriter) IsInteractive() bool                 { return false }
func (m *mockConsoleWriter) GetTerminalWidth() int               { return 80 }
func (m *mockConsoleWriter) FormatMessage(message string) string { return message }
func (m *mockConsoleWriter) ErrorMessage(message string) {
	m.messages = append(m.messages, "ERROR: "+message)
}
func (m *mockConsoleWriter) WarningMessage(message string) {
	m.messages = append(m.messages, "WARNING: "+message)
}
func (m *mockConsoleWriter) SuccessMessage(message string) {
	m.messages = append(m.messages, "SUCCESS: "+message)
}
func (m *mockConsoleWriter) StartStatusTracking(modelNames []string) {}
func (m *mockConsoleWriter) UpdateModelStatus(modelName string, status logutil.ModelStatus, duration time.Duration, errorMsg string) {
}
func (m *mockConsoleWriter) UpdateModelRateLimited(modelName string, retryAfter time.Duration) {}
func (m *mockConsoleWriter) RefreshStatusDisplay()                                             {}
func (m *mockConsoleWriter) FinishStatusTracking()                                             {}

func TestGatherContext(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     map[string][]byte
		config         interfaces.GatherConfig
		dryRun         bool
		expectError    bool
		expectFiles    int
		expectAuditOps []string
	}{
		{
			name: "successful context gathering with multiple files",
			setupFiles: map[string][]byte{
				"main.go":   []byte("package main\n\nfunc main() {\n\tprintln(\"Hello World\")\n}"),
				"config.go": []byte("package main\n\ntype Config struct {\n\tValue string\n}"),
				"README.md": []byte("# Project\n\nThis is a test project."),
			},
			config: interfaces.GatherConfig{
				Paths:        []string{}, // Will be set to temp dir
				Include:      "",
				Exclude:      "",
				ExcludeNames: "",
				Format:       "{path}\n{content}",
				Verbose:      false,
				LogLevel:     logutil.InfoLevel,
			},
			dryRun:         false,
			expectError:    false,
			expectFiles:    3,
			expectAuditOps: []string{"GatherContext"},
		},
		{
			name: "dry run mode with file collection",
			setupFiles: map[string][]byte{
				"test.go":   []byte("package test"),
				"helper.go": []byte("package test\n\nfunc Helper() {}"),
			},
			config: interfaces.GatherConfig{
				Paths:        []string{}, // Will be set to temp dir
				Include:      ".go",
				Exclude:      "",
				ExcludeNames: "",
				Format:       "{path}\n{content}",
				Verbose:      true,
				LogLevel:     logutil.DebugLevel,
			},
			dryRun:         true,
			expectError:    false,
			expectFiles:    2,
			expectAuditOps: []string{"GatherContext"},
		},
		{
			name:       "empty directory",
			setupFiles: map[string][]byte{},
			config: interfaces.GatherConfig{
				Paths:        []string{}, // Will be set to temp dir
				Include:      "",
				Exclude:      "",
				ExcludeNames: "",
				Format:       "{path}\n{content}",
				Verbose:      false,
				LogLevel:     logutil.InfoLevel,
			},
			dryRun:         false,
			expectError:    false,
			expectFiles:    0,
			expectAuditOps: []string{"GatherContext"},
		},
		{
			name: "filtered files with exclude",
			setupFiles: map[string][]byte{
				"main.go":     []byte("package main"),
				"test.txt":    []byte("test content"),
				"config.json": []byte(`{"test": true}`),
			},
			config: interfaces.GatherConfig{
				Paths:        []string{}, // Will be set to temp dir
				Include:      "",
				Exclude:      ".txt,.json",
				ExcludeNames: "",
				Format:       "{path}\n{content}",
				Verbose:      false,
				LogLevel:     logutil.InfoLevel,
			},
			dryRun:         false,
			expectError:    false,
			expectFiles:    1, // Only main.go should be included
			expectAuditOps: []string{"GatherContext"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tempDir := testutil.SetupTempDir(t, "context-test-")
			testutil.CreateTestFiles(t, tempDir, tt.setupFiles)

			// Set the temp directory as the path to process
			tt.config.Paths = []string{tempDir}

			// Create mock dependencies
			mockLogger := testutil.NewMockLogger()
			mockConsoleWriter := &mockConsoleWriter{}
			mockClient := &llm.MockLLMClient{}

			// Create context gatherer
			gatherer := NewContextGatherer(mockLogger, mockConsoleWriter, tt.dryRun, mockClient, mockLogger)

			// Execute the function
			ctx := context.Background()
			files, stats, err := gatherer.GatherContext(ctx, tt.config)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("GatherContext() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if err != nil {
				return // Skip further checks if we expected an error
			}

			// Verify file count
			if len(files) != tt.expectFiles {
				t.Errorf("Expected %d files, got %d", tt.expectFiles, len(files))
			}

			// Verify stats
			if stats == nil {
				t.Fatal("Expected stats to be non-nil")
			}

			if stats.ProcessedFilesCount != tt.expectFiles {
				t.Errorf("Expected ProcessedFilesCount %d, got %d", tt.expectFiles, stats.ProcessedFilesCount)
			}

			// For dry run mode, verify processed files list is populated
			if tt.dryRun && tt.expectFiles > 0 {
				if len(stats.ProcessedFiles) != tt.expectFiles {
					t.Errorf("Dry run: expected %d processed files, got %d", tt.expectFiles, len(stats.ProcessedFiles))
				}
			}

			// Verify character and line counts are calculated
			if tt.expectFiles > 0 {
				if stats.CharCount <= 0 {
					t.Error("Expected CharCount to be positive")
				}
				if stats.LineCount <= 0 {
					t.Error("Expected LineCount to be positive")
				}
			}

			// Verify audit logging
			auditCalls := mockLogger.GetLogOpCalls()
			if len(auditCalls) < len(tt.expectAuditOps) {
				t.Errorf("Expected at least %d audit operations, got %d", len(tt.expectAuditOps), len(auditCalls))
			}

			// Check for specific audit operations
			for _, expectedOp := range tt.expectAuditOps {
				found := false
				for _, call := range auditCalls {
					if call.Operation == expectedOp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected audit operation %s not found", expectedOp)
				}
			}

			// Verify appropriate log messages based on mode
			logMessages := mockLogger.GetInfoMessages()
			if tt.dryRun {
				// Should contain dry run specific messages
				found := false
				for _, msg := range logMessages {
					if strings.Contains(msg, "Dry run mode") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected dry run mode message in logs")
				}
			} else if tt.expectFiles > 0 {
				// Should contain context gathered message
				found := false
				for _, msg := range logMessages {
					if strings.Contains(msg, "Context gathered") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected context gathered message in logs")
				}
			}

			// Verify no warning for empty directories with files > 0
			if tt.expectFiles == 0 {
				warnMessages := mockLogger.GetWarnMessages()
				found := false
				for _, msg := range warnMessages {
					if strings.Contains(msg, "No files were processed") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected warning message for empty directory")
				}
			}
		})
	}
}

func TestDisplayDryRunInfo(t *testing.T) {
	tests := []struct {
		name                string
		stats               *interfaces.ContextStats
		expectedLogMessages []string
	}{
		{
			name: "display info with multiple files",
			stats: &interfaces.ContextStats{
				ProcessedFilesCount: 3,
				CharCount:           150,
				LineCount:           25,
				ProcessedFiles:      []string{"file1.go", "file2.go", "README.md"},
			},
			expectedLogMessages: []string{
				"Dry run results:",
				"Files that would be included: 3",
				"1. file1.go",
				"2. file2.go",
				"3. README.md",
				"Context statistics:",
				"Files: 3",
				"Lines: 25",
				"Characters: 150",
				"Dry run completed successfully",
				"To generate content, run without the --dry-run flag",
			},
		},
		{
			name: "display info with no files",
			stats: &interfaces.ContextStats{
				ProcessedFilesCount: 0,
				CharCount:           0,
				LineCount:           0,
				ProcessedFiles:      []string{},
			},
			expectedLogMessages: []string{
				"Dry run results:",
				"No files matched the current filters",
				"Context statistics:",
				"Files: 0",
				"Lines: 0",
				"Characters: 0",
				"Dry run completed successfully",
				"To generate content, run without the --dry-run flag",
			},
		},
		{
			name: "display info with single file",
			stats: &interfaces.ContextStats{
				ProcessedFilesCount: 1,
				CharCount:           42,
				LineCount:           3,
				ProcessedFiles:      []string{"main.go"},
			},
			expectedLogMessages: []string{
				"Dry run results:",
				"Files that would be included: 1",
				"1. main.go",
				"Context statistics:",
				"Files: 1",
				"Lines: 3",
				"Characters: 42",
				"Dry run completed successfully",
				"To generate content, run without the --dry-run flag",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock dependencies
			mockLogger := testutil.NewMockLogger()
			mockConsoleWriter := &mockConsoleWriter{}
			mockClient := &llm.MockLLMClient{}

			// Create context gatherer
			gatherer := NewContextGatherer(mockLogger, mockConsoleWriter, true, mockClient, mockLogger) // dryRun=true

			// Execute the function
			ctx := context.Background()
			err := gatherer.DisplayDryRunInfo(ctx, tt.stats)

			// Should not return an error
			if err != nil {
				t.Errorf("DisplayDryRunInfo() unexpected error: %v", err)
				return
			}

			// Check that expected messages are in the console output, not the logger
			for _, expectedMsg := range tt.expectedLogMessages {
				found := false
				for _, actualMsg := range mockConsoleWriter.messages {
					if strings.Contains(actualMsg, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected console message %q not found. All messages: %v", expectedMsg, mockConsoleWriter.messages)
				}
			}

			// Verify that console messages are used for user output, not structured logs
			logMessages := mockLogger.GetInfoMessages()
			for _, msg := range logMessages {
				// The structured logs should only have debug/info about the operation,
				// not the actual dry run output shown to users
				if strings.Contains(msg, "Files that would be included") ||
					strings.Contains(msg, "Context statistics:") ||
					strings.Contains(msg, "To generate content") {
					t.Errorf("User-facing message should not appear in structured logs: %s", msg)
				}
			}
		})
	}
}

func TestGatherContext_AuditLogging(t *testing.T) {
	// Test that audit logging works properly with mock logger error simulation
	mockLogger := testutil.NewMockLogger()
	mockClient := &llm.MockLLMClient{}

	// Set up a temporary directory with a test file
	tempDir := testutil.SetupTempDir(t, "audit-test-")
	testutil.CreateTestFile(t, tempDir, "test.go", []byte("package main"))

	// Create context gatherer
	mockConsoleWriter := &mockConsoleWriter{}
	gatherer := NewContextGatherer(mockLogger, mockConsoleWriter, false, mockClient, mockLogger)

	// Create config
	config := interfaces.GatherConfig{
		Paths:        []string{tempDir},
		Include:      "",
		Exclude:      "",
		ExcludeNames: "",
		Format:       "{path}\n{content}",
		Verbose:      false,
		LogLevel:     logutil.InfoLevel,
	}

	// Test with successful audit logging
	t.Run("successful audit logging", func(t *testing.T) {
		ctx := context.Background()
		files, stats, err := gatherer.GatherContext(ctx, config)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 file, got %d", len(files))
		}

		if stats == nil {
			t.Fatal("Expected stats to be non-nil")
		}

		// Verify audit calls were made
		auditCalls := mockLogger.GetLogOpCalls()
		if len(auditCalls) < 2 { // Should have InProgress and Success calls
			t.Errorf("Expected at least 2 audit calls, got %d", len(auditCalls))
		}

		// Verify both InProgress and Success audit entries
		hasInProgress := false
		hasSuccess := false
		for _, call := range auditCalls {
			if call.Operation == "GatherContext" {
				if call.Status == "InProgress" {
					hasInProgress = true
				}
				if call.Status == "Success" {
					hasSuccess = true
				}
			}
		}

		if !hasInProgress {
			t.Error("Expected InProgress audit log entry")
		}
		if !hasSuccess {
			t.Error("Expected Success audit log entry")
		}
	})

	// Test audit logging error handling
	t.Run("audit logging with errors", func(t *testing.T) {
		// Set an error to be returned by audit logging methods
		mockLogger.SetLogError(fmt.Errorf("audit error"))
		defer mockLogger.ClearLogError()

		ctx := context.Background()
		files, stats, err := gatherer.GatherContext(ctx, config)

		// The operation should still succeed despite audit logging errors
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 file, got %d", len(files))
		}

		if stats == nil {
			t.Fatal("Expected stats to be non-nil")
		}

		// Verify error messages were logged for audit failures
		errorMessages := mockLogger.GetErrorMessages()
		found := false
		for _, msg := range errorMessages {
			if strings.Contains(msg, "Failed to write audit log") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected error message about audit log failure")
		}
	})
}

func TestNewContextGatherer(t *testing.T) {
	mockLogger := testutil.NewMockLogger()
	mockConsoleWriter := &mockConsoleWriter{}
	mockClient := &llm.MockLLMClient{}

	gatherer := NewContextGatherer(mockLogger, mockConsoleWriter, true, mockClient, mockLogger)

	if gatherer == nil {
		t.Fatal("NewContextGatherer should not return nil")
	}

	// Verify the gatherer implements the interface
	var _ interfaces.ContextGatherer = gatherer
}
