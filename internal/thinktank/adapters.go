// Package thinktank provides core functionality for the thinktank application
package thinktank

import (
	"context"

	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
)

// APIServiceAdapter provides an adapter for different APIService implementations
// It allows using different implementations of APIService while maintaining
// backward compatibility with older code.
type APIServiceAdapter struct {
	// The underlying APIService implementation
	APIService interfaces.APIService
}

// InitLLMClient delegates to the underlying APIService implementation
// .nocover - pure wrapper method that simply delegates to underlying implementation
func (a *APIServiceAdapter) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return a.APIService.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
}

// ProcessLLMResponse delegates to the underlying APIService implementation
// .nocover - pure wrapper method that simply delegates to underlying implementation
func (a *APIServiceAdapter) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return a.APIService.ProcessLLMResponse(result)
}

// GetErrorDetails delegates to the underlying APIService implementation
// .nocover - pure wrapper method that simply delegates to underlying implementation
func (a *APIServiceAdapter) GetErrorDetails(err error) string {
	return a.APIService.GetErrorDetails(err)
}

// IsEmptyResponseError delegates to the underlying APIService implementation
// .nocover - pure wrapper method that simply delegates to underlying implementation
func (a *APIServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.APIService.IsEmptyResponseError(err)
}

// IsSafetyBlockedError delegates to the underlying APIService implementation
// .nocover - pure wrapper method that simply delegates to underlying implementation
func (a *APIServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.APIService.IsSafetyBlockedError(err)
}

// GetModelParameters delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return a.APIService.GetModelParameters(ctx, modelName)
}

// GetModelDefinition delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	return a.APIService.GetModelDefinition(ctx, modelName)
}

// GetModelTokenLimits delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	return a.APIService.GetModelTokenLimits(ctx, modelName)
}

// ValidateModelParameter validates a parameter value against its constraints.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	return a.APIService.ValidateModelParameter(ctx, modelName, paramName, value)
}

// Note: TokenManagerAdapter was removed as part of T032A
// to remove token handling from the application.

// ContextGathererAdapter provides an adapter for different ContextGatherer implementations
// It adapts the internal ContextGatherer interface to the interfaces.ContextGatherer interface
type ContextGathererAdapter struct {
	// The underlying ContextGatherer implementation from the internal package
	ContextGatherer ContextGatherer
}

// Convert internal GatherConfig to interfaces.GatherConfig
func internalToInterfacesGatherConfig(config interfaces.GatherConfig) GatherConfig {
	return GatherConfig{
		Paths:        config.Paths,
		Include:      config.Include,
		Exclude:      config.Exclude,
		ExcludeNames: config.ExcludeNames,
		Format:       config.Format,
		Verbose:      config.Verbose,
		LogLevel:     config.LogLevel,
	}
}

// Convert internal ContextStats to interfaces.ContextStats
func internalToInterfacesContextStats(stats *ContextStats) *interfaces.ContextStats {
	if stats == nil {
		return nil
	}
	return &interfaces.ContextStats{
		ProcessedFilesCount: stats.ProcessedFilesCount,
		CharCount:           stats.CharCount,
		LineCount:           stats.LineCount,
		// TokenCount field removed as part of T032F - token handling refactoring
		ProcessedFiles: stats.ProcessedFiles,
	}
}

// Convert interfaces.ContextStats to internal ContextStats
func interfacesToInternalContextStats(stats *interfaces.ContextStats) *ContextStats {
	if stats == nil {
		return nil
	}
	return &ContextStats{
		ProcessedFilesCount: stats.ProcessedFilesCount,
		CharCount:           stats.CharCount,
		LineCount:           stats.LineCount,
		// TokenCount field removed as part of T032F - token handling refactoring
		ProcessedFiles: stats.ProcessedFiles,
	}
}

// GatherContext adapts between interfaces.GatherConfig and internal GatherConfig
func (c *ContextGathererAdapter) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	// Convert interfaces.GatherConfig to internal GatherConfig
	internalConfig := internalToInterfacesGatherConfig(config)

	// Call the underlying internal implementation
	files, stats, err := c.ContextGatherer.GatherContext(ctx, internalConfig)

	// Convert internal ContextStats to interfaces.ContextStats if no error
	if err != nil {
		return nil, nil, err
	}

	return files, internalToInterfacesContextStats(stats), nil
}

// DisplayDryRunInfo adapts between interfaces.ContextStats and internal ContextStats
func (c *ContextGathererAdapter) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	// Convert interfaces.ContextStats to internal ContextStats
	internalStats := interfacesToInternalContextStats(stats)

	// Call the underlying internal implementation
	return c.ContextGatherer.DisplayDryRunInfo(ctx, internalStats)
}

