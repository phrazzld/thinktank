// Package architect contains the core application logic for the architect tool
package architect

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/llm"
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
	GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)

	// DisplayDryRunInfo shows detailed information for dry run mode
	DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error
}

// contextGatherer implements the ContextGatherer interface
type contextGatherer struct {
	logger       logutil.LoggerInterface
	dryRun       bool
	tokenManager TokenManager
	client       llm.LLMClient
	auditLogger  auditlog.AuditLogger
}

// NewContextGatherer creates a new ContextGatherer instance
func NewContextGatherer(logger logutil.LoggerInterface, dryRun bool, tokenManager TokenManager, client llm.LLMClient, auditLogger auditlog.AuditLogger) ContextGatherer {
	return &contextGatherer{
		logger:       logger,
		dryRun:       dryRun,
		tokenManager: tokenManager,
		client:       client,
		auditLogger:  auditLogger,
	}
}

// estimateTokenCount counts tokens simply by whitespace boundaries.
// This is kept as a fallback method in case the API token counting fails.
func estimateTokenCount(text string) int {
	count := 0
	inToken := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if inToken {
				count++
				inToken = false
			}
		} else {
			inToken = true
		}
	}
	if inToken {
		count++
	}
	return count
}

// GatherContext collects and processes files based on configuration
func (cg *contextGatherer) GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
	// Log start of context gathering operation to audit log
	gatherStartTime := time.Now()
	if logErr := cg.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: gatherStartTime,
		Operation: "GatherContextStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"paths":         config.Paths,
			"include":       config.Include,
			"exclude":       config.Exclude,
			"exclude_names": config.ExcludeNames,
			"format":        config.Format,
		},
		Message: "Starting to gather project context files",
	}); logErr != nil {
		cg.logger.Error("Failed to write audit log: %v", logErr)
	}

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

	// Calculate duration in milliseconds
	gatherDurationMs := time.Since(gatherStartTime).Milliseconds()

	if err != nil {
		cg.logger.Error("Failed during project context gathering: %v", err)

		// Log the failure of context gathering to audit log
		if logErr := cg.auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GatherContextEnd",
			Status:     "Failure",
			DurationMs: &gatherDurationMs,
			Inputs: map[string]interface{}{
				"paths":         config.Paths,
				"include":       config.Include,
				"exclude":       config.Exclude,
				"exclude_names": config.ExcludeNames,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to gather project context: %v", err),
				Type:    "ContextGatheringError",
			},
		}); logErr != nil {
			cg.logger.Error("Failed to write audit log: %v", logErr)
		}

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

	// Calculate basic statistics
	cg.logger.Info("Calculating token statistics for %d processed files...", stats.ProcessedFilesCount)
	startTime := time.Now()

	// Calculate character and line counts directly
	charCount := len(projectContext)
	lineCount := strings.Count(projectContext, "\n") + 1

	// Use token manager for token counting via the gemini client
	tokenCount := 0
	if cg.client != nil {
		// Use the gemini client to count tokens
		tokenResult, err := cg.client.CountTokens(ctx, projectContext)
		if err != nil {
			cg.logger.Warn("Failed to count tokens accurately: %v. Using estimation instead.", err)
			// Fall back to basic statistics using internal estimation
			tokenCount = estimateTokenCount(projectContext)
		} else {
			tokenCount = int(tokenResult.Total)
			cg.logger.Debug("Accurate token count: %d tokens", tokenCount)
		}
	} else {
		// Fall back to basic statistics if no client
		tokenCount = estimateTokenCount(projectContext)
		cg.logger.Debug("Using estimated token count: %d tokens", tokenCount)
	}

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

	// Log the successful completion of context gathering to audit log
	if logErr := cg.auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GatherContextEnd",
		Status:     "Success",
		DurationMs: &gatherDurationMs,
		Inputs: map[string]interface{}{
			"paths":         config.Paths,
			"include":       config.Include,
			"exclude":       config.Exclude,
			"exclude_names": config.ExcludeNames,
		},
		Outputs: map[string]interface{}{
			"processed_files_count": stats.ProcessedFilesCount,
			"char_count":            stats.CharCount,
			"line_count":            stats.LineCount,
			"token_count":           stats.TokenCount,
			"files_count":           len(contextFiles),
		},
		Message: "Successfully gathered project context files",
	}); logErr != nil {
		cg.logger.Error("Failed to write audit log: %v", logErr)
	}

	return contextFiles, stats, nil
}

// DisplayDryRunInfo shows detailed information for dry run mode
func (cg *contextGatherer) DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error {
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

	// Try to get token limits from registry first
	modelName := cg.client.GetModelName()

	// Get input token limit from the registry if available
	var inputTokenLimit int32
	var limitSource string

	// Check if we have an APIService that supports registry lookup
	if apiService, ok := cg.client.(interfaces.APIService); ok && apiService != nil {
		contextWindow, _, err := apiService.GetModelTokenLimits(modelName)
		if err == nil && contextWindow > 0 {
			// Use the context window from the registry
			inputTokenLimit = contextWindow
			limitSource = "registry"
			cg.logger.Debug("Using token limits from registry for model %s", modelName)
		}
	}

	// Fall back to client.GetModelInfo if registry lookup failed
	if inputTokenLimit == 0 {
		// Get model info for token limit comparison
		modelInfo, modelInfoErr := cg.client.GetModelInfo(ctx)
		if modelInfoErr != nil {
			// Check if it's a categorized error with enhanced details
			if catErr, ok := llm.IsCategorizedError(modelInfoErr); ok {
				category := catErr.Category()
				cg.logger.Warn("Could not get model information: %v (category: %s)",
					modelInfoErr, category.String())

				// Only show detailed error category in debug logs
				cg.logger.Debug("Model info error category: %s", category.String())
			} else {
				cg.logger.Warn("Could not get model information: %v", modelInfoErr)
			}
			// Continue - this is not a fatal error for dry run mode
		} else {
			// Use the input token limit from the client
			inputTokenLimit = modelInfo.InputTokenLimit
			limitSource = "client"
		}
	}

	// If we have a valid token limit, show usage information
	if inputTokenLimit > 0 {
		// Convert to int32 for comparison with model limits
		tokenCountInt32 := stats.TokenCount
		percentOfLimit := float64(tokenCountInt32) / float64(inputTokenLimit) * 100
		cg.logger.Info("Token usage: %d / %d (%.1f%% of model's limit) [source: %s]",
			tokenCountInt32, inputTokenLimit, percentOfLimit, limitSource)

		// Check if token count exceeds limit
		if tokenCountInt32 > inputTokenLimit {
			cg.logger.Error("WARNING: Token count exceeds model's limit by %d tokens",
				tokenCountInt32-inputTokenLimit)
			cg.logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		} else {
			cg.logger.Info("Context size is within the model's token limit")
		}
	}

	cg.logger.Info("Dry run completed successfully.")
	cg.logger.Info("To generate content, run without the --dry-run flag.")

	return nil
}
