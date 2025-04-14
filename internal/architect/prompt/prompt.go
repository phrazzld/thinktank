// Package prompt handles the creation and manipulation of prompts
// sent to the generative AI models. It provides functions for constructing
// prompts with context files and proper escaping of content.
package prompt

import (
	"strings"

	"github.com/phrazzld/architect/internal/fileutil"
)

// EscapeContent helps prevent conflicts with XML-like tags by escaping < and > characters
func EscapeContent(content string) string {
	escaped := strings.ReplaceAll(content, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	return escaped
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
