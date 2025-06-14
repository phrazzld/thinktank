# Todo

## URGENT: CI Resolution Tasks (PR #88)
- [x] **CI001 · Chore · P0: research and validate alternative property-based testing library**
    - **Context:** CI failure - pgregory.net/rapid uses forbidden MPL-2.0 license
    - **Action:**
        1. Research `github.com/leanovate/gopter` (MIT license) as replacement for `pgregory.net/rapid`
        2. Verify gopter provides equivalent property-based testing functionality
        3. Review API compatibility and migration complexity
        4. Confirm MIT license is in approved allowlist
    - **Done‑when:**
        1. Alternative library identified with compatible API and approved license
        2. Migration approach documented with effort estimate
    - **Depends‑on:** none

- [x] **CI002 · Chore · P0: replace problematic dependency with approved alternative**
    - **Context:** CI failure - license policy violation blocking merge
    - **Action:**
        1. Remove `pgregory.net/rapid v1.2.0` from go.mod
        2. Add `github.com/leanovate/gopter` (or validated alternative) to go.mod
        3. Run `go mod tidy` to update dependencies
        4. Verify no MPL-2.0 licenses remain in dependency tree
    - **Done‑when:**
        1. `go mod` contains only approved-license dependencies
        2. License compliance CI check passes
    - **Depends‑on:** [CI001]

- [x] **CI003 · Chore · P0: migrate property-based tests to new library**
    - **Context:** API migration required after dependency replacement
    - **Action:**
        1. Update `internal/testutil/property_testing.go` to use new library API
        2. Migrate all property-based test generators in affected files
        3. Update import statements across codebase
        4. Verify all property-based tests continue to pass
    - **Done‑when:**
        1. All property-based tests use new library and pass
        2. No references to old library remain in codebase
    - **Depends‑on:** [CI002]

- [x] **CI004 · Chore · P0: fix errcheck violations in integration test files**
    - **Context:** CI failure - golangci-lint errcheck violations
    - **Action:**
        1. Fix `internal/gemini/integration_test.go:32:29` - add error handling to JSON encoding
        2. Fix `internal/openai/integration_test.go:31:29` - add error handling to JSON encoding
        3. Use pattern: `if err := json.NewEncoder(w).Encode(response); err != nil { t.Fatalf("failed to encode response: %v", err) }`
    - **Done‑when:**
        1. No errcheck violations remain in integration test files
        2. golangci-lint errcheck passes for entire codebase
    - **Depends‑on:** none

- [x] **CI005 · Chore · P0: remove unused test utility functions**
    - **Context:** CI failure - golangci-lint unused function violations
    - **Action:**
        1. Remove `setupMockServer` function from `internal/gemini/integration_test.go:17`
        2. Remove `createJSONHandler` function from `internal/gemini/integration_test.go:27`
        3. Remove `createErrorHandler` function from `internal/gemini/integration_test.go:38`
        4. Verify functions are truly unused before removal
    - **Done‑when:**
        1. No unused function violations remain in codebase
        2. golangci-lint unused passes for entire codebase
    - **Depends‑on:** none

- [x] **CI006 · Feature · P0: implement expected error test logger**
    - **Context:** CI failure - test fails due to expected error logs being treated as failures
    - **Action:**
        1. Create `ExpectedErrorTestLogger` in `internal/logutil/test_logger.go`
        2. Add `ExpectError(pattern string)` method to pre-declare expected error messages
        3. Modify error detection to ignore declared expected patterns
        4. Maintain fail-fast behavior for unexpected errors
    - **Done‑when:**
        1. Test logger supports expected error pattern declarations
        2. Expected errors don't cause test failures
        3. Unexpected errors still cause test failures
    - **Depends‑on:** none

- [ ] **CI007 · Chore · P0: update failing synthesis test to use expected error patterns**
    - **Context:** CI failure - TestSynthesisWithModelFailuresFlow fails on expected error logs
    - **Action:**
        1. Modify `TestSynthesisWithModelFailuresFlow` in `internal/integration/synthesis_with_failures_test.go`
        2. Use `ExpectedErrorTestLogger` to declare expected error patterns
        3. Pre-declare expected error messages: "Generation failed for model model2", "Processing model model2 failed", etc.
        4. Verify test still validates correct partial failure behavior
    - **Done‑when:**
        1. `TestSynthesisWithModelFailuresFlow` passes without suppressing error detection
        2. Test continues to verify partial failure handling correctly
    - **Depends‑on:** [CI006]

- [ ] **CI008 · Chore · P0: verify all CI checks pass**
    - **Context:** Final validation before merge
    - **Action:**
        1. Run full local CI simulation: `go test ./...`, `golangci-lint run ./...`, license check
        2. Push changes and verify all CI jobs pass in GitHub Actions
        3. Confirm dependency license compliance, lint checks, and test suite
    - **Done‑when:**
        1. All CI checks pass: License Compliance ✅, Lint and Format ✅, Test ✅
        2. PR ready for merge with no quality gate violations
    - **Depends‑on:** [CI003, CI004, CI005, CI007]

