// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
)

// OutputWriter handles writing model outputs to files
type OutputWriter interface {
	// SaveIndividualOutputs saves individual model outputs to separate files
	// It returns the number of files successfully saved, map of model name to file paths, and any errors encountered
	SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, map[string]string, error)

	// SaveSynthesisOutput saves the synthesis result to a file
	// It returns the path to the saved synthesis file and any error encountered
	SaveSynthesisOutput(ctx context.Context, content string, modelName string, outputDir string) (string, error)
}

// DefaultOutputWriter implements the OutputWriter interface
type DefaultOutputWriter struct {
	fileWriter  interfaces.FileWriter
	auditLogger auditlog.AuditLogger
	logger      logutil.LoggerInterface
}

// LegacyOutputWriter is used for backward compatibility with tests
// This interface matches the old interface signature before T029
type LegacyOutputWriter interface {
	SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, error)
	SaveSynthesisOutput(ctx context.Context, content string, modelName string, outputDir string) error
}

// LegacyOutputWriterAdapter adapts the new OutputWriter to the old interface
type LegacyOutputWriterAdapter struct {
	outputWriter OutputWriter
}

// SaveIndividualOutputs adapts the new method to the old signature
func (a *LegacyOutputWriterAdapter) SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, error) {
	count, _, err := a.outputWriter.SaveIndividualOutputs(ctx, modelOutputs, outputDir)
	return count, err
}

// SaveSynthesisOutput adapts the new method to the old signature
func (a *LegacyOutputWriterAdapter) SaveSynthesisOutput(ctx context.Context, content string, modelName string, outputDir string) error {
	_, err := a.outputWriter.SaveSynthesisOutput(ctx, content, modelName, outputDir)
	return err
}

// NewOutputWriter creates a new OutputWriter instance with the specified dependencies
func NewOutputWriter(
	fileWriter interfaces.FileWriter,
	auditLogger auditlog.AuditLogger,
	logger logutil.LoggerInterface,
) OutputWriter {
	return &DefaultOutputWriter{
		fileWriter:  fileWriter,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

// SaveIndividualOutputs saves individual model outputs to separate files
// It iterates through the modelOutputs map, sanitizes the model names for use in filenames,
// and saves each output to a separate file. It tracks both successful saves and failures,
// returning a count of successful saves, a map of model names to file paths, and any errors that occurred.
func (w *DefaultOutputWriter) SaveIndividualOutputs(
	ctx context.Context,
	modelOutputs map[string]string,
	outputDir string,
) (int, map[string]string, error) {
	// Get logger with context
	contextLogger := w.logger.WithContext(ctx)

	// Track stats for logging and error reporting
	totalCount := len(modelOutputs)
	savedCount := 0
	errorCount := 0
	outputPaths := make(map[string]string)

	// Log start of output saving
	contextLogger.InfoContext(ctx, "Saving individual model outputs")
	contextLogger.DebugContext(ctx, "Preparing to save %d model outputs", totalCount)

	// Iterate over the model outputs and save each to a file
	for modelName, content := range modelOutputs {
		// Sanitize model name for use in filename
		sanitizedModelName := modelproc.SanitizeFilename(modelName)

		// Construct output file path
		outputFilePath := filepath.Join(outputDir, sanitizedModelName+".md")

		// Save the output to file
		contextLogger.DebugContext(ctx, "Saving output for model %s to %s", modelName, outputFilePath)
		if err := w.fileWriter.SaveToFile(ctx, content, outputFilePath); err != nil {
			contextLogger.ErrorContext(ctx, "Failed to save output for model %s: %v", modelName, err)
			errorCount++
		} else {
			savedCount++
			outputPaths[modelName] = outputFilePath
			contextLogger.InfoContext(ctx, "Successfully saved output for model %s", modelName)
		}
	}

	// Log summary of file operations
	if errorCount > 0 {
		contextLogger.ErrorContext(ctx, "Completed with errors: %d files saved successfully, %d files failed",
			savedCount, errorCount)

		// Create a descriptive error for the file save failures using proper categorization
		return savedCount, outputPaths, WrapOrchestratorError(
			ErrOutputFileSaveFailed,
			fmt.Sprintf("%d/%d files failed to save", errorCount, totalCount),
		)
	}

	contextLogger.InfoContext(ctx, "All %d model outputs saved successfully", savedCount)
	return savedCount, outputPaths, nil
}

// SaveSynthesisOutput saves the synthesis result to a file
// It sanitizes the model name for use in the filename, constructs the output path with
// a -synthesis suffix, and saves the content to that file. It returns the path to the saved
// file and an error if the save fails.
func (w *DefaultOutputWriter) SaveSynthesisOutput(
	ctx context.Context,
	content string,
	modelName string,
	outputDir string,
) (string, error) {
	// Get logger with context
	contextLogger := w.logger.WithContext(ctx)

	// Sanitize model name for use in filename
	sanitizedModelName := modelproc.SanitizeFilename(modelName)

	// Construct output file path with -synthesis suffix
	outputFilePath := filepath.Join(outputDir, sanitizedModelName+"-synthesis.md")

	// Save the synthesis output to file
	contextLogger.DebugContext(ctx, "Saving synthesis output to %s", outputFilePath)
	if err := w.fileWriter.SaveToFile(ctx, content, outputFilePath); err != nil {
		contextLogger.ErrorContext(ctx, "Failed to save synthesis output: %v", err)
		return "", WrapOrchestratorError(
			ErrOutputFileSaveFailed,
			fmt.Sprintf("failed to save synthesis output to %s: %v", outputFilePath, err),
		)
	}

	contextLogger.InfoContext(ctx, "Successfully saved synthesis output to %s", outputFilePath)
	return outputFilePath, nil
}
