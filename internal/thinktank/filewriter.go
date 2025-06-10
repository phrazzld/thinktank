// Package thinktank contains the core application logic for the thinktank tool.
// This file implements the FileWriter interface for saving generated content to files,
// handling file creation, directory resolution, and related audit logging.
// The FileWriter component is responsible for the final output step in the workflow,
// writing the generated plan content to disk with proper error handling.
package thinktank

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// FileWriter defines the interface for file output writing
type FileWriter interface {
	// SaveToFile writes content to the specified file
	SaveToFile(ctx context.Context, content, outputFile string) error
}

// fileWriter implements the FileWriter interface
type fileWriter struct {
	logger          logutil.LoggerInterface
	auditLogger     auditlog.AuditLogger
	dirPermissions  os.FileMode
	filePermissions os.FileMode
}

// NewFileWriter creates a new FileWriter instance with the specified dependencies.
// It injects the required logger and audit logger to ensure proper output
// handling and audit trail generation during file operations.
func NewFileWriter(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, dirPermissions, filePermissions os.FileMode) FileWriter {
	return &fileWriter{
		logger:          logger,
		auditLogger:     auditLogger,
		dirPermissions:  dirPermissions,
		filePermissions: filePermissions,
	}
}

// SaveToFile writes the content to the specified file and handles audit logging.
// It ensures proper directory existence, resolves relative paths to absolute paths,
// and generates appropriate audit log entries for the operation's start and completion.
// The method handles errors gracefully and ensures they are properly logged.
func (fw *fileWriter) SaveToFile(ctx context.Context, content, outputFile string) error {
	// Log the start of output saving
	saveStartTime := time.Now()
	inputs := map[string]interface{}{
		"output_path":    outputFile,
		"content_length": len(content),
	}
	if logErr := fw.auditLogger.LogOp(ctx, "SaveOutput", "InProgress", inputs, nil, nil); logErr != nil {
		fw.logger.Error("Failed to write audit log: %v", logErr)
	}

	// Ensure output path is absolute
	outputPath := outputFile
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			fw.logger.Error("Error getting current working directory: %v", err)

			// Log failure to save output
			saveDurationMs := time.Since(saveStartTime).Milliseconds()
			inputs["duration_ms"] = saveDurationMs
			if logErr := fw.auditLogger.LogOp(ctx, "SaveOutput", "Failure", inputs, nil, err); logErr != nil {
				fw.logger.Error("Failed to write audit log: %v", logErr)
			}

			return fmt.Errorf("error getting current working directory: %w", err)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, fw.dirPermissions); err != nil {
		fw.logger.Error("Error creating output directory %s: %v", outputDir, err)

		// Log failure to save output
		saveDurationMs := time.Since(saveStartTime).Milliseconds()
		inputs["duration_ms"] = saveDurationMs
		if logErr := fw.auditLogger.LogOp(ctx, "SaveOutput", "Failure", inputs, nil, err); logErr != nil {
			fw.logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
	}

	// Write to file
	fw.logger.Info("Writing to file %s...", outputPath)
	err := os.WriteFile(outputPath, []byte(content), fw.filePermissions)

	// Calculate duration in milliseconds
	saveDurationMs := time.Since(saveStartTime).Milliseconds()

	if err != nil {
		fw.logger.Error("Error writing to file %s: %v", outputPath, err)

		// Log failure to save output
		inputs["duration_ms"] = saveDurationMs
		if logErr := fw.auditLogger.LogOp(ctx, "SaveOutput", "Failure", inputs, nil, err); logErr != nil {
			fw.logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("error writing to file %s: %w", outputPath, err)
	}

	// Log successful saving of output
	inputs["duration_ms"] = saveDurationMs
	outputs := map[string]interface{}{
		"content_length": len(content),
	}
	if logErr := fw.auditLogger.LogOp(ctx, "SaveOutput", "Success", inputs, outputs, nil); logErr != nil {
		fw.logger.Error("Failed to write audit log: %v", logErr)
	}

	fw.logger.Info("Successfully saved to %s", outputPath)
	return nil
}
