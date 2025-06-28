// Package cli provides the simplified command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/phrazzld/thinktank/internal/thinktank/orchestrator"
)

// Variable to allow mocking os.Exit in tests
var osExit = os.Exit

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

// Main is the entry point for the thinktank CLI
func Main() {
	// Parse simplified arguments directly
	simplifiedConfig, err := ParseSimpleArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		fmt.Fprintln(os.Stderr, "\nRun 'thinktank --help' for usage information.")
		osExit(ExitCodeInvalidRequest)
	}

	// Handle help request
	if simplifiedConfig.HelpRequested() {
		PrintHelpToStdout()
		osExit(ExitCodeSuccess)
	}

	// Determine model selection strategy
	modelNames, synthesisModel := selectModelsForConfig(simplifiedConfig)

	// Convert to MinimalConfig
	minimalConfig := &config.MinimalConfig{
		InstructionsFile: simplifiedConfig.InstructionsFile,
		TargetPaths:      strings.Fields(simplifiedConfig.TargetPath), // Split space-joined paths
		ModelNames:       modelNames,
		OutputDir:        "", // Will be set by output manager
		DryRun:           simplifiedConfig.HasFlag(FlagDryRun),
		Verbose:          simplifiedConfig.HasFlag(FlagVerbose),
		SynthesisModel:   synthesisModel, // Set by intelligent selection
		LogLevel:         logutil.InfoLevel,
		Timeout:          config.DefaultTimeout,
		Quiet:            simplifiedConfig.HasFlag(FlagQuiet),
		NoProgress:       simplifiedConfig.HasFlag(FlagNoProgress),
		JsonLogs:         simplifiedConfig.HasFlag(FlagJsonLogs),
		Format:           config.DefaultFormat,
		Exclude:          config.DefaultExcludes,
		ExcludeNames:     config.DefaultExcludeNames,
	}

	// Apply environment variables
	if err := applyEnvironmentVars(minimalConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(ExitCodeInvalidRequest)
	}

	// If verbose or debug flag is set, upgrade log level
	if minimalConfig.Verbose || simplifiedConfig.HasFlag(FlagDebug) {
		minimalConfig.LogLevel = logutil.DebugLevel
	}

	// Create logger with proper routing based on flags
	logger, loggerWrapper := createLoggerWithRouting(minimalConfig, "")
	defer func() { _ = loggerWrapper.Close() }()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), minimalConfig.Timeout)
	defer cancel()

	// Setup graceful shutdown
	ctx = setupGracefulShutdown(ctx, logger)

	// Add correlation ID
	correlationID := uuid.New().String()
	ctx = logutil.WithCorrelationID(ctx, correlationID)
	contextLogger := logger.WithContext(ctx)

	// Create output directory if not set
	if minimalConfig.OutputDir == "" {
		outputManager := NewOutputManager(contextLogger)
		outputDir, err := outputManager.CreateOutputDirectory("", 0755)
		if err != nil {
			contextLogger.ErrorContext(ctx, "Failed to create output directory: %v", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to create output directory: %v\n", err)
			osExit(ExitCodeGenericError)
		}
		minimalConfig.OutputDir = outputDir

		// Now that we have output directory, recreate logger with proper file routing
		// Close the previous logger wrapper first
		_ = loggerWrapper.Close()
		logger, loggerWrapper = createLoggerWithRouting(minimalConfig, outputDir)
		defer func() { _ = loggerWrapper.Close() }()
		contextLogger = logger.WithContext(ctx)
	}

	// Run the application
	err = runApplication(ctx, minimalConfig, contextLogger)
	if err != nil {
		handleError(ctx, err, contextLogger)
	}
}

// applyEnvironmentVars applies environment variables to MinimalConfig
// Only handles essential environment variables - API keys are handled elsewhere during validation
func applyEnvironmentVars(cfg *config.MinimalConfig) error {
	// No configuration environment variables - keep it simple!
	// Use CLI flags for all configuration options.
	// Environment variables are only for authentication (API keys).
	return nil
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func setupGracefulShutdown(ctx context.Context, logger logutil.LoggerInterface) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			logger.InfoContext(ctx, "Received signal %v, initiating graceful shutdown", sig)
			fmt.Fprintln(os.Stderr, "\nReceived interrupt signal. Shutting down gracefully...")
			cancel()
		case <-ctx.Done():
			// Context cancelled by other means
		}
	}()

	return ctx
}

