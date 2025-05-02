# Todo

## CR-01: Remove API Key Prefix Logging
- [x] **T001 · Bugfix · P0: remove api key prefix logging**
    - **Context:** cr-01 Remove API Key Prefix Logging
    - **Action:**
        1. Delete logging statement(s) emitting `effectiveApiKey` substrings in `internal/thinktank/registry_api.go` (approx. lines 154-169).
        2. Replace with logging only the *source* of the key (e.g., env var name) if diagnostic info is required.
    - **Done‑when:**
        1. Code removing key logging is merged.
        2. Manual log inspection (including debug level) confirms zero key fragments are logged.
    - **Verification:**
        1. Run the application locally, trigger API key usage, inspect logs (stdout/stderr/files) to confirm no API key fragments appear.
    - **Depends‑on:** none

## CR-04: Secret Detection Tests
- [x] **T002 · Test · P0: audit secret detection test coverage gaps**
    - **Context:** cr-04 Audit & Enhance Secret Detection Test Coverage (Steps 1-2)
    - **Action:**
        1. Review `internal/providers/*/provider_secrets_test.go` for Gemini, OpenAI, OpenRouter.
        2. Document specific scenarios (client creation, API calls, error handling) lacking secret detection tests using `logutil.WithSecretDetection`.
    - **Done‑when:**
        1. Audit document/comment outlining coverage gaps for each provider is complete.
    - **Depends‑on:** none
- [x] **T003 · Test · P0: implement secret detection tests for coverage gaps**
    - **Context:** cr-04 Audit & Enhance Secret Detection Test Coverage (Steps 3-4)
    - **Action:**
        1. Write new test cases for identified gaps (from T002) using `logutil.WithSecretDetection`.
        2. Ensure assertions check `HasDetectedSecrets() == false` for relevant scenarios.
    - **Done‑when:**
        1. New test cases covering identified gaps are implemented and pass locally.
    - **Depends‑on:** [T002]
- [x] **T004 · Chore · P0: ensure secret detection tests are mandatory in ci**
    - **Context:** cr-04 Audit & Enhance Secret Detection Test Coverage (Step 5)
    - **Action:**
        1. Verify CI config includes `internal/providers/...` tests.
        2. Ensure these tests are mandatory and block merges on failure.
    - **Done‑when:**
        1. CI configuration verified/updated to enforce secret detection tests.
    - **Verification:**
        1. Intentionally break a secret detection test, push branch, confirm CI fails and blocks merge.
    - **Depends‑on:** [T003]

## CR-03: E2E Testing
- [x] **T005 · Test · P0: fix failing e2e tests**
    - **Context:** cr-03 Fix & Enforce E2E Test Suite in CI (Steps 1-2)
    - **Action:**
        1. Execute `./internal/e2e/run_e2e_tests.sh` locally.
        2. Debug and fix root cause(s) of any failures.
    - **Done‑when:**
        1. `./internal/e2e/run_e2e_tests.sh` passes reliably locally.
    - **Depends‑on:** none
- [x] **T006 · Chore · P0: enforce e2e test success in ci**
    - **Context:** cr-03 Fix & Enforce E2E Test Suite in CI (Steps 3-4)
    - **Action:**
        1. Add job/step to CI pipeline to execute `./internal/e2e/run_e2e_tests.sh`.
        2. Configure CI platform (e.g., branch protection rules) to require this job to pass for merges to main.
    - **Done‑when:**
        1. CI configuration updated; merge to main is blocked if the E2E job fails.
    - **Verification:**
        1. Push a branch where E2E tests fail; confirm merge is blocked.
    - **Depends‑on:** [T005]

## CR-02: Default Model Consistency
- [x] **T007 · Chore · P1: align default model name in `readme.md` and `config.go`**
    - **Context:** cr-02 Ensure Consistent Default Model (+ CI Check) (Step 1)
    - **Action:**
        1. Verify default model name (`gemini-2.5-pro-preview-03-25`) is identical in `README.md` and `internal/config/config.go`.
        2. Update file(s) if mismatch exists.
    - **Done‑when:**
        1. The default model name is identical in both files.
    - **Depends‑on:** none
