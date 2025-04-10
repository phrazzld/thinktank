// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// OutputWriter defines the interface for file output and plan writing
type OutputWriter interface {
	// SaveToFile writes the generated plan to the specified file
	SaveToFile(content, outputFile string) error

	// GenerateAndSavePlan creates and saves the plan to a file
	GenerateAndSavePlan(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, promptManager prompt.ManagerInterface) error

	// GenerateAndSavePlanWithConfig creates and saves the plan to a file using the config system
	GenerateAndSavePlanWithConfig(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, configManager config.ManagerInterface) error
}

// outputWriter implements the OutputWriter interface
type outputWriter struct {
	logger       logutil.LoggerInterface
	tokenManager TokenManager
}

// NewOutputWriter creates a new OutputWriter instance
func NewOutputWriter(logger logutil.LoggerInterface, tokenManager TokenManager) OutputWriter {
	return &outputWriter{
		logger:       logger,
		tokenManager: tokenManager,
	}
}

// SaveToFile writes the generated plan to the specified file
func (ow *outputWriter) SaveToFile(content, outputFile string) error {
	// Ensure output path is absolute
	outputPath := outputFile
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			ow.logger.Error("Error getting current working directory: %v", err)
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	// Write to file
	ow.logger.Info("Writing plan to %s...", outputPath)
	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		ow.logger.Error("Error writing plan to file %s: %v", outputPath, err)
		return fmt.Errorf("error writing plan to file %s: %w", outputPath, err)
	}

	ow.logger.Info("Successfully generated plan and saved to %s", outputPath)
	return nil
}

// GenerateAndSavePlan creates and saves the plan to a file
func (ow *outputWriter) GenerateAndSavePlan(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, promptManager prompt.ManagerInterface) error {
	// Stub implementation - will be replaced with actual code from main.go
	return fmt.Errorf("not implemented yet")
}

// GenerateAndSavePlanWithConfig creates and saves the plan to a file using the config system
func (ow *outputWriter) GenerateAndSavePlanWithConfig(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, configManager config.ManagerInterface) error {
	// Stub implementation - will be replaced with actual code from main.go
	return fmt.Errorf("not implemented yet")
}
