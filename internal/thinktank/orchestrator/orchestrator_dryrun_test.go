package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// MockContextGathererWithConfigurableError allows controlling error responses
type MockContextGathererWithConfigurableError struct {
	displayDryRunError error
}

// GatherContext is a mock implementation
func (m *MockContextGathererWithConfigurableError) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return []fileutil.FileMeta{}, &interfaces.ContextStats{}, nil
}

// DisplayDryRunInfo returns the configured error or nil
func (m *MockContextGathererWithConfigurableError) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return m.displayDryRunError
}

// MockLoggerWithRecorder records log calls for verification
type MockLoggerWithRecorder struct {
	MockLogger
	infoMessages []string
}

// InfoContext records messages
func (m *MockLoggerWithRecorder) InfoContext(ctx context.Context, format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

// WithContext returns self (maintains recording)
func (m *MockLoggerWithRecorder) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
}

// TestRunDryRunFlow tests the runDryRunFlow method with various configurations
func TestRunDryRunFlow(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                string
		dryRunEnabled       bool
		displayDryRunError  error
		expectedExecution   bool
		expectError         bool
		expectDryRunInfoMsg bool
	}{
		{
			name:                "Not in dry run mode",
			dryRunEnabled:       false,
			displayDryRunError:  nil,
			expectedExecution:   false,
			expectError:         false,
			expectDryRunInfoMsg: false,
		},
		{
			name:                "Dry run mode - success",
			dryRunEnabled:       true,
			displayDryRunError:  nil,
			expectedExecution:   true,
			expectError:         false,
			expectDryRunInfoMsg: true,
		},
		{
			name:                "Dry run mode - error",
			dryRunEnabled:       true,
			displayDryRunError:  errors.New("mock display error"),
			expectedExecution:   true,
			expectError:         true,
			expectDryRunInfoMsg: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks with configured behaviors
			mockContextGatherer := &MockContextGathererWithConfigurableError{
				displayDryRunError: tt.displayDryRunError,
			}
			mockLogger := &MockLoggerWithRecorder{}

			// Create orchestrator with test configuration
			orch := &Orchestrator{
				contextGatherer: mockContextGatherer,
				logger:          mockLogger,
				config:          &config.CliConfig{DryRun: tt.dryRunEnabled},
			}

			// Create context statistics
			stats := &interfaces.ContextStats{
				ProcessedFilesCount: 10,
				CharCount:           1000,
				LineCount:           50,
				ProcessedFiles:      []string{"file1.go", "file2.go"},
			}

			// Call the method under test
			executed, err := orch.runDryRunFlow(context.Background(), stats)

			// Verify execution flag is correctly set
			if executed != tt.expectedExecution {
				t.Errorf("Expected executed to be %v, got %v", tt.expectedExecution, executed)
			}

			// Verify error behavior
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			// Verify logging behavior
			loggedDryRunMessage := false
			for _, msg := range mockLogger.infoMessages {
				if msg == "Running in dry-run mode" {
					loggedDryRunMessage = true
					break
				}
			}

			if loggedDryRunMessage != tt.expectDryRunInfoMsg {
				t.Errorf("Expected dry run info message: %v, got: %v", tt.expectDryRunInfoMsg, loggedDryRunMessage)
			}
		})
	}
}
