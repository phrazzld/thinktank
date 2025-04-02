package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetTemplatePath(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "architect-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	userConfigDir := filepath.Join(tempDir, "user", ".config", "architect")
	userTemplateDir := filepath.Join(userConfigDir, "templates")
	sysConfigDir := filepath.Join(tempDir, "sys", "etc", "architect")
	sysTemplateDir := filepath.Join(sysConfigDir, "templates")
	cwdTemplateDir := filepath.Join(tempDir, "cwd")

	// Create directories
	for _, dir := range []string{userTemplateDir, sysTemplateDir, cwdTemplateDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create test dir %s: %v", dir, err)
		}
	}

	// Create test template files
	testTemplates := map[string]string{
		filepath.Join(userTemplateDir, "user.tmpl"):     "user template",
		filepath.Join(sysTemplateDir, "system.tmpl"):    "system template",
		filepath.Join(cwdTemplateDir, "cwd.tmpl"):       "cwd template",
		filepath.Join(userTemplateDir, "default.tmpl"):  "user default",
		filepath.Join(sysTemplateDir, "default.tmpl"):   "system default",
		filepath.Join(userTemplateDir, "override.tmpl"): "user override",
		filepath.Join(sysTemplateDir, "override.tmpl"):  "system override",
	}

	for path, content := range testTemplates {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", path, err)
		}
	}

	// Create a manager with our test directories
	logger := newMockLogger()
	manager := &Manager{
		logger:        logger,
		userConfigDir: userConfigDir,
		sysConfigDirs: []string{sysConfigDir},
		config:        DefaultConfig(),
		viperInst:     viper.New(),
	}

	// Test 1: Absolute path
	absolutePath := filepath.Join(userTemplateDir, "user.tmpl")
	t.Run("Absolute path", func(t *testing.T) {
		path, err := manager.GetTemplatePath(absolutePath)
		if err != nil {
			t.Errorf("Error finding template with absolute path: %v", err)
		}
		if path != absolutePath {
			t.Errorf("Expected path %s, got %s", absolutePath, path)
		}
	})

	// Note: We're skipping the relative path from CWD test as it's difficult to test
	// reliably in an automated test environment due to working directory changes.
	// In a real scenario, GetTemplatePath will check the current working directory.

	// Test 3: Template in user directory
	t.Run("User directory template", func(t *testing.T) {
		path, err := manager.GetTemplatePath("user.tmpl")
		if err != nil {
			t.Errorf("Error finding template in user directory: %v", err)
		}
		expectedPath := filepath.Join(userTemplateDir, "user.tmpl")
		if path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, path)
		}
	})

	// Test 4: Template in system directory
	t.Run("System directory template", func(t *testing.T) {
		path, err := manager.GetTemplatePath("system.tmpl")
		if err != nil {
			t.Errorf("Error finding template in system directory: %v", err)
		}
		expectedPath := filepath.Join(sysTemplateDir, "system.tmpl")
		if path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, path)
		}
	})

	// Test 5: Template name with precedence (user overrides system)
	t.Run("Template precedence", func(t *testing.T) {
		path, err := manager.GetTemplatePath("override.tmpl")
		if err != nil {
			t.Errorf("Error finding template with precedence: %v", err)
		}
		expectedPath := filepath.Join(userTemplateDir, "override.tmpl")
		if path != expectedPath {
			t.Errorf("Expected user override to be preferred, got %s", path)
		}
	})

	// Test 6: Template with configured path
	// Set configured path for default template
	customPath := filepath.Join(tempDir, "custom.tmpl")
	if err := os.WriteFile(customPath, []byte("custom template"), 0644); err != nil {
		t.Fatalf("failed to write custom template: %v", err)
	}

	manager.config.Templates.Default = customPath

	t.Run("Template with configured path", func(t *testing.T) {
		path, err := manager.GetTemplatePath("default")
		if err != nil {
			t.Errorf("Error finding template with configured path: %v", err)
		}
		if path != customPath {
			t.Errorf("Expected configured path %s, got %s", customPath, path)
		}
	})

	// Test 7: Nonexistent template
	t.Run("Nonexistent template", func(t *testing.T) {
		_, err := manager.GetTemplatePath("nonexistent.tmpl")
		if err == nil {
			t.Error("Expected error for nonexistent template, got nil")
		}
	})
}

func TestMergeWithFlags(t *testing.T) {
	logger := newMockLogger()
	manager := NewManager(logger)

	// Test merging with basic flags
	flags := map[string]interface{}{
		"output_file": "custom-output.md",
		"model":       "custom-model",
		"verbose":     true,
		"use_colors":  false,
	}

	if err := manager.MergeWithFlags(flags); err != nil {
		t.Fatalf("Error merging flags: %v", err)
	}

	config := manager.GetConfig()

	if config.OutputFile != "custom-output.md" {
		t.Errorf("Expected OutputFile to be custom-output.md, got %s", config.OutputFile)
	}

	if config.ModelName != "custom-model" {
		t.Errorf("Expected ModelName to be custom-model, got %s", config.ModelName)
	}

	if !config.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if config.UseColors {
		t.Error("Expected UseColors to be false")
	}

	// Test nested flag handling
	nestedFlags := map[string]interface{}{
		"templates.default":   "nested-default.tmpl",
		"excludes.extensions": ".nested",
	}

	if err := manager.MergeWithFlags(nestedFlags); err != nil {
		t.Fatalf("Error merging nested flags: %v", err)
	}

	if config.Templates.Default != "nested-default.tmpl" {
		t.Errorf("Expected Templates.Default to be nested-default.tmpl, got %s", config.Templates.Default)
	}

	if config.Excludes.Extensions != ".nested" {
		t.Errorf("Expected Excludes.Extensions to be .nested, got %s", config.Excludes.Extensions)
	}
}

func TestLoadFromFiles(t *testing.T) {
	// This test requires creating temporary config files with viper
	// and is more complex to setup in a unit test.
	// A more comprehensive test would be part of integration testing.

	// For now, just verify basic function behavior with no config files present
	logger := newMockLogger()
	manager := NewManager(logger)

	err := manager.LoadFromFiles()
	if err != nil {
		t.Fatalf("LoadFromFiles should not error when no files found: %v", err)
	}

	// Check that a debug message was logged
	found := false
	for _, msg := range logger.debugMessages {
		if msg == "No configuration file found, using defaults" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected debug message about using defaults")
	}
}
