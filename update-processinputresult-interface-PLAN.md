# Update ProcessInputResult interface

## Goal
Update the ProcessInputResult interface in runThinktankTypes.ts to properly reflect the combined prompt+context content structure, ensuring type safety throughout the workflow.

## Implementation Approach
The ProcessInputResult interface currently returns an ExtendedInputResult (which includes context metadata) and an optional array of context files. However, the _executeQueries function in runThinktank.ts expects a simple prompt string, not the full ExtendedInputResult.

I'll modify the ProcessInputResult interface to ensure that it properly reflects the combined content structure while maintaining compatibility with the rest of the workflow. The main changes will be:

1. Ensure the inputResult.content field already contains the combined prompt+context content (this is already done in _processInput)
2. Add a new property to the ProcessInputResult interface called "combinedContent" that directly exposes the content field from inputResult
3. This provides a more explicit access path for downstream components like _executeQueries

This approach maintains backward compatibility with existing code while providing a more explicit way to access the combined content.

## Reasoning
I chose this approach because:

1. **Minimal Changes**: It requires minimal changes to the codebase while achieving the desired result.
2. **Type Safety**: It maintains type safety throughout the workflow.
3. **Backward Compatibility**: It doesn't break existing code that might be using the current interface.
4. **Clear Intent**: The new "combinedContent" property makes it explicit that the content field contains combined prompt+context.
5. **Follows Project Patterns**: This approach aligns with the project's pattern of providing accessor properties for important data.

By adding this property, we make it clearer to developers that the content includes both the prompt and context files when applicable, which improves code readability and maintainability.