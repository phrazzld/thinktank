# Implementation Plan: Update flag.Usage() message in parseFlags()

## Task
Update the usage message to show --task-file as the primary means of providing input, as shown in the Implementation Specifications.

## Chosen Approach
I'll implement **Approach 2: Update Usage String and Flag Description** as recommended in the analysis.

This approach involves:
1. Modifying the usage message in the flag.Usage function to only mention --task-file
2. Ensuring the taskFlagDescription constant clearly indicates that --task is deprecated

### Implementation Steps:
1. Locate the `flag.Usage` function definition in `parseFlags()` function
2. Modify the first `fmt.Fprintf` line to remove the mention of the `--task` flag and show only `--task-file` as the primary input method
3. Verify the `taskFlagDescription` constant is already updated appropriately from our previous task

### Code Changes:

```go
// Inside parseFlags() function
flag.Usage = func() {
    // --- CHANGE THIS LINE ---
    // Original: fmt.Fprintf(os.Stderr, "Usage: %s (--task \"<description>\" | --task-file <path>) [options] <path1> [path2...]\n\n", os.Args[0])
    // New:
    fmt.Fprintf(os.Stderr, "Usage: %s --task-file <path> [options] <path1> [path2...]\n\n", os.Args[0])
    // --- END CHANGE ---

    // Rest of the function remains unchanged
    fmt.Fprintf(os.Stderr, "Arguments:\n")
    fmt.Fprintf(os.Stderr, "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")
    fmt.Fprintf(os.Stderr, "Options:\n")
    flag.PrintDefaults()
    fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
    fmt.Fprintf(os.Stderr, "  %s: Required. Your Google AI Gemini API key.\n", apiKeyEnvVar)
}
```

## Reasoning for Choice

I selected Approach 2 for the following reasons:

1. **Clarity and Consistency**: This approach clearly communicates that `--task-file` is now the primary method for providing input while still showing the deprecated `--task` option in the detailed help output with an appropriate deprecation notice.

2. **Balanced Approach**: It strikes a good balance between fulfilling the immediate requirement (updating the usage message) and supporting the broader goal of deprecating the `--task` flag. It doesn't prematurely remove functionality but clearly marks it as deprecated.

3. **Low Risk**: The changes are minimal and targeted, reducing the risk of introducing bugs. We're only modifying the help text output, not changing the actual flag parsing or validation logic.

4. **Standards Alignment**: This approach follows standard practices for deprecating command-line flags by keeping them functional but marking them as deprecated in the help text.

5. **Minimal Maintenance Burden**: The implementation is straightforward and doesn't require complex logic changes in multiple places, making it easy to maintain.

## Note
This change only affects the help text displayed to users. The changes to make `--task-file` required and to handle the deprecated `--task` flag will be implemented in separate tasks, specifically in the `validateInputs()` function.