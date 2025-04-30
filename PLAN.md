```markdown
# Remediation Plan – Sprint <n>

## Executive Summary
This plan addresses critical code review findings that currently block merges and introduce significant security, stability, and maintainability risks. We prioritize unblocking the CI/CD pipeline by fixing fundamental code errors (duplicated errors, production test stubs, legacy logic) and restoring critical automated safeguards (E2E tests, secret leakage detection). Subsequent steps focus on reinstating broader test coverage and improving documentation hygiene, following a sequence optimized for dependency resolution and impact reduction.

## Strike List
| Seq | CR‑ID | Title                                                        | Effort | Owner?   |
|-----|-------|--------------------------------------------------------------|--------|----------|
| 1   | cr-03 | Remove Duplicated Sentinel Errors                            | s      | backend  |
| 2   | cr-04 | Remove Test Stubs from Production Code                       | s      | backend  |
| 3   | cr-08 | Remove Legacy Provider Detection Logic                     | s      | backend  |
| 4   | cr-11 | Remove/Centralize Local `ProviderType` Enum                  | xs     | backend  |
| 5   | cr-09 | Remove Dead `min` Helper Function                            | xs     | backend  |
| 6   | cr-02 | Reinstate Secret Leakage Detection Tests                     | m      | backend  |
| 7   | cr-07 | Add Justifications for Remaining `//nolint:unused`           | s      | backend  |
| 8   | cr-05 | Restore Essential Provider/Adapter Boundary Tests            | m/l    | backend  |
| 9   | cr-01 | Investigate & Fix E2E Test Suite Failure                     | m/l    | backend  |
| 10  | cr-06 | Consolidate & Simplify Project Documentation/Tasks           | s      | docs     |
| 11  | cr-10 | Remove Duplicate BACKLOG Entry                             | xs     | docs     |

## Detailed Remedies

### cr-03 Remove Duplicated Sentinel Errors
- **Problem:** Sentinel errors are duplicated in `internal/thinktank/registry_api.go`, conflicting with the intended canonical definitions in `internal/llm/errors.go`.
- **Impact:** Breaks `errors.Is` checks, making robust error handling impossible (logic bomb); creates multiple sources of truth, violating Simplicity (DRY) and Centralized Error Handling.
- **Options:**
    1.  **Remove Local, Use Central:** Delete local vars, import `llm/errors.go`, replace usages. (Simplest, aligns with intent).
    2.  **Alias Local to Central:** Define local vars as aliases (`var ErrEmptyResponse = llm.ErrXxx`), then refactor usages later. (Adds temporary indirection).
    3.  **Refactor Error Handling:** Overhaul error creation/checking to use a custom system or always wrap central errors. (Overkill for this issue).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                |
    |----------------------------|---------|-------------------------------------|
    | Simplicity                 | ✔       | Single source of truth for errors   |
    | Modularity                 | ✔       | Centralized error definitions       |
    | Testability                | ✔       | `errors.Is` works predictably       |
    | Coding Std                 | ✔       | DRY, standard Go error handling     |
    | Security                   | N/A     |                                     |
- **Recommendation:** Option 1. Immediately remove duplicates and enforce use of the central package.
- **Effort:** s
- **Chosen Fix:** Remove local definitions and use canonical ones from `internal/llm/errors.go`.
- **Steps:**
    1.  Delete the `var (...)` block defining sentinel errors (lines 16-30) from `internal/thinktank/registry_api.go`.
    2.  Add/update the import for `github.com/phrazzld/thinktank/internal/llm`.
    3.  Replace all local usages (e.g., `ErrEmptyResponse`) with the corresponding imported versions (e.g., `llm.ErrEmptyResponse`). If specific errors don't exist centrally, define them *once* in `llm/errors.go`.
    4.  Verify `go test ./...` passes and error handling logic behaves as expected.
- **Done-When:** Duplicate error definitions removed, `registry_api.go` uses errors exclusively from `llm`, `errors.Is` checks function correctly, tests pass.

