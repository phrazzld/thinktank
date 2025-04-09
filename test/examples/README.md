# Test Examples

This directory contains example tests that demonstrate how to use the testing utilities and patterns recommended for the thinktank project.

## Available Examples

- **io-mocks-example.test.ts**: Demonstrates how to use the mock implementations for `FileSystem`, `ConsoleLogger`, and `UISpinner` interfaces in tests.

## Best Practices Demonstrated

These examples follow the testing philosophy described in `TESTING_PHILOSOPHY.md`, including:

1. **Minimizing Mocking**: Only mock external boundaries (File System, Console, UI).
2. **Behavior Testing**: Test behavior through public interfaces, not implementation details.
3. **Simple Test Setup**: Use standard setup helpers to reduce boilerplate.
4. **Clear Assertions**: Focus on verifying the behavior that matters.

## Using the Examples

To run the examples:

```bash
npm test -- test/examples/io-mocks-example.test.ts
```

Use these examples as reference when writing tests for your components that use the same interfaces.
