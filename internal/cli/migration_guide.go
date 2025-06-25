package cli

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MigrationGuideGenerator generates migration guides from deprecation telemetry
// Implements hybrid template+algorithmic approach for O(1) common cases
type MigrationGuideGenerator struct {
	templates    *TemplateRegistry
	customRules  []*MigrationRule
	cache        map[string]*MigrationGuide
	stringIntern map[string]string
}

// MigrationGuide represents a complete migration guide with suggestions and analysis
type MigrationGuide struct {
	Suggestions      []MigrationSuggestion
	CoOccurrenceData *CoOccurrenceAnalysis
	GeneratedAt      time.Time
}

// MigrationSuggestion represents a single migration recommendation
type MigrationSuggestion struct {
	DeprecatedPattern string
	Steps             []MigrationStep
	Confidence        ConfidenceLevel
	Priority          int
	UsageCount        int
}

// MigrationStep represents a specific migration action
type MigrationStep struct {
	Description     string
	OldPattern      string
	NewPattern      string
	AutomationLevel AutomationLevel
}

// MigrationRule represents a custom migration rule
type MigrationRule struct {
	FlagPattern   string
	TargetFlag    string
	Confidence    ConfidenceLevel
	Description   string
	TransformFunc func(string) string
}

// TemplateRegistry manages migration templates for O(1) lookup
type TemplateRegistry struct {
	templates map[string]*MigrationTemplate
}

// MigrationTemplate represents a pre-defined migration pattern
type MigrationTemplate struct {
	Pattern    string
	Steps      []MigrationStep
	Confidence ConfidenceLevel
}

// CoOccurrenceAnalysis provides flag correlation data
type CoOccurrenceAnalysis struct {
	correlations map[string]map[string]float64
}

// ConfidenceLevel represents automation confidence
type ConfidenceLevel int

const (
	ConfidenceUnknown ConfidenceLevel = iota
	ConfidenceLow
	ConfidenceMedium
	ConfidenceHigh
)

// AutomationLevel represents how automated a migration can be
type AutomationLevel int

const (
	AutomationManual AutomationLevel = iota
	AutomationSemiAuto
	AutomationFull
)

// NewMigrationGuideGenerator creates a new generator with built-in templates
func NewMigrationGuideGenerator() *MigrationGuideGenerator {
	generator := &MigrationGuideGenerator{
		templates:    &TemplateRegistry{templates: make(map[string]*MigrationTemplate)},
		customRules:  make([]*MigrationRule, 0),
		cache:        make(map[string]*MigrationGuide),
		stringIntern: make(map[string]string),
	}

	// Load built-in templates for common thinktank patterns
	generator.loadBuiltinTemplates()

	return generator
}

// GenerateFromTelemetry creates a migration guide from deprecation telemetry data
func (g *MigrationGuideGenerator) GenerateFromTelemetry(telemetry *DeprecationTelemetry) *MigrationGuide {
	// Check cache first for performance
	cacheKey := g.computeCacheKey(telemetry)
	if cached, exists := g.cache[cacheKey]; exists {
		return cached
	}

	suggestions := make([]MigrationSuggestion, 0)

	// Get usage patterns from telemetry
	patterns := telemetry.GetUsagePatterns()

	// Compute co-occurrence analysis for contextual suggestions
	coOccurrenceData := g.computeCoOccurrence(patterns)

	// Generate suggestions for each pattern
	for _, pattern := range patterns {
		suggestion := g.generateSuggestionForPattern(pattern)
		if suggestion != nil {
			// Enhance suggestion with co-occurrence data
			g.enhanceSuggestionWithContext(suggestion, coOccurrenceData, patterns)
			suggestions = append(suggestions, *suggestion)
		}
	}

	// Sort by priority (usage count descending) for impact-based ordering
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Priority > suggestions[j].Priority
	})

	guide := &MigrationGuide{
		Suggestions:      suggestions,
		CoOccurrenceData: coOccurrenceData,
		GeneratedAt:      time.Now(),
	}

	// Cache the result for subsequent calls
	g.cache[cacheKey] = guide

	return guide
}

// GetBuiltinTemplates returns the template registry
func (g *MigrationGuideGenerator) GetBuiltinTemplates() *TemplateRegistry {
	return g.templates
}

// AddCustomRule adds a custom migration rule to the generator
func (g *MigrationGuideGenerator) AddCustomRule(rule *MigrationRule) {
	g.customRules = append(g.customRules, rule)
}

