# TODO

This document provides a comprehensive and detailed implementation plan for refactoring the `thinktank` codebase based on the technical roadmap outlined in `PLAN.md`. Each task includes specific implementation steps, technical details, file paths, and clear success criteria.

## Phase 1: Complete Filesystem Test Refactoring

- [x] **Audit and catalog all tests using deprecated FS mocking**
  - **Action:** Systematically search through all test files to identify those using direct `fs` mocking or `mockFsUtils.ts`.
  - **Technical Details:**
    - Created an automated script (`scripts/audit-fs-mocks.js`) that scans test files for patterns.
    - Implemented regular expression patterns to identify: direct fs mocking, legacy mockFsUtils usage, and virtual FS usage.
    - Generated report at `FS_MOCK_AUDIT.md` with prioritization table and migration guidelines.
    - Found 17 test files using deprecated or mixed mocking approaches.
  - **Status:** Completed. Audit script created and executed successfully with full test coverage.
  - **Success Criteria:** ✅ Complete inventory of tests needing migration with assessment structure provided.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1.

- [x] **Implement helpers for standard FS test scenarios**
  - **Action:** Create or enhance test helpers in `test/setup/fs.ts` for common filesystem test scenarios.
  - **Technical Details:**
    - Added `setupWithFiles` helper that accepts a directory structure object.
    - Added `setupWithSingleFile` helper for the common single-file test case.
    - Added `setupWithNestedDirectories` helper for testing deep directory structures.
    - Ensured all helpers use `normalizePathForMemfs` consistently via `createVirtualFs`.
    - Added comprehensive tests for all new helpers in `test/setup/__tests__/fs.test.ts`.
  - **Files Modified:**
    - Enhanced `test/setup/fs.ts` with new helper functions and documentation.
    - Added new tests in `test/setup/__tests__/fs.test.ts`.
  - **Success Criteria:** ✅ Completed set of helpers covering common FS test scenarios with documentation.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1.

- [x] **Migrate fileReader.test.ts to use virtual filesystem**
  - **Action:** Refactor `src/utils/__tests__/fileReader.test.ts` to use `memfs` via the `setupBasicFs` helper.
  - **Technical Details:**
    - Removed `jest.mock('fs')` and direct mock implementations.
    - Replaced setup code with `setupBasicFs()` calls to configure test files.
    - Used `normalizePathForMemfs()` to ensure cross-platform path compatibility.
    - Restructured tests with better describe/beforeEach organization.
    - Fixed platform-specific test assumptions.
    - Updated jest.config.js to re-enable previously skipped test.
  - **Implementation Example:**
    ```typescript
    // Before:
    jest.mock('fs', () => ({
      promises: {
        readFile: jest.fn().mockResolvedValue('file content')
      }
    }));
    
    // After:
    import { setupBasicFs } from '../../../test/setup/fs';
    
    beforeEach(() => {
      setupBasicFs({
        ['/path/to/file.txt']: 'file content'
      });
    });
    ```
  - **Success Criteria:** ✅ All tests in `fileReader.test.ts` pass using virtual filesystem.
  - **Depends On:** Implement helpers for standard FS test scenarios.
  - **AC Ref:** AC 1.1.

- [x] **Migrate readDirectoryContents.test.ts**
  - **Action:** Reimplement the skipped tests in `src/utils/__tests__/readDirectoryContents.test.ts` using virtual filesystem.
  - **Technical Details:**
    - Remove any existing `jest.mock('fs')` and import the virtual filesystem helpers.
    - Set up test directory structures using `setupBasicFs`.
    - Use `normalizePathForMemfs` for all paths.
    - Focus on testing directory traversal behavior with different directory structures.
    - Implement tests for error handling scenarios (permission denied, not a directory).
  - **Success Criteria:** ✅ All tests in file pass, including previously skipped tests.
  - **Depends On:** Implement helpers for standard FS test scenarios.
  - **AC Ref:** AC 1.1.

- [x] **Migrate virtualFsUtils.test.ts**
  - **Action:** Complete the implementation of tests for `virtualFsUtils.ts` itself.
  - **Technical Details:**
    - Properly implement tests for `normalizePathForMemfs` with various path formats.
    - Test `createVirtualFs` with complex nested directory structures.
    - Test `addVirtualGitignoreFile` functionality.
    - Ensure reset and volume management functions work correctly.
    - Added tests for `createFsError`, `createMockStats`, and `createMockDirent`.
  - **Success Criteria:** ✅ All tests in `virtualFsUtils.test.ts` pass and provide good coverage.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1.

