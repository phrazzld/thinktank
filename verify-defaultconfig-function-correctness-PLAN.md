# Verify DefaultConfig() Function Correctness

## Task Title
Verify `DefaultConfig()` function correctness

## Implementation Approach
Ensure the `DefaultConfig()` function in `internal/config/config.go` correctly initializes all fields of the simplified `AppConfig` struct with appropriate default values after removing fields from the last task.

## Steps
1. Review the current `AppConfig` struct fields in `internal/config/config.go`
2. Verify the `DefaultConfig()` function initializes all these fields with appropriate values
3. Check for any fields that might be missing initialization
4. Ensure that appropriate default values are used from the defined constants