// FileWriterAdapter provides an adapter for different FileWriter implementations
// It adapts the internal FileWriter interface to the interfaces.FileWriter interface
type FileWriterAdapter struct {
	// The underlying FileWriter implementation from the internal package
	FileWriter interfaces.FileWriter
}

// SaveToFile delegates to the underlying FileWriter implementation
// .nocover - pure wrapper method that simply delegates to underlying implementation
func (f *FileWriterAdapter) SaveToFile(ctx context.Context, content, outputFile string) error {
	return f.FileWriter.SaveToFile(ctx, content, outputFile)
}

// TokenCountingServiceAdapter adapts thinktank.TokenCountingService to interfaces.TokenCountingService
// This allows the orchestrator to use the concrete TokenCountingService implementation
// while depending only on the interface package.
type TokenCountingServiceAdapter struct {
	// The underlying TokenCountingService implementation
	TokenCountingService TokenCountingService
}

// CountTokens adapts thinktank types to interfaces types and delegates to the underlying service
func (t *TokenCountingServiceAdapter) CountTokens(ctx context.Context, req interfaces.TokenCountingRequest) (interfaces.TokenCountingResult, error) {
	// Convert interfaces types to thinktank types
	thinktankReq := TokenCountingRequest{
		Instructions:        req.Instructions,
		Files:               convertFileContent(req.Files),
		SafetyMarginPercent: req.SafetyMarginPercent,
	}

	// Call underlying service
	result, err := t.TokenCountingService.CountTokens(ctx, thinktankReq)
	if err != nil {
		return interfaces.TokenCountingResult{}, err
	}

	// Convert result back to interfaces types
	return interfaces.TokenCountingResult{
		TotalTokens:       result.TotalTokens,
		InstructionTokens: result.InstructionTokens,
		FileTokens:        result.FileTokens,
		Overhead:          result.Overhead,
	}, nil
}

// CountTokensForModel adapts types and delegates to the underlying service
func (t *TokenCountingServiceAdapter) CountTokensForModel(ctx context.Context, req interfaces.TokenCountingRequest, modelName string) (interfaces.ModelTokenCountingResult, error) {
	// Convert interfaces types to thinktank types
	thinktankReq := TokenCountingRequest{
		Instructions:        req.Instructions,
		Files:               convertFileContent(req.Files),
		SafetyMarginPercent: req.SafetyMarginPercent,
	}

	// Call underlying service
	result, err := t.TokenCountingService.CountTokensForModel(ctx, thinktankReq, modelName)
	if err != nil {
		return interfaces.ModelTokenCountingResult{}, err
	}

	// Convert result back to interfaces types
	return interfaces.ModelTokenCountingResult{
		TokenCountingResult: interfaces.TokenCountingResult{
			TotalTokens:       result.TotalTokens,
			InstructionTokens: result.InstructionTokens,
			FileTokens:        result.FileTokens,
			Overhead:          result.Overhead,
		},
		ModelName:     result.ModelName,
		TokenizerUsed: result.TokenizerUsed,
		Provider:      result.Provider,
		IsAccurate:    result.IsAccurate,
	}, nil
}

// GetCompatibleModels adapts types and delegates to the underlying service
func (t *TokenCountingServiceAdapter) GetCompatibleModels(ctx context.Context, req interfaces.TokenCountingRequest, availableProviders []string) ([]interfaces.ModelCompatibility, error) {
	// Convert interfaces types to thinktank types
	thinktankReq := TokenCountingRequest{
		Instructions:        req.Instructions,
		Files:               convertFileContent(req.Files),
		SafetyMarginPercent: req.SafetyMarginPercent,
	}

	// Call underlying service
	results, err := t.TokenCountingService.GetCompatibleModels(ctx, thinktankReq, availableProviders)
	if err != nil {
		return nil, err
	}

	// Convert results back to interfaces types
	interfaceResults := make([]interfaces.ModelCompatibility, len(results))
	for i, result := range results {
		interfaceResults[i] = interfaces.ModelCompatibility{
			ModelName:     result.ModelName,
			IsCompatible:  result.IsCompatible,
			TokenCount:    result.TokenCount,
			ContextWindow: result.ContextWindow,
			UsableContext: result.UsableContext,
			Provider:      result.Provider,
			TokenizerUsed: result.TokenizerUsed,
			IsAccurate:    result.IsAccurate,
			Reason:        result.Reason,
		}
	}

	return interfaceResults, nil
}

// convertFileContent converts interfaces.FileContent slice to thinktank.FileContent slice
func convertFileContent(interfaceFiles []interfaces.FileContent) []FileContent {
	thinktankFiles := make([]FileContent, len(interfaceFiles))
	for i, file := range interfaceFiles {
		thinktankFiles[i] = FileContent(file)
	}
	return thinktankFiles
}
