// Package modelproc provides model processing functionality for the architect tool.
// It encapsulates the logic for interacting with AI models, managing tokens,
// writing outputs, and logging operations.
package modelproc

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// APIService defines the interface for API-related operations
type APIService interface {
	// InitClient initializes and returns a Gemini client
	InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error)

	// ProcessResponse processes the API response and extracts content
	ProcessResponse(result *gemini.GenerationResult) (string, error)

	// IsEmptyResponseError checks if an error is related to empty API responses
	IsEmptyResponseError(err error) bool

	// IsSafetyBlockedError checks if an error is related to safety filters
	IsSafetyBlockedError(err error) bool

	// GetErrorDetails extracts detailed information from an error
	GetErrorDetails(err error) string
}

// TokenResult holds information about token counts and limits
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// TokenManager defines the interface for token counting and management
type TokenManager interface {
	// GetTokenInfo retrieves token count information and checks limits
	GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error)

	// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
	CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error

	// PromptForConfirmation asks for user confirmation to proceed if token count exceeds threshold
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// FileWriter defines the interface for file output writing
type FileWriter interface {
	// SaveToFile writes content to the specified file
	SaveToFile(content, outputFile string) error
}

// ModelProcessor handles all interactions with AI models including initialization,
// token management, request generation, response processing, and output handling.
type ModelProcessor struct {
	// Dependencies
	apiService   APIService
	tokenManager TokenManager
	fileWriter   FileWriter
	auditLogger  auditlog.AuditLogger
	logger       logutil.LoggerInterface
	config       *config.CliConfig
}

