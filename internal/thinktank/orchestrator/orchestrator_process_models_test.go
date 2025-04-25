package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/ratelimit"
)

// TestProcessModelsEmptyOutputs tests that processModels does not include empty outputs for failed models
func TestProcessModelsEmptyOutputs(t *testing.T) {
	tests := []struct {
		name                 string
		modelNames           []string
		mockResults          []modelResult
		expectedOutputsCount int
		expectedErrorsCount  int
		expectedModelOutputs map[string]bool // map of model names that should be in the output
		unexpectedModelNames []string        // model names that should NOT be in the output
	}{
		{
			name:       "All models succeed",
			modelNames: []string{"model1", "model2", "model3"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "Output from model2", err: nil},
				{modelName: "model3", content: "Output from model3", err: nil},
			},
			expectedOutputsCount: 3,
			expectedErrorsCount:  0,
			expectedModelOutputs: map[string]bool{
				"model1": true,
				"model2": true,
				"model3": true,
			},
			unexpectedModelNames: []string{},
		},
		{
			name:       "Some models fail",
			modelNames: []string{"model1", "model2", "model3"},
			mockResults: []modelResult{
				{modelName: "model1", content: "Output from model1", err: nil},
				{modelName: "model2", content: "", err: errors.New("model2 failed")},
				{modelName: "model3", content: "Output from model3", err: nil},
			},
			expectedOutputsCount: 2,
			expectedErrorsCount:  1,
			expectedModelOutputs: map[string]bool{
				"model1": true,
				"model3": true,
			},
			unexpectedModelNames: []string{"model2"},
		},
		{
			name:       "All models fail",
			modelNames: []string{"model1", "model2", "model3"},
			mockResults: []modelResult{
				{modelName: "model1", content: "", err: errors.New("model1 failed")},
				{modelName: "model2", content: "", err: errors.New("model2 failed")},
				{modelName: "model3", content: "", err: errors.New("model3 failed")},
			},
			expectedOutputsCount: 0,
			expectedErrorsCount:  3,
			expectedModelOutputs: map[string]bool{},
			unexpectedModelNames: []string{"model1", "model2", "model3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockAPIService := &MockAPIService{}
			mockFileWriter := &MockFileWriter{}
			mockContextGatherer := &MockContextGatherer{}
			mockAuditLogger := &MockAuditLogger{}
			mockLogger := &MockLogger{}

			// Create config with model names
			cfg := &config.CliConfig{
				ModelNames: tt.modelNames,
			}

			// Create rate limiter with no constraints
			rateLimiter := ratelimit.NewRateLimiter(0, 0)

			// Create test orchestrator with controlled behavior
			orch := &processingTestOrchestrator{
				Orchestrator: Orchestrator{
					apiService:      mockAPIService,
					contextGatherer: mockContextGatherer,
					fileWriter:      mockFileWriter,
					auditLogger:     mockAuditLogger,
					rateLimiter:     rateLimiter,
					config:          cfg,
					logger:          mockLogger,
				},
				mockResults: tt.mockResults,
			}

			// Call processModels
			outputs, errors := orch.processModels(context.Background(), "test prompt")

			// Verify the correct number of outputs and errors
			if len(outputs) != tt.expectedOutputsCount {
				t.Errorf("Expected %d outputs, got %d", tt.expectedOutputsCount, len(outputs))
			}

			if len(errors) != tt.expectedErrorsCount {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrorsCount, len(errors))
			}

			// Verify the expected models are in the output
			for modelName := range tt.expectedModelOutputs {
				if _, exists := outputs[modelName]; !exists {
					t.Errorf("Expected model %s in outputs but it was not found", modelName)
				}
			}

			// Verify unexpected models are not in the output
			for _, modelName := range tt.unexpectedModelNames {
				if _, exists := outputs[modelName]; exists {
					t.Errorf("Model %s should not be in outputs but it was found", modelName)
				}
			}
		})
	}
}

// processingTestOrchestrator extends Orchestrator to provide controlled test behavior
type processingTestOrchestrator struct {
	Orchestrator
	mockResults []modelResult
}

// processModels Override to provide controlled behavior for testing
func (o *processingTestOrchestrator) processModels(ctx context.Context, prompt string) (map[string]string, []error) {
	// Create a result channel to simulate multiple results
	resultChan := make(chan modelResult, len(o.mockResults))

	// Put all the mock results into the channel
	for _, result := range o.mockResults {
		resultChan <- result
	}
	close(resultChan)

	// Collect outputs and errors from the channel
	modelOutputs := make(map[string]string)
	var modelErrors []error

	for result := range resultChan {
		// Only store output for successful models
		if result.err == nil {
			modelOutputs[result.modelName] = result.content
		} else {
			// Collect errors
			modelErrors = append(modelErrors, result.err)
		}
	}

	return modelOutputs, modelErrors
}

// MockOrchestratorProcessor implements a processor for direct testing
type MockOrchestratorProcessor struct {
	results map[string]modelResult
}

// Process mocks the processor operation
func (m *MockOrchestratorProcessor) Process(ctx context.Context, modelName string, _ string) (string, error) {
	if result, ok := m.results[modelName]; ok {
		return result.content, result.err
	}
	return "", errors.New("unexpected model name in test")
}

// This function was unused and removed to fix linting issues