- [x] **Refactor FileSystem.test.ts for behavior testing**
  - **Action:** Rewrite `src/core/__tests__/FileSystem.test.ts` to test behavior using the virtual filesystem.
  - **Technical Details:**
    - Remove mocking of internal `fileReader` calls.
    - Use `test/setup/fs.ts` helpers to set up test files in the virtual filesystem.
    - Test actual file reading/writing on the virtual filesystem.
    - Test error handling by creating error scenarios in the virtual filesystem.
    - Verify that errors are properly wrapped in `FileSystemError` instances.
    - Use `createFsError` for expected error creation.
  - **Implementation Example:**
    ```typescript
    it('should wrap fs errors in FileSystemError', async () => {
      // Create a directory where we expect a file - will cause EISDIR error
      setupBasicFs({
        [normalizePathForMemfs('/test/dir')]: '' // Empty content = directory
      });
      
      const fs = new ConcreteFileSystem();
      
      await expect(
        fs.readFileContent('/test/dir')
      ).rejects.toThrow(FileSystemError);
      
      await expect(
        fs.readFileContent('/test/dir')
      ).rejects.toMatchObject({
        code: 'EISDIR',
        originalError: expect.any(Error)
      });
    });
    ```
  - **Success Criteria:** `FileSystem.test.ts` tests the behavior of `ConcreteFileSystem` against `memfs`, not internal implementation details.
  - **Depends On:** None.
  - **AC Ref:** AC 1.2.

- [x] **Audit and standardize path normalization across tests**
  - **Action:** Ensure all tests consistently use the correct path normalization functions.
  - **Technical Details:**
    - Search for string paths in test files, especially Windows-style paths (`\\`).
    - Replace direct path strings with calls to the appropriate normalizer:
      - Use `normalizePathForMemfs` for paths used with the virtual filesystem.
      - Use `normalizePathGeneral` for general path normalization needs.
    - Document the normalization strategy in test comments.
  - **Success Criteria:** All tests use normalized paths and work cross-platform.
  - **Depends On:** None.
  - **AC Ref:** AC 1.3.

- [x] **Create path normalization testing guide**
  - **Action:** Document the path normalization strategy for tests in READMEs.
  - **Technical Details:**
    - Add a section to `jest/README.md` explaining when to use each normalizer.
    - Add examples of correct path handling for Windows and Unix paths.
    - Document how to test with cross-platform path separators.
  - **Success Criteria:** Clear documentation on path normalization for testing.
  - **Depends On:** Audit and standardize path normalization across tests.
  - **AC Ref:** AC 1.3.

- [x] **Delete mockFsUtils.ts**
  - **Action:** Remove the legacy `src/__tests__/utils/mockFsUtils.ts` file after all migrations.
  - **Technical Details:**
    - Verify all tests that used it have been migrated.
    - Remove the file and any related utility functions.
    - Update imports in any files that might still reference it.
  - **Success Criteria:** File is removed and no references to it remain in the codebase.
  - **Depends On:** Migrate all tests using legacy FS mocking.
  - **AC Ref:** AC 1.4.

- [ ] **Update test README files with new patterns**
  - **Action:** Update `jest/README.md` and `src/__tests__/utils/README.md` to document the virtual filesystem testing approach.
  - **Technical Details:**
    - Add detailed explanation of the `memfs`-based testing strategy.
    - Include code examples for common testing scenarios.
    - Document available helpers (`setupBasicFs`, `createVirtualFs`, etc.)
    - Add deprecation notices for old mocking approaches.
    - Link to canonical examples in the codebase.
  - **Success Criteria:** Complete, up-to-date documentation of the testing approach.
  - **Depends On:** Delete mockFsUtils.ts.
  - **AC Ref:** AC 1.4.

## Phase 2: Refactor Gitignore Testing

- [ ] **Test addVirtualGitignoreFile functionality**
  - **Action:** Verify that `addVirtualGitignoreFile` correctly creates .gitignore files in the virtual filesystem.
  - **Technical Details:**
    - Create tests that set up a virtual filesystem and add gitignore files.
    - Verify file creation and content through direct virtual filesystem access.
    - Test handling of different path formats and nested directories.
    - Test both with and without trailing slashes in directory paths.
  - **Success Criteria:** Confirmed functionality of `addVirtualGitignoreFile` helper.
  - **Depends On:** None.
  - **AC Ref:** AC 2.1.

- [ ] **Implement gitignore test helpers**
  - **Action:** Create helper functions for testing gitignore functionality.
  - **Technical Details:**
    - Implement `setupWithGitignore` helper in `test/setup/gitignore.ts` that:
      - Sets up a virtual filesystem with specific files and directories.
      - Creates a .gitignore file with provided patterns.
    - Implement `setupMultiGitignore` for testing nested .gitignore files.
    - Add helper to create test file structures that match/don't match patterns.
  - **Implementation Example:**
    ```typescript
    // In test/setup/gitignore.ts
    export async function setupWithGitignore(
      rootPath: string, 
      gitignoreContent: string,
      files: Record<string, string>
    ): Promise<void> {
      // Set up base filesystem
      setupBasicFs(files);
      
      // Create .gitignore file
      const gitignorePath = path.join(rootPath, '.gitignore');
      await addVirtualGitignoreFile(
        normalizePathForMemfs(gitignorePath), 
        gitignoreContent
      );
    }
    ```
  - **Success Criteria:** Complete set of helper functions for gitignore testing.
  - **Depends On:** Test addVirtualGitignoreFile functionality.
  - **AC Ref:** AC 2.1.

