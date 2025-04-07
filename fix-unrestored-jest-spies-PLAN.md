# Fix Unrestored Jest Spies

## Goal
Fix tests that use `jest.spyOn` but don't properly restore the spies, which can lead to test interference and unreliable test results.

## Chosen Implementation Approach
Systematically identify test files that use Jest spies without proper restoration, then add `jest.restoreAllMocks()` in `afterEach` blocks or individual `mockRestore()` calls as appropriate.

### Implementation Steps:
1. Search the codebase for files containing `jest.spyOn` to identify tests using spies
2. For each file, check if they already have proper cleanup (`mockRestore`, `restoreAllMocks`)
3. For files without proper cleanup:
   - Add `jest.restoreAllMocks()` in the `afterEach` block if the file already has an `afterEach`
   - Create a new `afterEach` block with `jest.restoreAllMocks()` if no such block exists
   - For isolated spies with specific cleanup needs, use individual `mockRestore()` calls

## Reasoning
I chose this approach over alternatives for these reasons:

1. **Global vs. Individual Restoration:**
   - Using `jest.restoreAllMocks()` in `afterEach` is more maintainable than individual `mockRestore()` calls as it catches all spies automatically
   - However, for specific cases where precise control is needed, individual `mockRestore()` calls will be used

2. **Alternative Considered - Centralized Setup:**
   - Creating a centralized Jest setup file for all tests would be more comprehensive but introduces a larger change that should be its own task
   - The selected approach addresses the immediate issue while laying groundwork for later centralization (task "Create Centralized Mock Setup")

3. **Alternative Considered - Migrate to Jest Reset Functions:**
   - Replacing `mockRestore()` with `mockReset()` where full restoration isn't needed would be more performant but increases risk of subtle test behavior changes
   - The selected approach is safer by fully restoring the original implementation
