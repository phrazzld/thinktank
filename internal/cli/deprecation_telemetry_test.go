package cli

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestDeprecationTelemetryBasics tests the basic functionality of the telemetry system
func TestDeprecationTelemetryBasics(t *testing.T) {
	// This test will fail until we implement the telemetry system
	telemetry := NewDeprecationTelemetry()

	// Record some usage
	telemetry.RecordUsage("--instructions")
	telemetry.RecordUsage("--model")
	telemetry.RecordUsage("--instructions") // Record again

	// Get usage statistics
	stats := telemetry.GetUsageStats()

	// Verify counting
	if stats["--instructions"] != 2 {
		t.Errorf("Expected 2 uses of --instructions, got %d", stats["--instructions"])
	}

	if stats["--model"] != 1 {
		t.Errorf("Expected 1 use of --model, got %d", stats["--model"])
	}

	// Test reset functionality
	telemetry.Reset()
	statsAfterReset := telemetry.GetUsageStats()

	if len(statsAfterReset) != 0 {
		t.Errorf("Expected empty stats after reset, got %v", statsAfterReset)
	}
}

// TestDeprecationTelemetryUsagePatterns tests tracking of usage patterns
func TestDeprecationTelemetryUsagePatterns(t *testing.T) {
	telemetry := NewDeprecationTelemetry()

	// Record usage patterns with timestamps
	telemetry.RecordUsagePattern("--instructions file.txt path", []string{"--instructions", "file.txt", "path"})
	telemetry.RecordUsagePattern("--model gpt-4", []string{"--model", "gpt-4"})
	telemetry.RecordUsagePattern("--instructions file.txt path", []string{"--instructions", "file.txt", "path"}) // Same pattern

	// Get pattern statistics
	patterns := telemetry.GetUsagePatterns()

	// Should have 2 distinct patterns
	if len(patterns) != 2 {
		t.Errorf("Expected 2 distinct patterns, got %d", len(patterns))
	}

	// Find the --instructions pattern
	var instructionsPattern *UsagePattern
	for _, pattern := range patterns {
		if pattern.Pattern == "--instructions file.txt path" {
			instructionsPattern = &pattern
			break
		}
	}

	if instructionsPattern == nil {
		t.Fatal("Expected to find --instructions pattern")
	}

	if instructionsPattern.Count != 2 {
		t.Errorf("Expected --instructions pattern count 2, got %d", instructionsPattern.Count)
	}
}

// TestDeprecationTelemetryMostCommonFlags tests identification of most commonly used flags
func TestDeprecationTelemetryMostCommonFlags(t *testing.T) {
	telemetry := NewDeprecationTelemetry()

	// Simulate realistic usage distribution
	flagUsage := map[string]int{
		"--instructions": 50,
		"--model":        25,
		"--output-dir":   15,
		"--verbose":      10,
		"--include":      5,
	}

	// Record usage
	for flag, count := range flagUsage {
		for i := 0; i < count; i++ {
			telemetry.RecordUsage(flag)
		}
	}

	// Get most common flags
	mostCommon := telemetry.GetMostCommonFlags(3)

	// Should return top 3
	if len(mostCommon) != 3 {
		t.Errorf("Expected 3 most common flags, got %d", len(mostCommon))
	}

	// Should be in descending order
	expectedOrder := []string{"--instructions", "--model", "--output-dir"}
	for i, expected := range expectedOrder {
		if mostCommon[i].Flag != expected {
			t.Errorf("Expected flag %d to be %s, got %s", i, expected, mostCommon[i].Flag)
		}
	}

	// Verify counts
	if mostCommon[0].Count != 50 {
		t.Errorf("Expected --instructions count 50, got %d", mostCommon[0].Count)
	}
}

// TestDeprecationTelemetryThreadSafety tests concurrent usage tracking
func TestDeprecationTelemetryThreadSafety(t *testing.T) {
	telemetry := NewDeprecationTelemetry()

	// Run concurrent recording
	done := make(chan bool, 10)

	// Start 10 goroutines recording usage
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				telemetry.RecordUsage("--test-flag")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have recorded 1000 total uses
	stats := telemetry.GetUsageStats()
	if stats["--test-flag"] != 1000 {
		t.Errorf("Expected 1000 uses of --test-flag, got %d", stats["--test-flag"])
	}
}

// TestDeprecationTelemetryPersistence tests saving and loading telemetry data
func TestDeprecationTelemetryPersistence(t *testing.T) {
	t.Skip("Persistence not yet implemented")
}

// TestParserRouterWithTelemetry tests that the parser router can collect telemetry
func TestParserRouterWithTelemetry(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	// Create router with telemetry enabled
	mockLogger := testutil.NewMockLogger()
	router := NewParserRouterWithTelemetry(mockLogger, true)

	// Parse some arguments that should trigger telemetry
	testCases := [][]string{
		{"thinktank", "--instructions", instructionsFile, targetDir},
		{"thinktank", "--model", "gpt-4.1", instructionsFile, targetDir},
		{"thinktank", "--instructions", instructionsFile, "--output-dir", "output", targetDir},
	}

	for _, args := range testCases {
		result := router.ParseArguments(args)
		if !result.IsSuccess() {
			t.Errorf("Parsing failed for args %v: %v", args, result.Error)
		}
	}

	// Get telemetry data
	telemetry := router.GetTelemetry()
	if telemetry == nil {
		t.Fatal("Expected telemetry to be available")
	}

	stats := telemetry.GetUsageStats()

	// Should have recorded usage
	if stats["--instructions"] < 2 {
		t.Errorf("Expected at least 2 uses of --instructions, got %d", stats["--instructions"])
	}

	// Should have recorded complex flag usage
	if stats["complex_flags"] == 0 {
		t.Error("Expected some complex_flags usage to be recorded")
	}

	// Verify that parsing was logged (validates that logger was used correctly)
	logMessages := mockLogger.GetMessages()
	var foundParsingLog bool
	for _, msg := range logMessages {
		if strings.Contains(msg, "Detected parsing mode") || strings.Contains(msg, "parsed using") {
			foundParsingLog = true
			break
		}
	}
	if !foundParsingLog {
		t.Error("Expected to find parsing-related log messages")
	}
}

// TestTelemetryDisabledByDefault tests that telemetry doesn't run by default
func TestTelemetryDisabledByDefault(t *testing.T) {
	// Create router with default settings (telemetry disabled)
	mockLogger := testutil.NewMockLogger()
	router := NewParserRouter(mockLogger)

	// This should work but not collect telemetry
	args := []string{"thinktank", "--instructions", "test.txt", "src/"}
	result := router.ParseArguments(args)

	// Should parse successfully but not have telemetry
	if result.Error != nil {
		t.Errorf("Parsing failed: %v", result.Error)
	}

	// Should not have telemetry available
	if router.GetTelemetry() != nil {
		t.Error("Expected no telemetry when disabled by default")
	}

	// Verify that parsing was still logged even with telemetry disabled
	logMessages := mockLogger.GetMessages()
	var foundParsingLog bool
	for _, msg := range logMessages {
		if strings.Contains(msg, "Detected parsing mode") || strings.Contains(msg, "parsed using") {
			foundParsingLog = true
			break
		}
	}
	if !foundParsingLog {
		t.Error("Expected to find parsing-related log messages even with telemetry disabled")
	}
}

// Helper functions for router integration are now implemented in parser_router.go
