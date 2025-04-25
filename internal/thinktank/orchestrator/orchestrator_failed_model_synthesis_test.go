package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/ratelimit"
)

// TestSynthesizeResultsWithFailedModels verifies that the synthesis process
// doesn't include failed models in the synthesis prompt
func TestSynthesizeResultsWithFailedModels(t *testing.T) {
	// Set up test configuration
	instructions := "Test instructions for synthesis"

	// Create successful and failed model outputs
	// Only successful models should be in the modelOutputs map
	// This simulates the behavior of processModels which only includes successful models
	modelOutputs := map[string]string{
		"success_model1": "Output from successful model 1",
		"success_model2": "Output from successful model 2",
	}

	// Failed model names that should NOT appear in synthesis prompt
	failedModels := []string{"failed_model1", "failed_model2"}

	// Create a special mock API service that captures the synthesis prompt
	promptCapturingService := &PromptCapturingAPIService{
		MockAPIService: MockAPIService{
			modelParams:    map[string]interface{}{"temperature": 0.7},
			generateResult: &llm.ProviderResult{},
			processOutput:  "Synthesized output from successful models only",
		},
	}

	// Create orchestrator with mocks
	mockContextGatherer := &MockContextGatherer{}
	mockFileWriter := &MockFileWriter{}
	mockRateLimiter := ratelimit.NewRateLimiter(0, 0)
	mockAuditLogger := &MockAuditLogger{}
	mockLogger := &MockLogger{}

	// Create config
	cfg := &config.CliConfig{
		SynthesisModel: "synthesis-model",
		ModelNames:     append([]string{"success_model1", "success_model2"}, failedModels...),
	}

	// Create orchestrator
	orchestrator := NewOrchestrator(
		promptCapturingService,
		mockContextGatherer,
		mockFileWriter,
		mockAuditLogger,
		mockRateLimiter,
		cfg,
		mockLogger,
	)

	// Call synthesizeResults
	_, err := orchestrator.synthesizeResults(context.Background(), instructions, modelOutputs)
	if err != nil {
		t.Fatalf("Unexpected error in synthesizeResults: %v", err)
	}

	// Get the captured synthesis prompt
	capturedPrompt := promptCapturingService.capturedPrompt
	if capturedPrompt == "" {
		t.Fatal("No synthesis prompt was captured")
	}

	// Verify successful models are included in the synthesis prompt
	for modelName := range modelOutputs {
		expectedModelTag := "<model_result model=\"" + modelName + "\">"
		if !strings.Contains(capturedPrompt, expectedModelTag) {
			t.Errorf("Synthesis prompt missing expected successful model: %s", modelName)
		}
	}

	// Verify failed models are NOT included in the synthesis prompt
	for _, failedModel := range failedModels {
		unexpectedModelTag := "<model_result model=\"" + failedModel + "\">"
		if strings.Contains(capturedPrompt, unexpectedModelTag) {
			t.Errorf("Synthesis prompt contains unexpected failed model: %s", failedModel)
		}
	}
}

// PromptCapturingAPIService extends MockAPIService to capture the synthesis prompt
type PromptCapturingAPIService struct {
	MockAPIService
	capturedPrompt string
}

// Override GenerateContent to capture the prompt before passing to the mock
func (m *PromptCapturingAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	if m.clientInitError != nil {
		return nil, m.clientInitError
	}
	return &PromptCapturingLLMClient{
		MockLLMClient: MockLLMClient{
			generateResult: m.generateResult,
			generateError:  m.generateError,
		},
		capturePrompt: &m.capturedPrompt,
	}, nil
}

// PromptCapturingLLMClient extends MockLLMClient to capture the prompt
type PromptCapturingLLMClient struct {
	MockLLMClient
	capturePrompt *string
}

// Override GenerateContent to capture the prompt
func (m *PromptCapturingLLMClient) GenerateContent(ctx context.Context, prompt string, parameters map[string]interface{}) (*llm.ProviderResult, error) {
	// Capture the prompt
	*m.capturePrompt = prompt
	return m.generateResult, m.generateError
}

