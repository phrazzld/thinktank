# TODO

## CLI Error Handling Improvements
- [x] **Implement CLI command error handling tests**
  - **Action:** Create comprehensive test cases in `cli-command-error-handling.test.ts` that validate error handling for each CLI command, including input validation, configuration errors, and runtime errors.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1 (Missing Implementation in Placeholder Test).

- [x] **Refactor CLI error testing logic**
  - **Action:** Remove duplicated error handling logic in `cli-error-handling.test.ts` by having tests import and use the actual `handleError` function from `src/cli/index.ts` with properly mocked dependencies.
  - **Depends On:** None.
  - **AC Ref:** AC 1.2 (DRY Violation in CLI Error Tests).

## Error System Refinements
- [x] **Create helper method for error name property setting**
  - **Action:** Add a helper method in `ThinktankError` base class to standardize setting the error name property across all error subclasses, ensuring proper inheritance and correct `instanceof` checks.
  - **Depends On:** None.
  - **AC Ref:** AC 2.1 (Repeated manual setting of error "name").

- [x] **Split errors.ts into multiple files**
  - **Action:** Refactor the large `errors.ts` (996 lines) into multiple files organized by error category (e.g., ApiErrors.ts, ConfigErrors.ts) to improve maintainability.
  - **Depends On:** None.
  - **AC Ref:** AC 2.2 (Large file size of `errors.ts`).

- [ ] **Simplify error message construction**
  - **Action:** Refactor error message construction in provider error handlers to eliminate duplication and improve consistency across different providers.
  - **Depends On:** None.
  - **AC Ref:** AC 2.3 (Duplicate error message construction).

- [ ] **Abstract error categorization logic**
  - **Action:** Create utility functions or mappings to standardize error message categorization, replacing string-based conditionals with more maintainable approaches.
  - **Depends On:** None.
  - **AC Ref:** AC 2.4 (String-based error categorization).

## Deprecated Code and Documentation
- [ ] **Complete migration from deprecated utility functions**
  - **Action:** Identify all uses of deprecated functions in `consoleUtils.ts` and update them to use the new error system directly. Remove the deprecated functions once migration is complete.
  - **Depends On:** None.
  - **AC Ref:** AC 3.1 (Deprecated functions in `consoleUtils.ts`).

- [ ] **Review and integrate documentation from removed files**
  - **Action:** Review content from removed files (`BEST-PRACTICES.md`, `TASK.md`) and integrate relevant information into appropriate documentation (README.md, code comments, etc.).
  - **Depends On:** None.
  - **AC Ref:** AC 3.2 (Removal of documentation).

## Test Improvements
- [ ] **Ensure proper cleanup of mocks in tests**
  - **Action:** Review all tests that use global mocks (e.g., process.cwd) and ensure they properly restore the original state using finally blocks or afterEach hooks.
  - **Depends On:** None.
  - **AC Ref:** AC 4.1 (Global mocks in tests).

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** The placeholder test file might be part of work-in-progress and not intended to be fixed in this iteration.
  - **Context:** The code review mentions a placeholder test in `cli-command-error-handling.test.ts`, but it's unclear if this is a planned item for future work.

- [ ] **Issue/Assumption:** The level of backward compatibility needed for deprecated functions is not specified.
  - **Context:** The code review mentions deprecated functions in `consoleUtils.ts`, but doesn't specify how long they need to be maintained for backward compatibility.

- [ ] **Issue/Assumption:** The code review doesn't specify whether a complete refactoring of the error message categorization logic is necessary or if incremental improvements are acceptable.
  - **Context:** The "Error Categorization Logic" critical issue mentions catch-all blocks that categorize errors based on message content.