# Jest Testing Setup

This directory contains centralized Jest configuration and setup files that standardize mocking and testing across the project.

## Overview

We use a centralized approach to Jest configuration and mocking to:
- Reduce code duplication across test files
- Ensure consistent mock behavior
- Simplify the process of writing new tests
- Provide standardized mock utilities for common operations
- Align with our testing philosophy of minimal mocking and focusing on behavior

## The Standard Mocking Approach

**PREFERRED APPROACH:** Use the centralized mock setup from `jest/setupFiles/` for all new tests and when refactoring existing tests.

The codebase has historically used multiple approaches to mocking:
1. ~~Manual mocks in `src/utils/__mocks__/`~~ (Deprecated)
2. ~~Factory functions in `src/__tests__/utils/mockFactories.ts`~~ (Deprecated)
3. **Centralized setup helpers in `jest/setupFiles/`** (Preferred)

The centralized approach based on the virtual filesystem (memfs) is now the standard. This approach:
- Aligns better with our testing philosophy
- Provides more realistic behavior testing
- Reduces brittle implementation-specific mocks
- Simplifies test setup and maintenance

## Directory Structure

- `setup.js` - Global setup file that runs once before all tests
- `setupFilesAfterEnv.js` - Setup file that runs before each test file
- `setupFiles/` - Directory containing specific mock configurations:
  - `fs.js` - Mock setup for filesystem operations
  - `gitignore.js` - Mock setup for gitignore utilities
  - `testHelpers.js` - Common test helper functions
- `examples/` - Example test files showing how to use the standardized setup

## Standard Setup Helpers

### Filesystem Testing (`setupFiles/fs.js`)

```typescript
import { 
  setupBasicFs,
  resetFs,
  createFsError,
  getFs,
  createStats,
  createDirent,
  normalizePath
} from '../../../jest/setupFiles/fs';

describe('Filesystem Testing', () => {
  beforeEach(() => {
    // Set up a virtual filesystem with test files
    setupBasicFs({
      '/path/to/file.txt': 'File content',
      '/path/to/dir/nested.txt': 'Nested content'
    });
  });

  it('should read file content', async () => {
    const fs = await import('fs/promises');
    const content = await fs.readFile('/path/to/file.txt', 'utf-8');
    expect(content).toBe('File content');
  });
  
  it('should handle file errors', async () => {
    // Create a standard filesystem error
    const error = createFsError('ENOENT', 'File not found', 'open', '/missing.txt');
    
    // Use the error in your tests
    expect(error.code).toBe('ENOENT');
    expect(error.path).toBe('/missing.txt');
  });
});
```

### Gitignore Testing (`setupFiles/gitignore.js`)

```typescript
import { 
  setupBasicGitignore,
  addGitignoreFile,
  clearGitignoreCache,
  createGitignoreMock
} from '../../../jest/setupFiles/gitignore';

describe('Gitignore Testing', () => {
  beforeEach(() => {
    // Clear the cache between tests for isolation
    clearGitignoreCache();
    
    // Set up a basic gitignore environment
    setupBasicGitignore();
    
    // Or create a custom gitignore file
    addGitignoreFile('/project/custom/.gitignore', '*.log\n/build/');
  });

  it('should properly ignore specified patterns', async () => {
    // The actual gitignoreUtils module works with our virtual filesystem
    const { shouldIgnorePath } = await import('../../src/utils/gitignoreUtils');
    
    const result = await shouldIgnorePath('/project', '/project/node_modules/file.js');
    expect(result).toBe(true);
  });
});
```

### General Testing Helpers (`setupFiles/testHelpers.js`)

