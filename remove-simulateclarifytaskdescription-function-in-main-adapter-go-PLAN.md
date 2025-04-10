# Remove simulateClarifyTaskDescription function in main_adapter.go

## Goal
The goal of this task is to remove the `simulateClarifyTaskDescription` function from internal/integration/main_adapter.go as part of the ongoing effort to remove all clarify-related functionality from the codebase.

## Implementation Approach
After examining the codebase, I've identified that this task will require several steps:

1. Remove the `simulateClarifyTaskDescription` function (lines 195-297) from main_adapter.go
2. Update any test files that may be using this function:
   - The TestTaskClarification test in integration_test.go (lines 376-434) depends on this function
   - Remove or update the test to work without the clarify functionality
3. Keep the test environment helper methods (like `SimulateUserInput`) that may be used by other tests
4. Run tests to ensure all changes are compatible

## Reasoning
This approach is most suitable because:

1. Systematically removing functions that are directly related to the clarify functionality follows the pattern established in previous tasks
2. Since the function is used in a test, we need to either remove or update the test to maintain test coverage for the rest of the application
3. The `SimulateUserInput` helper method is potentially used by other tests, so it should be preserved
4. The clean approach aligns with the overall goal of removing the clarify feature while maintaining code quality and test coverage
5. This keeps the codebase clean by removing dead code that's no longer needed since the clarify functionality is being eliminated