// NewProcessor creates a new ModelProcessor with all required dependencies.
func NewProcessor(
	apiService APIService,
	tokenManager TokenManager,
	fileWriter FileWriter,
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

// Process handles the entire model processing workflow for a single model.
// It implements the logic from the previous processModel/processModelConcurrently functions,
// including initialization, token checking, generation, response processing, and output saving.
func (p *ModelProcessor) Process(ctx context.Context, modelName string, stitchedPrompt string) error {
	p.logger.Info("Processing model: %s", modelName)

	// 1. Initialize model-specific client
	geminiClient, err := p.apiService.InitClient(ctx, p.config.ApiKey, modelName)
	if err != nil {
		errorDetails := p.apiService.GetErrorDetails(err)
		if apiErr, ok := gemini.IsAPIError(err); ok {
			p.logger.Error("Error creating Gemini client for model %s: %s", modelName, apiErr.Message)
			if apiErr.Suggestion != "" {
				p.logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			if p.config.LogLevel == logutil.DebugLevel {
				p.logger.Debug("Error details: %s", apiErr.DebugInfo())
			}
		} else {
			p.logger.Error("Error creating Gemini client for model %s: %s", modelName, errorDetails)
		}
		return fmt.Errorf("failed to initialize API client for model %s: %w", modelName, err)
	}
	defer geminiClient.Close()

	// 2. Check token limits for this model
	p.logger.Info("Checking token limits for model %s...", modelName)

	// Log token check start with prompt information
	if logErr := p.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokensStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    modelName,
		},
		Message: "Starting token count check for model " + modelName,
	}); logErr != nil {
		p.logger.Error("Failed to write audit log: %v", logErr)
	}

	tokenInfo, err := p.tokenManager.GetTokenInfo(ctx, geminiClient, stitchedPrompt)
	if err != nil {
		p.logger.Error("Token count check failed for model %s", modelName)

		// Determine error type for better categorization
		errorType := "TokenCheckError"
		errorMessage := fmt.Sprintf("Failed to check token count for model %s: %v", modelName, err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			p.logger.Error("Token count check failed for model %s: %s", modelName, apiErr.Message)
			if apiErr.Suggestion != "" {
				p.logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			p.logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			p.logger.Error("Token count check failed for model %s: %v", modelName, err)
			p.logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		// Log the token check failure
		if logErr := p.auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Token count check failed for model " + modelName,
		}); logErr != nil {
			p.logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token count check failed for model %s: %w", modelName, err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		p.logger.Error("Token limit exceeded for model %s", modelName)
		p.logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		p.logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")

		// Log the token limit exceeded case
		if logErr := p.auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    modelName,
			},
			TokenCounts: &auditlog.TokenCountInfo{
				PromptTokens: tokenInfo.TokenCount,
				TotalTokens:  tokenInfo.TokenCount,
				Limit:        tokenInfo.InputLimit,
			},
			Error: &auditlog.ErrorInfo{
				Message: tokenInfo.LimitError,
				Type:    "TokenLimitExceededError",
			},
			Message: "Token limit exceeded for model " + modelName,
		}); logErr != nil {
			p.logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token limit exceeded for model %s: %s", modelName, tokenInfo.LimitError)
	}

	// Prompt for confirmation if token count exceeds threshold
	if !p.tokenManager.PromptForConfirmation(tokenInfo.TokenCount, p.config.ConfirmTokens) {
		p.logger.Info("Operation cancelled by user due to token count.")
		return nil
	}

	p.logger.Info("Token check passed for model %s: %d / %d tokens (%.1f%% of limit)",
		modelName, tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// Log the successful token check
	if logErr := p.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokens",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    modelName,
		},
		Outputs: map[string]interface{}{
			"percentage": tokenInfo.Percentage,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: tokenInfo.TokenCount,
			TotalTokens:  tokenInfo.TokenCount,
			Limit:        tokenInfo.InputLimit,
		},
		Message: fmt.Sprintf("Token check passed for model %s: %d / %d tokens (%.1f%% of limit)",
			modelName, tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage),
	}); logErr != nil {
		p.logger.Error("Failed to write audit log: %v", logErr)
	}

	// 3. Generate content with this model
	p.logger.Info("Generating output with model %s (Temperature: %.2f, MaxOutputTokens: %d)...",
		modelName,
		geminiClient.GetTemperature(),
		geminiClient.GetMaxOutputTokens())

	// Log the start of content generation
	generateStartTime := time.Now()
	if logErr := p.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: generateStartTime,
		Operation: "GenerateContentStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"model_name":    modelName,
			"prompt_length": len(stitchedPrompt),
		},
		Message: "Starting content generation with Gemini model " + modelName,
	}); logErr != nil {
		p.logger.Error("Failed to write audit log: %v", logErr)
	}

	result, err := geminiClient.GenerateContent(ctx, stitchedPrompt)

	// Calculate duration in milliseconds
	generateDurationMs := time.Since(generateStartTime).Milliseconds()

	if err != nil {
		p.logger.Error("Generation failed for model %s", modelName)

		// Determine error type for better categorization
		errorType := "ContentGenerationError"
		errorMessage := fmt.Sprintf("Failed to generate content with model %s: %v", modelName, err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			p.logger.Error("Error generating content with model %s: %s", modelName, apiErr.Message)
			if apiErr.Suggestion != "" {
				p.logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			p.logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			p.logger.Error("Error generating content with model %s: %v (Current token count: %d)", modelName, err, tokenInfo.TokenCount)
		}

		// Log the content generation failure
		if logErr := p.auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GenerateContentEnd",
			Status:     "Failure",
			DurationMs: &generateDurationMs,
			Inputs: map[string]interface{}{
				"model_name":    modelName,
				"prompt_length": len(stitchedPrompt),
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Content generation failed for model " + modelName,
		}); logErr != nil {
			p.logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("output generation failed for model %s: %w", modelName, err)
	}

	// Log successful content generation
	if logErr := p.auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GenerateContentEnd",
		Status:     "Success",
		DurationMs: &generateDurationMs,
		Inputs: map[string]interface{}{
			"model_name":    modelName,
			"prompt_length": len(stitchedPrompt),
		},
		Outputs: map[string]interface{}{
			"finish_reason":      result.FinishReason,
			"has_safety_ratings": len(result.SafetyRatings) > 0,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: int32(tokenInfo.TokenCount),
			OutputTokens: int32(result.TokenCount),
			TotalTokens:  int32(tokenInfo.TokenCount + result.TokenCount),
		},
		Message: "Content generation completed successfully for model " + modelName,
	}); logErr != nil {
		p.logger.Error("Failed to write audit log: %v", logErr)
	}

	// 4. Process API response
	generatedOutput, err := p.apiService.ProcessResponse(result)
	if err != nil {
		// Get detailed error information
		errorDetails := p.apiService.GetErrorDetails(err)

		// Provide specific error messages based on error type
		if p.apiService.IsEmptyResponseError(err) {
			p.logger.Error("Received empty or invalid response from Gemini API for model %s", modelName)
			p.logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response for model %s due to empty content: %w", modelName, err)
		} else if p.apiService.IsSafetyBlockedError(err) {
			p.logger.Error("Content was blocked by Gemini safety filters for model %s", modelName)
			p.logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response for model %s due to safety restrictions: %w", modelName, err)
		} else {
			// Generic API error handling
			return fmt.Errorf("failed to process API response for model %s: %w", modelName, err)
		}
	}
	contentLength := len(generatedOutput)
	p.logger.Info("Output generated successfully with model %s (content length: %d characters, tokens: %d)",
		modelName, contentLength, result.TokenCount)

	// 5. Sanitize model name for use in filename
	sanitizedModelName := sanitizeFilename(modelName)

	// 6. Construct output file path
	outputFilePath := filepath.Join(p.config.OutputDir, sanitizedModelName+".md")

	// 7. Save the output to file
	if err := p.saveOutputToFile(outputFilePath, generatedOutput); err != nil {
		return fmt.Errorf("failed to save output for model %s: %w", modelName, err)
	}

	p.logger.Info("Successfully processed model: %s", modelName)
	return nil
}

