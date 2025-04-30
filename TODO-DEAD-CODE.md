# Todo: Dead Code Elimination (R2)

*Priority: P0 (High Impact, Low Risk)*

## Dead Code Elimination (R2)
- [ ] **T001 · Chore · P0**: verify registry api service sufficiency for legacy apiService removal
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 2
    - **Action:**
        1. Analyze functionality of legacy `apiService` (`internal/thinktank/api.go`, `DetectProviderFromModel`).
        2. Confirm `internal/thinktank/registry_api.go` covers all necessary functionality or plan migration.
        3. Document findings (e.g., PR comment).
    - **Done‑when:**
        1. Analysis complete, decision made on `apiService` removal feasibility.
    - **Depends‑on:** none
- [ ] **T002 · Refactor · P0**: remove obsolete token management files
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 1
    - **Action:**
        1. `git rm internal/thinktank/token.go internal/thinktank/registry_token.go internal/thinktank/token_stubs.go internal/thinktank/modelproc/token_stubs.go`
        2. Remove any remaining imports or references.
    - **Done‑when:**
        1. Files are removed.
        2. `go build ./...` succeeds.
    - **Depends‑on:** none
- [ ] **T003 · Refactor · P0**: remove deprecated `internal/runutil/` package
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 1
    - **Action:**
        1. `git rm -rf internal/runutil/`
        2. Remove any remaining imports or references.
    - **Done‑when:**
        1. Directory is removed.
        2. `go build ./...` succeeds.
    - **Depends‑on:** none
- [ ] **T004 · Chore · P0**: remove all `*.go.disabled` files
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 1
    - **Action:**
        1. Find and `git rm` all files matching `*.go.disabled`.
    - **Done‑when:**
        1. No `.go.disabled` files remain in the repository.
    - **Depends‑on:** none
- [ ] **T005 · Refactor · P0**: remove functions marked `//nolint:unused`
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 1
    - **Action:**
        1. Search for `//nolint:unused` comments associated with functions/methods.
        2. Delete the identified unused functions/methods.
    - **Done‑when:**
        1. Marked unused functions are removed.
        2. `go build ./...` succeeds.
    - **Depends‑on:** none
- [ ] **T006 · Refactor · P0**: remove commented-out code blocks
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 1
    - **Action:**
        1. Search codebase for significant blocks of commented-out Go code.
        2. Delete these blocks.
    - **Done‑when:**
        1. Commented-out code blocks are removed.
        2. `go build ./...` succeeds.
    - **Depends‑on:** none
- [ ] **T007 · Refactor · P0**: remove legacy `apiService` and related code
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 2
    - **Action:**
        1. `git rm internal/thinktank/api.go internal/thinktank/api_test.go`
        2. Remove `DetectProviderFromModel` function and calls.
        3. Remove any remaining imports or references to the legacy `apiService`.
    - **Done‑when:**
        1. Legacy API service files and references are removed.
        2. `go build ./...` succeeds.
    - **Depends‑on:** [T001]
- [ ] **T008 · Chore · P0**: run `go mod tidy` after code removal
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 4
    - **Action:**
        1. Execute `go mod tidy`.
        2. Commit changes to `go.mod` and `go.sum`.
    - **Done‑when:**
        1. `go mod tidy` completes successfully.
        2. `go.mod` and `go.sum` are updated.
    - **Depends‑on:** [T002, T003, T004, T005, T006, T007]
- [ ] **T009 · Bugfix · P0**: fix build and test errors after dead code removal
    - **Context:** SHRINK_PLAN.md § 3.2 (R2), Implementation Step 5
    - **Action:**
        1. Run `go build ./...` and `go test ./...`.
        2. Fix any compilation or test failures resulting from T002-T007.
    - **Done‑when:**
        1. `go build ./...` succeeds.
        2. `go test ./...` passes.
        3. `./internal/e2e/run_e2e_tests.sh` passes.
    - **Depends‑on:** [T008]