### cr-04 Remove Test Stubs from Production Code
- **Problem:** Test-only compatibility stubs (`TokenResult`, `TokenManager`, etc.) were added to production packages (`internal/thinktank`, `internal/thinktank/modelproc`).
- **Impact:** Introduces dead, misleading code into the main application (tech-debt anchor); violates Simplicity and Strict Separation of Concerns; provides a false sense of test coverage.
- **Options:**
    1.  **Remove Stubs & Obsolete Tests:** Identify tests requiring stubs, confirm they test removed functionality, delete both stubs and tests. (Cleanest).
    2.  **Move Stubs to Test Helpers:** Relocate stub code to `_test.go` files or a dedicated test helper package. (Keeps potentially obsolete tests compiling).
    3.  **Refactor Tests:** Modify tests to not rely on the specific stub implementations, potentially using interfaces or mocks. (May be complex if tests are truly obsolete).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                     |
    |----------------------------|---------|------------------------------------------|
    | Simplicity                 | ✔       | No dead/test code in production          |
    | Modularity                 | ✔       | Strict separation of prod/test code      |
    | Testability                | ✔       | Tests reflect actual production code     |
    | Coding Std                 | ✔       | No test shims in application logic       |
    | Security                   | N/A     |                                          |
- **Recommendation:** Option 1. Remove the stubs and the obsolete tests that required them.
- **Effort:** s
- **Chosen Fix:** Remove stubs and refactor/delete obsolete tests.
- **Steps:**
    1.  Identify specific `*_test.go` files that fail to compile after deleting the stub files/variables.
    2.  Analyze each failing test: Delete tests solely verifying the removed token management logic. Refactor tests that have other value to work without the stubs.
    3.  Delete the stub files: `internal/thinktank/modelproc/token_fixes.go`, `internal/thinktank/token_fixes.go`.
    4.  Delete the associated `NewTokenManagerWithClient` variable.
    5.  Verify `go test ./...` passes.
- **Done-When:** Stub files and variables are deleted, no production code references them, obsolete tests are removed/refactored, all remaining tests pass.

### cr-08 Remove Legacy Provider Detection Logic
- **Problem:** Fallback provider detection logic (`detectProviderFromModelName`, `createLLMClientFallback`) duplicates removed logic and bypasses the registry using fragile string matching. *(Reviewer Note: Severity upgraded from MEDIUM to HIGH due to violation of core architectural principles - Simplicity/DRY, Explicit > Implicit, Single Source of Truth).*
- **Impact:** Undermines the registry's purpose (tech-debt anchor); risks inconsistent behavior; violates Simplicity (DRY) and Explicit > Implicit principles.
- **Options:**
    1.  **Remove Fallback, Enforce Registry:** Delete detection functions, modify `InitLLMClient` to return specific error if model not in registry. (Enforces design).
    2.  **Move Fallback to Test Helper:** Keep logic but only for specific legacy tests, remove from production path. (Adds complexity).
    3.  **Deprecate Fallback:** Mark functions as deprecated, log warnings, plan removal later. (Delays necessary fix).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                   |
    |----------------------------|---------|----------------------------------------|
    | Simplicity                 | ✔       | Removes redundant/implicit logic       |
    | Modularity                 | ✔       | Registry is the single source of truth |
    | Testability                | ✔       | Behavior is explicit and predictable   |
    | Coding Std                 | ✔       | Explicit > Implicit, DRY               |
    | Security                   | N/A     |                                        |
- **Recommendation:** Option 1. Remove the fallback logic entirely and enforce registry usage.
- **Effort:** s
- **Chosen Fix:** Remove fallback logic and enforce registry usage.
- **Steps:**
    1.  Delete the `createLLMClientFallback` function from `internal/thinktank/registry_api.go`.
    2.  Delete the `detectProviderFromModelName` function from `internal/thinktank/registry_api.go`.
    3.  Modify the `InitLLMClient` function: Remove the call to the fallback logic in the error path of `s.registry.GetModel(modelName)`.
    4.  Ensure a specific error (e.g., `llm.ErrModelNotFound`) is returned when the model isn't found in the registry.
    5.  Verify behavior and tests for non-existent models.
- **Done-When:** Fallback functions are deleted, `InitLLMClient` relies solely on the registry, appropriate error returned for non-existent models, tests pass.

### cr-11 Remove/Centralize Local `ProviderType` Enum
- **Problem:** A local `ProviderType` enum exists in `internal/thinktank/registry_api.go`, likely only used by the removed legacy detection logic (cr-08).
- **Impact:** Dead code/duplication after removing fallback logic, violates Simplicity (DRY).
- **Recommendation:** Remove the enum as it becomes unused after fixing cr-08.
- **Effort:** xs
- **Chosen Fix:** Remove the local enum.
- **Steps:**
    1.  Confirm `ProviderType` enum (lines 206-215) in `internal/thinktank/registry_api.go` is unused after removing fallback logic (cr-08).
    2.  Delete the `ProviderType` type definition and its constants.
    3.  Verify `go test ./...` passes.
