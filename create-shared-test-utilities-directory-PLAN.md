# Create shared test utilities directory

## Goal
Create a new directory structure to house shared test helper functions that will be used across multiple test files, specifically for mocking filesystem and gitignore utilities.

## Implementation Approach
The chosen approach is to create a directory at `src/__tests__/utils/` with a basic README.md file explaining the purpose of the utilities.

### Alternatives Considered:
1. **Create at src/__tests__/utils/:** Place utilities directly under the src directory's tests folder, keeping test utilities close to the source code they support.
2. **Create at tests/utils/:** Create a top-level tests directory with a utils subdirectory, separating test code from source code.
3. **Create at src/utils/__tests__/mocks/:** Place utilities as a subdirectory of the specific module being tested.

### Reasoning for Selected Approach:
I've chosen the first approach (`src/__tests__/utils/`) for these reasons:

1. **Consistency with project structure:** Looking at the existing project structure, test files are already organized within `__tests__` directories alongside the code they test. This follows Jest's conventional approach.

2. **Accessibility:** Placing utilities in a common location within the src tree makes them easily importable from any test file without complex relative paths.

3. **Scope:** These utilities are specifically for testing, not part of the main application, so they belong in a test-related directory rather than mixed with production code.

4. **Organization:** Having a dedicated utils directory within tests provides clear separation between test files and test helpers/utilities.

The implementation will be simple - creating the directory and a basic README.md file to explain the purpose and usage patterns for the utilities that will be added later.