# Test Setup Helpers

This directory contains standardized helpers for setting up test environments. These utilities make tests more readable, maintainable, and aligned with our [testing philosophy](../../TESTING_PHILOSOPHY.md) by reducing boilerplate and focusing on testing behavior rather than implementation details.

## Quick Start

Most tests should use the `setupTestHooks()` function to establish a clean test environment:

```typescript
import { setupTestHooks } from '../../test/setup';

describe('My Feature', () => {
  setupTestHooks(); // Sets up standard beforeEach/afterEach hooks
  
  it('should handle the happy path', async () => {
    // Your test here...
  });
});
```

## Helper Modules

### Common Utilities (`common.ts`)

Core utilities used by most tests:

- **`setupTestHooks()`**: Sets up standard `beforeEach` and `afterEach` hooks to reset the virtual filesystem, clear gitignore cache, and reset/restore Jest mocks
- **`mockEnv()`**: Mocks environment variables and returns a function to restore the original environment
- **`createTestId()`**: Creates a random string for test fixtures
- **`wait()`**: Awaitable delay function for testing asynchronous behavior

```typescript
import { setupTestHooks, mockEnv } from '../../test/setup';

describe('Environment Testing', () => {
  setupTestHooks();
  
  it('should respect environment variables', async () => {
    const restore = mockEnv({ NODE_ENV: 'testing', API_KEY: 'test-key' });
    
    try {
      // Test code that uses process.env
      expect(process.env.NODE_ENV).toBe('testing');
    } finally {
      restore(); // Always restore the original environment
    }
  });
});
```

### Filesystem Testing (`fs.ts`)

Utilities for testing filesystem operations using an in-memory filesystem:

- **`setupBasicFs()`**: Creates a structure of files/directories in the virtual filesystem
- **`setupProjectStructure()`**: Creates a project directory with files and standard structure
- **`createMockFileSystem()`**: Creates a Jest mock implementation of the `FileSystem` interface
- **`setupWithGitignore()`**: Creates a project structure with a gitignore file

```typescript
import { setupTestHooks, setupBasicFs } from '../../test/setup';
import { readFileContent } from '../../src/utils/fileReader';

describe('File Reading', () => {
  setupTestHooks();
  
  beforeEach(() => {
    setupBasicFs({
      '/path/to/file.txt': 'File content',
      '/config/settings.json': '{"key": "value"}',
      '/empty-dir/': '' // Creates an empty directory
    });
  });
  
  it('should read file content', async () => {
    const result = await readFileContent('/path/to/file.txt');
    expect(result).toBe('File content');
  });
});
```

Interface mocking example:

```typescript
import { setupTestHooks, createMockFileSystem } from '../../test/setup';

describe('Using FileSystem interface mock', () => {
  setupTestHooks();
  
  it('should use a mocked FileSystem', async () => {
    // Create a mock FileSystem implementation
    const mockFs = createMockFileSystem({
      readFileContent: jest.fn().mockResolvedValue('Mocked content'),
      fileExists: jest.fn().mockResolvedValue(true)
    });
    
    // Function under test
    const result = await myFunctionThatUsesFileSystem(mockFs);
    
    // Assertions
    expect(mockFs.readFileContent).toHaveBeenCalledWith('/expected/path.txt');
    expect(result).toContain('Mocked content');
  });
});
```

### I/O Testing (`io.ts`)

Utilities for testing console output and UI spinners:

- **`createMockConsoleLogger()`**: Creates a Jest mock implementation of the `ConsoleLogger` interface
- **`createMockUISpinner()`**: Creates a Jest mock implementation of the `UISpinner` interface

```typescript
import { setupTestHooks, createMockConsoleLogger } from '../../test/setup';

describe('Logging Function', () => {
  setupTestHooks();
  
  it('should log messages at different levels', async () => {
    // Create mock logger
    const mockLogger = createMockConsoleLogger();
    
    // Function under test
    const result = await myFunctionThatLogs(mockLogger);
    
    // Assertions
    expect(mockLogger.info).toHaveBeenCalledWith('Starting process');
    expect(mockLogger.error).toHaveBeenCalledWith(expect.stringContaining('error'));
  });
});
```