- [ ] **Refactor gitignoreFiltering.test.ts**
  - **Action:** Rewrite `gitignoreFiltering.test.ts` to use the virtual filesystem and actual gitignore files.
  - **Technical Details:**
    - Remove all mocks of `gitignoreUtils`.
    - Use the new `setupWithGitignore` helper to create test scenarios.
    - Test `shouldIgnorePath` and `createIgnoreFilter` with actual gitignore patterns.
    - Test handling of nested .gitignore files.
    - Test with various gitignore pattern types (literals, wildcards, negations).
  - **Success Criteria:** Tests verify actual behavior against virtual .gitignore files.
  - **Depends On:** Implement gitignore test helpers.
  - **AC Ref:** AC 2.2.

- [ ] **Refactor gitignoreFilterIntegration.test.ts**
  - **Action:** Rewrite integration tests for gitignore filtering using virtual filesystem.
  - **Technical Details:**
    - Replace mocks with virtual filesystem setup.
    - Create complex directory structures with multiple .gitignore files.
    - Test actual filtering of paths based on gitignore patterns.
    - Verify that nested .gitignore files override parent patterns correctly.
  - **Success Criteria:** Integration tests verify filtering behavior with real implementations.
  - **Depends On:** Refactor gitignoreFiltering.test.ts.
  - **AC Ref:** AC 2.2.

- [ ] **Update readContextPaths.test.ts**
  - **Action:** Refactor `readContextPaths.test.ts` to use virtual .gitignore files.
  - **Technical Details:**
    - Replace any mock implementations with `setupWithGitignore` or `setupMultiGitignore`.
    - Test actual path filtering with directory traversal.
    - Focus on the integration between gitignore filtering and directory reading.
  - **Success Criteria:** Tests verify actual behavior without mocks.
  - **Depends On:** Refactor gitignoreFiltering.test.ts.
  - **AC Ref:** AC 2.2.

- [ ] **Investigate complex gitignore pattern support**
  - **Action:** Examine issues with complex gitignore patterns in `gitignoreComplexPatterns.test.ts`.
  - **Technical Details:**
    - Identify which complex patterns are problematic (brace expansion, wildcards).
    - Review the gitignore implementation for limitations.
    - Test each pattern type individually to isolate issues.
    - Compare behavior with actual Git implementation using `git check-ignore`.
  - **Success Criteria:** Clear understanding of pattern limitations with test cases.
  - **Depends On:** Refactor gitignoreFiltering.test.ts.
  - **AC Ref:** AC 2.3.

- [ ] **Fix or document complex pattern limitations**
  - **Action:** Either implement fixes for complex pattern support or document limitations.
  - **Technical Details:**
    - If fixing:
      - Enhance pattern handling in `gitignoreUtils.ts`.
      - Add tests for each complex pattern type.
    - If documenting:
      - Update `gitignoreUtils.ts` with clear JSDoc comments about limitations.
      - Add warnings for unsupported patterns.
      - Document in `TESTING.md` and READMEs.
  - **Success Criteria:** Either fixed support for complex patterns or clear documentation of limitations.
  - **Depends On:** Investigate complex gitignore pattern support.
  - **AC Ref:** AC 2.3.

- [ ] **Remove mockGitignoreUtils.ts**
  - **Action:** Delete the mock implementation once all tests are migrated.
  - **Technical Details:**
    - Verify all tests that used it have been migrated.
    - Remove the file and any direct dependencies.
    - Update `jest/README.md` to deprecate its usage.
  - **Success Criteria:** File removed with no references remaining.
  - **Depends On:** All gitignore test refactorings.
  - **AC Ref:** AC 2.4.

## Phase 3: Implement Dependency Injection

- [ ] **Review and enhance core interfaces**
  - **Action:** Ensure all interfaces in `src/core/interfaces.ts` are complete and well-documented.
  - **Technical Details:**
    - Review each interface: `FileSystem`, `LLMClient`, `ConfigManagerInterface`, `ConsoleLogger`, `UISpinner`.
    - Compare interface methods with actual usage in the codebase.
    - Add missing methods and parameters based on usage patterns.
    - Ensure comprehensive JSDoc documentation for each interface and method.
    - Verify typings are strict and accurate.
  - **Success Criteria:** Complete, well-documented interfaces that cover all required functionality.
  - **Depends On:** None.
  - **AC Ref:** AC 3.1.

