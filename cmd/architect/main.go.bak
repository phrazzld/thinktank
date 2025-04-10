// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/phrazzld/architect/internal/adapters/cliui"
	"github.com/phrazzld/architect/internal/adapters/filesystem"
	"github.com/phrazzld/architect/internal/adapters/git"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	contextService "github.com/phrazzld/architect/internal/context"
	"github.com/phrazzld/architect/internal/core"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/plan"
	"github.com/phrazzld/architect/internal/prompt"
)

// Main is the entry point for the architect CLI
func Main() {
	// Parse command line flags
	cliConfig, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging early for error reporting
	logger := SetupLogging(cliConfig)

	// Initialize the CLI UI adapter
	ui := cliui.NewAdapter(logger, os.Stdin)

	// Display startup info
	ui.DisplayInfo("Architect - An AI-assisted planning tool")
	ui.DisplayInfo("Version: 0.1.0")

	// Initialize XDG-compliant configuration system
	configManager := config.NewManager(logger)

	// Create a temporary NoopLogger for early initialization
	// This is needed because we need to load configuration before we can
	// initialize the real audit logger, but the config system needs an audit logger.
	tempAuditLogger := auditlog.NewNoopLogger()

	// Update the configManager with the temporary audit logger
	configManager.SetAuditLogger(tempAuditLogger)

	// Load configuration from files
	err = configManager.LoadFromFiles()
	if err != nil {
		ui.DisplayWarning("Failed to load configuration: %v", err)
		ui.DisplayInfo("Using default configuration")
	}

	// Ensure configuration directories exist
	if err := configManager.EnsureConfigDirs(); err != nil {
		ui.DisplayWarning("Failed to create configuration directories: %v", err)
	}

	// Convert CLI flags to the format needed for merging
	cliFlags := ConvertConfigToMap(cliConfig)

	// Merge CLI flags with loaded configuration
	if err := configManager.MergeWithFlags(cliFlags); err != nil {
		ui.DisplayWarning("Failed to merge CLI flags with configuration: %v", err)
	}

	// Get the final configuration
	appConfig := configManager.GetConfig()

	// Initialize structured audit logger with the loaded configuration
	auditLogger, err := initAuditLogger(appConfig, logger)
	if err != nil {
		ui.DisplayWarning("Failed to initialize audit logger: %v", err)
		ui.DisplayInfo("Using no-op audit logger")
		auditLogger = auditlog.NewNoopLogger()
	}
	defer auditLogger.Close() // Ensure logger is closed at program exit

	// Create LLM client (will be implemented in a future task, using a stub for now)
	llmClient := &stubLLMClient{}

	// Create prompt manager
	promptManager := prompt.NewManager(logger)

	// Create filesystem adapter
	fsAdapter := filesystem.NewOSFileSystem()
	
	// Create git checker
	gitChecker := git.NewCLIChecker(logger)
	
	// Create context gatherer
	contextGatherer := contextService.NewService(
		fsAdapter,
		gitChecker,
		logger,
		auditLogger,
	)
	
	// Create file writer
	fileWriter := filesystem.NewOSFileWriter()
	
	// Create plan generator
	planGenerator := plan.NewService(
		llmClient,
		promptManager,
		logger,
		auditLogger,
		ui,
	)

	// Create the core service
	service := core.NewService(
		ui,
		configManager,
		newLogProvider(logger, auditLogger),
		llmClient,
		promptManager,
		contextGatherer,
		planGenerator,
		fileWriter,
	)

	// Set up context
	ctx := context.Background()

	// For now, we'll still rely on the original main function for most functionality
	// In the future, this will be replaced by calls to service methods
	// For now, we're just integrating the task clarification functionality
	refinedTaskDesc := cliConfig.TaskDescription
	if cliConfig.ClarifyTask && !cliConfig.DryRun {
		var err error
		refinedTaskDesc, err = service.ClarifyTask(ctx, cliConfig.TaskDescription, cliConfig.ModelName)
		if err != nil {
			ui.DisplayError("Task clarification failed: %v", err)
			// Continue with original task description
			refinedTaskDesc = cliConfig.TaskDescription
		} else {
			ui.DisplayInfo("Using refined task: %s", refinedTaskDesc)
			// Update the CLI config with the refined task
			cliConfig.TaskDescription = refinedTaskDesc
		}
	}

	// For now, call the original main function with our initialized components
	// This will be replaced by proper service orchestration in future refactoring
	OriginalMain(cliConfig, logger, configManager, auditLogger, ui)
}