- [ ] **CI009 · Feature · P2: implement proactive license checking**
    - **Context:** Prevention - avoid future license policy violations
    - **Action:**
        1. Add `go-licenses` check to pre-commit hooks in `.pre-commit-config.yaml`
        2. Create `scripts/check-licenses.sh` for local license validation
        3. Document license policy and checking process in `DEVELOPMENT.md`
    - **Done‑when:**
        1. Pre-commit hooks prevent commits with forbidden licenses
        2. Developers can check licenses locally before committing
    - **Depends‑on:** [CI008]

- [ ] **CI010 · Feature · P2: enhance local development workflow**
    - **Context:** Prevention - catch lint issues before CI
    - **Action:**
        1. Add golangci-lint to pre-commit hooks
        2. Create `make ci-check` target for full local CI simulation
        3. Update `DEVELOPMENT.md` with local validation workflow
    - **Done‑when:**
        1. Pre-commit hooks catch lint violations locally
        2. Developers can simulate full CI pipeline locally
    - **Depends‑on:** [CI008]

## Phase 0: Decisions & Setup
- [x] **T001 · Chore · P1: decide on and document CI test API key management strategy**
    - **Context:** Open Questions & Dependencies > 1. Test Environment Management
    - **Action:**
        1. Define the strategy for securely managing test API keys in the CI environment (e.g., using repository/organization secrets).
        2. Document the setup process for developers and the CI configuration.
        3. Formally confirm the plan's direction of using in-memory mocks for external HTTP APIs, not real provider APIs, in the standard test suite.
    - **Done‑when:**
        1. The key management and external service mocking strategy is documented in `TESTING.md` or a similar guide.
    - **Depends‑on:** none
- [x] **T002 · Chore · P1: select and document the property-based testing library**
    - **Context:** Open Questions & Dependencies > 2. Property-Based Testing Framework
    - **Action:**
        1. Evaluate Go property-based testing libraries (e.g., rapid, gopter) based on features and ease of integration.
        2. Choose a library and document the decision with a brief rationale.
    - **Done‑when:**
        1. A library is selected and added to `go.mod`.
        2. The choice is documented in `TESTING.md`.
    - **Depends‑on:** none
- [x] **T003 · Chore · P2: decide and configure CI coverage measurement strategy**
    - **Context:** Open Questions & Dependencies > 4. Coverage Measurement
    - **Action:**
        1. Decide whether to exclude test utility packages (e.g., `internal/testutil`) from coverage threshold calculations.
        2. Configure the CI job to calculate and report coverage according to the decided strategy.
    - **Done‑when:**
        1. The coverage measurement strategy is documented.
        2. The CI configuration is updated to reflect the strategy.
    - **Depends‑on:** none

## Phase 1: Critical Foundations
- [x] **T004 · Feature · P0: implement `internal/testutil` package with basic helpers**
    - **Context:** 1.1 Test Infrastructure Foundation
    - **Action:**
        1. Create the `internal/testutil` package with `integration.go`.
        2. Add helpers for common test tasks like creating temporary directories/files and ensuring their cleanup via `t.Cleanup`.
    - **Done‑when:**
        1. `internal/testutil/integration.go` exists with file/directory helpers.
        2. Helpers are covered by their own unit tests.
    - **Depends‑on:** none
- [x] **T005 · Feature · P0: implement in-memory HTTP server utility for mocking external APIs**
    - **Context:** 1.1 Test Infrastructure Foundation > In-memory External System Implementations
    - **Action:**
        1. Create a helper function in `internal/testutil/providers.go` that sets up and tears down an `httptest.Server`.
        2. The helper should allow test functions to easily define handlers for specific API endpoints to simulate provider responses (success, errors, malformed JSON, etc.).
    - **Done‑when:**
        1. A test can easily create a mock HTTP server to act as a provider's external API endpoint.
    - **Depends‑on:** [T004]
- [x] **T006 · Feature · P0: implement test data factories for provider configs and API objects**
    - **Context:** 1.1 Test Infrastructure Foundation > Test data factories for complex structures
    - **Action:**
        1. Implement a Test Data Builder/Factory pattern within `internal/testutil`.
        2. Create factories for `ProviderConfig`, API requests (e.g., `ChatCompletionRequest`), and API responses.
        3. Include methods for creating both valid and invalid data variations (e.g., `InvalidAPIKey()`, `InvalidTemperature()`).
    - **Done‑when:**
        1. Factories for core data structures are implemented and available for use in tests.
    - **Depends‑on:** [T004]