- [ ] **Implement ConsoleAdapter class**
  - **Action:** Ensure `ConsoleAdapter` is fully implemented for the `ConsoleLogger` interface.
  - **Technical Details:**
    - Create or enhance `src/core/ConsoleAdapter.ts` to implement all methods in `ConsoleLogger`.
    - Use the singleton `logger` internally but expose the interface methods.
    - Ensure proper error handling and formatting.
    - Add comprehensive tests in `src/core/__tests__/ConsoleAdapter.test.ts`.
  - **Implementation Example:**
    ```typescript
    // src/core/ConsoleAdapter.ts
    import { ConsoleLogger } from './interfaces';
    import { logger } from '../utils/logger';
    
    export class ConsoleAdapter implements ConsoleLogger {
      error(message: string, error?: Error): void {
        logger.error(message, error);
      }
      
      warn(message: string): void {
        logger.warn(message);
      }
      
      // Implement remaining methods...
    }
    ```
  - **Success Criteria:** Fully implemented adapter with test coverage.
  - **Depends On:** Review and enhance core interfaces.
  - **AC Ref:** AC 3.2.

- [ ] **Verify and complete other concrete implementations**
  - **Action:** Ensure concrete implementations exist for all other interfaces.
  - **Technical Details:**
    - Verify `ConcreteFileSystem` implements all `FileSystem` methods.
    - Verify `ConcreteConfigManager` implements all `ConfigManagerInterface` methods.
    - Ensure provider implementations fulfill the `LLMClient` interface.
    - Create any missing implementations or methods.
  - **Success Criteria:** Complete set of implementations for all interfaces.
  - **Depends On:** Review and enhance core interfaces.
  - **AC Ref:** AC 3.2.

- [ ] **Define workflow dependencies context object**
  - **Action:** Create a context object type to encapsulate all dependencies for workflow modules.
  - **Technical Details:**
    - Define a `WorkflowDependencies` interface in `src/workflow/types.ts`:
      ```typescript
      export interface WorkflowDependencies {
        fileSystem: FileSystem;
        configManager: ConfigManagerInterface;
        llmClient: LLMClient;
        consoleLogger: ConsoleLogger;
        // Add other dependencies as needed
      }
      ```
    - Add factory function to create an instance with default implementations.
  - **Success Criteria:** Well-defined context object for dependency injection.
  - **Depends On:** Verify and complete other concrete implementations.
  - **AC Ref:** AC 3.3.

- [ ] **Refactor runThinktank for dependency injection**
  - **Action:** Modify `runThinktank.ts` to accept injected dependencies.
  - **Technical Details:**
    - Update function signature to accept the `WorkflowDependencies` object.
    - Remove direct imports of concrete implementations.
    - Use injected dependencies throughout the function.
    - Ensure proper error handling if dependencies are missing.
    - Update tests to provide mock implementations.
  - **Implementation Example:**
    ```typescript
    // Before
    import { logger } from '../utils/logger';
    import { ConcreteFileSystem } from '../core/ConcreteFileSystem';
    
    export async function runThinktank(options: RunOptions): Promise<string> {
      const fileSystem = new ConcreteFileSystem();
      // ...
      logger.info('Starting operation...');
      // ...
    }
    
    // After
    export async function runThinktank(
      options: RunOptions,
      deps: WorkflowDependencies
    ): Promise<string> {
      const { fileSystem, consoleLogger } = deps;
      // ...
      consoleLogger.info('Starting operation...');
      // ...
    }
    ```
  - **Success Criteria:** `runThinktank` uses injected dependencies with no direct imports.
  - **Depends On:** Define workflow dependencies context object.
  - **AC Ref:** AC 3.3.

- [ ] **Refactor runThinktankHelpers.ts for DI**
  - **Action:** Update helper functions to accept relevant dependencies.
  - **Technical Details:**
    - Pass dependencies to helper functions from `runThinktank`.
    - Remove direct imports of concrete implementations.
    - For each helper function, accept only the dependencies it needs.
  - **Success Criteria:** Helper functions use injected dependencies.
  - **Depends On:** Refactor runThinktank for dependency injection.
  - **AC Ref:** AC 3.3.

- [ ] **Refactor _executeQueries for LLMClient injection**
  - **Action:** Update `_executeQueries` to accept an `LLMClient` instance.
  - **Technical Details:**
    - Remove direct LLM provider access or imports.
    - Pass the `LLMClient` instance from `runThinktank`.
    - Use the interface methods for query execution.
  - **Success Criteria:** `_executeQueries` uses the injected `LLMClient`.
  - **Depends On:** Refactor runThinktankHelpers.ts for DI.
  - **AC Ref:** AC 3.3.

- [ ] **Create CLI handler interfaces**
  - **Action:** Define interfaces for CLI command handlers.
  - **Technical Details:**
    - Create `src/cli/interfaces.ts` with handler interfaces:
      ```typescript
      export interface RunCommandHandler {
        execute(options: RunCommandOptions): Promise<void>;
      }
      
      export interface ModelsCommandHandler {
        execute(options: ModelsCommandOptions): Promise<void>;
      }
      
      // Add others as needed
      ```
    - Define option types for each command.
  - **Success Criteria:** Well-defined interfaces for all CLI command handlers.
  - **Depends On:** None.
  - **AC Ref:** AC 3.4.

