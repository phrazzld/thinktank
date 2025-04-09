# PLAN

This plan outlines the technical steps required to refactor the `thinktank` codebase to enhance testability, improve maintainability, and fully adhere to the project's established guidelines (`BEST_PRACTICES.md`, `CONTRIBUTING.md`, `TESTING_PHILOSOPHY.md`), with a strong focus on minimizing mocks.

**Phase 1: Complete Filesystem Test Refactoring (Eliminate `fs` Mocking)**

* **Objective:** Ensure all tests involving filesystem interactions use the `memfs`-based virtual filesystem (`test/setup/fs.ts`, `virtualFsUtils.ts`) instead of mocking the `fs` module directly or using legacy mocks (`mockFsUtils.ts`).
* **Tasks:**
    1.  **Identify & Migrate Tests:** Systematically review all `*.test.ts` files[cite: 299]. Migrate any test still mocking `fs` or using `mockFsUtils.ts` to use `setupBasicFs` or `createVirtualFs` helpers[cite: 302]. Address skipped tests listed in `BACKLOG.md` (`readDirectoryContents.test.ts`, `virtualFsUtils.test.ts`).
    2.  **Refactor `ConcreteFileSystem` Test:** Rewrite `src/core/__tests__/FileSystem.test.ts` to use `memfs` via `test/setup/fs.ts`[cite: 90, 138]. Focus on testing the *behavior* (reading/writing to the virtual FS, correct error wrapping) rather than mocking internal `fileReader` calls[cite: 91, 103, 139, 151].
    3.  **Standardize Path Handling:** Ensure all tests use `normalizePathGeneral` or `normalizePathForMemfs` consistently for paths interacting with the virtual filesystem to guarantee cross-platform compatibility[cite: 11].
    4.  **Remove Legacy Utils:** Once all migrations are complete, delete `src/__tests__/utils/mockFsUtils.ts` [cite: 298, 303] and any related helpers. Update `jest/README.md` [cite: 96] and `src/__tests__/utils/README.md` [cite: 96, 144] to deprecate old patterns.

**Phase 2: Refactor Gitignore Testing (Test Implementation, Not Mocks)**

* **Objective:** Test the actual `gitignoreUtils` implementation against virtual `.gitignore` files using `memfs`. Eliminate mocking of `gitignoreUtils` itself[cite: 308].
* **Tasks:**
    1.  **Verify `memfs` Gitignore Support:** Ensure `test/setup/fs.ts` or `virtualFsUtils.ts` correctly handles creating and reading `.gitignore` files (including hidden file handling)[cite: 308]. The `addVirtualGitignoreFile` helper seems appropriate.
    2.  **Migrate Gitignore Tests:** Refactor tests listed in `BACKLOG.md` (`gitignoreFiltering.test.ts`, `gitignoreFilterIntegration.test.ts`) [cite: 9] and potentially `readContextPaths.test.ts`. Remove mocks (`jest.mock('../gitignoreUtils')`, `mockGitignoreUtils`). Use `setupWithGitignore` or `setupMultiGitignore` helpers to create virtual `.gitignore` files within `memfs` setups[cite: 309]. Assert the behavior of `shouldIgnorePath` and `createIgnoreFilter` against these virtual files[cite: 310, 311].
    3.  **Address Complex Patterns:** Investigate and address limitations with complex gitignore patterns (e.g., brace expansion, prefix wildcards) noted in `gitignoreComplexPatterns.test.ts`[cite: 10]. Either fix the implementation/testing approach or clearly document the limitations.
    4.  **Remove Mocks:** Delete `mockGitignoreUtils.ts` once it's no longer used[cite: 311].

**Phase 3: Implement Dependency Injection (Decouple Components)**

