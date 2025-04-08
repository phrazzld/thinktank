# Replace jest.spyOn with memfs helpers

## Task Goal
Refactor tests to use memfs helpers like createVirtualFs and createFsError instead of jest.spyOn for filesystem operations, improving test consistency and reducing direct mocking.

## Chosen Approach
After analyzing the thinktank suggestions and reviewing the project's testing philosophy, I have chosen the **Pragmatic Hybrid Approach** which uses virtual filesystem state for happy paths and targeted spies only for specific error conditions that are difficult to simulate through state alone.

### Justification for Chosen Approach

This approach provides the best balance between:

1. **Minimizing Mocking**: By replacing most jest.spyOn calls with virtual filesystem state setup, we adhere to the project's principle of minimizing mocks.

2. **Testing Behavior Over Implementation**: Using the virtual filesystem allows tests to verify behavior based on filesystem state rather than implementation details of how FS functions are called.

3. **Completeness of Error Testing**: While we minimize spies, we recognize that some specific error conditions (like permission errors) are difficult to simulate reliably through memfs state alone, so we retain targeted spies only for these cases.

4. **Consistency with Project Patterns**: The project already has sophisticated virtual filesystem utilities that are underutilized in some tests. This approach leverages these existing patterns more fully.

5. **Test Readability**: By clearly separating success cases (using filesystem state) and error cases (using targeted spies with createFsError), tests become more readable and purpose-focused.

### Principles for Implementation

1. **For Happy Paths**: Always use createVirtualFs or higher-level helpers like setupBasicFiles to set up filesystem state instead of spying on fs methods.

2. **For ENOENT Errors**: Prefer not creating the file/directory in the virtual filesystem instead of using spies to simulate "not found" errors.

3. **For Permission/Special Errors**: Use targeted jest.spyOn combined with createFsError helper for errors that cannot be easily simulated via state (EACCES, EPERM, etc.).

4. **For Test Clarity**: Ensure test setup clearly indicates whether it's testing a success path or error condition, with appropriate comments.

## Implementation Steps

1. **Files to Target**:
   - Focus on `src/utils/__tests__/readDirectoryContents.test.ts` (priority as it has many skipped tests)
   - `src/utils/__tests__/fileReader.test.ts` 
   - Any other test files using filesystem-related jest.spyOn calls

2. **For Each Test File**:
   - Replace jest.spyOn calls for successful filesystem operations with virtual filesystem setup
   - Convert ENOENT error simulation to simple absence of files in virtual filesystem
   - Retain only necessary spies for specific error conditions using createFsError
   - Unskip tests that can now be implemented with memfs helpers
   - Ensure all beforeEach/afterEach hooks properly reset the virtual filesystem

3. **Test Coverage Verification**:
   - Run tests after each file is refactored to ensure functionality is preserved
   - Verify that previously skipped tests now pass with the new approach

This implementation strategy ensures we maximize the use of the virtual filesystem while maintaining comprehensive test coverage, aligning with the project's testing philosophy of minimizing mocks while ensuring tests are reliable and maintainable.