### Configuration Testing (`config.ts`)

Utilities for testing configuration-related functionality:

- **`setupConfigTest()`**: Sets up a test environment with a configuration file
- **`createMockConfigManager()`**: Creates a Jest mock implementation of the `ConfigManagerInterface`

```typescript
import { setupTestHooks, setupConfigTest } from '../../test/setup';

describe('Config Loading', () => {
  setupTestHooks();
  
  it('should load config from file', async () => {
    // Setup a test config environment
    const { configPath } = setupConfigTest('/project', {
      models: [{ provider: 'openai', modelId: 'gpt-4' }]
    });
    
    // Import the module under test
    const { loadConfig } = await import('../../src/core/configManager');
    
    // Test the function
    const config = await loadConfig(configPath);
    
    // Assertions
    expect(config.models[0].provider).toBe('openai');
  });
});
```

### CLI Testing (`cli.ts`)

Utilities for testing command-line interface functionality:

- **`setupCliTest()`**: Sets up a test environment for CLI testing
- **`mockCliArguments()`**: Mocks command-line arguments
- **`mockConsoleOutput()`**: Mocks console.log, error, and warn functions

```typescript
import { setupTestHooks, setupCliTest, mockCliArguments, mockConsoleOutput } from '../../test/setup';

describe('CLI Command', () => {
  setupTestHooks();
  
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

### Provider Testing (`providers.ts`)

Utilities for testing LLM providers and API integrations:

- **`setupProviderMock()`**: Sets up a mock for provider API responses
- **`createMockLlmClient()`**: Creates a Jest mock implementation of the `LLMClient` interface
- **`createMockLlmProvider()`**: Creates a Jest mock implementation of the `LLMProvider` interface

```typescript
import { setupTestHooks, createMockLlmClient } from '../../test/setup';
import { createLlmResponse } from '../../test/factories';

describe('LLM Integration', () => {
  setupTestHooks();
  
  it('should generate text using LLM', async () => {
    // Create mock LLM client
    const mockLlmClient = createMockLlmClient();
    
    // Configure the mock to return a specific response
    mockLlmClient.generate.mockResolvedValue(
      createLlmResponse({ response: 'Generated response' })
    );
    
    // Function under test
    const result = await myFunctionThatUsesLlm(mockLlmClient);
    
    // Assertions
    expect(mockLlmClient.generate).toHaveBeenCalledWith(
      expect.objectContaining({ prompt: expect.any(String) })
    );
    expect(result).toContain('Generated response');
  });
});
```

### Workflow Testing (`workflow.ts`)

Utilities for testing complex workflows and integration scenarios:

- **`setupWorkflowTestEnvironment()`**: Sets up a complete test environment for workflow testing
- **`setupMocksForSuccessfulRun()`**: Configures mocks for a successful workflow run
- **`setupMocksForApiError()`**: Configures mocks for a workflow with API errors
- **`setupMocksForFileSystemError()`**: Configures mocks for a workflow with filesystem errors

```typescript
import { setupTestHooks, setupWorkflowTestEnvironment } from '../../test/setup';

describe('Workflow Integration', () => {
  setupTestHooks();
  
  it('should complete a successful workflow', async () => {
    // Set up a complete test environment with all required mocks
    const {
      mockFileSystem,
      mockConfigManager,
      mockLlmClient,
      mockConsoleLogger,
      options
    } = setupWorkflowTestEnvironment();
    
    // Configure specific behavior for this test
    mockLlmClient.generate.mockResolvedValue({
      model: 'test-model',
      response: 'Generated content',
      error: null
    });
    
    // Import and run the workflow
    const { runThinktank } = await import('../../src/workflow/runThinktank');
    await runThinktank(
      options,
      mockFileSystem,
      mockConfigManager,
      mockLlmClient,
      mockConsoleLogger
    );
    
    // Assertions
    expect(mockFileSystem.writeFile).toHaveBeenCalled();
    expect(mockConsoleLogger.success).toHaveBeenCalled();
  });
});
```

### Gitignore Testing (`gitignore.ts`)

Utilities for testing gitignore pattern matching:

- **`setupWithGitignore()`**: Creates a project with files and a gitignore file
- **`createIgnoreChecker()`**: Creates a function to check if a path should be ignored

```typescript
import { setupTestHooks, setupWithGitignore, createIgnoreChecker } from '../../test/setup';

