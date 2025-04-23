// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// Main is the entry point for the thinktank CLI
func Main() {
	// Seed the random number generator at program start
	rand.Seed(time.Now().UnixNano())

	// Create a base context
	ctx := context.Background()

	// Parse command line flags
	config, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging early for error reporting
	logger := SetupLogging(config)
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
		logger.Warn("Failed to initialize registry: %v. Falling back to legacy provider detection.", err)
	} else {
		logger.Info("Registry initialized successfully")

		// Set the registry manager getter for token management
		thinktank.SetRegistryManagerGetter(func() interface{} {
			return registryManager
		})
	}

	// Validate inputs before proceeding
	if err := ValidateInputs(config, logger); err != nil {
		os.Exit(1)
	}

	// Initialize APIService using Registry
	apiService := thinktank.NewRegistryAPIService(registryManager, logger)

	// Execute the core application logic
	err = thinktank.Execute(ctx, config, logger, auditLogger, apiService)
	if err != nil {
		logger.Error("Application failed: %v", err)
		os.Exit(1)
	}
}
