**Task: Update Config-Flag Mapping (`cmd/architect/cli.go`)**

## Goal
Update the `ConvertConfigToMap` function in `cmd/architect/cli.go` to reflect the removal of template-related flags and potential addition of the instructions flag, ensuring proper configuration merging in the application.

## Implementation Approach
1. Locate the `ConvertConfigToMap` function in `cmd/architect/cli.go`
2. Remove map entries for removed fields:
   - `promptTemplate`
   - `listExamples`
   - `showExample`
   - `taskFile`
3. Add a map entry for the new `instructionsFile` field if needed for configuration merging
4. Ensure proper type conversions are maintained for any remaining fields

## Reasoning
This approach directly addresses the task requirements by updating the configuration-to-map conversion function to match the new CLI flag structure. The change is straightforward: remove entries for flags that no longer exist and add an entry for the new instructions flag if needed.

I've chosen to specifically look for usage patterns of the map result in the codebase before deciding whether to add the `instructionsFile` key. If the map is used for config file merging and the instructions file path needs to be mergeable from config files, I'll add it. Otherwise, it may not be necessary.

This implementation maintains consistency with the previous flag refactoring tasks and prepares the application for proper handling of the new instruction-based design.