- [ ] **Implement concrete CLI handlers**
  - **Action:** Create handler classes implementing the interfaces.
  - **Technical Details:**
    - Create handlers for each command (Run, Models, Config).
    - Inject dependencies via constructor.
    - Move logic from commander callbacks to handlers.
    - Implement proper error handling.
  - **Implementation Example:**
    ```typescript
    // src/cli/handlers/RunCommandHandler.ts
    import { RunCommandHandler } from '../interfaces';
    import { RunCommandOptions } from '../types';
    import { WorkflowDependencies } from '../../workflow/types';
    import { runThinktank } from '../../workflow/runThinktank';
    
    export class ConcreteRunCommandHandler implements RunCommandHandler {
      constructor(private deps: WorkflowDependencies) {}
      
      async execute(options: RunCommandOptions): Promise<void> {
        // Convert command options to runThinktank options
        const runOptions = { /* ... */ };
        
        try {
          const result = await runThinktank(runOptions, this.deps);
          // Handle result
        } catch (error) {
          // Handle error
          this.deps.consoleLogger.error('Command failed', error as Error);
        }
      }
    }
    ```
  - **Success Criteria:** Command logic moved to testable handler classes.
  - **Depends On:** Create CLI handler interfaces.
  - **AC Ref:** AC 3.4.

- [ ] **Refactor CLI setup to use handlers**
  - **Action:** Update CLI module to use the handler classes.
  - **Technical Details:**
    - Modify `src/cli/index.ts` to instantiate the handlers with dependencies.
    - Update commander action callbacks to delegate to handlers.
    - Inject default or configured dependencies.
  - **Implementation Example:**
    ```typescript
    // In command setup
    program
      .command('run')
      // ... options ...
      .action(async (promptFile, contextPaths, options) => {
        const deps = createWorkflowDependencies();
        const handler = new ConcreteRunCommandHandler(deps);
        
        try {
          await handler.execute({ promptFile, contextPaths, ...options });
        } catch (error) {
          process.exit(1);
        }
      });
    ```
  - **Success Criteria:** CLI setup uses handler classes instead of inline logic.
  - **Depends On:** Implement concrete CLI handlers.
  - **AC Ref:** AC 3.4.

- [ ] **Create DI test helpers**
  - **Action:** Implement helper functions for creating mock dependencies.
  - **Technical Details:**
    - Add `createMockLlmClient` to `test/setup/providers.ts`.
    - Add `createMockWorkflowDependencies` for full test context.
    - Ensure all mock creators return strict types matching interfaces.
  - **Implementation Example:**
    ```typescript
    // test/setup/workflow.ts
    export function createMockWorkflowDependencies(): jest.Mocked<WorkflowDependencies> {
      return {
        fileSystem: createMockFileSystem(),
        configManager: createMockConfigManager(),
        llmClient: createMockLlmClient(),
        consoleLogger: createMockConsoleLogger()
      };
    }
    ```
  - **Success Criteria:** Complete set of helpers for creating mock dependencies.
  - **Depends On:** None.
  - **AC Ref:** AC 3.5.

- [ ] **Update workflow tests for DI**
  - **Action:** Refactor workflow tests to use injected dependencies.
  - **Technical Details:**
    - Replace `jest.mock()` calls with dependency injection.
    - Use `createMockWorkflowDependencies()` to create test context.
    - Verify interactions with mock dependencies instead of module mocks.
  - **Implementation Example:**
    ```typescript
    // Before
    jest.mock('../../core/ConcreteFileSystem');
    jest.mock('../../utils/logger');
    
    // After
    import { createMockWorkflowDependencies } from '../../../test/setup/workflow';
    
    describe('runThinktank', () => {
      let mockDeps: jest.Mocked<WorkflowDependencies>;
      
      beforeEach(() => {
        mockDeps = createMockWorkflowDependencies();
      });
      
      it('should log success message on completion', async () => {
        await runThinktank(options, mockDeps);
        
        expect(mockDeps.consoleLogger.success).toHaveBeenCalledWith(
          expect.stringContaining('completed successfully')
        );
      });
    });
    ```
  - **Success Criteria:** All workflow tests use mock dependencies via injection.
  - **Depends On:** Create DI test helpers.
  - **AC Ref:** AC 3.5.

- [ ] **Update CLI tests for DI**
  - **Action:** Refactor CLI tests to use injected dependencies.
  - **Technical Details:**
    - Replace direct mocks with handler mocks or dependency injection.
    - Test handler classes directly with mock dependencies.
    - Verify correct operation and error handling.
  - **Success Criteria:** CLI tests use dependency injection pattern.
  - **Depends On:** Update workflow tests for DI.
  - **AC Ref:** AC 3.5.

