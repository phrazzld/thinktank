package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// MockOutputWriter allows controlling output writer behavior for testing
type MockOutputWriter struct {
	savedCount           int
	saveError            error
	capturedModelOutputs map[string]string
	capturedOutputDir    string
}

// SaveIndividualOutputs is a mock implementation that returns configured results
func (m *MockOutputWriter) SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, error) {
	m.capturedModelOutputs = modelOutputs
	m.capturedOutputDir = outputDir
	return m.savedCount, m.saveError
}

// SaveSynthesisOutput is a mock implementation (not used in these tests)
func (m *MockOutputWriter) SaveSynthesisOutput(ctx context.Context, content string, modelName string, outputDir string) error {
	return nil
}

// MockLoggerWithOutputRecorder records log calls for verification
type MockLoggerWithOutputRecorder struct {
	MockLogger
	infoMessages  []string
	errorMessages []string
	debugMessages []string
}

// InfoContext records info messages
func (m *MockLoggerWithOutputRecorder) InfoContext(ctx context.Context, format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(format, args...))
}

// ErrorContext records error messages
func (m *MockLoggerWithOutputRecorder) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

// DebugContext records debug messages
func (m *MockLoggerWithOutputRecorder) DebugContext(ctx context.Context, format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}

// WithContext returns self (maintains recording)
func (m *MockLoggerWithOutputRecorder) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
}

// TestRunIndividualOutputFlow tests the runIndividualOutputFlow method with various scenarios
func TestRunIndividualOutputFlow(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                  string
		modelOutputs          map[string]string
		outputDir             string
		savedCount            int
		saveError             error
		expectError           bool
		expectSuccessLogCount int
		expectErrorLogCount   int
	}{
		{
			name: "All outputs saved successfully",
			modelOutputs: map[string]string{
				"model-1": "Content for model 1",
				"model-2": "Content for model 2",
				"model-3": "Content for model 3",
			},
			outputDir:             "/test/output/dir",
			savedCount:            3,
			saveError:             nil,
			expectError:           false,
			expectSuccessLogCount: 1,
			expectErrorLogCount:   0,
		},
		{
			name: "Some outputs failed to save",
			modelOutputs: map[string]string{
				"model-1": "Content for model 1",
				"model-2": "Content for model 2",
			},
			outputDir:             "/test/output/dir",
			savedCount:            1, // Only one saved, indicating an error
			saveError:             errors.New("failed to save some files"),
			expectError:           true,
			expectSuccessLogCount: 0,
			expectErrorLogCount:   1,
		},
		{
			name:                  "Empty model outputs",
			modelOutputs:          map[string]string{},
			outputDir:             "/test/output/dir",
			savedCount:            0,
			saveError:             nil,
			expectError:           false,
			expectSuccessLogCount: 1,
			expectErrorLogCount:   0,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks with configured behaviors
			mockOutputWriter := &MockOutputWriter{
				savedCount: tt.savedCount,
				saveError:  tt.saveError,
			}
			mockLogger := &MockLoggerWithOutputRecorder{}

			// Create orchestrator with test configuration
			orch := &Orchestrator{
				outputWriter: mockOutputWriter,
				logger:       mockLogger,
				config:       &config.CliConfig{OutputDir: tt.outputDir},
			}

			// Call the method under test
			err := orch.runIndividualOutputFlow(context.Background(), tt.modelOutputs)

			// Verify error behavior
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			// Verify the output writer was called with correct parameters
			if mockOutputWriter.capturedOutputDir != tt.outputDir {
				t.Errorf("Expected output dir %s, got %s", tt.outputDir, mockOutputWriter.capturedOutputDir)
			}

			// Verify model outputs were passed correctly
			if len(mockOutputWriter.capturedModelOutputs) != len(tt.modelOutputs) {
				t.Errorf("Expected %d model outputs, got %d", len(tt.modelOutputs), len(mockOutputWriter.capturedModelOutputs))
			}

			// Verify logging behavior for success case
			successLogCount := 0
			for _, msg := range mockLogger.infoMessages {
				if msg == fmt.Sprintf("All %d model outputs saved successfully", tt.savedCount) {
					successLogCount++
				}
			}
			if successLogCount != tt.expectSuccessLogCount {
				t.Errorf("Expected %d success log messages, got %d", tt.expectSuccessLogCount, successLogCount)
			}

			// Verify logging behavior for error case
			errorLogCount := 0
			for _, msg := range mockLogger.errorMessages {
				if len(tt.modelOutputs) > 0 && msg == fmt.Sprintf("Completed with errors: %d files saved successfully, %d files failed",
					tt.savedCount, len(tt.modelOutputs)-tt.savedCount) {
					errorLogCount++
				}
			}
			if errorLogCount != tt.expectErrorLogCount {
				t.Errorf("Expected %d error log messages, got %d", tt.expectErrorLogCount, errorLogCount)
			}

			// Verify the debug log message for the model output count
			debugLogFound := false
			for _, msg := range mockLogger.debugMessages {
				if msg == fmt.Sprintf("Collected %d model outputs", len(tt.modelOutputs)) {
					debugLogFound = true
					break
				}
			}
			if !debugLogFound && len(tt.modelOutputs) > 0 {
				t.Errorf("Expected debug log message about collected model outputs count")
			}
		})
	}
}
