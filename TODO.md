# TODO

## Design Principles (CORE_PRINCIPLES.md)
- [x] **Enhance setup.sh for pre-commit verification:**
  - **Action:** Modify the `scripts/setup.sh` script to check if `pre-commit` is installed in the user's environment. If not found, prompt the user to install it automatically or provide instructions.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: `setup.sh` verifies pre-commit installation)

- [x] **Remove TestImplementation struct from orchestrator.go:**
  - **Action:** Delete the `TestImplementation` struct and its associated methods (lines 303-319) from `internal/architect/orchestrator/orchestrator.go` as it appears to be leftover test code.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: Test code is removed from production file)

- [x] **Refactor large test file cli_test.go:**
  - **Action:** Break down the `cmd/architect/cli_test.go` file into multiple smaller, focused test files (e.g., `cli_args_test.go`, `cli_logging_test.go`, `cli_validation_test.go`). Each new file should test a specific aspect of the CLI functionality.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: `cli_test.go` is refactored into smaller, focused files)

## Architectural Patterns (ARCHITECTURE_GUIDELINES.md)
- [x] **Add pre-commit check step to CI workflow:**
  - **Action:** Modify the `.github/workflows/ci.yml` file to include a new step that executes `pre-commit run --all-files`. This ensures that pre-commit checks are enforced in the CI pipeline.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: CI workflow runs pre-commit checks)

## Code Quality (CODING_STANDARDS.md)
- [ ] **Address unused code warnings:**
  - **Action:** Run `golangci-lint` and identify all instances of unused code warnings. Remove the unused code or add comments explaining why it needs to be kept (e.g., `//nolint:unused` with justification).
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: `golangci-lint` reports no unused code warnings)

- [ ] **Verify XML escaping logic in prompt.go:**
  - **Action:** Review the changes made to the `EscapeContent` function in `internal/architect/prompt/prompt.go` (lines 14-15). Determine if the removal of XML escaping was intentional and correct for the LLM prompt structure. Revert the changes if escaping is necessary. Add tests to cover escaping logic.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: XML escaping logic is confirmed correct or fixed, and tested)

## Test Quality (TESTING_STRATEGY.md)
- [ ] **Verify test coverage meets 80% threshold:**
  - **Action:** Run the test suite with coverage analysis (`go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`). Check if the total coverage meets or exceeds the 80% threshold defined in `.github/workflows/ci.yml`.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: Current test coverage percentage is determined)

- [ ] **Add tests for logger adapter methods:**
  - **Action:** Implement unit tests for the standard logger adapter methods (Debug, Info, Warn, Error, Fatal) in `internal/logutil/logutil.go` (`StdLoggerAdapter`). Ensure adequate coverage for these methods.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: `StdLoggerAdapter` methods have sufficient test coverage)

- [ ] **Add tests for integration test helpers:**
  - **Action:** Implement unit tests for the helper functions defined in `internal/integration/test_helpers.go`. Focus on covering the logic within these helpers to increase their coverage.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: Test helper functions in `test_helpers.go` have sufficient test coverage)

## Documentation Practices (DOCUMENTATION_APPROACH.md)
- [ ] **Enhance pre-commit documentation in hooks/README.md:**
  - **Action:** Update `hooks/README.md` to provide more context on the benefits of using `pre-commit` hooks. Explain how they help maintain code quality, enforce standards, and improve the development workflow, aligning with project principles.
  - **Depends On:** None
  - **AC Ref:** N/A (Implied: `hooks/README.md` clearly explains the benefits of pre-commit)
