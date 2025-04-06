# Implement Query Execution Helper

## Task Goal
Create the `_executeQueries` helper function that handles query execution to LLM providers with proper spinner updates and error handling, specifically using `ApiError` when appropriate.

## Implementation Approach

After analyzing the task and existing codebase patterns, I've chosen the following implementation approach:

### Approach: Stateless Wrapper with Error Categorization
I'll implement the `_executeQueries` function as a stateless wrapper around the existing `queryExecutor` module with enhanced error handling and spinner updates. This approach will:

1. Update the spinner with clear progress messages for each model being queried
2. Call the existing queryExecutor functionality to perform the actual queries
3. Implement robust error handling that categorizes and wraps errors as ApiError when appropriate
4. Return a properly structured result object that matches the interface defined in runThinktankTypes.ts

### Alternatives Considered

1. **Replacing the queryExecutor logic directly in the helper function**
   - Pros: Could potentially simplify the code by reducing layers of abstraction
   - Cons: Would duplicate logic, violate the single responsibility principle, and make future changes to the query execution logic harder to maintain
   - Rejected because: The existing queryExecutor module already has well-defined responsibilities and is likely tested independently

2. **Implementing parallel execution of queries with Promise.all**
   - Pros: Might improve performance by running all model queries in parallel 
   - Cons: Would make spinner updates more complex and less informative, as we wouldn't know which model is being processed at any given moment
   - Rejected because: Clear user feedback via the spinner is more important than potential performance gains for this feature

## Key Reasoning

I selected the wrapper approach because:

1. **Separation of concerns**: The query execution logic remains in the queryExecutor module, while the helper function focuses on workflow integration, spinner management, and error handling
2. **Consistent error handling**: Following the error handling contract defined in runThinktankTypes.ts ensures errors are properly categorized and wrapped
3. **User feedback**: Providing clear, model-specific progress updates via the spinner improves the user experience
4. **Maintainability**: This approach minimizes duplication and follows the established pattern seen in the other helper functions already implemented

The implementation will need to gracefully handle both successful responses and errors from any provider, ensuring that an error with one model doesn't prevent queries to other models from being attempted.