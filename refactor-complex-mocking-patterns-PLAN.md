# Refactor complex mocking patterns

## Task Goal
The task aims to simplify and improve test maintainability by replacing complex mocking patterns (like `Object.defineProperty` and tests that mock the functions they're testing) with cleaner approaches that leverage the virtual filesystem setup we've already implemented.

## Chosen Implementation Approach: Direct Virtual Filesystem Replacement

Our approach will be to directly replace complex mocking patterns with virtual filesystem state setup and assertions. This will involve:

1. Using `createVirtualFs` or our newer setup helpers (`setupBasicFiles`, `setupProjectStructure`, etc.) to establish the initial filesystem state for each test.
2. Running the actual function under test without complex mocks.
3. Asserting on the output filesystem state using `getVirtualFs()` or mocked fs modules.
4. For error conditions, using targeted `jest.spyOn` for specific test cases rather than persistent complex mocks.

### Step-by-Step Process:
1. Identify test files that currently use `Object.defineProperty` or mock the functions they're testing.
2. For each identified file:
   - Replace setup code with our reusable helpers to configure the initial filesystem state.
   - Remove `Object.defineProperty` usages, replacing them with virtual filesystem setup.
   - For error-case testing, use limited-scope `jest.spyOn` instead of global overrides.
   - Update assertions to check the actual filesystem state or function output.
3. For commonly repeated patterns, extract additional reusable helpers if needed.

## Reasoning for Selection
This approach was selected for several compelling reasons:

1. **Direct Alignment with AC 3.4**: It specifically addresses the requirement to replace complex mocking patterns with the virtual filesystem approach.

2. **Builds on Existing Work**: It leverages the recently created reusable filesystem setup helpers and the virtual filesystem infrastructure.

3. **Improves Test Quality**: Shifting tests from verifying mock interactions to verifying actual behavior creates more meaningful tests that are closer to real-world usage.

4. **Simplicity and Readability**: Tests become more readable and focused on asserting behavior rather than complex mock setup.

5. **Consistency**: All models recommended this approach as the primary solution (Google Gemini, Qwen), noting it aligns with the project's established patterns.

The main alternative was a mixed approach that retained some complex mocking, but this would introduce inconsistency and limit the long-term maintainability benefits. A more dramatic approach involving dependency injection was considered but deemed outside the scope of this specific task.