// runApplication executes the core application logic with MinimalConfig
func runApplication(ctx context.Context, cfg *config.MinimalConfig, logger logutil.LoggerInterface) error {
	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return err
	}

	// Create audit logger
	var auditLogger auditlog.AuditLogger
	if cfg.DryRun {
		// In dry run mode, use no-op audit logger
		auditLogger = auditlog.NewNoOpAuditLogger()
	} else {
		// Use file audit logger writing to a log file
		auditLogPath := filepath.Join(cfg.OutputDir, "audit.jsonl")
		var err error
		auditLogger, err = auditlog.NewFileAuditLogger(auditLogPath, logger)
		if err != nil {
			return fmt.Errorf("failed to create audit logger: %w", err)
		}
	}

	// Log start
	logger.InfoContext(ctx, "Starting thinktank - AI-assisted content generation tool")

	// Read instructions
	instructionsContent, err := os.ReadFile(cfg.InstructionsFile)
	if err != nil {
		return fmt.Errorf("failed to read instructions file: %w", err)
	}
	instructions := string(instructionsContent)

	// In dry run mode, just show what would be processed
	if cfg.DryRun {
		return runDryRun(ctx, cfg, instructions, logger)
	}

	// Create necessary services
	consoleWriter := logutil.NewConsoleWriter()

	// Create registry API service that works with multiple providers
	apiService := thinktank.NewRegistryAPIService(logger)

	// Create a dummy LLM client for context gatherer (it's only needed for dry run)
	// In non-dry-run mode, the orchestrator will handle the actual client creation
	var dummyClient llm.LLMClient
	if cfg.DryRun {
		dummyClient = &llm.MockLLMClient{}
	}

	// Create context gatherer with all required parameters
	contextGatherer := thinktank.NewContextGatherer(logger, consoleWriter, cfg.DryRun, dummyClient, auditLogger)

	// Create file writer
	fileWriter := thinktank.NewFileWriter(logger, auditLogger, 0755, 0644)

	// Create rate limiter with smart defaults based on provider
	rateLimiter := createRateLimiter(cfg)

	// Create adapter config that implements the interface expected by orchestrator
	// This is temporary until we update orchestrator to use ConfigInterface
	adapterConfig := createAdapterConfig(cfg)

	// Create orchestrator with adapters for type compatibility
	orch := orchestrator.NewOrchestrator(
		apiService,
		&thinktank.ContextGathererAdapter{ContextGatherer: contextGatherer},
		fileWriter,
		auditLogger,
		rateLimiter,
		adapterConfig,
		logger,
		consoleWriter,
	)

	// Run orchestrator
	return orch.Run(ctx, instructions)
}

// validateConfig validates the minimal configuration
func validateConfig(cfg *config.MinimalConfig) error {
	if cfg.InstructionsFile == "" {
		return fmt.Errorf("instructions file is required")
	}

	if len(cfg.TargetPaths) == 0 {
		return fmt.Errorf("at least one target path is required")
	}

	// Check if instructions file exists
	if _, err := os.Stat(cfg.InstructionsFile); err != nil {
		return fmt.Errorf("instructions file not found: %w", err)
	}

	// Check if target paths exist
	for _, path := range cfg.TargetPaths {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("target path not found: %s", path)
		}
	}

	// Validate API keys based on models
	for _, model := range cfg.ModelNames {
		provider := getProviderForModel(model)
		apiKey := getAPIKeyForProvider(provider)
		if apiKey == "" && !cfg.DryRun {
			return fmt.Errorf("%s API key not set for model %s", provider, model)
		}
	}

	return nil
}

// getProviderForModel returns the provider for a given model name
func getProviderForModel(model string) string {
	modelInfo, err := models.GetModelInfo(model)
	if err != nil {
		// Default to gemini for unknown models
		if strings.Contains(strings.ToLower(model), "gpt") || strings.Contains(strings.ToLower(model), "o3") {
			return "openai"
		}
		if strings.Contains(strings.ToLower(model), "openrouter") {
			return "openrouter"
		}
		return "gemini"
	}
	return modelInfo.Provider
}

// getAPIKeyForProvider returns the API key for a given provider
func getAPIKeyForProvider(provider string) string {
	switch provider {
	case "openai":
		return os.Getenv(config.OpenAIAPIKeyEnvVar)
	case "openrouter":
		return os.Getenv(config.OpenRouterAPIKeyEnvVar)
	case "gemini":
		return os.Getenv(config.APIKeyEnvVar)
	default:
		return os.Getenv(config.APIKeyEnvVar)
	}
}

// createRateLimiter creates a rate limiter with smart defaults based on provider
func createRateLimiter(cfg *config.MinimalConfig) *ratelimit.RateLimiter {
	// Determine rate limits based on primary model provider
	if len(cfg.ModelNames) == 0 {
		return ratelimit.NewRateLimiter(5, 60) // Default
	}

	primaryModel := cfg.ModelNames[0]
	modelInfo, err := models.GetModelInfo(primaryModel)
	if err != nil {
		// Use conservative defaults
		return ratelimit.NewRateLimiter(5, 60)
	}

	// Use provider-specific defaults
	rpm := 60 // Default
	switch modelInfo.Provider {
	case "openai":
		rpm = 3000
	case "openrouter":
		rpm = 20
	}

	return ratelimit.NewRateLimiter(5, rpm)
}

