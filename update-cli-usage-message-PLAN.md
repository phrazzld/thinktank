# Task: Update CLI Usage Message (`cmd/architect/cli.go`)

## Goal
Update the usage message in the `flagSet.Usage` function to accurately reflect the new command structure using the `--instructions` flag instead of the old task/template flags, and to remove documentation for the deleted flags.

## Implementation Approach
I'll modify the `flagSet.Usage` function in `cmd/architect/cli.go` to:

1. **Update the Usage Header:**
   - Replace the current pattern `[--list-examples | --show-example NAME | --task-file <path> [options] <path1> [path2...]]` with the new pattern `--instructions <file> [options] <path1> [path2...]`
   - This makes it clear that `--instructions` is required and that it needs a file path

2. **Remove Example Commands Related to Templates:**
   - Remove the examples for `--list-examples` and `--show-example` since these flags no longer exist
   - Add a new example showing how to use the `--instructions` flag with context paths

3. **Keep Other Parts Unchanged:**
   - Maintain the explanation of positional arguments (`<path1> [path2...]`)
   - Keep the options section that uses `flagSet.PrintDefaults()` (this will automatically reflect our flag changes)
   - Keep the environment variables section unchanged

The updated usage message will make it clear that the primary command structure is now `architect --instructions <file> [options] <path1> [path2...]` without any references to the removed template functionality.

## Reasoning
I chose this straightforward approach because:

1. **Clear Communication:** The updated usage message will clearly communicate the new command structure to users
2. **Consistency:** The updated message will be consistent with the actual flag changes we've already made
3. **Simplicity:** This approach focuses on the minimal necessary changes to the usage message
4. **User Experience:** By giving clear examples with the new flag structure, we help users transition to the new syntax

This implementation directly addresses the requirements in the PLAN.md (Section 3, Task 1), specifically the bullet point "Update `flagSet.Usage` to reflect new flags and arguments."