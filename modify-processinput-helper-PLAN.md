# Modify _processInput helper

## Task Goal
Update the `_processInput` helper function in `src/workflow/runThinktankHelpers.ts` to accept context paths from command-line options, read content from these paths using the new `readContextPaths` function, and combine the prompt with context content using the `formatCombinedInput` function.

## Chosen Implementation Approach
I'll modify the `_processInput` helper function and associated types with the following changes:

1. Update the `ProcessInputParams` interface to include an optional `contextPaths` field from the options
2. Modify the `_processInput` function to:
   - Check if contextPaths are provided in the options
   - If provided, use the `readContextPaths` function to read all context files/directories
   - Use the `formatCombinedInput` function to combine prompt content with context files
   - Update the spinner text to reflect context processing status
   - Include information about context files in the result (counts, etc.)
3. Update the return type `ProcessInputResult` to include context information

The implementation will:
- Handle both cases (with and without context paths) gracefully
- Process context paths concurrently for better performance
- Provide appropriate error handling for context paths
- Update spinner messages to show context processing progress
- Log informative messages about the number of context files included

## Reasoning for Approach
I considered three approaches for integrating context paths:

1. **Integrated approach (selected)**: Modify the existing `_processInput` function to handle context paths directly.
   - Pros: Maintains the single responsibility of input processing, follows existing workflow pattern, keeps code modular
   - Cons: Makes the function slightly more complex

2. **Separate function approach**: Create a new dedicated function for context processing.
   - Pros: Cleaner separation of concerns, more focused functions
   - Cons: Complicates the workflow orchestration, requires changing the main workflow

3. **Combined function with conditional paths**: Process input and context paths together in one flow.
   - Pros: Potentially more efficient if there are common processing steps
   - Cons: More complex function, harder to test, less flexible

I selected the integrated approach because:
- It maintains the existing workflow structure
- It keeps all input processing logic in one place
- It follows the established pattern of the codebase
- It requires minimal changes to other parts of the application
- It keeps error handling consistent with the rest of the workflow
- It allows reuse of the existing spinner and logging infrastructure

The implementation provides a clean extension to the current functionality without disrupting the overall architecture of the application.