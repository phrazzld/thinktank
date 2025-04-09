# Jest Testing Setup

This directory contains centralized Jest configuration and setup files that standardize mocking and testing across the project.

## Overview

We use a centralized approach to Jest configuration and mocking to:
- Reduce code duplication across test files
- Ensure consistent mock behavior
- Simplify the process of writing new tests
- Provide standardized mock utilities for common operations
- Align with our [testing philosophy](../TESTING_PHILOSOPHY.md) of minimal mocking and focusing on behavior

## Standard Testing Approach

Our preferred testing approach is centered around:

1. **Interface Mocking**: Mock external boundaries (FileSystem, ConsoleLogger, LLMClient) through interfaces rather than implementation details
2. **Virtual Filesystem**: Use `memfs` for realistic filesystem testing without touching the real filesystem
3. **Standard Helpers**: Use the helpers in `test/setup/` for consistent test setup and teardown
4. **Data Factories**: Use factory functions in `test/factories/` to create test data with sensible defaults

For comprehensive documentation of these helpers and factories, see:
- **[Test Setup Helpers Documentation](../test/setup/README.md)** - Primary reference for all test setup helpers
- **[Test Data Factories Documentation](../test/factories/README.md)** - Reference for test data creation

### Quick Start Example

Here's a minimal example of the recommended testing approach:

```typescript
import { setupTestHooks, createMockFileSystem, createMockConsoleLogger } from '../../test/setup';
import { createRunOptions } from '../../test/factories';

describe('My Feature', () => {
  // Set up standard test hooks (resets virtual FS, clears mocks, etc.)
  setupTestHooks();
  
  it('should process files correctly', async () => {
    // Create test data using factories
    const options = createRunOptions({
      contextFiles: ['/path/to/file.txt']
    });
    
    // Create interface mocks
    const mockFs = createMockFileSystem();
    const mockLogger = createMockConsoleLogger();
    
    // Set up virtual filesystem
    mockFs.readFileContent.mockResolvedValue('File content');
    mockFs.fileExists.mockResolvedValue(true);
    
    // Import and run the function under test
    const { processFiles } = await import('../../src/utils/fileProcessor');
    const result = await processFiles(options, mockFs, mockLogger);
    
    // Assert on behavior, not implementation details
    expect(result).toContain('Processed content');
    expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Processing'));
  });
});
```

## Deprecated Patterns

The following approaches are deprecated and should not be used in new tests:

1. **❌ Manual mocks in `src/utils/__mocks__/`**
   - These mocks are tightly coupled to implementation details and make tests brittle
   - Instead, use the interface mocking approach in `test/setup/`

2. **❌ Factory functions in `src/__tests__/utils/mockFactories.ts`**
   - These older factories have been replaced by improved versions in `test/factories/`
   - The new factories provide better TypeScript support and more consistent defaults

3. **❌ Direct Jest mocking of individual functions**
   - Mocking individual functions makes tests brittle and prone to breaking when implementation changes
   - Instead, mock at the interface level using the `createMock*` functions in `test/setup/`

For existing tests using these patterns, consider migrating them to the new approach as part of normal maintenance and refactoring.

## Jest Configuration Structure

- `jest/` - Root directory for Jest configuration
  - `setup.js` - Global setup file that runs once before all tests
  - `setupFilesAfterEnv.js` - Setup file that runs before each test file
  - `setupFiles/` - Directory containing core mock configurations:
    - `fs.js` - Mock setup for filesystem operations
    - `gitignore.js` - Mock setup for gitignore utilities
    - `testHelpers.js` - Common test helper functions
  - `examples/` - Example test files showing how to use the standardized setup

## Migration Guide

If you're migrating from older testing patterns, see the [Migration Guide](../test/setup/README.md#migration-from-legacy-patterns) in the Test Setup Helpers documentation.