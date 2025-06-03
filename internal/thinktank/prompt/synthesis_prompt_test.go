package prompt_test

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/thinktank/prompt"
)

// TestStitchSynthesisPrompt tests the synthesis prompt creation function
func TestStitchSynthesisPrompt(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                 string
		originalInstructions string
		modelOutputs         map[string]string
		checks               []func(t *testing.T, result string)
	}{
		{
			name:                 "Standard case - instructions with multiple models",
			originalInstructions: "Analyze this code for potential bugs",
			modelOutputs: map[string]string{
				"model1": "The code has a potential null pointer issue in function X",
				"model2": "There might be an off-by-one error in the loop at line Y",
			},
			checks: []func(t *testing.T, result string){
				// Check synthesis instructions block
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<synthesis_instructions>") {
						t.Error("Missing synthesis instructions section")
					}
					if !strings.Contains(result, "You are a synthesis model") {
						t.Error("Missing synthesis model introduction")
					}
				},
				// Check original task context
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<original_task_context>") {
						t.Error("Missing original task context section")
					}
					if !strings.Contains(result, "Analyze this code for potential bugs") {
						t.Error("Original instructions not included in context")
					}
				},
				// Check model_outputs block
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<model_outputs>") || !strings.Contains(result, "</model_outputs>") {
						t.Error("Missing model_outputs section tags")
					}
				},
				// Check output blocks with model attribution
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<model_result model=\"model1\">") {
						t.Error("Missing or incorrectly formatted model1 output tag")
					}
					if !strings.Contains(result, "<model_result model=\"model2\">") {
						t.Error("Missing or incorrectly formatted model2 output tag")
					}
				},
				// Check model contents
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "null pointer issue") {
						t.Error("Model1 output content missing")
					}
					if !strings.Contains(result, "off-by-one error") {
						t.Error("Model2 output content missing")
					}
				},
				// Check final synthesis directive
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "Based on the above model outputs, create your comprehensive synthesis") {
						t.Error("Missing final synthesis directive")
					}
				},
				// Check closing tags
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "</model_result>") {
						t.Error("Missing closing model_result tags")
					}
				},
			},
		},
		{
			name:                 "Empty instructions",
			originalInstructions: "",
			modelOutputs: map[string]string{
				"model1": "Some analysis without instructions",
			},
			checks: []func(t *testing.T, result string){
				// Check that original task context exists even with empty instructions
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<original_task_context>") {
						t.Error("Missing original task context section")
					}
				},
				// Check that model output is still included
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "Some analysis without instructions") {
						t.Error("Model output not included")
					}
				},
			},
		},
		{
			name:                 "Empty model outputs map",
			originalInstructions: "Some instructions",
			modelOutputs:         map[string]string{},
			checks: []func(t *testing.T, result string){
				// Check basic structure exists
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<synthesis_instructions>") {
						t.Error("Missing synthesis instructions")
					}
					if !strings.Contains(result, "<model_outputs>") {
						t.Error("Missing model outputs section")
					}
					if !strings.Contains(result, "<original_task_context>") {
						t.Error("Missing original task context section")
					}
				},
			},
		},
		{
			name:                 "Empty model output content",
			originalInstructions: "Process this data",
			modelOutputs: map[string]string{
				"model1": "",
			},
			checks: []func(t *testing.T, result string){
				// Check model_result tag exists even with empty content
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<model_result model=\"model1\">") {
						t.Error("Missing model_result tag for empty output")
					}
				},
			},
		},
		{
			name:                 "Special characters in model name",
			originalInstructions: "Generate a summary",
			modelOutputs: map[string]string{
				"model-v2.1": "Summary generated by special model",
			},
			checks: []func(t *testing.T, result string){
				// Check special characters in model name are preserved
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<model_result model=\"model-v2.1\">") {
						t.Error("Special characters in model name not preserved")
					}
				},
			},
		},
		{
			name:                 "Model output with XML-like content",
			originalInstructions: "Format this data",
			modelOutputs: map[string]string{
				"model1": "Some <tag>content</tag> with XML-like structure",
			},
			checks: []func(t *testing.T, result string){
				// Check XML-like content is preserved
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "Some <tag>content</tag> with XML-like structure") {
						t.Error("XML-like content in model output not preserved")
					}
				},
			},
		},
		{
			name:                 "Multiple models with varying content sizes",
			originalInstructions: "Compare approaches",
			modelOutputs: map[string]string{
				"model1": "Short response",
				"model2": strings.Repeat("Long response. ", 50),
				"model3": strings.Repeat("A", 1000),
			},
			checks: []func(t *testing.T, result string){
				// Check all models are included
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "<model_result model=\"model1\">") ||
						!strings.Contains(result, "<model_result model=\"model2\">") ||
						!strings.Contains(result, "<model_result model=\"model3\">") {
						t.Error("Not all models included in synthesis prompt")
					}
				},
				// Check content preservation
				func(t *testing.T, result string) {
					t.Helper()
					if !strings.Contains(result, "Short response") {
						t.Error("Short response not preserved")
					}
					if !strings.Contains(result, strings.Repeat("Long response. ", 20)) {
						t.Error("Long response not preserved (checking partial)")
					}
					if !strings.Contains(result, strings.Repeat("A", 100)) {
						t.Error("Very long response not preserved (checking partial)")
					}
				},
			},
		},
	}

	// Run the test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := prompt.StitchSynthesisPrompt(tc.originalInstructions, tc.modelOutputs)

			// Apply all checks
			for _, check := range tc.checks {
				check(t, result)
			}
		})
	}
}
