# Test Examples

This directory contains example tests that demonstrate how to use the testing utilities and patterns recommended for the thinktank project.

## Available Examples

- **io-mocks-example.test.ts**: Demonstrates how to use the mock implementations for `FileSystem`, `ConsoleLogger`, and `UISpinner` interfaces in tests.
- **workflow-test-example.test.ts**: Demonstrates how to use the workflow test utilities, scenario helpers, and factories for testing the main application workflow.

## Best Practices Demonstrated

These examples follow the testing philosophy described in `TESTING_PHILOSOPHY.md`, including:

1. **Minimizing Mocking**: Only mock external boundaries (File System, Console, UI).
2. **Behavior Testing**: Test behavior through public interfaces, not implementation details.
3. **Simple Test Setup**: Use standard setup helpers to reduce boilerplate.
4. **Clear Assertions**: Focus on verifying the behavior that matters.
5. **Test Data Factories**: Use factories to create test data consistently.
6. **Scenario Helpers**: Use higher-level helpers to set up common test scenarios.

## Using the Examples

To run the examples:

```bash
# Run all examples
npm test -- test/examples/

# Run a specific example
npm test -- test/examples/io-mocks-example.test.ts
npm test -- test/examples/workflow-test-example.test.ts
```

Use these examples as reference when writing tests for your components that use the same interfaces and patterns.
