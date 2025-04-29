package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// MockSynthesisService allows controlling synthesis service behavior for testing
type MockSynthesisService struct {
	synthesizeError      error
	synthesizeContent    string
	capturedOutputs      map[string]string
	capturedInstructions string
}

// SynthesizeResults is a mock implementation that returns configured results
func (m *MockSynthesisService) SynthesizeResults(ctx context.Context, instructions string, modelOutputs map[string]string) (string, error) {
	m.capturedInstructions = instructions
	m.capturedOutputs = modelOutputs
	return m.synthesizeContent, m.synthesizeError
}

// MockSynthesisOutputWriter extends MockOutputWriter to record synthesis save calls
type MockSynthesisOutputWriter struct {
	MockOutputWriter
	saveSynthesisError     error
	capturedContent        string
	capturedSynthesisModel string
	saveSynthesisCalled    bool
}

// SaveSynthesisOutput is a mock implementation that captures parameters and returns configured error
func (m *MockSynthesisOutputWriter) SaveSynthesisOutput(ctx context.Context, content string, modelName string, outputDir string) error {
	m.capturedContent = content
	m.capturedSynthesisModel = modelName
	m.capturedOutputDir = outputDir
	m.saveSynthesisCalled = true
	return m.saveSynthesisError
}

// MockLoggerWithSynthesisRecorder records log calls for verification
type MockLoggerWithSynthesisRecorder struct {
	MockLoggerWithOutputRecorder
	warnMessages []string
}

// WarnContext records warn messages
func (m *MockLoggerWithSynthesisRecorder) WarnContext(ctx context.Context, format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, fmt.Sprintf(format, args...))
}

// WithContext returns self (maintains recording)
func (m *MockLoggerWithSynthesisRecorder) WithContext(ctx context.Context) logutil.LoggerInterface {
	return m
}

