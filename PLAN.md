```markdown
# SHRINK_PLAN.md

## Code Size Optimization Plan for `thinktank`

---

This document outlines a detailed plan for significantly reducing the codebase size of the `thinktank` CLI tool while preserving all essential functionality, maintaining test coverage, and improving overall code readability and maintainability.

## 1. Size Analysis

*(Initial analysis based on provided file context. Precise numbers require tools like `cloc` and `wc`.)*

**Estimated Current State:**

*   **Total Go Files:** ~390 (including tests, per GPT-4.1) / ~50-60 (excluding tests, per Gemini-2.5-Pro)
*   **Total Lines of Code (LOC):** ~52,000 (including tests, comments, blank lines, per GPT-4.1)
*   **Core Source LOC (Estimated):** ~5,000 - 8,000 (excluding tests, per Gemini-2.5-Pro) / ~26,000 (excluding tests, per GPT-4.1) -> *Let's assume a range, maybe 10k-20k core LOC is more realistic given the file count.*
*   **Test Code LOC:** ~26,000 (per GPT-4.1) - Clearly a significant portion.

**Largest Files/Modules (Estimated by LOC/Complexity):**

*   `internal/thinktank/orchestrator/orchestrator.go` (~400-600 LOC)
*   `cmd/thinktank/cli.go` (~200-550 LOC)
*   `internal/thinktank/registry_api.go` (~300 LOC)
*   `internal/thinktank/modelproc/processor.go` (~250-400 LOC)
*   `internal/registry/manager.go` (~400 LOC)
*   `internal/providers/*` (Gemini, OpenAI, OpenRouter clients/providers) (~200-300 LOC each)
*   `internal/llm/errors.go` (~300 LOC)
*   **Test Files:** Numerous `*_test.go` files, especially in `cmd/thinktank/`, `internal/integration/`, and `internal/providers/`, often containing >100-700 LOC each due to duplicated mocks and setup.
*   `internal/thinktank/api.go` (~250 LOC) - Legacy API service.

**LOC by Type (Estimated):**

*   **Test/Mock Code:** ~50% or more of total LOC.
*   **Core Business Logic (Orchestrator, ModelProc):** ~15-20%
*   **Provider Integrations:** ~10-15%
*   **CLI Handling:** ~5%
*   **Error Handling:** ~5%
*   **Registry/Config:** ~5%
*   **Utilities/File Handling:** ~5%

**Redundancy/Complexity Hotspots:**

*   **Mock Implementations:** Extensive duplication of mock loggers, API services, file writers across test files.
*   **Provider Logic:** Similar parameter handling, HTTP request logic, and error mapping across Gemini, OpenAI, OpenRouter implementations.
*   **Error Handling:** Repetitive error checking (e.g., empty response, safety blocks) and formatting logic.
*   **Adapters:** Multiple layers of interface adapters (`APIServiceAdapter`, `ContextGathererAdapter`, etc.) potentially adding unnecessary indirection.
*   **Legacy Code:** Obsolete token management code, disabled test files, deprecated `runutil` package, legacy `apiService`.
*   **API Key Resolution:** Logic duplicated across CLI and provider implementations.

## 2. Reduction Opportunities

Catalog of specific opportunities identified across the codebase:

1.  **R1: Redundant Mock Implementations:** Widespread duplication of mock loggers (`mockLogger`, `MockLogger`, `mockAPILogger`, etc.) and potentially other mocks (API service, file writer) across dozens of `_test.go` files.
2.  **R2: Unused/Dead/Deprecated Code:**
    *   Obsolete token management files (`internal/thinktank/token.go`, `registry_token.go`, `token_stubs.go`, `modelproc/token_stubs.go`).
    *   Disabled test files (`*.go.disabled`).
    *   Deprecated `internal/runutil/` package.
    *   Legacy `apiService` in `internal/thinktank/api.go` and associated tests/functions (`DetectProviderFromModel`).
    *   Unused functions marked `//nolint:unused` in tests.
    *   Commented-out code blocks.
3.  **R3: Duplicated Provider Logic:**
    *   Parameter application logic (mapping `map[string]interface{}` to client setters) in provider adapters.
    *   Common HTTP request/response handling patterns.
    *   Similar error mapping/categorization logic.
4.  **R4: Duplicated Error Handling Logic:** Manual string checks (`IsEmptyResponseError`, `IsSafetyBlockedError`) and inconsistent error formatting/categorization across providers and API services.
5.  **R5: Duplicated API Key Resolution:** Logic for finding the correct API key (CLI param, env var) repeated in CLI parsing and multiple provider implementations.
6.  **R6: Potentially Overengineered Abstractions:** Adapter layers (`internal/thinktank/adapters.go`) might be unnecessary if concrete types already satisfy required interfaces.
7.  **R7: Duplicated Helper Functions:** Simple helpers like `min`/`max` duplicated instead of using Go 1.21+ standard library.
8.  **R8: Verbose Implementations & Comments:** Overly verbose logging, comments explaining *what* instead of *why*, excessive error wrapping adding little value.

## 3. Proposed Solutions

### 3.1 (R1) Consolidate Mock Implementations

*   **Opportunity:** Massive duplication of mocks, especially loggers.
*   **Approach:** Standardize on a single set of reusable mocks in `internal/testutil`. Create a robust `testutil.MockLogger` implementing required interfaces (`logutil.LoggerInterface`, `auditlog.AuditLogger`). Do the same for other commonly mocked interfaces (e.g., `APIService`, `FileWriter`, `ExternalAPICaller`). Refactor all tests to use these shared mocks.
*   **Estimate Reduction:** 2,000 - 4,000+ LOC (Potentially 5-10% of total LOC, 10-20% of test LOC).
*   **Implementation:**
    1.  Enhance/Create `internal/testutil/mocklogger.go`, `mock_apiservice.go`, etc. with necessary methods and assertion helpers (e.g., `GetMessages`, `AssertCalledWith`).
    2.  Iterate through all `_test.go` files.
    3.  Remove local mock definitions.
    4.  Replace instantiation and usage with `testutil.NewMockLogger()`, `testutil.NewMockAPIService()`, etc.
    5.  Delete redundant mock files (e.g., `internal/fileutil/mock_logger.go`).
    6.  Run `go test ./...` frequently.

### 3.2 (R2) Eliminate Dead/Deprecated/Legacy Code

*   **Opportunity:** Significant amount of unused, disabled, or obsolete code.
*   **Approach:** Delete all identified dead code, disabled files, deprecated packages, and the legacy API service if the registry approach is stable and covers all needs (or refactor the fallback).
*   **Estimate Reduction:** 1,500 - 3,000+ LOC (Significant reduction, ~3-6% of total LOC).
*   **Implementation:**
    1.  `git rm` the following files/directories:
        *   `internal/thinktank/token*.go`
        *   `internal/thinktank/modelproc/token_stubs.go`
        *   `internal/runutil/`
        *   All `*.go.disabled` files.
        *   Functions marked `//nolint:unused`.
        *   Commented-out code blocks.
    2.  Remove the legacy `apiService` (`internal/thinktank/api.go`, `api_test.go`, `DetectProviderFromModel`) after verifying the registry service (`registry_api.go`) is sufficient or the fallback logic is self-contained within it.
    3.  Search for and remove any remaining imports or references.
    4.  Run `go mod tidy`.
    5.  Run `go test ./...` and fix any resulting errors.

### 3.3 (R3, R4) Refactor & Centralize Provider Logic / Error Handling

*   **Opportunity:** Duplication in provider clients (HTTP, params, errors) and error checking/formatting.
*   **Approach:**
    1.  **Base Client/Utils:** Create shared utilities in `internal/providers/base` or `internal/llm` for common HTTP requests, standard parameter application (using interfaces/reflection carefully), etc.
    2.  **Centralized Error Handling:** Move error checking logic (`IsEmptyResponseError`, `IsSafetyBlockedError`) to `internal/llm/errors.go` using `errors.Is` and standard error variables (e.g., `llm.ErrEmptyResponse`). Centralize `llm.LLMError` creation and formatting using `llm.CreateStandardErrorWithMessage` or similar, ensuring consistent `ErrorCategory` usage.
*   **Estimate Reduction:** 800 - 1,500 LOC (~2-3% of total LOC, significant % of provider/error code).
*   **Implementation:**
    1.  Define standard error variables (`ErrEmptyResponse`, `ErrContentFiltered`) in `internal/llm/errors.go`.
    2.  Refactor `IsEmptyResponseError`/`IsSafetyBlockedError` in `registry_api.go` to use `errors.Is` and `llm.IsCategory` with the new variables/categories.
    3.  Create `internal/llm/parameters.go` or similar with a function like `ApplyParameters(client interface{}, params map[string]interface{})`.
    4.  Refactor `internal/providers/*/provider.go` adapters to use `llm.ApplyParameters`.
    5.  Refactor provider clients (`internal/providers/*/client.go`) to use shared HTTP utilities (if applicable) and call centralized error formatting functions from `internal/llm/errors.go`.
    6.  Update unit and integration tests for providers and error handling.

### 3.4 (R5) Centralize API Key Resolution

*   **Opportunity:** API key precedence logic duplicated.
*   **Approach:** Consolidate API key resolution logic within `internal/thinktank/registry_api.go::InitLLMClient`. This function should be the single source of truth for checking CLI params, environment variables (using `ModelsConfig.APIKeySources`), and global keys before calling `Provider.CreateClient`.
*   **Estimate Reduction:** ~50-100 LOC.
*   **Implementation:**
    1.  Ensure `registryAPIService.InitLLMClient` implements the full key resolution logic.
    2.  Remove redundant lookups from individual provider `CreateClient` methods.
    3.  Remove redundant checks from `cmd/thinktank/cli.go` (rely on `InitLLMClient` during execution).
    4.  Update tests to verify correct key precedence.

### 3.5 (R6) Review & Flatten Adapters

*   **Opportunity:** Adapter layer might be unnecessary indirection.
*   **Approach:** Analyze if concrete types (`registryAPIService`, `contextGatherer`, `fileWriter`) directly implement the interfaces required by the `Orchestrator` (`interfaces.APIService`, etc.). If so, remove the adapter layer (`internal/thinktank/adapters.go`) and pass implementations directly.
*   **Estimate Reduction:** ~100-200 LOC.
*   **Implementation:**
    1.  Check interface satisfaction for `registryAPIService`, `contextGatherer`, `fileWriter`.
    2.  If satisfied, remove `internal/thinktank/adapters.go`.
    3.  Update `NewOrchestrator` and call sites (`app.go`) to inject concrete types or interfaces directly.
    4.  Run tests.

### 3.6 (R7) Use Standard Library Helpers

*   **Opportunity:** Duplicated `min`/`max`.
*   **Approach:** Remove custom `min`/`max` implementations and use the standard library versions (requires Go 1.21+).
*   **Estimate Reduction:** ~30 LOC.
*   **Implementation:**
    1.  Delete custom functions.
    2.  Replace calls with `min(...)` / `max(...)`.
    3.  Ensure `go.mod` specifies `go 1.21` or higher.

### 3.7 (R8) Prune Verbose Logging & Comments

*   **Opportunity:** Excessive logging and non-essential comments.
*   **Approach:** Review logs, keeping only those essential for debugging or user feedback. Remove comments that explain *what* obvious code does; keep only *why* comments for non-obvious decisions.
*   **Estimate Reduction:** ~500-1,200 LOC (~1-2% of total LOC).
*   **Implementation:** Audit logs in orchestrator, modelproc, providers. Remove trivial success logs. Review comments across the codebase.

## 4. Prioritization

| Opportunity                                 | Size Impact | Complexity | Risk      | Priority |
| :------------------------------------------ | :---------- | :--------- | :-------- | :------- |
| **Eliminate Dead/Deprecated/Legacy Code**   | Very High   | Low        | Very Low  | 1        |
| **Consolidate Mock Implementations**        | Very High   | Medium     | Low       | 2        |
| **Refactor Provider Logic / Error Handling**| High        | Medium-High| Medium    | 3        |
| **Prune Verbose Logging & Comments**        | Medium      | Low        | Very Low  | 4        |
| **Review & Flatten Adapters**               | Low-Medium  | Medium     | Low-Medium| 5        |
| **Centralize API Key Resolution**           | Low         | Low-Medium | Low       | 6        |
| **Use Standard Library Helpers (`min`/`max`)**| Very Low    | Very Low   | Very Low  | 7        |

## 5. Architectural Improvements

1.  **Standardized Mocks:** Centralize all mocks in `internal/testutil` for consistency and reuse.
2.  **Shared Provider Utilities:** Implement a base layer (`internal/providers/base` or `internal/llm`) for common provider logic (parameter application, potentially HTTP client wrappers).
3.  **Centralized Error Handling:** Consolidate error checking, categorization (`llm.ErrorCategory`), and formatting (`llm.LLMError`) in `internal/llm`.
4.  **Flattened Abstractions:** Remove unnecessary adapter layers where direct interface implementation suffices.
5.  **Boundary Pattern:** Reinforce clear boundaries for external interactions (HTTP, Filesystem) using interfaces and inject mocks in tests (as seen partially in `internal/integration`).

## 6. Testing Strategy

*   **Baseline:** Ensure `go test ./...`, `scripts/check-coverage.sh`, and `./internal/e2e/run_e2e_tests.sh` pass before starting.
*   **Coverage:** Maintain or increase test coverage (Target: 80%+). Add specific tests for refactored areas (provider logic, error handling, parameter application).
*   **Critical Test Cases:**
    *   All existing unit, integration, and E2E tests.
    *   Core orchestrator flows (single model, multi-model, synthesis).
    *   Provider-specific integration tests (using boundary mocks) covering parameter handling, content generation, and error mapping for Gemini, OpenAI, OpenRouter.
    *   Error handling scenarios (API errors, network errors, safety blocks, empty responses).
    *   CLI command execution, flag parsing, and validation.
    *   File input/output operations.
*   **Specific Verification Areas:**
    *   **Provider Refactoring:** Verify correct request formation and response/error parsing after introducing shared logic.
    *   **Error Handling:** Test that diverse errors are correctly categorized and formatted by the central `llm` logic.
    *   **Mock Consolidation:** Ensure tests using the new shared mocks still provide sufficient verification.
    *   **Legacy Code Removal:** Confirm no regressions occurred after removing `apiService` or token code.
*   **Risks & Mitigation:**
    *   *Risk:* Regressions due to large refactoring. **Mitigation:** Incremental changes, frequent testing, thorough code reviews.
    *   *Risk:* Centralized logic misses provider-specific nuances. **Mitigation:** Comprehensive provider-specific integration tests covering edge cases.
    *   *Risk:* Shared mocks become too generic. **Mitigation:** Parameterize mocks or add specific assertion helpers as needed.

## 7. Expected Results

*   **Estimated Size Reduction:**
    *   **Overall:** **20-35%** LOC reduction (potentially 10,000 - 18,000 LOC removed).
    *   **Test Code:** Significant reduction (up to 30-40%) due to mock consolidation and dead test removal.
    *   **Core Code:** Moderate reduction (15-25%) by removing legacy code, centralizing logic.
*   **Performance Impact:** Likely neutral. Minor potential improvements from reduced indirection and logging overhead.
*   **Maintainability/Readability:** Dramatically improved. Less code, reduced duplication, clearer architecture, easier testing, faster onboarding for new developers.

---
*End of SHRINK_PLAN.md*
```
