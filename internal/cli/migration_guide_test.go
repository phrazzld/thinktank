package cli

import (
	"strings"
	"testing"
	"time"
)

// TestMigrationGuideGeneration tests the core migration guide generation functionality
// This drives toward a high-performance, template-based system with algorithmic fallbacks
func TestMigrationGuideGeneration(t *testing.T) {
	telemetry := NewDeprecationTelemetry()

	// Record realistic usage patterns that would need migration guides
	telemetry.RecordUsagePattern("--instructions file.txt path/",
		[]string{"--instructions", "file.txt", "path/"})
	telemetry.RecordUsagePattern("--model gpt-4",
		[]string{"--model", "gpt-4"})
	telemetry.RecordUsagePattern("--output-dir build/ --verbose",
		[]string{"--output-dir", "build/", "--verbose"})

	// Generate migration guide from telemetry data
	generator := NewMigrationGuideGenerator()
	guide := generator.GenerateFromTelemetry(telemetry)

	// Should produce actionable migration suggestions
	if guide == nil {
		t.Fatal("Expected migration guide to be generated")
	}

	if len(guide.Suggestions) == 0 {
		t.Error("Expected migration suggestions to be generated")
	}

	// Should prioritize by usage frequency (most common patterns first)
	if !guide.IsSortedByPriority() {
		t.Error("Expected migration suggestions sorted by priority/frequency")
	}

	// Should include specific migration steps for each pattern
	instructionsSuggestion := findSuggestionForPattern(guide.Suggestions, "--instructions")
	if instructionsSuggestion == nil {
		t.Fatal("Expected migration suggestion for --instructions pattern")
	}

	if len(instructionsSuggestion.Steps) == 0 {
		t.Error("Expected concrete migration steps")
	}

	// Should provide confidence levels (template-based = high, algorithmic = lower)
	if instructionsSuggestion.Confidence == ConfidenceUnknown {
		t.Error("Expected confidence level to be set")
	}
}

// TestMigrationGuidePerformance tests that guide generation is efficient for large datasets
// Following Carmack's performance principles: measure first, optimize critical paths
func TestMigrationGuidePerformance(t *testing.T) {
	telemetry := NewDeprecationTelemetry()

	// Simulate high-volume usage patterns (1000 patterns, 10k total usages)
	patterns := generateRealisticUsagePatterns(1000, 10000)
	for pattern, count := range patterns {
		for i := 0; i < count; i++ {
			args := strings.Fields(pattern)
			telemetry.RecordUsagePattern(pattern, args)
		}
	}

	generator := NewMigrationGuideGenerator()

	// Measure generation time
	start := time.Now()
	guide := generator.GenerateFromTelemetry(telemetry)
	duration := time.Since(start)

	// Should complete within reasonable time bounds (< 100ms for 10k patterns)
	if duration > 100*time.Millisecond {
		t.Errorf("Migration guide generation too slow: %v", duration)
	}

	// Should handle high-volume data without memory issues
	if guide == nil {
		t.Fatal("Guide generation failed under load")
	}

	// Should still prioritize correctly under high volume
	if len(guide.Suggestions) > 0 && !guide.IsSortedByPriority() {
		t.Error("Priority sorting failed under high volume")
	}
}

// TestMigrationGuideTemplateSystem tests the template-based migration approach
// Templates provide O(1) lookup for common patterns vs O(n) algorithmic analysis
func TestMigrationGuideTemplateSystem(t *testing.T) {
	generator := NewMigrationGuideGenerator()

	// Load built-in templates for common migration patterns
	templates := generator.GetBuiltinTemplates()
	if templates == nil || len(templates.templates) == 0 {
		t.Fatal("Expected built-in migration templates to be available")
	}

	// Should have template for --instructions flag migration
	instructionsTemplate := templates.GetTemplate("--instructions")
	if instructionsTemplate == nil {
		t.Fatal("Expected template for --instructions migration")
	}

	// Template should specify concrete migration steps
	if len(instructionsTemplate.Steps) == 0 {
		t.Error("Expected template to define migration steps")
	}

	// Should support pattern matching for template selection
	pattern := "--instructions file.txt path/"
	matchedTemplate := templates.FindBestMatch(pattern)
	if matchedTemplate == nil {
		t.Error("Expected template matching to find appropriate template")
		return
	}

	// Template confidence should be high (pre-defined, tested patterns)
	if matchedTemplate.Confidence != ConfidenceHigh {
		t.Error("Expected high confidence for template-based migrations")
	}
}

