# Token Counting Implementation Analysis

## Task Description Recap

The task is to modify `ContextGatherer.GatherContext` to use the injected `gemini.Client` for accurate token counting after collecting file content. This involves replacing the existing estimation logic (`fileutil.CalculateStatistics`) with a call to `client.CountTokens`, handling potential errors or the absence of the client by falling back to the estimation, updating the `ContextStats` struct, and ensuring tests are updated.

## Proposed Implementation Approaches

### Approach 1: Direct Replacement with Fallback in `GatherContext`

**Steps:**

1.  Locate the section in `GatherContext` after `projectContext` (the combined content string) is created.
2.  Keep the existing direct calculation for `charCount` and `lineCount`.
3.  Add a check: `if cg.client != nil`.
4.  **Inside the `if` block:**
    *   Call `tokenResult, err := cg.client.CountTokens(ctx, projectContext)`.
    *   If `err == nil`:
        *   Set `stats.TokenCount = tokenResult.Total`.
        *   Log that an accurate token count was obtained.
    *   If `err != nil`:
        *   Log a warning about the failure to count tokens accurately (including the error).
        *   Call `_, _, estimatedTokenCount := fileutil.CalculateStatistics(projectContext)`.
        *   Set `stats.TokenCount = int32(estimatedTokenCount)`.
        *   Log that an estimated token count is being used.
5.  **Inside an `else` block (for `cg.client == nil`):**
    *   Log a debug message indicating fallback due to nil client.
    *   Call `_, _, estimatedTokenCount := fileutil.CalculateStatistics(projectContext)`.
    *   Set `stats.TokenCount = int32(estimatedTokenCount)`.
    *   Log that an estimated token count is being used.
6.  Ensure the final logging (`cg.logger.Info("Context gathered: ...")`) uses the calculated `stats.TokenCount`.
7.  Update the audit log entry (`GatherContextEnd`) to include the final `stats.TokenCount` in the `Outputs` map. Add a field indicating if the count is estimated or accurate, if desired.
8.  Update tests in `context_test.go`:
    *   Enhance `mockGeminiClient` to allow mocking `CountTokens` behavior (success, error).
    *   Add test cases verifying successful token counting via the mocked client.
    *   Add test cases verifying fallback to estimation when `CountTokens` returns an error.
    *   Add test cases verifying fallback to estimation when the injected client is `nil`.
    *   Verify audit log entries contain the correct token count.

**Pros:**

*   **Simple & Direct:** Integrates the logic directly into the existing flow with minimal structural changes.
*   **Localized Logic:** Keeps context gathering and its quantification (including token count) within the same method.
*   **Clear Fallback:** Explicit conditional logic for handling client absence or API errors.

**Cons:**

*   **Method Complexity:** Slightly increases the conditional logic and length of the `GatherContext` method.

**Evaluation Against Standards:**

*   **`CORE_PRINCIPLES.md`:**
    *   *Simplicity:* High. Minimal changes to existing structure.
    *   *Modularity:* Acceptable. Token counting is part of quantifying the gathered context.
    *   *Testability:* High. Relies on mocking the `gemini.Client` interface.
    *   *Maintainability:* High. Logic is straightforward.
    *   *Explicit:* Explicit checks for client and errors.
*   **`ARCHITECTURE_GUIDELINES.md`:**
    *   *Unix Philosophy:* `ContextGatherer` still focuses on gathering/quantifying context. Good.
    *   *Separation of Concerns:* Core file gathering (`fileutil`) is separate. Infrastructure interaction (`gemini.Client`) is via an interface. Good.
    *   *Dependency Inversion:* Depends on `gemini.Client` interface. Good.
    *   *Package Structure:* No change needed.
    *   *Error Handling:* Explicit handling of `CountTokens` error. Good.
*   **`CODING_STANDARDS.md`:**
    *   Adheres well. Requires clear comments for fallback logic.
