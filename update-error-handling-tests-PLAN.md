# Update Error Handling Tests

## Goal
Update the error handling tests to verify that each helper function properly catches, wraps, and propagates errors, and that the main error handling catch block in the runThinktank function correctly handles all error scenarios.

## Implementation Approach
The implementation will focus on updating the `runThinktank-error-handling.test.ts` file to work with the new helper function structure. The approach will:

1. Refactor the existing tests to use mocks of the new helper functions instead of mocking internal implementation details
2. Verify that each helper function properly catches, wraps, and propagates errors of different types
3. Test the integration between helper functions and the main error handling catch block
4. Ensure that error objects maintain their specific error types (e.g., ConfigError, ApiError, etc.) when propagated through the workflow
5. Test that the _handleWorkflowError helper function correctly categorizes unknown errors

## Reasoning
This approach aligns with the new architecture that separates concerns into distinct helper functions. By mocking the helper functions to throw specific errors, we can verify that the error handling contract is being followed throughout the workflow. This will ensure that users receive appropriate error messages and suggestions regardless of where in the workflow an error occurs.

The tests will focus on ensuring that:
1. Specific error types (ConfigError, ApiError, FileSystemError, etc.) are maintained when propagated
2. Error objects contain appropriate suggestions based on the error context
3. The workflow state is correctly passed to the error handler when an error occurs
4. User-facing error messages are clear and actionable

This will provide confidence that the refactored workflow handles errors consistently and provides helpful feedback to users.