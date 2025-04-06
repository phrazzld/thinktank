```markdown
# PLAN.md: Refactor Test Mocking Patterns

## 1. Overview

This plan outlines the steps to refactor the test suite by extracting common mocking patterns for the Node.js `fs/promises` module and the project's `gitignoreUtils` into shared test utilities. This addresses the "Test code duplication" issue identified in `CODE_REVIEW.md`, aiming to improve test code maintainability, reduce boilerplate, and ensure consistency in how dependencies are mocked across different test files.

The primary focus is on tests interacting with the file system and gitignore logic, as these were observed to have significant duplication in their mocking setup.

## 2. Task Breakdown

| # | Task Description | Effort | Affected Files/Modules | Notes |
|---|---|---|---|---|
| 1 | Create shared test utility directory | S | New: `src/__tests__/utils/` (or similar) | Establish a standard location for shared test helpers. |
| 2 | Develop `fs` Mock Utility | M | New: `src/__tests__/utils/mockFsUtils.ts` | Create functions to set up and configure `fs/promises` mocks. |
| 3 | Develop `gitignore` Mock Utility | S | New: `src/__tests__/utils/mockGitignoreUtils.ts` | Create functions to set up and configure `gitignoreUtils` mocks. |
| 4 | Refactor `src/utils/__tests__/fileReader.test.ts` | M | `src/utils/__tests__/fileReader.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks with the new utility. Handles `fs.access`, `fs.readFile`, `fs.writeFile`, `fs.mkdir`, `os`. |
| 5 | Refactor `src/utils/__tests__/binaryFileDetection.test.ts` | S | `src/utils/__tests__/binaryFileDetection.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks (`access`, `stat`). Logger mock can remain for now. |
| 6 | Refactor `src/utils/__tests__/fileSizeLimit.test.ts` | S | `src/utils/__tests__/fileSizeLimit.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks (`access`, `stat`, `readFile`). |
| 7 | Refactor `src/utils/__tests__/readContextFile.test.ts` | S | `src/utils/__tests__/readContextFile.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks (`access`, `readFile`, `stat`). Logger mock can remain. |
| 8 | Refactor `src/utils/__tests__/readContextPaths.test.ts` | S | `src/utils/__tests__/readContextPaths.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks (`access`, `stat`, `readdir`, `readFile`). |
| 9 | Refactor `src/utils/__tests__/readDirectoryContents.test.ts` | M | `src/utils/__tests__/readDirectoryContents.test.ts`, `src/__tests__/utils/mockFsUtils.ts`, `src/__tests__/utils/mockGitignoreUtils.ts` | Replace direct `fs` and `gitignoreUtils` mocks. |
| 10 | Refactor `src/utils/__tests__/gitignoreUtils.test.ts` | S | `src/utils/__tests__/gitignoreUtils.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks (`readFile`). `fileReader.fileExists` mock can remain. |
| 11 | Refactor `src/utils/__tests__/gitignoreFiltering.test.ts` | S | `src/utils/__tests__/gitignoreFiltering.test.ts`, `src/__tests__/utils/mockFsUtils.ts` | Replace direct `fs` mocks (`readFile`). `fileReader.fileExists` mock can remain. |
| 12 | Refactor `src/utils/__tests__/gitignoreFilterIntegration.test.ts` | M | `src/utils/__tests__/gitignoreFilterIntegration.test.ts`, `src/__tests__/utils/mockFsUtils.ts`, `src/__tests__/utils/mockGitignoreUtils.ts` | Replace direct `fs` and `gitignoreUtils` mocks. |
| 13 | Refactor `src/utils/__tests__/gitignoreFilteringIntegration.test.ts` | M | `src/utils/__tests__/gitignoreFilteringIntegration.test.ts`, `src/__tests__/utils/mockFsUtils.ts`, `src/__tests__/utils/mockGitignoreUtils.ts` | Replace direct `fs` and `gitignoreUtils` mocks. (Note: Duplicate filename in context, verify correct file). |
| 14 | Review and Refine | S | All refactored test files, utility files | Ensure consistency, remove unused old mocks, check for clarity. |
| 15 | Documentation (Optional) | S | README.md or CONTRIBUTING.md | Briefly document the new test utilities and their usage. |

**Effort Estimation:**
*   S = Small (<= 2 hours)
*   M = Medium (2-4 hours)
*   L = Large (> 4 hours)

## 3. Implementation Details

### 3.1. Test Utility Directory

*   Create a directory, e.g., `src/__tests__/utils/`. This location keeps test utilities close to the tests themselves.

### 3.2. `fs` Mock Utility (`mockFsUtils.ts`)

*   **Goal:** Provide functions to easily mock `fs/promises` methods.
*   **File:** `src/__tests__/utils/mockFsUtils.ts`
*   **Key Components:**
    *   Import `fs` and `jest`.
    *   Call `jest.mock('fs/promises');` at the top level.
    *   Export `const mockedFs = jest.mocked(fs);` for direct access if needed.
    *   **`setupMockFs(config?: FsMockConfig)`:**
        *   Optionally takes a configuration object to define default behaviors (e.g., all files exist, default read content).
        *   Resets mocks using `jest.clearAllMocks()` or specific mock resets.
        *   Sets up default mocks (e.g., `mockedFs.access.mockResolvedValue(undefined)`).
    *   **Helper Functions (Examples):**
        *   `mockAccess(path: string | RegExp, allowed: boolean)`: Configures `mockedFs.access` to resolve or reject for specific paths.
        *   `mockReadFile(path: string | RegExp, content: string | Error)`: Configures `mockedFs.readFile` to return content or throw an error.
        *   `mockStat(path: string | RegExp, stats: Partial<Stats> | Error)`: Configures `mockedFs.stat`. Allow passing partial `Stats` objects (e.g., `{ isFile: () => true, size: 100 }`).
        *   `mockReaddir(path: string | RegExp, entries: string[] | Error)`: Configures `mockedFs.readdir`.
        *   `mockMkdir(path: string | RegExp, success: boolean | Error)`: Configures `mockedFs.mkdir`.
        *   `mockWriteFile(path: string | RegExp, success: boolean | Error)`: Configures `mockedFs.writeFile`.
    *   **`resetMockFs()`:** Calls `jest.clearAllMocks()` or resets individual mocks on `mockedFs`.

*   **Example Usage (in a test file `beforeEach`):**
    ```typescript
    import { setupMockFs, mockReadFile, mockStat, resetMockFs, mockedFs } from './utils/mockFsUtils';

    beforeEach(() => {
      resetMockFs(); // Clear previous mocks
      setupMockFs(); // Setup default mocks (e.g., access resolves)

      // Configure specific mocks for this test suite
      mockReadFile('/path/to/file.txt', 'File content');
      mockStat('/path/to/file.txt', { isFile: () => true, size: 123 });
      mockStat('/path/to/dir', { isDirectory: () => true, isFile: () => false });
      // Can still use mockedFs directly for complex cases
      mockedFs.readdir.mockImplementation(async (p) => {
          if (String(p).includes('dir')) return ['file.txt'];
          throw new Error('ENOENT');
      });
    });
    ```

### 3.3. `gitignore` Mock Utility (`mockGitignoreUtils.ts`)

*   **Goal:** Provide functions to easily mock `gitignoreUtils`.
*   **File:** `src/__tests__/utils/mockGitignoreUtils.ts`
*   **Key Components:**
    *   Import `gitignoreUtils` and `jest`.
    *   Call `jest.mock('../gitignoreUtils');` at the top level (adjust path as needed).
    *   Export `const mockedGitignoreUtils = jest.mocked(gitignoreUtils);`
    *   **`setupMockGitignore(config?: GitignoreMockConfig)`:**
        *   Resets mocks.
        *   Sets up default behavior (e.g., `shouldIgnorePath` returns `false`).
    *   **Helper Functions:**
        *   `mockShouldIgnorePath(implementation: (basePath: string, filePath: string) => Promise<boolean> | boolean)`: Sets the implementation for `mockedGitignoreUtils.shouldIgnorePath`.
        *   `mockCreateIgnoreFilter(implementation: () => Promise<Ignore>)`: Sets the implementation for `mockedGitignoreUtils.createIgnoreFilter`.
    *   **`resetMockGitignore()`:** Calls `jest.clearAllMocks()` or resets individual mocks.

*   **Example Usage (in a test file `beforeEach`):**
    ```typescript
    import { setupMockGitignore, mockShouldIgnorePath, resetMockGitignore } from './utils/mockGitignoreUtils';

    beforeEach(() => {
      resetMockGitignore();
      setupMockGitignore(); // Setup default mocks (e.g., nothing ignored)

      // Configure specific mocks
      mockShouldIgnorePath(async (_base, file) => file.endsWith('.log'));
    });
    ```

### 3.4. Refactoring Test Files

*   Iterate through the files listed in the Task Breakdown.
*   In each file:
    *   Remove the direct `jest.mock('fs/promises')` and `jest.mock('../gitignoreUtils')` calls.
    *   Import the necessary setup/reset/helper functions from the new utility files.
    *   Modify the `beforeEach` (or `beforeAll`) blocks to use `setupMockFs`, `setupMockGitignore`, `resetMockFs`, `resetMockGitignore`.
    *   Replace direct mock configurations (e.g., `mockedFs.readFile.mockResolvedValue(...)`) with calls to the new helper functions (`mockReadFile(...)`) where appropriate and simpler.
    *   For complex mock implementations (like conditional logic in `stat` or `readdir`), continue using `mockedFs.stat.mockImplementation(...)` but ensure it's done *after* the `setupMockFs` call which might set defaults.
    *   Ensure `resetMockFs()` and `resetMockGitignore()` are called appropriately (likely in `beforeEach` or `afterEach`) to prevent mock state leaking between tests.

## 4. Potential Challenges & Considerations

1.  **Complex Mock Implementations:** Some tests have `fs.stat` or `fs.readdir` mocks that return different values based on the input path. The utility functions must either support this complexity directly or allow tests to easily override the default behavior using `mockedFs.<method>.mockImplementation`.
2.  **Mock Reset Strategy:** Ensure mocks are consistently reset between tests. Using `jest.clearAllMocks()` in `beforeEach` within the utilities might be sufficient, but verify it doesn't interfere with other mocks in the test files. Explicit `resetMockFs()` / `resetMockGitignore()` might be safer.
3.  **`os` Mocking:** `fileReader.test.ts` also mocks `os`. Decide if this should be part of the `mockFsUtils.ts` or a separate utility, or left as is if it's not widely duplicated. For now, leave as is.
4.  **`logger` Mocking:** Similar to `os`, logger mocking is present but less pervasive than `fs`. Keep logger mocks separate for this refactor unless significant duplication is found later.
5.  **Duplicate Test File:** The context lists both `src/utils/__tests__/gitignoreFilterIntegration.test.ts` and `src/utils/__tests__/gitignoreFilteringIntegration.test.ts`. Verify the correct file name and path. Assume both need refactoring for now.
6.  **Utility File Location:** Confirm the best location (`src/__tests__/utils/` vs. a top-level `tests/utils/`). `src/__tests__/utils/` seems conventional for utilities tied closely to the source tests.

## 5. Testing Strategy

1.  **Unit Tests (for Utilities):** Optionally, write simple unit tests for the `mockFsUtils.ts` and `mockGitignoreUtils.ts` helpers to ensure they configure the underlying Jest mocks correctly.
2.  **Existing Test Suite:** The primary validation method is ensuring that **all existing tests pass** after the refactoring. The goal is to change the *implementation* of the mocking setup without changing the *behavior* of the tests.
3.  **Code Review:** Review the refactored test files to confirm:
    *   Duplication has been significantly reduced.
    *   The new utilities are used correctly and consistently.
    *   Test readability is maintained or improved.
    *   No test logic has been inadvertently changed.
4.  **Coverage:** Ensure test coverage remains at least the same after refactoring.

## 6. Open Questions

1.  **File Naming:** What is the correct filename/path for the potentially duplicated `gitignoreFilter(ing)Integration.test.ts`?
2.  **Scope:** Should mocks for `os` or `logger` be included in this refactoring effort, or deferred? (Current plan defers them).
3.  **Utility Location:** Is `src/__tests__/utils/` the preferred location for these shared test utilities?
```