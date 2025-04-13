# TODO

## 1. Refactor for Parallel Integration Test Execution

- [x] **Task Title:** Modify `architect.Execute` Signature for Dependency Injection
    - **Action:** Change the signature of `architect.Execute` in `internal/architect/app.go` to accept `APIService` as an explicit parameter instead of relying on the global `NewAPIService` variable. Update the function's internal logic to use the passed parameter.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 1 (Modify `architect.Execute`...)

- [x] **Task Title:** Update `main.go` to Inject `APIService` into `Execute`
    - **Action:** Modify the main application entry point (`cmd/architect/main.go` - *path assumed*) to instantiate the concrete `APIService` and pass it into the refactored `architect.Execute` function.
    - **Depends On:** Modify `architect.Execute` Signature for Dependency Injection
    - **AC Ref:** Plan Recommendation 1 (Modify `architect.Execute`...)

- [ ] **Task Title:** Remove Global `architect.NewAPIService` Variable
    - **Action:** Delete the global `var NewAPIService = func(...)` definition from `internal/architect/api.go`. Remove any remaining code that reads or writes to this variable.
    - **Depends On:** Update `main.go` to Inject `APIService` into `Execute`, Update Integration Tests to Inject Dependencies
    - **AC Ref:** Plan Recommendation 1 (Refactor problematic pattern)

- [x] **Task Title:** Refactor `TestEnv.Setup` to Eliminate Global I/O Redirection
    - **Action:** Remove the code in `internal/integration/test_helpers.go` (`TestEnv.Setup`) that redirects `os.Stdout` and `os.Stderr` using `os.Pipe`. The `StdoutBuffer` and `StderrBuffer` should still exist but will need to be passed explicitly where needed.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 1 (Remove global I/O redirection...)

- [x] **Task Title:** Update Integration Tests to Pass I/O Buffers Explicitly
    - **Action:** Modify integration tests (`internal/integration/integration_test.go`) that use `TestEnv`. Pass `env.StdoutBuffer` and `env.StderrBuffer` directly to functions or components (like `logutil.NewLogger`) that require `io.Writer` instead of relying on the removed global redirection.
    - **Depends On:** Refactor `TestEnv.Setup` to Eliminate Global I/O Redirection
    - **AC Ref:** Plan Recommendation 1 (Pass buffers directly...)

- [x] **Task Title:** Refactor `TestEnv` Stdin Simulation to Use Buffers
    - **Action:** Modify `NewTestEnv` in `internal/integration/test_helpers.go` to remove the creation and use of a temporary file (`mockStdin`) for simulating standard input. Modify `SimulateUserInput` to write to an internal `bytes.Buffer` or similar. Provide a way for tests to pass this buffer as an `io.Reader` where stdin is needed.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 1 (Remove need for mock stdin files...)

- [x] **Task Title:** Update Integration Tests Using Stdin Simulation
    - **Action:** Modify integration tests (`internal/integration/integration_test.go`) that use `env.SimulateUserInput`. Update them to pass the `io.Reader` provided by the refactored `TestEnv` to the relevant functions (e.g., potentially `TokenManager.PromptForConfirmation`).
    - **Depends On:** Refactor `TestEnv` Stdin Simulation to Use Buffers
    - **AC Ref:** Plan Recommendation 1 (Remove need for mock stdin files...)

- [ ] **Task Title:** Remove Obsolete `integration.RunInternal` Function
    - **Action:** Delete the `RunInternal` function from `internal/integration/test_runner.go` as it relies on the removed global variable modification pattern.
    - **Depends On:** Remove Global `architect.NewAPIService` Variable
    - **AC Ref:** Plan Recommendation 1 (Refactor problematic pattern)

- [x] **Task Title:** Refactor or Remove `integration.RunTestWithConfig` Function
    - **Action:** Analyze `RunTestWithConfig` in `internal/integration/test_runner.go`. Refactor it to call the modified `architect.Execute` by injecting the mock `APIService` from the `TestEnv`, or remove it entirely if tests can call `Execute` directly.
    - **Depends On:** Modify `architect.Execute` Signature for Dependency Injection, Refactor `TestEnv.Setup` to Eliminate Global I/O Redirection, Refactor `TestEnv` Stdin Simulation to Use Buffers
    - **AC Ref:** Plan Recommendation 1 (Refactor problematic pattern)

