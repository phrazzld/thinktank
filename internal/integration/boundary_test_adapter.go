// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// BoundaryTestEnv represents a test environment where only external boundaries are mocked
type BoundaryTestEnv struct {
	// External boundaries that should be mocked
	APICaller    ExternalAPICaller
	Filesystem   FilesystemIO
	EnvProvider  EnvironmentProvider
	TimeProvider TimeProvider

	// Internal components that should use real implementations
	Logger          logutil.LoggerInterface
	APIService      interfaces.APIService
	ContextGatherer interfaces.ContextGatherer
	FileWriter      interfaces.FileWriter
	AuditLogger     auditlog.AuditLogger
	RateLimiter     *ratelimit.RateLimiter

	// Configuration
	Config *config.CliConfig

	// In-memory state for tests
	ModelOutputs map[string]string
	FileContents map[string][]byte
}

// NewBoundaryTestEnv creates a new test environment with mocked external boundaries
func NewBoundaryTestEnv(t testing.TB) *BoundaryTestEnv {
	// Create mocks for external boundaries
	apiCaller := &MockExternalAPICaller{}
	filesystem := NewMockFilesystemIO()
	envProvider := NewMockEnvironmentProvider()
	timeProvider := NewMockTimeProvider()

	// Parse log level
	loglevel, _ := logutil.ParseLogLevel("debug")

	// Create logger
	// Convert testing.TB to *testing.T if it is a T, otherwise use a simple fallback
	var logger logutil.LoggerInterface
	if tPtr, ok := t.(*testing.T); ok {
		logger = logutil.NewTestLogger(tPtr)
	} else {
		// Create a standard logger as fallback
		logger = logutil.NewStdLoggerAdapter(log.New(os.Stderr, "", log.LstdFlags))
	}

	// Create temp directory for outputs
	tempDir := filepath.Join(os.TempDir(), "thinktank-test")
	outputDir := filepath.Join(tempDir, "output")

	// Set up mock filesystem
	if err := filesystem.MkdirAll(tempDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create temp directory: %v", err))
	}
	if err := filesystem.MkdirAll(outputDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create output directory: %v", err))
	}

	// Set up configuration
	cfg := &config.CliConfig{
		OutputDir:                  outputDir,
		Verbose:                    true,
		LogLevel:                   loglevel,
		Format:                     "markdown",
		AuditLogFile:               filepath.Join(tempDir, "audit.log"),
		MaxConcurrentRequests:      2,
		RateLimitRequestsPerMinute: 60,
	}

	// Create API service with mocked external boundary
	apiService := NewBoundaryAPIService(apiCaller, envProvider, logger)

	// Create mock audit logger that uses the filesystem boundary
	auditLogger := NewBoundaryAuditLogger(filesystem, logger)

	// Create context gatherer
	contextGatherer := &BoundaryContextGatherer{
		filesystem: filesystem,
		logger:     logger,
	}

	// Create file writer
	fileWriter := &BoundaryFileWriter{
		filesystem: filesystem,
		logger:     logger,
	}

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(
		cfg.MaxConcurrentRequests,
		cfg.RateLimitRequestsPerMinute,
	)

	return &BoundaryTestEnv{
		APICaller:       apiCaller,
		Filesystem:      filesystem,
		EnvProvider:     envProvider,
		TimeProvider:    timeProvider,
		Logger:          logger,
		Config:          cfg,
		APIService:      apiService,
		ContextGatherer: contextGatherer,
		FileWriter:      fileWriter,
		AuditLogger:     auditLogger,
		RateLimiter:     rateLimiter,
		ModelOutputs:    make(map[string]string),
		FileContents:    make(map[string][]byte),
	}
}

