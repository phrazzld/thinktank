# Thinktank Test Suite and Architecture Refactoring Plan

> **Note:** This plan outlines the architectural refactoring approach for testing. For current development tasks and priorities, see [BACKLOG.md](/BACKLOG.md).

**Goal:** Refactor the test suite and application architecture to improve simplicity, maintainability, robustness, and extensibility, focusing on completing the migration to modern testing approaches and decoupling components for better testability.

## Phase 1: Complete `memfs` Migration for Filesystem Tests

**Objective:** Eliminate the legacy `mockFsUtils.ts` and ensure all filesystem-related unit/integration tests use the `virtualFsUtils.ts` (`memfs`) approach for speed, isolation, and consistency.

**Status: In Progress** - Standardizing the use of `createVirtualFs` helper across all tests.

### Steps:

1. **Identify Target Test Files:**
   - Review all files under `src/**/__tests__/*.test.ts`
   - Identify tests currently importing or relying on `mockFsUtils.ts`
   - Cross-reference with skipped tests in `jest.config.js` (`testPathIgnorePatterns`)
   
2. **Refactor Each Target Test File:**
   - Remove legacy imports from `../../__tests__/utils/mockFsUtils`
   - Add `memfs` setup:
     ```typescript
     import { mockFsModules } from '../../__tests__/utils/virtualFsUtils';
     jest.mock('fs', () => mockFsModules().fs);
     jest.mock('fs/promises', () => mockFsModules().fsPromises);
     // Now import fs, fs/promises, and the module under test
     import fsPromises from 'fs/promises';
     import { yourFunctionUnderTest } from '../yourModule';
     ```
   - Replace mock setups with `createVirtualFs`:
     ```typescript
     // Before (Legacy)
     // beforeEach(() => {
     //   resetMockFs();
     //   setupMockFs();
     //   mockReadFile('/path/to/file.txt', 'content');
     //   mockStat('/path/to/dir', { isDirectory: () => true });
     //   mockMkdir('/output/dir', true);
     // });

     // After (memfs)
     beforeEach(() => {
       resetVirtualFs();
       createVirtualFs({
         '/path/to/file.txt': 'content',
         '/path/to/dir/': '', // Creates a directory
         // '/output/dir/' will be created by the function under test
       });
     });
     ```
   - Update assertions to verify filesystem state:
     ```typescript
     // Before (Legacy)
     // await createDirectory('/output/dir');
     // expect(mockedFs.mkdir).toHaveBeenCalledWith('/output/dir', { recursive: true });

     // After (memfs)
     await createDirectory('/output/dir');
     const virtualFs = getVirtualFs();
     expect(virtualFs.existsSync('/output/dir')).toBe(true);
     expect(virtualFs.statSync('/output/dir').isDirectory()).toBe(true);
     ```
   - Refactor error testing:
     ```typescript
     it('should handle write permission errors', async () => {
       createVirtualFs({ '/path/to/': '' }); // Ensure parent dir exists
       const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
       writeFileSpy.mockRejectedValueOnce(
         createFsError('EACCES', 'Permission denied', 'writeFile', '/path/to/file.txt')
       );

       await expect(writeFileFunction('/path/to/file.txt', 'content'))
         .rejects.toThrow(/Permission denied/);

       writeFileSpy.mockRestore(); // Clean up the spy
     });
     ```
   - Update `jest.config.js` by removing its path from `testPathIgnorePatterns`

3. **Remove Legacy Utilities:**
   - Once all test files are migrated, delete `src/__tests__/utils/mockFsUtils.ts`
   - Remove any helper functions in `test-helpers.ts` related to the old mocking strategy

## Phase 2: Simplify Gitignore Testing

**Objective:** Test the actual `gitignoreUtils` implementation against virtual `.gitignore` files instead of mocking the utility itself.

### Steps:

1. **Enhance Filesystem Utilities:**
   - Ensure `virtualFsUtils.ts` can create hidden files like `.gitignore`
   - Verify `addVirtualGitignoreFile` or integrate its logic into `virtualFsUtils.ts`

