# Implement Error Handling Helper

## Task Goal
Create the `_handleWorkflowError` helper function that categorizes unknown errors, ensures proper ThinktankError types, logs contextual information, and rethrows for upstream handling.

## Implementation Approach

I'll implement a comprehensive error handling helper function that serves as the central error processing point for the workflow. This function will:

1. Take any caught error (`unknown` type) and convert it to a proper `ThinktankError` or its appropriate subclass based on error characteristics.

2. Use the existing error categorization utilities from `src/core/errors/utils/categorization.ts` to intelligently categorize errors that don't already have an assigned category.

3. Add contextual information to errors based on the current workflow state, making debugging and troubleshooting easier.

4. Update the spinner with appropriate failure messages using the error's formatted output.

5. Provide detailed, actionable error messages with suggestions tailored to the specific error context and category.

The function will maintain the existing error hierarchy and categorization system, ensuring that errors are consistently handled and displayed to users.

### Key Components:
1. **Type-Based Error Processing**: Handle errors differently based on their existing type (ThinktankError subclasses vs. generic Error vs. non-Error objects)
2. **Error Categorization**: Leverage the existing categorization utilities for unknown errors
3. **Spinner Updates**: Update the CLI spinner with appropriate failure messages
4. **Context-Aware Suggestions**: Generate helpful suggestions based on the error type and workflow state
5. **Error Propagation**: Rethrow properly formatted errors for upstream handling

## Alternatives Considered

1. **Simple Error Wrapper**: Instead of sophisticated categorization, simply wrap all errors in a generic ThinktankError. This would be simpler but would lose valuable error type information and context.

2. **Distributed Error Handling**: Keep the error handling logic in each specific workflow step rather than centralizing it. This would avoid needing to pass the workflow state but would lead to duplicated error handling code.

3. **Custom Error Logger**: Create a separate error logging system instead of focusing on error transformation. This would be useful for debugging but wouldn't improve the user experience with better error messages.

## Reasoning for Selected Approach

I've chosen to implement a comprehensive central error handler because:

1. **Centralized Logic**: This approach centralizes the error handling logic in one place, making it easier to maintain and update. The current runThinktank function has a large error handling section that can be refactored into this helper.

2. **Consistent Error Types**: By ensuring all errors are transformed into the appropriate ThinktankError subclass, we maintain a consistent error handling approach throughout the application.

3. **Rich Error Context**: By having access to the full workflow state, we can add rich contextual information to errors, making them more helpful to users and easier to debug.

4. **Reusable Logic**: The categorization and error processing logic can be reused anywhere in the application, not just in the runThinktank workflow.

5. **Improved User Experience**: This approach ensures that users see consistent, helpful error messages with actionable suggestions, regardless of where in the workflow the error occurred.

This implementation will align with the existing error handling patterns in the codebase while improving modularity and reusability.