- **Done-When:** Local `ProviderType` enum is removed, code compiles and tests pass.

### cr-09 Remove Dead `min` Helper Function
- **Problem:** The custom `min` function in `internal/thinktank/registry_api.go` is unused dead code.
- **Impact:** Code clutter, violates Simplicity. Go 1.21+ provides standard alternatives.
- **Recommendation:** Delete the unused function.
- **Effort:** xs
- **Chosen Fix:** Delete the unused function.
- **Steps:**
    1.  Delete the `min` function definition (lines 32-35) from `internal/thinktank/registry_api.go`.
    2.  Ensure `go.mod` specifies `go 1.21` or higher.
    3.  Verify `go test ./...` passes.
- **Done-When:** The `min` function is deleted, code compiles and tests pass.

### cr-02 Reinstate Secret Leakage Detection Tests
- **Problem:** Tests specifically designed to detect secrets (API keys, etc.) in logs were deleted.
- **Impact:** Removes critical automated security safeguard, drastically increasing risk of accidental secret leakage (security hole); violates Logging Strategy and Security Considerations.
- **Options:**
    1.  **Restore Disabled Tests:** Find deleted `.go.disabled` files, rename to `_test.go`, update if needed. (Fastest if code is recoverable/relevant).
    2.  **Reimplement Active Tests:** Create new `*_test.go` files using `logutil.WithSecretDetection` or similar, ensuring coverage for each provider. (Most robust).
    3.  **Integrate into Existing Tests:** Add secret detection checks to existing provider integration tests. (May bloat existing tests).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                          |
    |----------------------------|---------|-----------------------------------------------|
    | Simplicity                 | ✔       | Focused tests for specific security concern   |
    | Modularity                 | ✔       | Tests specific provider logging behavior      |
    | Testability                | ✔       | Automated verification of logging policy      |
    | Coding Std                 | ✔       | Adheres to testing strategy                   |
    | Security                   | ✔       | Directly enforces "NEVER log sensitive info"  |
- **Recommendation:** Option 2. Reinstate or create new, active tests for *each* provider integration, ensuring they run in CI.
- **Effort:** m
- **Chosen Fix:** Reinstate/create active tests verifying logs for secret absence.
- **Steps:**
    1.  Identify all provider integrations (OpenAI, Gemini, OpenRouter, etc.).
    2.  For each provider, create/restore a test file (e.g., `internal/providers/<provider>/provider_secrets_test.go`).
    3.  Write test cases triggering operations prone to logging secrets (client creation, errors).
    4.  Use a logging interceptor (`logutil.WithSecretDetection` or similar) to capture logs.
    5.  Assert that captured logs **do not** contain API keys, tokens, or credentials.
    6.  Ensure these tests run as part of the standard `go test ./...` suite.
- **Done-When:** Automated tests exist for each provider verifying absence of secrets in logs, tests run in CI, tests pass.

### cr-07 Add Justifications for Remaining `//nolint:unused`
- **Problem:** Existing `//nolint:unused` suppressions lack mandatory inline justification comments.
- **Impact:** Violates Coding Standards, hides rationale for suppression, potentially masking dead code.
- **Options:**
    1.  **Justify Inline:** Add a comment on the same line explaining the necessity for each suppression. (Standard compliant).
    2.  **Remove Unused Code:** If the suppression is no longer needed or the code is truly dead, remove both. (Reduces clutter).
    3.  **Centralized Justification:** Refer to a central document (e.g., `nolint_rationale.md`). (Not standard, less discoverable).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                         |
    |----------------------------|---------|----------------------------------------------|
    | Simplicity                 | ✔       | Clear rationale at point of suppression      |
    | Modularity                 | N/A     |                                              |
    | Testability                | N/A     |                                              |
    | Coding Std                 | ✔       | Adheres to suppression justification policy  |
    | Security                   | N/A     |                                              |
- **Recommendation:** Option 1 + 2. Add inline justifications or remove the unused code/suppression.
- **Effort:** s
- **Chosen Fix:** Add inline justification comments for all intentionally kept `//nolint:unused`.
- **Steps:**
    1.  Search the codebase for all instances of `//nolint:unused`.
    2.  For each instance: Verify necessity. If needed, add comment on the same line: `//nolint:unused // Reason: Kept for future integration X (T123)`. If not needed, remove suppression and code.
    3.  Run `golangci-lint` to confirm compliance.
- **Done-When:** Every remaining `//nolint:unused` suppression has an adjacent, clear justification comment.