2. **Refactor Gitignore-Related Tests:**
   - Remove mocks (`jest.mock('../gitignoreUtils')`) and imports from `mockGitignoreUtils`
   - Import actual functions from `src/utils/gitignoreUtils`
   - Setup virtual `.gitignore` files:
     ```typescript
     beforeEach(() => {
       resetVirtualFs();
       createVirtualFs({
         '/project/': '', // Create directory
         '/project/.gitignore': '*.log\n/dist/\n', // Add gitignore
         '/project/src/': '',
         '/project/src/app.ts': 'content',
         '/project/app.log': 'log content',
         '/project/dist/bundle.js': 'bundle content'
       });
       // Clear ignore cache if needed
       gitignoreUtils.clearIgnoreCache();
     });
     ```
   - Test the actual implementation:
     ```typescript
     it('should ignore files based on virtual .gitignore', async () => {
       // Get the filter for the virtual directory
       const filter = await gitignoreUtils.createIgnoreFilter('/project');
       expect(filter.ignores('app.log')).toBe(true);
       expect(filter.ignores('dist/bundle.js')).toBe(true);
       expect(filter.ignores('src/app.ts')).toBe(false);

       // Test shouldIgnorePath directly
       expect(await gitignoreUtils.shouldIgnorePath('/project', 'app.log')).toBe(true);
       expect(await gitignoreUtils.shouldIgnorePath('/project', 'src/app.ts')).toBe(false);
     });
     ```

3. **Remove `mockGitignoreUtils.ts`:**
   - Once all dependent tests are refactored, delete this file

## Phase 3: Decouple Dependencies with Interfaces

**Objective:** Reduce reliance on mocking by introducing interfaces and dependency injection.

### Steps:

1. **Identify Key Dependencies:**
   - Analyze modules interacting with external systems (APIs, filesystem, etc.)
   - Focus on `runThinktank.ts`, `runThinktankHelpers.ts`, and related functions

2. **Define Interfaces:**
   - Create interfaces for external dependencies:
     ```typescript
     // LLM API client interface
     interface LLMClient {
       generate(prompt: string, modelId: string, options?: ModelOptions): Promise<LLMResponse>;
     }

     // Config manager interface
     interface ConfigManagerInterface {
       loadConfig(): Promise<AppConfig>;
       saveConfig(config: AppConfig): Promise<void>;
       // ... other methods
     }

     // File system interface
     interface FileSystem {
       readDir(path: string): Promise<FileData[]>;
       writeFile(path: string, content: string): Promise<void>;
       // ... other methods
     }
     ```

3. **Implement Dependency Injection:**
   - Modify functions to accept interfaces as parameters:
     ```typescript
     export async function _executeQueries({
       spinner,
       config,
       models,
       combinedContent,
       options,
       llmClient, // Injected dependency
     }: ExecuteQueriesParams & { llmClient: LLMClient }): Promise<ExecuteQueriesResult> {
       // Use llmClient instead of direct API calls
     }
     ```

4. **Update Tests:**
   - Provide mock implementations:
     ```typescript
     const mockClient: LLMClient = { 
       generate: jest.fn().mockResolvedValue({ 
         text: 'mock response', 
         provider: 'mock', 
         modelId: 'test' 
       }) 
     };
     ```

## Phase 4: Isolate Side Effects

**Objective:** Separate I/O operations from core logic to improve testability.

### Steps:

1. **Identify Functions with Side Effects:**
   - Analyze functions performing I/O (file writing, console logging)
   - Focus on `_processOutput`, `_logCompletionSummary`, etc.

2. **Refactor to Return Data:**
   - Modify functions to return data rather than perform I/O:
     ```typescript
     export async function _processOutput({
       queryResults,
       options,
       friendlyRunName,
     }: ProcessOutputParams): Promise<{ files: FileData[]; consoleOutput: string }> {
       // Return file content, don't write it
       const files = queryResults.responses.map(r => ({
         filename: `${r.provider}-${r.modelId}.md`,
         content: formatResponse(r),
       }));
       return { files, consoleOutput: formatConsoleOutput(files) };
     }
     ```

3. **Move I/O to Higher-Level Functions:**
   - Push actual I/O operations to orchestration functions:
     ```typescript
     // In runThinktank or dedicated I/O module:
     const { files, consoleOutput } = await _processOutput({ ... });
     await writeFiles(outputDirectoryPath, files);
     console.log(consoleOutput);
     ```

## Phase 5: Simplify CLI Testing

**Objective:** Extract command logic to enable direct testing without CLI framework overhead.

### Steps:

1. **Extract Command Handlers:**
   - Move action callback logic to dedicated functions:
     ```typescript
     // In cli/commands/run.ts
     export async function runCommandHandler(input: string, options: any): Promise<void> {
       // All the logic previously in .action()
       await runThinktank({ input, options });
     }

     // In CLI setup
     program.command('run <input>').action(runCommandHandler);
     ```

