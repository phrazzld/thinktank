// Package architect provides the command-line interface for the architect tool
package architect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// PromptBuilder defines the interface for prompt building and template handling
type PromptBuilder interface {
	// ReadTaskFromFile reads task description from a file
	ReadTaskFromFile(taskFilePath string) (string, error)

	// BuildPrompt constructs the prompt string using a basic prompt manager
	BuildPrompt(task, context, customTemplateName string) (string, error)

	// BuildPromptWithConfig constructs the prompt string using the configuration system
	BuildPromptWithConfig(task, context, customTemplateName string, configManager config.ManagerInterface) (string, error)

	// ListExampleTemplates displays a list of available example templates
	ListExampleTemplates(configManager config.ManagerInterface) error

	// ShowExampleTemplate displays the content of a specific example template
	ShowExampleTemplate(name string, configManager config.ManagerInterface) error
}

// promptBuilder implements the PromptBuilder interface
type promptBuilder struct {
	logger logutil.LoggerInterface
}

// NewPromptBuilder creates a new PromptBuilder instance
func NewPromptBuilder(logger logutil.LoggerInterface) PromptBuilder {
	return &promptBuilder{
		logger: logger,
	}
}

// ReadTaskFromFile reads task description from a file
func (pb *promptBuilder) ReadTaskFromFile(taskFilePath string) (string, error) {
	// Check if path is absolute, if not make it absolute
	if !filepath.IsAbs(taskFilePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("error getting current working directory: %w", err)
		}
		taskFilePath = filepath.Join(cwd, taskFilePath)
	}

	// Enhanced file existence check with specific errors
	fileInfo, err := os.Stat(taskFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("task file not found: %s", taskFilePath)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("task file permission denied: %s", taskFilePath)
		}
		// Generic stat error
		return "", fmt.Errorf("error checking task file status: %w", err)
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return "", fmt.Errorf("task file path is a directory: %s", taskFilePath)
	}

	// Read file content
	content, err := os.ReadFile(taskFilePath)
	if err != nil {
		if os.IsPermission(err) {
			return "", fmt.Errorf("task file permission denied: %s", taskFilePath)
		}
		// Generic read error
		return "", fmt.Errorf("error reading task file content: %w", err)
	}

	// Check for empty content
	if len(strings.TrimSpace(string(content))) == 0 {
		return "", fmt.Errorf("task file is empty: %s", taskFilePath)
	}

	// Return content as string
	return string(content), nil
}

// buildPromptInternal is a helper method that handles the core prompt building logic
func (pb *promptBuilder) buildPromptInternal(task, context, templateName string, promptManager prompt.ManagerInterface) (string, error) {
	// Create template data
	data := &prompt.TemplateData{
		Task:    task,
		Context: context, // context already has the <context> tags from fileutil
	}

	// Determine which template to use
	finalTemplateName := "default.tmpl"
	if templateName != "" {
		finalTemplateName = templateName
		pb.logger.Debug("Using custom prompt template: %s", finalTemplateName)
	}

	// Build the prompt (template loading is handled by the manager)
	generatedPrompt, err := promptManager.BuildPrompt(finalTemplateName, data)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	return generatedPrompt, nil
}

// BuildPrompt constructs the prompt string using a basic prompt manager
func (pb *promptBuilder) BuildPrompt(task, context, customTemplateName string) (string, error) {
	// Create a basic prompt manager
	promptManager := prompt.NewManager(pb.logger)

	// Use the internal helper to build the prompt
	return pb.buildPromptInternal(task, context, customTemplateName, promptManager)
}

// BuildPromptWithConfig constructs the prompt string using the configuration system
func (pb *promptBuilder) BuildPromptWithConfig(task, context, customTemplateName string, configManager config.ManagerInterface) (string, error) {
	// Create a prompt manager with config support
	promptManager, err := prompt.SetupPromptManagerWithConfig(pb.logger, configManager)
	if err != nil {
		return "", fmt.Errorf("failed to set up prompt manager: %w", err)
	}

	// Use the internal helper to build the prompt
	return pb.buildPromptInternal(task, context, customTemplateName, promptManager)
}

// ListExampleTemplates displays a list of available example templates
func (pb *promptBuilder) ListExampleTemplates(configManager config.ManagerInterface) error {
	// Create prompt manager
	promptManager, err := prompt.SetupPromptManagerWithConfig(pb.logger, configManager)
	if err != nil {
		// Fall back to basic manager if config-based setup fails
		promptManager = prompt.NewManager(pb.logger)
	}

	// Get the list of examples
	examples, err := promptManager.ListExampleTemplates()
	if err != nil {
		return fmt.Errorf("error listing example templates: %w", err)
	}

	// Display the examples
	fmt.Println("Available Example Templates:")
	fmt.Println("---------------------------")
	if len(examples) == 0 {
		fmt.Println("No example templates found.")
	} else {
		for i, example := range examples {
			fmt.Printf("%d. %s\n", i+1, example)
		}
		fmt.Println("\nTo view an example template, use --show-example <template-name>")
		fmt.Println("Example: architect --show-example basic.tmpl")
	}

	return nil
}

// ShowExampleTemplate displays the content of a specific example template
func (pb *promptBuilder) ShowExampleTemplate(name string, configManager config.ManagerInterface) error {
	// Create prompt manager
	promptManager, err := prompt.SetupPromptManagerWithConfig(pb.logger, configManager)
	if err != nil {
		// Fall back to basic manager if config-based setup fails
		promptManager = prompt.NewManager(pb.logger)
	}

	// Get the template content
	content, err := promptManager.GetExampleTemplate(name)
	if err != nil {
		return fmt.Errorf("error: %w\nUse --list-examples to see available example templates", err)
	}

	// Print the content to stdout (allowing for redirection to a file)
	fmt.Print(content)

	return nil
}