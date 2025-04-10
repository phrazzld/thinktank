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
	logger        logutil.LoggerInterface
	defaultPrompt string
	// templatePath is currently unused but kept for future expansion
	// nolint:unused
	templatePath      string
	templates         map[string]*template.Template
	defaultTmplDir    string
	configManager     ConfigManagerInterface
	embeddedTemplates fs.FS
}

// ConfigManagerInterface defines the minimal interface for accessing configuration paths
type ConfigManagerInterface interface {
	GetTemplatePath(name string) (string, error)
}

// NewManager creates a new prompt manager instance
func NewManager(logger logutil.LoggerInterface) *Manager {
	return &Manager{
		logger:            logger,
		defaultPrompt:     "", // Will be loaded from embedded template
		templates:         make(map[string]*template.Template),
		defaultTmplDir:    filepath.Join("internal", "prompt", "templates"),
		embeddedTemplates: EmbeddedTemplates,
	}
}

// NewManagerWithConfig creates a new prompt manager with configuration support
func NewManagerWithConfig(logger logutil.LoggerInterface, configManager ConfigManagerInterface) *Manager {
	return &Manager{
		logger:            logger,
		defaultPrompt:     "", // Will be loaded from embedded template
		templates:         make(map[string]*template.Template),
		defaultTmplDir:    filepath.Join("internal", "prompt", "templates"),
		embeddedTemplates: EmbeddedTemplates,
		configManager:     configManager,
	}
}

// LoadTemplate loads a prompt template from a file or embedded templates.
// The lookup process is as follows:
// 1. If a configManager is available, ask it for the template path
// 2. If a filesystem path is found, load from there
// 3. Otherwise, fallback to embedded templates
//
// This ensures the application can run from any directory, and doesn't
// rely on relative paths that would break when installed globally.
func (m *Manager) LoadTemplate(templatePath string) error {
	var (
		templateContent []byte
		err             error
		isEmbedded      bool
		name            string
	)

	// Handle default case
	if templatePath == "" {
		templatePath = "default.tmpl"
	}

	// Extract the filename for storing in the templates map
	name = filepath.Base(templatePath)

	// First, try using config manager if available
	if m.configManager != nil {
		configPath, err := m.configManager.GetTemplatePath(templatePath)
		if err == nil {
			m.logger.Debug("Using template from config-managed path: %s", configPath)
			templatePath = configPath
		}
	}

	// Try loading from filesystem
	if _, err = os.Stat(templatePath); err == nil {
		// File exists in filesystem
		templateContent, err = os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}
		m.logger.Debug("Loaded template '%s' from filesystem", name)
	} else {
		// File doesn't exist in filesystem, try embedded templates
		embeddedPath := filepath.Join("templates", name)
		templateContent, err = fs.ReadFile(m.embeddedTemplates, embeddedPath)
		if err != nil {
			return fmt.Errorf("template not found in filesystem or embedded templates: %w", err)
		}
		isEmbedded = true
		m.logger.Debug("Loaded template '%s' from embedded templates", name)
	}

	// Parse template
	tmpl, err := template.New(name).Parse(string(templateContent))
	if err != nil {
		source := "filesystem"
		if isEmbedded {
			source = "embedded"
		}
		return fmt.Errorf("failed to parse template from %s: %w", source, err)
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
			// Ensure template has .tmpl extension
			if !strings.HasSuffix(templateName, ".tmpl") {
				templateName = templateName + ".tmpl"
			}

			// Try to load the template
			err := m.LoadTemplate(templateName)
			if err != nil {
				return "", fmt.Errorf("failed to load template %s: %w", templateName, err)
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
	templateSet := make(map[string]struct{}) // Use map to deduplicate templates

	// First, check filesystem templates if available
	if _, err := os.Stat(m.defaultTmplDir); err == nil {
		// List template files from filesystem
		err = filepath.WalkDir(m.defaultTmplDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".tmpl") {
				templateSet[d.Name()] = struct{}{}
			}
			return nil
		})

		if err != nil {
			m.logger.Warn("Failed to list filesystem templates: %v", err)
			// Continue to embedded templates even if filesystem listing fails
		}
	}

	// List templates from config manager if available
	if m.configManager != nil {
		// Try to check user template directory from config
		if configManager, ok := m.configManager.(interface {
			GetUserTemplateDir() string
			GetSystemTemplateDirs() []string
		}); ok {
			// Check user template directory
			userTemplateDir := configManager.GetUserTemplateDir()
			if _, err := os.Stat(userTemplateDir); err == nil {
				err = filepath.WalkDir(userTemplateDir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if !d.IsDir() && strings.HasSuffix(d.Name(), ".tmpl") {
						templateSet[d.Name()] = struct{}{}
					}
					return nil
				})

				if err != nil {
					m.logger.Warn("Failed to list templates from user template directory: %v", err)
				}
			}

			// Check system template directories
			for _, sysDir := range configManager.GetSystemTemplateDirs() {
				if _, err := os.Stat(sysDir); err == nil {
					err = filepath.WalkDir(sysDir, func(path string, d fs.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !d.IsDir() && strings.HasSuffix(d.Name(), ".tmpl") {
							templateSet[d.Name()] = struct{}{}
						}
						return nil
					})

					if err != nil {
						m.logger.Warn("Failed to list templates from system template directory: %v", err)
					}
				}
			}
		}
	}

	// Add embedded templates
	entries, err := fs.ReadDir(m.embeddedTemplates, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			templateSet[entry.Name()] = struct{}{}
		}
	}

	// Convert map to sorted slice
	templates := make([]string, 0, len(templateSet))
	for tmpl := range templateSet {
		templates = append(templates, tmpl)
	}

	return templates, nil
}
