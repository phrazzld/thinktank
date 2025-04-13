```markdown
## Approach 1: Modify `GatherContext` to Calculate Token Count Directly

### Steps:
1.  After collecting all file content in `GatherContext`, call `cg.client.CountTokens(ctx, projectContext)` to get the token count.
2.  Store the returned token count in the `ContextStats` struct.
3.  Update the return values of `GatherContext` to include the updated `ContextStats`.
4.  Update tests to assert the token count in the returned `ContextStats`.

### Pros:
*   Simple and straightforward implementation.
*   Keeps all context gathering logic in one place.

### Cons:
*   Violates Separation of Concerns (`ARCHITECTURE_GUIDELINES.md`) by mixing file gathering and token counting.
*   Makes `GatherContext` more complex and harder to test in isolation.
*   Tight coupling between file gathering and Gemini client.

### Evaluation Against Standards:

*   **CORE_PRINCIPLES.md:** Violates Simplicity by adding token counting logic to the context gathering process.
*   **ARCHITECTURE_GUIDELINES.md:** Violates Separation of Concerns by mixing file system operations with external API calls.
*   **CODING_STANDARDS.md:** No direct violations, but the increased complexity could lead to less readable code.
*   **TESTING_STRATEGY.md:** Makes `GatherContext` harder to test. Requires mocking the `gemini.Client` to test the file gathering logic, even if the token counting is not relevant to the test. This violates the principle of minimal mocking.
*   **DOCUMENTATION_APPROACH.md:** Increases the complexity of `GatherContext`, potentially requiring more detailed comments to explain the combined functionality.

## Approach 2: Introduce a `TokenCountingContextGatherer` Decorator

### Steps:
1.  Create a new struct `TokenCountingContextGatherer` that implements the `ContextGatherer` interface.
2.  `TokenCountingContextGatherer` takes an existing `ContextGatherer` and a `gemini.Client` as dependencies.
3.  Implement the `GatherContext` method in `TokenCountingContextGatherer`. This method calls the `GatherContext` method of the inner `ContextGatherer` to get the file metadata and initial stats.
4.  After the inner `GatherContext` returns, `TokenCountingContextGatherer` calculates the token count using the `gemini.Client` and updates the `ContextStats`.
5.  Return the updated `ContextStats` and file metadata.
6.  Update the orchestrator to use the `TokenCountingContextGatherer` when token counting is needed.
7.  Update tests to verify the token counting functionality of the decorator.

### Pros:
*   Adheres to Separation of Concerns (`ARCHITECTURE_GUIDELINES.md`). The original `contextGatherer` remains focused on file gathering.
*   Improves testability. The original `contextGatherer` can be tested independently of the token counting logic. The decorator can be tested separately with a mocked inner `ContextGatherer`.
*   Follows the Decorator pattern, allowing for flexible addition of token counting functionality.

### Cons:
*   More complex to implement than Approach 1.
*   Requires changes to the orchestrator to use the decorator.

### Evaluation Against Standards:

*   **CORE_PRINCIPLES.md:** Promotes Simplicity by keeping the core `contextGatherer` focused on its primary task.
*   **ARCHITECTURE_GUIDELINES.md:** Adheres to Separation of Concerns by isolating token counting logic in a separate component.
*   **CODING_STANDARDS.md:** May require more code, but the improved structure and testability can lead to more maintainable code.
*   **TESTING_STRATEGY.md:** Improves testability by allowing independent testing of the file gathering and token counting logic. Reduces the need for complex mocking.
*   **DOCUMENTATION_APPROACH.md:** The decorator pattern might require additional documentation to explain the interaction between the decorator and the decorated object.

## Approach 3: Post-Process ContextStats in the Orchestrator

### Steps:

1.  Keep the `GatherContext` method as is, returning the initial `ContextStats` without token counts.
2.  In the `Orchestrator.Run` method, after calling `contextGatherer.GatherContext`, calculate the token count using the `gemini.Client` and update the `ContextStats` struct.
3.  Update the orchestrator to use the `gemini.Client` to calculate the token count.
4.  Update tests to verify the token counting functionality in the orchestrator.

### Pros:

*   Minimally invasive, requiring the least amount of change to existing code.
*   Keeps the `ContextGatherer` focused on its original responsibility.

### Cons:

*   Violates Separation of Concerns (`ARCHITECTURE_GUIDELINES.md`) by placing token counting logic in the orchestrator, which is responsible for coordinating the overall workflow.
*   Tight coupling between the orchestrator and the `gemini.Client`.
*   Reduces testability of the context gathering process, as the token count is not directly associated with the gathered context.

### Evaluation Against Standards:

*   **CORE_PRINCIPLES.md:** Slightly violates Simplicity by adding token counting logic to the orchestrator.
*   **ARCHITECTURE_GUIDELINES.md:** Violates Separation of Concerns by mixing orchestration logic with external API calls.
*   **CODING_STANDARDS.md:** No direct violations, but the increased complexity in the orchestrator could lead to less readable code.
*   **TESTING_STRATEGY.md:** Makes it harder to test the context gathering process in isolation, as the token count is calculated separately in the orchestrator.
*   **DOCUMENTATION_APPROACH.md:** Increases the complexity of the orchestrator, potentially requiring more detailed comments to explain the added functionality.

## Recommendation

**Approach 2: Introduce a `TokenCountingContextGatherer` Decorator** is the best approach.

### Justification:

Approach 2 best aligns with the project's standards hierarchy:

1.  **Simplicity/Clarity (`CORE_PRINCIPLES.md`):** While it involves more initial code than Approach 1, it ultimately leads to a simpler and more focused design by adhering to Separation of Concerns. The core `contextGatherer` remains simple, and the token counting logic is encapsulated in a separate decorator.
2.  **Separation of Concerns (`ARCHITECTURE_GUIDELINES.md`):** This is the strongest argument for Approach 2. It cleanly separates the file gathering logic from the token counting logic, preventing the `contextGatherer` from becoming a monolithic component.
3.  **Testability (Minimal Mocking) (`TESTING_STRATEGY.md`):** Approach 2 significantly improves testability. The original `contextGatherer` can be tested in isolation without mocking the `gemini.Client`. The `TokenCountingContextGatherer` can be tested separately with a mocked inner `ContextGatherer` and `gemini.Client`. This minimizes the need for complex mocking and ensures that tests are focused on verifying the specific behavior of each component.
4.  **Coding Conventions (`CODING_STANDARDS.md`):** No direct violations, and the improved structure can lead to more maintainable code.
5.  **Documentability (`DOCUMENTATION_APPROACH.md`):** The decorator pattern might require additional documentation, but the improved structure and testability make the overall system easier to understand and maintain.

While Approach 3 is the least invasive initially, it introduces coupling and violates Separation of Concerns, making it a less desirable solution in the long run. Approach 1 is simpler than Approach 2 to implement initially, but it violates Separation of Concerns and makes testing more difficult.

The trade-off for the increased initial complexity of Approach 2 is a more modular, testable, and maintainable system that adheres to the project's core principles and architectural guidelines.
```