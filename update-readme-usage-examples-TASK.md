# Task: Update README.md Usage Examples

## Goal
Ensure all command examples in README.md work with the refactored code, providing clear and accurate guidance for users of the architect tool.

## Context
The architect codebase has undergone significant refactoring, moving functionality from the main.go file into dedicated component files (token.go, api.go, context.go, prompt.go, output.go) in the cmd/architect/ package. A new main.go entry point has been created that calls the cmd/architect/main.go's Main function. This task requires updating the README.md to reflect any changes in the command structure or usage patterns resulting from this refactoring.

## Requirements
1. Review the current README.md usage examples
2. Verify each example against the refactored codebase
3. Update examples that no longer work correctly
4. Ensure examples follow the new command structure
5. Maintain consistency with project standards and documentation approach

## Project Standards Considerations
When implementing this task, please consider the following project standards:

### Core Principles (CORE_PRINCIPLES.md)
- Maintain simplicity in examples
- Be explicit rather than implicit in usage instructions
- Keep a modular approach to documentation sections

### Architectural Guidelines (ARCHITECTURE_GUIDELINES.md)
- Ensure examples reflect the new component architecture
- Properly represent the interactions between components
- Maintain clean separation of concerns in examples

### Coding Standards (CODING_STANDARDS.md)
- Follow consistent formatting for code examples
- Use proper naming conventions in examples
- Show idiomatic Go usage that reflects the codebase

### Testing Strategy (TESTING_STRATEGY.md)
- Include examples of how to test the application if relevant
- Demonstrate proper usage that will lead to testable implementations

### Documentation Approach (DOCUMENTATION_APPROACH.md)
- Focus on user workflows and outcomes
- Provide clear, concise explanations with examples
- Structure documentation for different user experiences (novice to expert)

## Requested Implementation Approaches
Please provide 2-3 different approaches to updating the README.md usage examples, with pros and cons for each. For each approach, explain:

1. How you would analyze the current examples and identify needed changes
2. How you would structure the updated examples
3. How you would verify the examples work with the refactored code
4. How you would maintain consistency with the project's documentation style

Finally, recommend the best approach based on alignment with the project's standards and practices.