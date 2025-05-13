// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank"
)

// Exit codes for different error types
const (
	ExitCodeSuccess             = 0
	ExitCodeGenericError        = 1
	ExitCodeAuthError           = 2
	ExitCodeRateLimitError      = 3
	ExitCodeInvalidRequest      = 4
	ExitCodeServerError         = 5
	ExitCodeNetworkError        = 6
	ExitCodeInputError          = 7
	ExitCodeContentFiltered     = 8
	ExitCodeInsufficientCredits = 9
	ExitCodeCancelled           = 10
)

// handleError processes an error, logs it appropriately, and exits the application with the correct exit code.
// It determines the error category, creates a user-friendly message, and ensures proper logging and audit trail.
func handleError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string) {
	if err == nil {
		return
	}

	// Log detailed error with context for debugging
	logger.ErrorContext(ctx, "Error: %v", err)

	// Audit the error
	logErr := auditLogger.LogOp(ctx, operation, "Failure", nil, nil, err)
	if logErr != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", logErr)
	}

	// Determine error category and appropriate exit code
	exitCode := ExitCodeGenericError
	var userMsg string

	// Check if the error is an LLMError that implements CategorizedError
	if catErr, ok := llm.IsCategorizedError(err); ok {
		category := catErr.Category()

		// Determine exit code based on error category
		switch category {
		case llm.CategoryAuth:
			exitCode = ExitCodeAuthError
		case llm.CategoryRateLimit:
			exitCode = ExitCodeRateLimitError
		case llm.CategoryInvalidRequest:
			exitCode = ExitCodeInvalidRequest
		case llm.CategoryServer:
			exitCode = ExitCodeServerError
		case llm.CategoryNetwork:
			exitCode = ExitCodeNetworkError
		case llm.CategoryInputLimit:
			exitCode = ExitCodeInputError
		case llm.CategoryContentFiltered:
			exitCode = ExitCodeContentFiltered
		case llm.CategoryInsufficientCredits:
			exitCode = ExitCodeInsufficientCredits
		case llm.CategoryCancelled:
			exitCode = ExitCodeCancelled
		}

		// Try to get a user-friendly message if it's an LLMError
		if llmErr, ok := catErr.(*llm.LLMError); ok {
			userMsg = llmErr.UserFacingError()
		} else {
			userMsg = fmt.Sprintf("%v", err)
		}
	} else if errors.Is(err, thinktank.ErrPartialSuccess) {
		// Special case for partial success errors
		userMsg = "Some model executions failed, but partial results were generated. Use --partial-success-ok flag to exit with success code in this case."
	} else {
		// Generic error - try to create a user-friendly message
		userMsg = getFriendlyErrorMessage(err)
	}

	// Print user-friendly message to stderr
	fmt.Fprintf(os.Stderr, "Error: %s\n", userMsg)

	// Exit with appropriate code
	os.Exit(exitCode)
}

// getFriendlyErrorMessage creates a user-friendly error message based on the error type
func getFriendlyErrorMessage(err error) string {
	if err == nil {
		return "An unknown error occurred"
	}

	// Check for common error patterns and provide friendly messages
	errMsg := err.Error()
	lowerMsg := strings.ToLower(errMsg)

	// Common error patterns
	switch {
	case strings.Contains(lowerMsg, "api key"), strings.Contains(lowerMsg, "auth"), strings.Contains(lowerMsg, "unauthorized"):
		return "Authentication error: Please check your API key and permissions"

	case strings.Contains(lowerMsg, "rate limit"), strings.Contains(lowerMsg, "too many requests"):
		return "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits."

	case strings.Contains(lowerMsg, "timeout"), strings.Contains(lowerMsg, "deadline exceeded"), strings.Contains(lowerMsg, "timed out"):
		return "Operation timed out. Consider using a longer timeout with the --timeout flag."

	case strings.Contains(lowerMsg, "not found"):
		return "Resource not found. Please check that the specified file paths or models exist."

	case strings.Contains(lowerMsg, "file"):
		if strings.Contains(lowerMsg, "permission") {
			return "File permission error: Please check file permissions and try again."
		}
		return "File error: " + sanitizeErrorMessage(errMsg)

	case strings.Contains(lowerMsg, "flag"), strings.Contains(lowerMsg, "usage"), strings.Contains(lowerMsg, "help"):
		return "Invalid command line arguments. Use --help to see usage instructions."

	case strings.Contains(lowerMsg, "context"):
		if strings.Contains(lowerMsg, "canceled") || strings.Contains(lowerMsg, "cancelled") {
			return "Operation was cancelled. This might be due to timeout or user interruption."
		}
		return "Context error: " + sanitizeErrorMessage(errMsg)

	case strings.Contains(lowerMsg, "network"), strings.Contains(lowerMsg, "connection"):
		return "Network error: Please check your internet connection and try again."
	}

	// If we can't identify a specific error type, just sanitize the original message
	return sanitizeErrorMessage(errMsg)
}