// TestMigrationGuideCustomRules tests extensible rule-based migration generation
// Allows adding new migration strategies without changing core engine
func TestMigrationGuideCustomRules(t *testing.T) {
	generator := NewMigrationGuideGenerator()

	// Define custom migration rule for project-specific deprecated flags
	customRule := &MigrationRule{
		FlagPattern:   "--legacy-*",
		TargetFlag:    "--new-*",
		Confidence:    ConfidenceMedium,
		Description:   "Legacy flags replaced with new- prefix",
		TransformFunc: legacyToNewTransform,
	}

	// Register custom rule
	generator.AddCustomRule(customRule)

	// Create telemetry with custom pattern
	telemetry := NewDeprecationTelemetry()
	telemetry.RecordUsagePattern("--legacy-output output/",
		[]string{"--legacy-output", "output/"})

	// Generate guide with custom rule
	guide := generator.GenerateFromTelemetry(telemetry)

	// Should apply custom rule to matching patterns
	legacySuggestion := findSuggestionForPattern(guide.Suggestions, "--legacy-output")
	if legacySuggestion == nil {
		t.Fatal("Expected custom rule to generate migration suggestion")
	}

	// Should use custom transformation
	expectedTarget := "--new-output output/"
	if !containsTargetPattern(legacySuggestion.Steps, expectedTarget) {
		t.Errorf("Expected custom transformation to target %s", expectedTarget)
	}
}

// TestMigrationGuideContextualAnalysis tests algorithmic analysis for complex patterns
// Handles cases where templates aren't sufficient - analyzes argument relationships
func TestMigrationGuideContextualAnalysis(t *testing.T) {
	telemetry := NewDeprecationTelemetry()

	// Record complex usage patterns that need contextual analysis
	telemetry.RecordUsagePattern("--instructions file.txt --model gpt-4 --output-dir build/",
		[]string{"--instructions", "file.txt", "--model", "gpt-4", "--output-dir", "build/"})
	telemetry.RecordUsagePattern("--instructions other.txt --verbose --model claude",
		[]string{"--instructions", "other.txt", "--verbose", "--model", "claude"})

	generator := NewMigrationGuideGenerator()
	guide := generator.GenerateFromTelemetry(telemetry)

	// Should detect flag co-occurrence patterns
	coOccurrenceAnalysis := guide.GetCoOccurrenceAnalysis()
	if coOccurrenceAnalysis == nil {
		t.Fatal("Expected co-occurrence analysis for complex patterns")
	}

	// Should identify that --instructions and --model are commonly used together
	instructionsModelCorrelation := coOccurrenceAnalysis.GetCorrelation("--instructions", "--model")
	if instructionsModelCorrelation < 0.5 {
		t.Errorf("Expected high correlation between --instructions and --model, got %f",
			instructionsModelCorrelation)
	}

	// Should provide context-aware migration suggestions
	contextualSuggestion := findSuggestionForPattern(guide.Suggestions, "--instructions")
	if contextualSuggestion == nil {
		t.Fatal("Expected contextual migration suggestion")
	}

	// Should suggest migrating related flags together
	if !suggestsGroupMigration(contextualSuggestion, []string{"--instructions", "--model"}) {
		t.Error("Expected suggestion to recommend migrating related flags together")
	}
}

