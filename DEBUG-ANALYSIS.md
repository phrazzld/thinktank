# Debug Analysis

## Summary of Hypotheses
Based on the analysis from all Gemini models, the following key hypotheses emerged:

1. **Rate Limit Tests** - Tests in `rate_limit_test.go` contain intentional delays for simulating rate limiting (time.Sleep), which may be causing the timeouts. These tests have a `-short` flag option but it's not being used in CI.

2. **Parallelism Issues** - The `-parallel 4` setting may be:
   - Creating resource contention on the CI runner
   - Not optimal for the available resources
   - Causing subtle deadlocks or extreme slowdowns

3. **Specific Slow Tests** - Individual test cases may be taking exceptionally long to run, dragging down the overall execution time.

4. **Test Setup/Teardown Inefficiency** - The test environment setup or cleanup might be consuming significant time.

## Key Implementation Details

The integration tests have several characteristics that could contribute to timeouts:

1. **Deliberate Time Delays**: Several tests use `time.Sleep()` to simulate real-world conditions:
   - `rate_limit_test.go` has tests that deliberately introduce delays to simulate rate limiting
   - `multi_model_test.go` includes concurrency tests with sleeps

2. **Race Detection**: The `-race` flag adds significant overhead to test execution

3. **No Short Mode**: The CI configuration isn't using the `-short` flag, which would skip known long-running tests

## Recommended Next Test

All models agree that modifying the CI workflow is the logical next step, though they differ slightly on the approach:

1. **Option 1**: Add the `-short` flag to skip tests that are known to be slow:
   ```yaml
   run: go test -v -race -short -parallel 4 ./internal/integration/...
   ```

2. **Option 2**: Run tests sequentially to identify slow tests:
   ```yaml
   run: go test -v -race -parallel 1 ./internal/integration/...
   ```

3. **Option 3**: Run each test individually to measure execution times:
   ```bash
   for test_case in $(go test -list ./internal/integration/... | grep "^Test"); do
     echo "Running test case: $test_case"
     start_time=$(date +%s)
     go test -v -race -run "$test_case" ./internal/integration/...
     end_time=$(date +%s)
     duration=$((end_time - start_time))
     echo "Test case '$test_case' took $duration seconds"
   done
   ```

Given our constraint of diagnosing with minimal changes first, option 2 seems most appropriate as the first test - it maintains all tests while giving us valuable timing data without significant script changes.

## Long-term Recommendations

Once the specific cause is identified, potential fixes might include:

1. Always use `-short` flag in CI for integration tests
2. Refactor tests to use mock time instead of real `time.Sleep()`
3. Split integration tests into separate jobs with different timeouts
4. Increase the timeout for the integration test step (currently 5 minutes)
5. Optimize test setup/teardown procedures