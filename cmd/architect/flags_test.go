package architect

import (
	"testing"
)

// TestConvertConfigNoClarity verifies that clarify_task is not included in the config map
func TestConvertConfigNoClarity(t *testing.T) {
	// Create a minimal config for testing
	config := &CliConfig{
		TaskFile:   "task.txt",
		OutputFile: "output.md",
	}

	// Convert to map
	configMap := ConvertConfigToMap(config)

	// Check that clarify_task is not in the map
	_, hasClarifyTask := configMap["clarify_task"]
	if hasClarifyTask {
		t.Error("ConvertConfigToMap() should not include clarify_task key")
	}
}