- [x] **T007 · Feature · P1: implement `logutil.TestLogger` for structured test logging**
    - **Context:** Logging & Observability Approach > Test Logging Strategy
    - **Action:**
        1. Create a `TestLogger` in a `logutil` package that captures structured logs in memory.
        2. Implement methods to retrieve logs and assert that no error-level logs were captured during a test.
        3. Use `t.Cleanup` in the logger's setup function to automatically fail tests that logged an error.
    - **Done‑when:**
        1. `logutil.TestLogger` is implemented and available for use in tests.
    - **Depends‑on:** none
- [x] **T008 · Feature · P1: implement secure test configuration and API key handling helper**
    - **Context:** Security & Configuration Considerations > API Key Management in Tests
    - **Action:**
        1. Implement a `getTestAPIKey` helper function in `internal/testutil` that safely retrieves keys from environment variables.
        2. The helper must check that the key is a test key (e.g., has a `test-` prefix) and skip the test if not provided.
    - **Done‑when:**
        1. `getTestAPIKey` function is implemented and used in provider tests.
        2. Tests can securely access test-only API keys without hardcoding them.
    - **Depends‑on:** [T001, T004]
- [x] **T009 · Test · P0: add integration tests for `internal/gemini` entry points to 85%+ coverage**
    - **Context:** 1.2 Provider Entry Points (0% → 85%+)
    - **Action:**
        1. Add tests for `NewLLMClient`, `Close`, and `GetModelName` using the in-memory HTTP server.
        2. Add unit tests for pure functions `mapSafetyRatings` and `toProviderSafety`.
    - **Done‑when:**
        1. All target functions are tested, achieving at least 85% coverage for their respective files.
    - **Depends‑on:** [T005, T006]
- [x] **T010 · Test · P0: add integration tests for `internal/openai` entry points to 85%+ coverage**
    - **Context:** 1.2 Provider Entry Points (0% → 85%+)
    - **Action:**
        1. Add tests for `createChatCompletion`, `createChatCompletionWithParams`, and `Close` using the in-memory HTTP server.
    - **Done‑when:**
        1. All target functions are tested, achieving at least 85% coverage for their respective files.
    - **Depends‑on:** [T005, T006]
- [x] **T011 · Test · P0: add tests for `cmd/thinktank` flag parsing and input validation**
    - **Context:** 1.3 Core Application Entry Points (0% → 80%+)
    - **Action:**
        1. Add tests for `ParseFlags` and `ValidateInputs` covering valid cases, invalid cases, and edge cases.
        2. If necessary, refactor `main()` to move core logic into a testable `run()` function that returns an error.
    - **Done‑when:**
        1. `ParseFlags` and `ValidateInputs` have at least 80% test coverage.
    - **Depends‑on:** none
- [x] **T012 · Test · P0: add tests for `internal/thinktank` context and dry-run functions**
    - **Context:** 1.3 Core Application Entry Points (0% → 80%+)
    - **Action:**
        1. Add unit tests for `GatherContext` using a temporary file system created with helpers from `testutil`.
        2. Add unit tests for `DisplayDryRunInfo` to verify its output format by capturing stdout.
    - **Done‑when:**
        1. `GatherContext` and `DisplayDryRunInfo` have at least 80% test coverage.
    - **Depends‑on:** [T004]

## Phase 2: Core Business Logic
- [x] **T013 · Test · P1: add table-driven tests for `GenerateContent` parameter boundaries**
    - **Context:** 2.1 Provider Implementation Completion > Parameter boundary testing
    - **Action:**
        1. For each provider, add table-driven tests for the `GenerateContent` method.
        2. Test boundary conditions for parameters like `temperature` and `maxOutputTokens`, verifying that invalid values return errors.
    - **Done‑when:**
        1. Parameter validation for `GenerateContent` is comprehensively tested for all providers.
    - **Depends‑on:** [T009, T010]
- [x] **T014 · Test · P1: add integration tests for `GenerateContent` API error scenarios**
    - **Context:** 2.1 Provider Implementation Completion > Error scenario testing
    - **Action:**
        1. For each provider, add tests for `GenerateContent` that use the in-memory HTTP server to simulate various API failures (4xx/5xx status codes, malformed JSON, network timeouts).
    - **Done‑when:**
        1. API error handling for `GenerateContent` is tested for all providers, ensuring errors are correctly propagated.
    - **Depends‑on:** [T005, T009, T010]
- [x] **T015 · Feature · P2: implement property-based testing utilities and initial tests**
    - **Context:** 2.1 Provider Implementation Completion > Property-based testing for content processing
    - **Action:**
        1. Add the chosen PBT library to `go.mod` and create helper generators for project types in `internal/testutil/property_testing.go`.
        2. Write an initial property-based test for a function like `processProviderResponse` to verify invariants (e.g., token counts are non-negative).
    - **Done‑when:**
        1. Utilities for property-based testing are available in `internal/testutil`.
        2. An example property-based test is implemented and passes.
    - **Depends‑on:** [T002]
