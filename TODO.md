# TODO

## High Priority

- [x] **Fix skipped tests**
  - [x] Update `readDirectoryContents.test.ts` to work with the FileSystem interface
  - [x] Resolve issues in `readContextFile.centralized-mock.test.ts` (renamed to readContextFile.interface.test.ts)
  - [x] Fix any other skipped tests in the codebase (one test for empty files intentionally left skipped)

- [x] **Fix E2E testing**
  - [x] Remove internal mocking from `src/workflow/__tests__/runThinktank.e2e.test.ts`
  - [x] Replace `modelSelector` mock with CLI args or config files for controlling test behavior
  - [x] Verify E2E tests treat the application as a true black box

## Medium Priority

- [ ] **Refactor concrete implementation tests**
  - [ ] Revise `src/core/__tests__/ConcreteConfigManager.test.ts` to test behavior, not implementation
  - [ ] Update `src/core/__tests__/ConcreteFileSystem.test.ts` to use virtual filesystem
  - [ ] Refactor `src/core/__tests__/ConcreteLLMClient.test.ts` to focus on interface contract

- [ ] **Complete dependency injection adoption**
  - [ ] Make `FileSystem` a required parameter in `src/utils/fileReader.ts`
  - [ ] Update `src/utils/inputHandler.ts` to fully use DI without conditional branches
  - [ ] Refactor `src/utils/outputHandler.ts` to require injected dependencies
  - [ ] Update all call sites to provide required dependencies

- [ ] **Simplify ConcreteLLMClient**
  - [ ] Refactor `generate` method in `src/core/LLMClient.ts`
  - [ ] Remove duplication with `llmRegistry.callProvider`
  - [ ] Clarify responsibility boundaries between components

## Low Priority

- [ ] **Extract common error handling**
  - [ ] Create reusable error wrapper functions for filesystem operations
  - [ ] Create reusable error wrapper functions for configuration operations
  - [ ] Create reusable error wrapper functions for API operations
  - [ ] Replace repetitive error handling in concrete implementations

- [ ] **Remove logic duplication in fileReader.ts**
  - [ ] Extract common logic into private helper functions
  - [ ] Refactor `readContextFile` to use these helpers
  - [ ] Refactor `readContextPaths` to use these helpers