// SetupModelResponse configures the mock API caller to return specific responses for a model
func (env *BoundaryTestEnv) SetupModelResponse(modelName, response string) {
	env.ModelOutputs[modelName] = response

	// Update the mock API caller to return the model response
	mockAPICaller := env.APICaller.(*MockExternalAPICaller)
	mockAPICaller.CallLLMAPIFunc = func(ctx context.Context, model, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
		if model == modelName {
			return &llm.ProviderResult{
				Content:      response,
				FinishReason: "stop",
			}, nil
		}

		// For other models, check if we have a configured response
		if content, ok := env.ModelOutputs[model]; ok {
			return &llm.ProviderResult{
				Content:      content,
				FinishReason: "stop",
			}, nil
		}

		// Default response
		return &llm.ProviderResult{
			Content:      "Default response for " + model,
			FinishReason: "stop",
		}, nil
	}
}

// SetupInstructionsFile creates a mock instructions file
func (env *BoundaryTestEnv) SetupInstructionsFile(content string) string {
	instructionsPath := filepath.Join(env.Config.OutputDir, "instructions.md")
	err := env.Filesystem.WriteFile(instructionsPath, []byte(content), 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to write instructions file: %v", err))
	}

	// Verify file was created in filesystem mock
	env.FileContents[instructionsPath] = []byte(content)

	// Set in config
	env.Config.InstructionsFile = instructionsPath
	return instructionsPath
}

// SetupModels configures the models to use
func (env *BoundaryTestEnv) SetupModels(modelNames []string, synthesisModel string) {
	env.Config.ModelNames = modelNames
	env.Config.SynthesisModel = synthesisModel

	// Set up default responses for all models
	for _, model := range modelNames {
		if _, ok := env.ModelOutputs[model]; !ok {
			env.SetupModelResponse(model, fmt.Sprintf("# Output from %s\n\nThis is test output from %s.", model, model))
		}
	}

	// Set up synthesis model response if specified
	if synthesisModel != "" && synthesisModel != modelNames[0] {
		if _, ok := env.ModelOutputs[synthesisModel]; !ok {
			env.SetupModelResponse(synthesisModel, "# Synthesized Output\n\nThis content combines insights from all models.")
		}
	}
}

// Run executes the thinktank orchestrator with the configured environment
func (env *BoundaryTestEnv) Run(ctx context.Context, instructions string) error {
	// Set up instructions file if not already set
	if env.Config.InstructionsFile == "" {
		instructionsPath := env.SetupInstructionsFile(instructions)
		// Double check the file was created in the filesystem
		exists, _ := env.Filesystem.Stat(instructionsPath)
		if !exists {
			return fmt.Errorf("failed to create instructions file: %s", instructionsPath)
		}
	}

	// Replace the execute function to use our environment
	originalOrchestratorConstructor := thinktank.GetOrchestratorConstructor()
	defer thinktank.SetOrchestratorConstructor(originalOrchestratorConstructor)

	thinktank.SetOrchestratorConstructor(func(
		apiService interfaces.APIService,
		contextGatherer interfaces.ContextGatherer,
		fileWriter interfaces.FileWriter,
		auditLogger auditlog.AuditLogger,
		rateLimiter *ratelimit.RateLimiter,
		config *config.CliConfig,
		logger logutil.LoggerInterface,
	) thinktank.Orchestrator {
		// Use the injected components from our test environment
		return thinktank.NewOrchestrator(
			env.APIService,
			env.ContextGatherer,
			env.FileWriter,
			env.AuditLogger,
			env.RateLimiter,
			env.Config,
			env.Logger,
		)
	})

	// Execute the thinktank application
	return thinktank.Execute(ctx, env.Config, env.Logger, env.AuditLogger, env.APIService)
}

// BoundaryAPIService implements interfaces.APIService with mocked external boundaries
type BoundaryAPIService struct {
	apiCaller   ExternalAPICaller
	envProvider EnvironmentProvider
	logger      logutil.LoggerInterface
}

// NewBoundaryAPIService creates a new APIService with mocked external boundaries
func NewBoundaryAPIService(apiCaller ExternalAPICaller, envProvider EnvironmentProvider, logger logutil.LoggerInterface) interfaces.APIService {
	return &BoundaryAPIService{
		apiCaller:   apiCaller,
		envProvider: envProvider,
		logger:      logger,
	}
}