### cr-05 Restore Essential Provider/Adapter Boundary Tests
- **Problem:** Mass deletion of tests leaves critical interfaces (API services, adapters, providers) unverified.
- **Impact:** High risk of regressions (tech-debt anchor); impossible to ensure interface contracts are met; violates Testability and Testing Strategy.
- **Options:**
    1.  **Restore Deleted Tests:** `git checkout <commit>^ -- <path>` for all deleted test files. (May restore obsolete tests).
    2.  **Reimplement Essential Contract Tests:** Identify key interfaces/boundaries, write new tests focusing *only* on their contracts (inputs, outputs, errors). (Most targeted).
    3.  **Write High-Level Integration Tests:** Test components through broader use cases. (Less focused on specific contracts).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                          |
    |----------------------------|---------|-----------------------------------------------|
    | Simplicity                 | ✔       | Focused tests on contracts, not implementation|
    | Modularity                 | ✔       | Verifies interface boundaries                 |
    | Testability                | ✔       | Restores automated verification               |
    | Coding Std                 | ✔       | Standard testing practices                    |
    | Security                   | N/A     |                                               |
- **Recommendation:** Option 2. Restore or reimplement essential boundary tests focusing on the *contracts* of core interfaces.
- **Effort:** m/l
- **Chosen Fix:** Restore/reimplement essential boundary tests focusing on contracts.
- **Steps:**
    1.  Identify key interfaces/adapters (`APIService`, `FileWriter`, `Provider`, etc.).
    2.  Prioritize tests verifying contracts: parameter handling, error categorization, expected outputs/side-effects.
    3.  Review deleted tests (`git show <commit>^:<path>`) for relevant cases or restore/rewrite tests (e.g., `file_adapter_test.go`).
    4.  Focus on testing interaction logic (e.g., `CreateClient` behavior based on config).
    5.  Ensure adequate coverage for restored/new tests.
- **Done-When:** Key interface contracts and adapter logic have automated tests verifying their behavior, tests run in CI, tests pass.

### cr-01 Investigate & Fix E2E Test Suite Failure
- **Problem:** End-to-end test suite (`./internal/e2e/run_e2e_tests.sh`) is failing, but this failure is ignored, blocking merges improperly.
- **Impact:** Bypasses critical quality gate (regression risk); erodes confidence; violates Automation/Quality Gates and Testing Strategy.
- **Options:**
    1.  **Investigate & Fix Root Cause:** Debug E2E failures, fix underlying code/test/infra issues. (Required).
    2.  **Temporarily Disable Failing Tests:** Mark specific failing E2E tests as skipped with a ticket to fix later. (Risky, hides problems).
    3.  **Improve E2E Reliability:** Refactor flaky tests, improve mock stability, adjust timeouts. (May be part of Option 1).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                         |
    |----------------------------|---------|----------------------------------------------|
    | Simplicity                 | ✔       | Reliable tests are simpler to maintain       |
    | Modularity                 | N/A     |                                              |
    | Testability                | ✔       | E2E is a crucial layer of testing            |
    | Coding Std                 | ✔       | CI pipeline must enforce quality gates       |
    | Security                   | ✔       | E2E can catch integration-related security issues |
- **Recommendation:** Option 1. Investigate the root cause (code changes, test setup, flakiness) and resolve the E2E failures *before* merging. Ensure CI strictly enforces E2E success.
- **Effort:** m/l
- **Chosen Fix:** Investigate root cause and resolve E2E failures before merging.
- **Steps:**
    1.  Execute `./internal/e2e/run_e2e_tests.sh` locally against the branch to reproduce.
    2.  Isolate cause: Recent code changes (cr-03, cr-04, cr-08 fixes might resolve this), test environment/setup, flaky tests.
    3.  Remediate: Fix application code bugs, correct test setup/mocks, refactor flaky tests.
    4.  Verify `./internal/e2e/run_e2e_tests.sh` passes consistently locally.
    5.  Verify CI pipeline passes the E2E stage for this branch. **DO NOT MERGE** until green.
- **Done-When:** Root cause identified, fix implemented, `./internal/e2e/run_e2e_tests.sh` passes reliably locally and in CI against this branch.

