// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/orchestrator"
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
	result := processError(ctx, err, logger, auditLogger, operation)

	if !result.ShouldExit {
		return
	}

	// Print user-friendly message to stderr
	fmt.Fprintf(os.Stderr, "Error: %s\n", result.UserMessage)

	// Exit with appropriate code
	os.Exit(result.ExitCode)
}

// processError processes an error and returns structured result for testing
// This extracts the core error processing logic without os.Exit() side effects
func processError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string) *ErrorProcessingResult {
	if err == nil {
		return &ErrorProcessingResult{
			ExitCode:    ExitCodeSuccess,
			UserMessage: "",
			ShouldExit:  false,
			AuditLogged: false,
			AuditError:  nil,
		}
	}

	// Log detailed error with context for debugging
	logger.ErrorContext(ctx, "Error: %v", err)

	// Attempt audit logging
	auditErr := auditLogger.LogOp(ctx, operation, "Failure", nil, nil, err)
	auditLogged := true
	if auditErr != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", auditErr)
	}

	// Determine exit code and user message
	exitCode := getExitCodeFromError(err)
	userMessage := generateErrorMessage(err)

	return &ErrorProcessingResult{
		ExitCode:    exitCode,
		UserMessage: userMessage,
		ShouldExit:  true,
		AuditLogged: auditLogged,
		AuditError:  auditErr,
	}
}

// generateErrorMessage creates a user-friendly error message from any error
// This is the extracted, pure business logic for error message generation
func generateErrorMessage(err error) string {
	if err == nil {
		return "An unknown error occurred"
	}

	// Check if the error is an LLMError that implements CategorizedError
	if catErr, ok := llm.IsCategorizedError(err); ok {
		// Try to get a user-friendly message if it's an LLMError
		if llmErr, ok := catErr.(*llm.LLMError); ok {
			userMsg := llmErr.UserFacingError()

			// Add category-specific advice for better user experience
			switch llmErr.ErrorCategory {
			case llm.CategoryAuth:
				if userMsg == llmErr.Message {
					// If UserFacingError() just returns the raw message, enhance it
					return userMsg + ". Please check your API key and permissions."
				}
				return userMsg
			case llm.CategoryRateLimit:
				if userMsg == llmErr.Message {
					return userMsg + ". Please try again later or adjust rate limits."
				}
				return userMsg
			default:
				return userMsg
			}
		} else {
			return err.Error()
		}
	} else if errors.Is(err, thinktank.ErrPartialSuccess) {
		// Special case for partial success errors
		return "Some model executions failed, but partial results were generated. Use --partial-success-ok flag to exit with success code in this case."
	} else {
		// Generic error - try to create a user-friendly message
		return getFriendlyErrorMessage(err)
	}
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

// setupGracefulShutdown sets up signal handling for graceful shutdown on SIGINT (Ctrl+C) and SIGTERM.
// It returns a context that will be cancelled when an interrupt signal is received.
func setupGracefulShutdown(ctx context.Context) (context.Context, context.CancelFunc) {
	// Create a new context for signal handling
	signalCtx, signalCancel := context.WithCancel(ctx)

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register the channel to receive specific signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		select {
		case sig := <-sigChan:
			// User pressed Ctrl+C or sent SIGTERM
			fmt.Fprintf(os.Stderr, "\n🛑 Received %v signal, shutting down gracefully...\n", sig)
			signalCancel()
		case <-signalCtx.Done():
			// Context was cancelled for another reason, clean up signal handling
			signal.Stop(sigChan)
			close(sigChan)
		}
	}()

	return signalCtx, signalCancel
}

// RunMain executes the main application bootstrap logic with injected dependencies
// This function contains the extracted bootstrap logic from Main() to enable testing
func RunMain(mainConfig *MainConfig) *MainResult {
	// Parse command line flags first to get the timeout value
	// Use the injected Args and Getenv instead of os.Args and os.Getenv for testability
	config, err := ParseFlagsWithArgsAndEnv(mainConfig.Args, mainConfig.Getenv)
	if err != nil {
		// We don't have a logger or context yet, so handle this error specially
		// Return error result instead of calling os.Exit() for testability
		return &MainResult{
			ExitCode: ExitCodeInvalidRequest,
			Error:    err,
		}
	}

	// Create a base context with timeout
	rootCtx := context.Background()
	ctx, cancel := context.WithTimeout(rootCtx, config.Timeout)
	defer cancel() // Ensure resources are released when RunMain exits

	// Set up graceful shutdown on interrupt signals
	ctx, gracefulCancel := setupGracefulShutdown(ctx)
	defer gracefulCancel() // Ensure graceful cancel is called when RunMain exits

	// Add correlation ID to the context for tracing
	correlationID := ""
	ctx = logutil.WithCorrelationID(ctx, correlationID) // Empty string means generate a new UUID

	// Setup logging early for error reporting with context
	logger := SetupLogging(config)

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

	// Initialize APIService using models package
	apiService := thinktank.NewRegistryAPIService(logger)

	// Create and configure ConsoleWriter
	consoleWriter := logutil.NewConsoleWriter()

	// Create production dependencies and run
	runConfig := NewProductionRunConfig(ctx, config, logger, auditLogger, apiService, consoleWriter)
	result := Run(runConfig)

	// Handle the result - but don't call os.Exit(), return structured result
	if result.Error != nil && result.ExitCode != ExitCodeSuccess {
		// For bootstrap function, we need to handle errors differently
		// Use the injected ExitHandler for side effects (like error logging)
		// but don't actually exit - return the result instead
		errorResult := processError(ctx, result.Error, logger, auditLogger, "execution")
		return &MainResult{
			ExitCode:  errorResult.ExitCode,
			Error:     result.Error,
			RunResult: result,
		}
	}

	// Success case - return the run result
	return &MainResult{
		ExitCode:  result.ExitCode,
		Error:     result.Error,
		RunResult: result,
	}
}