// TestProcessModelsToSynthesis verifies the complete flow from processModels to synthesizeResults,
// ensuring that failed models are excluded throughout the entire process
func TestProcessModelsToSynthesis(t *testing.T) {
	// Create test orchestrator with controlled model processing
	orchestratorWithControlledModels := &synthesisTestOrchestrator{
		mockResults: []modelResult{
			{modelName: "success_model1", content: "Output from success model 1", err: nil},
			{modelName: "failed_model", content: "", err: errors.New("model failed")},
			{modelName: "success_model2", content: "Output from success model 2", err: nil},
		},
	}

	// Process models with controlled behavior
	modelOutputs, modelErrors := orchestratorWithControlledModels.processModels(context.Background(), "test prompt")

	// Verify correct number of outputs and errors
	if len(modelOutputs) != 2 {
		t.Errorf("Expected 2 successful model outputs, got %d", len(modelOutputs))
	}
	if len(modelErrors) != 1 {
		t.Errorf("Expected 1 model error, got %d", len(modelErrors))
	}

	// Verify expected models in output map
	expectedModels := map[string]bool{
		"success_model1": true,
		"success_model2": true,
	}
	for modelName := range modelOutputs {
		if _, exists := expectedModels[modelName]; !exists {
			t.Errorf("Unexpected model in outputs: %s", modelName)
		}
	}

	// Verify failed model is NOT in output map
	if _, exists := modelOutputs["failed_model"]; exists {
		t.Error("Failed model should not be in model outputs map")
	}

	// Create API service to capture synthesis prompt
	promptCapturingService := &PromptCapturingAPIService{
		MockAPIService: MockAPIService{
			modelParams:    map[string]interface{}{"temperature": 0.7},
			generateResult: &llm.ProviderResult{},
			processOutput:  "Synthesized output from successful models only",
		},
	}

	// Create actual orchestrator for synthesis
	mockContextGatherer := &MockContextGatherer{}
	mockFileWriter := &MockFileWriter{}
	mockRateLimiter := ratelimit.NewRateLimiter(0, 0)
	mockAuditLogger := &MockAuditLogger{}
	mockLogger := &MockLogger{}

	// Create config
	cfg := &config.CliConfig{
		SynthesisModel: "synthesis-model",
	}

	// Create orchestrator
	orchestrator := NewOrchestrator(
		promptCapturingService,
		mockContextGatherer,
		mockFileWriter,
		mockAuditLogger,
		mockRateLimiter,
		cfg,
		mockLogger,
	)

	// Run synthesis with outputs from processModels
	_, err := orchestrator.synthesizeResults(context.Background(), "test instructions", modelOutputs)
	if err != nil {
		t.Fatalf("Unexpected error in synthesizeResults: %v", err)
	}

	// Get the captured synthesis prompt
	capturedPrompt := promptCapturingService.capturedPrompt
	if capturedPrompt == "" {
		t.Fatal("No synthesis prompt was captured")
	}

	// Verify synthesis prompt contains successful models
	if !strings.Contains(capturedPrompt, "<model_result model=\"success_model1\">") {
		t.Error("Synthesis prompt missing successful model: success_model1")
	}
	if !strings.Contains(capturedPrompt, "<model_result model=\"success_model2\">") {
		t.Error("Synthesis prompt missing successful model: success_model2")
	}

	// Verify synthesis prompt does NOT contain failed model
	if strings.Contains(capturedPrompt, "<model_result model=\"failed_model\">") {
		t.Error("Synthesis prompt contains failed model which should be excluded")
	}
}

// synthesisTestOrchestrator is a specialized test orchestrator for synthesis tests
type synthesisTestOrchestrator struct {
	mockResults []modelResult
}

// processModels Simulates model processing with predetermined results
func (o *synthesisTestOrchestrator) processModels(ctx context.Context, prompt string) (map[string]string, []error) {
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
