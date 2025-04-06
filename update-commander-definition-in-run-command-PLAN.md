# Update `commander` definition in run command

## Goal
Update the Commander.js definition in the run command to accept variadic context path arguments after the prompt file argument. This will allow users to provide an arbitrary number of files and directories as additional context for their prompts.

## Implementation Approach
I will modify the commander definition in `src/cli/commands/run.ts` by:

1. Updating the `.argument()` call to add a second variadic argument `[contextPaths...]` after the existing `<promptFile>` argument
2. Adding clear help text for this new parameter explaining that it accepts multiple file/directory paths
3. Ensuring the action handler function signature is updated to receive this new parameter

The commander definition can support variadic arguments using the `...` syntax, which is exactly what we need for an arbitrary number of context paths.

## Reasoning for this Approach
I chose this approach because:

1. The Commander.js library has built-in support for variadic arguments, making it straightforward to implement
2. It maintains the current CLI structure and syntax, just extending it with additional arguments
3. It follows the common CLI convention where the main argument comes first followed by optional additional arguments
4. This approach will make the CLI intuitive for users who work with other command-line tools

Alternative approaches I considered:
- Using a separate option with a comma-separated list (like the existing models option), but this is less flexible and harder to work with file paths that might contain commas
- Using a configuration file to specify context paths, but this would make the CLI less convenient for ad-hoc usage