# Refactor fileReader.test.ts

## Goal
Refactor the fileReader.test.ts file to use the new virtualFsUtils approach instead of the deprecated mockFsUtils approach, ensuring all tests pass successfully.

## Implementation Approach
The refactoring will convert the tests from using direct Jest mocks with mockFsUtils to using the memfs-based virtualFsUtils approach. This will involve:

1. Updating the import structure to use virtualFsUtils and properly set up the Jest mocks
2. Refactoring the test setup to create a virtual filesystem structure instead of individual mocks
3. Adapting test assertions to verify actual filesystem state rather than checking mock calls
4. Converting error simulation approach to use spies on fs methods when needed

## Key Implementation Details

### 1. Import and Mock Setup
- Replace mockFsUtils imports with virtualFsUtils imports
- Add Jest mock setup for fs and fs/promises before importing these modules
- Move the import of os after the mock setup to ensure consistent mocking order

### 2. Test Setup Changes
- Replace `resetMockFs()` and `setupMockFs()` with `resetVirtualFs()`
- Replace individual file/directory mocks with a single `createVirtualFs()` call in beforeEach
- For each test case requiring specific filesystem state, add targeted createVirtualFs calls

### 3. Error Simulation
- Replace direct error mocking with Jest spies on fs functions
- Use the existing `createFsError()` function from virtualFsUtils
- Add proper spy cleanup (mockRestore) after each test to prevent test pollution

### 4. OS-Specific Tests
- Maintain the platform mocking approach for Windows and macOS tests
- Use spies on fs methods combined with createFsError to simulate platform-specific errors

## Reasoning
The virtualFsUtils approach was selected because:

1. It provides a more reliable testing environment by using an actual in-memory filesystem
2. It reduces the amount of boilerplate setup code needed for tests
3. It better simulates real filesystem behavior
4. It follows the recommended migration path documented in the test utilities README
5. It helps prevent Jest worker crashes that have been occurring with the mockFsUtils approach

This approach maintains the same test coverage and behavior verification while improving test stability and readability. The existing test scenarios (file reading, writing, error handling) will be preserved, just implemented with the new pattern.