// InitLLMClient initializes and returns a provider-agnostic LLM client
func (s *BoundaryAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// If API key is not provided, get it from environment variables
	if apiKey == "" {
		// Determine the environment variable name based on the model name
		envVarName := "OPENAI_API_KEY" // Default to OpenAI
		if modelName != "" {
			// Use model name to determine provider and env var
			if strings.Contains(strings.ToLower(modelName), "gemini") {
				envVarName = "GEMINI_API_KEY"
			} else if strings.Contains(strings.ToLower(modelName), "claude") {
				envVarName = "ANTHROPIC_API_KEY"
			}
		}

		apiKey = s.envProvider.GetEnv(envVarName)
	}

	// Return a mock client that uses our API caller
	return &BoundaryLLMClient{
		apiCaller: s.apiCaller,
		modelName: modelName,
		apiKey:    apiKey,
		endpoint:  apiEndpoint,
		logger:    s.logger,
	}, nil
}

// GetModelParameters retrieves parameter values for a given model
func (s *BoundaryAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	// Return default parameters
	return map[string]interface{}{
		"temperature": 0.7,
		"max_tokens":  1024,
	}, nil
}

// ValidateModelParameter validates a parameter value against its constraints
func (s *BoundaryAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	// Basic validation
	switch paramName {
	case "temperature":
		if temp, ok := value.(float64); ok && temp >= 0 && temp <= 1 {
			return true, nil
		}
		return false, fmt.Errorf("temperature must be a float between 0 and 1")
	case "max_tokens":
		if tokens, ok := value.(int); ok && tokens > 0 {
			return true, nil
		}
		return false, fmt.Errorf("max_tokens must be a positive integer")
	default:
		// Accept any value for other parameters
		return true, nil
	}
}

// GetModelDefinition retrieves the model definition
func (s *BoundaryAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	// Create a basic model definition
	return &registry.ModelDefinition{
		Name:     modelName,
		Provider: getProviderFromModelName(modelName),
	}, nil
}

// getProviderFromModelName determines the provider based on the model name
func getProviderFromModelName(modelName string) string {
	modelLower := strings.ToLower(modelName)
	if strings.Contains(modelLower, "gpt") {
		return "openai"
	} else if strings.Contains(modelLower, "gemini") {
		return "gemini"
	} else if strings.Contains(modelLower, "claude") {
		return "anthropic"
	} else {
		return "unknown"
	}
}

// GetModelTokenLimits retrieves token limits for a given model
func (s *BoundaryAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Return reasonable defaults based on model
	modelLower := strings.ToLower(modelName)
	if strings.Contains(modelLower, "gpt-4") {
		return 128000, 4096, nil
	} else if strings.Contains(modelLower, "gpt-3.5") {
		return 16385, 4096, nil
	} else if strings.Contains(modelLower, "gemini") {
		return 32768, 8192, nil
	} else if strings.Contains(modelLower, "claude") {
		return 100000, 4096, nil
	}

	// Default values
	return 4096, 1024, nil
}

// ProcessLLMResponse processes a provider-agnostic response
func (s *BoundaryAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil result")
	}
	return result.Content, nil
}

// IsEmptyResponseError checks if an error is related to empty API responses
func (s *BoundaryAPIService) IsEmptyResponseError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "empty") ||
		strings.Contains(strings.ToLower(err.Error()), "no content")
}

// IsSafetyBlockedError checks if an error is related to safety filters
func (s *BoundaryAPIService) IsSafetyBlockedError(err error) bool {
	if err == nil {
		return false
	}
	errLower := strings.ToLower(err.Error())
	return strings.Contains(errLower, "safety") ||
		strings.Contains(errLower, "content filter") ||
		strings.Contains(errLower, "moderation")
}

// GetErrorDetails extracts detailed information from an error
func (s *BoundaryAPIService) GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// BoundaryLLMClient implements llm.LLMClient using the API caller boundary
type BoundaryLLMClient struct {
	apiCaller ExternalAPICaller
	modelName string
	apiKey    string
	endpoint  string
	logger    logutil.LoggerInterface
}

