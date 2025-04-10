# Plan: Update Integration Tests To Work With Refactored Architecture

## Task Title
Update integration tests if needed

## Goal
Modify the existing integration tests to work properly with the refactored code structure, where functionality from the original `main.go` has been extracted into separate components in the `cmd/architect` package.

## Chosen Approach: Update `MainAdapter` to Orchestrate Real Components

The most effective approach is to update the `MainAdapter` in the integration tests to orchestrate the real, refactored components from `cmd/architect`. This approach will:

1. Keep the existing `TestEnv` and test case structure, minimizing disruption
2. Modify `MainAdapter` to use the actual components from `cmd/architect` rather than maintaining its own simplified versions
3. Leverage dependency injection principles to inject the mock Gemini client

### Implementation Steps

1. Update `internal/integration/main_adapter.go` to:
   - Replace its internal implementations of flag parsing, context gathering, and plan generation
   - Orchestrate the real components from `cmd/architect` package
   - Inject the MockClient into the appropriate components
   
2. Modify integration tests to work with the updated adapter:
   - Ensure test cases properly set up the environment
   - Verify the tests still check the appropriate outputs and behaviors

3. Add interface compatibility checks to ensure mocks and real implementations remain compatible

## Reasoning for This Approach

This approach was selected because it best aligns with the project's core standards and philosophy:

1. **Simplicity (Core Principles)**: Maintains simplicity by reusing existing test infrastructure rather than creating completely new testing patterns.

2. **Modularity (Core Principles & Architecture Guidelines)**: Directly tests how the newly extracted components work together, which is the key aspect of the refactoring effort. This approach properly respects the separation of concerns established in the refactored architecture.

3. **Testability (Core Principles & Testing Strategy)**: Provides a clean way to inject mocks for external dependencies (Gemini API) without complex subprocess handling or global state manipulation. This aligns perfectly with the testing philosophy of minimizing mocks while focusing on behavior rather than implementation details.

4. **Code Quality (Coding Standards)**: Follows good practices through dependency injection rather than relying on global state manipulation or subprocess execution, which would be more fragile.

5. **Documentation (Documentation Approach)**: The adapter's role as an orchestrator for real components will be clearer and easier to document than more complex approaches.

### Trade-offs Considered

While this approach doesn't directly test the absolute entry point (`architect.Main`), it tests the core application logic and component interactions, which is where most of the business logic and potential issues reside. The setup code in `architect.Main` is relatively minimal boilerplate that is less likely to contain significant bugs.

Alternative approaches considered were:
1. Testing the compiled binary via subprocess - rejected due to mocking complexity and slow test execution
2. Invoking `cmd/architect.Main()` directly - rejected due to reliance on global state manipulation and potential test fragility

The chosen approach offers the best balance between testing thoroughness and clean, maintainable test code.