// Package architect contains the core application logic for the architect tool
package architect

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// OutputWriter defines the interface for file output and plan writing
type OutputWriter interface {
	// SaveToFile writes the generated plan to the specified file
	SaveToFile(content, outputFile string) error

	// GenerateAndSavePlan creates and saves the plan to a file
	GenerateAndSavePlan(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, promptManager prompt.ManagerInterface) error

	// GenerateAndSavePlanWithConfig creates and saves the plan to a file using the config system
	GenerateAndSavePlanWithConfig(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, configManager config.ManagerInterface) error
}

// outputWriter implements the OutputWriter interface
type outputWriter struct {
	logger       logutil.LoggerInterface
	tokenManager TokenManager
	apiService   APIService
}

// NewOutputWriter creates a new OutputWriter instance
func NewOutputWriter(logger logutil.LoggerInterface, tokenManager TokenManager) OutputWriter {
	return &outputWriter{
		logger:       logger,
		tokenManager: tokenManager,
		apiService:   NewAPIService(logger),
	}
}

// SaveToFile writes the generated plan to the specified file
func (ow *outputWriter) SaveToFile(content, outputFile string) error {
	// Ensure output path is absolute
	outputPath := outputFile
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			ow.logger.Error("Error getting current working directory: %v", err)
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		ow.logger.Error("Error creating output directory %s: %v", outputDir, err)
		return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
	}

	// Write to file
	ow.logger.Info("Writing plan to %s...", outputPath)
	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		ow.logger.Error("Error writing plan to file %s: %v", outputPath, err)
		return fmt.Errorf("error writing plan to file %s: %w", outputPath, err)
	}

	ow.logger.Info("Successfully generated plan and saved to %s", outputPath)
	return nil
}

// GenerateAndSavePlan creates and saves the plan to a file
func (ow *outputWriter) GenerateAndSavePlan(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, promptManager prompt.ManagerInterface) error {
	// Create template data
	data := &prompt.TemplateData{
		Task:    taskDescription,
		Context: projectContext,
	}

	// First check if task file content is a template itself
	var generatedPrompt string
	var err error

	if prompt.IsTemplate(taskDescription) {
		// This is a template in the task file - process it directly
		ow.logger.Info("Task file contains template variables, processing as template...")
		ow.logger.Debug("Processing task file as a template")

		// Create a template from the task file content
		tmpl, err := template.New("task_file_template").Parse(taskDescription)
		if err != nil {
			ow.logger.Error("Failed to parse task file as template: %v", err)
			return fmt.Errorf("failed to parse task file as template: %w", err)
		}

		// Execute the template with the context data
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		if err != nil {
			ow.logger.Error("Failed to execute task file template: %v", err)
			return fmt.Errorf("failed to execute task file template: %w", err)
		}

		generatedPrompt = buf.String()
		ow.logger.Info("Task file template processed successfully")
	} else {
		// Standard approach - use the prompt manager
		ow.logger.Info("Building prompt template...")
		ow.logger.Debug("Building prompt template...")

		// Default template name
		templateName := "default.tmpl"

		// Build the prompt (template loading handled by manager)
		generatedPrompt, err = promptManager.BuildPrompt(templateName, data)
		if err != nil {
			ow.logger.Error("Failed to build prompt: %v", err)
			return fmt.Errorf("failed to build prompt: %w", err)
		}
		ow.logger.Info("Prompt template built successfully")
	}

	// Debug logging of prompt details
	ow.logger.Debug("Prompt length: %d characters", len(generatedPrompt))
	ow.logger.Debug("Sending task to Gemini: %s", taskDescription)

	// Get token count for confirmation and limit checking
	ow.logger.Info("Checking token limits...")
	ow.logger.Debug("Checking token limits...")

	// Get token info
	tokenInfo, err := ow.tokenManager.GetTokenInfo(ctx, client, generatedPrompt)
	if err != nil {
		ow.logger.Error("Token count check failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			ow.logger.Error("Token count check failed: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				ow.logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			ow.logger.Debug("Error details: %s", apiErr.DebugInfo())
		} else {
			ow.logger.Error("Token count check failed: %v", err)
			ow.logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		return fmt.Errorf("token count check failed: %w", err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		ow.logger.Error("Token limit exceeded")
		ow.logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		ow.logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		return fmt.Errorf("token limit exceeded: %s", tokenInfo.LimitError)
	}

	ow.logger.Info("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// Log token usage for regular (non-debug) mode
	ow.logger.Info("Token usage: %d / %d (%.1f%%)",
		tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// Call Gemini API
	ow.logger.Info("Generating plan...")
	ow.logger.Debug("Generating plan...")
	result, err := client.GenerateContent(ctx, generatedPrompt)
	if err != nil {
		ow.logger.Error("Generation failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			ow.logger.Error("Error generating content: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				ow.logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			ow.logger.Debug("Error details: %s", apiErr.DebugInfo())
		} else {
			ow.logger.Error("Error generating content: %v", err)
		}

		return fmt.Errorf("plan generation failed: %w", err)
	}

	// Process API response
	generatedPlan, err := ow.apiService.ProcessResponse(result)
	if err != nil {
		// Get detailed error information
		errorDetails := ow.apiService.GetErrorDetails(err)

		// Provide specific error messages based on error type
		if ow.apiService.IsEmptyResponseError(err) {
			ow.logger.Error("Received empty or invalid response from Gemini API")
			ow.logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response due to empty content: %w", err)
		} else if ow.apiService.IsSafetyBlockedError(err) {
			ow.logger.Error("Content was blocked by Gemini safety filters")
			ow.logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response due to safety restrictions: %w", err)
		} else {
			// Generic API error handling
			return fmt.Errorf("failed to process API response: %w", err)
		}
	}
	ow.logger.Info("Plan generated successfully")

	// Debug logging of results
	ow.logger.Debug("Plan received from Gemini.")
	if result.TokenCount > 0 {
		ow.logger.Debug("Token usage: %d tokens", result.TokenCount)
	}

	// Write the plan to file
	ow.logger.Info("Writing plan to %s...", outputFile)
	ow.logger.Debug("Writing plan to %s...", outputFile)

	// Save the content to the file
	err = ow.SaveToFile(generatedPlan, outputFile)
	if err != nil {
		return fmt.Errorf("error saving plan to file: %w", err)
	}

	ow.logger.Info("Plan saved to %s", outputFile)
	return nil
}

// SetupPromptManagerWithConfig is exported for testing
var SetupPromptManagerWithConfig = prompt.SetupPromptManagerWithConfig

// GenerateAndSavePlanWithConfig creates and saves the plan to a file using the config system
func (ow *outputWriter) GenerateAndSavePlanWithConfig(ctx context.Context, client gemini.Client, taskDescription, projectContext, outputFile string, configManager config.ManagerInterface) error {
	// Set up a prompt manager with config support
	promptManager, err := SetupPromptManagerWithConfig(ow.logger, configManager)
	if err != nil {
		ow.logger.Error("Failed to set up prompt manager: %v", err)

		// Create a fallback prompt manager without config
		fallbackManager := prompt.NewManager(ow.logger)
		ow.logger.Info("Falling back to default prompt manager")

		// Use the fallback manager instead
		return ow.GenerateAndSavePlan(ctx, client, taskDescription, projectContext, outputFile, fallbackManager)
	}

	// Use the config-based prompt manager
	return ow.GenerateAndSavePlan(ctx, client, taskDescription, projectContext, outputFile, promptManager)
}
