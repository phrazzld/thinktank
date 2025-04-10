# DONE

## Fixing Critical Issues

- [x] **Fix Broken Integration Tests** (2025-04-10)
  - **Action:** Update integration tests in `internal/integration/integration_test.go` to work with the refactored architecture. Update test expectations for the new `--task-file` requirement, removing tests that use `--task` flag only. Simplify `main_adapter.go` or test `architect.Main` more directly.
  - **Depends On:** None
  - **AC Ref:** Integration tests passing