# Task: Update CLI Input Validation (`cmd/architect/cli.go`)

## Goal
Modify the `ValidateInputs` function to validate the new command-line interface structure - removing the validation for template-related functionality and adding validation for the new `--instructions` flag.

## Implementation Approach
I'll update the `ValidateInputs` function in `cmd/architect/cli.go` to:

1. **Remove Template-Related Validation:**
   - Remove the check for example commands (the condition `config.ListExamples || config.ShowExample != ""` and its bypass logic)
   - This is no longer needed since those flags have been removed from the CLI configuration

2. **Update Required Flag Check:**
   - Replace the check for `config.TaskFile == ""` with a check for `config.InstructionsFile == ""`
   - Keep the dry-run exception (`&& !config.DryRun`) so that dry runs can still work without instructions
   - Update the error message to mention the new `--instructions` flag instead of `--task-file`

3. **Keep Other Validations:**
   - Maintain the validation for `config.Paths` (at least one path is required)
   - Maintain the validation for `config.ApiKey` (API key must be set)

The implementation will be straightforward, replacing the old validation checks with the new ones while preserving the existing validation pattern.

## Reasoning
I've chosen this direct approach because:

1. **Minimal Change Principle:** It follows the pattern of making the smallest necessary changes to accommodate the new CLI interface
2. **Consistency:** The validation maintains the same structure and error handling patterns as before
3. **User Experience:** The updated error messages will provide clear guidance on the new required flags
4. **Dry Run Support:** Preserving the dry-run exception maintains the existing functionality where users can run dry-runs without specifying a task file

This approach directly implements the requirements from PLAN.md (Section 3, Task 1), which specifies "Update `ValidateInputs`: Remove check for `TaskFile`. Add check for required `--instructions` flag. Remove validation bypass for example commands."