- [ ] **T008 · Feature · P1: create ci script (`check-defaults.sh`) to compare default models**
    - **Context:** cr-02 Ensure Consistent Default Model (+ CI Check) (Step 2)
    - **Action:**
        1. Create `scripts/ci/check-defaults.sh` to extract default model from `README.md` and `internal/config/config.go`.
        2. Implement comparison logic; exit non-zero on mismatch.
    - **Done‑when:**
        1. Script exists and correctly fails if default models mismatch.
    - **Depends‑on:** [T007]
- [ ] **T009 · Chore · P1: integrate `check-defaults.sh` into mandatory ci checks**
    - **Context:** cr-02 Ensure Consistent Default Model (+ CI Check) (Step 3)
    - **Action:**
        1. Add step to CI workflow to execute `scripts/ci/check-defaults.sh`.
        2. Ensure this step fails the build if the script exits non-zero.
    - **Done‑when:**
        1. CI configuration updated; build fails if default models mismatch.
    - **Verification:**
        1. Temporarily make defaults inconsistent, push branch, verify CI fails on the check.
    - **Depends‑on:** [T008]

## CR-09: Code Quality (`nolint:unused`)
- [ ] **T010 · Refactor · P2: remove or justify `nolint:unused` directives**
    - **Context:** cr-09 Remove/Justify all `nolint:unused`
    - **Action:**
        1. Search project for `//nolint:unused`.
        2. For each: remove unused code/directive OR add specific inline justification comment (`// Reason: ...`).
        3. Ensure `nolintlint` check passes in CI.
    - **Done‑when:**
        1. All `//nolint:unused` directives are removed or have specific justifications.
        2. CI linter check passes without relevant errors.
    - **Depends‑on:** none

## CR-05: Refactor Testability Helper (Constructor Injection)
- [ ] **T011 · Refactor · P1: modify `newregistryapiservice` signature for constructor injection**
    - **Context:** cr-05 Refactor Testability Helper via Constructor Injection (Step 1)
    - **Action:**
        1. Change `NewRegistryAPIService` signature in `internal/thinktank/registry_api.go` to accept `*registry.Registry`.
    - **Done‑when:**
        1. Function signature updated; code compiles (ignoring call site errors).
    - **Depends‑on:** none
- [ ] **T012 · Refactor · P1: update call sites for `newregistryapiservice` constructor injection**
    - **Context:** cr-05 Refactor Testability Helper via Constructor Injection (Steps 2, 4)
    - **Action:**
        1. Update production call site (e.g., `main.go`) to pass `*registry.Registry`.
        2. Update tests to instantiate `registryAPIService` directly via modified constructor, injecting mocks.
    - **Done‑when:**
        1. All production and test call sites updated; code compiles and tests pass.
    - **Depends‑on:** [T011]
- [ ] **T013 · Refactor · P1: remove `registryapiservicefortesting` struct and `setregistry` method**
    - **Context:** cr-05 Refactor Testability Helper via Constructor Injection (Step 3)
    - **Action:**
        1. Delete `RegistryAPIServiceForTesting` struct.
        2. Delete `SetRegistry` method.
    - **Done‑when:**
        1. Struct and method are removed from codebase; code compiles and tests pass.
    - **Verification:**
        1. Search codebase for deleted symbols; confirm zero results.
    - **Depends‑on:** [T012]

## CR-06: Relocate Provider Registry (Modularity)
- [ ] **T014 · Refactor · P2: create `internal/registry` package**
    - **Context:** cr-06 Relocate Provider Registry/Types for Modularity (Step 1)
    - **Action:**
        1. Create the new package `internal/registry`.
    - **Done‑when:**
        1. Directory `internal/registry` exists.
    - **Depends‑on:** none
- [ ] **T015 · Refactor · P2: move provider registry types/implementation to `internal/registry`**
    - **Context:** cr-06 Relocate Provider Registry/Types for Modularity (Steps 2-3)
    - **Action:**
        1. Move contents of `internal/thinktank/api_provider_types.go` to the new `internal/registry` package.
        2. Update package declaration in moved file(s).
    - **Done‑when:**
        1. Provider registry code resides in `internal/registry` with correct package declaration.
    - **Depends‑on:** [T014]
- [ ] **T016 · Refactor · P2: update imports and verify no circular dependencies for registry move**
    - **Context:** cr-06 Relocate Provider Registry/Types for Modularity (Steps 4-5)
    - **Action:**
        1. Update all import paths referencing the moved items to point to `internal/registry`.
        2. Check for and resolve any circular dependencies introduced.
    - **Done‑when:**
        1. All imports updated; project builds and tests pass; no circular dependencies reported.
    - **Verification:**
        1. Run build and static analysis checks for cycles.
    - **Depends‑on:** [T015]

