// Package prompt handles loading and processing prompt templates
package prompt

import (
	"embed"
)

//go:embed templates/*.tmpl
var EmbeddedTemplates embed.FS