*   **`TESTING_STRATEGY.md`:**
    *   *Testability:* Excellent. Mocking `gemini.Client` fits the strategy perfectly (mocking external boundaries via interfaces). Minimal mocking required.
*   **`DOCUMENTATION_APPROACH.md`:**
    *   Requires minor updates to code comments within `GatherContext`. No ADR needed.

### Approach 2: Dedicated Private Helper Method for Stats Calculation

**Steps:**

1.  Define a new private method: `func (cg *contextGatherer) calculateStats(ctx context.Context, content string) (charCount, lineCount int, tokenCount int32, isEstimate bool, err error)`.
2.  Move the logic for calculating char count, line count, and token count (API call, fallback) into `calculateStats`.
    *   Calculate `charCount` and `lineCount` directly.
    *   Check `cg.client != nil`.
    *   Attempt `cg.client.CountTokens`.
    *   Handle errors/nil client, falling back to `fileutil.CalculateStatistics` for `tokenCount`.
    *   Set `isEstimate` flag accordingly.
    *   Return the calculated values.
3.  In `GatherContext`, after creating `projectContext`:
    *   Call `charCount, lineCount, tokenCount, isEstimate, calcErr := cg.calculateStats(ctx, projectContext)`.
    *   Handle `calcErr` if necessary (though likely just logs warnings internally).
    *   Update the `stats` struct: `stats.CharCount = charCount`, `stats.LineCount = lineCount`, `stats.TokenCount = tokenCount`.
4.  Update logging messages to indicate if the count is estimated based on the `isEstimate` flag.
5.  Update audit logs as in Approach 1.
6.  Update tests as in Approach 1. Testing the private method directly is possible but testing via the public `GatherContext` is preferred.

**Pros:**

*   **Improved Internal Separation:** Isolates statistics calculation logic, making `GatherContext` slightly cleaner.
*   **Focused Helper:** The `calculateStats` method has a single, clear responsibility.

**Cons:**

*   **Increased Structure:** Adds a private method, slightly increasing the overall number of methods in the type.
*   **Data Flow:** Requires passing data back from the helper method to update the main `stats` struct.

**Evaluation Against Standards:**

*   **`CORE_PRINCIPLES.md`:**
    *   *Simplicity:* Comparable to Approach 1. `GatherContext` is simpler, but overall structure is slightly more complex.
    *   *Modularity:* Better internal modularity within `contextGatherer`.
    *   *Testability:* Same as Approach 1 (mocking `gemini.Client`).
    *   *Maintainability:* Potentially slightly improved due to separation.
    *   *Explicit:* Logic remains explicit.
*   **`ARCHITECTURE_GUIDELINES.md`:**
    *   *Unix Philosophy:* Helper method adheres well ("do one thing"). Good.
    *   *Separation of Concerns:* Good internal separation. External separation unchanged.
    *   *Dependency Inversion:* Same as Approach 1.
    *   *Package Structure:* No change needed.
    *   *Error Handling:* Encapsulated within the helper. Good.
*   **`CODING_STANDARDS.md`:**
    *   Adheres well. Requires clear naming and comments for the new method.
*   **`TESTING_STRATEGY.md`:**
    *   *Testability:* Excellent. Same reasoning as Approach 1 – relies on mocking the external `gemini.Client` interface.
*   **`DOCUMENTATION_APPROACH.md`:**
    *   Requires comments for the new private method and updates within `GatherContext`. No ADR needed.

### Approach 3: Inject a Dedicated `TokenCounter` Service (Overkill)

**Steps:**