// TestMigrationGuideOutput tests different output formats for migration guides
// Supports both human-readable and machine-processable formats
func TestMigrationGuideOutput(t *testing.T) {
	telemetry := NewDeprecationTelemetry()
	telemetry.RecordUsagePattern("--instructions file.txt",
		[]string{"--instructions", "file.txt"})

	generator := NewMigrationGuideGenerator()
	guide := generator.GenerateFromTelemetry(telemetry)

	// Should support human-readable text format
	textOutput := guide.FormatAsText()
	if textOutput == "" {
		t.Error("Expected text format output")
	}

	// Should include actionable steps in text format
	if !strings.Contains(textOutput, "Replace") && !strings.Contains(textOutput, "Change") {
		t.Error("Expected actionable language in text output")
	}

	// Should support structured JSON format for tooling integration
	jsonOutput := guide.FormatAsJSON()
	if jsonOutput == "" {
		t.Error("Expected JSON format output")
	}

	// Should support Markdown format for documentation
	markdownOutput := guide.FormatAsMarkdown()
	if markdownOutput == "" {
		t.Error("Expected Markdown format output")
	}

	// Markdown should include proper formatting
	if !strings.Contains(markdownOutput, "#") && !strings.Contains(markdownOutput, "##") {
		t.Error("Expected Markdown headers in output")
	}
}

// TestMigrationGuideMemoryEfficiency tests memory usage under high load
// Ensures the system doesn't consume excessive memory for large migration guides
func TestMigrationGuideMemoryEfficiency(t *testing.T) {
	// This test would measure memory allocation and usage patterns
	// Implementation would depend on runtime.MemStats and profiling tools
	t.Skip("Memory efficiency testing requires profiling setup")
}

// Test helper functions - these define the expected API

func generateRealisticUsagePatterns(numPatterns, totalUsages int) map[string]int {
	// Generate Zipfian distribution of usage patterns (realistic CLI usage)
	patterns := make(map[string]int)

	// Common patterns get more usage
	commonPatterns := []string{
		"--instructions file.txt path/",
		"--model gpt-4",
		"--output-dir build/",
		"--verbose",
		"--include *.go",
	}

	usageRemaining := totalUsages
	for i, pattern := range commonPatterns {
		if i >= numPatterns {
			break
		}
		// Zipfian: first pattern gets ~50% usage, second ~25%, etc.
		usage := usageRemaining / (2 * (i + 1))
		patterns[pattern] = usage
		usageRemaining -= usage
	}

	// Fill remaining patterns with low usage
	for i := len(commonPatterns); i < numPatterns && usageRemaining > 0; i++ {
		pattern := generateRandomPattern(i)
		usage := 1
		if usageRemaining > 0 {
			patterns[pattern] = usage
			usageRemaining--
		}
	}

	return patterns
}

func generateRandomPattern(seed int) string {
	// Generate diverse patterns for testing
	flags := []string{"--debug", "--trace", "--config", "--timeout", "--format"}
	return flags[seed%len(flags)]
}

func findSuggestionForPattern(suggestions []MigrationSuggestion, pattern string) *MigrationSuggestion {
	for _, suggestion := range suggestions {
		if strings.Contains(suggestion.DeprecatedPattern, pattern) {
			return &suggestion
		}
	}
	return nil
}

func containsTargetPattern(steps []MigrationStep, target string) bool {
	for _, step := range steps {
		if strings.Contains(step.NewPattern, target) {
			return true
		}
	}
	return false
}

func suggestsGroupMigration(suggestion *MigrationSuggestion, flags []string) bool {
	// Check if the suggestion recommends migrating multiple flags together
	for _, step := range suggestion.Steps {
		flagCount := 0
		for _, flag := range flags {
			if strings.Contains(step.Description, flag) {
				flagCount++
			}
		}
		if flagCount >= 2 {
			return true
		}
	}
	return false
}

func legacyToNewTransform(pattern string) string {
	// Example transformation function for custom rules
	return strings.ReplaceAll(pattern, "--legacy-", "--new-")
}

// Test helper functions for generating realistic data and validation
