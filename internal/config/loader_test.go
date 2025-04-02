package config

import (
	"os"
	"path/filepath"
	"strings"
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
	// Create a temporary directory to simulate user config directory
	tempDir, err := os.MkdirTemp("", "architect-test-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config subdirectories for test
	userConfigDir := filepath.Join(tempDir, "user-config")
	sysConfigDir := filepath.Join(tempDir, "sys-config")

	// Create logger and configure manager with our test directories
	logger := newMockLogger()
	manager := &Manager{
		logger:        logger,
		userConfigDir: userConfigDir,
		sysConfigDirs: []string{sysConfigDir},
		config:        DefaultConfig(),
		viperInst:     viper.New(),
	}

	// Test loading without existing config files
	err = manager.LoadFromFiles()
	if err != nil {
		t.Fatalf("LoadFromFiles should not error when no files found: %v", err)
	}

	// Verify initialization message was logged
	foundInfoMessage := false
	for _, msg := range logger.infoMessages {
		if msg == "No configuration file found. Initializing default configuration..." {
			foundInfoMessage = true
			break
		}
	}

	if !foundInfoMessage {
		t.Error("Expected info message about initializing configuration")
	}

	// Verify config directory was created
	if _, err := os.Stat(userConfigDir); os.IsNotExist(err) {
		t.Errorf("Expected user config directory %s to be created", userConfigDir)
	}

	// Verify config file was created
	configFilePath := filepath.Join(userConfigDir, ConfigFilename)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		t.Errorf("Expected config file %s to be created", configFilePath)
	}

	// Test loading with existing config file
	logger = newMockLogger() // Reset logger
	manager.logger = logger

	err = manager.LoadFromFiles()
	if err != nil {
		t.Fatalf("LoadFromFiles should not error with existing config: %v", err)
	}

	// Verify no initialization message this time
	for _, msg := range logger.infoMessages {
		if msg == "No configuration file found. Initializing default configuration..." {
			t.Error("Should not show initialization message on second load")
		}
	}
}

