// Package orchestrator provides compatibility analysis and display functionality.
package orchestrator

import (
	"fmt"
	"strconv"
	"strings"
)

// ModelCompatibilityInfo holds detailed information about a model's compatibility
type ModelCompatibilityInfo struct {
	ModelName     string
	ContextWindow int
	TokenCount    int
	Utilization   float64
	IsCompatible  bool
	FailureReason string
}

// CompatibilityAnalysis holds the complete analysis results for display
type CompatibilityAnalysis struct {
	TotalModels      int
	CompatibleModels int
	SkippedModels    int
	TotalTokens      int
	SafetyThreshold  float64
	MinUtilization   float64
	MaxUtilization   float64
	BestModel        ModelCompatibilityInfo
	WorstModel       ModelCompatibilityInfo
	CompatibleList   []string
	SkippedList      []string
	AllModels        []ModelCompatibilityInfo
}

// displayCompatibilityCard shows a clean, modern compatibility analysis
func (o *Orchestrator) displayCompatibilityCard(analysis CompatibilityAnalysis) {
	fmt.Println()

	// Main status line - clean and prominent
	if analysis.CompatibleModels == 0 {
		fmt.Printf("⚠️  No compatible models (%d/%d exceed %.0f%% context usage)\n",
			analysis.SkippedModels, analysis.TotalModels, analysis.SafetyThreshold)
	} else {
		fmt.Printf("✅ %d/%d models compatible", analysis.CompatibleModels, analysis.TotalModels)
		if analysis.SkippedModels > 0 {
			fmt.Printf(" (%d skipped)", analysis.SkippedModels)
		}
		fmt.Println()
	}

	// Usage statistics - concise and informative
	if analysis.TotalModels > 1 {
		fmt.Printf("   Context usage: %.1f%% - %.1f%%", analysis.MinUtilization, analysis.MaxUtilization)

		// Show extremes only if there's meaningful difference
		if analysis.MaxUtilization-analysis.MinUtilization > 10.0 {
			fmt.Printf(" (best: %s, worst: %s)", analysis.BestModel.ModelName, analysis.WorstModel.ModelName)
		}
		fmt.Println()
	} else if len(analysis.AllModels) == 1 {
		model := analysis.AllModels[0]
		if model.ContextWindow > 0 {
			fmt.Printf("   Context usage: %.1f%% (%s)\n", model.Utilization, model.ModelName)
		}
	}

	// Verbose mode: show all models
	if o.config.Verbose && len(analysis.AllModels) > 1 {
		fmt.Println()
		for _, model := range analysis.AllModels {
			status := "✓"
			color := "\033[32m" // green
			if !model.IsCompatible {
				status = "✗"
				color = "\033[31m" // red
			}

			if model.ContextWindow > 0 {
				fmt.Printf("   %s%s\033[0m %-20s %.1f%% (%s/%s)\n",
					color, status, model.ModelName, model.Utilization,
					formatWithCommas(model.TokenCount), formatWithCommas(model.ContextWindow))
			} else {
				fmt.Printf("   %s%s\033[0m %-20s %s\n",
					color, status, model.ModelName, model.FailureReason)
			}
		}
	}

	// Contextual suggestions - actionable and specific
	if analysis.CompatibleModels == 0 {
		fmt.Println()
		fmt.Println("💡 Try reducing input size:")
		fmt.Println("   • thinktank instructions.txt ./src          # focus on specific directories")
		fmt.Println("   • --exclude \"docs/,*.md,build/\"             # exclude documentation/build files")
		fmt.Println("   • --dry-run                                # check token count first")
	} else if analysis.SkippedModels > 0 && !o.config.Verbose {
		fmt.Printf("   Use --verbose to see all %d models\n", analysis.TotalModels)
	}

	fmt.Println()
}

// formatWithCommas formats an integer with comma separators for better readability.
// Example: 1234567 -> "1,234,567"
func formatWithCommas(n int) string {
	// Convert to string and work with the digits
	str := strconv.Itoa(n)

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(str, "-") {
		negative = true
		str = str[1:]
	}

	// Add commas from right to left
	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	// Add negative sign back if needed
	if negative {
		return "-" + result.String()
	}
	return result.String()
}