## Phase 4: Isolate Side Effects

- [ ] **Define FileData interface**
  - **Action:** Create a structured type for file output data.
  - **Technical Details:**
    - Define in `src/workflow/types.ts`:
      ```typescript
      export interface FileData {
        path: string;
        content: string;
        modelId: string;
        metadata?: Record<string, any>;
      }
      ```
    - Add documentation for each field.
  - **Success Criteria:** Well-defined type for file output data.
  - **Depends On:** None.
  - **AC Ref:** AC 4.1.

- [ ] **Refactor outputHandler.ts for pure functions**
  - **Action:** Modify `outputHandler.ts` to return data instead of performing I/O.
  - **Technical Details:**
    - Update `_processOutput` to return `FileData[]` instead of writing files.
    - Extract file path generation logic into a pure function.
    - Extract metadata formatting logic into a pure function.
    - Return a structured object with file data and console output.
  - **Implementation Example:**
    ```typescript
    // Before
    async function _processOutput(results, outputDir, options) {
      // ... logic and calculations ...
      await fileSystem.writeFile(filePath, content);
      // ... more I/O ...
    }
    
    // After
    function processOutput(results, options): { 
      files: FileData[], 
      consoleOutput: string 
    } {
      // ... logic and calculations ...
      return {
        files: [
          { path: filePath, content, modelId, metadata },
          // Other files
        ],
        consoleOutput: summaryText
      };
    }
    ```
  - **Success Criteria:** Pure function returning structured data without I/O.
  - **Depends On:** Define FileData interface.
  - **AC Ref:** AC 4.1.

- [ ] **Create formatCompletionSummary utility**
  - **Action:** Extract summary formatting logic to a pure function.
  - **Technical Details:**
    - Create `src/utils/formatCompletionSummary.ts`.
    - Move logic from `_logCompletionSummary` to the new function.
    - Return a formatted object with summary text and error details.
    - Add comprehensive tests.
  - **Implementation Example:**
    ```typescript
    // src/utils/formatCompletionSummary.ts
    export interface CompletionSummary {
      summaryText: string;
      errorDetails?: string[];
      verboseOutput?: string[];
    }
    
    export function formatCompletionSummary(
      results: LLMResponse[],
      options: { verbose?: boolean }
    ): CompletionSummary {
      // ... format summary text and collect error details ...
      
      return {
        summaryText,
        errorDetails: errors.length > 0 ? errors : undefined,
        verboseOutput: options.verbose ? verboseDetails : undefined
      };
    }
    ```
  - **Success Criteria:** Pure function with unit tests that handles summary formatting.
  - **Depends On:** None.
  - **AC Ref:** AC 4.2.

- [ ] **Create io.ts module**
  - **Action:** Implement a dedicated module for I/O operations.
  - **Technical Details:**
    - Create `src/workflow/io.ts` with functions for:
      - `writeOutputFiles(files: FileData[], fileSystem: FileSystem)`: Writes FileData to disk.
      - `logCompletionSummary(summary: CompletionSummary, logger: ConsoleLogger)`: Logs summary.
      - `ensureOutputDirectory(dir: string, fileSystem: FileSystem)`: Creates output directory.
    - Each function should accept dependencies as parameters.
    - Add robust error handling and logging.
    - Add comprehensive tests.
  - **Implementation Example:**
    ```typescript
    // src/workflow/io.ts
    export async function writeOutputFiles(
      files: FileData[],
      fileSystem: FileSystem
    ): Promise<void> {
      for (const file of files) {
        const dir = path.dirname(file.path);
        await fileSystem.mkdir(dir, { recursive: true });
        await fileSystem.writeFile(file.path, file.content);
      }
    }
    ```
  - **Success Criteria:** Centralized I/O module with dependency injection.
  - **Depends On:** Refactor outputHandler.ts for pure functions, Create formatCompletionSummary utility.
  - **AC Ref:** AC 4.3.

- [ ] **Update runThinktank to use io module**
  - **Action:** Modify `runThinktank.ts` to use the new I/O module.
  - **Technical Details:**
    - Update `runThinktank` to call the pure functions for data processing.
    - Pass the results to the I/O functions with the appropriate dependencies.
    - Maintain proper error handling and flow control.
  - **Implementation Example:**
    ```typescript
    // Inside runThinktank
    const { files, consoleOutput } = processOutput(results, options);
    
    // Handle file I/O with injected fileSystem
    await io.writeOutputFiles(files, deps.fileSystem);
    
    // Log completion with injected consoleLogger
    const summary = formatCompletionSummary(results, options);
    io.logCompletionSummary(summary, deps.consoleLogger);
    
    return consoleOutput;
    ```
  - **Success Criteria:** `runThinktank` uses pure functions and I/O module.
  - **Depends On:** Create io.ts module.
  - **AC Ref:** AC 4.3.