// IsSortedByPriority checks if suggestions are sorted by priority
func (guide *MigrationGuide) IsSortedByPriority() bool {
	if len(guide.Suggestions) <= 1 {
		return true
	}
	for i := 1; i < len(guide.Suggestions); i++ {
		if guide.Suggestions[i-1].Priority < guide.Suggestions[i].Priority {
			return false
		}
	}
	return true
}

// GetCoOccurrenceAnalysis returns the co-occurrence analysis data
func (guide *MigrationGuide) GetCoOccurrenceAnalysis() *CoOccurrenceAnalysis {
	return guide.CoOccurrenceData
}

// FormatAsText returns a human-readable text format
func (guide *MigrationGuide) FormatAsText() string {
	if len(guide.Suggestions) == 0 {
		return "Migration Guide: No deprecated patterns found."
	}

	output := "Migration Guide:\n\n"
	for i, suggestion := range guide.Suggestions {
		output += fmt.Sprintf("%d. Replace '%s' (used %d times)\n",
			i+1, suggestion.DeprecatedPattern, suggestion.UsageCount)

		for _, step := range suggestion.Steps {
			output += fmt.Sprintf("   • %s\n", step.Description)
			output += fmt.Sprintf("     Change '%s' to '%s'\n", step.OldPattern, step.NewPattern)
		}
		output += "\n"
	}

	return output
}

// FormatAsJSON returns a JSON representation
func (guide *MigrationGuide) FormatAsJSON() string {
	if len(guide.Suggestions) == 0 {
		return `{"suggestions": []}`
	}
	// Simple JSON format for now
	return fmt.Sprintf(`{"suggestions": [{"count": %d}]}`, len(guide.Suggestions))
}

// FormatAsMarkdown returns a Markdown representation
func (guide *MigrationGuide) FormatAsMarkdown() string {
	if len(guide.Suggestions) == 0 {
		return "# Migration Guide\n\nNo suggestions available."
	}

	output := "# Migration Guide\n\n"
	for i, suggestion := range guide.Suggestions {
		output += fmt.Sprintf("## %d. %s\n\n", i+1, suggestion.DeprecatedPattern)
		output += fmt.Sprintf("**Usage Count:** %d\n\n", suggestion.UsageCount)

		for _, step := range suggestion.Steps {
			output += fmt.Sprintf("- %s\n", step.Description)
		}
		output += "\n"
	}

	return output
}

// GetTemplate returns a specific template by flag name
func (tr *TemplateRegistry) GetTemplate(flag string) *MigrationTemplate {
	return tr.templates[flag]
}

// FindBestMatch finds the best matching template for a pattern
func (tr *TemplateRegistry) FindBestMatch(pattern string) *MigrationTemplate {
	for _, template := range tr.templates {
		if strings.Contains(pattern, template.Pattern) {
			return template
		}
	}
	return nil
}

// GetCorrelation returns the correlation between two flags
func (coa *CoOccurrenceAnalysis) GetCorrelation(flag1, flag2 string) float64 {
	if correlations, exists := coa.correlations[flag1]; exists {
		if correlation, exists := correlations[flag2]; exists {
			return correlation
		}
	}
	return 0.0
}

// Private helper methods

func (g *MigrationGuideGenerator) loadBuiltinTemplates() {
	// Template for --instructions flag
	g.templates.templates["--instructions"] = &MigrationTemplate{
		Pattern:    "--instructions",
		Confidence: ConfidenceHigh,
		Steps: []MigrationStep{
			{
				Description:     "Replace --instructions flag with new format",
				OldPattern:      "--instructions",
				NewPattern:      "--instruction-file",
				AutomationLevel: AutomationSemiAuto,
			},
		},
	}

	// Template for --model flag
	g.templates.templates["--model"] = &MigrationTemplate{
		Pattern:    "--model",
		Confidence: ConfidenceHigh,
		Steps: []MigrationStep{
			{
				Description:     "Update model specification format",
				OldPattern:      "--model",
				NewPattern:      "--ai-model",
				AutomationLevel: AutomationSemiAuto,
			},
		},
	}
}

func (g *MigrationGuideGenerator) generateSuggestionForPattern(pattern UsagePattern) *MigrationSuggestion {
	// Use string interning to reduce memory usage
	internedPattern := g.intern(pattern.Pattern)

	// Try to find a template match first (O(1) lookup)
	for templateFlag, template := range g.templates.templates {
		if strings.Contains(internedPattern, templateFlag) {
			return &MigrationSuggestion{
				DeprecatedPattern: internedPattern,
				Steps:             template.Steps,
				Confidence:        template.Confidence,
				Priority:          pattern.Count,
				UsageCount:        pattern.Count,
			}
		}
	}

	// Apply custom rules if no template matches
	for _, rule := range g.customRules {
		if g.matchesRule(internedPattern, rule) {
			steps := []MigrationStep{
				{
					Description:     g.intern(rule.Description),
					OldPattern:      internedPattern,
					NewPattern:      g.intern(rule.TransformFunc(internedPattern)),
					AutomationLevel: AutomationSemiAuto,
				},
			}
			return &MigrationSuggestion{
				DeprecatedPattern: internedPattern,
				Steps:             steps,
				Confidence:        rule.Confidence,
				Priority:          pattern.Count,
				UsageCount:        pattern.Count,
			}
		}
	}

	return nil
}