// sanitizeErrorMessage removes or masks sensitive information from error messages
// This is an additional layer beyond the sanitizing logger
func sanitizeErrorMessage(message string) string {
	// List of patterns to redact with corresponding replacements
	var redactedMsg string

	// API keys - OpenAI and all sk- patterns
	redactedMsg = "[REDACTED]"
	message = regexp.MustCompile(`sk[-_][a-zA-Z0-9]{16,}`).ReplaceAllString(message, redactedMsg)

	// API keys - Gemini and all key- patterns
	redactedMsg = "[REDACTED]"
	message = regexp.MustCompile(`key[-_][a-zA-Z0-9]{16,}`).ReplaceAllString(message, redactedMsg)

	// Long alphanumeric strings that might be API keys
	redactedMsg = "[REDACTED]"
	message = regexp.MustCompile(`[a-zA-Z0-9]{32,}`).ReplaceAllString(message, redactedMsg)

	// URLs with credentials
	redactedMsg = "[REDACTED]"
	message = regexp.MustCompile(`https?://[^:]+:[^@]+@[^/]+`).ReplaceAllString(message, redactedMsg)

	// Environment variables with secrets
	redactedMsg = "[REDACTED]"
	message = regexp.MustCompile(`GEMINI_API_KEY=.*`).ReplaceAllString(message, redactedMsg)
	message = regexp.MustCompile(`OPENAI_API_KEY=.*`).ReplaceAllString(message, redactedMsg)
	message = regexp.MustCompile(`OPENROUTER_API_KEY=.*`).ReplaceAllString(message, redactedMsg)
	message = regexp.MustCompile(`API_KEY=.*`).ReplaceAllString(message, redactedMsg)

	return message
}

// Main is the entry point for the thinktank CLI
func Main() {
	// As of Go 1.20, there's no need to seed the global random number generator
	// The runtime now automatically seeds it with a random value

	// Parse command line flags first to get the timeout value
	config, err := ParseFlags()
	if err != nil {
		// We don't have a logger or context yet, so handle this error specially
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitCodeInvalidRequest) // Use the appropriate exit code for invalid CLI flags
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
		// Use the central error handling mechanism
		handleError(ctx, err, logger, auditLogger, "initialize_registry")
	}

	logger.InfoContext(ctx, "Registry initialized successfully")
	if err := auditLogger.LogOp(ctx, "initialize_registry", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}

	// Validate inputs before proceeding
	if err := ValidateInputs(config, logger); err != nil {
		// Use the central error handling mechanism with input validation errors
		// These are considered invalid requests
		err = llm.Wrap(err, "thinktank", "Invalid input configuration", llm.CategoryInvalidRequest)
		handleError(ctx, err, logger, auditLogger, "validate_inputs")
	}

	if err := auditLogger.LogOp(ctx, "validate_inputs", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}

	// Initialize APIService using Registry
	apiService := thinktank.NewRegistryAPIService(registryManager.GetRegistry(), logger)

	// Execute the core application logic
	err = thinktank.Execute(ctx, config, logger, auditLogger, apiService)
	if err != nil {
		// Check if we're in tolerant mode (partial success is considered ok)
		if config.PartialSuccessOk && errors.Is(err, thinktank.ErrPartialSuccess) {
			logger.InfoContext(ctx, "Partial success accepted due to --partial-success-ok flag")
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

		// Use the central error handling for all other error types
		// The error might already be categorized, or handleError will categorize it
		handleError(ctx, err, logger, auditLogger, "execution")
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
