package config

import (
	"reflect"
	"testing"
)

func TestTemplateConfigStructHasNoFieldsClarifyOrRefine(t *testing.T) {
	// Get the type of TemplateConfig struct
	configType := reflect.TypeOf(TemplateConfig{})

	// Check if the Clarify field exists
	_, clarifyExists := configType.FieldByName("Clarify")
	if clarifyExists {
		t.Error("Clarify field still exists in TemplateConfig struct but should have been removed")
	}

	// Check if the Refine field exists
	_, refineExists := configType.FieldByName("Refine")
	if refineExists {
		t.Error("Refine field still exists in TemplateConfig struct but should have been removed")
	}
}

func TestGetTemplatePathFromConfigNoLongerHandlesClarifyOrRefine(t *testing.T) {
	// Create a minimal Manager for testing
	logger := newMockLogger()
	manager := &Manager{
		logger:        logger,
		userConfigDir: "/test/dir",
		sysConfigDirs: []string{"/test/sys/dir"},
		config:        DefaultConfig(),
		viperInst:     nil, // Not needed for this test
	}

	// Test that "clarify" no longer returns a valid path
	path, ok := manager.getTemplatePathFromConfig("clarify")
	if ok {
		t.Errorf("getTemplatePathFromConfig still handles 'clarify', returned: %s", path)
	}

	// Test that "refine" no longer returns a valid path
	path, ok = manager.getTemplatePathFromConfig("refine")
	if ok {
		t.Errorf("getTemplatePathFromConfig still handles 'refine', returned: %s", path)
	}

	// Test that valid templates still work
	path, ok = manager.getTemplatePathFromConfig("default")
	if !ok {
		t.Error("getTemplatePathFromConfig should still handle 'default' but it doesn't")
	}
}
