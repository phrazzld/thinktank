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

	// Create a base context with timeout
	rootCtx := context.Background()
	ctx, cancel := context.WithTimeout(rootCtx, config.Timeout)
	defer cancel() // Ensure resources are released when Main exits

	// Add correlation ID to the context for tracing
	correlationID := ""
	ctx = logutil.WithCorrelationID(ctx, correlationID) // Empty string means generate a new UUID
	currentCorrelationID := logutil.GetCorrelationID(ctx)

	// Setup logging early for error reporting with context
	logger := SetupLogging(config)
	// Ensure context with correlation ID is attached to logger
	logger = logger.WithContext(ctx)
	logger.InfoContext(ctx, "Starting thinktank - AI-assisted content generation tool")

	// Initialize the audit logger
	var auditLogger auditlog.AuditLogger
	if config.AuditLogFile != "" {
		fileLogger, err := auditlog.NewFileAuditLogger(config.AuditLogFile, logger)
		if err != nil {
			// Log error and fall back to NoOp implementation using context-aware method
			logger.ErrorContext(ctx, "Failed to initialize file audit logger: %v. Audit logging disabled.", err)
			auditLogger = auditlog.NewNoOpAuditLogger()
		} else {
			auditLogger = fileLogger
			logger.InfoContext(ctx, "Audit logging enabled to file: %s", config.AuditLogFile)
		}
	} else {
		auditLogger = auditlog.NewNoOpAuditLogger()
		logger.DebugContext(ctx, "Audit logging is disabled")
	}

	// Ensure the audit logger is properly closed when the application exits
	defer func() { _ = auditLogger.Close() }()

	// Log first audit entry with correlation ID
	if err := auditLogger.Log(ctx, auditlog.AuditEntry{
		Operation: "application_start",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"correlation_id": currentCorrelationID,
		},
		Message: "Application starting",
	}); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}

	// Initialize and load the Registry
	registryManager := registry.GetGlobalManager(logger)
	if err := registryManager.Initialize(); err != nil {
		logger.ErrorContext(ctx, "Failed to initialize registry: %v", err)
		if logErr := auditLogger.LogOp(ctx, "initialize_registry", "Failure", nil, nil, err); logErr != nil {
			logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
		}
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Registry initialized successfully")
	if err := auditLogger.LogOp(ctx, "initialize_registry", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}

	// Validate inputs before proceeding
	if err := ValidateInputs(config, logger); err != nil {
		if logErr := auditLogger.LogOp(ctx, "validate_inputs", "Failure", nil, nil, err); logErr != nil {
			logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
		}
		os.Exit(1)
	}

	if err := auditLogger.LogOp(ctx, "validate_inputs", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}

	// Initialize APIService using Registry
	apiService := thinktank.NewRegistryAPIService(registryManager.GetRegistry(), logger)

	// Execute the core application logic
	err = thinktank.Execute(ctx, config, logger, auditLogger, apiService)
	if err != nil {
		logger.ErrorContext(ctx, "Application failed: %v", err)
		if logErr := auditLogger.LogOp(ctx, "execution", "Failure", nil, nil, err); logErr != nil {
			logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
		}

		// Check if we're in tolerant mode (partial success is considered ok)
		if config.PartialSuccessOk {
			// Import the orchestrator errors package to check for partial failure
			if errors.Is(err, thinktank.ErrPartialSuccess) {
				logger.InfoContext(ctx, "Exiting with success code despite partial failure (--partial-success-ok is enabled)")
				if logErr := auditLogger.Log(ctx, auditlog.AuditEntry{
					Operation: "partial_success_exit",
					Status:    "Success",
					Inputs: map[string]interface{}{
						"reason": "tolerant_mode_enabled",
					},
					Message: "Exiting with success code despite partial failure",
				}); logErr != nil {
					logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
				}
				// Exit with success when some models succeed in tolerant mode
				return
			}
		}

		// Otherwise, or if it's not a partial failure, exit with error code
		os.Exit(1)
	}

	// Log successful completion
	if err := auditLogger.LogOp(ctx, "execution", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}

	if err := auditLogger.Log(ctx, auditlog.AuditEntry{
		Operation: "application_end",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"status": "success",
		},
		Message: "Application completed successfully",
	}); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}
}
