// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
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

	// Initialize the audit logger
	// Note: The auditLogger will be passed to Execute() in a future task
	var auditLogger auditlog.AuditLogger
	if cmdConfig.AuditLogFile != "" {
		fileLogger, err := auditlog.NewFileAuditLogger(cmdConfig.AuditLogFile, logger)
		if err != nil {
			// Log error and fall back to NoOp implementation
			logger.Error("Failed to initialize file audit logger: %v. Audit logging disabled.", err)
			auditLogger = auditlog.NewNoOpAuditLogger()
		} else {
			auditLogger = fileLogger
			logger.Info("Audit logging enabled to file: %s", cmdConfig.AuditLogFile)
		}
	} else {
		auditLogger = auditlog.NewNoOpAuditLogger()
		logger.Debug("Audit logging is disabled")
	}

	// Ensure the audit logger is properly closed when the application exits
	defer auditLogger.Close()

	// Configuration is now managed via CLI flags and environment variables only

	// Validate inputs before proceeding
	if err := ValidateInputs(cmdConfig, logger); err != nil {
		os.Exit(1)
	}

	// Convert cmdConfig to architect.CliConfig to pass to core logic
	coreConfig := convertToArchitectConfig(cmdConfig)

	// CLI flags and environment variables are now the only source of configuration

	// Execute the core application logic
	err = architect.Execute(ctx, coreConfig, logger, auditLogger)
	if err != nil {
		logger.Error("Application failed: %v", err)
		os.Exit(1)
	}
}

// convertToArchitectConfig converts the cmd package CliConfig to the internal architect package CliConfig
func convertToArchitectConfig(cmdConfig *CliConfig) *architect.CliConfig {
	return &architect.CliConfig{
		InstructionsFile:           cmdConfig.InstructionsFile,
		OutputDir:                  cmdConfig.OutputDir,
		AuditLogFile:               cmdConfig.AuditLogFile,
		Format:                     cmdConfig.Format,
		Paths:                      cmdConfig.Paths,
		Include:                    cmdConfig.Include,
		Exclude:                    cmdConfig.Exclude,
		ExcludeNames:               cmdConfig.ExcludeNames,
		DryRun:                     cmdConfig.DryRun,
		Verbose:                    cmdConfig.Verbose,
		ApiKey:                     cmdConfig.ApiKey,
		ModelNames:                 cmdConfig.ModelNames,
		ConfirmTokens:              cmdConfig.ConfirmTokens,
		LogLevel:                   cmdConfig.LogLevel,
		MaxConcurrentRequests:      cmdConfig.MaxConcurrentRequests,
		RateLimitRequestsPerMinute: cmdConfig.RateLimitRequestsPerMinute,
	}
}
