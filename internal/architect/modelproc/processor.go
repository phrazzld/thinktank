// Package modelproc provides model processing functionality for the architect tool.
// It encapsulates the logic for interacting with AI models, managing tokens,
// writing outputs, and logging operations.
package modelproc

import (
	"context"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

// ModelProcessor handles all interactions with AI models including initialization,
// token management, request generation, response processing, and output handling.
type ModelProcessor struct {
	// Dependencies
	apiService   architect.APIService
	tokenManager architect.TokenManager
	fileWriter   architect.FileWriter
	auditLogger  auditlog.AuditLogger
	logger       logutil.LoggerInterface
	config       *config.CliConfig
}

// NewProcessor creates a new ModelProcessor with all required dependencies.
func NewProcessor(
	apiService architect.APIService,
	tokenManager architect.TokenManager,
	fileWriter architect.FileWriter,
	auditLogger auditlog.AuditLogger,
	logger logutil.LoggerInterface,
	config *config.CliConfig,
) *ModelProcessor {
	return &ModelProcessor{
		apiService:   apiService,
		tokenManager: tokenManager,
		fileWriter:   fileWriter,
		auditLogger:  auditLogger,
		logger:       logger,
		config:       config,
	}
}

// ProcessModel handles the entire model processing workflow including:
// - Initializing the client for a specific model
// - Checking token limits
// - Sending the prompt to the model
// - Processing the response
// - Writing the output to file
// - Auditing the operation
func (p *ModelProcessor) ProcessModel(ctx context.Context, modelName, prompt, outputPath string) (string, error) {
	// Initialize client for the specified model
	client, err := p.apiService.InitClient(ctx, p.config.ApiKey, modelName)
	if err != nil {
		p.logger.Error("Failed to initialize client for model %s: %v", modelName, err)
		return "", err
	}

	// Check token limit
	if err := p.tokenManager.CheckTokenLimit(ctx, client, prompt); err != nil {
		p.logger.Error("Token limit exceeded for model %s: %v", modelName, err)
		return "", err
	}

	// Get token information for audit and confirmation
	tokenInfo, err := p.tokenManager.GetTokenInfo(ctx, client, prompt)
	if err != nil {
		p.logger.Error("Failed to get token information: %v", err)
		return "", err
	}

	// Prompt for confirmation if token count exceeds threshold
	if !p.tokenManager.PromptForConfirmation(tokenInfo.TokenCount, p.config.ConfirmTokens) {
		p.logger.Info("Operation cancelled by user due to token count.")
		return "", nil
	}

	// Generate content
	p.logger.Info("Generating content with model: %s", modelName)
	result, err := client.GenerateContent(ctx, prompt)
	if err != nil {
		p.logger.Error("Failed to generate content: %v", err)
		return "", err
	}

	// Process response
	content, err := p.apiService.ProcessResponse(result)
	if err != nil {
		p.logger.Error("Failed to process response: %v", err)
		return "", err
	}

	// Save to file if output path is provided
	if outputPath != "" {
		if err := p.fileWriter.SaveToFile(content, outputPath); err != nil {
			p.logger.Error("Failed to save content to file: %v", err)
			return content, err
		}
	}

	return content, nil
}
