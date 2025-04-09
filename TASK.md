# TASK: Create a Detailed Implementation Plan

## OBJECTIVES
1. Please review the current TODO.md and PLAN.md files
2. Create an extremely detailed, thorough, and comprehensive task breakdown that implements every point from PLAN.md
3. For each task, provide:
   - Explicit implementation steps with technical details
   - Complete file paths and function names
   - Specific code patterns to follow
   - Testing strategies with concrete examples
   - Clear success criteria

## GUIDANCE
- Make the task breakdown as granular and actionable as possible
- Include specific file paths, function names, and implementation details
- Provide concrete examples of code changes where helpful
- Ensure every acceptance criterion from PLAN.md is fully addressed
- Identify any technical dependencies, edge cases, or potential issues
- Maintain the dependency tracking and task relationships

The goal is to transform the TODO.md into an extremely precise implementation blueprint that leaves no ambiguity about what needs to be done.

## CURRENT TODO.md:

# TODO

## Phase 1: Complete Filesystem Test Refactoring

- [ ] **Review and migrate legacy FS test implementations**
  - **Action:** Identify all test files still using direct `fs` mocking or `mockFsUtils.ts` and migrate them to use `setupBasicFs` or `createVirtualFs` helpers.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1.

- [ ] **Migrate skipped tests to use virtual filesystem**
  - **Action:** Address skipped tests in `readDirectoryContents.test.ts` and `virtualFsUtils.test.ts` by implementing proper virtual filesystem testing.
  - **Depends On:** Review and migrate legacy FS test implementations.
  - **AC Ref:** AC 1.1.

- [ ] **Refactor ConcreteFileSystem test**
  - **Action:** Rewrite `src/core/__tests__/FileSystem.test.ts` to use `memfs` via `test/setup/fs.ts` instead of mocking internal `fileReader` calls.
  - **Depends On:** None.
  - **AC Ref:** AC 1.2.

- [ ] **Standardize path normalization in tests**
  - **Action:** Ensure all tests consistently use `normalizePathGeneral` or `normalizePathForMemfs` for path handling to guarantee cross-platform compatibility.
  - **Depends On:** None.
  - **AC Ref:** AC 1.3.

- [ ] **Remove mockFsUtils.ts after migration**
  - **Action:** Once all tests have been migrated, delete the legacy `src/__tests__/utils/mockFsUtils.ts` file.
  - **Depends On:** Review and migrate legacy FS test implementations, Migrate skipped tests to use virtual filesystem, Refactor ConcreteFileSystem test.
  - **AC Ref:** AC 1.4.

- [ ] **Update test README files**
  - **Action:** Update `jest/README.md` and `src/__tests__/utils/README.md` to reflect new testing patterns and deprecate old ones.
  - **Depends On:** Remove mockFsUtils.ts after migration.
  - **AC Ref:** AC 1.4.

## Phase 2: Refactor Gitignore Testing

- [ ] **Verify memfs gitignore support**
  - **Action:** Ensure `test/setup/fs.ts` or `virtualFsUtils.ts` correctly handles creating and reading `.gitignore` files, confirming the `addVirtualGitignoreFile` helper works as expected.
  - **Depends On:** None.
  - **AC Ref:** AC 2.1.

- [ ] **Refactor gitignore filtering tests**
  - **Action:** Refactor `gitignoreFiltering.test.ts` and `gitignoreFilterIntegration.test.ts` to use virtual `.gitignore` files with `setupWithGitignore` or `setupMultiGitignore` helpers instead of mocking.
  - **Depends On:** Verify memfs gitignore support.
  - **AC Ref:** AC 2.2.

- [ ] **Refactor readContextPaths tests**
  - **Action:** Update `readContextPaths.test.ts` to use virtual `.gitignore` files instead of mocks, ensuring it tests real behavior with the virtual filesystem.
  - **Depends On:** Refactor gitignore filtering tests.
  - **AC Ref:** AC 2.2.

- [ ] **Address complex gitignore pattern handling**
  - **Action:** Investigate issues with complex gitignore patterns in `gitignoreComplexPatterns.test.ts` and either fix implementation or document limitations clearly.
  - **Depends On:** Refactor gitignore filtering tests.
  - **AC Ref:** AC 2.3.

- [ ] **Remove mockGitignoreUtils.ts**
  - **Action:** Delete `mockGitignoreUtils.ts` once all gitignore tests have been refactored to use the virtual filesystem approach.
  - **Depends On:** Refactor gitignore filtering tests, Refactor readContextPaths tests.
  - **AC Ref:** AC 2.4.

