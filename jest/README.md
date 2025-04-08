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

- `jest/` - Root directory for Jest configuration
  - `setup.js` - Global setup file that runs once before all tests
  - `setupFilesAfterEnv.js` - Setup file that runs before each test file
  - `setupFiles/` - Directory containing core mock configurations:
    - `fs.js` - Mock setup for filesystem operations
    - `gitignore.js` - Mock setup for gitignore utilities
    - `testHelpers.js` - Common test helper functions
  - `examples/` - Example test files showing how to use the standardized setup
- `test/` - Root directory for test helpers and utilities
  - `setup/` - Directory containing domain-specific test setup helpers:
    - `common.ts` - Common setup utilities for all tests
    - `fs.ts` - File system specific setup helpers
    - `gitignore.ts` - Gitignore specific setup helpers
    - `config.ts` - Configuration testing helpers
    - `cli.ts` - CLI testing helpers
    - `providers.ts` - Provider/API testing helpers
    - `index.ts` - Re-exports all setup helpers

## Standard Setup Helpers

### Common Test Utilities (`test/setup/common.ts`)

```typescript
import { setupTestHooks } from '../../../test/setup/common';

describe('My Test Suite', () => {
  // Sets up standard beforeEach and afterEach hooks:
  // - Resets virtual filesystem
  // - Clears gitignore cache
  // - Resets Jest mocks
  // - Restores Jest mocks after each test
  setupTestHooks();
  
  it('should do something', () => {
    // Test with a clean environment
  });
});
```

### Filesystem Testing (`test/setup/fs.ts`)

```typescript
import { setupBasicFs, setupProjectStructure } from '../../../test/setup/fs';

describe('Filesystem Testing', () => {
  beforeEach(() => {
    // Set up a virtual filesystem with test files
    setupBasicFs({
      '/path/to/file.txt': 'File content',
      '/path/to/dir/nested.txt': 'Nested content'
    });
    
    // Or use a more structured approach
    setupProjectStructure('/project', {
      'src/index.ts': 'console.log("Hello");',
      'README.md': '# Project'
    });
  });

  it('should read file content', async () => {
    const fs = await import('fs/promises');
    const content = await fs.readFile('/path/to/file.txt', 'utf-8');
    expect(content).toBe('File content');
  });
});
```

### Gitignore Testing (`test/setup/gitignore.ts`)

```typescript
import { setupTestHooks } from '../../../test/setup/common';
import { setupWithGitignore, createIgnoreChecker } from '../../../test/setup/gitignore';

describe('Gitignore Testing', () => {
  setupTestHooks(); // Sets up standard hooks
  
  it('should properly ignore specified patterns', async () => {
    // Set up a project with files and a gitignore file
    await setupWithGitignore('/project', '*.log\n/build/', {
      'src/index.js': 'console.log("Hello");',
      'app.log': 'This should be ignored'
    });
    
    // Create a helper to check if paths should be ignored
    const shouldIgnore = createIgnoreChecker('/project');
    
    // Test gitignore behavior
    expect(await shouldIgnore('app.log')).toBe(true);
    expect(await shouldIgnore('src/index.js')).toBe(false);
  });
});
```

### Configuration Testing (`test/setup/config.ts`)

```typescript
import { setupTestHooks } from '../../../test/setup/common';
import { setupConfigTest, createMinimalConfig } from '../../../test/setup/config';

describe('Config Testing', () => {
  setupTestHooks(); // Sets up standard hooks
  
  it('should load configuration from file', async () => {
    // Set up a test environment with a configuration file
    const { configPath } = setupConfigTest('/project');
    
    // Import the configManager
    const { loadConfig } = await import('../../src/core/configManager');
    
    // Test configuration loading
    const config = await loadConfig(configPath);
    expect(config).toEqual(createMinimalConfig());
  });
});
```

### CLI Testing (`test/setup/cli.ts`)

