// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/config"
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

	// Initialize XDG-compliant configuration system
	configManager := config.NewManager(logger)

	// Load configuration from files
	err = configManager.LoadFromFiles()
	if err != nil {
		logger.Warn("Failed to load configuration: %v", err)
		logger.Info("Using default configuration")
	}

	// Ensure configuration directories exist
	if err := configManager.EnsureConfigDirs(); err != nil {
		logger.Warn("Failed to create configuration directories: %v", err)
	}

	// Convert cmdConfig to architect.CliConfig to pass to core logic
	coreConfig := convertToArchitectConfig(cmdConfig)

	// Handle special subcommands before regular flow
	if architect.HandleSpecialCommands(coreConfig, logger, configManager) {
		// Special command was executed, exit the program
		return
	}

	// Convert CLI flags to the format needed for merging
	cliFlags := ConvertConfigToMap(cmdConfig)

	// Merge CLI flags with loaded configuration
	if err := configManager.MergeWithFlags(cliFlags); err != nil {
		logger.Warn("Failed to merge CLI flags with configuration: %v", err)
	}

	// Get the final configuration
	_ = configManager.GetConfig()

	// Execute the core application logic
	err = architect.Execute(ctx, coreConfig, logger, configManager)
	if err != nil {
		logger.Error("Application failed: %v", err)
		os.Exit(1)
	}
}

// convertToArchitectConfig converts the cmd package CliConfig to the internal architect package CliConfig
func convertToArchitectConfig(cmdConfig *CliConfig) *architect.CliConfig {
	return &architect.CliConfig{
		TaskDescription: cmdConfig.TaskDescription,
		TaskFile:        cmdConfig.TaskFile,
		OutputFile:      cmdConfig.OutputFile,
		Format:          cmdConfig.Format,
		Paths:           cmdConfig.Paths,
		Include:         cmdConfig.Include,
		Exclude:         cmdConfig.Exclude,
		ExcludeNames:    cmdConfig.ExcludeNames,
		DryRun:          cmdConfig.DryRun,
		Verbose:         cmdConfig.Verbose,
		ApiKey:          cmdConfig.ApiKey,
		ModelName:       cmdConfig.ModelName,
		ConfirmTokens:   cmdConfig.ConfirmTokens,
		Template:        cmdConfig.PromptTemplate,
		ListExamples:    cmdConfig.ListExamples,
		ShowExample:     cmdConfig.ShowExample,
		LogLevel:        cmdConfig.LogLevel,
	}
}
