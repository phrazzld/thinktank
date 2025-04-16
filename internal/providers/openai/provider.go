// internal/providers/openai/provider.go
package openai

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/openai"
	"github.com/phrazzld/architect/internal/providers"
)

// OpenAIProvider implements the Provider interface for OpenAI models.
type OpenAIProvider struct {
	logger logutil.LoggerInterface
	apiKey string
}

// NewProvider creates a new instance of OpenAIProvider.
func NewProvider(logger logutil.LoggerInterface) providers.Provider {
	// If no logger provided, create a default one
	if logger == nil {
		logger = logutil.NewLogger(logutil.InfoLevel, nil, "[openai-provider] ")
	}

	return &OpenAIProvider{
		logger: logger,
	}
}

// CreateClient implements the Provider interface.
func (p *OpenAIProvider) CreateClient(
	ctx context.Context,
	apiKey string,
	modelID string,
	apiEndpoint string,
) (llm.LLMClient, error) {
	p.logger.Debug("Creating OpenAI client for model: %s", modelID)

	// Use provided API key or fall back to environment variable
	effectiveAPIKey := apiKey
	if effectiveAPIKey == "" {
		effectiveAPIKey = os.Getenv("OPENAI_API_KEY")
		if effectiveAPIKey == "" {
			return nil, fmt.Errorf("no OpenAI API key provided and OPENAI_API_KEY environment variable not set")
		}
	}

	// Store API key for later use
	p.apiKey = effectiveAPIKey

	// We need to set the environment variable temporarily for the openai client
	// Since the OpenAI client directly reads from the environment
	oldAPIKey := os.Getenv("OPENAI_API_KEY")
	if err := os.Setenv("OPENAI_API_KEY", effectiveAPIKey); err != nil {
		return nil, fmt.Errorf("failed to set OpenAI API key in environment: %w", err)
	}
	defer func() {
		var err error
		if oldAPIKey != "" {
			err = os.Setenv("OPENAI_API_KEY", oldAPIKey)
		} else {
			err = os.Unsetenv("OPENAI_API_KEY")
		}
		if err != nil {
			p.logger.Warn("Failed to restore original OpenAI API key environment variable: %v", err)
		}
	}()

	// Create the client using the existing OpenAI implementation
	baseClient, err := openai.NewClient(modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	// Wrap the client in our parameter-aware adapter
	return NewOpenAIClientAdapter(baseClient), nil
}

// OpenAIClientAdapter wraps the standard OpenAI client to provide
// parameter-aware capabilities
type OpenAIClientAdapter struct {
	client llm.LLMClient
	params map[string]interface{}
	mu     sync.RWMutex
}

// NewOpenAIClientAdapter creates a new adapter for the OpenAI client
func NewOpenAIClientAdapter(client llm.LLMClient) *OpenAIClientAdapter {
	return &OpenAIClientAdapter{
		client: client,
		params: make(map[string]interface{}),
		mu:     sync.RWMutex{},
	}
}

// SetParameters sets the parameters to use for API calls
func (a *OpenAIClientAdapter) SetParameters(params map[string]interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.params = params
}

// GenerateContent implements the llm.LLMClient interface and applies parameters
func (a *OpenAIClientAdapter) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	// TODO: When llm.LLMClient interface evolves to support parameters directly,
	// this adapter will apply them to the underlying API calls.
	// For now, we just pass through to the existing client.

	return a.client.GenerateContent(ctx, prompt)
}

// CountTokens implements the llm.LLMClient interface
func (a *OpenAIClientAdapter) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return a.client.CountTokens(ctx, prompt)
}

// GetModelInfo implements the llm.LLMClient interface and uses configuration data
func (a *OpenAIClientAdapter) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	// TODO: When Registry integration is complete (T007), this will return
	// model information from the ModelDefinition in the registry.

	// For now, delegate to the underlying client
	return a.client.GetModelInfo(ctx)
}

// GetModelName implements the llm.LLMClient interface
func (a *OpenAIClientAdapter) GetModelName() string {
	return a.client.GetModelName()
}

// Close implements the llm.LLMClient interface
func (a *OpenAIClientAdapter) Close() error {
	return a.client.Close()
}