- [ ] **Write unit tests for pure functions**
  - **Action:** Create comprehensive tests for the refactored pure functions.
  - **Technical Details:**
    - Test `processOutput` with various input scenarios.
    - Test `formatCompletionSummary` with different result types.
    - Verify correct output structures and error handling.
    - Focus on testing logic without mocking I/O.
  - **Implementation Example:**
    ```typescript
    describe('processOutput', () => {
      it('should generate correct FileData for each result', () => {
        const results = [/* mock LLMResponse objects */];
        const options = { outputDir: '/output' };
        
        const { files } = processOutput(results, options);
        
        expect(files).toHaveLength(results.length);
        expect(files[0]).toEqual({
          path: expect.stringContaining(results[0].modelId),
          content: expect.stringContaining(results[0].content),
          modelId: results[0].modelId
        });
      });
    });
    ```
  - **Success Criteria:** High coverage of pure function logic.
  - **Depends On:** Refactor outputHandler.ts for pure functions, Create formatCompletionSummary utility.
  - **AC Ref:** AC 4.5.

- [ ] **Write integration tests for io.ts**
  - **Action:** Implement tests for the I/O functions.
  - **Technical Details:**
    - Create `src/workflow/__tests__/io.test.ts`.
    - Test each function with mock dependencies.
    - Verify correct method calls on the mock dependencies.
    - Test error handling scenarios.
  - **Implementation Example:**
    ```typescript
    describe('writeOutputFiles', () => {
      it('should create directories and write files', async () => {
        const mockFs = createMockFileSystem();
        const files: FileData[] = [
          { path: '/output/model1.md', content: 'content1', modelId: 'model1' }
        ];
        
        await io.writeOutputFiles(files, mockFs);
        
        expect(mockFs.mkdir).toHaveBeenCalledWith(
          '/output', { recursive: true }
        );
        expect(mockFs.writeFile).toHaveBeenCalledWith(
          '/output/model1.md', 'content1'
        );
      });
    });
    ```
  - **Success Criteria:** Comprehensive tests for I/O functions with injected mocks.
  - **Depends On:** Create io.ts module.
  - **AC Ref:** AC 4.6.

## Phase 5: Refine and Finalize

- [ ] **Refactor error handling in ConcreteFileSystem**
  - **Action:** Clean up repetitive error wrapping code in `ConcreteFileSystem`.
  - **Technical Details:**
    - Extract common error wrapping logic to a helper function.
    - Use a decorator pattern or higher-order function to wrap methods.
    - Ensure consistent error messages and properties.
  - **Implementation Example:**
    ```typescript
    // Helper function
    private wrapFsError<T>(
      operation: () => Promise<T>,
      method: string,
      path: string
    ): Promise<T> {
      try {
        return operation();
      } catch (error) {
        throw new FileSystemError(
          `Failed to ${method} at path: ${path}`,
          error as Error
        );
      }
    }
    
    // Usage
    async readFileContent(filePath: string): Promise<string> {
      return this.wrapFsError(
        () => fileReader.readFileContent(filePath),
        'read file',
        filePath
      );
    }
    ```
  - **Success Criteria:** Reduced code duplication in error handling.
  - **Depends On:** None.
  - **AC Ref:** AC 5.1.

- [ ] **Replace direct console usage with ConsoleLogger**
  - **Action:** Find and replace any direct `console` calls with the injected logger.
  - **Technical Details:**
    - Search for patterns like `console.log`, `console.error`.
    - Replace with the appropriate `consoleLogger` method.
    - Add ConsoleLogger to any function that needs logging capability.
    - Update tests to verify correct logger usage.
  - **Success Criteria:** No direct console usage remains in the codebase.
  - **Depends On:** Implement ConsoleAdapter class.
  - **AC Ref:** AC 5.1.

- [ ] **Fix documentation links**
  - **Action:** Correct broken links in documentation files.
  - **Technical Details:**
    - Review all markdown files (README.md, etc.).
    - Check links to source files, documentation, and external resources.
    - Update or remove broken links.
    - Add links to new modules and interfaces.
  - **Success Criteria:** All documentation links work correctly.
  - **Depends On:** None.
  - **AC Ref:** AC 5.1.

- [ ] **Run code coverage analysis**
  - **Action:** Execute test coverage and analyze results.
  - **Technical Details:**
    - Run `pnpm test:cov` to generate coverage report.
    - Identify modules with low coverage (< 80%).
    - Prioritize critical code paths for additional testing.
    - Focus on behavioral coverage over line coverage.
  - **Success Criteria:** Understanding of coverage gaps with prioritization plan.
  - **Depends On:** Complete all previous test-related tasks.
  - **AC Ref:** AC 5.2.

