# Ensure proper cleanup of mocks in tests

## Goal
Review all tests that use global mocks (e.g., process.cwd) and ensure they properly restore the original state using finally blocks or afterEach hooks.

## Implementation Approaches

### Approach 1: Module-by-Module Review and Fix
Systematically review every test file in the codebase, identify tests with global mocks, and add appropriate cleanup code to each file individually.

#### Pros:
- Comprehensive coverage of all test files
- Each file can be fixed according to its specific needs
- Files can be fixed incrementally

#### Cons:
- Time-consuming to review every test file
- May lead to inconsistent cleanup patterns across the codebase
- Requires careful attention to detail for each file

### Approach 2: Standardized Jest Setup/Teardown File
Create a centralized Jest setup/teardown mechanism that automatically tracks and restores global mocks.

#### Pros:
- Centralized solution that enforces consistent cleanup
- Reduces the chance of missed cleanup in individual files
- More maintainable long-term

#### Cons:
- More complex to implement initially
- May not handle all specialized mock scenarios
- Could interfere with tests that intentionally modify globals between tests

### Approach 3: Hybrid Approach with Automated Detection
Fix the most critical global mocks (like process.cwd) in individual files, while also implementing a light automated detection system to warn about potential mock leakage.

#### Pros:
- Addresses the most critical issues directly
- Provides a safety net for future development
- Balances immediate fixes with long-term prevention

#### Cons:
- Requires both specific fixes and a general solution
- Automated detection could produce false positives
- Slightly higher implementation complexity

## Selected Approach
I've chosen **Approach 1: Module-by-Module Review and Fix** because:

1. This approach is the most straightforward to implement given the current structure of the codebase
2. The existing tests have a mix of different mocking approaches that would be difficult to standardize automatically
3. This task is specifically about ensuring proper cleanup in existing tests, not establishing a new testing pattern
4. The codebase already has many examples of proper mock cleanup that can be used as patterns
5. This approach allows for incremental improvements, fixing the most critical issues first

## Implementation Plan
1. Identify all test files with global mocks that lack proper cleanup
2. Prioritize files with mocks that could affect other tests (process.cwd, process.env, console methods)
3. Add appropriate cleanup using:
   - afterEach() hooks for test-suite-wide mocks
   - try/finally blocks for test-specific mocks
   - jest.spyOn().mockRestore() for function spies
4. Ensure any module-level jest.mock() calls have corresponding cleanup in afterAll() hooks
5. Verify that all test files follow consistent patterns for similar mocks