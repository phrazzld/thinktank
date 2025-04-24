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

// StitchSynthesisPrompt combines original instructions and multiple model outputs
// into a single prompt for a synthesis model. Each model output is clearly labeled
// with the model name for reference.
func StitchSynthesisPrompt(originalInstructions string, modelOutputs map[string]string) string {
	var builder strings.Builder

	// Format original instructions with clear delimiters
	builder.WriteString("<instructions>\n")
	builder.WriteString(originalInstructions)
	builder.WriteString("\n</instructions>\n\n")

	// Format model outputs section with model names as attributes
	builder.WriteString("<model_outputs>\n")
	for modelName, output := range modelOutputs {
		builder.WriteString(fmt.Sprintf("<output model=\"%s\">\n", modelName))
		builder.WriteString(output)
		builder.WriteString("\n</output>\n\n")
	}
	builder.WriteString("</model_outputs>\n\n")

	// Add synthesis instructions
	builder.WriteString("Please synthesize these outputs into a single, comprehensive response that addresses " +
		"the original instructions. Your synthesis should incorporate the strongest insights and information " +
		"from each model's output, resolving any contradictions and presenting a cohesive, well-structured result.")

	return builder.String()
}
