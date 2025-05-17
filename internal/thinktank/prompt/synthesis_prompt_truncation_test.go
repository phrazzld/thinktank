package prompt_test

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/thinktank/prompt"
)

func TestStitchSynthesisPromptTruncation(t *testing.T) {
	tests := []struct {
		name                 string
		instructionsLength   int
		expectedTruncation   bool
		expectedInstructions string
	}{
		{
			name:               "Short instructions - no truncation",
			instructionsLength: 1000,
			expectedTruncation: false,
		},
		{
			name:               "Exactly 50k characters - no truncation",
			instructionsLength: 50000,
			expectedTruncation: false,
		},
		{
			name:               "Just over 50k characters - truncation occurs",
			instructionsLength: 50001,
			expectedTruncation: true,
		},
		{
			name:               "Much longer than 50k - truncation occurs",
			instructionsLength: 100000,
			expectedTruncation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate instructions of the specified length
			instructions := strings.Repeat("x", tt.instructionsLength)
			
			modelOutputs := map[string]string{
				"model1": "Output from model 1",
			}
			
			// Generate synthesis prompt
			result := prompt.StitchSynthesisPrompt(instructions, modelOutputs)
			
			// Check if truncation occurred as expected
			if tt.expectedTruncation {
				// Should contain truncation notice
				if !strings.Contains(result, "[Note: Original instructions truncated") {
					t.Error("Expected truncation notice but didn't find it")
				}
				
				// Should contain the exact character counts
				if !strings.Contains(result, "to 50000 characters]") {
					t.Error("Expected truncation notice to mention 50000 characters")
				}
				
				// The actual content should be truncated to 50k characters
				// Find the content between the context tags
				startTag := "<original_task_context>"
				endTag := "</original_task_context>"
				startIdx := strings.Index(result, startTag)
				endIdx := strings.Index(result, endTag)
				
				if startIdx == -1 || endIdx == -1 {
					t.Fatal("Could not find context tags")
				}
				
				contextContent := result[startIdx+len(startTag):endIdx]
				
				// Find where the actual content starts after the header
				contentStartIdx := strings.Index(contextContent, ":\n")
				if contentStartIdx == -1 {
					t.Fatal("Could not find start of content")
				}
				contentStartIdx += 2 // Skip past ":\n"
				
				// Find where truncation note starts (if any)
				truncNoteIdx := strings.Index(contextContent, "\n\n[Note:")
				
				var actualInstructions string
				if truncNoteIdx != -1 && truncNoteIdx > contentStartIdx {
					actualInstructions = contextContent[contentStartIdx:truncNoteIdx]
				} else {
					// If no truncation note, take everything after the header
					endOfContent := len(contextContent)
					if strings.HasSuffix(contextContent, "\n") {
						endOfContent-- // Don't count trailing newline
					}
					actualInstructions = contextContent[contentStartIdx:endOfContent]
				}
				
				// Trim any trailing whitespace that might have been added
				actualInstructions = strings.TrimSpace(actualInstructions)
				
				if len(actualInstructions) != 50000 {
					t.Errorf("Expected truncated instructions to be 50000 characters, got %d", len(actualInstructions))
					t.Logf("Context content: %q", contextContent)
				}
			} else {
				// Should NOT contain truncation notice
				if strings.Contains(result, "[Note: Original instructions truncated") {
					t.Error("Didn't expect truncation notice but found it")
				}
				
				// The full instructions should be present
				if !strings.Contains(result, instructions) {
					t.Error("Expected full instructions to be present in the output")
				}
			}
			
			// Basic structure checks
			if !strings.Contains(result, "<synthesis_instructions>") {
				t.Error("Synthesis prompt missing synthesis_instructions section")
			}
			if !strings.Contains(result, "<original_task_context>") {
				t.Error("Synthesis prompt missing original_task_context section")
			}
			if !strings.Contains(result, "<model_outputs>") {
				t.Error("Synthesis prompt missing model_outputs section")
			}
		})
	}
}

// Test edge case where instructions are exactly at the boundary
func TestStitchSynthesisPromptBoundaryCase(t *testing.T) {
	// Create instructions that are exactly 50k characters
	instructions := strings.Repeat("a", 50000)
	
	modelOutputs := map[string]string{
		"model1": "Test output",
	}
	
	result := prompt.StitchSynthesisPrompt(instructions, modelOutputs)
	
	// Should not be truncated
	if strings.Contains(result, "[Note: Original instructions truncated") {
		t.Error("Instructions of exactly 50k characters should not be truncated")
	}
	
	// Full instructions should be present
	if !strings.Contains(result, instructions) {
		t.Error("Full instructions should be present for 50k character input")
	}
}