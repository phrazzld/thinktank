# Implement .gitignore-based filtering logic

## Task Goal
Add logic to parse .gitignore files and use their patterns to filter files and directories during traversal, ensuring that files specified in .gitignore rules are excluded from the context files used in prompts.

## Chosen Implementation Approach
I'll implement this task by:

1. Modifying the `readDirectoryContents` function in `fileReader.ts` to:
   - Use the `gitignoreUtils` module instead of the hardcoded `IGNORED_DIRECTORIES` array
   - Check each file/directory against the gitignore rules before including it in results
   - Maintain a cache of gitignore filters to improve performance during recursive traversal
   - Continue handling errors gracefully with the existing error reporting mechanism

2. Creating a wrapper function that:
   - Takes a directory path and recursively processes its contents
   - Applies gitignore rules at each level of the directory tree
   - Returns file contents that don't match ignore patterns

3. Ensuring proper relative path handling:
   - Convert absolute paths to relative ones for gitignore pattern matching
   - Maintain original paths in the result objects
   - Handle nested .gitignore files that may exist in subdirectories

## Reasoning for this Approach
I considered several approaches:

1. **Replace only the hardcoded IGNORED_DIRECTORIES array**:
   - Pros: Minimal code changes, simple to implement
   - Cons: Doesn't handle nested .gitignore files, doesn't respect complex patterns

2. **Use a file filtering library completely separate from our existing code**:
   - Pros: Might be more robust for complex edge cases
   - Cons: Duplicates functionality, requires more dependencies, breaks existing patterns

3. **Integrate gitignore filtering directly into the existing directory traversal**:
   - Pros: Maintains existing code structure and error handling, efficient (checks patterns during traversal)
   - Cons: More complex implementation, needs to handle relative paths carefully

I selected the third approach (integration) because:
- It leverages our existing, well-tested directory traversal logic
- It uses the newly created gitignore utilities without duplicating code
- It provides a seamless experience with proper error handling
- It follows the existing patterns in the codebase
- It can handle nested .gitignore files correctly
- It's more efficient than filtering after traversal is complete

This approach ensures that we respect .gitignore rules at every level of the directory tree, providing a consistent experience that matches how git itself handles ignore patterns.