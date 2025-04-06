# Simplify Error Handling Logic

## Goal
Refactor the complex error handling in `_handleWorkflowError` to be more maintainable by creating or enhancing utility functions for error categorization.

## Implementation Approach
After analyzing the current implementation, I've chosen to refactor the error handling logic with the following approach:

1. Create a new `createContextualError` utility function in the `categorization.ts` file that will:
   - Take an error object and context information
   - Use the existing `categorizeError` function to determine the error type
   - Create appropriate error instances with contextual suggestions
   - Handle special cases like file not found, permission errors, etc.

2. Modify the `_handleWorkflowError` function to:
   - Extract context information from workflow state
   - Call the new utility function
   - Apply any workflow-specific suggestions
   - Update the spinner and throw the error

3. Update failing tests to match the new behavior

## Rationale
I've chosen this approach for several reasons:

1. **Improved Separation of Concerns**: By moving error categorization logic to a dedicated utility function, we make both the utility and the `_handleWorkflowError` function simpler and more focused.

2. **Centralized Logic**: The new utility function centralizes the pattern matching and error creation logic that is currently duplicated across multiple helper functions.

3. **Reusability**: This approach makes the error categorization logic reusable across the codebase, not just in `_handleWorkflowError`.

4. **Testability**: Separating the error creation from the workflow context makes it easier to unit test both components independently.

5. **Maintainability**: The resulting code will be less complex and easier to modify when new error patterns need to be added, reducing the likelihood of errors.

Alternative approaches I considered but rejected:

1. **Simple Refactoring Within the Function**: Just restructuring the code within `_handleWorkflowError` would help but wouldn't address the root issue of complex, hard-to-maintain error categorization logic.

2. **Moving Logic to Error Classes**: Putting the categorization logic in the error classes themselves would create tight coupling between errors and their detection logic, making it harder to maintain and extend.

3. **Using a Pattern Matching Library**: Introducing a dependency for pattern matching would add complexity for limited gain, as the current regex-based approach is adequate when properly structured.