// Main is the entry point for the thinktank CLI
func Main() {
	// Create production configuration and run
	mainConfig := NewProductionMainConfig()
	result := RunMain(mainConfig)

	// Handle the result and exit
	if result.Error != nil && result.ExitCode != ExitCodeSuccess {
		// Print user-friendly message to stderr
		if result.RunResult != nil {
			// Error occurred during execution phase - use processError for consistent formatting
			noOpLogger := logutil.NewLogger(logutil.InfoLevel, nil, "")
			errorResult := processError(context.Background(), result.Error, noOpLogger, auditlog.NewNoOpAuditLogger(), "execution")
			fmt.Fprintf(os.Stderr, "Error: %s\n", errorResult.UserMessage)
		} else {
			// Error occurred during bootstrap phase - simple error output
			fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
		}
	}

	// Exit with the determined code
	os.Exit(result.ExitCode)
}

// Run executes the core application business logic with injected dependencies
// This function contains the extracted business logic from Main() to enable testing
func Run(runConfig *RunConfig) *RunResult {
	startTime := time.Now()
	stats := &ExecutionStats{}

	// Use the injected context
	ctx := runConfig.Context
	config := runConfig.Config
	logger := runConfig.Logger
	auditLogger := runConfig.AuditLogger
	apiService := runConfig.APIService
	consoleWriter := runConfig.ConsoleWriter

	// Ensure context is attached to logger
	logger = logger.WithContext(ctx)
	logger.InfoContext(ctx, "Starting thinktank - AI-assisted content generation tool")

	// Get correlation ID from context for logging
	currentCorrelationID := logutil.GetCorrelationID(ctx)

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
	stats.AuditEntriesWritten++

	// Models package is used directly, no initialization required
	logger.InfoContext(ctx, "Models package ready for use")

	// Validate inputs before proceeding
	if err := ValidateInputs(config, logger); err != nil {
		// Use the central error handling mechanism with input validation errors
		// These are considered invalid requests
		err = llm.Wrap(err, "thinktank", "Invalid input configuration", llm.CategoryInvalidRequest)

		// Instead of calling handleError which would exit, determine exit code and return
		exitCode := getExitCodeFromError(err)
		stats.Duration = time.Since(startTime)

		return &RunResult{
			ExitCode: exitCode,
			Error:    err,
			Stats:    stats,
		}
	}

	if err := auditLogger.LogOp(ctx, "validate_inputs", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}
	stats.AuditEntriesWritten++

	// Configure ConsoleWriter (using injected dependency)
	consoleWriter.SetQuiet(config.Quiet)
	consoleWriter.SetNoProgress(config.NoProgress)

	// Execute the core application logic
	var err error
	if runConfig.ContextGatherer != nil {
		// For testing: use custom ContextGatherer when provided
		err = executeWithCustomContextGatherer(ctx, config, logger, auditLogger, apiService, consoleWriter, runConfig.ContextGatherer)
	} else {
		// Normal production execution
		err = thinktank.Execute(ctx, config, logger, auditLogger, apiService, consoleWriter)
	}
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
			stats.AuditEntriesWritten++

			// Return success for partial success in tolerant mode
			stats.Duration = time.Since(startTime)
			return &RunResult{
				ExitCode: ExitCodeSuccess,
				Error:    nil,
				Stats:    stats,
			}
		}

		// For all other error types, determine exit code and return
		exitCode := getExitCodeFromError(err)
		stats.Duration = time.Since(startTime)

		return &RunResult{
			ExitCode: exitCode,
			Error:    err,
			Stats:    stats,
		}
	}

	// Log successful completion
	if err := auditLogger.LogOp(ctx, "execution", "Success", nil, nil, nil); err != nil {
		logger.ErrorContext(ctx, "Failed to write audit log: %v", err)
	}
	stats.AuditEntriesWritten++

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
	stats.AuditEntriesWritten++

	// Success case
	stats.Duration = time.Since(startTime)
	return &RunResult{
		ExitCode: ExitCodeSuccess,
		Error:    nil,
		Stats:    stats,
	}
}

