package config

import (
	"testing"
)

// TestCliConfig_SuppressDeprecationWarnings_DefaultValue tests that the
// SuppressDeprecationWarnings field defaults to false
func TestCliConfig_SuppressDeprecationWarnings_DefaultValue(t *testing.T) {
	cfg := NewDefaultCliConfig()
	if cfg.SuppressDeprecationWarnings {
		t.Error("Expected SuppressDeprecationWarnings to default to false")
	}
}

// TestCliConfig_SuppressDeprecationWarnings_CanBeSet tests that the
// SuppressDeprecationWarnings field can be set to true
func TestCliConfig_SuppressDeprecationWarnings_CanBeSet(t *testing.T) {
	cfg := NewDefaultCliConfig()
	cfg.SuppressDeprecationWarnings = true
	if !cfg.SuppressDeprecationWarnings {
		t.Error("Expected SuppressDeprecationWarnings to be settable to true")
	}
}