describe('Gitignore Filtering', () => {
  setupTestHooks();
  
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

## Using Data Factories

The `test/factories/` directory contains functions for creating test data objects. These factories provide defaults and allow customization through overrides:

```typescript
import { setupTestHooks } from '../../test/setup';
import { createAppConfig, createModelConfig, createLlmResponse } from '../../test/factories';

describe('Configuration Functions', () => {
  setupTestHooks();
  
  it('should process app configuration', () => {
    // Create test data with factories
    const appConfig = createAppConfig({
      models: [
        createModelConfig({ provider: 'openai', modelId: 'gpt-4o' }),
        createModelConfig({ provider: 'anthropic', modelId: 'claude-3-opus-20240229' })
      ]
    });
    
    // Test the function
    const result = processAppConfig(appConfig);
    
    // Assertions
    expect(result.availableModels.length).toBe(2);
  });
  
  it('should handle LLM responses', () => {
    // Create a successful response
    const successResponse = createLlmResponse({
      response: 'Success text',
      model: 'openai:gpt-4o',
      tokensUsed: { total: 100, prompt: 50, completion: 50 }
    });
    
    // Create an error response
    const errorResponse = createLlmResponse({
      error: new Error('API error'),
      model: 'openai:gpt-4o'
    });
    
    // Test the function
    expect(processLlmResponse(successResponse)).toBe('Success');
    expect(processLlmResponse(errorResponse)).toBe('Error');
  });
});
```

## Best Practices

1. **Always Use Standard Hooks**: Begin your describe blocks with `setupTestHooks()` to ensure a clean environment for each test.

2. **Test Behavior, Not Implementation**: Focus tests on the inputs and outputs of functions, not their internal implementation details.

3. **Minimize Mocking**: Only mock external boundaries (FileSystem, ConsoleLogger, LLMClient) and inject these mocks through interfaces, not by mocking internal functions.

4. **Use Interface Mocks**: Prefer the `createMock*` functions to create type-safe mocks for interfaces rather than mocking individual functions.

5. **Test the Happy Path First**: Prioritize testing the primary success scenario before adding tests for edge cases and errors.

6. **Use Data Factories**: Use the factory functions in `test/factories/` to create test data with sensible defaults.

7. **Co-locate Related Tests**: Keep tests close to the code they test to make them easier to find and maintain.

8. **Group Tests Logically**: Use describe blocks to group related tests and provide context.

9. **Keep Tests Simple**: If a test requires complex setup, consider if the code being tested should be refactored for better testability.

10. **Clean Up After Tests**: Ensure any global state (like environment variables) is properly restored after tests.

## Migration from Legacy Patterns

If you're currently using older patterns, here's how to migrate to the new approach:

### From Manual Mocks:

```typescript
// OLD APPROACH - Don't use this
jest.mock('../../utils/fileReader');
import { readFileContent } from '../../utils/fileReader';
readFileContent.mockResolvedValue('Mocked content');

// NEW APPROACH - Use this instead
import { createMockFileSystem } from '../../test/setup';
const mockFs = createMockFileSystem({
  readFileContent: jest.fn().mockResolvedValue('Mocked content')
});
myFunction(mockFs); // Inject the mock through the interface
```

### From src/__tests__/utils/mockFactories.ts:

```typescript
// OLD APPROACH - Don't use this
import { mockConsoleLogger } from '../../../__tests__/utils/mockFactories';
const mockLogger = mockConsoleLogger();

// NEW APPROACH - Use this instead
import { createMockConsoleLogger } from '../../test/setup';
const mockLogger = createMockConsoleLogger();
```

### From src/__tests__/utils/mockFsUtils.ts:

```typescript
// OLD APPROACH - Don't use this
import { setupMockFs, mockReadFile } from '../../../__tests__/utils/mockFsUtils';
setupMockFs();
mockReadFile('/file.txt', 'content');

// NEW APPROACH - Use this instead
import { setupBasicFs } from '../../test/setup';
setupBasicFs({
  '/file.txt': 'content'
});
```