// runDryRun executes a dry run showing what would be processed
func runDryRun(ctx context.Context, cfg *config.MinimalConfig, instructions string, logger logutil.LoggerInterface) error {
	// Respect quiet flag
	if !cfg.IsQuiet() {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Printf("Instructions file: %s\n", cfg.InstructionsFile)
		fmt.Printf("Target paths: %v\n", cfg.TargetPaths)
		fmt.Printf("Models: %v\n", cfg.ModelNames)
		fmt.Printf("Output directory: %s\n", cfg.OutputDir)
	}

	if !cfg.IsQuiet() && cfg.SynthesisModel != "" {
		fmt.Printf("Synthesis model: %s\n", cfg.SynthesisModel)
	}

	// Create console writer and dummy client for dry run
	consoleWriter := logutil.NewConsoleWriter()
	dummyClient := &llm.MockLLMClient{}
	noOpAuditLogger := auditlog.NewNoOpAuditLogger()

	// Create context gatherer to show what files would be processed
	contextGatherer := thinktank.NewContextGatherer(logger, consoleWriter, true, dummyClient, noOpAuditLogger)

	// Create app config for context gathering
	appConfig := &config.AppConfig{
		Format: cfg.Format,
		Excludes: config.ExcludeConfig{
			Extensions: cfg.Exclude,
			Names:      cfg.ExcludeNames,
		},
	}

	// Gather context
	// Create gather config
	gatherConfig := thinktank.GatherConfig{
		Paths:        cfg.TargetPaths,
		Format:       appConfig.Format,
		Exclude:      appConfig.Excludes.Extensions,
		ExcludeNames: appConfig.Excludes.Names,
	}

	files, stats, err := contextGatherer.GatherContext(ctx, gatherConfig)
	if err != nil {
		return fmt.Errorf("failed to gather context: %w", err)
	}
	_ = files // Files list is available if needed

	if !cfg.IsQuiet() {
		fmt.Printf("\nFiles that would be processed: %d\n", stats.ProcessedFilesCount)
		fmt.Printf("Total characters: %d\n", stats.CharCount)
		fmt.Printf("Total lines: %d\n", stats.LineCount)

		// Show first few files
		if len(stats.ProcessedFiles) > 0 {
			fmt.Println("\nSample files:")
			count := 10
			if len(stats.ProcessedFiles) < count {
				count = len(stats.ProcessedFiles)
			}
			for i := 0; i < count; i++ {
				fmt.Printf("  - %s\n", stats.ProcessedFiles[i])
			}
			if len(stats.ProcessedFiles) > count {
				fmt.Printf("  ... and %d more files\n", len(stats.ProcessedFiles)-count)
			}
		}
	}

	return nil
}

// createAdapterConfig creates a temporary adapter that makes MinimalConfig work with current orchestrator
// This will be removed once orchestrator is updated to use ConfigInterface
func createAdapterConfig(cfg *config.MinimalConfig) *config.CliConfig {
	return &config.CliConfig{
		InstructionsFile: cfg.InstructionsFile,
		Paths:            cfg.TargetPaths,
		ModelNames:       cfg.ModelNames,
		OutputDir:        cfg.OutputDir,
		DryRun:           cfg.DryRun,
		Verbose:          cfg.Verbose,
		SynthesisModel:   cfg.SynthesisModel,
		LogLevel:         cfg.LogLevel,
		Quiet:            cfg.Quiet,
		NoProgress:       cfg.NoProgress,
		Format:           cfg.Format,
		Exclude:          cfg.Exclude,
		ExcludeNames:     cfg.ExcludeNames,
		Timeout:          cfg.Timeout,
		// Set smart defaults for other fields
		MaxConcurrentRequests:      5,
		RateLimitRequestsPerMinute: 60,
		DirPermissions:             0755,
		FilePermissions:            0644,
		PartialSuccessOk:           false,
	}
}

// handleError processes an error and exits with appropriate code
func handleError(ctx context.Context, err error, logger logutil.LoggerInterface) {
	exitCode := getExitCode(err)
	userMessage := getUserMessage(err)

	logger.ErrorContext(ctx, "Application error: %v", err)
	fmt.Fprintf(os.Stderr, "Error: %s\n", userMessage)

	osExit(exitCode)
}

// getExitCode returns appropriate exit code for an error
func getExitCode(err error) int {
	if err == nil {
		return ExitCodeSuccess
	}

	// Check for specific error types
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

	// Check for CLI errors
	if cliErr, ok := IsCLIError(err); ok {
		switch cliErr.Type {
		case CLIErrorAuthentication:
			return ExitCodeAuthError
		case CLIErrorInvalidValue, CLIErrorMissingRequired:
			return ExitCodeInvalidRequest
		default:
			return ExitCodeGenericError
		}
	}

	// Check for context errors
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ExitCodeCancelled
	}

	// Check for partial success
	if errors.Is(err, thinktank.ErrPartialSuccess) {
		return ExitCodeGenericError
	}

	return ExitCodeGenericError
}