## CR-07: Documentation Consolidation
- [ ] **T017 · Chore · P2: consolidate actionable tasks into `todo.md`**
    - **Context:** cr-07 Consolidate Documentation Sprawl (Steps 1-2)
    - **Action:**
        1. Review `PLAN.md`, `T009-plan.md`, `TODO-*.md` files.
        2. Merge all actionable, non-duplicate tasks into this `TODO.md` file.
    - **Done‑when:**
        1. `TODO.md` contains all current/near-term actionable tasks from source files.
    - **Depends‑on:** none
- [ ] **T018 · Chore · P2: delete redundant plan and task files**
    - **Context:** cr-07 Consolidate Documentation Sprawl (Step 3)
    - **Action:**
        1. Delete `T009-plan.md` and all `TODO-*.md` files.
    - **Done‑when:**
        1. Redundant files are removed from the repository.
    - **Depends‑on:** [T017]
- [ ] **T019 · Chore · P2: refactor `plan.md` into high-level strategy document**
    - **Context:** cr-07 Consolidate Documentation Sprawl (Step 4)
    - **Action:**
        1. Edit `PLAN.md` (or `REMEDIATION_PLAN.md`) to remove task lists and markdown code blocks.
        2. Refocus content as a concise strategy/design overview.
    - **Done‑when:**
        1. `PLAN.md`/`REMEDIATION_PLAN.md` is refocused as a strategy doc.
    - **Depends‑on:** [T017]
- [ ] **T020 · Chore · P2: clarify `backlog.md` purpose and remove duplicates**
    - **Context:** cr-07 Consolidate Documentation Sprawl (Step 5)
    - **Action:**
        1. Remove tasks from `BACKLOG.md` that are now in `TODO.md`.
        2. Add a clarifying header note to `BACKLOG.md` (e.g., "Future ideas/non-sprint items").
    - **Done‑when:**
        1. `BACKLOG.md` has no duplicates from `TODO.md` and purpose is clarified.
    - **Depends‑on:** [T017]

## CR-08: Test Coverage
- [ ] **T021 · Test · P1: analyze test coverage for critical packages post-refactor**
    - **Context:** cr-08 Audit & Restore Test Coverage Post-Deletion (Steps 1-2)
    - **Action:**
        1. Run `go test ./... -coverprofile=coverage.out` focusing on `internal/thinktank`, `internal/providers`, `internal/registry`, `internal/llm`.
        2. Analyze `coverage.out` to identify low-coverage areas (interfaces, error paths).
    - **Done‑when:**
        1. Coverage analysis complete; list of specific gaps documented.
    - **Depends‑on:** [T013, T016]
- [ ] **T022 · Test · P1: implement tests to address identified coverage gaps**
    - **Context:** cr-08 Audit & Restore Test Coverage Post-Deletion (Step 3)
    - **Action:**
        1. Write new/restored unit/integration tests for gaps identified in T021.
        2. Focus on interface contracts and boundary conditions.
    - **Done‑when:**
        1. New/restored tests implemented and passing locally; coverage report shows improvement.
    - **Depends‑on:** [T021]
- [ ] **T023 · Chore · P1: configure ci to enforce minimum test coverage threshold**
    - **Context:** cr-08 Audit & Restore Test Coverage Post-Deletion (Step 4)
    - **Action:**
        1. Determine and configure a minimum coverage threshold in CI (e.g., using `go tool cover` or other tools).
        2. Ensure CI build fails if coverage drops below this threshold.
    - **Done‑when:**
        1. CI configuration updated; build fails if coverage is below threshold.
    - **Verification:**
        1. Check CI configuration file; observe CI run output enforcing coverage.
    - **Depends‑on:** [T022]

---

### Clarifications & Assumptions
- [ ] **Issue:** No specific minimum test coverage threshold defined in plan (cr-08).
    - **Context:** cr-08, Step 4 requires enforcing a minimum threshold.
    - **Blocking?:** no (Assume a reasonable default like 80% for T023, adjust later if needed).
- [ ] **Issue:** Confirm target package name for provider registry relocation (cr-06).
    - **Context:** cr-06 suggests `internal/registry` or `internal/providers/registry`.
    - **Blocking?:** no (Proceeding with `internal/registry` as per T014, easily changed if incorrect).
