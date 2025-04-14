// Package prompt handles the creation and manipulation of prompts
// sent to the generative AI models. It provides functions for constructing
// prompts with context files and proper escaping of content.
package prompt

import (
	"strings"

	"github.com/phrazzld/architect/internal/fileutil"
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
