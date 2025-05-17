// Package prompt handles the creation and manipulation of prompts
// sent to the generative AI models. It provides functions for constructing
// prompts with context files and proper escaping of content.
package prompt

import (
	"fmt"
	"strings"

	"github.com/phrazzld/thinktank/internal/fileutil"
)

// EscapeContent previously escaped XML-like characters, but this was determined to be incorrect
// as it modifies the actual code content. This function now returns content unchanged to preserve
// the original syntax including < and > characters.
func EscapeContent(content string) string {
	// No escaping is performed to preserve original code syntax
	return content
}

// StitchPrompt combines instructions and file context into the final prompt string with XML-like tags
func StitchPrompt(instructions string, contextFiles []fileutil.FileMeta) string {
	var sb strings.Builder

	// Add instructions block
	sb.WriteString("<instructions>\n")
	if instructions != "" {
		sb.WriteString(instructions)
		sb.WriteString("\n")
	}
	sb.WriteString("</instructions>\n")

	// Add context block
	sb.WriteString("<context>\n")
	for _, file := range contextFiles {
		// Add file path tag
		sb.WriteString("<path>")
		sb.WriteString(file.Path)
		sb.WriteString("</path>\n")

		// Add file content with escaping
		sb.WriteString(EscapeContent(file.Content))
		sb.WriteString("\n\n")
	}
	sb.WriteString("</context>")

	return sb.String()
}

// StitchSynthesisPrompt creates a specialized prompt for the synthesis model that focuses
// on the synthesis task itself, not the original user instructions.
//
// This function creates a structured prompt with XML-like tags that:
// 1. Provides clear synthesis-specific instructions
// 2. Includes original user task as context (not primary instructions)
// 3. Wraps all model outputs in a <model_outputs> section
// 4. Labels each model output with its source model name via <model_result model="name"> tags
//
// The structured format ensures the synthesis model understands its specific role:
// to create a comprehensive synthesis of the model outputs, not to solve the original problem.
func StitchSynthesisPrompt(originalInstructions string, modelOutputs map[string]string) string {
	var builder strings.Builder

	// Primary synthesis instructions - these are the actual instructions for the synthesis model
	builder.WriteString("<synthesis_instructions>\n")
	builder.WriteString("You are a synthesis model. Your task is to create a comprehensive, unified response by combining the outputs from multiple AI models.\n\n")
	builder.WriteString("Requirements:\n")
	builder.WriteString("1. Analyze all model outputs provided below\n")
	builder.WriteString("2. Identify common themes, insights, and recommendations across models\n")
	builder.WriteString("3. Reconcile any contradictions or differences between model outputs\n")
	builder.WriteString("4. Create a single, cohesive response that incorporates the best elements from each model\n")
	builder.WriteString("5. Preserve important details while eliminating redundancy\n")
	builder.WriteString("6. Structure the synthesis in a clear, logical manner\n")
	builder.WriteString("7. If models disagree on key points, present multiple perspectives with analysis\n")
	builder.WriteString("8. Focus on delivering a complete, actionable response to the original task\n")
	builder.WriteString("\n</synthesis_instructions>\n\n")

	// Include original instructions as context (truncated if necessary)
	builder.WriteString("<original_task_context>\n")
	builder.WriteString("The original task given to the models was:\n")

	// Truncate original instructions to 50k characters if needed
	const maxContextLength = 50000
	if len(originalInstructions) > maxContextLength {
		builder.WriteString(originalInstructions[:maxContextLength])
		builder.WriteString("\n\n[Note: Original instructions truncated from ")
		builder.WriteString(fmt.Sprintf("%d", len(originalInstructions)))
		builder.WriteString(" to ")
		builder.WriteString(fmt.Sprintf("%d", maxContextLength))
		builder.WriteString(" characters]")
	} else {
		builder.WriteString(originalInstructions)
	}

	builder.WriteString("\n</original_task_context>\n\n")

	// Format model outputs section with model names as attributes
	builder.WriteString("<model_outputs>\n")
	for modelName, output := range modelOutputs {
		builder.WriteString(fmt.Sprintf("<model_result model=\"%s\">\n", modelName))
		builder.WriteString(output)
		builder.WriteString("\n</model_result>\n\n")
	}
	builder.WriteString("</model_outputs>\n\n")

	// Final synthesis directive
	builder.WriteString("Based on the above model outputs, create your comprehensive synthesis that addresses the original task effectively.")

	return builder.String()
}
