package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestHandleProcessingOutcome tests the handleProcessingOutcome function
func TestHandleProcessingOutcome(t *testing.T) {
	tests := []struct {
		name           string
		processingErr  error
		fileSaveErr    error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:          "no errors",
			processingErr: nil,
			fileSaveErr:   nil,
			expectError:   false,
		},
		{
			name:           "processing error only",
			processingErr:  errors.New("model processing failed"),
			fileSaveErr:    nil,
			expectError:    true,
			expectedErrMsg: "model processing errors occurred",
		},
		{
			name:           "file save error only",
			processingErr:  nil,
			fileSaveErr:    errors.New("file save failed"),
			expectError:    true,
			expectedErrMsg: "file save operation failed",
		},
		{
			name:           "both processing and file save errors",
			processingErr:  errors.New("model processing failed"),
			fileSaveErr:    errors.New("file save failed"),
			expectError:    true,
			expectedErrMsg: "model processing errors and file save errors occurred",
		},
		{
			name:           "processing error with LLM category",
			processingErr:  llm.New("test", "AUTH_ERR", 401, "Authentication failed", "req123", errors.New("auth error"), llm.CategoryAuth),
			fileSaveErr:    nil,
			expectError:    true,
			expectedErrMsg: "model processing errors occurred",
		},
		{
			name:           "file save error with processing error",
			processingErr:  llm.New("test", "NET_ERR", 0, "Network error", "req456", errors.New("network error"), llm.CategoryNetwork),
			fileSaveErr:    errors.New("disk full"),
			expectError:    true,
			expectedErrMsg: "model processing errors and file save errors occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create orchestrator with minimal dependencies
			ctx := context.Background()
			logger := testutil.NewMockLogger()
			auditLogger := NewMockAuditLogger()
			config := &config.CliConfig{}

			o := &Orchestrator{
				logger:      logger,
				auditLogger: auditLogger,
				config:      config,
			}

			err := o.handleProcessingOutcome(ctx, tt.processingErr, tt.fileSaveErr, logger)

			if tt.expectError {
				if err == nil {
					t.Errorf("handleProcessingOutcome() error = nil, want error")
					return
				}
				if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("handleProcessingOutcome() error = %q, want to contain %q", err.Error(), tt.expectedErrMsg)
				}
			} else {
				if err != nil {
					t.Errorf("handleProcessingOutcome() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestHandleOutputFlow tests the handleOutputFlow function
func TestHandleOutputFlow(t *testing.T) {
	tests := []struct {
		name                string
		synthesisModel      string
		modelOutputs        map[string]string
		individualFlowError error
		synthesisFlowError  error
		expectedSynthesis   string
		expectedIndividual  map[string]string
		expectError         bool
	}{
		{
			name:           "no synthesis model - individual output flow",
			synthesisModel: "",
			modelOutputs: map[string]string{
				"model1": "output1",
				"model2": "output2",
			},
			individualFlowError: nil,
			expectedIndividual: map[string]string{
				"model1": "/test/output1.md",
				"model2": "/test/output2.md",
			},
			expectError: false,
		},
		{
			name:           "no synthesis model - individual flow fails",
			synthesisModel: "",
			modelOutputs: map[string]string{
				"model1": "output1",
			},
			individualFlowError: errors.New("individual flow failed"),
			expectError:         true,
		},
		{
			name:           "synthesis model - success",
			synthesisModel: "synthesis-model",
			modelOutputs: map[string]string{
				"model1": "output1",
				"model2": "output2",
			},
			synthesisFlowError: nil,
			expectedSynthesis:  "/test/synthesis.md",
			expectError:        false,
		},
		{
			name:           "synthesis model - synthesis fails, fallback to individual",
			synthesisModel: "synthesis-model",
			modelOutputs: map[string]string{
				"model1": "output1",
				"model2": "output2",
			},
			synthesisFlowError:  errors.New("synthesis failed"),
			individualFlowError: nil,
			expectedIndividual: map[string]string{
				"model1": "/test/output1.md",
				"model2": "/test/output2.md",
			},
			expectError: true, // should return synthesis error even though fallback succeeded
		},
		{
			name:           "synthesis model - both synthesis and fallback fail",
			synthesisModel: "synthesis-model",
			modelOutputs: map[string]string{
				"model1": "output1",
			},
			synthesisFlowError:  errors.New("synthesis failed"),
			individualFlowError: errors.New("fallback failed"),
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create orchestrator with all necessary dependencies
			ctx := context.Background()
			logger := testutil.NewMockLogger()
			auditLogger := NewMockAuditLogger()
			config := &config.CliConfig{
				SynthesisModel: tt.synthesisModel,
				OutputDir:      "/test",
			}

			// Create mock output writer and synthesis service
			mockOutputWriter := &TestOutputWriter{
				saveIndividualCount: len(tt.expectedIndividual),
				saveIndividualPaths: tt.expectedIndividual,
				saveIndividualError: tt.individualFlowError,
				saveSynthesisPath:   tt.expectedSynthesis,
				saveSynthesisError:  tt.synthesisFlowError,
			}

			var mockSynthesisService SynthesisService
			if tt.synthesisModel != "" {
				mockSynthesisService = &MockSynthesisService{
					synthesizeContent: "test synthesis content",
					synthesizeError:   tt.synthesisFlowError,
				}
			}

			consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return false }, // CI mode for tests
			})
			o := &Orchestrator{
				logger:           logger,
				auditLogger:      auditLogger,
				config:           config,
				consoleWriter:    consoleWriter,
				outputWriter:     mockOutputWriter,
				synthesisService: mockSynthesisService,
			}

			instructions := "test instructions"
			outputInfo, err := o.handleOutputFlow(ctx, instructions, tt.modelOutputs)

			if tt.expectError {
				if err == nil {
					t.Errorf("handleOutputFlow() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("handleOutputFlow() error = %v, want nil", err)
				}
			}

			// Verify outputInfo content
			if outputInfo == nil {
				t.Errorf("handleOutputFlow() outputInfo = nil, want non-nil")
				return
			}

			if tt.expectedSynthesis != "" && outputInfo.SynthesisFilePath != tt.expectedSynthesis {
				t.Errorf("handleOutputFlow() SynthesisFilePath = %q, want %q", outputInfo.SynthesisFilePath, tt.expectedSynthesis)
			}

			if tt.expectedIndividual != nil {
				if len(outputInfo.IndividualFilePaths) != len(tt.expectedIndividual) {
					t.Errorf("handleOutputFlow() IndividualFilePaths length = %d, want %d", len(outputInfo.IndividualFilePaths), len(tt.expectedIndividual))
				}
				for model, expectedPath := range tt.expectedIndividual {
					if actualPath, exists := outputInfo.IndividualFilePaths[model]; !exists {
						t.Errorf("handleOutputFlow() missing IndividualFilePaths[%q]", model)
					} else if actualPath != expectedPath {
						t.Errorf("handleOutputFlow() IndividualFilePaths[%q] = %q, want %q", model, actualPath, expectedPath)
					}
				}
			}
		})
	}
}
