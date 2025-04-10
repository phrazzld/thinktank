# Add coverage reporting

## Goal
Enhance the test job in the GitHub Actions workflow to generate test coverage reports for the Go project.

## Implementation Approach
Update the test job in the `.github/workflows/ci.yml` file to:

1. Add a new step after the race detection testing step to generate coverage reports
2. Use the `-coverprofile` flag with the `go test` command to save coverage data to a file
3. Ensure the coverage generation is separate from race detection to avoid conflicts
4. Optionally add a step to generate a human-readable coverage report for review

## Reasoning

1. **Separation of concerns**: While we could combine test execution with coverage reporting, keeping them separate provides more flexibility and clarity. Race detection and coverage measurement both add overhead to test execution, and combining them might lead to slower tests and less clear error messages.

2. **Comprehensive coverage reporting**: Using `-coverprofile=coverage.out ./...` ensures coverage data is collected for all packages in the project, providing a complete picture of code coverage.

3. **File-based approach**: Generating a coverage file creates an artifact that can be:
   - Used in subsequent steps for coverage threshold checking
   - Potentially uploaded as a workflow artifact
   - Used for integration with external code coverage tools in the future

4. **Performance considerations**: Coverage reporting adds some overhead but is generally faster than race detection. Since we already have a timeout set for the test job, we don't need to add a separate timeout specifically for the coverage step.