```markdown
# Todo

## Backend - Core Refactoring
- [x] **T001 · Refactor · P1: remove duplicate sentinel errors in registry_api**
    - **Context:** PLAN.md / cr-03 Remove Duplicated Sentinel Errors
    - **Action:**
        1. Delete the local sentinel error `var` block (lines 16-30) in `internal/thinktank/registry_api.go`.
        2. Ensure `internal/llm` is imported and replace local error usages with canonical versions (e.g., `llm.ErrEmptyResponse`).
        3. Define any missing sentinel errors *only* in `internal/llm/errors.go`.
    - **Done‑when:**
        1. Local error variables are removed from `internal/thinktank/registry_api.go`.
        2. Code compiles and uses only errors from `internal/llm`.
        3. `errors.Is` checks function correctly against central errors.
        4. `go test ./...` passes.
    - **Depends‑on:** none
- [x] **T002 · Refactor · P1: delete test stubs from production code**
    - **Context:** PLAN.md / cr-04 Remove Test Stubs from Production Code / Steps 1, 3, 4
    - **Action:**
        1. Delete stub files: `internal/thinktank/modelproc/token_fixes.go`, `internal/thinktank/token_fixes.go`.
        2. Delete the associated `NewTokenManagerWithClient` variable.
    - **Done‑when:**
        1. Specified stub files and variables are deleted.
        2. No production code references the removed stubs.
        3. Code compiles (tests addressed in T004).
    - **Depends‑on:** none
- [ ] **T003 · Refactor · P1: remove legacy provider detection fallback logic**
    - **Context:** PLAN.md / cr-08 Remove Legacy Provider Detection Logic
    - **Action:**
        1. Delete `createLLMClientFallback` and `detectProviderFromModelName` functions from `internal/thinktank/registry_api.go`.
        2. Modify `InitLLMClient` to remove the fallback call in the error path.
        3. Ensure `InitLLMClient` returns a specific error (e.g., `llm.ErrModelNotFound`) if a model is not found in the registry.
    - **Done‑when:**
        1. Fallback functions are deleted.
        2. `InitLLMClient` relies solely on the registry for model lookup.
        3. Appropriate error is returned for models not found in the registry.
        4. `go test ./...` passes, including tests for non-existent models.
    - **Depends‑on:** none
- [ ] **T005 · Refactor · P2: remove local providertype enum**
    - **Context:** PLAN.md / cr-11 Remove/Centralize Local `ProviderType` Enum
    - **Action:**
        1. Confirm `ProviderType` enum (lines 206-215) in `internal/thinktank/registry_api.go` is unused after T003 completion.
        2. Delete the `ProviderType` type definition and its associated constants.
    - **Done‑when:**
        1. `ProviderType` enum is removed from `internal/thinktank/registry_api.go`.
        2. Code compiles and `go test ./...` passes.
    - **Depends‑on:** [T003]
- [ ] **T006 · Refactor · P2: remove dead min helper function**
    - **Context:** PLAN.md / cr-09 Remove Dead `min` Helper Function
    - **Action:**
        1. Delete the `min` function definition (lines 32-35) from `internal/thinktank/registry_api.go`.
        2. Ensure `go.mod` specifies `go 1.21` or higher.
    - **Done‑when:**
        1. The `min` function is deleted.
        2. Code compiles and `go test ./...` passes.
    - **Depends‑on:** none

## Backend - Testing & CI
- [ ] **T004 · Test · P1: refactor or delete tests broken by stub removal**
    - **Context:** PLAN.md / cr-04 Remove Test Stubs from Production Code / Step 2
    - **Action:**
        1. Identify specific `*_test.go` files that fail to compile or run after completing T002.
        2. Analyze each failing test: Delete tests solely verifying the removed token management logic.
        3. Refactor tests that have other value to work without the stubs (e.g., using production interfaces/mocks).
    - **Done‑when:**
        1. Obsolete tests related to stubs are removed or refactored.
        2. All remaining tests compile and `go test ./...` passes.
    - **Depends‑on:** [T002]
- [ ] **T007 · Test · P0: reinstate secret leakage detection tests**
    - **Context:** PLAN.md / cr-02 Reinstate Secret Leakage Detection Tests
    - **Action:**
        1. For each provider integration (OpenAI, Gemini, OpenRouter, etc.), create/restore a test file (e.g., `internal/providers/<provider>/provider_secrets_test.go`).
        2. Implement test cases using a logging interceptor (`logutil.WithSecretDetection` or similar) that trigger potentially problematic operations (client creation, errors).
        3. Assert that captured logs **do not** contain API keys, tokens, or credentials.
    - **Done‑when:**
        1. Automated tests exist for each provider verifying absence of secrets in logs.
        2. These tests run as part of `go test ./...`.
        3. All secret detection tests pass.
        4. CI runs these tests.
    - **Depends‑on:** none # Best done after core refactors, but not strictly blocked
- [ ] **T008 · Test · P1: restore essential provider/adapter boundary tests**
    - **Context:** PLAN.md / cr-05 Restore Essential Provider/Adapter Boundary Tests
    - **Action:**
        1. Identify key interfaces/adapters (`APIService`, `FileWriter`, `Provider`, etc.) needing contract verification.
        2. Review deleted tests (`git show <commit>^:<path>`) or write new tests focusing on verifying contracts (inputs, outputs, errors, interaction logic like `CreateClient`).
        3. Ensure adequate test coverage for the contracts of restored/newly tested components.
    - **Done‑when:**
        1. Automated tests exist verifying the contracts of key interfaces and adapter logic.
        2. Tests cover parameter handling, error categorization, and expected outputs/side-effects.
        3. Tests run as part of `go test ./...` and pass.
        4. CI runs these tests.
    - **Depends‑on:** none # Best done after core refactors, but not strictly blocked
- [ ] **T009 · Bugfix · P0: investigate and fix e2e test suite failure**
    - **Context:** PLAN.md / cr-01 Investigate & Fix E2E Test Suite Failure
    - **Action:**
        1. Execute `./internal/e2e/run_e2e_tests.sh` locally against the branch incorporating fixes from T001, T003, T004 to reproduce the failure.
        2. Investigate the root cause (remaining code issues, test environment, setup, flaky tests).
        3. Implement the necessary fix (application code, test setup, mocks, test refactoring).
    - **Done‑when:**
        1. Root cause of E2E failure is identified and fixed.
        2. `./internal/e2e/run_e2e_tests.sh` passes consistently locally.
        3. CI pipeline passes the E2E stage reliably for the relevant branch.
    - **Verification:**
        1. Run `./internal/e2e/run_e2e_tests.sh` multiple times locally to check for flakiness.
        2. Observe CI run passing E2E stage.
    - **Depends‑on:** [T001, T003, T004] # Core refactors likely impact E2E

## Code Standards
- [ ] **T010 · Chore · P2: add justifications for remaining //nolint:unused directives**
    - **Context:** PLAN.md / cr-07 Add Justifications for Remaining `//nolint:unused`
    - **Action:**
        1. Search codebase for `//nolint:unused`.
        2. For each instance: Verify necessity. If needed, add comment on the same line: `//nolint:unused // Reason: <concise justification>`. If not needed, remove suppression and associated dead code.
    - **Done‑when:**
        1. Every remaining `//nolint:unused` suppression has an adjacent justification comment.
        2. Unnecessary suppressions and code are removed.
        3. `golangci-lint` passes without related warnings.
    - **Depends‑on:** none

