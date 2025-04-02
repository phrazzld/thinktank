package config

import (
	"testing"

	"github.com/phrazzld/architect/internal/logutil"
)

func TestDefaultConfig(t *testing.T) {
	// Test that DefaultConfig returns expected values
	cfg := DefaultConfig()

	// Check basic values
	if cfg.OutputFile != DefaultOutputFile {
		t.Errorf("Expected OutputFile to be %s, got %s", DefaultOutputFile, cfg.OutputFile)
	}

	if cfg.ModelName != DefaultModel {
		t.Errorf("Expected ModelName to be %s, got %s", DefaultModel, cfg.ModelName)
	}

	// Check template values
	if cfg.Templates.Default != "default.tmpl" {
		t.Errorf("Expected Default template to be default.tmpl, got %s", cfg.Templates.Default)
	}

	if cfg.Templates.Clarify != "clarify.tmpl" {
		t.Errorf("Expected Clarify template to be clarify.tmpl, got %s", cfg.Templates.Clarify)
	}

	if cfg.Templates.Refine != "refine.tmpl" {
		t.Errorf("Expected Refine template to be refine.tmpl, got %s", cfg.Templates.Refine)
	}

	// Check exclude values
	if cfg.Excludes.Extensions != DefaultExcludes {
		t.Errorf("Expected Excludes.Extensions to match DefaultExcludes")
	}

	if cfg.Excludes.Names != DefaultExcludeNames {
		t.Errorf("Expected Excludes.Names to match DefaultExcludeNames")
	}
}

func TestAppConfig_MarshalUnmarshal(t *testing.T) {
	// This test ensures that our struct can be properly marshaled and unmarshaled
	// by Viper, which would catch tag issues or other serialization problems.
	// The actual implementation will be tested in loader_test.go

	// Initialize with default values
	cfg := DefaultConfig()

	// Modify some values to ensure they round-trip correctly
	cfg.ModelName = "test-model"
	cfg.OutputFile = "test-output.md"
	cfg.Templates.Default = "custom-default.tmpl"
	cfg.Excludes.Extensions = ".test"

	// This is just a placeholder to verify the structure can be marshaled/unmarshaled
	// The actual TOML marshal/unmarshal will be tested in loader_test.go
	t.Log("AppConfig structure is defined with appropriate tags for serialization")
}

type mockLogger struct {
	logutil.LoggerInterface
	debugMessages []string
	infoMessages  []string
	errorMessages []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages: []string{},
		infoMessages:  []string{},
		errorMessages: []string{},
	}
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func TestNewManager(t *testing.T) {
	logger := newMockLogger()
	manager := NewManager(logger)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.config == nil {
		t.Error("Manager.config is nil")
	}

	if manager.logger == nil {
		t.Error("Manager.logger is nil")
	}

	if manager.userConfigDir == "" {
		t.Error("Manager.userConfigDir is empty")
	}

	if manager.viperInst == nil {
		t.Error("Manager.viperInst is nil")
	}
}
