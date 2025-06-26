package cli

import (
	"errors"
	"time"
)

// ConfidenceLevel represents the confidence level of a migration suggestion
type ConfidenceLevel int

const (
	ConfidenceUnknown ConfidenceLevel = iota
	ConfidenceLow
	ConfidenceMedium
	ConfidenceHigh
)

// MigrationSuggestion represents a single migration recommendation
type MigrationSuggestion struct {
	Pattern    string          `json:"pattern"`     // Original deprecated pattern
	Suggestion string          `json:"suggestion"`  // Suggested replacement
	Confidence ConfidenceLevel `json:"confidence"`  // Confidence level of suggestion
	UsageCount int             `json:"usage_count"` // Number of times pattern was used
}

// MigrationGuide represents a complete migration guide with multiple suggestions
type MigrationGuide struct {
	GeneratedAt   time.Time             `json:"generated_at"`   // When guide was generated
	Version       string                `json:"version"`        // Generator version
	TotalPatterns int                   `json:"total_patterns"` // Number of patterns analyzed
	Suggestions   []MigrationSuggestion `json:"suggestions"`    // Migration suggestions
}

// MigrationGuideGenerator generates migration guides from deprecation telemetry
// Following Rob Pike's simplicity principle: small, focused types
type MigrationGuideGenerator struct {
	// Will be populated with template registry and rule engine in later phases
}

// NewMigrationGuideGenerator creates a new migration guide generator
// Following TDD: minimal implementation to make tests pass
func NewMigrationGuideGenerator() *MigrationGuideGenerator {
	return &MigrationGuideGenerator{}
}

// GenerateFromTelemetry generates a migration guide from deprecation telemetry data
// GREEN: Basic implementation to make tests pass
func (g *MigrationGuideGenerator) GenerateFromTelemetry(telemetry *DeprecationTelemetry) (*MigrationGuide, error) {
	if telemetry == nil {
		return nil, errors.New("telemetry cannot be nil")
	}

	patterns := telemetry.GetUsagePatterns()
	suggestions := make([]MigrationSuggestion, 0, len(patterns))

	// Basic template matching for common patterns
	for _, pattern := range patterns {
		suggestion := g.generateSuggestion(pattern)
		if suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	guide := &MigrationGuide{
		GeneratedAt:   time.Now(),
		Version:       "0.1.0",
		TotalPatterns: len(patterns),
		Suggestions:   suggestions,
	}

	return guide, nil
}

// generateSuggestion creates a migration suggestion for a usage pattern
// GREEN: Basic template matching implementation
func (g *MigrationGuideGenerator) generateSuggestion(pattern UsagePattern) *MigrationSuggestion {
	suggestion := &MigrationSuggestion{
		Pattern:    pattern.Pattern,
		UsageCount: pattern.Count,
	}

	// Basic template matching for common thinktank patterns
	switch {
	case containsFlag(pattern.Args, "--instructions"):
		// Extract the file argument
		file := extractFlagValue(pattern.Args, "--instructions")
		sb := GetStringBuilder()
		defer PutStringBuilder(sb)
		sb.WriteString("thinktank ")
		sb.WriteString(file)
		sb.WriteString(" target_path")
		suggestion.Suggestion = sb.String()
		suggestion.Confidence = ConfidenceHigh

	case containsFlag(pattern.Args, "--model"):
		model := InternModelName(extractFlagValue(pattern.Args, "--model"))
		sb := GetStringBuilder()
		defer PutStringBuilder(sb)
		sb.WriteString("thinktank instructions.txt target_path --model ")
		sb.WriteString(model)
		suggestion.Suggestion = sb.String()
		suggestion.Confidence = ConfidenceHigh

	case containsFlag(pattern.Args, "--output-dir"):
		outputDir := extractFlagValue(pattern.Args, "--output-dir")
		sb := GetStringBuilder()
		defer PutStringBuilder(sb)
		sb.WriteString("thinktank instructions.txt target_path --output-dir ")
		sb.WriteString(outputDir)
		suggestion.Suggestion = sb.String()
		suggestion.Confidence = ConfidenceMedium

	default:
		// Generic suggestion for unknown patterns
		suggestion.Suggestion = "Use simplified interface: thinktank instructions.txt target_path [flags...]"
		suggestion.Confidence = ConfidenceLow
	}

	return suggestion
}

// containsFlag checks if args contains a specific flag
func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// extractFlagValue extracts the value for a flag from args
func extractFlagValue(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "value" // Default fallback
}
