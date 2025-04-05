```markdown
# Refactoring Plan: thinktank CLI

**Date:** 2024-07-27
**Author:** AI Software Architect

## 1. Overview

This plan outlines the steps to refactor the `thinktank` codebase. The primary goals are to improve code simplicity, readability, and maintainability, potentially reduce codebase size, and ensure 100% of existing functionality is preserved. The refactoring will address areas like CLI structure, configuration management complexity, error handling consistency, and test coverage, paving the way for easier future development as outlined in `BACKLOG.md`.

## 2. Current State Analysis

The codebase is a TypeScript CLI application using Commander.js for command parsing, various SDKs for LLM interactions, and a JSON-based configuration system with XDG support.

**Strengths:**

*   **Modular Structure:** Good separation of concerns into `core`, `providers`, `cli`, `utils`, and `workflow`.
*   **Provider Pattern:** Clear `LLMProvider` interface and registry (`llmRegistry.ts`) for extensibility.
*   **Configuration:** Powerful cascading configuration system, although complex. XDG support is good practice.
*   **Documentation:** Detailed `README.md` explaining usage and configuration. `BEST-PRACTICES.md` provides guidelines.

**Areas for Refactoring:**

*   **Configuration Complexity:** `configManager.ts` handles complex logic (loading, validation, normalization, cascading options). While functional, it could be a source of bugs and difficult to maintain.
*   **Error Handling:** Error types and handling are somewhat distributed (`ThinktankError`, provider-specific errors, `consoleUtils` helpers). Consolidation could improve clarity and consistency.
*   **Testing:** Coverage is moderate (50-60%). Some tests focus heavily on mock setup rather than behavior. E2E tests might be skipped if the build isn't present.
*   **Workflow Orchestration:** `runThinktank.ts` might be overly complex and could benefit from further decomposition.
*   **Code Consistency:** Minor inconsistencies in naming, style, or implementation details might exist.
*   **Resource Management:** The `BACKLOG.md` mentions the program hanging, suggesting potential issues with cleaning up resources (e.g., network connections) after execution.

## 3. Task Breakdown

| Task                                                     | Description                                                                                                                               | Affected Files/Modules                                                                 | Effort | Priority |
| :------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------- | :----- | :------- |
| **T1: CLI Entry Point Consolidation**                    | Verify if `src/cli/cli.ts` is redundant. If so, remove it and ensure `src/cli/index.ts` is the sole entry point using Commander.js.        | `src/cli/cli.ts`, `src/cli/index.ts`, `package.json` (bin entry)                       | S      | High     |
| **T2: Dependency Cleanup**                               | Investigate `yargs` usage. If unused, remove it from `package.json`. Update other dependencies if necessary.                              | `package.json`, potentially build/test scripts                                         | S      | High     |
| **T3: Error Handling Refinement**                        | Define a clear error hierarchy (e.g., `ConfigError`, `ProviderError`, `InputError` extending `ThinktankError`). Refactor error creation and handling across modules (`providers`, `workflow`, `configManager`, `cli`) to use these types consistently. Update `consoleUtils` error formatting. | `src/workflow/runThinktank.ts`, `src/core/*`, `src/providers/*`, `src/utils/consoleUtils.ts`, `src/cli/index.ts` | L      | High     |
| **T4: Configuration Logic Review**                       | Analyze `configManager.ts`. Simplify `loadConfig`, `normalizeConfig`, and `resolveModelOptions` if possible without losing functionality. Ensure Zod schemas are clear and maintainable. Improve error messages for config validation failures. | `src/core/configManager.ts`, `src/core/types.ts` (Zod schemas)                       | L      | Medium   |
| **T5: Provider Implementation Consistency**              | Review all provider implementations (`src/providers/*.ts`). Ensure strict adherence to `LLMProvider`, consistent error handling (using refined error types), and consistent option mapping (`mapOptions`). Check for resource cleanup (e.g., closing connections). | `src/providers/*.ts`, `src/core/types.ts`                                              | M      | Medium   |
| **T6: Workflow Orchestration (`runThinktank`) Refactor** | Break down the main `runThinktank` function into smaller, more focused steps (e.g., setup, select, execute, output). Improve clarity of the control flow and error propagation. Ensure proper resource cleanup. | `src/workflow/runThinktank.ts`                                                         | L      | Medium   |
| **T7: Utility Module Review**                            | Review `helpers.ts`, `fileReader.ts`, `outputFormatter.ts`, `logger.ts`, `nameGenerator.ts`. Consolidate duplicated logic, improve naming, ensure consistency. Remove unused helpers. | `src/utils/*`                                                                          | M      | Low      |
| **T8: Test Coverage and Quality Improvement**            | Increase unit test coverage for core logic, providers, and utils. Refactor tests focusing on behavior. Enhance integration tests for workflows. Ensure E2E tests cover main CLI commands and options. Fix skipped/incomplete tests. | `src/**/*.test.ts`, `src/**/*.e2e.test.ts`, `jest.config.js`                         | XL     | High     |
| **T9: Code Style and Readability Pass**                  | Run Prettier/ESLint fixes. Manually review code for clarity, naming conventions, and adherence to `BEST-PRACTICES.md`. Remove dead code or commented-out blocks. | Entire codebase (`src/`)                                                               | M      | Medium   |

**Effort Estimation:** S = Small (<1 day), M = Medium (1-3 days), L = Large (3-5 days), XL = Extra Large (5+ days)

## 4. Implementation Details

### T1: CLI Entry Point Consolidation ✅

*   **Goal:** Ensure `src/cli/index.ts` using `commander` is the single entry point.
*   **Action:**
    1.  ✅ Analyze `src/cli/cli.ts`. Determine if its simplified argument parsing logic is still needed or if it's fully superseded by the Commander setup in `src/cli/index.ts` and `src/cli/commands/`.
    2.  ✅ `cli.ts` was redundant:
        *   ✅ Deleted `src/cli/cli.ts`.
        *   ✅ Updated relevant test files that referenced the removed file.
        *   ✅ Verified `package.json` -> `bin` field already pointed to `dist/cli/index.js`.
    3.  ✅ Ensured `src/cli/index.ts` correctly imports and registers all commands from `src/cli/commands/`.
*   **Status:** Completed on 2025-04-05. CLI is now simplified with a single entry point using Commander.

### T2: Dependency Cleanup ✅

*   **Goal:** Remove unused dependencies, specifically `yargs`.
*   **Action:**
    1.  ✅ Search the codebase for any imports or usage of `yargs`.
    2.  ✅ If none are found, run `npm uninstall yargs` or `yarn remove yargs`.
    3.  ✅ Remove `yargs` and `@types/yargs` from `devDependencies` in `package.json`.
    4.  ✅ Run `npm install` or `yarn install` and ensure tests still pass.
*   **Status:** Completed on 2025-04-05. Verified that yargs is not used anywhere in the codebase.

### T3: Error Handling Refinement ⏳

*   **Goal:** Centralize error types and handling.
*   **Action:**
    1.  ✅ Define base `ThinktankError` in a central location - created `src/core/errors.ts` with the base class.
    2.  ✅ Define specific error types extending `ThinktankError` - implemented various subclasses and factory functions:
        ```typescript
        export class ThinktankError extends Error {
          category?: string;
          suggestions?: string[];
          examples?: string[];
          
          constructor(message: string, options?: ErrorOptions) {
            super(message);
            this.name = 'ThinktankError';
            
            if (options) {
              this.category = options.category;
              this.suggestions = options.suggestions;
              this.examples = options.examples;
            }
          }
          
          format(): string {
            // Formats error message with category, suggestions, and examples
          }
        }
        
        // Specialized error subclasses
        export class ConfigError extends ThinktankError { ... }
        export class ApiError extends ThinktankError { ... }
        export class FileSystemError extends ThinktankError { ... }
        // etc.
        
        // Factory functions
        export function createFileNotFoundError(filepath: string): ThinktankError { ... }
        export function createModelFormatError(model: string, ...): ThinktankError { ... }
        // etc.
        ```
    3.  ⏳ Refactor modules (`configManager`, `providers`, `inputHandler`, etc.) to throw these specific errors - in progress.
    4.  ⏳ Update `src/cli/index.ts` `handleError` function to recognize and format these errors appropriately - pending.
    5.  ✅ Refactor `consoleUtils` error creation functions to use new error system:
        * Updated `consoleUtils.ts` to properly handle ThinktankError instances in `formatError` and related functions
        * Added deprecated JSDoc tags to functions that should be replaced
        * Used ThinktankError.format() for error formatting
        * Updated tests to verify correct handling of ThinktankError instances
*   **Status:** Partially completed. Core error system implemented and `consoleUtils.ts` updated. Still need to update other modules to use the new error system.

### T6: Workflow Orchestration (`runThinktank`) Refactor

*   **Goal:** Improve readability and maintainability of the main workflow.
*   **Action:** Break down the `runThinktank` function:
    ```typescript
    // Pseudocode for refactored runThinktank
    async function runThinktank(options: RunOptions): Promise<string> {
      const spinner = ora(...);
      try {
        // Step 1: Setup & Initialization
        const config = await setupConfiguration(options, spinner);
        const friendlyRunName = await generateRunIdentifier(spinner);
        const inputResult = await processWorkflowInput(options, spinner);
        const outputDirectoryPath = await setupOutputDirectory(options, friendlyRunName, spinner);

        // Step 2: Model Selection
        const selectionResult = await selectWorkflowModels(config, options, spinner);
        if (selectionResult.models.length === 0) {
          // Handle no models selected case
          return handleNoModelsSelected(selectionResult, spinner);
        }
        displaySelectionInfo(selectionResult, options, spinner);

        // Step 3: Query Execution
        const queryResults = await executeModelQueries(config, selectionResult.models, inputResult, options, spinner);
        displayExecutionSummary(queryResults, friendlyRunName, spinner);

        // Step 4: Output Processing
        const fileOutputResult = await writeOutputFiles(queryResults.responses, outputDirectoryPath, options, spinner);
        displayFileOutputSummary(fileOutputResult, friendlyRunName, outputDirectoryPath, spinner);
        const consoleOutput = formatConsoleOutput(queryResults.responses, options);
        displayAdditionalMetadata(queryResults, fileOutputResult, options);

        return consoleOutput;

      } catch (error) {
        // Centralized error handling for the workflow
        return handleWorkflowError(error, spinner);
      }
    }

    // Define helper functions like setupConfiguration, selectWorkflowModels, etc.
    ```

## 5. Potential Challenges & Considerations

*   **Configuration Complexity:** Simplifying the cascading configuration logic (`resolveModelOptions`) without breaking existing behavior or expected overrides requires careful analysis and testing.
*   **Functionality Preservation:** Ensuring all CLI options, config interactions, and edge cases work exactly as before the refactor is critical and requires thorough testing.
*   **Testing Effort:** Achieving higher test coverage (e.g., 80%+) will require significant effort, especially for integration and E2E tests.
*   **Concurrent Development:** If new features are being developed simultaneously, merge conflicts could arise. Coordinate refactoring efforts with feature development.
*   **External SDK Changes:** Refactoring provider implementations might be affected by updates in the underlying LLM SDKs.
*   **Hanging Issue:** The refactoring should aim to improve resource management (e.g., closing SDK clients properly), but the root cause of the hanging issue mentioned in `BACKLOG.md` needs investigation. Ensure providers clean up connections. The `process.exit` calls in `cli.ts` might be masking underlying issues; removing `cli.ts` and ensuring graceful shutdown in `index.ts` is important.

## 6. Testing Strategy

*   **Baseline:** Run the existing test suite (`npm test`) before starting and ensure all tests pass. Record current coverage (`npm run test:cov`).
*   **Unit Tests:**
    *   Write new unit tests for any significantly refactored functions or modules (e.g., in `configManager`, `runThinktank` helpers, `providers`).
    *   Focus on testing logic, edge cases, and error handling.
    *   Improve existing tests to verify behavior rather than just mock interactions. Use Jest's `expect` matchers effectively.
    *   Target increased coverage for core modules (`src/core`, `src/workflow`, `src/utils`).
*   **Integration Tests:**
    *   Add integration tests for the `runThinktank` workflow, mocking the actual LLM API calls (`provider.generate`) but testing the interaction between `inputHandler`, `configManager`, `modelSelector`, `queryExecutor`, and `outputHandler`.
    *   Test different combinations of CLI options and configuration settings.
*   **End-to-End (E2E) Tests:**
    *   Review and enhance existing E2E tests (`src/cli/__tests__/cli.e2e.test.ts`).
    *   Ensure E2E tests cover:
        *   `thinktank run` with different inputs (file, group, specific model, multiple models).
        *   `thinktank config` subcommands (show, path, models add/remove/enable/disable, groups create/add-model/remove-model/set-prompt/remove).
        *   `thinktank models list` command.
        *   Key CLI options (`--config`, `--output`, `--thinking`, `--include-metadata`, etc.).
    *   Fix any issues causing E2E tests to be skipped (e.g., ensure build exists before running).
*   **Manual Testing:**
    *   Perform manual testing based on the examples in `README.md`.
    *   Test edge cases like invalid config files, missing API keys, non-existent files/groups.
    *   Verify console output formatting and file output structure.
*   **Coverage Goal:** Aim for >80% test coverage after refactoring. Use coverage reports (`coverage/lcov-report/index.html`) to identify gaps.
*   **Continuous Integration:** Ensure all tests run automatically in CI on each commit/PR.

## 7. Open Questions

1.  **Configuration Normalization:** Is the current behavior of `normalizeConfig` (especially regarding the default group) essential, or can it be simplified?

```