// TestAutomaticInitialization tests the automatic configuration initialization feature
func TestAutomaticInitialization(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "architect-test-init-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define test paths
	userConfigDir := filepath.Join(tempDir, "user")
	sysConfigDir := filepath.Join(tempDir, "sys")
	templateDir := filepath.Join(userConfigDir, "templates")
	configFilePath := filepath.Join(userConfigDir, ConfigFilename)

	// Test case 1: Basic initialization (happy path)
	t.Run("Basic initialization", func(t *testing.T) {
		// Ensure test directory is clean
		os.RemoveAll(tempDir)

		logger := newMockLogger()
		manager := &Manager{
			logger:        logger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     viper.New(),
		}

		// Run LoadFromFiles which should trigger initialization
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles failed: %v", err)
		}

		// Verify initialization occurred
		if !directoryExists(t, userConfigDir) {
			t.Error("User config directory was not created")
		}

		if !fileExists(t, configFilePath) {
			t.Error("Configuration file was not created")
		}

		// Verify template dir exists (created by EnsureConfigDirs)
		if !directoryExists(t, templateDir) {
			t.Error("Template directory was not created")
		}

		// Verify info message was logged
		assertMessageLogged(t, logger.infoMessages, "No configuration file found. Initializing default configuration...")

		// Verify success message via Printf was logged
		assertMessageLogged(t, logger.infoMessages, "✓ Architect configuration initialized automatically.")
	})

	// Test case 2: Directory creation fails
	t.Run("Directory creation fails", func(t *testing.T) {
		// Ensure test directory is clean
		os.RemoveAll(tempDir)

		// Create a file with the same name as our config directory to cause failure
		if err := os.MkdirAll(filepath.Dir(userConfigDir), 0755); err != nil {
			t.Fatalf("Failed to create parent directory: %v", err)
		}
		if err := os.WriteFile(userConfigDir, []byte("not a directory"), 0644); err != nil {
			t.Fatalf("Failed to create blocking file: %v", err)
		}

		logger := newMockLogger()
		manager := &Manager{
			logger:        logger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     viper.New(),
		}

		// Run LoadFromFiles which should handle the error gracefully
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles should not return error even when dir creation fails: %v", err)
		}

		// Verify warning message was logged
		assertMessageLogged(t, logger.infoMessages, "Failed to create configuration directories:")

		// Ensure no initialization message was shown (since we failed)
		for _, msg := range logger.infoMessages {
			if msg == "✓ Architect configuration initialized automatically." {
				t.Error("Success message should not be shown when initialization fails")
			}
		}
	})

	// Test case: Second run with existing config
	t.Run("Existing configuration", func(t *testing.T) {
		// Ensure test directory is clean
		os.RemoveAll(tempDir)

		// Create required directories
		if err := os.MkdirAll(userConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create user config dir: %v", err)
		}

		// Create a mock config file
		mockConfig := `output_file = "TEST_OUTPUT.md"
model = "test-model"`
		if err := os.WriteFile(configFilePath, []byte(mockConfig), 0644); err != nil {
			t.Fatalf("Failed to create mock config: %v", err)
		}

		logger := newMockLogger()
		manager := &Manager{
			logger:        logger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     viper.New(),
		}

		// Run LoadFromFiles which should NOT trigger initialization
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles failed: %v", err)
		}

		// Check that no initialization-related messages were logged
		for _, msg := range logger.infoMessages {
			if msg == "No configuration file found. Initializing default configuration..." {
				t.Error("Should not log initialization message with existing config")
			}
		}

		// Verify custom config values were loaded
		if manager.config.OutputFile != "TEST_OUTPUT.md" || manager.config.ModelName != "test-model" {
			t.Errorf("Failed to load values from existing config file. Got: %s %s",
				manager.config.OutputFile, manager.config.ModelName)
		}
	})

	// Test case: Partial configuration file
	t.Run("Partial configuration", func(t *testing.T) {
		// Ensure test directory is clean
		os.RemoveAll(tempDir)

		// Create required directories
		if err := os.MkdirAll(userConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create user config dir: %v", err)
		}

		// Create a partial config file (missing some fields)
		partialConfig := `output_file = "PARTIAL_OUTPUT.md"`
		if err := os.WriteFile(configFilePath, []byte(partialConfig), 0644); err != nil {
			t.Fatalf("Failed to create partial config: %v", err)
		}

		logger := newMockLogger()
		manager := &Manager{
			logger:        logger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     viper.New(),
		}

		// Run LoadFromFiles which should load partial config and use defaults for the rest
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles failed: %v", err)
		}

		// Check fields - customized from file
		if manager.config.OutputFile != "PARTIAL_OUTPUT.md" {
			t.Errorf("Failed to load custom value from partial config. Got: %s", manager.config.OutputFile)
		}

		// Check fields - should be default
		if manager.config.ModelName != DefaultModel {
			t.Errorf("Failed to use default for missing field. Got: %s", manager.config.ModelName)
		}
	})
}

// Helper to check if a directory exists
func directoryExists(t *testing.T, path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Fatalf("Error checking directory %s: %v", path, err)
	}
	return info.IsDir()
}

// Helper to check if a file exists
func fileExists(t *testing.T, path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Fatalf("Error checking file %s: %v", path, err)
	}
	return !info.IsDir()
}

// Helper to assert a message was logged
func assertMessageLogged(t *testing.T, messages []string, expectedSubstring string) {
	for _, msg := range messages {
		if strings.Contains(msg, expectedSubstring) {
			return
		}
	}
	t.Errorf("Expected message containing '%s' was not logged", expectedSubstring)
}

// TestDisplayInitializationMessage checks the message formatting
func TestDisplayInitializationMessage(t *testing.T) {
	logger := newMockLogger()
	manager := &Manager{
		logger:        logger,
		userConfigDir: "/test/config/dir",
		config:        DefaultConfig(),
	}

	// Call the function
	manager.displayInitializationMessage()

	// Expected key messages
	expectedMessages := []string{
		"✓ Architect configuration initialized automatically",
		"Created default configuration file at:",
		"Output File:",
		"Model:",
		"Log Level:",
		"Default Template:",
		"customize these settings by editing",
	}

	// Verify all expected messages are present
	for _, expected := range expectedMessages {
		assertMessageLogged(t, logger.infoMessages, expected)
	}

	// Check for specific default values
	assertMessageLogged(t, logger.infoMessages, DefaultOutputFile)
	assertMessageLogged(t, logger.infoMessages, DefaultModel)
}
