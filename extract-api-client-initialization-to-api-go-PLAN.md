# Extract API client initialization to api.go

## Goal
Move the initGeminiClient function from main.go to api.go, providing a clean interface for API operations that maintains the same behavior while improving testability and decoupling.

## Chosen Approach: Minimal Interface

Based on the thinktank analysis, I'll implement the "Minimal Interface" approach which keeps the `APIService` interface focused only on the actions it performs, requiring the necessary parameters for each operation. This maintains a clean separation of concerns while making the code more testable.

### Implementation Steps:

1. Update the existing `APIService` interface and `apiService` implementation in `cmd/architect/api.go` with a proper implementation of the `InitClient` method that:
   - Takes context, apiKey, and modelName as parameters
   - Calls gemini.NewClient with the provided parameters
   - Returns the client or an error (instead of calling logger.Fatal directly)

2. Update `main.go` to:
   - Create an instance of `APIService` via `architect.NewAPIService(logger)`
   - Call the `InitClient` method with the necessary parameters from the config
   - Handle the error returned by `InitClient` with the fatal logging that was in the original function
   - Remove the old `initGeminiClient` function

3. Add a transitional comment to the removed code in `main.go` to maintain the code trail

### Reasoning for Choice

This approach is the most aligned with the project's testing philosophy and design principles:

1. **Testability**: 
   - The clear interface with explicit parameters makes it easy to test without complex mocking
   - Following the "behavior over implementation" principle, we can test that `InitClient` properly creates the client or returns an error
   - Error handling in `main.go` allows testing of the API service in isolation

2. **Minimal Coupling**: 
   - The `APIService` has no dependency on the `main.Configuration` struct
   - It only depends on primitive types and the `gemini.Client` interface
   - This decoupling makes the component more reusable and maintainable

3. **Clear Intent and Responsibility**: 
   - The interface clearly states what data is required for client initialization
   - The responsibility for fatal error handling remains with the application's entry point (main.go)
   - Component responsibilities are cleanly separated

4. **Future Extensibility**:
   - This pattern sets a good foundation for adding other API-related methods like `ProcessResponse`
   - Consistent interface design will make the codebase more maintainable

While there's a slight change in where the fatal logging occurs, the overall observable behavior of the program remains identical, which is the primary goal of this refactoring.