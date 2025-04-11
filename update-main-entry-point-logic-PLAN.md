**Task: Update Main Entry Point Logic (`cmd/architect/main.go`)**

## Goal
Modify the main entry point logic to align with our new instructions-based design by updating the `convertToArchitectConfig` function to properly map between the CLI and internal CliConfig structs, and removing the special commands functionality that was related to templating.

## Implementation Approach
1. Update the `convertToArchitectConfig` function in `cmd/architect/main.go` to:
   - Remove mapping of removed fields: `TaskDescription`, `TaskFile`, `Template`, `ListExamples`, `ShowExample` 
   - Add mapping for the new `InstructionsFile` field
   - Preserve mapping for all other remaining fields that still exist in both structs

2. Remove the call to `architect.HandleSpecialCommands` in the `Main` function since this functionality was related to template examples and is no longer needed.

3. Ensure proper error handling and logging remains intact during these changes.

## Reasoning
This approach directly addresses the task requirements while maintaining consistency with the previous refactoring steps. By modifying the `convertToArchitectConfig` function, we ensure that the CLI config values are properly passed to the internal architecture while reflecting our new design decisions.

Removing the call to `HandleSpecialCommands` is appropriate because that function was specifically designed to handle template-related commands like listing examples, which are being removed as part of our simplification effort.

The implementation will be clean and focused, touching only what needs to be changed while preserving the overall flow and error handling of the application. This keeps the changes minimal and reduces the risk of introducing bugs.