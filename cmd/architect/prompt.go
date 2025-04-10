// Package architect provides the command-line interface for the architect tool
package architect

import (
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// PromptBuilder defines the interface for prompt building and template handling
type PromptBuilder interface {
	// ReadTaskFromFile reads task description from a file
	ReadTaskFromFile(taskFilePath string) (string, error)

	// BuildPrompt constructs the prompt string for the Gemini API
	BuildPrompt(taskDescription, context, promptTemplate string, promptManager prompt.ManagerInterface) (string, error)

	// BuildPromptWithConfig constructs the prompt string using the configuration system
	BuildPromptWithConfig(taskDescription, context, promptTemplate string, configManager config.ManagerInterface) (string, error)

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
	// Stub implementation - will be replaced with actual code from main.go
	return "", fmt.Errorf("not implemented yet")
}

// BuildPrompt constructs the prompt string for the Gemini API
func (pb *promptBuilder) BuildPrompt(taskDescription, context, promptTemplate string, promptManager prompt.ManagerInterface) (string, error) {
	// Stub implementation - will be replaced with actual code from main.go
	return "", fmt.Errorf("not implemented yet")
}

// BuildPromptWithConfig constructs the prompt string using the configuration system
func (pb *promptBuilder) BuildPromptWithConfig(taskDescription, context, promptTemplate string, configManager config.ManagerInterface) (string, error) {
	// Stub implementation - will be replaced with actual code from main.go
	return "", fmt.Errorf("not implemented yet")
}

// ListExampleTemplates displays a list of available example templates
func (pb *promptBuilder) ListExampleTemplates(configManager config.ManagerInterface) error {
	// Stub implementation - will be replaced with actual code from main.go
	return fmt.Errorf("not implemented yet")
}

// ShowExampleTemplate displays the content of a specific example template
func (pb *promptBuilder) ShowExampleTemplate(name string, configManager config.ManagerInterface) error {
	// Stub implementation - will be replaced with actual code from main.go
	return fmt.Errorf("not implemented yet")
}