// LoggerWrapper wraps a logger and manages file closure
type LoggerWrapper struct {
	logutil.LoggerInterface
	file *os.File
}

// Close closes the underlying file if it exists
func (lw *LoggerWrapper) Close() error {
	if lw.file != nil {
		err := lw.file.Close()
		lw.file = nil // Make Close idempotent
		return err
	}
	return nil
}

// createLoggerWithRouting creates a logger with proper output routing based on CLI flags
func createLoggerWithRouting(cfg *config.MinimalConfig, outputDir string) (logutil.LoggerInterface, *LoggerWrapper) {
	// Determine where JSON logs should go
	shouldShowJsonLogsOnConsole := cfg.ShouldShowJsonLogs() || cfg.IsVerbose()

	if shouldShowJsonLogsOnConsole {
		// Legacy behavior: JSON logs to stderr (console)
		logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, cfg.GetLogLevel())
		return logger, &LoggerWrapper{LoggerInterface: logger, file: nil}
	} else {
		// Default behavior: JSON logs to file
		var logFilePath string
		if outputDir != "" {
			logFilePath = filepath.Join(outputDir, "thinktank.log")
		} else {
			// Use current directory as fallback for temporary logging
			logFilePath = "thinktank.log"
		}

		if logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			logger := logutil.NewSlogLoggerFromLogLevel(logFile, cfg.GetLogLevel())
			return logger, &LoggerWrapper{LoggerInterface: logger, file: logFile}
		}

		// Fallback to stderr if file creation fails
		logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, cfg.GetLogLevel())
		return logger, &LoggerWrapper{LoggerInterface: logger, file: nil}
	}
}

// selectModelsForConfig intelligently selects models based on config flags and input size estimation.
// Returns the list of model names and an optional synthesis model.
func selectModelsForConfig(simplifiedConfig *SimplifiedConfig) ([]string, string) {
	// Check if synthesis flag is explicitly set
	forceSynthesis := simplifiedConfig.HasFlag(FlagSynthesis)

	// Try to estimate input size by reading the instructions file
	var estimatedTokens int
	if instructionsContent, err := os.ReadFile(simplifiedConfig.InstructionsFile); err == nil {
		// For initial estimation, use just the instructions file
		// We'll refine this later when we have actual file content
		estimatedTokens = models.EstimateTokensFromText(string(instructionsContent))

		// Add a rough estimate for the target file(s)
		// This is conservative - we'll get exact numbers during context gathering
		const averageFileEstimate = 10000 // ~10K tokens for typical files
		estimatedTokens += averageFileEstimate
	} else {
		// Fallback estimate if we can't read instructions
		estimatedTokens = 15000 // Conservative fallback
	}

	// Get available providers (those with API keys set)
	availableProviders := models.GetAvailableProviders()
	if len(availableProviders) == 0 {
		// No API keys available, fall back to default model
		return []string{config.DefaultModel}, ""
	}

	// Select models that can handle the estimated input size
	selectedModels := models.SelectModelsForInput(estimatedTokens, availableProviders)

	// Determine synthesis behavior
	var synthesisModel string

	// Use synthesis if:
	// 1. Multiple models are selected, OR
	// 2. --synthesis flag is explicitly set
	if len(selectedModels) > 1 || forceSynthesis {
		// Always use gemini-2.5-pro as the default synthesis model for predictable behavior
		synthesisModel = "gemini-2.5-pro"
	}

	// If no models were selected (shouldn't happen with safety margins), fall back to default
	if len(selectedModels) == 0 {
		return []string{config.DefaultModel}, ""
	}

	// If only one model and no forced synthesis, use single model mode
	if len(selectedModels) == 1 && !forceSynthesis {
		return selectedModels, ""
	}

	return selectedModels, synthesisModel
}

// getUserMessage returns a user-friendly error message
func getUserMessage(err error) string {
	if err == nil {
		return "An unknown error occurred"
	}

	// Check for CLI errors first
	if cliErr, ok := IsCLIError(err); ok {
		return cliErr.UserFacingMessage()
	}

	// Check for LLM errors
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) {
		return llmErr.UserFacingError()
	}

	// Check for context errors
	if errors.Is(err, context.Canceled) {
		return "Operation was cancelled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Operation timed out"
	}

	// Check for partial success
	if errors.Is(err, thinktank.ErrPartialSuccess) {
		return "Some model executions failed, but partial results were generated"
	}

	// Return the error as-is
	return err.Error()
}
