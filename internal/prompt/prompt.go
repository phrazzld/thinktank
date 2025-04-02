// Package prompt handles loading and processing prompt templates
package prompt

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/phrazzld/architect/internal/logutil"
)

// TemplateData holds the data to be passed to the prompt template
type TemplateData struct {
	Task    string
	Context string
}

// ManagerInterface defines the interface for prompt template management
type ManagerInterface interface {
	LoadTemplate(templatePath string) error
	BuildPrompt(templateName string, data *TemplateData) (string, error)
	ListTemplates() ([]string, error)
}

// Manager handles loading and processing prompt templates
type Manager struct {
	logger         logutil.LoggerInterface
	defaultPrompt  string
	templatePath   string
	templates      map[string]*template.Template
	defaultTmplDir string
}

// NewManager creates a new prompt manager instance
func NewManager(logger logutil.LoggerInterface) *Manager {
	return &Manager{
		logger:         logger,
		defaultPrompt:  "", // Will be loaded from embedded template
		templates:      make(map[string]*template.Template),
		defaultTmplDir: filepath.Join("internal", "prompt", "templates"),
	}
}

// LoadTemplate loads a prompt template from a file
func (m *Manager) LoadTemplate(templatePath string) error {
	if templatePath == "" {
		// Use default embedded template
		templatePath = filepath.Join(m.defaultTmplDir, "default.tmpl")
	}

	// Check if file exists
	_, err := os.Stat(templatePath)
	if err != nil {
		return fmt.Errorf("template file not found: %w", err)
	}

	// Load template content
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template
	name := filepath.Base(templatePath)
	tmpl, err := template.New(name).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Store template
	m.templates[name] = tmpl
	return nil
}

// BuildPrompt generates a prompt from a template and data
func (m *Manager) BuildPrompt(templateName string, data *TemplateData) (string, error) {
	// If templateName is a path and not loaded yet, try to load it
	if _, exists := m.templates[templateName]; !exists {
		// Check if it's a file path
		if strings.Contains(templateName, string(os.PathSeparator)) {
			err := m.LoadTemplate(templateName)
			if err != nil {
				return "", fmt.Errorf("failed to load template %s: %w", templateName, err)
			}
			templateName = filepath.Base(templateName)
		} else {
			// Try to load from default directory
			defaultPath := filepath.Join(m.defaultTmplDir, templateName)
			err := m.LoadTemplate(defaultPath)
			if err != nil {
				return "", fmt.Errorf("template not found: %s", templateName)
			}
		}
	}

	// Get template
	tmpl, exists := m.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// Execute template
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ListTemplates returns a list of available template names
func (m *Manager) ListTemplates() ([]string, error) {
	// Check if default template directory exists
	_, err := os.Stat(m.defaultTmplDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to access template directory: %w", err)
	}

	// List template files
	var templates []string
	err = filepath.WalkDir(m.defaultTmplDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".tmpl") {
			templates = append(templates, d.Name())
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}
