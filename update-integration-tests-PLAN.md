# Update Integration Tests

## Goal
Refactor the existing integration tests to work with the new helper function structure of the runThinktank workflow.

## Implementation Approach
The implementation will focus on updating the `runThinktank.test.ts` file to mock the new helper functions and verify that the orchestration logic works correctly. The approach will:

1. Update imports to include the new helper functions
2. Replace direct mocking of internal implementation details with mocks of the new helper functions
3. Verify that the main runThinktank function correctly orchestrates the helper functions in sequence
4. Test that errors from each helper function are properly propagated through the workflow

## Reasoning
This approach aligns with the task requirements and maintains test isolation. By mocking the helper functions rather than their internal implementations, we ensure that the tests verify the orchestration logic without being tightly coupled to implementation details. This will make the tests more resilient to future changes in the helper functions while still providing confidence that the workflow behaves correctly as a whole.