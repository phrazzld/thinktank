package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/spf13/viper"
)

// mockStructuredLogger captures log events for testing
type mockStructuredLogger struct {
	events []auditlog.AuditEvent
}

func newMockStructuredLogger() *mockStructuredLogger {
	return &mockStructuredLogger{
		events: []auditlog.AuditEvent{},
	}
}

// Log implements the StructuredLogger interface
func (m *mockStructuredLogger) Log(event auditlog.AuditEvent) {
	m.events = append(m.events, event)
}

// Close implements the StructuredLogger interface
func (m *mockStructuredLogger) Close() error {
	return nil
}

// TestConfigLoggingWithAuditLogger tests that configuration loading events are properly logged
func TestConfigLoggingWithAuditLogger(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "architect-test-audit-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define test paths
	userConfigDir := filepath.Join(tempDir, "user")
	sysConfigDir := filepath.Join(tempDir, "sys")
	configFilePath := filepath.Join(userConfigDir, ConfigFilename)

	// Create mock loggers
	stdLogger := newMockLogger()
	auditLogger := newMockStructuredLogger()

	// Test case 1: No config file exists (initialization)
	t.Run("Config initialization with audit logging", func(t *testing.T) {
		// Clean test directory
		os.RemoveAll(tempDir)

		// Create manager with audit logger
		manager := &Manager{
			logger:        stdLogger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     createTestViper(),
			auditLogger:   auditLogger,
		}

		// Load configuration (should create default)
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles failed: %v", err)
		}

		// Verify config directory and file were created
		if !directoryExists(t, userConfigDir) {
			t.Error("User config directory was not created")
		}
		if !fileExists(t, configFilePath) {
			t.Error("Config file was not created")
		}

		// Check audit log events
		// We should have at least:
		// 1. An event for starting configuration loading
		// 2. An event for config file not found
		// 3. An event for writing default config
		// 4. An event for successful initialization
		if len(auditLogger.events) < 4 {
			t.Errorf("Expected at least 4 audit log events, got %d", len(auditLogger.events))
		}

		// Verify specific events are present
		verifyEventExists(t, auditLogger.events, "ConfigLoadStart", "INFO")
		verifyEventExists(t, auditLogger.events, "ConfigFileNotFound", "INFO")
		verifyEventExists(t, auditLogger.events, "DefaultConfigCreated", "INFO")
		verifyEventExists(t, auditLogger.events, "ConfigLoadComplete", "INFO")
	})

	// Test case 2: Config file exists and is loaded
	t.Run("Config loading with existing file", func(t *testing.T) {
		// Clean test directory and create new mock loggers
		os.RemoveAll(tempDir)
		stdLogger = newMockLogger()
		auditLogger = newMockStructuredLogger()

		// Create directories and config file
		if err := os.MkdirAll(userConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create user config dir: %v", err)
		}

		// Create a test config file
		testConfig := `output_file = "TEST_OUTPUT.md"
model = "test-model"
audit_log_enabled = true
audit_log_file = "test-audit.log"`
		if err := os.WriteFile(configFilePath, []byte(testConfig), 0644); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		// Create manager with audit logger
		manager := &Manager{
			logger:        stdLogger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     createTestViper(),
			auditLogger:   auditLogger,
		}

		// Load configuration
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles failed: %v", err)
		}

		// Check that configuration was loaded properly
		if manager.config.OutputFile != "TEST_OUTPUT.md" {
			t.Errorf("Config value not loaded correctly, got: %s", manager.config.OutputFile)
		}

		// Check audit log events
		// We should have at least:
		// 1. An event for starting configuration loading
		// 2. An event for finding and loading config file
		// 3. An event for successful loading
		if len(auditLogger.events) < 3 {
			t.Errorf("Expected at least 3 audit log events, got %d", len(auditLogger.events))
		}

		// Verify specific events are present
		verifyEventExists(t, auditLogger.events, "ConfigLoadStart", "INFO")
		verifyEventExists(t, auditLogger.events, "ConfigFileLoaded", "INFO")
		verifyEventExists(t, auditLogger.events, "ConfigLoadComplete", "INFO")

		// Check metadata in the ConfigFileLoaded event
		for _, event := range auditLogger.events {
			if event.Operation == "ConfigFileLoaded" {
				// Verify file path is in metadata
				if event.Metadata == nil || event.Metadata["file_path"] == nil {
					t.Error("ConfigFileLoaded event should include file_path in metadata")
				}
				break
			}
		}
	})

	// Test case 3: Error during loading
	t.Run("Error during config loading", func(t *testing.T) {
		// Clean test directory and create new mock loggers
		os.RemoveAll(tempDir)
		stdLogger = newMockLogger()
		auditLogger = newMockStructuredLogger()

		// Create directories
		if err := os.MkdirAll(userConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create user config dir: %v", err)
		}

		// Create an invalid config file
		invalidConfig := `output_file = "TEST_OUTPUT.md"
model = test-model" # Missing quote creates syntax error`
		if err := os.WriteFile(configFilePath, []byte(invalidConfig), 0644); err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		// Create manager with audit logger
		manager := &Manager{
			logger:        stdLogger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     createTestViper(),
			auditLogger:   auditLogger,
		}

		// Try to load configuration (should fail parsing but not return error)
		_ = manager.LoadFromFiles()

		// Check audit log events
		// We should have at least:
		// 1. An event for starting configuration loading
		// 2. An event for error during loading
		if len(auditLogger.events) < 2 {
			t.Errorf("Expected at least 2 audit log events, got %d", len(auditLogger.events))
		}

		// Verify specific events are present
		verifyEventExists(t, auditLogger.events, "ConfigLoadStart", "INFO")
		verifyEventExists(t, auditLogger.events, "ConfigLoadError", "ERROR")
	})

	// Test case 4: MergeWithFlags
	t.Run("MergeWithFlags audit logging", func(t *testing.T) {
		// Clean test directory and create new mock loggers
		os.RemoveAll(tempDir)
		stdLogger = newMockLogger()
		auditLogger = newMockStructuredLogger()

		// Create manager with audit logger
		manager := &Manager{
			logger:        stdLogger,
			userConfigDir: userConfigDir,
			sysConfigDirs: []string{sysConfigDir},
			config:        DefaultConfig(),
			viperInst:     createTestViper(),
			auditLogger:   auditLogger,
		}

		// Prepare flags
		flags := map[string]interface{}{
			"output_file":       "custom-output.md",
			"model":             "custom-model",
			"templates.default": "custom-template.tmpl",
		}

		// Merge flags
		err = manager.MergeWithFlags(flags)
		if err != nil {
			t.Fatalf("MergeWithFlags failed: %v", err)
		}

		// Check that flags were merged correctly
		if manager.config.OutputFile != "custom-output.md" {
			t.Errorf("Flag not merged correctly, got: %s", manager.config.OutputFile)
		}

		// Verify audit log events
		verifyEventExists(t, auditLogger.events, "MergeFlags", "INFO")
		verifyEventExists(t, auditLogger.events, "MergeFlagsComplete", "INFO")

		// Check flag count metadata
		for _, event := range auditLogger.events {
			if event.Operation == "MergeFlags" {
				if event.Metadata == nil || event.Metadata["flag_count"] == nil {
					t.Error("MergeFlags event should include flag_count in metadata")
				}
				break
			}
		}
	})
}