// sanitizeFilename replaces characters that are not valid in filenames
func sanitizeFilename(filename string) string {
	// Replace slashes and other problematic characters with hyphens
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(filename)
}

// saveOutputToFile is a helper method that saves the generated output to a file
// and includes audit logging around the file writing operation.
func (p *ModelProcessor) saveOutputToFile(outputFilePath, content string) error {
	// Log the start of output saving
	saveStartTime := time.Now()
	if logErr := p.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: saveStartTime,
		Operation: "SaveOutputStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"output_path":    outputFilePath,
			"content_length": len(content),
		},
		Message: "Starting to save output to file",
	}); logErr != nil {
		p.logger.Error("Failed to write audit log: %v", logErr)
	}

	// Save output file
	p.logger.Info("Writing output to %s...", outputFilePath)
	err := p.fileWriter.SaveToFile(content, outputFilePath)

	// Calculate duration in milliseconds
	saveDurationMs := time.Since(saveStartTime).Milliseconds()

	if err != nil {
		// Log failure to save output
		p.logger.Error("Error saving output to file %s: %v", outputFilePath, err)

		if logErr := p.auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "SaveOutputEnd",
			Status:     "Failure",
			DurationMs: &saveDurationMs,
			Inputs: map[string]interface{}{
				"output_path": outputFilePath,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to save output to file: %v", err),
				Type:    "FileIOError",
			},
			Message: "Failed to save output to file",
		}); logErr != nil {
			p.logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("error saving output to file %s: %w", outputFilePath, err)
	}

	// Log successful saving of output
	if logErr := p.auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "SaveOutputEnd",
		Status:     "Success",
		DurationMs: &saveDurationMs,
		Inputs: map[string]interface{}{
			"output_path": outputFilePath,
		},
		Outputs: map[string]interface{}{
			"content_length": len(content),
		},
		Message: "Successfully saved output to file",
	}); logErr != nil {
		p.logger.Error("Failed to write audit log: %v", logErr)
	}

	p.logger.Info("Output successfully generated and saved to %s", outputFilePath)
	return nil
}
