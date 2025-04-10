# TODO

## Fixing Critical Issues

- [x] **Fix Broken Integration Tests**
  - **Action:** Update integration tests in `internal/integration/integration_test.go` to work with the refactored architecture. Update test expectations for the new `--task-file` requirement, removing tests that use `--task` flag only. Simplify `main_adapter.go` or test `architect.Main` more directly.
  - **Depends On:** None
  - **AC Ref:** Integration tests passing

- [x] **Move Core Logic from cmd/architect to internal/architect**
  - **Action:** Create a new package `internal/architect` and move the component files (api.go, context.go, output.go, prompt.go, token.go) there. Update imports across the codebase. Keep minimal code in cmd/architect for the entry point. 
  - **Depends On:** None
  - **AC Ref:** Package structure aligns with Go conventions

- [ ] **Fix Skipped TestGenerateAndSavePlanWithConfig**
  - **Action:** Investigate and fix the "package reference issues" in `cmd/architect/output_test.go` that prevent `TestGenerateAndSavePlanWithConfig` from running.
  - **Depends On:** None
  - **AC Ref:** All non-integration tests passing

- [ ] **Remove Temporary/Backup Files from Git**
  - **Action:** Remove temporary files (main.go.bak and test-results/*) from the repository and update `.gitignore` to prevent future occurrences.
  - **Depends On:** None
  - **AC Ref:** Clean git status

## Secondary Improvements

- [ ] **Refactor Transitional main_test.go**
  - **Action:** Refactor or remove tests in `main_test.go` that are now covered by component unit tests. Consider moving remaining integration tests to the integration package.
  - **Depends On:** Fix Broken Integration Tests 
  - **AC Ref:** Simplified test structure

- [ ] **Ensure Consistent Flag Status Documentation**
  - **Action:** Update README and CLI usage message to be perfectly aligned on the status of the `--task` flag. If `--task` is non-functional, remove it entirely from usage/help text.
  - **Depends On:** None
  - **AC Ref:** Consistent documentation

- [ ] **Clean Up Duplicated Logging Messages**
  - **Action:** Review logging statements in `cmd/architect/context.go` to avoid duplication at different levels.
  - **Depends On:** None 
  - **AC Ref:** Clean, non-redundant logging

## Finalization

- [ ] **End-to-End Verification**
  - **Action:** After all the above tasks are completed, run end-to-end tests to verify the tool functions correctly with the new architecture.
  - **Depends On:** All other tasks
  - **AC Ref:** Working application

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Transitional Files Approach:** Is `legacy_main.go` intended to be kept for reference or should it be removed?
  - **Context:** The file has go:build ignore tags but is still tracked in git.

- [ ] **Expected Integration Test Behavior:** Should integration tests only use `--task-file` flag, or support both methods?
  - **Context:** Current integration tests use deprecated `--task` flag.

- [ ] **Development Branch Restrictions:** Are there restrictions on creating new packages (internal/architect) in the current branch?
  - **Context:** Moving core logic to internal/architect may require significant changes.