// GenerateContent calls the API caller boundary to generate content
func (c *BoundaryLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return c.apiCaller.CallLLMAPI(ctx, c.modelName, prompt, params)
}

// GetModelName returns the model name
func (c *BoundaryLLMClient) GetModelName() string {
	return c.modelName
}

// Close closes the client
func (c *BoundaryLLMClient) Close() error {
	return nil
}

// BoundaryContextGatherer implements interfaces.ContextGatherer using mocked filesystem
type BoundaryContextGatherer struct {
	filesystem FilesystemIO
	logger     logutil.LoggerInterface
}

// GatherContext collects and processes files based on configuration
func (g *BoundaryContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	// For testing purposes, we'll return minimal context files
	files := []fileutil.FileMeta{
		{
			Path:    "test_file.go",
			Content: "package test\n\nfunc TestFunc() {\n\t// Test function\n}\n",
			// No need to set FileInfo for testing
		},
	}

	stats := &interfaces.ContextStats{
		ProcessedFilesCount: 1,
		CharCount:           54,
		LineCount:           5,
		ProcessedFiles:      []string{"test_file.go"},
	}

	return files, stats, nil
}

// DisplayDryRunInfo shows detailed information for dry run mode
func (g *BoundaryContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	g.logger.InfoContext(ctx, "Dry run mode: would process %d files with %d characters",
		stats.ProcessedFilesCount, stats.CharCount)
	return nil
}

// BoundaryFileWriter implements interfaces.FileWriter using mocked filesystem
type BoundaryFileWriter struct {
	filesystem FilesystemIO
	logger     logutil.LoggerInterface
}

// SaveToFile writes content to the specified file
func (w *BoundaryFileWriter) SaveToFile(content, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := w.filesystem.MkdirAll(dir, 0755); err != nil {
		w.logger.Error("Failed to create directory %s: %v", dir, err)
		return err
	}

	// Write file
	if err := w.filesystem.WriteFile(filePath, []byte(content), 0644); err != nil {
		w.logger.Error("Failed to write file %s: %v", filePath, err)
		return err
	}

	w.logger.Debug("Successfully wrote file: %s", filePath)
	return nil
}

// BoundaryAuditLogger implements auditlog.AuditLogger using mocked filesystem
type BoundaryAuditLogger struct {
	filesystem FilesystemIO
	logger     logutil.LoggerInterface
	entries    []auditlog.AuditEntry
}

// NewBoundaryAuditLogger creates a new audit logger with mocked filesystem
func NewBoundaryAuditLogger(filesystem FilesystemIO, logger logutil.LoggerInterface) auditlog.AuditLogger {
	return &BoundaryAuditLogger{
		filesystem: filesystem,
		logger:     logger,
		entries:    make([]auditlog.AuditEntry, 0),
	}
}

// Log writes an audit entry to the log
func (a *BoundaryAuditLogger) Log(entry auditlog.AuditEntry) error {
	// Add entry to in-memory log
	a.entries = append(a.entries, entry)

	// In a real implementation, we would write to a file
	// For testing, we just log it
	a.logger.Debug("Audit log: %s - %s", entry.Operation, entry.Status)
	return nil
}

// LogOp logs an operation with the given status
func (a *BoundaryAuditLogger) LogOp(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	// Create audit entry
	entry := auditlog.AuditEntry{
		Operation: operation,
		Status:    status,
		Inputs:    inputs,
		Outputs:   outputs,
		Message:   fmt.Sprintf("%s - %s", operation, status),
	}

	// Add error info if present
	if err != nil {
		entry.Error = &auditlog.ErrorInfo{
			Message: err.Error(),
			Type:    "GeneralError",
		}
	}

	// Log the entry
	return a.Log(entry)
}

// Close closes the audit logger
func (a *BoundaryAuditLogger) Close() error {
	// Nothing to close in the mock implementation
	return nil
}
