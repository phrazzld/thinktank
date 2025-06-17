package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
)

// TestSynthesizeResultsWithFailedModels verifies that the orchestrator
// correctly excludes outputs from failed models when synthesizing results
func TestSynthesizeResultsWithFailedModels(t *testing.T) {
	// Create test configuration with synthesis model
	cfg := &config.CliConfig{
		ModelNames:     []string{"model1", "model2", "model3"},
		SynthesisModel: "synthesis-model",
	}

	// Create mocks for dependencies
	logger := &MockLogger{}
	mockAuditLogger := NewMockAuditLogger()

	// Create a mock API service that returns a specific synthesis result
	mockAPIService := &MockAPIService{}

	// Create an instance of the SynthesisService for testing
	synthesisService := NewSynthesisService(mockAPIService, mockAuditLogger, logger, cfg.SynthesisModel)

	// Setup tests with various scenarios of model outputs
	tests := []struct {
		name                 string
		instructions         string
		modelOutputs         map[string]string
		expectedFailedModels []string
		expectedError        bool
	}{
		{
			name:         "All models successful",
			instructions: "test instruction",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
				"model2": "Output from model 2",
				"model3": "Output from model 3",
			},
			expectedFailedModels: nil,
			expectedError:        false,
		},
		{
			name:         "One model failed",
			instructions: "test instruction",
			modelOutputs: map[string]string{
				"model1": "Output from model 1",
				"model3": "Output from model 3",
				// model2 is missing - simulating failure
			},
			expectedFailedModels: []string{"model2"},
			expectedError:        false,
		},
		{
			name:         "Multiple models failed",
			instructions: "test instruction",
			modelOutputs: map[string]string{
				"model3": "Output from model 3",
				// model1 and model2 are missing - simulating failures
			},
			expectedFailedModels: []string{"model1", "model2"},
			expectedError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call SynthesizeResults
			_, err := synthesisService.SynthesizeResults(context.Background(), tt.instructions, tt.modelOutputs)

			// Check error expectations
			if (err != nil) != tt.expectedError {
				t.Errorf("SynthesizeResults() error = %v, wantErr %v", err, tt.expectedError)
				return
			}

			// Validate that only available model outputs were included
			// by examining the inputs sent to the underlying service

			// Check which models were included in the synthesis
			for _, call := range mockAuditLogger.LogCalls {
				if call.Operation == "SynthesisStart" {
					if inputs, ok := call.Inputs["model_outputs"]; ok {
						// Check that the failed models are not in the inputs
						modelOutputMap, ok := inputs.(map[string]string)
						if !ok {
							t.Errorf("Expected model_outputs to be map[string]string, got %T", inputs)
							continue
						}

						// Check that all expected models are included
						for modelName := range tt.modelOutputs {
							if _, exists := modelOutputMap[modelName]; !exists {
								t.Errorf("Expected model %s to be included in synthesis inputs, but it wasn't", modelName)
							}
						}

						// Check that failed models are not included
						for _, failedModel := range tt.expectedFailedModels {
							if _, exists := modelOutputMap[failedModel]; exists {
								t.Errorf("Failed model %s was incorrectly included in synthesis inputs", failedModel)
							}
						}
					}
				}
			}
		})
	}
}

