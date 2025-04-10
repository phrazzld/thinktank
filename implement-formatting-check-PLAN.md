# Implement formatting check

## Goal
Add a step to the GitHub Actions workflow that checks if the Go code is properly formatted according to Go formatting standards using `go fmt`.

## Implementation Approach
Update the `.github/workflows/ci.yml` file to add a new step in the lint job after the dependency verification step:

1. Add a clearly named step for checking code formatting
2. Use a shell script that:
   - Runs `go fmt ./...` and captures any output (which indicates files that need formatting)
   - Checks if there's any output, and if so, fails the build with an appropriate error message
3. Ensure this step follows a logical order in the workflow (after dependency verification)

## Reasoning

1. **Fail on formatting issues**: Rather than just running `go fmt`, we'll check if any files need formatting and fail the build if so. This ensures that all code in the repository follows Go's formatting standards.

2. **Informative error message**: By providing a clear error message when formatting issues are found, developers will know exactly what's wrong and how to fix it.

3. **Shell script approach**: Using a multi-line shell script with conditional logic gives us more control over the behavior, specifically allowing us to fail the build only when formatting issues are detected rather than relying on the exit code of `go fmt` (which always returns 0 even if it makes formatting changes).

4. **Placement in the workflow**: The formatting check is a basic code quality step that should come early in the workflow, so placing it after dependency verification but before more complex linting steps is logical.