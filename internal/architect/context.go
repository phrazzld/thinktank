// Package architect contains the core application logic for the architect tool
package architect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// ContextStats holds information about processed files and context size
type ContextStats struct {
	ProcessedFilesCount int
	CharCount           int
	LineCount           int
	TokenCount          int32
	ProcessedFiles      []string
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
	GatherContext(ctx context.Context, client gemini.Client, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)

	// DisplayDryRunInfo shows detailed information for dry run mode
	DisplayDryRunInfo(ctx context.Context, client gemini.Client, stats *ContextStats) error
}

// contextGatherer implements the ContextGatherer interface
type contextGatherer struct {
	logger       logutil.LoggerInterface
	dryRun       bool
	tokenManager TokenManager
}

// NewContextGatherer creates a new ContextGatherer instance
func NewContextGatherer(logger logutil.LoggerInterface, dryRun bool, tokenManager TokenManager) ContextGatherer {
	return &contextGatherer{
		logger:       logger,
		dryRun:       dryRun,
		tokenManager: tokenManager,
	}
}

// GatherContext collects and processes files based on configuration
func (cg *contextGatherer) GatherContext(ctx context.Context, client gemini.Client, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
	// Log appropriate message based on mode
	if cg.dryRun {
		cg.logger.Info("Dry run mode: gathering files that would be included in context...")
		cg.logger.Debug("Processing files with include filters: %v", config.Include)
		cg.logger.Debug("Processing files with exclude filters: %v", config.Exclude)
		cg.logger.Debug("Processing files with exclude names: %v", config.ExcludeNames)
		cg.logger.Debug("Paths being processed: %v", config.Paths)
	} else {
		cg.logger.Info("Gathering project context from %d paths...", len(config.Paths))
		cg.logger.Debug("Include filters: %v", config.Include)
		cg.logger.Debug("Exclude filters: %v", config.Exclude)
		cg.logger.Debug("Exclude names: %v", config.ExcludeNames)
		cg.logger.Debug("Paths being processed: %v", config.Paths)
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
	if err != nil {
		cg.logger.Error("Failed during project context gathering: %v", err)
		return nil, nil, fmt.Errorf("failed during project context gathering: %w", err)
	}

	// Set the processed files count in stats
	stats.ProcessedFilesCount = processedFilesCount

	// Log warning if no files were processed
	if processedFilesCount == 0 {
		cg.logger.Warn("No files were processed for context. Check paths and filters.")
		return contextFiles, stats, nil
	}

	// Create a combined string for token counting
	var combinedContent strings.Builder
	for _, file := range contextFiles {
		combinedContent.WriteString(file.Content)
		combinedContent.WriteString("\n")
	}
	projectContext := combinedContent.String()

	// Calculate token statistics
	cg.logger.Info("Calculating token statistics for %d processed files...", stats.ProcessedFilesCount)
	startTime := time.Now()
	charCount, lineCount, tokenCount := fileutil.CalculateStatisticsWithTokenCounting(ctx, client, projectContext, cg.logger)
	duration := time.Since(startTime)
	cg.logger.Debug("Token calculation completed in %v", duration)

	// Store statistics in the stats struct
	stats.CharCount = charCount
	stats.LineCount = lineCount
	stats.TokenCount = int32(tokenCount)

	// Handle output based on mode
	if processedFilesCount > 0 {
		cg.logger.Info("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount)

		// Additional detailed debug information if needed
		if config.LogLevel == logutil.DebugLevel && !cg.dryRun {
			cg.logger.Debug("Context details: files=%d, lines=%d, chars=%d, tokens=%d",
				processedFilesCount, lineCount, charCount, tokenCount)
		}
	}

	return contextFiles, stats, nil
}

// DisplayDryRunInfo shows detailed information for dry run mode
func (cg *contextGatherer) DisplayDryRunInfo(ctx context.Context, client gemini.Client, stats *ContextStats) error {
	cg.logger.Info("Files that would be included in context:")
	if stats.ProcessedFilesCount == 0 {
		cg.logger.Info("  No files matched the current filters.")
	} else {
		for i, file := range stats.ProcessedFiles {
			cg.logger.Info("  %d. %s", i+1, file)
		}
	}

	cg.logger.Info("Context statistics:")
	cg.logger.Info("  Files: %d", stats.ProcessedFilesCount)
	cg.logger.Info("  Lines: %d", stats.LineCount)
	cg.logger.Info("  Characters: %d", stats.CharCount)
	cg.logger.Info("  Tokens: %d", stats.TokenCount)

	// Get model info for token limit comparison
	modelInfo, modelInfoErr := client.GetModelInfo(ctx)
	if modelInfoErr != nil {
		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(modelInfoErr); ok {
			cg.logger.Warn("Could not get model information: %s", apiErr.Message)
			// Only show detailed info in debug logs
			cg.logger.Debug("Model info error details: %s", apiErr.DebugInfo())
		} else {
			cg.logger.Warn("Could not get model information: %v", modelInfoErr)
		}
		// Continue - this is not a fatal error for dry run mode
	} else {
		// Convert to int32 for comparison with model limits
		tokenCountInt32 := stats.TokenCount
		percentOfLimit := float64(tokenCountInt32) / float64(modelInfo.InputTokenLimit) * 100
		cg.logger.Info("Token usage: %d / %d (%.1f%% of model's limit)",
			tokenCountInt32, modelInfo.InputTokenLimit, percentOfLimit)

		// Check if token count exceeds limit
		if tokenCountInt32 > modelInfo.InputTokenLimit {
			cg.logger.Error("WARNING: Token count exceeds model's limit by %d tokens",
				tokenCountInt32-modelInfo.InputTokenLimit)
			cg.logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		} else {
			cg.logger.Info("Context size is within the model's token limit")
		}
	}

	cg.logger.Info("Dry run completed successfully.")
	cg.logger.Info("To generate content, run without the --dry-run flag.")

	return nil
}
