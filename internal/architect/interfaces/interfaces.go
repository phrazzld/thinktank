// Package interfaces provides shared interfaces used across the architect tool.
// It helps prevent import cycles between packages that depend on each other.
package interfaces

import (
	"context"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// TokenResult holds information about token counts and limits
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// APIService defines the interface for API-related operations
type APIService interface {
	// InitClient initializes and returns a Gemini client
	// Deprecated: Use InitLLMClient for new code that needs provider-agnostic functionality
	InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error)

	// InitLLMClient initializes and returns a provider-agnostic LLM client
	InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)

	// ProcessResponse processes the Gemini API response and extracts content
	// Deprecated: Use ProcessLLMResponse for new code that needs provider-agnostic functionality
	ProcessResponse(result *gemini.GenerationResult) (string, error)

	// ProcessLLMResponse processes a provider-agnostic response and extracts content
	ProcessLLMResponse(result *llm.ProviderResult) (string, error)

	// IsEmptyResponseError checks if an error is related to empty API responses
	IsEmptyResponseError(err error) bool

	// IsSafetyBlockedError checks if an error is related to safety filters
	IsSafetyBlockedError(err error) bool

	// GetErrorDetails extracts detailed information from an error
	GetErrorDetails(err error) string
}

// TokenManager defines the interface for token counting and management
type TokenManager interface {
	// GetTokenInfo retrieves token count information and checks limits
	GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error)

	// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
	CheckTokenLimit(ctx context.Context, prompt string) error

	// PromptForConfirmation asks for user confirmation to proceed if token count exceeds threshold
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// ContextStats holds information about processed files and context size
type ContextStats struct {
	ProcessedFilesCount int
	CharCount           int
	LineCount           int
	TokenCount          int32
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
	SaveToFile(content, outputFile string) error
}

// AuditLogger defines the interface for writing audit logs
type AuditLogger interface {
	// Log writes an audit entry to the log
	Log(entry auditlog.AuditEntry) error

	// Close closes the audit logger
	Close() error
}