func (g *MigrationGuideGenerator) matchesRule(pattern string, rule *MigrationRule) bool {
	// Simple wildcard matching
	if strings.Contains(rule.FlagPattern, "*") {
		prefix := strings.TrimSuffix(rule.FlagPattern, "*")
		return strings.Contains(pattern, prefix)
	}
	return strings.Contains(pattern, rule.FlagPattern)
}

func (g *MigrationGuideGenerator) computeCoOccurrence(patterns []UsagePattern) *CoOccurrenceAnalysis {
	analysis := &CoOccurrenceAnalysis{correlations: make(map[string]map[string]float64)}

	// Extract flags from patterns
	flagCounts := make(map[string]int)
	flagCoOccurrences := make(map[string]map[string]int)

	for _, pattern := range patterns {
		flags := extractFlags(pattern.Pattern)

		// Count individual flag usage
		for _, flag := range flags {
			flagCounts[flag] += pattern.Count
		}

		// Count co-occurrences
		for i, flag1 := range flags {
			if _, exists := flagCoOccurrences[flag1]; !exists {
				flagCoOccurrences[flag1] = make(map[string]int)
			}
			for j, flag2 := range flags {
				if i != j {
					flagCoOccurrences[flag1][flag2] += pattern.Count
				}
			}
		}
	}

	// Compute correlations
	for flag1, coOccurrences := range flagCoOccurrences {
		if _, exists := analysis.correlations[flag1]; !exists {
			analysis.correlations[flag1] = make(map[string]float64)
		}

		for flag2, coCount := range coOccurrences {
			// Simple correlation: co-occurrence / min(count1, count2)
			correlation := float64(coCount) / float64(min(flagCounts[flag1], flagCounts[flag2]))
			analysis.correlations[flag1][flag2] = correlation
		}
	}

	return analysis
}

func (g *MigrationGuideGenerator) enhanceSuggestionWithContext(suggestion *MigrationSuggestion, analysis *CoOccurrenceAnalysis, patterns []UsagePattern) {
	// Check if this suggestion involves flags that commonly appear together
	patternFlags := extractFlags(suggestion.DeprecatedPattern)
	relatedFlags := make([]string, 0)

	for _, flag := range patternFlags {
		if correlations, exists := analysis.correlations[flag]; exists {
			for otherFlag, correlation := range correlations {
				if correlation > 0.5 { // High correlation threshold
					relatedFlags = append(relatedFlags, otherFlag)
				}
			}
		}
	}

	// If we found related flags, enhance the suggestion
	if len(relatedFlags) > 0 {
		for i := range suggestion.Steps {
			// Add both the original flag and related flags to the description
			allFlags := append(patternFlags, relatedFlags...)
			if len(allFlags) >= 2 {
				suggestion.Steps[i].Description += fmt.Sprintf(" (consider migrating together with %s)", strings.Join(relatedFlags, ", "))
			}
		}
	}
}

func (g *MigrationGuideGenerator) computeCacheKey(telemetry *DeprecationTelemetry) string {
	// Create a hash of the telemetry data for caching
	patterns := telemetry.GetUsagePatterns()

	// Sort patterns for consistent hashing
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Pattern < patterns[j].Pattern
	})

	// Create a string representation of the data
	var keyData strings.Builder
	for _, pattern := range patterns {
		keyData.WriteString(fmt.Sprintf("%s:%d;", pattern.Pattern, pattern.Count))
	}

	// Hash the data
	hash := md5.Sum([]byte(keyData.String()))
	return fmt.Sprintf("%x", hash)
}

// String interning for memory efficiency
func (g *MigrationGuideGenerator) intern(s string) string {
	if interned, exists := g.stringIntern[s]; exists {
		return interned
	}
	g.stringIntern[s] = s
	return s
}

func extractFlags(pattern string) []string {
	fields := strings.Fields(pattern)
	flags := make([]string, 0)

	for _, field := range fields {
		if strings.HasPrefix(field, "--") {
			flags = append(flags, field)
		}
	}

	return flags
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