- [x] **T016 · Test · P1: add tests for `thinktank.Execute` error handling paths**
    - **Context:** 2.2 Core Logic Implementation > Execute() method error handling
    - **Action:**
        1. Add tests for the `thinktank.Execute` method that trigger various error conditions, such as provider errors or file I/O errors.
        2. Verify that the correct errors are returned and logged.
    - **Done‑when:**
        1. Key error handling paths in `thinktank.Execute` are covered by tests.
    - **Depends‑on:** [T012]
- [x] **T017 · Test · P1: add tests for `thinktank.Execute` file processing and dry run**
    - **Context:** 2.2 Core Logic Implementation > Context gathering, Output directory, Dry run
    - **Action:**
        1. Add tests for the `thinktank.Execute` happy path, using a temporary directory to provide inputs and verify that output files are created correctly.
        2. Add tests for `thinktank.Execute` with the dry-run flag enabled, verifying correct info is logged and no files are written.
    - **Done‑when:**
        1. Core file processing, output management, and dry-run logic in `thinktank.Execute` is tested.
    - **Depends‑on:** [T004, T012]

## Phase 3: Integration & Completeness
- [x] **T018 · Test · P2: add integration tests for CLI flag parsing edge cases and exit codes**
    - **Context:** 3.1 CLI Interface Completion
    - **Action:**
        1. Using `os/exec` to run the compiled test binary, create tests for the full CLI workflow.
        2. Test edge cases for flag combinations, invalid inputs, and verify the application exits with the correct status code.
    - **Done‑when:**
        1. CLI behavior for various flag combinations and error conditions is tested.
    - **Depends‑on:** [T011]
- [x] **T019 · Test · P2: add tests for `orchestrator` logic**
    - **Context:** 3.2 Orchestrator Logic
    - **Action:**
        1. Add unit tests for error classification, model management, and result processing functions in the `orchestrator` package.
    - **Done‑when:**
        1. Critical functions in the `orchestrator` package reach at least 90% test coverage.
    - **Depends‑on:** none
- [x] **T020 · Test · P1: implement comprehensive end-to-end workflow integration test**
    - **Context:** 3.3 Integration Test Suite
    - **Action:**
        1. Implement a test that executes the entire application workflow using real internal components and in-memory implementations for external systems.
        2. The test should set up a test environment, execute the app, and verify the final observable outcomes (output files, audit logs).
    - **Done‑when:**
        1. An end-to-end test validating the primary success path of the application exists and passes.
    - **Depends‑on:** [T017, T019]
- [x] **T021 · Test · P2: add integration test to verify correlation ID propagation**
    - **Context:** Logging & Observability Approach > Test correlation ID propagation
    - **Action:**
        1. Create an integration test that executes an operation spanning multiple components, injecting a `correlation_id` into the initial context.
        2. Use the `logutil.TestLogger` to capture all logs and assert that every log entry contains the correct correlation ID.
    - **Done‑when:**
        1. Correlation ID propagation is verified by an automated test.
    - **Depends‑on:** [T007, T020]
- [x] **T022 · Chore · P1: document established testing patterns and infrastructure usage**
    - **Context:** Risk Matrix & Mitigation > 3. Pattern Documentation
    - **Action:**
        1. Create or update a `TESTING.md` document in the repository.
        2. Document the core testing principles (no internal mocking) and provide clear examples for using the test infrastructure (`testutil`, data factories, etc.).
    - **Done‑when:**
        1. `TESTING.md` is created and populated with patterns established in Phase 1.
    - **Depends‑on:** [T004, T005, T006]
- [x] **T023 · Chore · P1: coordinate with Issue #46 to enforce CI coverage thresholds**
    - **Context:** External Dependencies > 1. Issue #46
    - **Action:**
        1. Work with the owner of Issue #46 to configure the CI pipeline to fail if overall test coverage drops below 90% or if any critical package is below 90%.
    - **Done‑when:**
        1. CI pipeline enforces the 90% coverage target on pull requests.
    - **Depends‑on:** [T003]

### Clarifications & Assumptions
- [ ] **Issue:** Finalize decision on using real vs. mock LLM API calls for specific, non-default integration tests.
    - **Context:** Open Questions & Dependencies > 3. Integration Test Scope
    - **Blocking?:** no
- [ ] **Issue:** Ensure compatibility with upcoming changes from Issue #62 (Testing Infrastructure Overhaul) and #65 (Gordian Simplification).
    - **Context:** External Dependencies > 2. Issue #62, 3. Issue #65
    - **Blocking?:** no
