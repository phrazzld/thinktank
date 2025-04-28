// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// FileWriter defines the interface for file writing
type FileWriter interface {
	// SaveToFile writes content to the specified file
	SaveToFile(content, outputFile string) error
}

// fileWriter implements the FileWriter interface
type fileWriter struct {
	logger          logutil.LoggerInterface
	dirPermissions  os.FileMode
	filePermissions os.FileMode
}

// NewFileWriter creates a new FileWriter instance
func NewFileWriter(logger logutil.LoggerInterface, dirPermissions, filePermissions os.FileMode) FileWriter {
	return &fileWriter{
		logger:          logger,
		dirPermissions:  dirPermissions,
		filePermissions: filePermissions,
	}
}

// SaveToFile writes content to the specified file
func (fw *fileWriter) SaveToFile(content, outputFile string) error {
	// Ensure output path is absolute
	outputPath := outputFile
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			fw.logger.Error("Error getting current working directory: %v", err)
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, fw.dirPermissions); err != nil {
		fw.logger.Error("Error creating output directory %s: %v", outputDir, err)
		return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
	}

	// Write to file
	fw.logger.Info("Writing to file %s...", outputPath)
	err := os.WriteFile(outputPath, []byte(content), fw.filePermissions)
	if err != nil {
		fw.logger.Error("Error writing to file %s: %v", outputPath, err)
		return fmt.Errorf("error writing to file %s: %w", outputPath, err)
	}

	fw.logger.Info("Successfully saved to %s", outputPath)
	return nil
}
