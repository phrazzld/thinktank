package architect

import (
	"testing"
)

// TestConvertConfigBasic verifies basic functionality of ConvertConfigToMap
func TestConvertConfigBasic(t *testing.T) {
	// Create a minimal config for testing
	config := &CliConfig{
		InstructionsFile: "instructions.txt",
		OutputFile:       "output.md",
	}

	// Convert to map
	configMap := ConvertConfigToMap(config)

	// Check that basic fields are included
	instructionsFile, hasInstructionsFile := configMap["instructionsFile"]
	if !hasInstructionsFile || instructionsFile != "instructions.txt" {
		t.Error("ConvertConfigToMap() should include instructionsFile key with correct value")
	}

	outputFile, hasOutputFile := configMap["output"]
	if !hasOutputFile || outputFile != "output.md" {
		t.Error("ConvertConfigToMap() should include output key with correct value")
	}
}
