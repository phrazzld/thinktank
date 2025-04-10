package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/config"
)

// mockFullConfigManager implements a minimal config.ManagerInterface for testing
type mockFullConfigManager struct {
	userDir   string
	sysDirs   []string
	templates map[string]string
}

func newMockFullConfigManager() *mockFullConfigManager {
	return &mockFullConfigManager{
		userDir:   "/test/user/config/architect",
		sysDirs:   []string{"/test/sys/etc/architect"},
		templates: make(map[string]string),
	}
}

func (m *mockFullConfigManager) GetConfig() *config.AppConfig {
	return &config.AppConfig{} // Empty config for tests
}

func (m *mockFullConfigManager) GetUserConfigDir() string {
	return m.userDir
}

func (m *mockFullConfigManager) GetSystemConfigDirs() []string {
	return m.sysDirs
}

func (m *mockFullConfigManager) GetUserTemplateDir() string {
	return filepath.Join(m.userDir, "templates")
}

func (m *mockFullConfigManager) GetSystemTemplateDirs() []string {
	dirs := make([]string, 0, len(m.sysDirs))
	for _, dir := range m.sysDirs {
		dirs = append(dirs, filepath.Join(dir, "templates"))
	}
	return dirs
}

func (m *mockFullConfigManager) GetConfigDirs() config.ConfigDirectories {
	return config.ConfigDirectories{
		UserConfigDir:     m.userDir,
		SystemConfigDirs:  m.sysDirs,
		UserTemplateDir:   m.GetUserTemplateDir(),
		SystemTemplateDirs: m.GetSystemTemplateDirs(),
	}
}

func (m *mockFullConfigManager) GetTemplatePath(name string) (string, error) {
	if path, ok := m.templates[name]; ok {
		return path, nil
	}
	return "", os.ErrNotExist
}

func (m *mockFullConfigManager) LoadFromFiles() error {
	return nil // Not needed for tests
}

func (m *mockFullConfigManager) MergeWithFlags(cliFlags map[string]interface{}) error {
	return nil // Not needed for tests
}

func (m *mockFullConfigManager) EnsureConfigDirs() error {
	return nil // Not needed for tests
}

func (m *mockFullConfigManager) WriteDefaultConfig() error {
	return nil // Not needed for tests
}

func TestConfigAdapter(t *testing.T) {
	logger := newMockLogger()
	configManager := newMockFullConfigManager()

	// Set up a test template path
	configManager.templates["test.tmpl"] = "/custom/path/test.tmpl"

	// Create the adapter
	adapter := NewConfigAdapter(configManager, logger)

	// Test that GetTemplatePath properly delegates to the config manager
	path, err := adapter.GetTemplatePath("test.tmpl")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if path != "/custom/path/test.tmpl" {
		t.Errorf("Expected path to be /custom/path/test.tmpl, got: %s", path)
	}

	// Test with a non-existent template
	_, err = adapter.GetTemplatePath("nonexistent.tmpl")
	if err == nil {
		t.Error("Expected error for nonexistent template, got nil")
	}
}

func TestCreatePromptManager(t *testing.T) {
	logger := newMockLogger()
	configManager := newMockFullConfigManager()

	// Create a prompt manager using the config manager
	promptManager := CreatePromptManager(configManager, logger)

	// Verify that the prompt manager was created with the right configuration
	if promptManager == nil {
		t.Fatal("Expected non-nil prompt manager")
	}

	if promptManager.configManager == nil {
		t.Error("Expected prompt manager to have a config manager")
	}

	// Verify that the embedded templates are available
	templates, err := promptManager.ListTemplates()
	if err != nil {
		t.Fatalf("Error listing templates: %v", err)
	}

	if len(templates) == 0 {
		t.Error("Expected at least some templates to be available")
	}
}
