# TODO

## Test Suite Refactoring - Phase 4 (Aggressive Approach)

### 1. Create Pure Data Model

- [x] **Define result data interfaces**
  - **Action:** Create interfaces in `src/workflow/types.ts` for all data structures returned by pure functions: `ProcessOutputResult`, `CompletionSummary`, etc.
  - **Depends On:** None
  - **AC Ref:** AC 2.1 (Return data structures)

- [x] **Refactor output formatting utilities**
  - **Action:** Convert formatting functions in `src/utils/outputFormatter.ts` to pure functions that take data and return formatted strings.
  - **Depends On:** Define result data interfaces
  - **AC Ref:** AC 2.1 (Modify functions to return data)

### 2. Isolate I/O Layer

- [x] **Create I/O interface**
  - **Action:** Define compact I/O interface in `src/core/interfaces.ts` with clear separation of concerns.
  - **Depends On:** None
  - **AC Ref:** AC 3.1 (Move I/O to high-level functions)

- [x] **Implement file system adapter**
  - **Action:** Create adapter implementing I/O interface for file operations in `src/core/FileSystem.ts`.
  - **Depends On:** Create I/O interface
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

- [x] **Implement console adapter**
  - **Action:** Create adapter implementing I/O interface for console output in `src/core/ConsoleAdapter.ts`.
  - **Depends On:** Create I/O interface
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

### 3. Refactor Core Functions

- [x] **Refactor `_processOutput`**
  - **Action:** Convert `_processOutput` in `src/workflow/runThinktankHelpers.ts` to return structured data instead of performing I/O.
  - **Depends On:** Define result data interfaces
  - **AC Ref:** AC 2.1 (Modify functions to return data)

- [x] **Extract `_processOutput` I/O operations**
  - **Action:** Move file writing operations from `_processOutput` to `runThinktank.ts`.
  - **Depends On:** Refactor `_processOutput`
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

- [x] **Refactor `_logCompletionSummary`**
  - **Action:** Convert `_logCompletionSummary` in `src/workflow/runThinktankHelpers.ts` to return formatted strings instead of console logging.
  - **Depends On:** Define result data interfaces
  - **AC Ref:** AC 2.1 (Modify functions to return data)

- [x] **Extract `_logCompletionSummary` I/O operations**
  - **Action:** Move console output operations from `_logCompletionSummary` to `runThinktank.ts`.
  - **Depends On:** Refactor `_logCompletionSummary`
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

- [ ] **Refactor `processOutput` in outputHandler.ts**
  - **Action:** Convert to pure function returning data structures instead of performing I/O.
  - **Depends On:** Define result data interfaces
  - **AC Ref:** AC 2.1 (Modify functions to return data)

- [ ] **Refactor `writeResponsesToFiles` in outputHandler.ts**
  - **Action:** Convert to pure function that prepares file data without actual I/O.
  - **Depends On:** Define result data interfaces
  - **AC Ref:** AC 2.1 (Modify functions to return data)

### 4. Restructure Workflow

- [ ] **Simplify runThinktank orchestration**
  - **Action:** Rewrite `runThinktank` in `src/workflow/runThinktank.ts` to compose pure functions and perform I/O at boundaries.
  - **Depends On:** Refactor core functions
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

- [ ] **Create dedicated I/O module**
  - **Action:** Implement centralized I/O handling in `src/workflow/io.ts` for all workflow I/O operations.
  - **Depends On:** Implement file system adapter, Implement console adapter
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

### 5. Update Tests

- [ ] **Create testing utilities**
  - **Action:** Build mock implementations for I/O adapters in `src/__tests__/utils/mockFactories.ts`.
  - **Depends On:** Implement file system adapter, Implement console adapter
  - **AC Ref:** AC 3.1 (Improve testability)

- [ ] **Create unit tests for pure functions**
  - **Action:** Write unit tests for all refactored pure functions without I/O dependencies.
  - **Depends On:** Refactor core functions
  - **AC Ref:** AC 2.1 (Modify functions to return data)

- [ ] **Refactor workflow integration tests**
  - **Action:** Update integration tests to use I/O mocks consistently.
  - **Depends On:** Create testing utilities
  - **AC Ref:** AC 3.1 (Push I/O operations to orchestration)

- [ ] **Simplify test setup**
  - **Action:** Eliminate redundant test setup in workflow test files.
  - **Depends On:** Refactor workflow integration tests
  - **AC Ref:** AC 3.1 (Improve testability)