// For now, we'll still rely on the original main.go implementation
// This will be removed in future refactoring steps
func OriginalMain(cliConfig *CliConfig, loggerObj, configManagerObj, auditLoggerObj, uiObj interface{}) {
	// Recover the core service from the calling function's context
	// This is a bit of a hack, but it's temporary until we fully refactor main.go
	service := getService()
	if service == nil {
		fmt.Println("Error: Could not get service instance. This is a transitional implementation.")
		os.Exit(1)
	}

	// Create a context
	ctx := context.Background()

	// If the user specified a task file, read it
	taskDescription := cliConfig.TaskDescription
	if cliConfig.TaskFile != "" {
		content, err := os.ReadFile(cliConfig.TaskFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading task file: %v\n", err)
			os.Exit(1)
		}
		taskDescription = string(content)
	}

	// Get the project context
	// First, convert CLI flags to ContextConfig
	logger, ok := loggerObj.(logutil.LoggerInterface)
	if !ok {
		fmt.Println("Error: Logger is not of the expected type. This is a transitional implementation.")
		os.Exit(1)
	}

	logger.Info("Gathering project context from %d paths", len(cliConfig.Paths))
	
	projectContext, stats, err := service.GatherContext(
		ctx,
		cliConfig.Paths,
		strings.Split(cliConfig.Include, ","),
		strings.Split(cliConfig.Exclude, ","),
		strings.Split(cliConfig.ExcludeNames, ","),
		cliConfig.Format,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering context: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Gathered context from %d/%d files (%d skipped)",
		stats.ProcessedFiles, stats.TotalFiles, stats.SkippedFiles)

	// If in dry run mode, just display stats and exit
	if cliConfig.DryRun {
		ui, ok := uiObj.(core.UserInterface)
		if !ok {
			fmt.Println("Error: UI is not of the expected type. This is a transitional implementation.")
			os.Exit(1)
		}

		// Get model info for token limit
		modelInfo, err := service.GetModelInfo(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting model info: %v\n", err)
			os.Exit(1)
		}

		// Count tokens in the context
		tokenCount, _, _ := countTokens(ctx, projectContext, service.GetLLMClient())

		// Call the UI's dry run info display method
		ui.DisplayDryRunInfo(
			len(projectContext),
			strings.Count(projectContext, "\n")+1,
			int(tokenCount),
			stats.ProcessedFiles,
			stats.FileList,
			modelInfo.InputTokenLimit,
			float64(tokenCount)/float64(modelInfo.InputTokenLimit)*100,
		)
		os.Exit(0)
	}

	// Generate and save the plan
	err = service.GenerateAndSavePlan(
		ctx,
		taskDescription,
		projectContext,
		cliConfig.OutputFile,
		cliConfig.ModelName,
		cliConfig.ConfirmTokens,
		cliConfig.PromptTemplate,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating plan: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Plan successfully saved to %s\n", cliConfig.OutputFile)
	os.Exit(0)
}

// getService returns the service instance from the global state
// This is a temporary function for the transitional period
var globalService *core.Service

func setService(service *core.Service) {
	globalService = service
}

func getService() *core.Service {
	return globalService
}

// Helper function for token counting during the transition
func countTokens(ctx context.Context, text string, llmClient core.LLMClient) (int32, error) {
	result, err := llmClient.CountTokens(ctx, text)
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// initAuditLogger creates and configures the audit logger based on application configuration
func initAuditLogger(appConfig *config.AppConfig, logger interface{}) (auditlog.StructuredLogger, error) {
	// Temporary implementation - will be moved to an adapter in subsequent tasks
	// If audit logging is disabled, use a no-op logger
	if !appConfig.AuditLogEnabled {
		return auditlog.NewNoopLogger(), nil
	}

	// Create an audit logger that writes to the configured file
	auditLogger, err := auditlog.NewFileLogger(appConfig.AuditLogFile)
	if err != nil {
		return auditlog.NewNoopLogger(), fmt.Errorf("failed to create audit logger: %w", err)
	}

	return auditLogger, nil
}

// logProvider implements the core.LogProvider interface to provide access to loggers
type logProvider struct {
	logger      logutil.LoggerInterface
	auditLogger auditlog.StructuredLogger
}

func newLogProvider(logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) *logProvider {
	return &logProvider{
		logger:      logger,
		auditLogger: auditLogger,
	}
}

func (p *logProvider) GetLogger() logutil.LoggerInterface {
	return p.logger
}

func (p *logProvider) GetAuditLogger() auditlog.StructuredLogger {
	return p.auditLogger
}

// stubLLMClient is a temporary implementation of the core.LLMClient interface
// It will be replaced with a real implementation in a future refactoring task
type stubLLMClient struct{}

func (c *stubLLMClient) GenerateContent(ctx context.Context, prompt string) (*core.GenerationResult, error) {
	// For task clarification stage 1, return questions
	if strings.Contains(prompt, "clarify") {
		return &core.GenerationResult{
			Content:      `{"analysis":"Task needs clarification.","questions":["What is the expected output?","What constraints are there?","How should it integrate with existing code?"]}`,
			FinishReason: "STOP",
			TokenCount:   100,
		}, nil
	}
	// For task clarification stage 2, return refined task
	if strings.Contains(prompt, "refine") {
		return &core.GenerationResult{
			Content:      `{"refined_task":"Improved task description with details from answers.","key_points":["Output format specified","Performance constraints noted","Integration approach outlined"]}`,
			FinishReason: "STOP",
			TokenCount:   100,
		}, nil
	}

	// Default response
	return &core.GenerationResult{
		Content:      "Stub response",
		FinishReason: "STOP",
		TokenCount:   100,
	}, nil
}

func (c *stubLLMClient) CountTokens(ctx context.Context, text string) (*core.TokenResult, error) {
	return &core.TokenResult{Total: 100}, nil
}

func (c *stubLLMClient) GetModelInfo(ctx context.Context) (*core.ModelInfo, error) {
	return &core.ModelInfo{
		Name:             "stub-model",
		InputTokenLimit:  8192,
		OutputTokenLimit: 4096,
	}, nil
}

func (c *stubLLMClient) Close() error {
	return nil
}
