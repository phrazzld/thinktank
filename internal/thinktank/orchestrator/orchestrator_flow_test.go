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
			expectedIndividual: map[string]string{
				"model1": "/test/output1.md",
				"model2": "/test/output2.md",
			},
			expectError: false,
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

// TestHandleOutputFlow_DualRoleModel tests the critical edge case where a model
// serves as both an individual processor AND the synthesis model. This test
// prevents regression where the same model name could cause routing confusion.
func TestHandleOutputFlow_DualRoleModel(t *testing.T) {
	tests := []struct {
		name                string
		modelOutputs        map[string]string
		synthesisModel      string
		individualFlowError error
		synthesisFlowError  error
		expectedIndividual  map[string]string
		expectedSynthesis   string
		expectError         bool
		expectedErrorType   string
	}{
		{
			name: "dual role model - both individual and synthesis succeed",
			modelOutputs: map[string]string{
				"gpt-5.2":        "Individual analysis from gpt-5.2",
				"claude-3":       "Analysis from claude-3",
				"gemini-3-flash": "Analysis from gemini-3-flash",
			},
			synthesisModel: "gpt-5.2", // Same model serves dual role
			expectedIndividual: map[string]string{
				"gpt-5.2":        "/test/gpt-5.2.md",
				"claude-3":       "/test/claude-3.md",
				"gemini-3-flash": "/test/gemini-3-flash.md",
			},
			expectedSynthesis: "/test/gpt-5.2-synthesis.md",
			expectError:       false,
		},
		{
			name: "dual role model - individual succeeds, synthesis fails",
			modelOutputs: map[string]string{
				"gpt-5.2":  "Individual analysis from gpt-5.2",
				"claude-3": "Analysis from claude-3",
			},
			synthesisModel:      "gpt-5.2",
			synthesisFlowError:  errors.New("synthesis rate limit exceeded"),
			individualFlowError: nil,
			expectedIndividual: map[string]string{
				"gpt-5.2":  "/test/gpt-5.2.md",
				"claude-3": "/test/claude-3.md",
			},
			expectedSynthesis: "",
			expectError:       true,
			expectedErrorType: "synthesis",
		},
		{
			name: "dual role model - individual fails, synthesis succeeds",
			modelOutputs: map[string]string{
				"gpt-5.2":  "Individual analysis from gpt-5.2",
				"claude-3": "Analysis from claude-3",
			},
			synthesisModel:      "gpt-5.2",
			individualFlowError: errors.New("individual save failed"),
			synthesisFlowError:  nil,
			expectedSynthesis:   "/test/gpt-5.2-synthesis.md",
			expectError:         true,
			expectedErrorType:   "individual",
		},
		{
			name: "single model serves dual role",
			modelOutputs: map[string]string{
				"gpt-5.2": "Analysis from gpt-5.2",
			},
			synthesisModel: "gpt-5.2",
			expectedIndividual: map[string]string{
				"gpt-5.2": "/test/gpt-5.2.md",
			},
			expectedSynthesis: "/test/gpt-5.2-synthesis.md",
			expectError:       false,
		},
		{
			name: "dual role model - both individual and synthesis fail",
			modelOutputs: map[string]string{
				"gpt-5.2":  "Individual analysis from gpt-5.2",
				"claude-3": "Analysis from claude-3",
			},
			synthesisModel:      "gpt-5.2",
			individualFlowError: errors.New("individual save failed"),
			synthesisFlowError:  errors.New("synthesis failed"),
			expectError:         true,
			expectedErrorType:   "synthesis", // synthesis error takes precedence
		},
		{
			name: "synthesis model not in individual outputs",
			modelOutputs: map[string]string{
				"claude-3":       "Analysis from claude-3",
				"gemini-3-flash": "Analysis from gemini-3-flash",
			},
			synthesisModel: "gpt-5.2", // Different from individual models
			expectedIndividual: map[string]string{
				"claude-3":       "/test/claude-3.md",
				"gemini-3-flash": "/test/gemini-3-flash.md",
			},
			expectedSynthesis: "/test/gpt-5.2-synthesis.md",
			expectError:       false,
		},
		{
			name:              "empty model outputs with synthesis model",
			modelOutputs:      map[string]string{}, // Empty
			synthesisModel:    "gpt-5.2",
			expectedSynthesis: "", // No synthesis when no model outputs
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create orchestrator with test dependencies
			ctx := context.Background()
			logger := testutil.NewMockLogger()
			auditLogger := NewMockAuditLogger()
			config := &config.CliConfig{
				SynthesisModel: tt.synthesisModel,
				OutputDir:      "/test",
			}

			// Create enhanced mock output writer
			mockOutputWriter := &TestOutputWriter{
				saveIndividualCount: len(tt.expectedIndividual),
				saveIndividualPaths: tt.expectedIndividual,
				saveIndividualError: tt.individualFlowError,
				saveSynthesisPath:   tt.expectedSynthesis,
				saveSynthesisError:  tt.synthesisFlowError,
			}

			// Create mock synthesis service
			var mockSynthesisService SynthesisService
			if tt.synthesisModel != "" {
				mockSynthesisService = &MockSynthesisService{
					synthesizeContent: "Synthesized content combining all model outputs",
					synthesizeError:   tt.synthesisFlowError,
				}
			}

			consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return false }, // CI mode for tests
			})

			orchestrator := &Orchestrator{
				logger:           logger,
				auditLogger:      auditLogger,
				config:           config,
				consoleWriter:    consoleWriter,
				outputWriter:     mockOutputWriter,
				synthesisService: mockSynthesisService,
			}

			// Execute the method under test
			instructions := "Analyze the provided data"
			outputInfo, err := orchestrator.handleOutputFlow(ctx, instructions, tt.modelOutputs)

			// Verify error expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("handleOutputFlow() expected error but got nil")
				} else {
					// Verify error type if specified
					if tt.expectedErrorType != "" {
						if tt.expectedErrorType == "synthesis" && !strings.Contains(err.Error(), "synthesis") {
							t.Errorf("Expected synthesis error, got: %v", err)
						}
						if tt.expectedErrorType == "individual" && !strings.Contains(err.Error(), "individual") {
							t.Errorf("Expected individual error, got: %v", err)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("handleOutputFlow() unexpected error: %v", err)
				}
			}

			// Verify outputInfo is not nil
			if outputInfo == nil {
				t.Fatalf("handleOutputFlow() outputInfo = nil, want non-nil")
			}

			// Verify individual file paths
			if tt.expectedIndividual != nil {
				if len(outputInfo.IndividualFilePaths) != len(tt.expectedIndividual) {
					t.Errorf("IndividualFilePaths length = %d, want %d",
						len(outputInfo.IndividualFilePaths), len(tt.expectedIndividual))
				}

				for modelName, expectedPath := range tt.expectedIndividual {
					if actualPath, exists := outputInfo.IndividualFilePaths[modelName]; !exists {
						t.Errorf("Missing IndividualFilePaths[%q]", modelName)
					} else if actualPath != expectedPath {
						t.Errorf("IndividualFilePaths[%q] = %q, want %q",
							modelName, actualPath, expectedPath)
					}
				}

				// CRITICAL: Verify dual-role model appears in individual outputs
				if tt.synthesisModel != "" {
					if _, exists := tt.expectedIndividual[tt.synthesisModel]; exists {
						if _, actualExists := outputInfo.IndividualFilePaths[tt.synthesisModel]; !actualExists {
							t.Errorf("Dual-role model %q missing from IndividualFilePaths", tt.synthesisModel)
						}
					}
				}
			}

			// Verify synthesis file path
			if tt.expectedSynthesis != "" {
				if outputInfo.SynthesisFilePath != tt.expectedSynthesis {
					t.Errorf("SynthesisFilePath = %q, want %q",
						outputInfo.SynthesisFilePath, tt.expectedSynthesis)
				}
			} else {
				if outputInfo.SynthesisFilePath != "" {
					t.Errorf("SynthesisFilePath = %q, want empty", outputInfo.SynthesisFilePath)
				}
			}

			// CRITICAL: For dual-role scenarios, verify both outputs exist
			if tt.synthesisModel != "" && tt.expectedSynthesis != "" {
				// Check that synthesis model has both individual AND synthesis outputs
				if individualPath, exists := outputInfo.IndividualFilePaths[tt.synthesisModel]; exists {
					synthesisPath := outputInfo.SynthesisFilePath
					if individualPath == synthesisPath {
						t.Errorf("Dual-role model %q has same path for individual and synthesis: %q",
							tt.synthesisModel, individualPath)
					}

					// Verify different file extensions/naming
					if !strings.Contains(synthesisPath, "synthesis") {
						t.Errorf("Synthesis path %q should contain 'synthesis' to distinguish from individual path %q",
							synthesisPath, individualPath)
					}
				}
			}
		})
	}
}