// TestProcessModelsToSynthesis tests the complete flow from processModels to synthesizeResults,
// ensuring that failed models are properly excluded throughout the entire process
func TestProcessModelsToSynthesis(t *testing.T) {
	// Create test cases
	tests := []struct {
		name                  string
		modelNames            []string
		failingModels         []string
		synthesisModel        string
		expectedModelCount    int
		expectSynthesisCalled bool
	}{
		{
			name:                  "All models succeed, synthesis succeeds",
			modelNames:            []string{"model1", "model2", "model3"},
			failingModels:         []string{},
			synthesisModel:        "synthesis-model",
			expectedModelCount:    3,
			expectSynthesisCalled: true,
		},
		{
			name:                  "Some models fail, synthesis proceeds with available outputs",
			modelNames:            []string{"model1", "model2", "model3", "model4"},
			failingModels:         []string{"model2", "model4"},
			synthesisModel:        "synthesis-model",
			expectedModelCount:    2,
			expectSynthesisCalled: true,
		},
		{
			name:                  "Most models fail, synthesis still proceeds with available output",
			modelNames:            []string{"model1", "model2", "model3"},
			failingModels:         []string{"model1", "model3"},
			synthesisModel:        "synthesis-model",
			expectedModelCount:    1,
			expectSynthesisCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock API service with controlled behavior
			mockAPIService := &CustomMockAPIService{
				failingModels: tt.failingModels,
			}

			// Create mocks for other dependencies
			mockContextGatherer := &MockContextGatherer{}
			mockFileWriter := &MockFileWriter{}
			mockRateLimiter := ratelimit.NewRateLimiter(0, 0)
			mockAuditLogger := NewMockAuditLogger()
			logger := &MockLogger{}

			// Create test configuration
			cfg := &config.CliConfig{
				ModelNames:     tt.modelNames,
				SynthesisModel: tt.synthesisModel,
			}

			// Create the orchestrator with specified dependencies
			consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return false }, // CI mode for tests
			})
			orch := NewOrchestrator(
				mockAPIService,
				mockContextGatherer,
				mockFileWriter,
				mockAuditLogger,
				mockRateLimiter,
				cfg,
				logger,
				consoleWriter,
			)

			// Create a custom mock synthesis service that tracks calls
			synthesisMock := &MockSynthesisService{
				synthesizeContent: "Synthesized content",
			}

			// Replace the synthesis service with our mock
			orch.synthesisService = synthesisMock

			// Execute the orchestrator
			err := orch.Run(context.Background(), "test instructions")

			// For partial failures, we expect a specific error type
			if len(tt.failingModels) > 0 && len(tt.failingModels) < len(tt.modelNames) {
				if !errors.Is(err, ErrPartialProcessingFailure) {
					t.Errorf("Expected ErrPartialProcessingFailure, got %v", err)
				}
			} else if len(tt.failingModels) == len(tt.modelNames) {
				// If all models fail, we should get a different error
				if !errors.Is(err, ErrAllProcessingFailed) {
					t.Errorf("Expected ErrAllProcessingFailed, got %v", err)
				}
			}

			// Check if synthesis was called when expected
			if tt.expectSynthesisCalled && synthesisMock.capturedOutputs == nil {
				t.Errorf("Expected synthesis to be called, but it wasn't")
			}

			// If synthesis was called, check that only non-failing models were included
			if synthesisMock.capturedOutputs != nil {
				// Verify expected model count
				if len(synthesisMock.capturedOutputs) != tt.expectedModelCount {
					t.Errorf("Expected %d model outputs, got %d", tt.expectedModelCount, len(synthesisMock.capturedOutputs))
				}

				// Check that failed models are not included
				for _, failingModel := range tt.failingModels {
					if _, exists := synthesisMock.capturedOutputs[failingModel]; exists {
						t.Errorf("Failed model %s was incorrectly included in synthesis inputs", failingModel)
					}
				}

				// Check that all non-failing models are included
				for _, modelName := range tt.modelNames {
					isFailingModel := false
					for _, failingModel := range tt.failingModels {
						if modelName == failingModel {
							isFailingModel = true
							break
						}
					}

					if !isFailingModel {
						if _, exists := synthesisMock.capturedOutputs[modelName]; !exists {
							t.Errorf("Expected model %s to be included in synthesis, but it wasn't", modelName)
						}
					}
				}
			}
		})
	}
}

// CustomMockAPIService extends MockAPIService to simulate model failures
type CustomMockAPIService struct {
	MockAPIService
	failingModels []string
}

// InitLLMClient overrides the mock to simulate failures for specific models
func (m *CustomMockAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// Check if this is a model that should fail
	for _, failingModel := range m.failingModels {
		if modelName == failingModel {
			return nil, errors.New("simulated failure for model " + modelName)
		}
	}

	// For non-failing models, return a working mock client
	return &MockLLMClient{}, nil
}
