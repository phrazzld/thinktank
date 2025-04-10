# Add race detection testing

## Goal
Enhance the test job in the GitHub Actions workflow to include race condition detection when running Go tests.

## Implementation Approach
Update the test job in the `.github/workflows/ci.yml` file to:

1. Modify the existing test step to include the `-race` flag in the `go test` command
2. Keep other test parameters like verbosity and timeout
3. Ensure clear step naming to indicate race detection is now included

## Reasoning

1. **Detection of race conditions**: The `-race` flag enables Go's built-in race detector, which identifies potential race conditions in concurrent code. This is critical for ensuring thread-safety in Go applications, especially since Go's concurrency model makes it relatively easy to write concurrent code.

2. **Early detection of issues**: Identifying race conditions during CI rather than in production is invaluable, as race conditions can be extremely hard to reproduce and debug once they occur in production environments.

3. **Modification vs. addition**: Rather than adding a separate test step, it makes more sense to enhance the existing test step. This approach is cleaner and prevents duplication of test execution, since we want all tests to be checked for race conditions.

4. **Performance considerations**: Race detection does add significant overhead to test execution (both memory and CPU), but the benefits of catching race conditions early outweigh the performance cost in a CI environment. We'll maintain the existing timeout setting, which should provide enough time for the race-enabled tests to complete.