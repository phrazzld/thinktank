# Create standardized path handling helper

## Goal
Develop a shared helper function for path normalization to ensure consistent handling of leading slashes and path separators across all tests.

## Chosen Implementation Approach
Create a centralized `normalizePath` utility function that leverages the Node.js built-in `path.posix` module to normalize paths with consistent handling of separators and leading slashes.

The implementation will:
1. Use `path.posix.normalize` to handle path normalization consistently (ensuring forward slashes)
2. Add explicit handling for leading slashes based on the project's requirements
3. Place the function in a dedicated utility file (`src/__tests__/utils/pathUtils.ts`)
4. Include thorough unit tests for the function

## Reasoning
This approach was chosen because:
- It aligns with the project's standards by using Node.js built-in modules instead of adding external dependencies
- It leverages `path.posix` for robust handling of path normalization across different platforms
- It provides an easily maintainable and testable solution in a centralized location
- It's compatible with the virtual filesystem (`memfs`) which has specific requirements for path formats
- It strikes the right balance between simplicity and robustness, handling edge cases without overcomplicating the implementation