// Package orchestrator is responsible for coordinating the core application workflow.
package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[36m"
)

// ResultsSummary contains information about processing results
type ResultsSummary struct {
	TotalModels      int
	SuccessfulModels int
	FailedModels     []string
	SuccessfulNames  []string
	SynthesisPath    string
	OutputPaths      []string
}

// SummaryWriter handles generating and displaying summaries of processing results
type SummaryWriter interface {
	// GenerateSummary creates a summary string from the results data
	GenerateSummary(summary *ResultsSummary) string

	// DisplaySummary writes the summary to the appropriate output
	DisplaySummary(ctx context.Context, summary *ResultsSummary)
}

// DefaultSummaryWriter implements the SummaryWriter interface
type DefaultSummaryWriter struct {
	logger logutil.LoggerInterface
}

// NewSummaryWriter creates a new SummaryWriter
func NewSummaryWriter(logger logutil.LoggerInterface) SummaryWriter {
	return &DefaultSummaryWriter{
		logger: logger,
	}
}

// GenerateSummary creates a human-readable summary string with optional color coding
func (w *DefaultSummaryWriter) GenerateSummary(summary *ResultsSummary) string {
	var sb strings.Builder

	// Determine result status
	var statusText string
	var statusColor string
	failedCount := len(summary.FailedModels)

	if summary.SuccessfulModels == 0 {
		statusText = "FAILED"
		statusColor = colorRed
	} else if failedCount > 0 {
		statusText = "PARTIAL SUCCESS"
		statusColor = colorYellow
	} else {
		statusText = "SUCCESS"
		statusColor = colorGreen
	}

	// Build header with status
	sb.WriteString("\n")
	sb.WriteString("┌────────────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│ Thinktank Execution Summary                                    │\n")
	sb.WriteString("├────────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Status: %s%s%s", statusColor, statusText, colorReset))
	sb.WriteString(strings.Repeat(" ", 52-len(statusText)))
	sb.WriteString("│\n")

	// Add processing statistics
	sb.WriteString("├────────────────────────────────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Models: %d total, ", summary.TotalModels))
	sb.WriteString(fmt.Sprintf("%s%d successful%s, ", colorGreen, summary.SuccessfulModels, colorReset))
	sb.WriteString(fmt.Sprintf("%s%d failed%s", colorRed, failedCount, colorReset))
	sb.WriteString(strings.Repeat(" ", 28-digitCount(summary.SuccessfulModels)-digitCount(failedCount)))
	sb.WriteString("│\n")

	// Add synthesis file path if available
	if summary.SynthesisPath != "" {
		sb.WriteString("├────────────────────────────────────────────────────────────────┤\n")
		sb.WriteString(fmt.Sprintf("│ %sSynthesis file:%s %-43s │\n", colorBlue, colorReset, truncatePath(summary.SynthesisPath, 43)))
	}

	// Add successful models if any
	if summary.SuccessfulModels > 0 {
		sb.WriteString("├────────────────────────────────────────────────────────────────┤\n")
		sb.WriteString(fmt.Sprintf("│ %sSuccessful models:%s %-40s │\n", colorGreen, colorReset, truncateList(summary.SuccessfulNames, 40)))

		// List individual output paths if no synthesis and multiple successful models
		if summary.SynthesisPath == "" && len(summary.OutputPaths) > 0 {
			sb.WriteString("├────────────────────────────────────────────────────────────────┤\n")
			sb.WriteString(fmt.Sprintf("│ %sOutput files:%s                                                │\n", colorBlue, colorReset))
			for _, path := range summary.OutputPaths {
				sb.WriteString(fmt.Sprintf("│   %s │\n", truncatePath(path, 58)))
			}
		}
	}

	// Add failed models if any
	if failedCount > 0 {
		sb.WriteString("├────────────────────────────────────────────────────────────────┤\n")
		sb.WriteString(fmt.Sprintf("│ %sFailed models:%s %-43s │\n", colorRed, colorReset, truncateList(summary.FailedModels, 43)))
	}

	// Close the box
	sb.WriteString("└────────────────────────────────────────────────────────────────┘\n")

	return sb.String()
}

// DisplaySummary writes the summary to the appropriate output
func (w *DefaultSummaryWriter) DisplaySummary(ctx context.Context, summary *ResultsSummary) {
	// Generate the summary text
	summaryText := w.GenerateSummary(summary)

	// Log the summary info
	w.logger.InfoContext(ctx, "Execution summary: %d total models, %d successful, %d failed",
		summary.TotalModels, summary.SuccessfulModels, len(summary.FailedModels))

	// For synthesis path, log it if available
	if summary.SynthesisPath != "" {
		w.logger.InfoContext(ctx, "Synthesis output saved to: %s", summary.SynthesisPath)
	}

	// Print the formatted summary to stdout
	fmt.Print(summaryText)
}

// Helper functions

// truncatePath shortens a path if it's too long
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	// Try to keep the filename portion
	filename := path
	if idx := strings.LastIndexByte(path, '/'); idx != -1 {
		filename = path[idx+1:]
	}

	if len(filename) <= maxLen-4 {
		return "..." + path[len(path)-(maxLen-3):]
	}

	return "..." + path[len(path)-(maxLen-3):]
}

// truncateList formats a list of names, truncating if necessary
func truncateList(items []string, maxLen int) string {
	if len(items) == 0 {
		return "none"
	}

	// For single item, just truncate the name
	if len(items) == 1 {
		if len(items[0]) <= maxLen {
			return items[0]
		}
		return items[0][:maxLen-3] + "..."
	}

	// For multiple items, show count and some names
	text := fmt.Sprintf("%d models", len(items))
	remaining := maxLen - len(text) - 3 // space for " ()"

	if remaining <= 0 {
		return text
	}

	var names []string
	var usedChars int

	for _, item := range items {
		// Add 2 for separator ", "
		if usedChars+len(item)+2 > remaining {
			if len(names) == 0 {
				// If we can't fit even one name, truncate it
				truncated := item
				if len(item) > remaining-3 {
					truncated = item[:remaining-3] + "..."
				}
				names = append(names, truncated)
			} else {
				names = append(names, "...")
			}
			break
		}

		names = append(names, item)
		usedChars += len(item) + 2
	}

	return fmt.Sprintf("%s (%s)", text, strings.Join(names, ", "))
}

// digitCount returns the number of digits in an integer
func digitCount(n int) int {
	if n == 0 {
		return 1
	}

	count := 0
	for n > 0 {
		count++
		n /= 10
	}
	return count
}
