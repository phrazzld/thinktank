// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/architect"
)

// Main is the entry point for the architect CLI
func Main() {
	// Create a base context
	ctx := context.Background()

	// Parse command line flags
	cmdConfig, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging early for error reporting
	logger := SetupLogging(cmdConfig)
	logger.Info("Starting Architect - AI-assisted planning tool")

	// Configuration is now managed via CLI flags and environment variables only

	// Convert cmdConfig to architect.CliConfig to pass to core logic
	coreConfig := convertToArchitectConfig(cmdConfig)

	// CLI flags and environment variables are now the only source of configuration

	// Execute the core application logic
	err = architect.Execute(ctx, coreConfig, logger)
	if err != nil {
		logger.Error("Application failed: %v", err)
		os.Exit(1)
	}
}

// convertToArchitectConfig converts the cmd package CliConfig to the internal architect package CliConfig
func convertToArchitectConfig(cmdConfig *CliConfig) *architect.CliConfig {
	return &architect.CliConfig{
		InstructionsFile: cmdConfig.InstructionsFile,
		OutputFile:       cmdConfig.OutputFile,
		AuditLogFile:     cmdConfig.AuditLogFile,
		Format:           cmdConfig.Format,
		Paths:            cmdConfig.Paths,
		Include:          cmdConfig.Include,
		Exclude:          cmdConfig.Exclude,
		ExcludeNames:     cmdConfig.ExcludeNames,
		DryRun:           cmdConfig.DryRun,
		Verbose:          cmdConfig.Verbose,
		ApiKey:           cmdConfig.ApiKey,
		ModelName:        cmdConfig.ModelName,
		ConfirmTokens:    cmdConfig.ConfirmTokens,
		LogLevel:         cmdConfig.LogLevel,
	}
}
