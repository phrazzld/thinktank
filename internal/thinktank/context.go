// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// ContextStats holds information about processed files and context size
type ContextStats struct {
	ProcessedFilesCount int
	CharCount           int
	LineCount           int
	// TokenCount field removed as part of T032F - token handling refactoring
	ProcessedFiles []string
}

// GatherConfig holds parameters needed for gathering context
type GatherConfig struct {
	Paths        []string
	Include      string
	Exclude      string
	ExcludeNames string
	Format       string
	Verbose      bool
	LogLevel     logutil.LogLevel
}

// ContextGatherer defines the interface for gathering project context
type ContextGatherer interface {
	// GatherContext collects and processes files based on configuration
	GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)

	// DisplayDryRunInfo shows detailed information for dry run mode
	DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error
}

// contextGatherer implements the ContextGatherer interface
type contextGatherer struct {
	logger        logutil.LoggerInterface
	consoleWriter logutil.ConsoleWriter
	dryRun        bool
	client        llm.LLMClient
	auditLogger   auditlog.AuditLogger
}

// NewContextGatherer creates a new ContextGatherer instance
func NewContextGatherer(logger logutil.LoggerInterface, consoleWriter logutil.ConsoleWriter, dryRun bool, client llm.LLMClient, auditLogger auditlog.AuditLogger) ContextGatherer {
	return &contextGatherer{
		logger:        logger,
		consoleWriter: consoleWriter,
		dryRun:        dryRun,
		client:        client,
		auditLogger:   auditLogger,
	}
}

// Token counting functions removed as part of T032F - token handling refactoring

// GatherContext collects and processes files based on configuration
func (cg *contextGatherer) GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
	// Log start of context gathering operation to audit log
	gatherStartTime := time.Now()
	inputs := map[string]interface{}{
		"paths":         config.Paths,
		"include":       config.Include,
		"exclude":       config.Exclude,
		"exclude_names": config.ExcludeNames,
		"format":        config.Format,
	}
	if logErr := cg.auditLogger.LogOp(ctx, "GatherContext", "InProgress", inputs, nil, nil); logErr != nil {
		cg.logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
	}

	// Log appropriate message based on mode
	if cg.dryRun {
		cg.logger.InfoContext(ctx, "Dry run mode: gathering files that would be included in context...")
		cg.logger.DebugContext(ctx, "Processing files with include filters: %v", config.Include)
		cg.logger.DebugContext(ctx, "Processing files with exclude filters: %v", config.Exclude)
		cg.logger.DebugContext(ctx, "Processing files with exclude names: %v", config.ExcludeNames)
		cg.logger.DebugContext(ctx, "Paths being processed: %v", config.Paths)
	} else {
		cg.logger.InfoContext(ctx, "Gathering project context from %d paths...", len(config.Paths))
		cg.logger.DebugContext(ctx, "Include filters: %v", config.Include)
		cg.logger.DebugContext(ctx, "Exclude filters: %v", config.Exclude)
		cg.logger.DebugContext(ctx, "Exclude names: %v", config.ExcludeNames)
		cg.logger.DebugContext(ctx, "Paths being processed: %v", config.Paths)
	}

	// Setup file processing configuration
	fileConfig := fileutil.NewConfig(config.Verbose, config.Include, config.Exclude, config.ExcludeNames, config.Format, cg.logger)

	// Initialize ContextStats
	stats := &ContextStats{
		ProcessedFiles: make([]string, 0),
	}

	// Track processed files for dry run mode
	if cg.dryRun {
		collector := func(path string) {
			stats.ProcessedFiles = append(stats.ProcessedFiles, path)
		}
		fileConfig.SetFileCollector(collector)
	}

	// Gather project context
	contextFiles, processedFilesCount, err := fileutil.GatherProjectContext(config.Paths, fileConfig)

	// Calculate duration in milliseconds
	gatherDurationMs := time.Since(gatherStartTime).Milliseconds()

	if err != nil {
		cg.logger.ErrorContext(ctx, "Failed during project context gathering: %v", err)

		// Log the failure of context gathering to audit log
		inputs := map[string]interface{}{
			"paths":         config.Paths,
			"include":       config.Include,
			"exclude":       config.Exclude,
			"exclude_names": config.ExcludeNames,
			"duration_ms":   gatherDurationMs,
		}
		if logErr := cg.auditLogger.LogOp(ctx, "GatherContext", "Failure", inputs, nil, err); logErr != nil {
			cg.logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
		}

		return nil, nil, fmt.Errorf("failed during project context gathering: %w", err)
	}

	// Set the processed files count in stats
	stats.ProcessedFilesCount = processedFilesCount

	// Log warning if no files were processed
	if processedFilesCount == 0 {
		cg.logger.WarnContext(ctx, "No files were processed for context. Check paths and filters.")
		return contextFiles, stats, nil
	}

	// Create a combined string for calculating basic statistics
	var combinedContent strings.Builder
	for _, file := range contextFiles {
		combinedContent.WriteString(file.Content)
		combinedContent.WriteString("\n")
	}
	projectContext := combinedContent.String()

	// Calculate basic statistics
	cg.logger.InfoContext(ctx, "Calculating statistics for %d processed files...", stats.ProcessedFilesCount)
	startTime := time.Now()

	// Calculate character and line counts directly
	charCount := len(projectContext)
	lineCount := strings.Count(projectContext, "\n") + 1

	// Token counting code removed as part of T032F - token handling refactoring

	duration := time.Since(startTime)
	cg.logger.DebugContext(ctx, "Statistics calculation completed in %v", duration)

	// Store statistics in the stats struct
	stats.CharCount = charCount
	stats.LineCount = lineCount
	// TokenCount field removed as part of T032F - token handling refactoring

	// Handle output based on mode
	if processedFilesCount > 0 {
		cg.logger.InfoContext(ctx, "Context gathered: %d files, %d lines, %d chars",
			processedFilesCount, lineCount, charCount)

		// Additional detailed debug information if needed
		if config.LogLevel == logutil.DebugLevel && !cg.dryRun {
			cg.logger.DebugContext(ctx, "Context details: files=%d, lines=%d, chars=%d",
				processedFilesCount, lineCount, charCount)
		}
	}

	// Log the successful completion of context gathering to audit log
	outputs := map[string]interface{}{
		"processed_files_count": stats.ProcessedFilesCount,
		"char_count":            stats.CharCount,
		"line_count":            stats.LineCount,
		// token_count field removed as part of T032F - token handling refactoring
		"files_count": len(contextFiles),
	}
	if logErr := cg.auditLogger.LogOp(ctx, "GatherContext", "Success", inputs, outputs, nil); logErr != nil {
		cg.logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
	}

	return contextFiles, stats, nil
}

