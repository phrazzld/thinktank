package architect

import (
	"context"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
)

// APIServiceAdapter adapts internal APIService to interfaces.APIService
type APIServiceAdapter struct {
	APIService APIService
}

// InitClient implements the interfaces.APIService method with backward compatibility
func (a *APIServiceAdapter) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	// Delegate to the wrapped service's backward-compatible method
	return a.APIService.InitClient(ctx, apiKey, modelName, apiEndpoint)
}

// InitLLMClient implements the new interfaces.APIService method
func (a *APIServiceAdapter) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// Delegate to the wrapped service's new method
	return a.APIService.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
}

// ProcessResponse implements the interfaces.APIService method with backward compatibility
func (a *APIServiceAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	// Delegate to the wrapped service
	return a.APIService.ProcessResponse(result)
}

// ProcessLLMResponse implements the new interfaces.APIService method
func (a *APIServiceAdapter) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	// Delegate to the wrapped service
	return a.APIService.ProcessLLMResponse(result)
}

// IsEmptyResponseError implements the interfaces.APIService method
func (a *APIServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.APIService.IsEmptyResponseError(err)
}

// IsSafetyBlockedError implements the interfaces.APIService method
func (a *APIServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.APIService.IsSafetyBlockedError(err)
}

// GetErrorDetails implements the interfaces.APIService method
func (a *APIServiceAdapter) GetErrorDetails(err error) string {
	return a.APIService.GetErrorDetails(err)
}

// TokenResultAdapter adapts TokenResult to modelproc.TokenResult
func TokenResultAdapter(tr *TokenResult) *modelproc.TokenResult {
	if tr == nil {
		return nil
	}
	return &modelproc.TokenResult{
		TokenCount:   tr.TokenCount,
		InputLimit:   tr.InputLimit,
		ExceedsLimit: tr.ExceedsLimit,
		LimitError:   tr.LimitError,
		Percentage:   tr.Percentage,
	}
}

// TokenManagerAdapter adapts internal TokenManager to interfaces.TokenManager
type TokenManagerAdapter struct {
	TokenManager TokenManager
}

func (t *TokenManagerAdapter) CheckTokenLimit(ctx context.Context, prompt string) error {
	return t.TokenManager.CheckTokenLimit(ctx, prompt)
}

func (t *TokenManagerAdapter) GetTokenInfo(ctx context.Context, prompt string) (*interfaces.TokenResult, error) {
	result, err := t.TokenManager.GetTokenInfo(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Convert the TokenResult to interfaces.TokenResult
	return &interfaces.TokenResult{
		TokenCount:   result.TokenCount,
		InputLimit:   result.InputLimit,
		ExceedsLimit: result.ExceedsLimit,
		LimitError:   result.LimitError,
		Percentage:   result.Percentage,
	}, nil
}

func (t *TokenManagerAdapter) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return t.TokenManager.PromptForConfirmation(tokenCount, threshold)
}

// ContextGathererAdapter adapts internal ContextGatherer to interfaces.ContextGatherer
type ContextGathererAdapter struct {
	ContextGatherer ContextGatherer
}

func (c *ContextGathererAdapter) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	// Convert interfaces.GatherConfig to internal GatherConfig
	internalConfig := GatherConfig{
		Paths:        config.Paths,
		Include:      config.Include,
		Exclude:      config.Exclude,
		ExcludeNames: config.ExcludeNames,
		Format:       config.Format,
		Verbose:      config.Verbose,
		LogLevel:     config.LogLevel,
	}

	files, stats, err := c.ContextGatherer.GatherContext(ctx, internalConfig)
	if err != nil {
		return nil, nil, err
	}

	// Convert internal ContextStats to interfaces.ContextStats
	interfaceStats := &interfaces.ContextStats{
		ProcessedFilesCount: stats.ProcessedFilesCount,
		CharCount:           stats.CharCount,
		LineCount:           stats.LineCount,
		TokenCount:          stats.TokenCount,
		ProcessedFiles:      stats.ProcessedFiles,
	}

	return files, interfaceStats, nil
}

func (c *ContextGathererAdapter) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	// Convert interfaces.ContextStats to internal ContextStats
	internalStats := &ContextStats{
		ProcessedFilesCount: stats.ProcessedFilesCount,
		CharCount:           stats.CharCount,
		LineCount:           stats.LineCount,
		TokenCount:          stats.TokenCount,
		ProcessedFiles:      stats.ProcessedFiles,
	}

	return c.ContextGatherer.DisplayDryRunInfo(ctx, internalStats)
}

// FileWriterAdapter adapts internal FileWriter to interfaces.FileWriter
type FileWriterAdapter struct {
	FileWriter FileWriter
}

func (f *FileWriterAdapter) SaveToFile(content, outputFile string) error {
	return f.FileWriter.SaveToFile(content, outputFile)
}
