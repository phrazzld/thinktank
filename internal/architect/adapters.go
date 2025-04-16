// Package architect provides core functionality for the Architect application
package architect

import (
	"context"
	"fmt"
	"strings"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/registry"
)

// APIServiceAdapter provides an adapter for different APIService implementations
// It allows using different implementations of APIService while maintaining
// backward compatibility with older code.
type APIServiceAdapter struct {
	// The underlying APIService implementation
	APIService interfaces.APIService
}

// InitLLMClient delegates to the underlying APIService implementation
func (a *APIServiceAdapter) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return a.APIService.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
}

// ProcessLLMResponse delegates to the underlying APIService implementation
func (a *APIServiceAdapter) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return a.APIService.ProcessLLMResponse(result)
}

// GetErrorDetails delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetErrorDetails(err error) string {
	return a.APIService.GetErrorDetails(err)
}

// IsEmptyResponseError delegates to the underlying APIService implementation
func (a *APIServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.APIService.IsEmptyResponseError(err)
}

// IsSafetyBlockedError delegates to the underlying APIService implementation
func (a *APIServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.APIService.IsSafetyBlockedError(err)
}

// GetModelParameters delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetModelParameters(modelName string) (map[string]interface{}, error) {
	// Check if the underlying implementation supports this method
	if apiService, ok := a.APIService.(interface {
		GetModelParameters(string) (map[string]interface{}, error)
	}); ok {
		return apiService.GetModelParameters(modelName)
	}
	// Return empty map for implementations that don't support this method
	return make(map[string]interface{}), nil
}

// GetModelDefinition delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	// Check if the underlying implementation supports this method
	if apiService, ok := a.APIService.(interface {
		GetModelDefinition(string) (*registry.ModelDefinition, error)
	}); ok {
		return apiService.GetModelDefinition(modelName)
	}
	// Return error for implementations that don't support this method
	return nil, fmt.Errorf("model definition not available")
}

// GetModelTokenLimits delegates to the underlying APIService implementation
func (a *APIServiceAdapter) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Check if the underlying implementation supports this method
	if apiService, ok := a.APIService.(interface {
		GetModelTokenLimits(string) (int32, int32, error)
	}); ok {
		return apiService.GetModelTokenLimits(modelName)
	}

	// Try to determine token limits based on model name
	// Use hardcoded values for well-known models
	switch {
	// Gemini models
	case strings.HasPrefix(modelName, "gemini-1.5-pro"):
		return 1000000, 8192, nil
	case strings.HasPrefix(modelName, "gemini-1.5-flash"):
		return 1000000, 8192, nil
	case strings.HasPrefix(modelName, "gemini-1.0-pro"):
		return 32768, 8192, nil
	case strings.HasPrefix(modelName, "gemini-1.0-ultra"):
		return 32768, 8192, nil

	// OpenAI models
	case strings.HasPrefix(modelName, "gpt-4-turbo"):
		return 128000, 4096, nil
	case strings.HasPrefix(modelName, "gpt-4"):
		return 8192, 4096, nil
	case strings.HasPrefix(modelName, "gpt-3.5-turbo"):
		return 16385, 4096, nil

	// Default fallback
	default:
		// Return error for unknown models
		return 0, 0, fmt.Errorf("token limits not available for model: %s", modelName)
	}
}

// ValidateModelParameter validates a parameter value against its constraints.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	if apiService, ok := a.APIService.(interface {
		ValidateModelParameter(string, string, interface{}) (bool, error)
	}); ok {
		return apiService.ValidateModelParameter(modelName, paramName, value)
	}
	// Return true if the underlying implementation doesn't support this method
	return true, nil
}

// TokenManagerAdapter provides an adapter for different TokenManager implementations
type TokenManagerAdapter struct {
	// The underlying TokenManager implementation
	TokenManager interfaces.TokenManager
}

// CountTokens delegates to the underlying TokenManager implementation
func (t *TokenManagerAdapter) CountTokens(content string) (int, error) {
	return t.TokenManager.CountTokens(content)
}

// GetInputTokenLimit delegates to the underlying TokenManager implementation
func (t *TokenManagerAdapter) GetInputTokenLimit() int {
	return t.TokenManager.GetInputTokenLimit()
}

// GetOutputTokenLimit delegates to the underlying TokenManager implementation
func (t *TokenManagerAdapter) GetOutputTokenLimit() int {
	return t.TokenManager.GetOutputTokenLimit()
}

// GetModelName delegates to the underlying TokenManager implementation
func (t *TokenManagerAdapter) GetModelName() string {
	return t.TokenManager.GetModelName()
}

// IsContentWithinLimits delegates to the underlying TokenManager implementation
func (t *TokenManagerAdapter) IsContentWithinLimits(content string) (bool, int, error) {
	return t.TokenManager.IsContentWithinLimits(content)
}

// ContextGathererAdapter provides an adapter for different ContextGatherer implementations
type ContextGathererAdapter struct {
	// The underlying ContextGatherer implementation
	ContextGatherer interfaces.ContextGatherer
}

// GatherContext delegates to the underlying ContextGatherer implementation
func (c *ContextGathererAdapter) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return c.ContextGatherer.GatherContext(ctx, config)
}

// DisplayDryRunInfo delegates to the underlying ContextGatherer implementation
func (c *ContextGathererAdapter) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return c.ContextGatherer.DisplayDryRunInfo(ctx, stats)
}

// FileWriterAdapter provides an adapter for different FileWriter implementations
type FileWriterAdapter struct {
	// The underlying FileWriter implementation
	FileWriter interfaces.FileWriter
}

// SaveToFile delegates to the underlying FileWriter implementation
func (f *FileWriterAdapter) SaveToFile(content, outputFile string) error {
	return f.FileWriter.SaveToFile(content, outputFile)
}
