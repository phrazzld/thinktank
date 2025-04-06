# Add dependency for gitignore parsing

## Task Goal
Research and add a dependency for parsing .gitignore files to properly filter files and directories during traversal, enabling the system to respect .gitignore rules when gathering context files.

## Chosen Implementation Approach
I'll implement this task by:

1. Adding the `ignore` npm package as a project dependency
2. Creating a utility module that wraps this library to:
   - Load and parse .gitignore files from project directories
   - Provide a simple filtering function to check if a path should be ignored
   - Handle scenarios where .gitignore doesn't exist (fallback to default ignores)
   - Cache parsed ignore patterns for performance

This implementation will serve as the foundation for the upcoming "Implement .gitignore-based filtering logic" task that will integrate this functionality into the directory traversal system.

## Reasoning for this Approach
I considered several approaches:

1. **Custom implementation without dependencies**:
   - Pros: No external dependencies, complete control
   - Cons: Complex to implement correctly, time-consuming, likely to have edge cases

2. **Using `glob` or `globby` packages**:
   - Pros: Powerful glob matching with built-in .gitignore support 
   - Cons: Primarily focused on file finding rather than just pattern matching, heavier dependencies

3. **Using the lightweight `ignore` package**:
   - Pros: Simple API focused solely on gitignore pattern matching, lightweight, well-maintained
   - Cons: Requires some additional code to handle reading .gitignore files

I selected the `ignore` package because:
- It's purpose-built for .gitignore pattern parsing and matching
- It's lightweight and has minimal dependencies
- It has a simple, clear API
- It's well-maintained and widely used
- It follows the exact same rules as git's ignore pattern matching

This approach balances simplicity, performance, and correctness. The `ignore` package handles the complex pattern matching logic while we'll write the minimal code needed to integrate it with our application's workflow.