// getExitCodeFromError determines the appropriate exit code for an error
// This extracts the exit code determination logic from handleError for testability
func getExitCodeFromError(err error) int {
	if err == nil {
		return ExitCodeSuccess
	}

	// Check for LLM errors with specific categories (handles wrapped errors)
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) {
		switch llmErr.ErrorCategory {
		case llm.CategoryAuth:
			return ExitCodeAuthError
		case llm.CategoryRateLimit:
			return ExitCodeRateLimitError
		case llm.CategoryInvalidRequest:
			return ExitCodeInvalidRequest
		case llm.CategoryServer:
			return ExitCodeServerError
		case llm.CategoryNetwork:
			return ExitCodeNetworkError
		case llm.CategoryInputLimit:
			return ExitCodeInputError
		case llm.CategoryContentFiltered:
			return ExitCodeContentFiltered
		case llm.CategoryInsufficientCredits:
			return ExitCodeInsufficientCredits
		case llm.CategoryCancelled:
			return ExitCodeCancelled
		default:
			return ExitCodeGenericError
		}
	}

	// Check for partial success error
	if errors.Is(err, thinktank.ErrPartialSuccess) {
		return ExitCodeGenericError
	}

	// Check for context cancellation errors (fallback for wrapped errors)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ExitCodeCancelled
	}

	// Check error message for cancellation patterns (last resort)
	errMsg := err.Error()
	if llm.GetErrorCategoryFromMessage(errMsg) == llm.CategoryCancelled {
		return ExitCodeCancelled
	}

	// Default to generic error for unknown error types
	return ExitCodeGenericError
}

// executeWithCustomContextGatherer executes the core application logic with a custom ContextGatherer
// This is used for testing file filtering behavior by injecting a mock ContextGatherer
func executeWithCustomContextGatherer(
	ctx context.Context,
	cliConfig *config.CliConfig,
	logger logutil.LoggerInterface,
	auditLogger auditlog.AuditLogger,
	apiService interfaces.APIService,
	consoleWriter logutil.ConsoleWriter,
	contextGatherer interfaces.ContextGatherer,
) error {
	// Ensure the logger has the context attached
	logger = logger.WithContext(ctx)

	// Setup output directory (replicated from app.go setupOutputDirectory function)
	if cliConfig.OutputDir == "" {
		// Generate a unique timestamped run name (simplified for testing)
		runName := fmt.Sprintf("thinktank_test_%d", time.Now().Unix())
		cliConfig.OutputDir = runName
		logger.InfoContext(ctx, "Generated output directory: %s", cliConfig.OutputDir)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cliConfig.OutputDir, cliConfig.DirPermissions); err != nil {
		logger.ErrorContext(ctx, "Failed to create output directory %s: %v", cliConfig.OutputDir, err)
		return fmt.Errorf("failed to create output directory %s: %v", cliConfig.OutputDir, err)
	}

	// Read instructions file (replicated from app.go)
	instructionsContent, err := os.ReadFile(cliConfig.InstructionsFile)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to read instructions file %s: %v", cliConfig.InstructionsFile, err)
		return fmt.Errorf("failed to read instructions file %s: %v", cliConfig.InstructionsFile, err)
	}
	instructions := string(instructionsContent)
	logger.InfoContext(ctx, "Successfully read instructions from %s", cliConfig.InstructionsFile)

	// Create file writer
	fileWriter := thinktank.NewFileWriter(logger, auditLogger, cliConfig.DirPermissions, cliConfig.FilePermissions)

	// Create rate limiter from configuration
	rateLimiter := ratelimit.NewRateLimiter(
		cliConfig.MaxConcurrentRequests,
		cliConfig.RateLimitRequestsPerMinute,
	)

	// Create adapters for the interfaces (same pattern as normal Execute)
	apiServiceAdapter := &thinktank.APIServiceAdapter{APIService: apiService}
	fileWriterAdapter := &thinktank.FileWriterAdapter{FileWriter: fileWriter}

	// Create orchestrator with custom context gatherer (pass directly, no adapter needed)
	orch := orchestrator.NewOrchestrator(
		apiServiceAdapter,
		contextGatherer, // Pass the contextGatherer directly - it implements interfaces.ContextGatherer
		fileWriterAdapter,
		auditLogger,
		rateLimiter,
		cliConfig,
		logger,
		consoleWriter,
	)

	// Run the orchestrator with the instructions
	return orch.Run(ctx, instructions)
}
