package orchestrator

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
)

func TestFormatWithCommas(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "small number",
			input:    123,
			expected: "123",
		},
		{
			name:     "thousand",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "million",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "negative number",
			input:    -1234567,
			expected: "-1,234,567",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWithCommas(tt.input)
			if result != tt.expected {
				t.Errorf("formatWithCommas(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDisplayCompatibilityCard(t *testing.T) {
	tests := []struct {
		name        string
		analysis    CompatibilityAnalysis
		verbose     bool
		expectedOut []string
	}{
		{
			name: "all models compatible",
			analysis: CompatibilityAnalysis{
				TotalModels:      3,
				CompatibleModels: 3,
				SkippedModels:    0,
				TotalTokens:      1000,
				SafetyThreshold:  80,
				MinUtilization:   45.2,
				MaxUtilization:   65.8,
				BestModel: ModelCompatibilityInfo{
					ModelName:     "gpt-4o-mini",
					ContextWindow: 2650,
					TokenCount:    1200,
					Utilization:   45.2,
					IsCompatible:  true,
				},
				WorstModel: ModelCompatibilityInfo{
					ModelName:     "claude-3-opus",
					ContextWindow: 1835,
					TokenCount:    1200,
					Utilization:   65.8,
					IsCompatible:  true,
				},
				AllModels: []ModelCompatibilityInfo{
					{ModelName: "gpt-4o-mini", ContextWindow: 2650, TokenCount: 1200, Utilization: 45.2, IsCompatible: true},
					{ModelName: "gpt-4o", ContextWindow: 2300, TokenCount: 1200, Utilization: 52.1, IsCompatible: true},
					{ModelName: "claude-3-opus", ContextWindow: 1835, TokenCount: 1200, Utilization: 65.8, IsCompatible: true},
				},
			},
			verbose: false,
			expectedOut: []string{
				"3/3 models compatible",
				"Context usage range: 45.2% - 65.8%",
			},
		},
		{
			name: "no compatible models",
			analysis: CompatibilityAnalysis{
				TotalModels:      2,
				CompatibleModels: 0,
				SkippedModels:    2,
				TotalTokens:      5000,
				SafetyThreshold:  80,
				AllModels: []ModelCompatibilityInfo{
					{ModelName: "model1", ContextWindow: 4000, TokenCount: 5000, Utilization: 125.0, IsCompatible: false},
					{ModelName: "model2", ContextWindow: 4000, TokenCount: 5000, Utilization: 125.0, IsCompatible: false},
				},
			},
			verbose: false,
			expectedOut: []string{
				"No compatible models (2/2 exceed 80% context usage)",
				"💡 Try reducing input size:",
				"• thinktank instructions.txt ./src",
				"• --exclude \"docs/,*.md,build/\"",
				"• --dry-run",
			},
		},
		{
			name: "some models compatible with verbose",
			analysis: CompatibilityAnalysis{
				TotalModels:      3,
				CompatibleModels: 2,
				SkippedModels:    1,
				TotalTokens:      1000,
				SafetyThreshold:  80,
				MinUtilization:   45.2,
				MaxUtilization:   95.8,
				BestModel: ModelCompatibilityInfo{
					ModelName:     "gpt-4o-mini",
					ContextWindow: 2650,
					TokenCount:    1200,
					Utilization:   45.2,
					IsCompatible:  true,
				},
				WorstModel: ModelCompatibilityInfo{
					ModelName:     "claude-3-opus",
					ContextWindow: 1254,
					TokenCount:    1200,
					Utilization:   95.8,
					IsCompatible:  false,
				},
				AllModels: []ModelCompatibilityInfo{
					{ModelName: "gpt-4o-mini", ContextWindow: 2650, TokenCount: 1200, Utilization: 45.2, IsCompatible: true},
					{ModelName: "gpt-4o", ContextWindow: 2300, TokenCount: 1200, Utilization: 52.1, IsCompatible: true},
					{ModelName: "claude-3-opus", ContextWindow: 1254, TokenCount: 1200, Utilization: 95.8, IsCompatible: false},
				},
			},
			verbose: true,
			expectedOut: []string{
				"2/3 models compatible (1 skipped)",
				"Context usage range: 45.2% - 95.8%",
				"gpt-4o-mini",
				"gpt-4o",
				"claude-3-opus",
			},
		},
		{
			name: "single model",
			analysis: CompatibilityAnalysis{
				TotalModels:      1,
				CompatibleModels: 1,
				SkippedModels:    0,
				TotalTokens:      1000,
				SafetyThreshold:  80,
				AllModels: []ModelCompatibilityInfo{
					{ModelName: "gpt-4o", ContextWindow: 2300, TokenCount: 1200, Utilization: 52.1, IsCompatible: true},
				},
			},
			verbose: false,
			expectedOut: []string{
				"1/1 models compatible",
				"Context usage: 52.1% (gpt-4o)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create mock orchestrator with verbose setting
			cfg := &config.CliConfig{
				Verbose: tt.verbose,
			}
			orch := &Orchestrator{config: cfg}

			// Call the function
			orch.displayCompatibilityCard(tt.analysis)

			// Restore stdout and read captured output
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe writer: %v", err)
			}
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to copy output: %v", err)
			}
			output := buf.String()

			// Check for expected strings
			for _, expected := range tt.expectedOut {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestModelCompatibilityInfo(t *testing.T) {
	info := ModelCompatibilityInfo{
		ModelName:     "test-model",
		ContextWindow: 4000,
		TokenCount:    2000,
		Utilization:   50.0,
		IsCompatible:  true,
		FailureReason: "",
	}

	// Test the struct fields
	if info.ModelName != "test-model" {
		t.Errorf("Expected ModelName to be 'test-model', got %q", info.ModelName)
	}
	if info.ContextWindow != 4000 {
		t.Errorf("Expected ContextWindow to be 4000, got %d", info.ContextWindow)
	}
	if info.TokenCount != 2000 {
		t.Errorf("Expected TokenCount to be 2000, got %d", info.TokenCount)
	}
	if info.Utilization != 50.0 {
		t.Errorf("Expected Utilization to be 50.0, got %f", info.Utilization)
	}
	if !info.IsCompatible {
		t.Errorf("Expected IsCompatible to be true, got %v", info.IsCompatible)
	}
}

func TestCompatibilityAnalysis(t *testing.T) {
	analysis := CompatibilityAnalysis{
		TotalModels:      5,
		CompatibleModels: 3,
		SkippedModels:    2,
		TotalTokens:      1500,
		SafetyThreshold:  80.0,
		MinUtilization:   25.5,
		MaxUtilization:   85.2,
		CompatibleList:   []string{"model1", "model2", "model3"},
		SkippedList:      []string{"model4", "model5"},
	}

	// Test the struct fields
	if analysis.TotalModels != 5 {
		t.Errorf("Expected TotalModels to be 5, got %d", analysis.TotalModels)
	}
	if analysis.CompatibleModels != 3 {
		t.Errorf("Expected CompatibleModels to be 3, got %d", analysis.CompatibleModels)
	}
	if analysis.SkippedModels != 2 {
		t.Errorf("Expected SkippedModels to be 2, got %d", analysis.SkippedModels)
	}
	if len(analysis.CompatibleList) != 3 {
		t.Errorf("Expected CompatibleList to have 3 items, got %d", len(analysis.CompatibleList))
	}
	if len(analysis.SkippedList) != 2 {
		t.Errorf("Expected SkippedList to have 2 items, got %d", len(analysis.SkippedList))
	}
}
