```markdown
# PLAN.md: Replace Gemini-based Name Generation with Deterministic Random Name Generation

## 1. Overview

This plan outlines the steps to replace the current Gemini API-based friendly name generation for ThinkTank output directories with a simpler, deterministic, offline approach. The new method will randomly select one adjective and one noun from predefined lists and combine them in the format "adjective-noun". This change aims to eliminate external API calls, reduce startup latency, remove dependencies, and improve the diversity and reliability of generated names while maintaining the existing fallback mechanism.

## 2. Task Breakdown

| Task                                                     | Description                                                                                                                               | Files Affected                                                                 | Effort |
| :------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------------- | :----- |
| **1. Define Word Lists**                                 | Create and add arrays of at least 50 adjectives and 50 nouns to `nameGenerator.ts`.                                                       | `src/utils/nameGenerator.ts`                                                   | S      |
| **2. Refactor `generateFunName`**                        | Modify `generateFunName` to use the word lists for random name generation. Remove API call logic, API key checks, and related imports. Make the function synchronous. | `src/utils/nameGenerator.ts`                                                   | M      |
| **3. Update Integration Point**                          | Modify the call to `generateFunName` in `runThinktank.ts` to handle the synchronous nature of the updated function.                       | `src/workflow/runThinktank.ts`                                                 | S      |
| **4. Update Unit Tests**                                 | Rewrite tests in `nameGenerator.test.ts` to remove API mocks and verify the new deterministic random generation logic and format.       | `src/utils/__tests__/nameGenerator.test.ts`                                    | M      |
| **5. Code Cleanup**                                      | Remove unused imports (`@google/generative-ai`) and any dead code related to the previous API implementation in `nameGenerator.ts`.       | `src/utils/nameGenerator.ts`                                                   | S      |
| **6. (Bonus) Implement Themed Word Lists**               | *Optional:* Explore adding themed sets of adjectives/nouns and logic to select a theme.                                                   | `src/utils/nameGenerator.ts`, potentially config files or CLI options        | M      |
| **7. Manual Verification & Integration Testing**         | Run the application manually to ensure names are generated correctly, appear in logs/directory names, and the fallback works if needed. | CLI, Output directories                                                        | S      |

**Effort Estimation:** S = Small (<= 2 hours), M = Medium (2-4 hours), L = Large (4+ hours)

## 3. Implementation Details

### Task 1: Define Word Lists

-   Add two constant arrays near the top of `src/utils/nameGenerator.ts`:

    ```typescript
    // src/utils/nameGenerator.ts

    const ADJECTIVES: string[] = [
      "adaptable", "adventurous", "affable", "agile", "alert", "amiable",
      "analytical", "attentive", "balanced", "bold", "bright", "brilliant",
      "calm", "capable", "careful", "charming", "cheerful", "clever",
      "coherent", "communicative", "compassionate", "composed", "confident", "conscientious",
      "consistent", "cooperative", "courageous", "courteous", "creative", "curious",
      "daring", "decisive", "dedicated", "deep", "delightful", "determined",
      "diligent", "diplomatic", "disciplined", "discreet", "dynamic", "eager",
      "efficient", "elastic", "eloquent", "energetic", "enhanced", "enlightened",
      "enormous", "enterprising", "enthusiastic", "epic", "exact", "excellent",
      "expert", "fabulous", "fair", "faithful", "fantastic", "fearless",
      // ... add more to reach at least 50
    ];

    const NOUNS: string[] = [
      "acorn", "anchor", "apple", "arrow", "aurora", "automaton", "axiom",
      "badger", "beacon", "bear", "beaver", "blossom", "breeze", "brook",
      "butterfly", "canyon", "cascade", "cat", "comet", "compass", "coyote",
      "cricket", "crystal", "current", "cyborg", "dawn", "deer", "delta",
      "diamond", "dolphin", "dove", "dragon", "dream", "drift", "eagle",
      "echo", "eclipse", "element", "elixir", "engine", "envoy", "epoch",
      "falcon", "firefly", "fjord", "flame", "fleet", "flow", "forest",
      "formula", "fox", "galaxy", "gazelle", "geode", "glacier", "glimmer",
      // ... add more to reach at least 50
    ];
    ```

-   Ensure words are lowercase and suitable for use in directory names (avoid special characters beyond hyphens if extending).

### Task 2: Refactor `generateFunName`

-   Remove the `async` keyword and `Promise<string | null>` return type. The function should become synchronous and return `string`.
-   Remove all code related to `GoogleGenerativeAI`, API key checks (`process.env.GEMINI_API_KEY`), `try...catch` blocks for the API call, and response parsing/validation.
-   Implement the random selection logic:

    ```typescript
    // src/utils/nameGenerator.ts

    /**
     * Generates a fun, deterministic random name for a run.
     * Format is 'adjective-noun' (e.g., clever-otter, swift-breeze)
     *
     * @returns A fun name string.
     */
    export function generateFunName(): string {
      // Basic check to prevent errors if lists are empty (should not happen in practice)
      if (ADJECTIVES.length === 0 || NOUNS.length === 0) {
        logger.error("Adjective or Noun list is empty. Cannot generate fun name.");
        // Return a predictable placeholder instead of null to simplify caller
        // The fallback mechanism in runThinktank should ideally handle this upstream if needed.
        // Or, alternatively, throw an error here. Let's return a placeholder for now.
        // Consider throwing an Error for a more robust signal of configuration issue.
        // For now, let's ensure the caller (`runThinktank`) still uses its fallback.
        // Reverting to returning null to keep the caller's logic simple.
        // throw new Error("Adjective or Noun list is empty."); // Option 1: Throw
        // return "error-generating-name"; // Option 2: Placeholder
         return generateFallbackName(); // Option 3: Directly use fallback (simplest integration)
      }

      const adjIndex = Math.floor(Math.random() * ADJECTIVES.length);
      const nounIndex = Math.floor(Math.random() * NOUNS.length);

      const adjective = ADJECTIVES[adjIndex];
      const noun = NOUNS[nounIndex];

      return `${adjective}-${noun}`;
    }

    // Keep generateFallbackName as is
    export function generateFallbackName(): string {
      // ... (existing implementation)
    }

    // Remove logger import if no longer needed after removing API calls
    // Remove GoogleGenerativeAI imports
    ```
    *Self-correction during planning:* Decided to call `generateFallbackName` directly within `generateFunName` if lists are empty. This simplifies the caller (`runThinktank`) as `generateFunName` now *always* returns a valid string, fulfilling the format requirement even in the edge case of empty lists (which indicates a setup problem). This avoids changing the signature or return type expectations drastically in the caller.

### Task 3: Update Integration Point

-   Modify `src/workflow/runThinktank.ts` around lines 324-331:
    -   Remove the `await` keyword when calling `generateFunName`.
    -   Adjust the logic slightly as `generateFunName` now guarantees a string return (either random or fallback).

    ```typescript
    // src/workflow/runThinktank.ts (around lines 324-331)

    // 1.5 Generate a friendly run name
    spinner.text = 'Generating run identifier...';
    // Remove await, generateFunName is now synchronous and always returns a string
    const friendlyRunName = generateFunName();

    // The function now handles its own fallback if lists are empty,
    // so we can directly use the result. Log it for info.
    spinner.info(styleInfo(`Run name: ${styleSuccess(friendlyRunName)}`));
    spinner.start(); // Restart spinner for next step

    // ... rest of the function ...

    // Pass friendlyRunName to createOutputDirectory
    const outputDirectoryPath = await createOutputDirectory({
      outputDirectory: options.output,
      directoryIdentifier, // Keep existing identifier logic
      friendlyRunName // Pass the generated name
    });
    spinner.info(styleInfo(`Output directory: ${outputDirectoryPath} (Run: ${friendlyRunName})`));

    // ... ensure friendlyRunName is used later for logging etc. ...
    ```

### Task 4: Update Unit Tests

-   In `src/utils/__tests__/nameGenerator.test.ts`:
    -   Remove `jest.mock('@google/generative-ai');`.
    -   Remove mocks for `process.env` related to API keys.
    -   Remove all test cases related to API key presence/absence, API call success/failure, and response parsing (JSON/text/invalid).
    -   Add new test cases for `generateFunName`:
        -   Verify the output format is always `adjective-noun`.
        *   Verify that the returned adjective exists in the `ADJECTIVES` array.
        *   Verify that the returned noun exists in the `NOUNS` array.
        *   Call the function multiple times and assert that the outputs are different (probabilistically testing randomness).
        *   *Optional:* Test the edge case of empty lists (if the implementation throws or returns a specific value). If it calls `generateFallbackName`, test that the output matches the fallback format.
    -   Keep the existing test case for `generateFallbackName` to ensure it still works correctly.

    ```typescript
    // src/utils/__tests__/nameGenerator.test.ts (Example structure)
    import { generateFunName, generateFallbackName } from '../nameGenerator';
    // Import the actual lists if needed for verification, or mock them if preferred for isolation
    // import { ADJECTIVES, NOUNS } from '../nameGenerator'; // Requires exporting them

    describe('nameGenerator', () => {
      // Mock logger if needed
      beforeEach(() => {
        jest.spyOn(console, 'debug').mockImplementation(() => {});
        jest.spyOn(console, 'error').mockImplementation(() => {});
        // Restore Math.random if mocking it for specific tests
        jest.spyOn(global.Math, 'random').mockRestore();
      });

      describe('generateFunName', () => {
        it('should return a name in the format "adjective-noun"', () => {
          const name = generateFunName();
          expect(name).toMatch(/^[a-z]+-[a-z]+$/);
        });

        it('should use words from the predefined lists', () => {
          const name = generateFunName();
          const [adjective, noun] = name.split('-');
          // This requires exporting ADJECTIVES and NOUNS or having a way to access them
          // expect(ADJECTIVES).toContain(adjective);
          // expect(NOUNS).toContain(noun);
          // Alternative: Just check format if lists aren't exported
          expect(typeof adjective).toBe('string');
          expect(adjective.length).toBeGreaterThan(0);
          expect(typeof noun).toBe('string');
          expect(noun.length).toBeGreaterThan(0);
        });

        it('should generate different names on subsequent calls (usually)', () => {
          const name1 = generateFunName();
          const name2 = generateFunName();
          const name3 = generateFunName();
          // It's statistically unlikely they are all the same with >50 words each
          expect(name1).not.toBe(name2); // High probability
          expect(new Set([name1, name2, name3]).size).toBeGreaterThan(1); // More robust check
        });

        // Optional: Test empty list scenario if generateFunName handles it specifically
        // it('should return a fallback name if lists are empty', () => {
        //   // Mock or temporarily empty the lists before calling
        //   // ... setup ...
        //   const name = generateFunName();
        //   expect(name).toMatch(/^run-\d{8}-\d{6}$/); // Check fallback format
        //   // ... teardown ...
        // });
      });

      describe('generateFallbackName', () => {
        it('should generate a timestamp-based name in the correct format', () => {
          const result = generateFallbackName();
          expect(result).toMatch(/^run-\d{8}-\d{6}$/);
        });
      });
    });
    ```

## 4. Potential Challenges & Considerations

1.  **Word List Quality:** The quality and diversity of the adjective and noun lists directly impact the perceived quality of the generated names. Need to curate lists that are generally positive/neutral, non-offensive, and varied.
2.  **Randomness:** `Math.random()` is pseudo-random and sufficient for this purpose. True cryptographic randomness is not required. Collisions (generating the same name twice) are possible but statistically unlikely with lists of 50+ items each (1 in 2500+ chance per generation).
3.  **Maintainability:** Adding/modifying word lists requires code changes. The bonus task addresses this by potentially externalizing or structuring the lists.
4.  **Empty Lists Edge Case:** The implementation should gracefully handle the unlikely scenario where the lists might be empty (e.g., during development or due to a mistake). Calling `generateFallbackName` seems like the most robust approach within `generateFunName` itself.
5.  **"Deterministic" Meaning:** The goal is deterministic *execution* (no external factors like API availability/latency), not necessarily generating the *same* name sequence across runs (which would require seeding the RNG).

## 5. Testing Strategy

1.  **Unit Tests (`nameGenerator.test.ts`):**
    *   Verify `generateFunName` returns a string matching the `adjective-noun` pattern.
    *   Verify the components (adjective, noun) are non-empty strings.
    *   (If lists are exported/accessible) Verify words come from the defined lists.
    *   Run multiple times to check for *different* outputs (probabilistic check for randomness).
    *   Verify `generateFallbackName` still produces the correct timestamp format.
    *   Test the empty list edge case if handled within `generateFunName`.
2.  **Integration Tests:**
    *   Modify existing integration tests (if any) or add new ones that run the `runThinktank` workflow.
    *   Assert that the console output and generated directory name include a friendly name matching the `adjective-noun` format.
    *   Ensure the application starts up faster without the API call delay.
3.  **Manual Testing:**
    *   Run `thinktank` via the CLI multiple times.
    *   Observe the "Run name:" log message – check format and variety.
    *   Check the created output directory names in `thinktank-output/` – verify format (`adjective-noun` or `identifier-adjective-noun`) and variety.
    *   Temporarily modify `nameGenerator.ts` to have empty lists and verify that a fallback name (e.g., `run-YYYYMMDD-HHmmss`) is generated and used.

## 6. Open Questions

1.  **Word List Source:** Are there specific requirements for the source or theme of the initial 50+ adjectives and nouns? (Assuming general, positive/neutral words for now).
2.  **Bonus Feature - Themes:** If the bonus is pursued, how should themes be specified? (e.g., CLI flag `--name-theme=space`, config file setting). What themes are desired initially?
```