## Phase 3: Implement Dependency Injection

- [ ] **Define and refine core interfaces**
  - **Action:** Ensure interfaces in `src/core/interfaces.ts` are complete and well-documented for all key dependencies: `FileSystem`, `LLMClient`, `ConfigManagerInterface`, `ConsoleLogger`, and `UISpinner`.
  - **Depends On:** None.
  - **AC Ref:** AC 3.1.

- [ ] **Implement concrete adapter classes**
  - **Action:** Ensure all interfaces have corresponding concrete implementations (e.g., `ConcreteFileSystem`, `ConcreteConfigManager`, `ConsoleAdapter`, etc.).
  - **Depends On:** Define and refine core interfaces.
  - **AC Ref:** AC 3.2.

- [ ] **Refactor workflow modules to use DI**
  - **Action:** Modify `runThinktank.ts` and helpers to accept injected dependencies via parameters/context objects rather than direct imports.
  - **Depends On:** Define and refine core interfaces, Implement concrete adapter classes.
  - **AC Ref:** AC 3.3.

- [ ] **Refactor _executeQueries for DI**
  - **Action:** Update `_executeQueries` to accept an `LLMClient` instance rather than using direct imports.
  - **Depends On:** Refactor workflow modules to use DI.
  - **AC Ref:** AC 3.3.

- [ ] **Extract CLI logic from commander callbacks**
  - **Action:** Move logic from commander `.action()` callbacks to separate testable functions that accept dependencies via parameters.
  - **Depends On:** Define and refine core interfaces.
  - **AC Ref:** AC 3.4.

- [ ] **Restructure CLI module for DI**
  - **Action:** Restructure CLI setup to only parse args and call handler functions with injected dependencies.
  - **Depends On:** Extract CLI logic from commander callbacks.
  - **AC Ref:** AC 3.4.

- [ ] **Update tests to use DI patterns**
  - **Action:** Modify tests to inject mock implementations from `test/setup/` helpers instead of using `jest.mock` for entire modules.
  - **Depends On:** Refactor workflow modules to use DI, Refactor CLI module for DI.
  - **AC Ref:** AC 3.5.

## Phase 4: Isolate Side Effects

- [ ] **Refactor _processOutput to be pure**
  - **Action:** Modify `_processOutput` or equivalent in `outputHandler.ts` to return structured data (`FileData[]`) instead of performing I/O directly.
  - **Depends On:** Refactor workflow modules to use DI.
  - **AC Ref:** AC 4.1.

- [ ] **Refactor _logCompletionSummary to be pure**
  - **Action:** Convert `_logCompletionSummary` to a pure function that returns formatted data instead of logging directly.
  - **Depends On:** None.
  - **AC Ref:** AC 4.2.

- [ ] **Create formatCompletionSummary utility**
  - **Action:** Create a new utility function that handles formatting completion summary data without side effects.
  - **Depends On:** Refactor _logCompletionSummary to be pure.
  - **AC Ref:** AC 4.2.

- [ ] **Centralize I/O operations**
  - **Action:** Move actual file writing and console logging to `runThinktank` or a dedicated `io.ts` module.
  - **Depends On:** Refactor _processOutput to be pure, Refactor _logCompletionSummary to be pure.
  - **AC Ref:** AC 4.3.

- [ ] **Implement centralized file writing function**
  - **Action:** Create a single, centralized function for file writing that accepts `FileData[]` and handles the actual file system interactions.
  - **Depends On:** Centralize I/O operations.
  - **AC Ref:** AC 4.4.

- [ ] **Write tests for pure business logic**
  - **Action:** Create unit tests for the refactored pure functions, focusing on verifying their output based on input without mocking I/O.
  - **Depends On:** Refactor _processOutput to be pure, Create formatCompletionSummary utility, Implement centralized file writing function.
  - **AC Ref:** AC 4.5.

- [ ] **Create io.ts module tests**
  - **Action:** Write integration tests for the I/O functions, injecting mock dependencies and verifying the correct method calls.
  - **Depends On:** Centralize I/O operations, Implement centralized file writing function.
  - **AC Ref:** AC 4.6.

## Phase 5: Refine and Finalize

- [ ] **Fix repetitive error wrapping in ConcreteFileSystem**
  - **Action:** Refactor error handling in `ConcreteFileSystem` to reduce repetition and improve consistency.
  - **Depends On:** None.
  - **AC Ref:** AC 5.1.

