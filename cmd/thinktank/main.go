// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// Main is the entry point for the thinktank CLI
func Main() {
	// As of Go 1.20, there's no need to seed the global random number generator
	// The runtime now automatically seeds it with a random value

	// Parse command line flags first to get the timeout value
	config, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create a base context with timeout and correlation ID
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel() // Ensure resources are released when Main exits
	ctx = logutil.WithCorrelationID(ctx)

	// Setup logging early for error reporting with context
	logger := SetupLogging(config)
	// Ensure context with correlation ID is attached to logger
	logger = logger.WithContext(ctx)
	logger.Info("Starting thinktank - AI-assisted content generation tool")

	// Initialize the audit logger
	var auditLogger auditlog.AuditLogger
	if config.AuditLogFile != "" {
		fileLogger, err := auditlog.NewFileAuditLogger(config.AuditLogFile, logger)
		if err != nil {
			// Log error and fall back to NoOp implementation
			logger.Error("Failed to initialize file audit logger: %v. Audit logging disabled.", err)
			auditLogger = auditlog.NewNoOpAuditLogger()
		} else {
			auditLogger = fileLogger
			logger.Info("Audit logging enabled to file: %s", config.AuditLogFile)
		}
	} else {
		auditLogger = auditlog.NewNoOpAuditLogger()
		logger.Debug("Audit logging is disabled")
	}

	// Ensure the audit logger is properly closed when the application exits
	defer func() { _ = auditLogger.Close() }()

	// Initialize and load the Registry
	registryManager := registry.GetGlobalManager(logger)
	if err := registryManager.Initialize(); err != nil {
		logger.Error("Failed to initialize registry: %v", err)
		os.Exit(1)
	}

	logger.Info("Registry initialized successfully")

	// Validate inputs before proceeding
	if err := ValidateInputs(config, logger); err != nil {
		os.Exit(1)
	}

	// Initialize APIService using Registry
	apiService := thinktank.NewRegistryAPIService(registryManager.GetRegistry(), logger)

	// Execute the core application logic
	err = thinktank.Execute(ctx, config, logger, auditLogger, apiService)
	if err != nil {
		logger.Error("Application failed: %v", err)

		// Check if we're in tolerant mode (partial success is considered ok)
		if config.PartialSuccessOk {
			// Import the orchestrator errors package to check for partial failure
			if errors.Is(err, thinktank.ErrPartialSuccess) {
				logger.Info("Exiting with success code despite partial failure (--partial-success-ok is enabled)")
				// Exit with success when some models succeed in tolerant mode
				return
			}
		}

		// Otherwise, or if it's not a partial failure, exit with error code
		os.Exit(1)
	}
}
