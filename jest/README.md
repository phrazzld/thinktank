# Jest Testing Setup

This directory contains centralized Jest configuration and setup files that standardize mocking and testing across the project.

## Overview

We use a centralized approach to Jest configuration to:
- Reduce code duplication across test files
- Ensure consistent mock behavior
- Simplify the process of writing new tests
- Provide standardized mock utilities for common operations

## Structure

- `setup.js` - Global setup file that runs once before all tests
- `setupFilesAfterEnv.js` - Setup file that runs before each test file
- `setupFiles/` - Directory containing specific mock configurations:
  - `fs.js` - Mock setup for filesystem operations
  - `gitignore.js` - Mock setup for gitignore utilities
  - `testHelpers.js` - Common test helper functions

## Usage Examples

### Basic Test File with FS Mocks

```typescript
// Import the test utilities you need
import { resetVirtualFs, createVirtualFs } from '../../__tests__/utils/virtualFsUtils';

// No need to mock fs/promises or fs - it's already done in the global setup

describe('Your Test Suite', () => {
  beforeEach(() => {
    // Reset the virtual filesystem
    resetVirtualFs();
    
    // Create test files and directories
    createVirtualFs({
      '/path/to/file.txt': 'File content',
      '/path/to/dir/nested.txt': 'Nested content'
    });
  });

  it('should read file content', async () => {
    // Import fs after the mocks are set up
    const fs = await import('fs/promises');
    const content = await fs.readFile('/path/to/file.txt', 'utf-8');
    expect(content).toBe('File content');
  });
});
```

### Using the Helper Functions

```typescript
// Import the helper functions from the setup files
import { setupBasicFs, createFsError } from '../../../jest/setupFiles/fs';
import { setupBasicGitignore, addGitignoreFile } from '../../../jest/setupFiles/gitignore';

describe('Your Test Suite', () => {
  beforeEach(() => {
    // Set up a basic filesystem
    setupBasicFs({
      '/path/to/file.txt': 'File content',
      '/path/to/dir/nested.txt': 'Nested content'
    });
    
    // Set up basic gitignore filtering
    setupBasicGitignore();
    
    // Add a .gitignore file
    addGitignoreFile('/path/to/.gitignore', '*.log\nnode_modules/');
  });

  it('should handle filesystem errors', async () => {
    // Create an fs error
    const error = createFsError('ENOENT', 'File not found', 'open', '/path/to/missing.txt');
    
    // Test error handling
    // ...
  });
});
```

## Common Test Patterns

### Testing Input/Output Operations

```typescript
// Import the setup helpers
import { setupBasicFs } from '../../../jest/setupFiles/fs';
import { resetVirtualFs } from '../../__tests__/utils/virtualFsUtils';

describe('File Processing', () => {
  beforeEach(() => {
    resetVirtualFs();
    setupBasicFs({
      '/input/file.txt': 'Input content',
      '/input/file2.txt': 'More input'
    });
  });

  it('should process files and write output', async () => {
    // Call your function that processes files
    await processFiles('/input', '/output');
    
    // Verify the output using the fs module (which is already mocked)
    const fs = await import('fs/promises');
    
    // Check if output files were created
    const outputExists = await fs.access('/output/file.txt')
      .then(() => true)
      .catch(() => false);
    
    expect(outputExists).toBe(true);
    
    // Check file content
    const content = await fs.readFile('/output/file.txt', 'utf-8');
    expect(content).toContain('Processed content');
  });
});
```

## Adding New Mocks

If you need to add new mock configurations:

1. Create a new file in the `setupFiles/` directory
2. Export the necessary mock functions and utilities
3. Import and use the file in your tests or add it to the global `setup.js` if needed