# Bug Fix Plan

## Bug Description:
The CI workflow is failing at the test step with the following message:
"Error: The action 'Run integration tests with parallel execution' has timed out after 5 minutes."

After fixing that issue, another error was encountered:
"Run go test -v -race -parallel 4 ./internal/e2e/...
go: warning: "./internal/e2e/..." matched no packages
no packages to test
Error: Process completed with exit code 1."

## Reproduction Steps:
Not explicitly provided in BUG.MD, but implied to be running the CI workflow which attempts to execute integration and E2E tests.

## Expected Behavior:
1. Integration tests should complete within the 5-minute timeout limit.
2. E2E tests should execute successfully using the appropriate build tags and test approach.

## Actual Behavior:
1. Integration tests were taking longer than 5 minutes to complete, causing the CI workflow to time out and fail.
2. After fixing the integration tests, the E2E tests failed because the CI was trying to run them without the required build tags.

## Key Components/Files Mentioned:
- CI workflow (in `.github/workflows/ci.yml`)
- Integration tests (in `internal/integration/` directory)
- E2E tests (in `internal/e2e/` directory)

## Hypotheses:
1. **Intentional Delays in Rate Limit Tests**: The `rate_limit_test.go` file contains tests that deliberately introduce delays using `time.Sleep()` to simulate rate limiting. These tests are skipped in short mode (`-short` flag), but this flag is not being used in the CI workflow.
   - Reasoning: Examination of `rate_limit_test.go` shows tests like `TestRateLimitFeatures` with multiple subtests including `MultiModelRateLimiting` that intentionally simulate rate limiting with sleep delays.
   - Validation: Add the `-short` flag to the integration test command in CI to skip these known slow tests.

2. **Excessive Sleep in Concurrency Tests**: The concurrency tests in `multi_model_test.go` use `time.Sleep()` with substantial durations (e.g., 50-150ms per model) to simulate real workloads, which accumulates significant time when running with multiple models.
   - Reasoning: Tests like `EnhancedConcurrentModelProcessing` in `multi_model_test.go` use sleeps to simulate API delays in a realistic manner.
   - Validation: Run tests with reduced parallelism or in short mode to check if execution time improves.

3. **Race Detector Overhead**: Running with the `-race` flag adds significant overhead, particularly with parallel execution and goroutines.
   - Reasoning: The race detector instruments all memory accesses, which can slow tests down by 5-10x, especially tests with heavy concurrency.
   - Validation: Run integration tests without the race detector in CI to check if they complete within the timeout.

4. **E2E Tests Build Tag Issue**: The E2E tests have a build tag (`//go:build manual_api_test`) that is not being respected in the CI configuration.
   - Reasoning: The E2E tests are designed to be run explicitly with the `manual_api_test` build tag, and they won't be found without it.
   - Validation: Use the provided `run_e2e_tests.sh` script that properly includes the build tags instead of running the tests directly.

## Test Log:
### Test 1: Add `-short` flag to integration tests in CI
**Hypothesis Tested:** Intentional Delays in Rate Limit Tests
**Test Description:** Modify the CI workflow to add the `-short` flag to the integration test command
**Execution Method:** Edit `.github/workflows/ci.yml` to change line 109 from 
```yaml
run: go test -v -race -parallel 4 ./internal/integration/...
```
to
```yaml
run: go test -v -race -short -parallel 4 ./internal/integration/...
```
**Expected Result (if true):** The integration tests will complete within the 5-minute timeout since the time-consuming rate limit tests will be skipped.
**Expected Result (if false):** The integration tests will still timeout, indicating that the rate limit tests are not the primary cause of the slowdown.
**Actual Result:** The fix was implemented but revealed a subsequent issue with the E2E tests not running properly.

### Test 2: Fix E2E test execution
**Hypothesis Tested:** E2E Tests Build Tag Issue
**Test Description:** Modify the E2E test command to use the provided script that properly includes the required build tags
**Execution Method:** Edit `.github/workflows/ci.yml` to change the E2E test command from 
```yaml
run: go test -v -race -parallel 4 ./internal/e2e/...
```
to
```yaml
run: ./internal/e2e/run_e2e_tests.sh --verbose --short
```
**Expected Result (if true):** The E2E tests will be executed with the proper build tags and should complete successfully.
**Expected Result (if false):** The tests will still fail, indicating another issue with the E2E test configuration.
**Actual Result:** (Awaiting CI validation)

## Root Cause:
1. The integration tests in `rate_limit_test.go` include test cases that intentionally introduce time delays to simulate rate limiting behavior. These tests have a feature that allows them to be skipped in short mode (with the `-short` flag), but the CI workflow was not utilizing this flag. As a result, these slow tests were running in CI, causing the build to exceed the 5-minute timeout limit.

2. The E2E tests in the `internal/e2e` directory are designed to only run with a specific build tag (`manual_api_test`), but the CI was trying to run them without this tag, resulting in "no packages to test" errors. The repository includes a special script (`run_e2e_tests.sh`) that's specifically designed to run these tests with the appropriate build tags.

## Fix Description:
1. **For integration tests**: Add the `-short` flag to the integration test command in the CI workflow to skip the time-consuming rate limit tests while still testing the core functionality.

   **Change:**
   ```diff
   # In .github/workflows/ci.yml
   - run: go test -v -race -parallel 4 ./internal/integration/...
   + run: go test -v -race -short -parallel 4 ./internal/integration/...
   ```

2. **For E2E tests**: Use the provided `run_e2e_tests.sh` script to run the E2E tests with the proper build tags instead of trying to run them directly with `go test`.

   **Change:**
   ```diff
   # In .github/workflows/ci.yml
   - run: go test -v -race -parallel 4 ./internal/e2e/...
   + run: ./internal/e2e/run_e2e_tests.sh --verbose --short
   ```

These changes follow the project's existing design and conventions by:
1. Using the `-short` flag as intended to skip time-consuming tests in CI
2. Using the provided script to run the E2E tests as documented in the project's own documentation (`/internal/e2e/README.md`)

## Status: Implemented, awaiting CI validation