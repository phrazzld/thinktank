# Todo: Provider Logic & Error Handling Refactoring (R3, R4)

*Priority: P2 (Medium Impact, Medium-High Risk)*

## Provider Logic & Error Handling Refactoring (R3, R4)
- [ ] **T020 · Refactor · P2**: define standard error variables in `internal/llm/errors.go`
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 1
    - **Action:**
        1. Define exported sentinel error variables `ErrEmptyResponse`, `ErrContentFiltered` in `internal/llm/errors.go`.
    - **Done‑when:**
        1. Standard error variables are defined and exported.
        2. `go build ./...` succeeds.
    - **Depends‑on:** none
- [ ] **T021 · Refactor · P2**: refactor `registry_api.go` error checks to use standard errors
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 2
    - **Action:**
        1. Modify error checking functions (e.g., `IsEmptyResponseError`, `IsSafetyBlockedError`) in `internal/thinktank/registry_api.go`.
        2. Replace checks with `errors.Is(err, llm.ErrEmptyResponse)` or `llm.IsCategory(err, llm.CategoryContentFilter)`.
    - **Done‑when:**
        1. Error checking logic uses centralized `llm` errors/categories.
        2. Related tests pass.
    - **Depends‑on:** [T020]
- [ ] **T022 · Refactor · P2**: implement centralized error creation function in `internal/llm/errors.go`
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 2
    - **Action:**
        1. Create `CreateStandardErrorWithMessage(category ErrorCategory, message string, underlying error) *LLMError` in `internal/llm/errors.go`.
        2. Ensure consistent population of `LLMError` fields and error wrapping.
    - **Done‑when:**
        1. Centralized error creation function is implemented.
        2. Unit tests for the creation function pass.
    - **Depends‑on:** none
- [ ] **T023 · Refactor · P2**: refactor provider clients to use centralized error formatting
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 5
    - **Action:**
        1. Identify error creation points in `internal/providers/*/client.go`.
        2. Replace manual `LLMError` creation with calls to the centralized function (T022).
    - **Done‑when:**
        1. Provider clients use the centralized error formatting function.
        2. Provider client tests pass, verifying correct error structure.
    - **Depends‑on:** [T022]
- [ ] **T024 · Feature · P2**: implement shared parameter application utility in `internal/llm/parameters.go`
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 3
    - **Action:**
        1. Create `internal/llm/parameters.go`.
        2. Implement `ApplyParameters(client interface{}, params map[string]interface{}) error` using reflection to map parameters.
    - **Done‑when:**
        1. `ApplyParameters` utility function is implemented.
        2. Unit tests cover various parameter types and edge cases.
    - **Depends‑on:** none
- [ ] **T025 · Refactor · P2**: refactor provider adapters to use `llm.ApplyParameters`
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 4
    - **Action:**
        1. Modify `CreateClient` methods in `internal/providers/*/provider.go`.
        2. Replace provider-specific parameter application logic with calls to `llm.ApplyParameters`.
    - **Done‑when:**
        1. Provider adapters use the shared parameter utility.
        2. Provider adapter tests pass.
    - **Depends‑on:** [T024]
- [ ] **T026 · Refactor · P2**: identify and implement shared http utilities (optional)
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 1 & 5
    - **Action:**
        1. Analyze HTTP patterns in `internal/providers/*/client.go`.
        2. If beneficial, create shared functions/types in `internal/providers/base` or `internal/llm`. Define `ExternalAPICaller` interface.
    - **Done‑when:**
        1. Analysis complete. Shared HTTP utilities implemented if needed. Unit tests pass.
    - **Depends‑on:** none
- [ ] **T027 · Refactor · P2**: refactor provider clients to use shared http utilities (if created)
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 5
    - **Action:**
        1. If shared utilities were created (T026), refactor `internal/providers/*/client.go` to use them.
    - **Done‑when:**
        1. Provider clients utilize shared HTTP utilities where applicable.
        2. Provider client tests pass.
    - **Depends‑on:** [T026]
- [ ] **T028 · Test · P2**: update/add tests for centralized error handling
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 6
    - **Action:**
        1. Review/Update error tests in providers and `registry_api.go` to use standard errors (`errors.Is`, `errors.As`) and check categories/wrapping.
    - **Done‑when:**
        1. Tests accurately verify centralized error handling.
        2. All related tests pass.
    - **Depends‑on:** [T021, T023]
- [ ] **T029 · Test · P2**: update/add tests for centralized parameter application
    - **Context:** SHRINK_PLAN.md § 3.3 (R3, R4), Implementation Step 6
    - **Action:**
        1. Review/Update parameter handling tests in providers to ensure correct behavior with `llm.ApplyParameters`.
    - **Done‑when:**
        1. Tests accurately verify centralized parameter application.
        2. All related tests pass.
    - **Depends‑on:** [T025]
