# Create context file reader utility function

## Task Goal
Create a new utility function that reads content from a single file, properly handles errors, and returns both the file path and content. This function will be used as a building block for reading context files that will be included with prompts.

## Chosen Implementation Approach
I'll create a new function called `readContextFile` in the existing utils/fileReader.ts file that:

1. Takes a file path as input
2. Validates that the file exists and is readable
3. Reads the file content as text
4. Returns an object with both the file path and content
5. Appropriately handles and reports errors without throwing (returning error information instead)

## Reasoning for this Approach
I considered several approaches:

1. **Creating a new file for context-specific utilities**: This would separate context reading from other file operations but would add complexity and potentially duplicate functionality.

2. **Adding the function to fileReader.ts**: The project already has a dedicated fileReader.ts utility that handles file reading operations, making it the natural place for this new function. This follows the existing code organization pattern and keeps related functionality together.

3. **Using synchronous vs asynchronous file reading**: Since file reading is I/O bound, asynchronous operations would be more efficient for multiple files. However, examining the existing code shows a preference for Promise-based asynchronous file operations.

I selected the second approach with asynchronous reading because:
- It keeps related functionality grouped together (file reading utilities)
- It follows existing code patterns in the project
- It doesn't introduce unnecessary new files
- Asynchronous reading is more scalable when dealing with multiple files
- Returning objects with both path and content (vs just content) provides more context for error handling and formatting later

The error handling approach will match the project's pattern of returning objects with status information rather than throwing exceptions, allowing the caller to decide how to proceed when errors occur.