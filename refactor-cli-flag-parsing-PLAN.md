# Task: Refactor CLI Flag Parsing (`cmd/architect/cli.go`)

## Goal
Update the `ParseFlagsWithEnv` function in `cmd/architect/cli.go` to replace the old templating and task-based flags with the new instructions-based approach. This involves removing flags for template/task functionality and adding a new required `--instructions` flag.

## Implementation Approach
I'll modify the `ParseFlagsWithEnv` function and related parts of the CLI functionality with the following changes:

1. **Flag Definition Changes:**
   - Remove the flag definitions for: `--task-file`, `--prompt-template`, `--list-examples`, and `--show-example`
   - Add a new required string flag `--instructions` that accepts a file path to the instructions file

2. **Flag Parse Logic:**
   - Remove the code that assigns values from the removed flags to the config object
   - Add code to assign the value from the new `--instructions` flag to `config.InstructionsFile`
   - Update the basic validation in `ParseFlagsWithEnv` to check for the required `--instructions` flag instead of `--task-file`

3. **Development Strategy:**
   - Make these changes in a systematic way, starting with flag definitions, then updating assignment logic
   - Maintain compatibility with other flags that aren't changing
   - Ensure error messages are clear and helpful

This implementation focuses only on the flag parsing logic in `ParseFlagsWithEnv`. Subsequent tasks will handle updating the validation function and usage message.

## Reasoning
I've chosen this approach because:

1. It aligns with the single responsibility of the task (updating flag parsing logic)
2. It's a focused change that only modifies what's necessary for the flag parsing portion of the refactoring
3. It keeps the implementation clean by only handling the flag definitions and their mapping to the config struct
4. It maintains compatibility with the rest of the CLI functionality that isn't being changed yet
5. This approach allows for incremental validation (can test that flag parsing works correctly before moving on to validation logic changes)

The implementation will directly follow the requirements in PLAN.md, specifically Section 3, Task 1, which outlines the CLI flag refactoring steps.