### cr-06 Consolidate & Simplify Project Documentation/Tasks
- **Problem:** Proliferation of `TODO-*.md` files, `PLAN.md` wrapped in code block, redundancy across task files creates cognitive overhead.
- **Impact:** Fragments information, hinders tracking, violates Simplicity and Documentation Approach.
- **Options:**
    1.  **Consolidate Markdown:** Merge all tasks into `TODO.md`, refactor `PLAN.md` (remove code block, focus on design), deduplicate `BACKLOG.md`. (Improves current structure).
    2.  **Migrate to Project Tool:** Move all tasks to GitHub Issues/Projects, Jira, etc. Deprecate `TODO.md`/`BACKLOG.md`. (Better tooling, higher initial effort).
    3.  **Hybrid Approach:** Use project tool for active tasks, keep high-level `PLAN.md` and potentially `BACKLOG.md` in repo. (Balances tooling and repo context).
- **Standards Check:**
    | Philosophy                 | Passes? | Note                                         |
    |----------------------------|---------|----------------------------------------------|
    | Simplicity                 | ✔       | Reduces fragmentation and redundancy         |
    | Modularity                 | N/A     |                                              |
    | Testability                | N/A     |                                              |
    | Coding Std                 | N/A     | (Applies to documentation standards)         |
    | Security                   | N/A     |                                              |
- **Recommendation:** Option 1. Consolidate into a single `TODO.md`, refactor `PLAN.md`, deduplicate `BACKLOG.md`.
- **Effort:** s
- **Chosen Fix:** Consolidate tasks, refactor `PLAN.md`, deduplicate backlog.
- **Steps:**
    1.  Merge actionable tasks from all `TODO-*.md` files into the main `TODO.md`. Delete `TODO-*.md` files.
    2.  Refactor `PLAN.md`: Remove surrounding ```markdown block. Rename if desired (e.g., `DESIGN_NOTES.md`). Focus content on high-level design/architecture.
    3.  Deduplicate entries between `BACKLOG.md` and the consolidated `TODO.md`.
- **Done-When:** All tasks consolidated into one `TODO.md` or project tool, `TODO-*.md` files removed, `PLAN.md` refactored, `BACKLOG.md` deduplicated.

### cr-10 Remove Duplicate BACKLOG Entry
- **Problem:** Task T25 appears multiple times in `BACKLOG.md`.
- **Impact:** Minor confusion in backlog tracking, violates Simplicity (DRY).
- **Recommendation:** Remove the duplicate entry.
- **Effort:** xs
- **Chosen Fix:** Remove the duplicate entry.
- **Steps:**
    1.  Open `BACKLOG.md`.
    2.  Delete one of the duplicate entries for Task T25.
- **Done-When:** `BACKLOG.md` contains only one entry for T25.

## Standards Alignment
- This plan directly remedies violations across key development philosophies:
    - **Simplicity:** Fixes target DRY violations (cr-03, cr-10, cr-11), dead code (cr-09), unnecessary complexity/implicitness (cr-08), production test code (cr-04), and documentation overhead (cr-06).
    - **Modularity / Separation of Concerns:** Fixes target duplicated errors breaking centralization (cr-03), production test code violating boundaries (cr-04), fallback logic bypassing registry (cr-08), and restores boundary tests (cr-05).
    - **Testability / Testing Strategy:** Fixes target E2E failures blocking quality gates (cr-01), missing security tests (cr-02), production code dictated by tests (cr-04), deleted boundary tests (cr-05), and mandates test restoration.
    - **Coding Standards:** Fixes target unjustified lint suppressions (cr-07) and use of standard library alternatives (cr-09).
    - **Security:** Fixes target removal of secret detection tests (cr-02) and mandates their restoration as a critical safeguard.
    - **Automation / Quality Gates:** Fixes target merge blocking E2E failures (cr-01), ensuring CI enforces quality.
- Executing this plan will restore adherence to mandatory standards for structure, testing, security, and maintainability.

## Validation Checklist
- [ ] All automated tests (unit, integration, E2E) pass reliably in CI (`go test -race ./...`, `./internal/e2e/run_e2e_tests.sh`).
- [ ] `golangci-lint` passes with no new warnings or *unjustified* errors/suppressions.
- [ ] `govulncheck` passes with no Critical or High severity vulnerabilities.
- [ ] Secret detection tests (`cr-02`) confirm no sensitive data in logs for all providers.
- [ ] Manual check confirms duplicated errors (`cr-03`) are removed and centralized.
- [ ] Manual check confirms production test stubs (`cr-04`) are removed.
- [ ] Manual check confirms fallback logic (`cr-08`) and associated enum (`cr-11`) are removed; registry is enforced.
- [ ] Manual check confirms `//nolint:unused` directives (`cr-07`) have justifications or were removed.
- [ ] Manual check confirms documentation/task files (`cr-06`, `cr-10`) are consolidated and cleaned.
- [ ] Code coverage meets or exceeds project thresholds for critical components.
```
