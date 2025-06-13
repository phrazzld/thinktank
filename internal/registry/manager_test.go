package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestGetGlobalManager tests the singleton behavior of GetGlobalManager
func TestGetGlobalManager(t *testing.T) {
	// Reset the global manager for this test
	managerMu.Lock()
	globalManager = nil
	managerMu.Unlock()

	// Get the global manager
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
	manager1 := GetGlobalManager(logger)
	if manager1 == nil {
		t.Fatal("Expected non-nil manager")
	}

	// Get it again and check if it's the same instance
	manager2 := GetGlobalManager(logger)
	if manager2 != manager1 {
		t.Error("Expected the same manager instance")
	}
}

// TestSetGlobalManagerForTesting tests the SetGlobalManagerForTesting function
func TestSetGlobalManagerForTesting(t *testing.T) {
	// Reset the global manager for this test
	managerMu.Lock()
	originalManager := globalManager
	globalManager = nil
	managerMu.Unlock()

	// Defer restoration of original manager
	defer func() {
		managerMu.Lock()
		globalManager = originalManager
		managerMu.Unlock()
	}()

	// Create a test manager
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")
	testManager := NewManager(logger)

	// Set the global manager using the test function
	SetGlobalManagerForTesting(testManager)

	// Verify the global manager was set correctly
	retrievedManager := GetGlobalManager(logger)
	if retrievedManager != testManager {
		t.Error("SetGlobalManagerForTesting did not set the global manager correctly")
	}

	// Test setting to nil
	SetGlobalManagerForTesting(nil)

	// Getting the global manager should now create a new one
	newManager := GetGlobalManager(logger)
	if newManager == nil {
		t.Fatal("Expected non-nil manager after setting global manager to nil")
	}
	if newManager == testManager {
		t.Error("Expected a new manager instance after setting global manager to nil")
	}
}

// TestNewManager tests the NewManager function
func TestNewManager(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	// Test with valid logger
	manager := NewManager(logger)
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.logger != logger {
		t.Error("NewManager did not set the logger correctly")
	}
	if manager.loaded {
		t.Error("NewManager should initialize with loaded=false")
	}
	if manager.registry == nil {
		t.Error("NewManager should initialize with a non-nil registry")
	}

	// Test with nil logger (should create default logger)
	managerWithNilLogger := NewManager(nil)
	if managerWithNilLogger == nil {
		t.Fatal("NewManager returned nil with nil logger")
	}
	if managerWithNilLogger.logger == nil {
		t.Error("NewManager should create a default logger when passed nil")
	}
}

// TestInitialize tests the Initialize function with various scenarios
func TestInitialize(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "[test] ")

	t.Run("already loaded", func(t *testing.T) {
		manager := NewManager(logger)
		manager.loaded = true // Mark as already loaded

		err := manager.Initialize()
		if err != nil {
			t.Fatalf("Initialize should succeed when already loaded: %v", err)
		}
		if !manager.loaded {
			t.Error("loaded flag should remain true")
		}
	})

	t.Run("successful initialization", func(t *testing.T) {
		// Create a temporary config file for testing
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".config", "thinktank")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test config directory: %v", err)
		}

		configFile := filepath.Join(configDir, "models.yaml")
		minimalConfig := `
providers:
  - id: gemini
    name: Gemini
    api_key_env: GEMINI_API_KEY
    api_url: https://generativelanguage.googleapis.com/v1beta/models

models:
  - id: gemini-1.5-pro
    name: "Test Model"
    provider: gemini
    parameters:
      - name: temperature
        type: number
        default: 0.7
`
		err = os.WriteFile(configFile, []byte(minimalConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Set HOME to our temp directory for the test
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
		if err != nil {
			t.Fatalf("Initialize should succeed with valid config: %v", err)
		}
		if !manager.loaded {
			t.Error("loaded flag should be true after successful initialization")
		}
	})

	t.Run("config file not found - fallback to defaults", func(t *testing.T) {
		// Create a temporary directory that doesn't have a config file
		tempDir := t.TempDir()

		// Set HOME to our temp directory for the test
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
		err := manager.Initialize()

		// Should succeed because the system falls back to embedded defaults
		if err != nil {
			t.Fatalf("Initialize should succeed with embedded defaults when config file not found: %v", err)
		}
		if !manager.loaded {
			t.Error("loaded flag should be true after successful initialization with defaults")
		}
	})

	t.Run("double initialization", func(t *testing.T) {
		// Test that calling Initialize twice doesn't cause issues
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".config", "thinktank")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test config directory: %v", err)
		}

		configFile := filepath.Join(configDir, "models.yaml")
		minimalConfig := `
providers:
  - id: gemini
    name: Gemini
    api_key_env: GEMINI_API_KEY
    api_url: https://generativelanguage.googleapis.com/v1beta/models

models:
  - id: gemini-1.5-pro
    name: "Test Model"
    provider: gemini
    parameters:
      - name: temperature
        type: number
        default: 0.7
`
		err = os.WriteFile(configFile, []byte(minimalConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Set HOME to our temp directory for the test
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

		// First initialization
		err = manager.Initialize()
		if err != nil {
			t.Fatalf("First Initialize should succeed: %v", err)
		}

		// Second initialization should be a no-op
		err = manager.Initialize()
		if err != nil {
			t.Fatalf("Second Initialize should succeed (no-op): %v", err)
		}

		if !manager.loaded {
			t.Error("loaded flag should remain true after double initialization")
		}
	})
}
