**Task: Update CLI Tests (`cmd/architect/cli_test.go`)**

## Goal
Update the CLI tests to align with the new instructions-based design, ensuring that the tests verify the proper handling of the `--instructions` flag, validate the input, interpret the usage message, and confirm the removal of template-related flags and features.

## Implementation Approach
1. First, review the current test file to understand existing test coverage and patterns.

2. Update test cases for flag parsing:
   - Replace tests checking `--task-file`, `--prompt-template`, `--list-examples`, `--show-example` with tests for the new `--instructions` flag.
   - Ensure tests verify correct default behavior and custom values for the remaining flags.
   - Verify that CLI config is properly populated with the correct `InstructionsFile` value.

3. Update validation tests:
   - Replace tests verifying `TaskFile` validation with tests for `InstructionsFile`.
   - Ensure tests verify the error handling for missing required flags.
   - Confirm that the validation logic correctly handles the dry-run exception.

4. Update usage message tests:
   - Update tests to verify the new usage pattern.
   - Ensure tests confirm the usage examples have been updated.
   - Verify the correct help text for the `--instructions` flag.

5. Check environment variable handling:
   - Ensure tests for environment variable fallbacks are still working properly.

6. Clean up any references to removed functionality:
   - Remove test mocks for template-related features.
   - Remove test fixtures related to templates.

## Reasoning
This approach ensures comprehensive testing of the CLI interface after the refactoring changes. By following the same structure as the existing tests, we maintain consistency in the test suite while updating it to reflect the new design.

The tests must verify not only the presence of the new features but also the absence of removed functionality. This ensures that the refactoring is complete and working as expected.

Since tests are a form of documentation, clear and well-structured tests will help future developers understand how the CLI is supposed to work with the new instructions-based approach.