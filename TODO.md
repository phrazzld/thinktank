# TODO

## Code Review Fixes
- [ ] **Task Title:** Refactor Execute and RunInternal to Eliminate Duplication
  - **Action:** Create a shared internal helper function in `internal/architect/app.go` containing the common execution logic (ReadInstructions, GatherContext, CheckTokens, GenerateContent, SaveOutput, audit logging). Update `Execute` and `RunInternal` to call this helper, passing necessary parameters like the API service/client. Ensure all audit logging calls are correctly placed within the helper or handled appropriately by the callers.
  - **Depends On:** None
  - **AC Ref:** Issue 1 (`internal/architect/app.go`)

- [ ] **Task Title:** Define and Use Constants for Audit Log Error Types
  - **Action:** Define string constants for error types (e.g., `ErrorTypeFileIO`, `ErrorTypeAPIError`, `ErrorTypeExecutionError`, `ErrorTypeContextGatheringError`, `ErrorTypeTokenCheckError`, `ErrorTypeTokenLimitExceededError`, `ErrorTypeContentGenerationError`) in `internal/auditlog/entry.go`. Replace all hardcoded error type strings within `auditlog.ErrorInfo` creation in `internal/architect/app.go` (both `Execute` and `RunInternal`, or the refactored helper) with these constants.
  - **Depends On:** None
  - **AC Ref:** Issue 2 (`internal/architect/app.go`)

- [x] **Task Title:** Correct HTML Entity in GenerateContentEnd Audit Log
  - **Action:** In `internal/architect/app.go` (lines ~380, ~612, or corresponding lines after refactoring), replace the incorrect HTML entity `&gt;` with the correct Go operator `>` in the boolean comparison `len(result.SafetyRatings) > 0`.
  - **Depends On:** None
  - **AC Ref:** Issue 3 (`internal/architect/app.go`)
  - **Note:** This issue was already fixed. The code at lines 439 and 908 correctly uses the `>` operator.

- [ ] **Task Title:** Correct HTML Entities in Logger Test Loops
  - **Action:** In `internal/auditlog/logger_test.go` (lines ~401, ~416, ~512, ~515, ~550, ~551), replace the incorrect HTML entity `&lt;` with the correct Go operator `<` in the `for` loop conditions (e.g., `i < 3`, `i < numGoroutines`, `j < entriesPerGoroutine`).
  - **Depends On:** None
  - **AC Ref:** Issue 4 (`internal/auditlog/logger_test.go`)

- [ ] **Task Title:** Correct HTML Entity in Integration Test Helper
  - **Action:** In `internal/integration/test_helpers.go` (line ~176), replace the incorrect HTML entity `&gt;` with the correct Go operator `>` in the `if` condition `len(options) > 0`.
  - **Depends On:** None
  - **AC Ref:** Issue 5 (`internal/integration/test_helpers.go`)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
*(No clarifications needed based on the provided code review document)*