- [ ] **Standardize logger usage**
  - **Action:** Ensure consistent usage of the logger across the codebase, replacing direct console calls with the injected logger.
  - **Depends On:** Refactor workflow modules to use DI.
  - **AC Ref:** AC 5.1.

- [ ] **Fix broken documentation links**
  - **Action:** Correct any broken links in documentation, especially in README.md.
  - **Depends On:** None.
  - **AC Ref:** AC 5.1.

- [ ] **Review and fix remaining tests**
  - **Action:** Address any remaining skipped or failing tests not covered in previous phases.
  - **Depends On:** Complete all Phase 1-4 test-related tasks.
  - **AC Ref:** AC 5.1.

- [ ] **Run and analyze code coverage**
  - **Action:** Execute `pnpm test:cov` and identify significant coverage gaps that need addressing.
  - **Depends On:** Review and fix remaining tests.
  - **AC Ref:** AC 5.2.

- [ ] **Add missing high-value tests**
  - **Action:** Create additional tests for any identified coverage gaps, focusing on behavior rather than implementation details.
  - **Depends On:** Run and analyze code coverage.
  - **AC Ref:** AC 5.2.

- [ ] **Create/update TESTING.md**
  - **Action:** Create or update `TESTING.md` to reflect the new DI architecture and memfs-based testing strategy.
  - **Depends On:** Complete all Phase 1-4 tasks.
  - **AC Ref:** AC 5.3.

- [ ] **Update CONTRIBUTING.md**
  - **Action:** Ensure `CONTRIBUTING.md` is updated to match the new testing approach and project standards.
  - **Depends On:** Create/update TESTING.md.
  - **AC Ref:** AC 5.3.

- [ ] **Remove dead code and unused mocks**
  - **Action:** Identify and eliminate any dead code or unused mock implementations throughout the codebase.
  - **Depends On:** Complete all Phase 1-4 tasks.
  - **AC Ref:** AC 5.4.

- [ ] **Run linting and formatting**
  - **Action:** Execute `pnpm run lint:fix`, `pnpm run format`, and `pnpm run fix:newlines` to ensure code style compliance.
  - **Depends On:** Remove dead code and unused mocks.
  - **AC Ref:** AC 5.4.

- [ ] **Final verification**
  - **Action:** Run a complete test suite and manual verification to ensure all refactoring is working properly.
  - **Depends On:** All previous tasks.
  - **AC Ref:** All ACs.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Multiple Filesystem Testing Approaches**: There appear to be several related but different approaches to virtual filesystem testing (`setupBasicFs`, `createVirtualFs`, etc.).
  - **Context:** Phase 1 mentions using helpers like `setupBasicFs` or `createVirtualFs`, but it's not clear if one is preferred over the other or if they serve different purposes.
  - **Assumption:** We'll use `setupBasicFs` for simpler tests and `createVirtualFs` for more complex scenarios, keeping both as valid approaches.

- [ ] **Gitignore Helper Functions**: The plan mentions `setupWithGitignore` and `setupMultiGitignore` helpers, but their implementation is not shown.
  - **Context:** Phase 2 refers to using these helpers for migrating gitignore tests, but they don't appear in the reviewed files.
  - **Assumption:** These helpers need to be implemented as part of the task, building on the existing `addVirtualGitignoreFile` function.

- [ ] **CLI Handler Structure**: The specific structure for refactored CLI handlers is not detailed.
  - **Context:** Phase 3 mentions extracting logic from commander callbacks, but doesn't specify the exact pattern to follow.
  - **Assumption:** We'll create a separate module with handler functions that accept dependencies through parameters and return appropriate values/promises.

- [ ] **I/O Module Structure**: The exact structure of the proposed `io.ts` module is not specified.
  - **Context:** Phase 4 suggests creating a dedicated module for I/O operations, but details of its interface are not provided.
  - **Assumption:** The module will export functions for file writing and console output that accept the necessary data and dependencies (FileSystem, ConsoleLogger) as parameters.

- [ ] **FileData Structure**: The `FileData[]` structure is mentioned but not defined.
  - **Context:** Phase 4 refers to refactoring to return `FileData[]` objects, but their exact structure is not specified.
  - **Assumption:** This is likely an existing type that represents a file path and its content, or will need to be defined as part of the refactoring.

## PLAN.md CONTENT:

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
