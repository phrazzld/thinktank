// Package prompt handles loading and processing prompt templates
package prompt

import (
	"fmt"
)

// MockManager implements prompt.Manager for testing
type MockManager struct {
	LoadTemplateFunc  func(templatePath string) error
	BuildPromptFunc   func(templateName string, data *TemplateData) (string, error)
	ListTemplatesFunc func() ([]string, error)
}

// LoadTemplate calls the mocked implementation
func (m *MockManager) LoadTemplate(templatePath string) error {
	if m.LoadTemplateFunc != nil {
		return m.LoadTemplateFunc(templatePath)
	}
	return nil
}

// BuildPrompt calls the mocked implementation
func (m *MockManager) BuildPrompt(templateName string, data *TemplateData) (string, error) {
	if m.BuildPromptFunc != nil {
		return m.BuildPromptFunc(templateName, data)
	}
	return fmt.Sprintf("Mock prompt for task: %s", data.Task), nil
}

// ListTemplates calls the mocked implementation
func (m *MockManager) ListTemplates() ([]string, error) {
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc()
	}
	return []string{"default.tmpl", "custom.tmpl"}, nil
}

// NewMockManager creates a new mock prompt manager for testing
func NewMockManager() *MockManager {
	return &MockManager{}
}
