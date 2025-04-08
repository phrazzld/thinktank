# Create LLMClient Interface Implementation - Plan

## Task Title
Create LLMClient interface implementation

## Goal
Implement a concrete class that implements the `LLMClient` interface defined in `src/core/interfaces.ts`, which wraps the existing provider logic in a way that promotes dependency injection and testability.

## Chosen Approach: Thin Wrapper Around Registry
After analyzing multiple approaches, I've selected the **Thin Wrapper** approach that leverages the existing `llmRegistry.callProvider` function.

### Implementation Details
1. Create a `ConcreteLLMClient` class that implements the `LLMClient` interface
2. The client will:
   - Parse the combined "provider:modelId" format from the `modelId` parameter
   - Load configuration via `ConfigManagerInterface` (injected dependency)
   - Find the model configuration and group information
   - Determine the final system prompt based on priority (override > model > group)
   - Delegate to `llmRegistry.callProvider` with the appropriate parameters
   - Handle errors consistently, wrapping them in appropriate error types

### Key Considerations
- **Dependency Injection**: The `ConfigManagerInterface` is injected as a dependency rather than directly imported
- **Error Handling**: Errors from `callProvider` or config loading are wrapped in appropriate error types
- **Minimal Logic**: The implementation avoids duplicating complex logic from `callProvider` and `configManager`

## Testability Considerations
This approach strongly aligns with the testing philosophy in several ways:

1. **Minimizes Mocking**: By delegating to `callProvider`, we avoid having to mock complex LLM provider internals. We only need to mock:
   - `ConfigManagerInterface` (an external boundary)
   - `callProvider` function (an internal helper)

2. **Behavior Over Implementation**: Tests for the `ConcreteLLMClient` will focus on verifying that:
   - It correctly parses the provider:modelId format
   - It loads configuration properly
   - It calls `callProvider` with the correct parameters

3. **Clear External Boundary**: By implementing the `LLMClient` interface, we create a clean abstraction for other components, allowing them to:
   - Depend on the interface rather than concrete implementations
   - Easily mock the interface for their own tests

4. **Testability of Dependent Code**: Services that consume `LLMClient` can now easily be tested by providing a mock implementation of the interface, avoiding the need to mock actual LLM provider interactions.

## Why This Approach
I chose the thin wrapper approach because it:

1. **Follows Simplicity Principle**: Minimizes new code and complexity while fulfilling the requirement
2. **Leverages Existing Components**: Reuses well-tested logic from `llmRegistry.callProvider`
3. **Provides Clear Abstraction**: Creates a distinct `LLMClient` interface for dependency injection
4. **Prioritizes Testability**: Makes both the client itself and its consumers easier to test

The approach avoids complicating the codebase with redundant logic while still achieving the core goal of better dependency injection and testability.