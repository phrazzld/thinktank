# Implement core filesystem mock setup

## Goal
Implement the core setup and reset functions for mocking the Node.js fs/promises module in tests. These functions will provide a standardized way to configure filesystem mocks across all test files, making them more maintainable and consistent.

## Implementation Approach
Extend the existing `mockFsUtils.ts` file to implement the following core functions:

1. `setupMockFs(config?: FsMockConfig)`: A function that configures the mocked fs module with default behaviors based on the provided configuration or sensible defaults.
2. `resetMockFs()`: A function that resets all mocks to their initial state to prevent test pollution.

The implementation will use Jest's mocking capabilities to provide consistent mock behaviors for all file system operations.

### Alternatives Considered:

1. **Comprehensive Default Setup**: Implement setup functions that configure default behaviors for all fs operations (access, readFile, writeFile, stat, readdir, mkdir, etc.) at once with sensible defaults. This provides a "batteries included" approach where tests get common mocking behaviors without additional configuration.

2. **Minimal Setup + Per-Operation Config**: Implement a minimal setup function that only resets mocks without configuring default behaviors, requiring tests to explicitly configure each operation they need. This provides more explicit control at the cost of more verbose test setup.

3. **Global Fixture Approach**: Implement a Jest test fixture that automatically sets up and tears down mocks for all tests, reducing the need for explicit setup calls in each test file. This provides convenience but reduces flexibility and explicitness.

### Reasoning for Selected Approach:

I've chosen the **Comprehensive Default Setup** approach for these reasons:

1. **Test Simplicity**: By providing sensible defaults, tests can be more concise since they only need to override specific behaviors rather than configuring everything from scratch. This aligns with the goal of reducing duplication across test files.

2. **Flexibility**: The approach still allows for complete customization through the config parameter, so tests can override any default behavior when needed. This provides a good balance between convenience and control.

3. **Explicit Control**: Unlike a global fixture, this approach keeps setup explicit in test files, making it clear when mocks are being configured. This improves test readability and maintainability.

4. **Error Prevention**: By setting up defaults for all operations, we reduce the chance of unexpected behaviors when tests interact with operations they didn't explicitly mock. This helps catch test issues earlier.

The implementation will provide default behaviors that match common test patterns in the existing codebase (e.g., files exist by default, directories have basic contents) while allowing granular overrides for specific test scenarios.