// TestRunSynthesisFlow tests the runSynthesisFlow method with various scenarios
func TestRunSynthesisFlow(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                      string
		instructions              string
		modelOutputs              map[string]string
		outputDir                 string
		synthesisModel            string
		synthesizeContent         string
		synthesizeError           error
		saveSynthesisError        error
		expectError               bool
		expectSynthesisCalled     bool
		expectSaveSynthesisCalled bool
		expectSuccessLogCount     int
		expectErrorLogCount       int
		expectWarnLogCount        int
	}{
		{
			name:         "Successful synthesis and save",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model-1": "Content for model 1",
				"model-2": "Content for model 2",
			},
			outputDir:                 "/test/output/dir",
			synthesisModel:            "gpt-4",
			synthesizeContent:         "Synthesized content combining model 1 and model 2 outputs",
			synthesizeError:           nil,
			saveSynthesisError:        nil,
			expectError:               false,
			expectSynthesisCalled:     true,
			expectSaveSynthesisCalled: true,
			expectSuccessLogCount:     1, // Successfully saved synthesis output
			expectErrorLogCount:       0,
			expectWarnLogCount:        0,
		},
		{
			name:         "Synthesis error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model-1": "Content for model 1",
				"model-2": "Content for model 2",
			},
			outputDir:                 "/test/output/dir",
			synthesisModel:            "gpt-4",
			synthesizeContent:         "",
			synthesizeError:           errors.New("synthesis API error"),
			saveSynthesisError:        nil,
			expectError:               true,
			expectSynthesisCalled:     true,
			expectSaveSynthesisCalled: false, // Save should not be called on synthesis error
			expectSuccessLogCount:     0,
			expectErrorLogCount:       1, // Error log for synthesis failure
			expectWarnLogCount:        0,
		},
		{
			name:         "Save error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model-1": "Content for model 1",
				"model-2": "Content for model 2",
			},
			outputDir:                 "/test/output/dir",
			synthesisModel:            "gpt-4",
			synthesizeContent:         "Synthesized content combining model 1 and model 2 outputs",
			synthesizeError:           nil,
			saveSynthesisError:        errors.New("failed to save synthesis output"),
			expectError:               true,
			expectSynthesisCalled:     true,
			expectSaveSynthesisCalled: true,
			expectSuccessLogCount:     0, // No success logs when save fails
			expectErrorLogCount:       1, // Error log for save failure
			expectWarnLogCount:        0,
		},
		{
			name:                      "Empty model outputs",
			instructions:              "Test instructions",
			modelOutputs:              map[string]string{},
			outputDir:                 "/test/output/dir",
			synthesisModel:            "gpt-4",
			synthesizeContent:         "",
			synthesizeError:           nil,
			saveSynthesisError:        nil,
			expectError:               false,
			expectSynthesisCalled:     false, // Synthesis should not be called with empty outputs
			expectSaveSynthesisCalled: false,
			expectSuccessLogCount:     0,
			expectErrorLogCount:       0,
			expectWarnLogCount:        1, // Warning about no outputs available
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks with configured behaviors
			mockSynthesisService := &MockSynthesisService{
				synthesizeError:   tt.synthesizeError,
				synthesizeContent: tt.synthesizeContent,
			}

			mockOutputWriter := &MockSynthesisOutputWriter{
				saveSynthesisError: tt.saveSynthesisError,
			}

			mockLogger := &MockLoggerWithSynthesisRecorder{}

			// Create orchestrator with test configuration
			orch := &Orchestrator{
				synthesisService: mockSynthesisService,
				outputWriter:     mockOutputWriter,
				logger:           mockLogger,
				config: &config.CliConfig{
					OutputDir:      tt.outputDir,
					SynthesisModel: tt.synthesisModel,
				},
			}

			// Call the method under test
			err := orch.runSynthesisFlow(context.Background(), tt.instructions, tt.modelOutputs)

			// Verify error behavior
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			// Verify synthesis service was called (or not) as expected
			if tt.expectSynthesisCalled {
				if mockSynthesisService.capturedOutputs == nil {
					t.Errorf("Expected SynthesizeResults to be called, but it wasn't")
				}

				// Verify correct instructions and model outputs were passed
				if mockSynthesisService.capturedInstructions != tt.instructions {
					t.Errorf("Expected instructions %s, got %s", tt.instructions, mockSynthesisService.capturedInstructions)
				}

				if len(mockSynthesisService.capturedOutputs) != len(tt.modelOutputs) {
					t.Errorf("Expected %d model outputs, got %d", len(tt.modelOutputs), len(mockSynthesisService.capturedOutputs))
				}
			} else if mockSynthesisService.capturedOutputs != nil {
				t.Errorf("Expected SynthesizeResults not to be called, but it was")
			}

			// Verify output writer was called (or not) as expected
			if tt.expectSaveSynthesisCalled {
				if !mockOutputWriter.saveSynthesisCalled {
					t.Errorf("Expected SaveSynthesisOutput to be called, but it wasn't")
				}

				// Verify correct content and model name were passed
				if mockOutputWriter.capturedContent != tt.synthesizeContent {
					t.Errorf("Expected content %s, got %s", tt.synthesizeContent, mockOutputWriter.capturedContent)
				}

				if mockOutputWriter.capturedSynthesisModel != tt.synthesisModel {
					t.Errorf("Expected synthesis model %s, got %s", tt.synthesisModel, mockOutputWriter.capturedSynthesisModel)
				}

				if mockOutputWriter.capturedOutputDir != tt.outputDir {
					t.Errorf("Expected output dir %s, got %s", tt.outputDir, mockOutputWriter.capturedOutputDir)
				}
			} else if mockOutputWriter.saveSynthesisCalled {
				t.Errorf("Expected SaveSynthesisOutput not to be called, but it was")
			}

			// Verify logging behavior

			// Check synthesis start log
			syntesisStartLogCount := 0
			for _, msg := range mockLogger.infoMessages {
				if msg == fmt.Sprintf("Processing completed, synthesizing results with model: %s", tt.synthesisModel) {
					syntesisStartLogCount++
				}
			}
			if (len(tt.modelOutputs) > 0) && syntesisStartLogCount == 0 {
				t.Errorf("Expected synthesis start log message, but didn't find it")
			}

			// Check success logs
			successLogCount := 0
			for _, msg := range mockLogger.infoMessages {
				if msg == "Successfully saved synthesis output" {
					successLogCount++
				}
			}
			if successLogCount != tt.expectSuccessLogCount {
				t.Errorf("Expected %d success log messages, got %d", tt.expectSuccessLogCount, successLogCount)
			}

			// Check error logs
			synthesisErrorLogCount := 0
			saveErrorLogCount := 0
			for _, msg := range mockLogger.errorMessages {
				if msg == fmt.Sprintf("Synthesis failed: %v", tt.synthesizeError) {
					synthesisErrorLogCount++
				}
				if msg == fmt.Sprintf("Failed to save synthesis output: %v", tt.saveSynthesisError) {
					saveErrorLogCount++
				}
			}

			// Verify error log count matches expected
			totalErrorLogs := synthesisErrorLogCount + saveErrorLogCount
			if totalErrorLogs != tt.expectErrorLogCount {
				t.Errorf("Expected %d error log messages, got %d", tt.expectErrorLogCount, totalErrorLogs)
			}

			// Check warning logs
			warnLogCount := 0
			for _, msg := range mockLogger.warnMessages {
				if msg == "No model outputs available for synthesis" {
					warnLogCount++
				}
			}
			if warnLogCount != tt.expectWarnLogCount {
				t.Errorf("Expected %d warn log messages, got %d", tt.expectWarnLogCount, warnLogCount)
			}

			// Verify debug logs for model output count are present when appropriate
			if len(tt.modelOutputs) > 0 {
				debugLogFound := false
				for _, msg := range mockLogger.debugMessages {
					if msg == fmt.Sprintf("Synthesizing %d model outputs", len(tt.modelOutputs)) {
						debugLogFound = true
						break
					}
				}
				if !debugLogFound {
					t.Errorf("Expected debug log message about synthesizing model outputs count")
				}
			}
		})
	}
}
