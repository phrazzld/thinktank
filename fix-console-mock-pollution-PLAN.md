# Fix Console Mock Pollution

## Goal
Update all test files that mock `console` methods to use `jest.spyOn` and properly restore the mocks in `afterEach` blocks, consistent with our Jest spy cleanup standards.

## Approach
Based on my analysis of the codebase, I'll follow these steps:

1. Identify all test files that mock console methods by:
   - Finding files that use direct assignment to mock console methods (`console.log = jest.fn()`)
   - Checking files that use `jest.spyOn` on console methods but may not properly restore them

2. Update each identified file by:
   - Converting direct assignments to `jest.spyOn` if needed
   - Adding `jest.restoreAllMocks()` in `afterEach` blocks
   - Ensuring proper cleanup even if tests fail

3. Verify the solution by:
   - Running tests to ensure they still pass
   - Checking that console methods are properly restored between test cases

## Implementation Steps

1. **Replace Direct Assignments**: For files that use direct assignment to mock console methods, replace with `jest.spyOn`:
   ```typescript
   // BEFORE:
   console.log = jest.fn();
   
   // AFTER:
   jest.spyOn(console, 'log').mockImplementation(() => {});
   ```

2. **Add Restoration**: Ensure all files with console mocks include `jest.restoreAllMocks()` in their `afterEach` blocks:
   ```typescript
   afterEach(() => {
     jest.restoreAllMocks();
   });
   ```

3. **Remove Redundant Cleanup**: For files that already restore console methods manually, assess whether to:
   - Keep both approaches (for safety/redundancy)
   - Replace with standardized `jest.restoreAllMocks()`

4. **Address Edge Cases**: Some files may have complex setup requiring special handling, such as:
   - Files capturing console output for assertions
   - Files that need to maintain mocks across multiple tests
   - Files with nested test blocks

## Reasoning

Using `jest.spyOn` with proper restoration provides several benefits:

1. **Test Isolation**: Ensures that console mocks don't persist between tests, preventing test pollution
2. **Consistent Approach**: Follows the same pattern we've established for other Jest spies
3. **Safer Testing**: Automatic restoration makes tests more reliable even if they fail
4. **Less Boilerplate**: Reduces manual cleanup code when using `jest.restoreAllMocks()`

The approach builds on our recent work to fix unrestored Jest spies, extending the same standards to console mocks specifically.