* **Objective:** Introduce and consistently apply Dependency Injection (DI) throughout the application, especially for external dependencies (FileSystem, LLM APIs, Console/UI), to facilitate easier testing with minimal mocking[cite: 111, 120, 312].
* **Tasks:**
    1.  **Define Core Interfaces:** Formalize interfaces for key dependencies in `src/core/interfaces.ts`: `FileSystem`[cite: 18, 315], `LLMClient`[cite: 18, 313], `ConfigManagerInterface`[cite: 18, 314], `ConsoleLogger`[cite: 86, 102, 150], `UISpinner`[cite: 86, 102, 150]. Ensure these cover all necessary interactions.
    2.  **Create Concrete Implementations:** Ensure concrete implementations exist for each interface (e.g., `ConcreteFileSystem`[cite: 90], `ConcreteConfigManager`[cite: 1649], `ConsoleAdapter`[cite: 1684], provider implementations for `LLMClient`).
    3.  **Refactor Workflow Modules:** Modify `runThinktank.ts` [cite: 19, 141] and its helpers (`runThinktankHelpers.ts`) [cite: 312] to accept instances of these interfaces via parameters or a context object, instead of directly importing or instantiating concrete classes/modules[cite: 19, 111, 312]. For example, `_executeQueries` should accept an `LLMClient` instance[cite: 316].
    4.  **Refactor CLI Handlers:** Extract command logic from `commander` `.action()` callbacks into separate testable functions/classes that accept dependencies (like ConfigManager, LLMClient) via parameters[cite: 21, 322]. The CLI setup file should then only be responsible for parsing args and calling these handlers with the necessary dependencies[cite: 323, 340].
    5.  **Update Tests for DI:** Modify unit/integration tests to inject mock implementations (using `test/setup/` helpers like `createMockFileSystem`[cite: 85], `createMockConsoleLogger`[cite: 85], `createMockLlmClient`) instead of using `jest.mock` for entire modules[cite: 16]. Test the interaction *with* the injected mocks[cite: 317, 86, 134].

**Phase 4: Isolate Side Effects (Pure Logic vs. I/O)**

* **Objective:** Ensure a clean separation between pure data transformation logic and functions that perform I/O operations (file writing, console logging, API calls)[cite: 318, 107, 155]. Refactor functions to return data rather than performing side effects directly where appropriate.
* **Tasks:**
    1.  **Refactor `_processOutput`:** Ensure `_processOutput` (or equivalent logic in `outputHandler.ts`) returns structured file data (`FileData[]`) and console output strings, rather than writing files or logging directly[cite: 20, 319, 320].
    2.  **Refactor `_logCompletionSummary`:** Convert `_logCompletionSummary` logic (now likely inside `runThinktank.ts` or `outputHandler.ts`) into a pure function (`formatCompletionSummary`) that returns a formatted string or object containing the summary text and error details[cite: 20, 318].
    3.  **Centralize I/O:** Move the actual file writing (`fileSystem.writeFile`) and console logging (`consoleLogger.log`, etc.) calls to the main `runThinktank` orchestration function or a dedicated `src/workflow/io.ts` module, using the data returned by the pure functions[cite: 321].
    4.  **Consolidate File Writing:** Eliminate the duplicated file writing logic identified in `CODE_REVIEW.md`[cite: 89, 137]. Use a single, centralized function (likely in `src/workflow/io.ts`) that takes `FileData[]` and handles writing[cite: 108, 156]. Add tests for the new `io.ts` module[cite: 92, 140, 104, 152].
    5.  **Test Pure Functions:** Write unit tests for the refactored pure functions, verifying their output based on input data without needing I/O mocks[cite: 86, 133].
    6.  **Test I/O Interactions:** Write integration tests for the I/O functions (e.g., in `src/workflow/__tests__/io.test.ts`), injecting mock `FileSystem`, `ConsoleLogger`, etc., and verifying the correct methods were called on the mocks[cite: 92, 140].

**Phase 5: Refine and Finalize**

* **Objective:** Clean up remaining issues, update documentation, and ensure consistency.
* **Tasks:**
    1.  **Address Minor Issues:** Fix remaining issues from `CODE_REVIEW.md` and `BACKLOG.md` (e.g., repetitive error wrapping in `ConcreteFileSystem`[cite: 95, 143], inconsistent logger usage[cite: 93, 141], skipped tests[cite: 9], documentation links[cite: 47], other backlog items).
    2.  **Review Test Coverage:** Run code coverage (`pnpm test:cov`) [cite: 482, 350] and address significant gaps, focusing on high-value behavioral tests[cite: 17, 333].
    3.  **Update Documentation:** Update `TESTING.md`[cite: 48, 335], `CONTRIBUTING.md`[cite: 48, 110, 335], and any other relevant documentation to reflect the new DI architecture and `memfs`-based testing strategy. Ensure consistency with `TESTING_PHILOSOPHY.md`.
    4.  **Code Cleanup:** Remove dead code, unused mocks, and ensure adherence to linting/formatting rules (`pnpm run lint:fix`[cite: 73], `pnpm run format` [cite: 73, 114]). Ensure final newlines are present (`pnpm run fix:newlines` [cite: 73, 114]).

By following this plan, `thinktank` will become significantly more testable and maintainable, closely aligning with its excellent development guidelines and philosophy.
