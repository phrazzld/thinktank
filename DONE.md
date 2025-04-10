# DONE

## Fixing Critical Issues

- [x] **Fix Broken Integration Tests** (2025-04-10)
  - **Action:** Update integration tests in `internal/integration/integration_test.go` to work with the refactored architecture. Update test expectations for the new `--task-file` requirement, removing tests that use `--task` flag only. Simplify `main_adapter.go` or test `architect.Main` more directly.
  - **Depends On:** None
  - **AC Ref:** Integration tests passing

- [x] **Move Core Logic from cmd/architect to internal/architect** (2025-04-10)
  - **Action:** Create a new package `internal/architect` and move the component files (api.go, context.go, output.go, prompt.go, token.go) there. Update imports across the codebase. Keep minimal code in cmd/architect for the entry point. 
  - **Depends On:** None
  - **AC Ref:** Package structure aligns with Go conventions

- [x] **Fix Skipped TestGenerateAndSavePlanWithConfig** (2025-04-10)
  - **Action:** Investigate and fix the "package reference issues" in `cmd/architect/output_test.go` that prevent `TestGenerateAndSavePlanWithConfig` from running.
  - **Depends On:** None
  - **AC Ref:** All non-integration tests passing

- [x] **Remove Temporary/Backup Files from Git** (2025-04-10)
  - **Action:** Remove temporary files (main.go.bak and test-results/*) from the repository and update `.gitignore` to prevent future occurrences.
  - **Depends On:** None
  - **AC Ref:** Clean git status

## Secondary Improvements

- [x] **Ensure Consistent Flag Status Documentation** (2025-04-10)
  - **Action:** Update README and CLI usage message to be perfectly aligned on the status of the `--task` flag. If `--task` is non-functional, remove it entirely from usage/help text.
  - **Depends On:** None
  - **AC Ref:** Consistent documentation

- [x] **Clean Up Duplicated Logging Messages** (2025-04-10)
  - **Action:** Review logging statements in `cmd/architect/context.go` to avoid duplication at different levels.
  - **Depends On:** None 
  - **AC Ref:** Clean, non-redundant logging