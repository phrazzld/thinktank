# Implement coverage threshold check

## Goal
Add a step to the GitHub Actions workflow that verifies test coverage meets or exceeds the specified threshold of 80%.

## Implementation Approach
Update the test job in the `.github/workflows/ci.yml` file to:

1. Add a new step after the coverage reporting steps
2. Extract the total coverage percentage from the coverage report
3. Compare the extracted percentage against the 80% threshold
4. Fail the workflow if the coverage is below the threshold

## Reasoning

1. **Shell script approach**: Using a shell script with command substitution and conditional logic allows us to extract the coverage percentage, perform the comparison, and provide informative feedback all in one step.

2. **Use of standard Go tools**: We'll use `go tool cover -func=coverage.out` to get the coverage percentage, which is the standard way to read coverage data in Go. This command outputs coverage for each function and a total at the end, which we can parse with standard Unix tools like grep and awk.

3. **Pure shell implementation**: Rather than introduce an external dependency on the `bc` command (which was noted as a concern in the CLARIFICATIONS NEEDED section), we'll use shell arithmetic comparisons for a more portable solution. This ensures the workflow will work across different CI environments.

4. **Clear error messaging**: If the coverage threshold is not met, we'll provide an informative error message that includes the actual coverage percentage. This helps developers understand why the build failed and by how much they need to improve coverage.

5. **Flexibility for future adjustments**: By explicitly specifying the threshold in the script, it's easy to adjust it later if needed, without having to modify the script logic.