package architect

import (
	"context"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
)

// APIServiceAdapter adapts internal APIService to interfaces.APIService
type APIServiceAdapter struct {
	APIService APIService
}

func (a *APIServiceAdapter) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	return a.APIService.InitClient(ctx, apiKey, modelName)
}

func (a *APIServiceAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return a.APIService.ProcessResponse(result)
}

func (a *APIServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.APIService.IsEmptyResponseError(err)
}

func (a *APIServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.APIService.IsSafetyBlockedError(err)
}

func (a *APIServiceAdapter) GetErrorDetails(err error) string {
	return a.APIService.GetErrorDetails(err)
}

// TokenResultAdapter adapts TokenResult to modelproc.TokenResult
func TokenResultAdapter(tr *TokenResult) *modelproc.TokenResult {
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

func (t *TokenManagerAdapter) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return t.TokenManager.CheckTokenLimit(ctx, client, prompt)
}

func (t *TokenManagerAdapter) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*interfaces.TokenResult, error) {
	result, err := t.TokenManager.GetTokenInfo(ctx, client, prompt)
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

func (c *ContextGathererAdapter) GatherContext(ctx context.Context, client gemini.Client, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
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

	files, stats, err := c.ContextGatherer.GatherContext(ctx, client, internalConfig)
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

func (c *ContextGathererAdapter) DisplayDryRunInfo(ctx context.Context, client gemini.Client, stats *interfaces.ContextStats) error {
	// Convert interfaces.ContextStats to internal ContextStats
	internalStats := &ContextStats{
		ProcessedFilesCount: stats.ProcessedFilesCount,
		CharCount:           stats.CharCount,
		LineCount:           stats.LineCount,
		TokenCount:          stats.TokenCount,
		ProcessedFiles:      stats.ProcessedFiles,
	}

	return c.ContextGatherer.DisplayDryRunInfo(ctx, client, internalStats)
}

// FileWriterAdapter adapts internal FileWriter to interfaces.FileWriter
type FileWriterAdapter struct {
	FileWriter FileWriter
}

func (f *FileWriterAdapter) SaveToFile(content, outputFile string) error {
	return f.FileWriter.SaveToFile(content, outputFile)
}