// createTestViper creates a Viper instance for testing
func createTestViper() *viper.Viper {
	v := viper.New()
	v.SetConfigType("toml")
	return v
}

// Helper function to verify an event exists in the log
func verifyEventExists(t *testing.T, events []auditlog.AuditEvent, operation, level string) {
	for _, event := range events {
		if event.Operation == operation && event.Level == level {
			return
		}
	}
	t.Errorf("Expected %s %s event not found in audit log", level, operation)
}

// Test for NewManager constructor with and without audit logger
func TestNewManagerWithAuditLogger(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "architect-test-constructor-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create loggers
	stdLogger := newMockLogger()
	auditLogger := newMockStructuredLogger()

	// Test case 1: Create manager with audit logger
	t.Run("With audit logger", func(t *testing.T) {
		manager := NewManager(stdLogger, auditLogger)

		// Verify manager was created with the right audit logger
		if manager.auditLogger != auditLogger {
			t.Error("Manager created with wrong audit logger")
		}
	})

	// Test case 2: Create manager without audit logger (should use NoopLogger)
	t.Run("Without audit logger", func(t *testing.T) {
		manager := NewManager(stdLogger)

		// Verify manager was created with a NoopLogger
		if manager.auditLogger == nil {
			t.Error("Manager should create a NoopLogger when audit logger not provided")
		}

		// Check that it's a NoopLogger by examining its type
		_, isNoopLogger := manager.auditLogger.(*auditlog.NoopLogger)
		if !isNoopLogger {
			t.Error("Manager should use NoopLogger when audit logger not provided")
		}
	})

	// Test case 3: Operations should work with NoopLogger
	t.Run("Operations with NoopLogger", func(t *testing.T) {
		// Create manager without explicit audit logger
		manager := NewManager(stdLogger)

		// Define test paths
		manager.userConfigDir = filepath.Join(tempDir, "user")
		manager.sysConfigDirs = []string{filepath.Join(tempDir, "sys")}

		// Verify that config operations work with the NoopLogger
		err = manager.LoadFromFiles()
		if err != nil {
			t.Fatalf("LoadFromFiles failed with NoopLogger: %v", err)
		}

		// Merge flags - should not panic even with NoopLogger
		flags := map[string]interface{}{"output_file": "custom-output.md"}
		err = manager.MergeWithFlags(flags)
		if err != nil {
			t.Fatalf("MergeWithFlags failed with NoopLogger: %v", err)
		}

		// Verify standard logger was still used
		assertMessageLogged(t, stdLogger.infoMessages, "No configuration file found")
	})
}