1.  Define `type TokenCounter interface { CountTokens(ctx context.Context, content string) (int32, error) }`.
2.  Implement `type geminiTokenCounter struct { client gemini.Client }` which satisfies `TokenCounter`.
3.  Modify `contextGatherer` to take `tokenCounter TokenCounter` instead of `gemini.Client`.
4.  In `GatherContext`, call `tokenCount, err := cg.tokenCounter.CountTokens(ctx, projectContext)`.
5.  Handle `err` or `cg.tokenCounter == nil`, falling back to `fileutil.CalculateStatistics`. (Fallback logic still needs to live somewhere, likely `ContextGatherer`).
6.  Update dependency injection setup to create and pass `geminiTokenCounter`.
7.  Update tests to mock `TokenCounter` instead of `gemini.Client`.

**Pros:**

*   **Maximum Separation:** Token counting is fully decoupled.
*   **Simplified Gatherer:** `ContextGatherer` has fewer direct dependencies.

**Cons:**

*   **Over-Engineering:** Introduces significant structural complexity (new interface, implementation, DI changes) for a single API call. Violates YAGNI.
*   **Complexity Shift:** Fallback logic still needs handling, potentially complicating the `ContextGatherer` or requiring a complex composite `TokenCounter`.
*   **Increased Components:** More pieces to manage.

**Evaluation Against Standards:**

*   **`CORE_PRINCIPLES.md`:**
    *   *Simplicity:* Poor. Increases overall system complexity unnecessarily.
    *   *Modularity:* Excellent, but likely excessive.
    *   *Testability:* Excellent via the new interface, but Approach 1 is already highly testable per the strategy.
*   **`ARCHITECTURE_GUIDELINES.md`:**
    *   *Unix Philosophy:* Creates a focused `TokenCounter`. Good.
    *   *Separation of Concerns:* Maximized. Good.
    *   *Dependency Inversion:* Good, uses new abstraction.
*   **`CODING_STANDARDS.md`:**
    *   Adheres, but adds more code surface.
*   **`TESTING_STRATEGY.md`:**
    *   *Testability:* Excellent, but introduces an arguably unnecessary mock point. The existing `gemini.Client` interface is already the correct boundary according to the strategy.
*   **`DOCUMENTATION_APPROACH.md`:**
    *   Requires documenting the new interface/implementation.

## Recommendation

**Recommended Approach: Approach 1: Direct Replacement with Fallback in `GatherContext`**

**Justification:**

1.  **Simplicity/Clarity (`CORE_PRINCIPLES.md`):** Approach 1 is the simplest and most direct way to achieve the goal. It modifies existing code minimally without adding new structural elements like helper methods or interfaces, making it easy to understand and maintain. Approach 3 is overly complex (violates YAGNI), and Approach 2 adds a helper method which, while separating concerns internally, offers marginal benefit over Approach 1 for this specific task.
2.  **Separation of Concerns (`ARCHITECTURE_GUIDELINES.md`):** Approach 1 respects the crucial separation between the application logic (`ContextGatherer`) and infrastructure (`gemini.Client`) by using the injected interface. The token counting is a direct quantification of the context gathered, making its inclusion within `GatherContext` reasonable. The core file processing remains delegated to `fileutil`.
3.  **Testability (Minimal Mocking) (`TESTING_STRATEGY.md`):** This is a key differentiator. Approach 1 aligns perfectly with the testing strategy by requiring mocking only at the existing, appropriate external boundary – the `gemini.Client` interface. It avoids introducing new interfaces or unnecessary layers of mocking. Minimal, targeted mocking is preferred.
4.  **Coding Conventions (`CODING_STANDARDS.md`):** Approach 1 requires the least amount of new code, fitting naturally into the existing structure and adhering to conventions.
5.  **Documentability (`DOCUMENTATION_APPROACH.md`):** Requires the least documentation effort – only updating comments within the affected method.

Approach 1 delivers the required functionality with the least complexity, adheres strongly to the principle of simplicity, and leverages the existing architecture and testing strategy effectively without introducing unnecessary abstractions. The slight increase in the `GatherContext` method's length is an acceptable trade-off for overall system simplicity and adherence to the testing strategy's emphasis on mocking true external boundaries.