- [ ] **Write tests for identified coverage gaps**
  - **Action:** Implement tests for areas with insufficient coverage.
  - **Technical Details:**
    - Focus on behavioral tests for core workflow functionality.
    - Add edge case and error handling tests.
    - Ensure all public API methods have proper test coverage.
    - Use dependency injection for testability.
  - **Success Criteria:** Improved overall coverage, especially for critical paths.
  - **Depends On:** Run code coverage analysis.
  - **AC Ref:** AC 5.2.

- [ ] **Create or update TESTING.md**
  - **Action:** Document the testing approach and standards.
  - **Technical Details:**
    - Create comprehensive guide for testing in the project.
    - Include sections on:
      - Testing philosophy and principles
      - Mocking strategy with dependency injection
      - Virtual filesystem testing
      - Test helpers and utilities
      - Common patterns and anti-patterns
    - Link to example tests for each pattern.
  - **Success Criteria:** Complete documentation of testing approach.
  - **Depends On:** Complete all Phase 1-4 tasks.
  - **AC Ref:** AC 5.3.

- [ ] **Update CONTRIBUTING.md**
  - **Action:** Align contributing guidelines with new testing approach.
  - **Technical Details:**
    - Update testing section with new strategies.
    - Add guidance on dependency injection.
    - Document the pure function approach for testability.
    - Include code examples for common patterns.
  - **Success Criteria:** Up-to-date contributing guidelines.
  - **Depends On:** Create or update TESTING.md.
  - **AC Ref:** AC 5.3.

- [ ] **Remove dead code and unused mocks**
  - **Action:** Clean up any remaining unused code and mocks.
  - **Technical Details:**
    - Search for deprecated or commented-out code blocks.
    - Identify unused utilities and helper functions.
    - Remove legacy mock implementations.
    - Update imports to remove unused dependencies.
  - **Success Criteria:** Clean codebase without dead code.
  - **Depends On:** Complete all Phase 1-4 tasks.
  - **AC Ref:** AC 5.4.

- [ ] **Run linting and formatting**
  - **Action:** Apply consistent code style throughout the codebase.
  - **Technical Details:**
    - Run `pnpm run lint:fix` to fix linting issues.
    - Run `pnpm run format` to apply consistent formatting.
    - Run `pnpm run fix:newlines` to fix EOL issues.
    - Address any remaining lint warnings manually.
  - **Success Criteria:** Clean lint and format checks with no warnings.
  - **Depends On:** Remove dead code and unused mocks.
  - **AC Ref:** AC 5.4.

- [ ] **Conduct final verification**
  - **Action:** Perform a comprehensive test and verification.
  - **Technical Details:**
    - Run the full test suite (`pnpm test`).
    - Verify all features work as expected with manual testing.
    - Check for any regressions in functionality.
    - Verify the build process (`pnpm run build`).
  - **Success Criteria:** All tests pass and functionality works as expected.
  - **Depends On:** All previous tasks.
  - **AC Ref:** All ACs.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **FileData Structure Definition**
  - **Context:** Phase 4 refers to returning `FileData[]` objects, but their exact structure isn't defined.
  - **Assumption:** We'll define a new `FileData` interface in `src/workflow/types.ts` with fields for `path`, `content`, `modelId`, and optional `metadata`.
  - **Impact:** If this structure already exists elsewhere, we'll need to use that instead of creating a new definition.

- [ ] **Approach to CLI Command Handlers**
  - **Context:** Phase 3 mentions extracting command logic from commander callbacks, but doesn't specify the exact pattern.
  - **Assumption:** We'll implement a handler class pattern with dependency injection via constructor parameters, implementing interfaces defined in a new `src/cli/interfaces.ts` file.
  - **Impact:** If there's a preferred pattern already established in the codebase, we might need to adjust this approach.

- [ ] **Test Coverage Targets**
  - **Context:** Phase 5 mentions reviewing test coverage, but doesn't specify target percentages.
  - **Assumption:** We'll aim for at least 80% coverage for core modules (workflow, core) and 70% for utility modules.
  - **Impact:** If specific coverage targets are already established, we'll need to adjust our targets accordingly.

- [ ] **Gitignore Helper Functions Implementation**
  - **Context:** Phase 2 references helpers like `setupWithGitignore` and `setupMultiGitignore`, but they don't appear in the reviewed code.
  - **Assumption:** We'll implement these helpers in a new `test/setup/gitignore.ts` module, building on the existing `addVirtualGitignoreFile` utility.
  - **Impact:** If these functions are planned to be implemented elsewhere, we'll need to adjust our approach.

- [ ] **io.ts Module Location and Structure**
  - **Context:** Phase 4 suggests creating a dedicated `io.ts` module, but its exact location and structure aren't specified.
  - **Assumption:** We'll create `src/workflow/io.ts` with functions for file writing and console output that accept dependencies as parameters.
  - **Impact:** If there's a preferred structure or location for I/O utilities, we might need to adjust this approach.
