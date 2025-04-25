# T003 - Fail fast on invalid `--synthesis-model` - Implementation Plan

## Task
Add validation to immediately fail when an invalid `--synthesis-model` is provided, rather than failing later in the execution process.

## Context
Currently, the CLI accepts any value for the `--synthesis-model` flag and only fails later when it tries to use it. We need to add upfront validation to fail fast with a clear error message.

## Implementation Steps

1. First, understand the current validation flow and where we need to add the check
2. Examine how synthesis models are currently validated/loaded
3. Modify the validation function to check if the synthesis model is valid
4. Add appropriate error message for invalid synthesis models
5. Update or add tests to verify the validation works correctly

## Relevant files to examine:
- cmd/thinktank/cli.go (for CLI argument handling)
- cmd/thinktank/cli_validation_test.go (for existing validation tests)
- internal/registry files (for model validation)
