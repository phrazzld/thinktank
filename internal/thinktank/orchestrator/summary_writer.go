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
	logger        logutil.LoggerInterface
	consoleWriter logutil.ConsoleWriter
}

// NewSummaryWriter creates a new SummaryWriter
func NewSummaryWriter(logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter) SummaryWriter {
	return &DefaultSummaryWriter{
		logger:        logger,
		consoleWriter: consoleWriter,
	}
}

// GenerateSummary creates a human-readable summary string with emoji prefixes and color coding
func (w *DefaultSummaryWriter) GenerateSummary(summary *ResultsSummary) string {
	var sb strings.Builder

	// Determine result status
	var statusText string
	var statusColor string
	var statusEmoji string
	failedCount := len(summary.FailedModels)

	if summary.SuccessfulModels == 0 {
		statusText = "FAILED"
		statusColor = colorRed
		statusEmoji = "🔴"
	} else if failedCount > 0 {
		statusText = "PARTIAL SUCCESS"
		statusColor = colorYellow
		statusEmoji = "🟡"
	} else {
		statusText = "SUCCESS"
		statusColor = colorGreen
		statusEmoji = "🟢"
	}

	// Build header with status
	sb.WriteString("\n")
	sb.WriteString("✨ THINKTANK EXECUTION SUMMARY ✨\n\n")

	// Add status
	sb.WriteString(fmt.Sprintf("%s Status: %s%s%s\n", statusEmoji, statusColor, statusText, colorReset))

	// Add processing statistics
	sb.WriteString(fmt.Sprintf("🔢 Models: %d total, %s%d successful%s, %s%d failed%s\n",
		summary.TotalModels,
		colorGreen, summary.SuccessfulModels, colorReset,
		colorRed, failedCount, colorReset))

	// Add synthesis file path if available
	if summary.SynthesisPath != "" {
		sb.WriteString(fmt.Sprintf("📄 Synthesis file: %s%s%s\n",
			colorBlue, truncatePath(summary.SynthesisPath, 60), colorReset))
	}

	// Add successful models if any
	if summary.SuccessfulModels > 0 {
		sb.WriteString(fmt.Sprintf("🚀 Successful models: %s%s%s\n",
			colorGreen, truncateList(summary.SuccessfulNames, 60), colorReset))

		// List individual output paths if no synthesis and multiple successful models
		if summary.SynthesisPath == "" && len(summary.OutputPaths) > 0 {
			sb.WriteString("📂 Output files:\n")
			for _, path := range summary.OutputPaths {
				sb.WriteString(fmt.Sprintf("  - %s%s%s\n", colorBlue, truncatePath(path, 70), colorReset))
			}
		}
	}

	// Add failed models if any
	if failedCount > 0 {
		sb.WriteString(fmt.Sprintf("❌ Failed models: %s%s%s\n",
			colorRed, truncateList(summary.FailedModels, 60), colorReset))
	}

	sb.WriteString("\n")

	return sb.String()
}

// DisplaySummary writes the summary to the appropriate output using modern clean format
func (w *DefaultSummaryWriter) DisplaySummary(ctx context.Context, summary *ResultsSummary) {
	// Log the summary info for structured logging
	w.logger.InfoContext(ctx, "Execution summary: %d total models, %d successful, %d failed",
		summary.TotalModels, summary.SuccessfulModels, len(summary.FailedModels))

	// For synthesis path, log it if available
	if summary.SynthesisPath != "" {
		w.logger.InfoContext(ctx, "Synthesis output saved to: %s", summary.SynthesisPath)
	}

	// Convert to SummaryData format and display using modern clean output
	summaryData := w.convertToSummaryData(summary)
	w.consoleWriter.ShowSummarySection(summaryData)
}

// convertToSummaryData converts ResultsSummary to SummaryData format
func (w *DefaultSummaryWriter) convertToSummaryData(summary *ResultsSummary) logutil.SummaryData {
	// Determine synthesis status
	synthesisStatus := "skipped"
	if summary.SynthesisPath != "" {
		synthesisStatus = "completed"
	}

	// Determine output directory from synthesis path or output paths
	outputDirectory := ""
	if summary.SynthesisPath != "" {
		// Extract directory from synthesis path
		if idx := strings.LastIndexByte(summary.SynthesisPath, '/'); idx != -1 {
			outputDirectory = summary.SynthesisPath[:idx+1]
		}
	} else if len(summary.OutputPaths) > 0 {
		// Extract directory from first output path
		if idx := strings.LastIndexByte(summary.OutputPaths[0], '/'); idx != -1 {
			outputDirectory = summary.OutputPaths[0][:idx+1]
		}
	}

	return logutil.SummaryData{
		ModelsProcessed:  summary.TotalModels,
		SuccessfulModels: summary.SuccessfulModels,
		FailedModels:     len(summary.FailedModels),
		SynthesisStatus:  synthesisStatus,
		OutputDirectory:  outputDirectory,
	}
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
