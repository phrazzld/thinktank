package prompt

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/logutil"
)

// mockLogger implements a minimal logger for testing
type mockLogger struct {
	logutil.LoggerInterface
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages: []string{},
		infoMessages:  []string{},
		warnMessages:  []string{},
		errorMessages: []string{},
	}
}

func (l *mockLogger) Debug(format string, args ...interface{}) {
	l.debugMessages = append(l.debugMessages, format)
}
func (l *mockLogger) Info(format string, args ...interface{}) {
	l.infoMessages = append(l.infoMessages, format)
}
func (l *mockLogger) Warn(format string, args ...interface{}) {
	l.warnMessages = append(l.warnMessages, format)
}
func (l *mockLogger) Error(format string, args ...interface{}) {
	l.errorMessages = append(l.errorMessages, format)
}
func (l *mockLogger) Fatal(format string, args ...interface{}) {
	l.errorMessages = append(l.errorMessages, format)
}
func (l *mockLogger) Printf(format string, args ...interface{}) {
	l.infoMessages = append(l.infoMessages, format)
}

// mockConfigManager provides a test implementation of ConfigManagerInterface
type mockConfigManager struct {
	templates map[string]string
}

func newMockConfigManager() *mockConfigManager {
	return &mockConfigManager{
		templates: make(map[string]string),
	}
}

func (m *mockConfigManager) GetTemplatePath(name string) (string, error) {
	if path, ok := m.templates[name]; ok {
		return path, nil
	}
	return "", errors.New("template not found in mock config")
}

func TestEmbeddedTemplates(t *testing.T) {
	// Verify that the embedded templates were properly included
	entries, err := fs.ReadDir(EmbeddedTemplates, "templates")
	if err != nil {
		t.Fatalf("Failed to read embedded templates: %v", err)
	}

	// Check that we have at least the three basic templates
	minTemplates := map[string]bool{
		"default.tmpl": false,
		"clarify.tmpl": false,
		"refine.tmpl":  false,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			minTemplates[entry.Name()] = true
		}
	}

	// Verify all required templates are present
	for name, found := range minTemplates {
		if !found {
			t.Errorf("Required template %s not found in embedded templates", name)
		}
	}
}

func TestLoadTemplateFromEmbedded(t *testing.T) {
	logger := newMockLogger()
	manager := NewManager(logger)

	// Attempt to load a template that doesn't exist in the filesystem
	// but does exist in embedded templates
	err := manager.LoadTemplate("default.tmpl")
	if err != nil {
		t.Fatalf("Failed to load embedded template: %v", err)
	}

	// Check that it was loaded
	if _, exists := manager.templates["default.tmpl"]; !exists {
		t.Error("Template was not stored in templates map")
	}
}

func TestBuildPromptWithEmbedded(t *testing.T) {
	logger := newMockLogger()
	manager := NewManager(logger)

	// Build a prompt using the default template
	data := &TemplateData{
		Task:    "Test task",
		Context: "Test context",
	}

	prompt, err := manager.BuildPrompt("default.tmpl", data)
	if err != nil {
		t.Fatalf("Failed to build prompt: %v", err)
	}

	// Check that the prompt contains our task and context
	if !contains(prompt, "Test task") || !contains(prompt, "Test context") {
		t.Error("Built prompt does not contain expected data")
	}
}

func TestLoadTemplateWithConfig(t *testing.T) {
	// Create a temporary file to simulate a user-configured template
	tempDir, err := os.MkdirTemp("", "prompt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	userTemplatePath := filepath.Join(tempDir, "user-default.tmpl")
	customContent := "Custom template with {{.Task}} and {{.Context}}"
	if err := os.WriteFile(userTemplatePath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a mock config manager that will return our custom template path
	configMgr := newMockConfigManager()
	configMgr.templates["default.tmpl"] = userTemplatePath

	// Create a manager with the config
	logger := newMockLogger()
	manager := NewManagerWithConfig(logger, configMgr)

	// Load the default template, which should use the config path
	err = manager.LoadTemplate("default.tmpl")
	if err != nil {
		t.Fatalf("Failed to load template with config: %v", err)
	}

	// Build a prompt to ensure the custom template was used
	data := &TemplateData{
		Task:    "Custom task",
		Context: "Custom context",
	}

	prompt, err := manager.BuildPrompt("default.tmpl", data)
	if err != nil {
		t.Fatalf("Failed to build prompt with custom template: %v", err)
	}

	expected := "Custom template with Custom task and Custom context"
	if prompt != expected {
		t.Errorf("Expected custom template content, got different content:\nExpected: %s\nGot: %s", expected, prompt)
	}
}

// TestIsTemplate tests the IsTemplate function to verify it correctly identifies templates
func TestIsTemplate(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Basic template with .Task",
			content:  "This is a template with {{.Task}} variable",
			expected: true,
		},
		{
			name:     "Basic template with .Context",
			content:  "This is a template with {{.Context}} variable",
			expected: true,
		},
		{
			name:     "Template with whitespace",
			content:  "This is a template with {{ .Task }} variable",
			expected: true,
		},
		{
			name:     "Multiple template variables",
			content:  "Template with {{.Task}} and {{.Context}} variables",
			expected: true,
		},
		{
			name:     "Not a template - no variables",
			content:  "This is not a template, just plain text",
			expected: false,
		},
		{
			name:     "Not a template - different variables",
			content:  "This has {{.Name}} and {{.Something}} but not the right ones",
			expected: false,
		},
		{
			name:     "Braces but not template syntax",
			content:  "This has { braces } but not templates",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTemplate(tt.content)
			if result != tt.expected {
				t.Errorf("IsTemplate(%q) = %v, want %v", tt.content, result, tt.expected)
			}
		})
	}
}

func TestListTemplates(t *testing.T) {
	logger := newMockLogger()
	manager := NewManager(logger)

	templates, err := manager.ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	// Check that we have at least the three basic templates
	minTemplates := map[string]bool{
		"default.tmpl": false,
		"clarify.tmpl": false,
		"refine.tmpl":  false,
	}

	for _, tmpl := range templates {
		minTemplates[tmpl] = true
	}

	// Verify all required templates are present
	for name, found := range minTemplates {
		if !found {
			t.Errorf("Required template %s not found in template list", name)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