```typescript
import { setupTestHooks } from '../../../test/setup/common';
import { setupCliTest, mockCliArguments, mockConsoleOutput } from '../../../test/setup/cli';

describe('CLI Testing', () => {
  setupTestHooks(); // Sets up standard hooks
  
  it('should run CLI command', async () => {
    // Set up a CLI test environment
    const { promptFile } = setupCliTest('/test');
    
    // Mock CLI arguments and console output
    const restoreArgs = mockCliArguments('run', [promptFile]);
    const { mockLog, restore: restoreConsole } = mockConsoleOutput();
    
    try {
      // Import and run the CLI
      const { run } = await import('../../src/cli');
      await run();
      
      // Test CLI behavior
      expect(mockLog).toHaveBeenCalledWith(expect.stringContaining('Success'));
    } finally {
      // Clean up
      restoreArgs();
      restoreConsole();
    }
  });
});
```

### Provider Testing (`test/setup/providers.ts`)

```typescript
import { setupTestHooks } from '../../../test/setup/common';
import { setupProviderMock, createMockLlmResponse } from '../../../test/setup/providers';

describe('Provider Testing', () => {
  setupTestHooks(); // Sets up standard hooks
  
  it('should generate text with provider', async () => {
    // Set up provider mock
    const { mockFetch } = setupProviderMock('openai', 'Generated text');
    global.fetch = mockFetch;
    
    // Import and test provider
    const { OpenAIProvider } = await import('../../src/providers/openai');
    const provider = new OpenAIProvider();
    
    // Test provider behavior
    const response = await provider.generate({ prompt: 'Hello' });
    expect(response.response).toBe('Generated text');
  });
});
```

## Migration Guide

### Using the New Domain-Specific Setup Helpers

If you're currently using the centralized `jest/setupFiles/` helpers directly, consider migrating to the domain-specific helpers in `test/setup/`:

Before:
```typescript
import { setupBasicFs } from '../../../jest/setupFiles/fs';
import { clearGitignoreCache } from '../../../jest/setupFiles/gitignore';

describe('My Test', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    setupBasicFs({
      '/test.txt': 'content'
    });
    clearGitignoreCache();
  });
  
  // Tests...
});
```

After:
```typescript
import { setupTestHooks } from '../../../test/setup/common';
import { setupBasicFs } from '../../../test/setup/fs';

describe('My Test', () => {
  // Sets up all standard hooks (resetVirtualFs, clearIgnoreCache, jest.clearAllMocks)
  setupTestHooks();
  
  beforeEach(() => {
    // Only need to add specific setup beyond the standard hooks
    setupBasicFs({
      '/test.txt': 'content'
    });
  });
  
  // Tests...
});
```

### Converting from Manual Mocks

If you're currently using manual mocks from `src/utils/__mocks__/`, convert to the domain-specific helpers:

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
// Using domain-specific setup
import { setupTestHooks } from '../../../test/setup/common';
import { setupFileReaderTest } from '../../../test/setup/fs';

describe('My Test', () => {
  setupTestHooks();
  
  beforeEach(() => {
    const { testFile } = setupFileReaderTest('/test');
  });
  
  // Later in test
  const { readContextFile } = await import('../../utils/fileReader');
  const result = await readContextFile('/test.txt');
});
```

## Best Practices

1. **Use `setupTestHooks()`** as the first line in your describe block to ensure consistent environment setup for all tests:
   ```typescript
   import { setupTestHooks } from '../../../test/setup/common';
   
   describe('My Test Suite', () => {
     setupTestHooks();
     // Your tests...
   });
   ```

2. **Prefer domain-specific helpers** over general utilities when available:
   ```typescript
   // Good - uses domain-specific helper
   import { setupConfigTest } from '../../../test/setup/config';
   const { configPath } = setupConfigTest();
   
   // Not recommended - manually creating config file
   import { setupBasicFs } from '../../../test/setup/fs';
   setupBasicFs({
     '/project/config.json': JSON.stringify({ models: [] })
   });
   ```

3. **Use the helper module's re-exports** to simplify imports:
   ```typescript
   // Good - single import for multiple helpers
   import { setupTestHooks, setupBasicFs, createMockLlmResponse } from '../../../test/setup';
   
   // Not recommended - separate imports
   import { setupTestHooks } from '../../../test/setup/common';
   import { setupBasicFs } from '../../../test/setup/fs';
   import { createMockLlmResponse } from '../../../test/setup/providers';
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

For complete examples, see:
- `jest/examples/centralized-mock-example.test.ts` - Basic usage of the centralized mock setup
- Look at the refactored tests in `src/utils/__tests__/gitignoreFilterIntegration.test.ts` for examples of using the standard approach