// DisplayDryRunInfo shows detailed information for dry run mode
func (cg *contextGatherer) DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error {
	// Log detailed information to structured logs for debugging
	cg.logger.InfoContext(ctx, "Dry run mode: displaying context information")
	cg.logger.DebugContext(ctx, "Found %d files matching filters", stats.ProcessedFilesCount)

	// Use ConsoleWriter for user-facing output
	cg.consoleWriter.StatusMessage("Dry run results:")

	if stats.ProcessedFilesCount == 0 {
		cg.consoleWriter.StatusMessage("  No files matched the current filters.")
	} else {
		cg.consoleWriter.StatusMessage(fmt.Sprintf("  Files that would be included: %d", stats.ProcessedFilesCount))
		// Only show first 10 files to avoid overwhelming output
		displayCount := len(stats.ProcessedFiles)
		if displayCount > 10 {
			displayCount = 10
		}
		for i := 0; i < displayCount; i++ {
			cg.consoleWriter.StatusMessage(fmt.Sprintf("    %d. %s", i+1, stats.ProcessedFiles[i]))
		}
		if len(stats.ProcessedFiles) > 10 {
			cg.consoleWriter.StatusMessage(fmt.Sprintf("    ... and %d more files", len(stats.ProcessedFiles)-10))
		}
	}

	// Display context statistics
	cg.consoleWriter.StatusMessage("")
	cg.consoleWriter.StatusMessage("Context statistics:")
	cg.consoleWriter.StatusMessage(fmt.Sprintf("  Files: %d", stats.ProcessedFilesCount))
	cg.consoleWriter.StatusMessage(fmt.Sprintf("  Lines: %d", stats.LineCount))
	cg.consoleWriter.StatusMessage(fmt.Sprintf("  Characters: %d", stats.CharCount))

	// Token counting and limit comparison code removed as part of T032F - token handling refactoring

	cg.consoleWriter.StatusMessage("")
	cg.consoleWriter.SuccessMessage("Dry run completed successfully.")
	cg.consoleWriter.StatusMessage("To generate content, run without the --dry-run flag.")

	// Also log completion to structured logs
	cg.logger.InfoContext(ctx, "Dry run completed successfully",
		"files_count", stats.ProcessedFilesCount,
		"lines_count", stats.LineCount,
		"char_count", stats.CharCount)

	return nil
}