- [ ] **Task Title:** Update Integration Tests to Inject Dependencies
    - **Action:** Modify tests in `internal/integration/integration_test.go`. Instead of using `RunTestWithConfig` (or after refactoring it), directly call the modified `architect.Execute` or test smaller units, ensuring all necessary dependencies (mock `APIService`, loggers with buffers, config) are correctly instantiated and passed.
    - **Depends On:** Modify `architect.Execute` Signature for Dependency Injection, Refactor or Remove `integration.RunTestWithConfig` Function, Update Integration Tests to Pass I/O Buffers Explicitly, Update Integration Tests Using Stdin Simulation
    - **AC Ref:** Plan Recommendation 1, Plan Recommendation 4

- [ ] **Task Title:** Enable Parallel Execution for Integration Tests in CI
    - **Action:** Update the CI configuration script/pipeline step that runs integration tests (`go test ./internal/integration/...`) to include the `-parallel N` flag (where N is a suitable number based on CI runner cores, e.g., 4).
    - **Depends On:** Update Integration Tests to Inject Dependencies
    - **AC Ref:** Plan Recommendation 1 (Benefits: Enables `go test -parallel N`)

## 2. Optimize E2E Test Execution

- [ ] **Task Title:** Implement `TestMain` in E2E Tests to Build Binary Once
    - **Action:** Add or modify the `TestMain` function in `internal/e2e/e2e_test.go`. Implement logic within `TestMain` to call `findOrBuildBinary` (or a similar function) *once* before any tests are run. Store the resulting binary path in a package-level variable. Handle potential build errors gracefully.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 2 (Build Binary Once)

- [ ] **Task Title:** Update E2E Tests to Use Pre-Built Binary Path
    - **Action:** Modify the `RunArchitect` or `RunWithFlags` helper functions in `internal/e2e/e2e_test.go` to use the binary path stored in the package-level variable set by `TestMain`, instead of calling `findOrBuildBinary` repeatedly or relying on a hardcoded path.
    - **Depends On:** Implement `TestMain` in E2E Tests to Build Binary Once
    - **AC Ref:** Plan Recommendation 2 (Build Binary Once)

- [ ] **Task Title:** Investigate Sharing `httptest.Server` in E2E `TestMain`
    - **Action:** Analyze the mock server setup (`startMockServer`) and usage in `internal/e2e/e2e_test.go`. Determine if a single mock server instance started in `TestMain` can serve all E2E tests or if individual tests require distinct server behaviors that prevent sharing. Document findings.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 2 (Consider Shared Mock Server)

- [ ] **Task Title:** Implement Shared `httptest.Server` if Feasible
    - **Action:** Based on the investigation, if feasible, move the `httptest.Server` setup to `TestMain` in `internal/e2e/e2e_test.go`. Ensure proper server shutdown in `TestMain`. Update tests to use the shared server URL.
    - **Depends On:** Investigate Sharing `httptest.Server` in E2E `TestMain`
    - **AC Ref:** Plan Recommendation 2 (Consider Shared Mock Server)

- [ ] **Task Title:** Verify E2E Test Isolation
    - **Action:** Confirm that each E2E test uses isolated resources, particularly temporary directories (`t.TempDir()`) and potentially unique ports or configurations if not using a shared server, to allow for parallel execution.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 2 (Enable Parallel E2E Tests)

- [ ] **Task Title:** Enable Parallel Execution for E2E Tests in CI (Optional)
    - **Action:** If resources allow and tests are confirmed to be isolated, update the CI configuration script/pipeline step that runs E2E tests (`go test ./internal/e2e/...`) to include the `-parallel N` flag. Monitor for flakiness.
    - **Depends On:** Verify E2E Test Isolation
    - **AC Ref:** Plan Recommendation 2 (Enable Parallel E2E Tests)

- [ ] **Task Title:** Review and Reduce E2E Test Suite Scope
    - **Action:** Analyze the existing E2E tests in `internal/e2e/e2e_test.go`. Identify tests covering scenarios that are (or could be) adequately covered by integration tests. Remove redundant E2E tests, focusing on essential user flows.
    - **Depends On:** Update Integration Tests to Inject Dependencies
    - **AC Ref:** Plan Recommendation 2 (Reduce E2E Test Count)

## 3. Reduce Test Setup Overhead

- [ ] **Task Title:** Identify Repetitive Integration Tests for Consolidation
    - **Action:** Review tests in `internal/integration/integration_test.go`. Look for multiple tests that perform similar setup steps but vary slightly in inputs or configuration (e.g., testing different flags like `--dry-run`, different file filters).
    - **Depends On:** Update Integration Tests to Inject Dependencies
    - **AC Ref:** Plan Recommendation 3 (Convert repetitive tests...)

- [ ] **Task Title:** Convert Identified Tests to Table-Driven Format
    - **Action:** Refactor the identified repetitive integration tests into single test functions using the table-driven pattern (`tests := []struct{...}`). Define common setup once outside the loop and test-specific variations within the loop using `t.Run`.
    - **Depends On:** Identify Repetitive Integration Tests for Consolidation
    - **AC Ref:** Plan Recommendation 3 (Table-Driven Tests)

