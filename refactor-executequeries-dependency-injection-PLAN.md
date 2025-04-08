# Refactoring _executeQueries to Use Dependency Injection: Implementation Plan

## Goal
Modify the `_executeQueries` function to accept and use the `LLMClient` interface instead of making direct API calls, to improve testability and follow the established pattern of dependency injection.

## Chosen Approach: Direct Parameter Injection

After analyzing the three implementation approaches (Direct Parameter Injection, Dependency Object Injection, and Class-Based Workflow Step), I've chosen the **Direct Parameter Injection** approach as the most suitable for this refactoring task.

### Approach Details
1. Modify the `_executeQueries` function signature to accept an `LLMClient` instance as a parameter
2. Update the `ExecuteQueriesParams` interface to include the `llmClient` parameter
3. Replace the call to `executeQueries` with direct use of the injected `llmClient` instance
4. Update the error handling to work with the new approach
5. Update `runThinktank.ts` to instantiate and pass the `ConcreteLLMClient` when calling `_executeQueries`

### Implementation Specifics
- The function will no longer call `executeQueries` from `queryExecutor.ts`, which currently handles model selection and provider resolution
- Instead, it will loop through models and directly call `llmClient.generate` for each model
- All existing functionality related to status tracking, parallel execution, spinner updates, and error handling will be maintained
- The function will create the same structure of returned data as the current implementation

### Testability Considerations
This approach aligns extremely well with the project's testing philosophy for the following reasons:

1. **Behavior Over Implementation**: By injecting the `LLMClient` interface, tests can focus on verifying that `_executeQueries` calls the client correctly without knowing implementation details of the API interactions.

2. **Minimize Mocking**: Tests will only need to mock a single, direct dependency - the `LLMClient` interface. This adheres to the principle of "Mock External Boundaries" by mocking only the interface that represents external system boundaries (network I/O).

3. **Refactor Signal**: The current implementation makes direct use of `executeQueries` which itself calls the registry and providers. This creates a complex dependency chain for testing. By injecting the `LLMClient`, we're addressing a potential "refactor signal" by reducing coupling.

4. **Explicit Dependency**: The function signature will clearly show that `_executeQueries` requires an `LLMClient`, making dependencies explicit and clear.

### Why This Approach?
- **Simplicity**: Most straightforward modification with minimal structural changes
- **Consistency**: Follows the same pattern used for the other interfaces in the project
- **Backward Compatibility**: Maintains the same function return structure and error handling approach
- **Test Clarity**: Creates a clean dependency injection point that makes mocking straightforward

## Implementation Details

1. Update the `ExecuteQueriesParams` interface in `runThinktankTypes.ts` to include the `llmClient` parameter

2. Modify the `_executeQueries` function in `runThinktankHelpers.ts` to:
   - Accept the `llmClient` parameter
   - Replace the call to `executeQueries` with direct calls to `llmClient.generate`
   - Maintain the same status tracking and parallel execution logic
   - Preserve the existing error handling patterns

3. Update `runThinktank.ts` to:
   - Import and instantiate the `ConcreteLLMClient`
   - Pass it to `_executeQueries` when calling the function

4. Update the tests in `executeQueriesHelper.test.ts` to:
   - Mock the `LLMClient` interface instead of `queryExecutor.executeQueries`
   - Verify the correct parameters are passed to `llmClient.generate`
   - Test the error handling with the new approach
