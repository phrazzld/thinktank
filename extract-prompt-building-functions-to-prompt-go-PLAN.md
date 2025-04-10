# Extract Prompt Building Functions to prompt.go

## Task
Extract the prompt building functions (`buildPrompt`, `buildPromptWithConfig`, and `buildPromptWithManager`) from main.go to cmd/architect/prompt.go, maintaining the same functionality.

## Selected Approach
**Direct Port & Interface Adjustment**

This approach involves:
1. Creating an unexported helper method in prompt.go to handle the core prompt building logic
2. Adjusting the PromptBuilder interface to better reflect the required inputs
3. Implementing the interface methods by creating appropriate prompt managers and calling the helper
4. Updating main.go to use the new implementation

## Rationale
I've selected this approach because it:

1. **Enhances Testability**: 
   - The core logic in the helper method can be tested independently by mocking prompt.ManagerInterface
   - The public interface methods become simpler orchestrators that are easier to test
   - This aligns with the TESTING_PHILOSOPHY.md principles of designing for testability and testing behavior over implementation

2. **Improves Separation of Concerns**:
   - The prompt builder only needs specific inputs rather than the entire Configuration struct
   - The implementation is decoupled from main.go's configuration details
   - This aligns with the "Components should declare specific configuration dependencies" guideline

3. **Centralizes Core Logic**:
   - Places common template handling logic in a single helper method
   - Reduces duplication between BuildPrompt and BuildPromptWithConfig implementations
   - Simplifies future maintenance

4. **Creates Clear Interfaces**:
   - The modified interface better represents the actual requirements (task, context, template name)
   - Makes the expected inputs and behavior more explicit

## Implementation Steps

1. Adjust the PromptBuilder interface in cmd/architect/prompt.go:
   ```go
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
   ```

2. Create an internal helper method in the promptBuilder struct:
   ```go
   func (pb *promptBuilder) buildPromptInternal(task, context, templateName string, promptManager prompt.ManagerInterface) (string, error) {
       // Core logic from main.buildPromptWithManager
       data := &prompt.TemplateData{
           Task:    task,
           Context: context,
       }

       finalTemplateName := "default.tmpl"
       if templateName != "" {
           finalTemplateName = templateName
           pb.logger.Debug("Using custom prompt template: %s", finalTemplateName)
       }

       // Build the prompt
       generatedPrompt, err := promptManager.BuildPrompt(finalTemplateName, data)
       if err != nil {
           return "", fmt.Errorf("failed to build prompt: %w", err)
       }

       return generatedPrompt, nil
   }
   ```

3. Implement the BuildPrompt and BuildPromptWithConfig methods:
   ```go
   func (pb *promptBuilder) BuildPrompt(task, context, customTemplateName string) (string, error) {
       promptManager := prompt.NewManager(pb.logger)
       return pb.buildPromptInternal(task, context, customTemplateName, promptManager)
   }

   func (pb *promptBuilder) BuildPromptWithConfig(task, context, customTemplateName string, configManager config.ManagerInterface) (string, error) {
       promptManager, err := prompt.SetupPromptManagerWithConfig(pb.logger, configManager)
       if err != nil {
           return "", fmt.Errorf("failed to set up prompt manager: %w", err)
       }
       return pb.buildPromptInternal(task, context, customTemplateName, promptManager)
   }
   ```

4. Update main.go to use the new PromptBuilder implementation:
   - Create a promptBuilder instance
   - Replace calls to the original functions with calls to the PromptBuilder methods
   - Add transitional comments to the original functions

5. Implement tests for the new prompt.go functions