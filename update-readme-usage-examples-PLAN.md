# Plan: Update README.md Usage Examples

## Task Title
Update README.md usage examples

## Goal
Ensure all command examples in README.md work with the refactored code, providing clear and accurate guidance for users of the architect tool.

## Chosen Approach
After analyzing the codebase and considering multiple approaches, I've chosen to implement a comprehensive review and update of the README.md usage examples with a focus on preserving the current documentation structure while ensuring all command examples accurately reflect the refactored architecture.

### Implementation Steps:

1. **Review existing README.md structure and examples**
   - Identify all command examples in the README.md file
   - Document the current organization of usage examples

2. **Verify example validity against refactored codebase**
   - Test each example command against the current codebase
   - Identify any examples that no longer work correctly

3. **Review flag structure in cmd/architect/cli.go**
   - Ensure all examples use the correct command-line flags
   - Check for any new flags added during refactoring
   - Verify deprecated flags are properly marked

4. **Update examples to reflect refactored structure**
   - Modify examples that don't work with the new codebase
   - Ensure consistent use of the `--task-file` flag
   - Remove or update examples using deprecated `--task` flag
   - Add any new examples needed to showcase new functionality

5. **Maintain documentation consistency**
   - Preserve the clear explanation of the task file approach
   - Ensure all configuration options accurately reflect current code
   - Keep the structure and organization of the README familiar to users

6. **Review integrated examples**
   - Ensure "Features" section aligns with refactored codebase
   - Update the "Configuration Options" table
   - Verify troubleshooting section contains accurate information

7. **Test updated examples**
   - Verify documentation matches actual behavior
   - Confirm all examples run successfully

### Reasoning for Chosen Approach:

This approach was chosen because it aligns best with our project standards and priorities:

1. **Simplicity and Clarity (CORE_PRINCIPLES.md)**
   - The approach preserves the existing documentation structure, making it familiar to users
   - It prioritizes accurate, straightforward examples over clever or complex ones
   - It maintains explicit descriptions for all commands and options

2. **Clean Separation of Concerns (ARCHITECTURE_GUIDELINES.md)**
   - The examples will reflect the new component architecture without exposing internal details
   - Usage examples will focus on functional behavior rather than implementation details
   - The approach respects the architectural boundaries between components

3. **Testability (TESTING_STRATEGY.md)**
   - This approach includes testing the examples to verify they work as documented
   - It embraces the behavior-focused testing philosophy by focusing on how users interact with the tool
   - It avoids examples that would be difficult to test or maintain

4. **Coding Conventions (CODING_STANDARDS.md)**
   - All code examples will follow project coding standards
   - Examples will use consistent naming and formatting
   - Command examples will reflect the project's standard command structure

5. **Documentation Approach (DOCUMENTATION_APPROACH.md)**
   - The plan maintains a user-focused perspective
   - It provides examples for different user experiences (from basic to advanced)
   - It maintains clarity in explanations and examples

### Trade-offs and Considerations:

- **Minimal Change vs. Complete Rewrite**: This approach prioritizes updating only what's necessary rather than a complete documentation rewrite, minimizing user disruption.
- **Backward Compatibility**: By clearly explaining the changes around `--task-file` vs `--task`, we help users transition smoothly.
- **Scope Management**: The focus remains solely on updating usage examples rather than expanding to broader documentation improvements.