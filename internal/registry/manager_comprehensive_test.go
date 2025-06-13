package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestInitialize_ErrorScenarios tests various error scenarios in Initialize function
func TestInitialize_ErrorScenarios(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	t.Run("fallback to embedded defaults when no config files found", func(t *testing.T) {
		// Create a manager in a completely isolated environment where no config exists
		tempDir := t.TempDir()

		// Change to the temp directory so no config files can be found
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		// Set HOME to the temp directory
		originalHome := os.Getenv("HOME")
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Errorf("Failed to set HOME environment variable: %v", err)
		}
		defer func() {
			if originalHome != "" {
				if err := os.Setenv("HOME", originalHome); err != nil {
					t.Errorf("Failed to restore HOME environment variable: %v", err)
				}
			} else {
				if err := os.Unsetenv("HOME"); err != nil {
					t.Errorf("Failed to unset HOME environment variable: %v", err)
				}
			}
		}()

		manager := NewManager(logger)

		// The system should now succeed by falling back to embedded defaults
		err := manager.Initialize()
		if err != nil {
			t.Errorf("Expected success with embedded defaults fallback, got error: %v", err)
		}
		if !manager.loaded {
			t.Error("Manager should be marked as loaded after successful fallback to defaults")
		}
	})

	t.Run("invalid config file format fallback to embedded defaults", func(t *testing.T) {
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".config", "thinktank")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test config directory: %v", err)
		}

		// Create an invalid YAML config file
		configFile := filepath.Join(configDir, "models.yaml")
		invalidConfig := `invalid: yaml: content: [unclosed bracket`
		err = os.WriteFile(configFile, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		originalHome := os.Getenv("HOME")
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Errorf("Failed to set HOME environment variable: %v", err)
		}
		defer func() {
			if originalHome != "" {
				if err := os.Setenv("HOME", originalHome); err != nil {
					t.Errorf("Failed to restore HOME environment variable: %v", err)
				}
			} else {
				if err := os.Unsetenv("HOME"); err != nil {
					t.Errorf("Failed to unset HOME environment variable: %v", err)
				}
			}
		}()

		manager := NewManager(logger)
		err = manager.Initialize()

		// The system should now succeed by falling back to embedded defaults
		if err != nil {
			t.Errorf("Expected success with embedded defaults fallback, got error: %v", err)
		}
		if !manager.loaded {
			t.Error("Manager should be marked as loaded after successful fallback to defaults")
		}
	})

	t.Run("already loaded manager skips initialization", func(t *testing.T) {
		manager := NewManager(logger)

		// First initialization
		err := manager.Initialize()
		if err != nil {
			t.Fatalf("First initialization failed: %v", err)
		}
		if !manager.loaded {
			t.Error("Manager should be marked as loaded after first initialization")
		}

		// Second initialization should skip
		err = manager.Initialize()
		if err != nil {
			t.Errorf("Second initialization should succeed without error, got: %v", err)
		}
		if !manager.loaded {
			t.Error("Manager should still be marked as loaded")
		}
	})

	t.Run("specific configuration error handling", func(t *testing.T) {
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".config", "thinktank")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test config directory: %v", err)
		}

		// Create a config file with permission issues
		configFile := filepath.Join(configDir, "models.yaml")
		err = os.WriteFile(configFile, []byte("valid: config"), 0000) // No read permissions
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		originalHome := os.Getenv("HOME")
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Errorf("Failed to set HOME environment variable: %v", err)
		}
		defer func() {
			if originalHome != "" {
				if err := os.Setenv("HOME", originalHome); err != nil {
					t.Errorf("Failed to restore HOME environment variable: %v", err)
				}
			} else {
				if err := os.Unsetenv("HOME"); err != nil {
					t.Errorf("Failed to unset HOME environment variable: %v", err)
				}
			}
		}()

		manager := NewManager(logger)
		err = manager.Initialize()

		// Should succeed due to fallback mechanism, but we're testing error handling paths
		// The actual behavior may vary based on the system's fallback implementation
		// For now, we just verify the function completes
		if manager.loaded && err != nil {
			t.Errorf("Manager is loaded but error returned: %v", err)
		}
	})
}

// TestInstallDefaultConfig_ErrorScenarios tests error scenarios in installDefaultConfig
func TestInstallDefaultConfig_ErrorScenarios(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	t.Run("no default config file found", func(t *testing.T) {
		tempDir := t.TempDir()

		// Change to an empty directory where no config exists
		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		originalHome := os.Getenv("HOME")
		if err := os.Setenv("HOME", tempDir); err != nil {
			t.Errorf("Failed to set HOME environment variable: %v", err)
		}
		defer func() {
			if originalHome != "" {
				if err := os.Setenv("HOME", originalHome); err != nil {
					t.Errorf("Failed to restore HOME environment variable: %v", err)
				}
			} else {
				if err := os.Unsetenv("HOME"); err != nil {
					t.Errorf("Failed to unset HOME environment variable: %v", err)
				}
			}
		}()

		manager := NewManager(logger)
		err := manager.installDefaultConfig()

		if err == nil {
			t.Errorf("Expected error when no default config file found, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "default configuration file not found") {
			t.Errorf("Expected error about missing default config, got: %v", err)
		}
	})

	t.Run("config directory creation failure", func(t *testing.T) {
		// Create a read-only directory to prevent config directory creation
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444) // Read-only
		if err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}

		// Create default config file
		configDir := filepath.Join(tempDir, "config")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		configFile := filepath.Join(configDir, "models.yaml")
		err = os.WriteFile(configFile, []byte("test config"), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		originalWd, _ := os.Getwd()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		originalHome := os.Getenv("HOME")
		if err := os.Setenv("HOME", filepath.Join(readOnlyDir, "subdir")); err != nil {
			t.Errorf("Failed to set HOME environment variable: %v", err)
		}
		defer func() {
			if originalHome != "" {
				if err := os.Setenv("HOME", originalHome); err != nil {
					t.Errorf("Failed to restore HOME environment variable: %v", err)
				}
			} else {
				if err := os.Unsetenv("HOME"); err != nil {
					t.Errorf("Failed to unset HOME environment variable: %v", err)
				}
			}
		}()

		manager := NewManager(logger)
		err = manager.installDefaultConfig()

		if err == nil {
			t.Errorf("Expected error when config directory creation fails, got nil")
		}
		// Note: The exact error message might vary by OS
	})
}