```typescript
import { 
  promisify,
  wait,
  createMockObject,
  createMockSpinner,
  createNetworkMock,
  createNetworkErrorMock,
  createLlmResponseMock
} from '../../../jest/setupFiles/testHelpers';

describe('Using Test Helpers', () => {
  it('should create mock objects', async () => {
    // Create a mock API object
    const mockApi = createMockObject({
      getData: () => ({ result: 'test' }),
      processData: (data) => data.toUpperCase()
    });
    
    // Use the mock in tests
    const result = mockApi.getData();
    expect(result).toEqual({ result: 'test' });
    expect(mockApi.getData).toHaveBeenCalled();
  });
  
  it('should create network mocks', async () => {
    // Mock a successful network response
    const fetchMock = createNetworkMock({ data: 'success' }, 200);
    global.fetch = fetchMock;
    
    // Use the mock in tests
    const response = await fetch('https://api.example.com');
    const data = await response.json();
    expect(data).toEqual({ data: 'success' });
  });
  
  it('should mock LLM responses', () => {
    // Create a mock LLM response
    const mockResponse = createLlmResponseMock({ text: 'Generated text' });
    
    // Use the mock in tests
    expect(mockResponse.providerId).toBe('mock-provider');
    expect(mockResponse.response).toBe('Generated text');
    expect(mockResponse.error).toBeNull();
  });
});
```

## Migration Guide

### Converting from Manual Mocks

If you're currently using manual mocks from `src/utils/__mocks__/`, convert to the centralized approach:

Before:
```typescript
// Using manual mocks
jest.mock('../../utils/fileReader');
import { readContextFile } from '../../utils/fileReader';

// Later in test
readContextFile.mockResolvedValueOnce({ 
  path: '/test.txt', 
  content: 'mocked content',
  error: null 
});
```

After:
```typescript
// Using centralized setup
import { setupBasicFs } from '../../../jest/setupFiles/fs';

beforeEach(() => {
  setupBasicFs({
    '/test.txt': 'mocked content'
  });
});

// Later in test
const { readContextFile } = await import('../../utils/fileReader');
const result = await readContextFile('/test.txt');
```

### Converting from mockFactories

If you're currently using `mockFactories.ts`, convert to the centralized approach:

Before:
```typescript
import { createFileReaderMocks } from '../mockFactories';

const mocks = createFileReaderMocks();
jest.spyOn(fileReader, 'readContextFile').mockImplementation(mocks.readContextFile);
```

After:
```typescript
import { setupBasicFs } from '../../../jest/setupFiles/fs';

beforeEach(() => {
  setupBasicFs({
    '/test.txt': 'mocked content'
  });
});
```

## Best Practices

1. **Clear the gitignore cache** in beforeEach to prevent test interdependencies:
   ```typescript
   import { clearGitignoreCache } from '../../../jest/setupFiles/gitignore';
   
   beforeEach(() => {
     clearGitignoreCache();
   });
   ```

2. **Reset the filesystem** between tests for isolation:
   ```typescript
   import { resetFs } from '../../../jest/setupFiles/fs';
   
   beforeEach(() => {
     resetFs();
   });
   ```

3. **Use the provided helpers** instead of creating your own utility functions:
   ```typescript
   // Good - uses standard helpers
   import { createNetworkMock } from '../../../jest/setupFiles/testHelpers';
   const fetchMock = createNetworkMock({ data: 'test' });
   
   // Not recommended - custom implementation may diverge from standards
   const customMock = jest.fn().mockResolvedValue({ 
     json: () => Promise.resolve({ data: 'test' }) 
   });
   ```

4. **Test real behavior** rather than implementation details:
   ```typescript
   // Good - tests behavior through the API
   const result = await readContextFile('/test.txt');
   expect(result.content).toBe('mocked content');
   
   // Not recommended - tests implementation details
   expect(readFileContent).toHaveBeenCalledWith('/test.txt');
   ```

5. **Minimize mocking** of internal components - use the virtual filesystem to test real code paths when possible.

## Example Tests

For complete examples, see the `examples/` directory:
- `centralized-mock-example.test.ts` - Basic usage of the centralized mock setup