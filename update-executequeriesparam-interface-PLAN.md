# Update ExecuteQueriesParams interface

## Goal
Update the ExecuteQueriesParams interface to accept combined prompt+context content, ensuring proper handling of the combined content throughout the query execution workflow.

## Implementation Approach
I will rename the `prompt` property to `combinedContent` in the ExecuteQueriesParams interface. This explicit naming clarifies that the content may include both the user prompt and context files. This approach requires updating all references to the property throughout the codebase.

The implementation steps will be:
1. Update the ExecuteQueriesParams interface in runThinktankTypes.ts to rename `prompt` to `combinedContent`
2. Update the executeQueries function in queryExecutor.ts to use the new parameter name
3. Update the call to _executeQueries in runThinktank.ts to pass inputResult.combinedContent instead of inputResult.inputResult.content

## Reasoning
I chose this direct property name change approach over alternatives (keeping both properties or just updating comments) for several reasons:

1. **Explicit Intent**: The name `combinedContent` clearly indicates that the content may include both the prompt and additional context files.

2. **Consistency**: This aligns with our previous change of adding a `combinedContent` property to ProcessInputResult, maintaining consistency across the codebase.

3. **Clean Design**: Avoiding redundant properties with the same meaning keeps the interface clean and prevents potential confusion or bugs from inconsistent property usage.

4. **Development Stage**: Since the codebase is still in active development, making breaking changes now is less disruptive than it would be in a stable release.

This approach prioritizes code clarity and maintainability, making it explicit throughout the codebase that we're working with combined prompt+context content rather than just a simple prompt.