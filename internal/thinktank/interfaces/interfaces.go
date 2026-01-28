// Package interfaces provides shared interfaces used across the thinktank tool.
// It helps prevent import cycles between packages that depend on each other.
package interfaces

import (
	"context"

	"github.com/misty-step/thinktank/internal/auditlog"
	"github.com/misty-step/thinktank/internal/fileutil"
	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/models"
)

// APIService defines the interface for API-related operations
// It provides methods for initializing LLM clients, processing responses,
// handling errors, and accessing model configuration from the models package.
type APIService interface {
	// InitLLMClient initializes and returns a provider-agnostic LLM client
	// Parameters:
	//   - ctx: The context for client initialization
	//   - apiKey: The API key for authentication with the provider
	//   - modelName: The name of the model to use
	//   - apiEndpoint: Optional custom API endpoint (if empty, default is used)
	// Returns:
	//   - An initialized LLM client
	//   - An error if initialization fails
	InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)

	// GetModelParameters retrieves parameter values from the models package for a given model
	// It returns a map of parameter name to parameter value, applying defaults from the model definition
	// Parameters:
	//   - ctx: The context for the operation
	//   - modelName: The name of the model to get parameters for
	// Returns:
	//   - A map of parameter name to parameter value
	//   - An error if retrieval fails
	GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error)

	// ValidateModelParameter validates a parameter value against its constraints
	// It checks type validation and constraint validation (min/max for numeric, enum values for string)
	// Parameters:
	//   - ctx: The context for the operation
	//   - modelName: The name of the model the parameter belongs to
	//   - paramName: The name of the parameter to validate
	//   - value: The value to validate
	// Returns:
	//   - true if the parameter is valid, false otherwise
	//   - An error with details about why validation failed
	ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error)

	// GetModelDefinition retrieves the full model definition
	// Parameters:
	//   - ctx: The context for the operation
	//   - modelName: The name of the model to get the definition for
	// Returns:
	//   - The model definition
	//   - An error if retrieval fails
	GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error)

	// GetModelTokenLimits retrieves token limits from the models package for a given model
	// Parameters:
	//   - ctx: The context for the operation
	//   - modelName: The name of the model to get token limits for
	// Returns:
	//   - contextWindow: The maximum number of tokens the model can accept as input
	//   - maxOutputTokens: The maximum number of tokens the model can generate as output
	//   - An error if retrieval fails
	GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error)

	// ProcessLLMResponse processes a provider-agnostic response and extracts content
	// Parameters:
	//   - result: The provider result to process
	// Returns:
	//   - The processed content
	//   - An error if processing fails
	ProcessLLMResponse(result *llm.ProviderResult) (string, error)

	// IsEmptyResponseError checks if an error is related to empty API responses
	// Parameters:
	//   - err: The error to check
	// Returns:
	//   - true if the error is related to empty API responses, false otherwise
	IsEmptyResponseError(err error) bool

	// IsSafetyBlockedError checks if an error is related to safety filters
	// Parameters:
	//   - err: The error to check
	// Returns:
	//   - true if the error is related to safety filters, false otherwise
	IsSafetyBlockedError(err error) bool

	// GetErrorDetails extracts detailed information from an error
	// Parameters:
	//   - err: The error to get details for
	// Returns:
	//   - A user-friendly error message with details
	GetErrorDetails(err error) string
}

// ContextStats holds information about processed files and context size
type ContextStats struct {
	ProcessedFilesCount int
	CharCount           int
	LineCount           int
	ProcessedFiles      []string
}

// GatherConfig holds parameters needed for gathering context
type GatherConfig struct {
	Paths        []string
	Include      string
	Exclude      string
	ExcludeNames string
	Format       string
	Verbose      bool
	LogLevel     logutil.LogLevel
}

// ContextGatherer defines the interface for gathering project context
type ContextGatherer interface {
	// GatherContext collects and processes files based on configuration
	GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)

	// DisplayDryRunInfo shows detailed information for dry run mode
	DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error
}

// FileWriter defines the interface for file output writing
type FileWriter interface {
	// SaveToFile writes content to the specified file
	SaveToFile(ctx context.Context, content, outputFile string) error
}

// AuditLogger defines the interface for writing audit logs
type AuditLogger interface {
	// Log writes an audit entry to the log
	Log(entry auditlog.AuditEntry) error

	// Close closes the audit logger
	Close() error
}

// FileContent represents a single file's content for token counting.
type FileContent struct {
	// Path is the file path for identification and logging
	Path string

	// Content is the actual file content to count tokens for
	Content string
}

// TokenCountingRequest contains all inputs needed for token counting.
type TokenCountingRequest struct {
	// Instructions text to be sent to the model
	Instructions string

	// Files contains all file content to be processed
	Files []FileContent

	// SafetyMarginPercent is the percentage of context window to reserve for output (0-50%)
	// Default: 10% if not specified (0 value)
	SafetyMarginPercent uint8
}

// TokenCountingResult provides detailed breakdown of token usage.
type TokenCountingResult struct {
	// TotalTokens is the sum of all token counts
	TotalTokens int

	// InstructionTokens is tokens used by instructions
	InstructionTokens int

	// FileTokens is tokens used by file content
	FileTokens int

	// Overhead includes formatting and structural tokens
	Overhead int
}

// ModelTokenCountingResult provides model-specific token counting results.
type ModelTokenCountingResult struct {
	TokenCountingResult

	// ModelName is the model this count was calculated for
	ModelName string

	// TokenizerUsed indicates which tokenization method was used
	TokenizerUsed string // "tiktoken", "sentencepiece", "estimation"

	// Provider is the LLM provider for the model
	Provider string

	// IsAccurate indicates if accurate tokenization was used (vs estimation)
	IsAccurate bool
}

// ModelCompatibility provides detailed compatibility information for a model.
type ModelCompatibility struct {
	// ModelName is the name of the model being evaluated
	ModelName string

	// IsCompatible indicates if the model can handle the input
	IsCompatible bool

	// TokenCount is the actual token count for this model
	TokenCount int

	// ContextWindow is the model's maximum context size
	ContextWindow int

	// UsableContext is the context available after safety margin
	UsableContext int

	// Provider is the LLM provider for this model
	Provider string

	// TokenizerUsed indicates which tokenization method was used
	TokenizerUsed string

	// IsAccurate indicates if accurate tokenization was used
	IsAccurate bool

	// Reason explains why the model is incompatible (if applicable)
	Reason string
}

// TokenCountingService provides accurate token counting and model filtering capabilities.
// It replaces estimation-based model selection with precise token counting from
// instructions and file content, enabling better model compatibility decisions.
type TokenCountingService interface {
	// CountTokens calculates total tokens for instructions and file content.
	// Returns detailed breakdown of token usage for audit and logging purposes.
	CountTokens(ctx context.Context, req TokenCountingRequest) (TokenCountingResult, error)

	// CountTokensForModel calculates tokens using accurate tokenization for the specific model.
	// Falls back to estimation if accurate tokenization is not available for the model.
	CountTokensForModel(ctx context.Context, req TokenCountingRequest, modelName string) (ModelTokenCountingResult, error)

	// GetCompatibleModels returns models that can handle the input with detailed compatibility information.
	// Uses accurate token counting to determine which models can process the given request.
	GetCompatibleModels(ctx context.Context, req TokenCountingRequest, availableProviders []string) ([]ModelCompatibility, error)
}
