# Create directory reader utility function

## Task Goal
Implement a recursive directory reading function that traverses directories and collects file contents. This will enable users to include entire directories as context when using thinktank.

## Chosen Implementation Approach
I'll create a function called `readDirectoryContents` in the utils/fileReader.ts file that:

1. Takes a directory path as input
2. Recursively traverses the directory structure
3. For each file encountered:
   - Calls our existing `readContextFile` function to read the file content
   - Maintains the original relative path in results for better context
4. Returns an array of ContextFileResult objects (the same format as readContextFile)
5. Skip node_modules and other common "ignore" directories for this MVP version
   (actual .gitignore filtering will be handled in a separate task)

## Reasoning for this Approach
I considered several approaches:

1. **Using a third-party library** like `glob` or `fast-glob` to handle directory traversal:
   - Pros: Potentially more robust, handles edge cases
   - Cons: Adds external dependency, might be overkill for our needs, less control over exactly how traversal works

2. **Creating our own recursive function**:
   - Pros: No additional dependencies, complete control over behavior, consistent error handling
   - Cons: Need to handle various edge cases ourselves (symlinks, permissions, etc.)

3. **Non-recursive approach with predefined depth limit**:
   - Pros: Simpler implementation, controls maximum nesting depth
   - Cons: Arbitrary limitations, incomplete directory traversal

I selected the second approach (creating our own recursive function) because:
- It follows the project's preference for limiting external dependencies
- It provides consistent error handling with our existing readContextFile function 
- The file system operations are already async, so we'll maintain that pattern
- It allows easy integration with future .gitignore filtering
- The implementation is not overly complex for our needs
- We can add features like max depth control if needed later

Additionally, we'll use a basic ignore list for now (node_modules, .git, etc.) rather than trying to implement full .gitignore parsing, since that will be implemented in a separate task.