package prompt_test

import (
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/thinktank/prompt"
)

// TestStitchSynthesisPromptFailedModels verifies that the synthesis prompt generation
// doesn't include entries for failed models when passed only successful model outputs
func TestStitchSynthesisPromptFailedModels(t *testing.T) {
	// Define test instructions
	originalInstructions := "Analyze this code for potential issues"

	// Define test models and their outputs
	successfulModels := map[string]string{
		"success_model1": "Output from successful model 1",
		"success_model2": "Output from successful model 2",
	}

	// List of models that should have failed (not present in the map)
	failedModelNames := []string{"failed_model1", "failed_model2"}

	// Generate the synthesis prompt
	synthesisPrompt := prompt.StitchSynthesisPrompt(originalInstructions, successfulModels)

	// Verify that successful models are included
	for modelName, output := range successfulModels {
		expectedModelTag := "<model_result model=\"" + modelName + "\">"
		if !strings.Contains(synthesisPrompt, expectedModelTag) {
			t.Errorf("Synthesis prompt missing expected successful model: %s", modelName)
		}
		if !strings.Contains(synthesisPrompt, output) {
			t.Errorf("Synthesis prompt missing expected output from model: %s", modelName)
		}
	}

	// Verify that failed models are NOT included
	for _, failedModelName := range failedModelNames {
		unexpectedModelTag := "<model_result model=\"" + failedModelName + "\">"
		if strings.Contains(synthesisPrompt, unexpectedModelTag) {
			t.Errorf("Synthesis prompt contains unexpected failed model: %s", failedModelName)
		}
	}

	// Verify essential prompt structure is correct
	if !strings.Contains(synthesisPrompt, "<instructions>") {
		t.Error("Synthesis prompt missing instructions section")
	}
	if !strings.Contains(synthesisPrompt, "<model_outputs>") {
		t.Error("Synthesis prompt missing model_outputs section")
	}
	if !strings.Contains(synthesisPrompt, "Please synthesize these outputs") {
		t.Error("Synthesis prompt missing synthesis instructions")
	}
}