## Documentation
- [ ] **T011 · Chore · P2: consolidate and simplify project documentation/tasks**
    - **Context:** PLAN.md / cr-06 Consolidate & Simplify Project Documentation/Tasks
    - **Action:**
        1. Merge actionable items from all `TODO-*.md` files into this `TODO.md` file and delete the `TODO-*.md` files.
        2. Refactor `PLAN.md`: Remove the surrounding ```markdown code block, optionally rename (e.g., `DESIGN_NOTES.md`), and ensure content focuses on high-level design/architecture.
        3. Review `BACKLOG.md` and remove entries that are now duplicated in this `TODO.md`.
    - **Done‑when:**
        1. All `TODO-*.md` files are removed.
        2. Actionable tasks are consolidated into this `TODO.md`.
        3. `PLAN.md` is refactored (code block removed, content focused).
        4. `BACKLOG.md` is deduplicated against this `TODO.md`.
    - **Depends‑on:** none
- [ ] **T012 · Chore · P3: remove duplicate backlog entry for t25**
    - **Context:** PLAN.md / cr-10 Remove Duplicate BACKLOG Entry
    - **Action:**
        1. Open `BACKLOG.md`.
        2. Delete one instance of the duplicate entry for Task T25.
    - **Done‑when:**
        1. `BACKLOG.md` contains only one entry for T25.
    - **Depends‑on:** none

### Clarifications & Assumptions
- *(No clarifications needed based on the provided plan)*
```
