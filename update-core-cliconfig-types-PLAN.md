# Task: Update Core CliConfig Types

## Goal
Update the `CliConfig` struct in both `internal/architect/types.go` and `cmd/architect/cli.go` to reflect the new application design, removing fields related to templating and task files, and adding a field for instructions file path.

## Implementation Approach
I'll directly modify both `CliConfig` struct definitions to:
1. **Remove fields**:
   - In `internal/architect/types.go`: `TaskDescription`, `TaskFile`, `Template`, `ListExamples`, `ShowExample`
   - In `cmd/architect/cli.go`: `TaskFile`, `PromptTemplate`, `ListExamples`, `ShowExample`

2. **Add field**:
   - Add `InstructionsFile string` to both struct definitions
   
3. **Preserve remaining fields**:
   - All other fields not explicitly listed above will remain unchanged

This is a straightforward structural change to match the new design, removing fields related to the template system and replacing them with a single field for the instructions file.

## Reasoning
This direct, minimal approach is preferred because:
1. It follows the principle of making the smallest necessary change to achieve the goal
2. It directly maps to the requirements specified in the task
3. It avoids introducing any implementation details that might conflict with later tasks
4. By focusing only on the struct definitions without implementation logic, we minimize the risk of breaking existing functionality (implementation changes will be handled in subsequent tasks)

The changes strictly follow the design decisions from PLAN.md, particularly Section 2 (Type Definitions) and Section 4 (CLIConfig Structs).