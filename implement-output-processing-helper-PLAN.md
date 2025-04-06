# Implement Output Processing Helper

## Task Goal
Create the `_processOutput` helper function that handles file writing and console output formatting with spinner updates. This function will catch and wrap errors using `FileSystemError` when needed.

## Implementation Approach

I'll implement a modular helper function that takes responsibility for:
1. Updating the spinner status during the file writing process
2. Handling output operations via the existing `outputHandler` module
3. Providing consistent error handling and wrapping with appropriate `FileSystemError` types
4. Updating the spinner with success/warning messages based on output results

The function will follow the same pattern established by the existing helper functions, taking a parameters object that includes:
- The spinner instance for updating UI status
- Query results to process
- Output directory path
- User options for formatting the output

The implementation will leverage the existing `processOutput` function from the outputHandler module, but will add proper error handling and spinner updates throughout the process, ensuring that file system errors are properly caught and wrapped in the appropriate error types.

### Key Components:
1. **Spinner Status Updates**: Throughout the output processing, the spinner will be updated with clear status messages
2. **Error Handling**: Catch and categorize error types from the file system operations
3. **Result Processing**: Return a properly structured result object with file output and console output

## Alternatives Considered

1. **Direct File System Operations**: Instead of using the outputHandler module, implement direct file writing operations within this helper. This would increase control over the process but would duplicate functionality already available in the outputHandler module.

2. **Splitting File and Console Output**: Create separate helper functions for file output and console output. This would increase modularity but would require more complexity in the main workflow function.

3. **Async Spinner Updates**: Implement a more sophisticated status update system that shows real-time file write progress. This would improve user experience but add significant complexity.

## Reasoning for Selected Approach

I've chosen to implement a wrapper around the existing `processOutput` function from the outputHandler module because:

1. **Reuse of Existing Functionality**: The outputHandler module already contains robust file writing and console formatting functionality.

2. **Consistent Error Handling**: This approach allows us to maintain the same error handling pattern used in other helper functions, ensuring consistent error types and messages.

3. **Clear Responsibility Boundary**: The helper function will focus on spinner updates and error handling, while delegating the actual output processing to the appropriate module.

4. **Maintainability**: By keeping the implementation focused on integration rather than reimplementing output functionality, the code will be easier to maintain and update.

The approach aligns with the project's functional style and modular architecture, with clear separation of concerns between the workflow orchestration, spinners/UI updates, and the actual file system operations.