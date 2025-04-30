# Todo: Mock Consolidation (R1)

*Priority: P1 (High Impact, Medium Risk)*

## Mock Consolidation (R1)
- [ ] **T010 · Chore · P1**: create `internal/testutil` package directory
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 1
    - **Action:**
        1. Create the `internal/testutil` directory.
        2. Add a `.gitkeep` file if needed initially.
    - **Done‑when:**
        1. `internal/testutil/` directory exists in the repository.
    - **Depends‑on:** none
- [ ] **T011 · Feature · P1**: create/enhance `testutil.MockLogger`
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 1
    - **Action:**
        1. Implement `MockLogger` in `internal/testutil/mocklogger.go` satisfying `logutil.LoggerInterface` and `auditlog.AuditLogger`.
        2. Add assertion helpers (e.g., `GetMessages`, `AssertCalledWith`).
    - **Done‑when:**
        1. `testutil.MockLogger` exists and implements required interfaces.
        2. Basic tests for the mock logger pass.
    - **Depends‑on:** [T010]
- [ ] **T012 · Feature · P1**: create/enhance `testutil.MockAPIService`
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 1
    - **Action:**
        1. Implement `MockAPIService` in `internal/testutil/mock_apiservice.go` satisfying `internal/thinktank/interfaces.APIService`.
        2. Add necessary methods and assertion helpers.
    - **Done‑when:**
        1. `testutil.MockAPIService` exists and implements the required interface.
        2. Basic tests for the mock API service pass.
    - **Depends‑on:** [T010]
- [ ] **T013 · Feature · P1**: create/enhance `testutil.MockFileWriter`
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 1
    - **Action:**
        1. Implement `MockFileWriter` in `internal/testutil/mock_filewriter.go` satisfying `internal/fileutil.FileWriter`.
        2. Add methods/helpers for checking written content.
    - **Done‑when:**
        1. `testutil.MockFileWriter` exists and implements the required interface.
        2. Basic tests for the mock file writer pass.
    - **Depends‑on:** [T010]
- [ ] **T014 · Feature · P1**: create/enhance `testutil.MockExternalAPICaller`
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 1
    - **Action:**
        1. Implement `MockExternalAPICaller` in `internal/testutil/mock_externalapicaller.go` satisfying relevant HTTP client interface (e.g., one defined for T026).
        2. Add necessary methods and assertion helpers.
    - **Done‑when:**
        1. `testutil.MockExternalAPICaller` exists and implements the required interface.
        2. Basic tests for the mock caller pass.
    - **Depends‑on:** [T010]
- [ ] **T015 · Test · P1**: refactor tests in `cmd/thinktank/` to use shared `testutil` mocks
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Steps 2-4
    - **Action:**
        1. Iterate through `_test.go` files in `cmd/thinktank/`.
        2. Replace local mock definitions/usage with `testutil` equivalents.
        3. Update assertions as needed.
    - **Done‑when:**
        1. Tests in `cmd/thinktank/` use shared mocks.
        2. `go test ./cmd/thinktank/...` passes.
    - **Depends‑on:** [T011, T012, T013, T014]
- [ ] **T016 · Test · P1**: refactor tests in `internal/integration/` to use shared `testutil` mocks
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Steps 2-4
    - **Action:**
        1. Iterate through `_test.go` files in `internal/integration/`.
        2. Replace local mock definitions/usage with `testutil` equivalents.
        3. Update assertions as needed.
    - **Done‑when:**
        1. Tests in `internal/integration/` use shared mocks.
        2. `go test ./internal/integration/...` passes.
    - **Depends‑on:** [T011, T012, T013, T014]
- [ ] **T017 · Test · P1**: refactor tests in `internal/providers/` to use shared `testutil` mocks
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Steps 2-4
    - **Action:**
        1. Iterate through `_test.go` files in `internal/providers/`.
        2. Replace local mock definitions/usage with `testutil` equivalents.
        3. Update assertions as needed.
    - **Done‑when:**
        1. Tests in `internal/providers/` use shared mocks.
        2. `go test ./internal/providers/...` passes.
    - **Depends‑on:** [T011, T012, T013, T014]
- [ ] **T018 · Test · P1**: refactor remaining tests across codebase to use shared `testutil` mocks
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Steps 2-4
    - **Action:**
        1. Identify and refactor any remaining `_test.go` files using duplicated mocks.
        2. Replace local mock definitions/usage with `testutil` equivalents.
        3. Update assertions as needed.
    - **Done‑when:**
        1. All identified remaining tests use shared mocks.
        2. `go test ./...` passes.
    - **Depends‑on:** [T011, T012, T013, T014]
- [ ] **T019 · Refactor · P1**: delete redundant local mock definitions and files
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 5
    - **Action:**
        1. Identify and `git rm` any local mock files (e.g., `internal/fileutil/mock_logger.go`) made redundant.
        2. Remove struct definitions for mocks within `_test.go` files that were replaced.
    - **Done‑when:**
        1. Redundant mock files and inline definitions are removed.
        2. `go build ./...` succeeds.
        3. `go test ./...` passes.
    - **Depends‑on:** [T015, T016, T017, T018]

### Clarifications & Assumptions
- [ ] **Issue:** Define exact list of interfaces to consolidate into `internal/testutil` mocks.
    - **Context:** SHRINK_PLAN.md § 3.1 (R1), Implementation Step 1
    - **Blocking?:** yes (for T011-T014)
