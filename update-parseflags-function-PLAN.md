# Implementation Plan: Update parseFlags() function to modify task-file flag description

## Task
Modify the task-file flag description in parseFlags() to indicate it's required. Update the taskFlag description to indicate it's deprecated.

## Chosen Approach
I'll implement Approach 2 from the analysis: **Using Constants for Descriptions**

This approach introduces constants for the flag descriptions, similar to how other constants like `defaultOutputFile` and `defaultModel` are handled in the codebase. This aligns with the existing patterns in the code and provides better maintainability.

### Implementation Steps:
1. Define constants for the new flag descriptions at the top of `main.go` near the other constants.
2. Use these constants in the `flag.String` calls in the `parseFlags()` function.

### Code Changes:

```go
// main.go - near the top constants
const (
    // ... other constants ...
    taskFlagDescription     = "Description of the task (deprecated: use --task-file instead)."
    taskFileFlagDescription = "Path to a file containing the task description (required)."
)

// main.go - inside parseFlags()
// Use constant for taskFlag description
taskFlag := flag.String("task", "", taskFlagDescription)

// Use constant for taskFileFlag description
taskFileFlag := flag.String("task-file", "", taskFileFlagDescription)
```

## Reasoning for Choice

I selected this approach for the following reasons:

1. **Alignment with Existing Patterns**: The codebase already uses constants extensively for default values, environment variable names, and default exclude patterns. Using constants for descriptions maintains this established pattern.

2. **Maintainability**: Though slightly more verbose, it centralizes the description strings. If future changes require updating the wording (e.g., standardizing how "required" or "deprecated" is phrased), constants make this easier and less error-prone.

3. **Clarity**: Using constants clearly signals that these are predefined, standardized pieces of text, making the code more readable and self-documenting.

4. **Consistency**: This approach creates consistency with how other configuration values are handled in the codebase.

5. **Testability**: While flag descriptions themselves are not typically unit-tested directly, this approach makes no negative impact on the testability of the flag parsing or validation logic compared to other approaches.

## Note
This change only affects the help text displayed to users. The actual enforcement of the required flag will be implemented in the `validateInputs()` function, which is a separate task.