- [ ] **Task Title:** Group Related Integration Tests Using Sub-tests (`t.Run`)
    - **Action:** Identify groups of tests in `internal/integration/integration_test.go` that test different aspects of the same feature or component and could share common setup code. Refactor these groups to use a parent test function with shared setup, running individual test cases within `t.Run` blocks.
    - **Depends On:** Update Integration Tests to Inject Dependencies
    - **AC Ref:** Plan Recommendation 3 (Use Sub-tests)

- [ ] **Task Title:** Create Helper Functions for Common Test Setup Logic
    - **Action:** Identify recurring setup patterns within integration (`integration_test.go`) and E2E (`e2e_test.go`) tests (e.g., creating specific file structures, configuring mock responses). Extract this logic into reusable helper functions within the respective test packages.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 3 (Helper Functions)

## 4. Refine Integration Test Scope

- [ ] **Task Title:** Analyze Integration Tests Running Full `Execute`
    - **Action:** Review tests in `internal/integration/integration_test.go` that invoke the full `architect.Execute` function (or the refactored `RunTestWithConfig`). Determine if the primary goal of the test is to verify a smaller interaction (e.g., context gathering logic, token counting).
    - **Depends On:** Update Integration Tests to Inject Dependencies
    - **AC Ref:** Plan Recommendation 4 (Test Smaller Units)

- [ ] **Task Title:** Refactor Overly Broad Integration Tests
    - **Action:** For tests identified in the previous task, refactor them to directly test the specific components or collaborations involved (e.g., instantiate `ContextGatherer` and test its `GatherContext` method directly) instead of running the entire application flow via `Execute`. Inject necessary dependencies/mocks.
    - **Depends On:** Analyze Integration Tests Running Full `Execute`
    - **AC Ref:** Plan Recommendation 4 (Test Smaller Units, Focus on Boundaries)

## 5. Profile & Optimize

- [ ] **Task Title:** Integrate Test Profiling into Workflow/CI
    - **Action:** Add steps to the CI pipeline or document a manual process for running tests with profiling enabled (`go test ./... -cpuprofile cpu.prof -memprofile mem.prof -parallel N`). Ensure profiles can be easily collected and accessed.
    - **Depends On:** Enable Parallel Execution for Integration Tests in CI
    - **AC Ref:** Plan Recommendation 6 (Profile Your Tests)

- [ ] **Task Title:** Analyze Test Profiles and Identify Hotspots
    - **Action:** Run the profiled tests and use `go tool pprof` to analyze the generated `cpu.prof` and `mem.prof` files. Identify functions or test setups consuming disproportionate amounts of CPU time or memory.
    - **Depends On:** Integrate Test Profiling into Workflow/CI
    - **AC Ref:** Plan Recommendation 6 (Profile Your Tests)

- [ ] **Task Title:** Implement Optimizations Based on Profiling (Placeholder)
    - **Action:** Based on the analysis of profiling data, create and implement specific tasks to address identified bottlenecks in test code or setup. (Specific tasks TBD after analysis).
    - **Depends On:** Analyze Test Profiles and Identify Hotspots
    - **AC Ref:** Plan Recommendation 4, Plan Recommendation 6

## 6. Leverage CI Caching & Resources

- [ ] **Task Title:** Verify Go Module Caching in CI
    - **Action:** Inspect the CI configuration files (e.g., GitHub Actions workflow, GitLab CI YAML). Ensure that Go module dependencies (`GOPATH/pkg/mod`) are being cached between CI runs. Implement caching if missing.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 5 (Ensure CI config caches Go module dependencies)

- [ ] **Task Title:** Verify Go Build Caching (`GOCACHE`) in CI
    - **Action:** Inspect the CI configuration files. Ensure that the Go build cache directory (usually `~/.cache/go-build` or location specified by `GOCACHE`) is being cached between CI runs. Implement caching if missing.
    - **Depends On:** None
    - **AC Ref:** Plan Recommendation 5 (Cache Go build results)

- [ ] **Task Title:** Monitor CI Resource Usage Post-Parallelization
    - **Action:** After enabling parallel tests, monitor the CPU and memory usage of the CI runners during test execution. Observe if runners are becoming resource-constrained.
    - **Depends On:** Enable Parallel Execution for Integration Tests in CI, Enable Parallel Execution for E2E Tests in CI (Optional)
    - **AC Ref:** Plan Recommendation 5 (Monitor and adjust CPU/memory resources)