2. **Test Command Handlers Directly:**
   - Import and test like normal functions:
     ```typescript
     import { runCommandHandler } from '../commands/run';

     test('runCommand processes input', async () => {
       // Arrange
       const mockRunThinktank = jest.fn();
       jest.spyOn(workflowModule, 'runThinktank').mockImplementation(mockRunThinktank);
       
       // Act
       await runCommandHandler('prompt.txt', { model: 'gpt-4' });
       
       // Assert
       expect(mockRunThinktank).toHaveBeenCalledWith({
         input: 'prompt.txt', 
         options: { model: 'gpt-4' }
       });
     });
     ```

## Phase 6: Refine E2E Tests

**Objective:** Ensure E2E tests focus on CLI interaction using the real filesystem effectively.

### Steps:

1. **Review E2E Test Scope:**
   - Verify `cli.e2e.test.ts` and `runThinktank.e2e.test.ts` use real compiled CLI
   - Ensure they utilize `e2eTestUtils.ts` for temporary directories/files
   - Focus assertions on:
     - Command execution (arguments, options)
     - Standard output/error
     - Created output files and their contents
     - Exit codes (0 for success, non-zero for errors)
   - Remove internal logic checks; treat CLI as a black box

2. **Ensure Proper Cleanup:**
   - Verify `afterAll`/`afterEach` hooks use `cleanupTestDir` 
   - Ensure cleanup works even if tests fail

## Phase 7: Reduce Mock Complexity

**Objective:** Minimize mock complexity by simplifying dependencies and setup.

### Steps:

1. **Simplify Spinner Usage:**
   - Make spinner optional or use a null object in tests:
     ```typescript
     const nullSpinner = { 
       start: () => {}, 
       info: () => {}, 
       succeed: () => {},
       fail: () => {} 
     };
     ```

2. **Streamline Configuration:**
   - Use minimal config objects in tests:
     ```typescript
     const testConfig: AppConfig = { 
       models: [{ provider: 'mock', modelId: 'test', enabled: true }] 
     };
     ```

## Phase 8: Review Test Coverage and Fill Gaps

**Objective:** Ensure comprehensive test coverage after refactoring.

### Steps:

1. **Run Coverage Analysis:**
   - Execute `npm run test:cov`
   - Review `coverage/lcov-report/index.html`

2. **Identify and Fill Gaps:**
   - Pay attention to:
     - `src/utils/fileReader.ts` functions
     - `src/utils/gitignoreUtils.ts` pattern handling
     - `src/core/configManager.ts` configuration operations
     - Error classes and utilities
     - Workflow error propagation
   - Add missing tests for edge cases and error conditions

## Phase 9: Refine Error Handling Tests

**Objective:** Ensure error handling tests are robust and comprehensive.

### Steps:

1. **Review Error Tests:**
   - Check all error-related test files

2. **Verify Error Types:**
   - Assert specific error types (`FileSystemError`, `ConfigError`, etc.)
   - Confirm error messages, suggestions, and examples are tested

3. **Test Error Chaining:**
   - Verify that error cause is preserved when wrapped

4. **Test Filesystem Errors:**
   - Use `memfs` and spies to simulate various error conditions

## Phase 10: Update Documentation

**Objective:** Ensure documentation reflects the new testing strategy.

### Steps:

1. **Update `TESTING.md`:**
   - Remove references to legacy `mockFsUtils.ts`
   - Document the `memfs` approach with examples
   - Update gitignore testing section
   - Document dependency injection and interface mocking
   - Detail E2E testing strategy

2. **Update `src/__tests__/utils/README.md`:**
   - Remove legacy sections
   - Focus on how to use `virtualFsUtils.ts`
   - Update the gitignore testing documentation

3. **Review `CONTRIBUTING.md`:**
   - Ensure testing guidance aligns with new practices

## Implementation Notes and Best Practices

- **Commit incrementally** after refactoring each test file or logical group
- **Prioritize `memfs` migration** first as it provides the most immediate benefit
- **Follow Test-Driven Development (TDD)** when adding new features
- **Use code reviews** to ensure adherence to the new testing standards
- **Start with simplified examples** when introducing new patterns to the team

By systematically implementing this plan, the `thinktank` application will achieve better simplicity, modularity, maintainability, extensibility, and testability, resulting in a more robust and developer-friendly codebase that aligns with modern software engineering practices.