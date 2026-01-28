package prompt_test

import (
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/thinktank/prompt"
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
				// Check instructions block
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<instructions>\nAnalyze this code for potential bugs\n</instructions>") {
						t.Error("Instructions section not formatted correctly")
					}
				},
				// Check model_outputs block
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<model_outputs>") || !strings.Contains(result, "</model_outputs>") {
						t.Error("Missing model_outputs section tags")
					}
				},
				// Check output blocks with model attribution
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<model_result model=\"model1\">") {
						t.Error("Missing or incorrectly formatted model1 output tag")
					}
					if !strings.Contains(result, "<model_result model=\"model2\">") {
						t.Error("Missing or incorrectly formatted model2 output tag")
					}
				},
				// Check model contents
				func(t *testing.T, result string) {
					if !strings.Contains(result, "null pointer issue") {
						t.Error("Missing content from model1")
					}
					if !strings.Contains(result, "off-by-one error") {
						t.Error("Missing content from model2")
					}
				},
				// Check synthesis instruction text
				func(t *testing.T, result string) {
					if !strings.Contains(result, "Please synthesize these outputs into a single") {
						t.Error("Missing synthesis instructions")
					}
				},
				// Check closing tags
				func(t *testing.T, result string) {
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
				// Check that instructions tags exist even with empty content
				func(t *testing.T, result string) {
					expected := "<instructions>\n\n</instructions>"
					if !strings.Contains(result, expected) {
						t.Errorf("Empty instructions not properly formatted, expected %q", expected)
					}
				},
				// Check that model output is still included
				func(t *testing.T, result string) {
					if !strings.Contains(result, "Some analysis without instructions") {
						t.Error("Missing model output despite empty instructions")
					}
				},
			},
		},
		{
			name:                 "Empty model outputs map",
			originalInstructions: "Instructions with no model outputs",
			modelOutputs:         map[string]string{},
			checks: []func(t *testing.T, result string){
				// Verify instructions are included
				func(t *testing.T, result string) {
					if !strings.Contains(result, "Instructions with no model outputs") {
						t.Error("Missing instructions content")
					}
				},
				// Check that model_outputs tags exist even with no outputs
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<model_outputs>") || !strings.Contains(result, "</model_outputs>") {
						t.Error("Empty model_outputs not properly formatted")
					}
				},
				// Verify that no output tags are present
				func(t *testing.T, result string) {
					if strings.Contains(result, "<model_result model=") {
						t.Error("Unexpected model_result tag with empty model outputs map")
					}
				},
			},
		},
		{
			name:                 "Empty model output content",
			originalInstructions: "Instructions with empty model output",
			modelOutputs: map[string]string{
				"empty_model": "",
			},
			checks: []func(t *testing.T, result string){
				// Verify model tag is included
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<model_result model=\"empty_model\">") {
						t.Error("Missing model_result tag for empty model content")
					}
				},
				// Check that empty content is handled properly
				func(t *testing.T, result string) {
					expected := "<model_result model=\"empty_model\">\n\n</model_result>"
					if !strings.Contains(result, expected) {
						t.Errorf("Empty model output not properly formatted, expected substring %q", expected)
					}
				},
			},
		},
		{
			name:                 "Special characters in model name",
			originalInstructions: "Test with special characters in model name",
			modelOutputs: map[string]string{
				"model/with-special_chars": "Content from model with special characters in name",
			},
			checks: []func(t *testing.T, result string){
				// Verify model name is preserved in attribute
				func(t *testing.T, result string) {
					expected := "<model_result model=\"model/with-special_chars\">"
					if !strings.Contains(result, expected) {
						t.Errorf("Model name with special chars not properly formatted, expected %q", expected)
					}
				},
			},
		},
		{
			name:                 "Model output with XML-like content",
			originalInstructions: "Analyze HTML/XML code",
			modelOutputs: map[string]string{
				"xml_model": "The issue is in the <div> tag that's missing a closing </div>",
			},
			checks: []func(t *testing.T, result string){
				// Verify XML content is preserved (not escaped)
				func(t *testing.T, result string) {
					if !strings.Contains(result, "<div> tag that's missing a closing </div>") {
						t.Error("XML-like content in model output not preserved correctly")
					}
				},
			},
		},
		{
			name:                 "Multiple models with varying content sizes",
			originalInstructions: "Compare these implementations",
			modelOutputs: map[string]string{
				"short_model": "Brief analysis.",
				"long_model":  "This is a much longer analysis that spans multiple lines.\nIt contains detailed observations about the code.\nIt makes several recommendations for improvement.",
			},
			checks: []func(t *testing.T, result string){
				// Verify both models are included
				func(t *testing.T, result string) {
					if !strings.Contains(result, "Brief analysis") {
						t.Error("Missing content from short model")
					}
					if !strings.Contains(result, "multiple lines") {
						t.Error("Missing content from long model")
					}
				},
			},
		},
	}

	// Run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the stitched synthesis prompt result
			result := prompt.StitchSynthesisPrompt(tt.originalInstructions, tt.modelOutputs)

			// Run all checks for this test case
			for _, check := range tt.checks {
				check(t, result)
			}
		})
	}
}
