# End-to-End Testing Guide

This guide covers end-to-end (E2E) testing strategies for the Thinktank project, with a focus on testing the CLI as a complete package.

## E2E Testing Goals

End-to-end tests verify that the entire application works correctly as a cohesive unit by:

1. Testing the CLI interface as users would interact with it
2. Verifying output files are created correctly
3. Ensuring error handling works properly from the user's perspective
4. Testing integration with file I/O, configuration, and external APIs

## E2E Testing Approach

Thinktank's E2E tests:

- Use a temporary test directory for isolation
- Run the actual compiled CLI binary
- Test realistic workflows
- Verify standard output, error output, exit codes, and generated files

## Setting Up E2E Tests

E2E tests should use the dedicated `e2eTestUtils`:

```typescript
import { 
  setupTestDir, 
  cleanupTestDir, 
  runCliCommand,
  getCliPath
} from '../e2eTestUtils';

describe('CLI end-to-end tests', () => {
  // Create a unique test directory for this test suite
  const testDir = setupTestDir();
  
  // Clean up after tests complete
  afterAll(() => {
    cleanupTestDir(testDir);
  });
  
  it('should process a text file', async () => {
    // Create test input file
    const inputPath = path.join(testDir, 'prompt.txt');
    fs.writeFileSync(inputPath, 'How do I optimize a React app?');
    
    // Create test context file
    const contextPath = path.join(testDir, 'context.js');
    fs.writeFileSync(contextPath, 'function optimize() { /* ... */ }');
    
    // Run the CLI command
    const { stdout, stderr, exitCode } = await runCliCommand([
      'run',
      inputPath,
      contextPath,
      '--output', path.join(testDir, 'output')
    ]);
    
    // Verify success
    expect(exitCode).toBe(0);
    expect(stderr).toBe('');
    expect(stdout).toContain('Successfully completed');
    
    // Verify output files were created
    const outputFiles = fs.readdirSync(path.join(testDir, 'output'));
    expect(outputFiles.length).toBeGreaterThan(0);
    expect(outputFiles).toContain(expect.stringMatching(/.*\.md$/));
  });
});
```

## Testing Error Scenarios

Test how the CLI handles errors from a user perspective:

```typescript
it('should handle nonexistent input file', async () => {
  const { stdout, stderr, exitCode } = await runCliCommand([
    'run',
    'nonexistent-file.txt'
  ]);
  
  // Verify error handling
  expect(exitCode).not.toBe(0);
  expect(stderr).toContain('File not found');
  expect(stderr).toContain('Suggestions:');
});

it('should handle invalid configuration', async () => {
  // Create invalid config file
  const configPath = path.join(testDir, 'config.json');
  fs.writeFileSync(configPath, '{ "invalid": "config" }');
  
  const { stdout, stderr, exitCode } = await runCliCommand([
    'run',
    'prompt.txt',
    '--config', configPath
  ]);
  
  // Verify error handling
  expect(exitCode).not.toBe(0);
  expect(stderr).toContain('Invalid configuration');
});
```

## Testing Different Commands

Test each CLI command:

```typescript
it('should list available models', async () => {
  const { stdout, stderr, exitCode } = await runCliCommand([
    'models'
  ]);
  
  // Verify command output
  expect(exitCode).toBe(0);
  expect(stdout).toContain('Available models');
  expect(stdout).toContain('openai:');
  expect(stdout).toContain('anthropic:');
});

it('should show configuration', async () => {
  const { stdout, stderr, exitCode } = await runCliCommand([
    'config',
    'show'
  ]);
  
  // Verify command output
  expect(exitCode).toBe(0);
  expect(stdout).toContain('Current configuration');
});
```

## Testing Output Files

Verify the content of generated output files:

```typescript
it('should generate correct output files', async () => {
  // Setup and run command
  // ...
  
  // Check content of generated files
  const outputDir = path.join(testDir, 'output');
  const outputFiles = fs.readdirSync(outputDir);
  
  // Get content of first output file
  const firstOutputFile = path.join(outputDir, outputFiles[0]);
  const content = fs.readFileSync(firstOutputFile, 'utf-8');
  
  // Verify content format
  expect(content).toContain('# openai:gpt-4');
  expect(content).toContain('Generated:');
  expect(content).toContain('## Response');
});
```

## Testing with Mocked API Responses

For E2E tests, you may want to mock external API calls while still testing the rest of the system:

```typescript
it('should process with mocked API response', async () => {
  // Create a mock API server
  const mockServer = setupMockApiServer([
    {
      model: 'gpt-4',
      response: {
        text: 'This is a mock response',
        usage: {
          prompt_tokens: 50,
          completion_tokens: 20,
          total_tokens: 70
        }
      }
    }
  ]);
  
  // Set environment variables to point to mock server
  process.env.OPENAI_API_BASE = mockServer.url;
  
  // Run command
  const { stdout, stderr, exitCode } = await runCliCommand([
    'run',
    'prompt.txt',
    '--models', 'openai:gpt-4'
  ]);
  
  // Verify results
  expect(exitCode).toBe(0);
  expect(stdout).toContain('Successfully completed');
  
  // Clean up
  mockServer.close();
  delete process.env.OPENAI_API_BASE;
});
```

## Best Practices

1. **Isolation**: Each E2E test should work in isolation with its own test directory
2. **Complete cleanup**: Always clean up test directories, even if tests fail
3. **Test real CLI commands**: Use the actual command-line interface as users would
4. **Verify exit codes**: Always check process exit codes (0 for success, non-zero for errors)
5. **Check both stdout and stderr**: Verify output appears in the correct stream
6. **Test file content**: Don't just check that files exist; verify their content
7. **Test with realistic data**: Use realistic input data that resembles actual usage

## Troubleshooting E2E Tests

### Common Issues

1. **Test interference**: Tests affecting each other's files
   - Use unique test directories for each test suite
   - Clean up thoroughly in afterEach/afterAll hooks

2. **Timing issues**: Commands completing too quickly or timing out
   - Use async/await with appropriate timeout values
   - Verify that processes have fully completed

3. **Environment variables**: Tests affected by global environment
   - Reset any modified environment variables after tests
   - Consider using dotenv to isolate test environments

4. **File path issues**: Platform-specific path problems
   - Always use path.join() for constructing paths
   - Test on multiple platforms (or with mocked platforms)